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
                let verbose = parts.contains(&"--verbose") || parts.contains(&"-v");
                // Parse agent filter if provided
                let agent = parts.iter()
                    .position(|&p| p == "--agent" || p == "-a")
                    .and_then(|i| parts.get(i + 1))
                    .map(|&s| s.to_string());
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
                
                // Parse --ref arguments first
                let mut references = Vec::new();
                let mut remaining_parts = Vec::new();
                let mut i = 2; // Start after agent
                
                while i < parts.len() {
                    if parts[i] == "--ref" && i + 1 < parts.len() {
                        // Found --ref with a value
                        references.push(parts[i + 1].to_string());
                        i += 2; // Skip both --ref and its value
                    } else {
                        remaining_parts.push(parts[i]);
                        i += 1;
                    }
                }
                
                // Convert references to Option
                let ref_option = if references.is_empty() { None } else { Some(references) };
                
                // Parse session/message from remaining parts (after removing --ref arguments)
                let (session, message) = match remaining_parts.len() {
                    0 => (None, None), // Just agent (and possibly refs)
                    1 => {
                        // Could be memory ID or message
                        let arg = remaining_parts[0];
                        let looks_like_id = arg.len() <= 20 && 
                            !arg.contains(' ') && 
                            (arg.contains(char::is_numeric) || 
                             arg.starts_with("cli-") || 
                             arg.contains('-') ||
                             arg.contains('_'));
                        
                        if looks_like_id {
                            // Looks like a memory ID
                            (Some(arg.to_string()), None)
                        } else {
                            // It's a message
                            (None, Some(arg.to_string()))
                        }
                    }
                    _ => {
                        // 2+ remaining parts: check if first is memory ID
                        let first_arg = remaining_parts[0];
                        let looks_like_id = first_arg.len() <= 20 && 
                            !first_arg.contains(' ') && 
                            (first_arg.contains(char::is_numeric) || 
                             first_arg.starts_with("cli-") || 
                             first_arg.contains('-') ||
                             first_arg.contains('_'));
                        
                        if looks_like_id {
                            // Memory ID + message
                            (Some(first_arg.to_string()), Some(remaining_parts[1..].join(" ")))
                        } else {
                            // All message
                            (None, Some(remaining_parts.join(" ")))
                        }
                    }
                };
                
                // Show connection progress since we're entering a session
                show_connection_progress(&agent)?;
                
                // Use the reference-aware handler if we have references
                if ref_option.is_some() {
                    possess::handle_possess_with_references(self.port, agent, message, session, None, ref_option, false)?;
                } else {
                    possess::handle_possess_no_boot(self.port, agent, message, session)?;
                }
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
                // Try to execute as Port 42 command or system command
                if let Err(e) = self.execute_external_command(&parts) {
                    eprintln!("{}: {}", MSG_SHELL_ERROR.red(), e);
                }
            }
        }
        
        Ok(())
    }
    
    fn execute_external_command(&self, parts: &[&str]) -> Result<()> {
        use std::process::Command;
        
        if parts.is_empty() {
            return Ok(());
        }
        
        let command_name = parts[0];
        let args = &parts[1..];
        
        // Check for escape prefix to force system command
        let (force_system, actual_command) = if command_name.starts_with('!') {
            (true, &command_name[1..])
        } else {
            (false, command_name)
        };
        
        // If not forcing system command, check Port 42 commands first
        if !force_system {
            let port42_cmd_path = dirs::home_dir()
                .unwrap_or_else(|| PathBuf::from("."))
                .join(".port42")
                .join("commands")
                .join(actual_command);
            
            if port42_cmd_path.exists() && port42_cmd_path.is_file() {
                // Execute Port 42 command
                let mut cmd = Command::new(&port42_cmd_path);
                cmd.args(args);
                
                let status = cmd.status()?;
                
                if !status.success() {
                    if let Some(code) = status.code() {
                        return Err(anyhow::anyhow!("Command exited with code {}", code));
                    }
                }
                return Ok(());
            }
        }
        
        // Try system command
        let mut cmd = Command::new(actual_command);
        cmd.args(args);
        
        match cmd.status() {
            Ok(status) => {
                if !status.success() {
                    if let Some(code) = status.code() {
                        return Err(anyhow::anyhow!("Command exited with code {}", code));
                    }
                }
                Ok(())
            }
            Err(e) => {
                // Check if it's a "command not found" error
                if e.kind() == std::io::ErrorKind::NotFound {
                    Err(anyhow::anyhow!("Command not found: {}", actual_command))
                } else {
                    Err(anyhow::anyhow!("Failed to execute command: {}", e))
                }
            }
        }
    }
    
    fn show_help(&self) {
        println!();
        println!("{}", crate::help_text::shell_help_header());
        println!("{}", crate::help_text::shell_help_main());
        println!();
    }
}