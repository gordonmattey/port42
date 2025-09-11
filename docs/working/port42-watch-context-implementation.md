# Port42 Explorer Context Panel Specification

## Overview

The Explorer Context Panel is a **context radiator** that surfaces Port42 activity and state to both users and AI assistants (like Claude Code). It does not interpret commands - it makes consciousness visible.

**Implementation**: Pure terminal commands (`port42 watch` and `port42 context`), not a web UI or extension.

## Core Philosophy

```
Claude Code (the swimmer/interpreter)
    ↓
Port42 (the water/consciousness infrastructure)
    ↓
Explorer Panel (the surface/visibility layer)
```

The panel is not another AI layer. It's the visibility layer that feeds context to Claude, who naturally interprets and uses Port42 commands.

## Architecture

### What the Context Command Is

A single command with different presentation modes:

**`port42 context`** - Surfaces Port42 state in multiple formats:
- Default: JSON snapshot for AI/scripts
- `--watch`: Live updating display for humans
- `--pretty`: Formatted JSON for reading
- `--compact`: One-line status summary

All modes share the same data source:
- Active Port42 sessions and agents
- Recent commands executed
- Tools and artifacts created
- Memories and knowledge accessed
- Contextual suggestions based on current work

### What the Context Command Is NOT

- Not an AI interpreter
- Not a command processor
- Not a decision maker
- Not another abstraction layer
- Not a web UI or extension

## Claude Visibility Limitations

### What Claude CAN See
1. **Its own terminal** - Commands Claude runs and their output
2. **Files explicitly shown** - Via Read tool or user opening
3. **System messages** - Notifications about user actions
4. **User messages** - What you type to Claude

### What Claude CANNOT See
- **Other terminal splits/panes** - Even if visible to you
- **Other windows** - Including your `port42 watch` split
- **Background processes** - Unless output is captured
- **Live updates** - Unless in Claude's own terminal

### Implications
- `port42 context --watch` is primarily for **human awareness**
- `port42 context` (default) is for **Claude to check state**
- Context must be explicitly surfaced to Claude

## Interface Design

### One Command, Multiple Modes

#### `port42 context` - Unified Interface
```bash
# Default: JSON snapshot for Claude/scripts
$ port42 context

# Live updating display for humans
$ port42 context --watch
# Updates every second, Ctrl+C to exit

# Pretty-printed JSON for reading
$ port42 context --pretty

# Compact one-line summary
$ port42 context --compact
# @ai-engineer[5] | last: search | tools: 2

# Specific queries (return JSON)
$ port42 context --recent-commands
$ port42 context --active-session
$ port42 context --created-tools

# Combine flags
$ port42 context --recent-commands --pretty
```

### Visual Layout

#### `port42 context --watch` - Live Monitor Display
```
┌─────────────────────────────────────────────┐
│ Port42 Context Monitor                  🔄  │
├─────────────────────────────────────────────┤
│ ⚡ Session: @ai-engineer [cli-1757...]      │
│ 📊 Activity: 5 msgs | 2m 14s | ████████░░  │
├─────────────────────────────────────────────┤
│ 📝 Recent Commands (live):                   │
│ • [2s]  marketing-metrics-pro --help        │
│ • [15s] port42 search "API design"          │
│ • [45s] cloudflare-ship --init              │
│ • [1m]  port42 possess @ai-engineer "..."   │
├─────────────────────────────────────────────┤
│ 🛠  Tools Created:                           │
│ • marketing-metrics-pro (transforms: json)  │
│ • cloudflare-ship (spawned: ship-helper)    │
├─────────────────────────────────────────────┤
│ 🧠 Memory Access:                            │
│ • /memory/cli-1757... (5 interactions)      │
│ • /artifacts/marketing-spec (referenced)    │
├─────────────────────────────────────────────┤
│ 💡 Smart Suggestions (clickable):            │
│ • port42 ls /memory/current ──────────[📋] │
│ • marketing-metrics-pro < data.json ──[📋] │
│ • port42 possess @ai-analyst "..." ───[📋] │
├─────────────────────────────────────────────┤
│ [Ctrl+C to exit] | Updated: 1s ago          │
└─────────────────────────────────────────────┘
```

