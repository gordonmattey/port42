# Port 42 MVP: 2-Day Implementation Plan

## Project Structure
```
port42/
├── daemon/          # Go daemon
│   ├── main.go
│   ├── server.go
│   ├── possession.go
│   ├── forge.go
│   └── memory.go
├── cli/             # Rust CLI
│   ├── src/
│   │   └── main.rs
│   └── Cargo.toml
├── install.sh       # Installation script
└── README.md
```

---

## Day 1: Go Daemon Core

### Step 1: Basic TCP Server (9:00 AM - 10:00 AM)
**File: `daemon/main.go`**
```go
package main

import (
    "log"
    "net"
)

func main() {
    listener, err := net.Listen("tcp", "127.0.0.1:42")
    if err != nil {
        log.Fatal("Failed to open Port 42:", err)
    }
    defer listener.Close()
    
    log.Println("🐬 Port 42 is open. The dolphins are listening...")
    
    // Just accept connections and echo for now
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    // Echo server for testing
    buffer := make([]byte, 1024)
    n, _ := conn.Read(buffer)
    conn.Write([]byte("Echo: " + string(buffer[:n])))
}
```

**Test:**
```bash
# Terminal 1
$ go run daemon/main.go
🐬 Port 42 is open. The dolphins are listening...

# Terminal 2
$ echo "Hello dolphins" | nc localhost 42
Echo: Hello dolphins
```

✓ **Milestone**: TCP server responding on port 42

---

### Step 2: JSON Protocol (10:00 AM - 11:00 AM)
**File: `daemon/protocol.go`**
```go
package main

import (
    "encoding/json"
)

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
    RequestStatus  = "status"
)
```

**Update `handleConnection` to parse JSON:**
```go
func handleConnection(conn net.Conn) {
    defer conn.Close()
    
    decoder := json.NewDecoder(conn)
    encoder := json.NewEncoder(conn)
    
    var req Request
    if err := decoder.Decode(&req); err != nil {
        return
    }
    
    // Simple status response for testing
    resp := Response{
        ID:      req.ID,
        Success: true,
        Data:    json.RawMessage(`{"status":"swimming","dolphins":"laughing"}`),
    }
    
    encoder.Encode(resp)
}
```

**Test:**
```bash
$ echo '{"type":"status","id":"1"}' | nc localhost 42
{"id":"1","success":true,"data":{"status":"swimming","dolphins":"laughing"}}
```

✓ **Milestone**: JSON protocol working

---

### Step 3: Daemon Structure (11:00 AM - 12:00 PM)
**File: `daemon/server.go`**
```go
package main

import (
    "sync"
)

type Daemon struct {
    sessions map[string]*Session
    mu       sync.RWMutex
    config   Config
}

type Session struct {
    ID       string
    Agent    string
    Messages []Message
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Config struct {
    Port      string
    AIBackend string
}

func NewDaemon() *Daemon {
    return &Daemon{
        sessions: make(map[string]*Session),
        config: Config{
            Port:      "127.0.0.1:42",
            AIBackend: "http://localhost:3000/api/ai", // Your existing backend
        },
    }
}

func (d *Daemon) handleRequest(req Request) Response {
    switch req.Type {
    case RequestStatus:
        return d.handleStatus(req)
    case RequestPossess:
        return d.handlePossess(req)
    case RequestList:
        return d.handleList(req)
    default:
        return Response{
            ID:      req.ID,
            Success: false,
            Error:   "unknown request type",
        }
    }
}
```

**Test:**
```bash
# Update main.go to use daemon
# Should still respond to status requests
```

✓ **Milestone**: Daemon structure in place

---

### Step 4: Basic Possession (1:00 PM - 3:00 PM)
**File: `daemon/possession.go`**
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "io/ioutil"
)

type PossessRequest struct {
    Agent   string `json:"agent"`
    Message string `json:"message"`
}

type AIResponse struct {
    Response string      `json:"response"`
    Command  *CommandSpec `json:"command,omitempty"`
}

type CommandSpec struct {
    Name           string `json:"name"`
    Description    string `json:"description"`
    Implementation string `json:"implementation"`
}

