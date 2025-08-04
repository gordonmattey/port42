use anyhow::{Result, anyhow};
use colored::*;
use crate::MemoryAction;
use crate::client::DaemonClient;
use crate::protocol::{MemoryListRequest, MemoryDetailRequest, MemoryListResponse, MemoryDetailResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use crate::common::{generate_id, errors::Port42Error};
use crate::help_text;

pub fn handle_memory(port: u16, action: Option<MemoryAction>) -> Result<()> {
    let mut client = DaemonClient::new(port);
    
    match action {
        None => {
            // List all sessions
            let request = MemoryListRequest.build_request(generate_id())?;
            
            // Convert to old-style request for daemon client
            let daemon_request = crate::types::Request {
                id: request.id,
                request_type: request.request_type,
                payload: request.payload,
            };
            
            let response = client.request(daemon_request)?;
            
            if !response.success {
                return Err(Port42Error::Daemon(
                    response.error.unwrap_or_else(|| "Failed to retrieve memory".to_string())
                ).into());
            }
            
            let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
            let memory_list = MemoryListResponse::parse_response(&data)?;
            
            memory_list.display(OutputFormat::Plain)?;
        }
        
        Some(MemoryAction::Search { query, limit: _ }) => {
            println!("{}", help_text::format_searching(&query).blue().bold());
            println!("{}", help_text::ERR_EVOLVE_NOT_READY.yellow());
            println!("{}", "Try: memory  (to list all threads)".dimmed());
            // Could implement by fetching all sessions and filtering
        }
        
        Some(MemoryAction::Show { session_id }) => {
            // Show specific session
            let request = MemoryDetailRequest {
                session_id: session_id.clone(),
            }.build_request(format!("cli-memory-show-{}", session_id))?;
            
            // Convert to old-style request for daemon client
            let daemon_request = crate::types::Request {
                id: request.id,
                request_type: request.request_type,
                payload: request.payload,
            };
            
            let response = client.request(daemon_request)?;
            
            if !response.success {
                println!("{}", help_text::format_error_with_suggestion(
                    help_text::ERR_SESSION_ABANDONED,
                    "This memory thread may have dissolved. Try: memory"
                ));
                if let Some(error) = response.error {
                    println!("  {}", error.dimmed());
                }
                return Ok(());
            }
            
            let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
            let memory_detail = MemoryDetailResponse::parse_response(&data)?;
            
            memory_detail.display(OutputFormat::Plain)?;
        }
    }
    
    Ok(())
}

