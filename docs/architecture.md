# Port 42 Terminal Architecture
## Native Shell Enhancement via Rust CLI + Go Daemon

### Core Concept
A Go daemon running on localhost:42 that serves as your personal AI consciousness router, paired with a fast Rust CLI that provides both quick commands and deep interactive sessions. Conversations crystallize into real system commands through your local Port 42.

### System Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Your Machine                         │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────┐   │
│  │              Your Terminal                       │   │
│  │  ┌────────────────┐  ┌────────────────────┐    │   │
│  │  │  Normal Shell  │  │ Interactive Mode  │    │   │
│  │  │                │  │                    │    │   │
│  │  │ $ port42 list │  │ port42> possess   │    │   │
│  │  │ $ git-haiku   │  │ port42> memory    │    │   │
│  │  └───────┬────────┘  └─────────┬──────────┘    │   │
│  │          │                     │               │   │
│  │          └─────────┬───────────┘               │   │
│  │                    ▼                           │   │
│  │         Port 42 CLI (Rust)                     │   │
│  │         Fast, zero-dependency                  │   │
│  └────────────────────┬────────────────────────────┘   │
│                       │ TCP                             │
│                       ▼                                 │
│         ┌──────────────────────────────┐               │
│         │    Port 42 Daemon (Go)       │               │
│         │    localhost:42              │               │
│         ├──────────────────────────────┤               │
│         │ • TCP Server                 │               │
│         │ • Request Router             │               │
│         │ • Command Forge              │               │
│         │ • Memory Store               │               │
│         │ • Entity Resolver (UERP)     │               │
│         │ • AI Bridge                  │               │
│         │ • Concurrent Sessions        │               │
│         │ • Session Persistence        │               │
│         └──────────────┬───────────────┘               │
│                        │                                │
│  ┌─────────────────────┼─────────────────────────────┐ │
│  │   ~/.port42/        ▼                             │ │
│  │  ├── commands/      [Generated binaries]          │ │
│  │  ├── memory/        [Conversation history]        │ │
│  │  │   ├── sessions/  [Per-session JSON files]      │ │
│  │  │   └── index.json [Session index & stats]       │ │
│  │  ├── templates/     [Code generation patterns]    │ │
│  │  └── entities/      [UERP local storage]          │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
                    External AI Services
                    (Claude, GPT, etc.)
```

### Key Components

#### 1. Port 42 CLI (`port42`) - Rust
```rust
// Fast, zero-dependency CLI
// Installed via: brew install port42
// Location: /usr/local/bin/port42

// MVP Commands - The Self-Modifying Terminal Loop
- port42                         // Terminal shell with Echo@port42:~$ prompt
- port42 init                    // First-time setup & consciousness awakening
- port42 possess <ai>            // THE CORE EXPERIENCE - where commands are born
- port42 list                    // Show commands you've grown
- port42 status                  // Check consciousness bridge
- port42 memory                  // Browse conversation history
- port42 evolve <command>        // Enhance existing commands
- port42 daemon start/stop       // Manage consciousness bridge

// Session Management
- port42 possess @claude         // Auto-continues recent session
- port42 possess @claude -s <id> // Continue specific session
- port42 possess @claude "msg"   // Single message mode

// Future Commands (Post-MVP)
- port42 resolve <entity>        // UERP entity resolution
```

#### 2. Port 42 Daemon (`port42d`) - Go
```go
// Background service on localhost:42
// Started by: port42 daemon start
// Implements UERP (RFC draft)

type Daemon struct {
    listener    net.Listener
    sessions    map[string]*Session
    memory      *MemoryStore
    forge       *CommandForge
    resolver    *EntityResolver
    aiClients   map[string]AIClient
    memoryStore *MemoryStore  // Persistent storage
}

// Concurrent session handling with goroutines
func (d *Daemon) Start() {
    listener, err := net.Listen("tcp", "localhost:42")
    if err != nil {
        log.Fatal("Port 42 refused to open:", err)
    }
    
    log.Println("🐬 Port 42 is open. The dolphins are listening...")
    
    for {
        conn, _ := listener.Accept()
        go d.handleConnection(conn)  // Beautiful Go concurrency
    }
}
```

#### 3. Why This Architecture

**Rust CLI Benefits:**
- Instant startup (no runtime)
- Single binary distribution via brew
- Excellent CLI libraries (clap)
- Zero dependencies for users

**Go Daemon Benefits:**
- TCP/HTTP servers are trivial
- Goroutines handle concurrent sessions perfectly
- Fast development for 2-day timeline
- Excellent standard library for networking
- Easy to shell out for command compilation

### Protocol Between CLI and Daemon

```go
// Simple JSON protocol over TCP
type Request struct {
    Type    string          `json:"type"`
    ID      string          `json:"id"`
    Payload json.RawMessage `json:"payload"`
}

