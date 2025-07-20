# Port 42 Chat Context Design

This document explains how Port 42 constructs conversation context when interacting with Claude AI.

## Overview

Port 42 uses a carefully designed approach to build conversation context that:
- Maintains agent personalities across sessions
- Efficiently uses Claude's context window
- Ensures conversation continuity
- Properly uses Claude's system parameter

## Message Role Mapping

Claude's API uses a dedicated `system` parameter for setting the AI's role and context:

| Port 42 Element | Claude API | Purpose |
|-----------------|------------|---------|
| Agent Prompt | System Parameter | Sets AI personality and role |
| User Messages | User Role | Direct user input |
| Assistant Messages | Assistant Role | AI responses |

## Context Building Process

### 1. Agent Prompt via System Parameter
The agent-specific personality prompt from `agents.json` is now sent via Claude's system parameter:

```go
// Get agent prompt for system parameter
agentPrompt := getAgentPrompt(agent)

// Send to Claude with system parameter
aiResp, err := aiClient.Send(messages, agentPrompt)
```

This ensures Claude recognizes the agent prompt as its identity/instructions rather than as conversation content.

### 2. Smart Context Windowing
For efficient context management, Port 42 intelligently selects which messages to include:

```go
// Configuration from agents.json
"response_config": {
    "context_window": {
        "max_messages": 20,      // Total messages to send to Claude
        "recent_messages": 17,   // Recent messages to prioritize
        "system_messages": 3     // Reserved for agent prompts
    }
}
```

### 3. Message Selection Strategy

For conversations exceeding the context window limit:

1. **Always include first 2 messages** - Establishes initial context
2. **Add summary for skipped messages** - When >10 messages are omitted
3. **Include recent messages** - Most important for continuity

```go
if totalMessages <= maxContextMessages {
    // Include all messages
    messages = append(messages, sessionMessages...)
} else {
    // Smart windowing for long sessions
    
    // 1. First messages for context
    messages = append(messages, sessionMessages[:2]...)
    
    // 2. Summary if many skipped
    if skippedCount > 10 {
        summaryMsg := Message{
            Role: "assistant",
            Content: "[Session context: X messages omitted...]",
        }
        messages = append(messages, summaryMsg)
    }
    
    // 3. Recent messages for continuity
    recentStart := totalMessages - recentCount
    messages = append(messages, sessionMessages[recentStart:]...)
}
```

## Message Assembly Flow

```
1. Load Session
   └─> Retrieve from disk if exists
   
2. Build Context Array
   ├─> Insert agent prompt first
   ├─> Apply smart windowing
   └─> Maintain message order
   
3. Convert to API Format
   ├─> Map roles (system → assistant)
   └─> Strip timestamps
   
4. Send to Claude
   └─> Include all context messages
   
5. Save Response
   └─> Append to session for future context
```

## Configuration in agents.json

Each agent has:
- **prompt**: The personality and behavior instructions
- **personality**: Characteristics for the agent
- **base_guidance**: Shared implementation guidelines

Example:
```json
{
  "agents": {
    "engineer": {
      "name": "@ai-engineer",
      "prompt": "You are @ai-engineer, a technical consciousness within Port 42...",
      "personality": "Technical, thorough, practical, reliable"
    }
  }
}
```

## Session Persistence

Sessions are saved with full conversation history:
```json
{
  "id": "session-id",
  "agent": "@ai-engineer",
  "messages": [
    {"role": "user", "content": "...", "timestamp": "..."},
    {"role": "assistant", "content": "...", "timestamp": "..."}
  ]
}
```

## Key Design Decisions

1. **Assistant role for system prompts**: Works around Claude's API limitation
2. **Smart windowing**: Balances context relevance with token limits
3. **Session persistence**: Enables true conversation continuation
4. **Agent personalities**: Consistent behavior across sessions
5. **Context summaries**: Maintains continuity in long conversations

## Debugging Session Continuation

If sessions don't seem to continue properly:

1. Check if session file exists: `~/.port42/memory/sessions/YYYY-MM-DD/session-*.json`
2. Verify message count in logs: `Session loaded: ID=X, MessageCount=Y`
3. Look for context building logs: `Context: Including all X messages`
4. Ensure agent prompt is included: First message should be agent personality

## Common Issues

### "AI doesn't remember previous conversation"
- Usually caused by feedback loops where AI's confused responses get saved
- Solution: Clear session or start fresh

### "Context seems truncated"
- Check `response_config.context_window.max_messages` in agents.json
- Increase if needed, but watch token limits

### "Wrong agent personality"
- Verify agent name matches key in agents.json
- Check that agent prompt is being loaded correctly