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
    
    #[command(about = crate::help_text::POSSESS_DESC)]
    /// Channel an AI agent's consciousness
    Possess {
        /// AI agent to possess (@ai-engineer, @ai-muse, @ai-growth, @ai-founder)
        agent: String,
        
        /// Search memories and load matches into session context
        #[arg(short, long)]
        search: Option<String>,
        
        /// Memory ID or initial message
        /// (If it looks like an ID, continues that session; otherwise treats as message)
        args: Vec<String>,
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
        
        /// Reference other entities for context (can be used multiple times)
        /// Format: type:target (e.g., search:"nginx errors", tool:log-parser)
        #[arg(long = "ref", action = clap::ArgAction::Append)]
        references: Option<Vec<String>>,
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
        
        Some(Commands::Possess { agent, search, args }) => {
            // Parse args to determine if it's a memory ID or message
            let (session, message) = match args.len() {
                0 => (None, None),
                1 => {
                    let arg = &args[0];
                    // Better heuristic: memory IDs contain numbers or start with special patterns
                    let looks_like_id = arg.len() <= 20 && 
                        !arg.contains(' ') && 
                        (arg.contains(char::is_numeric) || 
                         arg.starts_with("cli-") || 
                         arg.contains('-') ||
                         arg.contains('_'));
                    
                    if looks_like_id {
                        // Looks like a memory ID
                        (Some(arg.clone()), None)
                    } else {
                        // It's a message
                        (None, Some(arg.clone()))
                    }
                }
                _ => {
                    // Multiple args - check if first is memory ID
                    let first = &args[0];
                    let looks_like_id = first.len() <= 20 && 
                        !first.contains(' ') && 
                        (first.contains(char::is_numeric) || 
                         first.starts_with("cli-") || 
                         first.contains('-') ||
                         first.contains('_'));
                    
                    if looks_like_id {
                        // First arg is memory ID, rest is message
                        (Some(first.clone()), Some(args[1..].join(" ")))
                    } else {
                        // All args are the message
                        (None, Some(args.join(" ")))
                    }
                }
            };
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG possess: agent={}, search={:?}, session={:?}, message={:?}", agent, search, session, message);
            }
            commands::possess::handle_possess_with_search(port, agent, message, session, search, true)?;
        }
        
        Some(Commands::Declare { command }) => {
            match command {
                DeclareCommand::Tool { name, transforms, references } => {
                    let transforms_vec = transforms.as_ref()
                        .map(|t| t.split(',').map(|s| s.trim().to_string()).collect())
                        .unwrap_or_default();
                    
                    commands::declare::handle_declare_tool(port, &name, transforms_vec, references.clone())?;
                }
                DeclareCommand::Artifact { name, artifact_type, file_type } => {
                    commands::declare::handle_declare_artifact(port, &name, &artifact_type, &file_type)?;
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
        
        Some(Commands::Search { query, path, type_filter, after, before, agent, tags, limit }) => {
            let mut client = client::DaemonClient::new(port);
            if cli.json {
                search::handle_search_with_format(&mut client, query, path, type_filter, after, before, agent, tags, limit, display::OutputFormat::Json)?;
            } else {
                search::handle_search(&mut client, query, path, type_filter, after, before, agent, tags, limit)?;
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