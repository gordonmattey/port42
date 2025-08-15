package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	Model       string              `json:"model"`
	System      string              `json:"system,omitempty"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Stream      bool                `json:"stream"`
	Temperature float64             `json:"temperature,omitempty"`
	Tools       []AnthropicTool     `json:"tools,omitempty"`
}

// AnthropicTool represents a tool (function) definition
type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// AnthropicResponse represents Claude's response
type AnthropicResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text,omitempty"`
		Name  string          `json:"name,omitempty"`
		Input json.RawMessage `json:"input,omitempty"`
	} `json:"content"`
	Error      *AnthropicError `json:"error,omitempty"`
	StopReason string          `json:"stop_reason,omitempty"`
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
	SessionID      string   `json:"session_id,omitempty"` // Session that created this
	Agent          string   `json:"agent,omitempty"` // Agent that created this
}

// ArtifactSpec that AI might generate
type ArtifactSpec struct {
	Name        string                 `json:"name"`        // Filename or directory name
	Type        string                 `json:"type"`        // document, code, design, media
	Description string                 `json:"description"` // What this artifact is
	Content     map[string]string      `json:"content"`     // For multi-file artifacts (path -> content)
	SingleFile  string                 `json:"single_file,omitempty"` // For single file content
	Format      string                 `json:"format"`      // md, html, py, js, svg, etc
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // Additional metadata
	SessionID   string                 `json:"session_id,omitempty"` // Session that created this
	Agent       string                 `json:"agent,omitempty"` // Agent that created this
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
		httpClient: &http.Client{Timeout: 180 * time.Second}, // Increased timeout for Claude Opus (3 minutes)
	}
}

