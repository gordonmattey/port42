use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat, components};
use crate::help_text;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::json;
use colored::*;
use chrono::DateTime;
use std::collections::HashMap;

// Memory request types
#[derive(Debug, Serialize)]
pub struct MemoryListRequest;

#[derive(Debug, Serialize)]
pub struct MemoryDetailRequest {
    pub session_id: String,
}

impl RequestBuilder for MemoryListRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "memory".to_string(),
            id,
            payload: serde_json::Value::Null,
            references: None,
            session_context: None,
        })
    }
}

impl RequestBuilder for MemoryDetailRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "memory".to_string(),
            id,
            payload: json!({
                "session_id": self.session_id
            }),
            references: None,
            session_context: None,
        })
    }
}

// Memory response types
#[derive(Debug, Deserialize, Serialize)]
pub struct MemoryListResponse {
    pub active_sessions: Vec<SessionSummary>,
    pub recent_sessions: Vec<SessionSummary>,
    pub stats: Option<SessionMemoryStats>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct SessionSummary {
    pub id: String,
    pub agent: String,
    pub state: String,
    pub message_count: u64,
    pub command_generated: bool,
    pub date: String,
    pub created_at: Option<String>,
    pub last_activity: Option<String>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct SessionMemoryStats {
    pub total_sessions: u64,
    pub total_size_mb: f64,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct MemoryDetailResponse {
    pub id: String,
    pub agent: String,
    pub state: String,
    pub created_at: String,
    pub last_activity: String,
    pub command_generated: Option<SessionCommandInfo>,
    pub messages: Vec<Message>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct SessionCommandInfo {
    pub name: String,
    pub description: Option<String>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct Message {
    pub role: String,
    pub content: String,
    pub timestamp: String,
}

impl ResponseParser for MemoryListResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        // Parse arrays with default empty vectors
        let active_sessions = data.get("active_sessions")
            .and_then(|v| v.as_array())
            .map(|arr| {
                arr.iter()
                    .filter_map(|v| parse_session_summary(v).ok())
                    .collect()
            })
            .unwrap_or_default();
            
        let recent_sessions = data.get("recent_sessions")
            .and_then(|v| v.as_array())
            .map(|arr| {
                arr.iter()
                    .filter_map(|v| parse_session_summary(v).ok())
                    .collect()
            })
            .unwrap_or_default();
            
        let stats = data.get("stats")
            .and_then(|v| serde_json::from_value(v.clone()).ok());
        
        Ok(MemoryListResponse {
            active_sessions,
            recent_sessions,
            stats,
        })
    }
}

fn parse_session_summary(value: &serde_json::Value) -> Result<SessionSummary> {
    Ok(SessionSummary {
        id: value.get("id")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown")
            .to_string(),
        agent: value.get("agent")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown")
            .to_string(),
        state: value.get("state")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown")
            .to_string(),
        message_count: value.get("message_count")
            .and_then(|v| v.as_u64())
            .unwrap_or(0),
        command_generated: value.get("command_generated")
            .and_then(|v| v.as_bool())
            .unwrap_or(false),
        date: value.get("date")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string(),
        created_at: value.get("created_at")
            .and_then(|v| v.as_str())
            .map(|s| s.to_string()),
        last_activity: value.get("last_activity")
            .and_then(|v| v.as_str())
            .map(|s| s.to_string()),
    })
}

impl ResponseParser for MemoryDetailResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        Ok(MemoryDetailResponse {
            id: data.get("id")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .to_string(),
            agent: data.get("agent")
                .and_then(|v| v.as_str())
                .unwrap_or("unknown")
                .to_string(),
            state: data.get("state")
                .and_then(|v| v.as_str())
                .unwrap_or("unknown")
                .to_string(),
            created_at: data.get("created_at")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .to_string(),
            last_activity: data.get("last_activity")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .to_string(),
            command_generated: data.get("command_generated")
                .and_then(|v| {
                    if v.is_null() {
                        None
                    } else {
                        Some(SessionCommandInfo {
                            name: v.get("name")?.as_str()?.to_string(),
                            description: v.get("description")
                                .and_then(|d| d.as_str())
                                .map(|s| s.to_string()),
                        })
                    }
                }),
            messages: data.get("messages")
                .and_then(|v| v.as_array())
                .map(|arr| {
                    arr.iter()
                        .filter_map(|msg| {
                            Some(Message {
                                role: msg.get("role")?.as_str()?.to_string(),
                                content: msg.get("content")?.as_str()?.to_string(),
                                timestamp: msg.get("timestamp")?.as_str()?.to_string(),
                            })
                        })
                        .collect()
                })
                .unwrap_or_default(),
        })
    }
}

impl Displayable for MemoryListResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                // Active sessions table
                if !self.active_sessions.is_empty() {
                    println!("{}", help_text::MSG_ACTIVE_SESSIONS.bright_green().bold());
                    let mut table = components::TableBuilder::new();
                    table.add_header(vec!["ID", "Agent", "State", "Messages", "Command"]);
                    
                    for session in &self.active_sessions {
                        table.add_row(vec![
                            session.id.clone(),
                            session.agent.clone(),
                            format_state(&session.state),
                            session.message_count.to_string(),
                            if session.command_generated { "✨" } else { "-" }.to_string(),
                        ]);
                    }
                    table.print();
                    println!();
                }
                
