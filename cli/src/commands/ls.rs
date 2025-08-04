use anyhow::{Result, Context};
use crate::client::DaemonClient;
use crate::help_text::*;
use crate::protocol::{LsRequest, LsResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};

pub fn handle_ls(client: &mut DaemonClient, path: Option<String>) -> Result<()> {
    handle_ls_with_format(client, path, OutputFormat::Plain)
}

pub fn handle_ls_with_format(client: &mut DaemonClient, path: Option<String>, format: OutputFormat) -> Result<()> {
    // Default to root if no path specified
    let path = path.unwrap_or_else(|| "/".to_string());
    
    // Create request
    let request = LsRequest { path: path.clone() };
    let daemon_request = request.build_request(format!("ls-{}", chrono::Utc::now().timestamp()))?;
    
    // Send request and get response
    let response = client.request(daemon_request.into())
        .context(ERR_CONNECTION_LOST)?;
    
    if !response.success {
        anyhow::bail!(format_error_with_suggestion(
            ERR_PATH_NOT_FOUND,
            &format!("Path '{}' does not exist in reality", path)
        ));
    }
    
    // Parse response
    let data = response.data.context(ERR_INVALID_RESPONSE)?;
    let ls_response = LsResponse::parse_response(&data)?;
    
    // Display using the displayable trait
    ls_response.display(format)?;
    
    Ok(())
}

// Helper to convert DaemonRequest to old Request type
impl From<crate::protocol::DaemonRequest> for crate::types::Request {
    fn from(req: crate::protocol::DaemonRequest) -> Self {
        Self {
            request_type: req.request_type,
            id: req.id,
            payload: req.payload,
        }
    }
}