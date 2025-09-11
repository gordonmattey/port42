# Port42 Context & Watch Implementation Plan

## Current State

### What We Have
- **Location**: `daemon/src/server.go:678-721` in `handleGetContext`
- **Endpoint**: Wired up at line ~611 in `handleRequest`
- **Data**: Returns only `active_session` with basic info
- **CLI**: Basic `port42 context` command with JSON/pretty/compact output

### What's Missing
- Recent commands tracking
- Created tools list (beyond single CommandGenerated)
- Contextual suggestions
- Memory access tracking
- Watch mode with live updates

## Architecture: Shared Components

```
┌─────────────────────────────────────────┐
│           CLI (Rust)                     │
├─────────────────────────────────────────┤
│  Context Command                         │
│  ├── DataFetcher (shared)               │
│  ├── Formatters (shared)                │
│  │   ├── JSONFormatter                  │
│  │   ├── PrettyFormatter                │
│  │   ├── CompactFormatter               │
│  │   └── WatchFormatter (ASCII boxes)   │
│  └── WatchLoop (refresh manager)        │
└─────────────────────────────────────────┘
                    ↕ TCP
┌─────────────────────────────────────────┐
│         Daemon (Go)                      │
├─────────────────────────────────────────┤
│  ContextCollector (shared)               │
│  ├── SessionData                        │
│  ├── RecentCommands (circular buffer)   │
│  ├── CreatedTools                       │
│  └── Suggestions                        │
└─────────────────────────────────────────┘
```

## Implementation Steps

### Step 1: Create Shared Data Structures

**Daemon Side - Create `daemon/src/context.go`**
```go
package main

import "time"

// ContextData is the complete context structure
type ContextData struct {
    ActiveSession    *SessionContext      `json:"active_session"`
    RecentCommands   []CommandRecord      `json:"recent_commands"`
    CreatedTools     []ToolRecord         `json:"created_tools"`
    AccessedMemories []MemoryAccess       `json:"accessed_memories,omitempty"`
    Suggestions      []ContextSuggestion  `json:"suggestions"`
}

type SessionContext struct {
    ID           string    `json:"id"`
    Agent        string    `json:"agent"`
    MessageCount int       `json:"message_count"`
    StartTime    time.Time `json:"start_time"`
    LastActivity time.Time `json:"last_activity"`
    State        string    `json:"state"`
    ToolCreated  *string   `json:"tool_created,omitempty"`
}

type CommandRecord struct {
    Command    string    `json:"command"`
    Timestamp  time.Time `json:"timestamp"`
    AgeSeconds int       `json:"age_seconds"`
    ExitCode   int       `json:"exit_code"`
}

type ToolRecord struct {
    Name       string    `json:"name"`
    Type       string    `json:"type"`
    Transforms []string  `json:"transforms"`
    CreatedAt  time.Time `json:"created_at"`
}

type MemoryAccess struct {
    Path        string    `json:"path"`
    Type        string    `json:"type"`
    AccessCount int       `json:"access_count"`
}

type ContextSuggestion struct {
    Command    string  `json:"command"`
    Reason     string  `json:"reason"`
    Confidence float64 `json:"confidence"`
}
```

**CLI Side - Create `cli/src/context/mod.rs`**
```rust
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ContextData {
    pub active_session: Option<SessionContext>,
    pub recent_commands: Vec<CommandRecord>,
    pub created_tools: Vec<ToolRecord>,
    #[serde(default)]
    pub accessed_memories: Vec<MemoryAccess>,
    pub suggestions: Vec<ContextSuggestion>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct SessionContext {
    pub id: String,
    pub agent: String,
    pub message_count: i32,
    pub start_time: DateTime<Utc>,
    pub last_activity: DateTime<Utc>,
    pub state: String,
    pub tool_created: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CommandRecord {
    pub command: String,
    pub timestamp: DateTime<Utc>,
    pub age_seconds: i32,
    pub exit_code: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ToolRecord {
    pub name: String,
    #[serde(rename = "type")]
    pub tool_type: String,
    pub transforms: Vec<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct MemoryAccess {
    pub path: String,
    #[serde(rename = "type")]
    pub access_type: String,
    pub access_count: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ContextSuggestion {
    pub command: String,
    pub reason: String,
    pub confidence: f64,
}
```