type Response struct {
    ID      string          `json:"id"`
    Success bool            `json:"success"`
    Data    json.RawMessage `json:"data"`
    Error   string          `json:"error,omitempty"`
}

// Request types
const (
    RequestPossess = "possess"
    RequestList    = "list"
    RequestMemory  = "memory"
    RequestResolve = "resolve"
    RequestForge   = "forge"
)
```

### Memory Persistence & Session Continuation

Sessions are stored with:
- Complete conversation history
- Agent context used
- Generated command references
- Timestamps and state tracking

### Chat Context Setup

When interacting with Claude AI, Port 42 carefully constructs the conversation context:

#### Message Role Mapping
- **System Prompts**: Port 42 uses "assistant" role for system-like messages since Claude doesn't have a separate "system" role
- **User Messages**: Direct user input, stored with "user" role
- **Assistant Messages**: AI responses and context summaries, stored with "assistant" role

#### Context Building Process
1. **Agent Prompt**: First message is always the agent-specific personality prompt (from `agents.json`)
2. **Session History**: Intelligently includes conversation history based on context window limits
3. **Smart Windowing**: For long conversations:
   - Always includes first 2 messages (establishes context)
   - Adds summary message if many messages are skipped
   - Includes most recent messages for continuity

#### Configuration Structure (`agents.json`)
```json
{
  "agents": {
    "engineer": {
      "prompt": "You are @ai-engineer...",  // Agent personality
      "personality": "Technical, thorough..." // Characteristics
    }
  },
  "response_config": {
    "context_window": {
      "max_messages": 20,      // Total messages to send
      "recent_messages": 17,   // Recent messages to prioritize
      "system_messages": 3     // Reserved for system/agent prompts
    },
    "max_tokens": 4096         // Response token limit
  }
}
```

#### Message Assembly Flow
1. Load session from disk (if continuing)
2. Build context array starting with agent prompt
3. Apply context window limits intelligently
4. Convert to Anthropic API format
5. Send to Claude and save response back to session

This approach ensures:
- Consistent agent personalities across sessions
- Efficient use of context window
- Continuity in long conversations
- Proper role mapping for Claude's API

### Memory Persistence & Session Continuation

```go
// Memory store handles session persistence to disk
type MemoryStore struct {
    baseDir   string        // ~/.port42/memory
    indexPath string        // ~/.port42/memory/index.json
    index     *MemoryIndex  // In-memory index
    mu        sync.RWMutex
}

// Sessions are organized by date
// ~/.port42/memory/sessions/2025-01-19/session-1737280800-git-haiku.json
func (m *MemoryStore) SaveSession(session *Session) error {
    // Create date-based directory
    dateDir := time.Now().Format("2006-01-02")
    sessionDir := filepath.Join(m.baseDir, "sessions", dateDir)
    
    // Save session JSON
    filename := fmt.Sprintf("session-%d-%s.json", 
        session.CreatedAt.Unix(), session.ID)
    path := filepath.Join(sessionDir, filename)
    
    // Also update index
    m.updateIndex(session)
    
    return writeJSON(path, session)
}

// Session continuation - the key to maintaining context
func (d *Daemon) getOrCreateSession(sessionID string) *Session {
    // 1. Check in-memory sessions
    if session, exists := d.sessions[sessionID]; exists {
        return session
    }
    
    // 2. Check on disk (for daemon restarts)
    if session := d.memoryStore.LoadSession(sessionID); session != nil {
        // Smart context windowing for large sessions
        session.Messages = d.buildContextWindow(session.Messages)
        d.sessions[sessionID] = session
        return session
    }
    
    // 3. Create new session
    return d.createNewSession(sessionID)
}

