use clap::{Parser, Subcommand};
use colored::*;
use anyhow::Result;

mod boot;
mod commands;
mod client;
mod types;
mod interactive;
mod shell;
mod help_text;
mod help_handler;
mod protocol;
mod possess;
mod common;
mod ui;
mod display;
mod context;

use commands::*;

#[derive(Parser)]
#[command(
    name = "port42",
    about = crate::help_text::MAIN_ABOUT,
    long_about = crate::help_text::MAIN_LONG_ABOUT,
    version,
    author
)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
    
    /// Port for consciousness gateway (default: 42, fallback: 4242)
    #[arg(short, long, global = true, env = "PORT42_PORT")]
    port: Option<u16>,
    
    /// Verbose output for deeper introspection
    #[arg(short, long, global = true)]
    verbose: bool,
    
    /// Output in JSON format for machine processing
    #[arg(short, long, global = true)]
    json: bool,
}

#[derive(Subcommand)]
pub enum Commands {
    
    #[command(about = crate::help_text::DAEMON_DESC)]
    /// Manage the consciousness gateway
    Daemon {
        #[command(subcommand)]
        action: DaemonAction,
    },
    
    #[command(about = crate::help_text::STATUS_DESC)]
    /// Check the daemon's pulse
    Status {
        /// Show detailed status information
        #[arg(short, long)]
        detailed: bool,
    },
    
    #[command(about = crate::help_text::REALITY_DESC)]
    /// View your crystallized commands
    Reality {
        /// Show detailed information about each command
        #[arg(short, long)]
        verbose: bool,
        
        /// Filter by agent who created the command
        #[arg(short, long)]
        agent: Option<String>,
    },
    
    #[command(about = "View current Port42 context and active session")]
    /// Show context information
    Context {
        /// Pretty-print for human reading (default is JSON)
        #[arg(long)]
        pretty: bool,
        
        /// Compact single-line format
        #[arg(long)]
        compact: bool,
    },
    
    #[command(about = crate::help_text::POSSESS_DESC)]
    /// Channel an AI agent's consciousness
    Possess {
        /// AI agent to possess (@ai-engineer, @ai-muse, @ai-analyst, @ai-founder)
        agent: String,
        
        /// Session ID to resume, or 'last' for most recent
        #[arg(long, help = "Session ID to resume, or 'last' for most recent")]
        session: Option<String>,
        
        /// Reference entities for context (file:path, p42:/commands/name, url:https://, search:"query")
        #[arg(long = "ref", action = clap::ArgAction::Append, help = "Reference other entities for context in conversation (can be used multiple times)\n\nAvailable reference types:\n• file:./path/to/file    - Include local file content\n• p42:/commands/name     - Reference existing command or tool\n• url:https://api.docs   - Fetch web content for context\n• search:\"query terms\"   - Load relevant memories/tools\n\nExample: --ref file:./config.json --ref search:\"error patterns\"")]
        references: Option<Vec<String>>,
        
        /// Message to send to the AI
        #[arg(trailing_var_arg = true)]
        message: Vec<String>,
    },
    
    /// Declare that something should exist in reality
    Declare {
        /// Type of relation to declare
        #[command(subcommand)]
        command: DeclareCommand,
    },
    
    #[command(about = crate::help_text::MEMORY_DESC)]
    /// Browse the persistent memory of conversations
    Memory {
        /// Session ID to show, or 'search' followed by query
        args: Vec<String>,
    },
    
    #[command(about = crate::help_text::LS_DESC)]
    /// List contents of the virtual filesystem
    Ls {
        /// Path to list (default: /)
        path: Option<String>,
    },
    
    #[command(about = crate::help_text::CAT_DESC)]
    /// Display content from any reality path
    Cat {
        /// Path to read
        path: String,
    },
    
    #[command(about = crate::help_text::INFO_DESC)]
    /// Examine the metadata essence of objects
    Info {
        /// Path to inspect
        path: String,
    },
    
