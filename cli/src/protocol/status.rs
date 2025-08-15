use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use crate::help_text;
use crate::client::DaemonClient;
use crate::types::Response;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::json;
use colored::*;

#[derive(Debug, Serialize)]
pub struct StatusRequest;

impl RequestBuilder for StatusRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "status".to_string(),
            id,
            payload: serde_json::Value::Null,
            references: None,
            session_context: None,
            user_prompt: None,
        })
    }
}

#[derive(Debug, Deserialize, Serialize)]
pub struct StatusResponse {
    pub port: u64,
    pub uptime: String,
    pub active_sessions: u64,
    pub memory_stats: Option<MemoryStats>,
    pub recent_activity: Option<Vec<RecentActivity>>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct MemoryStats {
    pub total_sessions: u64,
    pub commands_generated: u64,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct RecentActivity {
    pub session_id: String,
    pub agent: String,
    pub timestamp: u64,
}

impl ResponseParser for StatusResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        // Extract fields from the daemon response - handle both string and number
        let port = data.get("port")
            .and_then(|v| {
                // Try as number first, then as string
                v.as_u64().or_else(|| {
                    v.as_str().and_then(|s| s.parse().ok())
                })
            })
            .unwrap_or(42);
            
        let uptime = data.get("uptime")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown")
            .to_string();
            
        let active_sessions = data.get("active_sessions")
            .and_then(|v| v.as_u64())
            .unwrap_or(0);
            
        let memory_stats = data.get("memory_stats")
            .and_then(|v| serde_json::from_value(v.clone()).ok());
            
        let recent_activity = data.get("recent_activity")
            .and_then(|v| serde_json::from_value(v.clone()).ok());
        
        Ok(StatusResponse {
            port,
            uptime,
            active_sessions,
            memory_stats,
            recent_activity,
        })
    }
}

impl Displayable for StatusResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Plain => {
                println!("{}", help_text::MSG_DAEMON_RUNNING.green().bold());
                println!("\n{}", help_text::MSG_CONNECTION_INFO.bright_white());
                println!("{}", help_text::format_port_info(&self.port.to_string().bright_cyan().to_string()));
                println!("{}", help_text::format_uptime_info(&self.uptime.bright_cyan().to_string()));
                println!("{}", help_text::format_sessions_info(&self.active_sessions.to_string().bright_cyan().to_string()));
                
                // Display memory stats if available
                if let Some(ref stats) = self.memory_stats {
                    println!("\n  {}", "Memory Store:".yellow());
                    println!("    Total Sessions: {}", stats.total_sessions.to_string().bright_cyan());
                    println!("    Commands Made:  {}", stats.commands_generated.to_string().bright_cyan());
                }
                
                println!("\n{}", help_text::MSG_DOLPHINS_LISTENING.blue().italic());
            }
            OutputFormat::Table => {
                // Status doesn't really make sense as a table, use plain format
                self.display(OutputFormat::Plain)?;
            }
        }
        Ok(())
    }
}

// Watch request function for real-time monitoring
pub fn send_watch_request(port: u16, target: &str) -> Result<serde_json::Value> {
    let mut client = DaemonClient::new(port);
    
    let payload = json!({
        "target": target
    });
    
    let request = DaemonRequest {
        request_type: "watch".to_string(),
        id: format!("watch-{}", chrono::Utc::now().timestamp_millis()),
        payload,
        references: None,
        session_context: None,
        user_prompt: None,
    };
    
    let response = client.request(request)?;
    
    if !response.success {
        let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
        return Err(anyhow::anyhow!("Watch request failed: {}", error));
    }
    
    let data = response.data.ok_or_else(|| anyhow::anyhow!("No data in watch response"))?;
    Ok(data)
}