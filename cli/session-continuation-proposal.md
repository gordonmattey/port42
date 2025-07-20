# Session Continuation Proposal for Port 42

## Current Issue
Sessions are saved but new conversations always start fresh because the CLI generates new session IDs based on timestamps.

## Proposed Solutions

### 1. Add `--continue` flag
```bash
port42 possess @ai-muse --continue
```

This would:
1. Query daemon for recent sessions
2. Show a list:
   ```
   Recent sessions:
   1. cli-1752985361 (2 minutes ago, 3 messages, @ai-muse)
   2. my-project (1 hour ago, 15 messages, @ai-engineer) 
   3. git-tools (yesterday, command created: git-haiku, @ai-muse)
   
   Select session to continue (1-3) or press Enter for new:
   ```

### 2. Add `port42 sessions` command
```bash
port42 sessions
# Lists all recent sessions

port42 sessions show cli-1752985361
# Shows conversation history
```

### 3. Smart session naming
Instead of timestamp-based IDs, use:
- Topic extraction from first message
- Command name if one was generated
- Agent + date combination

### 4. Per-agent memory (Advanced)
```bash
port42 possess @ai-muse
# Automatically continues last @ai-muse session if recent
```

## Implementation Priority

1. **Quick fix**: Document the `--session` flag better
2. **Medium**: Add `--continue` flag to possess command
3. **Long term**: Full session management commands

The system already has all the infrastructure - we just need to surface it better in the CLI!