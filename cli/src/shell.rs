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
                    println!("{}", "Usage: possess <agent> [message]".red());
                    println!("{}", "Example: possess @claude".dimmed());
                    return Ok(());
                }
                
                let agent = parts[1].to_string();
                let message = if parts.len() > 2 {
                    Some(parts[2..].join(" "))
                } else {
                    None
                };
                
                // Show connection progress since we're entering a session
                show_connection_progress(&agent)?;
                possess::handle_possess_no_boot(self.port, agent, message, None)?;
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
        println!("  {} - Start AI possession session", "possess <agent>".bright_green());
        println!("    Example: possess @claude");
        println!("    Example: possess @ai-engineer \"create a git helper\"");
        println!();
        
        println!("  {} - Check daemon status", "status".bright_green());
        println!("  {} - List generated commands", "list".bright_green());
        println!("  {} - Browse conversation memory", "memory".bright_green());
        println!("  {} - Evolve an existing command", "evolve <command>".bright_green());
        println!();
        
        println!("{}", "System Commands:".bright_cyan());
        println!("  {} - Manage daemon (start/stop/restart/status)", "daemon <action>".bright_green());
        println!("  {} - Show this help", "help".bright_green());
        println!("  {} - Clear screen", "clear".bright_green());
        println!("  {} - Exit Port 42", "exit".bright_green());
        println!();
        
        println!("{}", "Available Agents:".bright_cyan());
        println!("  {} - Creative muse for ideation", "@ai-muse".bright_blue());
        println!("  {} - Technical implementation expert", "@ai-engineer".bright_blue());
        println!("  {} - Command assistant", "@claude".bright_blue());
        println!();
    }
}