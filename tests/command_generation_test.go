package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommandGenerationWithObjectStore(t *testing.T) {
	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), "port42-test-cmd-"+time.Now().Format("20060102-150405"))
	defer os.RemoveAll(tempDir)

	// Create object store
	objectStore, err := NewObjectStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create object store: %v", err)
	}

	// Create test daemon with object store
	daemon := &Daemon{
		objectStore: objectStore,
		config: Config{
			CommandsPath: filepath.Join(tempDir, "commands"),
		},
	}

	// Test cases
	tests := []struct {
		name string
		spec *CommandSpec
		want struct {
			hasContent  bool
			hasPaths    bool
			pathCount   int
			hasSession  bool
		}
	}{
		{
			name: "Basic bash command",
			spec: &CommandSpec{
				Name:           "test-hello",
				Description:    "A simple hello world command",
				Implementation: "echo 'Hello, Port 42!'",
				Language:       "bash",
				SessionID:      "sess-test-123",
				Agent:          "@test-agent",
			},
			want: struct {
				hasContent  bool
				hasPaths    bool
				pathCount   int
				hasSession  bool
			}{
				hasContent: true,
				hasPaths:   true,
				pathCount:  4, // commands/, by-date/, by-type/, memory/sessions/
				hasSession: true,
			},
		},
		{
			name: "Python command with dependencies",
			spec: &CommandSpec{
				Name:           "data-analyzer",
				Description:    "Analyze data with pandas",
				Implementation: "import pandas as pd\nprint('Analyzing data...')",
				Language:       "python",
				Dependencies:   []string{"pandas", "numpy"},
				SessionID:      "sess-test-456",
				Agent:          "@ai-analyst",
			},
			want: struct {
				hasContent  bool
				hasPaths    bool
				pathCount   int
				hasSession  bool
			}{
				hasContent: true,
				hasPaths:   true,
				pathCount:  4,
				hasSession: true,
			},
		},
		{
			name: "Command without session",
			spec: &CommandSpec{
				Name:           "standalone-cmd",
				Description:    "A command without session context",
				Implementation: "echo 'Standalone'",
				Language:       "bash",
			},
			want: struct {
				hasContent  bool
				hasPaths    bool
				pathCount   int
				hasSession  bool
			}{
				hasContent: true,
				hasPaths:   true,
				pathCount:  3, // No session path
				hasSession: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate command
			err := daemon.generateCommand(tt.spec)
			if err != nil {
				t.Fatalf("Failed to generate command: %v", err)
			}

			// List all objects in store
			objects, err := objectStore.List()
			if err != nil {
				t.Fatalf("Failed to list objects: %v", err)
			}

			if len(objects) == 0 {
				t.Fatal("No objects stored")
			}

			// Find the object for this command by checking metadata
			var objectID string
			for _, id := range objects {
				meta, err := objectStore.LoadMetadata(id)
				if err != nil {
					continue
				}
				if meta.Title == tt.spec.Name {
					objectID = id
					break
				}
			}

			if objectID == "" {
				t.Fatalf("Could not find object for command %s", tt.spec.Name)
			}

			// Load metadata
			metadata, err := objectStore.LoadMetadata(objectID)
			if err != nil {
				t.Fatalf("Failed to load metadata: %v", err)
			}

			// Verify metadata
			if metadata.Type != "command" {
				t.Errorf("Expected type 'command', got %s", metadata.Type)
			}

			if metadata.Title != tt.spec.Name {
				t.Errorf("Expected title %s, got %s", tt.spec.Name, metadata.Title)
			}

			if metadata.Description != tt.spec.Description {
				t.Errorf("Expected description %s, got %s", tt.spec.Description, metadata.Description)
			}

			// Check paths
			if len(metadata.Paths) != tt.want.pathCount {
				t.Errorf("Expected %d paths, got %d: %v", tt.want.pathCount, len(metadata.Paths), metadata.Paths)
			}

			// Check session path if expected
			if tt.want.hasSession {
				hasSessionPath := false
				for _, path := range metadata.Paths {
					if strings.Contains(path, "memory/sessions/") {
						hasSessionPath = true
						break
					}
				}
				if !hasSessionPath {
					t.Error("Expected session path but not found")
				}
			}

			// Verify content is stored
			content, err := objectStore.Read(objectID)
			if err != nil {
				t.Fatalf("Failed to read object content: %v", err)
			}

			// Check content has shebang
			contentStr := string(content)
			switch tt.spec.Language {
			case "python":
				if !strings.HasPrefix(contentStr, "#!/usr/bin/env python3") {
					t.Error("Python command missing shebang")
				}
			case "bash":
				if !strings.HasPrefix(contentStr, "#!/bin/bash") {
					t.Error("Bash command missing shebang")
				}
			}

			// Check content contains implementation
			if !strings.Contains(contentStr, tt.spec.Implementation) {
				t.Error("Command content doesn't contain implementation")
			}

			// Verify tags
			if len(metadata.Tags) == 0 {
				t.Error("No tags generated")
			}

			// Check language tag
			hasLangTag := false
			for _, tag := range metadata.Tags {
				if tag == tt.spec.Language {
					hasLangTag = true
					break
				}
			}
			if !hasLangTag {
				t.Errorf("Missing language tag: %s", tt.spec.Language)
			}

			// Verify lifecycle
			if metadata.Lifecycle != "active" {
				t.Errorf("Expected lifecycle 'active', got %s", metadata.Lifecycle)
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		spec     *CommandSpec
		wantTags []string
		skipTags []string
	}{
		{
			name: "Bash command tags",
			spec: &CommandSpec{
				Name:        "git-helper",
				Description: "Helper tool for git operations",
				Language:    "bash",
			},
			wantTags: []string{"bash", "script", "shell", "helper", "tool", "operations"},
			skipTags: []string{"for"},
		},
		{
			name: "Python with dependencies",
			spec: &CommandSpec{
				Name:         "data-processor",
				Description:  "Process CSV files with pandas",
				Language:     "python",
				Dependencies: []string{"pandas", "numpy"},
			},
			wantTags: []string{"python", "python3", "pandas", "numpy", "process", "files"},
			skipTags: []string{"with"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := extractTags(tt.spec)

			// Check wanted tags
			for _, want := range tt.wantTags {
				found := false
				for _, tag := range tags {
					if tag == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag %s not found in %v", want, tags)
				}
			}

			// Check unwanted tags
			for _, skip := range tt.skipTags {
				for _, tag := range tags {
					if tag == skip {
						t.Errorf("Common word %s should not be a tag", skip)
					}
				}
			}
		})
	}
}