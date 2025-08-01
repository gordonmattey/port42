package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	fmt.Println("🐬 Testing Port 42 Object Store...")
	fmt.Println()

	// Create temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "port42-test-"+time.Now().Format("20060102-150405"))
	defer os.RemoveAll(tempDir)

	// Create object store
	store, err := NewObjectStore(tempDir)
	if err != nil {
		panic(fmt.Sprintf("Failed to create object store: %v", err))
	}

	// Run tests
	testBasicStorage(store)
	testMetadata(store)
	testDuplicateContent(store)
	testLargeContent(store)
	testStoreWithMetadata(store)
	testListObjects(store)
	testVirtualPaths(store)

	fmt.Println("\n✅ All object store tests passed!")
}

func testBasicStorage(store *ObjectStore) {
	fmt.Println("1. Testing basic storage and retrieval...")

	content := []byte("Hello, Port 42! The dolphins are here.")
	
	// Store content
	id, err := store.Store(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to store content: %v", err))
	}
	
	fmt.Printf("   Stored object with ID: %s\n", id[:12]+"...")

	// Verify it's stored in git-like structure
	path := store.GetPath(id)
	if !strings.Contains(path, filepath.Join(id[:2], id[2:4])) {
		panic("Object not stored in git-like structure")
	}
	fmt.Printf("   Stored at path: %s\n", path)

	// Read content back
	retrieved, err := store.Read(id)
	if err != nil {
		panic(fmt.Sprintf("Failed to read content: %v", err))
	}

	if string(retrieved) != string(content) {
		panic("Retrieved content doesn't match original")
	}

	fmt.Println("   ✓ Content retrieved successfully")
}

func testMetadata(store *ObjectStore) {
	fmt.Println("\n2. Testing metadata storage...")

	// Create metadata
	meta := &Metadata{
		ID:          "test-object-123",
		Paths:       []string{"commands/test-cmd", "by-date/2024-01-15/test-cmd"},
		Type:        "command",
		Title:       "Test Command",
		Description: "A test command for the object store",
		Tags:        []string{"test", "command", "example"},
		Lifecycle:   "active",
		Importance:  "high",
	}

	// Save metadata
	if err := store.SaveMetadata(meta); err != nil {
		panic(fmt.Sprintf("Failed to save metadata: %v", err))
	}

	fmt.Println("   ✓ Metadata saved")

	// Load metadata
	loaded, err := store.LoadMetadata("test-object-123")
	if err != nil {
		panic(fmt.Sprintf("Failed to load metadata: %v", err))
	}

	// Verify fields
	if loaded.Title != meta.Title {
		panic("Loaded metadata doesn't match")
	}
	if len(loaded.Tags) != len(meta.Tags) {
		panic("Tags don't match")
	}
	if loaded.Created.IsZero() {
		panic("Created timestamp not set")
	}
	if loaded.Accessed.Before(loaded.Created) {
		panic("Access time is before creation time")
	}

	fmt.Println("   ✓ Metadata loaded and verified")
}

func testDuplicateContent(store *ObjectStore) {
	fmt.Println("\n3. Testing duplicate content handling...")

	content := []byte("Duplicate content test")
	
	// Store once
	id1, err := store.Store(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to store content: %v", err))
	}

	// Store same content again
	id2, err := store.Store(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to store duplicate content: %v", err))
	}

	// Should get same ID (content-addressed)
	if id1 != id2 {
		panic("Different IDs for same content")
	}

	fmt.Printf("   ✓ Same content produces same ID: %s\n", id1[:12]+"...")
}

func testLargeContent(store *ObjectStore) {
	fmt.Println("\n4. Testing large content...")

	// Create 1MB of content
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	// Store large content
	id, err := store.Store(largeContent)
	if err != nil {
		panic(fmt.Sprintf("Failed to store large content: %v", err))
	}

	// Read it back
	retrieved, err := store.Read(id)
	if err != nil {
		panic(fmt.Sprintf("Failed to read large content: %v", err))
	}

	if len(retrieved) != len(largeContent) {
		panic("Large content size mismatch")
	}

	fmt.Printf("   ✓ Stored and retrieved %d bytes\n", len(largeContent))
}

