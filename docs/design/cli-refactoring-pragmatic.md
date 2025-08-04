# Port 42 CLI Refactoring: Pragmatic Approach

## Executive Summary

This document outlines a focused, incremental refactoring plan for the Port 42 CLI that prioritizes quick wins and clean code over comprehensive architectural changes. The plan focuses on eliminating code duplication starting with the possess command as a proof of concept.

## Constraints & Assumptions

### What We're Building
- Single-user, local-only CLI tool
- Clean code architecture without over-engineering
- CLI-only changes (daemon remains unchanged)
- No backward compatibility requirements

### What We're NOT Building
- No async/streaming operations
- No authentication or security layers
- No plugin system
- No protocol versioning
- No daemon modifications

## Current Problems (Focused)

### Possess Command Duplication
- **Response parsing**: Duplicated between `possess.rs` (lines 192-230) and `interactive.rs` (lines 164-221)
- **Display logic**: Different implementations for same data
- **Request construction**: Repeated pattern with manual JSON
- **Error handling**: Inconsistent approaches

### Root Causes
1. Interactive mode added after non-interactive without shared abstractions
2. No separation between data parsing and presentation
3. Direct client usage without protocol layer

## Plan Tracker

| Step | Name | Status |
|------|------|--------|
| 1 | Create Integration Tests for Possess | ✅ Complete |
| 2 | Create Protocol Types and Traits | ✅ Complete |
| 3 | Create Display Trait with Help Text Integration | ✅ Complete |
| 4 | Create Common Libraries | ✅ Complete |
| 5 | Create Shared Session Handler | ✅ Complete |
| 6 | Refactor Possess Command | ✅ Complete |
| 7 | Refactor Interactive Mode | ✅ Complete |
| 8 | Create General Display Framework | ✅ Complete |
| 9 | Integrate Possess with Display Framework | ✅ Complete |
| 10 | Apply Pattern to Status Command | ✅ Complete |
| 11 | Apply Pattern to Reality Command | ✅ Complete |
| 12 | Apply Pattern to Daemon and Init Commands | ⏸️ Paused (no daemon interaction) |
| 13 | Apply Pattern to Memory Command | ✅ Complete |
| 14 | Apply Pattern to Cat and Info Commands | ✅ Complete|
| 15 | Apply Pattern to Ls Command | ✅ Complete |
| 16 | Apply Pattern to Search Command | ✅ Complete |
| 17 | Update Main Entry Point | ✅ Complete |
| 18 | Remove Old Duplicate Code | ✅ Complete |
| 19 | Update Documentation | ✅ Complete |

## Implementation Plan

### Step 1: Create Integration Tests for Possess

Start with tests to ensure we don't break existing functionality during refactoring.

**File: `cli/tests/possess_integration.rs`**

```rust
use port42_cli::client::DaemonClient;
use port42_cli::commands::possess;
use std::process::Command;
use std::time::Duration;
use std::thread;

#[test]
fn test_possess_non_interactive() {
    // Start test daemon
    let daemon = start_test_daemon();
    thread::sleep(Duration::from_millis(500)); // Let daemon start
    
    let mut client = DaemonClient::new(daemon.port());
    
    // Test basic possess
    let result = possess::handle_possess(
        &mut client,
        "@ai-engineer".to_string(),
        "test message".to_string(),
        None,
        false,
    );
    
    assert!(result.is_ok());
    // Could check for specific response patterns
}

#[test]
fn test_possess_with_session_id() {
    let daemon = start_test_daemon();
    thread::sleep(Duration::from_millis(500));
    
    let mut client = DaemonClient::new(daemon.port());
    
    // First message creates session
    let result1 = possess::handle_possess(
        &mut client,
        "@ai-muse".to_string(),
        "first message".to_string(),
        Some("test-session-123".to_string()),
        false,
    );
    assert!(result1.is_ok());
    
    // Second message continues session
    let result2 = possess::handle_possess(
        &mut client,
        "@ai-muse".to_string(),
        "second message".to_string(),
        Some("test-session-123".to_string()),
        false,
    );
    assert!(result2.is_ok());
}

#[test]
fn test_possess_invalid_agent() {
    let daemon = start_test_daemon();
    thread::sleep(Duration::from_millis(500));
    
    let mut client = DaemonClient::new(daemon.port());
    
    let result = possess::handle_possess(
        &mut client,
        "@invalid-agent".to_string(),
        "test message".to_string(),
        None,
        false,
    );
    
    assert!(result.is_err());
    // Should contain appropriate error message
}

// Helper to start daemon for tests
fn start_test_daemon() -> TestDaemon {
    let port = find_free_port();
    let child = Command::new("./bin/port42d")
        .env("PORT42_PORT", port.to_string())
        .env("PORT42_TEST_MODE", "1")
        .spawn()
        .expect("Failed to start test daemon");
    
    TestDaemon { child, port }
}

struct TestDaemon {
    child: std::process::Child,
    port: u16,
}

impl TestDaemon {
    fn port(&self) -> u16 {
        self.port
    }
}

impl Drop for TestDaemon {
    fn drop(&mut self) {
        let _ = self.child.kill();
    }
}
```

### Step 2: Create Protocol Types and Traits

**File: `cli/src/protocol/mod.rs`**

```rust
use serde::{Deserialize, Serialize};
use anyhow::Result;

// Base request that all commands use
#[derive(Debug, Serialize)]
pub struct DaemonRequest {
    #[serde(rename = "type")]
    pub request_type: String,
    pub id: String,
    pub payload: serde_json::Value,
}

// Base response from daemon
#[derive(Debug, Deserialize)]
pub struct DaemonResponse {
    pub id: String,
    pub success: bool,
    pub data: Option<serde_json::Value>,
    pub error: Option<String>,
}

// Common trait for request builders
pub trait RequestBuilder {
    fn build_request(&self, id: String) -> Result<DaemonRequest>;
}

// Common trait for response parsers
pub trait ResponseParser {
    type Output;
    fn parse_response(data: &serde_json::Value) -> Result<Self::Output>;
}
```

