package main

import (
	"encoding/json"
	"fmt"
)

// Request represents an incoming request from the CLI
type Request struct {
	Type           string          `json:"type"`
	ID             string          `json:"id"`
	Payload        json.RawMessage `json:"payload"`
	SessionContext *SessionContext `json:"session_context,omitempty"` // Optional session info
	References     []Reference     `json:"references,omitempty"`      // Universal references
	UserPrompt     string          `json:"user_prompt,omitempty"`     // Universal user prompt
}

// SessionContext provides memory session information for relation tracking
type SessionContext struct {
	SessionID string `json:"session_id,omitempty"` // Memory session ID
	Agent     string `json:"agent,omitempty"`      // AI agent name if from conversation
}

// Reference represents a contextual reference to enhance tool generation
type Reference struct {
	Type    string `json:"type"`              // "search", "tool", "file", "p42", "url"
	Target  string `json:"target"`            // The thing being referenced
	Context string `json:"context,omitempty"` // Optional additional context
}

// Response represents the daemon's response
type Response struct {
	ID      string          `json:"id"`
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// Request types
const (
	RequestSwim = "swim"
	RequestList    = "list"
	RequestStatus  = "status"
	RequestMemory  = "memory"
	RequestWatch   = "watch"
	RequestEnd     = "end"
)

// SwimPayload for swim requests
type SwimPayload struct {
	Agent         string   `json:"agent"`
	Message       string   `json:"message"`
	SessionID     string   `json:"session_id,omitempty"`
	MemoryContext []string `json:"memory_context,omitempty"`
}

// StatusData for status responses
type StatusData struct {
	Status    string `json:"status"`
	Port      string `json:"port"`
	Sessions  int    `json:"sessions"`
	Uptime    string `json:"uptime"`
	Dolphins  string `json:"dolphins"`
	RuleCount int    `json:"rule_count,omitempty"`
	Rules     string `json:"rules,omitempty"`
}

// WatchPayload for watch requests
type WatchPayload struct {
	Target string `json:"target"` // "rules", "sessions", etc.
}

// WatchData for watch responses - streams rule activity
type WatchData struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`    // "rule_triggered", "rule_completed", "rule_failed"
	RuleID    string `json:"rule_id"`
	RuleName  string `json:"rule_name"`
	Details   string `json:"details,omitempty"`
}

// ListData for list responses
type ListData struct {
	Commands []string `json:"commands"`
}

// Helper functions
func NewResponse(id string, success bool) Response {
	return Response{
		ID:      id,
		Success: success,
	}
}

func (r *Response) SetData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.Data = jsonData
	return nil
}

func (r *Response) SetError(err string) {
	r.Success = false
	r.Error = err
}

// NewErrorResponse creates an error response
func NewErrorResponse(id string, errorMsg string) Response {
	resp := NewResponse(id, false)
	resp.SetError(errorMsg)
	return resp
}

// ValidateReference validates a single reference
func ValidateReference(ref Reference) error {
	validTypes := map[string]bool{
		"search": true,
		"tool":   true,
		"file":   true,
		"p42":    true,
		"url":    true,
	}
	
	if !validTypes[ref.Type] {
		return fmt.Errorf("invalid reference type: %s", ref.Type)
	}
	
	if ref.Target == "" {
		return fmt.Errorf("reference target cannot be empty")
	}
	
	return nil
}

// ValidateReferences validates an array of references
func ValidateReferences(refs []Reference) error {
	for i, ref := range refs {
		if err := ValidateReference(ref); err != nil {
			return fmt.Errorf("reference %d: %w", i, err)
		}
	}
	return nil
}