func testStoreWithMetadata(store *ObjectStore) {
	fmt.Println("\n5. Testing store with metadata...")

	content := []byte("#!/bin/bash\necho 'Hello from Port 42!'")
	meta := &Metadata{
		Paths:       []string{"commands/hello-p42", "by-date/2024-01-15/hello-p42"},
		Type:        "command",
		Title:       "Hello Port 42",
		Description: "A simple greeting command",
		Tags:        []string{"greeting", "bash", "example"},
		Session:     "sess-test-123",
		Agent:       "@ai-engineer",
	}

	// Store with metadata
	id, err := store.StoreWithMetadata(content, meta)
	if err != nil {
		panic(fmt.Sprintf("Failed to store with metadata: %v", err))
	}

	fmt.Printf("   Stored object: %s\n", id[:12]+"...")

	// Verify metadata was saved
	loaded, err := store.LoadMetadata(id)
	if err != nil {
		panic(fmt.Sprintf("Failed to load metadata: %v", err))
	}

	if loaded.ID != id {
		panic("Metadata ID doesn't match object ID")
	}
	if loaded.Session != "sess-test-123" {
		panic("Session not preserved")
	}

	fmt.Println("   ✓ Content and metadata stored together")
}

func testListObjects(store *ObjectStore) {
	fmt.Println("\n6. Testing list objects...")

	// Store a few objects
	contents := []string{
		"First object",
		"Second object", 
		"Third object",
	}

	var ids []string
	for _, content := range contents {
		id, err := store.Store([]byte(content))
		if err != nil {
			panic(fmt.Sprintf("Failed to store content: %v", err))
		}
		ids = append(ids, id)
	}

	// List all objects
	allIDs, err := store.List()
	if err != nil {
		panic(fmt.Sprintf("Failed to list objects: %v", err))
	}

	// Verify our objects are in the list
	found := 0
	for _, id := range ids {
		for _, listedID := range allIDs {
			if id == listedID {
				found++
				break
			}
		}
	}

	if found != len(ids) {
		panic("Not all objects found in listing")
	}

	fmt.Printf("   ✓ Listed %d objects\n", len(allIDs))
}

func testVirtualPaths(store *ObjectStore) {
	fmt.Println("\n7. Testing virtual path concepts...")

	// Simulate storing the same file accessible from multiple paths
	content := []byte("# Architecture Document\n\nThis appears in multiple places.")
	
	// Store content
	id, err := store.Store(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to store content: %v", err))
	}

	// Create metadata with multiple virtual paths
	meta := &Metadata{
		ID: id,
		Paths: []string{
			"projects/realtime-sync/architecture.md",
			"by-date/2024-01-15/architecture.md",
			"by-session/sess-abc123/architecture.md",
			"by-tag/websocket/architecture.md",
		},
		Type:        "artifact",
		Subtype:     "document",
		Title:       "Real-time Sync Architecture",
		Description: "System design for websocket-based synchronization",
		Tags:        []string{"architecture", "websocket", "realtime", "sync"},
		Session:     "sess-abc123",
		Agent:       "@ai-engineer",
		Lifecycle:   "active",
		Importance:  "high",
	}

	// Save metadata
	if err := store.SaveMetadata(meta); err != nil {
		panic(fmt.Sprintf("Failed to save metadata: %v", err))
	}

	// Load and verify
	loaded, err := store.LoadMetadata(id)
	if err != nil {
		panic(fmt.Sprintf("Failed to load metadata: %v", err))
	}

	fmt.Printf("   Object %s is accessible from:\n", id[:12]+"...")
	for _, path := range loaded.Paths {
		fmt.Printf("   - %s\n", path)
	}

	if len(loaded.Paths) != 4 {
		panic("Virtual paths not preserved")
	}

	fmt.Println("   ✓ Virtual paths concept working")
}