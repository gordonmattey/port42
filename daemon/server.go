package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Daemon represents the Port 42 daemon
type Daemon struct {
	listener    net.Listener
	sessions    map[string]*Session
	mu          sync.RWMutex
	config      Config
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
	memoryStore *MemoryStore
	objectStore *ObjectStore
}

// Session represents an active possession session
type Session struct {
	ID               string       `json:"id"`
	Agent            string       `json:"agent"`
	CreatedAt        time.Time    `json:"created_at"`
	LastActivity     time.Time    `json:"last_activity"`
	State            SessionState `json:"state"`
	Messages         []Message    `json:"messages"`
	CommandGenerated *CommandSpec `json:"command_generated,omitempty"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	mu               sync.Mutex
}

// Message represents a conversation message
type Message struct {
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}


// Config holds daemon configuration
type Config struct {
	Port         string
	AIBackend    string
	MaxSessions  int
	SessionTTL   time.Duration
	MemoryPath   string
	CommandsPath string
}

// NewDaemon creates a new daemon instance
func NewDaemon(listener net.Listener, port string) *Daemon {
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".port42")
	
	// Initialize memory store
	log.Printf("🔍 Initializing memory store with base dir: %s", baseDir)
	memoryStore, err := NewMemoryStore(baseDir)
	if err != nil {
		log.Printf("❌ Failed to initialize memory store: %v", err)
		// Continue without persistence
	} else {
		log.Printf("✅ Memory store initialized successfully (not nil: %v)", memoryStore != nil)
	}
	
	// Initialize object store
	log.Printf("🗄️ Initializing object store...")
	objectStore, err := NewObjectStore(baseDir)
	if err != nil {
		log.Printf("❌ Failed to initialize object store: %v", err)
		// Continue without object store for now
	} else {
		log.Printf("✅ Object store initialized successfully")
	}
	
	return &Daemon{
		listener:    listener,
		sessions:    make(map[string]*Session),
		shutdownCh:  make(chan struct{}),
		memoryStore: memoryStore,
		objectStore: objectStore,
		config: Config{
			Port:         port,
			AIBackend:    "http://localhost:3000/api/ai", // Default, can be overridden
			MaxSessions:  100,
			SessionTTL:   24 * time.Hour,
			MemoryPath:   filepath.Join(homeDir, ".port42", "memory"),
			CommandsPath: filepath.Join(homeDir, ".port42", "commands"),
		},
	}
}

// Start begins accepting connections
func (d *Daemon) Start() {
	log.Printf("🐬 Daemon starting with config: %+v", d.config)
	
	// Load recent sessions from disk
	if d.memoryStore != nil {
		d.loadRecentSessions()
	}
	
	// Start session cleanup goroutine
	d.wg.Add(1)
	go d.cleanupSessions()
	
	// Accept connections
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.shutdownCh:
				return
			default:
				log.Printf("Error accepting connection: %v", err)
				continue
			}
		}
		
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			d.handleConnection(conn)
		}()
	}
}

// Shutdown gracefully stops the daemon
func (d *Daemon) Shutdown() {
	log.Println("🐬 Daemon shutting down...")
	close(d.shutdownCh)
	d.listener.Close()
	d.wg.Wait()
	log.Println("🐬 Daemon stopped")
}

// handleConnection processes a single connection
func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	clientAddr := conn.RemoteAddr().String()
	log.Printf("◊ New consciousness connected from %s", clientAddr)
	
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)
	
	// Read JSON request
	var req Request
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error decoding request from %s: %v", clientAddr, err)
		resp := Response{
			ID:      "error",
			Success: false,
			Error:   "Invalid JSON request",
		}
		encoder.Encode(resp)
		return
	}
	
	log.Printf("◊ Request [%s] type: %s", req.ID, req.Type)
	
	// Process request
	resp := d.handleRequest(req)
	
	// Send response
	if err := encoder.Encode(resp); err != nil {
		log.Printf("Error encoding response to %s: %v", clientAddr, err)
		return
	}
	
	log.Printf("◊ Response sent [%s] success: %v", resp.ID, resp.Success)
	log.Printf("◊ Consciousness disconnected: %s", clientAddr)
}

// handleRequest routes requests to appropriate handlers
func (d *Daemon) handleRequest(req Request) Response {
	switch req.Type {
	case RequestStatus:
		return d.handleStatus(req)
	case RequestPossess:
		return d.handlePossess(req)
	case RequestList:
		return d.handleList(req)
	case RequestMemory:
		return d.handleMemory(req)
	case RequestEnd:
		return d.handleEnd(req)
	case "ping":
		// Simple ping handler for connection checks
		return NewResponse(req.ID, true)
	default:
		resp := NewResponse(req.ID, false)
		resp.SetError(fmt.Sprintf("Unknown request type: %s", req.Type))
		return resp
	}
}

// Session management methods
func (d *Daemon) getOrCreateSession(sessionID, agent string) *Session {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Step 1: Check in-memory sessions
	if session, exists := d.sessions[sessionID]; exists {
		// Update last activity
		session.LastActivity = time.Now()
		if session.State == SessionIdle {
			session.State = SessionActive
			log.Printf("🔄 Session %s reactivated from memory", sessionID)
		}
		return session
	}
	
	// Step 2: Check on disk (NEW)
	if d.memoryStore != nil {
		if persistedSession, err := d.memoryStore.LoadSession(sessionID); err == nil {
			// Convert from PersistentSession to Session
			session := &Session{
				ID:               persistedSession.ID,
				Agent:            persistedSession.Agent,
				CreatedAt:        persistedSession.CreatedAt,
				LastActivity:     time.Now(), // Update to current time
				State:            SessionActive, // Reactivate session
				Messages:         persistedSession.Messages,
				CommandGenerated: nil,
				IdleTimeout:      30 * time.Minute,
			}
			
			// Convert command info if exists
			if persistedSession.CommandGenerated != nil {
				// Note: CommandGenerationInfo only stores basic info (name, path, created_at)
				// The full CommandSpec is not persisted, just tracking that a command was generated
				session.CommandGenerated = &CommandSpec{
					Name: persistedSession.CommandGenerated.Name,
					// Other fields would need to be loaded from the actual command file if needed
				}
			}
			
			// Add to active sessions
			d.sessions[sessionID] = session
			
			log.Printf("🔄 Session %s restored from disk (%d messages)", 
				sessionID, len(session.Messages))
			return session
		}
	}
	
	// Step 3: Create new session (existing logic)
	now := time.Now()
	session := &Session{
		ID:           sessionID,
		Agent:        agent,
		CreatedAt:    now,
		LastActivity: now,
		State:        SessionActive,
		Messages:     []Message{},
		IdleTimeout:  30 * time.Minute, // Default 30 minutes
	}
	
	d.sessions[sessionID] = session
	log.Printf("📊 Session added to map. Current map size: %d", len(d.sessions))
	
	// Save new session to disk
	log.Printf("🔍 Memory store check: memoryStore != nil: %v", d.memoryStore != nil)
	if d.memoryStore != nil {
		log.Printf("💾 Queuing save for new session %s", sessionID)
		go func() {
			log.Printf("🏃 Goroutine started for saving session %s", sessionID)
			if err := d.memoryStore.SaveSession(session); err != nil {
				log.Printf("❌ Failed to save new session: %v", err)
			} else {
				log.Printf("✅ Successfully saved session %s", sessionID)
			}
		}()
	} else {
		log.Printf("⚠️  Memory store is nil, skipping save")
	}
	
	log.Printf("✨ New session created: %s with agent %s", sessionID, agent)
	return session
}

func (d *Daemon) getSession(sessionID string) (*Session, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	session, exists := d.sessions[sessionID]
	return session, exists
}

// loadRecentSessions loads active/idle sessions from disk on startup
func (d *Daemon) loadRecentSessions() {
	sessions, err := d.memoryStore.LoadRecentSessions(1) // Last 24 hours
	if err != nil {
		log.Printf("Failed to load recent sessions: %v", err)
		return
	}
	
	d.mu.Lock()
	defer d.mu.Unlock()
	
	loaded := 0
	for _, ps := range sessions {
		// Only load active or idle sessions
		if ps.State == SessionActive || ps.State == SessionIdle {
			session := &Session{
				ID:               ps.ID,
				Agent:            ps.Agent,
				CreatedAt:        ps.CreatedAt,
				LastActivity:     ps.LastActivity,
				State:            ps.State,
				Messages:         ps.Messages,
				CommandGenerated: nil,
				IdleTimeout:      30 * time.Minute,
			}
			
			// Convert command info if exists
			if ps.CommandGenerated != nil {
				session.CommandGenerated = &CommandSpec{
					Name:        ps.CommandGenerated.Name,
					Description: "", // Not stored in persistent format
					Implementation: "", // Not needed after generation
					Language:    "",
				}
			}
			
			d.sessions[ps.ID] = session
			loaded++
		}
	}
	
	if loaded > 0 {
		log.Printf("📚 Loaded %d sessions from disk", loaded)
	}
}

func (d *Daemon) endSession(sessionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if session, exists := d.sessions[sessionID]; exists {
		session.State = SessionCompleted
		log.Printf("◊ Session ended: %s", sessionID)
	}
}

// cleanupSessions manages session lifecycle based on activity
func (d *Daemon) cleanupSessions() {
	defer d.wg.Done()
	
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			now := time.Now()
			
			for id, session := range d.sessions {
				session.mu.Lock()
				
				timeSinceActivity := now.Sub(session.LastActivity)
				
				switch session.State {
				case SessionActive:
					// Check if session should go idle
					if timeSinceActivity > session.IdleTimeout {
						session.State = SessionIdle
						log.Printf("⏸️  Session %s is now idle (no activity for %v)", id, session.IdleTimeout)
						
						// Save idle state to disk
						if d.memoryStore != nil {
							go d.memoryStore.SaveSession(session)
						}
					}
					
				case SessionIdle:
					// Check if session should be abandoned (2x idle timeout)
					if timeSinceActivity > session.IdleTimeout*2 {
						session.State = SessionAbandoned
						log.Printf("🚪 Session %s abandoned (idle for %v)", id, timeSinceActivity)
						
						// Save final state and remove from memory
						if d.memoryStore != nil {
							go d.memoryStore.SaveSession(session)
						}
						delete(d.sessions, id)
					}
					
				case SessionCompleted, SessionAbandoned:
					// Remove from active memory (already saved to disk)
					delete(d.sessions, id)
				}
				
				session.mu.Unlock()
			}
			
			d.mu.Unlock()
			
		case <-d.shutdownCh:
			// Save all active sessions before shutdown
			d.mu.RLock()
			for _, session := range d.sessions {
				if d.memoryStore != nil && (session.State == SessionActive || session.State == SessionIdle) {
					d.memoryStore.SaveSession(session)
				}
			}
			d.mu.RUnlock()
			return
		}
	}
}

// Handler methods (moved from main.go, now with daemon context)
func (d *Daemon) handleStatus(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	uptime := time.Since(startTime).Round(time.Second).String()
	
	d.mu.RLock()
	activeSessions := 0
	for _, session := range d.sessions {
		if session.State == SessionActive {
			activeSessions++
		}
	}
	d.mu.RUnlock()
	
	status := StatusData{
		Status:   "swimming",
		Port:     d.config.Port,
		Sessions: activeSessions,
		Uptime:   uptime,
		Dolphins: "🐬🐬🐬 laughing in the digital waves",
	}
	
	resp.SetData(status)
	return resp
}

func (d *Daemon) handlePossess(req Request) Response {
	// Use the AI-powered possession handler
	return d.handlePossessWithAI(req)
}

func (d *Daemon) handleList(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	// Read from commands directory
	homeDir, _ := os.UserHomeDir()
	cmdDir := filepath.Join(homeDir, ".port42", "commands")
	
	commands := []string{}
	
	// Check if directory exists
	if _, err := os.Stat(cmdDir); err == nil {
		// Read all files in commands directory
		files, err := os.ReadDir(cmdDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					commands = append(commands, file.Name())
				}
			}
		}
	}
	
	list := ListData{
		Commands: commands,
	}
	
	resp.SetData(list)
	return resp
}

func (d *Daemon) handleMemory(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	// Check if payload contains a session ID
	var payload struct {
		SessionID string `json:"session_id,omitempty"`
	}
	
	if req.Payload != nil && len(req.Payload) > 0 {
		if err := json.Unmarshal(req.Payload, &payload); err == nil && payload.SessionID != "" {
			// Handle specific session request
			return d.handleMemoryShow(req, payload.SessionID)
		}
	}
	
	// Handle list all sessions
	d.mu.RLock()
	log.Printf("🔍 Memory endpoint: Current map size: %d", len(d.sessions))
	log.Printf("🔍 Session IDs in map:")
	for id := range d.sessions {
		log.Printf("   - %s", id)
	}
	
	// Create summaries for active sessions
	activeSummaries := make([]SessionSummary, 0, len(d.sessions))
	for _, session := range d.sessions {
		activeSummaries = append(activeSummaries, SessionSummary{
			ID:           session.ID,
			Agent:        session.Agent,
			CreatedAt:    session.CreatedAt,
			LastActivity: session.LastActivity,
			MessageCount: len(session.Messages),
			State:        string(session.State),
		})
	}
	d.mu.RUnlock()
	
	// Get recent sessions from disk if memory store available
	var recentSummaries []SessionSummary
	var stats *MemoryStats
	
	if d.memoryStore != nil {
		// Load last 7 days of sessions
		if sessions, err := d.memoryStore.LoadRecentSessions(7); err == nil {
			// Convert to summaries
			recentSummaries = make([]SessionSummary, 0, len(sessions))
			for _, ps := range sessions {
				recentSummaries = append(recentSummaries, SessionSummary{
					ID:           ps.ID,
					Agent:        ps.Agent,
					CreatedAt:    ps.CreatedAt,
					LastActivity: ps.LastActivity,
					MessageCount: len(ps.Messages),
					State:        string(ps.State),
				})
			}
		}
		stats = d.memoryStore.GetStats()
	}
	
	data := map[string]interface{}{
		"active_sessions": activeSummaries,
		"active_count":    len(activeSummaries),
		"recent_sessions": recentSummaries,
		"stats":           stats,
		"uptime":          time.Since(startTime).String(),
	}
	
	resp.SetData(data)
	return resp
}

func (d *Daemon) handleEnd(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	d.endSession(req.ID)
	
	data := map[string]string{
		"message": "Session crystallized. The dolphins remember...",
	}
	
	resp.SetData(data)
	return resp
}

// handleMemoryShow returns full details for a specific session
func (d *Daemon) handleMemoryShow(req Request, sessionID string) Response {
	resp := NewResponse(req.ID, true)
	
	// First check in-memory sessions
	d.mu.RLock()
	if session, exists := d.sessions[sessionID]; exists {
		d.mu.RUnlock()
		
		data := map[string]interface{}{
			"id":           session.ID,
			"agent":        session.Agent,
			"state":        session.State,
			"created_at":   session.CreatedAt,
			"last_activity": session.LastActivity,
			"messages":     session.Messages,
			"command_generated": session.CommandGenerated,
		}
		resp.SetData(data)
		return resp
	}
	d.mu.RUnlock()
	
	// Try to load from disk
	if d.memoryStore != nil {
		if persistedSession, err := d.memoryStore.LoadSession(sessionID); err == nil {
			data := map[string]interface{}{
				"id":           persistedSession.ID,
				"agent":        persistedSession.Agent,
				"state":        persistedSession.State,
				"created_at":   persistedSession.CreatedAt,
				"last_activity": persistedSession.LastActivity,
				"updated_at":   persistedSession.UpdatedAt,
				"messages":     persistedSession.Messages,
				"command_generated": persistedSession.CommandGenerated,
				"metadata":     persistedSession.Metadata,
			}
			resp.SetData(data)
			return resp
		}
	}
	
	// Session not found
	resp.SetError(fmt.Sprintf("Session '%s' not found", sessionID))
	return resp
}

// Command generation functionality
func (d *Daemon) generateCommand(spec *CommandSpec) error {
	log.Printf("🌊 Crystallizing command '%s'...", spec.Name)
	
	// Check for dependencies
	if len(spec.Dependencies) > 0 {
		log.Printf("📦 Command requires dependencies: %v", spec.Dependencies)
	}
	
	// Generate dependency check code based on language
	var depCheckCode string
	log.Printf("🔍 Language: %s, Dependencies: %v", spec.Language, spec.Dependencies)
	if len(spec.Dependencies) > 0 && spec.Language == "bash" {
		// Only add bash dependency check for bash scripts
		depCheckCode = d.generateDependencyCheck(spec.Dependencies)
		log.Printf("✅ Adding bash dependency check for bash script")
	} else if len(spec.Dependencies) > 0 {
		log.Printf("⚠️  Skipping dependency check for %s script with deps: %v", spec.Language, spec.Dependencies)
	}
	
	// Use implementation as-is - Go's json.Unmarshal already handled unescaping
	implementation := spec.Implementation
	
	// No need for any unescaping - JSON parsing already converted:
	// - \\n to \n (for line breaks)
	// - \\t to \t (for tabs)  
	// - \\\" to \" (for quotes)
	// The implementation should already be valid code!
	
	// Remove any shebang from the implementation (we'll add the correct one)
	lines := strings.Split(implementation, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		implementation = strings.Join(lines[1:], "\n")
	}
	
	// Determine file extension based on language
	var code string
	switch spec.Language {
	case "python":
		code = fmt.Sprintf("#!/usr/bin/env python3\n# Generated by Port 42 - %s\n# %s\n\n%s\n%s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			spec.Description,
			depCheckCode,
			implementation)
	case "node", "javascript":
		code = fmt.Sprintf("#!/usr/bin/env node\n// Generated by Port 42 - %s\n// %s\n\n%s\n%s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			spec.Description,
			depCheckCode,
			implementation)
	default: // bash
		code = fmt.Sprintf("#!/bin/bash\n# Generated by Port 42 - %s\n# %s\n\n%s%s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			spec.Description,
			depCheckCode,
			implementation)
	}
	
	// Store in object store
	if d.objectStore == nil {
		return fmt.Errorf("object store not initialized")
	}
	
	// Create metadata for the command
	metadata := &Metadata{
		Type:        "command",
		Title:       spec.Name,
		Description: spec.Description,
		Tags:        extractTags(spec),
		Session:     spec.SessionID,
		Agent:       spec.Agent,
		Lifecycle:   "active",
		Importance:  "medium",
		Paths: []string{
			fmt.Sprintf("commands/%s", spec.Name),
			fmt.Sprintf("by-date/%s/%s", time.Now().Format("2006-01-02"), spec.Name),
			fmt.Sprintf("by-type/command/%s", spec.Name),
		},
	}
	
	// Add session path if we have a session ID
	if spec.SessionID != "" {
		metadata.Paths = append(metadata.Paths, 
			fmt.Sprintf("memory/sessions/%s/generated/%s", spec.SessionID, spec.Name))
	}
	
	// Store content with metadata
	objectID, err := d.objectStore.StoreWithMetadata([]byte(code), metadata)
	if err != nil {
		return fmt.Errorf("failed to store command in object store: %v", err)
	}
	
	log.Printf("✨ Command '%s' crystallized in object store: %s", spec.Name, objectID[:12]+"...")
	log.Printf("🌊 Virtual paths: %v", metadata.Paths)
	
	// Log to memory (simple for now)
	d.logCommandGeneration(spec)
	
	return nil
}

// extractTags extracts relevant tags from a command spec
func extractTags(spec *CommandSpec) []string {
	tags := []string{spec.Language}
	
	// Add language-specific tags
	switch spec.Language {
	case "python":
		tags = append(tags, "script", "python3")
	case "node", "javascript":
		tags = append(tags, "script", "nodejs", "javascript")
	default:
		tags = append(tags, "script", "bash", "shell")
	}
	
	// Add dependency tags
	for _, dep := range spec.Dependencies {
		tags = append(tags, dep)
	}
	
	// Extract keywords from name and description
	words := strings.Fields(spec.Name + " " + spec.Description)
	for _, word := range words {
		word = strings.ToLower(word)
		// Add meaningful words as tags (skip common words)
		if len(word) > 3 && !isCommonWord(word) {
			tags = append(tags, word)
		}
	}
	
	return tags
}

// isCommonWord checks if a word is too common to be a useful tag
func isCommonWord(word string) bool {
	common := map[string]bool{
		"that": true, "this": true, "with": true, "from": true,
		"have": true, "been": true, "will": true, "your": true,
		"what": true, "when": true, "where": true, "which": true,
	}
	return common[word]
}

// Generate dependency check code for commands
func (d *Daemon) generateDependencyCheck(deps []string) string {
	if len(deps) == 0 {
		return ""
	}
	
	// Create dependency install script
	d.createDependencyInstaller(deps)
	
	// Bash dependency check
	check := `# Dependency check
missing_deps=()
`
	for _, dep := range deps {
		check += fmt.Sprintf("if ! command -v %s &> /dev/null; then\n", dep)
		check += fmt.Sprintf("  missing_deps+=(%s)\n", dep)
		check += "fi\n"
	}
	
	check += `
if [ ${#missing_deps[@]} -ne 0 ]; then
  echo "❌ Missing dependencies: ${missing_deps[*]}"
  echo ""
  echo "To install dependencies, run:"
  echo "  ~/.port42/install-deps.sh ${missing_deps[*]}"
  echo ""
  echo "Or install manually:"
  for dep in "${missing_deps[@]}"; do
    case "$dep" in
      lolcat) echo "  brew install lolcat  # or: gem install lolcat" ;;
      tree) echo "  brew install tree    # or: apt-get install tree" ;;
      figlet) echo "  brew install figlet  # or: apt-get install figlet" ;;
      jq) echo "  brew install jq      # or: apt-get install jq" ;;
      rg|ripgrep) echo "  brew install ripgrep # or: cargo install ripgrep" ;;
      fzf) echo "  brew install fzf     # or: git clone https://github.com/junegunn/fzf.git" ;;
      *) echo "  # Install $dep using your package manager" ;;
    esac
  done
  exit 1
