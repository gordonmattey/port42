//! Reality Compiler Help Text Constants
//! 
//! Centralized help text for Port 42's reality compiler interface.
//! This module contains all help strings to ensure consistency across
//! interactive and non-interactive modes.

use colored::*;

// Main descriptions
pub const MAIN_ABOUT: &str = "Your personal AI stream router 🐬";
pub const MAIN_LONG_ABOUT: &str = r#"Port 42 transforms your terminal into a gateway for AI streams.

A reality compiler where thoughts crystallize into tools and knowledge.

Through natural conversations, AI agents help you create custom commands 
that become permanent parts of your system.

The dolphins are listening on Port 42. Will you let them in?"#;

// Command descriptions for Clap
pub const SWIM_DESC: &str = "Swim into an AI agent's stream";
pub const MEMORY_DESC: &str = "Browse the persistent memory of conversations";
pub const REALITY_DESC: &str = "View your crystallized commands";
pub const LS_DESC: &str = "List contents of the virtual filesystem";
pub const CAT_DESC: &str = "Display content from any reality path";
pub const INFO_DESC: &str = "Examine the metadata essence of objects";
pub const SEARCH_DESC: &str = "Search across all crystallized knowledge";
pub const DAEMON_DESC: &str = "Manage the gateway daemon";
pub const STATUS_DESC: &str = "Check the daemon's pulse";

// Agent descriptions
pub const AGENT_ENGINEER_DESC: &str = "Technical manifestation for code and systems";
pub const AGENT_MUSE_DESC: &str = "Creative expression for art and narrative";
pub const AGENT_ANALYST_DESC: &str = "Analytical agent for data and insights";
pub const AGENT_FOUNDER_DESC: &str = "Visionary synthesis for product and leadership";

