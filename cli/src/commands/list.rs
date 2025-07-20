use anyhow::{Context, Result};
use colored::*;
use std::fs;

pub fn handle_list(_port: u16, verbose: bool, agent: Option<String>) -> Result<()> {
    println!("{}", "📋 Generated Commands".blue().bold());
    println!();
    
    let commands_dir = dirs::home_dir()
        .context("Could not find home directory")?  
        .join(".port42")
        .join("commands");
    
    if !commands_dir.exists() {
        println!("{}", "No commands directory found".dimmed());
        println!("\n{}", "Generate your first command:".yellow());
        println!("  {}", "port42 possess @ai-muse".bright_white());
        return Ok(());
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
    
    if commands.is_empty() {
        println!("{}", "No commands found".dimmed());
        println!("\n{}", "Generate your first command:".yellow());
        println!("  {}", "port42 possess @ai-muse".bright_white());
    } else {
        // Sort by name
        commands.sort_by(|a, b| a.0.cmp(&b.0));
        
        // Filter by agent if specified
        let filtered_commands: Vec<_> = if let Some(ref agent_filter) = agent {
            commands.iter()
                .filter(|(_, path)| {
                    // Check if command was created by this agent
                    if let Ok(content) = fs::read_to_string(path) {
                        content.lines().any(|line| 
                            line.contains(&format!("Agent: {}", agent_filter)) ||
                            line.contains(&format!("agent: {}", agent_filter)) ||
                            line.contains(&format!("@{}", agent_filter))
                        )
                    } else {
                        false
                    }
                })
                .cloned()
                .collect()
        } else {
            commands.iter().cloned().collect()
        };
        
        if filtered_commands.is_empty() {
            if agent.is_some() {
                println!("{}", format!("No commands found for agent: {}", agent.unwrap()).dimmed());
            }
        } else {
            // Display commands
            for (name, path) in &filtered_commands {
                print!("{:<20}", name.bright_cyan());
                
                if verbose {
                    // Extract metadata from file
                    if let Ok(content) = fs::read_to_string(path) {
                        let mut description = None;
                        let mut created_by = None;
                        let mut language = "unknown";
                        
                        // Detect language from shebang
                        if let Some(first_line) = content.lines().next() {
                            if first_line.starts_with("#!/") {
                                if first_line.contains("python") {
                                    language = "python";
                                } else if first_line.contains("node") {
                                    language = "node";
                                } else if first_line.contains("bash") || first_line.contains("sh") {
                                    language = "bash";
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
                                if let Some(agent) = line.split(':').nth(1) {
                                    created_by = Some(agent.trim().to_string());
                                }
                            }
                        }
                        
                        // Display metadata
                        print!(" [{:<6}]", language.yellow());
                        if let Some(desc) = description {
                            print!(" {}", desc.dimmed());
                        }
                        if let Some(agent) = created_by {
                            print!(" (by {})", agent.bright_blue());
                        }
                    }
                } else {
                    // Simple view - just try to get description
                    if let Ok(content) = fs::read_to_string(path) {
                        for line in content.lines().take(10) {
                            if line.contains("Description:") || line.contains("description:") {
                                if let Some(desc) = line.split(':').nth(1) {
                                    print!(" - {}", desc.trim().dimmed());
                                    break;
                                }
                            }
                        }
                    }
                }
                println!();
            }
            
            println!("\n{}", format!("Total: {} commands", filtered_commands.len()).dimmed());
            
            if verbose {
                println!("\n{}", "Command Location:".yellow());
                println!("  {}", commands_dir.display().to_string().bright_white());
            }
        }
        
        println!("\n{}", "Add to PATH:".yellow());
        println!("  {}", format!("export PATH=\"$PATH:{}\":", commands_dir.display()).bright_white());
    }
    
    Ok(())
}