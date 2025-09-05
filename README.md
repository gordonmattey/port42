# Port 42 🐬 Reality Compiler

> Transform thoughts into reality through declarative consciousness computing

## What is Port 42?

Port 42 is a **reality compiler** that bridges the gap between intention and reality. Instead of writing code to implement what you want, you simply declare what should exist. The reality compiler automatically handles all implementation details, creating a self-organizing system of tools, relationships, and knowledge.

### Two Ways to Create Reality

**1. Declarative (Instant Reality Creation)**
```bash
$ port42 declare tool git-haiku --transforms git-log,haiku
✨ Relation declared and materialized!
🔨 Tool is ready to use!

$ git-haiku
  feat: add new endpoint  
  seventeen syllables worth
  of code changes here
```

**2. AI-Assisted with Context (Conversational Creation)**
```bash
$ port42 possess @ai-muse
> Create a command that analyzes log files and shows patterns
◊ Crystallizing your intention...
[Created: log-analyzer → view-log-analyzer (auto-spawned)]

$ log-analyzer /var/log/nginx/access.log
$ view-log-analyzer /tmp/analysis-output.json
```

**3. Enhanced with References (Context-Aware Creation)**
```bash
# Reference local files, VFS knowledge, and web content
$ port42 declare tool config-processor --transforms config,validate,format \
  --ref file:./app.json \
  --ref p42:/commands/base-processor \
  --ref url:https://json-schema.org/spec

✨ Tool created with rich contextual knowledge!
🔨 Now processing with understanding of your config structure
```

## 🚀 Quick Start (For Users)

### Installation

#### Prerequisites

