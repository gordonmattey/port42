package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	
	"port42/daemon/resolution"
	"port42/daemon/validation"
)

// Daemon represents the Port 42 daemon
type Daemon struct {
	listener        net.Listener
	sessions        map[string]*Session
	mu              sync.RWMutex
	config          Config
	shutdownCh      chan struct{}
	wg              sync.WaitGroup
	storage         *Storage
	baseDir         string
	realityCompiler *RealityCompiler // NEW: Reality compiler component
	resolutionService resolution.ResolutionService // Phase 2: Reference resolution service
	validator       *validation.RequestValidator // Step 5: Request validation
	referenceHandler *ReferenceHandler // Common reference resolution logic
	contextCollector *ContextCollector // Step 2: Context tracking and suggestions
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
	
	// Initialize relation store first
	relationStore, err := NewFileRelationStore(baseDir)
	if err != nil {
		log.Printf("❌ Failed to initialize relation store: %v", err)
		relationStore = nil // Continue without relations
	}
	
	// Initialize unified storage with relation store
	log.Printf("🗄️ Initializing storage...")
	storage, err := NewStorage(baseDir, relationStore)
	if err != nil {
		log.Printf("❌ Failed to initialize storage: %v", err)
		// Continue without storage for now
	} else {
		log.Printf("✅ Storage initialized successfully")
	}
	
	// Debug logging
	log.Printf("DEBUG: NewDaemon called with port = '%s'", port)
	
	daemon := &Daemon{
		listener:   listener,
		sessions:   make(map[string]*Session),
		shutdownCh: make(chan struct{}),
		storage:    storage,
		baseDir:    baseDir,
		config: Config{
			Port:         port,
			AIBackend:    "http://localhost:3000/api/ai", // Default, can be overridden
			MaxSessions:  100,
			SessionTTL:   24 * time.Hour,
			MemoryPath:   filepath.Join(homeDir, ".port42", "memory"),
			CommandsPath: filepath.Join(homeDir, ".port42", "commands"),
		},
	}
	
	// Initialize Context Collector FIRST (before Reality Compiler needs it)
	log.Printf("📊 Initializing Context Collector...")
	daemon.contextCollector = NewContextCollector(daemon)
	log.Printf("✅ Context Collector initialized")
	
	// Initialize Reality Compiler (now has access to context collector)
	log.Printf("🌟 Initializing Reality Compiler...")
	if err := daemon.initializeRealityCompiler(); err != nil {
		log.Printf("⚠️ Failed to initialize Reality Compiler: %v", err)
		log.Printf("💡 Declarative commands will not be available")
	} else {
		log.Printf("✅ Reality Compiler initialized successfully")
	}
	
	// Initialize Reference Resolution Manager (Phase 2)
	log.Printf("📎 Initializing Reference Resolution Manager...")
	if err := daemon.initializeResolutionManager(); err != nil {
		log.Printf("⚠️ Failed to initialize Resolution Manager: %v", err)
		log.Printf("💡 Reference resolution will not be available")
	} else {
		log.Printf("✅ Resolution Manager initialized successfully")
	}
	
	// Initialize Request Validator (Step 5)
	log.Printf("🛡️ Initializing Request Validator...")
	daemon.validator = validation.NewRequestValidator()
	log.Printf("✅ Request Validator initialized successfully")
	
	// Initialize Reference Handler (common reference resolution logic)
	log.Printf("🔗 Initializing Reference Handler...")
	daemon.referenceHandler = NewReferenceHandler(daemon.resolutionService)
	log.Printf("✅ Reference Handler initialized successfully")
	
	log.Printf("DEBUG: Created daemon with config.Port = '%s'", daemon.config.Port)
	return daemon
}

