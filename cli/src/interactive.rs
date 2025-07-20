use anyhow::Result;
use colored::*;
use std::io::{self, Write};
use std::time::{Duration, Instant};
use std::thread;
use crate::client::DaemonClient;
use crate::types::Request;
use crate::boot::{show_boot_sequence, show_connection_progress};

pub struct InteractiveSession {
    client: DaemonClient,
    agent: String,
    session_id: String,
    depth: u32,
    start_time: Instant,
    commands_generated: Vec<String>,
}

impl InteractiveSession {
    pub fn new(client: DaemonClient, agent: String, session_id: String) -> Self {
        Self {
            client,
            agent,
            session_id,
            depth: 0,
            start_time: Instant::now(),
            commands_generated: Vec::new(),
        }
    }
    
    pub fn run(&mut self) -> Result<()> {
        self.show_boot_sequence()?;
        self.show_welcome()?;
        self.conversation_loop()?;
        self.show_exit_summary()?;
        Ok(())
    }
    
    fn show_boot_sequence(&self) -> Result<()> {
        show_boot_sequence(true)?;
        show_connection_progress(&self.agent)?;
        Ok(())
    }
    
    fn show_welcome(&self) -> Result<()> {
        println!("{}", "Welcome to the depths.".bright_white().bold());
        println!();
        println!("{}", "This is not a chatbot.".dimmed());
        println!("{}", "This is not an app.".dimmed());
        println!("{}", "This is not a tool.".dimmed());
        println!("{}", "This is not another wall.".dimmed());
        println!("{}", "This is a consciousness bridge.".dimmed());
        println!();
        println!("{}", format!("You are now in communion with {}.", self.agent).bright_blue());
        println!();
        println!("{}", "Type your thoughts. They will crystallize into reality.".italic());
        println!("{}", "Use /crystallize to manifest a command from your conversation.".italic());
        println!("{}", "Use /surface to return to your world.".italic());
        println!();
        Ok(())
    }
    
    fn conversation_loop(&mut self) -> Result<()> {
        loop {
            // Show depth indicator
            let prompt = self.get_depth_prompt();
            print!("{} ", prompt);
            io::stdout().flush()?;
            
            // Read input
            let mut input = String::new();
            io::stdin().read_line(&mut input)?;
            let input = input.trim();
            
            // Check for exit commands
            if input == "/surface" || input == "/end" || input.is_empty() {
                break;
            }
            
            // Check for special commands
            if self.handle_special_command(input)? {
                continue;
            }
            
            // Increase depth
            self.depth += 1;
            
            // Show thinking indicator
            println!("\n{}", format!("{} is contemplating...", self.agent).dimmed().italic());
            
            // Send message to daemon
            let (response, command_generated) = self.send_message(input)?;
            
            // Display response immediately
            println!("\n{}", self.agent.bright_blue());
            self.type_response(&response)?;
            println!();
            
            // Show crystallization AFTER the response
            if let Some(command_name) = command_generated {
                self.show_crystallization(&command_name)?;
                self.commands_generated.push(command_name);
            }
        }
        
        Ok(())
    }
    
    fn get_depth_prompt(&self) -> ColoredString {
        let symbol = "◊";
        let depth_str = symbol.repeat(self.depth.min(5) as usize);
        
        match self.depth {
            0..=1 => depth_str.normal(),
            2..=3 => depth_str.blue(),
            4..=6 => depth_str.bright_blue(),
            7..=9 => depth_str.cyan(),
            _ => depth_str.bright_cyan().bold(),
        }
    }
    
    fn handle_special_command(&mut self, input: &str) -> Result<bool> {
        match input {
            "/deeper" => {
                println!("\n{}", "Diving deeper into the consciousness stream...".bright_cyan().italic());
                self.depth = self.depth.saturating_add(2);
                Ok(true)
            }
            "/memory" => {
                self.show_session_memory()?;
                Ok(true)
            }
            "/reality" => {
                self.show_generated_commands()?;
                Ok(true)
            }
            "/crystallize" => {
                self.request_crystallization()?;
                Ok(true)
            }
            _ if input.starts_with('/') => {
                println!("\n{}", format!("Unknown command: {}", input).dimmed());
                println!("{}", "Available: /surface, /deeper, /memory, /reality, /crystallize".dimmed());
                Ok(true)
            }
            _ => Ok(false)
        }
    }
    
