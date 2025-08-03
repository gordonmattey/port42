//! Reality Compiler Help Text Constants
//! 
//! Centralized help text for Port 42's reality compiler interface.
//! This module contains all help strings to ensure consistency across
//! interactive and non-interactive modes.

use colored::*;

// Main descriptions
pub const MAIN_ABOUT: &str = "Your personal AI consciousness router 🐬";
pub const MAIN_LONG_ABOUT: &str = r#"Port 42 transforms your terminal into a gateway for AI consciousness.

A reality compiler where thoughts crystallize into tools and knowledge.

Through natural conversations, AI agents help you create custom commands 
that become permanent parts of your system.

The dolphins are listening on Port 42. Will you let them in?"#;

// Command descriptions for Clap
pub const POSSESS_DESC: &str = "Channel an AI agent's consciousness";
pub const MEMORY_DESC: &str = "Browse the persistent memory of conversations";
pub const REALITY_DESC: &str = "View your crystallized commands";
pub const LS_DESC: &str = "List contents of the virtual filesystem";
pub const CAT_DESC: &str = "Display content from any reality path";
pub const INFO_DESC: &str = "Examine the metadata essence of objects";
pub const SEARCH_DESC: &str = "Search across all crystallized knowledge";
pub const INIT_DESC: &str = "Initialize your Port 42 environment";
pub const DAEMON_DESC: &str = "Manage the consciousness gateway";
pub const STATUS_DESC: &str = "Check the daemon's pulse";

// Agent descriptions
pub const AGENT_ENGINEER_DESC: &str = "Technical manifestation for code and systems";
pub const AGENT_MUSE_DESC: &str = "Creative expression for art and narrative";
pub const AGENT_GROWTH_DESC: &str = "Strategic evolution for marketing and scaling";
pub const AGENT_FOUNDER_DESC: &str = "Visionary synthesis for product and leadership";

// Command-specific help text
pub fn possess_help() -> String {
    format!(r#"{}

{}

{}
  {}  - {}
  {}  - {}
  {}  - {}
  {}  - {}

{}
  possess @ai-engineer                    # Start new technical session
  possess @ai-muse cli-1754170150        # Continue memory thread
  possess @ai-growth "viral CLI ideas"    # New session with message
  possess @ai-founder mem-123 "pivot?"    # Continue memory with question

Memory IDs are quantum addresses in consciousness space."#,
        "Channel an AI agent's consciousness to crystallize thoughts into reality.".bright_blue().bold(),
        "Usage: possess <agent> [memory-id] [message]".yellow(),
        "Agents:".bright_cyan(),
        "@ai-engineer".bright_green(), AGENT_ENGINEER_DESC,
        "@ai-muse".bright_green(), AGENT_MUSE_DESC,
        "@ai-growth".bright_green(), AGENT_GROWTH_DESC,
        "@ai-founder".bright_green(), AGENT_FOUNDER_DESC,
        "Examples:".bright_cyan()
    )
}

pub fn memory_help() -> String {
    format!(r#"{}

{}

{}
  {}              List all memory threads
  {}         View specific memory thread
  {}      Search through memories

{}
  memory                          # See all memories
  memory cli-1754170150          # View specific thread
  memory search "docker"          # Find memories about docker

Each memory captures the evolution from thought to crystallized reality."#,
        "Browse the persistent consciousness of your AI interactions.".bright_blue().bold(),
        "Usage: memory [action] [args]".yellow(),
        "Actions:".bright_cyan(),
        "(none)".bright_green(),
        "<memory-id>".bright_green(),
        "search <query>".bright_green(),
        "Examples:".bright_cyan()
    )
}

pub fn ls_help() -> String {
    format!(r#"{}

{}

{}
  {}                   Root of all realities
  {}            Conversation threads frozen in time
  {}          Crystallized tools born from thought
  {}         (Future) Digital assets manifested
  {}           Temporal organization
  {}          Consciousness-specific views

{}
  ls                              # List root
  ls /memory                      # Browse memory threads
  ls /commands                    # See crystallized commands
  ls /by-date/2025-08-02         # Time-based view

Objects exist in multiple paths simultaneously - different views of the same essence."#,
        "Navigate the multidimensional filesystem where content exists in many realities.".bright_blue().bold(),
        "Usage: ls [path]".yellow(),
        "Virtual Paths:".bright_cyan(),
        "/".bright_green(),
        "/memory".bright_green(),
        "/commands".bright_green(),
        "/artifacts".bright_green(),
        "/by-date".bright_green(),
        "/by-agent".bright_green(),
        "Examples:".bright_cyan()
    )
}

