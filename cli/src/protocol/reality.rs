use crate::display::{Displayable, OutputFormat, components};
use crate::help_text;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use colored::*;
use std::path::PathBuf;

// Reality doesn't need request/response types since it reads filesystem directly
// But we create structured types for business logic and display separation

#[derive(Debug, Serialize)]
pub struct RealityData {
    pub commands: Vec<CommandInfo>,
    pub total: usize,
    pub commands_dir: PathBuf,
}

#[derive(Debug, Serialize, Clone)]
pub struct CommandInfo {
    pub name: String,
    pub path: PathBuf,
    pub language: String,
    pub description: Option<String>,
    pub agent: Option<String>,
}

impl Displayable for RealityData {
    fn display(&self, format: OutputFormat) -> Result<()> {
        match format {
            OutputFormat::Json => {
                println!("{}", serde_json::to_string_pretty(self)?);
            }
            OutputFormat::Table => {
                if self.commands.is_empty() {
                    self.display_empty();
                } else {
                    let mut table = components::TableBuilder::new();
                    table.add_header(vec!["Command", "Language", "Agent", "Description"]);
                    
                    for cmd in &self.commands {
                        table.add_row(vec![
                            cmd.name.clone(),
                            cmd.language.clone(),
                            cmd.agent.as_deref().unwrap_or("-").to_string(),
                            cmd.description.as_deref().unwrap_or("-").to_string(),
                        ]);
                    }
                    
                    table.print();
                    println!("\n{}", help_text::format_total_commands(self.total).dimmed());
                    println!("\n{}", "Command Location:".yellow());
                    println!("  {}", self.commands_dir.display().to_string().bright_white());
                }
                self.display_path_hint();
            }
            OutputFormat::Plain => {
                if self.commands.is_empty() {
                    self.display_empty();
                } else {
                    for cmd in &self.commands {
                        print!("{:<20}", cmd.name.bright_cyan());
                        if let Some(ref desc) = cmd.description {
                            print!(" - {}", desc.dimmed());
                        }
                        println!();
                    }
                    println!("\n{}", help_text::format_total_commands(self.total).dimmed());
                }
                self.display_path_hint();
            }
        }
        Ok(())
    }
}

impl RealityData {
    fn display_empty(&self) {
        println!("{}", "No commands found".dimmed());
        println!("\n{}", "Generate your first command:".yellow());
        println!("  {}", "port42 possess @ai-muse".bright_white());
    }
    
    fn display_path_hint(&self) {
        println!("\n{}", "Add to PATH:".yellow());
        println!("  {}", format!("export PATH=\"$PATH:{}\"", self.commands_dir.display()).bright_white());
    }
}