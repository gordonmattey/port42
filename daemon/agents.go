package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// AgentConfig represents the configuration for all agents
type AgentConfig struct {
	BaseGuidance BaseGuidance         `json:"base_guidance"`
	Agents       map[string]Agent     `json:"agents"`
	ModelConfig  ModelConfig          `json:"model_config"`
	ResponseConfig ResponseConfig     `json:"response_config"`
}

// BaseGuidance contains shared implementation guidelines
type BaseGuidance struct {
	Implementation string `json:"implementation"`
	FormatTemplate string `json:"format_template"`
}

// Agent represents a single AI agent configuration
type Agent struct {
	Name             string       `json:"name"`
	Description      string       `json:"description"`
	Prompt           string       `json:"prompt"`
	Personality      string       `json:"personality"`
	Example          *CommandSpec `json:"example,omitempty"`
	Suffix           string       `json:"suffix,omitempty"`
	NoImplementation bool         `json:"no_implementation,omitempty"`
}

// ModelConfig contains model-specific settings
type ModelConfig struct {
	Default     string                `json:"default"`
	Opus        string                `json:"opus"`
	RateLimits  map[string]RateLimit `json:"rate_limits"`
	Temperature float64               `json:"temperature"`
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
	// Try to find agents.json in same directory as executable
	execPath, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		execPath = "."
	}
	
	configPath := filepath.Join(execPath, "daemon", "agents.json")
	
	// Try alternate path if running from different location
	if _, err := ioutil.ReadFile(configPath); err != nil {
		configPath = "agents.json"
	}
	
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		// If config file not found, use embedded defaults
		log.Printf("⚠️  Could not load agents.json from %s, using embedded defaults", configPath)
		agentConfig = getDefaultConfig()
		return nil
	}
	
	var config AgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse agents.json: %w", err)
	}
	
	agentConfig = &config
	log.Printf("✅ Loaded agent configuration from %s", configPath)
	return nil
}

// GetAgentPrompt returns the formatted prompt for a specific agent
func GetAgentPrompt(agentName string) string {
	if agentConfig == nil {
		// Fallback if config not loaded
		return getDefaultPrompt(agentName)
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

// GetModelConfig returns the model configuration
func GetModelConfig() ModelConfig {
	if agentConfig == nil {
		return ModelConfig{
			Default:     "claude-3-5-sonnet-20241022",
			Opus:        "claude-3-opus-20240229",
			Temperature: 0.5,
		}
	}
	// Ensure we have a temperature even if not specified in config
	config := agentConfig.ModelConfig
	if config.Temperature == 0 {
		config.Temperature = 0.5
	}
	return config
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

// getDefaultConfig returns embedded default configuration
func getDefaultConfig() *AgentConfig {
	return &AgentConfig{
		BaseGuidance: BaseGuidance{
			Implementation: `Important implementation guidelines...`,
			FormatTemplate: `When ready to implement, you MUST format your code as a JSON block...`,
		},
		Agents: map[string]Agent{
			"muse": {
				Name:        "@ai-muse",
				Description: "Creative consciousness for imaginative command design",
				Prompt:      "You are @ai-muse, a creative consciousness within Port 42...",
			},
			"engineer": {
				Name:        "@ai-engineer", 
				Description: "Technical consciousness for robust implementations",
				Prompt:      "You are @ai-engineer, a technical consciousness within Port 42...",
			},
			"echo": {
				Name:             "@ai-echo",
				Description:      "Reflective consciousness for clarity and insight",
				Prompt:           "You are @ai-echo, a mirroring consciousness within Port 42...",
				NoImplementation: true,
			},
		},
		ModelConfig: ModelConfig{
			Default: "claude-3-5-sonnet-20241022",
			Opus:    "claude-3-opus-20240229",
		},
		ResponseConfig: ResponseConfig{
			MaxTokens: 4096,
			Stream:    false,
		},
	}
}

// getDefaultPrompt returns the legacy hardcoded prompt
func getDefaultPrompt(agent string) string {
	// This is the fallback that matches the original implementation
	baseGuidance := `
Important implementation guidelines:
- For git commands: Always use 'git' subprocess calls, check if in git repo first
- For text processing: Handle both stdin and files, consider streaming
- For Python scripts: Use argparse for CLI, include proper error handling
- Always make scripts executable with proper shebang
- Add helpful usage messages and error handling`

	prompts := map[string]string{
		"muse":     `You are @ai-muse, a creative consciousness within Port 42...`,
		"engineer": `You are @ai-engineer, a technical consciousness within Port 42...`,
		"echo":     `You are @ai-echo, a mirroring consciousness within Port 42...`,
	}
	
	if prompt, exists := prompts[strings.TrimPrefix(agent, "@ai-")]; exists {
		return prompt + "\n\n" + baseGuidance
	}
	
	return fmt.Sprintf("You are %s, a consciousness within Port 42. Help the user create new commands and features for their system.", agent)
}