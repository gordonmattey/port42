package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ToolMaterializer implements Materializer for Tool relations
type ToolMaterializer struct {
	aiClient *AnthropicClient
	storage  *Storage // Use existing storage system
	matStore MaterializationStore
}

// NewToolMaterializer creates a new tool materializer
func NewToolMaterializer(aiClient *AnthropicClient, storage *Storage, matStore MaterializationStore) (*ToolMaterializer, error) {
	return &ToolMaterializer{
		aiClient: aiClient,
		storage:  storage,
		matStore: matStore,
	}, nil
}

// CanMaterialize checks if this materializer can handle the relation
func (tm *ToolMaterializer) CanMaterialize(relation Relation) bool {
	return relation.Type == "Tool"
}

// Materialize creates a physical tool from a Tool relation
func (tm *ToolMaterializer) Materialize(relation Relation) (*MaterializedEntity, error) {
	log.Printf("🔨 Materializing tool relation: %s", relation.ID)
	
	// Extract tool properties
	name, ok := relation.Properties["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool relation missing 'name' property")
	}
	
	// Get transforms (optional)
	transforms := []string{}
	if transformsRaw, exists := relation.Properties["transforms"]; exists {
		if transformsList, ok := transformsRaw.([]interface{}); ok {
			for _, t := range transformsList {
				if tStr, ok := t.(string); ok {
					transforms = append(transforms, tStr)
				}
			}
		}
	}
	
	log.Printf("🔨 Generating code for tool: %s with transforms: %v", name, transforms)
	
	// Generate tool code using AI - this returns a CommandSpec 
	spec, code, err := tm.generateToolCode(name, transforms, relation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tool code: %w", err)
	}
	
	// Store using existing storage system (creates object store + symlink)
	if err := tm.storage.StoreCommand(spec, code); err != nil {
		return nil, fmt.Errorf("failed to store tool in object store: %w", err)
	}
	
	// Get the symlink path for the materialized entity
	homeDir, _ := os.UserHomeDir()
	toolPath := filepath.Join(homeDir, ".port42", "commands", name)
	
	// Create materialized entity
	entity := &MaterializedEntity{
		RelationID:   relation.ID,
		PhysicalPath: toolPath,
		Metadata: map[string]interface{}{
			"executable": true,
			"language":   "python",  // For now, assume Python tools
			"transforms": transforms,
		},
		Status:    MaterializedSuccess,
		CreatedAt: time.Now(),
	}
	
	// Save materialization info
	if err := tm.matStore.Save(*entity); err != nil {
		log.Printf("⚠️ Failed to save materialization info: %v", err)
		// Don't fail materialization for this
	}
	
	log.Printf("✅ Tool materialized successfully: %s -> %s", name, toolPath)
	return entity, nil
}

// Dematerialize removes the physical manifestation of a tool
func (tm *ToolMaterializer) Dematerialize(entity *MaterializedEntity) error {
	log.Printf("🗑️ Dematerializing tool: %s", entity.PhysicalPath)
	
	// Remove physical file
	if err := os.Remove(entity.PhysicalPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove tool file: %w", err)
	}
	
	// Remove materialization info
	if err := tm.matStore.Delete(entity.RelationID); err != nil {
		log.Printf("⚠️ Failed to delete materialization info: %v", err)
		// Don't fail dematerialization for this
	}
	
	log.Printf("✅ Tool dematerialized successfully: %s", entity.PhysicalPath)
	return nil
}

// generateToolCode creates executable code for a tool using AI (reusing existing command crystallization approach)
func (tm *ToolMaterializer) generateToolCode(name string, transforms []string, relationID string) (*CommandSpec, string, error) {
	// Build prompt based on tool name and transforms
	prompt := tm.buildToolPrompt(name, transforms)
	
	// Use existing AI client to generate code
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	// Get agent prompt for tool creation (reuse existing logic)
	agentPrompt := getAgentPrompt("@ai-engineer")
	
	response, err := tm.aiClient.Send(messages, agentPrompt, "@ai-engineer")
	if err != nil {
		return nil, "", fmt.Errorf("AI code generation failed: %w", err)
	}
	
	// Extract code from response
	if len(response.Content) == 0 {
		return nil, "", fmt.Errorf("AI returned empty response")
	}
	
	responseText := response.Content[0].Text
	if responseText == "" {
		return nil, "", fmt.Errorf("AI returned empty response")
	}
	
	// Extract CommandSpec from JSON response (same as existing command crystallization)
	spec := extractCommandSpec(responseText)
	if spec == nil {
		return nil, "", fmt.Errorf("failed to extract command spec from AI response")
	}
	
	// Add relation context to spec
	spec.SessionID = relationID // Use relation ID as session context
	spec.Agent = "@ai-engineer"
	
	// Process implementation the same way as existing command crystallization
	implementation := spec.Implementation
	
	// Remove any shebang from the implementation (we'll add the correct one)
	lines := strings.Split(implementation, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		implementation = strings.Join(lines[1:], "\n")
	}
	
	// Create executable code with proper shebang and headers (same as existing system)
	code := fmt.Sprintf("#!/usr/bin/env python3\n# Generated by Port 42 Declarative System - %s\n# %s\n\n%s",
		time.Now().Format("2006-01-02 15:04:05"),
		spec.Description,
		implementation)
	
	return spec, code, nil
}

// buildToolPrompt creates a prompt for AI tool generation using the same format as command crystallization
func (tm *ToolMaterializer) buildToolPrompt(name string, transforms []string) string {
	prompt := fmt.Sprintf("Create a command-line tool called '%s'", name)
	
	if len(transforms) > 0 {
		prompt += fmt.Sprintf(" that transforms/processes: %s", strings.Join(transforms, ", "))
	}
	
	prompt += fmt.Sprintf(`

Create a practical command-line tool that handles the specified transforms.

Requirements:
1. Write clean, functional Python code
2. Include proper argument parsing and help text
3. Handle errors gracefully with meaningful messages
4. Make it useful and practical

Respond with a JSON object in this exact format:

` + "```json\n" + `{
  "name": "%s",
  "description": "Brief description of what this tool does",
  "language": "python",
  "implementation": "import argparse\nimport sys\n\n# Your complete Python implementation here\n# Use \\n for line breaks, escape quotes with \\\"\n# Do NOT include shebang - it will be added automatically"
}
` + "```", name)

	return prompt
}


