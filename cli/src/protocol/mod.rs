use serde::{Deserialize, Serialize};
use anyhow::Result;

// Base request that all commands use
#[derive(Debug, Serialize)]
pub struct DaemonRequest {
    #[serde(rename = "type")]
    pub request_type: String,
    pub id: String,
    pub payload: serde_json::Value,
}

// Base response from daemon
#[derive(Debug, Deserialize)]
pub struct DaemonResponse {
    pub id: String,
    pub success: bool,
    pub data: Option<serde_json::Value>,
    pub error: Option<String>,
}

// Common trait for request builders
pub trait RequestBuilder {
    fn build_request(&self, id: String) -> Result<DaemonRequest>;
}

// Common trait for response parsers
pub trait ResponseParser {
    type Output;
    fn parse_response(data: &serde_json::Value) -> Result<Self::Output>;
}

// Re-export submodules
pub mod possess;

pub use possess::*;