// Start begins accepting connections
func (d *Daemon) Start() {
	log.Printf("🐬 Daemon starting with config: %+v", d.config)
	
	// Load recent sessions from disk
	if d.storage != nil {
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
	
	// Only log non-context requests to reduce noise
	if req.Type != "context" {
		log.Printf("◊ New consciousness connected from %s", clientAddr)
		log.Printf("◊ Request [%s] type: %s", req.ID, req.Type)
	}
	
	// Process request
	resp := d.handleRequest(req)
	
	// Debug: Check response size (skip for context)
	var respJSON []byte
	if req.Type != "context" {
		respJSON, _ = json.Marshal(resp)
		log.Printf("🔍 Response size for [%s]: %d bytes", resp.ID, len(respJSON))
		
		// For very large responses, log a warning
		if len(respJSON) > 1024*1024 { // 1MB
			log.Printf("⚠️ Large response detected: %.2f MB", float64(len(respJSON))/(1024*1024))
		}
	}
	
	// Send response
	if err := encoder.Encode(resp); err != nil {
		log.Printf("Error encoding response to %s: %v", clientAddr, err)
		return
	}
	
	// Only log non-context responses
	if req.Type != "context" {
		log.Printf("◊ Response sent [%s] success: %v", resp.ID, resp.Success)
		log.Printf("◊ Consciousness disconnected: %s", clientAddr)
	}
}

// handleRequest routes requests to appropriate handlers
func (d *Daemon) handleRequest(req Request) Response {
	// Track command execution (except context and ping)
	if d.contextCollector != nil && req.Type != "context" && req.Type != "ping" {
		d.contextCollector.TrackCommand(req.Type, 0)
	}
	
	switch req.Type {
	case RequestStatus:
		return d.handleStatus(req)
	case RequestPossess:
		return d.handlePossess(req)
	case RequestList:
		return d.handleList(req)
	case RequestMemory:
		return d.handleMemory(req)
	case RequestWatch:
		return d.handleWatch(req)
	case RequestEnd:
		return d.handleEnd(req)
	case "ping":
		// Simple ping handler for connection checks
		return NewResponse(req.ID, true)
	case "store_path":
		return d.handleStorePath(req)
	case "update_path":
		return d.handleUpdatePath(req)
	case "delete_path":
		return d.handleDeletePath(req)
	case "create_memory":
		return d.handleCreateMemory(req)
	case "list_path":
		return d.handleListPath(req)
	case "read_path":
		return d.handleReadPath(req)
	case "get_metadata":
		return d.handleGetMetadata(req)
	case "search":
		return d.handleSearch(req)
	case "get_last_session":
		return d.handleGetLastSession(req)
	case "declare_relation":
		return d.handleDeclareRelation(req)
	case "get_relation":
		return d.handleGetRelation(req)
	case "list_relations":
		return d.handleListRelations(req)
	case "delete_relation":
		return d.handleDeleteRelation(req)
	case "context":
		return d.handleGetContext(req)
	default:
		resp := NewResponse(req.ID, false)
		resp.SetError(fmt.Sprintf("Unknown request type: %s", req.Type))
		return resp
	}
}

// Virtual filesystem handlers - thin wrappers that delegate to storage

// handleStorePath stores content at a virtual path
func (d *Daemon) handleStorePath(req Request) Response {
	var payload struct {
		Path     string                 `json:"path"`
		Content  string                 `json:"content"` // base64 encoded
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Decode content
	content, err := base64.StdEncoding.DecodeString(payload.Content)
	if err != nil {
		return NewErrorResponse(req.ID, "Failed to decode content: "+err.Error())
	}

	// Delegate to storage
	result, err := d.storage.HandleStorePath(payload.Path, content, payload.Metadata)
	if err != nil {
		return NewErrorResponse(req.ID, err.Error())
	}

	resp := NewResponse(req.ID, true)
	resp.SetData(result)
	return resp
}

// handleUpdatePath updates content at a virtual path
func (d *Daemon) handleUpdatePath(req Request) Response {
	var payload struct {
		Path            string                 `json:"path"`
		Content         string                 `json:"content,omitempty"` // base64, optional
		MetadataUpdates map[string]interface{} `json:"metadata_updates,omitempty"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Decode content if provided
	var content []byte
	if payload.Content != "" {
		var err error
		content, err = base64.StdEncoding.DecodeString(payload.Content)
		if err != nil {
			return NewErrorResponse(req.ID, "Failed to decode content: "+err.Error())
		}
	}

	// Delegate to storage
	result, err := d.storage.HandleUpdatePath(payload.Path, content, payload.MetadataUpdates)
	if err != nil {
		return NewErrorResponse(req.ID, err.Error())
	}

	resp := NewResponse(req.ID, true)
	resp.SetData(result)
	return resp
}

// handleDeletePath removes a virtual path
func (d *Daemon) handleDeletePath(req Request) Response {
	var payload struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Delegate to storage
	result, err := d.storage.HandleDeletePath(payload.Path)
	if err != nil {
		return NewErrorResponse(req.ID, err.Error())
	}

	resp := NewResponse(req.ID, true)
	resp.SetData(result)
	return resp
}

// handleCreateMemory creates a new memory (session) thread
func (d *Daemon) handleCreateMemory(req Request) Response {
	var payload struct {
		Agent          string `json:"agent"`
		InitialMessage string `json:"initial_message,omitempty"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Delegate to storage for memory ID generation
	result, err := d.storage.HandleCreateMemory(payload.Agent, payload.InitialMessage)
	if err != nil {
		return NewErrorResponse(req.ID, err.Error())
	}

	// Extract memory ID from result
	memoryID := result["memory_id"].(string)

	// Create actual session
	session := d.getOrCreateSession(memoryID, payload.Agent)

	// Add initial message if provided
	if payload.InitialMessage != "" {
		session.mu.Lock()
		session.Messages = append(session.Messages, Message{
			Role:      "user",
			Content:   payload.InitialMessage,
			Timestamp: time.Now(),
		})
		session.mu.Unlock()

		// Save to disk
		if d.storage != nil {
			d.storage.SaveSession(session)
		}
	}

	// Add session details to result
	result["created_at"] = session.CreatedAt

	resp := NewResponse(req.ID, true)
	resp.SetData(result)
	return resp
}

// handleListPath lists entries in a virtual directory
func (d *Daemon) handleListPath(req Request) Response {
	var payload struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Default to root if no path specified
	path := payload.Path
	if path == "" {
		path = "/"
	}

	// Track path access (browsing)
	if d.contextCollector != nil && path != "/" {
		accessType := "browse"
		if strings.HasPrefix(path, "/commands/") {
			accessType = "browse-commands"
		} else if strings.HasPrefix(path, "/tools/") {
			accessType = "browse-tools"
		} else if strings.HasPrefix(path, "/memory/") {
			accessType = "browse-memory"
		}
		d.contextCollector.TrackMemoryAccess(path, accessType)
	}

	// Get directory listing
	entries := d.listVirtualPath(path)
	
	log.Printf("🔍 List operation for path '%s' returned %d entries", path, len(entries))

	resp := NewResponse(req.ID, true)
	resp.SetData(map[string]interface{}{
		"path":    path,
		"entries": entries,
	})
	return resp
}

// handleReadPath reads content from a virtual path
func (d *Daemon) handleReadPath(req Request) Response {
	var payload struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Track artifact access
	if d.contextCollector != nil {
		accessType := "artifact"
		if strings.HasPrefix(payload.Path, "/commands/") {
			accessType = "command"
		} else if strings.HasPrefix(payload.Path, "/tools/") {
			accessType = "tool"
		} else if strings.HasPrefix(payload.Path, "/memory/") {
			accessType = "memory"
		}
		d.contextCollector.TrackMemoryAccess(payload.Path, accessType)
	}

	// Resolve path to object ID
	objID := d.resolvePath(payload.Path)
	if objID == "" {
		return NewErrorResponse(req.ID, fmt.Sprintf("Path not found: %s", payload.Path))
	}

	// Read content
	content, err := d.storage.Read(objID)
	if err != nil {
		return NewErrorResponse(req.ID, fmt.Sprintf("Failed to read content: %v", err))
	}

	// Load metadata
	metadata, err := d.storage.LoadMetadata(objID)
	if err != nil {
		// Continue without metadata - it's optional
		log.Printf("Warning: Failed to load metadata for %s: %v", objID, err)
	}

	// Prepare response data
	responseData := map[string]interface{}{
		"content": base64.StdEncoding.EncodeToString(content),
		"size":    len(content),
		"path":    payload.Path,
	}

	// Add metadata if available
	if metadata != nil {
		responseData["metadata"] = map[string]interface{}{
			"type":        metadata.Type,
			"created":     metadata.Created,
			"modified":    metadata.Modified,
			"agent":       metadata.Agent,
			"session":     metadata.Session,
			"title":       metadata.Title,
			"description": metadata.Description,
		}
	}

	resp := NewResponse(req.ID, true)
	resp.SetData(responseData)
	return resp
}

// handleGetMetadata retrieves enriched metadata for a virtual path
func (d *Daemon) handleGetMetadata(req Request) Response {
	var payload struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Track metadata access
	if d.contextCollector != nil {
		accessType := "info"
		if strings.HasPrefix(payload.Path, "/commands/") {
			accessType = "info-command"
		} else if strings.HasPrefix(payload.Path, "/tools/") {
			accessType = "info-tool"
		} else if strings.HasPrefix(payload.Path, "/memory/") {
			accessType = "info-memory"
		}
		d.contextCollector.TrackMemoryAccess(payload.Path, accessType)
	}

	// Resolve path to object ID
	objID := d.resolvePath(payload.Path)
	if objID == "" {
		return NewErrorResponse(req.ID, fmt.Sprintf("Path not found: %s", payload.Path))
	}

	// Special handling for relation IDs - extract data directly from relation
	if strings.HasPrefix(objID, "relation:") {
		return d.handleRelationInfo(req.ID, payload.Path, objID)
	}

	// Load metadata
	metadata, err := d.storage.LoadMetadata(objID)
	if err != nil {
		// Try to create basic metadata if none exists
		metadata = &Metadata{
			ID:      objID,
			Type:    "unknown",
			Created: time.Now(),
			Paths:   []string{payload.Path},
		}
	}

	// Get actual content size
	content, err := d.storage.Read(objID)
	actualSize := int64(0)
	if err == nil {
		actualSize = int64(len(content))
		metadata.Size = actualSize
	}

	// Prepare enriched metadata response
	responseData := map[string]interface{}{
		"path":      payload.Path,
		"object_id": objID,
		"type":      metadata.Type,
		"subtype":   metadata.Subtype,
		"created":   metadata.Created,
		"modified":  metadata.Modified,
		"accessed":  metadata.Accessed,
		"size":      actualSize,
		
		// Content info
		"title":       metadata.Title,
		"description": metadata.Description,
		"tags":        metadata.Tags,
		
		// Context
		"session":    metadata.Session,
		"agent":      metadata.Agent,
		"lifecycle":  metadata.Lifecycle,
		"importance": metadata.Importance,
		"usage_count": metadata.UsageCount,
		
		// Relationships
		"paths":         metadata.Paths,
		"relationships": metadata.Relationships,
		
		// Computed fields
		"age_seconds":       time.Since(metadata.Created).Seconds(),
		"modified_seconds":  time.Since(metadata.Modified).Seconds(),
	}

	// Add active session info if this is a memory path
	if strings.HasPrefix(payload.Path, "/memory/") {
		parts := strings.Split(payload.Path, "/")
		if len(parts) >= 3 {
			sessionID := parts[2]
			d.mu.RLock()
			if session, exists := d.sessions[sessionID]; exists {
				responseData["active_session"] = map[string]interface{}{
					"state":         string(session.State),
					"message_count": len(session.Messages),
					"last_activity": session.LastActivity,
				}
			}
			d.mu.RUnlock()
		}
	}

	resp := NewResponse(req.ID, true)
	resp.SetData(responseData)
	return resp
}

// handleSearch searches across the virtual filesystem
func (d *Daemon) handleSearch(req Request) Response {
	var payload struct {
		Query   string        `json:"query"`
		Mode    string        `json:"mode,omitempty"`
		Filters SearchFilters `json:"filters"`
	}

	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, "Invalid payload: "+err.Error())
	}

	// Default to OR mode if not specified
	if payload.Mode == "" {
		payload.Mode = "or"
	}

	// Perform search with mode
	results, err := d.storage.SearchObjects(payload.Query, payload.Mode, payload.Filters)
	if err != nil {
		return NewErrorResponse(req.ID, fmt.Sprintf("Search failed: %v", err))
	}

	resp := NewResponse(req.ID, true)
	resp.SetData(map[string]interface{}{
		"query":   payload.Query,
		"mode":    payload.Mode,
		"filters": payload.Filters,
		"results": results,
		"count":   len(results),
	})
	return resp
}

func (d *Daemon) handleGetLastSession(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	// Parse payload
	var payload struct {
		Agent string `json:"agent"`
	}
	
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		resp.SetError("Invalid payload: " + err.Error())
		return resp
	}
	
	if payload.Agent == "" {
		resp.SetError("agent parameter required")
		return resp
	}
	
	// Normalize agent name (remove @ if present)
	agent := strings.TrimPrefix(payload.Agent, "@")
	
	// Get most recent session from storage for this agent
	if d.storage == nil {
		resp.SetError("Storage not initialized")
		return resp
	}
	
	sessionID, err := d.storage.GetLastSession(agent)
	if err != nil {
		resp.SetError(fmt.Sprintf("No sessions found for %s: %v", agent, err))
		return resp
	}
	
	// Load session to get metadata (returns PersistentSession)
	session, err := d.storage.LoadSession(sessionID)
	if err != nil {
		resp.SetError(fmt.Sprintf("Failed to load session: %v", err))
		return resp
	}
	
	data := map[string]interface{}{
		"session_id":    session.ID,
		"agent":         session.Agent,
		"last_activity": session.LastActivity.Format(time.RFC3339),
		"message_count": len(session.Messages),
	}
	
	resp.SetData(data)
	return resp
}

