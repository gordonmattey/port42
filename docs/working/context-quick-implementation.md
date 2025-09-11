# Port42 Context Commands - Quick Implementation

## Phase 1: Minimal Viable Context (What We Can Ship Today)

### 1. Extend Session Tracking

We already track sessions. Let's add to the session data:

```go
// daemon/src/types.go - Add to Session struct
type Session struct {
    // ... existing fields ...
    
    // New tracking fields
    CreatedTools    []string    `json:"created_tools"`    // Tool names created
    ExecutedCommands []string   `json:"executed_commands"` // Port42 commands run
    AccessedPaths   []string    `json:"accessed_paths"`    // VFS paths accessed
}
```

### 2. Add Command History Buffer

Simple in-memory circular buffer in daemon:

```go
// daemon/src/daemon.go
type CommandHistory struct {
    commands []CommandRecord
    maxSize  int
    mutex    sync.RWMutex
}

type CommandRecord struct {
    Command   string    `json:"command"`
    Timestamp time.Time `json:"timestamp"`
    SessionID string    `json:"session_id,omitempty"`
}

func (d *Daemon) TrackCommand(cmd string, sessionID string) {
    d.commandHistory.Add(CommandRecord{
        Command:   cmd,
        Timestamp: time.Now(),
        SessionID: sessionID,
    })
}
```

### 3. Simple Context Endpoint

```go
// daemon/src/server.go
func (d *Daemon) handleGetContext(w http.ResponseWriter, r *http.Request) {
    ctx := d.buildContext()
    json.NewEncoder(w).Encode(ctx)
}

func (d *Daemon) buildContext() ContextResponse {
    d.mu.RLock()
    defer d.mu.RUnlock()
    
    // Get active session
    var activeSession *SessionInfo
    if d.currentSession != nil {
        activeSession = &SessionInfo{
            ID:           d.currentSession.ID,
            Agent:        d.currentSession.Agent,
            MessageCount: len(d.currentSession.Messages),
            StartTime:    d.currentSession.CreatedAt,
            LastActivity: d.currentSession.LastActivity,
        }
    }
    
    // Get recent commands
    recentCommands := d.commandHistory.GetRecent(10)
    
    // Get created tools from current session
    var createdTools []string
    if d.currentSession != nil {
        createdTools = d.currentSession.CreatedTools
    }
    
    return ContextResponse{
        ActiveSession:  activeSession,
        RecentCommands: recentCommands,
        CreatedTools:   createdTools,
        Timestamp:      time.Now(),
    }
}
```

### 4. CLI Context Command

```rust
// cli/src/commands/context.rs
use crate::client::DaemonClient;
use anyhow::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
struct ContextResponse {
    active_session: Option<SessionInfo>,
    recent_commands: Vec<CommandRecord>,
    created_tools: Vec<String>,
    timestamp: String,
}

pub fn handle_context(port: u16, json: bool) -> Result<()> {
    let mut client = DaemonClient::new(port);
    
    // Simple GET request to daemon
    let response = client.request(DaemonRequest {
        request_type: "context".to_string(),
        id: generate_id(),
        payload: json!({})
    })?;
    
    if json {
        println!("{}", serde_json::to_string_pretty(&response.data)?);
    } else {
        display_context(response.data)?;
    }
    
    Ok(())
}

fn display_context(data: Value) -> Result<()> {
    let ctx: ContextResponse = serde_json::from_value(data)?;
    
    // Active session
    if let Some(session) = ctx.active_session {
        println!("🔄 Active: {} session ({} messages)", 
            session.agent, session.message_count);
        println!("   Started: {}", format_duration(session.start_time));
    } else {
        println!("💤 No active session");
    }
    
    // Recent commands
    if !ctx.recent_commands.is_empty() {
        println!("\n📝 Recent Commands:");
        for cmd in &ctx.recent_commands[..5.min(ctx.recent_commands.len())] {
            println!("  • {} [{}]", 
                cmd.command, 
                format_age(cmd.timestamp));
        }
    }
    
    // Created tools
    if !ctx.created_tools.is_empty() {
        println!("\n🛠 Created This Session:");
        for tool in &ctx.created_tools {
            println!("  • {}", tool);
        }
    }
    
    Ok(())
}
```

