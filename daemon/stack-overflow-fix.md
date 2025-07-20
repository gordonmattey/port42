# Stack Overflow Fix for Daemon

## Problem
The daemon's memory endpoint returns full session history (all messages) for all recent sessions, causing:
- 141KB+ responses
- Stack overflow in CLI when parsing
- Crashes when continuing sessions

## Debug Evidence
```
DEBUG: Response line length: 141367 bytes
DEBUG: Found 43 recent sessions
thread 'main' has overflowed its stack
```

## Root Cause
In `daemon/main.go`, the memory endpoint returns complete Session objects including all messages:
```go
recent_sessions: recentSessions  // <- Contains ALL messages!
```

## Fix Required
The memory endpoint should return only session metadata:
```go
type SessionSummary struct {
    ID           string    `json:"id"`
    Agent        string    `json:"agent"`
    CreatedAt    time.Time `json:"created_at"`
    LastActivity time.Time `json:"last_activity"`
    MessageCount int       `json:"message_count"`
    State        string    `json:"state"`
}
```

## Changes Needed in daemon/main.go

1. In the memory handler, create summaries instead of full sessions:
```go
var summaries []SessionSummary
for _, session := range recentSessions {
    summaries = append(summaries, SessionSummary{
        ID:           session.ID,
        Agent:        session.Agent,
        CreatedAt:    session.CreatedAt,
        LastActivity: session.LastActivity,
        MessageCount: len(session.Messages),
        State:        session.State,
    })
}
```

2. Return summaries in the response:
```go
"recent_sessions": summaries,
```

## Testing
After fix:
1. Memory responses should be <10KB
2. No stack overflow when continuing sessions
3. CLI can still find recent sessions by ID and agent