package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ObjectStore provides content-addressed storage for all Port 42 artifacts
type ObjectStore struct {
	baseDir     string
	metadataDir string
}

// NewObjectStore creates a new object store
func NewObjectStore(baseDir string) (*ObjectStore, error) {
	objectsDir := filepath.Join(baseDir, "objects")
	metadataDir := filepath.Join(baseDir, "metadata")

	// Create directories
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create objects directory: %w", err)
	}
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %w", err)
	}

	return &ObjectStore{
		baseDir:     objectsDir,
		metadataDir: metadataDir,
	}, nil
}

// Store saves content and returns its hash ID
func (o *ObjectStore) Store(content []byte) (string, error) {
	// Calculate SHA256 hash
	hash := sha256.Sum256(content)
	id := hex.EncodeToString(hash[:])

	// Store in git-like structure: objects/3a/4f/2b8c9d...
	dir := filepath.Join(o.baseDir, id[:2], id[2:4])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create object directory: %w", err)
	}

	path := filepath.Join(dir, id[4:])
	
	// Check if object already exists
	if _, err := os.Stat(path); err == nil {
		return id, nil // Object already stored
	}

	// Write content
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write object: %w", err)
	}

	return id, nil
}

// Read retrieves content by hash ID
func (o *ObjectStore) Read(id string) ([]byte, error) {
	if len(id) < 4 {
		return nil, fmt.Errorf("invalid object ID: %s", id)
	}

	path := filepath.Join(o.baseDir, id[:2], id[2:4], id[4:])
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return content, nil
}

// Exists checks if an object exists
func (o *ObjectStore) Exists(id string) bool {
	if len(id) < 4 {
		return false
	}

	path := filepath.Join(o.baseDir, id[:2], id[2:4], id[4:])
	_, err := os.Stat(path)
	return err == nil
}

// GetPath returns the filesystem path for an object
func (o *ObjectStore) GetPath(id string) string {
	if len(id) < 4 {
		return ""
	}
	return filepath.Join(o.baseDir, id[:2], id[2:4], id[4:])
}

// Metadata represents object metadata
type Metadata struct {
	ID       string    `json:"id"`
	Paths    []string  `json:"paths"`
	Type     string    `json:"type"`
	Subtype  string    `json:"subtype,omitempty"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	Accessed time.Time `json:"accessed"`
	Session  string    `json:"session,omitempty"`
	Agent    string    `json:"agent,omitempty"`
	
	// Rich metadata
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	
	// Lifecycle
	Lifecycle   string `json:"lifecycle,omitempty"` // draft, active, stable, archived, deprecated
	Importance  string `json:"importance,omitempty"`
	UsageCount  int    `json:"usage_count"`
	
	// AI context
	Summary    string    `json:"summary,omitempty"`
	Embeddings []float32 `json:"embeddings,omitempty"`
	
	// Relationships
	Relationships struct {
		Session           string   `json:"session,omitempty"`
		ParentArtifacts   []string `json:"parent_artifacts,omitempty"`
		ChildArtifacts    []string `json:"child_artifacts,omitempty"`
		GeneratedCommands []string `json:"generated_commands,omitempty"`
		References        []string `json:"references,omitempty"`
	} `json:"relationships,omitempty"`
}

// SaveMetadata stores metadata for an object
func (o *ObjectStore) SaveMetadata(meta *Metadata) error {
	if meta.ID == "" {
		return fmt.Errorf("metadata ID cannot be empty")
	}

	// Update timestamps
	if meta.Created.IsZero() {
		meta.Created = time.Now()
	}
	meta.Modified = time.Now()
	meta.Accessed = time.Now()

	// Set default lifecycle if not specified
	if meta.Lifecycle == "" {
		meta.Lifecycle = "draft"
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Store in metadata directory
	metaPath := filepath.Join(o.metadataDir, meta.ID+".json")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// LoadMetadata retrieves metadata for an object
func (o *ObjectStore) LoadMetadata(id string) (*Metadata, error) {
	metaPath := filepath.Join(o.metadataDir, id+".json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("metadata not found for object: %s", id)
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Update access time
	meta.Accessed = time.Now()
	o.SaveMetadata(&meta) // Update the access time

	return &meta, nil
}

// StoreWithMetadata stores content with associated metadata
func (o *ObjectStore) StoreWithMetadata(content []byte, meta *Metadata) (string, error) {
	// Store content
	id, err := o.Store(content)
	if err != nil {
		return "", err
	}

	// Update metadata with ID
	meta.ID = id
	
	// Save metadata
	if err := o.SaveMetadata(meta); err != nil {
		// TODO: Consider cleanup of stored object on metadata failure
		return "", fmt.Errorf("failed to save metadata: %w", err)
	}

	return id, nil
}

// List returns all object IDs in the store
func (o *ObjectStore) List() ([]string, error) {
	var ids []string
	
	// Walk through the object store directory structure
	err := filepath.Walk(o.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Reconstruct ID from path
		rel, err := filepath.Rel(o.baseDir, path)
		if err != nil {
			return err
		}
		
		// Convert path back to ID: 3a/4f/2b8c9d... -> 3a4f2b8c9d...
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) == 3 {
			id := parts[0] + parts[1] + parts[2]
			ids = append(ids, id)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	
	return ids, nil
}

// Copy copies a reader to the object store
func (o *ObjectStore) CopyFrom(r io.Reader) (string, error) {
	// Read all content
	content, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}
	
	return o.Store(content)
}