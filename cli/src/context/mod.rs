use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

/// Complete context data structure matching daemon's ContextData
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ContextData {
    pub active_session: Option<ActiveSessionInfo>,
    pub recent_commands: Vec<CommandRecord>,
    pub created_tools: Vec<ToolRecord>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub accessed_memories: Vec<MemoryAccess>,
    pub suggestions: Vec<ContextSuggestion>,
}

/// Active session information for display
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ActiveSessionInfo {
    pub id: String,
    pub agent: String,
    pub message_count: i32,
    pub start_time: DateTime<Utc>,
    pub last_activity: DateTime<Utc>,
    pub state: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_created: Option<String>,
}

/// Recently executed command
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CommandRecord {
    pub command: String,
    pub timestamp: DateTime<Utc>,
    pub age_seconds: i32,
    pub exit_code: i32,
}

/// Tool created in current session
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ToolRecord {
    pub name: String,
    #[serde(rename = "type")]
    pub tool_type: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub transforms: Vec<String>,
    pub created_at: DateTime<Utc>,
}

/// Memory or artifact access tracking
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct MemoryAccess {
    pub path: String,
    #[serde(rename = "type")]
    pub access_type: String,
    pub access_count: i32,
}

/// Smart command suggestion
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ContextSuggestion {
    pub command: String,
    pub reason: String,
    pub confidence: f64,
}

// Re-export submodules
pub mod formatters;
pub mod watch;