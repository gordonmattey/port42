# Session Resumption Improvements - Implementation Plan

## Overview
Remove ambiguity in session resumption by adding explicit `--session` parameter and eliminating guesswork about whether arguments are session IDs or messages.

## Current Problems
1. CLI guesses whether "cli-1234" is a message or session ID
2. Limited to specific patterns (cli-, numbers, etc.)
3. Can't handle custom session IDs
4. Ambiguous user intent

## Proposed Solution
- Add explicit `--session` parameter
- Support `--session last` for most recent session
- All positional args always treated as message
- No backward compatibility needed (clean break)

## Implementation Phases

### Phase 1: CLI Changes

#### 1.1 Update Command Structure (`cli/src/main.rs`)

**Current:**
```rust
Possess {
    agent: String,
    #[arg(long = "ref", action = clap::ArgAction::Append)]
    references: Option<Vec<String>>,
    args: Vec<String>,  // Could be session ID or message
}
```

**New:**
```rust
Possess {
    agent: String,
    
    #[arg(long, help = "Session ID to resume, or 'last' for most recent")]
    session: Option<String>,
    
    #[arg(long = "ref", action = clap::ArgAction::Append)]
    references: Option<Vec<String>>,
    
    /// Message to send to the AI
    #[arg(trailing_var_arg = true)]
    message: Vec<String>,
}
```

#### 1.2 Update Command Handler (`cli/src/main.rs:340-387`)

**Remove:** All session ID detection logic (lines 342-380)

**Replace with:**
```rust
Some(Commands::Possess { agent, session, references, message }) => {
    // Simple: session is explicit, message is always the args
    let message_text = if message.is_empty() { 
        None 
    } else { 
        Some(message.join(" ")) 
    };
    
    // Handle special "last" value
    let session_id = match session.as_deref() {
        Some("last") => {
            // Query daemon for last session
            let client = DaemonClient::new(port);
            match client.get_last_session() {
                Ok(id) => Some(id),
                Err(_) => {
                    eprintln!("No previous sessions found");
                    std::process::exit(1);
                }
            }
        },
        Some(id) => Some(id.to_string()),
        None => None,
    };
    
    if std::env::var("PORT42_DEBUG").is_ok() {
        eprintln!("DEBUG possess: agent={}, session={:?}, message={:?}", 
                 agent, session_id, message_text);
    }
    
    let show_boot = message_text.is_none();
    commands::possess::handle_possess_with_references(
        port, agent, message_text, session_id, references, show_boot
    )?;
}
```

### Phase 2: Protocol Changes

#### 2.1 Add New Request Type (`cli/src/protocol/mod.rs`)

```rust
pub struct GetLastSessionRequest {
    pub agent: Option<String>,  // None means any agent
}

pub struct GetLastSessionResponse {
    pub session_id: String,
    pub agent: String,
    pub last_activity: String,
    pub message_count: usize,
}
```

#### 2.2 Add Client Method (`cli/src/client.rs`)

```rust
impl DaemonClient {
    pub fn get_last_session(&self) -> Result<String> {
        let request = Request {
            id: generate_request_id(),
            request_type: "get_last_session".to_string(),
            payload: json!({}),
            references: vec![],
        };
        
        let response = self.send_request(request)?;
        
        if !response.success {
            bail!("Failed to get last session: {}", 
                  response.error.unwrap_or_default());
        }
        
        // Extract session_id from response
        let data = response.data
            .ok_or_else(|| anyhow!("No data in response"))?;
        
        data["session_id"]
            .as_str()
            .ok_or_else(|| anyhow!("No session_id in response"))
            .map(|s| s.to_string())
    }
}
```

### Phase 3: Daemon Changes

#### 3.1 Add Request Handler (`daemon/src/server.go`)

**Add to handleRequest switch:**
```go
case "get_last_session":
    return d.handleGetLastSession(req)
```

**Add handler function:**
```go
func (d *Daemon) handleGetLastSession(req Request) Response {
    resp := NewResponse(req.ID, true)
    
    // Get most recent session from storage
    if d.storage == nil {
        resp.SetError("Storage not initialized")
        return resp
    }
    
    sessionID, err := d.storage.GetLastSession()
    if err != nil {
        resp.SetError(fmt.Sprintf("No sessions found: %v", err))
        return resp
    }
    
    // Load session to get metadata
    session, err := d.storage.LoadSession(sessionID)
    if err != nil {
        resp.SetError(fmt.Sprintf("Failed to load session: %v", err))
        return resp
    }
    
    data := map[string]interface{}{
        "session_id":    session.ID,
        "agent":         session.Agent,
        "last_activity": session.LastActivity.Format(time.RFC3339),
        "message_count": len(session.Messages),
    }
    
    resp.SetData(data)
    return resp
}
```

#### 3.2 Update Storage Interface (`daemon/src/storage.go`)

**Add to Storage interface:**
```go
type Storage interface {
    // ... existing methods ...
    GetLastSession() (string, error)
    UpdateLastSession(sessionID string) error
}
```

