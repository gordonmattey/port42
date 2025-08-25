package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMemoryStoreWithObjectStore(t *testing.T) {
	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), "port42-test-memory-"+time.Now().Format("20060102-150405"))
	defer os.RemoveAll(tempDir)

	// Create object store
	objectStore, err := NewObjectStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
	}

	// Create memory store
	memoryStore, err := NewMemoryStore(tempDir, objectStore)
	if err != nil {
		t.Fatalf("Failed to create memory store: %v", err)
	}

	// Test 1: Save and load basic session
	t.Run("BasicSessionSaveLoad", func(t *testing.T) {
		session := &Session{
			ID:           "test-basic-session",
			Agent:        "@test-agent",
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
			State:        SessionActive,
			Messages: []Message{
				{Role: "user", Content: "Test message", Timestamp: time.Now()},
				{Role: "assistant", Content: "Test response", Timestamp: time.Now()},
			},
		}

		// Save session
		if err := memoryStore.SaveSession(session); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Load session
		loaded, err := memoryStore.LoadSession("test-basic-session")
		if err != nil {
			t.Fatalf("Failed to load session: %v", err)
		}

		// Verify
		if loaded.ID != session.ID {
			t.Errorf("ID mismatch: expected %s, got %s", session.ID, loaded.ID)
		}
		if loaded.Agent != session.Agent {
			t.Errorf("Agent mismatch: expected %s, got %s", session.Agent, loaded.Agent)
		}
		if len(loaded.Messages) != len(session.Messages) {
			t.Errorf("Message count mismatch: expected %d, got %d", 
				len(session.Messages), len(loaded.Messages))
		}
	})

	// Test 2: Session with command generation
	t.Run("SessionWithCommand", func(t *testing.T) {
		session := &Session{
			ID:           "test-cmd-session",
			Agent:        "@ai-builder",
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
			State:        SessionCompleted,
			Messages: []Message{
				{Role: "user", Content: "Create a test command", Timestamp: time.Now()},
			},
			CommandGenerated: &CommandSpec{
				Name:        "test-cmd",
				Description: "Test command",
				Language:    "bash",
			},
		}

		// Save
		if err := memoryStore.SaveSession(session); err != nil {
			t.Fatalf("Failed to save session with command: %v", err)
		}

		// Verify object store has proper metadata
		objects, err := objectStore.List()
		if err != nil {
			t.Fatalf("Failed to list objects: %v", err)
		}

		found := false
		for _, objID := range objects {
			meta, err := objectStore.LoadMetadata(objID)
			if err != nil {
				continue
			}
			if meta.Session == "test-cmd-session" {
				found = true
				
				// Check metadata
				if meta.Type != "session" {
					t.Errorf("Expected type 'session', got %s", meta.Type)
				}
				
				// Check for command-generated tag
				hasTag := false
				for _, tag := range meta.Tags {
					if tag == "command-generated" {
						hasTag = true
						break
					}
				}
				if !hasTag {
					t.Error("Missing 'command-generated' tag")
				}
				
				// Check lifecycle
				if meta.Lifecycle != "stable" {
					t.Errorf("Expected lifecycle 'stable' for completed session, got %s", 
						meta.Lifecycle)
				}
				
				break
			}
		}
		
		if !found {
			t.Error("Session object not found in object store")
		}
	})

	// Test 3: Virtual paths
	t.Run("VirtualPaths", func(t *testing.T) {
		session := &Session{
			ID:           "test-paths-session",
			Agent:        "@ai-analyst",
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
			State:        SessionActive,
			Messages:     []Message{{Role: "user", Content: "Test", Timestamp: time.Now()}},
		}

		if err := memoryStore.SaveSession(session); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Check virtual paths in metadata
		objects, _ := objectStore.List()
		for _, objID := range objects {
			meta, err := objectStore.LoadMetadata(objID)
			if err != nil || meta.Session != "test-paths-session" {
				continue
			}
			
			// Should have at least 3 paths
			if len(meta.Paths) < 3 {
				t.Errorf("Expected at least 3 virtual paths, got %d", len(meta.Paths))
			}
			
			// Check path patterns
			hasMemoryPath := false
			hasDatePath := false
			hasAgentPath := false
			
			for _, path := range meta.Paths {
				if strings.HasPrefix(path, "memory/sessions/") {
					hasMemoryPath = true
				}
				if strings.Contains(path, "by-date/") {
					hasDatePath = true
				}
				if strings.Contains(path, "by-agent/") {
					hasAgentPath = true
				}
			}
			
			if !hasMemoryPath {
				t.Error("Missing memory/sessions path")
			}
			if !hasDatePath {
				t.Error("Missing by-date path")
			}
			if !hasAgentPath {
				t.Error("Missing by-agent path")
			}
			
			break
		}
	})

	// Test 4: Recent sessions
	t.Run("RecentSessions", func(t *testing.T) {
		// Create sessions with different timestamps
		sessions := []*Session{
			{
				ID:           "recent-1",
				Agent:        "@test",
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
				State:        SessionActive,
				Messages:     []Message{{Role: "user", Content: "Recent 1", Timestamp: time.Now()}},
			},
			{
				ID:           "old-session",
				Agent:        "@test",
				CreatedAt:    time.Now().AddDate(0, 0, -10), // 10 days ago
				LastActivity: time.Now().AddDate(0, 0, -10),
				State:        SessionCompleted,
				Messages:     []Message{{Role: "user", Content: "Old", Timestamp: time.Now().AddDate(0, 0, -10)}},
			},
		}

		for _, s := range sessions {
			if err := memoryStore.SaveSession(s); err != nil {
				t.Fatalf("Failed to save session %s: %v", s.ID, err)
			}
		}

		// Load recent (last 7 days)
		recent, err := memoryStore.LoadRecentSessions(7)
		if err != nil {
			t.Fatalf("Failed to load recent sessions: %v", err)
		}

		// Should include recent-1 but not old-session
		foundRecent := false
		foundOld := false
		
		for _, s := range recent {
			if s.ID == "recent-1" {
				foundRecent = true
			}
			if s.ID == "old-session" {
				foundOld = true
			}
		}
		
		if !foundRecent {
			t.Error("Recent session not found in LoadRecentSessions")
		}
		if foundOld {
			t.Error("Old session should not be in recent sessions")
		}
	})

	// Test 5: Memory stats
	t.Run("MemoryStats", func(t *testing.T) {
		stats := memoryStore.GetStats()
		
		if stats.TotalSessions == 0 {
			t.Error("Expected non-zero total sessions")
		}
		
		// Count command sessions
		expectedCmdCount := 0
		for _, summary := range memoryStore.index.Sessions {
			if summary.CommandGenerated {
				expectedCmdCount++
			}
		}
		
		if stats.CommandsGenerated != expectedCmdCount {
			t.Errorf("Command count mismatch: expected %d, got %d", 
				expectedCmdCount, stats.CommandsGenerated)
		}
	})

	// Test 6: Tag extraction
	t.Run("TagExtraction", func(t *testing.T) {
		session := &Session{
			ID:           "test-tags",
			Agent:        "@ai-analyst",
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
			State:        SessionActive,
			Messages: []Message{
				{Role: "user", Content: "Analyze performance metrics dashboard", Timestamp: time.Now()},
			},
		}

		tags := memoryStore.extractSessionTags(session)
		
		// Should have basic tags
		hasConversation := false
		hasAI := false
		hasAgent := false
		
		for _, tag := range tags {
			if tag == "conversation" {
				hasConversation = true
			}
			if tag == "ai" {
				hasAI = true
			}
			if tag == "@ai-analyst" {
				hasAgent = true
			}
		}
		
		if !hasConversation || !hasAI || !hasAgent {
			t.Errorf("Missing basic tags. Tags: %v", tags)
		}
		
		// Should extract keywords
		hasKeyword := false
		for _, tag := range tags {
			if tag == "analyze" || tag == "performance" || tag == "metrics" || tag == "dashboard" {
				hasKeyword = true
				break
			}
		}
		if !hasKeyword {
			t.Errorf("No keywords extracted from message. Tags: %v", tags)
		}
	})
}

