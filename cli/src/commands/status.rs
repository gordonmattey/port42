use anyhow::{Result, anyhow};
use colored::*;

use crate::client::DaemonClient;
use crate::protocol::{StatusRequest, StatusResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use crate::common::{generate_id, errors::Port42Error};
use crate::help_text;

pub fn handle_status(port: u16, detailed: bool) -> Result<()> {
    let mut client = DaemonClient::new(port);
    handle_status_with_format(&mut client, detailed, OutputFormat::Plain)
}

pub fn handle_status_with_format(client: &mut DaemonClient, _detailed: bool, format: OutputFormat) -> Result<()> {
    if format != OutputFormat::Json {
        println!("{}", help_text::MSG_CHECKING_STATUS.blue().bold());
    }
    
    // Build request using protocol types
    let request = StatusRequest.build_request(generate_id())?;
    
    // Send to daemon
    match client.request(request) {
        Ok(response) => {
            if !response.success {
                let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
                return Err(Port42Error::Daemon(error).into());
            }
            
            // Parse response using protocol trait
            let data = response.data.ok_or_else(|| anyhow!("No data in response"))?;
            let status_response = StatusResponse::parse_response(&data)?;
            
            // Display using framework
            status_response.display(format)?;
        }
        Err(e) => {
            if format == OutputFormat::Json {
                // For JSON, output an offline status
                println!(r#"{{"status":"offline","port":{},"error":"Connection failed"}}"#, client.port());
            } else {
                // Connection failed - show offline message
                println!("{}", help_text::format_daemon_connection_error(client.port()));
            }
            // Return error so exit code is non-zero (important for scripts checking status)
            return Err(anyhow!("Daemon not running: {}", e));
        }
    }
    
    Ok(())
}