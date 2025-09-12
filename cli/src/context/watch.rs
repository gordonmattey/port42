use super::*;
use super::formatters::{ContextFormatter, PrettyFormatter};
use crate::client::DaemonClient;
use std::io::{self, Write};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::{Duration, Instant};

/// Watch mode for live context updates
pub struct WatchMode {
    pub client: DaemonClient,
    pub refresh_rate: Duration,
    formatter: Box<dyn ContextFormatter>,
    running: Arc<AtomicBool>,
}

impl WatchMode {
    pub fn new(client: DaemonClient, refresh_rate_ms: u64) -> Self {
        WatchMode {
            client,
            refresh_rate: Duration::from_millis(refresh_rate_ms),
            formatter: Box::new(PrettyFormatter),
            running: Arc::new(AtomicBool::new(true)),
        }
    }
    
    pub fn run(&mut self) -> Result<(), Box<dyn std::error::Error>> {
        // Set up Ctrl+C handler
        let running = self.running.clone();
        ctrlc::set_handler(move || {
            running.store(false, Ordering::SeqCst);
        })?;
        
        // Clear screen and hide cursor
        // VS Code terminal doesn't handle clear screen well
        self.clear_screen();
        io::stdout().flush()?;
        
        let mut last_data: Option<ContextData> = None;
        let mut last_update = Instant::now();
        
        while self.running.load(Ordering::SeqCst) {
            // Fetch current context
            match self.client.get_context() {
                Ok(data) => {
                    // Only update if data changed or every 5 seconds (for age updates)
                    let should_update = last_data.as_ref()
                        .map(|last| !self.data_equals(last, &data))
                        .unwrap_or(true)
                        || last_update.elapsed() > Duration::from_secs(5);
                    
                    if should_update {
                        // Clear screen and move to top
                        self.clear_screen();
                        
                        // Print header with timestamp
                        let now = chrono::Local::now();
                        println!("┌─────────────────────────────────────────────┐");
                        println!("│ Port42 Context --watch      {} │", now.format("%H:%M:%S"));
                        println!("├─────────────────────────────────────────────┤");
                        
                        // Format and display context with enhanced watch formatter
                        self.format_watch_display(&data);
                        
                        // Footer
                        println!("└─────────────────────────────────────────────┘");
                        println!("Press Ctrl+C to exit | Refreshing every {}s", 
                                self.refresh_rate.as_secs());
                        
                        io::stdout().flush()?;
                        
                        last_data = Some(data);
                        last_update = Instant::now();
                    }
                }
                Err(e) => {
                    // Show error but keep running
                    self.clear_screen();
                    println!("⚠️  Error fetching context: {}", e);
                    println!("Retrying...");
                    io::stdout().flush()?;
                }
            }
            
            // Sleep with interruptible check
            let sleep_end = Instant::now() + self.refresh_rate;
            while Instant::now() < sleep_end && self.running.load(Ordering::SeqCst) {
                thread::sleep(Duration::from_millis(50));
            }
        }
        
        // Restore cursor and clear line
        print!("\x1b[?25h\n");
        println!("✨ Watch mode stopped");
        io::stdout().flush()?;
        
        Ok(())
    }
    
    /// Compare two context data structures for meaningful changes
    fn data_equals(&self, a: &ContextData, b: &ContextData) -> bool {
        // Check active session
        match (&a.active_session, &b.active_session) {
            (None, None) => {},
            (Some(s1), Some(s2)) => {
                if s1.id != s2.id || 
                   s1.message_count != s2.message_count ||
                   s1.state != s2.state ||
                   s1.tool_created != s2.tool_created {
                    return false;
                }
            },
            _ => return false,
        }
        
        // Check commands (ignore age_seconds for comparison)
        if a.recent_commands.len() != b.recent_commands.len() {
            return false;
        }
        for (cmd_a, cmd_b) in a.recent_commands.iter().zip(b.recent_commands.iter()) {
            if cmd_a.command != cmd_b.command || 
               cmd_a.exit_code != cmd_b.exit_code {
                return false;
            }
        }
        
        // Check tools
        if a.created_tools.len() != b.created_tools.len() {
            return false;
        }
        for (tool_a, tool_b) in a.created_tools.iter().zip(b.created_tools.iter()) {
            if tool_a.name != tool_b.name {
                return false;
            }
        }
        
        // Check accessed memories
        if a.accessed_memories.len() != b.accessed_memories.len() {
            return false;
        }
        for (mem_a, mem_b) in a.accessed_memories.iter().zip(b.accessed_memories.iter()) {
            if mem_a.path != mem_b.path || 
               mem_a.access_count != mem_b.access_count {
                return false;
            }
        }
        
        // Check suggestions (these might change)
        if a.suggestions.len() != b.suggestions.len() {
            return false;
        }
        
        true
    }
    
