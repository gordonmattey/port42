use anyhow::{Result, Context, bail};
use crate::client::DaemonClient;
use crate::help_text::*;
use crate::protocol::{InfoRequest, InfoResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};

pub fn handle_info(client: &mut DaemonClient, path: String) -> Result<()> {
    handle_info_with_format(client, path, OutputFormat::Plain)
}

pub fn handle_info_with_format(client: &mut DaemonClient, path: String, format: OutputFormat) -> Result<()> {
    // Create request
    let request = InfoRequest { path: path.clone() };
    let daemon_request = request.build_request(format!("info-{}", chrono::Utc::now().timestamp()))?;
    
    // Send request and get response
    let response = client.request(daemon_request)
        .context(ERR_CONNECTION_LOST)?;
    
    if !response.success {
        bail!(format_error_with_suggestion(
            ERR_PATH_NOT_FOUND,
            &format!("Cannot inspect essence of '{}'", path)
        ));
    }
    
    // Parse response
    let data = response.data.context(ERR_INVALID_RESPONSE)?;
    let mut info_response = InfoResponse::parse_response(&data)?;
    
    // Set path if not provided by response
    if info_response.path.is_empty() {
        info_response.path = path;
    }
    
    // Display using the displayable trait
    info_response.display(format)?;
    
    Ok(())
}