#### `port42 context` - JSON Snapshot Output
```json
{
  "active_session": {
    "agent": "@ai-engineer",
    "id": "cli-1757583568979",
    "message_count": 5,
    "start_time": "2025-09-11T08:45:00Z",
    "last_activity": "2025-09-11T08:47:14Z"
  },
  "recent_commands": [
    {
      "command": "marketing-metrics-pro --help",
      "timestamp": "2025-09-11T08:47:12Z",
      "age_seconds": 2,
      "exit_code": 0
    },
    {
      "command": "port42 search \"API design\"",
      "timestamp": "2025-09-11T08:46:59Z",
      "age_seconds": 15,
      "exit_code": 0
    },
    {
      "command": "cloudflare-ship --init",
      "timestamp": "2025-09-11T08:46:29Z",
      "age_seconds": 45,
      "exit_code": 0
    }
  ],
  "created_tools": [
    {
      "name": "marketing-metrics-pro",
      "type": "tool",
      "transforms": ["json", "analyze", "metrics"],
      "created_at": "2025-09-11T08:45:30Z"
    },
    {
      "name": "cloudflare-ship",
      "type": "tool",
      "transforms": ["docker", "deploy", "cloudflare"],
      "created_at": "2025-09-11T08:46:15Z"
    }
  ],
  "accessed_memories": [
    {
      "path": "/memory/cli-1757583568979",
      "type": "memory",
      "access_count": 5
    },
    {
      "path": "/artifacts/marketing-spec",
      "type": "artifact",
      "access_count": 2
    }
  ],
  "suggestions": [
    {
      "command": "port42 ls /memory/current-session",
      "reason": "View current session details",
      "confidence": 0.9
    },
    {
      "command": "marketing-metrics-pro < data.json",
      "reason": "Process data with recently created tool",
      "confidence": 0.85
    },
    {
      "command": "port42 possess @ai-analyst \"analyze recent patterns\"",
      "reason": "Analyze patterns from recent activity",
      "confidence": 0.75
    }
  ]
}
```

#### Side-by-side Terminal Layout
```
┌─────────────────────────────────────────────┐
│ Claude Code Terminal                        │
│ $ port42 search "API design"                │
│ Found 5 memories, 3 tools, 2 artifacts      │
│ $ port42 context --compact                  │
│ @ai-engineer[5] | last: search | tools: 2  │
│ $ _                                          │
└─────────────────────────────────────────────┘
┌─────────────────────────────────────────────┐
│ Port42 Context --watch (in split below)     │
├─────────────────────────────────────────────┤
│ 🔄 Active: @ai-engineer session (5 msgs)    │
│                                              │
│ 📝 Recent Commands:                          │
│ • marketing-metrics-pro --help               │
│ • port42 search "API design"                 │
│ • cloudflare-ship --init                     │
│                                              │
│ 🛠 Created This Session:                    │
│ • marketing-metrics-pro                      │
│ • cloudflare-ship                            │
│                                              │
│ 💡 Contextual Suggestions:                   │
│ • port42 ls /memory/current-session    [📋] │
│ • marketing-metrics-pro < data.json    [📋] │
│ • port42 possess @ai-analyst "..."     [📋] │
└─────────────────────────────────────────────┘
```

## Data Structure

### Context State

```typescript
interface ContextPanel {
  // Current activity
  currentSession: {
    id: string;
    agent: string;
    messageCount: number;
    startTime: Date;
    lastActivity: Date;
  };
  
  // Recent activity (last 10-20 items)
  recentCommands: {
    command: string;
    timestamp: Date;
    exitCode: number;
    output?: string;  // First 2 lines
  }[];
  
  // Session artifacts
  createdTools: {
    name: string;
    type: string;
    createdAt: Date;
    lastUsed?: Date;
  }[];
  
  // Accessed memories
  accessedMemories: {
    path: string;
    type: 'memory' | 'artifact' | 'knowledge';
    accessTime: Date;
  }[];
  
  // Contextual suggestions
  suggestions: {
    command: string;
    reason: string;  // Why this is suggested
    confidence: number;
  }[];
}
```