func TestCleanAgentName(t *testing.T) {
	tempDir := os.TempDir()
	objectStore, _ := NewObjectStore(tempDir)
	memoryStore, _ := NewMemoryStore(tempDir, objectStore)
	
	tests := []struct {
		input    string
		expected string
	}{
		{"@ai-analyst", "ai-analyst"},
		{"@AI Builder", "ai-builder"},
		{"test/agent", "test-agent"},
		{"Test Agent", "test-agent"},
	}

	for _, test := range tests {
		result := memoryStore.cleanAgentName(test.input)
		if result != test.expected {
			t.Errorf("cleanAgentName(%s) = %s, expected %s", 
				test.input, result, test.expected)
		}
	}
}

func TestMapStateToLifecycle(t *testing.T) {
	tempDir := os.TempDir()
	objectStore, _ := NewObjectStore(tempDir)
	memoryStore, _ := NewMemoryStore(tempDir, objectStore)
	
	tests := []struct {
		state     SessionState
		lifecycle string
	}{
		{SessionActive, "active"},
		{SessionCompleted, "stable"},
		{SessionAbandoned, "archived"},
		{SessionIdle, "draft"},
	}

	for _, test := range tests {
		result := memoryStore.mapStateToLifecycle(test.state)
		if result != test.lifecycle {
			t.Errorf("mapStateToLifecycle(%s) = %s, expected %s", 
				test.state, result, test.lifecycle)
		}
	}
}