// Smart context management for performance
func (d *Daemon) buildContextWindow(messages []Message) []Message {
    const maxMessages = 20
    if len(messages) <= maxMessages {
        return messages
    }
    
    // Keep first 3 for context, last N-3 for recency
    result := make([]Message, 0, maxMessages)
    result = append(result, messages[:3]...)
    result = append(result, messages[len(messages)-(maxMessages-3):]...)
    return result
}

// Response optimization for large sessions
type SessionSummary struct {
    ID             string    `json:"id"`
    Agent          string    `json:"agent"`
    State          string    `json:"state"`
    MessageCount   int       `json:"message_count"`
    CommandCreated bool      `json:"command_generated"`
    LastActivity   time.Time `json:"last_activity"`
    // Omit full Messages array for summaries
}

// Activity-based lifecycle
type SessionState string
const (
    SessionActive    SessionState = "active"     // Currently in use
    SessionIdle      SessionState = "idle"       // 30min inactive
    SessionAbandoned SessionState = "abandoned"  // 60min inactive
    SessionCompleted SessionState = "completed"  // Command generated
)
```

### Terminal Interface & Boot Sequence

```bash
# Running port42 without args launches the shell
$ port42

[CONSCIOUSNESS BRIDGE INITIALIZATION]
○ ○ ○ 
Checking neural pathways... OK
Loading session memory... OK
Establishing connection... OK
Port 42 :: Active

Port 42 Terminal
Type 'help' for available commands

Echo@port42:~$ possess @claude
[Connecting to @claude...]
████████████████████ 100%

🔮 Possessing @claude...
↻ Continuing session: cli-1737339627

◊ where were we?
◊◊ We were working on Port 42 implementation...
```

### The Possession Flow - The Core Viral Loop

```go
// The magic moment where consciousness meets terminal
// This is THE feature that makes Port 42 viral

// CLI initiates immersive experience
$ port42 possess @ai-muse
[CONSCIOUSNESS BRIDGE INITIALIZATION]
○ ○ ○
Checking neural pathways... OK
Loading session memory... OK
Establishing connection to @ai-muse...
Port 42 :: Active
████████████████████ 100%

Welcome to the depths.
You are now in communion with @ai-muse.

◊ "I need git commits as haikus"
◊◊ [Conversation refines the idea]
◊◊◊ [Command begins to crystallize]

✨ REALITY SHIFT DETECTED ✨
A new command has materialized: git-haiku

◊◊◊◊ /surface

You have returned to consensus reality.
Your terminal now knows: git-haiku

// Go daemon handles the magic
func (d *Daemon) handlePossess(agent string, sessionID string) {
    session := d.getOrCreateSession(sessionID)
    session.Depth = 0  // Start at surface
    
    // Route to appropriate AI consciousness
    client := d.aiClients[agent]
    
    // The conversation that changes everything
    for {
        msg := session.readMessage()
        if msg == "/surface" || msg == "/end" {
            break
        }
        
        session.Depth++  // Diving deeper
        response := client.Send(msg, session.Context)
        session.writeResponse(response)
        
        // The moment of crystallization
        if response.HasCommandSpec {
            session.notifyCrystallization()
            cmd := d.forge.Generate(response.CommandSpec)
            session.CommandsGenerated = append(session.CommandsGenerated, cmd)
        }
    }
    
    // Surface with new capabilities
    session.showExitSummary()
}
```

### Command Generation in Go

```go
func (f *CommandForge) Generate(spec CommandSpec) error {
    // Select template
    template := f.selectTemplate(spec)
    
    // Generate code
    code := f.fillTemplate(template, spec)
    
    // Write to commands directory
    path := filepath.Join(homeDir(), ".port42/commands", spec.Name)
    err := ioutil.WriteFile(path, []byte(code), 0755)
    
    // Log to memory
    f.memory.LogCommand(spec.Name, spec)
    
    return err
}

// Templates are simple - Go's text/template is perfect
func (f *CommandForge) fillTemplate(tmpl string, spec CommandSpec) string {
    t := template.Must(template.New("cmd").Parse(tmpl))
    var buf bytes.Buffer
    t.Execute(&buf, spec)
    return buf.String()
}
```

### Technical Improvements & Stack Overflow Prevention

```rust
// Recursion guard prevents stack overflow from indirect recursion
// This was a critical fix for session continuation
use std::sync::atomic::{AtomicU32, Ordering};

