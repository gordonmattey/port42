use anyhow::{Result, anyhow};
use colored::*;

use crate::client::DaemonClient;
use crate::protocol::{StatusRequest, StatusResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use crate::common::{generate_id, errors::Port42Error};
use crate::help_text;

pub fn handle_status(port: u16, detailed: bool) -> Result<()> {
    println!("{}", help_text::MSG_CHECKING_STATUS.blue().bold());
    
    // Create client
    let mut client = DaemonClient::new(port);
    
    // Build request using protocol types
    let request = StatusRequest.build_request(generate_id())?;
    
    // Convert to old-style request for daemon client (temporary compatibility)
    let daemon_request = crate::types::Request {
        id: request.id,
        request_type: request.request_type,
        payload: request.payload,
    };
    
    // Send to daemon
    match client.request(daemon_request) {
        Ok(response) => {
            if !response.success {
                let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
                return Err(Port42Error::Daemon(error).into());
            }
            
            // Parse response using protocol trait
            let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
            let status_response = StatusResponse::parse_response(&data)?;
            
            // Display using framework
            let format = if detailed {
                // For now, detailed mode uses plain format with more info
                // In the future, we could add a DetailedStatusResponse type
                OutputFormat::Plain
            } else {
                OutputFormat::Plain
            };
            
            status_response.display(format)?;
        }
        Err(e) => {
            // Connection failed - show offline message
            println!("{}", help_text::format_daemon_connection_error(port));
            return Ok(());
        }
    }
    
    Ok(())
}