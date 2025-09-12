use crate::help_text;
use crate::protocol::{CommandSpec, ArtifactSpec};
use crate::display::{StatusIndicator, ProgressIndicator};
use colored::*;
use std::io::{self, Write};
use std::thread;
use std::time::Duration;

pub trait SwimDisplay {
    fn show_ai_message(&self, agent: &str, message: &str);
    fn show_command_created(&self, spec: &CommandSpec);
    fn show_artifact_created(&self, spec: &ArtifactSpec);
    fn show_session_info(&self, session_id: &str, is_new: bool);
    fn show_session_complete(&self, session_id: &str);
    fn show_error(&self, error: &str);
}

pub struct SimpleDisplay;

impl SimpleDisplay {
    pub fn new() -> Self {
        SimpleDisplay
    }
}

impl SwimDisplay for SimpleDisplay {
    fn show_ai_message(&self, agent: &str, message: &str) {
        println!("\n{}", agent.bright_blue());
        println!("{}", message);
        println!();
    }
    
    fn show_command_created(&self, spec: &CommandSpec) {
        println!("{} {}", StatusIndicator::success(), help_text::format_command_born(&spec.name).bright_green().bold());
        println!("{}", "Add to PATH to use:".yellow());
        println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
        println!();
    }
    
    fn show_artifact_created(&self, spec: &ArtifactSpec) {
        println!("{} {}", StatusIndicator::success(), format!("Artifact created: {} ({})", spec.name, spec.artifact_type).bright_cyan().bold());
        println!("{}", "View with:".yellow());
        println!("  {}", format!("port42 cat {}", spec.path).bright_white());
        println!();
    }
    
    fn show_session_info(&self, session_id: &str, is_new: bool) {
        if is_new {
            println!("{}", format!("✨ Swimming session started: {}", session_id).bright_cyan());
        } else {
            println!("{}", help_text::format_session_continuing(session_id).bright_cyan());
        }
    }
    
    fn show_session_complete(&self, session_id: &str) {
        println!("{}", help_text::format_new_session(session_id).bright_cyan());
    }
    
    fn show_error(&self, error: &str) {
        eprintln!("{} {}", StatusIndicator::error(), error.red());
    }
}

pub struct AnimatedDisplay {
    depth: u32,
}

impl AnimatedDisplay {
    pub fn new() -> Self {
        AnimatedDisplay { depth: 0 }
    }
    
    pub fn with_depth(depth: u32) -> Self {
        AnimatedDisplay { depth }
    }
    
    fn animate_text(&self, text: &str, delay_ms: u64) {
        for ch in text.chars() {
            print!("{}", ch);
            io::stdout().flush().unwrap();
            thread::sleep(Duration::from_millis(delay_ms));
        }
        println!();
    }
    
    fn show_thinking(&self) {
        let mut progress = ProgressIndicator::new("Thinking");
        
        for _ in 0..6 {
            progress.tick();
            thread::sleep(Duration::from_millis(200));
        }
        
        print!("\r{}\r", " ".repeat(30));
        io::stdout().flush().unwrap();
    }
}

impl SwimDisplay for AnimatedDisplay {
    fn show_ai_message(&self, agent: &str, message: &str) {
        // Show thinking animation
        self.show_thinking();
        
        // Animated agent name
        println!("\n{}", agent.bright_blue());
        
        // Animate message with typing effect
        let delay = match self.depth {
            0..=5 => 15,
            6..=10 => 10,
            _ => 5,
        };
        
        self.animate_text(message, delay);
        println!();
    }
    
    fn show_command_created(&self, spec: &CommandSpec) {
        // Dramatic pause
        thread::sleep(Duration::from_millis(500));
        
        // Progress animation
        let mut progress = ProgressIndicator::new("Crystallizing swim into reality");
        for _ in 0..8 {
            progress.tick();
            thread::sleep(Duration::from_millis(150));
        }
        progress.finish(&help_text::format_command_born(&spec.name));
        
        println!("{}", format!("   {}", spec.description).dimmed());
        println!();
        
        thread::sleep(Duration::from_millis(500));
        println!("{}", "Add to PATH to use:".yellow());
        println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
        println!();
    }
    
    fn show_artifact_created(&self, spec: &ArtifactSpec) {
        // Dramatic pause
        thread::sleep(Duration::from_millis(500));
        
        // Progress animation
        let mut progress = ProgressIndicator::new("Surfacing artifact from stream");
        for _ in 0..8 {
            progress.tick();
            thread::sleep(Duration::from_millis(150));
        }
        progress.finish(&format!("Artifact created: {} ({})", spec.name, spec.artifact_type));
        
        println!("{}", format!("   {}", spec.description).dimmed());
        println!();
        
        thread::sleep(Duration::from_millis(500));
        println!("{}", "View with:".yellow());
        println!("  {}", format!("port42 cat {}", spec.path).bright_white());
        println!();
    }
    
    fn show_session_info(&self, session_id: &str, is_new: bool) {
        if is_new {
            println!("{}", format!("✨ Swimming session started: {}", session_id).bright_cyan());
        } else {
            println!("{}", help_text::format_session_continuing(session_id).bright_cyan());
        }
        thread::sleep(Duration::from_millis(300));
    }
    
    fn show_session_complete(&self, session_id: &str) {
        println!("{}", help_text::format_new_session(session_id).bright_cyan());
    }
    
    fn show_error(&self, error: &str) {
        eprintln!("{} {}", StatusIndicator::error(), error.red());
    }
}