    #[command(about = crate::help_text::SEARCH_DESC)]
    /// Search across all crystallized knowledge
    Search {
        /// Search query
        query: String,
        
        /// Match ALL terms (AND mode)
        #[arg(long = "all", short = 'a', conflicts_with_all = &["any", "exact"])]
        all: bool,
        
        /// Match ANY terms (OR mode - default)
        #[arg(long = "any", short = 'o', conflicts_with_all = &["all", "exact"])]
        any: bool,
        
        /// Match exact phrase
        #[arg(long = "exact", short = 'e', conflicts_with_all = &["all", "any"])]
        exact: bool,
        
        /// Limit search to paths under this prefix
        #[arg(long)]
        path: Option<String>,
        
        /// Filter by object type
        #[arg(long = "type")]
        type_filter: Option<String>,
        
        /// Filter by creation date after (YYYY-MM-DD)
        #[arg(long)]
        after: Option<String>,
        
        /// Filter by creation date before (YYYY-MM-DD)
        #[arg(long)]
        before: Option<String>,
        
        /// Filter by agent name
        #[arg(long)]
        agent: Option<String>,
        
        /// Filter by tags (can specify multiple)
        #[arg(long = "tag")]
        tags: Vec<String>,
        
        /// Maximum number of results to show
        #[arg(long, short = 'n', default_value = "20")]
        limit: Option<usize>,
    },
    
    /// Watch real-time system activity
    Watch {
        /// What to watch (rules, sessions)
        target: String,
    },
}

#[derive(Subcommand)]
pub enum DaemonAction {
    /// Start the daemon
    Start {
        /// Run in background (default: foreground)
        #[arg(short, long)]
        background: bool,
    },
    
    /// Stop the daemon
    Stop,
    
    /// Restart the daemon
    Restart,
    
    /// Show daemon logs
    Logs {
        /// Number of lines to show
        #[arg(short = 'n', long, default_value = "50")]
        lines: usize,
        
        /// Follow log output
        #[arg(short, long)]
        follow: bool,
    },
}

#[derive(Subcommand)]
pub enum MemoryAction {
    /// Search through memories
    Search {
        /// Search query
        query: String,
        
        /// Limit number of results
        #[arg(short, long, default_value = "10")]
        limit: usize,
    },
    
    /// Show a specific session
    Show {
        /// Session ID
        session_id: String,
    },
    
    /// Rename a memory/session
    Rename {
        /// Session ID to rename
        session_id: String,
        /// New name for the session
        new_name: String,
    },
}

#[derive(Subcommand)]
enum DeclareCommand {
    /// Declare that a tool should exist
    Tool {
        /// Name of the tool
        name: String,
        
        /// What the tool transforms/processes (comma-separated)
        #[arg(long)]
        transforms: Option<String>,
        
        /// Reference entities for context (file:path, p42:/commands/name, url:https://, search:"query")
        #[arg(long = "ref", action = clap::ArgAction::Append, help = "Reference other entities for context (can be used multiple times)\n\nAvailable reference types:\n• file:./path/to/file    - Local file reference\n• p42:/commands/name     - Port 42 VFS reference\n• url:https://api.docs   - Web URL reference\n• search:\"query terms\"   - Search-based reference\n\nExample: --ref file:./config.json --ref search:\"error patterns\"")]
        references: Option<Vec<String>>,
        
        /// Custom prompt to guide AI tool generation  
        #[arg(long, help = "Custom prompt to guide AI tool generation\n\nProvide specific instructions for how the tool should work.\nCombined with references to create contextually-aware tools.\n\nExample: --prompt \"Create a tool that analyzes logs and highlights errors\"")]
        prompt: Option<String>,
    },
    
    /// Declare that an artifact should exist
    Artifact {
        /// Name of the artifact
        name: String,
        
        /// Type of artifact (document, config, schema, etc.)
        #[arg(long, default_value = "document")]
        artifact_type: String,
        
        /// File type/extension
        #[arg(long, default_value = ".md")]
        file_type: String,
        
        /// Custom prompt to guide AI artifact generation
        #[arg(long, help = "Custom prompt to guide AI artifact generation\n\nProvide specific instructions for the artifact content and structure.\nWorks with references to create contextually-aware documentation.\n\nExample: --prompt \"Create API documentation with examples and error codes\"")]
        prompt: Option<String>,
    },
}

