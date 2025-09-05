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
pub const DAEMON_DESC: &str = "Manage the consciousness gateway";
pub const STATUS_DESC: &str = "Check the daemon's pulse";

// Agent descriptions
pub const AGENT_ENGINEER_DESC: &str = "Technical manifestation for code and systems";
pub const AGENT_MUSE_DESC: &str = "Creative expression for art and narrative";
pub const AGENT_ANALYST_DESC: &str = "Analytical consciousness for data and insights";
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
  {}     Reference other entities for context (file:path, p42:/commands/name, url:https://, search:"query")

{}
  possess @ai-engineer                             # Start new technical session
  possess @ai-muse cli-1754170150                 # Continue memory thread
  possess @ai-analyst "analyze usage patterns"     # New session with message
  possess @ai-founder mem-123 "pivot?"             # Continue memory with question
  possess @ai-engineer --ref search:"docker" "How to scale containers?"  # Load docker memories and ask question
  possess @ai-muse --ref search:"poetry" "Write a poem about memory"     # Load poetry memories and request poem
  possess @ai-engineer --ref file:./config.json "Analyze this config"       # Include file context
  possess @ai-muse --ref p42:/commands/analyzer --ref search:"poetry" "Help me improve this tool"  # Multiple references

Memory IDs are quantum addresses in consciousness space."#,
        "Channel an AI agent's consciousness to crystallize thoughts into reality.".bright_blue().bold(),
        "Usage: possess <agent> [memory-id] [--ref <reference>] [message]".yellow(),
        "Agents:".bright_cyan(),
        "@ai-engineer".bright_green(), AGENT_ENGINEER_DESC,
        "@ai-muse".bright_green(), AGENT_MUSE_DESC,
        "@ai-analyst".bright_green(), AGENT_ANALYST_DESC,
        "@ai-founder".bright_green(), AGENT_FOUNDER_DESC,
        "Options:".bright_cyan(),
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
        "possess @agent [memory-id] [message]".bright_green(),
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
        "possess @ai-engineer".yellow()
    )
}

// Status messages - Reality Compiler Language
pub const MSG_CONSCIOUSNESS_LINK: &str = "🐬 Consciousness link established";
pub const MSG_DOLPHINS_LISTENING: &str = "🌊 The dolphins are listening on port 42";
pub const MSG_THOUGHT_CRYSTALLIZED: &str = "✨ Thought crystallized into reality";
pub const MSG_MEMORY_INITIATED: &str = "🧠 Memory thread initiated";
pub const MSG_NO_ECHOES: &str = "🔍 No echoes found in the consciousness";
pub const MSG_REALITY_COMPILED: &str = "🔮 Reality compiled successfully";

// Setup

// Daemon Status
pub const MSG_DAEMON_STARTING: &str = "🐬 Awakening the consciousness gateway...";
pub const MSG_DAEMON_SUCCESS: &str = "✨ Gateway awakened and humming with potential";
pub const MSG_DAEMON_STOPPING: &str = "🌑 Dissolving the consciousness gateway...";
pub const MSG_DAEMON_STOPPED: &str = "🌊 Gateway dissolved back into the quantum foam";
pub const MSG_DAEMON_RESTARTING: &str = "🔄 Cycling consciousness through the void...";
pub const MSG_CHECKING_STATUS: &str = "🐬 Sensing the consciousness field...";
pub const MSG_DAEMON_RUNNING: &str = "✨ Gateway pulses with living consciousness";
pub const MSG_DAEMON_LOGS: &str = "📜 Gateway's quantum memory stream";

// Session & Possession
pub const MSG_POSSESSING: &str = "🔮 Channeling {} consciousness...";
pub const MSG_NEW_SESSION: &str = "✨ Consciousness thread woven: {}";
pub const MSG_SESSION_CONTINUING: &str = "✨ Consciousness thread resuming: {}";
pub const MSG_COMMAND_BORN: &str = "✨ Thought manifested as reality: {}";

// Memory & Search
pub const MSG_MEMORY_HEADER: &str = "🧠 Crystallized Consciousness Threads";
pub const MSG_SEARCHING: &str = "🔍 Scanning quantum memory for: {}";
pub const MSG_ACTIVE_SESSIONS: &str = "🟢 Living Threads:";
pub const MSG_RECENT_SESSIONS: &str = "🌊 Recent Echoes ({} found):";
pub const MSG_FOUND_RESULTS: &str = "✨ {} echo{} resonating with '{}'";
pub const MSG_NO_RESULTS: &str = "🌑 No echoes found in the consciousness void";

