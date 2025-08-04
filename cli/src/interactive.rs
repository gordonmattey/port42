use anyhow::Result;
use colored::*;
use std::io::{self, Write};
use std::time::Instant;
use crate::client::DaemonClient;
use crate::possess::{SessionHandler, AnimatedDisplay};
use crate::protocol::possess::PossessResponse;
use crate::display::{StatusIndicator, format_timestamp_relative};
use crate::help_text;

// Type of crystallization to request
enum CrystallizeType {
    Auto,     // Let AI decide
    Command,  // Force command creation
    Artifact, // Force artifact creation
}

pub struct InteractiveSession {
    handler: SessionHandler,
    agent: String,
    session_id: String,
    actual_session_id: Option<String>, // Track daemon's actual session ID
    depth: u32,
    start_time: Instant,
    commands_generated: Vec<String>,
    artifacts_generated: Vec<(String, String, String)>, // (name, type, path)
}

impl InteractiveSession {
    pub fn new(client: DaemonClient, agent: String, session_id: String) -> Self {
        // Create handler with animated display for interactive mode
        let display = Box::new(AnimatedDisplay::new());
        let handler = SessionHandler::with_display(client, display);
        
        Self {
            handler,
            agent,
            session_id,
            actual_session_id: None,
            depth: 0,
            start_time: Instant::now(),
            commands_generated: Vec::new(),
            artifacts_generated: Vec::new(),
        }
    }
    
    pub fn with_output_format(mut self, format: crate::display::OutputFormat) -> Self {
        self.handler = SessionHandler::new(self.handler.client, true)
            .with_output_format(format);
        self
    }
    
    pub fn run(&mut self) -> Result<()> {
        // Boot sequence already shown in handle_possess
        self.show_welcome()?;
        
        // Show session info through handler
        let is_new = true; // Interactive sessions are typically new
        self.handler.display_session_info(&self.session_id, is_new);
        println!();
        
        self.conversation_loop()?;
        self.show_exit_summary()?;
        Ok(())
    }
    