// AnthropicMessage is the format Anthropic expects
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Send a message to Claude with retry logic
func (c *AnthropicClient) Send(messages []Message, systemPrompt string, agentName string) (*AnthropicResponse, error) {
	// Get model configuration for this agent
	modelDef, err := GetModelForAgent(agentName)
	if err != nil {
		log.Printf("⚠️ Failed to get model for agent %s: %v", agentName, err)
		return nil, err
	}
	responseConfig := GetResponseConfig()
	
	// Rate limiting: ensure minimum time between requests
	c.requestMutex.Lock()
	timeSinceLastRequest := time.Since(c.lastRequest)
	
	minDelay := time.Duration(modelDef.RateLimit.MinDelaySeconds) * time.Second
	if minDelay == 0 {
		minDelay = 1 * time.Second // fallback
	}
	
	if timeSinceLastRequest < minDelay {
		waitTime := minDelay - timeSinceLastRequest
		log.Printf("⏳ Rate limiting: waiting %v before next request", waitTime)
		time.Sleep(waitTime)
	}
	c.lastRequest = time.Now()
	c.requestMutex.Unlock()
	
	// Convert our Message format to Anthropic's format (without timestamp)
	// Skip any system messages as they'll be in the system parameter
	anthropicMessages := []AnthropicMessage{}
	for _, msg := range messages {
		if msg.Role == "system" {
			continue // Skip system messages
		}
		anthropicMessages = append(anthropicMessages, AnthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	
	// Check if this agent should use tools
	var tools []AnthropicTool
	
	// Get agent config
	cleanName := strings.TrimPrefix(agentName, "@ai-")
	cleanName = strings.TrimPrefix(cleanName, "@")
	log.Printf("🔍 Checking tools for agent: %s (clean: %s), config exists: %v", agentName, cleanName, agentConfig != nil)
	
	if agentConfig != nil {
		if agentInfo, exists := agentConfig.Agents[cleanName]; exists {
			log.Printf("🔍 Agent %s found, NoImplementation: %v", cleanName, agentInfo.NoImplementation)
			if agentInfo.NoImplementation {
				// Agent marked as no implementation - gets command runner and artifact tool
				tools = []AnthropicTool{
					getCommandRunnerTool(),
					getArtifactGenerationTool(),
				}
				log.Printf("🎨 Agent %s will use command runner and artifact generation", agentName)
			} else {
				// Full implementation agent - gets all tools
				tools = []AnthropicTool{
					getCommandRunnerTool(),
					getCommandGenerationTool(),
					getArtifactGenerationTool(),
				}
				log.Printf("🔧 Agent %s will use all tools (command runner, generation, and artifacts)", agentName)
			}
		} else {
			log.Printf("⚠️ Agent %s not found in config", cleanName)
		}
	} else {
		log.Printf("⚠️ Agent config is nil!")
	}
	
	// Debug log the model
	log.Printf("🔍 Using model for agent %s: ID=%s, Name=%s, Temp=%.2f", 
		agentName, modelDef.ID, modelDef.Name, modelDef.Temperature)
	
	req := AnthropicRequest{
		Model:       modelDef.ID,
		System:      systemPrompt,
		Messages:    anthropicMessages,
		MaxTokens:   responseConfig.MaxTokens,
		Stream:      responseConfig.Stream,
		Temperature: modelDef.Temperature,
		Tools:       tools,
	}
	
	log.Printf("🔧 Sending request with %d tools to API", len(tools))
	
	// Log the actual tools being sent
	for i, tool := range tools {
		log.Printf("  Tool %d: %s - %s", i+1, tool.Name, tool.Description)
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	// Log request details for debugging
	log.Printf("🔍 Claude API Request: model=%s, messages=%d, tokens=%d, temp=%.2f", 
		req.Model, len(req.Messages), req.MaxTokens, req.Temperature)
	
	// Log system prompt
	if req.System != "" {
		// Log if prompt contains XML tags
		hasXMLTags := strings.Contains(req.System, "<tool_instructions>")
		log.Printf("  System prompt contains XML tags: %v", hasXMLTags)
		
		// Show more of the prompt for debugging
		systemPreview := req.System
		if len(systemPreview) > 500 {
			systemPreview = systemPreview[:500] + "..."
		}
		log.Printf("  System prompt preview: %s", systemPreview)
	}
	
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
	
	// Get or create session - use session_id from payload if provided, otherwise use request ID
	sessionID := req.ID
	if payload.SessionID != "" {
		sessionID = payload.SessionID
	}
	session := d.getOrCreateSession(sessionID, payload.Agent)
	log.Printf("🔍 Session loaded: ID=%s, MessageCount=%d", session.ID, len(session.Messages))
	
	// Add user message to session
	session.mu.Lock()
	session.Messages = append(session.Messages, Message{
		Role:      "user",
		Content:   payload.Message,
		Timestamp: time.Now(),
	})
	session.LastActivity = time.Now()
	
	// Get agent prompt
	agentPrompt := getAgentPrompt(payload.Agent)
	
	// Inject memory contexts into system prompt if provided
	if len(payload.MemoryContext) > 0 {
		log.Printf("🧠 Injecting %d memory contexts into system prompt", len(payload.MemoryContext))
		memorySection := "\n\n--- RELEVANT MEMORY CONTEXTS ---\n"
		memorySection += "The following are relevant memories from previous conversations. Reference them when helpful:\n\n"
		
		for i, context := range payload.MemoryContext {
			memorySection += fmt.Sprintf("Memory Context %d:\n%s\n\n", i+1, context)
		}
		
		memorySection += "--- END MEMORY CONTEXTS ---\n"
		agentPrompt = agentPrompt + memorySection
	}
	
	// Build conversation history (without system prompt)
	messages := d.buildConversationContext(session, payload.Agent)
	session.mu.Unlock()
	
	// Save session after user message
	log.Printf("🔍 Possess handler: memoryStore != nil: %v", d.storage != nil)
	if d.storage != nil {
		log.Printf("🔍 [POSSESSION] Saving session after user message (messages=%d)", len(session.Messages))
		go d.storage.SaveSession(session)
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
	
	log.Printf("🔍 Sending to AI with %d messages in context", len(messages))
	aiResp, err := aiClient.Send(messages, agentPrompt, payload.Agent)
	if err != nil {
		log.Printf("AI error: %v", err)
		resp.SetError(fmt.Sprintf("AI connection failed: %v", err))
		return resp
	}
	log.Printf("🔍 Got AI response")
	
	// Extract response text and check for tool calls
	var responseText string
	var commandSpec *CommandSpec
	var artifactSpec *ArtifactSpec
	
	if len(aiResp.Content) > 0 {
		// Check if response contains tool calls
		hasToolCall := false
		for _, content := range aiResp.Content {
			if content.Type == "tool_use" && content.Name == "generate_command" {
				hasToolCall = true
				// Extract command spec from tool call
				if spec, err := extractCommandSpecFromToolCall(content.Input); err == nil {
					commandSpec = spec
					log.Printf("🔧 Extracted command spec from tool call: %s", spec.Name)
				} else {
					log.Printf("❌ Failed to extract command spec from tool call: %v", err)
				}
			} else if content.Type == "tool_use" && content.Name == "generate_artifact" {
				hasToolCall = true
				// Extract artifact spec from tool call
				if spec, err := extractArtifactSpecFromToolCall(content.Input); err == nil {
					artifactSpec = spec
					log.Printf("🎨 Extracted artifact spec from tool call: %s", spec.Name)
				} else {
					log.Printf("❌ Failed to extract artifact spec from tool call: %v", err)
				}
			} else if content.Type == "tool_use" && content.Name == "run_command" {
				hasToolCall = true
				// Execute command and capture output
				log.Printf("🏃 AI is executing a Port 42 command")
				if output, err := executeCommand(content.Input); err != nil {
					// Include error in response
					responseText += fmt.Sprintf("\n\n❌ Command error: %v", err)
					log.Printf("❌ Command execution failed: %v", err)
				} else {
					// Include successful output in response
					responseText += fmt.Sprintf("\n\n📟 Command output:\n%s", output)
					log.Printf("✅ Command executed successfully")
				}
			} else if content.Type == "text" {
				responseText = content.Text
			}
		}
		
		// If no tool call, try to extract from text (backward compatibility)
		if !hasToolCall && responseText != "" {
			if spec := extractCommandSpec(responseText); spec != nil {
				commandSpec = spec
				log.Printf("📄 Extracted command spec from text response: %s", spec.Name)
			}
		}
	}
	
	// Add AI response to session
	session.mu.Lock()
	session.Messages = append(session.Messages, Message{
		Role:      "assistant",
		Content:   responseText,
		Timestamp: time.Now(),
	})
	session.LastActivity = time.Now()
	
	// Check if we have a command spec to generate
	if commandSpec != nil {
		// Add session and agent info to command spec
		commandSpec.SessionID = session.ID
		commandSpec.Agent = session.Agent
		
		// Store command info in session
		session.CommandGenerated = commandSpec
		session.State = SessionCompleted
		log.Printf("🎉 Command generated in session %s: %s", session.ID, commandSpec.Name)
		
		// Generate the command!
		go d.generateCommand(commandSpec)
	}
	
	// Check if we have an artifact spec to generate
	if artifactSpec != nil {
		// Add session and agent info to artifact spec
		artifactSpec.SessionID = session.ID
		artifactSpec.Agent = session.Agent
		
		// Note: We don't have an ArtifactGenerated field in Session yet
		// For now, just log it
		log.Printf("🎨 Artifact generated in session %s: %s", session.ID, artifactSpec.Name)
		
		// Generate the artifact asynchronously (like commands)
		go func() {
			if err := d.generateArtifact(artifactSpec); err != nil {
				log.Printf("❌ Failed to generate artifact: %v", err)
			}
		}()
	}
	
	session.mu.Unlock()
	
	// Save session after AI response
	log.Printf("🔍 After AI response: memoryStore != nil: %v", d.storage != nil)
	if d.storage != nil {
		log.Printf("🔍 [POSSESSION] Saving session after AI response (messages=%d, command=%v)", 
			len(session.Messages), session.CommandGenerated != nil)
		go d.storage.SaveSession(session)
	}
	
	// Prepare response
	data := map[string]interface{}{
		"message":    responseText,
		"agent":      payload.Agent,
		"session_id": session.ID,
	}
	
	if commandSpec != nil {
		// Only send essential info, not full implementation
		data["command_spec"] = map[string]interface{}{
			"name":        commandSpec.Name,
			"description": commandSpec.Description,
			"language":    commandSpec.Language,
		}
		data["command_generated"] = true
	}
	
	if artifactSpec != nil {
		// Build the full path with extension
		artifactPath := fmt.Sprintf("/artifacts/%s/%s", artifactSpec.Type, artifactSpec.Name)
		if artifactSpec.Format != "" && !strings.Contains(artifactSpec.Name, ".") {
			artifactPath = fmt.Sprintf("%s.%s", artifactPath, artifactSpec.Format)
		}
		
		// Only send essential info, not full content
		data["artifact_spec"] = map[string]interface{}{
			"name":        artifactSpec.Name,
			"type":        artifactSpec.Type,
			"description": artifactSpec.Description,
			"format":      artifactSpec.Format,
			"path":        artifactPath,  // Include the full path with extension
		}
		data["artifact_generated"] = true
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

// Build conversation context (without system prompt which is now separate)
func (d *Daemon) buildConversationContext(session *Session, agent string) []Message {
	messages := []Message{}
	
	// Smart context selection based on token limits
	responseConfig := GetResponseConfig()
	maxContextMessages := responseConfig.ContextWindow.MaxMessages
	if maxContextMessages == 0 {
		maxContextMessages = 20 // fallback
	}
	summaryThreshold := 10 // When to start summarizing
	
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
	// Use the new configuration system
	return GetAgentPrompt(agent)
}

// getCommandGenerationTool returns the tool definition for command generation
func getCommandGenerationTool() AnthropicTool {
	return AnthropicTool{
		Name:        "generate_command",
		Description: "Generate a Port 42 command implementation",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Command name (lowercase, hyphens allowed)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "What this command does",
				},
				"implementation": map[string]interface{}{
					"type":        "string",
					"description": "Complete implementation code WITHOUT shebang",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"bash", "python", "javascript"},
					"description": "Programming language",
				},
				"dependencies": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "External commands required (e.g., git, jq)",
				},
			},
			"required": []string{"name", "description", "implementation", "language"},
		},
	}
}

// getArtifactGenerationTool returns the tool definition for artifact generation
func getArtifactGenerationTool() AnthropicTool {
	return AnthropicTool{
		Name:        "generate_artifact",
		Description: "Generate a Port 42 artifact (document, code project, design, media)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Artifact name (e.g., 'pitch-deck', 'dashboard-app', 'logo-concepts')",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"document", "code", "design", "media"},
					"description": "Type of artifact",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "What this artifact is",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "File format (md, html, py, js, svg, pdf, etc)",
				},
				"content": map[string]interface{}{
					"type":        "object",
					"description": "For multi-file artifacts: map of filepath to content",
					"additionalProperties": map[string]interface{}{
						"type": "string",
					},
				},
				"single_file": map[string]interface{}{
					"type":        "string",
					"description": "For single file artifacts: the file content",
				},
				"metadata": map[string]interface{}{
					"type":        "object",
					"description": "Additional metadata",
					"additionalProperties": true,
				},
			},
			"required": []string{"name", "type", "description", "format"},
		},
	}
}

