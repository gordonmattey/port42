use anyhow::Result;
use colored::*;
use crate::help_text::*;

pub fn handle_evolve(_port: u16, command: String, message: Option<String>) -> Result<()> {
    println!("{}", format!("🦋 Evolving command: {}", command).blue().bold());
    
    println!("{}", ERR_EVOLVE_NOT_READY.yellow());
    println!("{}", "The evolution chamber awaits activation".dimmed());
    
    if let Some(msg) = message {
        println!("\n{}", "Changes requested:".bright_white());
        println!("  {}", msg.dimmed());
    }
    
    Ok(())
}