package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ToolMaterializer implements Materializer for Tool relations
type ToolMaterializer struct {
	aiClient         *AnthropicClient
	storage          *Storage // Use existing storage system
	matStore         MaterializationStore
	contextCollector *ContextCollector
}

// NewToolMaterializer creates a new tool materializer
func NewToolMaterializer(aiClient *AnthropicClient, storage *Storage, matStore MaterializationStore, contextCollector *ContextCollector) (*ToolMaterializer, error) {
	return &ToolMaterializer{
		aiClient:         aiClient,
		storage:          storage,
		matStore:         matStore,
		contextCollector: contextCollector,
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
	spec, code, err := tm.generateToolCode(name, transforms, relation.ID, relation)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tool code: %w", err)
	}
	
	// Store using existing storage system (creates object store + symlink)
	if err := tm.storage.StoreCommand(spec, code); err != nil {
		return nil, fmt.Errorf("failed to store tool in object store: %w", err)
	}
	
	// Track tool creation in context collector
	if tm.contextCollector != nil {
		log.Printf("🛠 Tracking tool creation: %s", name)
		tm.contextCollector.TrackToolCreation(name, "command", transforms)
	} else {
		log.Printf("⚠️ Context collector is nil, cannot track tool: %s", name)
	}
	
	// Get the canonical object ID for the executable
	executableID, err := tm.storage.Store([]byte(code))
	if err != nil {
		return nil, fmt.Errorf("failed to get object ID for executable: %w", err)
	}
	
	// Update the relation to store the canonical object ID instead of content
	if relation.Properties == nil {
		relation.Properties = make(map[string]interface{})
	}
	relation.Properties["executable_id"] = executableID
	
	// Remove legacy executable content if it exists to save memory
	delete(relation.Properties, "executable")
	
	// Save the updated relation with the object ID
	relationStore := tm.storage.relationStore
	if relationStore != nil {
		if err := relationStore.Save(relation); err != nil {
			log.Printf("⚠️ Failed to update relation with executable_id: %v", err)
		} else {
			log.Printf("✅ Updated relation %s with executable_id: %s", relation.ID, executableID[:12]+"...")
		}
	}
	
	// Get the symlink path for the materialized entity
	homeDir, _ := os.UserHomeDir()
	toolPath := filepath.Join(homeDir, ".port42", "commands", name)
	
	// Create materialized entity with actual language from spec
	entity := &MaterializedEntity{
		RelationID:   relation.ID,
		PhysicalPath: toolPath,
		Metadata: map[string]interface{}{
			"executable": true,
			"language":   spec.Language,  // Use actual selected language
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

// validateLanguageSelection ensures the selected language is supported (B2.4 Error Handling)
func (tm *ToolMaterializer) validateLanguageSelection(language string) error {
	supportedLanguages := []string{"bash", "python", "node"}
	for _, supported := range supportedLanguages {
		if language == supported {
			return nil
		}
	}
	return fmt.Errorf("unsupported language '%s', supported languages: %v", language, supportedLanguages)
}

// validateGeneratedCode performs basic syntax validation for generated code (B2.4 Error Handling)  
func (tm *ToolMaterializer) validateGeneratedCode(code string, language string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("generated code is empty")
	}
	
	switch language {
	case "bash":
		// Basic bash validation - check for common syntax errors
		if !strings.HasPrefix(code, "#!/bin/bash") && !strings.HasPrefix(code, "#!/usr/bin/env bash") {
			return fmt.Errorf("bash code missing proper shebang")
		}
		
	case "python":
		// Basic python validation
		if !strings.HasPrefix(code, "#!/usr/bin/env python3") {
			return fmt.Errorf("python code missing proper shebang")
		}
		
	case "node":
		// Basic node validation
		if !strings.HasPrefix(code, "#!/usr/bin/env node") {
			return fmt.Errorf("node code missing proper shebang")
		}
	}
	
	return nil
}

// generateToolCode creates executable code for a tool using AI (reusing existing command crystallization approach)
func (tm *ToolMaterializer) generateToolCode(name string, transforms []string, relationID string, relation Relation) (*CommandSpec, string, error) {
	// Build base prompt based on tool name and transforms
	prompt := tm.buildToolPrompt(name, transforms)
	
	// Phase 2: Add resolved context from references to enhance AI generation
	if resolvedContext, exists := relation.Properties["resolved_context"]; exists {
		if contextStr, ok := resolvedContext.(string); ok && contextStr != "" {
			log.Printf("🔗 Enhancing AI prompt with resolved context (%d chars)", len(contextStr))
			prompt = prompt + "\n\nAdditional Context from References:\n" + contextStr
		}
	}
	
	// Phase 3: Add user prompt requirements to customize generation
	if userPrompt, exists := relation.Properties["user_prompt"]; exists {
		if promptStr, ok := userPrompt.(string); ok && promptStr != "" {
			log.Printf("💬 Enhancing AI prompt with user requirements (%d chars)", len(promptStr))
			prompt = prompt + "\n\nUser Requirements:\n" + promptStr + "\n\nPlease incorporate these specific requirements into the tool implementation."
		}
	}
	
	// Use existing AI client to generate code
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	// Get agent prompt for tool creation (reuse existing logic)
	agentPrompt := getAgentPrompt("@ai-engineer")
	
	// Use SendWithoutTools for pure text generation (we want JSON, not tool execution)
	response, err := tm.aiClient.SendWithoutTools(messages, agentPrompt, "@ai-engineer")
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
	
	
	// Extract tool specification from our new unified AI response format
	spec, err := tm.extractToolSpecFromResponse(responseText)
	if err != nil {
		log.Printf("❌ DEBUG: Failed to extract tool spec. Response was:\n%s", responseText)
		
		// Write failed response to file for debugging
		if debugErr := tm.writeDebugResponse(responseText, err, relationID); debugErr != nil {
			log.Printf("⚠️ Failed to write debug file: %v", debugErr)
		}
		
		return nil, "", fmt.Errorf("failed to extract tool spec from AI response: %w", err)
	}
	
	// Validate language selection (B2.4 Error Handling)
	if err := tm.validateLanguageSelection(spec.Language); err != nil {
		log.Printf("⚠️ Invalid language selection, falling back to Python: %v", err)
		spec.Language = "python" // Fallback to Python
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
	
	// Create executable code with language-appropriate shebang (B1.2 integration)
	var shebang string
	switch spec.Language {
	case "bash":
		shebang = "#!/bin/bash"
	case "node":
		shebang = "#!/usr/bin/env node"
	default: // python
		shebang = "#!/usr/bin/env python3"
	}
	
	code := fmt.Sprintf("%s\n%s",
		shebang,
		implementation)
	
	// Validate generated code (B2.4 Error Handling)
	if err := tm.validateGeneratedCode(code, spec.Language); err != nil {
		log.Printf("⚠️ Generated code validation failed: %v", err)
		// Try to fix common issues by regenerating shebang
		lines := strings.Split(code, "\n")
		if len(lines) > 0 {
			switch spec.Language {
			case "bash":
				lines[0] = "#!/bin/bash"
			case "node":
				lines[0] = "#!/usr/bin/env node"  
			default:
				lines[0] = "#!/usr/bin/env python3"
			}
			code = strings.Join(lines, "\n")
			log.Printf("✅ Fixed shebang for %s code", spec.Language)
		}
	}
	
	return spec, code, nil
}

// extractToolSpecFromResponse parses our new clean slate AI response format
func (tm *ToolMaterializer) extractToolSpecFromResponse(responseText string) (*CommandSpec, error) {
	// Define our new clean slate response structure
	type ToolResponse struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`  
		Language       string   `json:"language"`
		Implementation string   `json:"implementation"`
		Tags           []string `json:"tags"`
	}
	
	// Look for JSON code block (same as legacy)
	startMarker := "```json"
	endMarker := "```"
	
	startIdx := strings.Index(responseText, startMarker)
	if startIdx == -1 {
		return nil, fmt.Errorf("no JSON code block found in response")
	}
	
	startIdx += len(startMarker)
	endIdx := strings.Index(responseText[startIdx:], endMarker)
	if endIdx == -1 {
		return nil, fmt.Errorf("unclosed JSON code block in response")
	}
	
	jsonStr := strings.TrimSpace(responseText[startIdx : startIdx+endIdx])
	log.Printf("🔍 DEBUG: Extracted JSON:\n%s", jsonStr)
	
	// Parse our new clean slate format
	var toolResp ToolResponse
	if err := json.Unmarshal([]byte(jsonStr), &toolResp); err != nil {
		return nil, fmt.Errorf("failed to parse tool response JSON: %w", err)
	}
	
	// Validate required fields
	if toolResp.Name == "" {
		return nil, fmt.Errorf("missing required field: name")
	}
	if toolResp.Implementation == "" {
		return nil, fmt.Errorf("missing required field: implementation")
	}
	if toolResp.Language == "" {
		log.Printf("⚠️ No language specified, defaulting to python")
		toolResp.Language = "python"
	}
	if toolResp.Description == "" {
		toolResp.Description = fmt.Sprintf("A %s tool for processing data", toolResp.Language)
	}
	
	// Convert to CommandSpec (for compatibility with existing materialization)
	spec := &CommandSpec{
		Name:           toolResp.Name,
		Description:    toolResp.Description,
		Language:       toolResp.Language,
		Implementation: toolResp.Implementation,
		Dependencies:   []string{}, // AI handles dependencies in implementation
		Tags:           toolResp.Tags, // Use AI-generated tags
		// Other fields will be set by materialization process
	}
	
	log.Printf("✅ Successfully parsed clean slate tool spec: %s (%s)", spec.Name, spec.Language)
	return spec, nil
}

// Dependencies are now intelligently determined by AI as part of tool generation
// instead of using hardcoded keyword matching

// Dependency checking is now handled intelligently by AI as part of code generation

// getLanguageTemplate returns language-specific template structure (B2.1-B2.3)
func (tm *ToolMaterializer) getLanguageTemplate(language string) string {
	switch language {
	case "bash":
		// B2.1 Bash Template
		return `#!/bin/bash
# Tool: {name}
# Description: {description}
# Transforms: {transforms}

set -euo pipefail  # Error handling

# Dependency checks will be added here
{dependency_checks}

# Argument parsing
{argument_parsing}

# Main implementation
{main_implementation}

# Error handling
{error_handling}`

	case "node":
		// B2.3 Node Template
		return `#!/usr/bin/env node
// {name}: {description}
// Transforms: {transforms}

{dependency_checks}
{argument_parsing}
{main_implementation}
{error_handling}`

	default: // python
		// B2.2 Python Template
		return `#!/usr/bin/env python3
"""
{name}: {description}
Transforms: {transforms}
"""

import argparse
import sys
{additional_imports}

{dependency_checks}
{main_implementation}
{error_handling}

if __name__ == "__main__":
    main()`
	}
}

// selectLanguageForTool uses AI to determine the best language based on transforms
func (tm *ToolMaterializer) selectLanguageForTool(transforms []string) string {
	// For simple cases, use fast heuristics
	if len(transforms) == 0 {
		return "python" // Default for empty transforms
	}
	
	// Use AI to make the language selection decision
	return tm.selectLanguageWithAI(transforms)
}

// selectLanguageWithAI uses AI to intelligently select language based on transforms
func (tm *ToolMaterializer) selectLanguageWithAI(transforms []string) string {
	prompt := fmt.Sprintf(`Given these transforms for a command-line tool: %s

Select the most appropriate programming language (bash, python, or node) based on:

BASH is ideal for:
- File system operations, Git operations, system administration
- Text processing with pipes and streams
- Process management and system integration
- Simple automation and glue scripts

PYTHON is ideal for:
- Data processing, analysis, and transformation
- API clients and HTTP operations  
- JSON/XML/YAML processing and validation
- Mathematical calculations and statistics
- General-purpose automation with libraries

NODE is ideal for:
- Web servers, REST APIs, and GraphQL
- Interactive tools and user interfaces
- Modern web development and frontends
- Real-time applications and WebSocket handling

Respond with ONLY the language name: bash, python, or node`, strings.Join(transforms, ", "))

	// Get AI response for language selection
	messages := []Message{{Role: "user", Content: prompt}}
	response, err := tm.aiClient.Send(messages, "", "language-selector")
	if err != nil {
		// Fallback to simple heuristics if AI fails
		return tm.selectLanguageWithHeuristics(transforms)
	}
	
	// Parse AI response
	var responseText string
	if len(response.Content) > 0 {
		responseText = response.Content[0].Text
	} else {
		return tm.selectLanguageWithHeuristics(transforms)
	}
	
	language := strings.ToLower(strings.TrimSpace(responseText))
	switch language {
	case "bash":
		return "bash"
	case "python":
		return "python" 
	case "node":
		return "node"
	default:
		// Fallback if AI returns unexpected response
		return tm.selectLanguageWithHeuristics(transforms)
	}
}

// selectLanguageWithHeuristics provides fallback logic when AI is unavailable
func (tm *ToolMaterializer) selectLanguageWithHeuristics(transforms []string) string {
	transformsStr := strings.ToLower(strings.Join(transforms, " "))
	
	// Simple heuristics as fallback
	bashIndicators := []string{"git", "file", "directory", "system", "process", "pipe", "stream", "service", "os", "native"}
	nodeIndicators := []string{"web", "server", "ui", "graphql", "frontend", "interactive"}
	
	bashScore := 0
	nodeScore := 0
	
	for _, indicator := range bashIndicators {
		if strings.Contains(transformsStr, indicator) {
			bashScore++
		}
	}
	
	for _, indicator := range nodeIndicators {
		if strings.Contains(transformsStr, indicator) {
			nodeScore++
		}
	}
	
	if bashScore > nodeScore {
		return "bash"
	} else if nodeScore > 0 {
		return "node"
	}
	
	return "python" // Default fallback
}

// buildToolPrompt creates a unified prompt for AI tool generation with integrated language selection
func (tm *ToolMaterializer) buildToolPrompt(name string, transforms []string) string {
	// Check if this is a viewer tool (auto-spawned)
	isViewer := strings.HasPrefix(name, "view-") && contains(transforms, "view")
	
	var prompt string
	if isViewer {
		// Special prompt for viewer tools
		originalTool := strings.TrimPrefix(name, "view-")
		prompt = fmt.Sprintf("Create a viewer tool called '%s' that takes output from the '%s' tool and formats/displays it in a more visual or readable way. It should accept the '%s' tool's output (via stdin, pipe, or file) and transform it for better presentation such as tables, charts, colored output, or formatted text. Focus on visualization and formatting, not new analysis.", name, originalTool, originalTool)
	} else {
		prompt = fmt.Sprintf("Create a command-line tool called '%s'", name)
		if len(transforms) > 0 {
			prompt += fmt.Sprintf(" that transforms/processes: %s", strings.Join(transforms, ", "))
		}
	}
	
	// XML-structured prompt for better organization and response control
	prompt += fmt.Sprintf(`

<task>
Analyze the transforms and create a practical command-line tool.
</task>

<language_selection>
BASH: File system operations, Git operations, system administration, native OS interaction, text processing with pipes
PYTHON: Data processing, analysis, transformation, API clients, JSON/XML/YAML processing, mathematical calculations
NODE: Web servers, REST APIs, GraphQL, interactive tools, user interfaces, real-time applications
</language_selection>

<implementation_guidelines>
1. Use native OS APIs and tools when appropriate (e.g., osascript on macOS, notify-send on Linux, systemctl for services)
2. Prefer built-in OS tools for OS-specific tasks rather than reimplementing functionality
3. Use the right tool for the platform (e.g., AppleScript via osascript for macOS GUI/sound/notifications)
</implementation_guidelines>

<code_quality>
1. Write clean, functional code with proper error handling and argument parsing
2. Include basic error messages and help text
3. Keep implementation concise but complete
4. Make it useful and practical
5. DO NOT include any comments - generate clean code without commentary
</code_quality>

<dependency_management>
1. Use only standard library modules when possible
2. For Python: Handle missing modules gracefully with helpful install messages
3. Check for and declare any required external dependencies
</dependency_management>

<metadata>
1. Select the best language (bash, python, or node) based on the transforms
2. Generate 3-5 semantic tags that describe the tool's purpose and domain
3. Write a clear, concise description of what the tool does
</metadata>

<output_format>
Respond with a JSON object in this EXACT format (no additional text):
</output_format>

` + "```json\n" + `{
  "name": "%s",
  "description": "Brief description of what this tool does",
  "language": "your_selected_language_here",
  "tags": ["semantic-tag1", "domain-tag", "tool-type", "functionality"],
  "implementation": "Your complete implementation here\\nUse \\\\n for line breaks, escape quotes with \\\\\\\"\\nDo NOT include shebang - it will be added automatically"
}
` + "```", name)

	return prompt
}

// writeDebugResponse writes failed AI responses to ~/.port42/debug/ for analysis
func (tm *ToolMaterializer) writeDebugResponse(response string, err error, relationID string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	debugDir := filepath.Join(homeDir, ".port42", "debug")
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		return fmt.Errorf("failed to create debug directory: %w", err)
	}
	
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("failed_response_%s_%s.txt", timestamp, relationID[:8])
	filepath := filepath.Join(debugDir, filename)
	
	content := fmt.Sprintf("FAILED AI RESPONSE DEBUG LOG\n")
	content += fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("Relation ID: %s\n", relationID)
	content += fmt.Sprintf("Error: %v\n", err)
	content += fmt.Sprintf("=== AI RESPONSE START ===\n%s\n=== AI RESPONSE END ===\n", response)
	
	return os.WriteFile(filepath, []byte(content), 0644)
}