fi

`
	return check
}

// Create a dependency installer script
func (d *Daemon) createDependencyInstaller(deps []string) {
	homeDir, _ := os.UserHomeDir()
	installerPath := filepath.Join(homeDir, ".port42", "install-deps.sh")
	
	installer := `#!/bin/bash
# Port 42 Dependency Installer
# Generated automatically to help install command dependencies

set -e

echo "🐬 Port 42 Dependency Installer"
echo ""

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
  OS="macos"
elif [[ -f /etc/debian_version ]]; then
  OS="debian"
elif [[ -f /etc/redhat-release ]]; then
  OS="redhat"
else
  OS="unknown"
fi

# Function to install a dependency
install_dep() {
  local dep=$1
  echo "📦 Installing $dep..."
  
  case "$OS" in
    macos)
      if command -v brew &> /dev/null; then
        brew install "$dep" || true
      else
        echo "❌ Homebrew not found. Please install: https://brew.sh"
        return 1
      fi
      ;;
    debian)
      sudo apt-get update && sudo apt-get install -y "$dep" || true
      ;;
    redhat)
      sudo yum install -y "$dep" || true
      ;;
    *)
      echo "❌ Unknown OS. Please install $dep manually."
      return 1
      ;;
  esac
}

# Install each dependency passed as argument
for dep in "$@"; do
  if ! command -v "$dep" &> /dev/null; then
    install_dep "$dep"
  else
    echo "✅ $dep is already installed"
  fi
