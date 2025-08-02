use anyhow::{Result, Context, bail};
use colored::*;
use serde_json::json;
use crate::client::DaemonClient;
use crate::types::Request;
use chrono::{DateTime, Utc, Local};

pub fn handle_info(client: &mut DaemonClient, path: String) -> Result<()> {
    // Create request
    let request = Request {
        request_type: "get_metadata".to_string(),
        id: format!("info-{}", chrono::Utc::now().timestamp()),
        payload: json!({
            "path": path
        }),
    };
    
    // Send request and get response
    let response = client.request(request)
        .context("Failed to get metadata")?;
    
    if !response.success {
        bail!("Failed to get info for {}: {}", path, 
            response.error.unwrap_or_else(|| "Unknown error".to_string()));
    }
    
    // Extract data
    let data = response.data.context("No data in response")?;
    
    // Display formatted metadata
    display_metadata(&path, &data)?;
    
    Ok(())
}

fn display_metadata(path: &str, data: &serde_json::Value) -> Result<()> {
    // Header
    println!("{}", "╔══════════════════════════════════════════════════════════════════╗".dimmed());
    println!("{} {}", "Path:".bright_blue().bold(), path.bright_white());
    
    // Basic info
    if let Some(obj_type) = data["type"].as_str() {
        println!("{} {}", "Type:".bright_blue().bold(), obj_type.yellow());
    }
    
    if let Some(obj_id) = data["object_id"].as_str() {
        let short_id = if obj_id.len() > 12 {
            format!("{}...", &obj_id[..12])
        } else {
            obj_id.to_string()
        };
        println!("{} {}", "Object ID:".bright_blue().bold(), short_id.dimmed());
    }
    
    println!("{}", "╚══════════════════════════════════════════════════════════════════╝".dimmed());
    
    // Metadata section
    println!("\n{}", "Metadata:".bright_green().bold());
    
    // Dates
    if let Some(created) = data["created"].as_str() {
        if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
            let local: DateTime<Local> = dt.into();
            println!("  {} {}", "Created:".cyan(), local.format("%Y-%m-%d %H:%M:%S").to_string());
            
            // Show age
            if let Some(age_secs) = data["age_seconds"].as_f64() {
                let age = format_duration(age_secs);
                println!("  {} {} ago", "Age:".cyan(), age.dimmed());
            }
        }
    }
    
    if let Some(modified) = data["modified"].as_str() {
        if let Ok(dt) = DateTime::parse_from_rfc3339(modified) {
            let local: DateTime<Local> = dt.into();
            println!("  {} {}", "Modified:".cyan(), local.format("%Y-%m-%d %H:%M:%S").to_string());
            
            // Show time since modified
            if let Some(mod_secs) = data["modified_seconds"].as_f64() {
                if mod_secs < 86400.0 { // Less than a day
                    let duration = format_duration(mod_secs);
                    println!("  {} {} ago", "Last Modified:".cyan(), duration.dimmed());
                }
            }
        }
    }
    
    if let Some(accessed) = data["accessed"].as_str() {
        if let Ok(dt) = DateTime::parse_from_rfc3339(accessed) {
            let local: DateTime<Local> = dt.into();
            println!("  {} {}", "Last Access:".cyan(), local.format("%Y-%m-%d %H:%M:%S").to_string());
        }
    }
    
    // Size
    if let Some(size) = data["size"].as_i64() {
        println!("  {} {} ({})", "Size:".cyan(), format_size(size), size);
    }
    
    // Description and title
    if let Some(title) = data["title"].as_str() {
        if !title.is_empty() {
            println!("\n{}", "Description:".bright_green().bold());
            println!("  {}", title.bright_white());
        }
    }
    
    if let Some(desc) = data["description"].as_str() {
        if !desc.is_empty() {
            if data["title"].is_null() {
                println!("\n{}", "Description:".bright_green().bold());
            }
            println!("  {}", desc);
        }
    }
    
    // Properties
    let has_properties = !data["agent"].is_null() || 
                        !data["session"].is_null() || 
                        !data["lifecycle"].is_null() || 
                        !data["importance"].is_null() ||
                        data["usage_count"].as_i64().unwrap_or(0) > 0;
    
    if has_properties {
        println!("\n{}", "Properties:".bright_green().bold());
        
        if let Some(agent) = data["agent"].as_str() {
            if !agent.is_empty() {
                println!("  {} {}", "Agent:".cyan(), agent.bright_cyan());
            }
        }
        
        if let Some(session) = data["session"].as_str() {
            if !session.is_empty() {
                println!("  {} {}", "Session:".cyan(), session);
            }
        }
        
        if let Some(lifecycle) = data["lifecycle"].as_str() {
            if !lifecycle.is_empty() {
                let color = match lifecycle {
                    "active" => lifecycle.bright_green(),
                    "stable" => lifecycle.green(),
                    "archived" => lifecycle.yellow(),
                    "deprecated" => lifecycle.red(),
                    _ => lifecycle.normal()
                };
                println!("  {} {}", "Lifecycle:".cyan(), color);
            }
        }
        
        if let Some(importance) = data["importance"].as_str() {
            if !importance.is_empty() {
                println!("  {} {}", "Importance:".cyan(), importance);
            }
        }
        
        if let Some(usage) = data["usage_count"].as_i64() {
            if usage > 0 {
                println!("  {} {}", "Usage Count:".cyan(), usage);
            }
        }
    }
    
    // Tags
    if let Some(tags) = data["tags"].as_array() {
        if !tags.is_empty() {
            println!("\n{}", "Tags:".bright_green().bold());
            for tag in tags {
                if let Some(tag_str) = tag.as_str() {
                    println!("  • {}", tag_str.bright_yellow());
                }
            }
        }
    }
    
    // Virtual paths
    if let Some(paths) = data["paths"].as_array() {
        if !paths.is_empty() {
            println!("\n{}", "Virtual Paths:".bright_green().bold());
            for path in paths {
                if let Some(path_str) = path.as_str() {
                    println!("  • {}", path_str.bright_blue());
                }
            }
        }
    }
    
    // Active session info
    if let Some(active) = data.get("active_session") {
        println!("\n{}", "Active Session Info:".bright_green().bold());
        if let Some(state) = active["state"].as_str() {
            let state_color = match state {
                "active" => state.bright_green(),
                "idle" => state.yellow(),
                "completed" => state.green(),
                "abandoned" => state.red(),
                _ => state.normal()
            };
            println!("  {} {}", "State:".cyan(), state_color);
        }
        if let Some(count) = active["message_count"].as_i64() {
            println!("  {} {}", "Messages:".cyan(), count);
        }
        if let Some(last) = active["last_activity"].as_str() {
            if let Ok(dt) = DateTime::parse_from_rfc3339(last) {
                let local: DateTime<Local> = dt.into();
                println!("  {} {}", "Last Activity:".cyan(), 
                    local.format("%Y-%m-%d %H:%M:%S").to_string());
            }
        }
    }
    
    // Relationships
    if let Some(rels) = data.get("relationships") {
        if let Some(rel_obj) = rels.as_object() {
            if !rel_obj.is_empty() {
                println!("\n{}", "Relationships:".bright_green().bold());
                for (key, value) in rel_obj {
                    if let Some(val_str) = value.as_str() {
                        if !val_str.is_empty() {
                            println!("  {} {}", format!("{}:", key).cyan(), val_str);
                        }
                    } else if let Some(val_arr) = value.as_array() {
                        if !val_arr.is_empty() {
                            println!("  {}:", key.cyan());
                            for item in val_arr {
                                if let Some(item_str) = item.as_str() {
                                    println!("    • {}", item_str);
                                }
                            }
                        }
                    }
                }
            }
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
        format!("{}{}", size as i64, UNITS[unit_index])
    } else {
        format!("{:.1}{}", size, UNITS[unit_index])
    }
}

fn format_duration(seconds: f64) -> String {
    if seconds < 60.0 {
        format!("{:.0} seconds", seconds)
    } else if seconds < 3600.0 {
        format!("{:.0} minutes", seconds / 60.0)
    } else if seconds < 86400.0 {
        format!("{:.1} hours", seconds / 3600.0)
    } else if seconds < 604800.0 {
        format!("{:.1} days", seconds / 86400.0)
    } else if seconds < 2592000.0 {
        format!("{:.1} weeks", seconds / 604800.0)
    } else if seconds < 31536000.0 {
        format!("{:.1} months", seconds / 2592000.0)
    } else {
        format!("{:.1} years", seconds / 31536000.0)
    }
}