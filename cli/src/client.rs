use anyhow::{Context, Result};
use std::io::{Read, Write};
use std::net::TcpStream;
use std::time::Duration;

use crate::types::{Request, Response};

pub struct DaemonClient {
    port: u16,
}

impl DaemonClient {
    pub fn new(port: u16) -> Self {
        Self { port }
    }
    
    pub fn send_request(&self, request: Request) -> Result<Response> {
        // Connect to daemon
        let mut stream = TcpStream::connect_timeout(
            &format!("127.0.0.1:{}", self.port).parse()?,
            Duration::from_secs(2)
        )
        .context("Failed to connect to daemon. Is it running?")?;
        
        // Send request
        let request_json = serde_json::to_string(&request)?;
        stream.write_all(request_json.as_bytes())?;
        stream.write_all(b"\n")?;
        
        // Read response
        let mut buffer = vec![0; 16384]; // Larger buffer for possess responses
        let n = stream.read(&mut buffer)
            .context("Failed to read response from daemon")?;
        
        // Parse response
        let response: Response = serde_json::from_slice(&buffer[..n])
            .context("Failed to parse daemon response")?;
        
        Ok(response)
    }
    
    pub fn is_running(&self) -> bool {
        TcpStream::connect_timeout(
            &format!("127.0.0.1:{}", self.port).parse().unwrap(),
            Duration::from_millis(500)
        ).is_ok()
    }
}

// Helper function to detect which port the daemon is on
pub fn detect_daemon_port() -> Option<u16> {
    if TcpStream::connect_timeout(&"127.0.0.1:42".parse().unwrap(), Duration::from_millis(100)).is_ok() {
        Some(42)
    } else if TcpStream::connect_timeout(&"127.0.0.1:4242".parse().unwrap(), Duration::from_millis(100)).is_ok() {
        Some(4242)
    } else {
        None
    }
}