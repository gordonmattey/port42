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

## 🚀 Quick Start (For Users)

### Installation

```bash
# Install from official script (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/yourusername/port42/main/install.sh | bash

# Or install locally from source
git clone https://github.com/yourusername/port42.git
cd port42
./install-local.sh
```

### First Steps

```bash
# Check daemon status
port42 status

# Enter the Port 42 shell (recommended)
port42

# Or use commands directly
port42 possess @ai-muse         # Start AI conversation
port42 list                     # List your commands
port42 memory                   # View past conversations
```

### Creating Your First Command

```bash
# Method 1: Interactive shell (recommended)
$ port42
Echo@port42:~$ possess @ai-muse
> Help me create a command that shows disk usage beautifully
[Conversation flows...]
Echo@port42:~$ exit

# Method 2: Direct command
$ port42 possess @ai-muse "Create a command that explains any function in my codebase"
```

### Managing the Daemon

```bash
# The daemon starts automatically after installation

# Manual control
port42 daemon start          # Start daemon
port42 daemon stop           # Stop daemon
port42 daemon restart        # Restart daemon
port42 daemon logs           # View logs
port42 daemon logs -f        # Follow logs (tail -f)

# Run on port 42 (requires sudo)
sudo -E port42 daemon start -b

# Otherwise it runs on port 4242
```

### Continuing Conversations

```bash
# List your past sessions
port42 memory

# Continue a specific session
port42 possess @ai-muse --session myproject

# Sessions persist across daemon restarts!
```

### Setting Your API Key

```bash
# Option 1: During installation (recommended)
# The installer will prompt you

# Option 2: Manual setup
export ANTHROPIC_API_KEY='your-key-here'
port42 daemon restart

# Option 3: Add to shell profile
echo "export ANTHROPIC_API_KEY='your-key-here'" >> ~/.zshrc
source ~/.zshrc
```

## 🛠️ For Developers

### Architecture Overview

Port 42 consists of:
- **Go Daemon** (`daemon/`): TCP server handling AI sessions and command generation
- **Rust CLI** (`cli/`): Fast, user-friendly command-line interface
- **Generated Commands** (`~/.port42/commands/`): Your personalized command library

### Building from Source

```bash
# Prerequisites
# - Go 1.21+ 
# - Rust 1.70+
# - Anthropic API key (optional for testing)

# Clone and build
git clone https://github.com/yourusername/port42.git
cd port42

# Build everything
./build.sh

# Or build individually
cd daemon && go build -o ../bin/port42d
cd ../cli && cargo build --release && cp target/release/port42 ../bin/
```

### Development Workflow

```bash
# 1. Set up development environment
export ANTHROPIC_API_KEY='your-key-here'
export PORT42_DEV=1  # Enables debug logging

# 2. Run daemon in foreground for debugging
./bin/port42d

# 3. In another terminal, test CLI commands
./bin/port42 status
./bin/port42 possess @ai-muse "test message"

# 4. Run test suite
./tests/run_all_tests.sh
```

### Key Components

#### Daemon (`daemon/`)
- `main.go`: Entry point, port binding, signal handling
- `server.go`: TCP server, session management
- `protocol.go`: JSON protocol definitions
- `possession.go`: AI integration (Claude API)
- `memory_store.go`: Session persistence
- `forge.go`: Command generation logic

#### CLI (`cli/src/`)
- `main.rs`: CLI argument parsing with clap
- `commands/`: Command implementations
- `client.rs`: TCP client for daemon communication
- `interactive.rs`: Interactive shell mode
- `boot.rs`: Boot sequence animations

### Extending Port 42

#### Adding a New AI Agent

```go
// In daemon/possession.go
func getAgentPrompt(agent string) string {
    prompts := map[string]string{
        "@your-agent": `You are @your-agent, a specialized AI...
        Your personality and capabilities...`,
    }
    // ...
}
```

#### Adding a New CLI Command

```rust
// In cli/src/main.rs
#[derive(Subcommand)]
enum Commands {
    /// Your new command description
    YourCommand {
        #[arg(short, long)]
        your_arg: String,
    },
    // ...
}

// In cli/src/commands/mod.rs
pub mod your_command;

// Create cli/src/commands/your_command.rs
pub fn handle_your_command(port: u16, your_arg: String) -> Result<()> {
    // Implementation
}
```

#### Protocol Extension

```go
// In daemon/protocol.go
type YourRequest struct {
    Type    string      `json:"type"`    // "your_type"
    ID      string      `json:"id"`
    Payload YourPayload `json:"payload"`
}

// In daemon/server.go handleRequest()
case "your_type":
    // Handle your new request type
```

### Testing

```bash
# Run all tests
./tests/run_all_tests.sh

# Individual test suites
./tests/test_tcp.sh                    # Basic connectivity
./tests/test_json_protocol.py          # Protocol compliance
./tests/test_daemon_structure.py       # Session management
./tests/test_ai_possession.py          # AI integration
./tests/test_memory_persistence.py     # Persistence layer

# Integration testing
./tests/integration/test_full_flow.sh  # End-to-end test
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`./tests/run_all_tests.sh`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Debugging Tips

```bash
# Enable debug logging
export PORT42_DEBUG=1

# Check daemon logs
tail -f ~/.port42/daemon.log

# Test raw TCP connection
echo '{"type":"status","id":"test"}' | nc localhost 42

# Monitor memory usage
watch -n 1 'ps aux | grep port42d'

# Inspect session files
ls -la ~/.port42/memory/sessions/$(date +%Y-%m-%d)/
```

## 📁 Project Structure
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

```
port42/
├── bin/                      # Built binaries (git-ignored)
│   ├── port42d              # The daemon executable  
│   └── port42               # The CLI executable
├── cli/                     # Rust CLI source
│   ├── Cargo.toml           # Rust dependencies
│   └── src/                 # CLI implementation
├── daemon/                  # Go daemon source
│   ├── *.go                 # Daemon implementation
│   └── go.mod               # Go dependencies
├── tests/                   # Test suite
├── docs/                    # Documentation
└── install.sh               # Production installer

~/.port42/                   # User data directory
├── commands/                # Your generated commands
├── memory/                  # Conversation history
│   └── sessions/           # Organized by date
├── daemon.log              # Daemon logs
└── activate.sh             # Shell activation helper
```

## 🌟 Features

### ✅ What Works Today

- **AI Conversations**: Natural dialogue with multiple AI personalities
- **Command Generation**: Your conversations become executable commands
- **Memory Persistence**: Sessions continue across daemon restarts
- **Interactive Shell**: Immersive terminal experience with boot sequences
- **Smart Context**: Handles long conversations intelligently
- **Dependency Management**: Commands auto-check for required tools
- **Session Management**: Continue conversations with `--session`
- **Multiple Agents**: @ai-muse (creative), @ai-engineer (technical), @ai-echo (adaptive)

### 🚧 Coming Soon

- Web dashboard for session browsing
- Command sharing marketplace
- Team synchronization
- More AI agents and personalities

## 🐬 The Vision

In 1970, we called it the "personal computer" but it was just a box you owned. In 2025, Port 42 makes computers truly personal - they know you, extend you, and think with you.

The dolphins are listening on Port 42. Will you let them in?

## 🤝 Community

- **Issues**: [GitHub Issues](https://github.com/yourusername/port42/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/port42/discussions)
- **Twitter**: [@port42ai](https://twitter.com/port42ai)
- **Discord**: Coming soon

## 📄 License

MIT License - see [LICENSE](LICENSE) for details

---

*Building the future of human-AI collaboration, one command at a time.*