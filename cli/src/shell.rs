use anyhow::Result;
use colored::*;
use std::io::{self, Write};
use crate::commands::*;
use crate::boot::{show_boot_sequence, show_connection_progress};

pub struct Port42Shell {
    port: u16,
    running: bool,
}

impl Port42Shell {
    pub fn new(port: u16) -> Self {
        Self {
            port,
            running: true,
        }
    }
    
    pub fn run(&mut self) -> Result<()> {
        // Show boot sequence
        show_boot_sequence(true, self.port)?;
        
        println!("{}", "Port 42 Terminal".bright_white().bold());
        println!("{}", "Type 'help' for available commands".dimmed());
        println!();
        
        // Main shell loop
        while self.running {
            // Show prompt
            print!("{}", "Echo@port42:~$ ".bright_green());
            io::stdout().flush()?;
            
            // Read input
            let mut input = String::new();
            io::stdin().read_line(&mut input)?;
            let input = input.trim();
            
            if input.is_empty() {
                continue;
            }
            
            // Parse and execute command
            self.execute_command(input)?;
        }
        
        Ok(())
    }
    
    fn execute_command(&mut self, input: &str) -> Result<()> {
        let parts: Vec<&str> = input.split_whitespace().collect();
        if parts.is_empty() {
            return Ok(());
        }
        
        match parts[0] {
            "help" => self.show_help(),
            "exit" | "quit" => {
                println!("{}", "Exiting Port 42...".dimmed());
                self.running = false;
            }
            "clear" => {
                print!("\x1B[2J\x1B[1;1H");
            }
            "status" => {
                let detailed = parts.get(1).map(|&s| s == "--detailed").unwrap_or(false);
                status::handle_status(self.port, detailed)?;
            }
            "list" => {
                let verbose = parts.contains(&"--verbose");
                let agent = None; // Could parse agent filter
                list::handle_list(self.port, verbose, agent)?;
            }
            "possess" => {
                if parts.len() < 2 {
                    println!("{}", "Usage: possess <agent> [memory-id | message]".red());
                    println!("{}", "Example: possess @claude".dimmed());
                    println!("{}", "Example: possess @ai-engineer x1".dimmed());
                    return Ok(());
                }
                
                let agent = parts[1].to_string();
                let (session, message) = match parts.len() {
                    2 => (None, None), // Just agent
                    3 => {
                        // Could be memory ID or message
                        let second_arg = parts[2];
                        let looks_like_id = second_arg.len() <= 20 && 
                            !second_arg.contains(' ') && 
                            (second_arg.contains(char::is_numeric) || 
                             second_arg.starts_with("cli-") || 
                             second_arg.contains('-') ||
                             second_arg.contains('_'));
                        
                        if looks_like_id {
                            // Looks like a memory ID
                            (Some(second_arg.to_string()), None)
                        } else {
                            // It's a message
                            (None, Some(second_arg.to_string()))
                        }
                    }
                    _ => {
                        // 4+ parts: check if second is memory ID
                        let second_arg = parts[2];
                        let looks_like_id = second_arg.len() <= 20 && 
                            !second_arg.contains(' ') && 
                            (second_arg.contains(char::is_numeric) || 
                             second_arg.starts_with("cli-") || 
                             second_arg.contains('-') ||
                             second_arg.contains('_'));
                        
                        if looks_like_id {
                            // Memory ID + message
                            (Some(second_arg.to_string()), Some(parts[3..].join(" ")))
                        } else {
                            // All message
                            (None, Some(parts[2..].join(" ")))
                        }
                    }
                };
                
                // Show connection progress since we're entering a session
                show_connection_progress(&agent)?;
                possess::handle_possess_no_boot(self.port, agent, message, session)?;
            }
            "memory" => {
                memory::handle_memory(self.port, None)?;
            }
            "evolve" => {
                if parts.len() < 2 {
                    println!("{}", "Usage: evolve <command> [message]".red());
                    return Ok(());
                }
                
                let command = parts[1].to_string();
                let message = if parts.len() > 2 {
                    Some(parts[2..].join(" "))
                } else {
                    None
                };
                
                evolve::handle_evolve(self.port, command, message)?;
            }
            "daemon" => {
                if parts.len() < 2 {
                    println!("{}", "Usage: daemon <start|stop|restart|status>".red());
                    return Ok(());
                }
                
                use crate::DaemonAction;
                let action = match parts[1] {
                    "start" => DaemonAction::Start { background: false },
                    "stop" => DaemonAction::Stop,
                    "restart" => DaemonAction::Restart,
                    "status" => {
                        // Just check status directly
                        status::handle_status(self.port, false)?;
                        return Ok(());
                    }
                    _ => {
                        println!("{}", "Unknown daemon action".red());
                        return Ok(());
                    }
                };
                
                daemon::handle_daemon(action, self.port)?;
            }
            _ => {
                println!("{}", format!("Unknown command: {}", parts[0]).red());
                println!("{}", "Type 'help' for available commands".dimmed());
            }
        }
        
        Ok(())
    }
    
    fn show_help(&self) {
        println!();
        println!("{}", "Port 42 Terminal Commands".bright_white().bold());
        println!("{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".dimmed());
        println!();
        
        println!("{}", "Core Commands:".bright_cyan());
        println!("  {} - Start AI possession session", "possess <agent> [memory-id] [message]".bright_green());
        println!("    Example: possess @claude                      (starts new session)");
        println!("    Example: possess @ai-engineer x1              (continues memory x1)");
        println!("    Example: possess @claude \"help with git\"      (new session + message)");
        println!("    Example: possess @ai-engineer x1 \"continue\"   (memory x1 + message)");
        println!();
        
        println!("  {} - Check daemon status", "status".bright_green());
        println!("  {} - List generated commands", "list".bright_green());
        println!("  {} - Browse conversation memory", "memory".bright_green());
        println!("    Use: memory list - to see all sessions");
        println!("    Use: memory show <id> - to view a session");
        println!("  {} - Evolve an existing command", "evolve <command>".bright_green());
        println!();
        
        println!("{}", "System Commands:".bright_cyan());
        println!("  {} - Manage daemon (start/stop/restart/status)", "daemon <action>".bright_green());
        println!("  {} - Show this help", "help".bright_green());
        println!("  {} - Clear screen", "clear".bright_green());
        println!("  {} - Exit Port 42", "exit".bright_green());
        println!();
        
        println!("{}", "Available Agents:".bright_cyan());
        println!("  {} - Technical implementation expert", "@ai-engineer".bright_blue());
        println!("  {} - Creative muse for ideation", "@ai-muse".bright_blue());
        println!("  {} - Growth strategist for viral developer tools", "@ai-growth".bright_blue());
        println!("  {} - Strategic founder wisdom", "@ai-founder".bright_blue());
        println!();
    }
}