use anyhow::Result;
use colored::*;
use crate::MemoryAction;
use crate::client::DaemonClient;
use crate::types::Request;
use chrono::{DateTime, Utc};

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
    println!("{}", "🚧 Show session not yet implemented".yellow().dimmed());
    println!("\n{}", "For now, check session files in:".yellow());
    println!("  {}", "~/.port42/memory/sessions/".bright_white());
    Ok(())
}