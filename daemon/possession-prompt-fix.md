# Fix for Command Generation False Positives

## Problem
The AI agents are generating commands when users ask questions like "what did we do last time?", causing stack overflows and crashes.

## Root Cause
The agent prompts are too focused on command generation and don't distinguish between:
- Questions about the conversation/history
- Actual requests to create commands

## Solution
Update the agent prompts to include:

```go
baseGuidance := `
IMPORTANT: Only generate a command specification when the user EXPLICITLY asks you to create, build, or generate a command. 

Do NOT generate commands when the user:
- Asks questions about previous conversations
- Asks what we did before
- Asks for clarification or explanation
- Is having a general discussion

DO generate commands when the user:
- Says "create a command that..."
- Says "I need a command for..."
- Says "build me a tool that..."
- Uses /crystallize
- Explicitly requests command generation

When in doubt, ask for clarification rather than generating a command.
` + existingGuidance
```

## Temporary Workaround
Users should be explicit when they want commands:
- "Create a command that..." ✓
- "What did we do last time?" ✗ (will not generate command)

## Testing
1. Ask "what did we do last time?" - should NOT generate a command
2. Ask "create a command that shows recent activity" - SHOULD generate a command