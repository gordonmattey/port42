package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FileRelationStore implements RelationStore using JSON files
type FileRelationStore struct {
	baseDir string // ~/.port42/relations/
}

// NewFileRelationStore creates a new file-based relation store
func NewFileRelationStore(baseDir string) (*FileRelationStore, error) {
	relationsDir := filepath.Join(baseDir, "relations")
	
	// Ensure relations directory exists
	if err := os.MkdirAll(relationsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create relations directory: %w", err)
	}
	
	return &FileRelationStore{
		baseDir: relationsDir,
	}, nil
}

// Save stores a relation as a JSON file
func (store *FileRelationStore) Save(relation Relation) error {
	filename := fmt.Sprintf("relation-%s.json", relation.ID)
	filePath := filepath.Join(store.baseDir, filename)
	
	// DEBUG: Log what we're saving for URLArtifacts
	if relation.Type == "URLArtifact" {
		fetchedAt, _ := relation.Properties["fetched_at"].(float64)
		if fetchedAtInt, ok := relation.Properties["fetched_at"].(int64); ok {
			fetchedAt = float64(fetchedAtInt)
		}
		log.Printf("🔍 [STORE] Saving URLArtifact %s: fetched_at=%v, UpdatedAt=%s", 
			relation.ID, fetchedAt, relation.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	
	data, err := json.MarshalIndent(relation, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal relation: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write relation file: %w", err)
	}
	
	return nil
}

// Load retrieves a relation by ID
func (store *FileRelationStore) Load(id string) (*Relation, error) {
	filename := fmt.Sprintf("relation-%s.json", id)
	filePath := filepath.Join(store.baseDir, filename)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("relation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read relation file: %w", err)
	}
	
	var relation Relation
	if err := json.Unmarshal(data, &relation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relation: %w", err)
	}
	
	// DEBUG: Log what we're loading for URLArtifacts
	if relation.Type == "URLArtifact" {
		fetchedAt, _ := relation.Properties["fetched_at"].(float64)
		if fetchedAtInt, ok := relation.Properties["fetched_at"].(int64); ok {
			fetchedAt = float64(fetchedAtInt)
		}
		log.Printf("🔍 [LOAD] Loading URLArtifact %s: fetched_at=%v, UpdatedAt=%s", 
			relation.ID, fetchedAt, relation.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	
	return &relation, nil
}

// LoadByType retrieves all relations of a specific type
func (store *FileRelationStore) LoadByType(relationType string) ([]Relation, error) {
	relations, err := store.List()
	if err != nil {
		return nil, err
	}
	
	var filtered []Relation
	for _, rel := range relations {
		if rel.Type == relationType {
			filtered = append(filtered, rel)
		}
	}
	
	return filtered, nil
}

// LoadByProperty retrieves relations with a specific property value
func (store *FileRelationStore) LoadByProperty(key, value string) ([]Relation, error) {
	relations, err := store.List()
	if err != nil {
		return nil, err
	}
	
	var filtered []Relation
	for _, rel := range relations {
		if propValue, exists := rel.Properties[key]; exists {
			if fmt.Sprintf("%v", propValue) == value {
				filtered = append(filtered, rel)
			}
		}
	}
	
	return filtered, nil
}

// Delete removes a relation
func (store *FileRelationStore) Delete(id string) error {
	filename := fmt.Sprintf("relation-%s.json", id)
	filePath := filepath.Join(store.baseDir, filename)
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("relation not found: %s", id)
		}
		return fmt.Errorf("failed to delete relation: %w", err)
	}
	
	return nil
}

// List retrieves all relations
func (store *FileRelationStore) List() ([]Relation, error) {
	var relations []Relation
	
	err := filepath.WalkDir(store.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() || !strings.HasPrefix(d.Name(), "relation-") || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read relation file %s: %w", path, err)
		}
		
		var relation Relation
		if err := json.Unmarshal(data, &relation); err != nil {
			return fmt.Errorf("failed to unmarshal relation file %s: %w", path, err)
		}
		
		relations = append(relations, relation)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list relations: %w", err)
	}
	
	return relations, nil
}

// FileMaterializationStore implements MaterializationStore using JSON files
type FileMaterializationStore struct {
	baseDir string // ~/.port42/relations/
}

// NewFileMaterializationStore creates a new file-based materialization store
func NewFileMaterializationStore(baseDir string) (*FileMaterializationStore, error) {
	relationsDir := filepath.Join(baseDir, "relations")
	
	if err := os.MkdirAll(relationsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create relations directory: %w", err)
	}
	
	return &FileMaterializationStore{
		baseDir: relationsDir,
	}, nil
}

// Save stores materialization info
func (fms *FileMaterializationStore) Save(entity MaterializedEntity) error {
	filename := fmt.Sprintf("materialized-%s.json", entity.RelationID)
	filePath := filepath.Join(fms.baseDir, filename)
	
	data, err := json.MarshalIndent(entity, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal materialization: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write materialization file: %w", err)
	}
	
	return nil
}

// Load retrieves materialization info by relation ID
func (fms *FileMaterializationStore) Load(relationID string) (*MaterializedEntity, error) {
	filename := fmt.Sprintf("materialized-%s.json", relationID)
	filePath := filepath.Join(fms.baseDir, filename)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("materialization not found for relation: %s", relationID)
		}
		return nil, fmt.Errorf("failed to read materialization file: %w", err)
	}
	
	var entity MaterializedEntity
	if err := json.Unmarshal(data, &entity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal materialization: %w", err)
	}
	
	return &entity, nil
}

// Delete removes materialization info
func (fms *FileMaterializationStore) Delete(relationID string) error {
	filename := fmt.Sprintf("materialized-%s.json", relationID)
	filePath := filepath.Join(fms.baseDir, filename)
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted, that's fine
		}
		return fmt.Errorf("failed to delete materialization: %w", err)
	}
	
	return nil
}

// List retrieves all materializations
func (fms *FileMaterializationStore) List() ([]MaterializedEntity, error) {
	var entities []MaterializedEntity
	
	err := filepath.WalkDir(fms.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if d.IsDir() || !strings.HasPrefix(d.Name(), "materialized-") || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read materialization file %s: %w", path, err)
		}
		
		var entity MaterializedEntity
		if err := json.Unmarshal(data, &entity); err != nil {
			return fmt.Errorf("failed to unmarshal materialization file %s: %w", path, err)
		}
		
		entities = append(entities, entity)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list materializations: %w", err)
	}
	
	return entities, nil
}