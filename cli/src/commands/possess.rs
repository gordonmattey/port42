use anyhow::Result;
use colored::*;
use crate::client::DaemonClient;
use crate::interactive::InteractiveSession;
use crate::types::Request;
use std::io::{self, Write};
use chrono::{DateTime, Utc};

pub fn handle_possess(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>
) -> Result<()> {
    println!("{}", format!("🔮 Possessing {}...", agent).blue().bold());
    
    let mut client = DaemonClient::new(port);
    
    // Determine session ID: use provided, continue recent, or generate new
    let session_id = if let Some(id) = session {
        // User explicitly provided a session ID
        id
    } else {
        // Try to find and continue the most recent session with this agent
        match find_recent_session(&mut client, &agent)? {
            Some(recent_id) => {
                println!("{}", format!("↻ Continuing recent session: {}", recent_id).dimmed());
                recent_id
            }
            None => {
                // No recent session, create new
                format!("cli-{}", chrono::Utc::now().timestamp())
            }
        }
    };
    
    if let Some(msg) = message {
        // Single message mode
        println!("{}", "Sending single message...".dimmed());
        send_message(&mut client, &session_id, &agent, &msg)?;
    } else {
        // Check if terminal supports interactive features
        let is_tty = atty::is(atty::Stream::Stdout);
        let has_term = std::env::var("TERM").is_ok();
        
        if is_tty && has_term {
            // Full immersive interactive mode
            let mut session = InteractiveSession::new(client, agent, session_id.clone());
            session.run()?;
        } else {
            // Fallback to simple interactive mode (for pipes, non-TTY, etc)
            if !is_tty {
                eprintln!("{}", "Note: Not a TTY, using simple mode".dimmed());
            }
            if !has_term {
                eprintln!("{}", "Note: TERM not set, using simple mode".dimmed());
            }
            simple_interactive_mode(&mut client, &session_id, &agent)?;
        }
        
        // End session with a new client (ownership was moved to interactive session)
        let mut end_client = DaemonClient::new(port);
        let end_request = Request {
            request_type: "end".to_string(),
            id: session_id.clone(),
            payload: serde_json::json!({
                "session_id": session_id
            }),
        };
        
        if let Err(e) = end_client.request(end_request) {
            eprintln!("{}", format!("⚠️  Failed to end session: {}", e).yellow());
        }
    }
    
    Ok(())
}

fn simple_interactive_mode(client: &mut DaemonClient, session_id: &str, agent: &str) -> Result<()> {
    println!("{}", "Entering interactive mode. Type '/end' to finish.".dimmed());
    println!();
    
    loop {
        // Prompt
        print!("{} ", ">".bright_blue());
        io::stdout().flush()?;
        
        // Read input
        let mut input = String::new();
        io::stdin().read_line(&mut input)?;
        let input = input.trim();
        
        // Check for exit
        if input == "/end" || input.is_empty() {
            break;
        }
        
        // Send message
        send_message(client, session_id, agent, input)?;
    }
    
    Ok(())
}

