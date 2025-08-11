package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"
)

// SimilarityCalculator handles tool similarity detection and scoring
type SimilarityCalculator struct {
	relationStore RelationStore
}

// SimilarTool represents a tool with its similarity score to a target tool
type SimilarTool struct {
	Tool       Relation `json:"tool"`
	Similarity float64  `json:"similarity"`
	Reason     []string `json:"reason"`
}

// NewSimilarityCalculator creates a new similarity calculator
func NewSimilarityCalculator(relationStore RelationStore) *SimilarityCalculator {
	return &SimilarityCalculator{
		relationStore: relationStore,
	}
}

// calculateTransformSimilarity computes similarity between two transform arrays
// Uses Jaccard similarity coefficient with semantic enhancements
func calculateTransformSimilarity(transforms1, transforms2 []string) float64 {
	if len(transforms1) == 0 || len(transforms2) == 0 {
		return 0.0
	}
	
	// Convert to sets for intersection/union calculation
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, t := range transforms1 {
		set1[strings.ToLower(strings.TrimSpace(t))] = true
	}
	for _, t := range transforms2 {
		set2[strings.ToLower(strings.TrimSpace(t))] = true
	}
	
	// Calculate intersection and union
	intersection := make(map[string]bool)
	union := make(map[string]bool)
	
	for t := range set1 {
		union[t] = true
		if set2[t] {
			intersection[t] = true
		}
	}
	for t := range set2 {
		union[t] = true
	}
	
	if len(union) == 0 {
		return 0.0
	}
	
	// Base Jaccard similarity coefficient
	baseSimilarity := float64(len(intersection)) / float64(len(union))
	
	// Add semantic boost for related transforms
	semanticBoost := calculateSemanticBoost(transforms1, transforms2)
	
	// Cap total similarity at 1.0
	return math.Min(1.0, baseSimilarity+semanticBoost)
}

// calculateSemanticBoost adds similarity for semantically related transforms
func calculateSemanticBoost(transforms1, transforms2 []string) float64 {
	// Semantic relationships between transforms
	semanticGroups := map[string][]string{
		"analyze":  {"analysis", "analyzer", "inspect", "examine", "review"},
		"parse":    {"parsing", "parser", "process", "extract", "decode"},
		"format":   {"formatting", "formatter", "display", "render", "beautify"},
		"test":     {"testing", "verify", "validation", "check", "validate"},
		"log":      {"logs", "logging", "logger", "audit", "record"},
		"data":     {"dataset", "info", "information", "content"},
		"file":     {"files", "document", "doc", "fs", "filesystem"},
		"network":  {"net", "http", "web", "api", "service"},
		"security": {"secure", "auth", "authentication", "authorization", "crypt"},
		"config":   {"configuration", "settings", "setup", "preferences"},
	}
	
	boost := 0.0
	
	// Normalize transforms to lowercase
	norm1 := make([]string, len(transforms1))
	norm2 := make([]string, len(transforms2))
	for i, t := range transforms1 {
		norm1[i] = strings.ToLower(strings.TrimSpace(t))
	}
	for i, t := range transforms2 {
		norm2[i] = strings.ToLower(strings.TrimSpace(t))
	}
	
	// Check each transform in first set against second set
	for _, t1 := range norm1 {
		for _, t2 := range norm2 {
			if t1 == t2 {
				continue // Already counted in base similarity
			}
			
			// Check if they're in the same semantic group
			for baseWord, synonyms := range semanticGroups {
				t1InGroup := (t1 == baseWord)
				t2InGroup := (t2 == baseWord)
				
				for _, synonym := range synonyms {
					if t1 == synonym {
						t1InGroup = true
					}
					if t2 == synonym {
						t2InGroup = true
					}
				}
				
				if t1InGroup && t2InGroup {
					boost += 0.15 // 15% boost per semantic match
					break
				}
			}
		}
	}
	
	// Cap semantic boost at 30%
	return math.Min(0.3, boost)
}

// findSimilarTools finds all tools similar to the target tool above the threshold
func (sc *SimilarityCalculator) findSimilarTools(targetTool Relation, threshold float64) ([]SimilarTool, error) {
	if targetTool.Type != "Tool" {
		return nil, fmt.Errorf("target relation is not a Tool: %s", targetTool.Type)
	}
	
	// Get target tool transforms
	targetTransforms, err := sc.extractTransforms(targetTool)
	if err != nil {
		return nil, fmt.Errorf("failed to extract target transforms: %v", err)
	}
	
	if len(targetTransforms) == 0 {
		return []SimilarTool{}, nil // No transforms to compare against
	}
	
	// Load all relations
	allRelations, err := sc.relationStore.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load relations: %v", err)
	}
	
	var similarTools []SimilarTool
	
	// Compare against each tool
	for _, relation := range allRelations {
		// Skip non-tools and self
		if relation.Type != "Tool" || relation.ID == targetTool.ID {
			continue
		}
		
		// Extract transforms
		candidateTransforms, err := sc.extractTransforms(relation)
		if err != nil {
			log.Printf("Warning: failed to extract transforms from %s: %v", relation.ID, err)
			continue
		}
		
		if len(candidateTransforms) == 0 {
			continue // No transforms to compare
		}
		
		// Calculate similarity
		similarity := calculateTransformSimilarity(targetTransforms, candidateTransforms)
		
		// Include if above threshold
		if similarity >= threshold {
			reasons := sc.generateReasons(targetTransforms, candidateTransforms, similarity)
			
			similarTool := SimilarTool{
				Tool:       relation,
				Similarity: similarity,
				Reason:     reasons,
			}
			
			similarTools = append(similarTools, similarTool)
		}
	}
	
	// Sort by similarity score (highest first)
	for i := 0; i < len(similarTools)-1; i++ {
		for j := i + 1; j < len(similarTools); j++ {
			if similarTools[i].Similarity < similarTools[j].Similarity {
				similarTools[i], similarTools[j] = similarTools[j], similarTools[i]
			}
		}
	}
	
	return similarTools, nil
}

