# Port 42 CLI Architecture Overhaul Plan

## Executive Summary

This document outlines a comprehensive architectural overhaul of the Port 42 CLI to address systemic issues of code duplication, lack of abstractions, and architectural debt. The plan establishes proper separation of concerns through a layered architecture with clear boundaries and responsibilities.

## Current State Analysis

### Systemic Problems

#### 1. No Protocol Abstraction
- Request types are magic strings scattered throughout code
- Manual JSON construction in every command
- No type safety or validation
- Protocol changes require updates in multiple places
- No versioning capability

#### 2. Command Implementation Chaos
- Each command reimplements:
  - Client connection logic
  - Request construction
  - Response parsing
  - Error handling
  - Display formatting
- No standard command interface or framework
- Business logic mixed with presentation

#### 3. Display Logic Scattered
- Each command formats its own output
- No consistent styling or formatting rules
- Interactive vs non-interactive display duplicated
- No reusable display components

#### 4. Error Handling Inconsistency
- Each command handles errors differently
- No standard error types
- User-facing error messages vary wildly
- Debug logging implemented differently everywhere

### Specific Examples of Duplication

#### Request Construction Pattern (appears in EVERY command)
```rust
// From possess.rs, memory.rs, list.rs, status.rs, etc.
let response = client.request(Request {
    request_type: "some_type".to_string(),
    id: generate_id(),
    payload: serde_json::json!({ /* manual JSON */ }),
})?;
```

#### Response Handling Pattern (appears in EVERY command)
```rust
// Duplicated with variations in all commands
if response.success {
    if let Some(data) = response.data {
        // Command-specific parsing
        // Command-specific display
    } else {
        println!("No data");
    }
} else {
    println!("Error: {}", response.error.unwrap_or("Unknown".to_string()));
}
```

#### Display Logic Examples
- `memory.rs`: 40+ lines of custom session formatting
- `list.rs`: 20+ lines of command list formatting
- `status.rs`: Custom status display logic
- `possess.rs` + `interactive.rs`: Duplicate artifact/command display

### Root Cause Analysis

The codebase suffers from **organic growth without architectural planning**:

1. **No initial framework** - Commands added ad-hoc without patterns
2. **Copy-paste development** - New commands copied from existing ones
3. **No abstraction layers** - Direct client usage everywhere
4. **Mixed concerns** - Business logic, display, and protocol handling intertwined
5. **No shared components** - Every command is an island

## Proposed Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   CLI Entry Point                        │
│                    (main.rs)                            │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                 Command Registry                         │
│  - Command discovery and routing                        │
│  - Argument parsing                                     │
│  - Help generation                                      │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              Command Implementations                     │
│  - Implement Command trait                              │
│  - Focus on business logic only                         │
│  - Return structured data                              │
├─────────────────────────┴───────────────────────────────┤
│ possess │ list │ memory │ status │ cat │ etc.        │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              Protocol Abstraction Layer                  │
│  - Type-safe request/response definitions               │
│  - Protocol versioning and negotiation                  │
│  - Automatic serialization/deserialization              │
│  - Request builders and response parsers                │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                 Display Framework                        │
│  - Output format selection (plain, JSON, table)         │
│  - Consistent styling and colors                        │
│  - Progress indicators and animations                   │
│  - Interactive vs non-interactive modes                 │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  Error Framework                         │
│  - Error type hierarchy                                 │
│  - User-friendly error messages                        │
│  - Debug information when requested                     │
│  - Consistent error display                            │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                Transport Layer (Client)                  │
│  - Low-level TCP communication                          │
│  - Connection pooling                                   │
│  - Retry logic                                          │
└─────────────────────────────────────────────────────────┘
```

## Implementation Steps

### Step 1: Define Protocol Types

**File: `cli/src/protocol/mod.rs`**

```rust
// Version-aware protocol definition
pub const PROTOCOL_VERSION: &str = "1.0";

// Strongly-typed request types
#[derive(Debug, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum ProtocolRequest {
    // Basic operations
    Status,
    Ping,
    
    // AI Agent operations
    Possess {
        agent: String,
        message: String,
        session_id: Option<String>,
    },
    
    // Command management
    List {
        filter: Option<String>,
    },
    
    // Memory/Session operations
    Memory {
        session_id: Option<String>,
    },
    CreateMemory {
        agent: String,
        initial_message: Option<String>,
    },
    End {
        session_id: String,
    },
    
    // Virtual filesystem operations
    StorePath {
        path: String,
        content: String, // base64 encoded
        metadata: Option<StoreMetadata>,
    },
    UpdatePath {
        path: String,
        content: Option<String>, // base64 encoded
        metadata_updates: Option<MetadataUpdates>,
    },
    DeletePath {
        path: String,
    },
    ListPath {
        path: Option<String>, // defaults to "/"
    },
    ReadPath {
        path: String,
    },
    GetMetadata {
        path: String,
    },
    Search {
        query: String,
        filters: Option<SearchFilters>,
    },
}

