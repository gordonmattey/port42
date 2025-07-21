use anyhow::Result;
use colored::*;
use crate::MemoryAction;
use crate::client::DaemonClient;
use crate::types::Request;
use chrono::{DateTime, FixedOffset};

pub fn handle_memory(port: u16, action: Option<MemoryAction>) -> Result<()> {
    let mut client = DaemonClient::new(port);
    
    match action {
        None | Some(MemoryAction::List { .. }) => {
            list_sessions(&mut client)?;
        }
        
        Some(MemoryAction::Search { query, limit: _ }) => {
            println!("{}", format!("🔍 Searching for: {}", query).blue().bold());
            println!("{}", "🚧 Memory search not yet implemented".yellow().dimmed());
            // Could implement by fetching all sessions and filtering
        }
        
        Some(MemoryAction::Show { session_id }) => {
            show_session(&mut client, &session_id)?;
        }
    }
    
    Ok(())
}

fn list_sessions(client: &mut DaemonClient) -> Result<()> {
    println!("{}", "🧠 Conversation Memory".blue().bold());
    println!();
    
    // Query daemon for memory
    let request = Request {
        request_type: "memory".to_string(),
        id: "cli-memory-list".to_string(),
        payload: serde_json::Value::Null,
    };
    
    let response = client.request(request)?;
    
    if response.success {
        if let Some(data) = response.data {
            // Show active sessions
            if let Some(active) = data.get("active_sessions").and_then(|v| v.as_array()) {
                if !active.is_empty() {
                    println!("{}", "Active Sessions:".bright_green().bold());
                    for session in active {
                        print_session_summary(session);
                    }
                    println!();
                }
            }
            
            // Show recent sessions
            if let Some(recent) = data.get("recent_sessions").and_then(|v| v.as_array()) {
                println!("{}", format!("Recent Sessions ({} found):", recent.len()).bright_cyan().bold());
                
                // Group by date
                let mut by_date: std::collections::HashMap<String, Vec<&serde_json::Value>> = std::collections::HashMap::new();
                
                for session in recent {
                    if let Some(date) = session.get("date").and_then(|v| v.as_str()) {
                        by_date.entry(date.to_string()).or_insert_with(Vec::new).push(session);
                    }
                }
                
                // Sort dates in reverse order (most recent first)
                let mut dates: Vec<_> = by_date.keys().cloned().collect();
                dates.sort_by(|a, b| b.cmp(a));
                
                for date in dates.iter().take(7) { // Show last 7 days
                    println!("\n{}", format!("  📅 {}", date).yellow());
                    
                    if let Some(sessions) = by_date.get(date) {
                        for session in sessions {
                            print_session_summary(session);
                        }
                    }
                }
            }
            
            // Show stats
            if let Some(stats) = data.get("stats") {
                println!("\n{}", "Statistics:".dimmed());
                if let Some(total) = stats.get("total_sessions").and_then(|v| v.as_u64()) {
                    println!("  Total sessions: {}", total);
                }
                if let Some(size) = stats.get("total_size_mb").and_then(|v| v.as_f64()) {
                    println!("  Storage used: {:.1} MB", size);
                }
            }
        }
    } else {
        println!("{}", "❌ Failed to retrieve memory".red());
        if let Some(error) = response.error {
            println!("  {}", error.dimmed());
        }
    }
    
    Ok(())
}

fn print_session_summary(session: &serde_json::Value) {
    let id = session.get("id").and_then(|v| v.as_str()).unwrap_or("unknown");
    let agent = session.get("agent").and_then(|v| v.as_str()).unwrap_or("unknown");
    let state = session.get("state").and_then(|v| v.as_str()).unwrap_or("unknown");
    let msg_count = session.get("message_count").and_then(|v| v.as_u64()).unwrap_or(0);
    let cmd_generated = session.get("command_generated").and_then(|v| v.as_bool()).unwrap_or(false);
    
    let state_icon = match state {
        "active" => "🟢",
        "idle" => "🟡",
        "completed" => "✅",
        "abandoned" => "❌",
        _ => "❓",
    };
    
    print!("    {} {} ", state_icon, id.bright_white());
    print!("({}) ", agent.bright_blue());
    print!("{} messages", msg_count);
    
    if cmd_generated {
        print!(" {}", "✨ command".bright_green());
    }
    
    println!();
}

