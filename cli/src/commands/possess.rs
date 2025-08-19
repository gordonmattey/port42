use anyhow::{Result, bail};
use colored::*;
use crate::client::DaemonClient;
use crate::interactive::InteractiveSession;
use crate::boot::{show_boot_sequence, show_connection_progress};
use crate::help_text;
use crate::possess::{SessionHandler, determine_session_id};
use crate::common::{errors::Port42Error, references::parse_references};
use crate::commands::search;

pub fn handle_possess(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>
) -> Result<()> {
    handle_possess_with_search(port, agent, message, session, None, true)
}

pub fn handle_possess_with_references(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>,
    search_query: Option<String>,
    references: Option<Vec<String>>,
    show_boot: bool
) -> Result<()> {
    // Handle both search and references
    let mut memory_context = Vec::new();
    
    // Load memory contexts from search if provided
    if let Some(ref query) = search_query {
        let mut client = DaemonClient::new(port);
        let search_contexts = load_search_results_as_context(&mut client, &query, &agent)?;
        memory_context.extend(search_contexts);
        
        if show_boot {
            let is_tty = atty::is(atty::Stream::Stdout);
            show_boot_sequence(is_tty, port)?;
            show_connection_progress(&agent)?;
        }
        
        println!("{}", format!("🔍 Searching memories with query: '{}'", query.bright_yellow()));
        println!("{}", "Loading matching memories into consciousness...".blue().italic());
        
        // Show search results  
        search::handle_search_with_format(
            &mut client,
            query.clone(),
            None, None, None, None, 
            Some(agent.clone()),
            vec![], Some(10),
            crate::display::OutputFormat::Plain,
        )?;
        println!();
    }
    
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
    
    if memory_context.is_empty() && search_query.is_some() {
        println!("{}", "No memories found to load into session context.".yellow());
    } else if !memory_context.is_empty() {
        println!("{}", format!("✨ Loaded {} contexts into session.", memory_context.len()).green());
    }
    println!();
    
    // Use unified flow with all context and references
    handle_possess_with_boot_and_context(port, agent, message, session, show_boot, memory_context, parsed_refs)
}

pub fn handle_possess_with_search(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>,
    search_query: Option<String>,
    show_boot: bool
) -> Result<()> {
    if let Some(query) = search_query {
        // Load memory contexts from search
        let mut client = DaemonClient::new(port);
        let memory_context = load_search_results_as_context(&mut client, &query, &agent)?;
        
        // Display search results summary (existing behavior)
        if show_boot {
            let is_tty = atty::is(atty::Stream::Stdout);
            show_boot_sequence(is_tty, port)?;
            show_connection_progress(&agent)?;
        }
        
        println!("{}", format!("🔍 Searching memories with query: '{}'", query.bright_yellow()));
        println!("{}", "Loading matching memories into consciousness...".blue().italic());
        
        // Show search results  
        search::handle_search_with_format(
            &mut client,
            query.clone(),
            None, // path
            None, // type_filter  
            None, // after
            None, // before
            Some(agent.clone()), // agent filter
            vec![], // tags
            Some(10), // limit
            crate::display::OutputFormat::Plain,
        )?;
        
        println!();
        if memory_context.is_empty() {
            println!("{}", "No memories found to load into session context.".yellow());
        } else {
            println!("{}", format!("✨ Loaded {} memories into session context.", memory_context.len()).green());
        }
        println!();
        
        // Use unified flow with memory context
        handle_possess_with_boot_and_context(port, agent, message, session, false, memory_context, None)
    } else {
        handle_possess_with_boot(port, agent, message, session, show_boot)
    }
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
        handler.send_message_with_context(session_id, agent, input, memory_ctx.clone(), references.clone())?;
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

fn load_search_results_as_context(
    client: &mut DaemonClient,
    search_query: &str,
    agent: &str,
) -> Result<Vec<String>> {
    use crate::protocol::{SearchRequest, SearchFilters, ResponseParser, RequestBuilder};
    use crate::protocol::search::SearchResponse;
    
    // Build search request
    let mut filters = SearchFilters::default();
    filters.agent = Some(agent.to_string());
    filters.limit = Some(5); // Limit to top 5 memories to avoid overwhelming context
    
    let request = SearchRequest::new(search_query.to_string())
        .with_filters(filters)
        .build_request(format!("search-context-{}", 
            std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).unwrap().as_millis()))?;
    
    // Execute search
    let response = client.request(request)?;
    
    if !response.success {
        bail!("Search failed: {}", response.error.unwrap_or_else(|| "Unknown error".to_string()));
    }
    
    let data = response.data.ok_or_else(|| anyhow::anyhow!("No search results"))?;
    let search_response = SearchResponse::parse_response(&data)?;
    
    // Load full memory content for each result
    let mut memory_contexts = Vec::new();
    
    for result in search_response.results.iter().take(5) {
        // Only load session type results (memory sessions)
        if result.result_type == "session" {
            eprintln!("DEBUG: Attempting to load memory from path: {}", result.path);
            match load_memory_content(client, &result.path) {
                Ok(content) => {
                    eprintln!("DEBUG: Successfully loaded memory content ({} chars)", content.len());
                    memory_contexts.push(content);
                }
                Err(e) => {
                    eprintln!("DEBUG: Failed to load memory content: {}", e);
                }
            }
        }
    }
    
    Ok(memory_contexts)
}

