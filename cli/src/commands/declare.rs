use anyhow::Result;
use colored::*;
use std::time::Duration;

use crate::client::DaemonClient;
use crate::protocol::{
    DeclareRelationRequest, DeclareRelationResponse, 
    Relation, RequestBuilder, ResponseParser
};
use crate::display::{Displayable, OutputFormat};
use crate::common::{generate_id, references::parse_references};

/// Handle declaring a new tool relation
pub fn handle_declare_tool(port: u16, name: &str, transforms: Vec<String>, references: Option<Vec<String>>, prompt: Option<String>) -> Result<()> {
    println!("{}", format!("🌟 Declaring tool: {}", name).bright_blue());
    
    if !transforms.is_empty() {
        println!("  {}: {}", "Transforms".bright_cyan(), transforms.join(", ").bright_green());
    }
    
    // Parse references if provided using common logic
    let parsed_refs = if let Some(ref_strings) = references {
        match parse_references(ref_strings, true) {
            Ok(refs) => Some(refs),
            Err(e) => {
                eprintln!("{} {}", "❌ Invalid reference:".red(), e);
                std::process::exit(1);
            }
        }
    } else {
        None
    };
    
    // Create tool relation
    let relation = Relation::new_tool(name, transforms);
    
    // Create request
    let request = DeclareRelationRequest { relation, references: parsed_refs, user_prompt: prompt };
    
    // Send to daemon with extended timeout for AI generation
    let mut client = DaemonClient::new(port);
    let daemon_request = request.build_request(generate_id())?;
    let response = client.request_timeout(daemon_request, Duration::from_secs(300))?; // 5 minutes for AI - matches daemon timeout
    
    if !response.success {
        let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
        eprintln!("{} {}", "❌ Failed to declare tool:".red(), error);
        std::process::exit(1);
    }
    
    // Parse and display response
    if let Some(data) = response.data {
        let declare_response = DeclareRelationResponse::parse_response(&data)?;
        declare_response.display(OutputFormat::Plain)?;
    }
    
    Ok(())
}

/// Handle declaring a new artifact relation
pub fn handle_declare_artifact(port: u16, name: &str, artifact_type: &str, file_type: &str, prompt: Option<String>) -> Result<()> {
    println!("{}", format!("🌟 Declaring artifact: {}", name).bright_blue());
    println!("  {}: {}", "Type".bright_cyan(), artifact_type.bright_green());
    println!("  {}: {}", "File Type".bright_cyan(), file_type.bright_green());
    
    // Create artifact relation
    let relation = Relation::new_artifact(name, artifact_type, file_type);
    
    // Create request
    let request = DeclareRelationRequest { relation, references: None, user_prompt: prompt };
    
    // Send to daemon with extended timeout for AI generation
    let mut client = DaemonClient::new(port);
    let daemon_request = request.build_request(generate_id())?;
    let response = client.request_timeout(daemon_request, Duration::from_secs(300))?; // 5 minutes for AI - matches daemon timeout
    
    if !response.success {
        let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
        eprintln!("{} {}", "❌ Failed to declare artifact:".red(), error);
        std::process::exit(1);
    }
    
    // Parse and display response
    if let Some(data) = response.data {
        let declare_response = DeclareRelationResponse::parse_response(&data)?;
        declare_response.display(OutputFormat::Plain)?;
    }
    
    Ok(())
}