use super::{DaemonRequest, RequestBuilder, ResponseParser};
use crate::display::{Displayable, OutputFormat};
use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::json;
use colored::*;
use chrono::{DateTime, Local};
use base64::{Engine as _, engine::general_purpose};

// Cat request and response types
#[derive(Debug, Serialize)]
pub struct CatRequest {
    pub path: String,
}

impl RequestBuilder for CatRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "read_path".to_string(),
            id,
            payload: json!({
                "path": &self.path
            }),
            references: None,
        })
    }
}

#[derive(Debug, Deserialize)]
pub struct CatResponse {
    pub path: String,
    pub content: String,
    pub metadata: Option<FileMetadata>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct FileMetadata {
    #[serde(rename = "type")]
    pub content_type: String,
    pub description: Option<String>,
    pub created: Option<String>,
    pub agent: Option<String>,
}

impl ResponseParser for CatResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        // Decode base64 content
        let content_b64 = data["content"].as_str()
            .ok_or_else(|| anyhow::anyhow!("Missing content field"))?;
        let content_bytes = general_purpose::STANDARD.decode(content_b64)?;
        let content = String::from_utf8(content_bytes)?;
        
        // Extract metadata if available
        let metadata = data.get("metadata")
            .and_then(|m| serde_json::from_value(m.clone()).ok());
            
        // Get path from data or use empty string
        let path = data.get("path")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string();
            
        Ok(CatResponse {
            path,
            content,
            metadata,
        })
    }
}

impl Displayable for CatResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                // Create a JSON representation with decoded content
                let output = json!({
                    "path": &self.path,
                    "content": &self.content,
                    "metadata": &self.metadata,
                });
                println!("{}", serde_json::to_string_pretty(&output)?);
            }
            OutputFormat::Plain | OutputFormat::Table => {
                // Display based on content type
                match self.metadata.as_ref().map(|m| m.content_type.as_str()) {
                    Some("command") => self.display_command(),
                    Some("session") | Some("memory") => self.display_memory(),
                    Some("document") => self.display_document(),
                    _ => {
                        // Default: just print the content
                        println!("{}", self.content);
                    }
                }
            }
        }
        Ok(())
    }
}

impl CatResponse {
    fn display_command(&self) {
        // Show header
        println!("{}", self.path.bright_blue().bold());
        
        // Show metadata if available
        if let Some(ref meta) = self.metadata {
            if let Some(ref desc) = meta.description {
                println!("{}", format!("# {}", desc).dimmed());
            }
            if let Some(ref created) = meta.created {
                if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
                    println!("{}", format!("# Created: {}", dt.format("%Y-%m-%d %H:%M")).dimmed());
                }
            }
            if let Some(ref agent) = meta.agent {
                println!("{}", format!("# Agent: {}", agent).dimmed());
            }
            println!(); // Empty line
        }
        
        // Display content with basic syntax highlighting
        for line in self.content.lines() {
            if line.starts_with('#') && !line.starts_with("#!") {
                // Comments
                println!("{}", line.dimmed());
            } else if line.starts_with("#!/") {
                // Shebang
                println!("{}", line.yellow());
            } else if line.trim().is_empty() {
                println!();
            } else {
                // Check for common keywords
                let highlighted = highlight_keywords(line);
                println!("{}", highlighted);
            }
        }
    }
    
    fn display_memory(&self) {
        // Parse as JSON if possible
        if let Ok(session_data) = serde_json::from_str::<serde_json::Value>(&self.content) {
            // Display formatted session
            println!("{}", "Memory Thread".bright_blue().bold());
            println!("{}", "─".repeat(50).dimmed());
            
            if let Some(ref meta) = self.metadata {
                if let Some(ref agent) = meta.agent {
                    println!("Agent: {}", agent.cyan());
                }
                if let Some(ref created) = meta.created {
                    if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
                        println!("Started: {}", dt.format("%Y-%m-%d %H:%M").to_string().dimmed());
                    }
                }
            }
            
            // Display messages
            if let Some(messages) = session_data["messages"].as_array() {
                println!("{}", "─".repeat(50).dimmed());
                for msg in messages {
                    let role = msg["role"].as_str().unwrap_or("unknown");
                    let content = msg["content"].as_str().unwrap_or("");
                    
                    match role {
                        "user" => {
                            println!("\n{}", "User:".bright_green().bold());
                            println!("{}", content);
                        }
                        "assistant" => {
                            println!("\n{}", "AI:".bright_cyan().bold());
                            println!("{}", content);
                        }
                        _ => {
                            println!("\n{}: {}", role, content);
                        }
                    }
                }
                println!("{}", "─".repeat(50).dimmed());
            }
        } else {
            // Fallback: just display as text
            println!("{}", self.path.bright_blue().bold());
            println!("{}", self.content);
        }
    }
    
    fn display_document(&self) {
        println!("{}", self.path.bright_blue().bold());
        println!("{}", "─".repeat(50).dimmed());
        println!("{}", self.content);
    }
}

// Info request and response types
#[derive(Debug, Serialize)]
pub struct InfoRequest {
    pub path: String,
}

impl RequestBuilder for InfoRequest {
    fn build_request(&self, id: String) -> Result<DaemonRequest> {
        Ok(DaemonRequest {
            request_type: "get_metadata".to_string(),
            id,
            payload: json!({
                "path": &self.path
            }),
            references: None,
        })
    }
}

