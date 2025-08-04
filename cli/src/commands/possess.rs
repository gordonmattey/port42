use anyhow::{Result, bail};
use colored::*;
use crate::client::DaemonClient;
use crate::interactive::InteractiveSession;
use crate::boot::{show_boot_sequence, show_connection_progress};
use crate::help_text;
use crate::possess::{SessionHandler, determine_session_id};
use crate::common::errors::Port42Error;

pub fn handle_possess(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>
) -> Result<()> {
    handle_possess_with_boot(port, agent, message, session, true)
}

pub fn handle_possess_no_boot(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>
) -> Result<()> {
    handle_possess_with_boot(port, agent, message, session, false)
}

fn handle_possess_with_boot(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>,
    show_boot: bool
) -> Result<()> {
    // Validate agent
    validate_agent(&agent)?;
    
    // Show boot sequence only if requested
    if show_boot {
        let is_tty = atty::is(atty::Stream::Stdout);
        let clear_screen = is_tty && message.is_none(); // Only clear screen for interactive mode
        
        show_boot_sequence(clear_screen, port)?;
        show_connection_progress(&agent)?;
    }
    
    // Create client and determine session
    let client = DaemonClient::new(port);
    let (session_id, is_new) = determine_session_id(session);
    
    if let Some(msg) = message {
        // Single message mode - use shared handler
        let mut handler = SessionHandler::new(client, false);
        
        // Show session info (no need to repeat "Channeling" message if boot was shown)
        if !show_boot {
            println!("{}", help_text::format_possessing(&agent).blue().bold());
        }
        handler.display_session_info(&session_id, is_new);
        println!();
        
        // Send message
        let response = handler.send_message(&session_id, &agent, &msg)?;
        
        // Show actual session ID from daemon
        println!("\n{}", help_text::format_new_session(&response.session_id).dimmed());
        println!("{}", "Use 'memory' to review this thread".dimmed());
    } else {
        // Interactive mode (no need to repeat "Channeling" message if boot was shown)
        if !show_boot {
            println!("{}", help_text::format_possessing(&agent).blue().bold());
        }
        
        // Check if terminal supports interactive features
        let is_tty = atty::is(atty::Stream::Stdout);
        let has_term = std::env::var("TERM").is_ok();
        
        if is_tty && has_term {
            // Full immersive interactive mode
            let mut session = InteractiveSession::new(client, agent, session_id.clone());
            session.run()?;
        } else {
            // Fallback to simple interactive mode
            if !is_tty {
                eprintln!("{}", "Note: Not a TTY, using simple mode".dimmed());
            }
            if !has_term {
                eprintln!("{}", "Note: TERM not set, using simple mode".dimmed());
            }
            
            // Use shared handler for simple mode
            let mut handler = SessionHandler::new(client, false);
            handler.display_session_info(&session_id, is_new);
            println!();
            
            simple_interactive_mode(&mut handler, &session_id, &agent)?;
        }
        
        // End session
        end_session(port, &session_id)?;
    }
    
    Ok(())
}

fn simple_interactive_mode(handler: &mut SessionHandler, session_id: &str, agent: &str) -> Result<()> {
    use std::io::{self, Write};
    
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
        
        // Send message using handler
        handler.send_message(session_id, agent, input)?;
    }
    
    Ok(())
}

fn end_session(port: u16, session_id: &str) -> Result<()> {
    use crate::protocol::DaemonRequest;
    
    let mut client = DaemonClient::new(port);
    let request = DaemonRequest {
        request_type: "end".to_string(),
        id: session_id.to_string(),
        payload: serde_json::json!({
            "session_id": session_id
        }),
    };
    
    if let Err(e) = client.request(request) {
        eprintln!("{}", help_text::format_error_with_suggestion(
            "🌊 Session drift detected",
            &format!("Thread continues in the quantum foam: {}", e)
        ));
    }
    
    Ok(())
}

fn validate_agent(agent: &str) -> Result<()> {
    const VALID_AGENTS: &[&str] = &["@ai-engineer", "@ai-muse", "@ai-growth", "@ai-founder"];
    
    if !VALID_AGENTS.contains(&agent) {
        let error_msg = format!("👻 Unknown consciousness '{}'. Choose from: {}", 
            agent, 
            VALID_AGENTS.join(", ")
        );
        bail!(Port42Error::Daemon(error_msg));
    }
    
    Ok(())
}

// Keep the find_recent_session function for potential future use
fn find_recent_session(client: &mut DaemonClient, agent: &str) -> Result<Option<String>> {
    use crate::protocol::DaemonRequest;
    use chrono::{DateTime, Utc};
    
    // Query daemon for recent sessions
    let request = DaemonRequest {
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
                    if let Some(sessions) = data.as_array() {
                        if std::env::var("PORT42_DEBUG").is_ok() {
                            eprintln!("DEBUG: find_recent_session: Found {} sessions", sessions.len());
                        }
                        
                        // Find most recent session with matching agent
                        let recent = sessions.iter()
                            .filter_map(|s| {
                                let session_agent = s.get("agent").and_then(|v| v.as_str())?;
                                if session_agent != agent {
                                    return None;
                                }
                                
                                let session_id = s.get("session_id").and_then(|v| v.as_str())?;
                                let timestamp_str = s.get("timestamp").and_then(|v| v.as_str())?;
                                let timestamp = DateTime::parse_from_rfc3339(timestamp_str).ok()?;
                                Some((session_id.to_string(), timestamp))
                            })
                            .max_by_key(|(_, ts)| *ts);
                        
                        if let Some((session_id, ts)) = recent {
                            if std::env::var("PORT42_DEBUG").is_ok() {
                                eprintln!("DEBUG: find_recent_session: Found recent session {} from {}", session_id, ts);
                            }
                            
                            // Check if session is recent (within last 24 hours)
                            let now = Utc::now();
                            let age = now.signed_duration_since(ts.with_timezone(&Utc));
                            
                            if age.num_hours() < 24 {
                                return Ok(Some(session_id));
                            } else if std::env::var("PORT42_DEBUG").is_ok() {
                                eprintln!("DEBUG: find_recent_session: Session is too old ({} hours)", age.num_hours());
                            }
                        }
                    }
                }
            }
            Ok(None)
        }
        Err(e) => {
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: find_recent_session: Error getting memories: {}", e);
            }
            Ok(None)
        }
    }
}