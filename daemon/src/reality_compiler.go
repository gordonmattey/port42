package main

import (
	"fmt"
	"log"
	"time"
)

// RealityCompiler orchestrates the materialization of declared relations
type RealityCompiler struct {
	relationStore RelationStore
	materializers []Materializer
	ruleEngine    *RuleEngine // Step 2: Auto-spawning rules
}

// NewRealityCompiler creates a new reality compiler
func NewRealityCompiler(relationStore RelationStore, materializers []Materializer) *RealityCompiler {
	rc := &RealityCompiler{
		relationStore: relationStore,
		materializers: materializers,
	}
	
	// Initialize rule engine with default rules
	rc.ruleEngine = NewRuleEngine(rc, defaultRules())
	log.Printf("🔧 Initialized rule engine with %d default rules", len(defaultRules()))
	
	return rc
}

// GetRelationStore returns the relation store (for similarity calculator access)
func (rc *RealityCompiler) GetRelationStore() RelationStore {
	return rc.relationStore
}

// DeclareRelation declares that a relation should exist and materializes it
func (rc *RealityCompiler) DeclareRelation(relation Relation) (*MaterializedEntity, error) {
	log.Printf("🌟 Declaring relation: %s (type: %s)", relation.ID, relation.Type)
	
	// Set timestamps
	now := time.Now()
	if relation.CreatedAt.IsZero() {
		relation.CreatedAt = now
	}
	relation.UpdatedAt = now
	
	// Store the relation (what should exist)
	if err := rc.relationStore.Save(relation); err != nil {
		return nil, fmt.Errorf("failed to store relation: %w", err)
	}
	
	log.Printf("✅ Relation stored: %s", relation.ID)
	
	// Check if this relation type needs materialization
	if !rc.shouldMaterialize(relation) {
		log.Printf("📊 Data-only relation stored: %s (type: %s)", relation.ID, relation.Type)
		// Return a virtual entity for data-only relations
		entity := &MaterializedEntity{
			RelationID:   relation.ID,
			PhysicalPath: "", // No physical path for data relations
			Metadata:     relation.Properties,
			Status:       MaterializedSuccess,
			CreatedAt:    time.Now(),
		}
		
		// Still trigger rules for data relations
		if rc.ruleEngine != nil {
			spawned, err := rc.ruleEngine.ProcessRelation(relation)
			if err != nil {
				log.Printf("⚠️ Rule processing failed: %v", err)
			} else if len(spawned) > 0 {
				log.Printf("🌱 Auto-spawned %d entities from rules", len(spawned))
			}
		}
		
		return entity, nil
	}
	
	// Find appropriate materializer for physical relations
	materializer := rc.findMaterializer(relation)
	if materializer == nil {
		return nil, fmt.Errorf("no materializer found for relation type: %s", relation.Type)
	}
	
	log.Printf("🔨 Found materializer for type: %s", relation.Type)
	
	// Materialize into physical reality
	entity, err := materializer.Materialize(relation)
	if err != nil {
		return nil, fmt.Errorf("materialization failed: %w", err)
	}
	
	log.Printf("🎉 Relation materialized successfully: %s -> %s", relation.ID, entity.PhysicalPath)
	
	// Step 2: Trigger auto-spawning rules (for materialized relations)
	if rc.ruleEngine != nil {
		spawned, err := rc.ruleEngine.ProcessRelation(relation)
		if err != nil {
			log.Printf("⚠️ Rule processing failed: %v", err)
			// Don't fail the entire operation for rule processing errors
		} else if len(spawned) > 0 {
			log.Printf("🌱 Auto-spawned %d entities from rules", len(spawned))
		}
	}
	
	return entity, nil
}

// shouldMaterialize determines if a relation type needs physical materialization
func (rc *RealityCompiler) shouldMaterialize(relation Relation) bool {
	// Data-only relation types don't need physical materialization
	dataOnlyTypes := map[string]bool{
		"URLArtifact": true,
		"Artifact":    true, // Documentation and other artifacts are metadata-only
		// Add other data-only types as needed:
		// "SearchResult": true,
		// "MemoryContext": true,
	}
	return !dataOnlyTypes[relation.Type]
}

// findMaterializer finds the appropriate materializer for a relation
func (rc *RealityCompiler) findMaterializer(relation Relation) Materializer {
	for _, materializer := range rc.materializers {
		if materializer.CanMaterialize(relation) {
			return materializer
		}
	}
	return nil
}

// GetRelation retrieves a relation by ID
func (rc *RealityCompiler) GetRelation(id string) (*Relation, error) {
	return rc.relationStore.Load(id)
}

// ListRelations retrieves all relations
func (rc *RealityCompiler) ListRelations() ([]Relation, error) {
	return rc.relationStore.List()
}

// ListRelationsByType retrieves relations of a specific type
func (rc *RealityCompiler) ListRelationsByType(relationType string) ([]Relation, error) {
	return rc.relationStore.LoadByType(relationType)
}

// DeleteRelation removes a relation and attempts to dematerialize it
func (rc *RealityCompiler) DeleteRelation(id string) error {
	log.Printf("🗑️ Deleting relation: %s", id)
	
	// Load relation first
	relation, err := rc.relationStore.Load(id)
	if err != nil {
		return fmt.Errorf("failed to load relation for deletion: %w", err)
	}
	
	// Find materializer and attempt dematerialization
	materializer := rc.findMaterializer(*relation)
	if materializer != nil {
		// Try to load materialization info
		if matStore, ok := materializer.(*ToolMaterializer); ok {
			if entity, err := matStore.matStore.Load(id); err == nil {
				if err := materializer.Dematerialize(entity); err != nil {
					log.Printf("⚠️ Failed to dematerialize relation %s: %v", id, err)
					// Continue with deletion even if dematerialization fails
				}
			}
		}
	}
	
	// Delete the relation
	if err := rc.relationStore.Delete(id); err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}
	
	log.Printf("✅ Relation deleted successfully: %s", id)
	return nil
}

// SetRuleEngine sets the rule engine for auto-spawning behavior
func (rc *RealityCompiler) SetRuleEngine(ruleEngine *RuleEngine) {
	rc.ruleEngine = ruleEngine
	log.Printf("🎯 Rule engine attached to reality compiler")
}