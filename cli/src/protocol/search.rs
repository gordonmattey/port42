use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat, components};
use crate::help_text;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::json;
use colored::*;
use chrono::{DateTime, Local, NaiveDate, TimeZone};

// Search request types
#[derive(Debug, Serialize, Deserialize, Default)]
pub struct SearchFilters {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub path: Option<String>,
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    pub type_filter: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub after: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub before: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agent: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub limit: Option<usize>,
}

#[derive(Debug, Serialize)]
pub struct SearchRequest {
    pub query: String,
    pub filters: SearchFilters,
}

impl SearchRequest {
    pub fn new(query: String) -> Self {
        Self {
            query,
            filters: SearchFilters::default(),
        }
    }
    
    pub fn with_filters(mut self, filters: SearchFilters) -> Self {
        self.filters = filters;
        self
    }
}

impl RequestBuilder for SearchRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "search".to_string(),
            id,
            payload: json!({
                "query": &self.query,
                "filters": &self.filters
            }),
            references: None,
            session_context: None,
            user_prompt: None,
        })
    }
}

// Search response types
#[derive(Debug, Deserialize, Serialize)]
pub struct SearchResponse {
    pub query: String,
    pub count: u64,
    pub results: Vec<SearchResult>,
    pub filters: Option<SearchFilters>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct SearchResult {
    pub path: String,
    #[serde(rename = "type")]
    pub result_type: String,
    pub score: f64,
    pub snippet: Option<String>,
    pub match_fields: Vec<String>,
    pub metadata: Option<SearchMetadata>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct SearchMetadata {
    pub created: Option<String>,
    pub agent: Option<String>,
    pub title: Option<String>,
    pub description: Option<String>,
}

impl ResponseParser for SearchResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        let results = data["results"].as_array()
            .ok_or_else(|| anyhow::anyhow!("Missing results array"))?
            .iter()
            .filter_map(|r| serde_json::from_value(r.clone()).ok())
            .collect();
            
        let query = data.get("query")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string();
            
        let count = data["count"].as_u64().unwrap_or(0);
        
        let filters = data.get("filters")
            .and_then(|f| serde_json::from_value(f.clone()).ok());
            
        Ok(SearchResponse {
            query,
            count,
            results,
            filters,
        })
    }
}

impl Displayable for SearchResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                self.display_table()?;
            }
            OutputFormat::Plain => {
                self.display_plain()?;
            }
        }
        Ok(())
    }
}

impl SearchResponse {
    fn display_plain(&self) -> Result<()> {
        if self.results.is_empty() {
            println!("{}", help_text::MSG_NO_RESULTS.dimmed());
            return Ok(());
        }
        
        // Display header
        println!("{}", help_text::format_found_results(
            self.count,
            if self.count == 1 { "" } else { "s" },
            &self.query
        ).replace(&self.count.to_string(), &self.count.to_string().green().bold().to_string())
         .replace(&self.query, &self.query.yellow().to_string()));
        
        // Display active filters
        if let Some(ref filters) = self.filters {
            if let Some(ref path) = filters.path {
                println!("  {} {}", "in:".dimmed(), path.cyan());
            }
            if let Some(ref type_filter) = filters.type_filter {
                println!("  {} {}", "type:".dimmed(), type_filter.cyan());
            }
            if let Some(ref agent) = filters.agent {
                println!("  {} {}", "agent:".dimmed(), agent.cyan());
            }
            if let Some(ref tags) = filters.tags {
                if !tags.is_empty() {
                    println!("  {} {}", "tags:".dimmed(), tags.join(", ").cyan());
                }
            }
        }
        
        println!();
        
        // Display results
        for (idx, result) in self.results.iter().enumerate() {
            self.display_search_result(idx + 1, result)?;
            
            // Add separator between results (except last)
            if idx < self.results.len() - 1 {
                println!();
            }
        }
        
        Ok(())
    }
    
    fn display_table(&self) -> Result<()> {
        if self.results.is_empty() {
            println!("{}", help_text::MSG_NO_RESULTS.dimmed());
            return Ok(());
        }
        
        println!("{} results for '{}'", self.count, self.query.yellow());
        println!();
        
        let mut table = components::TableBuilder::new();
        table.add_header(vec!["Path", "Type", "Score", "Created", "Match In"]);
        
        for result in &self.results {
            let created = result.metadata.as_ref()
                .and_then(|m| m.created.as_ref())
                .and_then(|c| DateTime::parse_from_rfc3339(c).ok())
                .map(|dt| dt.format("%Y-%m-%d").to_string())
                .unwrap_or_else(|| "-".to_string());
                
            let match_fields = if result.match_fields.is_empty() {
                "-".to_string()
            } else {
                result.match_fields.join(", ")
            };
            
            table.add_row(vec![
                result.path.clone(),
                result.result_type.clone(),
                format!("{:.2}", result.score),
                created,
                match_fields,
            ]);
        }
        
        table.print();
        Ok(())
    }
    
    fn display_search_result(&self, index: usize, result: &SearchResult) -> Result<()> {
        // Type indicator with color
        let type_indicator = match result.result_type.as_str() {
            "session" => "[memory]".blue(),
            "command" => "[command]".green(),
            "artifact" => "[artifact]".yellow(),
            _ => format!("[{}]", result.result_type).dimmed(),
        };
        
        // Display result header
        println!("{} {} {} {}",
            format!("{}.", index).dimmed(),
            result.path.bold(),
            type_indicator,
            format!("(score: {:.2})", result.score).dimmed()
        );
        
        // Display metadata
        if let Some(ref metadata) = result.metadata {
            // Created date and agent
            if let Some(ref created) = metadata.created {
                if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
                    let local: DateTime<Local> = dt.into();
                    print!("   {} {}", "Created:".dimmed(), local.format("%Y-%m-%d %H:%M").to_string());
                    
                    if let Some(ref agent) = metadata.agent {
                        if !agent.is_empty() {
                            print!(" {} {}", "by".dimmed(), agent.cyan());
                        }
                    }
                    println!();
                }
            }
            
            // Match fields
            if !result.match_fields.is_empty() {
                println!("   {} {}", 
                    "Match in:".dimmed(), 
                    result.match_fields.join(", ").yellow()
                );
            }
        }
        
        // Display snippet with highlighted query
        if let Some(ref snippet) = result.snippet {
            if !snippet.is_empty() {
                let highlighted = highlight_query(snippet, &self.query);
                println!("   {}", format!("\"{}\"", highlighted).italic());
            }
        }
        
        Ok(())
    }
}

// Helper functions
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

pub fn parse_date(date_str: &str) -> Result<String> {
    // Try parsing as full date-time
    if let Ok(dt) = DateTime::parse_from_rfc3339(date_str) {
        return Ok(dt.to_rfc3339());
    }
    
    // Try parsing as date only (YYYY-MM-DD)
    if let Ok(date) = NaiveDate::parse_from_str(date_str, "%Y-%m-%d") {
        let dt = date.and_hms_opt(0, 0, 0)
            .ok_or_else(|| anyhow::anyhow!(help_text::ERR_INVALID_DATE))?;
        let local = Local::now()
            .timezone()
            .from_local_datetime(&dt)
            .single()
            .ok_or_else(|| anyhow::anyhow!(help_text::ERR_INVALID_DATE))?;
        return Ok(local.to_rfc3339());
    }
    
    Err(anyhow::anyhow!(help_text::format_error_with_suggestion(
        help_text::ERR_INVALID_DATE,
        "Examples: 2025-08-02 or 2025-08-02T15:30:00Z"
    )))
}