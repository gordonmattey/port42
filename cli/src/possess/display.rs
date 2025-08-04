use crate::help_text;
use crate::protocol::{CommandSpec, ArtifactSpec};
use colored::*;
use std::io::{self, Write};
use std::thread;
use std::time::Duration;

pub trait PossessDisplay {
    fn show_ai_message(&self, agent: &str, message: &str);
    fn show_command_created(&self, spec: &CommandSpec);
    fn show_artifact_created(&self, spec: &ArtifactSpec);
    fn show_session_info(&self, session_id: &str, is_new: bool);
    fn show_error(&self, error: &str);
}

pub struct SimpleDisplay;

impl SimpleDisplay {
    pub fn new() -> Self {
        SimpleDisplay
    }
}

impl PossessDisplay for SimpleDisplay {
    fn show_ai_message(&self, agent: &str, message: &str) {
        println!("\n{}", agent.bright_blue());
        println!("{}", message);
        println!();
    }
    
    fn show_command_created(&self, spec: &CommandSpec) {
        println!("{}", help_text::format_command_born(&spec.name).bright_green().bold());
        println!("{}", "Add to PATH to use:".yellow());
        println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
        println!();
    }
    
    fn show_artifact_created(&self, spec: &ArtifactSpec) {
        println!("{}", format!("✨ Artifact created: {} ({})", spec.name, spec.artifact_type).bright_cyan().bold());
        println!("{}", "View with:".yellow());
        println!("  {}", format!("port42 cat {}", spec.path).bright_white());
        println!();
    }
    
    fn show_session_info(&self, session_id: &str, is_new: bool) {
        if is_new {
            println!("{}", help_text::format_new_session(session_id).bright_cyan());
        } else {
            println!("{}", help_text::format_session_continuing(session_id).bright_cyan());
        }
    }
    
    fn show_error(&self, error: &str) {
        eprintln!("{}", error.red());
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
        print!("{}", "🤔 Thinking".dimmed());
        io::stdout().flush().unwrap();
        
        for _ in 0..3 {
            thread::sleep(Duration::from_millis(500));
            print!("{}", ".".dimmed());
            io::stdout().flush().unwrap();
        }
        
        print!("\r{}\r", " ".repeat(20));
        io::stdout().flush().unwrap();
    }
}

impl PossessDisplay for AnimatedDisplay {
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
        
        // Animated reveal
        println!("{}", "✨ Crystallizing thought into reality...".yellow().italic());
        thread::sleep(Duration::from_millis(1000));
        
        println!("{}", help_text::format_command_born(&spec.name).bright_green().bold());
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
        
        // Animated reveal
        println!("{}", "🎨 Manifesting artifact in reality...".cyan().italic());
        thread::sleep(Duration::from_millis(1000));
        
        println!("{}", format!("✨ Artifact created: {} ({})", spec.name, spec.artifact_type).bright_cyan().bold());
        println!("{}", format!("   {}", spec.description).dimmed());
        println!();
        
        thread::sleep(Duration::from_millis(500));
        println!("{}", "View with:".yellow());
        println!("  {}", format!("port42 cat {}", spec.path).bright_white());
        println!();
    }
    
    fn show_session_info(&self, session_id: &str, is_new: bool) {
        if is_new {
            println!("{}", help_text::format_new_session(session_id).bright_cyan());
        } else {
            println!("{}", help_text::format_session_continuing(session_id).bright_cyan());
        }
        thread::sleep(Duration::from_millis(300));
    }
    
    fn show_error(&self, error: &str) {
        eprintln!("{}", format!("⚡ {}", error).red());
    }
}