fn load_memory_content(client: &mut DaemonClient, memory_path: &str) -> Result<String> {
    use crate::protocol::DaemonRequest;
    
    // Extract session ID from path (e.g., "/memory/cli-1754709496765" -> "cli-1754709496765")
    let session_id = memory_path.strip_prefix("/memory/").unwrap_or(memory_path);
    eprintln!("DEBUG: load_memory_content - path: {}, extracted session_id: {}", memory_path, session_id);
    
    // Request memory content from daemon
    let request = DaemonRequest {
        request_type: "memory".to_string(),
        id: format!("memory-load-{}", session_id),
        payload: serde_json::json!({
            "session_id": session_id,
            "include_content": true
        }),
        references: None,
        session_context: None,
        user_prompt: None,
    };
    
    eprintln!("DEBUG: load_memory_content - sending request: {:?}", request);
    let response = client.request(request)?;
    eprintln!("DEBUG: load_memory_content - got response: success={}, data_present={}", 
              response.success, response.data.is_some());
    
    if !response.success {
        bail!("Failed to load memory {}: {}", session_id, 
              response.error.unwrap_or_else(|| "Unknown error".to_string()));
    }
    
    // Extract conversation content
    if let Some(data) = response.data {
        eprintln!("DEBUG: load_memory_content - data keys: {:?}", data.as_object().map(|o| o.keys().collect::<Vec<_>>()));
        if let Some(messages) = data.get("messages") {
            eprintln!("DEBUG: load_memory_content - found messages field, type: {:?}", messages);
            if let Some(messages) = messages.as_array() {
                eprintln!("DEBUG: load_memory_content - messages array length: {}", messages.len());
                let mut content = String::new();
                content.push_str(&format!("=== Memory {} ===\n", session_id));
                
                for message in messages {
                    if let (Some(role), Some(msg_content)) = 
                        (message.get("role").and_then(|r| r.as_str()),
                         message.get("content").and_then(|c| c.as_str())) {
                        content.push_str(&format!("{}: {}\n\n", role, msg_content));
                    }
                }
                
                eprintln!("DEBUG: load_memory_content - extracted content length: {}", content.len());
                return Ok(content);
            } else {
                eprintln!("DEBUG: load_memory_content - messages field is not an array");
            }
        } else {
            eprintln!("DEBUG: load_memory_content - no messages field found");
        }
    } else {
        eprintln!("DEBUG: load_memory_content - no data in response");
    }
    
    bail!("No conversation content found for memory {}", session_id)
}


fn start_session_with_context(session: InteractiveSession, memory_contexts: Vec<String>) -> Result<()> {
    // For now, we'll just display the loaded context and start the session
    // In the future, we could pre-load this context by sending it to the AI
    if !memory_contexts.is_empty() {
        println!("{}", "🧠 Memory context summary:".bright_cyan());
        for (i, context) in memory_contexts.iter().enumerate() {
            let lines: Vec<&str> = context.lines().collect();
            let summary = if lines.len() > 3 {
                format!("{}...", lines[0..3].join("\n"))
            } else {
                context.clone()
            };
            println!("{}: {}", format!("{}", i + 1).dimmed(), summary.dimmed());
        }
        println!();
        println!("{}", "This context is available to reference during the session.".green());
        println!();
    }
    
    // Start the normal interactive session
    // The user can use /import commands to pull specific memories if needed
    let mut session = session;
    session.run()?;
    
    Ok(())
}

// Removed handle_possess_search_mode - now using unified flow via handle_possess_with_boot_and_context