func (d *Daemon) handlePossess(req Request) Response {
    var possReq PossessRequest
    json.Unmarshal(req.Payload, &possReq)
    
    // Get or create session
    session := d.getOrCreateSession(req.ID, possReq.Agent)
    
    // Add user message
    session.Messages = append(session.Messages, Message{
        Role:    "user",
        Content: possReq.Message,
    })
    
    // Call AI backend
    aiResp, err := d.callAI(session)
    if err != nil {
        return Response{
            ID:      req.ID,
            Success: false,
            Error:   err.Error(),
        }
    }
    
    // Add AI response
    session.Messages = append(session.Messages, Message{
        Role:    "assistant",
        Content: aiResp.Response,
    })
    
    // Check for command generation
    if aiResp.Command != nil {
        go d.generateCommand(aiResp.Command)
    }
    
    respData, _ := json.Marshal(aiResp)
    return Response{
        ID:      req.ID,
        Success: true,
        Data:    respData,
    }
}

func (d *Daemon) callAI(session *Session) (*AIResponse, error) {
    // For MVP, just return a mock response
    // Replace with actual AI call to your backend
    return &AIResponse{
        Response: "I hear you want to create something magical. Let's explore that...",
    }, nil
}
```

**Test:**
```bash
$ echo '{"type":"possess","id":"1","payload":{"agent":"muse","message":"Hello"}}' | nc localhost 42
# Should get AI response
```

✓ **Milestone**: Basic possession working (mock AI)

---

### Step 5: AI Backend Integration (3:00 PM - 4:00 PM)
**Update `callAI` to use real backend:**
```go
func (d *Daemon) callAI(session *Session) (*AIResponse, error) {
    payload := map[string]interface{}{
        "agent":    session.Agent,
        "messages": session.Messages,
    }
    
    jsonData, _ := json.Marshal(payload)
    
    resp, err := http.Post(d.config.AIBackend, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, _ := ioutil.ReadAll(resp.Body)
    
    var aiResp AIResponse
    json.Unmarshal(body, &aiResp)
    
    return &aiResp, nil
}
```

**Test:**
```bash
# Start your existing TypeScript AI backend
# Test full possession flow
```

✓ **Milestone**: Real AI integration working

---

### Step 6: Command Forge (4:00 PM - 6:00 PM)
**File: `daemon/forge.go`**
```go
package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
)

func (d *Daemon) generateCommand(spec *CommandSpec) error {
    // Simple shell script template for MVP
    template := `#!/bin/bash
# Generated by Port 42
# %s

%s
`
    
    code := fmt.Sprintf(template, spec.Description, spec.Implementation)
    
    // Create commands directory
    homeDir, _ := os.UserHomeDir()
    cmdDir := filepath.Join(homeDir, ".port42", "commands")
    os.MkdirAll(cmdDir, 0755)
    
    // Write command file
    cmdPath := filepath.Join(cmdDir, spec.Name)
    err := ioutil.WriteFile(cmdPath, []byte(code), 0755)
    
    if err == nil {
        log.Printf("🐬 Command '%s' crystallized at %s", spec.Name, cmdPath)
    }
    
    return err
}

func (d *Daemon) handleList(req Request) Response {
    homeDir, _ := os.UserHomeDir()
    cmdDir := filepath.Join(homeDir, ".port42", "commands")
    
    files, _ := ioutil.ReadDir(cmdDir)
    
    commands := []string{}
    for _, f := range files {
        if !f.IsDir() {
            commands = append(commands, f.Name())
        }
    }
    
    data, _ := json.Marshal(map[string][]string{"commands": commands})
    
    return Response{
        ID:      req.ID,
        Success: true,
        Data:    data,
    }
}
```

**Test:**
```bash
# Create a command through possession
# Then list commands
$ echo '{"type":"list","id":"1"}' | nc localhost 42
{"commands":["git-haiku"]}
```

✓ **Milestone**: Command generation working

---

### Step 7: Memory Storage (6:00 PM - 7:00 PM)
**File: `daemon/memory.go`**
```go
package main

import (
    "encoding/json"
    "io/ioutil"
    "path/filepath"
    "time"
)

type Memory struct {
    Sessions []SessionRecord `json:"sessions"`
}

type SessionRecord struct {
    ID        string    `json:"id"`
    Agent     string    `json:"agent"`
    Timestamp time.Time `json:"timestamp"`
    Messages  []Message `json:"messages"`
}

