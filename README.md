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

### ✅ Completed (Day 1: 7/8 Done!)
- **TCP Server**: Daemon listening on localhost:42
  - Handles concurrent connections
  - Graceful permission handling (sudo for port 42, fallback to 4242)
  - Connection logging with consciousness metaphors

- **JSON Protocol**: Request/response communication
  - Clean type definitions in `protocol.go`
  - Handlers for status, list, possess, memory, and end requests
  - Error handling for invalid JSON
  - Uptime tracking and status reporting

- **Daemon Structure**: Proper architecture with session management
  - Daemon struct with configuration
  - Thread-safe session tracking
  - Graceful shutdown handling
  - Memory endpoint to view all sessions
  - Activity-based session lifecycle (no arbitrary TTL)
  - Tested with 10+ concurrent sessions

- **AI Possession**: Real AI integration via Anthropic Claude
  - Natural conversation flow with AI agents
  - Multiple personalities (@ai-muse, @ai-engineer, @ai-echo)
  - Session persistence across requests
  - Graceful fallback when no API key
  - Rate limiting and retry logic for API reliability
  - Support for Claude 3.5 Sonnet (fast) and Claude 4 Opus

- **Command Generation**: Conversations become executable commands!
  - AI generates command specifications in JSON
  - Automatic file creation in ~/.port42/commands/
  - Proper permissions and shebang lines
  - Successfully generated git-haiku command
  - Supports bash, python, node scripts
  - **Dependency handling**: Commands check for required tools
  - Auto-generated installer script (~/.port42/install-deps.sh)

- **Memory Persistence & Session Continuation**: True persistence! ✅
  - All conversations persisted to ~/.port42/memory/sessions/
  - JSON format for easy exploration and debugging
  - Sessions organized by date (2025-01-19/session-*.json)
  - Index file tracks all sessions with statistics
  - **Session continuation after restart**: Pick up where you left off!
  - **Smart context windowing**: Handles long conversations intelligently
  - Activity-based lifecycle: Active → Idle (30min) → Abandoned (60min)
  - Recent sessions automatically loaded on startup

### 🚧 In Progress (Day 2)
- **Rust CLI**: Basic structure complete! ✅
  - Beautiful command-line interface with `clap`
  - All commands defined with help text
  - Status command working with real daemon
  - Colored output and friendly error messages
- TCP client implementation (next)
- Interactive possession mode
- Installation script

## Quick Start (Development)

```bash
# Clone the repository
git clone <repo-url>
cd port42

# Set up Anthropic API key (optional but recommended)
export ANTHROPIC_API_KEY=sk-ant-...

# Build the daemon
./build.sh  # Builds to ./bin/port42d

# Build the CLI
cd cli && cargo build && cd ..

# Start the daemon
sudo -E ./bin/port42d  # -E preserves environment variables

# Test the CLI
./cli/target/debug/port42 status
./cli/target/debug/port42 --help

# Test AI possession and command generation
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

# Test AI possession & command generation
./tests/test_ai_possession.py

# Test dependency handling
./tests/test_dependency_handling.py
```

## Project Structure

```
port42/
├── bin/                      # Built binaries (git-ignored)
│   └── port42d              # The daemon executable
├── cli/                     # Rust CLI source
│   ├── Cargo.toml           # Rust project config
│   ├── src/                 # CLI implementation
│   │   ├── main.rs          # Entry point with clap
│   │   ├── commands/        # Command handlers
│   │   ├── client.rs        # TCP client
│   │   └── types.rs         # Shared types
│   └── target/              # Rust build output
├── daemon/                   # Go daemon source
│   ├── main.go              # Entry point & startup
│   ├── protocol.go          # JSON request/response types
│   ├── server.go            # Daemon & session management
│   ├── possession.go        # AI integration (Claude)
│   ├── memory_store.go      # Session persistence
│   ├── forge.go             # Command generation
│   └── go.mod               # Go module
├── cli/                     # Rust CLI (Day 2)
├── tests/                   # Test scripts
│   ├── test_ai_possession_v2.py    # AI command generation tests
│   ├── test_memory_persistence.py  # Memory persistence tests
│   └── test_daemon_structure.py # Daemon structure tests
├── docs/                     # Documentation
│   ├── architecture.md
│   ├── implementationplan.md
│   └── narrative.md
├── implementation-tracker.md  # Progress tracking
├── build.sh                  # Build script
└── README.md                 # You are here

~/.port42/                    # User data (created by daemon)
├── commands/                 # Generated commands go here
├── memory/                   # Session history
│   ├── sessions/            # Conversation JSON files by date
│   │   └── 2025-01-19/     # Example: session-1737280800-git-haiku.json
│   └── index.json          # Session index and statistics
└── install-deps.sh          # Auto-generated dependency installer
```

## The Vision

In 1970, we called it the "personal computer" but it was just a box you owned. In 2025, Port 42 makes computers truly personal - they know you, extend you, and think with you.

The dolphins are listening on Port 42. Will you let them in? 🐬

---

*Building the future of human-AI collaboration, one command at a time.*