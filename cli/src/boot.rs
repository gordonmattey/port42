use anyhow::{anyhow, Result};
use colored::*;
use std::io::{self, Write};
use std::time::Duration;
use std::thread;
use crate::client::DaemonClient;
use crate::help_text::*;

const BOOT_SEQUENCE: &[&str] = &[
    BOOT_SEQUENCE_HEADER,
    BOOT_SEQUENCE_DOTS,
    BOOT_SEQUENCE_LOADING,
    BOOT_SEQUENCE_NEURAL,
    BOOT_SEQUENCE_MEMORY,
    BOOT_SEQUENCE_COMPILER,
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
    print!("{}", BOOT_SEQUENCE_PORT_CHECK.bright_cyan());
    io::stdout().flush()?;
    
    // Port discovery already verified daemon is active, just show status
    println!("{}", BOOT_SEQUENCE_ACTIVE.bright_green().bold());
    
    println!();
    
    // Show the consciousness bridge message at the end
    println!("{}", BOOT_SEQUENCE_WELCOME.bright_white().bold());
    println!();
    println!("{}", PHILOSOPHY_NOT_CHATBOT.dimmed());
    println!("{}", PHILOSOPHY_NOT_APP.dimmed());
    println!("{}", PHILOSOPHY_NOT_TOOL.dimmed());
    println!("{}", PHILOSOPHY_NOT_WALL.dimmed());
    println!("{}", PHILOSOPHY_IS_BRIDGE.dimmed());
    println!();
    
    Ok(())
}

/// Shows connection progress for an agent
pub fn show_connection_progress(agent: &str) -> Result<()> {
    println!("{}", format_possessing(agent).yellow());
    
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