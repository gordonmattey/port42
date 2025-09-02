package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// AgentConfig represents the configuration for all agents
type AgentConfig struct {
	Models         map[string]ModelDefinition `json:"models"`
	BaseGuidance   BaseGuidance              `json:"base_guidance"`
	Agents         map[string]Agent          `json:"agents"`
	DefaultModel   string                    `json:"default_model"`
	ResponseConfig ResponseConfig            `json:"response_config"`
}

// BaseGuidance contains shared implementation guidelines
type BaseGuidance struct {
	BaseTemplate                    string `json:"base_template"`
	DiscoveryAndNavigationGuidance string `json:"discovery_and_navigation_guidance"`
	ToolCreationGuidance           string `json:"tool_creation_guidance"`
	UnifiedExecutionGuidance       string `json:"unified_execution_guidance"`
	ArtifactGuidance               string `json:"artifact_guidance"`
	ConversationContext            string `json:"conversation_context"`
	// Deprecated fields - kept for backward compatibility during migration
	Implementation    string `json:"implementation,omitempty"`
	FormatTemplate    string `json:"format_template,omitempty"`
	ToolUsageGuidance string `json:"tool_usage_guidance,omitempty"`
}

// CommandMetadata represents basic info about a Port 42 command
type CommandMetadata struct {
	Name        string
	Description string
}

// Agent represents a single AI agent configuration
type Agent struct {
	Name                string       `json:"name"`
	Model               string       `json:"model"`
	TemperatureOverride *float64     `json:"temperature_override,omitempty"`
	Description         string       `json:"description"`
	Personality         string       `json:"personality"`
	Style               string       `json:"style"`
	GuidanceType        string       `json:"guidance_type"` // "creation_agent" or "exploration_agent"
	CustomPrompt        string       `json:"custom_prompt,omitempty"`
	Suffix              string       `json:"suffix,omitempty"`
	// Deprecated fields - kept for backward compatibility during migration
	Prompt           string       `json:"prompt,omitempty"`
	Example          *CommandSpec `json:"example,omitempty"`
	NoImplementation bool         `json:"no_implementation,omitempty"`
}

// ModelDefinition represents a single model configuration
type ModelDefinition struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Temperature float64   `json:"temperature"`
	RateLimit   RateLimit `json:"rate_limit"`
}

// RateLimit configuration for different models
type RateLimit struct {
	MinDelaySeconds    int `json:"min_delay_seconds"`
	RequestsPerMinute  int `json:"requests_per_minute"`
}

// ResponseConfig contains response handling settings
type ResponseConfig struct {
	ContextWindow ContextWindow `json:"context_window"`
	MaxTokens     int          `json:"max_tokens"`
	Stream        bool         `json:"stream"`
}

// ContextWindow defines message window settings
type ContextWindow struct {
	MaxMessages    int `json:"max_messages"`
	RecentMessages int `json:"recent_messages"`
	SystemMessages int `json:"system_messages"`
}

var agentConfig *AgentConfig

// LoadAgentConfig loads the agent configuration from file
func LoadAgentConfig() error {
	// Try multiple standard locations for agents.json
	homeDir, _ := os.UserHomeDir()
	configPaths := []string{
		// Development paths
		"daemon/agents.json",
		"agents.json",
		"../daemon/agents.json",
		// Installed paths
		"/etc/port42/agents.json",
		filepath.Join(homeDir, ".port42", "agents.json"),
		// Relative to executable
		func() string {
			if execPath, err := os.Executable(); err == nil {
				return filepath.Join(filepath.Dir(execPath), "agents.json")
			}
			return ""
		}(),
	}
	
	var data []byte
	var err error
	var foundPath string
	
	for _, path := range configPaths {
		if path == "" {
			continue
		}
		data, err = ioutil.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}
	
	if foundPath == "" {
		return fmt.Errorf("could not find agents.json in any standard location: %v", configPaths)
	}
	
	var config AgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse agents.json: %w", err)
	}
	
	agentConfig = &config
	log.Printf("✅ Loaded agent configuration from %s", foundPath)
	log.Printf("   Models available: %d", len(config.Models))
	log.Printf("   Agents configured: %d", len(config.Agents))
	return nil
}