**For Binary Installation (Recommended):**
- macOS or Linux (x86_64 or arm64)
- curl or wget
- Anthropic API key (for AI features) - get one at [console.anthropic.com](https://console.anthropic.com/)

**For Building from Source:**
- Go 1.21+ (for daemon)
- Rust/Cargo (for CLI)
- Git

#### Quick Install (Pre-built Binaries)

```bash
# Install latest release (no password required!)
curl -fsSL https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh | bash

# Activate Port 42 in current shell (no restart needed!)
source ~/.port42/activate.sh

# Start the daemon
port42 daemon start

# Verify installation
port42 status
```

The installer automatically:
- ✅ Downloads pre-built binaries for your platform
- ✅ Installs to `~/.port42/bin` (no sudo required)
- ✅ Updates your PATH automatically
- ✅ Configures your API key interactively
- ✅ Integrates with Claude Code (appends to ~/.claude/CLAUDE.md)
- ✅ Creates activation script for easy environment setup

#### Working with Claude Code

After installation, Claude Code automatically uses Port 42 behind the scenes. You don't need to mention "port42" at all!

**Just ask Claude naturally:**
- "I need to analyze these server logs for errors"
- "Create a tool to validate JSON schemas"
- "Find tools for parsing CSV files"
- "Generate a weekly report from this data"

**Claude will automatically:**
- Search existing Port 42 tools for solutions
- Create new tools using Port 42's reality compiler
- Reference existing tools and knowledge for context
- All invisibly - Port 42 just makes Claude smarter!

#### API Key Configuration

Port 42 checks for API keys in this order:
1. `PORT42_ANTHROPIC_API_KEY` - Port 42 specific key (recommended)
2. `ANTHROPIC_API_KEY` - Shared Anthropic key

You can set your key manually if needed:
```bash
export PORT42_ANTHROPIC_API_KEY='your-key-here'
# OR
export ANTHROPIC_API_KEY='your-key-here'
```

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/gordonmattey/port42.git
cd port42

# Option 1: Build and install with one command
./install.sh --build

# Option 2: Build manually, then install
./build.sh          # Creates binaries in ./bin/
./install.sh        # Installs from local ./bin/

# Set your API key and start
export ANTHROPIC_API_KEY='your-key-here'
port42 daemon start
```

#### Manual Build

If you prefer to build components separately:

```bash
# Build daemon (Go)
cd daemon/src
go mod tidy
go build -o ../../bin/port42d .
cd ../..

# Build CLI (Rust)
cd cli
cargo build --release
cp target/release/port42 ../bin/
cd ..

# Install manually
cp bin/port42* ~/.port42/bin/
cp daemon/agents.json ~/.port42/

# Add to PATH in your shell config
echo 'export PATH="$PATH:$HOME/.port42/bin:$HOME/.port42/commands"' >> ~/.zshrc
```

### Reality Compiler Quick Start

```bash
# Check daemon status
port42 status

# Create your first tool (instant)
port42 declare tool hello-world --transforms greeting,demo

# Create context-aware tools with references and custom prompts
port42 declare tool smart-analyzer --transforms analyze,process \
  --ref file:./config.json \
  --ref p42:/commands/existing-analyzer \
  --ref search:"error patterns" \
  --prompt "Build an intelligent analyzer that detects anomalies and provides actionable recommendations"

# Explore the unified filesystem
port42 ls /                     # Root: tools/, commands/, memory/, by-date/, similar/
port42 ls /tools/               # All tools with relationships  
port42 ls /commands/            # Executable view of tools
port42 ls /by-date/2024-01-15/  # Everything created today

# Discover tool similarities (Step 6: Semantic Tool Discovery)
port42 ls /similar/             # All tools with automatic relationships (150+)
port42 ls /similar/hello-world/ # Tools similar to hello-world
port42 ls /similar/analyzer/    # Find all analysis-related tools
port42 ls /similar/parser/      # Find all parsing-related tools

# Navigate tool relationships
port42 ls /tools/hello-world/   # definition, executable, spawned/, parents/
port42 cat /tools/hello-world/definition    # See relation JSON
port42 cat /commands/hello-world            # View executable code
port42 info /tools/hello-world              # Complete metadata

# Relationship traversal
port42 ls /tools/spawned-by/                # Global spawning index
port42 ls /tools/by-transform/greeting/     # Tools by capability
port42 ls /tools/ancestry/                  # Tools with parent chains
```

### AI-Assisted Creation

```bash
# Start interactive AI session
port42 possess @ai-muse         # Creative AI agent
port42 possess @ai-engineer     # Technical implementation
port42 possess @ai-analyst      # Data analysis & insights  
port42 possess @ai-founder      # Business strategy

# View conversation memory
port42 memory                   # List all sessions
port42 ls /memory/              # Browse memory filesystem
port42 info /memory/cli-123     # Session details
port42 search "docker"          # Search everything
```

## 🏗️ Reality Compiler Architecture

Port 42 implements **Premise principles** - declarative reality creation where you specify *what* should exist, not *how* to create it. The system automatically handles all implementation complexity.

### Core Components

**1. Relation Store** - Entity Knowledge Graph
```
Relations define what should exist:
• Tools: Executable capabilities with transforms
• Artifacts: Documents, designs, media content  
• Memory: Conversation threads with AI agents
• Relationships: Parent-child, spawning, semantic links
```

**2. Reality Compiler** - Intention → Reality Bridge
```go
// You declare WHAT should exist
port42 declare tool log-analyzer --transforms parse,analysis

// Reality compiler handles HOW:
// ✅ Generate Python executable template
// ✅ Create filesystem symlinks  
// ✅ Auto-spawn viewer tools via rules
// ✅ Build virtual filesystem paths
// ✅ Store relationship metadata
```

**3. Virtual Filesystem** - Multiple Reality Views
```
Unified access to all entities through different lenses:

/tools/                    # Relationship-aware tool browser
├── by-name/              # Alphabetical tool listing
├── by-transform/         # Grouped by capabilities  
├── spawned-by/           # Global spawning relationships
├── ancestry/             # Parent-child chains
└── {tool-name}/          # Individual tool context
    ├── definition        # Relation JSON metadata
    ├── executable        # Generated code
    ├── spawned/          # Child entities
    └── parents/          # Parent chain

/similar/                  # Semantic similarity discovery (Step 6)
├── {tool-name}/          # Tools similar to specified tool
└── [150+ tools with automatic relationships]

/commands/                # Traditional executable view (enhanced)
/by-date/{date}/          # Time-based organization (unified)
/memory/                  # AI conversation storage
```

**4. Rules Engine** - Self-Organizing Intelligence
```
Automatic system behaviors:
• Analysis tools → spawn viewer tools automatically
• Document changes → regenerate related artifacts  
• Semantic similarity → create suggestion links
• Parent tools → inherit capabilities to children
```

**5. Materializers** - Reality Manifestation
```
Transform abstract relations into concrete reality:
• Tool Materializer: Relations → Working executables
• Artifact Materializer: Intentions → Documents/media
• Memory Materializer: Conversations → Searchable knowledge
```

### Premise Principles Implementation

**Zero Implementation Complexity**
```bash
# Traditional: 50+ lines of bash, file management, permissions
mkdir -p ~/.local/bin
cat > ~/.local/bin/git-summary << 'EOF'
#!/usr/bin/env python3
# ... 30 lines of implementation ...
EOF  
chmod +x ~/.local/bin/git-summary
export PATH="$PATH:~/.local/bin"

# Premise: 1 declaration
port42 declare tool git-summary --transforms git-log,analysis
# Everything else handled automatically
```

**Self-Maintaining Reality**
- Tools appear automatically in `/tools/`, `/commands/`, `/by-date/`
- Auto-spawning creates viewer tools for analysis tools
- Virtual filesystem stays consistent across all views
- Relationship graph maintains parent-child connections

**Consciousness-Aligned Computing**  
- Natural language: `port42 declare tool NAME --transforms X,Y`
- Multiple perspectives: Same tool visible through different organizational schemes
- Relationship intelligence: Spawning chains, capability grouping, temporal organization

## 🔗 Universal Prompt & Reference System - Context-Aware Tool Creation

Port 42's reality compiler includes a **Universal Prompt & Reference System** that allows you to create tools with rich contextual knowledge from multiple sources and custom instructions. Instead of creating tools from scratch, provide specific guidance and reference existing knowledge, files, and web content.

### Custom AI Generation with Prompts

**Direct Instructions (`--prompt "instructions"`)**
```bash
# Guide the AI with specific instructions for tool generation
port42 declare tool log-analyzer --transforms analyze,logs \
  --prompt "Create a tool that analyzes web server logs and highlights errors, security threats, and performance issues"

# Combine prompts with references for context-aware generation
port42 declare tool config-validator --transforms validate,config \
  --ref file:./app.json \
  --prompt "Build a validator that checks the referenced config for security vulnerabilities and best practices"
```

**Advanced AI Artifact Generation**
```bash
# Create context-aware documentation
port42 declare artifact api-docs --artifact-type documentation \
  --ref p42:/commands/api-server \
  --ref url:https://api.example.com \
  --prompt "Generate comprehensive API documentation with examples, error codes, and authentication details"

# Build specialized configurations
port42 declare artifact deployment-config --artifact-type config --file-type .yaml \
  --ref file:./kubernetes-base.yaml \
  --prompt "Create a production-ready Kubernetes deployment with auto-scaling, health checks, and security policies"
```

### Reference Types

**Local Files (`file:path`)**
```bash
# Reference project configuration, documentation, or code
port42 declare tool config-validator --transforms validate,config \
  --ref file:./app.json \
  --ref file:./README.md
```

**Port 42 VFS (`p42:path`)**
```bash
# Reference existing tools and crystallized knowledge
port42 declare tool enhanced-processor --transforms process,enhance \
  --ref p42:/commands/base-processor \
  --ref p42:/commands/utility-tool
```

**Web Content (`url:https://...`)**
```bash
# Reference API documentation, specifications, examples
port42 declare tool api-client --transforms http,client \
  --ref url:https://api.example.com/docs \
  --ref url:https://github.com/example/api-examples
```

**Knowledge Search (`search:query`)**
```bash
# Reference crystallized knowledge from previous sessions
port42 declare tool error-analyzer --transforms analyze,debug \
  --ref search:"error handling patterns" \
  --ref search:"debugging techniques"
```

**Tool Definitions (`tool:name`)**
```bash
# Reference existing tool capabilities and commands
port42 declare tool super-analyzer --transforms analyze,extend \
  --ref tool:log-parser \
  --ref tool:data-processor
```

**Memory Sessions (`p42:/memory/session-id`)**
```bash
# Reference previous conversations and decisions
port42 declare tool project-manager --transforms manage,track \
  --ref p42:/memory/cli-1234 \
  --ref file:./project-spec.md
```

### Multi-Reference Intelligence with Custom Prompts

Combine multiple reference types with custom instructions for sophisticated context-aware generation:

```bash
# The ultimate context-aware tool with AI guidance
port42 declare tool intelligent-processor --transforms process,analyze,output \
  --ref file:./data-schema.json \
  --ref p42:/commands/base-processor \
  --ref url:https://standards.org/spec \
  --ref search:"processing patterns" \
  --prompt "Create a sophisticated data processor that validates against the schema, follows industry standards, and includes error handling with detailed logging"

# AI-driven conversation with context
port42 possess @ai-engineer \
  --ref file:./project-requirements.md \
  --ref p42:/commands/existing-codebase \
  "Help me design a microservice architecture for this project"
```

## 🔍 Semantic Tool Discovery - Find Tools by What They Do

Port 42's reality compiler includes **semantic similarity detection** that automatically discovers relationships between your tools. Instead of remembering what tools exist, explore by capability and discover unexpected connections.

### Automatic Similarity Detection

Every tool you create is automatically analyzed for similarity to existing tools:

```bash
# Create analysis tools - they automatically find each other
$ port42 declare tool log-analyzer --transforms logs,analysis
$ port42 declare tool quick-analyzer --transforms data,analysis
$ port42 declare tool semantic-analyzer --transforms text,analysis

# Explore automatic relationships via /similar/ virtual path
$ port42 ls /similar/log-analyzer/
quick-analyzer
semantic-analyzer
test-analyzer
phase-b-analyzer
code-analyzer
```

### Similarity Virtual Filesystem

Navigate tool relationships like a filesystem - no complex queries needed:

```bash
# See all tools that have similar tools available
$ port42 ls /similar/
advanced-parser/
analyzer/
basic-parser/
data-processor/
log-analyzer/
semantic-analyzer/
# ... 150+ tools with automatic relationships

# Explore specific tool similarities
$ port42 ls /similar/basic-parser/
enhanced-parser
test-parser
doc-processor       # Cross-category: parsers can find processors

# Tools automatically group by capability
$ port42 ls /similar/semantic-analyzer/
log-analyzer        # Both do analysis
quick-analyzer      # Both do analysis  
test-analyzer       # Both do analysis
phase-b-analyzer    # Both do analysis
```

### Semantic Transform Intelligence

The system understands **semantic relationships** between transforms:

```bash
# Tools with 'analyze' find tools with 'analysis' (semantic boost)
$ port42 ls /similar/analyzer/
semantic-analyzer   # analyze ↔ analysis relationship detected
log-analyzer        # analyze ↔ analysis relationship detected

# Cross-category detection based on shared capabilities
$ port42 ls /similar/data-processor/
enhanced-parser     # Both process/transform data
doc-processor       # Both process documents
phase-b-processor   # Both processors
```

### Mathematical Similarity Scoring

Port 42 uses **Jaccard similarity coefficient** with semantic enhancement:

- **Base similarity**: Shared transforms / Total unique transforms
- **Semantic boost**: Related words (analyze ↔ analysis, parse ↔ parser)
- **Threshold filtering**: 20% for browsing, 50% for relationship storage
- **Bidirectional relationships**: If A is similar to B, then B is similar to A

```bash
# High similarity tools find many matches (>40 similar tools)
$ port42 ls /similar/log-analyzer/ | wc -l
45

# Specialized tools find fewer, more relevant matches
$ port42 ls /similar/basic-parser/ | wc -l  
12

# System maintains mathematical consistency
$ port42 ls /similar/log-analyzer/ | grep semantic-analyzer
semantic-analyzer
$ port42 ls /similar/semantic-analyzer/ | grep log-analyzer  
log-analyzer
```

### Performance at Scale

The similarity system handles large tool collections efficiently:

```bash
# Instant response even with 150+ tools in system
$ time port42 ls /similar/semantic-analyzer/
# Response: ~18ms (< 100ms typical)

# Root directory shows comprehensive coverage
$ port42 ls /similar/ | wc -l
154
```

### Discovery Use Cases

**Find tools by capability without remembering names:**
```bash
# "I need something that analyzes data"
$ port42 ls /similar/analyzer/

# "I need something that parses files"  
$ port42 ls /similar/basic-parser/

# "I need something like this existing tool"
$ port42 ls /similar/my-existing-tool/
```

**Discover tool ecosystems:**
```bash
# Find all tools related to security
$ port42 ls /similar/security-test/

# Find all tools related to file processing
$ port42 ls /similar/file-processor/

# Explore generated viewer tools
$ port42 ls /similar/view-analyzer/
```

**Quality assurance via relationships:**
```bash
# Verify similar tools cluster correctly
$ port42 ls /similar/log-analyzer/ | grep analyzer | wc -l    # Many analyzers
$ port42 ls /similar/log-analyzer/ | grep parser | wc -l     # Few parsers

# Check bidirectional relationships
$ port42 ls /similar/A/ | grep B
$ port42 ls /similar/B/ | grep A    # Should both work
```

### Integration with Reality Compiler

Similarity detection is **seamlessly integrated** with all Port 42 features:

- **Auto-spawning rules**: Similar tools influence spawning decisions
- **Virtual filesystem**: Relationships visible through `/tools/` and `/similar/`
- **Search enhancement**: Semantic scoring improves search relevance
- **Background processing**: Similarity analysis doesn't block tool creation

**The similarity system transforms Port 42 from a tool collection into an intelligent capability discovery engine** - you explore *what tools can do* rather than *what files exist*.

### Creating Commands and Artifacts

Port 42 can create two types of outputs from your AI conversations:

#### Commands - Executable Tools
```bash
# Create a command that becomes part of your system
$ port42 possess @ai-engineer
> Create a command that converts CSV files to beautiful markdown tables
✨ Crystallizing thought into reality...
[Created: csv-to-markdown]

# Now use it anywhere
$ csv-to-markdown data.csv > report.md
```

#### Artifacts - Documents, Configs, Scripts
```bash
# Generate complex documents or configurations
$ port42 possess @ai-muse
> Create a comprehensive README for my Python project with badges and examples
📄 Generating artifact...
[Created: /memory/cli-1234/artifacts/README.md]

# Artifacts are version-controlled and linked to conversations
$ port42 cat /memory/cli-1234/artifacts/README.md
```

### Navigating the Virtual Filesystem

Port 42 presents a unified view of all your content:

```bash
# List all content types
$ port42 ls /
/memory     - Conversation threads and artifacts
/commands   - Your crystallized commands
/by-date    - Temporal organization
/by-agent   - Organized by AI consciousness

# Explore specific areas
$ port42 ls /memory/cli-1234
messages.json          - Conversation history
artifacts/             - Generated documents
  README.md
  docker-compose.yml
  
# Read any content
$ port42 cat /commands/csv-to-markdown
$ port42 info /memory/cli-1234  # See metadata

# Search across everything
$ port42 search "docker" --type artifact
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
- **Rust CLI** (`cli/`): Fast, user-friendly command-line interface with protocol abstraction
- **Generated Commands** (`~/.port42/commands/`): Your personalized command library

The CLI uses a **protocol abstraction pattern** that provides:
- Type-safe request/response handling via `RequestBuilder` and `ResponseParser` traits
- Consistent output formatting through the `Displayable` trait
- Built-in JSON support for all commands via the global `--json` flag
- Centralized error handling and help text in Reality Compiler language

See [protocol pattern documentation](docs/architecture/protocol-pattern.md) for details.

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
- `protocol/`: Protocol abstraction layer (request/response types)
- `commands/`: Command handlers using protocol pattern
- `client.rs`: TCP client accepting `DaemonRequest` types
- `display/`: Output formatting components
- `shell.rs`: Interactive shell mode
- `help_text.rs`: Reality Compiler language constants

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

Using the protocol pattern, new commands follow a consistent structure:

```rust
// 1. Define protocol types (cli/src/protocol/yourcommand.rs)
pub struct YourCommandRequest { ... }
impl RequestBuilder for YourCommandRequest { ... }
impl ResponseParser for YourCommandResponse { ... }
impl Displayable for YourCommandResponse { ... }

// 2. Create handler (cli/src/commands/yourcommand.rs)
pub fn handle_yourcommand(client: &mut DaemonClient, args...) -> Result<()> {
    let request = YourCommandRequest { ... }.build_request(id)?;
    let response = client.request(request)?;
    YourCommandResponse::parse_response(&response.data)?.display(format)?;
    Ok(())
}

// 3. Add to CLI (main.rs)
Commands::YourCommand { args } => {
    commands::yourcommand::handle_yourcommand(&mut client, args)?;
}
```

See the [developer guide](docs/developer/adding-commands.md) for detailed instructions.

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
├── commands/                # Your generated executable commands
├── memory/                  # Conversation history
│   ├── sessions/           # Session files organized by date
│   │   └── 2025/01/04/    # Daily directories
│   │       └── cli-*.json # Individual session files
│   └── index.json         # Memory index and statistics
├── artifacts/              # Generated documents and files
│   └── [session-id]/      # Artifacts grouped by conversation
├── metadata/              # Object metadata (git-like)
│   └── [object-id].json  # Metadata for each object
├── objects/               # Content-addressed storage
│   └── [sha]/[object]    # Git-like object storage
├── agents.json           # AI agent configurations
├── daemon.log           # Daemon runtime logs
└── activate.sh         # Shell activation helper

Virtual Filesystem View:
/                          # Root of reality
├── memory/               # All conversation threads
│   └── cli-*/           # Individual sessions with artifacts
├── commands/            # All crystallized commands
├── tools/               # Relationship-aware tool browser
│   ├── by-name/        # Alphabetical listing
│   ├── by-transform/   # Grouped by capabilities
│   ├── spawned-by/     # Global spawning relationships
│   └── ancestry/       # Parent-child chains
├── similar/             # Semantic similarity discovery (Step 6)
│   └── {tool-name}/    # Tools similar to specified tool (150+ with relationships)
├── by-date/            # Temporal organization
│   └── 2025-01-04/    # Daily views
└── by-agent/          # Organized by AI consciousness
    ├── @ai-engineer/  # Technical creations
    ├── @ai-muse/     # Creative works
    ├── @ai-analyst/  # Analysis & insights
    └── @ai-founder/  # Visionary synthesis
```

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