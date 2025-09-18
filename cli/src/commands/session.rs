use anyhow::{Result, Context, bail};
use colored::*;
use serde_json::Value;
use crate::client::DaemonClient;
use crate::protocol::{LsRequest, InfoRequest, CatRequest, RequestBuilder, ResponseParser, LsResponse, InfoResponse, CatResponse};
use crate::help_text::*;
use chrono::{DateTime, Local};

pub fn handle_session(port: u16, id_prefix: String) -> Result<()> {
    let mut client = DaemonClient::new(port);

    // Create request to list memory sessions
    let ls_request = LsRequest { path: "/memory".to_string() };
    let daemon_request = ls_request.build_request(format!("ls-session-{}", chrono::Utc::now().timestamp()))?;

    // Send request and get response
    let response = client.request(daemon_request)
        .context(ERR_CONNECTION_LOST)?;

    if !response.success {
        bail!("Failed to list memory sessions");
    }

    // Parse the ls response
    let data = response.data.context(ERR_INVALID_RESPONSE)?;
    let ls_response = LsResponse::parse_response(&data)?;

    // Find sessions matching the prefix (any session name format)
    let matching_sessions: Vec<String> = ls_response.entries
        .iter()
        .filter_map(|entry| {
            // Simple prefix matching - works with any session name
            if entry.name.starts_with(&id_prefix) {
                Some(entry.name.clone())
            } else {
                None
            }
        })
        .collect();

    // Handle matches
    match matching_sessions.len() {
        0 => {
            bail!("No session found matching prefix '{}'", id_prefix);
        }
        1 => {
            let session_name = &matching_sessions[0];
            let full_path = format!("/memory/{}", session_name);

            // Get info first
            println!("\n{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".bright_blue());
            println!("{} {}", "📊 Session Info:".bright_cyan(), session_name.bright_yellow());
            println!("{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".bright_blue());

            // Get metadata
            let info_request = InfoRequest { path: full_path.clone() };
            let daemon_request = info_request.build_request(format!("info-session-{}", chrono::Utc::now().timestamp()))?;
            let response = client.request(daemon_request)
                .context(ERR_CONNECTION_LOST)?;

            if response.success {
                if let Some(data) = response.data {
                    let info_response = InfoResponse::parse_response(&data)?;
                    let metadata = &info_response.metadata;

                    // Display key metadata fields
                    if let Some(agent) = metadata.get("agent").and_then(Value::as_str) {
                        println!("  {} {}", "Agent:".bright_cyan(), agent.green());
                    }
                    if let Some(summary) = metadata.get("summary").and_then(Value::as_str) {
                        println!("  {} {}", "Summary:".bright_cyan(), summary);
                    }
                    if let Some(messages) = metadata.get("messageCount").and_then(Value::as_u64) {
                        println!("  {} {}", "Messages:".bright_cyan(), messages.to_string().yellow());
                    }
                    if let Some(created) = metadata.get("createdAt").and_then(Value::as_str) {
                        if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
                            let local: DateTime<Local> = dt.into();
                            println!("  {} {}", "Created:".bright_cyan(), local.format("%Y-%m-%d %H:%M:%S").to_string());
                        }
                    }
                    if let Some(updated) = metadata.get("updatedAt").and_then(Value::as_str) {
                        if let Ok(dt) = DateTime::parse_from_rfc3339(updated) {
                            let local: DateTime<Local> = dt.into();
                            println!("  {} {}", "Updated:".bright_cyan(), local.format("%Y-%m-%d %H:%M:%S").to_string());
                        }
                    }
                    if let Some(size) = metadata.get("size").and_then(Value::as_u64) {
                        let size_kb = size as f64 / 1024.0;
                        println!("  {} {:.1} KB", "Size:".bright_cyan(), size_kb);
                    }
                }
            }

            // Get and display content
            println!("\n{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".bright_blue());
            println!("{} {}", "📝 Session Transcript:".bright_cyan(), session_name.bright_yellow());
            println!("{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".bright_blue());

            // Get content
            let cat_request = CatRequest { path: full_path };
            let daemon_request = cat_request.build_request(format!("cat-session-{}", chrono::Utc::now().timestamp()))?;
            let response = client.request(daemon_request)
                .context(ERR_CONNECTION_LOST)?;

            if !response.success {
                bail!("Failed to read session content");
            }

            let data = response.data.context(ERR_INVALID_RESPONSE)?;
            let cat_response = CatResponse::parse_response(&data)?;

            // Parse and format the session content
            if let Ok(session_data) = serde_json::from_str::<Value>(&cat_response.content) {
                if let Some(messages) = session_data.get("messages").and_then(Value::as_array) {
                    for (i, message) in messages.iter().enumerate() {
                        if i > 0 {
                            println!();  // Add spacing between messages
                        }

                        let role = message.get("role").and_then(Value::as_str).unwrap_or("unknown");
                        let content = message.get("content").and_then(Value::as_str).unwrap_or("");
                        let timestamp = message.get("timestamp").and_then(Value::as_str).unwrap_or("");

                        // Format based on role
                        match role {
                            "user" => {
                                println!("{} {}", "👤 User".bright_green(), format!("[{}]", timestamp).dimmed());
                                for line in content.lines() {
                                    println!("{}", line);
                                }
                            }
                            "assistant" => {
                                println!("{} {}", "🤖 Assistant".bright_blue(), format!("[{}]", timestamp).dimmed());
                                for line in content.lines() {
                                    println!("{}", line);
                                }
                            }
                            _ => {
                                println!("{} {} {}", "💬".dimmed(), role.dimmed(), format!("[{}]", timestamp).dimmed());
                                for line in content.lines() {
                                    println!("{}", line.dimmed());
                                }
                            }
                        }
                    }
                }
            } else {
                // Fallback to raw content if not JSON
                println!("{}", cat_response.content);
            }

            println!("\n{}", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━".bright_blue());
            Ok(())
        }
        _ => {
            // Multiple matches
            println!("⚠️  Multiple sessions match prefix '{}':", id_prefix.yellow());
            for session in &matching_sessions {
                println!("  • {}", session.bright_cyan());
            }
            println!("\nPlease provide a more specific prefix.");
            std::process::exit(1);
        }
    }
}