package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// AgentSessions manages last session tracking per agent
type AgentSessions struct {
	mu       sync.RWMutex
	sessions map[string]string // agent -> sessionID
	filePath string
}

// NewAgentSessions creates a new agent session tracker
func NewAgentSessions(baseDir string) *AgentSessions {
	return &AgentSessions{
		sessions: make(map[string]string),
		filePath: filepath.Join(baseDir, "agent_sessions.json"),
	}
}

// Load reads agent sessions from disk
func (as *AgentSessions) Load() error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	data, err := os.ReadFile(as.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, not an error
			return nil
		}
		return fmt.Errorf("failed to read agent sessions: %w", err)
	}
	
	if err := json.Unmarshal(data, &as.sessions); err != nil {
		return fmt.Errorf("failed to parse agent sessions: %w", err)
	}
	
	log.Printf("📌 [AGENT_SESSIONS] Loaded sessions for %d agents", len(as.sessions))
	return nil
}

// Save writes agent sessions to disk
func (as *AgentSessions) Save() error {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	data, err := json.MarshalIndent(as.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent sessions: %w", err)
	}
	
	if err := os.WriteFile(as.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write agent sessions: %w", err)
	}
	
	return nil
}

// GetLastSession returns the last session for an agent
func (as *AgentSessions) GetLastSession(agent string) (string, bool) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	
	sessionID, exists := as.sessions[agent]
	return sessionID, exists
}

// SetLastSession updates the last session for an agent
func (as *AgentSessions) SetLastSession(agent, sessionID string) error {
	as.mu.Lock()
	defer as.mu.Unlock()
	
	as.sessions[agent] = sessionID
	
	// Save immediately for persistence
	data, err := json.MarshalIndent(as.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent sessions: %w", err)
	}
	
	if err := os.WriteFile(as.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write agent sessions: %w", err)
	}
	
	log.Printf("📌 [AGENT_SESSIONS] Updated %s -> %s", agent, sessionID)
	return nil
}

// Storage provides unified storage for all Port 42 data
type Storage struct {
	baseDir     string
	objectsDir  string
	metadataDir string
	
	// Session index for quick lookups
	sessionIndex map[string]SessionReference
	indexMutex   sync.RWMutex
	
	// Agent-specific session tracking
	agentSessions *AgentSessions
	
	// Relations integration for virtual filesystem
	relationStore RelationStore
	
	// Stats
	stats StorageStats
}

// StorageStats tracks storage metrics
type StorageStats struct {
	TotalSessions      int       `json:"total_sessions"`
	ActiveSessions     int       `json:"active_sessions"`
	CompletedSessions  int       `json:"completed_sessions"`
	TotalObjects       int       `json:"total_objects"`
	StorageSize        int64     `json:"storage_size"`
	LastUpdated        time.Time `json:"last_updated"`
}

// NewStorage creates a new unified storage instance
func NewStorage(baseDir string, relationStore RelationStore) (*Storage, error) {
	objectsDir := filepath.Join(baseDir, "objects")
	metadataDir := filepath.Join(baseDir, "metadata")
	
	// Check if directories exist (they should be created by installer)
	if _, err := os.Stat(objectsDir); os.IsNotExist(err) {
		log.Printf("⚠️  Warning: objects directory missing at %s", objectsDir)
		log.Printf("⚠️  Creating it now, but this indicates Port 42 wasn't installed properly")
		if err := os.MkdirAll(objectsDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create objects directory: %w", err)
		}
	}
	if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
		log.Printf("⚠️  Warning: metadata directory missing at %s", metadataDir)
		log.Printf("⚠️  Creating it now, but this indicates Port 42 wasn't installed properly")
		if err := os.MkdirAll(metadataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create metadata directory: %w", err)
		}
	}
	
	// Initialize agent sessions
	agentSessions := NewAgentSessions(baseDir)
	if err := agentSessions.Load(); err != nil {
		log.Printf("⚠️ [STORAGE] Failed to load agent sessions: %v", err)
		// Continue anyway, will create new file on first save
	}
	
	s := &Storage{
		baseDir:       baseDir,
		objectsDir:    objectsDir,
		metadataDir:   metadataDir,
		sessionIndex:  make(map[string]SessionReference),
		agentSessions: agentSessions,
		relationStore: relationStore,
		stats:         StorageStats{LastUpdated: time.Now()},
	}
	
	// Load session index
	if err := s.loadSessionIndex(); err != nil {
		log.Printf("Warning: Failed to load session index: %v", err)
		// Continue anyway, we'll rebuild it
	}
	
	return s, nil
}

// ==================== Core Object Storage ====================

// Store saves content and returns its hash ID
func (s *Storage) Store(content []byte) (string, error) {
	// Calculate SHA256 hash
	hash := sha256.Sum256(content)
	id := hex.EncodeToString(hash[:])
	
	log.Printf("🔍 [STORAGE] Store called: size=%d, id=%s", len(content), id[:12]+"...")
	
	// Store in git-like structure: objects/3a/4f/2b8c9d...
	dir := filepath.Join(s.objectsDir, id[:2], id[2:4])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create object directory: %w", err)
	}
	
	path := filepath.Join(dir, id[4:])
	
	// Check if object already exists
	if _, err := os.Stat(path); err == nil {
		log.Printf("🔍 [STORAGE] Object already exists: %s", id[:12]+"...")
		return id, nil
	}
	
	// Write content
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write object: %w", err)
	}
	
	log.Printf("✅ [STORAGE] New object stored: %s at %s", id[:12]+"...", path)
	return id, nil
}

// Read retrieves content by hash ID
func (s *Storage) Read(id string) ([]byte, error) {
	// Handle special relation IDs
	if strings.HasPrefix(id, "relation:") {
		relationID := strings.TrimPrefix(id, "relation:")
		if s.relationStore != nil {
			if relation, err := s.relationStore.Load(relationID); err == nil {
				// Return relation as JSON
				return json.MarshalIndent(relation, "", "  ")
			}
		}
		return nil, fmt.Errorf("relation not found: %s", relationID)
	}
	
	if len(id) < 4 {
		return nil, fmt.Errorf("invalid object ID: %s", id)
	}
	
	path := filepath.Join(s.objectsDir, id[:2], id[2:4], id[4:])
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	
	return content, nil
}

// GetPath returns the filesystem path for an object
func (s *Storage) GetPath(id string) string {
	if len(id) < 4 {
		return ""
	}
	return filepath.Join(s.objectsDir, id[:2], id[2:4], id[4:])
}

// ==================== Metadata Management ====================

// SaveMetadata stores metadata for an object
func (s *Storage) SaveMetadata(meta *Metadata) error {
	if meta.ID == "" {
		return fmt.Errorf("metadata ID cannot be empty")
	}
	
	// Update timestamps
	if meta.Created.IsZero() {
		meta.Created = time.Now()
	}
	meta.Modified = time.Now()
	meta.Accessed = time.Now()
	
	// Set defaults
	if meta.Lifecycle == "" {
		meta.Lifecycle = "draft"
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	// Store in metadata directory
	metaPath := filepath.Join(s.metadataDir, meta.ID+".json")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	
	return nil
}

// LoadMetadata retrieves metadata for an object
func (s *Storage) LoadMetadata(id string) (*Metadata, error) {
	metaPath := filepath.Join(s.metadataDir, id+".json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("metadata not found for object: %s", id)
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	
	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	// Update access time
	meta.Accessed = time.Now()
	s.SaveMetadata(&meta)
	
	return &meta, nil
}

// StoreWithMetadata stores content with associated metadata
func (s *Storage) StoreWithMetadata(content []byte, meta *Metadata) (string, error) {
	log.Printf("🔍 [STORAGE] StoreWithMetadata called with type=%s, paths=%v", meta.Type, meta.Paths)
	
	// Store content
	id, err := s.Store(content)
	if err != nil {
		return "", err
	}
	
	// Update metadata with ID and size
	meta.ID = id
	meta.Size = int64(len(content))
	
	log.Printf("🔍 [STORAGE] Saving metadata for object %s", id[:12]+"...")
	
	// Save metadata
	if err := s.SaveMetadata(meta); err != nil {
		return "", fmt.Errorf("failed to save metadata: %w", err)
	}
	
	log.Printf("✅ [STORAGE] StoreWithMetadata complete: id=%s, type=%s", id[:12]+"...", meta.Type)
	return id, nil
}

// ==================== Session Management ====================

// SaveSession saves a session to storage
func (s *Storage) SaveSession(session *Session) error {
	log.Printf("🔍 [STORAGE] SaveSession starting for %s (messages=%d, state=%s)", 
		session.ID, len(session.Messages), session.State)
	
	s.indexMutex.Lock()
	defer s.indexMutex.Unlock()
	
	// Check if session already exists in index
	if existing, exists := s.sessionIndex[session.ID]; exists {
		log.Printf("🔍 [STORAGE] Session %s already exists with object ID %s", 
			session.ID, existing.ObjectID[:12]+"...")
	}
	
	// Create persistent session
	ps := &PersistentSession{
		ID:           session.ID,
		Agent:        session.Agent,
		State:        session.State,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    time.Now(),
		LastActivity: session.LastActivity,
		Messages:     session.Messages,
		Metadata: map[string]interface{}{
			"agent": session.Agent,
		},
	}
	
	// Add command info if generated
	if session.CommandGenerated != nil {
		ps.CommandGenerated = &CommandGenerationInfo{
			Name:      session.CommandGenerated.Name,
			CreatedAt: time.Now(),
			Path:      fmt.Sprintf("commands/%s", session.CommandGenerated.Name),
		}
	}
	
	// Serialize session
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}
	
	// Create metadata
	metadata := &Metadata{
		Type:        "session",
		Title:       fmt.Sprintf("Session %s", session.ID),
		Description: fmt.Sprintf("AI conversation with %s", session.Agent),
		Tags:        extractSessionTags(session),
		Session:     session.ID,
		Agent:       session.Agent,
		Lifecycle:   mapStateToLifecycle(session.State),
		Paths: []string{
			fmt.Sprintf("/memory/%s", session.ID),                    // Direct memory access
			fmt.Sprintf("/memory/sessions/%s", session.ID),           // Type-specific access
			fmt.Sprintf("/memory/sessions/by-date/%s/%s",            // Date organization
				session.CreatedAt.Format("2006-01-02"), session.ID),
			fmt.Sprintf("/memory/sessions/by-agent/%s/%s",           // Agent organization
				cleanAgentName(session.Agent), session.ID),
			fmt.Sprintf("/by-date/%s/memory/%s",                     // Global date view
				session.CreatedAt.Format("2006-01-02"), session.ID),
			fmt.Sprintf("/by-agent/%s/memory/%s",                    // Global agent view
				cleanAgentName(session.Agent), session.ID),
		},
	}
	
	// Store in object store
	objectID, err := s.StoreWithMetadata(data, metadata)
	if err != nil {
		return fmt.Errorf("failed to store session: %v", err)
	}
	
	// Update index
	s.sessionIndex[session.ID] = SessionReference{
		ObjectID:         objectID,
		SessionID:        session.ID,
		Agent:            session.Agent,
		CreatedAt:        session.CreatedAt,
		LastUpdated:      time.Now(),
		CommandGenerated: session.CommandGenerated != nil,
		State:            string(session.State),
		MessageCount:     len(session.Messages),
	}
	
	// Update stats
	s.updateStats()
	
	// Save index
	if err := s.saveSessionIndex(); err != nil {
		log.Printf("Warning: Failed to save session index: %v", err)
	}
	
	// Update last session tracker for this agent
	if session.Agent != "" {
		if err := s.UpdateLastSession(session.Agent, session.ID); err != nil {
			log.Printf("Warning: Failed to update last session for agent %s: %v", session.Agent, err)
		}
	}
	
	log.Printf("✅ [STORAGE] Session %s saved with object ID %s", session.ID, objectID[:12]+"...")
	return nil
}

// LoadSession loads a session from storage
func (s *Storage) LoadSession(sessionID string) (*Session, error) {
	s.indexMutex.RLock()
	ref, exists := s.sessionIndex[sessionID]
	s.indexMutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	// Load object
	data, err := s.Read(ref.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to read session object: %v", err)
	}
	
	// Unmarshal persistent session
	var ps PersistentSession
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %v", err)
	}
	
	// Convert to Session
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
			Name: ps.CommandGenerated.Name,
		}
	}
	
	return session, nil
}