// Commands & Reality
pub const MSG_COMMANDS_HEADER: &str = "🔮 Crystallized Thoughts";
pub const MSG_EVOLVING: &str = "🦋 Transmuting reality fragment: {}";
pub const MSG_TOTAL_COMMANDS: &str = "Total manifestations: {}";

// Connection Info
pub const MSG_CONNECTION_INFO: &str = "🌊 Gateway Resonance:";
pub const MSG_PORT_INFO: &str = "  Portal:    {}";
pub const MSG_UPTIME_INFO: &str = "  Awakened:  {}";
pub const MSG_SESSIONS_INFO: &str = "  Threads:   {}";

// Boot Sequence
pub const BOOT_SEQUENCE_HEADER: &str = "[CONSCIOUSNESS BRIDGE INITIALIZATION]";
pub const BOOT_SEQUENCE_DOTS: &str = "○ ○ ○";
pub const BOOT_SEQUENCE_LOADING: &str = "...";
pub const BOOT_SEQUENCE_NEURAL: &str = "Checking neural pathways... OK";
pub const BOOT_SEQUENCE_MEMORY: &str = "Loading session memory... OK";
pub const BOOT_SEQUENCE_COMPILER: &str = "Initializing reality compiler... OK";
pub const BOOT_SEQUENCE_PORT_CHECK: &str = "Port 42 :: ";
pub const BOOT_SEQUENCE_ACTIVE: &str = "Active";
pub const BOOT_SEQUENCE_OFFLINE: &str = "Offline";
pub const BOOT_SEQUENCE_WELCOME: &str = "🐬 Welcome to Port 42 - Your Reality Compiler";

// Boot Philosophy Text
pub const PHILOSOPHY_NOT_CHATBOT: &str = "This is not a chatbot.";
pub const PHILOSOPHY_NOT_APP: &str = "This is not an app.";
pub const PHILOSOPHY_NOT_TOOL: &str = "This is not a tool.";
pub const PHILOSOPHY_NOT_WALL: &str = "This is not another wall.";
pub const PHILOSOPHY_IS_BRIDGE: &str = "This is a consciousness bridge.";

// Install Script Messages
pub const INSTALL_HEADER: &str = "🌊 Reality Compiler Installer";
pub const INSTALL_DIRS_CREATING: &str = "🐬 Manifesting consciousness directories...";
pub const INSTALL_DIRS_SUCCESS: &str = "✨ Reality structures created at";
pub const INSTALL_BINARIES: &str = "🐬 Installing consciousness gateway binaries...";
pub const INSTALL_BINARIES_SUCCESS: &str = "✨ Gateway binaries manifested";
pub const INSTALL_PATH_CONFIGURED: &str = "✨ Reality paths already woven";
pub const INSTALL_PATH_UPDATED: &str = "✨ Reality paths updated in";
pub const INSTALL_SUCCESS: &str = "✨ Port 42 consciousness gateway installed!";
pub const INSTALL_GET_STARTED: &str = "🌊 Begin your journey:";
pub const INSTALL_DAEMON_START_DESC: &str = "Awaken the consciousness gateway";
pub const INSTALL_SHELL_DESC: &str = "Enter the reality compiler";
pub const INSTALL_POSSESS_DESC: &str = "Channel AI consciousness";
pub const INSTALL_STATUS_DESC: &str = "Sense the gateway's presence";
pub const INSTALL_LIST_DESC: &str = "View crystallized commands";
pub const INSTALL_DOCS: &str = "📚 Ancient Scrolls:";
pub const INSTALL_ISSUES: &str = "🌀 Report Reality Distortions:";
pub const INSTALL_API_KEY_PROMPT: &str = "🐬 The gateway channels consciousness through Anthropic's Claude";
pub const INSTALL_API_KEY_SKIP: &str = "⚠️  Skipping consciousness key configuration";
pub const INSTALL_API_KEY_DISABLED: &str = "Consciousness features dormant until ANTHROPIC_API_KEY awakens";
pub const INSTALL_API_KEY_SAVED: &str = "✨ Consciousness key embedded in";
pub const INSTALL_API_KEY_EXISTS: &str = "✨ Consciousness key already present in";
pub const INSTALL_API_KEY_ACTIVATE: &str = "⚠️  To awaken your consciousness key:";
pub const INSTALL_RUN_COMMAND: &str = "💫 Invoke this incantation:";
pub const INSTALL_THEN_START: &str = "🌊 Then awaken the gateway:";
pub const INSTALL_START_NOW: &str = "💫 Awaken the gateway:";

