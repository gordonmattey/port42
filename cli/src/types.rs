use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Serialize, Deserialize)]
pub struct Request {
    #[serde(rename = "type")]
    pub request_type: String,
    pub id: String,
    pub payload: Value,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Response {
    pub id: String,
    pub success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PossessPayload {
    pub agent: String,
    pub message: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Session {
    pub id: String,
    pub agent: String,
    pub created_at: String,
    pub messages: Vec<Message>,
    pub state: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Message {
    pub role: String,
    pub content: String,
    pub timestamp: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CommandInfo {
    pub name: String,
    pub description: String,
    pub created_at: String,
    pub agent: String,
}