**Add to ContentStorage implementation:**
```go
func (cs *ContentStorage) GetLastSession() (string, error) {
    // Read from special metadata path
    lastSessionPath := "/memory/sessions/last-active"
    
    // Use VFS to read the symlink/reference
    paths, err := cs.vfs.List(lastSessionPath)
    if err != nil || len(paths) == 0 {
        return "", fmt.Errorf("no last session found")
    }
    
    // The first path should be the session ID
    // Extract session ID from path like /memory/sessions/last-active/cli-123456
    parts := strings.Split(paths[0], "/")
    if len(parts) > 0 {
        return parts[len(parts)-1], nil
    }
    
    return "", fmt.Errorf("invalid last session path")
}

func (cs *ContentStorage) UpdateLastSession(sessionID string) error {
    // Update the last-active reference
    lastSessionPath := fmt.Sprintf("/memory/sessions/last-active/%s", sessionID)
    
    // Store a reference/symlink via VFS
    metadata := map[string]interface{}{
        "type":           "session-reference",
        "session_id":     sessionID,
        "updated":        time.Now().Unix(),
    }
    
    // Use StoreWithMetadata to create/update the reference
    _, err := cs.StoreWithMetadata(
        []byte(sessionID),  // Simple content: just the session ID
        "session-reference",
        metadata,
        []string{
            "/memory/sessions/last-active",  // Clear old, set new
            lastSessionPath,
        },
    )
    
    return err
}
```

#### 3.3 Update Session Save (`daemon/src/storage.go`)

**Modify SaveSession to update last-active:**
```go
func (cs *ContentStorage) SaveSession(session *Session) error {
    // ... existing save logic ...
    
    // After successful save, update last-active
    if err == nil {
        cs.UpdateLastSession(session.ID)
    }
    
    return err
}
```

### Phase 4: Help Text Updates

#### 4.1 Update Help (`cli/src/help_text.rs`)

```rust
pub const POSSESS_HELP: &str = "Channel an AI agent's consciousness

USAGE:
    port42 possess <AGENT> [OPTIONS] [MESSAGE...]

ARGUMENTS:
    <AGENT>       AI agent (@ai-engineer, @ai-muse, @ai-analyst, @ai-founder)
    [MESSAGE...]  Optional message to send

OPTIONS:
    --session <ID>    Resume specific session (use 'last' for most recent)
    --ref <REF>       Add reference context (file:, p42:, url:, search:)
    -h, --help        Print help information

EXAMPLES:
    # Start new conversation
    port42 possess @ai-engineer \"help me build a parser\"
    
    # Resume last session
    port42 possess @ai-engineer --session last \"continue\"
    
    # Resume specific session
    port42 possess @ai-engineer --session cli-1234567890
    
    # With references
    port42 possess @ai-engineer --ref file:./spec.md \"implement this\"";
```

## Testing Plan

### Test Scenarios

1. **New Session Creation**
   ```bash
   port42 possess @ai-engineer "test message"
   # Verify new session created
   ```

2. **Resume Last Session**
   ```bash
   # Create session A
   port42 possess @ai-engineer "first message"
   # Resume it
   port42 possess @ai-engineer --session last
   # Verify session A resumed with history
   ```

3. **Resume Specific Session**
   ```bash
   port42 possess @ai-engineer --session cli-123
   # Verify correct session loaded
   ```

4. **No Previous Sessions**
   ```bash
   # Clear all sessions
   port42 possess @ai-engineer --session last
   # Verify appropriate error message
   ```

5. **Agent Switching**
   ```bash
   # Create session with @ai-muse
   port42 possess @ai-muse "hello"
   # Try to resume with different agent
   port42 possess @ai-engineer --session last
   # Verify warning but proceeds
   ```

6. **Multiple Sessions**
   ```bash
   # Create multiple sessions
   # Verify 'last' picks most recent by timestamp
   ```

## Edge Cases to Handle

1. **Daemon Restart**: Last session tracking must persist
2. **Corrupted Session**: Graceful error handling
3. **Missing Storage**: Clear error message
4. **Race Conditions**: Multiple clients updating last session
5. **Session Cleanup**: Old sessions don't affect "last"

## Implementation Order

1. **Step 1**: CLI changes (Phase 1)
   - Update command structure
   - Remove guessing logic
   - Add session parameter handling

2. **Step 2**: Protocol additions (Phase 2)  
   - Add get_last_session request/response
   - Update client with new method

3. **Step 3**: Daemon handlers (Phase 3.1)
   - Add get_last_session handler
   - Return session metadata

4. **Step 4**: Storage updates (Phase 3.2-3.3)
   - Implement last session tracking
   - Update on each save

5. **Step 5**: Testing and edge cases
   - Run through all test scenarios
   - Fix any issues

6. **Step 6**: Help text and documentation
   - Update all help text
   - Update examples

## Success Criteria

- [ ] No more guessing if string is session or message
- [ ] `--session last` works reliably
- [ ] Clean error messages for edge cases
- [ ] Session resumption works after daemon restart
- [ ] Clear, unambiguous interface

## Example Usage After Implementation

```bash
# Start new session (no --session = new)
$ port42 possess @ai-engineer "help me build a tool"
🤖 Channeling @ai-engineer...
Session: cli-1757350466684
[AI responds...]

# Resume that specific session
$ port42 possess @ai-engineer --session cli-1757350466684 "add error handling"
🔄 Resuming session cli-1757350466684 (3 messages)
[AI continues with context...]

# Resume last session (convenience)
$ port42 possess @ai-engineer --session last "what were we working on?"
🔄 Resuming session cli-1757350466684 (4 messages)
[AI continues...]

# Start fresh even though sessions exist
$ port42 possess @ai-engineer "new topic about testing"
🤖 Channeling @ai-engineer...
Session: cli-1757360123456
[AI responds to new topic...]

# Check available sessions
$ port42 ls /memory/sessions/
cli-1757350466684  (2024-01-09 10:30:15) @ai-engineer - 5 messages
cli-1757360123456  (2024-01-09 11:45:30) @ai-engineer - 2 messages
cli-1757340987654  (2024-01-08 16:20:00) @ai-muse - 8 messages
```

## Notes

- No backward compatibility layer needed
- Clean break for cleaner UX
- Existing `port42 ls /memory` already provides session listing
- Focus on clarity over cleverness