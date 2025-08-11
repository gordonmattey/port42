pub mod errors;
pub mod utils;

use std::time::{SystemTime, UNIX_EPOCH};

/// Generate unique request ID
pub fn generate_id() -> String {
    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis();
    format!("cli-{}", timestamp)
}

/// Generate CLI session ID for memory-relation bridge
/// CLI sessions represent individual command invocations for tracking tool creation context
pub fn generate_session_id() -> String {
    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis();
    format!("cli-session-{}", timestamp)
}