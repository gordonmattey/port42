use anyhow::{Context, Result};
use colored::*;
use std::io::{Read, Write};
use std::net::TcpStream;
use std::time::Duration;

use crate::types::{Request, Response};
use crate::help_text::*;

pub fn handle_status(port: u16, detailed: bool) -> Result<()> {
    println!("{}", MSG_CHECKING_STATUS.blue().bold());
    
    // Try to connect
    let mut stream = match TcpStream::connect_timeout(
        &format!("127.0.0.1:{}", port).parse()?,
        Duration::from_secs(2)
    ) {
        Ok(stream) => stream,
        Err(_) => {
            println!("{}", format_daemon_connection_error(port));
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
        .context(ERR_CONNECTION_LOST)?;
    
    let response: Response = serde_json::from_slice(&buffer[..n])
        .context(ERR_INVALID_RESPONSE)?;
    
    if response.success {
        println!("{}", MSG_DAEMON_RUNNING.green().bold());
        
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
            
            println!("\n{}", MSG_CONNECTION_INFO.bright_white());
            println!("{}", format_port_info(&port.to_string().bright_cyan().to_string()));
            println!("{}", format_uptime_info(&uptime.bright_cyan().to_string()));
            println!("{}", format_sessions_info(&sessions.to_string().bright_cyan().to_string()));
            
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
        
        println!("\n{}", MSG_DOLPHINS_LISTENING.blue().italic());
    } else {
        println!("{}", ERR_CONNECTION_LOST.red());
        if let Some(error) = response.error {
            println!("  {}", error.dimmed());
        }
    }
    
    Ok(())
}