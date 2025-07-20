package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SessionState represents the lifecycle state of a session
type SessionState string

const (
	SessionActive    SessionState = "active"
	SessionIdle      SessionState = "idle"
	SessionCompleted SessionState = "completed"
	SessionAbandoned SessionState = "abandoned"
)

// MemoryStore handles persistent storage of sessions
type MemoryStore struct {
	baseDir   string
	indexPath string
	mu        sync.RWMutex
	index     *MemoryIndex
}

// MemoryIndex tracks all sessions for quick lookup
type MemoryIndex struct {
	Sessions []SessionSummary `json:"sessions"`
	Stats    MemoryStats      `json:"stats"`
}

// SessionSummary is a lightweight representation for the index
type SessionSummary struct {
	ID               string    `json:"id"`
	Date             string    `json:"date"`
	Slug             string    `json:"slug"`
	Agent            string    `json:"agent"`
	CommandGenerated bool      `json:"command_generated"`
	State            string    `json:"state"`
	File             string    `json:"file"`
	CreatedAt        time.Time `json:"created_at"`
}

// MemoryStats tracks usage statistics
type MemoryStats struct {
	TotalSessions      int    `json:"total_sessions"`
	CommandsGenerated  int    `json:"commands_generated"`
	ActiveSessions     int    `json:"active_sessions"`
	LastSessionTime    time.Time `json:"last_session_time"`
}

// PersistentSession is the full session data saved to disk
type PersistentSession struct {
	ID               string                 `json:"id"`
	Agent            string                 `json:"agent"`
	State            SessionState           `json:"state"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	LastActivity     time.Time              `json:"last_activity"`
	Messages         []Message              `json:"messages"`
	CommandGenerated *CommandGenerationInfo `json:"command_generated,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CommandGenerationInfo tracks generated commands
type CommandGenerationInfo struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Path      string    `json:"path"`
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(baseDir string) (*MemoryStore, error) {
	store := &MemoryStore{
		baseDir:   baseDir,
		indexPath: filepath.Join(baseDir, "memory", "index.json"),
	}

	// Create directory structure
	memoryDir := filepath.Join(baseDir, "memory")
	sessionsDir := filepath.Join(memoryDir, "sessions")
	
	log.Printf("🔍 Creating memory directories at: %s", sessionsDir)
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directories: %v", err)
	}
	log.Printf("✅ Memory directories created successfully")

	// Load or create index
	if err := store.loadIndex(); err != nil {
		log.Printf("Creating new memory index: %v", err)
		store.index = &MemoryIndex{
			Sessions: []SessionSummary{},
			Stats:    MemoryStats{},
		}
	}

	return store, nil
}

// SaveSession persists a session to disk
func (m *MemoryStore) SaveSession(session *Session) error {
	log.Printf("💾 Attempting to save session %s", session.ID)
	
	m.mu.Lock()
	defer m.mu.Unlock()

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
			"model": "claude-3-5-sonnet-20241022", // Could be dynamic
		},
	}

	// Check if command was generated
	if session.CommandGenerated != nil {
		ps.CommandGenerated = &CommandGenerationInfo{
			Name:      session.CommandGenerated.Name,
			CreatedAt: time.Now(),
			Path:      filepath.Join(m.baseDir, "commands", session.CommandGenerated.Name),
		}
	}

	// Generate filename
	date := ps.CreatedAt.Format("2006-01-02")
	slug := m.generateSlug(session)
	filename := fmt.Sprintf("session-%d-%s.json", ps.CreatedAt.Unix(), slug)
	
	// Create date directory
	dateDir := filepath.Join(m.baseDir, "memory", "sessions", date)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return fmt.Errorf("failed to create date directory: %v", err)
	}

	// Save session file
	filePath := filepath.Join(dateDir, filename)
	log.Printf("🔍 Saving session to file: %s", filePath)
	
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %v", err)
	}
	log.Printf("✅ Session file written successfully")

	// Update index
	m.updateIndex(session, date, slug, filepath.Join(date, filename))

	log.Printf("💾 Saved session %s to %s", session.ID, filename)
	return nil
}

