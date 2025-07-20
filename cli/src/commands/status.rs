use anyhow::{Context, Result};
use colored::*;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::time::Duration;

use crate::types::{Request, Response};

pub fn handle_status(port: u16, detailed: bool) -> Result<()> {
    println!("{}", "🐬 Checking Port 42 status...".blue().bold());
    
    // Try to connect
    let mut stream = match TcpStream::connect_timeout(
        &format!("127.0.0.1:{}", port).parse()?,
        Duration::from_secs(2)
    ) {
        Ok(stream) => stream,
        Err(_) => {
            println!("{}", "❌ Daemon not running".red());
            println!("\n{}", "To start the daemon:".yellow());
            println!("  {}", "sudo -E ./bin/port42d".bright_white());
            println!("\n{}", "Or use:".yellow());
            println!("  {}", "port42 daemon start".bright_white());
            return Ok(());
        }
    };
    
    // Send status request
    let request = Request {
        request_type: "status".to_string(),
        id: "cli-status".to_string(),
        payload: serde_json::Value::Null,
    };
    
    let request_json = serde_json::to_string(&request)?;
    stream.write_all(request_json.as_bytes())?;
    stream.write_all(b"\n")?;
    
    // Read response
    let mut buffer = vec![0; 4096];
    let n = stream.read(&mut buffer)
        .context("Failed to read response from daemon")?;
    
    let response: Response = serde_json::from_slice(&buffer[..n])
        .context("Failed to parse daemon response")?;
    
    if response.success {
        println!("{}", "✅ Daemon is running".green().bold());
        
        if let Some(data) = response.data {
            // Extract daemon info
            let port = data.get("port")
                .and_then(|v| v.as_u64())
                .unwrap_or(42);
            let uptime = data.get("uptime")
                .and_then(|v| v.as_str())
                .unwrap_or("unknown");
            let sessions = data.get("active_sessions")
                .and_then(|v| v.as_u64())
                .unwrap_or(0);
            
            println!("\n{}", "Connection Info:".bright_white());
            println!("  Port:     {}", port.to_string().bright_cyan());
            println!("  Uptime:   {}", uptime.bright_cyan());
            println!("  Sessions: {}", sessions.to_string().bright_cyan());
            
            if detailed {
                println!("\n{}", "Detailed Status:".bright_white());
                
                // Memory stats
                if let Some(memory_stats) = data.get("memory_stats") {
                    println!("\n  {}", "Memory Store:".yellow());
                    if let Some(total) = memory_stats.get("total_sessions") {
                        println!("    Total Sessions: {}", total.to_string().bright_cyan());
                    }
                    if let Some(commands) = memory_stats.get("commands_generated") {
                        println!("    Commands Made:  {}", commands.to_string().bright_cyan());
                    }
                }
                
                // Recent activity
                if let Some(_recent) = data.get("recent_activity") {
                    println!("\n  {}", "Recent Activity:".yellow());
                    // Would parse and display recent sessions
                }
            }
        }
        
        println!("\n{}", "The dolphins are listening... 🐬".blue().italic());
    } else {
        println!("{}", "❌ Daemon returned error".red());
        if let Some(error) = response.error {
            println!("  {}", error);
        }
    }
    
    Ok(())
}