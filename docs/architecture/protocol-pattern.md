# Port 42 Protocol Pattern

## Overview

The Port 42 CLI uses a protocol abstraction pattern to eliminate code duplication and provide a consistent interface for all commands. This pattern separates request building, response parsing, and display logic into distinct, reusable components.

## Core Components

### 1. Protocol Types (`cli/src/protocol/mod.rs`)

The foundation of the pattern consists of two base types:

```rust
pub struct DaemonRequest {
    #[serde(rename = "type")]
    pub request_type: String,
    pub id: String,
    pub payload: serde_json::Value,
}

pub struct DaemonResponse {
    pub id: String,
    pub success: bool,
    pub data: Option<serde_json::Value>,
    pub error: Option<String>,
}
```

### 2. Request Builder Trait

Every command implements the `RequestBuilder` trait to construct daemon requests:

```rust
pub trait RequestBuilder {
    fn build_request(&self, id: String) -> Result<DaemonRequest>;
}
```

### 3. Response Parser Trait

Commands implement `ResponseParser` to handle daemon responses:

```rust
pub trait ResponseParser {
    type Output;
    fn parse_response(data: &serde_json::Value) -> Result<Self::Output>;
}
```

### 4. Display Trait

The `Displayable` trait provides consistent output formatting:

```rust
pub trait Displayable {
    fn display(&self, format: OutputFormat) -> Result<()>;
}

pub enum OutputFormat {
    Plain,
    Json,
    Table,
}
```

## Implementation Example: Status Command

Here's how the status command implements the pattern:

### Protocol Types (`protocol/status.rs`)

```rust
// Request type
pub struct StatusRequest;

impl RequestBuilder for StatusRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "status".to_string(),
            id,
            payload: json!({}),
        })
    }
}

// Response type
pub struct StatusResponse {
    pub status: String,
    pub uptime: String,
    pub active_sessions: u64,
    pub total_commands: u64,
    pub version: String,
}

impl ResponseParser for StatusResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        Ok(StatusResponse {
            status: data["status"].as_str().unwrap_or("unknown").to_string(),
            uptime: data["uptime"].as_str().unwrap_or("0s").to_string(),
            active_sessions: data["active_sessions"].as_u64().unwrap_or(0),
            total_commands: data["total_commands"].as_u64().unwrap_or(0),
            version: data["version"].as_str().unwrap_or("unknown").to_string(),
        })
    }
}

impl Displayable for StatusResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Plain => {
                println!("{}", "✨ Gateway Status:".green().bold());
                println!("  Status:   {}", self.status.bright_green());
                println!("  Uptime:   {}", self.uptime);
                println!("  Sessions: {}", self.active_sessions);
                println!("  Commands: {}", self.total_commands);
            }
            OutputFormat::Table => {
                // Table format implementation
            }
        }
        Ok(())
    }
}
```

### Command Handler (`commands/status.rs`)

```rust
pub fn handle_status_with_format(
    client: &mut DaemonClient, 
    detailed: bool, 
    format: OutputFormat
) -> Result<()> {
    // Build request
    let request = StatusRequest.build_request(generate_id())?;
    
    // Send to daemon
    let response = client.request(request)?;
    
    if !response.success {
        return Err(anyhow!("Status request failed: {}", 
            response.error.unwrap_or_default()));
    }
    
    // Parse response
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    let status_response = StatusResponse::parse_response(&data)?;
    
    // Display results
    status_response.display(format)?;
    
    Ok(())
}
```

## Adding a New Command

To add a new command using this pattern:

### 1. Create Protocol Types

Create a new file in `cli/src/protocol/` for your command:

```rust
// protocol/mycommand.rs
use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use anyhow::Result;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize)]
pub struct MyCommandRequest {
    pub parameter: String,
}

impl RequestBuilder for MyCommandRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "my_command".to_string(),
            id,
            payload: serde_json::to_value(self)?,
        })
    }
}

#[derive(Debug, Deserialize)]
pub struct MyCommandResponse {
    pub result: String,
}

impl ResponseParser for MyCommandResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        serde_json::from_value(data.clone())
            .map_err(|e| anyhow!("Failed to parse response: {}", e))
    }
}

impl Displayable for MyCommandResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Plain => {
                println!("Result: {}", self.result);
            }
            _ => unimplemented!()
        }
        Ok(())
    }
}
```

### 2. Add to Protocol Module

Update `cli/src/protocol/mod.rs`:

```rust
pub mod mycommand;
pub use mycommand::*;
```

### 3. Create Command Handler

Create `cli/src/commands/mycommand.rs`:

```rust
use anyhow::Result;
use crate::client::DaemonClient;
use crate::protocol::{MyCommandRequest, MyCommandResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use crate::common::generate_id;

pub fn handle_mycommand(client: &mut DaemonClient, parameter: String) -> Result<()> {
    handle_mycommand_with_format(client, parameter, OutputFormat::Plain)
}

pub fn handle_mycommand_with_format(
    client: &mut DaemonClient,
    parameter: String,
    format: OutputFormat
) -> Result<()> {
    // Create request
    let request = MyCommandRequest { parameter };
    let daemon_request = request.build_request(generate_id())?;
    
    // Send request
    let response = client.request(daemon_request)?;
    
    if !response.success {
        anyhow::bail!("Command failed: {}", 
            response.error.unwrap_or_else(|| "Unknown error".to_string()));
    }
    
    // Parse response
    let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
    let cmd_response = MyCommandResponse::parse_response(&data)?;
    
    // Display
    cmd_response.display(format)?;
    
    Ok(())
}
```

### 4. Wire into CLI

Add to `cli/src/main.rs`:

```rust
#[derive(Subcommand)]
pub enum Commands {
    // ... existing commands ...
    
    /// Description of your command
    MyCommand {
        /// Parameter description
        parameter: String,
    },
}

// In main():
Some(Commands::MyCommand { parameter }) => {
    let mut client = DaemonClient::new(port);
    if cli.json {
        mycommand::handle_mycommand_with_format(&mut client, parameter, OutputFormat::Json)?;
    } else {
        mycommand::handle_mycommand(&mut client, parameter)?;
    }
}
```

## Benefits

1. **Consistency**: All commands follow the same pattern
2. **Reusability**: Common functionality is shared
3. **Type Safety**: Strong typing throughout
4. **Flexibility**: Easy to add new output formats
5. **Testability**: Each component can be tested independently
6. **Maintainability**: Changes to the protocol only need updates in one place

## Best Practices

1. Keep request types simple and focused
2. Parse responses defensively with proper error handling
3. Support JSON output for all commands
4. Use the help_text module for consistent error messages
5. Follow the Reality Compiler language for user-facing messages