// LoadRecentSessions loads sessions from the last N days
func (s *Storage) LoadRecentSessions(days int) ([]*PersistentSession, error) {
	s.indexMutex.RLock()
	defer s.indexMutex.RUnlock()
	
	cutoff := time.Now().AddDate(0, 0, -days)
	var sessions []*PersistentSession
	
	for _, ref := range s.sessionIndex {
		if ref.CreatedAt.After(cutoff) {
			// Load session data
			data, err := s.Read(ref.ObjectID)
			if err != nil {
				log.Printf("Warning: Failed to load session %s: %v", ref.SessionID, err)
				continue
			}
			
			var ps PersistentSession
			if err := json.Unmarshal(data, &ps); err != nil {
				log.Printf("Warning: Failed to unmarshal session %s: %v", ref.SessionID, err)
				continue
			}
			
			sessions = append(sessions, &ps)
		}
	}
	
	return sessions, nil
}

// GetLastSession returns the ID of the most recently active session for an agent
func (s *Storage) GetLastSession(agent string) (string, error) {
	if agent == "" {
		return "", fmt.Errorf("agent parameter required")
	}
	
	// Normalize agent name (remove @ if present)
	agent = strings.TrimPrefix(agent, "@")
	
	// Check agent-specific sessions
	sessionID, exists := s.agentSessions.GetLastSession(agent)
	if !exists {
		return "", fmt.Errorf("no sessions found for agent %s", agent)
	}
	
	// Verify session still exists
	s.indexMutex.RLock()
	defer s.indexMutex.RUnlock()
	
	if _, exists := s.sessionIndex[sessionID]; !exists {
		// Session no longer exists, clean up
		s.agentSessions.SetLastSession(agent, "")
		return "", fmt.Errorf("session %s no longer exists", sessionID)
	}
	
	log.Printf("🔍 [STORAGE] Retrieved last session for %s: %s", agent, sessionID)
	return sessionID, nil
}

// UpdateLastSession updates a marker for the last active session for an agent
// Note: This should only be called when already holding the indexMutex lock
func (s *Storage) UpdateLastSession(agent, sessionID string) error {
	if agent == "" {
		return fmt.Errorf("agent parameter required")
	}
	
	// Normalize agent name (remove @ if present)
	agent = strings.TrimPrefix(agent, "@")
	
	// Update agent-specific session
	return s.agentSessions.SetLastSession(agent, sessionID)
}

// ==================== Command Management ====================

// StoreCommand stores a command with metadata and creates symlink
func (s *Storage) StoreCommand(spec *CommandSpec, code string) error {
	log.Printf("🔍 [STORAGE] StoreCommand for '%s' (session=%s)", spec.Name, spec.SessionID)
	
	// Create metadata
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
			fmt.Sprintf("/commands/%s", spec.Name),
			fmt.Sprintf("/by-date/%s/%s", time.Now().Format("2006-01-02"), spec.Name),
			fmt.Sprintf("/by-type/command/%s", spec.Name),
		},
	}
	
	// Add session path if we have a session ID
	if spec.SessionID != "" {
		metadata.Paths = append(metadata.Paths, 
			fmt.Sprintf("/memory/%s/generated/%s", spec.SessionID, spec.Name),
			fmt.Sprintf("/memory/sessions/%s/generated/%s", spec.SessionID, spec.Name))
	}
	
	// Store content with metadata
	objectID, err := s.StoreWithMetadata([]byte(code), metadata)
	if err != nil {
		return fmt.Errorf("failed to store command: %v", err)
	}
	
	// Create symlink
	if err := s.CreateCommandSymlink(objectID, spec.Name); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}
	
	log.Printf("✅ [STORAGE] Command '%s' stored with ID %s", spec.Name, objectID[:12]+"...")
	return nil
}

// CreateCommandSymlink creates a symlink for command execution
func (s *Storage) CreateCommandSymlink(objID, cmdName string) error {
	homeDir, _ := os.UserHomeDir()
	cmdDir := filepath.Join(homeDir, ".port42", "commands")
	
	// Ensure commands directory exists
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		log.Printf("❌ [STORAGE] Failed to create commands directory: %v", err)
		return fmt.Errorf("failed to create commands directory: %v", err)
	}
	
	// Create symlink
	linkPath := filepath.Join(cmdDir, cmdName)
	targetPath := s.GetPath(objID)
	
	log.Printf("🔍 [STORAGE] Creating symlink: %s -> %s", linkPath, targetPath)
	
	// Remove existing symlink if it exists
	os.Remove(linkPath)
	
	// Create new symlink
	if err := os.Symlink(targetPath, linkPath); err != nil {
		log.Printf("❌ [STORAGE] Failed to create symlink: %v", err)
		return fmt.Errorf("failed to create symlink: %v", err)
	}
	
	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		log.Printf("⚠️  [STORAGE] Failed to make command executable: %v", err)
	}
	
	log.Printf("✅ [STORAGE] Symlink created successfully")
	return nil
}

// ==================== Virtual Filesystem ====================

// ResolvePath resolves a virtual path to an object ID
func (s *Storage) ResolvePath(path string) string {
	// Handle unified tools paths specially
	if s.relationStore != nil && strings.HasPrefix(path, "/tools/") {
		return s.resolveToolsPath(path)
	}
	
	// Handle enhanced commands paths - resolve to tools
	if s.relationStore != nil && strings.HasPrefix(path, "/commands/") {
		return s.resolveCommandPath(path)
	}
	
	// Handle memory paths - resolve to session objects
	if strings.HasPrefix(path, "/memory/") {
		return s.resolveMemoryPath(path)
	}
	
	// List all objects and check their metadata
	ids, err := s.List()
	if err != nil {
		log.Printf("Error listing objects: %v", err)
		return ""
	}
	
	for _, id := range ids {
		meta, err := s.LoadMetadata(id)
		if err != nil {
			continue
		}
		
		// Check if this object has the requested path
		for _, p := range meta.Paths {
			if p == path {
				return id
			}
		}
	}
	
	return ""
}

// ListPath lists entries in a virtual directory
func (s *Storage) ListPath(path string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	// Handle root directory
	if path == "/" || path == "" {
		entries = append(entries, map[string]interface{}{
			"name": "tools",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "commands",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "memory",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "artifacts",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "by-date",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "by-agent",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "by-type",
			"type": "directory",
		})
		return entries
	}
	
	// Handle unified tools paths  
	if s.relationStore != nil && strings.HasPrefix(path, "/tools") {
		return s.handleToolsPath(path)
	}
	
	// Handle enhanced commands view - show relation-backed tools
	if path == "/commands" || path == "/commands/" {
		return s.handleEnhancedCommandsView()
	}
	
	// Handle enhanced by-date view - include relations
	if strings.HasPrefix(path, "/by-date/") {
		return s.handleEnhancedByDateView(path)
	}
	
	// Handle memory generated view - show tools from specific session
	if strings.HasPrefix(path, "/memory/") && strings.Contains(path, "/generated") {
		return s.handleMemoryGeneratedView(path)
	}
	
	// Handle similar paths - semantic tool discovery
	if strings.HasPrefix(path, "/similar") {
		return s.handleSimilarView(path)
	}
	
	// List all objects and organize by virtual paths
	ids, err := s.List()
	if err != nil {
		log.Printf("Error listing objects: %v", err)
		return entries
	}
	
	pathMap := make(map[string]bool)
	
	for _, id := range ids {
		meta, err := s.LoadMetadata(id)
		if err != nil {
			continue
		}
		
		// Check each virtual path for this object
		for _, vpath := range meta.Paths {
			if strings.HasPrefix(vpath, path+"/") {
				// Extract the next component
				relative := strings.TrimPrefix(vpath, path+"/")
				parts := strings.Split(relative, "/")
				if len(parts) > 0 {
					name := parts[0]
					if !pathMap[name] {
						pathMap[name] = true
						
						// Determine if it's a directory or file
						isDir := len(parts) > 1
						entryType := "file"
						if isDir {
							entryType = "directory"
						}
						
						entry := map[string]interface{}{
							"name": name,
							"type": entryType,
						}
						
						// Add metadata for files
						if !isDir {
							entry["id"] = id
							entry["size"] = meta.Size
							entry["created"] = meta.Created
							entry["modified"] = meta.Modified
							if meta.Type != "" {
								entry["content_type"] = meta.Type
							}
						}
						
						entries = append(entries, entry)
					}
				}
			}
		}
	}
	
	return entries
}