**File: `cli/src/protocol/possess.rs`**

```rust
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize)]
pub struct PossessRequest {
    pub agent: String,
    pub message: String,
}

impl RequestBuilder for PossessRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "possess".to_string(),
            id,
            payload: serde_json::to_value(self)?,
        })
    }
}

#[derive(Debug, Deserialize)]
pub struct PossessResponse {
    pub message: String,
    pub session_id: String,
    pub agent: String,
    pub command_generated: Option<bool>,
    pub command_spec: Option<CommandSpec>,
    pub artifact_generated: Option<bool>,
    pub artifact_spec: Option<ArtifactSpec>,
}

#[derive(Debug, Deserialize)]
pub struct CommandSpec {
    pub name: String,
    pub description: String,
    pub language: String,
}

#[derive(Debug, Deserialize)]
pub struct ArtifactSpec {
    pub name: String,
    #[serde(rename = "type")]
    pub artifact_type: String,
    pub path: String,
    pub description: String,
    pub format: String,
}

impl ResponseParser for PossessResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        // Single place for all response parsing logic
        // Extract message, check flags, build specs
        serde_json::from_value(data.clone())
            .map_err(|e| anyhow!("Failed to parse possess response: {}", e))
    }
}
```

### Step 2: Create Display Trait with Help Text Integration

**File: `cli/src/possess/display.rs`**

```rust
use crate::help_text;

pub trait PossessDisplay {
    fn show_ai_message(&self, agent: &str, message: &str);
    fn show_command_created(&self, spec: &CommandSpec);
    fn show_artifact_created(&self, spec: &ArtifactSpec);
}

pub struct SimpleDisplay;

impl PossessDisplay for SimpleDisplay {
    fn show_ai_message(&self, agent: &str, message: &str) {
        println!("\n{}", agent.bright_blue());
        println!("{}", message);
        println!();
    }
    
    fn show_command_created(&self, spec: &CommandSpec) {
        println!("{}", help_text::format_command_born(&spec.name).bright_green().bold());
        println!("{}", "Add to PATH to use:".yellow());
        println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
        println!();
    }
    
    fn show_artifact_created(&self, spec: &ArtifactSpec) {
        println!("{}", format!("✨ Artifact created: {} ({})", spec.name, spec.artifact_type).bright_cyan().bold());
        println!("{}", "View with:".yellow());
        println!("  {}", format!("port42 cat {}", spec.path).bright_white());
        println!();
    }
}

pub struct AnimatedDisplay {
    depth: u32,
}

impl PossessDisplay for AnimatedDisplay {
    // Animated implementations with delays and progress bars
}
```

### Step 3: Create Common Libraries

**File: `cli/src/common/mod.rs`**

```rust
pub mod errors;
pub mod utils;

use crate::help_text;
use std::time::{SystemTime, UNIX_EPOCH};

/// Generate unique request ID
pub fn generate_id() -> String {
    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis();
    format!("cli-{}", timestamp)
}
```

**File: `cli/src/common/errors.rs`**

```rust
use crate::help_text;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum Port42Error {
    #[error("Connection failed: {0}")]
    Connection(String),
    
    #[error("Daemon error: {0}")]
    Daemon(String),
    
    #[error("Parse error: {0}")]
    Parse(String),
}

impl Port42Error {
    /// Get user-friendly error message using help_text constants
    pub fn user_message(&self) -> String {
        match self {
            Self::Connection(_) => help_text::format_daemon_connection_error(42),
            Self::Daemon(msg) => help_text::format_error_with_help(msg, "possess"),
            _ => self.to_string(),
        }
    }
}
```

### Step 4: Create Shared Session Handler

**File: `cli/src/possess/session.rs`**

```rust
pub struct SessionHandler {
    client: DaemonClient,
    display: Box<dyn PossessDisplay>,
}

impl SessionHandler {
    pub fn new(client: DaemonClient, interactive: bool) -> Self {
        let display: Box<dyn PossessDisplay> = if interactive {
            Box::new(AnimatedDisplay::new())
        } else {
            Box::new(SimpleDisplay)
        };
        
        Self { client, display }
    }
    
    pub fn send_message(&mut self, session_id: &str, agent: &str, message: &str) -> Result<PossessResponse> {
        use crate::protocol::{RequestBuilder, ResponseParser};
        use crate::common::generate_id;
        
        // Build request using protocol traits
        let possess_req = PossessRequest {
            agent: agent.to_string(),
            message: message.to_string(),
        };
        let request = possess_req.build_request(session_id.to_string())?;
        
        // Send to daemon
        let response = self.client.request(request)?;
        
        if !response.success {
            return Err(Port42Error::Daemon(response.error.unwrap_or_default()).into());
        }
        
        // Parse response using protocol trait
        let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
        let possess_response = PossessResponse::parse_response(&data)?;
        
        // Display results
        self.display.show_ai_message(agent, &possess_response.message);
        
        if let Some(ref spec) = possess_response.command_spec {
            self.display.show_command_created(spec);
        }
        
        if let Some(ref spec) = possess_response.artifact_spec {
            self.display.show_artifact_created(spec);
        }
        
        Ok(possess_response)
    }
}
```

### Step 5: Refactor Possess Command

**File: `cli/src/commands/possess.rs`**

