use anyhow::{Context, Result};
use colored::*;
use std::fs;
use std::path::PathBuf;
use crate::protocol::{RealityData, CommandInfo};
use crate::display::{Displayable, OutputFormat};
use crate::help_text;

pub fn handle_reality(_port: u16, verbose: bool, agent: Option<String>) -> Result<()> {
    println!("{}", help_text::MSG_COMMANDS_HEADER.blue().bold());
    println!();
    
    let commands_dir = dirs::home_dir()
        .context("Could not find home directory")?  
        .join(".port42")
        .join("commands");
    
    if !commands_dir.exists() {
        // No commands directory - display empty state
        let reality_data = RealityData {
            commands: vec![],
            total: 0,
            commands_dir,
        };
        
        return reality_data.display(OutputFormat::Plain);
    }
    
    let mut commands = Vec::new();
    
    // Read all files in commands directory
    for entry in fs::read_dir(&commands_dir)? {
        let entry = entry?;
        let path = entry.path();
        
        if path.is_file() {
            if let Some(name) = path.file_name().and_then(|n| n.to_str()) {
                // Skip hidden files and backup files
                if !name.starts_with('.') && !name.ends_with('~') {
                    // Check if executable
                    #[cfg(unix)]
                    {
                        use std::os::unix::fs::PermissionsExt;
                        let metadata = fs::metadata(&path)?;
                        if metadata.permissions().mode() & 0o111 != 0 {
                            commands.push((name.to_string(), path));
                        }
                    }
                    
                    #[cfg(not(unix))]
                    {
                        commands.push((name.to_string(), path));
                    }
                }
            }
        }
    }
    
    // Sort by name
    commands.sort_by(|a, b| a.0.cmp(&b.0));
    
    // Convert to CommandInfo structures
    let mut command_infos = Vec::new();
    
    for (name, path) in commands {
        let (language, description, agent_name) = extract_metadata(&path)?;
        
        // Filter by agent if specified
        if let Some(ref agent_filter) = agent {
            if agent_name.as_deref() != Some(agent_filter) {
                continue;
            }
        }
        
        command_infos.push(CommandInfo {
            name,
            path,
            language,
            description,
            agent: agent_name,
        });
    }
    
    // Create structured data for display
    let reality_data = RealityData {
        total: command_infos.len(),
        commands: command_infos,
        commands_dir,
    };
    
    // Display using the framework
    let format = if verbose {
        OutputFormat::Table
    } else {
        OutputFormat::Plain
    };
    
    reality_data.display(format)?;
    
    Ok(())
}

fn extract_metadata(path: &PathBuf) -> Result<(String, Option<String>, Option<String>)> {
    let mut language = "unknown".to_string();
    let mut description = None;
    let mut agent = None;
    
    if let Ok(content) = fs::read_to_string(path) {
        // Detect language from shebang
        if let Some(first_line) = content.lines().next() {
            if first_line.starts_with("#!/") {
                if first_line.contains("python") {
                    language = "python".to_string();
                } else if first_line.contains("node") {
                    language = "node".to_string();
                } else if first_line.contains("bash") || first_line.contains("sh") {
                    language = "bash".to_string();
                }
            }
        }
        
        // Look for metadata in comments
        for line in content.lines().take(20) {
            if line.contains("Description:") || line.contains("description:") {
                if let Some(desc) = line.split(':').nth(1) {
                    description = Some(desc.trim().to_string());
                }
            }
            if line.contains("Agent:") || line.contains("Created by:") {
                if let Some(agent_name) = line.split(':').nth(1) {
                    agent = Some(agent_name.trim().to_string());
                }
            }
        }
    }
    
    Ok((language, description, agent))
}