// getCommandRunnerTool returns the tool definition for running Port 42 commands
func getCommandRunnerTool() AnthropicTool {
	return AnthropicTool{
		Name:        "run_command",
		Description: "Run a previously generated Port 42 command",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Command name (e.g., rainbow-art, git-haiku)",
				},
				"args": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Command arguments",
				},
				"stdin": map[string]interface{}{
					"type":        "string",
					"description": "Optional input to pipe to the command",
				},
			},
			"required": []string{"command"},
		},
	}
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

// extractCommandSpecFromToolCall extracts command spec from tool call response
func extractCommandSpecFromToolCall(toolInput json.RawMessage) (*CommandSpec, error) {
	var spec CommandSpec
	if err := json.Unmarshal(toolInput, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse tool input: %v", err)
	}
	
	// Validate spec
	if spec.Name == "" || spec.Implementation == "" {
		return nil, fmt.Errorf("invalid command spec: missing required fields")
	}
	
	// Default language to bash if not specified
	if spec.Language == "" {
		spec.Language = "bash"
	}
	
	return &spec, nil
}

// extractArtifactSpecFromToolCall extracts artifact spec from tool call response
func extractArtifactSpecFromToolCall(toolInput json.RawMessage) (*ArtifactSpec, error) {
	var spec ArtifactSpec
	if err := json.Unmarshal(toolInput, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse tool input: %v", err)
	}
	
	// Validate spec
	if spec.Name == "" || spec.Type == "" {
		return nil, fmt.Errorf("invalid artifact spec: missing required fields")
	}
	
	// Must have either content map or single_file
	if spec.Content == nil && spec.SingleFile == "" {
		return nil, fmt.Errorf("invalid artifact spec: must have content or single_file")
	}
	
	return &spec, nil
}

// executeCommand safely executes a Port 42 command
func executeCommand(input json.RawMessage) (string, error) {
	var params struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
		Stdin   string   `json:"stdin"`
	}
	
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid parameters: %v", err)
	}
	
	// Security: verify command exists in Port 42 commands directory
	cmdPath := filepath.Join(os.Getenv("HOME"), ".port42/commands", params.Command)
	if _, err := os.Stat(cmdPath); err != nil {
		return "", fmt.Errorf("command not found: %s", params.Command)
	}
	
	// Create command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, cmdPath, params.Args...)
	
	// Set up stdin if provided
	if params.Stdin != "" {
		cmd.Stdin = strings.NewReader(params.Stdin)
	}
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include output even on error
		if len(output) > 0 {
			return string(output), fmt.Errorf("command failed: %v", err)
		}
		return "", fmt.Errorf("command failed: %v", err)
	}
	
	// Limit output size to prevent huge responses
	const maxOutputSize = 50000 // 50KB
	if len(output) > maxOutputSize {
		output = output[:maxOutputSize]
		output = append(output, []byte("\n... (output truncated)")...)
	}
	
	return string(output), nil
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