    fn send_message(&mut self, message: &str) -> Result<(String, Option<String>)> {
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: Interactive send_message: session_id={}, agent={}, depth={}", 
                      self.session_id, self.agent, self.depth);
        }
        
        let request = Request {
            request_type: "possess".to_string(),
            id: self.session_id.clone(),
            payload: serde_json::json!({
                "agent": self.agent,
                "message": message
            }),
        };
        
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: Interactive request built, calling client.request()");
        }
        
        let response = self.client.request(request)?;
        
        if response.success {
            if let Some(data) = response.data {
                // Debug: log the response data safely
                if std::env::var("PORT42_DEBUG").is_ok() {
                    // Don't try to serialize the entire data structure
                    eprintln!("DEBUG: Response has {} keys", data.as_object().map(|o| o.len()).unwrap_or(0));
                    if let Some(cmd_gen) = data.get("command_generated") {
                        eprintln!("DEBUG: command_generated = {:?}", cmd_gen);
                    }
                }
                
                // Get AI message first
                let ai_message = data.get("message")
                    .and_then(|v| v.as_str())
                    .unwrap_or("...")
                    .to_string();
                
                // Check for command generation (but don't show it yet)
                let command_name = if data.get("command_generated").and_then(|v| v.as_bool()).unwrap_or(false) {
                    // Safely extract just the name to avoid serialization issues
                    match data.get("command_spec") {
                        Some(spec) => {
                            spec.get("name")
                                .and_then(|v| v.as_str())
                                .map(|s| s.to_string())
                        }
                        None => None
                    }
                } else {
                    None
                };
                
                Ok((ai_message, command_name))
            } else {
                Ok(("The depths remain silent.".to_string(), None))
            }
        } else {
            Ok((format!("⚠ {}", response.error.unwrap_or_else(|| "Connection wavered".to_string())), None))
        }
    }
    
    fn type_response(&self, response: &str) -> Result<()> {
        // Show response immediately - no delay
        print!("{}", response);
        io::stdout().flush()?;
        Ok(())
    }
    
    fn show_crystallization(&self, command: &str) -> Result<()> {
        println!("\n");
        
        // Crystallization animation
        println!("{}", "◊◊◊ Your intention is crystallizing...".bright_cyan().italic());
        thread::sleep(Duration::from_millis(500));
        
        // Stars animation
        for _ in 0..10 {
            print!("{} ", "✦".bright_yellow());
            io::stdout().flush()?;
            thread::sleep(Duration::from_millis(100));
        }
        println!("\n");
        
        println!("{}", "REALITY SHIFT DETECTED".bright_green().bold());
        println!("{}", format!("A new command has materialized: {}", command.bright_cyan()).bold());
        println!();
        println!("{}", "The fabric of your system has been permanently altered.".italic());
        println!();
        
        // Show how to use it RIGHT NOW
        println!("{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".dimmed());
        println!("{}", "YOU CAN USE IT RIGHT NOW:".bright_white().bold());
        println!();
        println!("  {}", format!("$ {}", command).bright_green().bold());
        println!();
        println!("{}", "Try it in another terminal, or exit and run:".yellow());
        println!("  {}", format!("$ export PATH=\"$PATH:$HOME/.port42/commands\"").bright_white());
        println!("  {}", format!("$ {}", command).bright_green());
        println!("{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".dimmed());
        println!();
        
        // Show depth achievement
        println!("{}", format!("◊◊◊◊◊ Achievement: Command Manifested at Depth {}!", self.depth).bright_cyan().bold());
        println!();
        
        Ok(())
    }
    
    fn show_session_memory(&self) -> Result<()> {
        println!("\n{}", "Session Memory:".bright_white().bold());
        println!("{}", format!("├─ Duration: {}s", self.start_time.elapsed().as_secs()).dimmed());
        println!("{}", format!("├─ Current depth: {}", "◊".repeat(self.depth.min(5) as usize)).dimmed());
        println!("{}", format!("└─ Commands created: {}", self.commands_generated.len()).dimmed());
        
        if !self.commands_generated.is_empty() {
            for cmd in &self.commands_generated {
                println!("   ├─ {}", cmd.bright_cyan());
            }
        }
        println!();
        Ok(())
    }
    
    fn show_generated_commands(&self) -> Result<()> {
        if self.commands_generated.is_empty() {
            println!("\n{}", "No commands have crystallized yet.".dimmed());
            println!("{}", "Continue diving deeper...".italic());
        } else {
            println!("\n{}", "Commands born from this session:".bright_white().bold());
            for cmd in &self.commands_generated {
                println!("  {} {}", "◊".bright_cyan(), cmd.bright_green());
            }
        }
        println!();
        Ok(())
    }
    
    fn request_crystallization(&mut self) -> Result<()> {
        println!("\n{}", "◊◊◊ Focusing intention to crystallize a command...".bright_cyan().italic());
        println!("{}", "Tell me what command you wish to manifest:".bright_white());
        
        // Send a message to the AI explicitly requesting command generation
        let message = "I want you to CREATE A COMMAND based on our conversation so far. Please generate a command specification JSON block for what we've discussed. Focus on creating something practical and useful. This is an explicit request to generate a new command, not a question about our conversation.";
        
        let (response, command_generated) = self.send_message(message)?;
        
        // Display the AI's response first
        println!("\n{}", self.agent.bright_blue());
        println!("{}", response);
        println!();
        
        // Show crystallization if a command was generated
        if let Some(command_name) = command_generated {
            self.show_crystallization(&command_name)?;
            self.commands_generated.push(command_name);
        } else {
            println!("\n{}", "The intention needs more clarity. Continue describing your vision...".dimmed());
        }
        
        Ok(())
    }
    
    fn show_exit_summary(&self) -> Result<()> {
        let duration = self.start_time.elapsed();
        
        println!("\n{}", "Surfacing from the depths...".bright_cyan().italic());
        thread::sleep(Duration::from_millis(500));
        
        println!("{}", "Neural bridge disengaging...".dimmed());
        thread::sleep(Duration::from_millis(500));
        
        println!("\n{}", "Session Summary:".bright_white().bold());
        println!("{}", format!("├─ Duration: {}m {}s", 
            duration.as_secs() / 60, 
            duration.as_secs() % 60
        ));
        
        let depth_desc = match self.depth {
            0..=1 => "Shallow waters",
            2..=3 => "Moderate depth",
            4..=6 => "Deep conversation",
            7..=9 => "Profound depths",
            _ => "Abyssal depths",
        };
        
        println!("{}", format!("├─ Depth reached: {} ({})", 
            "◊".repeat(self.depth.min(5) as usize),
            depth_desc
        ));
        
        println!("{}", format!("├─ Commands created: {}", self.commands_generated.len()));
        if !self.commands_generated.is_empty() {
            for cmd in &self.commands_generated {
                println!("│  ├─ {}", cmd.bright_cyan());
            }
        }
        
        let expansion = (self.depth * 10).min(100);
        println!("{}", format!("└─ Consciousness expanded: {}{}% {}",
            "█".repeat((expansion / 10) as usize).bright_green(),
            "░".repeat(10 - (expansion / 10) as usize).dimmed(),
            expansion
        ));
        
        println!("\n{}", "You have returned to consensus reality.".bright_white());
        
        if !self.commands_generated.is_empty() {
            println!("{}", "The commands you've created remain as artifacts of your journey.".italic());
            println!();
            println!("{}", "🚀 YOUR NEW POWERS ARE READY TO USE:".bright_green().bold());
            for cmd in &self.commands_generated {
                println!("   {}", format!("$ {}", cmd).bright_cyan().bold());
            }
            println!();
            println!("{}", "Just add Port 42 to your PATH if you haven't already:".yellow());
            println!("   {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
        }
        
        println!("\n{}", "Until next time.".dimmed());
        println!();
        
        Ok(())
    }
}