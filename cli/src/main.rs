use clap::{Parser, Subcommand};
use colored::*;
use anyhow::Result;

mod commands;
mod client;
mod types;

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
    command: Commands,
    
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
    
    /// List generated commands
    List {
        /// Show detailed information about each command
        #[arg(short, long)]
        verbose: bool,
        
        /// Filter by agent who created the command
        #[arg(short, long)]
        agent: Option<String>,
    },
    
    /// Start AI possession session
    Possess {
        /// AI agent to possess (@ai-muse, @ai-engineer, @ai-echo)
        agent: String,
        
        /// Initial message (starts interactive mode if not provided)
        message: Option<String>,
        
        /// Session ID (generates one if not provided)
        #[arg(short, long)]
        session: Option<String>,
    },
    
    /// Browse conversation memory
    Memory {
        #[command(subcommand)]
        action: Option<MemoryAction>,
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
    /// List recent sessions
    List {
        /// Number of days to look back
        #[arg(short, long, default_value = "7")]
        days: u32,
        
        /// Filter by agent
        #[arg(short, long)]
        agent: Option<String>,
    },
    
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
        Commands::Init { no_start, force } => {
            init::handle_init(no_start, force)?;
        }
        
        Commands::Daemon { action } => {
            daemon::handle_daemon(action, port)?;
        }
        
        Commands::Status { detailed } => {
            status::handle_status(port, detailed)?;
        }
        
        Commands::List { verbose, agent } => {
            list::handle_list(port, verbose, agent)?;
        }
        
        Commands::Possess { agent, message, session } => {
            possess::handle_possess(port, agent, message, session)?;
        }
        
        Commands::Memory { action } => {
            memory::handle_memory(port, action)?;
        }
        
        Commands::Evolve { command, message } => {
            evolve::handle_evolve(port, command, message)?;
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
                Commands::Status { detailed } => {
                    assert!(!detailed);
                }
                _ => panic!("Expected Status command"),
            }
        }
    }
}