```rust
use crate::possess::{SessionHandler, session::determine_session_id};
use crate::common::Port42Error;
use crate::help_text;

pub fn handle_possess(
    client: &mut DaemonClient,
    agent: String,
    message: String,
    session_id: Option<String>,
    interactive: bool,
) -> Result<()> {
    // Determine session
    let (session_id, is_new) = determine_session_id(session_id);
    
    if interactive {
        // Interactive mode still handles its own loop
        let session = InteractiveSession::new(client, agent, session_id)?;
        session.run()
    } else {
        // Non-interactive uses shared handler
        let mut handler = SessionHandler::new(client.clone(), false);
        handler.send_message(&session_id, &agent, &message)?;
        Ok(())
    }
}
```

### Step 6: Refactor Interactive Mode

**File: `cli/src/interactive.rs`**

```rust
use crate::possess::{SessionHandler, AnimatedDisplay};

pub struct InteractiveSession {
    handler: SessionHandler,
    agent: String,
    session_id: String,
    depth: u32,
}

impl InteractiveSession {
    pub fn run(&mut self) -> Result<()> {
        self.show_welcome()?;
        
        loop {
            // Get input
            let input = self.read_input()?;
            if input == "/surface" {
                break;
            }
            
            // Use shared handler for sending messages
            let response = self.handler.send_message(&self.session_id, &self.agent, &input)?;
            
            // Track any generated commands/artifacts
            if let Some(spec) = response.command_spec {
                self.commands_generated.push(spec.name);
            }
            if let Some(spec) = response.artifact_spec {
                self.artifacts_generated.push((spec.name, spec.artifact_type, spec.path));
            }
        }
        
        self.show_exit_summary()?;
        Ok(())
    }
}
```

**Integration Test for Interactive Mode**:

```rust
#[test]
fn test_possess_interactive_mode() {
    // This is harder to test due to terminal interaction
    // Could use a pty or mock the terminal input/output
    // For now, ensure the interactive flag is handled correctly
    
    let daemon = start_test_daemon();
    thread::sleep(Duration::from_millis(500));
    
    // Test that interactive mode is properly detected
    // Real interactive testing would require terminal emulation
}
```

### Step 7: Run Possess Integration Tests

Before moving to other commands, ensure all possess tests pass:

```bash
cargo test --test possess_integration
```

This validates that the refactoring maintains backward compatibility.

**File: `cli/src/interactive.rs`**

```rust
use crate::possess::{SessionHandler, AnimatedDisplay};

pub struct InteractiveSession {
    handler: SessionHandler,
    agent: String,
    session_id: String,
    depth: u32,
}

impl InteractiveSession {
    pub fn run(&mut self) -> Result<()> {
        self.show_welcome()?;
        
        loop {
            // Get input
            let input = self.read_input()?;
            if input == "/surface" {
                break;
            }
            
            // Use shared handler for sending messages
            let response = self.handler.send_message(&self.session_id, &self.agent, &input)?;
            
            // Track any generated commands/artifacts
            if let Some(spec) = response.command_spec {
                self.commands_generated.push(spec.name);
            }
            if let Some(spec) = response.artifact_spec {
                self.artifacts_generated.push((spec.name, spec.artifact_type, spec.path));
            }
        }
        
        self.show_exit_summary()?;
        Ok(())
    }
}
```

### Step 8: Create General Display Framework

**File: `cli/src/display/mod.rs`**

```rust
use colored::*;
use std::fmt::Display;

pub enum OutputFormat {
    Plain,
    Json,
    Table,
}

pub trait Displayable {
    fn display(&self, format: OutputFormat) -> Result<()>;
}

// Reusable display components
pub mod components {
    use super::*;
    use prettytable::{Table, Row, Cell};
    
    pub struct TableBuilder {
        table: Table,
    }
    
    impl TableBuilder {
        pub fn new() -> Self {
            Self { table: Table::new() }
        }
        
        pub fn add_header(&mut self, headers: Vec<&str>) -> &mut Self {
            let cells: Vec<Cell> = headers.iter()
                .map(|h| Cell::new(h).style_spec("Fb"))
                .collect();
            self.table.add_row(Row::new(cells));
            self
        }
        
        pub fn add_row(&mut self, values: Vec<String>) -> &mut Self {
            let cells: Vec<Cell> = values.iter()
                .map(|v| Cell::new(v))
                .collect();
            self.table.add_row(Row::new(cells));
            self
        }
        
        pub fn print(&self) {
            self.table.printstd();
        }
    }
    
    pub fn format_size(size: usize) -> String {
        const UNITS: &[&str] = &["B", "K", "M", "G"];
        let mut size = size as f64;
        let mut unit_index = 0;
        
        while size >= 1024.0 && unit_index < UNITS.len() - 1 {
            size /= 1024.0;
            unit_index += 1;
        }
        
        if unit_index == 0 {
            format!("{:>4}B", size as usize)
        } else {
            format!("{:>4.1}{}", size, UNITS[unit_index])
        }
    }
    
    pub fn format_time_ago(timestamp: &str) -> Result<String> {
        // Parse timestamp and return human-readable time ago
        // "2 hours ago", "3 days ago", etc.
    }
}
```

**File: `cli/src/display/impls.rs`**

