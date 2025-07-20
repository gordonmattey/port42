# Possess Response Fix for Stack Overflow

## Problem
When continuing a session, the daemon's possess endpoint returns the ENTIRE conversation history in the response, causing stack overflow.

## Evidence
```
DEBUG: Sending message to session: cli-1752987755
DEBUG: About to send request to daemon
thread 'main' has overflowed its stack  <- Happens AFTER possess request
```

## Root Cause
The possess handler is likely returning the full session object including all messages:
```go
// BAD - returns entire session with all messages
response := map[string]interface{}{
    "message": aiResponse,
    "session": session,  // <- Contains ALL messages!
}
```

## Fix Required
Return only the new AI response:
```go
// GOOD - returns only what's needed
response := map[string]interface{}{
    "message": aiResponse,
    "command_generated": commandGenerated,
    "command_spec": commandSpec,
}
```

## Why This Happens
When continuing a session with history, the daemon:
1. Loads the full session from disk
2. Adds all history to the AI context (correct)
3. Gets AI response
4. Returns the full session in response (WRONG)

The CLI only needs:
- The new AI message
- Command generation info
- NOT the conversation history

## Testing
After fix:
1. Possess responses should be <5KB even for continued sessions
2. No stack overflow when asking "what did we do?"
3. Session continuation still works with context