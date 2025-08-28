package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Relation represents a declarative entity that should exist
type Relation struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // "Tool", "Artifact", "Memory"
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// MaterializedEntity represents the physical manifestation of a relation
type MaterializedEntity struct {
	RelationID   string                 `json:"relation_id"`
	PhysicalPath string                 `json:"physical_path"`
	Metadata     map[string]interface{} `json:"metadata"`
	Status       MaterializationStatus  `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
}

// MaterializationStatus tracks the state of materialization
type MaterializationStatus string

const (
	MaterializedSuccess MaterializationStatus = "success"
	MaterializedFailed  MaterializationStatus = "failed"
	MaterializedPending MaterializationStatus = "pending"
)

// RelationStore interface for storing and querying relations
type RelationStore interface {
	Save(relation Relation) error
	Load(id string) (*Relation, error)
	LoadByType(relationType string) ([]Relation, error)
	LoadByProperty(key, value string) ([]Relation, error)
	Delete(id string) error
	List() ([]Relation, error)
}

// Materializer interface for turning relations into physical reality
type Materializer interface {
	CanMaterialize(relation Relation) bool
	Materialize(relation Relation) (*MaterializedEntity, error)
	Dematerialize(entity *MaterializedEntity) error
}

// MaterializationStore tracks what has been materialized
type MaterializationStore interface {
	Save(entity MaterializedEntity) error
	Load(relationID string) (*MaterializedEntity, error)
	Delete(relationID string) error
	List() ([]MaterializedEntity, error)
}

// generateID creates a unique identifier
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateRelationID creates a relation-specific ID
func generateRelationID(relationType, name string) string {
	prefix := fmt.Sprintf("%s-%s", strings.ToLower(relationType), name)
	suffix := generateID()
	return fmt.Sprintf("%s-%s", prefix, suffix)
}