fn main() -> Result<()> {
    // Set up colored output first
    colored::control::set_override(true);
    
    // Check if this is a help request and handle it with our custom help
    if help_handler::handle_help_request() {
        return Ok(());
    }
    
    // Otherwise, let Clap parse normally
    let cli = Cli::parse();
    
    // Handle verbose flag
    if cli.verbose {
        eprintln!("{}", "🔍 Verbose mode enabled".dimmed());
    }
    
    // Determine port
    let port = cli.port.unwrap_or_else(|| {
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: main() - no explicit port, calling detect_daemon_port()");
        }
        // Use proper daemon ping to discover port
        let discovered_port = client::detect_daemon_port().unwrap_or(42);
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: main() - discovered port: {}", discovered_port);
        }
        discovered_port
    });
    
    // Determine output format
    let output_format = if cli.json {
        display::OutputFormat::Json
    } else {
        display::OutputFormat::Plain
    };
    
    // Route to command handlers
    match cli.command {
        
        Some(Commands::Daemon { action }) => {
            daemon::handle_daemon(action, port)?;
        }
        
        Some(Commands::Status { detailed }) => {
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: main() - handling Status command with port {}", port);
            }
            let mut client = client::DaemonClient::new(port);
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: main() - created new DaemonClient for Status command");
            }
            if cli.json {
                status::handle_status_with_format(&mut client, detailed, display::OutputFormat::Json)?;
            } else {
                status::handle_status(port, detailed)?;
            }
        }
        
        Some(Commands::Reality { verbose, agent }) => {
            if cli.json {
                reality::handle_reality_with_format(port, verbose, agent, display::OutputFormat::Json)?;
            } else {
                reality::handle_reality(port, verbose, agent)?;
            }
        }
        
        Some(Commands::Context { pretty, compact }) => {
            use crate::context::formatters::{ContextFormatter, JsonFormatter, PrettyFormatter, CompactFormatter};
            
            let mut client = crate::client::DaemonClient::new(port);
            let response = client.request(crate::protocol::DaemonRequest {
                request_type: "context".to_string(),
                id: format!("context-{}", std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap()
                    .as_millis()),
                payload: serde_json::json!({}),
                references: None,
                session_context: None,
                user_prompt: None,
            })?;
            
            if !response.success {
                eprintln!("❌ Failed to get context: {}", 
                    response.error.unwrap_or_else(|| "Unknown error".to_string()));
                std::process::exit(1);
            }
            
            if let Some(data) = response.data {
                // Parse into typed structure
                let context_data: crate::context::ContextData = serde_json::from_value(data)?;
                
                // Choose formatter based on flags
                let formatter: Box<dyn ContextFormatter> = if compact {
                    Box::new(CompactFormatter)
                } else if pretty {
                    Box::new(PrettyFormatter)
                } else {
                    Box::new(JsonFormatter)
                };
                
                // Format and print
                println!("{}", formatter.format(&context_data));
            }
        }
        
        Some(Commands::Possess { agent, session, references, message }) => {
            // Simple: session is explicit, message is always the args
            let message_text = if message.is_empty() { 
                None 
            } else { 
                Some(message.join(" ")) 
            };
            
            // Handle special "last" value with agent context
            let session_id = match session.as_deref() {
                Some("last") => {
                    // Query daemon for last session for this specific agent
                    let mut client = crate::client::DaemonClient::new(port);
                    match client.get_last_session(&agent) {
                        Ok(id) => {
                            eprintln!("🔄 Resuming last session for {}: {}", agent, id);
                            Some(id)
                        },
                        Err(_) => {
                            eprintln!("❌ No previous sessions found for {}", agent);
                            std::process::exit(1);
                        }
                    }
                },
                Some(id) => Some(id.to_string()),
                None => None,
            };
            
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG possess: agent={}, session={:?}, message={:?}", 
                         agent, session_id, message_text);
            }
            
            // Auto-detect output mode: show boot only for interactive mode (no message)
            let show_boot = message_text.is_none();
            commands::possess::handle_possess_with_references(port, agent, message_text, session_id, references, show_boot)?;
        }
        
        Some(Commands::Declare { command }) => {
            match command {
                DeclareCommand::Tool { name, transforms, references, prompt } => {
                    let transforms_vec = transforms.as_ref()
                        .map(|t| t.split(',').map(|s| s.trim().to_string()).collect())
                        .unwrap_or_default();
                    
                    commands::declare::handle_declare_tool(port, &name, transforms_vec, references.clone(), prompt.clone())?;
                }
                DeclareCommand::Artifact { name, artifact_type, file_type, prompt } => {
                    commands::declare::handle_declare_artifact(port, &name, &artifact_type, &file_type, prompt.clone())?;
                }
            }
        }
        
        Some(Commands::Memory { args }) => {
            // Parse memory args similar to shell
            let action = if args.is_empty() {
                None // List all
            } else if args[0] == "search" {
                if args.len() < 2 {
                    eprintln!("{}", help_text::ERR_MEMORY_SEARCH_USAGE.red());
                    std::process::exit(1);
                }
                Some(MemoryAction::Search {
                    query: args[1..].join(" "),
                    limit: 10,
                })
            } else if args[0] == "rename" {
                if args.len() < 3 {
                    eprintln!("{}", "Usage: memory rename <session_id> <new_name>".red());
                    std::process::exit(1);
                }
                Some(MemoryAction::Rename {
                    session_id: args[1].clone(),
                    new_name: args[2..].join(" "),
                })
            } else {
                // First arg is session ID
                Some(MemoryAction::Show {
                    session_id: args[0].clone(),
                })
            };
            
            if cli.json {
                memory::handle_memory_with_format(port, action, display::OutputFormat::Json)?;
            } else {
                memory::handle_memory(port, action)?;
            }
        }
        
        
        Some(Commands::Ls { path }) => {
            let mut client = client::DaemonClient::new(port);
            if cli.json {
                ls::handle_ls_with_format(&mut client, path, display::OutputFormat::Json)?;
            } else {
                ls::handle_ls(&mut client, path)?;
            }
        }
        
        Some(Commands::Cat { path }) => {
            let mut client = client::DaemonClient::new(port);
            if cli.json {
                cat::handle_cat_with_format(&mut client, path, display::OutputFormat::Json)?;
            } else {
                cat::handle_cat(&mut client, path)?;
            }
        }
        
        Some(Commands::Info { path }) => {
            let mut client = client::DaemonClient::new(port);
            if cli.json {
                info::handle_info_with_format(&mut client, path, display::OutputFormat::Json)?;
            } else {
                info::handle_info(&mut client, path)?;
            }
        }
        
        Some(Commands::Search { query, all, any, exact, path, type_filter, after, before, agent, tags, limit }) => {
            let mut client = client::DaemonClient::new(port);
            
            // Determine search mode
            let mode = if all {
                "and"
            } else if exact {
                "phrase"
            } else {
                "or"  // default, also covers explicit --any
            };
            
            if cli.json {
                search::handle_search_with_format(&mut client, query, mode, path, type_filter, after, before, agent, tags, limit, display::OutputFormat::Json)?;
            } else {
                search::handle_search(&mut client, query, mode, path, type_filter, after, before, agent, tags, limit)?;
            }
        }
        
        Some(Commands::Watch { target }) => {
            match target.as_str() {
                "rules" => {
                    commands::watch::watch_rules(port)?;
                }
                _ => {
                    eprintln!("❌ Unsupported watch target: {}. Supported: rules", target);
                    std::process::exit(1);
                }
            }
        }
        
        None => {
            // No command provided - launch Port 42 shell
            let mut shell = shell::Port42Shell::new(port);
            shell.run()?;
        }
    }
    
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use clap::CommandFactory;
    
    #[test]
    fn verify_cli() {
        // This will catch CLI parsing errors at compile time
        Cli::command().debug_assert();
    }
    
    #[test]
    fn test_help() {
        let result = Cli::try_parse_from(&["port42", "--help"]);
        assert!(result.is_err()); // --help returns an error with help message
    }
    
    #[test]
    fn test_status_command() {
        let result = Cli::try_parse_from(&["port42", "status"]);
        assert!(result.is_ok());
        
        if let Ok(cli) = result {
            match cli.command {
                Some(Commands::Status { detailed }) => {
                    assert!(!detailed);
                }
                _ => panic!("Expected Status command"),
            }
        }
    }
}