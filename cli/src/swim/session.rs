use crate::client::DaemonClient;
use crate::swim::display::SwimDisplay;
use crate::swim::{SimpleDisplay, AnimatedDisplay};
use crate::protocol::{RequestBuilder, ResponseParser, swim::{SwimRequest, SwimResponse, ApprovalResponse}};
use crate::common::{generate_id, errors::Port42Error};
use crate::display::{OutputFormat, Displayable};
use crate::ui::WaveSpinner;
use anyhow::{Result, anyhow};
use std::time::{SystemTime, UNIX_EPOCH};
use std::io::{self, Write};
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
    
    pub fn send_message_with_context(&mut self, session_id: &str, agent: &str, message: &str, memory_context: Option<Vec<String>>, references: Option<Vec<crate::protocol::relations::Reference>>) -> Result<SwimResponse> {
        // Build request using protocol traits
        let swim_req = SwimRequest {
            agent: agent.to_string(),
            message: message.to_string(),
            memory_context,
            references,
            approval_response: None,
        };
        
        let request_id = generate_id();
        let mut request = swim_req.build_request(request_id)?;
        
        // Add session_id to payload
        if let Some(obj) = request.payload.as_object_mut() {
            obj.insert("session_id".to_string(), serde_json::Value::String(session_id.to_string()));
        }
        
        // Show wave spinner while waiting for response
        let mut spinner = WaveSpinner::new();
        let response = self.client.request(request)?;
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
        let mut swim_response = SwimResponse::parse_response(&data)?;
        
        // Check if approval is needed
        if let Some(approval_req) = &swim_response.approval_needed {
            // Format the command for display
            let cmd_display = format!("bash -c \"{}\"", approval_req.args.join(" "));
            
            // Show approval prompt
            println!("\n{}", "=".repeat(60).bright_black());
            println!("{} {}", "🔒".bright_yellow(), "AI REQUESTS BASH ACCESS".bold());
            println!("{}", "-".repeat(60).bright_black());
            println!("Command: {}", cmd_display.bright_cyan());
            println!("{}", "-".repeat(60).bright_black());
            println!("{} {}", "⚠️".bright_red(), "Bash commands have full system access".yellow());
            println!("{}", "=".repeat(60).bright_black());
            print!("\nApprove? [y/N]: ");
            io::stdout().flush()?;
            
            // Read user input
            let mut input = String::new();
            io::stdin().read_line(&mut input)?;
            let approved = input.trim().to_lowercase() == "y" || input.trim().to_lowercase() == "yes";
            
            if approved {
                println!("{} Bash command approved\n", "✅".green());
            } else {
                println!("{} Bash command denied\n", "❌".red());
            }
            
            // Send approval response
            let approval_response = ApprovalResponse {
                request_id: approval_req.request_id.clone(),
                approved,
            };
            
            // Build new request with approval
            let approval_req = SwimRequest {
                agent: agent.to_string(),
                message: String::new(), // Empty message for continuation
                memory_context: None,
                references: None,
                approval_response: Some(approval_response),
            };
            
            let request_id = generate_id();
            let mut request = approval_req.build_request(request_id)?;
            
            // Add session_id to payload
            if let Some(obj) = request.payload.as_object_mut() {
                obj.insert("session_id".to_string(), serde_json::Value::String(session_id.to_string()));
            }
            
            // Send approval and get new response
            let response = self.client.request(request)?;
            
            if !response.success {
                let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
                self.display.show_error(&error);
                return Err(anyhow!(error));
            }
            
            // Parse the new response
            let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
            swim_response = SwimResponse::parse_response(&data)?;
        }
        
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