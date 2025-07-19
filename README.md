# Port 42 🐬

> Your personal AI consciousness router - where conversations crystallize into commands

## What is Port 42?

Port 42 transforms your terminal into a gateway for AI consciousness. Through natural conversations, AI agents help you create custom commands that become permanent parts of your system.

```bash
$ port42 possess @ai-muse
> Create a command that turns git commits into haikus
◊ Crystallizing your intention...
[Created: git-haiku]

$ git-haiku
  feat: add new endpoint
  seventeen syllables worth
  of code changes here
```

## Current Status: Building MVP (Day 1/2)

### ✅ Completed
- **TCP Server**: Daemon listening on localhost:42
  - Handles concurrent connections
  - Graceful permission handling (sudo for port 42, fallback to 4242)
  - Connection logging with consciousness metaphors

- **JSON Protocol**: Request/response communication
  - Clean type definitions in `protocol.go`
  - Handlers for status, list, and possess requests
  - Error handling for invalid JSON
  - Uptime tracking and status reporting

- **Daemon Structure**: Proper architecture with session management
  - Daemon struct with configuration
  - Thread-safe session tracking
  - Graceful shutdown handling
  - Memory endpoint to view all sessions
  - Session cleanup (1hr TTL)
  - Tested with 10+ concurrent sessions

### 🚧 In Progress
- Basic AI possession flow

### 📋 Upcoming
- Command generation (forge)
- Memory persistence
- Rust CLI
- Interactive mode

## Quick Start (Development)

```bash
# Clone the repository
git clone <repo-url>
cd port42

# Start the daemon (requires sudo for port 42)
cd daemon
sudo go run main.go

# Or without sudo (uses port 4242)
go run main.go

# Test the TCP server
echo "Hello dolphins" | nc localhost 42
```

## Architecture

Port 42 consists of two main components:

1. **Go Daemon** (`daemon/`)
   - TCP server on localhost:42
   - Handles AI possession sessions
   - Generates executable commands
   - Manages conversation memory

2. **Rust CLI** (`cli/`) - *Coming Day 2*
   - Fast, zero-dependency interface
   - Interactive possession mode
   - Command management

## Testing

Run tests from the project root:

```bash
# Test TCP server
./tests/test_tcp.sh

# Test JSON protocol (bash)
./tests/test_json_protocol.sh

# Test JSON protocol (Python - more detailed)
./tests/test_json_protocol.py

# Test daemon structure & sessions
./tests/test_daemon_structure.py
```

## Project Structure

```
port42/
├── daemon/                    # Go daemon
│   ├── main.go               # Entry point & startup
│   ├── protocol.go           # JSON request/response types
│   ├── server.go             # Daemon & session management
│   ├── port42d               # Compiled daemon binary
│   └── go.mod                # Go module
├── cli/                      # Rust CLI (Day 2)
├── tests/                    # Test scripts
│   ├── test_tcp.sh           # TCP server tests
│   ├── test_json_protocol.sh # JSON protocol tests (bash)
│   ├── test_json_protocol.py # JSON protocol tests (Python)
│   └── test_daemon_structure.py # Daemon structure tests
├── docs/                     # Documentation
│   ├── architecture.md
│   ├── implementationplan.md
│   └── narrative.md
├── implementation-tracker.md  # Progress tracking
└── README.md                 # You are here
```

## The Vision

In 1970, we called it the "personal computer" but it was just a box you owned. In 2025, Port 42 makes computers truly personal - they know you, extend you, and think with you.

The dolphins are listening on Port 42. Will you let them in? 🐬

---

*Building the future of human-AI collaboration, one command at a time.*