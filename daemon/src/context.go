package main

import "time"

// ContextData is the complete context structure shared between daemon and CLI
type ContextData struct {
	ActiveSession    *ActiveSessionInfo   `json:"active_session"`
	RecentCommands   []CommandRecord      `json:"recent_commands"`
	CreatedTools     []ToolRecord         `json:"created_tools"`
	AccessedMemories []MemoryAccess       `json:"accessed_memories,omitempty"`
	Suggestions      []ContextSuggestion  `json:"suggestions"`
}

// ActiveSessionInfo represents the active session information for display
type ActiveSessionInfo struct {
	ID           string    `json:"id"`
	Agent        string    `json:"agent"`
	MessageCount int       `json:"message_count"`
	StartTime    time.Time `json:"start_time"`
	LastActivity time.Time `json:"last_activity"`
	State        string    `json:"state"`
	ToolCreated  *string   `json:"tool_created,omitempty"`
}

// CommandRecord represents a recently executed command
type CommandRecord struct {
	Command    string    `json:"command"`
	Timestamp  time.Time `json:"timestamp"`
	AgeSeconds int       `json:"age_seconds"`
	ExitCode   int       `json:"exit_code"`
}

// ToolRecord represents a tool created in the current session
type ToolRecord struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Transforms []string  `json:"transforms,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// MemoryAccess tracks accessed memory/artifact paths
type MemoryAccess struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	AccessCount int    `json:"access_count"`
	DisplayName string `json:"display_name,omitempty"` // Human-readable name
}

// ContextSuggestion provides smart command suggestions
type ContextSuggestion struct {
	Command    string  `json:"command"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}