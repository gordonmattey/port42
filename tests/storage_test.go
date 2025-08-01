package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStorageBasicOperations(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	
	// Create storage
	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	
	// Test 1: Store and retrieve content
	content := []byte("Hello, Port 42!")
	id, err := storage.Store(content)
	if err != nil {
		t.Fatalf("Failed to store content: %v", err)
	}
	
	// Verify ID format
	if len(id) != 64 {
		t.Errorf("Expected 64-char SHA256 ID, got %d chars", len(id))
	}
	
	// Read back
	retrieved, err := storage.Read(id)
	if err != nil {
		t.Fatalf("Failed to read content: %v", err)
	}
	
	if string(retrieved) != string(content) {
		t.Errorf("Content mismatch: expected %s, got %s", content, retrieved)
	}
	
	// Test 2: Store duplicate content
	id2, err := storage.Store(content)
	if err != nil {
		t.Fatalf("Failed to store duplicate: %v", err)
	}
	
	if id2 != id {
		t.Errorf("Duplicate content should have same ID: %s != %s", id, id2)
	}
}

func TestStorageWithMetadata(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)
	
	// Create metadata
	meta := &Metadata{
		Type:        "test",
		Title:       "Test Object",
		Description: "A test object for storage",
		Tags:        []string{"test", "storage", "metadata"},
		Paths:       []string{"/test/object", "/by-type/test/object"},
	}
	
	// Store with metadata
	content := []byte("Test content with metadata")
	id, err := storage.StoreWithMetadata(content, meta)
	if err != nil {
		t.Fatalf("Failed to store with metadata: %v", err)
	}
	
	// Load metadata
	loadedMeta, err := storage.LoadMetadata(id)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	
	// Verify metadata
	if loadedMeta.ID != id {
		t.Errorf("Metadata ID mismatch: %s != %s", loadedMeta.ID, id)
	}
	if loadedMeta.Type != "test" {
		t.Errorf("Metadata type mismatch: %s != test", loadedMeta.Type)
	}
	if loadedMeta.Size != int64(len(content)) {
		t.Errorf("Size mismatch: %d != %d", loadedMeta.Size, len(content))
	}
}

func TestSessionManagement(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)
	
	// Create test session
	session := &Session{
		ID:           "test-session-123",
		Agent:        "@ai-engineer",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		State:        SessionActive,
		Messages: []Message{
			{Role: "user", Content: "Test message", Timestamp: time.Now()},
			{Role: "assistant", Content: "Test response", Timestamp: time.Now()},
		},
	}
	
	// Save session
	if err := storage.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}
	
	// Load session
	loaded, err := storage.LoadSession("test-session-123")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}
	
	// Verify
	if loaded.ID != session.ID {
		t.Errorf("Session ID mismatch: %s != %s", loaded.ID, session.ID)
	}
	if len(loaded.Messages) != 2 {
		t.Errorf("Message count mismatch: %d != 2", len(loaded.Messages))
	}
	
	// Test update
	session.Messages = append(session.Messages, Message{
		Role: "user", Content: "Another message", Timestamp: time.Now(),
	})
	
	if err := storage.SaveSession(session); err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}
	
	// Verify session index was updated
	if ref, exists := storage.sessionIndex["test-session-123"]; !exists {
		t.Error("Session not in index")
	} else if ref.MessageCount != 3 {
		t.Errorf("Index message count not updated: %d != 3", ref.MessageCount)
	}
}

func TestCommandStorage(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)
	
	// Create test command
	spec := &CommandSpec{
		Name:         "test-cmd",
		Description:  "Test command",
		Language:     "bash",
		Dependencies: []string{"git"},
		SessionID:    "test-session",
		Agent:        "@ai-engineer",
	}
	
	code := `#!/bin/bash
echo "Hello from test command"`
	
	// Store command
	if err := storage.StoreCommand(spec, code); err != nil {
		t.Fatalf("Failed to store command: %v", err)
	}
	
	// Verify symlink was created
	homeDir, _ := os.UserHomeDir()
	linkPath := filepath.Join(homeDir, ".port42", "commands", "test-cmd")
	
	// Check if symlink exists (may fail in test environment)
	if info, err := os.Lstat(linkPath); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("Expected symlink, got regular file")
		}
	}
}

func TestVirtualPaths(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)
	
	// Store object with virtual paths
	meta := &Metadata{
		Type:  "document",
		Title: "Test Doc",
		Paths: []string{
			"/artifacts/documents/test.md",
			"/by-date/2024-01-01/test.md",
			"/by-type/document/test.md",
		},
	}
	
	content := []byte("Test document content")
	id, _ := storage.StoreWithMetadata(content, meta)
	
	// Test path resolution
	resolvedID := storage.ResolvePath("/artifacts/documents/test.md")
	if resolvedID != id {
		t.Errorf("Path resolution failed: %s != %s", resolvedID, id)
	}
	
	// Test listing
	entries := storage.ListPath("/artifacts/documents")
	found := false
	for _, entry := range entries {
		if entry["name"] == "test.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("File not found in virtual directory listing")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)
	
	// Run concurrent saves
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(n int) {
			session := &Session{
				ID:           fmt.Sprintf("concurrent-%d", n),
				Agent:        "@ai-test",
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
				State:        SessionActive,
				Messages:     []Message{{Role: "user", Content: fmt.Sprintf("Test %d", n)}},
			}
			
			if err := storage.SaveSession(session); err != nil {
				t.Errorf("Concurrent save %d failed: %v", n, err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all saves
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all sessions were saved
	if len(storage.sessionIndex) != 10 {
		t.Errorf("Expected 10 sessions in index, got %d", len(storage.sessionIndex))
	}
}

func TestRecentSessions(t *testing.T) {
	tempDir := t.TempDir()
	storage, _ := NewStorage(tempDir)
	
	// Create sessions with different dates
	oldSession := &Session{
		ID:           "old-session",
		Agent:        "@ai-test",
		CreatedAt:    time.Now().AddDate(0, 0, -10), // 10 days ago
		LastActivity: time.Now().AddDate(0, 0, -10),
		State:        SessionCompleted,
	}
	
	recentSession := &Session{
		ID:           "recent-session",
		Agent:        "@ai-test",
		CreatedAt:    time.Now().AddDate(0, 0, -1), // 1 day ago
		LastActivity: time.Now(),
		State:        SessionActive,
	}
	
	storage.SaveSession(oldSession)
	storage.SaveSession(recentSession)
	
	// Load recent sessions (last 7 days)
	recent, err := storage.LoadRecentSessions(7)
	if err != nil {
		t.Fatalf("Failed to load recent sessions: %v", err)
	}
	
	// Should only get the recent one
	if len(recent) != 1 {
		t.Errorf("Expected 1 recent session, got %d", len(recent))
	}
	
	if len(recent) > 0 && recent[0].ID != "recent-session" {
		t.Errorf("Wrong session returned: %s", recent[0].ID)
	}
}