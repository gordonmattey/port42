package main

import (
	"time"
)

// SessionState represents the current state of a session
type SessionState string

const (
	SessionActive     SessionState = "active"
	SessionIdle       SessionState = "idle"
	SessionCompleted  SessionState = "completed"
	SessionAbandoned  SessionState = "abandoned"
)

// Metadata represents object metadata
type Metadata struct {
	ID       string    `json:"id"`
	Paths    []string  `json:"paths"`
	Type     string    `json:"type"`
	Subtype  string    `json:"subtype,omitempty"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	Accessed time.Time `json:"accessed"`
	Session  string    `json:"session,omitempty"`
	Agent    string    `json:"agent,omitempty"`
	
	// Rich metadata
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	
	// Lifecycle
	Lifecycle   string `json:"lifecycle,omitempty"` // draft, active, stable, archived, deprecated
	Importance  string `json:"importance,omitempty"`
	UsageCount  int    `json:"usage_count"`
	Size        int64  `json:"size,omitempty"`
	
	// AI context
	Summary    string    `json:"summary,omitempty"`
	Embeddings []float32 `json:"embeddings,omitempty"`
	
	// Relationships
	Relationships struct {
		Session           string   `json:"session,omitempty"`
		ParentArtifacts   []string `json:"parent_artifacts,omitempty"`
		ChildArtifacts    []string `json:"child_artifacts,omitempty"`
		GeneratedCommands []string `json:"generated_commands,omitempty"`
		References        []string `json:"references,omitempty"`
	} `json:"relationships,omitempty"`
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
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CommandGenerationInfo stores info about a generated command
type CommandGenerationInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// MemoryStats tracks memory store statistics
type MemoryStats struct {
	TotalSessions      int       `json:"total_sessions"`
	CommandsGenerated  int       `json:"commands_generated"`
	ActiveSessions     int       `json:"active_sessions"`
	LastSessionTime    time.Time `json:"last_session_time"`
}

// SessionSummary provides a lightweight view of a session
type SessionSummary struct {
	ID           string    `json:"id"`
	Agent        string    `json:"agent"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
	MessageCount int       `json:"message_count"`
	State        string    `json:"state"`
}

// SearchFilters defines filters for searching objects
type SearchFilters struct {
	Path   string    `json:"path,omitempty"`   // Limit to paths under this prefix
	Type   string    `json:"type,omitempty"`   // Object type filter
	After  time.Time `json:"after,omitempty"`  // Created after
	Before time.Time `json:"before,omitempty"` // Created before
	Agent  string    `json:"agent,omitempty"`  // Filter by agent
	Tags   []string  `json:"tags,omitempty"`   // Must have all these tags
	Limit  int       `json:"limit,omitempty"`  // Max results (default 20)
}

// SearchResult represents a search match
type SearchResult struct {
	Path        string   `json:"path"`
	ObjectID    string   `json:"object_id"`
	Type        string   `json:"type"`
	Score       float64  `json:"score"`
	Snippet     string   `json:"snippet"`      // Context around match
	Metadata    Metadata `json:"metadata"`     // Full metadata
	MatchFields []string `json:"match_fields"` // Which fields matched
}

// SessionIndex represents the complete session storage (v2.0 format)
type SessionIndex struct {
	Sessions     map[string]SessionReference `json:"sessions"`
	LastSessions map[string]string           `json:"last_sessions"`
	Metadata     SessionIndexMetadata        `json:"metadata"`
}

// SessionIndexMetadata contains index-level metadata
type SessionIndexMetadata struct {
	Version       string    `json:"version"`
	LastUpdated   time.Time `json:"last_updated"`
	TotalSessions int       `json:"total_sessions"`
}