// Request metadata types
#[derive(Debug, Serialize, Deserialize)]
pub struct StoreMetadata {
    pub r#type: String,
    pub description: Option<String>,
    pub agent: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct MetadataUpdates {
    pub description: Option<String>,
    pub tags: Option<Vec<String>>,
    pub importance: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SearchFilters {
    pub path: Option<String>,
    pub r#type: Option<String>,
    pub after: Option<String>,
    pub before: Option<String>,
    pub agent: Option<String>,
    pub tags: Option<Vec<String>>,
    pub limit: Option<usize>,
}

// Strongly-typed responses
#[derive(Debug, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum ProtocolResponse {
    Status(StatusData),
    Possess(PossessData),
    List(ListData),
    Memory(MemoryData),
    SessionList(SessionListData),
    StorePath(StorePathData),
    ListPath(ListPathData),
    ReadPath(ReadPathData),
    GetMetadata(MetadataData),
    Search(SearchData),
    Simple(SimpleData),
    Empty,
}

// Response data structures
#[derive(Debug, Serialize, Deserialize)]
pub struct StatusData {
    pub status: String,
    pub port: String,
    pub sessions: usize,
    pub uptime: String,
    pub dolphins: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PossessData {
    pub message: String,
    pub session_id: String,
    pub agent: String,
    pub command_generated: bool,
    pub command_spec: Option<CommandSpec>,
    pub artifact_generated: bool,
    pub artifact_spec: Option<ArtifactSpec>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CommandSpec {
    pub name: String,
    pub description: String,
    pub language: String,
    pub implementation: String,
    pub dependencies: Option<Vec<String>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ArtifactSpec {
    pub name: String,
    pub r#type: String,
    pub description: String,
    pub format: String,
    pub path: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ListData {
    pub commands: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SessionListData {
    pub active_sessions: Vec<SessionInfo>,
    pub recent_sessions: Vec<SessionInfo>,
    pub stats: SessionStats,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SessionInfo {
    pub id: String,
    pub agent: String,
    pub created_at: String,
    pub last_activity: String,
    pub message_count: usize,
    pub state: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct MemoryData {
    pub id: String,
    pub agent: String,
    pub state: String,
    pub created_at: String,
    pub last_activity: String,
    pub messages: Vec<Message>,
    pub command_generated: Option<GeneratedCommand>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Message {
    pub role: String,
    pub content: String,
    pub timestamp: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ListPathData {
    pub path: String,
    pub entries: Vec<PathEntry>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PathEntry {
    pub name: String,
    pub r#type: String,
    pub size: Option<usize>,
    pub created: Option<String>,
    pub executable: Option<bool>,
    pub state: Option<String>, // for memory entries
    pub messages: Option<usize>, // for memory entries
}

// Request/Response wrapper for versioning
#[derive(Debug, Serialize, Deserialize)]
pub struct Request {
    pub version: String,
    pub id: String,
    pub body: ProtocolRequest,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Response {
    pub version: String,
    pub id: String,
    pub body: Result<ProtocolResponse, ErrorData>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ErrorData {
    pub code: String,
    pub message: String,
    pub details: Option<serde_json::Value>,
}
```

### Step 2: Create Protocol Client

**File: `cli/src/protocol/client.rs`**

```rust
pub struct ProtocolClient {
    transport: DaemonClient,
    version: String,
}

impl ProtocolClient {
    /// Type-safe request sending
    pub async fn send<Req, Resp>(&mut self, request: Req) -> Result<Resp>
    where
        Req: Into<ProtocolRequest>,
        Resp: TryFrom<ProtocolResponse>,
    {
        let proto_request = Request {
            version: self.version.clone(),
            id: generate_request_id(),
            body: request.into(),
        };
        
        let response = self.transport.send(proto_request).await?;
        
        // Version compatibility check
        if response.version != self.version {
            return Err(ProtocolError::VersionMismatch);
        }
        
        // Type-safe response extraction
        match response.body {
            Ok(data) => Resp::try_from(data)
                .map_err(|_| ProtocolError::UnexpectedResponse),
            Err(error) => Err(ProtocolError::ServerError(error)),
        }
    }
}
```

### Step 3: Define Command Framework

**File: `cli/src/framework/command.rs`**

```rust
/// Standard interface for all commands
pub trait Command: Send + Sync {
    /// Command name (e.g., "possess", "list")
    fn name(&self) -> &'static str;
    
    /// Short description for help text
    fn description(&self) -> &'static str;
    
    /// Define command-line arguments
    fn args(&self) -> clap::Command;
    
    /// Execute the command with parsed arguments
    fn execute(&self, args: &ArgMatches, context: &mut Context) -> Result<Box<dyn Display>>;
}

/// Execution context provided to commands
pub struct Context {
    pub client: ProtocolClient,
    pub config: Config,
    pub display: DisplayManager,
}

/// Display trait for command results
pub trait Display {
    fn render(&self, format: OutputFormat) -> Result<String>;
}
```

### Step 4: Implement Display Framework

**File: `cli/src/display/mod.rs`**

```rust
pub struct DisplayManager {
    format: OutputFormat,
    style: StyleConfig,
    interactive: bool,
}

#[derive(Clone, Copy)]
pub enum OutputFormat {
    Plain,
    Json,
    Table,
    Interactive,
}

/// Trait for displayable data
pub trait Displayable {
    fn display(&self, manager: &DisplayManager) -> Result<()>;
}

/// Reusable display components
pub mod components {
    pub struct Table { /* ... */ }
    pub struct ProgressBar { /* ... */ }
    pub struct Tree { /* ... */ }
    pub struct AnimatedText { /* ... */ }
}

/// Implementations for common data types
impl Displayable for PossessData {
    fn display(&self, manager: &DisplayManager) -> Result<()> {
        match manager.format {
            OutputFormat::Json => manager.render_json(self),
            OutputFormat::Plain => manager.render_possess_plain(self),
            OutputFormat::Interactive => manager.render_possess_interactive(self),
            _ => Err(DisplayError::UnsupportedFormat),
        }
    }
}
```

### Step 5: Create Error Framework

**File: `cli/src/error.rs`**

```rust
#[derive(Debug, thiserror::Error)]
pub enum Port42Error {
    #[error("Connection failed: {0}")]
    Connection(#[from] ConnectionError),
    
    #[error("Protocol error: {0}")]
    Protocol(#[from] ProtocolError),
    
    #[error("Command failed: {0}")]
    Command(#[from] CommandError),
    
    #[error("Display error: {0}")]
    Display(#[from] DisplayError),
}

/// User-friendly error display
impl Port42Error {
    pub fn display_for_user(&self, verbose: bool) -> String {
        match self {
            Self::Connection(e) => format!(
                "{}\n\nTry:\n  - Check if daemon is running\n  - Run: port42 daemon start",
                e
            ),
            Self::Protocol(e) if verbose => format!("{:?}", e),
            Self::Protocol(e) => e.to_string(),
            // ... other cases
        }
    }
}
```

### Step 6: Refactor Existing Commands

**Example: Refactored possess command**

```rust
pub struct PossessCommand;

impl Command for PossessCommand {
    fn name(&self) -> &'static str { "possess" }
    
    fn description(&self) -> &'static str {
        "Channel AI consciousness through an agent"
    }
    
    fn args(&self) -> clap::Command {
        clap::Command::new(self.name())
            .about(self.description())
            .arg(arg!(<AGENT> "Agent to possess"))
            .arg(arg!(<MESSAGE> "Initial message"))
            .arg(arg!(-s --session <ID> "Session ID"))
            .arg(arg!(-i --interactive "Interactive mode"))
    }
    
    fn execute(&self, args: &ArgMatches, ctx: &mut Context) -> Result<Box<dyn Display>> {
        let agent = args.get_one::<String>("AGENT").unwrap();
        let message = args.get_one::<String>("MESSAGE").unwrap();
        let session_id = args.get_one::<String>("session");
        let interactive = args.get_flag("interactive");
        
        if interactive {
            Ok(Box::new(InteractiveSession::start(ctx, agent, session_id)?))
        } else {
            let request = ProtocolRequest::Possess {
                agent: agent.clone(),
                message: message.clone(),
                session_id: session_id.cloned(),
            };
            
            let response: PossessData = ctx.client.send(request).await?;
            Ok(Box::new(response))
        }
    }
}
```

### Step 7: Create Command Registry

**File: `cli/src/framework/registry.rs`**

```rust
pub struct CommandRegistry {
    commands: HashMap<String, Box<dyn Command>>,
}

impl CommandRegistry {
    pub fn new() -> Self {
        let mut registry = Self {
            commands: HashMap::new(),
        };
        
        // Register all commands
        registry.register(Box::new(PossessCommand));
        registry.register(Box::new(ListCommand));
        registry.register(Box::new(MemoryCommand));
        registry.register(Box::new(StatusCommand));
        // ... etc
        
        registry
    }
    
    pub fn execute(&self, args: &ArgMatches) -> Result<()> {
        let (cmd_name, cmd_args) = args.subcommand().unwrap();
        
        let command = self.commands.get(cmd_name)
            .ok_or(Error::UnknownCommand)?;
            
        let mut context = Context::new()?;
        let result = command.execute(cmd_args, &mut context)?;
        
        context.display.render(result)?;
        Ok(())
    }
}
```

### Step 8: Update Main Entry Point

**File: `cli/src/main.rs`**

```rust
fn main() -> Result<()> {
    // Initialize
    let config = Config::load()?;
    env_logger::init();
    
    // Build CLI
    let app = build_cli();
    let matches = app.get_matches();
    
    // Create registry and execute
    let registry = CommandRegistry::new();
    
    match registry.execute(&matches) {
        Ok(()) => Ok(()),
        Err(e) => {
            let verbose = matches.get_flag("verbose");
            eprintln!("{}", e.display_for_user(verbose));
            std::process::exit(1);
        }
    }
}
```

### Step 9: Protocol Compatibility Layer

**File: `cli/src/protocol/compat.rs`**

```rust
/// Maintains compatibility with current daemon protocol
pub mod v0 {
    pub fn adapt_request(req: &ProtocolRequest) -> LegacyRequest {
        match req {
            ProtocolRequest::Possess { agent, message, session_id } => {
                LegacyRequest {
                    request_type: "possess".to_string(),
                    id: session_id.clone().unwrap_or_else(generate_id),
                    payload: json!({
                        "agent": agent,
                        "message": message,
                    }),
                }
            }
            // ... other conversions
        }
    }
    
    pub fn adapt_response(resp: LegacyResponse) -> Result<ProtocolResponse> {
        // Convert old format to new
    }
}
```

### Step 10: Testing Framework

**File: `cli/src/framework/testing.rs`**

```rust
/// Test utilities for commands
pub struct CommandTester {
    client: MockProtocolClient,
}

impl CommandTester {
    pub fn test_command<C: Command>(cmd: C, args: &[&str]) -> TestResult {
        let matches = cmd.args().get_matches_from(args);
        let mut context = Context::new_mock();
        let result = cmd.execute(&matches, &mut context)?;
        
        TestResult {
            output: result.render(OutputFormat::Plain)?,
            requests: context.client.get_requests(),
        }
    }
}
```

## Migration Strategy

### Step 1: Parallel Implementation
- Build new framework alongside existing code
- Implement protocol layer with compatibility adapter
- No breaking changes to daemon

### Step 2: Command Migration
- Migrate one simple command (e.g., `status`) as proof of concept
- Verify behavior parity
- Migrate remaining commands one by one

### Step 3: Interactive Mode Integration
- Refactor interactive possess to use new framework
- Eliminate duplicate display logic
- Unify session management

### Step 4: Cleanup
- Remove old request/response types
- Delete duplicate display code
- Update all imports and dependencies

### Step 5: Daemon Protocol Update
- Once all commands migrated, update daemon to use new protocol
- Maintain compatibility layer for older CLI versions
- Eventually deprecate legacy protocol

## Benefits

### Code Quality
- **DRY Principle**: ~60% reduction in code duplication
- **Single Responsibility**: Clear separation of concerns
- **Type Safety**: Compile-time protocol validation
- **Testability**: Mockable at every layer

### Developer Experience
- **Adding new commands**: Implement trait, register, done
- **Protocol changes**: Update once in protocol layer
- **Display changes**: Modify display framework, affects all commands
- **Error handling**: Consistent across entire CLI

### User Experience
- **Consistent output**: All commands follow same patterns
- **Better errors**: Helpful, actionable error messages
- **Multiple formats**: JSON, table, plain text output
- **Predictable behavior**: Same options work everywhere

### Maintainability
- **Clear architecture**: Easy to understand and modify
- **Isolated changes**: Modifications don't cascade
- **Version compatibility**: Protocol versioning built-in
- **Comprehensive tests**: Every layer independently testable

## Risks and Mitigations

### Risk: Large Scope
**Mitigation**: Incremental migration, maintain compatibility throughout

### Risk: Breaking Changes
**Mitigation**: Compatibility layer, extensive testing, gradual rollout

### Risk: Performance Impact
**Mitigation**: Benchmark critical paths, optimize hot spots

### Risk: Team Adoption
**Mitigation**: Clear documentation, migration guide, pair programming

## Success Metrics

1. **Code Reduction**: Target 60% less code overall
2. **Test Coverage**: 90%+ on framework components
3. **Command Consistency**: All commands use framework
4. **Zero Regressions**: Full behavior compatibility
5. **Developer Velocity**: 50% faster to add new commands

## Conclusion

This architectural overhaul addresses fundamental issues in the Port 42 CLI codebase. By establishing proper abstractions and frameworks, we create a sustainable foundation for future development. The investment in architecture will pay dividends in maintainability, consistency, and developer productivity.