### 5. CLI Watch Command (Simple Version)

```rust
// cli/src/commands/watch.rs
use std::time::Duration;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, Ordering};

pub fn handle_watch(port: u16, refresh_ms: u64) -> Result<()> {
    let client = DaemonClient::new(port);
    let running = Arc::new(AtomicBool::new(true));
    let r = running.clone();
    
    // Ctrl+C handler
    ctrlc::set_handler(move || {
        r.store(false, Ordering::SeqCst);
    })?;
    
    while running.load(Ordering::SeqCst) {
        // Clear screen
        print!("\x1B[2J\x1B[1;1H");
        
        // Get and display context
        if let Ok(response) = client.request(/* context request */) {
            draw_panel(response.data);
        }
        
        // Wait
        thread::sleep(Duration::from_millis(refresh_ms));
    }
    
    println!("\n👋 Context monitor stopped");
    Ok(())
}

fn draw_panel(data: Value) {
    let width = 50;
    
    // Simple ASCII box
    println!("┌{}┐", "─".repeat(width - 2));
    println!("│ Port42 Context Monitor {:>23} │", "🔄");
    println!("├{}┤", "─".repeat(width - 2));
    
    // ... display context data ...
    
    println!("└{}┘", "─".repeat(width - 2));
}
```

## What to Track Where

### In Daemon Memory (ephemeral)
- Recent commands (last 50-100)
- Current active session pointer
- Command execution times

### In Session Data (persistent)
- Tools/artifacts created in session
- Commands executed in session  
- Paths accessed in session

### Not Tracking Yet (Phase 2)
- Memory access patterns
- Cross-session analytics
- Smart suggestions
- Command success/failure rates

## Dead Simple Suggestions

Just two for now:

```go
func generateSimpleSuggestions(ctx *ContextState) []string {
    suggestions := []string{}
    
    // 1. Continue session if active
    if ctx.ActiveSession != nil {
        suggestions = append(suggestions, 
            fmt.Sprintf("port42 possess %s --session last", ctx.ActiveSession.Agent))
    }
    
    // 2. Help for newly created tool
    if len(ctx.CreatedTools) > 0 {
        suggestions = append(suggestions,
            fmt.Sprintf("%s --help", ctx.CreatedTools[0]))
    }
    
    return suggestions
}
```

## Implementation Order (Quick Win)

1. **Add command tracking to daemon** (30 min)
   - Add CommandHistory buffer
   - Track in handleRequest

2. **Add context endpoint** (30 min)
   - New /api/context handler
   - Aggregate current data

3. **Implement context command** (1 hour)
   - Basic CLI command
   - Text + JSON output

4. **Implement watch command** (1 hour)
   - Simple refresh loop
   - Basic ASCII display

5. **Add tool tracking to sessions** (30 min)
   - Extend session struct
   - Track in possess handler

Total: ~3.5 hours for basic working version

## Testing Manually

```bash
# Start daemon with new endpoints
$ port42 daemon

# In another terminal, create activity
$ port42 possess @ai-engineer "create a test tool"
$ port42 search "test"
$ port42 ls /tools/

# Check context
$ port42 context
🔄 Active: ai-engineer session (1 messages)
   Started: 30 seconds ago

📝 Recent Commands:
  • port42 search "test" [5s]
  • port42 ls /tools/ [2s]

# Watch live
$ port42 watch
[Live updating display]
```

## What This Gives Us

1. **Immediate visibility** into Port42 activity
2. **Session awareness** for users and Claude
3. **Command history** without complex tracking
4. **Foundation** for smarter features later

This is enough to be useful TODAY while keeping the door open for the advanced features in the full spec.