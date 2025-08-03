use anyhow::{Result, Context, bail};
use colored::*;
use serde_json::json;
use crate::client::DaemonClient;
use crate::types::Request;
use crate::help_text::*;
use base64::{Engine as _, engine::general_purpose};

pub fn handle_cat(client: &mut DaemonClient, path: String) -> Result<()> {
    // Create request
    let request = Request {
        request_type: "read_path".to_string(),
        id: format!("cat-{}", chrono::Utc::now().timestamp()),
        payload: json!({
            "path": path
        }),
    };
    
    // Send request and get response
    let response = client.request(request)
        .context(ERR_CONNECTION_LOST)?;
    
    if !response.success {
        bail!(format_error_with_suggestion(
            ERR_PATH_NOT_FOUND,
            &format!("Reality fragment '{}' cannot be accessed", path)
        ));
    }
    
    // Extract data
    let data = response.data.context(ERR_INVALID_RESPONSE)?;
    
    // Decode content
    let content_b64 = data["content"].as_str()
        .context(ERR_INVALID_RESPONSE)?;
    let content_bytes = general_purpose::STANDARD.decode(content_b64)
        .context("Reality encoding corrupted")?;
    let content = String::from_utf8(content_bytes)
        .context("Reality fragment contains non-textual essence")?;
    
    // Get metadata if available
    let metadata = data.get("metadata");
    let content_type = metadata
        .and_then(|m| m["type"].as_str())
        .unwrap_or("file");
    
    // Display content based on type
    match content_type {
        "command" => display_command(&path, &content, metadata),
        "session" | "memory" => display_memory(&path, &content, metadata),
        "document" => display_document(&path, &content),
        _ => {
            // Default: just print the content
            println!("{}", content);
        }
    }
    
    Ok(())
}

fn display_command(path: &str, content: &str, metadata: Option<&serde_json::Value>) {
    // Show header
    println!("{}", path.bright_blue().bold());
    
    // Show metadata if available
    if let Some(meta) = metadata {
        if let Some(desc) = meta["description"].as_str() {
            println!("{}", format!("# {}", desc).dimmed());
        }
        if let Some(created) = meta["created"].as_str() {
            if let Ok(dt) = chrono::DateTime::parse_from_rfc3339(created) {
                println!("{}", format!("# Created: {}", dt.format("%Y-%m-%d %H:%M")).dimmed());
            }
        }
        if let Some(agent) = meta["agent"].as_str() {
            println!("{}", format!("# Agent: {}", agent).dimmed());
        }
        println!(); // Empty line
    }
    
    // Display content with basic syntax highlighting
    for line in content.lines() {
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

fn display_memory(path: &str, content: &str, metadata: Option<&serde_json::Value>) {
    // Parse as JSON if possible
    if let Ok(session_data) = serde_json::from_str::<serde_json::Value>(content) {
        // Display formatted session
        println!("{}", "Memory Thread".bright_blue().bold());
        println!("{}", "─".repeat(50).dimmed());
        
        if let Some(meta) = metadata {
            if let Some(agent) = meta["agent"].as_str() {
                println!("Agent: {}", agent.cyan());
            }
            if let Some(created) = meta["created"].as_str() {
                if let Ok(dt) = chrono::DateTime::parse_from_rfc3339(created) {
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
        println!("{}", path.bright_blue().bold());
        println!("{}", content);
    }
}

fn display_document(path: &str, content: &str) {
    println!("{}", path.bright_blue().bold());
    println!("{}", "─".repeat(50).dimmed());
    println!("{}", content);
}

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