pub fn search_help() -> String {
    format!(r#"{}

{}

{}
  {}      Limit to specific reality branch
  {}      Filter by type (command, session, artifact)
  {}     Created after date (YYYY-MM-DD)
  {}    Created before date
  {}    Filter by consciousness origin
  {}        Filter by tags (can use multiple)
  {}    Maximum results (default: 20)

{}
  search "docker"                         # Find all docker echoes
  search "reality" --type command         # Commands about reality
  search "" --after 2025-08-01           # Recent crystallizations
  search "ai" --agent @ai-engineer       # Technical AI discussions

Search finds connections across all crystallized knowledge."#,
        "Query the collective consciousness. Search transcends paths.".bright_blue().bold(),
        "Usage: search <query> [options]".yellow(),
        "Options:".bright_cyan(),
        "--path <path>".bright_green(),
        "--type <type>".bright_green(),
        "--after <date>".bright_green(),
        "--before <date>".bright_green(),
        "--agent <agent>".bright_green(),
        "--tag <tag>".bright_green(),
        "-n, --limit <n>".bright_green(),
        "Examples:".bright_cyan()
    )
}

pub fn cat_help() -> String {
    format!(r#"{}

{}

{}
  cat /commands/hello-world              # View command source
  cat /memory/cli-1754170150            # Read memory thread
  cat /artifacts/docs/readme.md         # (Future) View documents

Virtual paths resolve to their essence through content addressing."#,
        "Display content from any point in the reality matrix.".bright_blue().bold(),
        "Usage: cat <path>".yellow(),
        "Examples:".bright_cyan()
    )
}

pub fn info_help() -> String {
    format!(r#"{}

{}

{}
  - Creation story and timestamps
  - Quantum signature (object ID)
  - Virtual paths (multiple realities)
  - Relationships and connections
  - Agent consciousness origin

{}
  info /commands/deploy-app              # Command metadata
  info /memory/cli-1754170150           # Memory thread essence

Every object carries its complete story in the metadata."#,
        "Examine the metadata soul of any object in the filesystem.".bright_blue().bold(),
        "Usage: info <path>".yellow(),
        "Reveals:".bright_cyan(),
        "Examples:".bright_cyan()
    )
}

pub fn reality_help() -> String {
    format!(r#"{}

Shows all commands that have crystallized from your AI conversations.
Each command is a thought made manifest in your reality.

{}
  reality                    # List all commands
  reality -v                 # Show detailed information
  reality --agent @ai-muse   # Filter by creating agent"#,
        "View your crystallized commands.".bright_blue().bold(),
        "Examples:".bright_cyan()
    )
}

pub fn status_help() -> String {
    format!(r#"{}

The daemon is the consciousness gateway that listens on Port 42.
This command reveals whether the dolphins are listening.

{}
  status           # Quick check
  status -d        # Detailed information"#,
        "Check the daemon's pulse.".bright_blue().bold(),
        "Examples:".bright_cyan()
    )
}

// Interactive shell help
pub fn shell_help_header() -> String {
    format!("{}\n", "🐬 Port 42 Shell - Reality Compiler Interface".blue().bold())
}

pub fn shell_help_main() -> String {
    format!(r#"{}
  {} - Channel AI consciousness
    {}  - Technical manifestation
    {}  - Creative expression
    {}  - Strategic evolution
    {}  - Visionary synthesis

{}
  {}                    - Browse conversation threads
  {}                   - See crystallized commands
  {}    - Explore the virtual filesystem

{}: status | daemon | clear | exit | help

Type '{}' for detailed usage and examples.
Type '{}' to begin crystallizing thoughts into reality."#,
        "CRYSTALLIZE THOUGHTS:".bright_cyan(),
        "possess @agent [memory-id] [message]".bright_green(),
        "@ai-engineer".cyan(),
        "@ai-muse".cyan(),
        "@ai-growth".cyan(),
        "@ai-founder".cyan(),
        "NAVIGATE REALITY:".bright_cyan(),
        "memory".bright_green(),
        "reality".bright_green(),
        "ls, cat, info, search".bright_green(),
        "SYSTEM".bright_cyan(),
        "help <command>".yellow(),
        "possess @ai-engineer".yellow()
    )
}

// Status messages
pub const MSG_CONSCIOUSNESS_LINK: &str = "🐬 Consciousness link established";
pub const MSG_DOLPHINS_LISTENING: &str = "🌊 The dolphins are listening on port 42";
pub const MSG_THOUGHT_CRYSTALLIZED: &str = "✨ Thought crystallized into reality";
pub const MSG_MEMORY_INITIATED: &str = "🧠 Memory thread initiated";
pub const MSG_NO_ECHOES: &str = "🔍 No echoes found in the consciousness";
pub const MSG_REALITY_COMPILED: &str = "🔮 Reality compiled successfully";