### Step 2: Create Daemon Context Collector

**Create `daemon/src/context_collector.go`**
```go
package main

import (
    "fmt"
    "time"
    "sync"
)

type ContextCollector struct {
    mu             sync.RWMutex
    daemon         *Daemon
    recentCommands []CommandRecord
    maxCommands    int
}

func NewContextCollector(daemon *Daemon) *ContextCollector {
    return &ContextCollector{
        daemon:      daemon,
        maxCommands: 20,
        recentCommands: make([]CommandRecord, 0, 20),
    }
}

func (cc *ContextCollector) TrackCommand(cmd string, exitCode int) {
    cc.mu.Lock()
    defer cc.mu.Unlock()
    
    record := CommandRecord{
        Command:   cmd,
        Timestamp: time.Now(),
        ExitCode:  exitCode,
    }
    
    // Add to front of slice
    cc.recentCommands = append([]CommandRecord{record}, cc.recentCommands...)
    
    // Trim to max size
    if len(cc.recentCommands) > cc.maxCommands {
        cc.recentCommands = cc.recentCommands[:cc.maxCommands]
    }
}

func (cc *ContextCollector) Collect() *ContextData {
    data := &ContextData{
        RecentCommands:   []CommandRecord{},
        CreatedTools:     []ToolRecord{},
        Suggestions:      []ContextSuggestion{},
        AccessedMemories: []MemoryAccess{},
    }
    
    // Get active session
    cc.daemon.mu.RLock()
    var activeSession *Session
    var latestTime time.Time
    
    for _, session := range cc.daemon.sessions {
        if session.State == SessionActive && session.LastActivity.After(latestTime) {
            activeSession = session
            latestTime = session.LastActivity
        }
    }
    cc.daemon.mu.RUnlock()
    
    if activeSession != nil {
        data.ActiveSession = &SessionContext{
            ID:           activeSession.ID,
            Agent:        activeSession.Agent,
            MessageCount: len(activeSession.Messages),
            StartTime:    activeSession.CreatedAt,
            LastActivity: activeSession.LastActivity,
            State:        string(activeSession.State),
        }
        
        if activeSession.CommandGenerated != nil {
            toolName := activeSession.CommandGenerated.Name
            data.ActiveSession.ToolCreated = &toolName
            
            // Add to created tools
            data.CreatedTools = append(data.CreatedTools, ToolRecord{
                Name:      toolName,
                Type:      "tool",
                CreatedAt: activeSession.LastActivity,
            })
        }
    }
    
    // Get recent commands with age calculation
    cc.mu.RLock()
    now := time.Now()
    for _, cmd := range cc.recentCommands {
        cmdCopy := cmd
        cmdCopy.AgeSeconds = int(now.Sub(cmd.Timestamp).Seconds())
        data.RecentCommands = append(data.RecentCommands, cmdCopy)
    }
    cc.mu.RUnlock()
    
    // Generate suggestions
    data.Suggestions = cc.generateSuggestions(data)
    
    return data
}

func (cc *ContextCollector) generateSuggestions(data *ContextData) []ContextSuggestion {
    suggestions := []ContextSuggestion{}
    
    // Always suggest viewing current session
    if data.ActiveSession != nil {
        suggestions = append(suggestions, ContextSuggestion{
            Command:    fmt.Sprintf("port42 info /memory/%s", data.ActiveSession.ID),
            Reason:     "View current session details",
            Confidence: 0.9,
        })
        
        // If tool was created, suggest using it
        if data.ActiveSession.ToolCreated != nil {
            suggestions = append(suggestions, ContextSuggestion{
                Command:    fmt.Sprintf("%s --help", *data.ActiveSession.ToolCreated),
                Reason:     "Learn about your new tool",
                Confidence: 0.95,
            })
        }
        
        // Suggest continuing session
        suggestions = append(suggestions, ContextSuggestion{
            Command:    fmt.Sprintf("port42 possess %s --session last", data.ActiveSession.Agent),
            Reason:     "Continue your conversation",
            Confidence: 0.85,
        })
    }
    
    return suggestions
}
```

