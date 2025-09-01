use anyhow::{Result, Context, bail};
use colored::*;
use std::process::{Command, Stdio};
use std::io::{BufRead, BufReader, Write};
use std::fs;
use std::env;
use std::path::PathBuf;
use crate::DaemonAction;
use crate::help_text::*;

const DAEMON_BINARY: &str = "port42d";
const PID_FILE: &str = "/tmp/port42d.pid";
const LOG_FILE: &str = ".port42/daemon.log";

fn get_log_path() -> PathBuf {
    let home = env::var("HOME").unwrap_or_else(|_| ".".to_string());
    PathBuf::from(home).join(LOG_FILE)
}

fn is_daemon_running() -> bool {
    // Check if PID file exists and process is running
    if let Ok(pid_str) = fs::read_to_string(PID_FILE) {
        if let Ok(pid) = pid_str.trim().parse::<u32>() {
            // Check if process exists (signal 0)
            unsafe {
                libc::kill(pid as i32, 0) == 0
            }
        } else {
            false
        }
    } else {
        // Also check by process name
        Command::new("pgrep")
            .arg("-f")
            .arg(DAEMON_BINARY)
            .output()
            .map(|output| output.status.success())
            .unwrap_or(false)
    }
}

fn start_daemon(background: bool) -> Result<()> {
    if is_daemon_running() {
        println!("{}", ERR_DAEMON_ALREADY_RUNNING.green());
        return Ok(());
    }
    
    // Check for API key - PORT42_ANTHROPIC_API_KEY first, then ANTHROPIC_API_KEY
    let api_key = env::var("PORT42_ANTHROPIC_API_KEY")
        .or_else(|_| env::var("ANTHROPIC_API_KEY"))
        .ok();
    if api_key.is_none() {
        println!("{}", ERR_NO_API_KEY.yellow());
        println!("{}", "To channel consciousness:".yellow());
        println!("  export PORT42_ANTHROPIC_API_KEY='your-key-here'");
        println!("  # or");
        println!("  export ANTHROPIC_API_KEY='your-key-here'");
        println!("  port42 daemon restart\n");
    }
    
    // Check if daemon binary exists
    let daemon_path = which::which(DAEMON_BINARY)
        .context(format!("{}
💡 Install Port 42 to manifest the daemon", ERR_BINARY_NOT_FOUND))?;
    
    println!("{}", MSG_DAEMON_STARTING.blue().bold());
    
    // Provide sudo hint
    println!("{}", "💡 Tip: For port 42, use: sudo -E port42 daemon start -b".dimmed());
    println!("{}", "   (Otherwise daemon will use port 4242)".dimmed());
    println!();
    
    if background {
        // Start in background using nohup
        let log_path = get_log_path();
        
        // Create log directory if needed
        if let Some(parent) = log_path.parent() {
            fs::create_dir_all(parent)?;
        }
        
        let mut cmd = Command::new("nohup");
        cmd.arg(&daemon_path)
            .stdout(Stdio::from(fs::File::create(&log_path)?))
            .stderr(Stdio::from(fs::File::create(&log_path)?))
            .stdin(Stdio::null());
        
        // The daemon should inherit all environment variables by default
        // No need to explicitly set them unless we want to override
        
        let child = cmd.spawn()
            .context(ERR_DAEMON_START_FAILED)?;
        
        // Save PID
        fs::write(PID_FILE, child.id().to_string())?;
        
        // Wait a moment to check if it started successfully
        std::thread::sleep(std::time::Duration::from_secs(2));
        
        if is_daemon_running() {
            println!("{}", MSG_DAEMON_SUCCESS.green());
            println!("{}", format!("📋 Log file: {}", log_path.display()).dimmed());
        } else {
            bail!(format_error_with_suggestion(
                ERR_DAEMON_START_FAILED,
                &format!("Check the log file: {}", log_path.display())
            ));
        }
    } else {
        // Start in foreground - but still log to file
        let log_path = get_log_path();
        
        // Create log directory if needed
        if let Some(parent) = log_path.parent() {
            fs::create_dir_all(parent)?;
        }
        
        println!("{}", "Starting in foreground mode (Ctrl+C to stop)...".dimmed());
        println!("{}", format!("📋 Log file: {}", log_path.display()).dimmed());
        
        // Open log file for writing
        let log_file = fs::File::create(&log_path)?;
        
        // Start daemon directly, capturing output to both terminal and file
        let mut cmd = Command::new(&daemon_path);
        
        // The daemon should inherit all environment variables by default
        
        // Spawn the process with piped stdout/stderr
        let mut child = cmd
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .context(ERR_DAEMON_START_FAILED)?;
        
        // Read from daemon and write to both terminal and file
        let stdout = child.stdout.take().expect("Failed to capture stdout");
        let stderr = child.stderr.take().expect("Failed to capture stderr");
        
        // Use threads to handle both streams
        let log_file_stdout = log_file.try_clone()?;
        let log_file_stderr = log_file.try_clone()?;
        
        std::thread::spawn(move || {
            let reader = BufReader::new(stdout);
            let mut writer = std::io::BufWriter::new(log_file_stdout);
            for line in reader.lines() {
                if let Ok(line) = line {
                    println!("{}", line);
                    writeln!(writer, "{}", line).ok();
                    writer.flush().ok();
                }
            }
        });
        
        std::thread::spawn(move || {
            let reader = BufReader::new(stderr);
            let mut writer = std::io::BufWriter::new(log_file_stderr);
            for line in reader.lines() {
                if let Ok(line) = line {
                    eprintln!("{}", line);
                    writeln!(writer, "{}", line).ok();
                    writer.flush().ok();
                }
            }
        });
        
        // Wait for the child process to exit
        let status = child.wait()?;
        
        if !status.success() {
            bail!(format_error_with_suggestion(
                ERR_DAEMON_START_FAILED,
                &format!("Process exited with status: {}", status)
            ));
        }
    }
    
    Ok(())
}

fn stop_daemon() -> Result<()> {
    if !is_daemon_running() {
        println!("{}", format_daemon_connection_error(42));
        return Ok(());
    }
    
    println!("{}", MSG_DAEMON_STOPPING.red().bold());
    
    // Try to read PID and kill gracefully
    if let Ok(pid_str) = fs::read_to_string(PID_FILE) {
        if let Ok(pid) = pid_str.trim().parse::<u32>() {
            unsafe {
                // Send SIGTERM
                if libc::kill(pid as i32, libc::SIGTERM) == 0 {
                    // Wait for process to stop
                    for _ in 0..10 {
                        std::thread::sleep(std::time::Duration::from_millis(500));
                        if !is_daemon_running() {
                            println!("{}", MSG_DAEMON_STOPPED.green());
                            fs::remove_file(PID_FILE).ok();
                            return Ok(());
                        }
                    }
                    
                    // Force kill if still running
                    libc::kill(pid as i32, libc::SIGKILL);
                }
            }
        }
    }
    
    // Fallback: kill by name
    Command::new("pkill")
        .arg("-f")
        .arg(DAEMON_BINARY)
        .status()
        .context(ERR_FAILED_TO_STOP)?;
    
    fs::remove_file(PID_FILE).ok();
    println!("{}", MSG_DAEMON_STOPPED.green());
    
    Ok(())
}

fn show_logs(lines: usize, follow: bool) -> Result<()> {
    let log_path = get_log_path();
    
    if !log_path.exists() {
        bail!(format_error_with_suggestion(
            ERR_LOG_NOT_FOUND,
            &format!("Expected at: {}", log_path.display())
        ));
    }
    
    println!("{}", MSG_DAEMON_LOGS.bright_white().bold());
    println!("{}", format!("File: {}", log_path.display()).dimmed());
    println!("{}", "─".repeat(50).dimmed());
    
    if follow {
        // Follow logs using tail -f
        let mut child = Command::new("tail")
            .arg("-f")
            .arg(&log_path)
            .stdout(Stdio::piped())
            .spawn()
            .context("Failed to follow log stream")?;
        
        if let Some(stdout) = child.stdout.take() {
            let reader = BufReader::new(stdout);
            for line in reader.lines() {
                println!("{}", line?);
            }
        }
    } else {
        // Show last N lines
        let output = Command::new("tail")
            .arg(format!("-{}", lines))
            .arg(&log_path)
            .output()
            .context(ERR_LOG_NOT_FOUND)?;
        
        print!("{}", String::from_utf8_lossy(&output.stdout));
    }
    
    Ok(())
}

pub fn handle_daemon(action: DaemonAction, _port: u16) -> Result<()> {
    match action {
        DaemonAction::Start { background } => {
            start_daemon(background)?;
        }
        
        DaemonAction::Stop => {
            stop_daemon()?;
        }
        
        DaemonAction::Restart => {
            println!("{}", MSG_DAEMON_RESTARTING.yellow().bold());
            
            // Stop if running
            if is_daemon_running() {
                stop_daemon()?;
                std::thread::sleep(std::time::Duration::from_secs(1));
            }
            
            // Start again
            start_daemon(true)?;
        }
        
        DaemonAction::Logs { lines, follow } => {
            show_logs(lines, follow)?;
        }
    }
    
    Ok(())
}