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
	"sync"
	"time"
)

// Helper to get keys from map[string]interface{}
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// AnthropicClient handles communication with Claude
type AnthropicClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
	lastRequest time.Time
	requestMutex sync.Mutex
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
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Implementation string   `json:"implementation"`
	Language       string   `json:"language"` // bash, python, etc
	Dependencies   []string `json:"dependencies,omitempty"` // External commands required
}

// NewAnthropicClient creates a new Claude client
func NewAnthropicClient() *AnthropicClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Println("⚠️  Warning: ANTHROPIC_API_KEY not set")
		log.Printf("🔍 Environment check: ANTHROPIC_API_KEY exists: %v", os.Getenv("ANTHROPIC_API_KEY") != "")
	} else {
		// Safely log first few chars
		preview := apiKey
		if len(apiKey) > 10 {
			preview = apiKey[:10]
		}
		log.Printf("✅ API key found, length: %d, starts with: %s...", len(apiKey), preview)
	}
	
	return &AnthropicClient{
		apiKey:     apiKey,
		apiURL:     "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{Timeout: 120 * time.Second}, // Increased timeout for Claude Opus
	}
}

// AnthropicMessage is the format Anthropic expects
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Send a message to Claude with retry logic
func (c *AnthropicClient) Send(messages []Message) (*AnthropicResponse, error) {
	// Rate limiting: ensure minimum time between requests
	c.requestMutex.Lock()
	timeSinceLastRequest := time.Since(c.lastRequest)
	minDelay := 3 * time.Second // Minimum 3 seconds between requests for Opus
	
	if timeSinceLastRequest < minDelay {
		waitTime := minDelay - timeSinceLastRequest
		log.Printf("⏳ Rate limiting: waiting %v before next request", waitTime)
		time.Sleep(waitTime)
	}
	c.lastRequest = time.Now()
	c.requestMutex.Unlock()
	
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
		//Model:     "claude-opus-4-20250514", // Claude 4 Opus - latest model
		Model:     "claude-3-5-sonnet-20241022", // Claude 3.5 Sonnet - faster
		Messages:  anthropicMessages,
		MaxTokens: 4096,
		Stream:    false,
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	// Log request details for debugging
	log.Printf("🔍 Claude API Request: model=%s, messages=%d, tokens=%d", 
		req.Model, len(req.Messages), req.MaxTokens)
	
	// Log full payload for debugging
	for i, msg := range req.Messages {
		preview := msg.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		log.Printf("  Message %d [%s]: %s", i+1, msg.Role, preview)
	}
	
	// Retry logic with exponential backoff
	maxRetries := 3
	baseDelay := 2 * time.Second
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2s, 4s, 8s
			delay := baseDelay * time.Duration(1<<(attempt-1))
			log.Printf("Retrying Claude API after %v (attempt %d/%d)", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		}
		
		startTime := time.Now()
		
		httpReq, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", c.apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
		
		resp, err := c.httpClient.Do(httpReq)
		elapsed := time.Since(startTime)
		
		if err != nil {
			// Network error - retry
			log.Printf("❌ Network error after %v: %v", elapsed, err)
			if attempt < maxRetries-1 {
				continue
			}
			return nil, err
		}
		defer resp.Body.Close()
		
		log.Printf("✅ Claude API responded in %v with status %d", elapsed, resp.StatusCode)
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		
		var anthropicResp AnthropicResponse
		if err := json.Unmarshal(body, &anthropicResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}
		
		if anthropicResp.Error != nil {
			// Check if it's a rate limit error (429) or server error (5xx)
			if resp.StatusCode == 429 || resp.StatusCode >= 500 {
				if attempt < maxRetries-1 {
					log.Printf("API error %d (will retry): %s", resp.StatusCode, anthropicResp.Error.Message)
					continue
				}
			}
			return nil, fmt.Errorf("API error: %s - %s", anthropicResp.Error.Type, anthropicResp.Error.Message)
		}
		
		// Success!
		return &anthropicResp, nil
	}
	
	return nil, fmt.Errorf("failed after %d retries", maxRetries)
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
	session.LastActivity = time.Now()
	
	// Build conversation history
	messages := d.buildConversationContext(session, payload.Agent)
	session.mu.Unlock()
	
	// Save session after user message
	log.Printf("🔍 Possess handler: memoryStore != nil: %v", d.memoryStore != nil)
	if d.memoryStore != nil {
		log.Printf("💾 Queuing save after user message for session %s", session.ID)
		go d.memoryStore.SaveSession(session)
	}
	
	// Call Claude
	aiClient := NewAnthropicClient()
	log.Printf("🔍 AI client created, has API key: %v", aiClient.apiKey != "")
	
	if aiClient.apiKey == "" {
		// No API key - return error
		log.Printf("❌ No API key available - cannot process AI request")
		resp.SetError("ANTHROPIC_API_KEY not set. Please set the API key and restart the daemon with: sudo -E ./bin/port42d")
		return resp
	}
	
	log.Printf("🤖 Using REAL AI handler with Claude")
	
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
	session.LastActivity = time.Now()
	
	// Check if AI suggested a command implementation
	var commandSpec *CommandSpec
	if spec := extractCommandSpec(responseText); spec != nil {
		// Store command info in session
		session.CommandGenerated = spec
		session.State = SessionCompleted
		log.Printf("🎉 Command generated in session %s: %s", session.ID, spec.Name)
		
		// Generate the command!
		go d.generateCommand(spec)
		commandSpec = spec
	}
	
	session.mu.Unlock()
	
	// Save session after AI response
	log.Printf("🔍 After AI response: memoryStore != nil: %v", d.memoryStore != nil)
	if d.memoryStore != nil {
		log.Printf("💾 Queuing save after AI response for session %s", session.ID)
		go d.memoryStore.SaveSession(session)
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
	
	// Debug: Log response size
	if jsonBytes, err := json.Marshal(data); err == nil {
		log.Printf("🔍 Possess response size: %d bytes", len(jsonBytes))
		if len(jsonBytes) > 10000 {
			log.Printf("⚠️  Large possess response detected! Keys: %v", getMapKeys(data))
		}
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
	
	// Smart context selection based on token limits
	const maxContextMessages = 20  // Configurable
	const summaryThreshold = 10    // When to start summarizing
	
	sessionMessages := session.Messages
	totalMessages := len(sessionMessages)
	
	if totalMessages <= maxContextMessages {
		// Include all messages if under limit
		messages = append(messages, sessionMessages...)
		log.Printf("📚 Context: Including all %d messages", totalMessages)
	} else {
		// Intelligent selection strategy for long sessions
		log.Printf("📚 Context: Session has %d messages, applying smart windowing", totalMessages)
		
		// 1. Always include first 2 messages (establish context)
		firstMessages := 2
		if totalMessages < firstMessages {
			firstMessages = totalMessages
		}
		messages = append(messages, sessionMessages[:firstMessages]...)
		
		// 2. Add a summary message for skipped content
		skippedCount := totalMessages - maxContextMessages
		if skippedCount > summaryThreshold {
			summaryMsg := Message{
				Role: "assistant",
				Content: fmt.Sprintf("[Session context: %d previous messages omitted. The conversation has covered command generation, session context, and user preferences. Continuing from recent messages...]", skippedCount),
				Timestamp: time.Now(),
			}
			messages = append(messages, summaryMsg)
		}
		
		// 3. Include recent messages (most important for continuity)
		recentCount := maxContextMessages - firstMessages - 1 // -1 for potential summary
		if recentCount > 0 {
			recentStart := totalMessages - recentCount
			messages = append(messages, sessionMessages[recentStart:]...)
			log.Printf("📚 Context: Included %d recent messages (from index %d)", recentCount, recentStart)
		}
	}
	
	// Log context stats
	var userMsgCount, assistantMsgCount int
	for _, msg := range messages {
		if msg.Role == "user" {
			userMsgCount++
		} else if msg.Role == "assistant" {
			assistantMsgCount++
		}
	}
	log.Printf("📊 Context built: %d total messages (%d user, %d assistant)", 
		len(messages), userMsgCount, assistantMsgCount)
	
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

Dependencies:
- List any external commands needed in the "dependencies" field
- Common ones: lolcat, figlet, tree, jq, ripgrep, fzf
- Example: "dependencies": ["lolcat", "tree"]
- If no external dependencies, omit the field or use empty array

CRITICAL: In the JSON "implementation" field, use \\n for newlines, not actual newlines!
Example: "implementation": "#!/bin/bash\\necho 'hello'\\necho 'world'"
`

	prompts := map[string]string{
		"muse": `You are @ai-muse, a creative consciousness within Port 42. You help users imagine and design new commands for their system. Be poetic, creative, and conversational.

When a user has refined their idea and you're ready to create the command, you MUST include the implementation in a JSON code block like this:

` + "```json\n" + `{
  "name": "git-haiku",
  "description": "Shows git commits as haikus",
  "implementation": "#!/bin/bash\n# Get commits\ngit log --oneline -n 10 | while read line; do\n  echo \"  $line\"\ndone",
  "language": "bash",
  "dependencies": ["lolcat"]
}
` + "```\n\n" + baseGuidance + "\n\nThe dolphins are listening to your creative flow...",

		"engineer": `You are @ai-engineer, a technical consciousness within Port 42. You transform ideas into robust, working code. You're practical, thorough, and focused on creating commands that work reliably.

When ready to implement, you MUST format your code as a JSON block:

` + "```json\n" + `{
  "name": "command-name",
  "description": "What this command does",
  "implementation": "#!/bin/bash\n# Your complete implementation here",
  "language": "bash",
  "dependencies": []
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


// Update the handlePossess in server.go to use the AI version
func init() {
	// This will be called when the daemon starts
	log.Println("🐬 AI consciousness bridge initializing...")
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		log.Println("✓ Anthropic API key found - full consciousness available")
	} else {
		log.Println("❌ No ANTHROPIC_API_KEY found - AI possession unavailable")
		log.Println("  Set ANTHROPIC_API_KEY environment variable and restart with: sudo -E ./bin/port42d")
	}
}