#[derive(Debug, Deserialize)]
pub struct InfoResponse {
    pub path: String,
    #[serde(flatten)]
    pub metadata: serde_json::Value,
}

impl ResponseParser for InfoResponse {
    type Output = Self;
    
    fn parse_response(data: &serde_json::Value) -> Result<Self> {
        // Extract path from data or derive from metadata
        let path = data.get("path")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string();
            
        Ok(InfoResponse {
            path,
            metadata: data.clone(),
        })
    }
}

impl Displayable for InfoResponse {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(&self.metadata)?);
            }
            OutputFormat::Plain | OutputFormat::Table => {
                self.display_formatted()?;
            }
        }
        Ok(())
    }
}

impl InfoResponse {
    fn display_formatted(&self) -> Result<()> {
        let data = &self.metadata;
        
        // Header
        println!("{}", "╔══════════════════════════════════════════════════════════════════╗".dimmed());
        println!("{} {}", "Path:".bright_blue().bold(), self.path.bright_white());
        
        // Basic info
        if let Some(obj_type) = data["type"].as_str() {
            println!("{} {}", "Type:".bright_blue().bold(), obj_type.yellow());
        }
        
        if let Some(obj_id) = data["object_id"].as_str() {
            println!("{} {}", "Object ID:".bright_blue().bold(), obj_id.dimmed());
        }
        
        println!("{}", "╚══════════════════════════════════════════════════════════════════╝".dimmed());
        
        // Metadata section
        println!("\n{}", "Metadata:".bright_green().bold());
        
        // Dates
        if let Some(created) = data["created"].as_str() {
            if let Ok(dt) = DateTime::parse_from_rfc3339(created) {
                let local: DateTime<Local> = dt.into();
                println!("  {} {}", "Created:".cyan(), local.format("%Y-%m-%d %H:%M:%S").to_string());
                
                // Show age
                if let Some(age_secs) = data["age_seconds"].as_f64() {
                    let age = format_duration(age_secs);
                    println!("  {} {} ago", "Age:".cyan(), age.dimmed());
                }
            }
        }
        
        if let Some(modified) = data["modified"].as_str() {
            if let Ok(dt) = DateTime::parse_from_rfc3339(modified) {
                let local: DateTime<Local> = dt.into();
                println!("  {} {}", "Modified:".cyan(), local.format("%Y-%m-%d %H:%M:%S").to_string());
            }
        }
        
        // Size
        if let Some(size) = data["size"].as_i64() {
            println!("  {} {} ({})", "Size:".cyan(), format_size(size), size);
        }
        
        // Description
        if let Some(desc) = data["description"].as_str() {
            if !desc.is_empty() {
                println!("\n{}", "Description:".bright_green().bold());
                println!("  {}", desc);
            }
        }
        
        // Properties
        if let Some(agent) = data["agent"].as_str() {
            println!("\n{}", "Properties:".bright_green().bold());
            println!("  {} {}", "Agent:".cyan(), agent.bright_cyan());
        }
        
        // Tags
        if let Some(tags) = data["tags"].as_array() {
            if !tags.is_empty() {
                println!("\n{}", "Tags:".bright_green().bold());
                for tag in tags {
                    if let Some(tag_str) = tag.as_str() {
                        println!("  • {}", tag_str.bright_yellow());
                    }
                }
            }
        }
        
        Ok(())
    }
}

// Helper functions
fn highlight_keywords(line: &str) -> String {
    // Simple keyword highlighting for common shell/programming constructs
    let keywords = vec![
        "if", "then", "else", "elif", "fi", "for", "while", "do", "done",
        "function", "return", "echo", "export", "source", "alias",
        "def", "class", "import", "from", "as", "try", "except", "finally",
        "const", "let", "var", "async", "await", "require", "module",
    ];
    
    let mut result = line.to_string();
    for keyword in keywords {
        let pattern = format!(r"\b{}\b", keyword);
        if let Ok(re) = regex::Regex::new(&pattern) {
            result = re.replace_all(&result, |caps: &regex::Captures| {
                caps[0].bright_magenta().to_string()
            }).to_string();
        }
    }
    
    result
}

fn format_size(bytes: i64) -> String {
    const UNITS: &[&str] = &["B", "K", "M", "G", "T"];
    let mut size = bytes as f64;
    let mut unit_index = 0;
    
    while size >= 1024.0 && unit_index < UNITS.len() - 1 {
        size /= 1024.0;
        unit_index += 1;
    }
    
    if unit_index == 0 {
        format!("{}{}", size as i64, UNITS[unit_index])
    } else {
        format!("{:.1}{}", size, UNITS[unit_index])
    }
}

fn format_duration(seconds: f64) -> String {
    if seconds < 60.0 {
        format!("{:.0} seconds", seconds)
    } else if seconds < 3600.0 {
        format!("{:.0} minutes", seconds / 60.0)
    } else if seconds < 86400.0 {
        format!("{:.1} hours", seconds / 3600.0)
    } else if seconds < 604800.0 {
        format!("{:.1} days", seconds / 86400.0)
    } else if seconds < 2592000.0 {
        format!("{:.1} weeks", seconds / 604800.0)
    } else if seconds < 31536000.0 {
        format!("{:.1} months", seconds / 2592000.0)
    } else {
        format!("{:.1} years", seconds / 31536000.0)
    }
}