fn show_session(client: &mut DaemonClient, session_id: &str) -> Result<()> {
    println!("{}", format!("📖 Session: {}", session_id).blue().bold());
    println!();
    
    // Query daemon for specific session
    let payload = serde_json::json!({
        "session_id": session_id
    });
    
    // Debug log the payload
    eprintln!("🔍 CLI sending memory show request with payload: {}", payload);
    
    let request = Request {
        request_type: "memory".to_string(),
        id: format!("cli-memory-show-{}", session_id),
        payload,
    };
    
    // Debug log the full request
    eprintln!("🔍 CLI full request: {:?}", serde_json::to_string(&request));
    
    let response = client.request(request)?;
    
    if response.success {
        if let Some(data) = response.data {
            // Display session details
            if let Some(agent) = data.get("agent").and_then(|v| v.as_str()) {
                println!("{}: {}", "Agent".dimmed(), agent.bright_blue());
            }
            
            if let Some(state) = data.get("state").and_then(|v| v.as_str()) {
                let state_display = match state {
                    "active" => "🟢 Active".green(),
                    "idle" => "🟡 Idle".yellow(),
                    "completed" => "✅ Completed".bright_green(),
                    "abandoned" => "❌ Abandoned".red(),
                    _ => state.normal(),
                };
                println!("{}: {}", "State".dimmed(), state_display);
            }
            
            if let Some(created) = data.get("created_at").and_then(|v| v.as_str()) {
                if let Ok(datetime) = DateTime::parse_from_rfc3339(created) {
                    println!("{}: {}", "Created".dimmed(), datetime.format("%Y-%m-%d %H:%M:%S"));
                }
            }
            
            if let Some(last_activity) = data.get("last_activity").and_then(|v| v.as_str()) {
                if let Ok(datetime) = DateTime::parse_from_rfc3339(last_activity) {
                    println!("{}: {}", "Last Activity".dimmed(), datetime.format("%Y-%m-%d %H:%M:%S"));
                }
            }
            
            if let Some(cmd) = data.get("command_generated") {
                if !cmd.is_null() {
                    if let Some(name) = cmd.get("name").and_then(|v| v.as_str()) {
                        println!("{}: {} {}", "Command Generated".dimmed(), "✨".bright_green(), name.bright_white());
                    }
                }
            }
            
            println!("\n{}", "Conversation:".bright_cyan().bold());
            
            if let Some(messages) = data.get("messages").and_then(|v| v.as_array()) {
                for (i, msg) in messages.iter().enumerate() {
                    if i > 0 {
                        println!();
                    }
                    
                    let role = msg.get("role").and_then(|v| v.as_str()).unwrap_or("unknown");
                    let content = msg.get("content").and_then(|v| v.as_str()).unwrap_or("");
                    let timestamp = msg.get("timestamp").and_then(|v| v.as_str()).unwrap_or("");
                    
                    // Format timestamp
                    let time_str = if let Ok(datetime) = DateTime::parse_from_rfc3339(timestamp) {
                        datetime.format("%H:%M:%S").to_string()
                    } else {
                        String::new()
                    };
                    
                    // Get agent name from session data
                    let agent_name = data.get("agent").and_then(|v| v.as_str()).unwrap_or("Assistant");
                    
                    match role {
                        "user" => {
                            println!("{} {} {}", "→".bright_green(), "User".bright_green().bold(), time_str.dimmed());
                            println!("  {}", content.bright_white());
                        }
                        "assistant" => {
                            println!("{} {} {}", "←".bright_blue(), agent_name.bright_blue().bold(), time_str.dimmed());
                            // Handle multiline assistant responses
                            for line in content.lines() {
                                println!("  {}", line);
                            }
                        }
                        _ => {
                            println!("{} {} {}", "•".dimmed(), role.dimmed(), time_str.dimmed());
                            println!("  {}", content.dimmed());
                        }
                    }
                }
            } else {
                println!("{}", "  No messages found".dimmed());
            }
        }
    } else {
        println!("{}", "❌ Failed to retrieve session".red());
        if let Some(error) = response.error {
            println!("  {}", error.dimmed());
        }
    }
    
    Ok(())
}