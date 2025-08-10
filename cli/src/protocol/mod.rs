use serde::{Deserialize, Serialize};
use anyhow::Result;

// Base request that all commands use
#[derive(Debug, Serialize)]
pub struct DaemonRequest {
    #[serde(rename = "type")]
    pub request_type: String,
    pub id: String,
    pub payload: serde_json::Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub references: Option<Vec<crate::protocol::relations::Reference>>,
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
pub mod status;
pub mod reality;
pub mod memory;
pub mod filesystem;
pub mod file_ops;
pub mod search;
pub mod relations;

pub use possess::*;
pub use status::*;
pub use reality::*;
pub use memory::*;
pub use filesystem::*;
pub use file_ops::*;
pub use search::*;
pub use relations::*;