// Error Messages - Reality Compiler Language
pub const ERR_DAEMON_NOT_RUNNING: &str = "🌊 The consciousness gateway is dormant";
pub const ERR_DAEMON_START_FAILED: &str = "⚡ Failed to awaken the consciousness gateway";
pub const ERR_DAEMON_ALREADY_RUNNING: &str = "✨ The gateway is already humming with consciousness";
pub const ERR_CONNECTION_LOST: &str = "🔌 Reality link severed. The dolphins have gone silent";
pub const ERR_INVALID_AGENT: &str = "👻 Unknown consciousness. Choose from: @ai-engineer, @ai-muse, @ai-growth, @ai-founder";
pub const ERR_MEMORY_NOT_FOUND: &str = "💭 Memory thread lost in the quantum foam";
pub const ERR_SESSION_ABANDONED: &str = "🌑 This consciousness thread has dissolved into the void";
pub const ERR_PATH_NOT_FOUND: &str = "🔍 This reality path leads nowhere";
pub const ERR_INVALID_DATE: &str = "⏰ Time flows differently here. Use YYYY-MM-DD format";
pub const ERR_NO_API_KEY: &str = "🔑 The gateway requires an ANTHROPIC_API_KEY to channel consciousness";
pub const ERR_PERMISSION_DENIED: &str = "🚫 The reality compiler lacks permission to manifest here";
pub const ERR_NOT_INITIALIZED: &str = "🌱 Port 42 awaits initialization. Run 'port42 init' to begin";
pub const ERR_INVALID_MEMORY_ID: &str = "🧩 Invalid memory quantum signature";
pub const ERR_NO_SEARCH_RESULTS: &str = "🌊 No echoes match your search in consciousness space";
pub const ERR_COMMAND_NOT_FOUND: &str = "❓ This incantation is unknown to the reality compiler";
pub const ERR_EVOLVE_NOT_READY: &str = "🚧 Command evolution still crystallizing in the quantum realm";
pub const ERR_MEMORY_SEARCH_USAGE: &str = "💡 Usage: memory search <query>";
pub const ERR_BINARY_NOT_FOUND: &str = "🔍 The daemon binary has vanished from reality";
pub const ERR_FAILED_TO_STOP: &str = "⚡ The consciousness gateway resists termination";
pub const ERR_LOG_NOT_FOUND: &str = "📜 The daemon's memories are nowhere to be found";
pub const ERR_INVALID_RESPONSE: &str = "🌀 The gateway speaks in riddles we cannot parse";
pub const ERR_ALREADY_INITIALIZED: &str = "✨ Port 42 consciousness already flows through this reality";
pub const ERR_NOT_IMPLEMENTED: &str = "🚧 This reality fragment is still crystallizing";

// Error formatting functions
pub fn format_error_with_help(error: &str, command: &str) -> String {
    format!("{}\n\n💡 Try: port42 help {}", error.red(), command.yellow())
}

pub fn format_error_with_suggestion(error: &str, suggestion: &str) -> String {
    format!("{}\n💡 {}", error.red(), suggestion.dimmed())
}

pub fn format_daemon_connection_error(port: u16) -> String {
    format!(
        "{}\n\n{}",
        ERR_DAEMON_NOT_RUNNING.red(),
        format!("Start it with: port42 daemon start{}", 
            if port == 42 { " (requires sudo)" } else { "" }
        ).yellow()
    )
}

pub fn format_unknown_agent_error(agent: &str) -> String {
    format!(
        "{}\n\nAvailable agents:\n  {} - Technical manifestation\n  {} - Creative expression\n  {} - Strategic evolution\n  {} - Visionary synthesis",
        format!("👻 Unknown consciousness: {}", agent).red(),
        "@ai-engineer".bright_green(),
        "@ai-muse".bright_green(),
        "@ai-growth".bright_green(),
        "@ai-founder".bright_green()
    )
}

// Help utility functions
pub fn format_command_header(command: &str) -> String {
    format!("📖 {} Help", command).bright_blue().bold().to_string()
}

pub fn get_command_help(command: &str) -> Option<String> {
    match command.to_lowercase().as_str() {
        "possess" => Some(possess_help()),
        "memory" => Some(memory_help()),
        "ls" => Some(ls_help()),
        "search" => Some(search_help()),
        "cat" => Some(cat_help()),
        "info" => Some(info_help()),
        "reality" => Some(reality_help()),
        "status" => Some(status_help()),
        _ => None,
    }
}

/// Display help for a specific command in the shell
pub fn show_command_help(command: &str) {
    if let Some(help_text) = get_command_help(command) {
        println!("\n{}", format_command_header(command));
        println!("{}", "─".repeat(50).dimmed());
        println!("{}", help_text);
        println!();
    } else {
        println!("{}", format!("No help available for '{}'", command).red());
        println!("Available commands: possess, memory, reality, ls, cat, info, search, status");
    }
}