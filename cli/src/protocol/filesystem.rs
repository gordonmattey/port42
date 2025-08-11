use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat, components};
use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::json;
use colored::*;
use chrono::DateTime;

// Ls request and response types
#[derive(Debug, Serialize)]
pub struct LsRequest {
    pub path: String,
}

impl RequestBuilder for LsRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "list_path".to_string(),
            id,
            payload: json!({
                "path": &self.path
            }),
            references: None,
            session_context: None,
        })
    }
}

#[derive(Debug, Deserialize, Serialize)]
pub struct LsResponse {
    pub path: String,
    pub entries: Vec<FileSystemEntry>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct FileSystemEntry {
    pub name: String,
    #[serde(rename = "type")]
    pub entry_type: String,
    pub size: Option<i64>,
    pub created: Option<String>,
    pub executable: Option<bool>,
    pub state: Option<String>,
    pub messages: Option<i64>,
}

impl ResponseParser for LsResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        let path = data.get("path")
            .and_then(|v| v.as_str())
            .unwrap_or("/")
            .to_string();
            
        let entries = data.get("entries")
            .and_then(|v| v.as_array())
            .map(|arr| {
                arr.iter()
                    .filter_map(|entry| {
                        Some(FileSystemEntry {
                            name: entry.get("name")?.as_str()?.to_string(),
                            entry_type: entry.get("type")
                                .and_then(|v| v.as_str())
                                .unwrap_or("file")
                                .to_string(),
                            size: entry.get("size").and_then(|v| v.as_i64()),
                            created: entry.get("created")
                                .and_then(|v| v.as_str())
                                .map(|s| s.to_string()),
                            executable: entry.get("executable")
                                .and_then(|v| v.as_bool()),
                            state: entry.get("state")
                                .and_then(|v| v.as_str())
                                .map(|s| s.to_string()),
                            messages: entry.get("messages")
                                .and_then(|v| v.as_i64()),
                        })
                    })
                    .collect()
            })
            .unwrap_or_default();
            
        Ok(LsResponse { path, entries })
    }
}

impl Displayable for LsResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                // Display path header
                if self.path != "/" {
                    println!("{}", self.path.bright_blue().bold());
                }
                
                if self.entries.is_empty() {
                    println!("{}", "(empty)".dimmed());
                } else {
                    let mut table = components::TableBuilder::new();
                    
                    // Build headers based on what data we have
                    let has_size = self.entries.iter().any(|e| e.size.is_some());
                    let has_created = self.entries.iter().any(|e| e.created.is_some());
                    let has_messages = self.entries.iter().any(|e| e.messages.is_some());
                    
                    let mut headers = vec!["Name", "Type"];
                    if has_size { headers.push("Size"); }
                    if has_created { headers.push("Created"); }
                    if has_messages { headers.push("Messages"); }
                    
                    table.add_header(headers);
                    
                    // Add rows
                    for entry in &self.entries {
                        let mut row = vec![
                            format_entry_name(entry),
                            entry.entry_type.clone(),
                        ];
                        
                        if has_size {
                            row.push(entry.size
                                .map(format_size)
                                .unwrap_or_else(|| "-".to_string()));
                        }
                        
                        if has_created {
                            row.push(entry.created.as_ref()
                                .and_then(|c| DateTime::parse_from_rfc3339(c).ok())
                                .map(|dt| dt.format("%Y-%m-%d %H:%M").to_string())
                                .unwrap_or_else(|| "-".to_string()));
                        }
                        
                        if has_messages {
                            row.push(entry.messages
                                .map(|m| m.to_string())
                                .unwrap_or_else(|| "-".to_string()));
                        }
                        
                        table.add_row(row);
                    }
                    
                    table.print();
                }
            }
            OutputFormat::Plain => {
                // Display path header
                if self.path != "/" {
                    println!("{}", self.path.bright_blue().bold());
                }
                
                if self.entries.is_empty() {
                    println!("{}", "(empty)".dimmed());
                } else {
                    for entry in &self.entries {
                        print!("{}", format_entry_name_colored(entry, &self.path));
                        
                        // Show additional info if available
                        if let Some(size) = entry.size {
                            print!("  {}", format_size(size).dimmed());
                        }
                        
                        if let Some(ref created) = entry.created {
                            if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
                                print!("  {}", dt.format("%Y-%m-%d %H:%M").to_string().dimmed());
                            }
                        }
                        
                        if entry.entry_type == "directory" {
                            // For memory entries, show state if available
                            if let Some(ref state) = entry.state {
                                print!("  [{}]", state.yellow());
                            }
                            if let Some(msg_count) = entry.messages {
                                print!("  {} messages", msg_count);
                            }
                        }
                        
                        println!();
                    }
                }
            }
        }
        Ok(())
    }
}

// Helper functions
fn format_entry_name(entry: &FileSystemEntry) -> String {
    match entry.entry_type.as_str() {
        "directory" => format!("{}/", entry.name),
        _ => entry.name.clone(),
    }
}

fn format_entry_name_colored(entry: &FileSystemEntry, path: &str) -> ColoredString {
    match entry.entry_type.as_str() {
        "directory" => format!("{}/", entry.name).bright_blue(),
        "file" => {
            // Check if it's a command (executable)
            if path.starts_with("/commands") || entry.executable.unwrap_or(false) {
                entry.name.bright_green()
            } else {
                entry.name.normal()
            }
        },
        _ => entry.name.normal(),
    }
}

fn format_size(bytes: i64) -> String {
    const UNITS: &[&str] = &["B", "K", "M", "G", "T"];
    let mut size = bytes as f64;
    let mut unit_index = 0;
    
    while size >= 1024.0 && unit_index < UNITS.len() - 1 {
        size /= 1024.0;
        unit_index += 1;
    }
    
    if unit_index == 0 {
        format!("{:>4}{}", size as i64, UNITS[unit_index])
    } else {
        format!("{:>4.1}{}", size, UNITS[unit_index])
    }
}