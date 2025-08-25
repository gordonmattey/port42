package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestObjectStore(t *testing.T) {
	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), "port42-test-"+time.Now().Format("20060102-150405"))
	defer os.RemoveAll(tempDir)

	// Create object store
	store, err := NewObjectStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
	}

	t.Run("BasicStorage", func(t *testing.T) {
		content := []byte("Hello, Port 42!")
		
		// Store content
		id, err := store.Store(content)
		if err != nil {
			t.Fatalf("Failed to store content: %v", err)
		}
		
		// Read it back
		retrieved, err := store.Read(id)
		if err != nil {
			t.Fatalf("Failed to read content: %v", err)
		}
		
		if !bytes.Equal(retrieved, content) {
			t.Error("Retrieved content doesn't match original")
		}
	})

	t.Run("Metadata", func(t *testing.T) {
		meta := &Metadata{
			ID:          "test-123",
			Type:        "command",
			Title:       "Test Command",
			Tags:        []string{"test", "command"},
			Lifecycle:   "active",
		}
		
		// Save metadata
		err := store.SaveMetadata(meta)
		if err != nil {
			t.Fatalf("Failed to save metadata: %v", err)
		}
		
		// Load it back
		loaded, err := store.LoadMetadata("test-123")
		if err != nil {
			t.Fatalf("Failed to load metadata: %v", err)
		}
		
		if loaded.Title != meta.Title {
			t.Error("Loaded title doesn't match")
		}
		
		if len(loaded.Tags) != len(meta.Tags) {
			t.Error("Tags count doesn't match")
		}
	})

	t.Run("ContentAddressing", func(t *testing.T) {
		content := []byte("Same content")
		
		// Store twice
		id1, err := store.Store(content)
		if err != nil {
			t.Fatalf("Failed to store content: %v", err)
		}
		
		id2, err := store.Store(content)
		if err != nil {
			t.Fatalf("Failed to store content again: %v", err)
		}
		
		// Should get same ID
		if id1 != id2 {
			t.Error("Same content should produce same ID")
		}
	})

	t.Run("StoreWithMetadata", func(t *testing.T) {
		content := []byte("#!/bin/bash\necho 'Hello'")
		meta := &Metadata{
			Type:  "command",
			Title: "Hello Command",
			Paths: []string{"commands/hello", "by-date/2024-01-15/hello"},
			Tags:  []string{"bash", "greeting"},
		}
		
		// Store with metadata
		id, err := store.StoreWithMetadata(content, meta)
		if err != nil {
			t.Fatalf("Failed to store with metadata: %v", err)
		}
		
		// Verify metadata was saved with correct ID
		loaded, err := store.LoadMetadata(id)
		if err != nil {
			t.Fatalf("Failed to load metadata: %v", err)
		}
		
		if loaded.ID != id {
			t.Error("Metadata ID doesn't match object ID")
		}
		
		if loaded.Title != meta.Title {
			t.Error("Title not preserved")
		}
	})
}