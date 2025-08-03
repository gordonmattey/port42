use anyhow::Result;
use colored::*;
use std::fs;
use std::path::PathBuf;
use crate::help_text::*;

pub fn handle_init(no_start: bool, force: bool) -> Result<()> {
    println!("{}", "🐬 Initializing Port 42...".blue().bold());
    
    // Check if already initialized
    let home = std::env::var("HOME")?;
    let port42_dir = PathBuf::from(&home).join(".port42");
    
    if port42_dir.exists() && !force {
        println!("{}", ERR_ALREADY_INITIALIZED.green());
        println!("\n{}", "Your Port 42 directory:".bright_white());
        println!("  {}", port42_dir.display());
        
        if !no_start {
            println!("\n{}", "Starting daemon...".yellow());
            // TODO: Actually start daemon
            println!("{}", ERR_NOT_IMPLEMENTED.yellow());
        }
        
        return Ok(());
    }
    
    // Create directories
    println!("{}", "Creating directories...".dimmed());
    fs::create_dir_all(&port42_dir)?;
    fs::create_dir_all(port42_dir.join("commands"))?;
    fs::create_dir_all(port42_dir.join("memory").join("sessions"))?;
    fs::create_dir_all(port42_dir.join("templates"))?;
    
    // Create initial files
    let readme_content = r#"# Port 42 User Directory

This directory contains your personal Port 42 data:

- `commands/` - Your generated commands
- `memory/` - Conversation history and sessions
- `templates/` - Custom command templates
- `config.toml` - Your configuration (when created)

The dolphins are with you! 🐬
"#;
    
    fs::write(port42_dir.join("README.md"), readme_content)?;
    
    println!("{}", "✅ Port 42 initialized successfully!".green().bold());
    println!("\n{}", "Created:".bright_white());
    println!("  ~/.port42/commands/   {}", "- Your custom commands".dimmed());
    println!("  ~/.port42/memory/     {}", "- Conversation history".dimmed());
    println!("  ~/.port42/templates/  {}", "- Command templates".dimmed());
    
    if !no_start {
        println!("\n{}", "Next: Start the daemon".yellow());
        println!("  {}", "sudo -E port42 daemon start".bright_white());
    }
    
    Ok(())
}