// handleGetContext returns current session context information
func (d *Daemon) handleGetContext(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	// Use the context collector if available
	if d.contextCollector != nil {
		contextData := d.contextCollector.Collect()
		resp.SetData(contextData)
	} else {
		// Fallback to basic implementation if collector not initialized
		contextData := &ContextData{
			RecentCommands:   []CommandRecord{},
			CreatedTools:     []ToolRecord{},
			AccessedMemories: []MemoryAccess{},
			Suggestions:      []ContextSuggestion{},
		}
		resp.SetData(contextData)
	}
	
	return resp
}

// Path resolution methods

// resolvePath resolves a virtual path to an object ID
func (d *Daemon) resolvePath(path string) string {
	if d.storage == nil {
		return ""
	}
	return d.storage.ResolvePath(path)
}

// listVirtualPath lists entries in a virtual directory
func (d *Daemon) listVirtualPath(path string) []map[string]interface{} {
	if d.storage == nil {
		return []map[string]interface{}{}
	}
	
	// Use storage method that includes active sessions
	return d.storage.ListPathWithActiveSessions(path, d.sessions)
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
	if d.storage != nil {
		if persistedSession, err := d.storage.LoadSession(sessionID); err == nil {
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
	
	// Track memory creation in context collector
	if d.contextCollector != nil {
		memoryPath := fmt.Sprintf("/memory/%s", sessionID)
		d.contextCollector.TrackMemoryAccess(memoryPath, "created")
		log.Printf("🧠 Tracked new memory creation: %s", memoryPath)
	}
	
	// Save new session to disk
	log.Printf("🔍 Memory store check: memoryStore != nil: %v", d.storage != nil)
	if d.storage != nil {
		log.Printf("🔍 [NEW_SESSION] Saving newly created session %s", sessionID)
		go func() {
			if err := d.storage.SaveSession(session); err != nil {
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
	sessions, err := d.storage.LoadRecentSessions(1) // Last 24 hours
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
						if d.storage != nil {
							go d.storage.SaveSession(session)
						}
					}
					
				case SessionIdle:
					// Check if session should be abandoned (2x idle timeout)
					if timeSinceActivity > session.IdleTimeout*2 {
						session.State = SessionAbandoned
						log.Printf("🚪 Session %s abandoned (idle for %v)", id, timeSinceActivity)
						
						// Save final state and remove from memory
						if d.storage != nil {
							go d.storage.SaveSession(session)
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
				if d.storage != nil && (session.State == SessionActive || session.State == SessionIdle) {
					d.storage.SaveSession(session)
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
	
	// Debug logging to see what port is stored
	log.Printf("DEBUG: handleStatus called, d.config.Port = '%s'", d.config.Port)
	
	uptime := time.Since(startTime).Round(time.Second).String()
	
	d.mu.RLock()
	activeSessions := 0
	for _, session := range d.sessions {
		if session.State == SessionActive {
			activeSessions++
		}
	}
	d.mu.RUnlock()
	
	// Get rule engine status
	var ruleCount int
	var ruleNames []string
	if d.realityCompiler != nil && d.realityCompiler.ruleEngine != nil {
		rules := d.realityCompiler.ruleEngine.ListRules()
		ruleCount = len(rules)
		for _, rule := range rules {
			if rule.Enabled {
				ruleNames = append(ruleNames, rule.Name)
			}
		}
	}
	
	var rulesStatus string
	if ruleCount > 0 {
		rulesStatus = fmt.Sprintf("%d active rules: %s", len(ruleNames), strings.Join(ruleNames, ", "))
	} else {
		rulesStatus = "No rules loaded"
	}

	status := StatusData{
		Status:    "swimming",
		Port:      d.config.Port,
		Sessions:  activeSessions,
		Uptime:    uptime,
		Dolphins:  "🐬🐬🐬 laughing in the digital waves",
		RuleCount: ruleCount,
		Rules:     rulesStatus,
	}
	
	resp.SetData(status)
	return resp
}

// handleWatch handles watch requests for real-time monitoring
func (d *Daemon) handleWatch(req Request) Response {
	// Parse the watch payload
	var payload WatchPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return NewErrorResponse(req.ID, fmt.Sprintf("Invalid watch payload: %v", err))
	}
	
	resp := NewResponse(req.ID, true)
	
	// Handle different watch targets
	switch payload.Target {
	case "rules":
		return d.handleWatchRules(req)
	default:
		return NewErrorResponse(req.ID, fmt.Sprintf("Unsupported watch target: %s", payload.Target))
	}
	
	return resp
}

// handleWatchRules provides real-time rule engine activity monitoring
func (d *Daemon) handleWatchRules(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	// For now, return current rule status since we don't have streaming yet
	// This would be enhanced to stream real-time events in a full implementation
	if d.realityCompiler == nil || d.realityCompiler.ruleEngine == nil {
		data := WatchData{
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "status",
			RuleID:    "system",
			RuleName:  "Rule Engine Status",
			Details:   "Rule engine not initialized",
		}
		resp.SetData(data)
		return resp
	}
	
	rules := d.realityCompiler.ruleEngine.ListRules()
	var watchData []WatchData
	
	for _, rule := range rules {
		status := "enabled"
		if !rule.Enabled {
			status = "disabled"
		}
		
		data := WatchData{
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "rule_status",
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Details:   fmt.Sprintf("Status: %s, Description: %s", status, rule.Description),
		}
		watchData = append(watchData, data)
	}
	
	resp.SetData(watchData)
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
		SessionID      string `json:"session_id,omitempty"`
		IncludeContent bool   `json:"include_content,omitempty"`
	}
	
	log.Printf("🔍 [DEBUG] Memory endpoint - request ID: %s", req.ID)
	log.Printf("🔍 [DEBUG] Memory endpoint - payload: %s", string(req.Payload))
	
	if req.Payload != nil && len(req.Payload) > 0 {
		if err := json.Unmarshal(req.Payload, &payload); err == nil && payload.SessionID != "" {
			log.Printf("🔍 [DEBUG] Memory endpoint - parsed session_id: %s, include_content: %t", payload.SessionID, payload.IncludeContent)
			// Handle specific session request
			return d.handleMemoryShow(req, payload.SessionID)
		} else if err != nil {
			log.Printf("🔍 [DEBUG] Memory endpoint - payload unmarshal error: %v", err)
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
	
	if d.storage != nil {
		// Load last 7 days of sessions
		if sessions, err := d.storage.LoadRecentSessions(7); err == nil {
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
		// Convert StorageStats to MemoryStats for compatibility
		sStats := d.storage.GetStats()
		stats = &MemoryStats{
			TotalSessions:     sStats.TotalSessions,
			ActiveSessions:    sStats.ActiveSessions,
			CommandsGenerated: 0, // TODO: track commands
			LastSessionTime:   sStats.LastUpdated,
		}
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
	
	log.Printf("🔍 [DEBUG] handleMemoryShow - looking for session: %s", sessionID)
	
	// Track memory access
	if d.contextCollector != nil {
		memoryPath := fmt.Sprintf("/memory/%s", sessionID)
		d.contextCollector.TrackMemoryAccess(memoryPath, "session")
	}
	
	// First check in-memory sessions
	d.mu.RLock()
	if session, exists := d.sessions[sessionID]; exists {
		d.mu.RUnlock()
		
		log.Printf("🔍 [DEBUG] handleMemoryShow - found session in memory, messages: %d", len(session.Messages))
		
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
		
		log.Printf("🔍 [DEBUG] handleMemoryShow - returning data with %d messages", len(session.Messages))
		return resp
	}
	d.mu.RUnlock()
	
	log.Printf("🔍 [DEBUG] handleMemoryShow - session not in memory, checking storage")
	
	// Try to load from disk
	if d.storage != nil {
		if session, err := d.storage.LoadSession(sessionID); err == nil {
			log.Printf("🔍 [DEBUG] handleMemoryShow - found session on disk, messages: %d", len(session.Messages))
			
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
			
			log.Printf("🔍 [DEBUG] handleMemoryShow - returning data from disk with %d messages", len(session.Messages))
			return resp
		} else {
			log.Printf("🔍 [DEBUG] handleMemoryShow - failed to load from storage: %v", err)
		}
	} else {
		log.Printf("🔍 [DEBUG] handleMemoryShow - no storage available")
	}
	
	// Session not found
	log.Printf("🔍 [DEBUG] handleMemoryShow - session not found: %s", sessionID)
	resp.SetError(fmt.Sprintf("Session '%s' not found", sessionID))
	return resp
}

// Reality Compiler handlers

// handleDeclareRelation declares a new relation and materializes it
func (d *Daemon) handleDeclareRelation(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	// Step 5: Early validation - fail fast with helpful messages
	if d.validator != nil {
		// Create validation request from raw request data
		validationReq := map[string]interface{}{
			"user_prompt": req.UserPrompt,
			"references":  req.References,
		}
		
		if validationResult := d.validator.ValidateRequest(validationReq); validationResult.HasErrors() {
			userFriendlyError := d.validator.FormatErrors(validationResult.Errors)
			resp.SetError(userFriendlyError)
			return resp
		}
	}
	
	if d.realityCompiler == nil {
		resp.SetError("Reality compiler not initialized")
		return resp
	}
	
	// Parse relation from payload
	var payload struct {
		Relation Relation `json:"relation"`
	}
	
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		resp.SetError("Invalid relation payload: " + err.Error())
		return resp
	}
	
	// Set ID if not provided
	if payload.Relation.ID == "" {
		payload.Relation.ID = generateRelationID(payload.Relation.Type, 
			fmt.Sprintf("%v", payload.Relation.Properties["name"]))
	}
	
	// Step 5: Capture session context for memory-relation bridge
	if req.SessionContext != nil && req.SessionContext.SessionID != "" {
		// Add memory session properties to relation
		if payload.Relation.Properties == nil {
			payload.Relation.Properties = make(map[string]interface{})
		}
		payload.Relation.Properties["memory_session"] = req.SessionContext.SessionID
		if req.SessionContext.Agent != "" {
			payload.Relation.Properties["crystallized_agent"] = req.SessionContext.Agent
		}
		log.Printf("🔗 Linking relation %s to memory session %s", 
			payload.Relation.ID, req.SessionContext.SessionID)
	}
	
	// Phase 1: Universal References - Validate, store, and resolve references
	if len(req.References) > 0 {
		// Store original references in relation properties
		if payload.Relation.Properties == nil {
			payload.Relation.Properties = make(map[string]interface{})
		}
		payload.Relation.Properties["references"] = req.References
		log.Printf("📎 References stored for %s: %d references", 
			payload.Relation.ID, len(req.References))
		
		// Use common reference handler for resolution
		if d.referenceHandler != nil {
			result := d.referenceHandler.ResolveReferences(req.References, "declare")
			if result.Success {
				// Store resolved context for AI generation
				contextStr := d.referenceHandler.FormatForDeclare(result.ResolvedText)
				payload.Relation.Properties["resolved_context"] = contextStr
				log.Printf("✨ Resolved context stored (%d chars)", len(contextStr))
			} else if result.Error != nil {
				log.Printf("⚠️ Reference resolution failed: %v", result.Error)
				// For declare mode, we could fail the request or continue with graceful degradation
				// Continuing with graceful degradation for consistency
			}
		} else {
			log.Printf("⚠️ No reference handler available - skipping reference resolution")
		}
	}
	
	// Phase 3: Universal User Prompt - Store user prompt if provided
	if req.UserPrompt != "" {
		if payload.Relation.Properties == nil {
			payload.Relation.Properties = make(map[string]interface{})
		}
		payload.Relation.Properties["user_prompt"] = req.UserPrompt
		
		log.Printf("💬 User prompt stored for %s: %.100s...", 
			payload.Relation.ID, req.UserPrompt)
	}
	
	// Add default agent for Tool relations created via direct declare
	if payload.Relation.Type == "Tool" {
		if payload.Relation.Properties == nil {
			payload.Relation.Properties = make(map[string]interface{})
		}
		// Only set agent if not already set (preserve session agents)
		if _, hasAgent := payload.Relation.Properties["agent"]; !hasAgent {
			payload.Relation.Properties["agent"] = "@ai-engineer"
		}
	}
	
	// Declare and materialize the relation
	entity, err := d.realityCompiler.DeclareRelation(payload.Relation)
	if err != nil {
		resp.SetError("Failed to declare relation: " + err.Error())
		return resp
	}
	
	// Step 6 Phase C: Create similarity relationships for new tools
	// Process in background after successful response to avoid any blocking
	if payload.Relation.Type == "Tool" && d.realityCompiler != nil {
		relationCopy := payload.Relation // Copy for goroutine safety
		go func() {
			// Add a small delay to ensure main response is sent first
			time.Sleep(100 * time.Millisecond)
			
			defer func() {
				// Catch any panics in similarity processing
				if r := recover(); r != nil {
					log.Printf("⚠️ Panic in similarity processing: %v", r)
				}
			}()
			
			similarityCalculator := NewSimilarityCalculator(d.realityCompiler.GetRelationStore())
			if similarityCalculator != nil {
				err := similarityCalculator.createSimilarityRelationships(relationCopy, 0.5)
				if err != nil {
					log.Printf("⚠️ Failed to create similarity relationships for %s: %v", 
						relationCopy.ID, err)
				} else {
					log.Printf("🔗 Similarity relationships processed for %s", 
						relationCopy.Properties["name"])
				}
			}
		}()
	}
	
	// Return success with materialized entity info
	data := map[string]interface{}{
		"relation_id":    payload.Relation.ID,
		"type":          payload.Relation.Type,
		"materialized":  true,
		"physical_path": entity.PhysicalPath,
		"status":        entity.Status,
	}
	
	resp.SetData(data)
	return resp
}

// handleGetRelation retrieves a relation by ID
func (d *Daemon) handleGetRelation(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	if d.realityCompiler == nil {
		resp.SetError("Reality compiler not initialized")
		return resp
	}
	
	var payload struct {
		RelationID string `json:"relation_id"`
	}
	
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		resp.SetError("Invalid payload: " + err.Error())
		return resp
	}
	
	relation, err := d.realityCompiler.GetRelation(payload.RelationID)
	if err != nil {
		resp.SetError("Failed to get relation: " + err.Error())
		return resp
	}
	
	resp.SetData(map[string]interface{}{
		"relation": relation,
	})
	return resp
}

// handleListRelations lists all relations or relations of a specific type
func (d *Daemon) handleListRelations(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	if d.realityCompiler == nil {
		resp.SetError("Reality compiler not initialized")
		return resp
	}
	
	var payload struct {
		Type string `json:"type,omitempty"`
	}
	
	// Parse payload (optional)
	if req.Payload != nil {
		json.Unmarshal(req.Payload, &payload)
	}
	
	var relations []Relation
	var err error
	
	if payload.Type != "" {
		relations, err = d.realityCompiler.ListRelationsByType(payload.Type)
	} else {
		relations, err = d.realityCompiler.ListRelations()
	}
	
	if err != nil {
		resp.SetError("Failed to list relations: " + err.Error())
		return resp
	}
	
	resp.SetData(map[string]interface{}{
		"relations": relations,
		"count":     len(relations),
	})
	return resp
}

// handleDeleteRelation deletes a relation and dematerializes it
func (d *Daemon) handleDeleteRelation(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	if d.realityCompiler == nil {
		resp.SetError("Reality compiler not initialized")
		return resp
	}
	
	var payload struct {
		RelationID string `json:"relation_id"`
	}
	
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		resp.SetError("Invalid payload: " + err.Error())
		return resp
	}
	
	if err := d.realityCompiler.DeleteRelation(payload.RelationID); err != nil {
		resp.SetError("Failed to delete relation: " + err.Error())
		return resp
	}
	
	resp.SetData(map[string]interface{}{
		"deleted": true,
		"relation_id": payload.RelationID,
	})
	return resp
}

// initializeRealityCompiler sets up the reality compiler with materializers
func (d *Daemon) initializeRealityCompiler() error {
	// Initialize relation store
	relationStore, err := NewFileRelationStore(d.baseDir)
	if err != nil {
		return fmt.Errorf("failed to initialize relation store: %w", err)
	}
	
	// Initialize materialization store  
	matStore, err := NewFileMaterializationStore(d.baseDir)
	if err != nil {
		return fmt.Errorf("failed to initialize materialization store: %w", err)
	}
	
	// Initialize AI client for tool generation
	aiClient := NewAnthropicClient()
	
	// Initialize tool materializer with context collector
	log.Printf("🔧 Creating tool materializer with context collector: %v", d.contextCollector != nil)
	toolMaterializer, err := NewToolMaterializer(aiClient, d.storage, matStore, d.contextCollector)
	if err != nil {
		return fmt.Errorf("failed to initialize tool materializer: %w", err)
	}
	
	// Create reality compiler with materializers
	materializers := []Materializer{
		toolMaterializer,
		// TODO: Add more materializers in future steps (artifact, memory, etc.)
	}
	
	d.realityCompiler = NewRealityCompiler(relationStore, materializers)
	
	// Initialize rule engine with default rules
	ruleEngine := NewRuleEngine(d.realityCompiler, defaultRules())
	d.realityCompiler.SetRuleEngine(ruleEngine)
	
	log.Printf("🎯 Reality compiler initialized with %d rules", len(ruleEngine.ListRules()))
	
	return nil
}

// relationsAdapter adapts daemon's Relations system to resolution.RelationsManager interface
type relationsAdapter struct {
	realityCompiler *RealityCompiler
	storage         *Storage
}

func (ra *relationsAdapter) DeclareRelation(relation *resolution.URLArtifactRelation) error {
	// Store content in object storage first
	contentID, err := ra.storage.Store([]byte(relation.Content))
	if err != nil {
		return fmt.Errorf("failed to store URL content: %w", err)
	}
	
	// Create daemon Relation
	daemonRelation := Relation{
		ID:         relation.ID,
		Type:       relation.Type,
		Properties: relation.Properties,
		CreatedAt:  relation.CreatedAt,
		UpdatedAt:  relation.UpdatedAt,
	}
	
	// Add content reference to properties
	if daemonRelation.Properties == nil {
		daemonRelation.Properties = make(map[string]interface{})
	}
	daemonRelation.Properties["content_id"] = contentID
	
	// Declare via reality compiler
	_, err = ra.realityCompiler.DeclareRelation(daemonRelation)
	return err
}

func (ra *relationsAdapter) GetRelationByID(id string) (*resolution.URLArtifactRelation, error) {
	// Get relation from reality compiler
	relation, err := ra.realityCompiler.GetRelation(id)
	if err != nil {
		return nil, err
	}
	
	// Convert to resolution type
	urlRelation := &resolution.URLArtifactRelation{
		ID:         relation.ID,
		Type:       relation.Type,
		Properties: relation.Properties,
		CreatedAt:  relation.CreatedAt,
		UpdatedAt:  relation.UpdatedAt,
	}
	
	// Load content if available
	if contentID, exists := relation.Properties["content_id"].(string); exists {
		content, err := ra.storage.Read(contentID)
		if err == nil {
			urlRelation.Content = string(content)
			urlRelation.ContentID = contentID
		}
	}
	
	return urlRelation, nil
}

func (ra *relationsAdapter) ListRelationsByType(relationType string) ([]*resolution.URLArtifactRelation, error) {
	// Get relations from reality compiler
	relations, err := ra.realityCompiler.ListRelationsByType(relationType)
	if err != nil {
		return nil, err
	}
	
	var urlRelations []*resolution.URLArtifactRelation
	for _, relation := range relations {
		urlRelation := &resolution.URLArtifactRelation{
			ID:         relation.ID,
			Type:       relation.Type,
			Properties: relation.Properties,
			CreatedAt:  relation.CreatedAt,
			UpdatedAt:  relation.UpdatedAt,
		}
		
		// Load content if available (optional for listing)
		if contentID, exists := relation.Properties["content_id"].(string); exists {
			urlRelation.ContentID = contentID
			// Don't load content for listings to save memory
		}
		
		urlRelations = append(urlRelations, urlRelation)
	}
	
	return urlRelations, nil
}

// initializeResolutionManager initializes the Phase 2 reference resolution system
func (d *Daemon) initializeResolutionManager() error {
	// Create handlers that bridge to daemon functionality
	handlers := resolution.Handlers{
		// Search handler - queries the storage system
		SearchHandler: func(query string, limit int) ([]resolution.SearchResult, error) {
			log.Printf("🔍 Search handler called for: %s (limit: %d)", query, limit)
			
			if d.storage == nil {
				log.Printf("⚠️ Storage not available for search")
				return []resolution.SearchResult{}, nil
			}
			
			// Create search filters
			filters := SearchFilters{
				Limit: limit,
			}
			
			// Execute search using storage system
			results, err := d.storage.SearchObjects(query, "or", filters)
			if err != nil {
				log.Printf("❌ Search failed: %v", err)
				return []resolution.SearchResult{}, nil // Return empty results, don't fail resolution
			}
			
			// Convert daemon SearchResult to resolution SearchResult
			var resolverResults []resolution.SearchResult
			for _, result := range results {
				resolverResults = append(resolverResults, resolution.SearchResult{
					Path:       result.Path,
					Type:       result.Type,
					Score:      result.Score,
					Title:      result.Metadata.Title,
					Summary:    result.Snippet,
					Properties: map[string]interface{}{
						"object_id":    result.ObjectID,
						"match_fields": result.MatchFields,
						"created":      result.Metadata.Created,
						"tags":         result.Metadata.Tags,
					},
				})
			}
			
			log.Printf("✅ Search completed: %d results found", len(resolverResults))
			return resolverResults, nil
		},
		
		// Tool handler - queries relations store for existing tools
		ToolHandler: func(toolName string) (*resolution.ToolDefinition, error) {
			log.Printf("🔧 Tool handler called for: %s", toolName)
			
			if d.realityCompiler == nil {
				log.Printf("⚠️ Reality compiler not available for tool lookup")
				return nil, nil // Don't fail resolution, just return empty
			}
			
			// Get all tool relations
			toolRelations, err := d.realityCompiler.ListRelationsByType("Tool")
			if err != nil {
				log.Printf("❌ Failed to list tool relations: %v", err)
				return nil, nil // Don't fail resolution, just return empty
			}
			
			// Search for tool by name in relation ID or properties
			for _, relation := range toolRelations {
				// Check if relation ID contains the tool name
				if strings.Contains(strings.ToLower(relation.ID), strings.ToLower(toolName)) {
					// Found matching tool relation
					var transforms []string
					var commands []string
					
					// Extract transforms from properties
					if transformsVal, exists := relation.Properties["transforms"]; exists {
						if transformsList, ok := transformsVal.([]interface{}); ok {
							for _, t := range transformsList {
								if tStr, ok := t.(string); ok {
									transforms = append(transforms, tStr)
								}
							}
						} else if transformsStr, ok := transformsVal.(string); ok {
							transforms = []string{transformsStr}
						}
					}
					
					// Check for generated commands in properties
					if commandsVal, exists := relation.Properties["generated_commands"]; exists {
						if commandsList, ok := commandsVal.([]interface{}); ok {
							for _, cmd := range commandsList {
								if cmdStr, ok := cmd.(string); ok {
									commands = append(commands, cmdStr)
								}
							}
						}
					}
					
					// Extract agent from properties if available
					agent := ""
					if agentVal, exists := relation.Properties["agent"]; exists {
						if agentStr, ok := agentVal.(string); ok {
							agent = agentStr
						}
					}
					
					toolDef := &resolution.ToolDefinition{
						ID:         relation.ID,
						Name:       toolName,
						Type:       relation.Type,
						Transforms: transforms,
						Commands:   commands,
						Properties: relation.Properties,
						Created:    relation.CreatedAt.Format("2006-01-02T15:04:05Z"),
						Agent:      agent,
					}
					
					log.Printf("✅ Tool found: %s (ID: %s)", toolName, relation.ID)
					return toolDef, nil
				}
			}
			
			log.Printf("⚠️ Tool '%s' not found in %d tool relations", toolName, len(toolRelations))
			return nil, nil // Don't fail resolution, just return empty
		},
		
		
		// File handler - local filesystem with security boundaries
		FileHandler: func(path string) (*resolution.FileContent, error) {
			log.Printf("📄 File handler called for: %s", path)
			return d.handleLocalFile(path)
		},
		
		// P42 handler - Port 42 VFS and crystallized knowledge access
		P42Handler: func(p42Path string) (*resolution.FileContent, error) {
			log.Printf("🏗️ P42 handler called for: %s", p42Path)
			return d.handleP42File(p42Path)
		},
		
		// Relations handler - provides access to relations for URL artifact caching
		RelationsHandler: func() resolution.RelationsManager {
			return &relationsAdapter{
				realityCompiler: d.realityCompiler,
				storage:         d.storage,
			}
		},
	}
	
	d.resolutionService = resolution.NewResolutionService(handlers)
	log.Printf("🔗 Resolution service initialized")
	return nil
}

// Command generation functionality
func (d *Daemon) generateCommand(spec *CommandSpec) error {
	log.Printf("🔍 [GENERATE_COMMAND] Starting generation for '%s' (session=%s)", spec.Name, spec.SessionID)
	
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
	
	// Store command using unified storage
	if d.storage == nil {
		return fmt.Errorf("storage not initialized")
	}
	
	// Store command with metadata and symlink
	if err := d.storage.StoreCommand(spec, code); err != nil {
		return fmt.Errorf("failed to store command: %v", err)
	}
	
	// Create a relation for VFS integration (same pattern as declare tool)
	if d.realityCompiler != nil {
		relation := Relation{
			ID:   fmt.Sprintf("tool-%s-%d", spec.Name, time.Now().Unix()),
			Type: "Tool",
			Properties: map[string]interface{}{
				"name":        spec.Name,
				"description": spec.Description,
				"language":    spec.Language,
				"source":      "possess", // Track creation method
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		// Add dependencies if present
		if len(spec.Dependencies) > 0 {
			relation.Properties["dependencies"] = spec.Dependencies
		}
		
		// Add session info if present
		if spec.SessionID != "" {
			relation.Properties["session_id"] = spec.SessionID
		}
		
		// Add agent info if present  
		if spec.Agent != "" {
			relation.Properties["agent"] = spec.Agent
		}
		
		log.Printf("🔗 Creating relation for possess-generated command: %s", relation.ID)
		if _, err := d.realityCompiler.DeclareRelation(relation); err != nil {
			log.Printf("⚠️ Failed to create relation for command %s: %v", spec.Name, err)
			// Don't fail the command generation, just log the issue
		} else {
			log.Printf("✅ Relation created for possess-generated command: %s", spec.Name)
		}
	}
	
	// Log to memory (simple for now)
	d.logCommandGeneration(spec)
	
	return nil
}

// Artifact generation functionality
func (d *Daemon) generateArtifact(spec *ArtifactSpec) error {
	log.Printf("🔍 [GENERATE_ARTIFACT] Starting generation for '%s' (type=%s, session=%s)", 
		spec.Name, spec.Type, spec.SessionID)
	
	// Check if storage is available
	if d.storage == nil {
		return fmt.Errorf("storage not initialized")
	}
	
	// Determine the base path for the artifact
	basePath := fmt.Sprintf("/artifacts/%s/%s", spec.Type, spec.Name)
	
	// Handle single file vs multi-file artifacts
	if spec.SingleFile != "" {
		// Single file artifact
		fullPath := basePath
		if spec.Format != "" {
			// Add extension if not already present
			if !strings.Contains(spec.Name, ".") {
				fullPath = fmt.Sprintf("%s.%s", basePath, spec.Format)
			}
		}
		
		// Store the single file
		metadata := map[string]interface{}{
			"type":                 spec.Type,
			"format":               spec.Format,
			"description":          spec.Description,
			"crystallization_type": "artifact",
			"session":              spec.SessionID,
			"agent":                spec.Agent,
		}
		
		// Add any additional metadata
		for k, v := range spec.Metadata {
			metadata[k] = v
		}
		
		result, err := d.storage.HandleStorePath(fullPath, []byte(spec.SingleFile), metadata)
		if err != nil {
			return fmt.Errorf("failed to store artifact: %v", err)
		}
		
		log.Printf("✨ Artifact stored: %s (id=%s)", fullPath, result["id"])
		
	} else if spec.Content != nil && len(spec.Content) > 0 {
		// Multi-file artifact (e.g., a web app with multiple files)
		for filePath, content := range spec.Content {
			fullPath := fmt.Sprintf("%s/%s", basePath, filePath)
			
			// Infer file type from extension
			fileType := "file"
			if strings.HasSuffix(filePath, ".md") {
				fileType = "document"
			} else if strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".py") {
				fileType = "code"
			} else if strings.HasSuffix(filePath, ".html") || strings.HasSuffix(filePath, ".css") {
				fileType = "web"
			}
			
			metadata := map[string]interface{}{
				"type":                 fileType,
				"parent_type":          spec.Type,
				"description":          spec.Description,
				"crystallization_type": "artifact",
				"session":              spec.SessionID,
				"agent":                spec.Agent,
				"artifact_name":        spec.Name,
			}
			
			result, err := d.storage.HandleStorePath(fullPath, []byte(content), metadata)
			if err != nil {
				log.Printf("❌ Failed to store file %s: %v", filePath, err)
				continue
			}
			
			log.Printf("✨ Artifact file stored: %s (id=%s)", fullPath, result["id"])
		}
	}
	
	log.Printf("🎨 Artifact generation completed: %s", spec.Name)
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
	
	// Use AI-generated tags if available (preferred over word-splitting)
	if len(spec.Tags) > 0 {
		tags = append(tags, spec.Tags...)
	} else {
		// Fallback to word extraction for legacy tools
		words := strings.Fields(spec.Name + " " + spec.Description)
		for _, word := range words {
			word = strings.ToLower(word)
			// Add meaningful words as tags (skip common words)
			if len(word) > 3 && !isCommonWord(word) {
				tags = append(tags, word)
			}
		}
	}
	
	return tags
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

// handleLocalFile implements secure local file access for file: references
func (d *Daemon) handleLocalFile(path string) (*resolution.FileContent, error) {
	// Expand tilde (~) to home directory if present
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("❌ Failed to get home directory: %v", err)
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
		log.Printf("🏠 Expanded tilde path to: %s", path)
	}
	
	// Security: Clean path and prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		log.Printf("🚨 SECURITY WARNING: Path traversal attempt blocked - %s", path)
		return nil, fmt.Errorf("path traversal not allowed: %s", path)
	}
	
	// Security: Convert relative paths to absolute to check boundaries
	var absPath string
	var err error
	if filepath.IsAbs(cleanPath) {
		// For absolute paths, use as-is but be restrictive
		absPath = cleanPath
	} else {
		// For relative paths, resolve against current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		absPath = filepath.Join(cwd, cleanPath)
	}
	
	// Security: Only allow files within reasonable boundaries
	if !d.isFileAccessAllowed(absPath) {
		log.Printf("🚨 SECURITY WARNING: File access boundary violation blocked - %s", path)
		return nil, fmt.Errorf("file access not allowed: %s", path)
	}
	
	// Check if file exists and get info
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to access file: %w", err)
	}
	
	// Security: Check file size (prevent memory exhaustion)
	const maxFileSize = 1 * 1024 * 1024 // 1MB
	if fileInfo.Size() > maxFileSize {
		log.Printf("🚨 SECURITY WARNING: Large file access attempt blocked - %s (%d bytes > %d bytes)", 
			path, fileInfo.Size(), maxFileSize)
		return nil, fmt.Errorf("file too large: %s (size: %d bytes, max: %d bytes)", 
			path, fileInfo.Size(), maxFileSize)
	}
	
	// Security: Only allow certain file types
	if !d.isFileTypeAllowed(absPath) {
		log.Printf("🚨 SECURITY WARNING: Disallowed file type access attempt blocked - %s", path)
		return nil, fmt.Errorf("file type not allowed: %s", path)
	}
	
	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Detect content type
	contentType := d.detectFileType(absPath, content)
	
	log.Printf("✅ Local file accessed: %s (%d bytes, type: %s)", path, len(content), contentType)
	
	return &resolution.FileContent{
		Path:    path, // Return original path requested
		Content: string(content),
		Size:    fileInfo.Size(),
		Type:    contentType,
		Metadata: map[string]interface{}{
			"absolute_path": absPath,
			"modified":      fileInfo.ModTime(),
			"permissions":   fileInfo.Mode().String(),
			"is_dir":        fileInfo.IsDir(),
		},
	}, nil
}

// isFileAccessAllowed checks if file access is within security boundaries
func (d *Daemon) isFileAccessAllowed(absPath string) bool {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("⚠️ Failed to get working directory for security check: %v", err)
		return false
	}
	
	// Allow files within current working directory tree
	if strings.HasPrefix(absPath, cwd) {
		return true
	}
	
	// Allow common config directories relative to home
	if homeDir, err := os.UserHomeDir(); err == nil {
		// Allow .port42 directory
		port42Dir := filepath.Join(homeDir, ".port42")
		if strings.HasPrefix(absPath, port42Dir) {
			return true
		}
		
		// Allow files directly in home directory (like ~/test.txt)
		// but still exclude sensitive system paths
		if strings.HasPrefix(absPath, homeDir) && !strings.Contains(absPath, "/.ssh/") && 
		   !strings.Contains(absPath, "/.gnupg/") && !strings.Contains(absPath, "/.aws/") {
			return true
		}
	}
	
	// Deny access to system directories and files outside project
	systemPaths := []string{"/etc", "/usr", "/var", "/bin", "/sbin", "/sys", "/proc"}
	for _, sysPath := range systemPaths {
		if strings.HasPrefix(absPath, sysPath) {
			return false
		}
	}
	
	log.Printf("⚠️ File access denied for security: %s (not within working directory: %s)", absPath, cwd)
	return false
}

// isFileTypeAllowed checks if file extension/type is allowed
func (d *Daemon) isFileTypeAllowed(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	
	allowedExtensions := map[string]bool{
		// Text files
		".txt": true, ".md": true, ".rst": true,
		// Config files  
		".json": true, ".yaml": true, ".yml": true, ".toml": true, ".ini": true,
		// Code files
		".go": true, ".py": true, ".js": true, ".ts": true, ".sh": true, ".bash": true,
		".php": true, ".rb": true, ".java": true, ".c": true, ".cpp": true, ".h": true,
		// Log files
		".log": true, ".out": true,
		// Data files
		".csv": true, ".xml": true,
		// Files without extensions (often config files)
		"": true,
	}
	
	return allowedExtensions[ext]
}

// detectFileType determines file content type
func (d *Daemon) detectFileType(path string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	
	// Detect by extension first
	switch ext {
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"  
	case ".xml":
		return "application/xml"
	case ".csv":
		return "text/csv"
	case ".md":
		return "text/markdown"
	case ".log", ".out":
		return "text/log"
	case ".go", ".py", ".js", ".ts", ".sh", ".bash", ".php", ".rb", ".java", ".c", ".cpp", ".h":
		return "text/code"
	}
	
	// Detect by content if no clear extension
	contentStr := string(content)
	if len(contentStr) > 0 {
		// Check for JSON
		if strings.HasPrefix(strings.TrimSpace(contentStr), "{") || strings.HasPrefix(strings.TrimSpace(contentStr), "[") {
			return "application/json"
		}
		// Check for YAML
		if strings.Contains(contentStr, ":") && (strings.Contains(contentStr, "\n") || strings.Contains(contentStr, " ")) {
			return "application/yaml"
		}
	}
	
	return "text/plain"
}

// handleP42File implements Port 42 VFS access for p42: references
func (d *Daemon) handleP42File(p42Path string) (*resolution.FileContent, error) {
	// Clean the path
	cleanPath := strings.TrimPrefix(p42Path, "/")
	if cleanPath == "" {
		return nil, fmt.Errorf("empty P42 path: %s", p42Path)
	}
	
	log.Printf("🔍 P42 VFS access: %s", p42Path)
	
	// Method 1: Handle /tools/ paths via Relations store
	if strings.HasPrefix(p42Path, "/tools/") {
		return d.handleP42ToolPath(p42Path)
	}
	
	// Method 2: Handle /commands/ paths via direct storage lookup
	if strings.HasPrefix(p42Path, "/commands/") {
		return d.handleP42CommandPath(p42Path)  
	}
	
	// Method 3: Handle /memory/ paths via session lookup
	if strings.HasPrefix(p42Path, "/memory/") {
		return d.handleP42MemoryPath(p42Path)
	}
	
	// Method 4: General path resolution via search
	return d.handleP42SearchPath(p42Path)
}

// handleP42ToolPath resolves /tools/name paths via Relations store
func (d *Daemon) handleP42ToolPath(p42Path string) (*resolution.FileContent, error) {
	toolName := strings.TrimPrefix(p42Path, "/tools/")
	if toolName == "" {
		return nil, fmt.Errorf("invalid tool path: %s", p42Path)
	}
	
	if d.realityCompiler == nil {
		return nil, fmt.Errorf("reality compiler not available for tool lookup")
	}
	
	// Get all tool relations
	toolRelations, err := d.realityCompiler.ListRelationsByType("Tool")
	if err != nil {
		return nil, fmt.Errorf("failed to list tool relations: %w", err)
	}
	
	// Search for tool by name in relation ID
	for _, relation := range toolRelations {
		if strings.Contains(strings.ToLower(relation.ID), strings.ToLower(toolName)) {
			// Build content from tool relation
			content := d.formatToolRelationAsP42Content(relation)
			
			log.Printf("✅ P42 tool found: %s -> %s", toolName, relation.ID)
			// Extract agent from properties for proper info display
			metadata := map[string]interface{}{
				"relation_id": relation.ID,
				"relation_type": relation.Type,
				"created": relation.CreatedAt,
				"updated": relation.UpdatedAt,
				"properties": relation.Properties,
			}
			
			// Map agent from properties to expected field for info command
			if agent, exists := relation.Properties["agent"]; exists {
				metadata["Agent"] = agent
			}
			
			return &resolution.FileContent{
				Path:    p42Path,
				Content: content,
				Size:    int64(len(content)),
				Type:    "application/port42-tool",
				Metadata: metadata,
			}, nil
		}
	}
	
	return nil, fmt.Errorf("tool not found: %s", toolName)
}

// handleP42CommandPath resolves /commands/name paths via VFS direct access
func (d *Daemon) handleP42CommandPath(p42Path string) (*resolution.FileContent, error) {
	// Use VFS to resolve path to object ID - same pattern as port42 cat
	log.Printf("🔍 P42 command path resolution via VFS: %s", p42Path)
	
	objID := d.resolvePath(p42Path)
	if objID == "" {
		return nil, fmt.Errorf("command not found in VFS: %s", p42Path)
	}
	
	// Read content from storage using resolved object ID
	content, err := d.storage.Read(objID)
	if err != nil {
		return nil, fmt.Errorf("failed to read command content: %w", err)
	}
	
	log.Printf("✅ P42 command found via VFS: %s -> %s (size: %d bytes)", p42Path, objID, len(content))
	return &resolution.FileContent{
		Path:    p42Path,
		Content: string(content),
		Size:    int64(len(content)),
		Type:    "application/port42-command",
		Metadata: map[string]interface{}{
			"vfs_path": p42Path,
			"object_id": objID,
		},
	}, nil
}

// handleP42MemoryPath resolves /memory/session-id paths via session lookup
func (d *Daemon) handleP42MemoryPath(p42Path string) (*resolution.FileContent, error) {
	sessionID := strings.TrimPrefix(p42Path, "/memory/")
	if sessionID == "" {
		return nil, fmt.Errorf("invalid memory path: %s", p42Path)
	}
	
	// Remove any sub-paths (e.g., "/memory/session-123/generated" -> "session-123")
	if strings.Contains(sessionID, "/") {
		sessionID = strings.Split(sessionID, "/")[0]
	}
	
	if d.storage == nil {
		return nil, fmt.Errorf("storage not available for memory lookup")
	}
	
	// Load session from storage
	session, err := d.storage.LoadSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session %s: %w", sessionID, err)
	}
	
	log.Printf("✅ P42 memory found: %s -> session with %d messages", sessionID, len(session.Messages))
	
	// Format session as conversation transcript
	var content strings.Builder
	content.WriteString(fmt.Sprintf("=== Session: %s ===\n", sessionID))
	content.WriteString(fmt.Sprintf("Agent: %s\n", session.Agent))
	content.WriteString(fmt.Sprintf("Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Last Activity: %s\n", session.LastActivity.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Messages: %d\n\n", len(session.Messages)))
	
	// Add conversation transcript
	for i, msg := range session.Messages {
		timestamp := ""
		if !msg.Timestamp.IsZero() {
			timestamp = fmt.Sprintf(" [%s]", msg.Timestamp.Format("15:04:05"))
		}
		content.WriteString(fmt.Sprintf("%d. %s%s:\n%s\n\n", i+1, strings.ToUpper(msg.Role), timestamp, msg.Content))
	}
	
	return &resolution.FileContent{
		Path:    p42Path,
		Content: content.String(),
		Size:    int64(content.Len()),
		Type:    "session",
		Metadata: map[string]interface{}{
			"session_id":    sessionID,
			"agent":         session.Agent,
			"message_count": len(session.Messages),
			"created":       session.CreatedAt.Format("2006-01-02 15:04:05"),
			"last_activity": session.LastActivity.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// handleP42SearchPath resolves general paths via search
func (d *Daemon) handleP42SearchPath(p42Path string) (*resolution.FileContent, error) {
	if d.storage == nil {
		return nil, fmt.Errorf("storage not available for P42 path resolution")
	}
	
	// Extract search term from path
	searchTerm := strings.Trim(p42Path, "/")
	searchTerm = strings.ReplaceAll(searchTerm, "/", " ")
	
	// Search for content
	results, err := d.storage.SearchObjects(searchTerm, "or", SearchFilters{
		Limit: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search P42 path: %w", err)
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("P42 path not found: %s", p42Path)
	}
	
	// Use best match
	bestResult := results[0]
	content, err := d.storage.Read(bestResult.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to read P42 content: %w", err)
	}
	
	log.Printf("✅ P42 path resolved via search: %s -> %s (score: %.2f)", 
		p42Path, bestResult.Path, bestResult.Score)
	
	return &resolution.FileContent{
		Path:    p42Path,
		Content: string(content),
		Size:    int64(len(content)),
		Type:    "application/port42-knowledge",
		Metadata: map[string]interface{}{
			"object_id": bestResult.ObjectID,
			"storage_path": bestResult.Path,
			"score": bestResult.Score,
			"search_term": searchTerm,
			"created": bestResult.Metadata.Created,
			"title": bestResult.Metadata.Title,
		},
	}, nil
}

// formatToolRelationAsP42Content formats a tool relation as readable content
func (d *Daemon) formatToolRelationAsP42Content(relation Relation) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Port 42 Tool: %s", relation.ID))
	parts = append(parts, fmt.Sprintf("Type: %s", relation.Type))
	
	if transforms, exists := relation.Properties["transforms"]; exists {
		parts = append(parts, fmt.Sprintf("Transforms: %v", transforms))
	}
	
	if agent, exists := relation.Properties["agent"]; exists {
		parts = append(parts, fmt.Sprintf("Agent: %v", agent))
	}
	
	if description, exists := relation.Properties["description"]; exists {
		parts = append(parts, fmt.Sprintf("Description: %v", description))
	}
	
	// Add generated commands if available
	if commands, exists := relation.Properties["generated_commands"]; exists {
		parts = append(parts, fmt.Sprintf("Generated Commands: %v", commands))
	}
	
	parts = append(parts, fmt.Sprintf("Created: %s", relation.CreatedAt.Format("2006-01-02 15:04:05")))
	parts = append(parts, fmt.Sprintf("Updated: %s", relation.UpdatedAt.Format("2006-01-02 15:04:05")))
	
	return strings.Join(parts, "\n")
}

// handleRelationInfo handles info requests for relation objects
func (d *Daemon) handleRelationInfo(requestID, path, objID string) Response {
	resp := NewResponse(requestID, true)
	
	// Extract relation ID from objID (remove "relation:" prefix)
	relationID := strings.TrimPrefix(objID, "relation:")
	
	// Load relation from relation store
	if d.realityCompiler == nil || d.realityCompiler.relationStore == nil {
		return NewErrorResponse(requestID, "Relation store not available")
	}
	
	relation, err := d.realityCompiler.relationStore.Load(relationID)
	if err != nil {
		return NewErrorResponse(requestID, fmt.Sprintf("Failed to load relation: %v", err))
	}
	
	// Extract data from relation properties
	var objectType, title, description, agent string
	var size int64
	
	// Type from relation.Type
	objectType = strings.ToLower(relation.Type)  // "Tool" -> "tool"
	
	// Extract properties
	if name, ok := relation.Properties["name"].(string); ok {
		title = name
	}
	if desc, ok := relation.Properties["description"].(string); ok {
		description = desc
	}
	if ag, ok := relation.Properties["agent"].(string); ok {
		agent = ag
	}
	
	// Calculate size from relation JSON
	if relationData, err := json.Marshal(relation); err == nil {
		size = int64(len(relationData))
	}
	
	// Prepare response data
	responseData := map[string]interface{}{
		"path":      path,
		"object_id": objID,
		"type":      objectType,
		"created":   relation.CreatedAt,
		"modified":  relation.UpdatedAt,
		"size":      size,
		
		// Content info from relation
		"title":       title,
		"description": description,
		
		// Context
		"agent":       agent,
		"session":     relation.Properties["session_id"],
		"source":      relation.Properties["source"], // "declare" or "possess"
		"language":    relation.Properties["language"],
		"transforms":  relation.Properties["transforms"],
	}
	
	resp.SetData(responseData)
	return resp
}