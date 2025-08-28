package main

import (
	"fmt"
	"log"
	"strings"
	"time"
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
	return []Rule{
		viewerRule(),
		documentationRule(),
		gitToolsRule(),
		testSuiteRule(),
		documentationEmergenceRule(),
	}
}

// viewerRule creates a rule that auto-spawns viewer tools for analysis tools
func viewerRule() Rule {
	return Rule{
		ID:          "spawn-viewer",
		Name:        "Auto-spawn viewer for analysis tools",
		Description: "When a tool with 'analysis' transform is declared, automatically create a corresponding viewer tool",
		Enabled:     true,
		Condition: func(relation Relation) bool {
			// Only process Tool relations
			if relation.Type != "Tool" {
				return false
			}
			
			// Skip auto-spawned tools to prevent infinite recursion
			if autoSpawned, exists := relation.Properties["auto_spawned"]; exists {
				if spawned, ok := autoSpawned.(bool); ok && spawned {
					return false
				}
			}
			
			// Check if it has 'analysis' in its transforms
			transforms := getTransforms(relation)
			return contains(transforms, "analysis")
		},
		Action: func(relation Relation, compiler *RealityCompiler) error {
			toolName := getRelationName(relation)
			if toolName == "" {
				return fmt.Errorf("tool relation missing name property")
			}
			
			// Create viewer relation
			viewerName := "view-" + toolName
			viewerRelation := Relation{
				ID:   generateRelationID("tool", viewerName),
				Type: "Tool",
				Properties: map[string]interface{}{
					"name":         viewerName,
					"transforms":   []string{"view", "display", "format"},
					"parent":       toolName,
					"spawned_by":   relation.ID,
					"auto_spawned": true,
				},
				CreatedAt: time.Now(),
			}
			
			log.Printf("🌱 Auto-spawning viewer tool: %s", viewerName)
			
			// Use the reality compiler to declare the viewer relation
			// This will trigger full materialization including storage
			_, err := compiler.DeclareRelation(viewerRelation)
			if err != nil {
				return fmt.Errorf("failed to materialize viewer tool: %w", err)
			}
			
			log.Printf("✅ Successfully spawned viewer tool: %s", viewerName)
			return nil
		},
	}
}

// documentationRule creates a rule that auto-spawns documentation for complex tools
func documentationRule() Rule {
	return Rule{
		ID:          "spawn-documentation",
		Name:        "Auto-spawn documentation for complex tools",
		Description: "When a tool with 3+ transforms is declared, automatically create documentation artifact",
		Enabled:     true,
		Condition: func(relation Relation) bool {
			// Only process Tool relations
			if relation.Type != "Tool" {
				return false
			}
			
			// Skip auto-spawned tools to prevent infinite recursion
			if autoSpawned, exists := relation.Properties["auto_spawned"]; exists {
				if spawned, ok := autoSpawned.(bool); ok && spawned {
					return false
				}
			}
			
			// Check if tool has 3+ transforms (complex tool)
			transforms := getTransforms(relation)
			return len(transforms) >= 3
		},
		Action: func(relation Relation, compiler *RealityCompiler) error {
			toolName := getRelationName(relation)
			if toolName == "" {
				return fmt.Errorf("tool relation missing name property")
			}
			
			// Get transforms for documentation content
			transforms := getTransforms(relation)
			
			// Create documentation relation
			docsName := toolName + "-docs"
			docsRelation := Relation{
				ID:   generateRelationID("artifact", docsName),
				Type: "Artifact",
				Properties: map[string]interface{}{
					"name":         docsName,
					"type":         "documentation",
					"format":       "markdown",
					"documents":    toolName,
					"spawned_by":   relation.ID,
					"auto_spawned": true,
					"transforms":   transforms, // Include original transforms for context
					"description":  fmt.Sprintf("Auto-generated documentation for %s", toolName),
				},
				CreatedAt: time.Now(),
			}
			
			log.Printf("📚 Auto-spawning documentation: %s", docsName)
			
			// Use the reality compiler to declare the documentation relation
			_, err := compiler.DeclareRelation(docsRelation)
			if err != nil {
				return fmt.Errorf("failed to materialize documentation: %w", err)
			}
			
			log.Printf("✅ Successfully spawned documentation: %s", docsName)
			return nil
		},
	}
}

