use super::*;

/// Trait for formatting context data in different ways
pub trait ContextFormatter {
    fn format(&self, data: &ContextData) -> String;
}

/// JSON formatter (default)
pub struct JsonFormatter;

impl ContextFormatter for JsonFormatter {
    fn format(&self, data: &ContextData) -> String {
        serde_json::to_string_pretty(data).unwrap_or_else(|_| "{}".to_string())
    }
}

/// Pretty formatter for human reading
pub struct PrettyFormatter;

impl ContextFormatter for PrettyFormatter {
    fn format(&self, data: &ContextData) -> String {
        let mut output = String::new();
        
        if let Some(session) = &data.active_session {
            output.push_str(&format!("🔄 Active: {} session ({} messages)\n", 
                session.agent, session.message_count));
            output.push_str(&format!("   Session ID: {}\n", session.id));
            output.push_str(&format!("   Started: {}\n", session.start_time));
            output.push_str(&format!("   Last activity: {}\n", session.last_activity));
            output.push_str(&format!("   State: {}\n", session.state));
            
            if let Some(tool) = &session.tool_created {
                output.push_str(&format!("   Created tool: {}\n", tool));
            }
        } else {
            output.push_str("No active session\n");
        }
        
        output
    }
}

/// Compact formatter for status lines
pub struct CompactFormatter;

impl ContextFormatter for CompactFormatter {
    fn format(&self, data: &ContextData) -> String {
        let session = data.active_session.as_ref()
            .map(|s| format!("{}[{}]", s.agent, s.message_count))
            .unwrap_or_else(|| "no session".to_string());
        
        format!("{} | tools: {}", session, data.created_tools.len())
    }
}

/// Watch formatter with ASCII boxes
pub struct WatchFormatter;

impl ContextFormatter for WatchFormatter {
    fn format(&self, data: &ContextData) -> String {
        // Placeholder - will be implemented in Step 3
        format!("Watch mode: {:?}", data.active_session.is_some())
    }
}