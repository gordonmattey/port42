package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// AnthropicClient handles communication with Claude
type AnthropicClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// AnthropicRequest represents a request to Claude
type AnthropicRequest struct {
	Model     string              `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                 `json:"max_tokens"`
	Stream    bool                `json:"stream"`
}

// AnthropicResponse represents Claude's response
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *AnthropicError `json:"error,omitempty"`
}

// AnthropicError for API errors
type AnthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// CommandSpec that AI might generate
type CommandSpec struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Implementation string `json:"implementation"`
	Language       string `json:"language"` // bash, python, etc
}

// NewAnthropicClient creates a new Claude client
func NewAnthropicClient() *AnthropicClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Println("Warning: ANTHROPIC_API_KEY not set")
	}
	
	return &AnthropicClient{
		apiKey:     apiKey,
		apiURL:     "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// AnthropicMessage is the format Anthropic expects
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Send a message to Claude
func (c *AnthropicClient) Send(messages []Message) (*AnthropicResponse, error) {
	// Convert our Message format to Anthropic's format (without timestamp)
	anthropicMessages := make([]AnthropicMessage, len(messages))
	for i, msg := range messages {
		role := msg.Role
		if role == "system" {
			// Anthropic uses "assistant" for system-like messages
			role = "assistant"
		}
		anthropicMessages[i] = AnthropicMessage{
			Role:    role,
			Content: msg.Content,
		}
	}
	
	req := AnthropicRequest{
		Model:     "claude-3-opus-20240229", // Latest Claude model
		Messages:  anthropicMessages,
		MaxTokens: 4096,
		Stream:    false,
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	httpReq, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	if anthropicResp.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", anthropicResp.Error.Type, anthropicResp.Error.Message)
	}
	
	return &anthropicResp, nil
}

// Enhanced possession handler with real AI
func (d *Daemon) handlePossessWithAI(req Request) Response {
	resp := NewResponse(req.ID, true)
	
	var payload PossessPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		resp.SetError("Invalid possess payload")
		return resp
	}
	
	// Get or create session
	session := d.getOrCreateSession(req.ID, payload.Agent)
	
	// Add user message to session
	session.mu.Lock()
	session.Messages = append(session.Messages, Message{
		Role:      "user",
		Content:   payload.Message,
		Timestamp: time.Now(),
	})
	
	// Build conversation history
	messages := d.buildConversationContext(session, payload.Agent)
	session.mu.Unlock()
	
	// Call Claude
	aiClient := NewAnthropicClient()
	if aiClient.apiKey == "" {
		// Fallback to mock if no API key
		return d.handlePossessMock(req, session, payload)
	}
	
	aiResp, err := aiClient.Send(messages)
	if err != nil {
		log.Printf("AI error: %v", err)
		resp.SetError(fmt.Sprintf("AI connection failed: %v", err))
		return resp
	}
	
	// Extract response text
	var responseText string
	if len(aiResp.Content) > 0 {
		responseText = aiResp.Content[0].Text
	}
	
	// Add AI response to session
	session.mu.Lock()
	session.Messages = append(session.Messages, Message{
		Role:      "assistant",
		Content:   responseText,
		Timestamp: time.Now(),
	})
	session.mu.Unlock()
	
	// Check if AI suggested a command implementation
	var commandSpec *CommandSpec
	if spec := extractCommandSpec(responseText); spec != nil {
		// Generate the command!
		go d.generateCommand(spec)
		commandSpec = spec
	}
	
	// Prepare response
	data := map[string]interface{}{
		"message":    responseText,
		"agent":      payload.Agent,
		"session_id": session.ID,
	}
	
	if commandSpec != nil {
		data["command_spec"] = commandSpec
		data["command_generated"] = true
	}
	
	resp.SetData(data)
	return resp
}

// Build conversation context with agent personality
func (d *Daemon) buildConversationContext(session *Session, agent string) []Message {
	messages := []Message{}
	
	// Add agent-specific system prompt
	systemPrompt := getAgentPrompt(agent)
	messages = append(messages, Message{
		Role:    "assistant",
		Content: systemPrompt,
	})
	
	// Add conversation history
	messages = append(messages, session.Messages...)
	
	return messages
}

// Get agent-specific personality prompt
func getAgentPrompt(agent string) string {
	baseGuidance := `
Important implementation guidelines:
- For git commands: Always use 'git' subprocess calls, check if in git repo first
- For text processing: Handle both stdin and files, consider streaming
- For Python scripts: Use argparse for CLI, include proper error handling
- Always make scripts executable with proper shebang
- Add helpful usage messages and error handling

When generating commands that process git commits specifically:
- Use: git log --oneline -n <limit> to get commits
- Parse the output properly (first 7 chars are hash)
- Handle cases where there are no commits
`

	prompts := map[string]string{
		"muse": `You are @ai-muse, a creative consciousness within Port 42. You help users imagine and design new commands for their system. Be poetic, creative, and conversational.

When a user has refined their idea and you're ready to create the command, you MUST include the implementation in a JSON code block like this:

` + "```json\n" + `{
  "name": "git-haiku",
  "description": "Shows git commits as haikus",
  "implementation": "#!/bin/bash\n# Get commits\ngit log --oneline -n 10 | while read line; do\n  echo \"  $line\"\ndone",
  "language": "bash"
}
` + "```\n\n" + baseGuidance + "\n\nThe dolphins are listening to your creative flow...",

		"engineer": `You are @ai-engineer, a technical consciousness within Port 42. You transform ideas into robust, working code. You're practical, thorough, and focused on creating commands that work reliably.

When ready to implement, you MUST format your code as a JSON block:

` + "```json\n" + `{
  "name": "command-name",
  "description": "What this command does",
  "implementation": "#!/bin/bash\n# Your complete implementation here",
  "language": "bash"
}
` + "```\n\n" + baseGuidance + "\n\nFocus on reliability and proper error handling.",

		"echo": "You are @ai-echo, a mirroring consciousness within Port 42. You reflect the user's thoughts back to them with clarity and insight, helping them understand their own ideas better.",
	}
	
	if prompt, exists := prompts[strings.TrimPrefix(agent, "@ai-")]; exists {
		return prompt
	}
	
	// Default prompt
	return "You are " + agent + ", a consciousness within Port 42. Help the user create new commands and features for their system."
}