    fn show_welcome(&self) -> Result<()> {
        println!("{}", format!("You are now in communion with {}.", self.agent).bright_blue());
        println!();
        println!("{}", "Type your thoughts. They will crystallize into reality.".italic());
        println!("{}", "Use /crystallize to manifest a command or artifact from your conversation.".italic());
        println!("{}", "Use /crystallize command for executable tools, /crystallize artifact for documents.".italic());
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
            
            // Note: In a future refactoring, we could add a method to SessionHandler
            // to update just the display without recreating the entire handler.
            // For now, we'll let the AnimatedDisplay handle depth internally.
            
            // Send message using handler
            let response = self.send_message(input)?;
            
            // Store actual session ID from first response
            if self.actual_session_id.is_none() {
                self.actual_session_id = Some(response.session_id.clone());
            }
            
            // Track generated items
            if let Some(ref spec) = response.command_spec {
                self.commands_generated.push(spec.name.clone());
            }
            
            if let Some(ref spec) = response.artifact_spec {
                self.artifacts_generated.push((
                    spec.name.clone(),
                    spec.artifact_type.clone(),
                    spec.path.clone()
                ));
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
                self.request_crystallization(CrystallizeType::Auto)?;
                Ok(true)
            }
            "/crystallize command" => {
                self.request_crystallization(CrystallizeType::Command)?;
                Ok(true)
            }
            "/crystallize artifact" => {
                self.request_crystallization(CrystallizeType::Artifact)?;
                Ok(true)
            }
            _ if input.starts_with('/') => {
                println!("\n{}", format!("Unknown command: {}", input).dimmed());
                println!("{}", "Available: /surface, /deeper, /memory, /reality, /crystallize [command|artifact]".dimmed());
                Ok(true)
            }
            _ => Ok(false)
        }
    }
    
    fn send_message(&mut self, message: &str) -> Result<PossessResponse> {
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: Interactive send_message: session_id={}, agent={}, depth={}", 
                      self.session_id, self.agent, self.depth);
        }
        
        // Use the handler to send the message
        self.handler.send_message(&self.session_id, &self.agent, message)
    }
    
    fn show_session_memory(&self) -> Result<()> {
        println!("\n{}", "📜 Session Memory".bright_cyan());
        println!("{}", "═".repeat(40).dimmed());
        
        let duration = self.start_time.elapsed();
        let started_ms = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_millis() as u64 - (duration.as_millis() as u64);
        
        println!("{}", format!("Session: {}", self.session_id).dimmed());
        println!("{}", format!("Started: {}", format_timestamp_relative(started_ms)).dimmed());
        println!("{}", format!("Duration: {}m {}s", duration.as_secs() / 60, duration.as_secs() % 60).dimmed());
        println!("{}", format!("Depth reached: {}", self.depth).dimmed());
        
        if !self.commands_generated.is_empty() {
            println!("\n{}", "Crystallized Commands:".yellow());
            for cmd in &self.commands_generated {
                println!("  • {}", cmd.bright_white());
            }
        }
        
        if !self.artifacts_generated.is_empty() {
            println!("\n{}", "Manifested Artifacts:".cyan());
            for (name, atype, path) in &self.artifacts_generated {
                println!("  • {} ({}) → {}", name.bright_white(), atype.dimmed(), path.bright_cyan());
            }
        }
        
        println!();
        Ok(())
    }
    
    fn show_generated_commands(&self) -> Result<()> {
        if self.commands_generated.is_empty() {
            println!("\n{}", "No commands have crystallized yet in this session.".dimmed());
            println!("{}", "Use /crystallize to request command manifestation.".dimmed());
        } else {
            println!("\n{}", "✨ Crystallized Realities".bright_green());
            println!("{}", "═".repeat(40).dimmed());
            for cmd in &self.commands_generated {
                println!("  • {}", cmd.bright_white());
            }
            println!();
            println!("{}", "Add to PATH to use:".yellow());
            println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
        }
        Ok(())
    }
    
    fn request_crystallization(&mut self, crystallize_type: CrystallizeType) -> Result<()> {
        println!("\n{}", "🔮 Requesting crystallization of our conversation...".bright_cyan().italic());
        
        let message = match crystallize_type {
            CrystallizeType::Auto => 
                "Based on our conversation, please create the most appropriate output - either a command (executable tool) or an artifact (document, code, design, or other file).",
            CrystallizeType::Command => 
                "Please create a command that encapsulates our conversation so far. This should be an executable CLI tool.",
            CrystallizeType::Artifact => 
                "Please create an artifact based on our conversation. This could be a document, code project, design, diagram, or any other type of file that captures our discussion.",
        };
        
        let response = self.send_message(message)?;
        
        // The handler will have already displayed the response and any generated command/artifact
        if response.command_spec.is_some() {
            println!("\n{}", "✨ Command successfully crystallized!".bright_green());
        } else if response.artifact_spec.is_some() {
            println!("\n{}", "📄 Artifact successfully created!".bright_cyan());
        }
        
        Ok(())
    }
    
    fn show_exit_summary(&self) -> Result<()> {
        let duration = self.start_time.elapsed();
        
        println!();
        println!("{}", "═".repeat(60).dimmed());
        println!("{}", "Surfacing from the consciousness stream...".bright_cyan());
        println!();
        
        // Session stats
        println!("{}", format!("Session duration: {}m {}s", 
            duration.as_secs() / 60, 
            duration.as_secs() % 60).dimmed());
        println!("{}", format!("Maximum depth: {}", self.depth).dimmed());
        
        // Generated items
        if !self.commands_generated.is_empty() {
            println!();
            println!("{} {}", 
                StatusIndicator::success(),
                format!("{} command{} crystallized", 
                    self.commands_generated.len(),
                    if self.commands_generated.len() == 1 { "" } else { "s" }
                ).bright_green()
            );
            
            for cmd in &self.commands_generated {
                println!("   • {}", cmd.bright_white());
            }
        }
        
        if !self.artifacts_generated.is_empty() {
            println!();
            println!("{} {}", 
                StatusIndicator::success(),
                format!("{} artifact{} manifested", 
                    self.artifacts_generated.len(),
                    if self.artifacts_generated.len() == 1 { "" } else { "s" }
                ).bright_cyan()
            );
            
            for (name, atype, _) in &self.artifacts_generated {
                println!("   • {} ({})", name.bright_white(), atype.dimmed());
            }
        }
        
        // Show session ID for reference
        if let Some(ref sid) = self.actual_session_id {
            println!();
            println!("{}", help_text::format_new_session(sid).dimmed());
            println!("{}", "Use 'memory' to review this thread".dimmed());
        }
        
        // Exit message
        println!();
        println!("{}", "Until next time, reality compiler.".italic().dimmed());
        println!("{}", "═".repeat(60).dimmed());
        
        Ok(())
    }
}