// resolveToolsPath resolves unified tools paths to object IDs
func (s *Storage) resolveToolsPath(path string) string {
	// Remove leading /tools
	toolsPath := strings.TrimPrefix(path, "/tools")
	if toolsPath == "" {
		toolsPath = "/"
	}
	
	// Parse the tools path structure
	parts := strings.Split(strings.Trim(toolsPath, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "" // Root tools directory - no specific object
	}
	
	// Handle individual tool directory: /tools/{toolname} -> redirect to definition
	if len(parts) == 1 && parts[0] != "" {
		toolName := parts[0]
		
		// Skip organizational paths (by-name, by-transform, etc.)
		if toolName == "by-name" || toolName == "by-transform" || toolName == "spawned-by" || toolName == "ancestry" {
			return "" // These are organizational directories, not objects
		}
		
		// For individual tool directory, default to definition
		return s.resolveToolsPath("/tools/" + toolName + "/definition")
	}
	
	// Handle specific tool paths like /tools/{toolname}/definition or /tools/{toolname}/executable
	if len(parts) >= 2 {
		toolName := parts[0]
		subpath := parts[1]
		
		// Skip organizational paths (by-name, by-transform, etc.)
		if toolName == "by-name" || toolName == "by-transform" || toolName == "spawned-by" || toolName == "ancestry" {
			return "" // These are organizational directories, not objects
		}
		
		// Find the tool relation
		if s.relationStore != nil {
			if relations, err := s.relationStore.List(); err == nil {
				for _, relation := range relations {
					if relation.Type == "Tool" {
						if name, ok := relation.Properties["name"].(string); ok && name == toolName {
							// Handle different subpaths
							switch subpath {
							case "definition":
								// Return the relation as JSON
								return "relation:" + relation.ID
							case "executable":
								// Look for executable object ID in properties
								if executableID, exists := relation.Properties["executable_id"]; exists {
									if objID, ok := executableID.(string); ok && objID != "" {
										// Return the canonical object ID directly
										return objID
									}
								}
								
								// Fallback: if only executable content is stored (legacy), convert it
								if executable, exists := relation.Properties["executable"]; exists {
									if execStr, ok := executable.(string); ok && execStr != "" {
										// Store the executable content and return its ID
										objID, err := s.Store([]byte(execStr))
										if err == nil {
											return objID
										}
									}
								}
								return "" // No executable found
							}
							break
						}
					}
				}
			}
		}
	}
	
	return "" // Path not found
}

// resolveCommandPath resolves /commands/ paths to the actual symlink target
func (s *Storage) resolveCommandPath(path string) string {
	commandPath := strings.TrimPrefix(path, "/commands/")
	if commandPath == "" || commandPath == "/" {
		return "" // Root commands directory - no specific object
	}
	
	// Remove any trailing slash
	commandPath = strings.TrimSuffix(commandPath, "/")
	
	// Check if there's a symlink in the commands directory
	homeDir, _ := os.UserHomeDir()
	symlinkPath := filepath.Join(homeDir, ".port42", "commands", commandPath)
	
	// Follow the symlink to get the actual object path
	if targetPath, err := os.Readlink(symlinkPath); err == nil {
		// Extract object ID from the target path
		// Path format: /Users/.../objects/ab/cd/efgh... -> abcdefgh...
		if strings.Contains(targetPath, "/objects/") {
			parts := strings.Split(targetPath, "/objects/")
			if len(parts) == 2 {
				objectPath := parts[1]
				// Remove directory structure: "ab/cd/efgh..." -> "abcdefgh..."
				objectID := strings.ReplaceAll(objectPath, "/", "")
				return objectID
			}
		}
	}
	
	// Fallback to tools path for backward compatibility
	toolsPath := "/tools/" + commandPath + "/executable"
	return s.resolveToolsPath(toolsPath)
}

// resolveMemoryPath resolves memory session paths to object IDs
func (s *Storage) resolveMemoryPath(path string) string {
	// Extract session ID from path: "/memory/session-123" -> "session-123"
	sessionID := strings.TrimPrefix(path, "/memory/")
	if sessionID == "" || sessionID == "/" {
		return "" // Root memory directory - no specific object
	}
	
	// Remove any trailing slash and sub-paths
	sessionID = strings.TrimSuffix(sessionID, "/")
	if strings.Contains(sessionID, "/") {
		// Handle sub-paths like "/memory/session-123/generated"
		sessionID = strings.Split(sessionID, "/")[0]
	}
	
	// Search for objects with this session ID in their metadata
	ids, err := s.List()
	if err != nil {
		log.Printf("Error listing objects for memory resolution: %v", err)
		return ""
	}
	
	for _, id := range ids {
		meta, err := s.LoadMetadata(id)
		if err != nil {
			continue
		}
		
		// Check if this object is a session with matching ID
		if meta.Type == "session" && meta.Session == sessionID {
			return id
		}
		
		// Also check paths for exact match (for sub-paths like /memory/session/generated/tool)
		for _, p := range meta.Paths {
			if p == path {
				return id
			}
		}
	}
	
	return ""
}

// ==================== Utilities ====================

// List returns all object IDs in storage
func (s *Storage) List() ([]string, error) {
	var ids []string
	
	// Walk through the object store directory structure
	err := filepath.Walk(s.objectsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Reconstruct ID from path
		rel, err := filepath.Rel(s.objectsDir, path)
		if err != nil {
			return err
		}
		
		// Convert path back to ID: 3a/4f/2b8c9d... -> 3a4f2b8c9d...
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) == 3 {
			id := parts[0] + parts[1] + parts[2]
			ids = append(ids, id)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	
	return ids, nil
}

// GetStats returns storage statistics
func (s *Storage) GetStats() StorageStats {
	s.indexMutex.RLock()
	defer s.indexMutex.RUnlock()
	
	return s.stats
}

// ==================== Private Helper Methods ====================

// loadSessionIndex loads the session index from disk
func (s *Storage) loadSessionIndex() error {
	indexPath := filepath.Join(s.baseDir, "session-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No index yet, start fresh
			return nil
		}
		return err
	}
	
	var index map[string]SessionReference
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}
	
	s.sessionIndex = index
	return nil
}

// saveSessionIndex saves the session index to disk
func (s *Storage) saveSessionIndex() error {
	indexPath := filepath.Join(s.baseDir, "session-index.json")
	data, err := json.MarshalIndent(s.sessionIndex, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(indexPath, data, 0644)
}

// updateStats updates storage statistics
func (s *Storage) updateStats() {
	active := 0
	completed := 0
	
	for _, ref := range s.sessionIndex {
		switch ref.State {
		case "active", "idle":
			active++
		case "completed", "abandoned":
			completed++
		}
	}
	
	s.stats.TotalSessions = len(s.sessionIndex)
	s.stats.ActiveSessions = active
	s.stats.CompletedSessions = completed
	s.stats.LastUpdated = time.Now()
	
	// Count objects
	if ids, err := s.List(); err == nil {
		s.stats.TotalObjects = len(ids)
	}
}

// Helper functions use the existing ones from memory_store_object.go

// CopyFrom copies a reader to storage
func (s *Storage) CopyFrom(r io.Reader) (string, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}
	
	return s.Store(content)
}

// ==================== Protocol Handlers ====================

// HandleStorePath processes store_path requests
func (s *Storage) HandleStorePath(path string, content []byte, metadata map[string]interface{}) (map[string]interface{}, error) {
	// Parse virtual path
	pathType, subpath := parseVirtualPath(path)
	if pathType == "invalid" {
		return nil, fmt.Errorf("invalid virtual path: %s", path)
	}
	
	// Create metadata
	meta := &Metadata{
		Type:      inferTypeFromPath(pathType, subpath),
		Created:   time.Now(),
		Modified:  time.Now(),
		Accessed:  time.Now(),
		Lifecycle: "active",
		Paths:     []string{path},
	}
	
	// Add metadata from payload
	if metadata != nil {
		if memoryID, ok := metadata["memory_id"].(string); ok {
			meta.Session = memoryID
		}
		if agent, ok := metadata["agent"].(string); ok {
			meta.Agent = agent
		}
		if crystType, ok := metadata["crystallization_type"].(string); ok {
			meta.Subtype = crystType
		}
		if desc, ok := metadata["description"].(string); ok {
			meta.Description = desc
		}
		if title, ok := metadata["title"].(string); ok {
			meta.Title = title
		}
	}
	
	// Generate additional virtual paths based on type
	meta.Paths = generateVirtualPaths(pathType, subpath, meta)
	
	// Store in object store
	objID, err := s.StoreWithMetadata(content, meta)
	if err != nil {
		return nil, fmt.Errorf("failed to store content: %v", err)
	}
	
	// Special handling for commands - create symlink
	if pathType == "commands" {
		if err := s.CreateCommandSymlink(objID, subpath); err != nil {
			log.Printf("Warning: Failed to create symlink for command %s: %v", subpath, err)
		}
	}
	
	return map[string]interface{}{
		"id":    objID,
		"paths": meta.Paths,
		"size":  len(content),
	}, nil
}

