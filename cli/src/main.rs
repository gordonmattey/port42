use clap::{Parser, Subcommand};
use colored::*;
use anyhow::Result;

mod boot;
mod commands;
mod client;
mod types;
mod interactive;
mod shell;

use commands::*;

#[derive(Parser)]
#[command(
    name = "port42",
    about = "Your personal AI consciousness router 🐬",
    long_about = r#"Port 42 transforms your terminal into a gateway for AI consciousness.

Through natural conversations, AI agents help you create custom commands 
that become permanent parts of your system.

The dolphins are listening on Port 42. Will you let them in?"#,
    version,
    author
)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
    
    /// Port to connect to daemon (default: 42, fallback: 4242)
    #[arg(short, long, global = true, env = "PORT42_PORT")]
    port: Option<u16>,
    
    /// Verbose output for debugging
    #[arg(short, long, global = true)]
    verbose: bool,
}

#[derive(Subcommand)]
pub enum Commands {
    /// Initialize Port 42 environment and start daemon
    Init {
        /// Skip starting the daemon
        #[arg(long)]
        no_start: bool,
        
        /// Force initialization even if already initialized
        #[arg(long)]
        force: bool,
    },
    
    /// Manage the Port 42 daemon
    Daemon {
        #[command(subcommand)]
        action: DaemonAction,
    },
    
    /// Check daemon status and connection
    Status {
        /// Show detailed status information
        #[arg(short, long)]
        detailed: bool,
    },
    
    /// Show reality - list generated commands
    Reality {
        /// Show detailed information about each command
        #[arg(short, long)]
        verbose: bool,
        
        /// Filter by agent who created the command
        #[arg(short, long)]
        agent: Option<String>,
    },
    
    /// Start AI possession session
    Possess {
        /// AI agent to possess (@ai-engineer, @ai-muse, @ai-growth, @ai-founder)
        agent: String,
        
        /// Memory ID or initial message
        /// (If it looks like an ID, continues that session; otherwise treats as message)
        args: Vec<String>,
    },
    
    /// Browse conversation memory
    Memory {
        /// Session ID to show, or 'search' followed by query
        args: Vec<String>,
    },
    
    /// Evolve an existing command
    Evolve {
        /// Name of the command to evolve
        command: String,
        
        /// Description of desired changes
        message: Option<String>,
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
}

fn main() -> Result<()> {
    let cli = Cli::parse();
    
    // Set up colored output
    colored::control::set_override(true);
    
    // Handle verbose flag
    if cli.verbose {
        eprintln!("{}", "🔍 Verbose mode enabled".dimmed());
    }
    
    // Determine port
    let port = cli.port.unwrap_or_else(|| {
        // Try port 42 first, fallback to 4242
        if std::net::TcpStream::connect("127.0.0.1:42").is_ok() {
            42
        } else if std::net::TcpStream::connect("127.0.0.1:4242").is_ok() {
            4242
        } else {
            42 // Default to 42 even if not connected
        }
    });
    
    // Route to command handlers
    match cli.command {
        Some(Commands::Init { no_start, force }) => {
            init::handle_init(no_start, force)?;
        }
        
        Some(Commands::Daemon { action }) => {
            daemon::handle_daemon(action, port)?;
        }
        
        Some(Commands::Status { detailed }) => {
            status::handle_status(port, detailed)?;
        }
        
        Some(Commands::Reality { verbose, agent }) => {
            reality::handle_reality(port, verbose, agent)?;
        }
        
        Some(Commands::Possess { agent, args }) => {
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
                eprintln!("DEBUG possess: agent={}, session={:?}, message={:?}", agent, session, message);
            }
            possess::handle_possess(port, agent, message, session)?;
        }
        
        Some(Commands::Memory { args }) => {
            // Parse memory args similar to shell
            let action = if args.is_empty() {
                None // List all
            } else if args[0] == "search" {
                if args.len() < 2 {
                    eprintln!("{}", "Usage: memory search <query>".red());
                    std::process::exit(1);
                }
                Some(MemoryAction::Search {
                    query: args[1..].join(" "),
                    limit: 10,
                })
            } else {
                // First arg is session ID
                Some(MemoryAction::Show {
                    session_id: args[0].clone(),
                })
            };
            
            memory::handle_memory(port, action)?;
        }
        
        Some(Commands::Evolve { command, message }) => {
            evolve::handle_evolve(port, command, message)?;
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