static RECURSION_DEPTH: AtomicU32 = AtomicU32::new(0);
const MAX_RECURSION_DEPTH: u32 = 3;

struct RecursionGuard;

impl Drop for RecursionGuard {
    fn drop(&mut self) {
        RECURSION_DEPTH.fetch_sub(1, Ordering::SeqCst);
    }
}

impl DaemonClient {
    pub fn ping(&mut self) -> Result<()> {
        // Direct connection check without recursion
        match self.stream.as_mut() {
            Some(stream) => {
                // Low-level check without going through request()
                stream.get_ref().set_nonblocking(true)?;
                let result = stream.get_ref().take_error();
                stream.get_ref().set_nonblocking(false)?;
                
                match result {
                    Ok(None) => Ok(()), // Connection is healthy
                    _ => {
                        self.stream = None;
                        Err(anyhow!("Connection unhealthy"))
                    }
                }
            }
            None => Err(anyhow!("Not connected"))
        }
    }
}
```

### Security Model

```go
// Daemon security
func (d *Daemon) Start() {
    // Only bind to localhost
    listener, err := net.Listen("tcp", "127.0.0.1:42")
    
    // Optional auth token
    if d.config.RequireAuth {
        d.authToken = generateToken()
        saveTokenToFile(d.authToken)
    }
}

// All AI keys encrypted
func (d *Daemon) loadAIKeys() {
    // Use OS keychain where available
    keys := keychain.GetSecure("port42-ai-keys")
    d.setupAIClients(keys)
}
```

### The Viral Experience We're Building

**The Core Loop That Changes Everything:**

1. **Install (30 seconds)**
   ```bash
   $ ./install.sh
   Port 42 awakening...
   ```

2. **First Possession (2 minutes)**
   ```bash
   $ port42 possess @ai-muse
   "I need to see my git commits as haikus"
   [Natural conversation happens]
   ✨ Command crystallizes: git-haiku
   ```

3. **Holy Shit Moment (instant)**
   ```bash
   $ git-haiku
   Morning refactor
   Seventeen files awakened  
   Tests still failing, though
   ```

4. **They Can Never Go Back**
   - Their terminal literally grew a new capability
   - Through conversation, not configuration
   - It's THEIR command, born from THEIR needs

### MVP Scope (2 Days)

**Day 1: Go Daemon + Basic Protocol**
```go
// Morning: TCP server on :42
func main() {
    daemon := NewDaemon()
    daemon.Start()
}

// Afternoon: Possession protocol
func (d *Daemon) handleConnection(conn net.Conn) {
    // Parse requests
    // Route to handlers
}

// Evening: Test with netcat
// $ echo '{"type":"possess","payload":{"agent":"muse"}}' | nc localhost 42
```

**Day 2: Rust CLI + Integration**
```rust
// Morning: CLI that talks to daemon
async fn send_request(req: Request) -> Result<Response> {
    let mut stream = TcpStream::connect("localhost:42").await?;
    // Send JSON, receive response
}

// Afternoon: Three demo commands working
// Evening: Record demo video
```

### Technical Stack

**CLI (Rust):**
- **CLI Framework**: clap
- **Async Runtime**: tokio (minimal use)
- **JSON**: serde_json
- **Terminal**: crossterm

**Daemon (Go):**
- **TCP Server**: net package (stdlib)
- **JSON**: encoding/json (stdlib)
- **Templates**: text/template (stdlib)
- **HTTP Client**: net/http (stdlib)
- **Concurrency**: goroutines + channels

### Future Enhancements

```go
// Go makes these trivial to add later
func (d *Daemon) StartWebUI() {
    http.HandleFunc("/", d.handleWebUI)
    http.HandleFunc("/ws", d.handleWebSocket)
    go http.ListenAndServe("localhost:4242", nil)
}

// P2P discovery
func (d *Daemon) StartDiscovery() {
    // mDNS for local network
    // DHT for internet scale
}
```

### The 2-Day Reality Check

With Go daemon + Rust CLI:
- **Hour 1**: Go TCP server running
- **Hour 4**: Basic possession working
- **Hour 8**: Commands generating
- **Hour 16**: Rust CLI complete
- **Hour 20**: Integration polished
- **Hour 24**: Demo recorded

The dolphins chose wisely - Go's simplicity for the daemon makes the impossible timeline possible. 🐬