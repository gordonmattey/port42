use anyhow::{anyhow, Result};
use colored::*;
use std::io::{BufRead, BufReader, Write};
use std::net::{TcpStream, SocketAddr};
use std::time::{Duration, Instant};
use std::sync::atomic::{AtomicU32, Ordering};

use crate::protocol::DaemonRequest;
use crate::types::Response; // Keep old Response for now

// Track recursion depth to prevent stack overflow
static RECURSION_DEPTH: AtomicU32 = AtomicU32::new(0);

// RAII guard to ensure recursion depth is decremented
struct RecursionGuard;

impl Drop for RecursionGuard {
    fn drop(&mut self) {
        // Saturating sub to prevent underflow
        let current = RECURSION_DEPTH.load(Ordering::SeqCst);
        if current > 0 {
            RECURSION_DEPTH.fetch_sub(1, Ordering::SeqCst);
        }
    }
}

pub struct DaemonClient {
    port: u16,
    stream: Option<TcpStream>,
    reader: Option<BufReader<TcpStream>>,
    connection_timeout: Duration,
    request_timeout: Duration,
}

impl DaemonClient {
    pub fn new(port: u16) -> Self {
        Self {
            port,
            stream: None,
            reader: None,
            connection_timeout: Duration::from_secs(2),
            request_timeout: Duration::from_secs(30), // Longer for AI requests
        }
    }
    
    pub fn port(&self) -> u16 {
        self.port
    }
    
    /// Ensure we have a valid connection to the daemon
    pub fn ensure_connected(&mut self) -> Result<()> {
        // Guard against recursion
        let depth = RECURSION_DEPTH.fetch_add(1, Ordering::SeqCst);
        
        // Create guard immediately after incrementing
        let _guard = RecursionGuard;
        
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: ensure_connected: Recursion depth = {}", depth);
        }
        
        // Prevent stack overflow from recursive calls
        if depth > 3 {
            return Err(anyhow!("Connection recursion detected - possible stack overflow"));
        }
        