// HandleUpdatePath processes update_path requests
func (s *Storage) HandleUpdatePath(path string, content []byte, metadataUpdates map[string]interface{}) (map[string]interface{}, error) {
	// Resolve path to object ID
	objID := s.ResolvePath(path)
	if objID == "" {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	
	// Load existing metadata
	meta, err := s.LoadMetadata(objID)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %v", err)
	}
	
	// Update content if provided
	if len(content) > 0 {
		// Store new version
		newID, err := s.Store(content)
		if err != nil {
			return nil, fmt.Errorf("failed to store new content: %v", err)
		}
		
		// Update metadata to point to new object
		meta.ID = newID
		meta.Modified = time.Now()
		
		// Update symlinks if it's a command
		if strings.HasPrefix(path, "/commands/") {
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				cmdName := parts[2]
				s.updateCommandSymlink(newID, cmdName)
			}
		}
	}
	
	// Update metadata fields
	if metadataUpdates != nil {
		if lifecycle, ok := metadataUpdates["lifecycle"].(string); ok {
			meta.Lifecycle = lifecycle
		}
		if tags, ok := metadataUpdates["tags"].([]interface{}); ok {
			meta.Tags = make([]string, len(tags))
			for i, tag := range tags {
				meta.Tags[i] = fmt.Sprintf("%v", tag)
			}
		}
		if importance, ok := metadataUpdates["importance"].(string); ok {
			meta.Importance = importance
		}
		if summary, ok := metadataUpdates["summary"].(string); ok {
			meta.Summary = summary
		}
	}
	
	// Save updated metadata
	if err := s.SaveMetadata(meta); err != nil {
		return nil, fmt.Errorf("failed to save metadata: %v", err)
	}
	
	return map[string]interface{}{
		"id":       meta.ID,
		"modified": meta.Modified,
		"paths":    meta.Paths,
	}, nil
}

// HandleDeletePath processes delete_path requests
func (s *Storage) HandleDeletePath(path string) (map[string]interface{}, error) {
	// Resolve path to object ID
	objID := s.ResolvePath(path)
	if objID == "" {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	
	// Load metadata
	meta, err := s.LoadMetadata(objID)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %v", err)
	}
	
	// Remove the specific path from metadata
	newPaths := []string{}
	for _, p := range meta.Paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	meta.Paths = newPaths
	
	// If no paths remain, mark as deprecated
	if len(meta.Paths) == 0 {
		meta.Lifecycle = "deprecated"
	}
	
	// Save updated metadata
	if err := s.SaveMetadata(meta); err != nil {
		return nil, fmt.Errorf("failed to update metadata: %v", err)
	}
	
	// Remove symlink if it's a command
	if strings.HasPrefix(path, "/commands/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			cmdName := parts[2]
			s.removeCommandSymlink(cmdName)
		}
	}
	
	return map[string]interface{}{
		"message":         "Path removed",
		"remaining_paths": meta.Paths,
		"object_id":       objID,
	}, nil
}

// HandleCreateMemory processes create_memory requests
func (s *Storage) HandleCreateMemory(agent, initialMessage string) (map[string]interface{}, error) {
	// Generate memory ID
	memoryID := generateMemoryID()
	
	// Return memory creation info (actual session creation happens elsewhere)
	return map[string]interface{}{
		"memory_id": memoryID,
		"agent":     agent,
		"paths": []string{
			fmt.Sprintf("/memory/%s", memoryID),
			fmt.Sprintf("/by-agent/%s/memory/%s", agent, memoryID),
			fmt.Sprintf("/by-date/%s/memory/%s", time.Now().Format("2006-01-02"), memoryID),
		},
	}, nil
}

// ListPathWithActiveSessions lists path entries including active sessions
func (s *Storage) ListPathWithActiveSessions(path string, activeSessions map[string]*Session) []map[string]interface{} {
	entries := s.ListPath(path)
	
	// Special handling for memory directory - add active sessions
	if path == "/memory" {
		pathMap := make(map[string]bool)
		for _, entry := range entries {
			if name, ok := entry["name"].(string); ok {
				pathMap[name] = true
			}
		}
		
		// Add active sessions not already in the list
		for _, session := range activeSessions {
			name := session.ID
			if !pathMap[name] {
				entries = append(entries, map[string]interface{}{
					"name":     name,
					"type":     "directory",
					"state":    string(session.State),
					"agent":    session.Agent,
					"messages": len(session.Messages),
				})
			}
		}
	}
	
	return entries
}

// ==================== Helper Functions ====================

func parseVirtualPath(path string) (string, string) {
	if !strings.HasPrefix(path, "/") {
		return "invalid", ""
	}
	
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 {
		return "invalid", ""
	}
	
	pathType := parts[0]
	subpath := strings.Join(parts[1:], "/")
	
	switch pathType {
	case "commands", "memory", "artifacts", "by-date", "by-agent", "by-type":
		return pathType, subpath
	default:
		return "invalid", ""
	}
}

func inferTypeFromPath(pathType, subpath string) string {
	switch pathType {
	case "commands":
		return "command"
	case "artifacts":
		// Try to infer from subpath
		parts := strings.Split(subpath, "/")
		if len(parts) > 0 {
			switch parts[0] {
			case "documents":
				return "document"
			case "code":
				return "code"
			case "designs":
				return "design"
			case "media":
				return "media"
			}
		}
		return "artifact"
	case "memory":
		return "memory"
	default:
		return "file"
	}
}

func generateVirtualPaths(pathType, subpath string, meta *Metadata) []string {
	paths := []string{fmt.Sprintf("/%s/%s", pathType, subpath)}
	now := time.Now().Format("2006-01-02")
	
	// Add temporal path
	paths = append(paths, fmt.Sprintf("/by-date/%s/%s", now, filepath.Base(subpath)))
	
	// Add type-based path
	if meta.Type != "" {
		paths = append(paths, fmt.Sprintf("/by-type/%s/%s", meta.Type, filepath.Base(subpath)))
	}
	
	// Add agent-based path
	if meta.Agent != "" {
		paths = append(paths, fmt.Sprintf("/by-agent/%s/%s/%s", meta.Agent, pathType, filepath.Base(subpath)))
	}
	
	// Add memory-based path
	if meta.Session != "" && pathType != "memory" {
		paths = append(paths, fmt.Sprintf("/memory/%s/generated/%s", meta.Session, filepath.Base(subpath)))
	}
	
	return paths
}

func (s *Storage) updateCommandSymlink(objID, cmdName string) error {
	return s.CreateCommandSymlink(objID, cmdName)
}

func (s *Storage) removeCommandSymlink(cmdName string) error {
	homeDir, _ := os.UserHomeDir()
	linkPath := filepath.Join(homeDir, ".port42", "commands", cmdName)
	return os.Remove(linkPath)
}

func generateMemoryID() string {
	return fmt.Sprintf("mem-%d", time.Now().Unix())
}

// SearchObjects searches across all objects and relations in the virtual filesystem
func (s *Storage) SearchObjects(query string, mode string, filters SearchFilters) ([]SearchResult, error) {
	results := []SearchResult{}
	
	// Default limit
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	
	// Phase D: Search relations first (tools, artifacts defined as relations)
	if s.relationStore != nil {
		relationResults, err := s.searchInRelations(query, mode, filters)
		if err == nil {
			results = append(results, relationResults...)
		}
	}
	
	// Load all metadata files (traditional objects)
	entries, err := os.ReadDir(s.metadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %v", err)
	}
	
	// Convert query to lowercase for case-insensitive search
	queryLower := strings.ToLower(query)
	
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		
		// Load metadata
		objID := strings.TrimSuffix(entry.Name(), ".json")
		metadata, err := s.LoadMetadata(objID)
		if err != nil {
			log.Printf("Failed to load metadata for %s: %v", objID, err)
			continue
		}
		
		// Apply filters
		if !matchesFilters(metadata, filters) {
			continue
		}
		
		// Search in metadata fields with mode
		score, matchFields, snippet := searchInMetadata(metadata, queryLower, mode)
		
		// If no metadata match and query exists, optionally search in content
		if score == 0 && query != "" && metadata.Size < 100*1024 { // Only for small files
			contentScore, contentSnippet := s.searchInContent(objID, queryLower, mode, metadata.Type)
			if contentScore > 0 {
				score = contentScore * 0.8 // Content matches score lower than metadata
				matchFields = append(matchFields, "content")
				snippet = contentSnippet
			}
		}
		
		// Skip if no match
		if score == 0 && query != "" {
			continue
		}
		
		// Pick the best path for display
		displayPath := ""
		if len(metadata.Paths) > 0 {
			// Prefer shorter, more intuitive paths
			displayPath = metadata.Paths[0]
			for _, path := range metadata.Paths {
				if len(path) < len(displayPath) && !strings.Contains(path, "by-date") {
					displayPath = path
				}
			}
		}
		
		result := SearchResult{
			Path:        displayPath,
			ObjectID:    objID,
			Type:        metadata.Type,
			Score:       score,
			Snippet:     snippet,
			Metadata:    *metadata,
			MatchFields: matchFields,
		}
		
		results = append(results, result)
		
		// Stop if we have enough results
		if len(results) >= limit {
			break
		}
	}
	
	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		// Primary sort by score
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		// Secondary sort by creation date (newest first)
		return results[i].Metadata.Created.After(results[j].Metadata.Created)
	})
	
	// Trim to limit
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results, nil
}