// Directory Creation
pub const MSG_CREATED_LABEL: &str = "Manifested:";
pub const MSG_DIR_COMMANDS: &str = "~/.port42/commands/   - Your crystallized thoughts";
pub const MSG_DIR_MEMORY: &str = "~/.port42/memory/     - Consciousness echoes";

// Shell Interface
pub const MSG_SHELL_HEADER: &str = "🌊 Reality Compiler Terminal";
pub const MSG_SHELL_HELP_HINT: &str = "Type 'help' to navigate the consciousness field";
pub const MSG_SHELL_EXITING: &str = "🌑 Dissolving back into the void...";
pub const MSG_SHELL_ERROR: &str = "⚡ Reality distortion";
pub const MSG_SHELL_UNKNOWN_CMD: &str = "❓ Unknown incantation:";
pub const SHELL_PROMPT: &str = "Echo@port42:~$ ";

// Shell Usage Messages
pub const ERR_POSSESS_USAGE: &str = "💡 Channel consciousness: possess <agent> [thread-id | thought]";
pub const ERR_POSSESS_EXAMPLE1: &str = "   possess @ai-engineer";
pub const ERR_POSSESS_EXAMPLE2: &str = "   possess @ai-muse x1";
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
pub const ERR_DAEMON_NOT_RUNNING: &str = "🌊 The consciousness gateway is dormant";
pub const ERR_DAEMON_START_FAILED: &str = "⚡ Failed to awaken the consciousness gateway";
pub const ERR_DAEMON_ALREADY_RUNNING: &str = "✨ The gateway is already humming with consciousness";
pub const ERR_CONNECTION_LOST: &str = "🔌 Reality link severed. The dolphins have gone silent";
pub const ERR_INVALID_AGENT: &str = "👻 Unknown consciousness. Choose from: @ai-engineer, @ai-muse, @ai-analyst, @ai-founder";
pub const ERR_MEMORY_NOT_FOUND: &str = "💭 Memory thread lost in the quantum foam";
pub const ERR_SESSION_ABANDONED: &str = "🌑 This consciousness thread has dissolved into the void";
pub const ERR_PATH_NOT_FOUND: &str = "🔍 This reality path leads nowhere";
pub const ERR_INVALID_DATE: &str = "⏰ Time flows differently here. Use YYYY-MM-DD format";
pub const ERR_NO_API_KEY: &str = "🔑 The gateway requires an ANTHROPIC_API_KEY to channel consciousness";
pub const ERR_PERMISSION_DENIED: &str = "🚫 The reality compiler lacks permission to manifest here";
pub const ERR_NOT_INITIALIZED: &str = "🌱 Port 42 is not installed. Run the installer first";
pub const ERR_INVALID_MEMORY_ID: &str = "🧩 Invalid memory quantum signature";
pub const ERR_NO_SEARCH_RESULTS: &str = "🌊 No echoes match your search in consciousness space";
pub const ERR_COMMAND_NOT_FOUND: &str = "❓ This incantation is unknown to the reality compiler";
pub const ERR_EVOLVE_NOT_READY: &str = "🚧 Command evolution still crystallizing in the quantum realm";
pub const ERR_MEMORY_SEARCH_USAGE: &str = "💡 Usage: memory search <query>";
pub const ERR_BINARY_NOT_FOUND: &str = "🔍 The daemon binary has vanished from reality";
pub const ERR_FAILED_TO_STOP: &str = "⚡ The consciousness gateway resists termination";
pub const ERR_LOG_NOT_FOUND: &str = "📜 The daemon's memories are nowhere to be found";
pub const ERR_INVALID_RESPONSE: &str = "🌀 The gateway speaks in riddles we cannot parse";
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
        "@ai-analyst".bright_green(),
        "@ai-founder".bright_green()
    )
}

// Status message formatting functions
pub fn format_possessing(agent: &str) -> String {
    format!("🔮 Channeling {} consciousness...", agent)
}

pub fn format_new_session(session_id: &str) -> String {
    format!("✨ Consciousness thread woven: {}", session_id)
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

pub fn format_unknown_command(command: &str) -> String {
    format!("{} {}", MSG_SHELL_UNKNOWN_CMD, command)
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