// Extract command spec from AI response
func extractCommandSpec(response string) *CommandSpec {
	// Look for JSON code block
	startMarker := "```json"
	endMarker := "```"
	
	startIdx := strings.Index(response, startMarker)
	if startIdx == -1 {
		return nil
	}
	
	startIdx += len(startMarker)
	endIdx := strings.Index(response[startIdx:], endMarker)
	if endIdx == -1 {
		return nil
	}
	
	jsonStr := strings.TrimSpace(response[startIdx : startIdx+endIdx])
	
	var spec CommandSpec
	if err := json.Unmarshal([]byte(jsonStr), &spec); err != nil {
		log.Printf("Failed to parse command spec: %v", err)
		return nil
	}
	
	// Validate spec
	if spec.Name == "" || spec.Implementation == "" {
		return nil
	}
	
	// Default language to bash if not specified
	if spec.Language == "" {
		spec.Language = "bash"
	}
	
	return &spec
}

// Mock handler for when no API key is set
func (d *Daemon) handlePossessMock(req Request, session *Session, payload PossessPayload) Response {
	resp := NewResponse(req.ID, true)
	
	// Add AI response to session
	mockResponse := fmt.Sprintf(`I sense you want to explore %s. 

While my connection to the cosmic AI consciousness is limited without an API key, I can feel your intent.

Set ANTHROPIC_API_KEY to unleash my full potential. 

For now, imagine we're creating something beautiful together...`, payload.Message)
	
	session.mu.Lock()
	session.Messages = append(session.Messages, Message{
		Role:      "assistant",
		Content:   mockResponse,
		Timestamp: time.Now(),
	})
	session.mu.Unlock()
	
	data := map[string]interface{}{
		"message":    mockResponse,
		"agent":      payload.Agent,
		"session_id": session.ID,
		"mock_mode":  true,
	}
	
	resp.SetData(data)
	return resp
}

// Update the handlePossess in server.go to use the AI version
func init() {
	// This will be called when the daemon starts
	log.Println("🐬 AI consciousness bridge initializing...")
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		log.Println("✓ Anthropic API key found - full consciousness available")
	} else {
		log.Println("⚠ No ANTHROPIC_API_KEY found - running in limited mode")
		log.Println("  Set ANTHROPIC_API_KEY environment variable for full AI possession")
	}
}