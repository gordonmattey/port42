use crate::protocol::relations::Reference;
use anyhow::{Result, bail};
use colored::*;

/// Parse reference strings into Reference structs
/// Common logic used by both declare and possess modes
pub fn parse_references(ref_strings: Vec<String>, show_output: bool) -> Result<Vec<Reference>> {
    let mut refs = Vec::new();
    
    for ref_str in ref_strings {
        match Reference::from_string(&ref_str) {
            Ok(reference) => {
                if show_output {
                    println!("  {}: {} → {}", 
                           "Reference".bright_cyan(), 
                           reference.ref_type.bright_yellow(), 
                           reference.target.bright_white());
                }
                refs.push(reference);
            }
            Err(e) => {
                bail!("Invalid reference {}: {}", ref_str.bright_white(), e);
            }
        }
    }
    
    Ok(refs)
}