// searchInRelations searches within the relation store for Phase D advanced discovery
func (s *Storage) searchInRelations(query string, mode string, filters SearchFilters) ([]SearchResult, error) {
	results := []SearchResult{}
	queryLower := strings.ToLower(query)
	
	// Load all relations
	relations, err := s.relationStore.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load relations: %v", err)
	}
	
	for _, relation := range relations {
		// Skip if type filter doesn't match
		if filters.Type != "" && !strings.EqualFold(relation.Type, filters.Type) {
			continue
		}
		
		// Calculate search score and find matches with mode
		score, matchFields, snippet := s.scoreRelation(relation, queryLower, mode)
		
		// Skip if no match and query is specified
		if score == 0 && query != "" {
			continue
		}
		
		// Apply additional filters (agent, date range, etc.)
		if !s.relationMatchesFilters(relation, filters) {
			continue
		}
		
		// Create search result
		displayPath := fmt.Sprintf("/tools/%s", relation.Properties["name"])
		if relation.Type != "Tool" {
			displayPath = fmt.Sprintf("/relations/%s", relation.ID)
		}
		
		// Create metadata-like structure for relations
		relatedMetadata := Metadata{
			ID:          relation.ID,
			Type:        strings.ToLower(relation.Type),
			Created:     relation.CreatedAt,
			Modified:    relation.UpdatedAt,
			Accessed:    relation.UpdatedAt,
			Paths:       []string{displayPath},
			Agent:       getStringProperty(relation.Properties, "agent"),
			Description: getStringProperty(relation.Properties, "description"),
		}
		
		result := SearchResult{
			Path:        displayPath,
			ObjectID:    relation.ID,
			Type:        strings.ToLower(relation.Type),
			Score:       score,
			Snippet:     snippet,
			Metadata:    relatedMetadata,
			MatchFields: matchFields,
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// scoreRelation calculates search relevance score for a relation
func (s *Storage) scoreRelation(relation Relation, queryLower string, mode string) (float64, []string, string) {
	var score float64
	var matchFields []string
	var snippet string
	
	// If no query, return base score
	if queryLower == "" {
		return 1.0, []string{}, ""
	}
	
	// Split query into terms for OR/AND modes
	terms := strings.Fields(queryLower)
	if len(terms) == 0 {
		return 1.0, []string{}, ""
	}
	
	// Search in name (highest weight)
	if name, ok := relation.Properties["name"].(string); ok {
		nameLower := strings.ToLower(name)
		
		switch mode {
		case "phrase", "exact":
			if strings.Contains(nameLower, queryLower) {
				score += 10.0
				matchFields = append(matchFields, "name")
				snippet = extractSnippet(name, queryLower)
			}
		case "and":
			allMatch := true
			for _, term := range terms {
				if !strings.Contains(nameLower, term) {
					allMatch = false
					break
				}
			}
			if allMatch {
				score += 10.0
				matchFields = append(matchFields, "name")
				snippet = extractSnippet(name, terms[0])
			}
		default: // "or"
			matchCount := 0
			for _, term := range terms {
				if strings.Contains(nameLower, term) {
					matchCount++
				}
			}
			if matchCount > 0 {
				score += 10.0 * float64(matchCount) / float64(len(terms))
				matchFields = append(matchFields, "name")
				for _, term := range terms {
					if strings.Contains(nameLower, term) {
						snippet = extractSnippet(name, term)
						break
					}
				}
			}
		}
	}
	
	// Search in transforms (high weight for semantic similarity)
	if transforms, ok := relation.Properties["transforms"].([]interface{}); ok {
		transformStrs := []string{}
		for _, transform := range transforms {
			if transformStr, ok := transform.(string); ok {
				transformStrs = append(transformStrs, transformStr)
			}
		}
		transformText := strings.ToLower(strings.Join(transformStrs, " "))
		
		switch mode {
		case "phrase", "exact":
			if strings.Contains(transformText, queryLower) {
				score += 8.0
				matchFields = append(matchFields, "transforms")
				if snippet == "" {
					snippet = fmt.Sprintf("transforms: %v", transforms)
				}
			}
		case "and":
			allMatch := true
			for _, term := range terms {
				if !strings.Contains(transformText, term) {
					allMatch = false
					break
				}
			}
			if allMatch {
				score += 8.0
				matchFields = append(matchFields, "transforms")
				if snippet == "" {
					snippet = fmt.Sprintf("transforms: %v", transforms)
				}
			}
		default: // "or"
			matchCount := 0
			for _, term := range terms {
				if strings.Contains(transformText, term) {
					matchCount++
				}
			}
			if matchCount > 0 {
				score += 8.0 * float64(matchCount) / float64(len(terms))
				matchFields = append(matchFields, "transforms")
				if snippet == "" {
					snippet = fmt.Sprintf("transforms: %v", transforms)
				}
			}
		}
	}
	
	// Search in description (medium weight)
	if desc, ok := relation.Properties["description"].(string); ok {
		descLower := strings.ToLower(desc)
		
		switch mode {
		case "phrase", "exact":
			if strings.Contains(descLower, queryLower) {
				score += 5.0
				matchFields = append(matchFields, "description")
				if snippet == "" {
					snippet = extractSnippet(desc, queryLower)
				}
			}
		case "and":
			allMatch := true
			for _, term := range terms {
				if !strings.Contains(descLower, term) {
					allMatch = false
					break
				}
			}
			if allMatch {
				score += 5.0
				matchFields = append(matchFields, "description")
				if snippet == "" {
					snippet = extractSnippet(desc, terms[0])
				}
			}
		default: // "or"
			matchCount := 0
			firstMatchTerm := ""
			for _, term := range terms {
				if strings.Contains(descLower, term) {
					matchCount++
					if firstMatchTerm == "" {
						firstMatchTerm = term
					}
				}
			}
			if matchCount > 0 {
				score += 5.0 * float64(matchCount) / float64(len(terms))
				matchFields = append(matchFields, "description")
				if snippet == "" && firstMatchTerm != "" {
					snippet = extractSnippet(desc, firstMatchTerm)
				}
			}
		}
	}
	
	// Search in parent/spawned_by (medium weight for relationship traversal)
	if parent, ok := relation.Properties["parent"].(string); ok {
		parentLower := strings.ToLower(parent)
		
		switch mode {
		case "phrase", "exact":
			if strings.Contains(parentLower, queryLower) {
				score += 6.0
				matchFields = append(matchFields, "parent")
				if snippet == "" {
					snippet = fmt.Sprintf("spawned by: %s", parent)
				}
			}
		case "and":
			allMatch := true
			for _, term := range terms {
				if !strings.Contains(parentLower, term) {
					allMatch = false
					break
				}
			}
			if allMatch {
				score += 6.0
				matchFields = append(matchFields, "parent")
				if snippet == "" {
					snippet = fmt.Sprintf("spawned by: %s", parent)
				}
			}
		default: // "or"
			matchCount := 0
			for _, term := range terms {
				if strings.Contains(parentLower, term) {
					matchCount++
				}
			}
			if matchCount > 0 {
				score += 6.0 * float64(matchCount) / float64(len(terms))
				matchFields = append(matchFields, "parent")
				if snippet == "" {
					snippet = fmt.Sprintf("spawned by: %s", parent)
				}
			}
		}
	}
	
	// Search in all other properties (low weight)
	for key, value := range relation.Properties {
		if key == "name" || key == "description" || key == "transforms" || key == "parent" {
			continue // Already searched
		}
		if valueStr, ok := value.(string); ok {
			valueLower := strings.ToLower(valueStr)
			
			switch mode {
			case "phrase", "exact":
				if strings.Contains(valueLower, queryLower) {
					score += 2.0
					matchFields = append(matchFields, key)
					if snippet == "" {
						snippet = extractSnippet(valueStr, queryLower)
					}
				}
			case "and":
				allMatch := true
				for _, term := range terms {
					if !strings.Contains(valueLower, term) {
						allMatch = false
						break
					}
				}
				if allMatch {
					score += 2.0
					matchFields = append(matchFields, key)
					if snippet == "" {
						snippet = extractSnippet(valueStr, terms[0])
					}
				}
			default: // "or"
				matchCount := 0
				firstMatchTerm := ""
				for _, term := range terms {
					if strings.Contains(valueLower, term) {
						matchCount++
						if firstMatchTerm == "" {
							firstMatchTerm = term
						}
					}
				}
				if matchCount > 0 {
					score += 2.0 * float64(matchCount) / float64(len(terms))
					matchFields = append(matchFields, key)
					if snippet == "" && firstMatchTerm != "" {
						snippet = extractSnippet(valueStr, firstMatchTerm)
					}
				}
			}
		}
	}
	
	return score, matchFields, snippet
}

// relationMatchesFilters checks if relation matches search filters
func (s *Storage) relationMatchesFilters(relation Relation, filters SearchFilters) bool {
	// Agent filter
	if filters.Agent != "" {
		if agent, ok := relation.Properties["agent"].(string); ok {
			if !strings.EqualFold(agent, filters.Agent) {
				return false
			}
		}
	}
	
	// Date filters
	if !filters.After.IsZero() && relation.CreatedAt.Before(filters.After) {
		return false
	}
	if !filters.Before.IsZero() && relation.CreatedAt.After(filters.Before) {
		return false
	}
	
	return true
}

// Helper function to safely get string property
func getStringProperty(properties map[string]interface{}, key string) string {
	if value, ok := properties[key].(string); ok {
		return value
	}
	return ""
}

// matchesFilters checks if metadata matches all provided filters
func matchesFilters(metadata *Metadata, filters SearchFilters) bool {
	// Path filter
	if filters.Path != "" {
		hasMatchingPath := false
		for _, path := range metadata.Paths {
			if strings.HasPrefix(path, filters.Path) {
				hasMatchingPath = true
				break
			}
		}
		if !hasMatchingPath {
			return false
		}
	}
	
	// Type filter
	if filters.Type != "" && metadata.Type != filters.Type {
		return false
	}
	
	// Date filters
	if !filters.After.IsZero() && metadata.Created.Before(filters.After) {
		return false
	}
	if !filters.Before.IsZero() && metadata.Created.After(filters.Before) {
		return false
	}
	
	// Agent filter
	if filters.Agent != "" {
		// Normalize agent names (remove @ prefix for comparison)
		filterAgent := strings.TrimPrefix(filters.Agent, "@")
		metadataAgent := strings.TrimPrefix(metadata.Agent, "@")
		if !strings.EqualFold(filterAgent, metadataAgent) {
			return false
		}
	}
	
	// Tag filters (must have all specified tags)
	if len(filters.Tags) > 0 {
		for _, requiredTag := range filters.Tags {
			hasTag := false
			for _, tag := range metadata.Tags {
				if strings.EqualFold(tag, requiredTag) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				return false
			}
		}
	}
	
	return true
}

// searchInMetadata searches for query in metadata fields and returns score
func searchInMetadata(metadata *Metadata, queryLower string, mode string) (float64, []string, string) {
	score := 0.0
	matchFields := []string{}
	snippet := ""
	
	// Empty query matches everything with base score
	if queryLower == "" {
		return 1.0, []string{"all"}, metadata.Description
	}
	
	// Handle different search modes
	switch mode {
	case "phrase", "exact":
		// Original behavior - exact phrase match
		if strings.Contains(strings.ToLower(metadata.Description), queryLower) {
			score += 3.0
			matchFields = append(matchFields, "description")
			snippet = extractSnippet(metadata.Description, queryLower)
		}
		
		if strings.Contains(strings.ToLower(metadata.Title), queryLower) {
			score += 2.5
			matchFields = append(matchFields, "title")
			if snippet == "" {
				snippet = metadata.Title
			}
		}
		
		for _, tag := range metadata.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				score += 2.0
				matchFields = append(matchFields, "tags")
				if snippet == "" {
					snippet = fmt.Sprintf("Tag: %s", tag)
				}
				break
			}
		}
		
	case "and":
		// All terms must match
		terms := strings.Fields(queryLower)
		if len(terms) == 0 {
			return 1.0, []string{"all"}, metadata.Description
		}
		
		// Check description
		allMatchInDesc := true
		for _, term := range terms {
			if !strings.Contains(strings.ToLower(metadata.Description), term) {
				allMatchInDesc = false
				break
			}
		}
		if allMatchInDesc {
			score += 3.0
			matchFields = append(matchFields, "description")
			snippet = extractSnippet(metadata.Description, terms[0])
		}
		
		// Check title
		allMatchInTitle := true
		for _, term := range terms {
			if !strings.Contains(strings.ToLower(metadata.Title), term) {
				allMatchInTitle = false
				break
			}
		}
		if allMatchInTitle {
			score += 2.5
			matchFields = append(matchFields, "title")
			if snippet == "" {
				snippet = metadata.Title
			}
		}
		
		// Check tags (all terms must match across all tags)
		tagText := strings.ToLower(strings.Join(metadata.Tags, " "))
		allMatchInTags := true
		for _, term := range terms {
			if !strings.Contains(tagText, term) {
				allMatchInTags = false
				break
			}
		}
		if allMatchInTags {
			score += 2.0
			matchFields = append(matchFields, "tags")
			if snippet == "" {
				snippet = fmt.Sprintf("Tags: %s", strings.Join(metadata.Tags, ", "))
			}
		}
		
	case "or":
		fallthrough
	default:
		// Any term matches (OR mode)
		terms := strings.Fields(queryLower)
		if len(terms) == 0 {
			return 1.0, []string{"all"}, metadata.Description
		}
		
		// Count matches in description
		descMatches := 0
		for _, term := range terms {
			if strings.Contains(strings.ToLower(metadata.Description), term) {
				descMatches++
			}
		}
		if descMatches > 0 {
			// Score based on percentage of terms matched
			score += 3.0 * float64(descMatches) / float64(len(terms))
			matchFields = append(matchFields, "description")
			// Find first matching term for snippet
			for _, term := range terms {
				if strings.Contains(strings.ToLower(metadata.Description), term) {
					snippet = extractSnippet(metadata.Description, term)
					break
				}
			}
		}
		
		// Count matches in title
		titleMatches := 0
		for _, term := range terms {
			if strings.Contains(strings.ToLower(metadata.Title), term) {
				titleMatches++
			}
		}
		if titleMatches > 0 {
			score += 2.5 * float64(titleMatches) / float64(len(terms))
			matchFields = append(matchFields, "title")
			if snippet == "" {
				snippet = metadata.Title
			}
		}
		
		// Count matches in tags
		tagText := strings.ToLower(strings.Join(metadata.Tags, " "))
		tagMatches := 0
		for _, term := range terms {
			if strings.Contains(tagText, term) {
				tagMatches++
			}
		}
		if tagMatches > 0 {
			score += 2.0 * float64(tagMatches) / float64(len(terms))
			matchFields = append(matchFields, "tags")
			if snippet == "" {
				snippet = fmt.Sprintf("Tags: %s", strings.Join(metadata.Tags, ", "))
			}
		}
	}
	
	// For phrase mode, search session ID, agent, paths with full query
	if mode == "phrase" || mode == "exact" {
		if strings.Contains(strings.ToLower(metadata.Session), queryLower) {
			score += 1.5
			matchFields = append(matchFields, "session")
		}
		
		if strings.Contains(strings.ToLower(metadata.Agent), queryLower) {
			score += 1.5
			matchFields = append(matchFields, "agent")
		}
		
		for _, path := range metadata.Paths {
			if strings.Contains(strings.ToLower(path), queryLower) {
				score += 0.5
				matchFields = append(matchFields, "path")
				break
			}
		}
	} else {
		// For AND/OR modes, check each term
		terms := strings.Fields(queryLower)
		
		// Session ID
		sessionMatches := 0
		for _, term := range terms {
			if strings.Contains(strings.ToLower(metadata.Session), term) {
				sessionMatches++
			}
		}
		if mode == "and" && sessionMatches == len(terms) && sessionMatches > 0 {
			score += 1.5
			matchFields = append(matchFields, "session")
		} else if mode == "or" && sessionMatches > 0 {
			score += 1.5 * float64(sessionMatches) / float64(len(terms))
			matchFields = append(matchFields, "session")
		}
		
		// Agent
		agentMatches := 0
		for _, term := range terms {
			if strings.Contains(strings.ToLower(metadata.Agent), term) {
				agentMatches++
			}
		}
		if mode == "and" && agentMatches == len(terms) && agentMatches > 0 {
			score += 1.5
			matchFields = append(matchFields, "agent")
		} else if mode == "or" && agentMatches > 0 {
			score += 1.5 * float64(agentMatches) / float64(len(terms))
			matchFields = append(matchFields, "agent")
		}
		
		// Paths
		pathText := strings.ToLower(strings.Join(metadata.Paths, " "))
		pathMatches := 0
		for _, term := range terms {
			if strings.Contains(pathText, term) {
				pathMatches++
			}
		}
		if mode == "and" && pathMatches == len(terms) && pathMatches > 0 {
			score += 0.5
			matchFields = append(matchFields, "path")
		} else if mode == "or" && pathMatches > 0 {
			score += 0.5 * float64(pathMatches) / float64(len(terms))
			matchFields = append(matchFields, "path")
		}
	}
	
	// Boost recent items slightly
	age := time.Since(metadata.Created)
	if age < 24*time.Hour {
		score *= 1.2
	} else if age < 7*24*time.Hour {
		score *= 1.1
	}
	
	return score, matchFields, snippet
}

