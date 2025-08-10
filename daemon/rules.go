package main

import (
	"fmt"
	"log"
	"strings"
)

// Rule defines an auto-spawning rule that can trigger when relations are declared
type Rule struct {
	ID          string
	Name        string
	Description string
	Condition   func(relation Relation) bool
	Action      func(relation Relation, compiler *RealityCompiler) error
	Enabled     bool
}

// RuleEngine manages and executes rules for auto-spawning entities
type RuleEngine struct {
	rules    []Rule
	compiler *RealityCompiler
}

// NewRuleEngine creates a new rule engine with the given rules
func NewRuleEngine(compiler *RealityCompiler, rules []Rule) *RuleEngine {
	return &RuleEngine{
		rules:    rules,
		compiler: compiler,
	}
}

// ProcessRelation evaluates all enabled rules against a relation and executes matching ones
func (re *RuleEngine) ProcessRelation(relation Relation) ([]string, error) {
	log.Printf("🔍 Processing relation %s through %d rules", relation.ID, len(re.rules))
	
	var spawnedIDs []string
	var errors []string
	
	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}
		
		// Check if rule condition matches
		if rule.Condition(relation) {
			log.Printf("🌱 Rule '%s' matched relation %s", rule.Name, relation.ID)
			
			// Execute rule action
			err := rule.Action(relation, re.compiler)
			if err != nil {
				errorMsg := fmt.Sprintf("Rule '%s' failed: %v", rule.Name, err)
				log.Printf("❌ %s", errorMsg)
				errors = append(errors, errorMsg)
			} else {
				log.Printf("✅ Rule '%s' executed successfully", rule.Name)
				// Note: We don't track spawned IDs yet, but rule actions can spawn relations
				// This will be enhanced in Phase 2
			}
		}
	}
	
	// Return any errors encountered
	if len(errors) > 0 {
		return spawnedIDs, fmt.Errorf("rule execution errors: %s", strings.Join(errors, "; "))
	}
	
	return spawnedIDs, nil
}

// AddRule adds a new rule to the engine
func (re *RuleEngine) AddRule(rule Rule) {
	re.rules = append(re.rules, rule)
	log.Printf("📋 Added rule: %s", rule.Name)
}

// EnableRule enables a rule by ID
func (re *RuleEngine) EnableRule(ruleID string) error {
	for i, rule := range re.rules {
		if rule.ID == ruleID {
			re.rules[i].Enabled = true
			log.Printf("✅ Enabled rule: %s", rule.Name)
			return nil
		}
	}
	return fmt.Errorf("rule not found: %s", ruleID)
}

// DisableRule disables a rule by ID
func (re *RuleEngine) DisableRule(ruleID string) error {
	for i, rule := range re.rules {
		if rule.ID == ruleID {
			re.rules[i].Enabled = false
			log.Printf("⏸️ Disabled rule: %s", rule.Name)
			return nil
		}
	}
	return fmt.Errorf("rule not found: %s", ruleID)
}

// ListRules returns all rules in the engine
func (re *RuleEngine) ListRules() []Rule {
	return re.rules
}

// Helper functions for rule conditions

// hasTransform checks if a relation has a specific transform
func hasTransform(relation Relation, transform string) bool {
	transforms := getTransforms(relation)
	return contains(transforms, transform)
}

// getTransforms extracts transforms array from relation properties
func getTransforms(relation Relation) []string {
	if transformsRaw, exists := relation.Properties["transforms"]; exists {
		if transformsList, ok := transformsRaw.([]interface{}); ok {
			var transforms []string
			for _, t := range transformsList {
				if tStr, ok := t.(string); ok {
					transforms = append(transforms, tStr)
				}
			}
			return transforms
		}
	}
	return []string{}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getRelationName extracts name from relation properties
func getRelationName(relation Relation) string {
	if name, ok := relation.Properties["name"].(string); ok {
		return name
	}
	return ""
}

// defaultRules returns the initial set of rules for the reality compiler
func defaultRules() []Rule {
	// For Phase 1, we'll return an empty slice
	// Phase 2 will add the actual ViewerRule
	return []Rule{}
}