done

echo ""
echo "✨ Installation complete!"
`
	
	os.WriteFile(installerPath, []byte(installer), 0755)
}


// Ensure ~/.port42/commands is in PATH
func (d *Daemon) ensureCommandsInPath() {
	homeDir, _ := os.UserHomeDir()
	cmdDir := filepath.Join(homeDir, ".port42", "commands")
	
	// Check if already in PATH
	path := os.Getenv("PATH")
	if strings.Contains(path, cmdDir) {
		return
	}
	
	// Create or update shell config hint file
	hintPath := filepath.Join(homeDir, ".port42", "setup-hint.txt")
	hint := fmt.Sprintf(`
To use Port 42 generated commands, add this to your shell config:

export PATH="$PATH:%s"

For bash: echo 'export PATH="$PATH:%s"' >> ~/.bashrc
For zsh:  echo 'export PATH="$PATH:%s"' >> ~/.zshrc

Then restart your shell or run: source ~/.bashrc (or ~/.zshrc)
`, cmdDir, cmdDir, cmdDir)
	
	os.WriteFile(hintPath, []byte(hint), 0644)
	
	log.Printf("💡 Add %s to your PATH to use generated commands", cmdDir)
	log.Printf("   See %s for instructions", hintPath)
}

// Simple command generation logging
func (d *Daemon) logCommandGeneration(spec *CommandSpec) {
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".port42", "command-history.json")
	
	// Read existing history
	var history []map[string]interface{}
	if data, err := os.ReadFile(logPath); err == nil {
		json.Unmarshal(data, &history)
	}
	
	// Add new entry
	entry := map[string]interface{}{
		"name":        spec.Name,
		"description": spec.Description,
		"language":    spec.Language,
		"generated":   time.Now().Format(time.RFC3339),
	}
	history = append(history, entry)
	
	// Write back
	if data, err := json.MarshalIndent(history, "", "  "); err == nil {
		os.WriteFile(logPath, data, 0644)
	}
}