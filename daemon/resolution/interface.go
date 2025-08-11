package resolution

import "time"

// ResolutionService provides the public interface for reference resolution
type ResolutionService interface {
	// ResolveForAI resolves references and formats for AI consumption
	// Returns formatted string, resolved contexts, and error
	ResolveForAI(references []Reference) (string, []*ResolvedContext, error)
	
	// GetResolutionStats returns statistics about resolution process (DEPRECATED)
	GetResolutionStats(references []Reference) (*Stats, error)
	
	// ComputeStatsFromContexts computes stats from already resolved contexts
	ComputeStatsFromContexts(references []Reference, contexts []*ResolvedContext) *Stats
}

// Reference represents a reference to resolve (from protocol)
type Reference struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Context string `json:"context,omitempty"`
}

// Stats provides resolution statistics
type Stats struct {
	TotalReferences  int            `json:"total_references"`
	ResolvedCount    int            `json:"resolved_count"`
	FailedCount      int            `json:"failed_count"`
	TotalContentSize int            `json:"total_content_size"`
	TypeBreakdown    map[string]int `json:"type_breakdown"`
	SuccessRate      float64        `json:"success_rate_percent"`
}

// ResolvedContext represents resolved reference content
type ResolvedContext struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Content string `json:"content"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Handlers are the interface points where daemon provides data access
type Handlers struct {
	SearchHandler    func(query string, limit int) ([]SearchResult, error)
	ToolHandler      func(toolName string) (*ToolDefinition, error)
	FileHandler      func(path string) (*FileContent, error)
	P42Handler       func(p42Path string) (*FileContent, error) // Port 42 VFS access
	RelationsHandler func() RelationsManager // NEW: For URL artifact Relations
}

// Data types for handlers (self-contained in this package)
type SearchResult struct {
	Path       string                 `json:"path"`
	Type       string                 `json:"type"`
	Score      float64                `json:"score"`
	Title      string                 `json:"title"`
	Summary    string                 `json:"summary"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type ToolDefinition struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Transforms []string               `json:"transforms"`
	Commands   []string               `json:"commands,omitempty"`
	Properties map[string]interface{} `json:"properties"`
	Created    string                 `json:"created"`
	Agent      string                 `json:"agent,omitempty"`
}

type MemorySession struct {
	SessionID string    `json:"session_id"`
	Agent     string    `json:"agent"`
	Title     string    `json:"title,omitempty"`
	Created   string    `json:"created"`
	Updated   string    `json:"updated"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

type FileContent struct {
	Path     string                 `json:"path"`
	Content  string                 `json:"content"`
	Size     int64                  `json:"size"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RelationsManager interface defines operations needed for URL artifact Relations
type RelationsManager interface {
	DeclareRelation(relation *URLArtifactRelation) error
	GetRelationByID(id string) (*URLArtifactRelation, error)
	ListRelationsByType(relationType string) ([]*URLArtifactRelation, error)
}

// URLArtifactRelation represents a URL artifact stored as a Relation
type URLArtifactRelation struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // "URLArtifact"
	Properties map[string]interface{} `json:"properties"` // URL metadata
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	
	// Content storage
	ContentID  string `json:"content_id"`  // Object ID in storage
	Content    string `json:"-"`           // Loaded content (not serialized)
}

// ReferenceContext captures the rich context of why a URL was fetched
type ReferenceContext struct {
	RelationID     string      `json:"relation_id,omitempty"`     // Tool/entity being created
	RelationType   string      `json:"relation_type,omitempty"`   // "Tool", "Memory", etc.
	SessionID      string      `json:"session_id,omitempty"`      // Memory session
	Agent          string      `json:"agent,omitempty"`           // AI agent name
	AllReferences  []Reference `json:"all_references,omitempty"`  // Full reference batch 
	ResolvedAt     time.Time   `json:"resolved_at"`               // When resolution occurred
}

// NewResolutionService creates a new resolution service with provided handlers
func NewResolutionService(handlers Handlers) ResolutionService {
	return newService(handlers)
}