// extractTransforms safely extracts transforms array from relation properties
func (sc *SimilarityCalculator) extractTransforms(relation Relation) ([]string, error) {
	transformsRaw, exists := relation.Properties["transforms"]
	if !exists {
		return []string{}, nil
	}
	
	// Handle []interface{} (JSON arrays)
	if transformsInterface, ok := transformsRaw.([]interface{}); ok {
		var transforms []string
		for _, t := range transformsInterface {
			if tStr, ok := t.(string); ok {
				transforms = append(transforms, tStr)
			}
		}
		return transforms, nil
	}
	
	// Handle []string (direct string arrays)
	if transformsString, ok := transformsRaw.([]string); ok {
		return transformsString, nil
	}
	
	return nil, fmt.Errorf("transforms property has unsupported type: %T", transformsRaw)
}

// generateReasons creates human-readable explanations for why tools are similar
func (sc *SimilarityCalculator) generateReasons(targetTransforms, candidateTransforms []string, similarity float64) []string {
	var reasons []string
	
	// Find exact matches
	exactMatches := []string{}
	targetSet := make(map[string]bool)
	for _, t := range targetTransforms {
		targetSet[strings.ToLower(t)] = true
	}
	
	for _, t := range candidateTransforms {
		if targetSet[strings.ToLower(t)] {
			exactMatches = append(exactMatches, t)
		}
	}
	
	if len(exactMatches) > 0 {
		reasons = append(reasons, fmt.Sprintf("Shared transforms: %s", strings.Join(exactMatches, ", ")))
	}
	
	// Add similarity score explanation
	if similarity >= 0.8 {
		reasons = append(reasons, "Very high similarity (>80%)")
	} else if similarity >= 0.6 {
		reasons = append(reasons, "High similarity (60-80%)")
	} else if similarity >= 0.4 {
		reasons = append(reasons, "Moderate similarity (40-60%)")
	} else {
		reasons = append(reasons, "Low similarity (20-40%)")
	}
	
	// Add semantic reasoning if applicable
	semanticBoost := calculateSemanticBoost(targetTransforms, candidateTransforms)
	
	if semanticBoost > 0.05 { // 5% threshold for semantic boost
		reasons = append(reasons, "Semantic relationship detected")
	}
	
	return reasons
}

// GetSimilarToolsForTool is a convenience method that finds a tool by name and returns its similar tools
func (sc *SimilarityCalculator) GetSimilarToolsForTool(toolName string, threshold float64) ([]SimilarTool, error) {
	// Find the target tool by name
	allRelations, err := sc.relationStore.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load relations: %v", err)
	}
	
	var targetTool *Relation
	for _, relation := range allRelations {
		if relation.Type == "Tool" {
			if name, ok := relation.Properties["name"].(string); ok && name == toolName {
				targetTool = &relation
				break
			}
		}
	}
	
	if targetTool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	
	return sc.findSimilarTools(*targetTool, threshold)
}

// createSimilarityRelationships stores similarity relationships in the relation system
func (sc *SimilarityCalculator) createSimilarityRelationships(tool Relation, threshold float64) error {
	if tool.Type != "Tool" {
		return fmt.Errorf("target relation is not a Tool: %s", tool.Type)
	}
	
	// Find similar tools using the specified threshold
	similarTools, err := sc.findSimilarTools(tool, threshold)
	if err != nil {
		return fmt.Errorf("failed to find similar tools: %v", err)
	}
	
	if len(similarTools) == 0 {
		return nil // No similar tools found, nothing to store
	}
	
	// Create bidirectional similarity relationships
	for _, simTool := range similarTools {
		// Create forward relationship (tool -> similar tool)
		forwardRelID := fmt.Sprintf("similarity-%s-%s", tool.ID, simTool.Tool.ID)
		forwardRelationship := Relation{
			ID:   forwardRelID,
			Type: "Relationship",
			Properties: map[string]interface{}{
				"relationship_type": "similar_to",
				"from":             tool.ID,
				"to":               simTool.Tool.ID,
				"similarity_score": simTool.Similarity,
				"reasons":          simTool.Reason,
				"auto_generated":   true,
				"created_by":       "similarity_calculator",
			},
			CreatedAt: time.Now(),
		}
		
		err := sc.relationStore.Save(forwardRelationship)
		if err != nil {
			log.Printf("Failed to store forward similarity relationship %s: %v", forwardRelID, err)
			continue
		}
		
		// Create reverse relationship (similar tool -> tool) for bidirectionality
		reverseRelID := fmt.Sprintf("similarity-%s-%s", simTool.Tool.ID, tool.ID)
		reverseRelationship := Relation{
			ID:   reverseRelID,
			Type: "Relationship", 
			Properties: map[string]interface{}{
				"relationship_type": "similar_to",
				"from":             simTool.Tool.ID,
				"to":               tool.ID,
				"similarity_score": simTool.Similarity,
				"reasons":          simTool.Reason,
				"auto_generated":   true,
				"created_by":       "similarity_calculator",
			},
			CreatedAt: time.Now(),
		}
		
		err = sc.relationStore.Save(reverseRelationship)
		if err != nil {
			log.Printf("Failed to store reverse similarity relationship %s: %v", reverseRelID, err)
			continue
		}
		
		log.Printf("✓ Created bidirectional similarity relationship: %s ↔ %s (%.0f%%)", 
			tool.Properties["name"], simTool.Tool.Properties["name"], simTool.Similarity*100)
	}
	
	return nil
}