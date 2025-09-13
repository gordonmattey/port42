use thiserror::Error;

#[derive(Error, Debug)]
pub enum Port42Error {
    #[error("Daemon error: {0}")]
    Daemon(String),
    
    #[error("Claude API error: {0}")]
    ClaudeApi(String),
    
    #[error("API key error: {0}")]
    ApiKey(String),
    
    #[error("Network error: {0}")]
    Network(String),
    
    #[error("External service error: {0}")]
    ExternalService(String),
}