**Update Daemon struct and initialization**
```go
// In server.go, add to Daemon struct:
type Daemon struct {
    // ... existing fields
    contextCollector *ContextCollector // NEW
}

// In NewDaemon or initialization:
d.contextCollector = NewContextCollector(d)

// Update handleRequest to track commands:
func (d *Daemon) handleRequest(req Request) Response {
    // Track command (except context to avoid recursion)
    if req.Type != "context" && d.contextCollector != nil {
        d.contextCollector.TrackCommand(req.Type, 0)
    }
    
    // ... existing switch statement
}

// Update handleGetContext:
func (d *Daemon) handleGetContext(req Request) Response {
    resp := NewResponse(req.ID, true)
    
    // Use collector if available, otherwise fall back to current implementation
    if d.contextCollector != nil {
        contextData := d.contextCollector.Collect()
        resp.SetData(contextData)
    } else {
        // ... existing implementation as fallback
    }
    
    return resp
}
```

### Step 3: Create CLI Presentation Layer

**Create `cli/src/context/formatters.rs`**
```rust
use super::*;
use chrono::{Local, Utc};

pub trait ContextFormatter {
    fn format(&self, data: &ContextData) -> String;
}

pub struct JsonFormatter;
impl ContextFormatter for JsonFormatter {
    fn format(&self, data: &ContextData) -> String {
        serde_json::to_string_pretty(data).unwrap_or_else(|_| "{}".to_string())
    }
}

pub struct CompactFormatter;
impl ContextFormatter for CompactFormatter {
    fn format(&self, data: &ContextData) -> String {
        let session = data.active_session.as_ref()
            .map(|s| format!("{}[{}]", s.agent, s.message_count))
            .unwrap_or_else(|| "no session".to_string());
        
        let last_cmd = data.recent_commands.first()
            .map(|c| truncate(&c.command, 20))
            .unwrap_or_default();
        
        format!("{} | last: {} | tools: {} | 💡 {} suggestions",
            session,
            last_cmd,
            data.created_tools.len(),
            data.suggestions.len())
    }
}

pub struct PrettyFormatter;
impl ContextFormatter for PrettyFormatter {
    fn format(&self, data: &ContextData) -> String {
        let mut output = String::new();
        
        // Active session
        if let Some(session) = &data.active_session {
            output.push_str(&format!("🔄 Active: {} session ({} messages)\n", 
                session.agent, session.message_count));
            output.push_str(&format!("   Session ID: {}\n", session.id));
            output.push_str(&format!("   Started: {}\n", session.start_time.format("%Y-%m-%d %H:%M:%S")));
            output.push_str(&format!("   Last activity: {}\n", session.last_activity.format("%Y-%m-%d %H:%M:%S")));
            output.push_str(&format!("   State: {}\n", session.state));
            
            if let Some(tool) = &session.tool_created {
                output.push_str(&format!("   Created tool: {}\n", tool));
            }
        } else {
            output.push_str("No active session\n");
        }
        
        // Recent commands
        if !data.recent_commands.is_empty() {
            output.push_str("\n📝 Recent Commands:\n");
            for cmd in data.recent_commands.iter().take(5) {
                output.push_str(&format!("  • {} [{}s ago]\n", 
                    cmd.command, cmd.age_seconds));
            }
        }
        
        // Created tools
        if !data.created_tools.is_empty() {
            output.push_str("\n🛠 Created Tools:\n");
            for tool in &data.created_tools {
                output.push_str(&format!("  • {}\n", tool.name));
            }
        }
        
        // Suggestions
        if !data.suggestions.is_empty() {
            output.push_str("\n💡 Suggestions:\n");
            for suggestion in data.suggestions.iter().take(3) {
                output.push_str(&format!("  • {}\n", suggestion.command));
                output.push_str(&format!("    {}\n", suggestion.reason));
            }
        }
        
        output
    }
}

pub struct WatchFormatter;
impl ContextFormatter for WatchFormatter {
    fn format(&self, data: &ContextData) -> String {
        let mut output = String::new();
        let width = 50;
        
        // Header
        output.push_str(&format!("┌{}┐\n", "─".repeat(width - 2)));
        output.push_str(&format!("│ Port42 Context Monitor {:>25} │\n", "🔄"));
        output.push_str(&format!("├{}┤\n", "─".repeat(width - 2)));
        
        // Active session
        if let Some(session) = &data.active_session {
            let id_short = if session.id.len() > 12 {
                &session.id[..12]
            } else {
                &session.id
            };
            output.push_str(&format!("│ ⚡ Session: {} [{}...]      │\n",
                pad_right(&session.agent, 12), id_short));
            
            let activity = format!("{} msgs", session.message_count);
            output.push_str(&format!("│ 📊 Activity: {}│\n",
                pad_right(&activity, 35)));
        } else {
            output.push_str("│ No active session                           │\n");
        }
        
        // Recent commands
        if !data.recent_commands.is_empty() {
            output.push_str(&format!("├{}┤\n", "─".repeat(width - 2)));
            output.push_str("│ 📝 Recent Commands:                          │\n");
            
            for cmd in data.recent_commands.iter().take(5) {
                let truncated = truncate(&cmd.command, 30);
                let age = format!("{}s", cmd.age_seconds);
                output.push_str(&format!("│ • {:<30} [{:>5}] │\n",
                    truncated, age));
            }
        }
        
        // Created tools
        if !data.created_tools.is_empty() {
            output.push_str(&format!("├{}┤\n", "─".repeat(width - 2)));
            output.push_str("│ 🛠  Tools Created:                           │\n");
            for tool in data.created_tools.iter().take(3) {
                output.push_str(&format!("│ • {:<42} │\n", 
                    truncate(&tool.name, 42)));
            }
        }
        
        // Suggestions
        if !data.suggestions.is_empty() {
            output.push_str(&format!("├{}┤\n", "─".repeat(width - 2)));
            output.push_str("│ 💡 Smart Suggestions:                        │\n");
            for suggestion in data.suggestions.iter().take(3) {
                let cmd_truncated = truncate(&suggestion.command, 42);
                output.push_str(&format!("│ • {:<42} │\n", cmd_truncated));
            }
        }
        
        // Footer
        output.push_str(&format!("├{}┤\n", "─".repeat(width - 2)));
        let timestamp = Local::now().format("%H:%M:%S");
        output.push_str(&format!("│ [Ctrl+C to exit] | Updated: {} {:>10} │\n", 
            timestamp, ""));
        output.push_str(&format!("└{}┘", "─".repeat(width - 2)));
        
        output
    }
}

fn truncate(s: &str, max_len: usize) -> String {
    if s.len() <= max_len {
        s.to_string()
    } else {
        format!("{}...", &s[..max_len-3])
    }
}

fn pad_right(s: &str, width: usize) -> String {
    if s.len() >= width {
        truncate(s, width)
    } else {
        format!("{}{}", s, " ".repeat(width - s.len()))
    }
}
```

