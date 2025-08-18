use crate::help_text;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum Port42Error {
    #[error("Connection failed: {0}")]
    Connection(String),
    
    #[error("Daemon error: {0}")]
    Daemon(String),
    
    #[error("Parse error: {0}")]
    Parse(String),
    
    #[error("Claude API error: {0}")]
    ClaudeApi(String),
    
    #[error("API key error: {0}")]
    ApiKey(String),
    
    #[error("Network error: {0}")]
    Network(String),
    
    #[error("External service error: {0}")]
    ExternalService(String),
}

impl Port42Error {
    /// Get user-friendly error message using help_text constants
    pub fn user_message(&self) -> String {
        match self {
            Self::Connection(_) => help_text::format_daemon_connection_error(42),
            Self::Daemon(msg) => help_text::format_error_with_help(msg, "possess"),
            _ => self.to_string(),
        }
    }
}