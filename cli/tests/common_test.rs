use port42::common::{generate_id, errors::Port42Error};
use port42::common::utils::{timestamp_millis, format_timestamp, extract_session_id};

#[test]
fn test_generate_id() {
    let id1 = generate_id();
    // Sleep to ensure different timestamp
    std::thread::sleep(std::time::Duration::from_millis(1));
    let id2 = generate_id();
    
    // Should start with "cli-"
    assert!(id1.starts_with("cli-"));
    assert!(id2.starts_with("cli-"));
    
    // Should be unique
    assert_ne!(id1, id2);
    
    // Should contain timestamp
    let timestamp_part = id1.strip_prefix("cli-").unwrap();
    assert!(timestamp_part.parse::<u128>().is_ok());
}

#[test]
fn test_port42_error() {
    let conn_err = Port42Error::Connection("localhost:42".to_string());
    assert!(conn_err.to_string().contains("Connection failed"));
    
    let daemon_err = Port42Error::Daemon("Unknown command".to_string());
    assert!(daemon_err.to_string().contains("Daemon error"));
    
    let parse_err = Port42Error::Parse("Invalid JSON".to_string());
    assert!(parse_err.to_string().contains("Parse error"));
}

#[test]
fn test_error_user_messages() {
    let conn_err = Port42Error::Connection("localhost:42".to_string());
    let msg = conn_err.user_message();
    // Should use help_text formatting
    assert!(msg.contains("daemon") || msg.contains("connection"));
    
    let daemon_err = Port42Error::Daemon("Test error".to_string());
    let msg = daemon_err.user_message();
    assert!(msg.contains("Test error") || msg.contains("help"));
}

#[test]
fn test_timestamp_millis() {
    let ts1 = timestamp_millis();
    std::thread::sleep(std::time::Duration::from_millis(10));
    let ts2 = timestamp_millis();
    
    assert!(ts2 > ts1);
    assert!(ts2 - ts1 >= 10);
}

#[test]
fn test_format_timestamp() {
    let now = timestamp_millis();
    // Current time or very recent should show as seconds ago
    let formatted_now = format_timestamp(now);
    assert!(formatted_now.contains("seconds ago") || formatted_now == "just now");
    
    let thirty_secs_ago = now - 30_000;
    let formatted = format_timestamp(thirty_secs_ago);
    assert!(formatted.contains("seconds ago"));
    
    let two_mins_ago = now - 120_000;
    let formatted = format_timestamp(two_mins_ago);
    assert!(formatted.contains("minutes ago"));
    
    let two_hours_ago = now - 7200_000;
    let formatted = format_timestamp(two_hours_ago);
    assert!(formatted.contains("hours ago"));
    
    let two_days_ago = now - 172800_000;
    let formatted = format_timestamp(two_days_ago);
    assert!(formatted.contains("days ago"));
}

#[test]
fn test_extract_session_id() {
    // Should recognize session IDs
    assert_eq!(extract_session_id("cli-123456"), Some("cli-123456".to_string()));
    assert_eq!(extract_session_id("test-session-123"), Some("test-session-123".to_string()));
    assert_eq!(extract_session_id("abc_123"), Some("abc_123".to_string()));
    assert_eq!(extract_session_id("12345"), Some("12345".to_string()));
    
    // Should not recognize regular text as session IDs
    assert_eq!(extract_session_id("hello world"), None);
    assert_eq!(extract_session_id("this is a test message"), None);
    // This contains dashes but no numbers, so it's still recognized as an ID
    assert_eq!(extract_session_id("no-numbers-here"), Some("no-numbers-here".to_string()));
    
    // Edge cases
    assert_eq!(extract_session_id(""), None);
    assert_eq!(extract_session_id("a very long string that is definitely not a session id"), None);
}