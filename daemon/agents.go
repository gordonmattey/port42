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
	Implementation   string `json:"implementation"`
	FormatTemplate   string `json:"format_template"`
	ArtifactGuidance string `json:"artifact_guidance"`
}

// Agent represents a single AI agent configuration
type Agent struct {
	Name                string       `json:"name"`
	Model               string       `json:"model"`
	TemperatureOverride *float64     `json:"temperature_override,omitempty"`
	Description         string       `json:"description"`
	Prompt              string       `json:"prompt"`
	Personality         string       `json:"personality"`
	Example             *CommandSpec `json:"example,omitempty"`
	Suffix              string       `json:"suffix,omitempty"`
	NoImplementation    bool         `json:"no_implementation,omitempty"`
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
	
	// Build the full prompt
	var prompt strings.Builder
	prompt.WriteString(agent.Prompt)
	
	// Add artifact guidance for all agents
	prompt.WriteString("\n\n")
	prompt.WriteString(agentConfig.BaseGuidance.ArtifactGuidance)
	
	// Debug log
	log.Printf("🔍 Building prompt for %s, artifact guidance length: %d", 
		agentName, len(agentConfig.BaseGuidance.ArtifactGuidance))
	
	// Add implementation guidance if agent creates commands
	if !agent.NoImplementation {
		prompt.WriteString("\n\n")
		prompt.WriteString(agentConfig.BaseGuidance.FormatTemplate)
		
		// Add example if provided
		if agent.Example != nil {
			prompt.WriteString("\n\nExample:\n```json\n")
			exampleJSON, _ := json.MarshalIndent(agent.Example, "", "  ")
			prompt.WriteString(string(exampleJSON))
			prompt.WriteString("\n```")
		}
		
		prompt.WriteString("\n\n")
		prompt.WriteString(agentConfig.BaseGuidance.Implementation)
	}
	
	// Add agent-specific suffix
	if agent.Suffix != "" {
		prompt.WriteString("\n\n")
		prompt.WriteString(agent.Suffix)
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