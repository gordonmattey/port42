use anyhow::Result;
use crate::protocol::status::send_watch_request;

pub fn watch_rules(port: u16) -> Result<()> {
    println!("🔍 Watching rule engine activity...");
    
    match send_watch_request(port, "rules") {
        Ok(watch_data) => {
            // Display current rule status
            if let Some(data) = watch_data.as_array() {
                for item in data {
                    if let (Some(timestamp), Some(rule_name), Some(details)) = (
                        item.get("timestamp").and_then(|v| v.as_str()),
                        item.get("rule_name").and_then(|v| v.as_str()),
                        item.get("details").and_then(|v| v.as_str())
                    ) {
                        println!("⚡ [{}] {}: {}", 
                                format_timestamp(timestamp), 
                                rule_name, 
                                details);
                    }
                }
            } else if let (Some(timestamp), Some(rule_name), Some(details)) = (
                watch_data.get("timestamp").and_then(|v| v.as_str()),
                watch_data.get("rule_name").and_then(|v| v.as_str()),
                watch_data.get("details").and_then(|v| v.as_str())
            ) {
                println!("⚡ [{}] {}: {}", 
                        format_timestamp(timestamp), 
                        rule_name, 
                        details);
            }
        }
        Err(e) => {
            eprintln!("❌ Failed to watch rules: {}", e);
            return Err(e);
        }
    }
    
    Ok(())
}

fn format_timestamp(timestamp: &str) -> String {
    // For now, just show time part
    if let Some(time_part) = timestamp.split('T').nth(1) {
        if let Some(time_only) = time_part.split('.').next() {
            return time_only.to_string();
        }
    }
    timestamp.to_string()
}