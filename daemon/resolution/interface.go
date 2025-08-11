package resolution

// ResolutionService provides the public interface for reference resolution
type ResolutionService interface {
	// ResolveForAI resolves references and formats for AI consumption
	ResolveForAI(references []Reference) (string, error)
	
	// GetResolutionStats returns statistics about resolution process
	GetResolutionStats(references []Reference) (*Stats, error)
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
	SearchHandler func(query string, limit int) ([]SearchResult, error)
	ToolHandler   func(toolName string) (*ToolDefinition, error)
	MemoryHandler func(sessionID string) (*MemorySession, error)
	FileHandler   func(path string) (*FileContent, error)
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

// NewResolutionService creates a new resolution service with provided handlers
func NewResolutionService(handlers Handlers) ResolutionService {
	return newService(handlers)
}