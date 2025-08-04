use anyhow::{Result, Context};
use crate::client::DaemonClient;
use crate::help_text::*;
use crate::protocol::{SearchRequest, SearchFilters, SearchResponse, RequestBuilder, ResponseParser, parse_date};
use crate::display::{Displayable, OutputFormat};

pub fn handle_search(
    client: &mut DaemonClient,
    query: String,
    path: Option<String>,
    type_filter: Option<String>,
    after: Option<String>,
    before: Option<String>,
    agent: Option<String>,
    tags: Vec<String>,
    limit: Option<usize>,
) -> Result<()> {
    handle_search_with_format(
        client,
        query,
        path,
        type_filter,
        after,
        before,
        agent,
        tags,
        limit,
        OutputFormat::Plain,
    )
}

pub fn handle_search_with_format(
    client: &mut DaemonClient,
    query: String,
    path: Option<String>,
    type_filter: Option<String>,
    after: Option<String>,
    before: Option<String>,
    agent: Option<String>,
    tags: Vec<String>,
    limit: Option<usize>,
    format: OutputFormat,
) -> Result<()> {
    // Build filters
    let mut filters = SearchFilters::default();
    
    filters.path = path;
    filters.type_filter = type_filter;
    
    if let Some(a) = after {
        filters.after = Some(parse_date(&a)?);
    }
    
    if let Some(b) = before {
        filters.before = Some(parse_date(&b)?);
    }
    
    filters.agent = agent;
    
    if !tags.is_empty() {
        filters.tags = Some(tags);
    }
    
    filters.limit = limit.or(Some(20));
    
    // Create request
    let request = SearchRequest::new(query.clone()).with_filters(filters);
    let daemon_request = request.build_request(format!("search-{}", chrono::Utc::now().timestamp_millis()))?;
    
    // Send request and get response
    let response = client.request(daemon_request)
        .context(ERR_CONNECTION_LOST)?;
    
    if !response.success {
        let error = response.error.as_deref().unwrap_or("Connection lost");
        eprintln!("{}", format_error_with_suggestion(
            ERR_CONNECTION_LOST,
            error
        ));
        return Ok(());
    }
    
    // Parse response
    let data = response.data.as_ref()
        .ok_or_else(|| anyhow::anyhow!(ERR_INVALID_RESPONSE))?;
    let mut search_response = SearchResponse::parse_response(data)?;
    
    // Ensure query is set (in case response doesn't include it)
    if search_response.query.is_empty() {
        search_response.query = query;
    }
    
    // Display using the displayable trait
    search_response.display(format)?;
    
    Ok(())
}