fn send_message(client: &mut DaemonClient, session_id: &str, agent: &str, message: &str) -> Result<()> {
    if std::env::var("PORT42_DEBUG").is_ok() {
        eprintln!("DEBUG: Sending message to session: {}", session_id);
        eprintln!("DEBUG: Message length: {} chars", message.len());
    }
    
    let request = Request {
        request_type: "possess".to_string(),
        id: session_id.to_string(),
        payload: serde_json::json!({
            "agent": agent,
            "message": message
        }),
    };
    
    if std::env::var("PORT42_DEBUG").is_ok() {
        eprintln!("DEBUG: About to send request to daemon");
    }
    
    match client.request(request) {
        Ok(response) => {
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: Got response from daemon, success={}", response.success);
                if let Some(data) = &response.data {
                    // Check size without serializing
                    if let Some(obj) = data.as_object() {
                        eprintln!("DEBUG: Response data has {} keys", obj.len());
                        for key in obj.keys() {
                            eprintln!("DEBUG:   Key: {}", key);
                        }
                    }
                }
            }
            
            if response.success {
                if let Some(data) = response.data {
                    if let Some(ai_message) = data.get("message").and_then(|v| v.as_str()) {
                        println!("\n{}", agent.bright_blue());
                        println!("{}", ai_message);
                        println!();
                        
                        // Check if command was generated
                        // The daemon sends command_generated=true and command_spec with the details
                        if data.get("command_generated").and_then(|v| v.as_bool()).unwrap_or(false) {
                            if let Some(spec) = data.get("command_spec") {
                                if let Some(name) = spec.get("name").and_then(|v| v.as_str()) {
                                    println!("{}", format!("✨ Command crystallized: {}", name).bright_green().bold());
                                    println!("{}", "Add to PATH to use:".yellow());
                                    println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
                                    println!();
                                }
                            }
                        }
                    } else {
                        println!("{}", "No message in response".dimmed());
                    }
                } else {
                    println!("{}", "No data in response".dimmed());
                }
            } else {
                println!("{}", "❌ Failed to send message".red());
                if let Some(error) = response.error {
                    println!("  {}", error.dimmed());
                }
            }
        }
        Err(e) => {
            eprintln!("{}", e);
            return Err(e);
        }
    }
    
    Ok(())
}

fn find_recent_session(client: &mut DaemonClient, agent: &str) -> Result<Option<String>> {
    // Query daemon for recent sessions
    let request = Request {
        request_type: "memory".to_string(),
        id: "cli-memory-query".to_string(),
        payload: serde_json::Value::Null,
    };
    
    if std::env::var("PORT42_DEBUG").is_ok() {
        eprintln!("DEBUG: find_recent_session: About to request memory from daemon");
    }
    
    match client.request(request) {
        Ok(response) => {
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: find_recent_session: Got memory response, success={}", response.success);
            }
            
            if response.success {
                if let Some(data) = response.data {
                    // Debug: Check response without serializing
                    if std::env::var("PORT42_DEBUG").is_ok() {
                        if let Some(obj) = data.as_object() {
                            eprintln!("DEBUG: Memory response has {} keys", obj.len());
                        }
                        if let Some(recent) = data.get("recent_sessions").and_then(|v| v.as_array()) {
                            eprintln!("DEBUG: Found {} recent sessions", recent.len());
                        }
                    }
                    
                    // Check recent_sessions array
                    if let Some(recent) = data.get("recent_sessions").and_then(|v| v.as_array()) {
                        // Find the most recent session with this agent
                        let mut best_session: Option<(String, DateTime<Utc>)> = None;
                        
                        for session in recent {
                            if let (Some(session_agent), Some(id), Some(last_activity)) = (
                                session.get("agent").and_then(|v| v.as_str()),
                                session.get("id").and_then(|v| v.as_str()),
                                session.get("last_activity").and_then(|v| v.as_str())
                            ) {
                                // Match agent (with @ prefix handling)
                                let session_agent_normalized = if session_agent.starts_with('@') {
                                    session_agent.to_string()
                                } else {
                                    format!("@{}", session_agent)
                                };
                                
                                if session_agent_normalized == agent {
                                    // Parse timestamp
                                    if let Ok(activity_time) = last_activity.parse::<DateTime<Utc>>() {
                                        // Only consider sessions less than 24 hours old
                                        let age = Utc::now() - activity_time;
                                        if age.num_hours() < 24 {
                                            match &best_session {
                                                None => best_session = Some((id.to_string(), activity_time)),
                                                Some((_, best_time)) => {
                                                    if activity_time > *best_time {
                                                        best_session = Some((id.to_string(), activity_time));
                                                    }
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        
                        Ok(best_session.map(|(id, _)| id))
                    } else {
                        Ok(None)
                    }
                } else {
                    Ok(None)
                }
            } else {
                // If memory query fails, just create new session
                Ok(None)
            }
        }
        Err(_) => {
            // If we can't query memory, just create new session
            Ok(None)
        }
    }
}