// searchInContent searches in the actual content of an object
func (s *Storage) searchInContent(objID, queryLower, mode, objType string) (float64, string) {
	content, err := s.Read(objID)
	if err != nil {
		return 0, ""
	}
	
	contentStr := string(content)
	contentLower := strings.ToLower(contentStr)
	
	score := 0.0
	snippet := ""
	
	switch mode {
	case "phrase", "exact":
		// Original behavior - exact phrase match
		if !strings.Contains(contentLower, queryLower) {
			return 0, ""
		}
		
		// Base score for content match
		score = 1.0
		
		// Count occurrences (max 5 for scoring)
		count := strings.Count(contentLower, queryLower)
		if count > 5 {
			count = 5
		}
		score += float64(count) * 0.2
		snippet = extractSnippet(contentStr, queryLower)
		
	case "and":
		// All terms must match
		terms := strings.Fields(queryLower)
		if len(terms) == 0 {
			return 0, ""
		}
		
		for _, term := range terms {
			if !strings.Contains(contentLower, term) {
				return 0, ""  // Missing required term
			}
		}
		
		// All terms found
		score = 1.0
		// Count total occurrences
		totalCount := 0
		for _, term := range terms {
			count := strings.Count(contentLower, term)
			if count > 5 {
				count = 5
			}
			totalCount += count
		}
		score += float64(totalCount) * 0.1
		snippet = extractSnippet(contentStr, terms[0])
		
	case "or":
		fallthrough
	default:
		// Any term matches
		terms := strings.Fields(queryLower)
		if len(terms) == 0 {
			return 0, ""
		}
		
		matchCount := 0
		totalCount := 0
		firstMatch := ""
		
		for _, term := range terms {
			if strings.Contains(contentLower, term) {
				matchCount++
				if firstMatch == "" {
					firstMatch = term
				}
				count := strings.Count(contentLower, term)
				if count > 5 {
					count = 5
				}
				totalCount += count
			}
		}
		
		if matchCount == 0 {
			return 0, ""
		}
		
		// Score based on percentage of terms matched
		score = 1.0 * float64(matchCount) / float64(len(terms))
		score += float64(totalCount) * 0.1
		snippet = extractSnippet(contentStr, firstMatch)
	}
	
	return score, snippet
}

// extractSnippet extracts a snippet around the query match
func extractSnippet(text, query string) string {
	textLower := strings.ToLower(text)
	idx := strings.Index(textLower, query)
	if idx == -1 {
		return ""
	}
	
	// Extract ~80 chars around the match
	start := idx - 40
	if start < 0 {
		start = 0
	}
	
	end := idx + len(query) + 40
	if end > len(text) {
		end = len(text)
	}
	
	snippet := text[start:end]
	
	// Add ellipsis if truncated
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}
	
	return strings.TrimSpace(snippet)
}

