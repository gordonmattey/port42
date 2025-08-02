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
	"strings"
	"sync"
	"time"
)

// Storage provides unified storage for all Port 42 data
type Storage struct {
	baseDir     string
	objectsDir  string
	metadataDir string
	
	// Session index for quick lookups
	sessionIndex map[string]SessionReference
	indexMutex   sync.RWMutex
	
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
func NewStorage(baseDir string) (*Storage, error) {
	objectsDir := filepath.Join(baseDir, "objects")
	metadataDir := filepath.Join(baseDir, "metadata")
	
	// Create directories
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create objects directory: %w", err)
	}
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}
	
	s := &Storage{
		baseDir:      baseDir,
		objectsDir:   objectsDir,
		metadataDir:  metadataDir,
		sessionIndex: make(map[string]SessionReference),
		stats:        StorageStats{LastUpdated: time.Now()},
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
			"model": "claude-3-5-sonnet-20241022",
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
		paths = append(paths, fmt.Sprintf("/memory/%s/crystallized/%s", meta.Session, filepath.Base(subpath)))
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