// gitToolsRule creates a rule that auto-spawns git management tools
func gitToolsRule() Rule {
	return Rule{
		ID:          "spawn-git-tools",
		Name:        "Auto-spawn git tools",
		Description: "When git-related tools are declared, automatically create git workflow helpers",
		Enabled:     true,
		Condition: func(relation Relation) bool {
			// Only process Tool relations
			if relation.Type != "Tool" {
				return false
			}
			
			// Skip auto-spawned tools to prevent infinite recursion
			if autoSpawned, exists := relation.Properties["auto_spawned"]; exists {
				if spawned, ok := autoSpawned.(bool); ok && spawned {
					return false
				}
			}
			
			// Check if this is a git-related tool
			toolName := getRelationName(relation)
			transforms := getTransforms(relation)
			
			// Look for git indicators in name
			gitNamePatterns := []string{"git", "commit", "push", "pull", "branch", "merge"}
			for _, pattern := range gitNamePatterns {
				if strings.Contains(strings.ToLower(toolName), pattern) {
					return true
				}
			}
			
			// Look for git indicators in transforms
			gitTransformPatterns := []string{"git", "version", "commit", "repository", "branch"}
			for _, transform := range transforms {
				for _, pattern := range gitTransformPatterns {
					if strings.Contains(strings.ToLower(transform), pattern) {
						return true
					}
				}
			}
			
			return false
		},
		Action: func(relation Relation, compiler *RealityCompiler) error {
			toolName := getRelationName(relation)
			if toolName == "" {
				return fmt.Errorf("git tool relation missing name property")
			}
			
			// Auto-spawn git status checker
			statusToolName := "git-status-enhanced"
			statusRelation := Relation{
				ID:   generateRelationID("Tool", statusToolName),
				Type: "Tool",
				Properties: map[string]interface{}{
					"name":         statusToolName,
					"transforms":   []string{"git", "status", "enhanced", "display"},
					"description":  "Enhanced git status with branch info and commit details",
					"spawned_by":   relation.ID,
					"auto_spawned": true,
				},
				CreatedAt: time.Now(),
			}
			
			log.Printf("🔧 Auto-spawning git tool: %s", statusToolName)
			
			// Use the reality compiler to declare the git tool relation
			_, err := compiler.DeclareRelation(statusRelation)
			if err != nil {
				return fmt.Errorf("failed to materialize git tool: %w", err)
			}
			
			log.Printf("✅ Successfully spawned git tool: %s", statusToolName)
			return nil
		},
	}
}

// testSuiteRule creates a rule that auto-spawns test management tools
func testSuiteRule() Rule {
	return Rule{
		ID:          "spawn-test-suite",
		Name:        "Auto-spawn test suite tools",
		Description: "When test-related tools are declared, automatically create test automation helpers",
		Enabled:     true,
		Condition: func(relation Relation) bool {
			// Only process Tool relations
			if relation.Type != "Tool" {
				return false
			}
			
			// Skip auto-spawned tools to prevent infinite recursion
			if autoSpawned, exists := relation.Properties["auto_spawned"]; exists {
				if spawned, ok := autoSpawned.(bool); ok && spawned {
					return false
				}
			}
			
			// Check if this is a test-related tool
			toolName := getRelationName(relation)
			transforms := getTransforms(relation)
			
			// Look for test indicators in name
			testNamePatterns := []string{"test", "spec", "unit", "integration", "e2e", "pytest", "jest"}
			for _, pattern := range testNamePatterns {
				if strings.Contains(strings.ToLower(toolName), pattern) {
					return true
				}
			}
			
			// Look for test indicators in transforms
			testTransformPatterns := []string{"test", "testing", "spec", "unit", "integration", "validation"}
			for _, transform := range transforms {
				for _, pattern := range testTransformPatterns {
					if strings.Contains(strings.ToLower(transform), pattern) {
						return true
					}
				}
			}
			
			return false
		},
		Action: func(relation Relation, compiler *RealityCompiler) error {
			toolName := getRelationName(relation)
			if toolName == "" {
				return fmt.Errorf("test tool relation missing name property")
			}
			
			// Auto-spawn test runner
			runnerToolName := "test-runner-enhanced"
			runnerRelation := Relation{
				ID:   generateRelationID("Tool", runnerToolName),
				Type: "Tool",
				Properties: map[string]interface{}{
					"name":         runnerToolName,
					"transforms":   []string{"test", "runner", "automation", "enhanced"},
					"description":  "Enhanced test runner with coverage and reporting",
					"spawned_by":   relation.ID,
					"auto_spawned": true,
				},
				CreatedAt: time.Now(),
			}
			
			log.Printf("🧪 Auto-spawning test tool: %s", runnerToolName)
			
			// Use the reality compiler to declare the test tool relation
			_, err := compiler.DeclareRelation(runnerRelation)
			if err != nil {
				return fmt.Errorf("failed to materialize test tool: %w", err)
			}
			
			log.Printf("✅ Successfully spawned test tool: %s", runnerToolName)
			return nil
		},
	}
}