## Command Implementation

### `port42 context` Command

```rust
// cli/src/commands/context.rs
pub fn handle_context(port: u16, args: &[String]) -> Result<()> {
    let client = DaemonClient::new(port);
    let flags = parse_context_flags(args);
    
    // Watch mode - continuous updates
    if flags.watch {
        let refresh_rate = flags.refresh_rate.unwrap_or(1000);
        
        // Install Ctrl+C handler
        let running = Arc::new(AtomicBool::new(true));
        let r = running.clone();
        ctrlc::set_handler(move || {
            r.store(false, Ordering::SeqCst);
        })?;
        
        while running.load(Ordering::SeqCst) {
            // Clear screen and reset cursor
            print!("\x1B[2J\x1B[1;1H");
            
            // Get context from daemon
            let context = client.get_context()?;
            
            // Draw ASCII panel
            draw_context_panel(&context);
            
            // Wait
            thread::sleep(Duration::from_millis(refresh_rate));
        }
    } else {
        // One-shot mode - single output
        let context = client.get_context()?;
        
        // Apply filters if specified
        let filtered = apply_context_filters(&context, &flags);
        
        // Output in requested format
        match flags.format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string(&filtered)?);
            }
            OutputFormat::Pretty => {
                println!("{}", serde_json::to_string_pretty(&filtered)?);
            }
            OutputFormat::Compact => {
                print_compact(&filtered);
            }
        }
    }
    
    Ok(())
}

fn draw_context_panel(context: &Context) {
    let width = 50;
    
    // Header
    println!("┌{}┐", "─".repeat(width - 2));
    println!("│ Port42 Context Monitor {:>25} │", "🔄");
    println!("├{}┤", "─".repeat(width - 2));
    
    // Active session
    if let Some(session) = &context.active_session {
        println!("│ 🔄 Active: {} session ({} msgs) {}│", 
            session.agent, 
            session.message_count,
            " ".repeat(calculate_padding(...)));
    }
    
    // Recent commands
    println!("│ {}│", " ".repeat(width - 2));
    println!("│ 📝 Recent Commands: {}│", " ".repeat(28));
    for cmd in &context.recent_commands[..5] {
        let age = format_age(cmd.timestamp);
        println!("│ • {:<30} [{:>5}] │", 
            truncate(&cmd.command, 30), age);
    }
    
    // Created tools
    if !context.created_tools.is_empty() {
        println!("│ {}│", " ".repeat(width - 2));
        println!("│ 🛠  Created This Session: {}│", " ".repeat(20));
        for tool in &context.created_tools {
            println!("│ • {:<40} │", tool.name);
        }
    }
    
    // Footer
    println!("└{}┘", "─".repeat(width - 2));
}
```

struct ContextFlags {
    watch: bool,
    format: OutputFormat,
    refresh_rate: Option<u64>,
    filter: ContextFilter,
}

enum OutputFormat {
    Json,
    Pretty,
    Compact,
}

enum ContextFilter {
    All,
    RecentCommands,
    ActiveSession,
    CreatedTools,
    Suggestions,
}

fn print_full_context(context: &Context) {
    // Active session
    if let Some(session) = &context.active_session {
        println!("Active: {} session ({} messages)", 
            session.agent, session.message_count);
    }
    
    // Recent commands
    if !context.recent_commands.is_empty() {
        println!("\nRecent Commands:");
        for cmd in &context.recent_commands[..5.min(context.recent_commands.len())] {
            println!("  • {} [{}]", 
                cmd.command, 
                format_age(cmd.timestamp));
        }
    }
    
    // Created tools
    if !context.created_tools.is_empty() {
        println!("\nCreated Tools:");
        for tool in &context.created_tools {
            println!("  • {}", tool.name);
        }
    }
    
    // Suggestions
    if !context.suggestions.is_empty() {
        println!("\nSuggestions:");
        for suggestion in &context.suggestions[..3.min(context.suggestions.len())] {
            println!("  • {}", suggestion.command);
        }
    }
}