        // Check if we already have a connection
        if self.stream.is_some() {
            // Test if still alive with a quick ping
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: ensure_connected: Testing existing connection with ping");
            }
            if self.ping().is_ok() {
                return Ok(());
            }
            // Connection is dead, reset
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: ensure_connected: Connection dead, resetting");
            }
            self.stream = None;
            self.reader = None;
        }
        
        // Try to connect
        let addr: SocketAddr = format!("127.0.0.1:{}", self.port).parse()?;
        
        match TcpStream::connect_timeout(&addr, self.connection_timeout) {
            Ok(stream) => {
                // Set timeouts on the stream
                stream.set_read_timeout(Some(self.request_timeout))?;
                stream.set_write_timeout(Some(Duration::from_secs(5)))?;
                
                // Clone for the reader
                let reader_stream = stream.try_clone()?;
                let reader = BufReader::with_capacity(65536, reader_stream); // 64KB buffer
                
                self.stream = Some(stream);
                self.reader = Some(reader);
                
                Ok(())
            }
            Err(e) => Err(self.enhance_connection_error(e)),
        }
    }
    
    /// Send a request and receive a response
    pub fn request(&mut self, request: DaemonRequest) -> Result<Response> {
        self.ensure_connected()?;
        
        let start = Instant::now();
        
        // Send request
        let stream = self.stream.as_mut().unwrap();
        let json = serde_json::to_string(&request)?;
        
        if std::env::var("PORT42_VERBOSE").is_ok() {
            eprintln!("{} {}", "→ Request:".dimmed(), json.dimmed());
        }
        
        stream.write_all(json.as_bytes())?;
        stream.write_all(b"\n")?;
        stream.flush()?;
        
        // Read response (line-based protocol)
        let reader = self.reader.as_mut().unwrap();
        let mut line = String::new();
        
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: About to read response line");
        }
        
        // Retry on EAGAIN (Resource temporarily unavailable)
        let mut retry_count = 0;
        let bytes_read = loop {
            match reader.read_line(&mut line) {
                Ok(bytes) => break bytes,
                Err(e) if e.kind() == std::io::ErrorKind::WouldBlock && retry_count < 3 => {
                    if std::env::var("PORT42_DEBUG").is_ok() {
                        eprintln!("DEBUG: Got EAGAIN, retry {} of 3", retry_count + 1);
                    }
                    retry_count += 1;
                    std::thread::sleep(Duration::from_millis(10));
                    continue;
                }
                Err(e) => return Err(self.enhance_io_error(e, "reading response")),
            }
        };
            
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: Read {} bytes, has_newline={}", bytes_read, line.ends_with('\n'));
            if bytes_read == 0 {
                eprintln!("DEBUG: Got 0 bytes - connection closed by daemon");
            }
        }
        
        let elapsed = start.elapsed();
        
        if std::env::var("PORT42_VERBOSE").is_ok() {
            eprintln!("{} {} {:?}", "← Response:".dimmed(), 
                     if line.len() > 200 { format!("{}...", &line[..200]) } else { line.clone() }.dimmed(),
                     elapsed);
        }
        
        // Debug: Check response size before parsing
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: Response line length: {} bytes", line.len());
            if line.len() > 1000 {
                eprintln!("DEBUG: Large response detected! First 200 chars: {}", &line[..200.min(line.len())]);
            } else if line.len() < 100 && line.len() > 0 {
                eprintln!("DEBUG: Small response: '{}'", line.trim());
            }
        }
        
        // Parse response
        let response: Response = serde_json::from_str(&line)
            .map_err(|e| anyhow!("Invalid response from daemon: {}\nRaw response: {}", e, 
                               if line.len() > 200 { format!("{}...", &line[..200]) } else { line.clone() }))?;
        
        Ok(response)
    }
    
    /// Send a request with a custom timeout
    pub fn request_timeout(&mut self, request: DaemonRequest, timeout: Duration) -> Result<Response> {
        let old_timeout = self.request_timeout;
        self.request_timeout = timeout;
        
        // Update stream timeout if connected
        if let Some(stream) = &self.stream {
            stream.set_read_timeout(Some(timeout))?;
        }
        
        let result = self.request(request);
        
        // Restore timeout
        self.request_timeout = old_timeout;
        if let Some(stream) = &self.stream {
            stream.set_read_timeout(Some(old_timeout))?;
        }
        
        result
    }
    
    /// Test if the connection is still alive
    fn ping(&mut self) -> Result<()> {
        if std::env::var("PORT42_DEBUG").is_ok() {
            eprintln!("DEBUG: ping() called");
        }
        
        let req = DaemonRequest {
            request_type: "ping".to_string(),
            id: "ping".to_string(),
            payload: serde_json::Value::Null,
        };
        
        // Don't use request_timeout as it might cause recursion
        // Instead, do a simple write/read test
        let stream = self.stream.as_mut().ok_or_else(|| anyhow!("No stream for ping"))?;
        let json = serde_json::to_string(&req)?;
        
        // Try to write
        if let Err(e) = stream.write_all(json.as_bytes()) {
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: ping write failed: {}", e);
            }
            return Err(anyhow!("Ping write failed"));
        }
        
        if let Err(e) = stream.write_all(b"\n") {
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: ping newline write failed: {}", e);
            }
            return Err(anyhow!("Ping write failed"));
        }
        
        if let Err(e) = stream.flush() {
            if std::env::var("PORT42_DEBUG").is_ok() {
                eprintln!("DEBUG: ping flush failed: {}", e);
            }
            return Err(anyhow!("Ping flush failed"));
        }
        
        // Try to read response
        let reader = self.reader.as_mut().ok_or_else(|| anyhow!("No reader for ping"))?;
        let mut line = String::new();
        
        match reader.read_line(&mut line) {
            Ok(0) => {
                if std::env::var("PORT42_DEBUG").is_ok() {
                    eprintln!("DEBUG: ping read returned 0 bytes - connection closed");
                }
                Err(anyhow!("Connection closed"))
            }
            Ok(n) => {
                if std::env::var("PORT42_DEBUG").is_ok() {
                    eprintln!("DEBUG: ping read {} bytes: {}", n, line.trim());
                }
                // Just check if we got a response, don't parse it
                if n > 0 {
                    Ok(())
                } else {
                    Err(anyhow!("Empty ping response"))
                }
            }
            Err(e) => {
                if std::env::var("PORT42_DEBUG").is_ok() {
                    eprintln!("DEBUG: ping read failed: {}", e);
                }
                Err(anyhow!("Ping read failed"))
            }
        }
    }
    
    /// Check if daemon is running (without connecting)
    pub fn is_running(&self) -> bool {
        TcpStream::connect_timeout(
            &format!("127.0.0.1:{}", self.port).parse().unwrap(),
            Duration::from_millis(500)
        ).is_ok()
    }
    
    /// Enhance connection errors with helpful context
    fn enhance_connection_error(&self, err: std::io::Error) -> anyhow::Error {
        use std::io::ErrorKind;
        
        match err.kind() {
            ErrorKind::ConnectionRefused => {
                anyhow!(
                    "{}\n\n{} {}\n\n{}\n  {}\n\n{}\n  {}",
                    "🔌 Cannot connect to Port 42 daemon".red().bold(),
                    "The daemon is not running on port".yellow(),
                    self.port.to_string().bright_white(),
                    "To start the daemon:".bright_white(),
                    "sudo -E ./bin/port42d".bright_cyan(),
                    "Or if installed:".bright_white(),
                    "sudo -E port42 daemon start".bright_cyan()
                )
            }
            ErrorKind::PermissionDenied => {
                anyhow!(
                    "{}\n\n{}\n  {}",
                    "🚫 Permission denied".red().bold(),
                    "Port 42 requires elevated permissions. Try:".yellow(),
                    "sudo -E port42".bright_cyan()
                )
            }
            ErrorKind::TimedOut => {
                anyhow!(
                    "{}\n\n{}\n{}",
                    "⏱️  Connection timed out".red().bold(),
                    "The daemon might be busy or unresponsive.".yellow(),
                    "Try again in a moment.".dimmed()
                )
            }
            _ => anyhow!("Connection failed: {}", err),
        }
    }
    
    /// Enhance IO errors with context
    fn enhance_io_error(&self, err: std::io::Error, context: &str) -> anyhow::Error {
        use std::io::ErrorKind;
        
        match err.kind() {
            ErrorKind::UnexpectedEof => {
                anyhow!(
                    "{}\n\n{}",
                    format!("🔌 Connection lost while {}", context).red().bold(),
                    "The daemon may have crashed or been stopped.".yellow()
                )
            }
            ErrorKind::TimedOut => {
                anyhow!(
                    "{}\n\n{}",
                    format!("⏱️  Timeout while {}", context).red().bold(),
                    "The operation took too long. The daemon might be processing another request.".yellow()
                )
            }
            ErrorKind::WouldBlock => {
                anyhow!(
                    "{}\n\n{}\n{}",
                    format!("🔄 Resource temporarily unavailable while {}", context).red().bold(),
                    "The socket buffer might be full or timing issue occurred.".yellow(),
                    "This usually resolves itself. Try again.".dimmed()
                )
            }
            _ => anyhow!("IO error while {}: {}", context, err),
        }
    }
}

/// Helper function to detect which port the daemon is on
pub fn detect_daemon_port() -> Option<u16> {
    if TcpStream::connect_timeout(&"127.0.0.1:42".parse().unwrap(), Duration::from_millis(100)).is_ok() {
        Some(42)
    } else if TcpStream::connect_timeout(&"127.0.0.1:4242".parse().unwrap(), Duration::from_millis(100)).is_ok() {
        Some(4242)
    } else {
        None
    }
}