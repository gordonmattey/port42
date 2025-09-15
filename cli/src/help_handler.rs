//! Custom help handler to unify interactive and CLI help
//! 
//! This module intercepts help requests and displays our rich,
//! reality compiler themed help instead of Clap's default.

use crate::help_text;
use std::env;

/// Check if this is a help request and handle it
/// Returns true if help was handled, false otherwise
pub fn handle_help_request() -> bool {
    let args: Vec<String> = env::args().collect();
    
    // Check for "port42 --help" or "port42 -h"
    if args.len() == 2 && (args[1] == "--help" || args[1] == "-h") {
        show_main_help();
        return true;
    }
    
    // Check for "port42 help <command>" pattern
    if args.len() >= 2 && args[1] == "help" {
        if args.len() == 2 {
            // Just "port42 help" - show main help
            show_main_help();
        } else {
            // "port42 help <command>"
            let command = &args[2];
            help_text::show_command_help(command);
        }
        return true;
    }
    
    // Check for "port42 <command> --help" or "port42 <command> -h" or "port42 <command> -help"
    if args.len() >= 3 && (args[args.len() - 1] == "--help" || args[args.len() - 1] == "-h" || args[args.len() - 1] == "-help") {
        // Extract command name (second argument)
        let command = &args[1];
        
        // Map command to our help
        match command.as_str() {
            "swim" | "memory" | "ls" | "cat" | "info" | "search" | "reality" | "status" | "init" | "daemon" => {
                help_text::show_command_help(command);
                return true;
            }
            _ => {
                // Let Clap handle unknown commands
                return false;
            }
        }
    }
    
    false
}

/// Show main help with reality compiler essence
fn show_main_help() {
    println!("{}", help_text::MAIN_ABOUT);
    println!();
    println!("{}", help_text::MAIN_LONG_ABOUT);
    println!();
    
    println!("{}", "CONSCIOUSNESS OPERATIONS:".bright_cyan());
    println!("  {} - {}", "swim <agent>".bright_green(), help_text::SWIM_DESC);
    println!("  {} - {}", "memory".bright_green(), help_text::MEMORY_DESC);
    println!("  {} - {}", "reality".bright_green(), help_text::REALITY_DESC);
    println!();
    
    println!("{}", "REALITY NAVIGATION:".bright_cyan());
    println!("  {} - {}", "ls [path]".bright_green(), help_text::LS_DESC);
    println!("  {} - {}", "cat <path>".bright_green(), help_text::CAT_DESC);
    println!("  {} - {}", "info <path>".bright_green(), help_text::INFO_DESC);
    println!("  {} - {}", "search <query>".bright_green(), help_text::SEARCH_DESC);
    println!();
    
    println!("{}", "SYSTEM:".bright_cyan());
    println!("  {} - {}", "daemon".bright_green(), help_text::DAEMON_DESC);
    println!("  {} - {}", "status".bright_green(), help_text::STATUS_DESC);
    println!();
    
    println!("{}", "OPTIONS:".bright_cyan());
    println!("  {} - Port for consciousness gateway", "-p, --port <PORT>".bright_green());
    println!("  {} - Verbose output for deeper introspection", "-v, --verbose".bright_green());
    println!("  {} - Print help", "-h, --help".bright_green());
    println!();
    
    println!("{}", "For detailed command help: port42 help <command>".yellow());
    println!();
    println!("{}", "The dolphins are listening on Port 42. Will you let them in?".bright_blue());
}

use colored::*;