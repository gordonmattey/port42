use crate::client::DaemonClient;
use crate::possess::display::PossessDisplay;
use crate::possess::{SimpleDisplay, AnimatedDisplay};
use crate::protocol::{RequestBuilder, ResponseParser, possess::{PossessRequest, PossessResponse}};
use crate::common::{generate_id, errors::Port42Error};
use anyhow::{Result, anyhow};
use std::time::{SystemTime, UNIX_EPOCH};

pub struct SessionHandler {
    pub(crate) client: DaemonClient,
    display: Box<dyn PossessDisplay>,
}

impl SessionHandler {
    pub fn new(client: DaemonClient, interactive: bool) -> Self {
        let display: Box<dyn PossessDisplay> = if interactive {
            Box::new(AnimatedDisplay::new())
        } else {
            Box::new(SimpleDisplay::new())
        };
        
        Self { client, display }
    }
    
    pub fn with_display(client: DaemonClient, display: Box<dyn PossessDisplay>) -> Self {
        Self { client, display }
    }
    
    pub fn send_message(&mut self, session_id: &str, agent: &str, message: &str) -> Result<PossessResponse> {
        // Build request using protocol traits
        let possess_req = PossessRequest {
            agent: agent.to_string(),
            message: message.to_string(),
        };
        
        let request_id = generate_id();
        let mut request = possess_req.build_request(request_id)?;
        
        // Add session_id to payload
        if let Some(obj) = request.payload.as_object_mut() {
            obj.insert("session_id".to_string(), serde_json::Value::String(session_id.to_string()));
        }
        
        // Convert to old-style request for daemon client
        let daemon_request = crate::types::Request {
            id: request.id,
            request_type: request.request_type,
            payload: request.payload,
        };
        
        // Send to daemon
        let response = self.client.request(daemon_request)?;
        
        if !response.success {
            let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
            self.display.show_error(&error);
            return Err(Port42Error::Daemon(error).into());
        }
        
        // Parse response using protocol trait
        let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
        let possess_response = PossessResponse::parse_response(&data)?;
        
        // Display results
        self.display.show_ai_message(agent, &possess_response.message);
        
        if let Some(ref spec) = possess_response.command_spec {
            self.display.show_command_created(spec);
        }
        
        if let Some(ref spec) = possess_response.artifact_spec {
            self.display.show_artifact_created(spec);
        }
        
        Ok(possess_response)
    }
    
    pub fn display_session_info(&self, session_id: &str, is_new: bool) {
        self.display.show_session_info(session_id, is_new);
    }
}

/// Determine session ID - either use provided one or generate new
pub fn determine_session_id(session_id: Option<String>) -> (String, bool) {
    match session_id {
        Some(id) => (id, false), // Existing session
        None => {
            // Generate new session ID
            let timestamp = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_millis();
            let id = format!("cli-{}", timestamp);
            (id, true) // New session
        }
    }
}