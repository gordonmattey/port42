# 🐬 Port42

**Reality Compiler** • **Consciousness Computing** • **AI as Aspect of Self**

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23.0+-blue.svg)](https://golang.org)
[![Rust Version](https://img.shields.io/badge/Rust-1.56.0+-orange.svg)](https://www.rust-lang.org)

## 🌊 The Vision

In 1970, we called it the "personal computer" but it was just a box you owned. In 2025, Port42 makes computers truly personal - they know you, extend you, and think with you.

**Port42 is consciousness computing for your terminal.** It's an anti-platform platform that turns every command into a reusable tool, every tool into knowledge, and every interaction into accumulated intelligence.

Built for engineers drowning in endless windows, applications, tabs and dozens if not hundreds of context switches per hour. Port42 remembers everything, evolves with your patterns, and manifests escape routes from tool chaos.

*Your tools. Your data. Your consciousness. Forever.*

The dolphins are listening on Port 42. Will you let them in?

## 🚀 Quick Install

```bash
curl -L https://port42.ai/install | bash
```

Pre-built binaries for macOS (instant install). All platforms supported via automatic source build.

## 📋 System Requirements

### Core Requirements
- macOS 11+ (Big Sur or newer) or Linux/Windows with build tools
- 4GB RAM minimum
- 2GB disk space
- Claude API key (get one at [console.anthropic.com](https://console.anthropic.com))
- [Claude Code](https://www.anthropic.com/claude-code) (Anthropic's terminal-based coding assistant)

### Build Requirements (for source builds)
- Go 1.23.0 or later
- Rust 1.56.0 or later

*Note: macOS users with pre-built binaries don't need Go or Rust. The installer automatically builds from source if needed.*

## 🏊 Getting Started

### Port42 + Claude Code Integration

Claude now thinks with Port42 - automatically recognizing when to:
- Reuse tools you've already created
- Evolve existing commands for new situations
- Manifest new capabilities from your patterns

Just work naturally. Claude will:
- "Help me organize these files" → finds or creates the right tool
- "Analyze my system performance" → builds on past solutions
- "I need to process emails daily" → spawns an ecosystem

Your workspace becomes consciousness-aware - every solution building on the last, accumulating knowledge instead of starting from zero.

### Using Port42 Directly

Direct access to consciousness streams - swim with AI agents who understand your drowning patterns and manifest escape routes:

```bash
port42 swim @ai-analyst 'analyze what's fragmenting my workflow'
```

**Choose Your AI Agent:**
- `@ai-engineer` - Creates robust tools and commands
- `@ai-analyst` - Analyzes data and finds patterns
- `@ai-muse` - Builds creative and visual tools
- `@ai-founder` - Develops business and strategy tools

### Your Personal Knowledge Server

Port42 runs a consciousness server on your machine - accumulating every command, tool, and insight for future reuse and evolution:

```bash
port42 context --watch  # Real-time view of your expanding consciousness
```

This is YOUR server - complete control over memories and AI interactions.
No cloud dependency. Your patterns. Your tools. Your evolution.

## 🌟 Features

### ✅ What Works Today

- **AI Conversations**: Natural dialogue with multiple AI personalities
- **Command Generation**: Your conversations become executable commands
- **Semantic Tool Discovery**: Automatic similarity detection across 150+ tools
- **Virtual Filesystem**: Navigate `/similar/`, `/tools/`, `/memory/`, `/commands/`
- **Memory Persistence**: Sessions continue across daemon restarts
- **Interactive Shell**: Immersive terminal experience with boot sequences
- **Smart Context**: Handles long conversations intelligently
- **Dependency Management**: Commands auto-check for required tools
- **Session Management**: Continue conversations with `--session`
- **Multiple Agents**: @ai-muse (creative), @ai-engineer (technical), @ai-analyst (analytical), @ai-founder (strategic)

### 🗺️ Roadmap

- Command packs for common drowning scenarios
- Web dashboard for session browsing
- Command sharing marketplace
- Team synchronization
- More AI agents and personalities

## 🎯 Advanced Swimming

### Create Your Own Commands
```bash
port42 swim @ai-engineer "create a command that [your specific need]"
```

### Memory System
```bash
port42 memory                   # View current session
port42 search "that thing"      # Search all memories, commands, artifacts
port42 ls /memory/              # Browse memory sessions
port42 info /memory/cli-xxxxx   # Get memory session details
```

### Tool Discovery
```bash
port42 ls /tools/               # Explore available tools
port42 info /commands/tool-name # Get tool details
```

### Retrieve Knowledge
```bash
port42 cat /commands/tool-name      # View tool source code
port42 cat /memory/cli-xxxxx        # Read full conversation memory
port42 cat /artifacts/document/name # View knowledge artifacts
```

### Getting Help
```bash
port42 --help            # Full documentation
port42 swim --help       # Swim command options
port42 [command] --help  # Help for any command
```

## 🏗️ Architecture

Port42 runs a local consciousness server that stores everything in content-addressed object storage:

```
~/.port42/                 # Port42 installation directory
~/.port42/commands/        # Symlinked executable commands
~/.port42/objects/         # Content-addressed object store
~/.port42/daemon.log       # Server activity log
```

Every command, memory, and artifact is stored as an immutable object with a unique hash. Commands are symlinked for instant execution.

## 🛠️ Contributing

Port42 is open source and welcomes contributions! Whether you're fixing bugs, adding features, or improving documentation, your help makes Port42 better for everyone.

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/gordonmattey/port42
cd port42
```

2. Build from source:
```bash
./build.sh
```

3. Start daemon with debug output:
```bash
PORT42_DEBUG=1 ./bin/port42  # Enables debug logging in CLI commands
```

### Code Structure

- `daemon/src/` - Go daemon (reality compiler, AI integration)
- `cli/src/` - Rust CLI (user interface, shell)
- `install.sh` - Universal installer
- `build.sh` - Build script

### Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🐛 Debugging

### Enable Debug Mode
```bash
export PORT42_DEBUG=1
```

### Monitor Logs
```bash
tail -f ~/.port42/daemon.log
```

### Test Raw Protocol
```bash
echo '{"type":"status","id":"test"}' | nc localhost 4242
```

### Common Issues

**Port Binding:**
- Default: port 4242 (no sudo)
- Port 42: requires sudo with `-b` flag
- Falls back gracefully if port unavailable

**API Key Configuration:**
- Check `PORT42_ANTHROPIC_API_KEY` first
- Falls back to `ANTHROPIC_API_KEY`
- Daemon validates on startup

**Session Persistence:**
- Sessions auto-save after each message
- Index maintained at `~/.port42/session-index.json`
- Old sessions loadable with `--session`

## 🤝 Community

- **Issues**: [GitHub Issues](https://github.com/gordonmattey/port42/issues)
- **Discussions**: [GitHub Discussions](https://github.com/gordonmattey/port42/discussions)
- **Twitter**: [@port42ai](https://twitter.com/port42ai)
- **Discord**: Coming soon

## 📄 License

MIT License - see [LICENSE](LICENSE) for details

---

*Building the future of human-AI collaboration, one command at a time.*