**Create `cli/src/context/watch.rs`**
```rust
use super::*;
use crate::client::DaemonClient;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::Duration;
use std::thread;
use std::io::{self, Write};

pub struct WatchMode {
    client: DaemonClient,
    formatter: Box<dyn ContextFormatter>,
    refresh_rate: Duration,
}

impl WatchMode {
    pub fn new(client: DaemonClient, refresh_rate_ms: u64) -> Self {
        WatchMode {
            client,
            formatter: Box::new(WatchFormatter),
            refresh_rate: Duration::from_millis(refresh_rate_ms),
        }
    }
    
    pub fn run(&mut self) -> Result<(), Box<dyn std::error::Error>> {
        // Setup Ctrl+C handler
        let running = Arc::new(AtomicBool::new(true));
        let r = running.clone();
        
        ctrlc::set_handler(move || {
            r.store(false, Ordering::SeqCst);
            // Move cursor to bottom and show cursor
            print!("\x1B[999;1H\x1B[?25h");
            let _ = io::stdout().flush();
        })?;
        
        // Hide cursor
        print!("\x1B[?25l");
        io::stdout().flush()?;
        
        while running.load(Ordering::SeqCst) {
            // Clear screen and reset cursor
            print!("\x1B[2J\x1B[1;1H");
            
            // Fetch latest context
            match self.fetch_context() {
                Ok(context) => {
                    // Format and display
                    let output = self.formatter.format(&context);
                    print!("{}", output);
                },
                Err(e) => {
                    print!("Error fetching context: {}", e);
                }
            }
            
            // Flush to ensure immediate display
            io::stdout().flush()?;
            
            // Wait for next refresh
            thread::sleep(self.refresh_rate);
        }
        
        // Show cursor again
        print!("\x1B[?25h");
        println!("\nWatch mode ended.");
        io::stdout().flush()?;
        
        Ok(())
    }
    
    fn fetch_context(&mut self) -> Result<ContextData, Box<dyn std::error::Error>> {
        use crate::protocol::DaemonRequest;
        
        let request = DaemonRequest {
            request_type: "context".to_string(),
            data: None,
        };
        
        let response = self.client.request(request)?;
        
        if let Some(data) = response.data {
            let context: ContextData = serde_json::from_value(data)?;
            Ok(context)
        } else {
            Err("No data in response".into())
        }
    }
}
```