```rust
use super::*;
use crate::protocol::*;

impl Displayable for StatusResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Plain | OutputFormat::Table => {
                println!("{}", help_text::MSG_CONNECTION_INFO.bright_blue().bold());
                println!("{}", "═══════════════════════".bright_blue());
                println!("{}", help_text::format_port_info(&self.port));
                println!("{}", help_text::format_uptime_info(&self.uptime));
                println!("{}", help_text::format_sessions_info(&self.sessions.to_string()));
                println!("\n{}", self.dolphins.bright_cyan());
            }
        }
        Ok(())
    }
}

impl Displayable for ListResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                let mut table = components::TableBuilder::new();
                table.add_header(vec!["Command", "Type"]);
                
                for cmd in &self.commands {
                    let extension = std::path::Path::new(cmd)
                        .extension()
                        .and_then(|e| e.to_str())
                        .unwrap_or("unknown");
                    table.add_row(vec![cmd.clone(), extension.to_string()]);
                }
                
                table.print();
            }
            OutputFormat::Plain => {
                if self.commands.is_empty() {
                    println!("{}", "No commands found in ~/.port42/commands".dimmed());
                } else {
                    println!("{}", format!("Found {} commands:", self.commands.len()).bright_white());
                    for cmd in &self.commands {
                        println!("  {}", cmd.bright_cyan());
                    }
                }
            }
        }
        Ok(())
    }
}

impl Displayable for MemoryListResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                if !self.active_sessions.is_empty() {
                    println!("{}", "Active Sessions".bright_green().bold());
                    let mut table = components::TableBuilder::new();
                    table.add_header(vec!["ID", "Agent", "Messages", "Last Activity"]);
                    
                    for session in &self.active_sessions {
                        table.add_row(vec![
                            session.id.clone(),
                            session.agent.clone(),
                            session.message_count.to_string(),
                            components::format_time_ago(&session.last_activity)?,
                        ]);
                    }
                    table.print();
                }
                
                if !self.recent_sessions.is_empty() {
                    println!("\n{}", "Recent Sessions".bright_blue().bold());
                    // Similar table for recent sessions
                }
            }
            OutputFormat::Plain => {
                // Plain text format similar to current implementation
            }
        }
        Ok(())
    }
}

impl Displayable for ListPathResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                let mut table = components::TableBuilder::new();
                table.add_header(vec!["Name", "Type", "Size", "Created"]);
                
                for entry in &self.entries {
                    table.add_row(vec![
                        entry.name.clone(),
                        entry.entry_type.clone(),
                        entry.size.map(components::format_size).unwrap_or_default(),
                        entry.created.as_deref().unwrap_or("-").to_string(),
                    ]);
                }
                
                table.print();
            }
            OutputFormat::Plain => {
                println!("{}", self.path.bright_white());
                for entry in &self.entries {
                    let type_icon = if entry.entry_type == "directory" { "📁" } else { "📄" };
                    print!("{} {}", type_icon, entry.name);
                    
                    if let Some(size) = entry.size {
                        print!("\n   {}", components::format_size(size).dimmed());
                    }
                    
                    if let Some(created) = &entry.created {
                        print!("  {}", created.dimmed());
                    }
                    
                    println!();
                }
            }
        }
        Ok(())
    }
}
```

### Unit Tests

```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_possess_response_parsing() {
        let json = json!({
            "message": "Hello",
            "session_id": "test-123",
            "agent": "@ai-muse",
            "command_generated": true,
            "command_spec": {
                "name": "test-cmd",
                "description": "Test",
                "language": "bash"
            }
        });
        
        let response = PossessResponse::from_daemon_response(&json).unwrap();
        assert_eq!(response.message, "Hello");
        assert!(response.command_spec.is_some());
    }
    
    #[test]
    fn test_display_formatting() {
        // Test that display implementations produce expected output
    }
}
```

### Integration Tests

```rust
#[test]
fn test_possess_full_flow() {
    // Start real daemon
    // Send possess request
    // Verify response
    // Check command creation
}
```

### Step 9: Apply Pattern to Status, Daemon, and Init Commands

**First, create integration tests**:

**File: `cli/tests/status_daemon_integration.rs`**

```rust
#[test]
fn test_status_command() {
    let daemon = start_test_daemon();
    thread::sleep(Duration::from_millis(500));
    
    let mut client = DaemonClient::new(daemon.port());
    let result = status::handle_status(&mut client, OutputFormat::Plain);
    
    assert!(result.is_ok());
    // Could capture stdout and verify output format
}

#[test]
fn test_daemon_lifecycle() {
    // Test start (already running in test)
    // Test status
    // Test restart
    // Test stop
    // Each step should verify expected behavior
}

#[test]
fn test_init_command() {
    let test_dir = tempdir::TempDir::new("port42_test").unwrap();
    std::env::set_var("HOME", test_dir.path());
    
    // First init should succeed
    let result1 = init::handle_init(false);
    assert!(result1.is_ok());
    
    // Check directories were created
    assert!(test_dir.path().join(".port42/commands").exists());
    assert!(test_dir.path().join(".port42/memory").exists());
    
    // Second init without force should indicate already initialized
    let result2 = init::handle_init(false);
    assert!(result2.is_ok()); // Should succeed but with different message
}
```

**Then implement the refactored commands**:

These commands are related as they all deal with daemon state and initialization.

**File: `cli/src/protocol/requests.rs`**

```rust
// Status request/response
#[derive(Debug, Serialize)]
pub struct StatusRequest;

#[derive(Debug, Deserialize)]
pub struct StatusResponse {
    pub status: String,
    pub port: String,
    pub sessions: usize,
    pub uptime: String,
    pub dolphins: String,
}

impl RequestBuilder for StatusRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "status".to_string(),
            id,
            payload: serde_json::Value::Null,
        })
    }
}

// Memory request/response
#[derive(Debug, Serialize)]
pub struct MemoryRequest {
    pub session_id: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct MemoryListResponse {
    pub active_sessions: Vec<SessionInfo>,
    pub recent_sessions: Vec<SessionInfo>,
    pub stats: SessionStats,
}

#[derive(Debug, Deserialize)]
pub struct SessionInfo {
    pub id: String,
    pub agent: String,
    pub created_at: String,
    pub last_activity: String,
    pub message_count: usize,
    pub state: String,
}

// Virtual filesystem requests
#[derive(Debug, Serialize)]
pub struct ReadPathRequest {
    pub path: String,
}

#[derive(Debug, Deserialize)]
pub struct ReadPathResponse {
    pub content: String, // base64
    pub size: usize,
    pub path: String,
    pub metadata: Option<FileMetadata>,
}

#[derive(Debug, Serialize)]
pub struct ListPathRequest {
    pub path: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ListPathResponse {
    pub path: String,
    pub entries: Vec<PathEntry>,
}

#[derive(Debug, Deserialize)]
pub struct PathEntry {
    pub name: String,
    #[serde(rename = "type")]
    pub entry_type: String,
    pub size: Option<usize>,
    pub created: Option<String>,
    pub executable: Option<bool>,
}
```