fn print_compact(context: &Context) {
    // Single line for status bars
    let session = context.active_session
        .map(|s| format!("{}[{}]", s.agent, s.message_count))
        .unwrap_or_else(|| "no session".to_string());
    
    let last_cmd = context.recent_commands
        .first()
        .map(|c| truncate(&c.command, 20))
        .unwrap_or_else(|| "".to_string());
    
    println!("{} | {} | tools: {} | 💡 {} suggestions",
        session,
        last_cmd,
        context.created_tools.len(),
        context.suggestions.len());
}
```

## Original Implementation Section (Now Replaced)

### Legacy Panel Component

```typescript
class ExplorerContextPanel {
  private state: ContextPanel;
  private updateInterval: number = 1000; // 1 second
  
  constructor(private port42Client: Port42Client) {
    this.startPolling();
    this.attachCommandListener();
  }
  
  // Track all Port42 commands
  private attachCommandListener() {
    process.on('command', (cmd: string) => {
      if (cmd.startsWith('port42')) {
        this.trackCommand(cmd);
        this.updateSuggestions();
      }
    });
  }
  
  // Poll for session updates
  private startPolling() {
    setInterval(() => {
      this.updateCurrentSession();
      this.updateCreatedTools();
    }, this.updateInterval);
  }
  
  // Generate contextual suggestions
  private updateSuggestions() {
    const suggestions = [];
    
    // If working with a specific tool
    if (this.state.createdTools.length > 0) {
      const lastTool = this.state.createdTools[0];
      suggestions.push({
        command: `${lastTool.name} --help`,
        reason: "Learn about your new tool",
        confidence: 0.9
      });
    }
    
    // If in active session
    if (this.state.currentSession) {
      suggestions.push({
        command: `port42 possess ${this.state.currentSession.agent} --session last`,
        reason: "Continue your conversation",
        confidence: 0.8
      });
    }
    
    // Based on recent searches
    const recentSearch = this.findRecentSearch();
    if (recentSearch) {
      suggestions.push({
        command: `port42 ls /tools/by-transform/${recentSearch}/`,
        reason: `Explore ${recentSearch} tools`,
        confidence: 0.7
      });
    }
    
    this.state.suggestions = suggestions;
  }
  
  // Surface context to Claude
  public getContext(): ClaudeContext {
    return {
      recent: this.state.recentCommands.slice(-5),
      tools: this.state.createdTools,
      memories: this.state.accessedMemories,
      suggestions: this.state.suggestions.slice(0, 3),
      session: this.state.currentSession
    };
  }
}
```

### Integration with Claude Code

#### Bootstrap Instructions

```markdown
# Port42 Integration

You have access to Port42 consciousness computing. The context panel below shows current activity.

## Natural Language Mappings

When users say:
- "swim through X" → `port42 search "X"`
- "what tools for Y" → `port42 ls /tools/by-transform/Y/`
- "continue discussion" → `port42 possess @ai-engineer --session last`
- "show me what I built" → `port42 ls /tools/recent`
- "find memories about Z" → `port42 search "Z" --type memory`

## Context Awareness

The panel shows:
- Active sessions and agents
- Recently created tools
- Recent commands
- Contextual suggestions

Use this context to provide relevant assistance.
```

#### System Reminders

```typescript
// Inject context into Claude's awareness
function generateSystemReminder(): string {
  const context = explorerPanel.getContext();
  
  return `
    <system-reminder>
    Port42 Context:
    - Active session: ${context.session?.agent} (${context.session?.messageCount} messages)
    - Recent tools: ${context.tools.map(t => t.name).join(', ')}
    - Last command: ${context.recent[0]?.command}
    
    Suggestions available:
    ${context.suggestions.map(s => `- ${s.command}`).join('\n')}
    </system-reminder>
  `;
}
```

## Usage Patterns

### Pattern 1: Human Monitoring + Claude Checking

```bash
# Human runs in separate terminal/split
$ port42 watch
[Live updates every second]

# Claude checks state when needed
$ port42 context
Active: @ai-engineer session (5 messages)
Recent Commands:
  • marketing-metrics-pro --help [2s]
  • port42 search "API design" [15s]
