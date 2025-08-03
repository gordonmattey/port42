use anyhow::Result;
use colored::*;
use std::fs;
use std::path::PathBuf;
use crate::help_text::*;

pub fn handle_init(no_start: bool, force: bool) -> Result<()> {
    println!("{}", MSG_INIT_BEGIN.blue().bold());
    
    // Check if already initialized
    let home = std::env::var("HOME")?;
    let port42_dir = PathBuf::from(&home).join(".port42");
    
    if port42_dir.exists() && !force {
        println!("{}", MSG_ALREADY_INIT.green());
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
    println!("{}", MSG_CREATING_DIRS.dimmed());
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
    
    println!("{}", MSG_INIT_SUCCESS.green().bold());
    println!("\n{}", MSG_CREATED_LABEL.bright_white());
    println!("  {}", MSG_DIR_COMMANDS.dimmed());
    println!("  {}", MSG_DIR_MEMORY.dimmed());
    println!("  {}", MSG_DIR_TEMPLATES.dimmed());
    
    if !no_start {
        println!("\n{}", "Next: Start the daemon".yellow());
        println!("  {}", "sudo -E port42 daemon start".bright_white());
    }
    
    Ok(())
}