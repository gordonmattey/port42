# Adding New Commands to Port 42

This guide walks through adding a new command to the Port 42 CLI using our protocol abstraction pattern.

## Quick Checklist

- [ ] Create protocol types in `cli/src/protocol/`
- [ ] Implement RequestBuilder, ResponseParser, and Displayable traits
- [ ] Create command handler in `cli/src/commands/`
- [ ] Add command to CLI enum in `main.rs`
- [ ] Add help text constants in `help_text.rs`
- [ ] Update command exports in module files
- [ ] Test the command with daemon running

## Detailed Steps

### Step 1: Design Your Command

Before coding, define:
- Command name and purpose
- Required parameters
- Expected daemon response format
- Display requirements (plain text, JSON, table)

### Step 2: Create Protocol Types

Create a new file: `cli/src/protocol/yourcommand.rs`

```rust
use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::json;
use colored::*;

// Request structure - what we send to daemon
#[derive(Debug, Serialize)]
pub struct YourCommandRequest {
    pub param1: String,
    pub param2: Option<String>,
}

// Implement request building
impl RequestBuilder for YourCommandRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "your_command".to_string(), // Must match daemon expectation
            id,
            payload: json!({
                "param1": self.param1,
                "param2": self.param2,
            }),
        })
    }
}

// Response structure - what we get from daemon
#[derive(Debug, Serialize, Deserialize)]
pub struct YourCommandResponse {
    pub result: String,
    pub details: Vec<String>,
    pub count: usize,
}

// Implement response parsing
impl ResponseParser for YourCommandResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        // Option 1: Direct deserialization
        serde_json::from_value(data.clone())
            .map_err(|e| anyhow::anyhow!("Failed to parse response: {}", e))
        
        // Option 2: Manual parsing for more control
        // Ok(YourCommandResponse {
        //     result: data["result"].as_str()
        //         .ok_or_else(|| anyhow::anyhow!("Missing result field"))?
        //         .to_string(),
        //     details: ... 
        // })
    }
}

// Implement display formatting
impl Displayable for YourCommandResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Plain => {
                // Use Reality Compiler language
                println!("{}", "✨ Command Results:".bright_blue().bold());
                println!("Result: {}", self.result.green());
                println!("Found {} items:", self.count);
                for detail in &self.details {
                    println!("  • {}", detail);
                }
            }
            OutputFormat::Table => {
                // Optional: implement table format using prettytable
                unimplemented!("Table format not yet implemented")
            }
        }
        Ok(())
    }
}
```

### Step 3: Export Protocol Types

Add to `cli/src/protocol/mod.rs`:

```rust
pub mod yourcommand;
pub use yourcommand::*;
```

### Step 4: Create Command Handler

Create `cli/src/commands/yourcommand.rs`:

```rust
use anyhow::{Result, Context};
use crate::client::DaemonClient;
use crate::protocol::{YourCommandRequest, YourCommandResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use crate::common::{generate_id, errors::Port42Error};
use crate::help_text;

// Main handler - delegates to format-aware version
pub fn handle_yourcommand(
    client: &mut DaemonClient,
    param1: String,
    param2: Option<String>
) -> Result<()> {
    handle_yourcommand_with_format(client, param1, param2, OutputFormat::Plain)
}

// Format-aware handler
pub fn handle_yourcommand_with_format(
    client: &mut DaemonClient,
    param1: String,
    param2: Option<String>,
    format: OutputFormat
) -> Result<()> {
    // Show status message (skip for JSON output)
    if format != OutputFormat::Json {
        println!("{}", help_text::MSG_PROCESSING.blue().bold());
    }
    
    // Build request
    let request = YourCommandRequest { param1, param2 };
    let daemon_request = request.build_request(generate_id())?;
    
    // Send to daemon
    let response = client.request(daemon_request)
        .context(help_text::ERR_CONNECTION_LOST)?;
    
    // Check for errors
    if !response.success {
        let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
        return Err(Port42Error::Daemon(error).into());
    }
    
    // Parse response
    let data = response.data
        .ok_or_else(|| anyhow::anyhow!(help_text::ERR_INVALID_RESPONSE))?;
    let cmd_response = YourCommandResponse::parse_response(&data)?;
    
    // Display results
    cmd_response.display(format)?;
    
    Ok(())
}
```