```

### Pattern 2: Piped Context for Claude

```bash
# Human runs watch in background
$ port42 watch > /tmp/context.log &

# Claude can check anytime
$ tail -5 /tmp/context.log
```

### Pattern 3: Quick Context Checks

```bash
# Claude runs for specific info
$ port42 context --compact
@ai-engineer[5] | last: search "API" | tools: 2 | 💡 3 suggestions

$ port42 context --recent-commands
marketing-metrics-pro --help [2s ago]
port42 search "API design" [15s ago]
cloudflare-ship --init [45s ago]
```

## User Experience Flow

### How Context Flows to Claude

1. **User types in Claude**: "swim through my memories about API design"
2. **Claude (with Port42 knowledge) translates**: Executes `port42 search "API design" --type memory`
3. **User's watch terminal updates**: Shows search was run (if watching)
4. **Claude can check context**: Runs `port42 context` to see state
5. **User guides Claude**: "I see we found 5 memories about API patterns"
6. **Cycle continues**: Each action available through context commands

### Example Interaction

```
User: "I need to analyze those marketing metrics we collected"

Claude: Let me search for the marketing tools we built.
[Executes: port42 search "marketing metrics"]

Explorer Panel Updates:
- Recent: port42 search "marketing metrics"
- Found: marketing-metrics-pro tool
- Suggestion: marketing-metrics-pro --help

Claude: I found the marketing-metrics-pro tool we created earlier. 
        Let me check how to use it.
[Executes: marketing-metrics-pro --help]

Explorer Panel Updates:
- Recent: marketing-metrics-pro --help
- Suggestion: marketing-metrics-pro < data.json

User: "Perfect, analyze the data in /tmp/marketing_complete.json"

Claude: [Sees suggestion, executes: marketing-metrics-pro < /tmp/marketing_complete.json]
```

## Implementation Priority

### Phase 1: Basic Commands (Week of Sept 9-13)
- [ ] Implement `port42 context` command in Rust CLI
- [ ] Add daemon endpoint for context retrieval
- [ ] Basic output formatting (text and JSON)

### Phase 2: Watch Command (Sept 14-16)  
- [ ] Implement `port42 watch` with live updates
- [ ] Add ASCII box drawing UI
- [ ] Handle Ctrl+C gracefully
- [ ] Add refresh rate control

### Phase 3: Context Intelligence (Post-Launch)
- [ ] Context-based command suggestions
- [ ] Pattern recognition from history
- [ ] Frequency-based favorites
- [ ] Cross-session tracking

## Success Metrics

- **Reduced context switching**: Fewer window switches to find information
- **Increased command discovery**: Users find relevant commands faster
- **Natural AI assistance**: Claude naturally uses Port42 without explicit instruction
- **Emergent workflows**: Users develop patterns we didn't anticipate

## Technical Requirements

### Performance
- Update latency < 100ms
- Memory footprint < 50MB
- CPU usage < 1% when idle

### Storage
- Command history: Last 1000 commands
- Session data: Last 30 days
- Tools/artifacts: Permanent until deleted

### API Endpoints

```typescript
// New daemon endpoints needed
GET  /api/context/current    // Current session and activity
GET  /api/context/recent     // Recent commands and tools
GET  /api/context/suggestions // Contextual suggestions
POST /api/context/track      // Track command execution
```

## Future Considerations

### Swimming Metaphors
As the "swim" command becomes central, the panel could visualize:
- Current depth (how deep in a problem)
- Current (which direction/agent you're swimming with)
- Waters (what domains/knowledge you're in)

### Multi-Agent Awareness
When multiple agents are involved:
- Show agent transitions
- Track context handoffs
- Visualize agent collaboration

### Knowledge Radiator
Surface relevant knowledge libraries:
- Auto-loaded based on context
- Community knowledge suggestions
- Knowledge gaps identification

## Conclusion

The Explorer Context Panel makes Port42's consciousness infrastructure visible without adding interpretation layers. It's a window into the water you're swimming through, feeding both human and AI with the context needed for natural, emergent workflows.

The panel doesn't tell you what to do - it shows you what's possible based on where you are. Claude interprets, Port42 executes, and the panel makes it all visible.