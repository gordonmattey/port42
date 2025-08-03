use anyhow::{Result, Context};
use colored::*;
use serde_json::json;
use crate::client::DaemonClient;
use crate::types::{Request, SearchFilters};
use crate::help_text::*;
use chrono::{DateTime, Local, NaiveDate, TimeZone};

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
    // Build filters
    let mut filters = SearchFilters::default();
    
    if let Some(p) = path {
        filters.path = Some(p);
    }
    
    if let Some(t) = type_filter {
        filters.type_filter = Some(t);
    }
    
    if let Some(a) = after {
        filters.after = Some(parse_date(&a)?);
    }
    
    if let Some(b) = before {
        filters.before = Some(parse_date(&b)?);
    }
    
    if let Some(ag) = agent {
        filters.agent = Some(ag);
    }
    
    if !tags.is_empty() {
        filters.tags = Some(tags);
    }
    
    filters.limit = limit.or(Some(20));
    
    // Send search request
    let request = Request {
        request_type: "search".to_string(),
        id: format!("search-{}", std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_millis()),
        payload: json!({
            "query": query,
            "filters": filters
        }),
    };
    
    let response = client.request(request)
        .context(ERR_CONNECTION_LOST)?;
    
    if !response.success {
        let error = response.error.as_deref().unwrap_or("Connection lost");
        eprintln!("{}", format_error_with_suggestion(
            ERR_CONNECTION_LOST,
            error
        ));
        return Ok(());
    }
    
    // Extract results
    let data = response.data.as_ref()
        .ok_or_else(|| anyhow::anyhow!(ERR_INVALID_RESPONSE))?;
    
    let results = data["results"].as_array()
        .ok_or_else(|| anyhow::anyhow!(ERR_INVALID_RESPONSE))?;
    
    let count = data["count"].as_u64().unwrap_or(0);
    
    // Display results
    if results.is_empty() {
        println!("{}", MSG_NO_RESULTS.dimmed());
        return Ok(());
    }
    
    println!("{}", format_found_results(
        count,
        if count == 1 { "" } else { "s" },
        &query
    ).replace(&count.to_string(), &count.to_string().green().bold().to_string())
     .replace(&query, &query.yellow().to_string()));
    
    if let Some(f) = &filters.path {
        println!("  {} {}", "in:".dimmed(), f.cyan());
    }
    if let Some(t) = &filters.type_filter {
        println!("  {} {}", "type:".dimmed(), t.cyan());
    }
    if let Some(a) = &filters.agent {
        println!("  {} {}", "agent:".dimmed(), a.cyan());
    }
    if let Some(tags) = &filters.tags {
        if !tags.is_empty() {
            println!("  {} {}", "tags:".dimmed(), tags.join(", ").cyan());
        }
    }
    
    println!();
    
    for (idx, result) in results.iter().enumerate() {
        display_search_result(idx + 1, result, &query)?;
        
        // Add separator between results (except last)
        if idx < results.len() - 1 {
            println!();
        }
    }
    
    Ok(())
}

fn display_search_result(index: usize, result: &serde_json::Value, query: &str) -> Result<()> {
    let path = result["path"].as_str().unwrap_or("unknown");
    let obj_type = result["type"].as_str().unwrap_or("unknown");
    let score = result["score"].as_f64().unwrap_or(0.0);
    let snippet = result["snippet"].as_str().unwrap_or("");
    let match_fields = result["match_fields"].as_array()
        .map(|arr| arr.iter()
            .filter_map(|v| v.as_str())
            .collect::<Vec<_>>())
        .unwrap_or_default();
    
    // Type indicator with color
    let type_indicator = match obj_type {
        "session" => "[memory]".blue(),
        "command" => "[command]".green(),
        "artifact" => "[artifact]".yellow(),
        _ => format!("[{}]", obj_type).dimmed(),
    };
    
    // Display result header
    println!("{} {} {} {}",
        format!("{}.", index).dimmed(),
        path.bold(),
        type_indicator,
        format!("(score: {:.2})", score).dimmed()
    );
    
    // Display metadata
    if let Some(metadata) = result.get("metadata") {
        // Created date and agent
        if let Some(created) = metadata["created"].as_str() {
            if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
                let local: DateTime<Local> = dt.into();
                print!("   {} {}", "Created:".dimmed(), local.format("%Y-%m-%d %H:%M").to_string());
                
                if let Some(agent) = metadata["agent"].as_str() {
                    if !agent.is_empty() {
                        print!(" {} {}", "by".dimmed(), agent.cyan());
                    }
                }
                println!();
            }
        }
        
        // Match fields
        if !match_fields.is_empty() {
            println!("   {} {}", 
                "Match in:".dimmed(), 
                match_fields.join(", ").yellow()
            );
        }
    }
    
    // Display snippet with highlighted query
    if !snippet.is_empty() {
        let highlighted = highlight_query(snippet, query);
        println!("   {}", format!("\"{}\"", highlighted).italic());
    }
    
    Ok(())
}

fn highlight_query(text: &str, query: &str) -> String {
    // Case-insensitive highlighting
    let lower_text = text.to_lowercase();
    let lower_query = query.to_lowercase();
    
    if let Some(idx) = lower_text.find(&lower_query) {
        let before = &text[..idx];
        let matched = &text[idx..idx + query.len()];
        let after = &text[idx + query.len()..];
        
        format!("{}{}{}", before, matched.yellow().bold(), after)
    } else {
        text.to_string()
    }
}

fn parse_date(date_str: &str) -> Result<String> {
    // Try parsing as full date-time
    if let Ok(dt) = DateTime::parse_from_rfc3339(date_str) {
        return Ok(dt.to_rfc3339());
    }
    
    // Try parsing as date only (YYYY-MM-DD)
    if let Ok(date) = NaiveDate::parse_from_str(date_str, "%Y-%m-%d") {
        let dt = date.and_hms_opt(0, 0, 0)
            .ok_or_else(|| anyhow::anyhow!(ERR_INVALID_DATE))?;
        let local = Local::now()
            .timezone()
            .from_local_datetime(&dt)
            .single()
            .ok_or_else(|| anyhow::anyhow!(ERR_INVALID_DATE))?;
        return Ok(local.to_rfc3339());
    }
    
    Err(anyhow::anyhow!(format_error_with_suggestion(
        ERR_INVALID_DATE,
        "Examples: 2025-08-02 or 2025-08-02T15:30:00Z"
    )))
}