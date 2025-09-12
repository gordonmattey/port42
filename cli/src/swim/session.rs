use crate::client::DaemonClient;
use crate::swim::display::SwimDisplay;
use crate::swim::{SimpleDisplay, AnimatedDisplay};
use crate::protocol::{RequestBuilder, ResponseParser, swim::{SwimRequest, SwimResponse}};
use crate::common::{generate_id, errors::Port42Error};
use crate::display::{OutputFormat, Displayable};
use crate::ui::SpinnerGuard;
use anyhow::{Result, anyhow};
use std::time::{SystemTime, UNIX_EPOCH};
use colored::*;

pub struct SessionHandler {
    pub(crate) client: DaemonClient,
    display: Box<dyn SwimDisplay>,
    output_format: OutputFormat,
}

impl SessionHandler {
    pub fn new(client: DaemonClient, interactive: bool) -> Self {
        let display: Box<dyn SwimDisplay> = if interactive {
            Box::new(AnimatedDisplay::new())
        } else {
            Box::new(SimpleDisplay::new())
        };
        
        Self { 
            client, 
            display,
            output_format: OutputFormat::Plain,
        }
    }
    
    pub fn with_display(client: DaemonClient, display: Box<dyn SwimDisplay>) -> Self {
        Self { 
            client, 
            display,
            output_format: OutputFormat::Plain,
        }
    }
    
    pub fn with_output_format(mut self, format: OutputFormat) -> Self {
        self.output_format = format;
        self
    }
    
    pub fn send_message(&mut self, session_id: &str, agent: &str, message: &str) -> Result<SwimResponse> {
        self.send_message_with_context(session_id, agent, message, None, None)
    }
    
    pub fn send_message_with_context(&mut self, session_id: &str, agent: &str, message: &str, memory_context: Option<Vec<String>>, references: Option<Vec<crate::protocol::relations::Reference>>) -> Result<SwimResponse> {
        // Build request using protocol traits
        let swim_req = SwimRequest {
            agent: agent.to_string(),
            message: message.to_string(),
            memory_context,
            references,
        };
        
        let request_id = generate_id();
        let mut request = swim_req.build_request(request_id)?;
        
        // Add session_id to payload
        if let Some(obj) = request.payload.as_object_mut() {
            obj.insert("session_id".to_string(), serde_json::Value::String(session_id.to_string()));
        }
        
        // Show spinner while waiting for AI response
        let spinner = SpinnerGuard::new("Swimming into consciousness stream...");
        
        // Send to daemon
        let response = self.client.request(request)?;
        
        // Stop spinner once we have a response
        spinner.stop();
        
        if !response.success {
            let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
            
            // Classify error and show appropriate message
            let classified_error = classify_error(&error);
            match &classified_error {
                Port42Error::ClaudeApi(_) => {
                    eprintln!("{} Claude API is currently experiencing issues. Please try again in a moment.", "🤖".bright_blue());
                },
                Port42Error::ApiKey(_) => {
                    eprintln!("{} API key issue. Please set PORT42_ANTHROPIC_API_KEY or ANTHROPIC_API_KEY and restart the daemon.", "🔑".bright_yellow());
                },
                Port42Error::Network(_) => {
                    eprintln!("{} Network connection issue. Please check your internet connection.", "🌐".bright_red());
                },
                _ => {
                    self.display.show_error(&error);
                }
            }
            
            return Err(classified_error.into());
        }
        
        // Parse response using protocol trait
        let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
        let swim_response = SwimResponse::parse_response(&data)?;
        
        // Display results based on output format
        match self.output_format {
            OutputFormat::Json => {
                // For JSON, use the Displayable trait
                swim_response.display(OutputFormat::Json)?;
            }
            OutputFormat::Plain | OutputFormat::Table => {
                // For Plain and Table, use the custom display trait for animations in interactive mode
                self.display.show_ai_message(agent, &swim_response.message);
                
                if let Some(ref spec) = swim_response.command_spec {
                    self.display.show_command_created(spec);
                }
                
                if let Some(ref spec) = swim_response.artifact_spec {
                    self.display.show_artifact_created(spec);
                }
            }
        }
        
        Ok(swim_response)
    }
    
    pub fn display_session_info(&self, session_id: &str, is_new: bool) {
        self.display.show_session_info(session_id, is_new);
    }
    
    pub fn display_session_complete(&self, session_id: &str) {
        self.display.show_session_complete(session_id);
    }
}

/// Classify daemon errors by source for better user messaging
fn classify_error(error: &str) -> Port42Error {
    if error.starts_with("CLAUDE_API_ERROR:") {
        let msg = error.strip_prefix("CLAUDE_API_ERROR:").unwrap_or(error).trim();
        Port42Error::ClaudeApi(msg.to_string())
    } else if error.starts_with("API_KEY_ERROR:") {
        let msg = error.strip_prefix("API_KEY_ERROR:").unwrap_or(error).trim();
        Port42Error::ApiKey(msg.to_string())
    } else if error.starts_with("NETWORK_ERROR:") {
        let msg = error.strip_prefix("NETWORK_ERROR:").unwrap_or(error).trim();
        Port42Error::Network(msg.to_string())
    } else if error.starts_with("AI_CONNECTION_ERROR:") {
        let msg = error.strip_prefix("AI_CONNECTION_ERROR:").unwrap_or(error).trim();
        Port42Error::ExternalService(msg.to_string())
    } else {
        // Fallback to daemon error for unclassified errors
        Port42Error::Daemon(error.to_string())
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