**File: `cli/src/commands/status.rs`**

```rust
use crate::protocol::{StatusRequest, StatusResponse, RequestBuilder, ResponseParser};
use crate::display::Displayable;
use crate::common::{generate_id, Port42Error};
use crate::help_text;

pub fn handle_status(client: &mut DaemonClient, format: OutputFormat) -> Result<()> {
    // Build request using protocol types
    let request = StatusRequest.build_request(generate_id())?;
    
    // Send to daemon
    let response = client.request(request)?;
    
    if !response.success {
        return Err(Port42Error::Daemon(response.error.unwrap_or_default()).into());
    }
    
    // Parse response using trait
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    let status = StatusResponse::parse_response(&data)?;
    
    // Display using framework
    status.display(format)?;
    
    Ok(())
}
```

// Daemon control requests/responses
#[derive(Debug, Serialize)]
pub struct DaemonRequest {
    pub action: DaemonAction,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum DaemonAction {
    Start,
    Stop,
    Restart,
    Status,
    Logs { lines: Option<usize> },
}

#[derive(Debug, Deserialize)]
pub struct DaemonResponse {
    pub success: bool,
    pub message: String,
    pub pid: Option<u32>,
    pub logs: Option<Vec<String>>,
}

// Init request/response
#[derive(Debug, Serialize)]
pub struct InitRequest {
    pub force: bool,
}

#[derive(Debug, Deserialize)]
pub struct InitResponse {
    pub created_dirs: Vec<String>,
    pub already_initialized: bool,
}

impl RequestBuilder for DaemonRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "daemon".to_string(),
            id,
            payload: serde_json::to_value(self)?,
        })
    }
}

impl ResponseParser for DaemonResponse {
    type Output = Self;
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        serde_json::from_value(data.clone())
            .map_err(|e| anyhow!("Failed to parse daemon response: {}", e))
    }
}
```

**File: `cli/src/commands/daemon.rs`**

```rust
use crate::protocol::{DaemonRequest, DaemonResponse, DaemonAction, RequestBuilder, ResponseParser};
use crate::common::{generate_id, Port42Error};
use crate::help_text;

pub fn handle_daemon(action: DaemonAction, format: OutputFormat) -> Result<()> {
    match action {
        DaemonAction::Start => {
            println!("{}", help_text::MSG_DAEMON_STARTING);
            // Special handling for daemon start (spawn process)
            start_daemon()?;
            println!("{}", help_text::MSG_DAEMON_SUCCESS);
        }
        DaemonAction::Stop => {
            let mut client = DaemonClient::new(detect_port());
            println!("{}", help_text::MSG_DAEMON_STOPPING);
            
            let request = DaemonRequest { action }.build_request(generate_id())?;
            let response = client.request(request)?;
            
            if response.success {
                println!("{}", help_text::MSG_DAEMON_STOPPED);
            } else {
                return Err(Port42Error::Daemon(response.error.unwrap_or_default()).into());
            }
        }
        DaemonAction::Status => {
            println!("{}", help_text::MSG_CHECKING_STATUS);
            // Reuse status command handler
            handle_status(&mut DaemonClient::new(detect_port()), format)?;
        }
        DaemonAction::Logs { lines } => {
            let mut client = DaemonClient::new(detect_port());
            let request = DaemonRequest { action }.build_request(generate_id())?;
            let response = client.request(request)?;
            
            if let Some(logs) = response.data.and_then(|d| d.get("logs")) {
                println!("{}", help_text::MSG_DAEMON_LOGS);
                // Display logs
            }
        }
        _ => {}
    }
    Ok(())
}
```

**File: `cli/src/commands/init.rs`**

```rust
use crate::help_text;
use std::fs;
use std::path::PathBuf;

pub fn handle_init(force: bool) -> Result<()> {
    let port42_dir = dirs::home_dir()
        .ok_or_else(|| anyhow!("Could not find home directory"))?
        .join(".port42");
    
    if port42_dir.exists() && !force {
        println!("{}", help_text::MSG_ALREADY_INIT);
        return Ok(());
    }
    
    println!("{}", help_text::MSG_INIT_BEGIN);
    println!("{}", help_text::MSG_CREATING_DIRS);
    
    // Create directories
    let dirs = vec!["commands", "memory", "templates", "artifacts"];
    for dir in &dirs {
        let path = port42_dir.join(dir);
        fs::create_dir_all(&path)?;
    }
    
    println!("\n{}", help_text::MSG_CREATED_LABEL);
    println!("{}", help_text::MSG_DIR_COMMANDS);
    println!("{}", help_text::MSG_DIR_MEMORY);
    println!("{}", help_text::MSG_DIR_TEMPLATES);
    println!("\n{}", help_text::MSG_INIT_SUCCESS);
    
    Ok(())
}
```

### Step 10: Apply Pattern to Reality Command and Remove Evolve

The reality command lists crystallized commands from the filesystem. Even though it doesn't use the daemon, we apply the same display patterns for consistency. The evolve command is not implemented and should be removed.

**File: `cli/src/commands/reality_types.rs`**

```rust
// Reality doesn't need request/response types since it reads filesystem directly
// But we create structured types for business logic and display separation
#[derive(Debug)]
pub struct RealityData {
    pub commands: Vec<CommandInfo>,
    pub total: usize,
}

#[derive(Debug)]
pub struct CommandInfo {
    pub name: String,
    pub path: PathBuf,
    pub language: String,
    pub created: Option<String>,
    pub agent: Option<String>,
}