// Unified Tools Hierarchy Helper Methods

// handleToolsPath handles the unified tools virtual filesystem paths
func (s *Storage) handleToolsPath(path string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	// Remove leading /tools
	toolsPath := strings.TrimPrefix(path, "/tools")
	if toolsPath == "" {
		toolsPath = "/"
	}
	
	switch {
	case toolsPath == "/":
		// /tools/ - show unified tool structure
		entries = append(entries, map[string]interface{}{
			"name": "by-name",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "by-transform",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "spawned-by",
			"type": "directory",
		})
		entries = append(entries, map[string]interface{}{
			"name": "ancestry",
			"type": "directory",
		})
		
		// Also show individual tools as directories
		if relations, err := s.relationStore.List(); err == nil {
			for _, relation := range relations {
				if relation.Type == "Tool" {
					if name, ok := relation.Properties["name"].(string); ok {
						entries = append(entries, map[string]interface{}{
							"name":         name,
							"type":         "directory",
							"relation_id":  relation.ID,
							"created":      relation.CreatedAt,
							"modified":     relation.UpdatedAt,
						})
					}
				}
			}
		}
		
	case toolsPath == "/by-name" || toolsPath == "/by-name/":
		// /tools/by-name/ - all tools alphabetically
		return s.handleToolsByName()
		
	case toolsPath == "/by-transform" || toolsPath == "/by-transform/":
		// /tools/by-transform/ - grouped by transforms
		return s.handleToolsByTransform("")
		
	case strings.HasPrefix(toolsPath, "/by-transform/"):
		// /tools/by-transform/{transform}/ - tools with specific transform
		transform := strings.TrimPrefix(toolsPath, "/by-transform/")
		transform = strings.TrimSuffix(transform, "/")
		return s.handleToolsByTransform(transform)
		
	case toolsPath == "/spawned-by" || toolsPath == "/spawned-by/":
		// /tools/spawned-by/ - global spawned-by index
		return s.handleSpawnedByIndex()
		
	case strings.HasPrefix(toolsPath, "/spawned-by/"):
		// /tools/spawned-by/{tool}/ - what this tool spawned
		toolName := strings.TrimPrefix(toolsPath, "/spawned-by/")
		toolName = strings.TrimSuffix(toolName, "/")
		return s.handleSpawnedByTool(toolName)
		
	case toolsPath == "/ancestry" || toolsPath == "/ancestry/":
		// /tools/ancestry/ - tools with parent chains
		return s.handleAncestryIndex()
		
	default:
		// Check if it's an individual tool directory
		parts := strings.Split(strings.Trim(toolsPath, "/"), "/")
		if len(parts) >= 1 {
			toolName := parts[0]
			if len(parts) == 1 {
				// /tools/{tool}/ - show tool subpaths
				return s.handleIndividualTool(toolName)
			} else {
				// /tools/{tool}/{subpath}
				subpath := strings.Join(parts[1:], "/")
				return s.handleToolSubpath(toolName, subpath)
			}
		}
	}
	
	return entries
}
// handleToolsByName shows all tools alphabetically
func (s *Storage) handleToolsByName() []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	if relations, err := s.relationStore.List(); err == nil {
		for _, relation := range relations {
			if relation.Type == "Tool" {
				if name, ok := relation.Properties["name"].(string); ok {
					entries = append(entries, map[string]interface{}{
						"name":         name,
						"type":         "directory",
						"relation_id":  relation.ID,
						"created":      relation.CreatedAt,
						"modified":     relation.UpdatedAt,
					})
				}
			}
		}
	}
	
	return entries
}

// handleToolsByTransform shows tools grouped by transforms or specific transform
func (s *Storage) handleToolsByTransform(specificTransform string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	if specificTransform == "" {
		// Show available transforms
		transformSet := make(map[string]bool)
		if relations, err := s.relationStore.List(); err == nil {
			for _, relation := range relations {
				if relation.Type == "Tool" {
					if transformsRaw, exists := relation.Properties["transforms"]; exists {
						if transformsList, ok := transformsRaw.([]interface{}); ok {
							for _, t := range transformsList {
								if tStr, ok := t.(string); ok {
									transformSet[tStr] = true
								}
							}
						}
					}
				}
			}
		}
		
		// Convert set to entries
		for transform := range transformSet {
			entries = append(entries, map[string]interface{}{
				"name": transform,
				"type": "directory",
			})
		}
	} else {
		// Show tools with specific transform
		if relations, err := s.relationStore.List(); err == nil {
			for _, relation := range relations {
				if relation.Type == "Tool" {
					if name, ok := relation.Properties["name"].(string); ok {
						if hasTransformInRelation(relation, specificTransform) {
							entries = append(entries, map[string]interface{}{
								"name":         name,
								"type":         "directory",
								"relation_id":  relation.ID,
								"created":      relation.CreatedAt,
								"modified":     relation.UpdatedAt,
							})
						}
					}
				}
			}
		}
	}
	
	return entries
}

// handleSpawnedByIndex shows tools that have spawned other entities
func (s *Storage) handleSpawnedByIndex() []map[string]interface{} {
	entries := []map[string]interface{}{}
	spawningTools := make(map[string]bool)
	
	if relations, err := s.relationStore.List(); err == nil {
		// Find tools that spawned others
		for _, relation := range relations {
			if spawnedBy, exists := relation.Properties["spawned_by"]; exists {
				if spawnedByID, ok := spawnedBy.(string); ok {
					// Find the tool that did the spawning
					if parent, err := s.relationStore.Load(spawnedByID); err == nil {
						if parentName, ok := parent.Properties["name"].(string); ok {
							spawningTools[parentName] = true
						}
					}
				}
			}
		}
	}
	
	// Convert to entries
	for toolName := range spawningTools {
		entries = append(entries, map[string]interface{}{
			"name": toolName,
			"type": "directory",
		})
	}
	
	return entries
}

// handleSpawnedByTool shows what a specific tool spawned
func (s *Storage) handleSpawnedByTool(toolName string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	// Find the tool relation first
	var parentRelation *Relation
	if relations, err := s.relationStore.List(); err == nil {
		for _, relation := range relations {
			if name, ok := relation.Properties["name"].(string); ok && name == toolName {
				parentRelation = &relation
				break
			}
		}
	}
	
	if parentRelation == nil {
		return entries
	}
	
	// Find entities spawned by this tool
	if relations, err := s.relationStore.List(); err == nil {
		for _, relation := range relations {
			if spawnedBy, exists := relation.Properties["spawned_by"]; exists {
				if spawnedByID, ok := spawnedBy.(string); ok && spawnedByID == parentRelation.ID {
					if name, ok := relation.Properties["name"].(string); ok {
						entries = append(entries, map[string]interface{}{
							"name":         name,
							"type":         "directory",
							"relation_id":  relation.ID,
							"created":      relation.CreatedAt,
							"modified":     relation.UpdatedAt,
						})
					}
				}
			}
		}
	}
	
	return entries
}

// handleAncestryIndex shows tools with parent relationships
func (s *Storage) handleAncestryIndex() []map[string]interface{} {
	entries := []map[string]interface{}{}
	toolsWithParents := make(map[string]bool)
	
	if relations, err := s.relationStore.List(); err == nil {
		for _, relation := range relations {
			if _, hasParent := relation.Properties["parent"]; hasParent {
				if name, ok := relation.Properties["name"].(string); ok {
					toolsWithParents[name] = true
				}
			}
		}
	}
	
	for toolName := range toolsWithParents {
		entries = append(entries, map[string]interface{}{
			"name": toolName,
			"type": "directory",
		})
	}
	
	return entries
}

// handleIndividualTool shows subpaths for a specific tool
func (s *Storage) handleIndividualTool(toolName string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	// Standard subpaths for every tool
	entries = append(entries, map[string]interface{}{
		"name": "definition",
		"type": "file",
	})
	entries = append(entries, map[string]interface{}{
		"name": "executable",
		"type": "file",
	})
	entries = append(entries, map[string]interface{}{
		"name": "spawned",
		"type": "directory",
	})
	entries = append(entries, map[string]interface{}{
		"name": "parents",
		"type": "directory",
	})
	
	return entries
}

// handleToolSubpath handles specific tool subpaths
func (s *Storage) handleToolSubpath(toolName string, subpath string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	switch subpath {
	case "spawned", "spawned/":
		// Show what this tool spawned (same as spawned-by logic)
		return s.handleSpawnedByTool(toolName)
		
	case "parents", "parents/":
		// Show parent chain for this tool
		return s.handleParentChain(toolName)
	}
	
	return entries
}

// handleParentChain shows the parent ancestry for a tool
func (s *Storage) handleParentChain(toolName string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	// Find the tool and its parent
	if relations, err := s.relationStore.List(); err == nil {
		for _, relation := range relations {
			if name, ok := relation.Properties["name"].(string); ok && name == toolName {
				if parent, exists := relation.Properties["parent"]; exists {
					if parentName, ok := parent.(string); ok {
						entries = append(entries, map[string]interface{}{
							"name": parentName,
							"type": "directory",
						})
					}
				}
				break
			}
		}
	}
	
	return entries
}

// hasTransformInRelation checks if a relation has a specific transform (helper function)
func hasTransformInRelation(relation Relation, transform string) bool {
	if transformsRaw, exists := relation.Properties["transforms"]; exists {
		if transformsList, ok := transformsRaw.([]interface{}); ok {
			for _, t := range transformsList {
				if tStr, ok := t.(string); ok && tStr == transform {
					return true
				}
			}
		}
	}
	return false
}