// Command-specific help text
pub fn swim_help() -> String {
    format!(r#"{}

{}

{}
  {}  - {}
  {}  - {}
  {}  - {}
  {}  - {}

{}
  {}     Resume specific session (use 'last' for most recent)
  {}     Reference other entities for context (file:path, p42:/commands/name, url:https://, search:"query")

{}
  swim @ai-engineer "help me build a parser"           # Start new conversation
  swim @ai-engineer --session last "continue"          # Resume last session
  swim @ai-engineer --session cli-1234567890           # Resume specific session
  swim @ai-engineer --ref file:./spec.md "implement this"  # With file reference
  swim @ai-engineer --ref search:"docker" "How to scale containers?"  # With search context
  swim @ai-muse --ref search:"poetry" "Write a poem"   # Load poetry memories
  swim @ai-engineer --ref p42:/commands/analyzer --ref search:"poetry" "Help me improve this tool"  # Multiple references

Sessions persist across daemon restarts. Use 'port42 ls /memory/sessions/' to list all sessions."#,
        "Swim into an AI agent's stream to crystallize thoughts into reality.".bright_blue().bold(),
        "Usage: swim <agent> [OPTIONS] [MESSAGE...]".yellow(),
        "Agents:".bright_cyan(),
        "@ai-engineer".bright_green(), AGENT_ENGINEER_DESC,
        "@ai-muse".bright_green(), AGENT_MUSE_DESC,
        "@ai-analyst".bright_green(), AGENT_ANALYST_DESC,
        "@ai-founder".bright_green(), AGENT_FOUNDER_DESC,
        "Options:".bright_cyan(),
        "--session <ID>".bright_green(),
        "--ref <reference>".bright_green(),
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
        "Browse the persistent memory of your AI interactions.".bright_blue().bold(),
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
  {}          Agent-specific views

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
  {}          Match ANY terms (OR mode - default)
  {}          Match ALL terms (AND mode)  
  {}        Match exact phrase
  {}      Limit to specific reality branch
  {}      Filter by type (command, session, artifact)
  {}     Created after date (YYYY-MM-DD)
  {}    Created before date
  {}    Filter by agent origin
  {}        Filter by tags (can use multiple)
  {}    Maximum results (default: 20)

{}
  search "docker"                         # Find all docker echoes (OR mode)
  search "test command"                   # Items with 'test' OR 'command'
  search --all "test runner"              # Items with 'test' AND 'runner'
  search --exact "test suite"             # Exact phrase "test suite"
  search "reality" --type command         # Commands about reality
  search "" --after 2025-08-01           # Recent crystallizations
  search "ai" --agent @ai-engineer       # Technical AI discussions

Search finds connections across all crystallized knowledge."#,
        "Query the collective memory. Search transcends paths.".bright_blue().bold(),
        "Usage: search <query> [options]".yellow(),
        "Options:".bright_cyan(),
        "-o, --any".bright_green(),
        "-a, --all".bright_green(),
        "-e, --exact".bright_green(),
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
  - Agent origin

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

The daemon is the gateway that listens on Port 42.
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
  {} - Channel AI streams
    {}  - Technical manifestation
    {}  - Creative expression
    {}  - Analytical insights
    {}  - Visionary synthesis

{}
  {}                    - Browse conversation threads
  {}                   - See crystallized commands
  {}    - Explore the virtual filesystem

{}
  {}              - Run any Port 42 or system command
  {}            - Force system command (e.g., !ls for system ls)

{}: status | daemon | clear | exit | help

Type '{}' for detailed usage and examples.
Type '{}' to begin crystallizing thoughts into reality."#,
        "CRYSTALLIZE THOUGHTS:".bright_cyan(),
        "swim @agent [memory-id] [message]".bright_green(),
        "@ai-engineer".cyan(),
        "@ai-muse".cyan(),
        "@ai-analyst".cyan(),
        "@ai-founder".cyan(),
        "NAVIGATE REALITY:".bright_cyan(),
        "memory".bright_green(),
        "reality".bright_green(),
        "ls, cat, info, search".bright_green(),
        "EXECUTE COMMANDS:".bright_cyan(),
        "<command>".bright_green(),
        "!<command>".bright_green(),
        "SYSTEM".bright_cyan(),
        "help <command>".yellow(),
        "swim @ai-engineer".yellow()
    )
}

// Status messages - Reality Compiler Language
pub const MSG_DOLPHINS_LISTENING: &str = "🌊 The dolphins are listening on port 42";

// Setup

// Daemon Status
pub const MSG_DAEMON_STARTING: &str = "🐬 Awakening the gateway...";
pub const MSG_DAEMON_SUCCESS: &str = "✨ Gateway awakened and humming with potential";
pub const MSG_DAEMON_STOPPING: &str = "🌑 Dissolving the gateway...";
pub const MSG_DAEMON_STOPPED: &str = "🌊 Gateway dissolved back into the quantum foam";
pub const MSG_DAEMON_RESTARTING: &str = "🔄 Cycling the gateway through the void...";
pub const MSG_CHECKING_STATUS: &str = "🐬 Sensing the gateway's presence...";
pub const MSG_DAEMON_RUNNING: &str = "✨ Gateway pulses with living energy";
pub const MSG_DAEMON_LOGS: &str = "📜 Gateway's quantum memory stream";

// Session & Possession
pub const MSG_SESSION_CONTINUING: &str = "✨ Swimming session resuming: {}";

// Memory & Search
pub const MSG_MEMORY_HEADER: &str = "🧠 Captured Streams";
pub const MSG_ACTIVE_SESSIONS: &str = "🟢 Active Sessions:";
pub const MSG_NO_RESULTS: &str = "🌑 No matches found";

// Commands & Reality
pub const MSG_COMMANDS_HEADER: &str = "🔮 Crystallized Thoughts";

// Connection Info
pub const MSG_CONNECTION_INFO: &str = "🌊 Gateway Resonance:";

// Boot Sequence
pub const BOOT_SEQUENCE_HEADER: &str = "[CONSCIOUSNESS BRIDGE INITIALIZATION]";
pub const BOOT_SEQUENCE_DOTS: &str = "○ ○ ○";
pub const BOOT_SEQUENCE_LOADING: &str = "...";
pub const BOOT_SEQUENCE_NEURAL: &str = "Checking neural pathways... OK";
pub const BOOT_SEQUENCE_MEMORY: &str = "Loading session memory... OK";
pub const BOOT_SEQUENCE_COMPILER: &str = "Initializing reality compiler... OK";
pub const BOOT_SEQUENCE_PORT_CHECK: &str = "Port 42 :: ";
pub const BOOT_SEQUENCE_ACTIVE: &str = "Active";
pub const BOOT_SEQUENCE_WELCOME: &str = "🐬 Welcome to Port 42 - Your Reality Compiler";

// Boot Philosophy Text
pub const PHILOSOPHY_NOT_CHATBOT: &str = "This is not a chatbot.";
pub const PHILOSOPHY_NOT_APP: &str = "This is not an app.";
pub const PHILOSOPHY_NOT_TOOL: &str = "This is not a tool.";
pub const PHILOSOPHY_NOT_WALL: &str = "This is not another wall.";
pub const PHILOSOPHY_IS_BRIDGE: &str = "This is a bridge between minds.";


// Directory Creation
// (Removed unused constants MSG_CREATED_LABEL, MSG_DIR_COMMANDS, MSG_DIR_MEMORY)

// Shell Interface
pub const MSG_SHELL_HEADER: &str = "🌊 Reality Compiler Terminal";
pub const MSG_SHELL_HELP_HINT: &str = "Type 'help' for available commands";
pub const MSG_SHELL_EXITING: &str = "🌑 Dissolving back into the void...";
pub const MSG_SHELL_ERROR: &str = "⚡ Reality distortion";
pub const SHELL_PROMPT: &str = "Echo@port42:~$ ";

// Shell Usage Messages
pub const ERR_SWIM_USAGE: &str = "💡 Swim into stream: swim <agent> [session-id | message]";
pub const ERR_SWIM_EXAMPLE1: &str = "   swim @ai-engineer";
pub const ERR_SWIM_EXAMPLE2: &str = "   swim @ai-muse x1";
pub const ERR_MEMORY_SEARCH_USAGE2: &str = "💡 Scan memories: memory search <echo>";
pub const ERR_EVOLVE_USAGE: &str = "💡 Transmute reality: evolve <fragment> [vision]";
pub const ERR_DAEMON_USAGE: &str = "💡 Gateway control: daemon <awaken|dissolve|cycle|sense>";
pub const ERR_DAEMON_UNKNOWN: &str = "❓ Unknown gateway ritual";
pub const ERR_CAT_USAGE: &str = "💡 Read essence: cat <reality-path>";
pub const ERR_CAT_EXAMPLE: &str = "   cat /commands/hello-world";
pub const ERR_INFO_USAGE: &str = "💡 Inspect metadata: info <reality-path>";
pub const ERR_INFO_EXAMPLE: &str = "   info /memory/cli-1754170150";
pub const ERR_SEARCH_USAGE: &str = "💡 Find echoes: search <resonance> [filters]";
pub const ERR_SEARCH_EXAMPLE: &str = "   search docker";
pub const ERR_SEARCH_HELP: &str = "Type 'help search' for quantum filters";

// Error Messages - Reality Compiler Language
pub const ERR_DAEMON_NOT_RUNNING: &str = "🌊 The gateway is dormant";
pub const ERR_DAEMON_START_FAILED: &str = "⚡ Failed to awaken the gateway";
pub const ERR_DAEMON_ALREADY_RUNNING: &str = "✨ The gateway is already humming with energy";
pub const ERR_CONNECTION_LOST: &str = "🔌 Reality link severed. The dolphins have gone silent";
pub const ERR_SESSION_ABANDONED: &str = "🌑 This session has expired";
pub const ERR_PATH_NOT_FOUND: &str = "🔍 This reality path leads nowhere";
pub const ERR_INVALID_DATE: &str = "⏰ Time flows differently here. Use YYYY-MM-DD format";
pub const ERR_NO_API_KEY: &str = "🔑 Port42 requires an ANTHROPIC_API_KEY to connect to Claude";
pub const ERR_EVOLVE_NOT_READY: &str = "🚧 Command evolution still crystallizing in the quantum realm";
pub const ERR_MEMORY_SEARCH_USAGE: &str = "💡 Usage: memory search <query>";
pub const ERR_BINARY_NOT_FOUND: &str = "🔍 The daemon binary has vanished from reality";
pub const ERR_FAILED_TO_STOP: &str = "⚡ The gateway resists termination";
pub const ERR_LOG_NOT_FOUND: &str = "📜 The daemon's memories are nowhere to be found";
pub const ERR_INVALID_RESPONSE: &str = "🌀 The gateway speaks in riddles we cannot parse";

// Error formatting functions
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

// Status message formatting functions
pub fn format_swimming(agent: &str) -> String {
    format!("🏊 Swimming into {}'s stream...", agent)
}

pub fn format_new_session(session_id: &str) -> String {
    format!("✨ Swimming session started: {}", session_id)
}

pub fn format_session_continuing(session_id: &str) -> String {
    MSG_SESSION_CONTINUING.replace("{}", session_id)
}

pub fn format_command_born(name: &str) -> String {
    format!("✨ Thought manifested as reality: {}", name)
}

pub fn format_searching(query: &str) -> String {
    format!("🔍 Scanning quantum memory for: {}", query)
}

pub fn format_recent_sessions(count: usize) -> String {
    format!("🌊 Recent Echoes ({} found):", count)
}

pub fn format_found_results(count: u64, plural: &str, query: &str) -> String {
    format!("✨ {} echo{} resonating with '{}'", count, plural, query)
}

pub fn format_evolving(command: &str) -> String {
    format!("🦋 Transmuting reality fragment: {}", command)
}

pub fn format_total_commands(count: usize) -> String {
    format!("Total manifestations: {}", count)
}

pub fn format_port_info(port: &str) -> String {
    format!("  Portal:    {}", port)
}

pub fn format_uptime_info(uptime: &str) -> String {
    format!("  Awakened:  {}", uptime)
}

pub fn format_sessions_info(sessions: &str) -> String {
    format!("  Threads:   {}", sessions)
}

// Help utility functions
pub fn format_command_header(command: &str) -> String {
    format!("📖 {} Help", command).bright_blue().bold().to_string()
}

pub fn get_command_help(command: &str) -> Option<String> {
    match command.to_lowercase().as_str() {
        "swim" => Some(swim_help()),
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
        println!("Available commands: swim, memory, reality, ls, cat, info, search, status");
    }
}