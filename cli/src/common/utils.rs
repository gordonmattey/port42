use std::time::{SystemTime, UNIX_EPOCH};

/// Generate a timestamp in milliseconds since UNIX epoch
pub fn timestamp_millis() -> u128 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis()
}

/// Format a timestamp for display
pub fn format_timestamp(millis: u128) -> String {
    let secs = millis / 1000;
    let datetime = UNIX_EPOCH + std::time::Duration::from_secs(secs as u64);
    if let Ok(duration) = SystemTime::now().duration_since(datetime) {
        let secs = duration.as_secs();
        if secs < 60 {
            format!("{} seconds ago", secs)
        } else if secs < 3600 {
            format!("{} minutes ago", secs / 60)
        } else if secs < 86400 {
            format!("{} hours ago", secs / 3600)
        } else {
            format!("{} days ago", secs / 86400)
        }
    } else {
        "just now".to_string()
    }
}

/// Extract session ID from a memory ID if it looks like one
pub fn extract_session_id(arg: &str) -> Option<String> {
    // Better heuristic: memory IDs contain numbers or start with special patterns
    let looks_like_id = arg.len() <= 20 && 
        !arg.contains(' ') && 
        (arg.contains(char::is_numeric) || 
         arg.starts_with("cli-") || 
         arg.contains('-') ||
         arg.contains('_'));
    
    if looks_like_id {
        Some(arg.to_string())
    } else {
        None
    }
}