// Implement Displayable for consistent output formatting
impl Displayable for RealityData {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                let mut table = components::TableBuilder::new();
                table.add_header(vec!["Command", "Language", "Agent", "Created"]);
                
                for cmd in &self.commands {
                    table.add_row(vec![
                        cmd.name.clone(),
                        cmd.language.clone(),
                        cmd.agent.as_deref().unwrap_or("-").to_string(),
                        cmd.created.as_deref().unwrap_or("-").to_string(),
                    ]);
                }
                
                table.print();
                println!("\n{}", help_text::format_total_commands(self.total));
            }
            OutputFormat::Plain => {
                if self.commands.is_empty() {
                    println!("{}", "No commands found in ~/.port42/commands".dimmed());
                    println!("\n{}", "Generate your first command:".yellow());
                    println!("  {}", "port42 possess @ai-muse".bright_white());
                } else {
                    for cmd in &self.commands {
                        println!("  {}", cmd.name.bright_cyan());
                    }
                    println!("\n{}", help_text::format_total_commands(self.total));
                }
            }
        }
        Ok(())
    }
}
```

**File: `cli/src/commands/reality.rs`**

```rust
use crate::display::Displayable;
use crate::help_text;
use std::fs;
use std::path::PathBuf;

pub fn handle_reality(
    verbose: bool,
    agent_filter: Option<String>,
) -> Result<()> {
    println!("{}", help_text::MSG_COMMANDS_HEADER);
    println!();
    
    let commands_dir = dirs::home_dir()
        .context("Could not find home directory")?  
        .join(".port42")
        .join("commands");
    
    if !commands_dir.exists() {
        println!("{}", "No commands directory found".dimmed());
        println!("\n{}", "Generate your first command:".yellow());
        println!("  {}", "port42 possess @ai-muse".bright_white());
        return Ok(());
    }
    
    let mut commands = Vec::new();
    
    // Read all files in commands directory
    for entry in fs::read_dir(&commands_dir)? {
        let entry = entry?;
        let path = entry.path();
        
        if path.is_file() {
            if let Some(name) = path.file_name().and_then(|n| n.to_str()) {
                // Skip hidden files and backup files
                if !name.starts_with('.') && !name.ends_with('~') {
                    // Check if executable
                    #[cfg(unix)]
                    {
                        use std::os::unix::fs::PermissionsExt;
                        let metadata = fs::metadata(&path)?;
                        if metadata.permissions().mode() & 0o111 != 0 {
                            commands.push(CommandInfo {
                                name: name.to_string(),
                                path: path.clone(),
                                language: detect_language(&path),
                                created: None, // Could read from metadata
                                agent: None, // Could parse from file header
                            });
                        }
                    }
                    
                    #[cfg(not(unix))]
                    {
                        commands.push(CommandInfo {
                            name: name.to_string(),
                            path: path.clone(),
                            language: detect_language(&path),
                            created: None,
                            agent: None,
                        });
                    }
                }
            }
        }
    }
    
    // Apply agent filter if provided
    if let Some(agent) = agent_filter {
        commands.retain(|cmd| cmd.agent.as_deref() == Some(&agent));
    }
    
    // Sort by name
    commands.sort_by(|a, b| a.name.cmp(&b.name));
    
    // Create structured data for display
    let reality_data = RealityData {
        total: commands.len(),
        commands,
    };
    
    // Display using the same framework as other commands
    reality_data.display(if verbose { OutputFormat::Table } else { OutputFormat::Plain })?;
    
    Ok(())
}

fn detect_language(path: &PathBuf) -> String {
    match path.extension().and_then(|e| e.to_str()) {
        Some("sh") => "bash".to_string(),
        Some("py") => "python".to_string(),
        Some("js") => "javascript".to_string(),
        Some("rb") => "ruby".to_string(),
        _ => "unknown".to_string(),
    }
}
```

**Key Points**:
- Reality reads filesystem directly, no daemon interaction needed
- Still uses the same display framework for consistency
- Business logic (filesystem reading) separated from display logic
- Remove `evolve.rs` from the commands directory as it's not implemented
- There is no separate 'list' command - 'reality' handles listing commands

### Step 11: Apply Pattern to Memory Command

**Integration tests first**:

```rust
#[test]
fn test_memory_list() {
    let daemon = start_test_daemon();
    thread::sleep(Duration::from_millis(500));
    
    // Create some sessions first
    create_test_sessions(&daemon);
    
    let mut client = DaemonClient::new(daemon.port());
    let result = memory::handle_memory(&mut client, None, OutputFormat::Plain);
    
    assert!(result.is_ok());
}

#[test]
fn test_memory_detail() {
    let daemon = start_test_daemon();
    thread::sleep(Duration::from_millis(500));
    
    let session_id = "test-session-123";
    create_test_session(&daemon, session_id);
    
    let mut client = DaemonClient::new(daemon.port());
    let result = memory::handle_memory(
        &mut client, 
        Some(session_id.to_string()), 
        OutputFormat::Plain
    );
    
    assert!(result.is_ok());
}
```

**Then the implementation**:

**File: `cli/src/commands/memory.rs`**

```rust
use crate::protocol::{MemoryRequest, MemoryListResponse, MemoryDetailResponse};
use crate::display::Displayable;

