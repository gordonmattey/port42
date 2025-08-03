use anyhow::Result;
use colored::*;
use rustyline::{DefaultEditor, error::ReadlineError};
use std::path::PathBuf;
use crate::commands::*;
use crate::boot::{show_boot_sequence, show_connection_progress};
use crate::help_text::*;

pub struct Port42Shell {
    port: u16,
    running: bool,
    editor: DefaultEditor,
    history_path: PathBuf,
}

impl Port42Shell {
    pub fn new(port: u16) -> Self {
        // Set up history file path
        let history_path = dirs::home_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join(".port42")
            .join("shell_history");
        
        // Create editor with history
        let mut editor = DefaultEditor::new().unwrap();
        
        // Load history if it exists
        if history_path.exists() {
            let _ = editor.load_history(&history_path);
        }
        
        Self {
            port,
            running: true,
            editor,
            history_path,
        }
    }
    
    pub fn run(&mut self) -> Result<()> {
        // Show boot sequence
        show_boot_sequence(true, self.port)?;
        
        println!("{}", MSG_SHELL_HEADER.bright_white().bold());
        println!("{}", MSG_SHELL_HELP_HINT.dimmed());
        println!();
        
        // Main shell loop
        while self.running {
            // Read input with rustyline
            match self.editor.readline(SHELL_PROMPT) {
                Ok(line) => {
                    let input = line.trim();
                    
                    if input.is_empty() {
                        continue;
                    }
                    
                    // Add to history
                    self.editor.add_history_entry(input)?;
                    
                    // Parse and execute command
                    if let Err(e) = self.execute_command(input) {
                        eprintln!("{}: {}", MSG_SHELL_ERROR.red(), e);
                    }
                    
                    // Save history after each command
                    let _ = self.editor.save_history(&self.history_path);
                }
                Err(ReadlineError::Interrupted) => {
                    // Ctrl-C pressed
                    println!("^C");
                    continue;
                }
                Err(ReadlineError::Eof) => {
                    // Ctrl-D pressed
                    println!();
                    println!("{}", MSG_SHELL_EXITING.dimmed());
                    break;
                }
                Err(err) => {
                    eprintln!("{}: {}", MSG_SHELL_ERROR.red(), err);
                    break;
                }
            }
        }
        
        Ok(())
    }
    
    fn execute_command(&mut self, input: &str) -> Result<()> {
        let parts: Vec<&str> = input.split_whitespace().collect();
        if parts.is_empty() {
            return Ok(());
        }
        
        match parts[0] {
            "help" => {
                if parts.len() > 1 {
                    // Show command-specific help
                    crate::help_text::show_command_help(parts[1]);
                } else {
                    // Show general help
                    self.show_help();
                }
            }
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
            "reality" => {
                let verbose = parts.contains(&"--verbose");
                let agent = None; // Could parse agent filter
                reality::handle_reality(self.port, verbose, agent)?;
            }
            "possess" => {
                if parts.len() < 2 {
                    println!("{}", ERR_POSSESS_USAGE.red());
                    println!("{}", ERR_POSSESS_EXAMPLE1.dimmed());
                    println!("{}", ERR_POSSESS_EXAMPLE2.dimmed());
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
                use crate::MemoryAction;
                
                // Parse memory arguments
                let action = if parts.len() > 1 {
                    if parts[1] == "search" {
                        // Handle search command
                        if parts.len() < 3 {
                            println!("{}", ERR_MEMORY_SEARCH_USAGE2.red());
                            return Ok(());
                        }
                        let query = parts[2..].join(" ");
                        Some(MemoryAction::Search { 
                            query,
                            limit: 10 
                        })
                    } else {
                        // Treat first arg as session ID
                        Some(MemoryAction::Show { 
                            session_id: parts[1].to_string() 
                        })
                    }
                } else {
                    // No args = list all
                    None
                };
                
                memory::handle_memory(self.port, action)?;
            }
            "evolve" => {
                if parts.len() < 2 {
                    println!("{}", ERR_EVOLVE_USAGE.red());
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
                    println!("{}", ERR_DAEMON_USAGE.red());
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
                        println!("{}", ERR_DAEMON_UNKNOWN.red());
                        return Ok(());
                    }
                };
                
                daemon::handle_daemon(action, self.port)?;
            }
            "ls" => {
                let path = parts.get(1).map(|s| s.to_string());
                let mut client = crate::client::DaemonClient::new(self.port);
                ls::handle_ls(&mut client, path)?;
            }
            "cat" => {
                if parts.len() < 2 {
                    println!("{}", ERR_CAT_USAGE.red());
                    println!("{}", ERR_CAT_EXAMPLE.dimmed());
                    return Ok(());
                }
                let mut client = crate::client::DaemonClient::new(self.port);
                cat::handle_cat(&mut client, parts[1].to_string())?;
            }
            "info" => {
                if parts.len() < 2 {
                    println!("{}", ERR_INFO_USAGE.red());
                    println!("{}", ERR_INFO_EXAMPLE.dimmed());
                    return Ok(());
                }
                let mut client = crate::client::DaemonClient::new(self.port);
                info::handle_info(&mut client, parts[1].to_string())?;
            }
            "search" => {
                if parts.len() < 2 {
                    println!("{}", ERR_SEARCH_USAGE.red());
                    println!("{}", ERR_SEARCH_EXAMPLE.dimmed());
                    println!("{}", ERR_SEARCH_HELP.dimmed());
                    return Ok(());
                }
                
                // Basic search - just query, no filters from shell yet
                let query = parts[1..].join(" ");
                let mut client = crate::client::DaemonClient::new(self.port);
                search::handle_search(
                    &mut client,
                    query,
                    None,      // path
                    None,      // type_filter
                    None,      // after
                    None,      // before
                    None,      // agent
                    vec![],    // tags
                    None,      // limit
                )?;
            }
            _ => {
                println!("{}", format_unknown_command(parts[0]).red());
                println!("{}", MSG_SHELL_HELP_HINT.dimmed());
            }
        }
        
        Ok(())
    }
    
    fn show_help(&self) {
        println!();
        println!("{}", crate::help_text::shell_help_header());
        println!("{}", crate::help_text::shell_help_main());
        println!();
    }
}