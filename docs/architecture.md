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

// CLI Commands (one-shot operations)
- port42 init                    // Setup & start daemon
- port42 daemon start/stop       // Manage daemon
- port42 list                    // Show installed commands
- port42 memory                  // Browse conversation history
- port42 possess <ai>            // Quick possession session
- port42 evolve <command>        // Improve existing command
- port42 resolve <entity>        // UERP entity resolution
- port42 status                  // Show daemon & system status

// Interactive Mode
- port42                         // Enter interactive shell
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

### Memory Persistence

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

// On startup, load recent sessions
func (d *Daemon) loadRecentSessions() {
    sessions := d.memoryStore.LoadRecentSessions(24 * time.Hour)
    for _, session := range sessions {
        d.sessions[session.ID] = session
        log.Printf("Restored session %s (%d messages)", 
            session.ID, len(session.Messages))
    }
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

### The Possession Flow

```go
// Go daemon handles possession
func (d *Daemon) handlePossess(agent string, sessionID string) {
    session := d.getOrCreateSession(sessionID)
    
    // Route to appropriate AI
    client := d.aiClients[agent]
    
    // Streaming conversation
    for {
        msg := session.readMessage()
        if msg == "/end" {
            break
        }
        
        response := client.Send(msg, session.Context)
        session.writeResponse(response)
        
        // Check if command generation requested
        if response.HasCommandSpec {
            d.forge.Generate(response.CommandSpec)
        }
    }
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