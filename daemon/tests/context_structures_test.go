package main

import (
	"encoding/json"
	"testing"
	"time"
)

// Test that ContextData serializes and deserializes correctly
func TestContextDataSerialization(t *testing.T) {
	// Create test data with all fields populated
	testData := &ContextData{
		ActiveSession: &ActiveSessionInfo{
			ID:           "test-session-123",
			Agent:        "@ai-engineer",
			MessageCount: 5,
			StartTime:    time.Now().Add(-10 * time.Minute),
			LastActivity: time.Now(),
			State:        "active",
			ToolCreated:  stringPtr("test-tool"),
		},
		RecentCommands: []CommandRecord{
			{
				Command:    "port42 search test",
				Timestamp:  time.Now().Add(-5 * time.Minute),
				AgeSeconds: 300,
				ExitCode:   0,
			},
		},
		CreatedTools: []ToolRecord{
			{
				Name:       "test-tool",
				Type:       "tool",
				Transforms: []string{"bash", "json"},
				CreatedAt:  time.Now().Add(-8 * time.Minute),
			},
		},
		AccessedMemories: []MemoryAccess{
			{
				Path:        "/memory/test-session",
				Type:        "memory",
				AccessCount: 3,
			},
		},
		Suggestions: []ContextSuggestion{
			{
				Command:    "test-tool --help",
				Reason:     "Learn about your new tool",
				Confidence: 0.95,
			},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Deserialize back
	var decoded ContextData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	// Verify core fields
	if decoded.ActiveSession == nil {
		t.Fatal("ActiveSession should not be nil")
	}
	if decoded.ActiveSession.ID != "test-session-123" {
		t.Errorf("Expected session ID 'test-session-123', got '%s'", decoded.ActiveSession.ID)
	}
	if len(decoded.RecentCommands) != 1 {
		t.Errorf("Expected 1 recent command, got %d", len(decoded.RecentCommands))
	}
	if len(decoded.CreatedTools) != 1 {
		t.Errorf("Expected 1 created tool, got %d", len(decoded.CreatedTools))
	}
	if len(decoded.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(decoded.Suggestions))
	}
	if decoded.ActiveSession.ToolCreated == nil || *decoded.ActiveSession.ToolCreated != "test-tool" {
		t.Error("ToolCreated field not preserved")
	}
}

// Test empty data structures
func TestEmptyContextData(t *testing.T) {
	testData := &ContextData{
		ActiveSession:    nil,
		RecentCommands:   []CommandRecord{},
		CreatedTools:     []ToolRecord{},
		AccessedMemories: []MemoryAccess{},
		Suggestions:      []ContextSuggestion{},
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to serialize empty data: %v", err)
	}

	var decoded ContextData
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to deserialize empty data: %v", err)
	}

	if decoded.ActiveSession != nil {
		t.Error("ActiveSession should be nil")
	}
	if len(decoded.RecentCommands) != 0 {
		t.Errorf("RecentCommands should be empty, got %d", len(decoded.RecentCommands))
	}
}

func stringPtr(s string) *string {
	return &s
}