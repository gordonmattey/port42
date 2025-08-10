use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat, StatusIndicator};
use crate::help_text;
use anyhow::{Result, anyhow};
use serde::{Deserialize, Serialize};
use serde_json::json;
use colored::*;

#[derive(Debug, Serialize)]
pub struct PossessRequest {
    pub agent: String,
    pub message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub memory_context: Option<Vec<String>>,
}

impl RequestBuilder for PossessRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        let mut payload = json!({
            "agent": &self.agent,
            "message": &self.message,
        });
        
        // Add memory context if present
        if let Some(ref context) = self.memory_context {
            payload["memory_context"] = json!(context);
        }
        
        Ok(DaemonRequest {
            request_type: "possess".to_string(),
            id,
            payload,
        })
    }
}

#[derive(Debug, Deserialize, Serialize)]
pub struct PossessResponse {
    pub message: String,
    pub session_id: String,
    pub agent: String,
    #[serde(default)]
    pub command_generated: bool,
    pub command_spec: Option<CommandSpec>,
    #[serde(default)]
    pub artifact_generated: bool,
    pub artifact_spec: Option<ArtifactSpec>,
}

#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct CommandSpec {
    pub name: String,
    pub description: String,
    pub language: String,
}

#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct ArtifactSpec {
    pub name: String,
    #[serde(rename = "type")]
    pub artifact_type: String,
    pub path: String,
    pub description: String,
    pub format: String,
}

impl ResponseParser for PossessResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        // Handle the nested data structure from daemon
        let message = data.get("message")
            .and_then(|v| v.as_str())
            .ok_or_else(|| anyhow!("Missing message in response"))?
            .to_string();
            
        let session_id = data.get("session_id")
            .and_then(|v| v.as_str())
            .ok_or_else(|| anyhow!("Missing session_id in response"))?
            .to_string();
            
        let agent = data.get("agent")
            .and_then(|v| v.as_str())
            .ok_or_else(|| anyhow!("Missing agent in response"))?
            .to_string();
            
        let command_generated = data.get("command_generated")
            .and_then(|v| v.as_bool())
            .unwrap_or(false);
            
        let command_spec = if command_generated {
            data.get("command_spec")
                .and_then(|spec| serde_json::from_value(spec.clone()).ok())
        } else {
            None
        };
        
        let artifact_generated = data.get("artifact_generated")
            .and_then(|v| v.as_bool())
            .unwrap_or(false);
            
        let artifact_spec = if artifact_generated {
            data.get("artifact_spec")
                .and_then(|spec| serde_json::from_value(spec.clone()).ok())
        } else {
            None
        };
        
        Ok(PossessResponse {
            message,
            session_id,
            agent,
            command_generated,
            command_spec,
            artifact_generated,
            artifact_spec,
        })
    }
}

impl Displayable for PossessResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Plain | OutputFormat::Table => {
                // Display AI message
                println!("\n{}", self.agent.bright_blue());
                println!("{}", self.message);
                println!();
                
                // Display command if created
                if let Some(ref spec) = self.command_spec {
                    println!("{} {}", StatusIndicator::success(), help_text::format_command_born(&spec.name).bright_green().bold());
                    println!("{}", "Add to PATH to use:".yellow());
                    println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
                    println!();
                }
                
                // Display artifact if created
                if let Some(ref spec) = self.artifact_spec {
                    println!("{} {}", StatusIndicator::success(), format!("Artifact created: {} ({})", spec.name, spec.artifact_type).bright_cyan().bold());
                    println!("{}", "View with:".yellow());
                    println!("  {}", format!("port42 cat {}", spec.path).bright_white());
                    println!();
                }
            }
        }
        Ok(())
    }
}