// LoadSession loads a specific session from disk
func (m *MemoryStore) LoadSession(id string) (*PersistentSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find session in index
	for _, summary := range m.index.Sessions {
		if summary.ID == id {
			filePath := filepath.Join(m.baseDir, "memory", "sessions", summary.File)
			return m.loadSessionFromFile(filePath)
		}
	}

	return nil, fmt.Errorf("session not found: %s", id)
}

// LoadRecentSessions loads sessions from the last N days
func (m *MemoryStore) LoadRecentSessions(days int) ([]*PersistentSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	var sessions []*PersistentSession

	for _, summary := range m.index.Sessions {
		if summary.CreatedAt.After(cutoff) {
			filePath := filepath.Join(m.baseDir, "memory", "sessions", summary.File)
			if session, err := m.loadSessionFromFile(filePath); err == nil {
				sessions = append(sessions, session)
			}
		}
	}

	return sessions, nil
}

// SearchSessions searches for sessions matching a query
func (m *MemoryStore) SearchSessions(query string) ([]*SessionSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []*SessionSummary

	for _, summary := range m.index.Sessions {
		if strings.Contains(strings.ToLower(summary.ID), query) ||
			strings.Contains(strings.ToLower(summary.Agent), query) ||
			strings.Contains(strings.ToLower(summary.Slug), query) {
			matches = append(matches, &summary)
		}
	}

	return matches, nil
}

// GetStats returns memory statistics
func (m *MemoryStore) GetStats() *MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return &m.index.Stats
}

// Private helper methods

func (m *MemoryStore) loadIndex() error {
	data, err := ioutil.ReadFile(m.indexPath)
	if err != nil {
		return err
	}

	index := &MemoryIndex{}
	if err := json.Unmarshal(data, index); err != nil {
		return err
	}

	m.index = index
	return nil
}

func (m *MemoryStore) saveIndex() error {
	data, err := json.MarshalIndent(m.index, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(m.indexPath, data, 0644)
}

func (m *MemoryStore) updateIndex(session *Session, date, slug, file string) {
	// Update or add session summary
	found := false
	for i, summary := range m.index.Sessions {
		if summary.ID == session.ID {
			m.index.Sessions[i] = SessionSummary{
				ID:               session.ID,
				Date:             date,
				Slug:             slug,
				Agent:            session.Agent,
				CommandGenerated: session.CommandGenerated != nil,
				State:            string(session.State),
				File:             file,
				CreatedAt:        session.CreatedAt,
			}
			found = true
			break
		}
	}

	if !found {
		m.index.Sessions = append(m.index.Sessions, SessionSummary{
			ID:               session.ID,
			Date:             date,
			Slug:             slug,
			Agent:            session.Agent,
			CommandGenerated: session.CommandGenerated != nil,
			State:            string(session.State),
			File:             file,
			CreatedAt:        session.CreatedAt,
		})
	}

	// Update stats
	m.index.Stats.TotalSessions = len(m.index.Sessions)
	m.index.Stats.LastSessionTime = time.Now()
	
	// Count commands and active sessions
	commandCount := 0
	activeCount := 0
	for _, s := range m.index.Sessions {
		if s.CommandGenerated {
			commandCount++
		}
		if s.State == string(SessionActive) || s.State == string(SessionIdle) {
			activeCount++
		}
	}
	m.index.Stats.CommandsGenerated = commandCount
	m.index.Stats.ActiveSessions = activeCount

	// Save index
	if err := m.saveIndex(); err != nil {
		log.Printf("Failed to save index: %v", err)
	}
}

func (m *MemoryStore) generateSlug(session *Session) string {
	// Try to extract a meaningful slug from the session
	if session.CommandGenerated != nil && session.CommandGenerated.Name != "" {
		return session.CommandGenerated.Name
	}

	// Use first few words of first user message
	for _, msg := range session.Messages {
		if msg.Role == "user" {
			words := strings.Fields(msg.Content)
			if len(words) > 0 {
				slug := strings.Join(words[:min(3, len(words))], "-")
				slug = strings.ToLower(slug)
				// Clean up slug
				slug = strings.ReplaceAll(slug, "/", "-")
				slug = strings.ReplaceAll(slug, "\\", "-")
				slug = strings.ReplaceAll(slug, ".", "")
				return slug
			}
		}
	}

	return "conversation"
}

func (m *MemoryStore) loadSessionFromFile(filePath string) (*PersistentSession, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	session := &PersistentSession{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil, err
	}

	return session, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}