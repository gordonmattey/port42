use anyhow::{Result, Context};
use colored::*;
use serde_json::json;
use crate::client::DaemonClient;
use crate::types::Request;

pub fn handle_ls(client: &mut DaemonClient, path: Option<String>) -> Result<()> {
    // Default to root if no path specified
    let path = path.unwrap_or_else(|| "/".to_string());
    
    // Create request
    let request = Request {
        request_type: "list_path".to_string(),
        id: format!("ls-{}", chrono::Utc::now().timestamp()),
        payload: json!({
            "path": path
        }),
    };
    
    // Send request and get response
    let response = client.request(request)
        .context("Failed to list path")?;
    
    if !response.success {
        anyhow::bail!("Failed to list {}: {}", path, 
            response.error.unwrap_or_else(|| "Unknown error".to_string()));
    }
    
    // Extract data
    let data = response.data.context("No data in response")?;
    let entries = data["entries"].as_array()
        .context("Invalid entries format")?;
    
    // Display path
    if path != "/" {
        println!("{}", path.bright_blue().bold());
    }
    
    // Display entries
    if entries.is_empty() {
        println!("{}", "(empty)".dimmed());
    } else {
        for entry in entries {
            let name = entry["name"].as_str().unwrap_or("?");
            let entry_type = entry["type"].as_str().unwrap_or("file");
            
            match entry_type {
                "directory" => {
                    println!("{}", format!("{}/", name).bright_blue());
                },
                "file" => {
                    // Check if it's a command (executable)
                    if path.starts_with("/commands") || 
                       entry.get("executable").and_then(|v| v.as_bool()).unwrap_or(false) {
                        println!("{}", name.bright_green());
                    } else {
                        println!("{}", name);
                    }
                },
                _ => {
                    println!("{}", name);
                }
            }
            
            // Show additional info if available
            if let Some(size) = entry.get("size").and_then(|v| v.as_i64()) {
                print!("  {}", format_size(size).dimmed());
            }
            
            if let Some(created) = entry.get("created").and_then(|v| v.as_str()) {
                if let Ok(dt) = chrono::DateTime::parse_from_rfc3339(created) {
                    print!("  {}", dt.format("%Y-%m-%d %H:%M").to_string().dimmed());
                }
            }
            
            if entry_type == "directory" {
                // For memory entries, show state if available
                if let Some(state) = entry.get("state").and_then(|v| v.as_str()) {
                    print!("  [{}]", state.yellow());
                }
                if let Some(msg_count) = entry.get("messages").and_then(|v| v.as_i64()) {
                    print!("  {} messages", msg_count);
                }
            }
            
            println!(); // End the line
        }
    }
    
    Ok(())
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