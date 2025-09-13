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
	GuidanceFile   string                    `json:"guidance_file"`
	BaseTemplate   string                    `json:"base_template"`
	Agents         map[string]Agent          `json:"agents"`
	DefaultModel   string                    `json:"default_model"`
	ResponseConfig ResponseConfig            `json:"response_config"`
	LoadedGuidance string                    // Loaded from guidance_file
}


// CommandMetadata represents basic info about a Port 42 command
type CommandMetadata struct {
	Name        string
	Description string
}

// Agent represents a single AI agent configuration
type Agent struct {
	Name                string   `json:"name"`
	Model               string   `json:"model"`
	TemperatureOverride *float64 `json:"temperature_override,omitempty"`
	Description         string   `json:"description"`
	Personality         string   `json:"personality"`
	Style               string   `json:"style"`
	CustomPrompt        string   `json:"custom_prompt,omitempty"`
	Suffix              string   `json:"suffix,omitempty"`
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
		// Installed paths (check these FIRST for production use)
		filepath.Join(homeDir, ".port42", "agents.json"),
		"/etc/port42/agents.json",
		// Relative to executable
		func() string {
			if execPath, err := os.Executable(); err == nil {
				return filepath.Join(filepath.Dir(execPath), "agents.json")
			}
			return ""
		}(),
		// Development paths (check these LAST)
		"daemon/agents.json",
		"agents.json",
		"../daemon/agents.json",
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
	
	// Load the guidance file if specified
	if config.GuidanceFile != "" {
		guidancePath := filepath.Join(filepath.Dir(foundPath), config.GuidanceFile)
		guidanceData, err := ioutil.ReadFile(guidancePath)
		if err != nil {
			return fmt.Errorf("failed to read guidance file %s: %w", guidancePath, err)
		}
		config.LoadedGuidance = string(guidanceData)
		log.Printf("✅ Loaded guidance from %s", guidancePath)
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
		return fmt.Sprintf("You are %s, swimming in Port 42's stream. Help the user create new commands and features for their system.", agentName)
	}
	
	// Build the base prompt using the template from configuration
	var prompt strings.Builder
	
	// 1. Base template with agent details and loaded guidance
	baseTemplate := agentConfig.BaseTemplate
	baseTemplate = strings.ReplaceAll(baseTemplate, "{name}", agent.Name)
	baseTemplate = strings.ReplaceAll(baseTemplate, "{personality}", agent.Personality)
	baseTemplate = strings.ReplaceAll(baseTemplate, "{style}", agent.Style)
	baseTemplate = strings.ReplaceAll(baseTemplate, "{guidance}", agentConfig.LoadedGuidance)
	prompt.WriteString(baseTemplate)
	
	// 2. Custom prompt if exists
	if agent.CustomPrompt != "" {
		prompt.WriteString("\n\n<role_details>\n")
		prompt.WriteString(agent.CustomPrompt)
		prompt.WriteString("\n</role_details>")
	}
	
	// 3. Available commands list (ALL agents)
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
	
	// 4. Agent-specific suffix
	if agent.Suffix != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agent.Suffix)
	}
	
	// Debug logging
	log.Printf("🔍 Building prompt for %s, personality: %s, style: %s", 
		agentName, agent.Personality, agent.Style)
	
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