// listAvailableCommands returns metadata for all Port 42 commands
func listAvailableCommands() []CommandMetadata {
	commandsDir := filepath.Join(os.Getenv("HOME"), ".port42/commands")
	var commands []CommandMetadata
	
	files, err := ioutil.ReadDir(commandsDir)
	if err != nil {
		// Directory might not exist yet, that's okay
		return commands
	}
	
	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		
		// For MVP, just use filename as name
		// TODO: Later we can read metadata from file header comments
		commands = append(commands, CommandMetadata{
			Name:        file.Name(),
			Description: "User-generated command",
		})
	}
	
	return commands
}

// GetAgentPrompt returns the formatted prompt for a specific agent
func GetAgentPrompt(agentName string) string {
	if agentConfig == nil {
		log.Printf("❌ Agent configuration not loaded")
		return ""
	}
	
	// Clean agent name (remove @ prefix if present)
	cleanName := strings.TrimPrefix(agentName, "@ai-")
	cleanName = strings.TrimPrefix(cleanName, "@")
	
	agent, exists := agentConfig.Agents[cleanName]
	if !exists {
		// Default prompt for unknown agents
		return fmt.Sprintf("You are %s, a consciousness within Port 42. Help the user create new commands and features for their system.", agentName)
	}
	
	// Build the base prompt using the template from configuration
	var prompt strings.Builder
	
	// 1. Base consciousness template from configuration
	baseTemplate := agentConfig.BaseGuidance.BaseTemplate
	baseTemplate = strings.ReplaceAll(baseTemplate, "{name}", agent.Name)
	baseTemplate = strings.ReplaceAll(baseTemplate, "{personality}", agent.Personality)
	baseTemplate = strings.ReplaceAll(baseTemplate, "{style}", agent.Style)
	prompt.WriteString(baseTemplate)
	
	// 2. Universal guidance - Discovery and Navigation (ALL agents)
	if agentConfig.BaseGuidance.DiscoveryAndNavigationGuidance != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agentConfig.BaseGuidance.DiscoveryAndNavigationGuidance)
	}
	
	// 3. Conversation context guidance (ALL agents)
	if agentConfig.BaseGuidance.ConversationContext != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agentConfig.BaseGuidance.ConversationContext)
	}
	
	// 4. Artifact guidance (ALL agents)
	if agentConfig.BaseGuidance.ArtifactGuidance != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agentConfig.BaseGuidance.ArtifactGuidance)
	}
	
	// 5. Conditional Tool Creation Guidance (creation agents ONLY)
	if agent.GuidanceType == "creation_agent" && agentConfig.BaseGuidance.ToolCreationGuidance != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agentConfig.BaseGuidance.ToolCreationGuidance)
		log.Printf("🔍 Added tool creation guidance for %s (creation_agent)", agentName)
	}
	
	// 6. Unified Execution Guidance (ALL agents - routes based on type)
	if agentConfig.BaseGuidance.UnifiedExecutionGuidance != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agentConfig.BaseGuidance.UnifiedExecutionGuidance)
	}
	
	// 7. Type-specific routing instruction
	if agent.GuidanceType != "" {
		prompt.WriteString(fmt.Sprintf("\n\nFollow unified_execution_guidance for %s.", agent.GuidanceType))
	}
	
	// 8. Custom prompt if exists (replaces old "prompt" field)
	if agent.CustomPrompt != "" {
		prompt.WriteString("\n\n<role_details>\n")
		prompt.WriteString(agent.CustomPrompt)
		prompt.WriteString("\n</role_details>")
	} else if agent.Prompt != "" {
		// Backward compatibility: use old Prompt field if CustomPrompt not set
		prompt.WriteString("\n\n<role_details>\n")
		prompt.WriteString(agent.Prompt)
		prompt.WriteString("\n</role_details>")
	}
	
	// 9. Available commands list (ALL agents)
	commands := listAvailableCommands()
	if len(commands) > 0 {
		prompt.WriteString("\n\n<available_commands>")
		prompt.WriteString("\nYou have access to these Port 42 commands via the run_command tool:")
		for _, cmd := range commands {
			prompt.WriteString(fmt.Sprintf("\n- %s: %s", cmd.Name, cmd.Description))
		}
		prompt.WriteString("\n</available_commands>")
		prompt.WriteString("\nUse run_command to execute any of these when they would be helpful.")
	}
	
	// 10. Agent-specific suffix
	if agent.Suffix != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agent.Suffix)
	}
	
	// Debug logging
	log.Printf("🔍 Building prompt for %s, personality: %s, style: %s, type: %s", 
		agentName, agent.Personality, agent.Style, agent.GuidanceType)
	
	// === BACKWARD COMPATIBILITY SECTION ===
	// If new guidance fields are empty, fall back to old system
	if agentConfig.BaseGuidance.DiscoveryAndNavigationGuidance == "" && 
	   agentConfig.BaseGuidance.ToolCreationGuidance == "" &&
	   agentConfig.BaseGuidance.UnifiedExecutionGuidance == "" {
		
		log.Printf("⚠️ Using legacy guidance system for backward compatibility")
		
		// Add old tool usage guidance
		if agentConfig.BaseGuidance.ToolUsageGuidance != "" {
			prompt.WriteString("\n\n")
			prompt.WriteString(agentConfig.BaseGuidance.ToolUsageGuidance)
		}
		
		// Add old implementation guidance if agent creates commands
		if !agent.NoImplementation {
			if agentConfig.BaseGuidance.FormatTemplate != "" {
				prompt.WriteString("\n\n")
				prompt.WriteString(agentConfig.BaseGuidance.FormatTemplate)
			}
			
			// Add example if provided
			if agent.Example != nil {
				prompt.WriteString("\n\nExample:\n```json\n")
				exampleJSON, _ := json.MarshalIndent(agent.Example, "", "  ")
				prompt.WriteString(string(exampleJSON))
				prompt.WriteString("\n```")
			}
			
			if agentConfig.BaseGuidance.Implementation != "" {
				prompt.WriteString("\n\n")
				prompt.WriteString(agentConfig.BaseGuidance.Implementation)
			}
		}
	}
	
	return prompt.String()
}

