package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ObjectMemoryStore handles persistent storage of sessions using object store
type ObjectMemoryStore struct {
	objectStore *ObjectStore
	baseDir     string
	mu          sync.RWMutex
	index       *ObjectMemoryIndex
}

// ObjectMemoryIndex tracks all sessions with their current object IDs
type ObjectMemoryIndex struct {
	Version  string                      `json:"version"`
	Sessions map[string]SessionReference `json:"sessions"` // session ID -> reference
	Stats    MemoryStats                 `json:"stats"`
}

// SessionReference points to the current object for a session
type SessionReference struct {
	ObjectID         string    `json:"object_id"`
	SessionID        string    `json:"session_id"`
	Agent            string    `json:"agent"`
	CreatedAt        time.Time `json:"created_at"`
	LastUpdated      time.Time `json:"last_updated"`
	CommandGenerated bool      `json:"command_generated"`
	State            string    `json:"state"`
	MessageCount     int       `json:"message_count"`
}

// NewObjectMemoryStore creates a new memory store backed by object store
func NewObjectMemoryStore(objectStore *ObjectStore) (*MemoryStore, error) {
	store := &ObjectMemoryStore{
		objectStore: objectStore,
		baseDir:     objectStore.baseDir,
		index: &ObjectMemoryIndex{
			Version:  "2.0",
			Sessions: make(map[string]SessionReference),
			Stats:    MemoryStats{},
		},
	}

	// Load existing index from object store if it exists
	if err := store.loadIndex(); err != nil {
		log.Printf("📝 Creating new memory index: %v", err)
	}

	// Return as MemoryStore interface for compatibility
	return &MemoryStore{
		baseDir:   store.baseDir,
		indexPath: "", // Not used in object store version
		index: &MemoryIndex{
			Sessions: store.getSessionSummaries(),
			Stats:    store.index.Stats,
		},
	}, nil
}

// SaveSession saves a session to the object store
func (m *ObjectMemoryStore) SaveSession(session *Session) error {
	log.Printf("💾 Saving session %s to object store", session.ID)

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

	// Create metadata for the session
	metadata := &Metadata{
		Type:        "session",
		Title:       fmt.Sprintf("Session %s", session.ID),
		Description: fmt.Sprintf("AI conversation with %s", session.Agent),
		Tags:        extractSessionTags(session),
		Session:     session.ID,
		Agent:       session.Agent,
		Lifecycle:   mapStateToLifecycle(session.State),
		Paths: []string{
			fmt.Sprintf("memory/sessions/%s", session.ID),
			fmt.Sprintf("memory/sessions/by-date/%s/%s", 
				session.CreatedAt.Format("2006-01-02"), session.ID),
			fmt.Sprintf("memory/sessions/by-agent/%s/%s", 
				cleanAgentName(session.Agent), session.ID),
		},
	}

	// Store in object store
	objectID, err := m.objectStore.StoreWithMetadata(data, metadata)
	if err != nil {
		return fmt.Errorf("failed to store session: %v", err)
	}

	// Update index
	m.index.Sessions[session.ID] = SessionReference{
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
	m.updateStats()

	// Save index
	if err := m.saveIndex(); err != nil {
		log.Printf("⚠️  Failed to save index: %v", err)
		// Don't fail the whole operation if index save fails
	}

	log.Printf("✅ Session %s saved to object store: %s", session.ID, objectID[:12]+"...")
	return nil
}

// LoadSession loads a session from object store
func (m *ObjectMemoryStore) LoadSession(sessionID string) (*PersistentSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ref, exists := m.index.Sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Load from object store
	data, err := m.objectStore.Read(ref.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to read session object: %v", err)
	}

	var session PersistentSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %v", err)
	}

	return &session, nil
}

// GetRecentSessions returns sessions from the last N days
func (m *ObjectMemoryStore) GetRecentSessions(days int) ([]*PersistentSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	var sessions []*PersistentSession

	for _, ref := range m.index.Sessions {
		if ref.CreatedAt.After(cutoff) {
			if session, err := m.LoadSession(ref.SessionID); err == nil {
				sessions = append(sessions, session)
			}
		}
	}

	return sessions, nil
}

