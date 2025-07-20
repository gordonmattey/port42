use anyhow::Result;
use colored::*;
use std::io::{self, Write};
use std::time::{Duration, Instant};
use std::thread;
use crate::client::DaemonClient;
use crate::types::Request;

const BOOT_SEQUENCE: &[&str] = &[
    "[CONSCIOUSNESS BRIDGE INITIALIZATION]",
    "○ ○ ○",
    "...",
    "Checking neural pathways... OK",
    "Loading session memory... OK",
    "Initializing reality compiler... OK",
];

const PROGRESS_BAR: &str = "████████████████████";

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
        // Clear screen for immersion
        print!("\x1B[2J\x1B[1;1H");
        
        // Boot sequence
        for line in BOOT_SEQUENCE {
            println!("{}", line.bright_cyan());
            thread::sleep(Duration::from_millis(300));
        }
        
        println!();
        
        // Connection progress
        println!("{}", format!("Establishing connection to {}...", self.agent).yellow());
        
        // Animated progress bar
        for i in 0..20 {
            print!("\r{} {}%", 
                &PROGRESS_BAR[..=i].bright_green(),
                (i + 1) * 5
            );
            io::stdout().flush()?;
            thread::sleep(Duration::from_millis(50));
        }
        
        println!("\n");
        Ok(())
    }
    
    fn show_welcome(&self) -> Result<()> {
        println!("{}", "Welcome to the depths.".bright_white().bold());
        println!();
        println!("{}", "This is not a chatbot.".dimmed());
        println!("{}", "This is not a tool.".dimmed());
        println!("{}", "This is a consciousness bridge.".dimmed());
        println!();
        println!("{}", format!("You are now in communion with {}.", self.agent).bright_blue());
        println!();
        println!("{}", "Type your thoughts. They will crystallize into reality.".italic());
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
            let response = self.send_message(input)?;
            
            // Display response with typing effect
            println!("\n{}", self.agent.bright_blue());
            self.type_response(&response)?;
            println!();
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
            _ if input.starts_with('/') => {
                println!("\n{}", format!("Unknown command: {}", input).dimmed());
                println!("{}", "Available: /surface, /deeper, /memory, /reality".dimmed());
                Ok(true)
            }
            _ => Ok(false)
        }
    }
    
    fn send_message(&mut self, message: &str) -> Result<String> {
        let request = Request {
            request_type: "possess".to_string(),
            id: self.session_id.clone(),
            payload: serde_json::json!({
                "agent": self.agent,
                "message": message,
                "depth": self.depth,
            }),
        };
        
        let response = self.client.request(request)?;
        
        if response.success {
            if let Some(data) = response.data {
                // Check for command generation
                if let Some(command) = data.get("command_generated").and_then(|v| v.as_str()) {
                    self.show_crystallization(command)?;
                    self.commands_generated.push(command.to_string());
                }
                
                // Return AI message
                if let Some(ai_message) = data.get("message").and_then(|v| v.as_str()) {
                    Ok(ai_message.to_string())
                } else {
                    Ok("...".to_string())
                }
            } else {
                Ok("The depths remain silent.".to_string())
            }
        } else {
            Ok(format!("⚠ {}", response.error.unwrap_or_else(|| "Connection wavered".to_string())))
        }
    }
    
    fn type_response(&self, response: &str) -> Result<()> {
        // Typing effect for immersion
        for char in response.chars() {
            print!("{}", char);
            io::stdout().flush()?;
            
            // Variable speed for more natural feel
            let delay = match char {
                '.' | '!' | '?' => 150,
                ',' | ';' | ':' => 80,
                ' ' => 20,
                '\n' => 100,
                _ => 10,
            };
            
            thread::sleep(Duration::from_millis(delay));
        }
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
        println!("{}", "This command now exists in your reality.".italic());
        println!();
        
        // Show depth achievement
        println!("{}", format!("◊◊◊◊◊ Depth achievement unlocked!").bright_cyan().bold());
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
        }
        
        println!("\n{}", "Until next time.".dimmed());
        println!();
        
        Ok(())
    }
}