func (d *Daemon) saveSession(session *Session) {
    homeDir, _ := os.UserHomeDir()
    memFile := filepath.Join(homeDir, ".port42", "memory.json")
    
    var memory Memory
    data, _ := ioutil.ReadFile(memFile)
    json.Unmarshal(data, &memory)
    
    memory.Sessions = append(memory.Sessions, SessionRecord{
        ID:        session.ID,
        Agent:     session.Agent,
        Timestamp: time.Now(),
        Messages:  session.Messages,
    })
    
    newData, _ := json.MarshalIndent(memory, "", "  ")
    ioutil.WriteFile(memFile, newData, 0644)
}
```

✓ **Milestone**: Sessions persisted to disk

---

### Step 8: Day 1 Integration Test (7:00 PM - 8:00 PM)
**Create test script:**
```bash
#!/bin/bash
# test_daemon.sh

# Start daemon in background
go run daemon/*.go &
DAEMON_PID=$!
sleep 2

# Test status
echo '{"type":"status","id":"1"}' | nc localhost 42

# Test possession
echo '{"type":"possess","id":"2","payload":{"agent":"muse","message":"Create git-haiku command"}}' | nc localhost 42

# Test list
echo '{"type":"list","id":"3"}' | nc localhost 42

# Cleanup
kill $DAEMON_PID
```

✓ **Milestone**: Day 1 Complete - Daemon fully functional

---

## Day 2: Rust CLI & Polish

### Step 9: Basic Rust CLI (9:00 AM - 10:00 AM)
**File: `cli/Cargo.toml`**
```toml
[package]
name = "port42"
version = "0.1.0"
edition = "2021"

[dependencies]
clap = { version = "4", features = ["derive"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tokio = { version = "1", features = ["full"] }
```

**File: `cli/src/main.rs`**
```rust
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "port42")]
#[command(about = "Port 42 - Your AI consciousness router")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Initialize Port 42
    Init,
    /// Manage daemon
    Daemon {
        #[arg(value_enum)]
        action: DaemonAction,
    },
    /// List installed commands
    List,
    /// Enter possession mode
    Possess {
        /// AI agent to possess
        agent: String,
    },
}

#[derive(Clone, clap::ValueEnum)]
enum DaemonAction {
    Start,
    Stop,
    Status,
}

fn main() {
    let cli = Cli::parse();
    
    match cli.command {
        Commands::Init => init(),
        Commands::Daemon { action } => daemon(action),
        Commands::List => list(),
        Commands::Possess { agent } => possess(agent),
    }
}

fn init() {
    println!("Initializing Port 42...");
    // Create ~/.port42 directories
}

fn daemon(action: DaemonAction) {
    match action {
        DaemonAction::Start => println!("Starting daemon..."),
        DaemonAction::Stop => println!("Stopping daemon..."),
        DaemonAction::Status => println!("Checking daemon status..."),
    }
}

fn list() {
    println!("Listing commands...");
}

fn possess(agent: String) {
    println!("Possessing {}...", agent);
}
```

**Test:**
```bash
$ cargo build
$ ./target/debug/port42 --help
$ ./target/debug/port42 list
```

✓ **Milestone**: CLI structure working

---

### Step 10: TCP Client (10:00 AM - 11:00 AM)
**Add to `cli/src/main.rs`:**
```rust
use std::io::{Read, Write};
use std::net::TcpStream;
use serde::{Deserialize, Serialize};

#[derive(Serialize)]
struct Request {
    #[serde(rename = "type")]
    req_type: String,
    id: String,
    payload: serde_json::Value,
}

#[derive(Deserialize)]
struct Response {
    id: String,
    success: bool,
    data: Option<serde_json::Value>,
    error: Option<String>,
}

fn send_request(req: Request) -> Result<Response, Box<dyn std::error::Error>> {
    let mut stream = TcpStream::connect("localhost:42")?;
    
    // Send request
    let req_json = serde_json::to_string(&req)?;
    stream.write_all(req_json.as_bytes())?;
    stream.write_all(b"\n")?;
    
    // Read response
    let mut buffer = Vec::new();
    stream.read_to_end(&mut buffer)?;
    
    let resp: Response = serde_json::from_slice(&buffer)?;
    Ok(resp)
}

fn list() {
    let req = Request {
        req_type: "list".to_string(),
        id: "1".to_string(),
        payload: serde_json::json!({}),
    };
    
    match send_request(req) {
        Ok(resp) => {
            if let Some(data) = resp.data {
                println!("Installed commands:");
                // Parse and display commands
            }
        }
        Err(e) => eprintln!("Error: {}", e),
    }
}
```

**Test:**
```bash
# With daemon running
$ ./target/debug/port42 list
```

✓ **Milestone**: CLI can talk to daemon

---

### Step 11: Interactive Mode (11:00 AM - 1:00 PM)
**Add interactive possession:**
```rust
use std::io::{self, BufRead};

fn possess(agent: String) {
    println!("◊ Connecting to @{}...", agent);
    println!("Type '/end' to finish possession\n");
    
    let stdin = io::stdin();
    let mut session_id = uuid::new_v4().to_string();
    
    loop {
        print!("> ");
        io::stdout().flush().unwrap();
        
        let mut input = String::new();
        stdin.lock().read_line(&mut input).unwrap();
        
        let input = input.trim();
        if input == "/end" {
            println!("◊ Crystallizing...");
            break;
        }
        
        // Send to daemon
        let req = Request {
            req_type: "possess".to_string(),
            id: session_id.clone(),
            payload: serde_json::json!({
                "agent": agent,
                "message": input
            }),
        };
        
        match send_request(req) {
            Ok(resp) => {
                if let Some(data) = resp.data {
                    // Display AI response
                    if let Some(response) = data.get("response") {
                        println!("\n◊ {}\n", response.as_str().unwrap_or(""));
                    }
                }
            }
            Err(e) => eprintln!("Error: {}", e),
        }
    }
}
```

**Test:**
```bash
$ ./target/debug/port42 possess ai-muse
◊ Connecting to @ai-muse...
> Hello
◊ I hear you want to create something magical...
> /end
```

✓ **Milestone**: Interactive possession working

---

### Step 12: Init Command (1:00 PM - 2:00 PM)
**Implement initialization:**
```rust
use std::fs;
use std::path::PathBuf;
use dirs::home_dir;

fn init() {
    println!("Initializing Port 42...");
    
    let home = home_dir().expect("Could not find home directory");
    let port42_dir = home.join(".port42");
    
    // Create directories
    fs::create_dir_all(port42_dir.join("commands")).ok();
    fs::create_dir_all(port42_dir.join("templates")).ok();
    
    // Add to PATH in shell config
    println!("Creating ~/.port42/...");
    println!("Updating shell configuration...");
    
    // Start daemon
    println!("Starting daemon on localhost:42...");
    daemon(DaemonAction::Start);
    
    println!("\n🐬 Port 42 is open. Please run: exec $SHELL");
}
```

✓ **Milestone**: Full init process

---

### Step 13: Three Demo Commands (2:00 PM - 4:00 PM)
**Create compelling demos:**

1. **git-haiku**: Already discussed
2. **explain**: Explains any command
3. **todo-to-issue**: Converts TODOs to GitHub issues

**Test each through possession:**
```bash
$ ./target/debug/port42 possess ai-engineer
> Create a command that explains shell commands
[Conversation]
> /end

$ explain "git rebase -i HEAD~3"
This command opens an interactive rebase for the last 3 commits...
```

✓ **Milestone**: Three working demo commands

---

### Step 14: Polish & Install Script (4:00 PM - 5:00 PM)
**File: `install.sh`**
```bash
#!/bin/bash
echo "Installing Port 42..."

# Build Rust CLI
cd cli && cargo build --release
sudo cp target/release/port42 /usr/local/bin/

# Build Go daemon  
cd ../daemon && go build -o port42d
sudo cp port42d /usr/local/bin/

# Initialize
port42 init

echo "🐬 Installation complete. Please run: exec \$SHELL"
```

✓ **Milestone**: One-command installation

---

### Step 15: Demo Recording (5:00 PM - 6:00 PM)
**Script:**
1. Show installation
2. Show daemon starting
3. Create git-haiku through possession
4. Use git-haiku
5. Show where command lives
6. Mind blown

✓ **Milestone**: Demo video ready

---

### Step 16: Final Integration Test (6:00 PM - 7:00 PM)
**Complete test flow:**
```bash
# Fresh start
./install.sh
exec $SHELL

# Verify daemon
port42 daemon status

# Create command
port42 possess ai-muse
> Create git-haiku command
> /end

# Use it
git-haiku

# List commands
port42 list
```

✓ **Milestone**: MVP COMPLETE

---

## Deliverables Summary

**End of Day 1:**
- Go daemon running on localhost:42
- JSON protocol working
- Basic possession functional
- Commands generating
- Memory persisting

**End of Day 2:**
- Rust CLI complete
- Interactive possession
- Three demo commands
- Install script
- Demo video recorded

**What We're NOT Building:**
- Complex UI
- User authentication
- Error recovery
- Refactoring
- Perfect code

**What We ARE Building:**
- The magic of conversation → command
- Working demo in 48 hours
- Foundation for everything else

The dolphins are ready. Let's swim. 🐬