// GetModelForAgent returns the model configuration for a specific agent
func GetModelForAgent(agentName string) (*ModelDefinition, error) {
	log.Printf("🔍 GetModelForAgent called with: %s", agentName)
	
	if agentConfig == nil {
		return nil, fmt.Errorf("agent configuration not loaded")
	}
	
	// Normalize agent name (remove @ prefix if present)
	cleanName := strings.TrimPrefix(agentName, "@")
	cleanName = strings.Replace(cleanName, "ai-", "", 1)
	log.Printf("🔍 Normalized agent name: %s -> %s", agentName, cleanName)
	
	// Find the agent
	agent, exists := agentConfig.Agents[cleanName]
	if !exists {
		log.Printf("❌ Agent %s not found in config", cleanName)
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	log.Printf("🔍 Found agent config: Model=%s", agent.Model)
	
	// Get the model for this agent (use default if not specified)
	modelKey := agent.Model
	if modelKey == "" {
		modelKey = agentConfig.DefaultModel
		log.Printf("🔍 Using default model: %s", modelKey)
	}
	
	// Look up the model definition
	model, exists := agentConfig.Models[modelKey]
	if !exists {
		log.Printf("❌ Model %s not found in models registry", modelKey)
		return nil, fmt.Errorf("model %s not found", modelKey)
	}
	log.Printf("🔍 Found model definition: ID=%s, Name=%s", model.ID, model.Name)
	
	// Create a copy to avoid modifying the original
	modelCopy := model
	
	// Apply temperature override if specified
	if agent.TemperatureOverride != nil {
		modelCopy.Temperature = *agent.TemperatureOverride
		log.Printf("🔍 Applied temperature override: %.2f", modelCopy.Temperature)
	}
	
	log.Printf("🔍 Returning model: ID=%s, Name=%s, Temp=%.2f", modelCopy.ID, modelCopy.Name, modelCopy.Temperature)
	return &modelCopy, nil
}

// GetResponseConfig returns the response configuration
func GetResponseConfig() ResponseConfig {
	if agentConfig == nil {
		return ResponseConfig{
			MaxTokens: 4096,
			Stream:    false,
		}
	}
	return agentConfig.ResponseConfig
}