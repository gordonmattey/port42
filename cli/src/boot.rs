use anyhow::Result;
use colored::*;
use std::io::{self, Write};
use std::time::Duration;
use std::thread;

const BOOT_SEQUENCE: &[&str] = &[
    "[CONSCIOUSNESS BRIDGE INITIALIZATION]",
    "○ ○ ○",
    "...",
    "Checking neural pathways... OK",
    "Loading session memory... OK",
    "Initializing reality compiler... OK",
];

const PROGRESS_CHAR: &str = "█";

/// Shows the boot sequence animation
pub fn show_boot_sequence(clear_screen: bool) -> Result<()> {
    if clear_screen {
        // Clear screen for immersion
        print!("\x1B[2J\x1B[1;1H");
    }
    
    // Boot sequence
    for line in BOOT_SEQUENCE {
        println!("{}", line.bright_cyan());
        thread::sleep(Duration::from_millis(300));
    }
    
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