                // Recent sessions table
                if !self.recent_sessions.is_empty() {
                    println!("{}", help_text::format_recent_sessions(self.recent_sessions.len()).bright_cyan().bold());
                    // Group by date for display
                    display_sessions_by_date(&self.recent_sessions);
                }
                
                // Stats
                if let Some(stats) = &self.stats {
                    println!("\n{}", "Statistics:".dimmed());
                    println!("  Total sessions: {}", stats.total_sessions);
                    println!("  Storage used: {:.1} MB", stats.total_size_mb);
                }
            }
            OutputFormat::Plain => {
                println!("{}", help_text::MSG_MEMORY_HEADER.blue().bold());
                println!();
                
                // Active sessions
                if !self.active_sessions.is_empty() {
                    println!("{}", help_text::MSG_ACTIVE_SESSIONS.bright_green().bold());
                    for session in &self.active_sessions {
                        print_session_summary(&session);
                    }
                    println!();
                }
                
                // Recent sessions
                if !self.recent_sessions.is_empty() {
                    println!("{}", help_text::format_recent_sessions(self.recent_sessions.len()).bright_cyan().bold());
                    display_sessions_by_date(&self.recent_sessions);
                }
                
                // Stats
                if let Some(stats) = &self.stats {
                    println!("\n{}", "Statistics:".dimmed());
                    println!("  Total sessions: {}", stats.total_sessions);
                    println!("  Storage used: {:.1} MB", stats.total_size_mb);
                }
            }
        }
        Ok(())
    }
}

impl Displayable for MemoryDetailResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            _ => {
                println!("{}", format!("📖 Session: {}", self.id).blue().bold());
                println!();
                
                println!("{}: {}", "Agent".dimmed(), self.agent.bright_blue());
                println!("{}: {}", "State".dimmed(), format_state_colored(&self.state));
                
                if let Ok(datetime) = DateTime::parse_from_rfc3339(&self.created_at) {
                    println!("{}: {}", "Created".dimmed(), datetime.format("%Y-%m-%d %H:%M:%S"));
                }
                
                if let Ok(datetime) = DateTime::parse_from_rfc3339(&self.last_activity) {
                    println!("{}: {}", "Last Activity".dimmed(), datetime.format("%Y-%m-%d %H:%M:%S"));
                }
                
                if let Some(cmd) = &self.command_generated {
                    println!("{}: {} {}", "Command Generated".dimmed(), "✨".bright_green(), cmd.name.bright_white());
                }
                
                println!("\n{}", "Conversation:".bright_cyan().bold());
                
                for (i, msg) in self.messages.iter().enumerate() {
                    if i > 0 {
                        println!();
                    }
                    
                    let time_str = if let Ok(datetime) = DateTime::parse_from_rfc3339(&msg.timestamp) {
                        datetime.format("%H:%M:%S").to_string()
                    } else {
                        String::new()
                    };
                    
                    match msg.role.as_str() {
                        "user" => {
                            println!("{} {} {}", "→".bright_green(), "User".bright_green().bold(), time_str.dimmed());
                            println!("  {}", msg.content.bright_white());
                        }
                        "assistant" => {
                            println!("{} {} {}", "←".bright_blue(), self.agent.bright_blue().bold(), time_str.dimmed());
                            for line in msg.content.lines() {
                                println!("  {}", line);
                            }
                        }
                        _ => {
                            println!("{} {} {}", "•".dimmed(), msg.role.dimmed(), time_str.dimmed());
                            println!("  {}", msg.content.dimmed());
                        }
                    }
                }
            }
        }
        Ok(())
    }
}

// Helper functions
fn format_state(state: &str) -> String {
    match state {
        "active" => "Active".to_string(),
        "idle" => "Idle".to_string(),
        "completed" => "Completed".to_string(),
        "abandoned" => "Dissolved".to_string(),
        _ => state.to_string(),
    }
}

fn format_state_colored(state: &str) -> ColoredString {
    match state {
        "active" => "🟢 Active".green(),
        "idle" => "🟡 Idle".yellow(),
        "completed" => "✅ Completed".bright_green(),
        "abandoned" => "🌑 Dissolved".red(),
        _ => state.normal(),
    }
}

fn print_session_summary(session: &SessionSummary) {
    let state_icon = match session.state.as_str() {
        "active" => "🟢",
        "idle" => "🟡",
        "completed" => "✅",
        "abandoned" => "❌",
        _ => "❓",
    };
    
    print!("    {} {} ", state_icon, session.id.bright_white());
    print!("({}) ", session.agent.bright_blue());
    print!("{} messages", session.message_count);
    
    if session.command_generated {
        print!(" {}", "✨ command".bright_green());
    }
    
    println!();
}

fn display_sessions_by_date(sessions: &[SessionSummary]) {
    let mut by_date: HashMap<String, Vec<&SessionSummary>> = HashMap::new();
    
    for session in sessions {
        by_date.entry(session.date.clone())
            .or_insert_with(Vec::new)
            .push(session);
    }
    
    let mut dates: Vec<_> = by_date.keys().cloned().collect();
    dates.sort_by(|a, b| b.cmp(a));
    
    for date in dates.iter().take(7) {
        println!("\n{}", format!("  📅 {}", date).yellow());
        
        if let Some(sessions) = by_date.get(date) {
            for session in sessions {
                print_session_summary(session);
            }
        }
    }
}