use anyhow::Result;
use colored::*;

use crate::client::DaemonClient;
use crate::protocol::{
    DeclareRelationRequest, DeclareRelationResponse, 
    Relation, RequestBuilder, ResponseParser
};
use crate::display::{Displayable, OutputFormat};
use crate::common::generate_id;

/// Handle declaring a new tool relation
pub fn handle_declare_tool(port: u16, name: &str, transforms: Vec<String>) -> Result<()> {
    println!("{}", format!("🌟 Declaring tool: {}", name).bright_blue());
    
    if !transforms.is_empty() {
        println!("  {}: {}", "Transforms".bright_cyan(), transforms.join(", ").bright_green());
    }
    
    // Create tool relation
    let relation = Relation::new_tool(name, transforms);
    
    // Create request
    let request = DeclareRelationRequest { relation };
    
    // Send to daemon
    let mut client = DaemonClient::new(port);
    let daemon_request = request.build_request(generate_id())?;
    let response = client.request(daemon_request)?;
    
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
pub fn handle_declare_artifact(port: u16, name: &str, artifact_type: &str, file_type: &str) -> Result<()> {
    println!("{}", format!("🌟 Declaring artifact: {}", name).bright_blue());
    println!("  {}: {}", "Type".bright_cyan(), artifact_type.bright_green());
    println!("  {}: {}", "File Type".bright_cyan(), file_type.bright_green());
    
    // Create artifact relation
    let relation = Relation::new_artifact(name, artifact_type, file_type);
    
    // Create request
    let request = DeclareRelationRequest { relation };
    
    // Send to daemon
    let mut client = DaemonClient::new(port);
    let daemon_request = request.build_request(generate_id())?;
    let response = client.request(daemon_request)?;
    
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

/// Handle listing all relations
pub fn handle_list_relations(port: u16, relation_type: Option<&str>) -> Result<()> {
    use crate::protocol::{ListRelationsRequest, ListRelationsResponse};
    
    let request = ListRelationsRequest {
        relation_type: relation_type.map(String::from),
    };
    
    let mut client = DaemonClient::new(port);
    let daemon_request = request.build_request(generate_id())?;
    let response = client.request(daemon_request)?;
    
    if !response.success {
        let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
        eprintln!("{} {}", "❌ Failed to list relations:".red(), error);
        std::process::exit(1);
    }
    
    if let Some(data) = response.data {
        let list_response = ListRelationsResponse::parse_response(&data)?;
        list_response.display(OutputFormat::Table)?;
    }
    
    Ok(())
}

/// Handle getting a specific relation
pub fn handle_get_relation(port: u16, relation_id: &str) -> Result<()> {
    use crate::protocol::{GetRelationRequest, GetRelationResponse};
    
    let request = GetRelationRequest {
        relation_id: relation_id.to_string(),
    };
    
    let mut client = DaemonClient::new(port);
    let daemon_request = request.build_request(generate_id())?;
    let response = client.request(daemon_request)?;
    
    if !response.success {
        let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
        eprintln!("{} {}", "❌ Failed to get relation:".red(), error);
        std::process::exit(1);
    }
    
    if let Some(data) = response.data {
        let get_response = GetRelationResponse::parse_response(&data)?;
        
        // Display relation info
        let relation = &get_response.relation;
        println!("{}", format!("📋 Relation: {}", relation.id).bright_blue());
        println!("  {}: {}", "Type".bright_cyan(), relation.relation_type.bright_green());
        
        if let Some(name) = relation.name() {
            println!("  {}: {}", "Name".bright_cyan(), name.bright_white());
        }
        
        let transforms = relation.transforms();
        if !transforms.is_empty() {
            println!("  {}: {}", "Transforms".bright_cyan(), transforms.join(", ").bright_green());
        }
        
        // Show other properties
        for (key, value) in &relation.properties {
            if key != "name" && key != "transforms" {
                println!("  {}: {}", key.bright_cyan(), 
                       format!("{}", value).bright_white());
            }
        }
    }
    
    Ok(())
}

/// Handle deleting a relation
pub fn handle_delete_relation(port: u16, relation_id: &str) -> Result<()> {
    use crate::protocol::{DeleteRelationRequest, DeleteRelationResponse};
    
    println!("{}", format!("🗑️ Deleting relation: {}", relation_id).bright_red());
    
    let request = DeleteRelationRequest {
        relation_id: relation_id.to_string(),
    };
    
    let mut client = DaemonClient::new(port);
    let daemon_request = request.build_request(generate_id())?;
    let response = client.request(daemon_request)?;
    
    if !response.success {
        let error = response.error.unwrap_or_else(|| "Unknown error".to_string());
        eprintln!("{} {}", "❌ Failed to delete relation:".red(), error);
        std::process::exit(1);
    }
    
    if let Some(data) = response.data {
        let delete_response = DeleteRelationResponse::parse_response(&data)?;
        delete_response.display(OutputFormat::Plain)?;
    }
    
    Ok(())
}