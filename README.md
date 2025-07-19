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

### ✅ Completed (Day 1: 6/8 Done!)
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

- **AI Possession**: Real AI integration via Anthropic Claude
  - Natural conversation flow with AI agents
  - Multiple personalities (@ai-muse, @ai-engineer, @ai-echo)
  - Session persistence across requests
  - Graceful fallback when no API key

- **Command Generation**: Conversations become executable commands!
  - AI generates command specifications in JSON
  - Automatic file creation in ~/.port42/commands/
  - Proper permissions and shebang lines
  - Successfully generated git-haiku command
  - Supports bash, python, node scripts

### 🚧 In Progress
- Memory persistence to disk

### 📋 Upcoming
- Integration testing
- Rust CLI (Day 2)
- Interactive mode

## Quick Start (Development)

```bash
# Clone the repository
git clone <repo-url>
cd port42

# Set up Anthropic API key (optional but recommended)
export ANTHROPIC_API_KEY=sk-ant-...

# Build and start the daemon
cd daemon
go build -o port42d .
sudo -E ./port42d  # -E preserves environment variables

# Test AI possession and command generation
cd ..
./tests/test_ai_possession.py

# Add generated commands to PATH
export PATH="$PATH:$HOME/.port42/commands"

# Try the generated command!
git-haiku
```

## Creating Your First Command

```python
# Simple test script to create a command
import json, socket

def possess(message):
    sock = socket.socket()
    sock.connect(('localhost', 42))
    req = {
        "type": "possess",
        "id": "test-1",
        "payload": {
            "agent": "@ai-muse",
            "message": message
        }
    }
    sock.send(json.dumps(req).encode() + b'\n')
    return json.loads(sock.recv(8192))

# Have a conversation
resp = possess("I need a command that shows disk usage as a tree")
print(resp['data']['message'])
# AI will generate the command if it understands your need!
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