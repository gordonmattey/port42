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
        
        // Show recent commands (more of them for activity summary)
        if !data.recent_commands.is_empty() {
            output.push_str("\n📝 Recent Activity:\n");
            for cmd in data.recent_commands.iter().take(10) {
                let age = if cmd.age_seconds < 60 {
                    format!("{}s ago", cmd.age_seconds)
                } else {
                    format!("{}m ago", cmd.age_seconds / 60)
                };
                output.push_str(&format!("   • {} ({})\n", cmd.command, age));
            }
        }
        
        // Show created tools
        if !data.created_tools.is_empty() {
            output.push_str("\n🛠  Created Tools:\n");
            for tool in &data.created_tools {
                output.push_str(&format!("   • {}\n", tool.name));
            }
        }
        
        // Show accessed memories/artifacts
        if !data.accessed_memories.is_empty() {
            output.push_str("\n📚 Recently Accessed:\n");
            for access in data.accessed_memories.iter().take(5) {
                let icon = match access.access_type.as_str() {
                    "created" => "✨",  // Memory/session created
                    "command" => "🔧",
                    "tool" => "⚙️",
                    "memory" | "session" => "🧠",
                    "info" | "info-command" | "info-tool" | "info-memory" => "ℹ️",
                    "browse" | "browse-commands" | "browse-tools" | "browse-memory" => "👁",
                    _ => "📄",
                };
                let times = if access.access_count > 1 {
                    format!(" ({}x)", access.access_count)
                } else {
                    String::new()
                };
                let display = access.display_name.as_ref().unwrap_or(&access.path);
                output.push_str(&format!("   {} {}{}\n", icon, display, times));
            }
        }
        
        // Show suggestions
        if !data.suggestions.is_empty() {
            output.push_str("\n💡 Suggestions:\n");
            for suggestion in data.suggestions.iter().take(3) {
                output.push_str(&format!("   • {}\n", suggestion.command));
            }
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