pub fn handle_memory(
    client: &mut DaemonClient,
    session_id: Option<String>,
    format: OutputFormat,
) -> Result<()> {
    let request = MemoryRequest { session_id: session_id.clone() }
        .build_request(generate_id())?;
    
    let response = client.request(request)?;
    
    if !response.success {
        return Err(anyhow!("Memory query failed: {}", response.error.unwrap_or_default()));
    }
    
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    
    // Parse based on request type
    if session_id.is_some() {
        let detail: MemoryDetailResponse = serde_json::from_value(data)?;
        detail.display(format)?;
    } else {
        let list: MemoryListResponse = serde_json::from_value(data)?;
        list.display(format)?;
    }
    
    Ok(())
}
```

### Step 12: Apply Pattern to Cat and Info Commands

These commands both read from the virtual filesystem but display different aspects.

**File: `cli/src/protocol/requests.rs`**

```rust
// Info request/response
#[derive(Debug, Serialize)]
pub struct InfoRequest {
    pub path: String,
}

#[derive(Debug, Deserialize)]
pub struct InfoResponse {
    pub path: String,
    pub object_type: String,
    pub created: String,
    pub modified: Option<String>,
    pub size: usize,
    pub hash: String,
    pub metadata: HashMap<String, serde_json::Value>,
    pub virtual_paths: Vec<String>,
}

impl RequestBuilder for InfoRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "info".to_string(),
            id,
            payload: serde_json::to_value(self)?,
        })
    }
}

impl ResponseParser for InfoResponse {
    type Output = Self;
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        serde_json::from_value(data.clone())
            .map_err(|e| anyhow!("Failed to parse info response: {}", e))
    }
}
```

**File: `cli/src/commands/info.rs`**

```rust
use crate::protocol::{InfoRequest, InfoResponse, RequestBuilder, ResponseParser};
use crate::display::Displayable;
use crate::common::{generate_id, Port42Error};
use crate::help_text;

pub fn handle_info(client: &mut DaemonClient, path: String, format: OutputFormat) -> Result<()> {
    let request = InfoRequest { path: path.clone() }
        .build_request(generate_id())?;
    
    let response = client.request(request)?;
    
    if !response.success {
        if response.error.as_deref() == Some("Path not found") {
            return Err(Port42Error::PathNotFound(path).into());
        }
        return Err(Port42Error::Daemon(response.error.unwrap_or_default()).into());
    }
    
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    let info = InfoResponse::parse_response(&data)?;
    
    match format {
        OutputFormat::Json => {
            println!("{}", serde_json::to_string_pretty(&info)?);
        }
        _ => {
            // Display metadata in human-readable format
            println!("📊 Object Metadata");
            println!("═══════════════════");
            println!("Path:     {}", info.path);
            println!("Type:     {}", info.object_type);
            println!("Size:     {}", components::format_size(info.size));
            println!("Created:  {}", info.created);
            if let Some(modified) = &info.modified {
                println!("Modified: {}", modified);
            }
            println!("Hash:     {}", info.hash);
            
            if !info.virtual_paths.is_empty() {
                println!("\nVirtual Paths:");
                for vpath in &info.virtual_paths {
                    println!("  • {}", vpath);
                }
            }
            
            if !info.metadata.is_empty() {
                println!("\nMetadata:");
                for (key, value) in &info.metadata {
                    println!("  {}: {}", key, value);
                }
            }
        }
    }
    
    Ok(())
}
```

**File: `cli/src/commands/cat.rs`**

```rust
use crate::protocol::{ReadPathRequest, ReadPathResponse};
use crate::display::Displayable;
use base64;

pub fn handle_cat(client: &mut DaemonClient, path: String, format: OutputFormat) -> Result<()> {
    let request = ReadPathRequest { path: path.clone() }
        .build_request(generate_id())?;
    
    let response = client.request(request)?;
    
    if !response.success {
        return Err(anyhow!("Read failed: {}", response.error.unwrap_or_default()));
    }
    
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    let read_response: ReadPathResponse = serde_json::from_value(data)?;
    
    match format {
        OutputFormat::Json => {
            println!("{}", serde_json::to_string_pretty(&read_response)?);
        }
        _ => {
            // Decode and display content
            let content = base64::decode(&read_response.content)?;
            let content_str = String::from_utf8(content)?;
            
            // Display with syntax highlighting if possible
            if let Some(metadata) = &read_response.metadata {
                if let Some(format) = &metadata.format {
                    // Use syntect or similar for syntax highlighting
                }
            }
            
            println!("{}", content_str);
        }
    }
    
    Ok(())
}
```

### Step 13: Apply Pattern to Ls Command

**File: `cli/src/commands/ls.rs`**

```rust
use crate::protocol::{ListPathRequest, ListPathResponse, RequestBuilder, ResponseParser};
use crate::display::Displayable;
use crate::common::{generate_id, Port42Error};
use crate::help_text;

pub fn handle_ls(client: &mut DaemonClient, path: String, format: OutputFormat) -> Result<()> {
    let request = ListPathRequest { path: Some(path.clone()) }
        .build_request(generate_id())?;
    
    let response = client.request(request)?;
    
    if !response.success {
        if response.error.as_deref() == Some("Path not found") {
            return Err(Port42Error::PathNotFound(path).into());
        }
        return Err(Port42Error::Daemon(response.error.unwrap_or_default()).into());
    }
    
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    let list_response = ListPathResponse::parse_response(&data)?;
    
    list_response.display(format)?;
    
    Ok(())
}
```

### Step 14: Apply Pattern to Search Command

**File: `cli/src/commands/search.rs`**

```rust
use crate::protocol::{SearchRequest, SearchResponse, RequestBuilder, ResponseParser};
use crate::display::Displayable;
use crate::common::{generate_id, Port42Error};
use crate::help_text;

