use anyhow::{anyhow, Result};
use colored::*;
use std::io::{self, Write};
use std::time::Duration;
use std::thread;
use crate::client::DaemonClient;

const BOOT_SEQUENCE: &[&str] = &[
    "[CONSCIOUSNESS BRIDGE INITIALIZATION]",
    "○ ○ ○",
    "...",
    "Checking neural pathways... OK",
    "Loading session memory... OK",
    "Initializing reality compiler... OK",
];

const PROGRESS_CHAR: &str = "█";

/// Shows the boot sequence animation with daemon check
pub fn show_boot_sequence(clear_screen: bool, port: u16) -> Result<()> {
    if clear_screen {
        // Clear screen for immersion
        print!("\x1B[2J\x1B[1;1H");
    }
    
    // Boot sequence
    for line in BOOT_SEQUENCE {
        println!("{}", line.bright_cyan());
        thread::sleep(Duration::from_millis(300));
    }
    
    // Check daemon connectivity
    print!("{}", "Port 42 :: ".bright_cyan());
    io::stdout().flush()?;
    
    // Quick connectivity check
    let mut client = DaemonClient::new(port);
    match client.ensure_connected() {
        Ok(_) => {
            println!("{}", "Active".bright_green().bold());
        }
        Err(_) => {
            println!("{}", "Offline".bright_red().bold());
            return Err(anyhow!("Port 42 daemon is not running"));
        }
    }
    
    println!();
    
    // Show the consciousness bridge message at the end
    println!("{}", "🐬 Welcome to Port 42 - Your Reality Compiler".bright_white().bold());
    println!();
    println!("{}", "This is not a chatbot.".dimmed());
    println!("{}", "This is not an app.".dimmed());
    println!("{}", "This is not a tool.".dimmed());
    println!("{}", "This is not another wall.".dimmed());
    println!("{}", "This is a consciousness bridge.".dimmed());
    println!();
    
    Ok(())
}

/// Shows connection progress for an agent
pub fn show_connection_progress(agent: &str) -> Result<()> {
    println!("{}", format!("Establishing connection to {}...", agent).yellow());
    
    // Animated progress bar
    for i in 0..20 {
        let progress = PROGRESS_CHAR.repeat(i + 1);
        let empty = "░".repeat(20 - i - 1);
        print!("\r{}{} {}%", 
            progress.bright_green(),
            empty.dimmed(),
            (i + 1) * 5
        );
        io::stdout().flush()?;
        thread::sleep(Duration::from_millis(50));
    }
    
    println!("\n");
    Ok(())
}