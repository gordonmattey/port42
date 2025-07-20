use anyhow::Result;
use colored::*;
use crate::client::DaemonClient;
use crate::types::Request;
use std::io::{self, Write};

pub fn handle_possess(
    port: u16, 
    agent: String, 
    message: Option<String>, 
    session: Option<String>
) -> Result<()> {
    println!("{}", format!("🔮 Possessing {}...", agent).blue().bold());
    
    let mut client = DaemonClient::new(port);
    
    // Generate session ID
    let session_id = session.unwrap_or_else(|| {
        format!("cli-{}", chrono::Utc::now().timestamp())
    });
    
    if let Some(msg) = message {
        // Single message mode
        send_message(&mut client, &session_id, &agent, &msg)?;
    } else {
        // Interactive mode
        println!("{}", "Entering interactive mode. Type '/end' to finish.".dimmed());
        println!();
        
        loop {
            // Prompt
            print!("{} ", ">".bright_blue());
            io::stdout().flush()?;
            
            // Read input
            let mut input = String::new();
            io::stdin().read_line(&mut input)?;
            let input = input.trim();
            
            // Check for exit
            if input == "/end" || input.is_empty() {
                break;
            }
            
            // Send message
            send_message(&mut client, &session_id, &agent, input)?;
        }
        
        // End session
        println!("\n{}", "Ending possession session...".dimmed());
        let end_request = Request {
            request_type: "end".to_string(),
            id: session_id.clone(),
            payload: serde_json::json!({
                "session_id": session_id
            }),
        };
        
        match client.request(end_request) {
            Ok(_) => println!("{}", "✅ Session ended".green()),
            Err(e) => eprintln!("{}", format!("⚠️  Failed to end session: {}", e).yellow()),
        }
    }
    
    Ok(())
}

fn send_message(client: &mut DaemonClient, session_id: &str, agent: &str, message: &str) -> Result<()> {
    let request = Request {
        request_type: "possess".to_string(),
        id: session_id.to_string(),
        payload: serde_json::json!({
            "agent": agent,
            "message": message
        }),
    };
    
    match client.request(request) {
        Ok(response) => {
            if response.success {
                if let Some(data) = response.data {
                    if let Some(ai_message) = data.get("message").and_then(|v| v.as_str()) {
                        println!("\n{}", agent.bright_blue());
                        println!("{}", ai_message);
                        println!();
                        
                        // Check if command was generated
                        if let Some(command) = data.get("command_generated").and_then(|v| v.as_str()) {
                            println!("{}", format!("✨ Command crystallized: {}", command).bright_green().bold());
                            println!("{}", "Add to PATH to use:".yellow());
                            println!("  {}", "export PATH=\"$PATH:$HOME/.port42/commands\"".bright_white());
                            println!();
                        }
                    } else {
                        println!("{}", "No message in response".dimmed());
                    }
                } else {
                    println!("{}", "No data in response".dimmed());
                }
            } else {
                println!("{}", "❌ Failed to send message".red());
                if let Some(error) = response.error {
                    println!("  {}", error.dimmed());
                }
            }
        }
        Err(e) => {
            eprintln!("{}", e);
            return Err(e);
        }
    }
    
    Ok(())
}