pub fn handle_search(
    client: &mut DaemonClient,
    query: String,
    filters: Option<SearchFilters>,
    format: OutputFormat,
) -> Result<()> {
    println!("{}", help_text::format_searching(&query));
    
    let request = SearchRequest { query, filters }
        .build_request(generate_id())?;
    
    let response = client.request(request)?;
    
    if !response.success {
        return Err(Port42Error::Daemon(response.error.unwrap_or_default()).into());
    }
    
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    let search_response = SearchResponse::parse_response(&data)?;
    
    if search_response.results.is_empty() {
        println!("{}", help_text::MSG_NO_RESULTS);
    } else {
        let plural = if search_response.results.len() == 1 { "" } else { "es" };
        println!("{}", help_text::format_found_results(
            search_response.results.len() as u64,
            plural,
            &search_response.query
        ));
        search_response.display(format)?;
    }
    
    Ok(())
}
```

### Step 15: Update Main Entry Point

**File: `cli/src/main.rs`** (updated)

```rust
use clap::{Parser, Subcommand};
use anyhow::Result;

mod client;
mod protocol;
mod display;
mod commands;
mod common;
mod possess;

use crate::display::OutputFormat;
use crate::common::CommonOpts;

#[derive(Parser)]
#[command(name = "port42")]
#[command(about = "Port 42 - Reality Compiler")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
    
    #[arg(short, long, value_enum, default_value = "plain")]
    format: OutputFormat,
    
    #[arg(short, long)]
    verbose: bool,
    
    #[arg(short, long)]
    debug: bool,
}

#[derive(Subcommand)]
enum Commands {
    /// Channel AI consciousness
    Possess {
        agent: String,
        message: String,
        #[arg(short, long)]
        session: Option<String>,
        #[arg(short, long)]
        interactive: bool,
    },
    
    /// Show daemon status
    Status,
    
    /// List available commands
    List {
        #[arg(short, long)]
        filter: Option<String>,
    },
    
    /// Query memory/sessions
    Memory {
        session_id: Option<String>,
    },
    
    /// Read virtual filesystem
    Cat {
        path: String,
    },
    
    /// List virtual filesystem
    Ls {
        #[arg(default_value = "/")]
        path: String,
    },
}

fn main() -> Result<()> {
    let cli = Cli::parse();
    
    // Set up logging
    if cli.debug {
        std::env::set_var("PORT42_DEBUG", "1");
    }
    if cli.verbose {
        std::env::set_var("PORT42_VERBOSE", "1");
    }
    
    // Create client
    let mut client = DaemonClient::new(detect_port());
    
    // Common options
    let opts = CommonOpts {
        format: cli.format,
        verbose: cli.verbose,
        debug: cli.debug,
    };
    
    // Execute command
    match cli.command {
        Commands::Possess { agent, message, session, interactive } => {
            commands::possess::handle_possess(&mut client, agent, message, session, interactive)?;
        }
        Commands::Status => {
            commands::status::handle_status(&mut client, opts.format)?;
        }
        Commands::List { filter } => {
            commands::list::handle_list(&mut client, filter, opts.format)?;
        }
        Commands::Memory { session_id } => {
            commands::memory::handle_memory(&mut client, session_id, opts.format)?;
        }
        Commands::Cat { path } => {
            commands::cat::handle_cat(&mut client, path, opts.format)?;
        }
        Commands::Ls { path } => {
            commands::ls::handle_ls(&mut client, path, opts.format)?;
        }
    }
    
    Ok(())
}
```

### Step 16: Remove Old Duplicate Code

Once all tests pass:
1. Remove old response parsing code from individual commands
2. Remove duplicate display logic
3. Remove manual JSON construction
4. Clean up imports and dependencies

### Step 17: Update Documentation

1. Update README with new architecture
2. Create developer guide for adding new commands
3. Document the protocol types and traits
4. Add examples of using the new patterns

## Architecture Diagram

After completing all steps, the architecture will look like:

```
┌─────────────────────────────────────────────────────────┐
│                   CLI Entry Point                        │
│                    (main.rs)                            │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              Command Implementations                     │
│  - Use protocol traits and types                        │
│  - Use common error handling                            │
│  - All messages via help_text.rs                        │
├─────────────────────────┴───────────────────────────────┤
│ possess │ list │ memory │ status │ cat │ ls │ etc.   │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              Protocol Abstraction Layer                  │
│  - DaemonRequest/DaemonResponse types                   │
│  - RequestBuilder and ResponseParser traits             │
│  - Type-safe command-specific types                     │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                 Display Framework                        │
│  - Command-specific display traits                      │
│  - Consistent formatting using help_text.rs             │
│  - Interactive vs non-interactive modes                 │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  Common Libraries                        │
│  - Error types with help_text integration               │
│  - Utility functions (generate_id, etc.)                │
│  - All user messages via help_text.rs                   │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                Transport Layer (Client)                  │
│  - Low-level TCP communication                          │
│  - Connection handling                                   │
│  - Retry logic                                          │
└─────────────────────────────────────────────────────────┘
```

## Migration Strategy

1. **Implement Steps 1-6** for possess command as proof of concept
2. **Test thoroughly** to ensure no regressions
3. **Apply Steps 7-14** to remaining commands one by one
4. **Run integration tests** after each command migration
5. **Clean up** with Steps 15-17

## Success Metrics

- **Code reduction**: ~60% less code overall
- **Single source of truth**: One place for each concern
- **Consistent display**: All commands use same formatting
- **Testable components**: Unit tests for all layers
- **Clean separation**: Each layer has single responsibility
- **Developer velocity**: Adding new commands is trivial

## Key Patterns

Each command follows the same pattern:
1. Build typed request using RequestBuilder trait
2. Send to daemon with proper error handling
3. Parse typed response using ResponseParser trait  
4. Display using framework with help_text.rs messages

This eliminates all duplicate request building, response parsing, error handling, and display logic across commands while ensuring consistent user messaging.

## Key Differences from Comprehensive Plan

- **No daemon changes** - Work with existing protocol
- **Minimal abstractions** - Just enough to eliminate duplication
- **Focused scope** - Possess command only initially
- **Incremental approach** - Prove concept before expanding
- **Pragmatic testing** - Real daemon for integration tests

This approach delivers immediate value while setting up patterns for future refactoring.