use std::process::Command;
use std::time::Duration;
use std::thread;

#[test]
fn test_possess_non_interactive_basic() {
    // Ensure daemon is running
    ensure_daemon_running();
    
    // Test basic possess command
    let output = Command::new("./target/debug/port42")
        .args(&["possess", "@ai-engineer", "test message"])
        .output()
        .expect("Failed to execute possess command");
    
    assert!(output.status.success(), "Command failed with output: {}", String::from_utf8_lossy(&output.stderr));
    
    let stdout = String::from_utf8_lossy(&output.stdout);
    // Should contain some response
    assert!(!stdout.is_empty(), "Expected output from possess command");
}

#[test]
fn test_possess_with_session_id() {
    ensure_daemon_running();
    
    let session_id = format!("test-session-{}", std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_millis());
    
    // First message creates session
    let output1 = Command::new("./target/debug/port42")
        .args(&["possess", "@ai-muse", &session_id, "first message"])
        .output()
        .expect("Failed to execute first possess command");
    
    assert!(output1.status.success());
    
    // Second message continues session
    let output2 = Command::new("./target/debug/port42")
        .args(&["possess", "@ai-muse", &session_id, "second message"])
        .output()
        .expect("Failed to execute second possess command");
    
    assert!(output2.status.success());
    
    // Both should have responses
    assert!(!String::from_utf8_lossy(&output1.stdout).is_empty());
    assert!(!String::from_utf8_lossy(&output2.stdout).is_empty());
}

#[test]
fn test_possess_invalid_agent() {
    ensure_daemon_running();
    
    let output = Command::new("./target/debug/port42")
        .args(&["possess", "@invalid-agent", "test message"])
        .output()
        .expect("Failed to execute possess command");
    
    // Should fail with invalid agent
    assert!(!output.status.success());
    
    let stderr = String::from_utf8_lossy(&output.stderr);
    assert!(stderr.contains("Unknown consciousness") || stderr.contains("invalid agent"),
        "Expected error about invalid agent, got: {}", stderr);
}

#[test]
fn test_possess_command_generation() {
    ensure_daemon_running();
    
    // This test would require ANTHROPIC_API_KEY to actually generate commands
    // For now, just test that the command runs without crashing
    let output = Command::new("./target/debug/port42")
        .args(&["possess", "@ai-engineer", "create a hello world script"])
        .output()
        .expect("Failed to execute possess command");
    
    // If no API key, it should still run but with appropriate message
    let stdout = String::from_utf8_lossy(&output.stdout);
    let stderr = String::from_utf8_lossy(&output.stderr);
    
    // Either success with command generation, or error about API key
    assert!(output.status.success() || stderr.contains("ANTHROPIC_API_KEY"),
        "Unexpected error: {}", stderr);
}

// Helper to ensure daemon is running
fn ensure_daemon_running() {
    // Check if daemon is already running
    let status = Command::new("./target/debug/port42")
        .arg("status")
        .output()
        .expect("Failed to check daemon status");
    
    let output = String::from_utf8_lossy(&status.stdout);
    
    if output.contains("dormant") || output.contains("not running") {
        // Start daemon in background
        println!("Starting test daemon...");
        
        // Use the debug binary to start daemon
        let daemon_start = Command::new("./target/debug/port42")
            .args(&["daemon", "start", "-b"])
            .output()
            .expect("Failed to start daemon");
        
        if !daemon_start.status.success() {
            eprintln!("Failed to start daemon: {}", String::from_utf8_lossy(&daemon_start.stderr));
            panic!("Cannot run tests without daemon");
        }
        
        // Give daemon time to start
        thread::sleep(Duration::from_secs(2));
    }
}

// Cleanup helper - could be used in a test suite teardown
#[allow(dead_code)]
fn stop_daemon() {
    let _ = Command::new("./target/debug/port42")
        .args(&["daemon", "stop"])
        .output();
}