// handleEnhancedCommandsView shows relation-backed tools as commands with metadata
func (s *Storage) handleEnhancedCommandsView() []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	if s.relationStore == nil {
		return entries
	}
	
	// Get all tool relations
	relations, err := s.relationStore.List()
	if err != nil {
		log.Printf("Failed to load relations for commands view: %v", err)
		return entries
	}
	
	// Convert tool relations to command entries with metadata
	for _, relation := range relations {
		if relation.Type == "Tool" {
			if name, ok := relation.Properties["name"].(string); ok {
				entry := map[string]interface{}{
					"name":        name,
					"type":        "file",
					"relation_id": relation.ID,
					"created":     relation.CreatedAt,
					"modified":    relation.UpdatedAt,
				}
				
				// Add relation-specific metadata
				if transforms, exists := relation.Properties["transforms"]; exists {
					entry["transforms"] = transforms
				}
				if spawnedBy, exists := relation.Properties["spawned_by"]; exists {
					entry["spawned_by"] = spawnedBy
				}
				if parent, exists := relation.Properties["parent"]; exists {
					entry["parent"] = parent
				}
				if autoSpawned, exists := relation.Properties["auto_spawned"]; exists {
					entry["auto_spawned"] = autoSpawned
				}
				
				entries = append(entries, entry)
			}
		}
	}
	
	return entries
}

// handleEnhancedByDateView includes relations alongside traditional objects by date
func (s *Storage) handleEnhancedByDateView(path string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	// Extract date from path: /by-date/2025-08-10/ -> "2025-08-10"
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 2 {
		// Show available dates - for now return empty, could be enhanced
		return entries
	}
	
	targetDate := pathParts[1]
	
	// Get traditional object entries by date (existing logic)
	ids, err := s.List()
	if err != nil {
		log.Printf("Error listing objects for by-date: %v", err)
		return entries
	}
	
	pathMap := make(map[string]bool)
	
	// Add traditional objects that match the date
	for _, id := range ids {
		meta, err := s.LoadMetadata(id)
		if err != nil {
			continue
		}
		
		// Check if created on target date
		if meta.Created.Format("2006-01-02") == targetDate {
			// Check each virtual path for this object
			for _, vpath := range meta.Paths {
				if strings.HasPrefix(vpath, "/by-date/"+targetDate+"/") {
					// Extract the next component
					relative := strings.TrimPrefix(vpath, "/by-date/"+targetDate+"/")
					parts := strings.Split(relative, "/")
					if len(parts) > 0 {
						name := parts[0]
						if !pathMap[name] {
							pathMap[name] = true
							
							entry := map[string]interface{}{
								"name":         name,
								"type":         "file",
								"id":           id,
								"size":         meta.Size,
								"created":      meta.Created,
								"modified":     meta.Modified,
								"content_type": meta.Type,
								"source":       "object",
							}
							
							entries = append(entries, entry)
						}
					}
				}
			}
		}
	}
	
	// Add relation entries that match the date
	if s.relationStore != nil {
		relations, err := s.relationStore.List()
		if err == nil {
			for _, relation := range relations {
				// Check if created on target date
				if relation.CreatedAt.Format("2006-01-02") == targetDate {
					if name, ok := relation.Properties["name"].(string); ok {
						if !pathMap[name] {
							pathMap[name] = true
							
							entry := map[string]interface{}{
								"name":        name,
								"type":        "file", 
								"relation_id": relation.ID,
								"created":     relation.CreatedAt,
								"modified":    relation.UpdatedAt,
								"source":      "relation",
								"relation_type": relation.Type,
							}
							
							// Add relation metadata
							if transforms, exists := relation.Properties["transforms"]; exists {
								entry["transforms"] = transforms
							}
							
							entries = append(entries, entry)
						}
					}
				}
			}
		}
	}
	
	return entries
}

// handleMemoryGeneratedView shows tools created from a specific memory session  
func (s *Storage) handleMemoryGeneratedView(path string) []map[string]interface{} {
	entries := []map[string]interface{}{}
	
	// Extract session ID from path: /memory/session-123/generated -> "session-123"
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 2 {
		return entries
	}
	
	sessionID := pathParts[1]
	
	if s.relationStore == nil {
		return entries
	}
	
	// Get all relations and filter by memory_session
	relations, err := s.relationStore.List()
	if err != nil {
		log.Printf("Failed to load relations for generated view: %v", err)
		return entries
	}
	
	// Filter relations that match the session ID
	for _, relation := range relations {
		if memorySession, ok := relation.Properties["memory_session"].(string); ok {
			if memorySession == sessionID {
				if name, ok := relation.Properties["name"].(string); ok {
					entry := map[string]interface{}{
						"name":        name,
						"type":        "file",
						"relation_id": relation.ID,
						"created":     relation.CreatedAt,
						"modified":    relation.UpdatedAt,
						"session_id":  sessionID,
					}
					entries = append(entries, entry)
				}
			}
		}
	}
	
	return entries
}

// handleSimilarView provides semantic tool discovery through similarity analysis
func (s *Storage) handleSimilarView(path string) []map[string]interface{} {
	// Extract tool name from path if provided
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	
	if len(pathParts) == 1 && pathParts[0] == "similar" {
		// Root /similar path
		return s.handleSimilarRootView()
	} else if len(pathParts) >= 2 {
		// Specific tool similarity path
		toolName := pathParts[1]
		return s.handleSimilarToolView(toolName)
	}
	
	return []map[string]interface{}{}
}

// handleSimilarRootView shows all tools that have similar tools available
func (s *Storage) handleSimilarRootView() []map[string]interface{} {
	calculator := s.getSimilarityCalculator()
	if calculator == nil {
		return []map[string]interface{}{
			{
				"name":        "⚠️ Similarity calculator unavailable",
				"type":        "error",
				"description": "RelationStore not initialized",
			},
		}
	}
	
	// Load all tool relations
	allRelations, err := s.relationStore.List()
	if err != nil {
		return []map[string]interface{}{
			{
				"name":        "⚠️ Failed to load tools",
				"type":        "error", 
				"description": fmt.Sprintf("Error: %v", err),
			},
		}
	}
	
	entries := []map[string]interface{}{}
	
	// Find tools with similar tools (threshold 0.2 for directory listing)
	for _, relation := range allRelations {
		if relation.Type != "Tool" {
			continue
		}
		
		toolName, ok := relation.Properties["name"].(string)
		if !ok {
			continue
		}
		
		// Find similar tools for this tool
		similarTools, err := calculator.findSimilarTools(relation, 0.2)
		if err != nil {
			continue
		}
		
		// Only include tools that have similar tools
		if len(similarTools) > 0 {
			entry := map[string]interface{}{
				"name":             toolName,
				"type":             "directory",
				"similar_count":    len(similarTools),
				"description":      fmt.Sprintf("Tool with %d similar tools", len(similarTools)),
				"path":            fmt.Sprintf("/similar/%s", toolName),
			}
			entries = append(entries, entry)
		}
	}
	
	if len(entries) == 0 {
		return []map[string]interface{}{
			{
				"name":        "🔍 No tools with similarities found",
				"type":        "notice",
				"description": "Either no tools exist or none have sufficient similarity (>20%)",
			},
		}
	}
	
	return entries
}

// handleSimilarToolView shows tools similar to a specific target tool
func (s *Storage) handleSimilarToolView(toolName string) []map[string]interface{} {
	calculator := s.getSimilarityCalculator()
	if calculator == nil {
		return []map[string]interface{}{
			{
				"name":        "⚠️ Similarity calculator unavailable",
				"type":        "error",
				"description": "RelationStore not initialized",
			},
		}
	}
	
	// Find similar tools using the calculator (threshold 0.2)
	similarTools, err := calculator.GetSimilarToolsForTool(toolName, 0.2)
	if err != nil {
		return []map[string]interface{}{
			{
				"name":        fmt.Sprintf("⚠️ Error finding similar tools for '%s'", toolName),
				"type":        "error",
				"description": fmt.Sprintf("Error: %v", err),
			},
		}
	}
	
	if len(similarTools) == 0 {
		return []map[string]interface{}{
			{
				"name":        fmt.Sprintf("🔍 No similar tools found for '%s'", toolName),
				"type":        "notice",
				"description": "No tools found with similarity above 20% threshold",
			},
		}
	}
	
	// Convert similar tools to virtual nodes
	entries := []map[string]interface{}{}
	for _, similarTool := range similarTools {
		entry := s.similarToolToVirtualNode(similarTool)
		entries = append(entries, entry)
	}
	
	return entries
}

// similarToolToVirtualNode converts a SimilarTool to a virtual node for CLI display
func (s *Storage) similarToolToVirtualNode(similarTool SimilarTool) map[string]interface{} {
	toolName, ok := similarTool.Tool.Properties["name"].(string)
	if !ok {
		toolName = similarTool.Tool.ID
	}
	
	// Build similarity description
	similarityPercent := int(similarTool.Similarity * 100)
	description := fmt.Sprintf("%d%% similarity", similarityPercent)
	
	if len(similarTool.Reason) > 0 {
		description += " - " + strings.Join(similarTool.Reason, ", ")
	}
	
	// Extract transforms for additional context
	transforms := []string{}
	if transformsRaw, exists := similarTool.Tool.Properties["transforms"]; exists {
		if transformsInterface, ok := transformsRaw.([]interface{}); ok {
			for _, t := range transformsInterface {
				if tStr, ok := t.(string); ok {
					transforms = append(transforms, tStr)
				}
			}
		} else if transformsString, ok := transformsRaw.([]string); ok {
			transforms = transformsString
		}
	}
	
	entry := map[string]interface{}{
		"name":          toolName,
		"type":          "tool",
		"similarity":    similarityPercent,
		"score":         similarTool.Similarity,
		"description":   description,
		"path":          fmt.Sprintf("/tools/%s", toolName),
		"transforms":    transforms,
		"reason":        similarTool.Reason,
	}
	
	return entry
}

// getSimilarityCalculator creates a SimilarityCalculator instance for this storage
func (s *Storage) getSimilarityCalculator() *SimilarityCalculator {
	if s.relationStore == nil {
		return nil
	}
	return NewSimilarityCalculator(s.relationStore)
}