### Step 5: Export Command Handler

Add to `cli/src/commands/mod.rs`:

```rust
pub mod yourcommand;
```

### Step 6: Add to CLI

Update `cli/src/main.rs`:

```rust
// Add to Commands enum
#[derive(Subcommand)]
pub enum Commands {
    // ... existing commands ...
    
    #[command(about = crate::help_text::YOURCOMMAND_DESC)]
    /// Brief description for --help
    YourCommand {
        /// Parameter 1 description
        param1: String,
        
        /// Optional parameter 2 description
        #[arg(short, long)]
        param2: Option<String>,
    },
}

// Add to match statement in main()
Some(Commands::YourCommand { param1, param2 }) => {
    let mut client = DaemonClient::new(port);
    if cli.json {
        commands::yourcommand::handle_yourcommand_with_format(
            &mut client, param1, param2, OutputFormat::Json
        )?;
    } else {
        commands::yourcommand::handle_yourcommand(&mut client, param1, param2)?;
    }
}
```

### Step 7: Add Help Text

Update `cli/src/help_text.rs`:

```rust
// Command description
pub const YOURCOMMAND_DESC: &str = "Brief description of what the command does";

// Add detailed help function if needed
pub fn yourcommand_help() -> String {
    format!(r#"{}

{}

{}
  yourcommand <param1>              Basic usage
  yourcommand <param1> -p value     With optional parameter
  yourcommand "quoted param"        Handle spaces

Each command crystallizes thought into reality."#,
        "Detailed command description.".bright_blue().bold(),
        "Usage: yourcommand <param1> [options]".yellow(),
        "Examples:".bright_cyan()
    )
}

// Add any command-specific messages
pub const MSG_PROCESSING: &str = "🔮 Processing your request...";
```

### Step 8: Testing

1. **Build the CLI**:
   ```bash
   cd cli
   cargo build
   ```

2. **Start the daemon** (make sure it handles your new request type):
   ```bash
   ./bin/port42d
   ```

3. **Test your command**:
   ```bash
   # Basic test
   ./target/debug/port42 yourcommand "test"
   
   # With optional parameter
   ./target/debug/port42 yourcommand "test" -p "value"
   
   # JSON output
   ./target/debug/port42 yourcommand "test" --json
   
   # Help text
   ./target/debug/port42 help yourcommand
   ```

## Common Patterns

### Commands with Subcommands

For commands with subcommands (like `daemon start/stop`):

```rust
#[derive(Subcommand)]
pub enum YourAction {
    Start { 
        #[arg(short, long)]
        background: bool 
    },
    Stop,
    Status,
}

// In Commands enum:
YourCommand {
    #[command(subcommand)]
    action: YourAction,
}
```

### Commands that List Resources

Use consistent patterns from memory/reality commands:

```rust
pub struct ListResponse {
    pub items: Vec<Item>,
    pub total: usize,
}

impl Displayable for ListResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Plain => {
                if self.items.is_empty() {
                    println!("{}", help_text::MSG_NO_ITEMS);
                } else {
                    for item in &self.items {
                        println!("• {}", item.name);
                    }
                    println!("\nTotal: {}", self.total);
                }
            }
            // ...
        }
        Ok(())
    }
}
```

### Error Handling

Use the existing error types and help text:

```rust
// For daemon errors
return Err(Port42Error::Daemon("Specific error message".to_string()).into());

// For connection errors
.context(help_text::ERR_CONNECTION_LOST)?;

// For invalid input
anyhow::bail!(help_text::format_error_with_suggestion(
    help_text::ERR_INVALID_INPUT,
    "Try: yourcommand --help"
));
```

## Tips

1. **Follow Reality Compiler Language**: Use metaphysical language in user-facing messages
2. **Be Consistent**: Look at similar commands for patterns
3. **Handle Edge Cases**: Empty results, missing fields, connection errors
4. **Support JSON**: Always implement JSON output for scripting
5. **Test Thoroughly**: Test success cases, errors, and edge cases
6. **Document Well**: Update help text and add examples

## Need Help?

- Review existing commands in `cli/src/commands/` for examples
- Check `protocol/` for similar request/response patterns
- Look at `help_text.rs` for consistent messaging
- Run `cargo clippy` for style suggestions