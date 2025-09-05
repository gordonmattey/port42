use anyhow::{Result, bail};
use colored::*;
use crate::client::DaemonClient;
use crate::interactive::InteractiveSession;
use crate::boot::{show_boot_sequence, show_connection_progress};
use crate::help_text;
use crate::possess::{SessionHandler, determine_session_id};
use crate::common::{errors::Port42Error, references::parse_references};

pub fn handle_possess(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>
) -> Result<()> {
    // Auto-detect output mode: show boot only for interactive mode (no message)
    let show_boot = message.is_none();
    handle_possess_with_references(port, agent, message, session, None, show_boot)
}

pub fn handle_possess_with_references(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>,
    references: Option<Vec<String>>,
    show_boot: bool
) -> Result<()> {
    // Parse references if provided - daemon will resolve them server-side
    let parsed_refs = if let Some(ref_strings) = references {
        println!("{}", format!("🔗 Preparing {} references for AI context...", ref_strings.len()).bright_cyan());
        match parse_references(ref_strings, true) {
            Ok(refs) => {
                println!("{}", format!("✅ Parsed {} references", refs.len()).green());
                Some(refs)
            },
            Err(e) => {
                eprintln!("{} {}", "❌ Invalid reference:".red(), e);
                std::process::exit(1);
            }
        }
    } else {
        None
    };
    
    // Use unified flow with references - no manual memory context loading
    handle_possess_with_boot_and_context(port, agent, message, session, show_boot, Vec::new(), parsed_refs)
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
    handle_possess_with_boot_and_context(port, agent, message, session, show_boot, Vec::new(), None)
}

fn handle_possess_with_boot_and_context(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>,
    show_boot: bool,
    memory_context: Vec<String>,
    references: Option<Vec<crate::protocol::relations::Reference>>
) -> Result<()> {
    // Validate agent
    validate_agent(&agent)?;
    
    // Show boot sequence only if requested
    if show_boot {
        let is_tty = atty::is(atty::Stream::Stdout);
        // Don't clear screen if we have references - user needs to see them
        let has_references = references.is_some() && !references.as_ref().unwrap().is_empty();
        let clear_screen = is_tty && message.is_none() && !has_references;
        
        show_boot_sequence(clear_screen, port)?;
        show_connection_progress(&agent)?;
    }
    
    // Create client and determine session
    let client = DaemonClient::new(port);
    let (session_id, is_new) = determine_session_id(session);
    
    if let Some(msg) = message {
        // Single message mode - use shared handler
        let mut handler = SessionHandler::new(client, false);
        
        // Show minimal connection info for CLI mode, full session info for interactive
        if !show_boot {
            // CLI mode: just show channeling message, no session details
            println!("{}", help_text::format_possessing(&agent).blue().bold());
        } else {
            // Interactive mode: show full session info
            handler.display_session_info(&session_id, is_new);
        }
        println!();
        
        // Show memory context summary if present
        if !memory_context.is_empty() {
            println!("{}", "🧠 Memory context summary:".bright_cyan());
            for (i, context) in memory_context.iter().enumerate() {
                // Extract just the reference header for display
                let lines: Vec<&str> = context.lines().collect();
                if let Some(first_line) = lines.first() {
                    if first_line.starts_with("=== Reference:") {
                        // Extract reference name from header
                        let ref_name = first_line.replace("=== Reference:", "").replace("===", "").trim().to_string();
                        let content_lines = lines.len().saturating_sub(2); // Subtract header and blank line
                        println!("{}: {} ({} lines)", 
                            format!("{}", i + 1).dimmed(), 
                            ref_name.bright_white(),
                            content_lines);
                    } else {
                        // Fallback for non-reference contexts (memory, etc.)
                        let summary = if lines.len() > 1 {
                            lines[0].chars().take(80).collect::<String>()
                        } else {
                            context.chars().take(80).collect::<String>()
                        };
                        println!("{}: {}...", 
                            format!("{}", i + 1).dimmed(), 
                            summary.dimmed());
                    }
                }
            }
            println!();
            println!("{}", "This context is available to reference during the session.".green());
            println!();
        }
        
        // Send message with memory context and references
        let memory_ctx = if memory_context.is_empty() { None } else { Some(memory_context) };
        let response = handler.send_message_with_context(&session_id, &agent, &msg, memory_ctx, references)?;
        
        // Show session completion with actual daemon session ID
        println!();
        handler.display_session_complete(&response.session_id);
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
            let memory_ctx = if memory_context.is_empty() { None } else { Some(memory_context) };
            let mut session = InteractiveSession::with_context(client, agent, session_id.clone(), memory_ctx, references);
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
            
            simple_interactive_mode_with_context(&mut handler, &session_id, &agent, memory_context, references)?;
        }
        
        // End session
        end_session(port, &session_id)?;
    }
    
    Ok(())
}

fn simple_interactive_mode_with_context(
    handler: &mut SessionHandler, 
    session_id: &str, 
    agent: &str,
    memory_context: Vec<String>,
    references: Option<Vec<crate::protocol::relations::Reference>>
) -> Result<()> {
    use std::io::{self, Write};
    
    println!("{}", "Entering interactive mode. Type '/end' to finish.".dimmed());
    println!();
    
    // Convert memory_context to Option for consistency
    let memory_ctx = if memory_context.is_empty() { None } else { Some(memory_context) };
    let mut actual_session_id = session_id.to_string();
    
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
        
        // Send message with session context
        let response = handler.send_message_with_context(session_id, agent, input, memory_ctx.clone(), references.clone())?;
        
        // Track the actual session ID from daemon response
        actual_session_id = response.session_id;
    }
    
    // Show session completion with actual session ID
    println!();
    handler.display_session_complete(&actual_session_id);
    println!("{}", "Use 'memory' to review this thread".dimmed());
    
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
        references: None,
        session_context: None,
        user_prompt: None,
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
    const VALID_AGENTS: &[&str] = &["@ai-engineer", "@ai-muse", "@ai-analyst", "@ai-founder"];
    
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
        references: None,
        session_context: None,
        user_prompt: None,
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





// Removed handle_possess_search_mode - now using unified flow via handle_possess_with_boot_and_context