// SearchSessions searches for sessions matching a query
func (m *ObjectMemoryStore) SearchSessions(query string) ([]*SessionSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []*SessionSummary

	for _, ref := range m.index.Sessions {
		if strings.Contains(strings.ToLower(ref.SessionID), query) ||
			strings.Contains(strings.ToLower(ref.Agent), query) {
			matches = append(matches, &SessionSummary{
				ID:               ref.SessionID,
				Agent:            ref.Agent,
				CommandGenerated: ref.CommandGenerated,
				State:            ref.State,
				CreatedAt:        ref.CreatedAt,
				LastActivity:     ref.LastUpdated,
				MessageCount:     ref.MessageCount,
			})
		}
	}

	return matches, nil
}

// Helper methods

func (m *ObjectMemoryStore) loadIndex() error {
	// Look for index in object store
	// The index itself is stored as an object with a well-known ID
	indexID := "memory-index-v2"
	
	data, err := m.objectStore.Read(indexID)
	if err != nil {
		return err
	}

	var index ObjectMemoryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	m.index = &index
	return nil
}

func (m *ObjectMemoryStore) saveIndex() error {
	// Serialize index
	data, err := json.MarshalIndent(m.index, "", "  ")
	if err != nil {
		return err
	}

	// Store with well-known ID (not content-addressed for index)
	// This is a special case - the index needs a stable ID
	indexPath := filepath.Join(m.objectStore.baseDir, "memory-index-v2.json")
	return os.WriteFile(indexPath, data, 0644)
}

func (m *ObjectMemoryStore) updateStats() {
	stats := &m.index.Stats
	stats.TotalSessions = len(m.index.Sessions)
	
	commandCount := 0
	activeCount := 0
	var lastTime time.Time
	
	for _, ref := range m.index.Sessions {
		if ref.CommandGenerated {
			commandCount++
		}
		if ref.State == string(SessionActive) {
			activeCount++
		}
		if ref.LastUpdated.After(lastTime) {
			lastTime = ref.LastUpdated
		}
	}
	
	stats.CommandsGenerated = commandCount
	stats.ActiveSessions = activeCount
	stats.LastSessionTime = lastTime
}

func (m *ObjectMemoryStore) getSessionSummaries() []SessionSummary {
	var summaries []SessionSummary
	
	for _, ref := range m.index.Sessions {
		summaries = append(summaries, SessionSummary{
			ID:               ref.SessionID,
			Agent:            ref.Agent,
			CommandGenerated: ref.CommandGenerated,
			State:            ref.State,
			CreatedAt:        ref.CreatedAt,
			LastActivity:     ref.LastUpdated,
			MessageCount:     ref.MessageCount,
		})
	}
	
	return summaries
}

// Utility functions

func extractSessionTags(session *Session) []string {
	tags := []string{"conversation", "ai", strings.ToLower(session.Agent)}
	
	if session.CommandGenerated != nil {
		tags = append(tags, "command-generated", session.CommandGenerated.Name)
	}
	
	// Add state as tag
	tags = append(tags, string(session.State))
	
	// Extract keywords from messages
	for _, msg := range session.Messages {
		words := strings.Fields(msg.Content)
		for _, word := range words {
			word = strings.ToLower(word)
			if len(word) > 5 && !isCommonWord(word) {
				tags = append(tags, word)
			}
		}
	}
	
	return tags
}

func mapStateToLifecycle(state SessionState) string {
	switch state {
	case SessionActive:
		return "active"
	case SessionCompleted:
		return "stable"
	case SessionAbandoned:
		return "archived"
	default:
		return "draft"
	}
}

func cleanAgentName(agent string) string {
	// Remove @ prefix and make filesystem-friendly
	agent = strings.TrimPrefix(agent, "@")
	agent = strings.ReplaceAll(agent, " ", "-")
	agent = strings.ToLower(agent)
	return agent
}