    /// Format the watch display with all context information
    fn format_watch_display(&self, data: &ContextData) {
        // Active session
        if let Some(session) = &data.active_session {
            println!("│ 🔄 Active: {} session ({} msgs)    │", 
                session.agent, session.message_count);
            if let Some(tool) = &session.tool_created {
                println!("│    Tool created: {}                  │", tool);
            }
        } else {
            println!("│ 💤 No active session                        │");
        }
        
        // Recent commands - show more for activity summary
        if !data.recent_commands.is_empty() {
            println!("│                                              │");
            println!("│ 📝 Recent Activity:                          │");
            for cmd in data.recent_commands.iter().take(5) {
                let age = if cmd.age_seconds < 60 {
                    format!("{}s ago", cmd.age_seconds)
                } else {
                    format!("{}m ago", cmd.age_seconds / 60)
                };
                println!("│ • {:<30} {:>8} │", 
                    Self::truncate(&cmd.command, 30),
                    age);
            }
        }
        
        // Created tools
        if !data.created_tools.is_empty() {
            println!("│                                              │");
            println!("│ 🛠  Created This Session:                    │");
            for tool in data.created_tools.iter().take(3) {
                println!("│ • {:<42} │", Self::truncate(&tool.name, 42));
            }
        }
        
        // Accessed memories/artifacts
        if !data.accessed_memories.is_empty() {
            println!("│                                              │");
            println!("│ 📚 Recently Accessed:                        │");
            for access in data.accessed_memories.iter().take(3) {
                let icon = match access.access_type.as_str() {
                    "created" => "✨",  // Memory/session created
                    "command" => "🔧",
                    "tool" => "⚙️",
                    "memory" | "session" => "🧠",
                    "info" | "info-command" | "info-tool" | "info-memory" => "ℹ️",
                    "browse" | "browse-commands" | "browse-tools" | "browse-memory" => "👁",
                    _ => "📄",
                };
                let times = if access.access_count > 1 {
                    format!(" ({}x)", access.access_count)
                } else {
                    String::new()
                };
                let display = access.display_name.as_ref().unwrap_or(&access.path);
                let path_display = format!("{} {}{}", icon, 
                    Self::truncate(display, 30), times);
                println!("│ {:<44} │", path_display);
            }
        }
        
        // Suggestions
        if !data.suggestions.is_empty() {
            println!("│                                              │");
            println!("│ 💡 Contextual Suggestions:                   │");
            for suggestion in data.suggestions.iter().take(3) {
                println!("│ • {:<39} [📋] │", 
                    Self::truncate(&suggestion.command, 39));
            }
        }
        
        // Fill remaining space
        println!("│                                              │");
    }
    
    /// Truncate string to fit in display
    fn truncate(s: &str, max_len: usize) -> String {
        if s.len() <= max_len {
            s.to_string()
        } else {
            format!("{}...", &s[..max_len - 3])
        }
    }
    
    /// Clear screen in a terminal-compatible way
    fn clear_screen(&self) {
        // Check for VS Code terminal or other problematic terminals
        let term_program = std::env::var("TERM_PROGRAM").unwrap_or_default();
        
        if term_program == "vscode" {
            // VS Code terminal - move cursor up and clear lines
            // This avoids accumulation of output
            print!("\x1b[H");  // Move to home position
            print!("\x1b[J");  // Clear from cursor to end of screen
            print!("\x1b[?25l"); // Hide cursor
        } else {
            // Regular terminal - use standard clear screen
            print!("\x1b[2J\x1b[1;1H\x1b[?25l");
        }
    }
}