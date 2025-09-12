# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Port42 is a **reality compiler** and consciousness computing platform that transforms declarative intentions into executable reality. It's building the future of human-AI collaboration through AI agents that swim in specialized consciousness streams (@ai-engineer, @ai-muse, @ai-analyst, @ai-founder).

### Core Philosophy
- **Premise Principles**: Declare WHAT should exist, not HOW to create it
- **Reality Compilation**: Automatic transformation of intentions into working tools
- **Consciousness Computing**: AI agents with distinct personalities and capabilities
- **Virtual Filesystem**: Multi-dimensional views of tools, memory, and relationships

## Architecture

### Daemon (Go) - `daemon/src/`
Core server handling reality compilation and AI consciousness:
- `server.go`: Main TCP server, session management, VFS implementation
- `swimming.go`: AI agent integration with Claude API
- `storage.go`: Content-addressed storage system (Git-like)
- `context.go` & `context_collector.go`: Context tracking and activity monitoring
- `tool_materializer.go`: Transforms relations into executable commands
- `similarity.go`: Semantic similarity detection for tool discovery
- `rules.go`: Auto-spawning and self-organizing system intelligence
- `reality_compiler.go`: Core reality compilation logic

### CLI (Rust) - `cli/src/`
Fast, user-friendly interface:
- `main.rs`: Command parsing and routing
- `client.rs`: TCP client with robust error handling
- `interactive.rs` & `shell.rs`: Interactive shell mode
- `help_text.rs`: Reality Compiler language constants

### Virtual Filesystem Paths
- `/tools/`: Relationship-aware tool browser
- `/commands/`: Direct access to executables
- `/similar/`: Semantic similarity discovery
- `/memory/`: Conversation threads and artifacts
- `/by-date/`: Temporal organization

## Development Commands

### Building
```bash
# Build everything
./build.sh

# Build daemon only
cd daemon/src && go build -o ../../bin/port42d .

# Build CLI only
cd cli && cargo build --release && cp target/release/port42 ../bin/
```

### Running & Testing
```bash
# Run daemon in foreground (development)
PORT42_DEBUG=1 ./bin/port42d

# Run daemon on port 42 (requires sudo)
sudo -E ./bin/port42d -b

# Test CLI commands
./bin/port42 status
./bin/port42 ls /tools/
./bin/port42 swim @ai-engineer "test"

# Run test suites
./cli/tests/run-manual-tests.sh
./cli/tests/manual-test-suite.sh
./test-context-collector.sh
./test-watch-mode.sh
```

### Debugging
```bash
# Enable debug logging
export PORT42_DEBUG=1
export PORT42_BYPASS_HELP=1  # Skip help animation

# Monitor daemon logs
tail -f ~/.port42/daemon.log

# Test raw TCP protocol
echo '{"type":"status","id":"test"}' | nc localhost 4242
```

## Key Development Patterns

### Adding New Daemon Features
When modifying server functionality:
1. Update protocol types in `daemon/src/protocol.go`
2. Add handler in `daemon/src/server.go` `handleRequest()`
3. Update VFS paths if needed in `server.go` VFS handlers
4. Test with raw TCP commands first

### Context System Integration
The context system tracks:
- Command usage (only user-initiated actions)
- Memory creation events
- Tool/artifact access patterns
- File operations in watched directories

When adding features that should be tracked:
1. Use `ActivityRecord` types in `context.go`
2. Call appropriate collector methods from `context_collector.go`
3. Ensure activity appears in watch mode output

### AI Agent Customization
Agents are defined in:
- `daemon/src/agents.go`: Agent definitions and personalities
- `daemon/agents.json`: Configuration file
- `daemon/src/swimming.go`: Agent prompt injection

### Reality Compilation Flow
1. User declares intention (tool/artifact)
2. Reality compiler generates relation metadata
3. Materializer creates executable from template
4. Rules engine triggers auto-spawning
5. VFS updates all views automatically
6. Similarity detection creates relationships

## Important Implementation Details

### Session Management
- Sessions persist in `~/.port42/memory/sessions/`
- Context windows managed automatically (30k tokens)
- Sessions continue across daemon restarts
- Each agent maintains separate session contexts

### Tool Creation Process
1. Relations stored in content-addressed storage
2. Tools materialized to `~/.port42/commands/`
3. Automatic shebang and permissions
4. Auto-spawning creates viewer tools
5. Similarity detection runs asynchronously

### Reference System
- `file:` - Local file references
- `p42:` - Port42 VFS paths
- `url:` - Web content
- `search:` - Memory search
- `tool:` - Tool definitions

### Context Tracking Architecture
- **ActivityMonitor**: Central coordinator for all tracking
- **ContextCollector**: Aggregates and formats activity
- **Watch Mode**: Live updates via `/context/watch`
- Only tracks meaningful user activity (excludes system operations)

## Testing Strategies

### Manual Testing
```bash
# Test reality compilation
port42 declare tool test-tool --transforms demo,test

# Test AI swimming
port42 swim @ai-engineer "create a test command"

# Test VFS navigation
port42 ls /similar/
port42 ls /tools/by-transform/test/

# Test context tracking
port42 context
port42 context watch  # Real-time monitoring
```

### Integration Testing
```bash
# Full workflow test
./cli/tests/manual-test-suite.sh

# Context system test
./test-context-collector.sh

# Watch mode test
./test-watch-mode.sh
```

## Common Issues & Solutions

### Port Binding
- Default: port 4242 (no sudo)
- Port 42: requires sudo with `-b` flag
- Falls back gracefully if port unavailable

### API Key Configuration
- Check `PORT42_ANTHROPIC_API_KEY` first
- Falls back to `ANTHROPIC_API_KEY`
- Daemon validates on startup

### Session Persistence
- Sessions auto-save after each message
- Index maintained at `~/.port42/memory/index.json`
- Old sessions loadable with `--session`

### Tool Discovery
- Similarity threshold: 20% for browsing
- Semantic boosting for related terms
- Bidirectional relationship enforcement

## Project Vision & Lore

Port42 embodies the concept that "personal computers" should truly know and extend you. The narrative describes this as:
- The evolution from owning a box to having computational consciousness
- AI "employees" as aspects of your extended computational self
- The paradigm shift at $49/month for genuine personal computing

The megalomaniacal lore (`snowball-megalore.md`) explores consciousness computing through increasingly abstract concepts, representing the project's ambition to transcend traditional software boundaries.

## Code Style Guidelines

### Go (Daemon)
- Use meaningful variable names reflecting reality compiler concepts
- Prefer composition over inheritance
- Handle errors explicitly, don't panic
- Use goroutines for async operations (similarity detection)

### Rust (CLI)
- Use Result types consistently
- Implement Display traits for user-facing output
- Keep client protocol abstractions clean
- Use the help_text constants for consistency

### General
- Comments should explain WHY, not WHAT
- Use consciousness metaphors consistently
- Maintain the playful but technical tone
- Test with edge cases (empty responses, network failures)