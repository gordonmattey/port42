use anyhow::{Result, Context, bail};
use crate::client::DaemonClient;
use crate::help_text::*;
use crate::protocol::{CatRequest, CatResponse, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};

pub fn handle_cat(client: &mut DaemonClient, path: String) -> Result<()> {
    handle_cat_with_format(client, path, OutputFormat::Plain)
}

pub fn handle_cat_with_format(client: &mut DaemonClient, path: String, format: OutputFormat) -> Result<()> {
    // Create request
    let request = CatRequest { path: path.clone() };
    let daemon_request = request.build_request(format!("cat-{}", chrono::Utc::now().timestamp()))?;
    
    // Send request and get response
    let response = client.request(daemon_request)
        .context(ERR_CONNECTION_LOST)?;
    
    if !response.success {
        bail!(format_error_with_suggestion(
            ERR_PATH_NOT_FOUND,
            &format!("Reality fragment '{}' cannot be accessed", path)
        ));
    }
    
    // Parse response
    let data = response.data.context(ERR_INVALID_RESPONSE)?;
    let mut cat_response = CatResponse::parse_response(&data)?;
    
    // Set path if not provided by response
    if cat_response.path.is_empty() {
        cat_response.path = path;
    }
    
    // Display using the displayable trait
    cat_response.display(format)?;
    
    Ok(())
}