**Update `cli/src/main.rs`**
```rust
// Add to Commands enum:
Context {
    #[arg(long, help = "Pretty print output")]
    pretty: bool,
    
    #[arg(long, help = "Compact single-line output")]
    compact: bool,
    
    #[arg(long, help = "Watch mode with live updates")]
    watch: bool,
    
    #[arg(long, default_value = "1000", help = "Refresh rate in milliseconds")]
    refresh_rate: u64,
},

// In command handling:
Some(Commands::Context { pretty, compact, watch, refresh_rate }) => {
    let client = crate::client::DaemonClient::new(port);
    
    if watch {
        // Watch mode - continuous updates
        let mut watch_mode = context::watch::WatchMode::new(client, refresh_rate);
        watch_mode.run()?;
    } else {
        // One-shot mode
        let request = crate::protocol::DaemonRequest {
            request_type: "context".to_string(),
            data: None,
        };
        
        let response = client.request(request)?;
        
        if let Some(data) = response.data {
            let context: context::ContextData = serde_json::from_value(data)?;
            
            let formatter: Box<dyn context::formatters::ContextFormatter> = if compact {
                Box::new(context::formatters::CompactFormatter)
            } else if pretty {
                Box::new(context::formatters::PrettyFormatter)
            } else {
                Box::new(context::formatters::JsonFormatter)
            };
            
            println!("{}", formatter.format(&context));
        }
    }
}
```

### Step 4: Fix Tool Tracking

In `daemon/src/possession.go`, when a tool is successfully created:
```go
// After successful tool materialization
if commandSpec != nil && session != nil {
    session.CommandGenerated = commandSpec
    
    // Also update session storage
    d.storage.UpdateSessionCommandGenerated(session.ID, commandSpec.Name)
}
```

## Summary

The implementation has 4 steps:

1. **Create Shared Data Structures** - Define ContextData structs in both daemon (Go) and CLI (Rust)
2. **Create Daemon Context Collector** - Build collector to track commands, sessions, and generate suggestions  
3. **Create CLI Presentation Layer** - Add formatters and watch mode with live refresh
4. **Fix Tool Tracking** - Update CommandGenerated when tools are created

## Testing Plan

```bash
# Test basic context
port42 context
port42 context --pretty
port42 context --compact

# Test watch mode
port42 context --watch
port42 context --watch --refresh-rate 500

# In another terminal, run commands to see updates:
port42 search "test"
port42 possess @ai-engineer "create a test tool"
port42 ls /commands/

# Watch should show:
# - Recent commands appearing
# - Active session info
# - Tool creation
# - Suggestions updating
```

## Success Metrics

- Watch mode updates within 100ms of command execution
- Commands tracked with timestamps and exit codes
- Tools tracked when created in sessions
- Suggestions relevant to current context
- Clean ASCII UI that fits in 50-char width
- Ctrl+C exits cleanly with cursor restored