// documentationEmergenceRule creates a rule that detects documentation ecosystem patterns
func documentationEmergenceRule() Rule {
	return Rule{
		ID:          "documentation-emergence",
		Name:        "Documentation Emergence Intelligence",
		Description: "Detects documentation-focused tools and auto-spawns documentation infrastructure",
		Enabled:     true,
		Condition: func(relation Relation) bool {
			// Only process Tool relations
			if relation.Type != "Tool" {
				return false
			}
			
			// Skip auto-spawned tools to prevent infinite recursion
			if autoSpawned, exists := relation.Properties["auto_spawned"]; exists {
				if spawned, ok := autoSpawned.(bool); ok && spawned {
					return false
				}
			}
			
			// Check if this is a documentation-related tool
			toolName := getRelationName(relation)
			transforms := getTransforms(relation)
			description := ""
			if desc, ok := relation.Properties["description"].(string); ok {
				description = desc
			}
			
			// Look for documentation indicators in name
			docNamePatterns := []string{"readme", "doc", "docs", "write", "note", "changelog", "manual", "guide", "wiki"}
			for _, pattern := range docNamePatterns {
				if strings.Contains(strings.ToLower(toolName), pattern) {
					return true
				}
			}
			
			// Look for documentation indicators in transforms
			docTransformPatterns := []string{"docs", "documentation", "readme", "write", "note", "changelog", "manual", "guide", "wiki", "markdown"}
			for _, transform := range transforms {
				for _, pattern := range docTransformPatterns {
					if strings.Contains(strings.ToLower(transform), pattern) {
						return true
					}
				}
			}
			
			// Look for documentation indicators in description
			docDescPatterns := []string{"documentation", "readme", "docs", "manual", "guide", "notes", "changelog"}
			for _, pattern := range docDescPatterns {
				if strings.Contains(strings.ToLower(description), pattern) {
					return true
				}
			}
			
			return false
		},
		Action: func(relation Relation, compiler *RealityCompiler) error {
			toolName := getRelationName(relation)
			if toolName == "" {
				return fmt.Errorf("documentation tool relation missing name property")
			}
			
			log.Printf("📝 Documentation emergence detected for: %s", toolName)
			
			// Check what documentation infrastructure already exists
			// This would be enhanced to query existing tools in the future
			var spawnedTools []string
			
			// 1. Auto-spawn documentation template generator
			templateToolName := "doc-template-generator"
			templateRelation := Relation{
				ID:   generateRelationID("Tool", templateToolName),
				Type: "Tool",
				Properties: map[string]interface{}{
					"name":         templateToolName,
					"transforms":   []string{"documentation", "template", "generator", "scaffold"},
					"description":  "Generates documentation templates for projects (README, CONTRIBUTING, API docs)",
					"spawned_by":   relation.ID,
					"auto_spawned": true,
					"emergence_type": "documentation_infrastructure",
				},
				CreatedAt: time.Now(),
			}
			
			_, err := compiler.DeclareRelation(templateRelation)
			if err != nil {
				log.Printf("❌ Failed to spawn %s: %v", templateToolName, err)
			} else {
				log.Printf("✅ Spawned documentation infrastructure: %s", templateToolName)
				spawnedTools = append(spawnedTools, templateToolName)
			}
			
			// 2. Auto-spawn documentation validator
			validatorToolName := "doc-validator"
			validatorRelation := Relation{
				ID:   generateRelationID("Tool", validatorToolName),
				Type: "Tool",
				Properties: map[string]interface{}{
					"name":         validatorToolName,
					"transforms":   []string{"documentation", "validation", "quality", "check"},
					"description":  "Validates documentation quality, checks for broken links, spelling, completeness",
					"spawned_by":   relation.ID,
					"auto_spawned": true,
					"emergence_type": "documentation_infrastructure",
				},
				CreatedAt: time.Now(),
			}
			
			_, err = compiler.DeclareRelation(validatorRelation)
			if err != nil {
				log.Printf("❌ Failed to spawn %s: %v", validatorToolName, err)
			} else {
				log.Printf("✅ Spawned documentation infrastructure: %s", validatorToolName)
				spawnedTools = append(spawnedTools, validatorToolName)
			}
			
			// 3. Auto-spawn documentation site generator
			siteGeneratorName := "doc-site-builder"
			siteRelation := Relation{
				ID:   generateRelationID("Tool", siteGeneratorName),
				Type: "Tool",
				Properties: map[string]interface{}{
					"name":         siteGeneratorName,
					"transforms":   []string{"documentation", "site", "generator", "static", "web"},
					"description":  "Builds static documentation sites from markdown files (supports themes, search, navigation)",
					"spawned_by":   relation.ID,
					"auto_spawned": true,
					"emergence_type": "documentation_infrastructure",
				},
				CreatedAt: time.Now(),
			}
			
			_, err = compiler.DeclareRelation(siteRelation)
			if err != nil {
				log.Printf("❌ Failed to spawn %s: %v", siteGeneratorName, err)
			} else {
				log.Printf("✅ Spawned documentation infrastructure: %s", siteGeneratorName)
				spawnedTools = append(spawnedTools, siteGeneratorName)
			}
			
			log.Printf("🎯 Documentation emergence complete - spawned %d infrastructure tools", len(spawnedTools))
			return nil
		},
	}
}