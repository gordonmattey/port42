use port42::display::{TableBuilder, format_size, format_timestamp_relative, truncate_string, StatusIndicator};

#[test]
fn test_table_builder() {
    let mut table = TableBuilder::new();
    table.add_header(vec!["Name", "Status", "Size"])
         .add_row(vec!["test.txt".to_string(), "Active".to_string(), "1.2K".to_string()])
         .add_row(vec!["config.json".to_string(), "Inactive".to_string(), "456B".to_string()]);
    
    // Should not panic
    let output = table.to_string();
    assert!(output.contains("Name"));
    assert!(output.contains("test.txt"));
}

#[test]
fn test_format_size() {
    assert_eq!(format_size(0), "0B");
    assert_eq!(format_size(512), "512B");
    assert_eq!(format_size(1024), "1.0K");
    assert_eq!(format_size(1536), "1.5K");
    assert_eq!(format_size(1048576), "1.0M");
    assert_eq!(format_size(1073741824), "1.0G");
}

#[test]
fn test_truncate_string() {
    assert_eq!(truncate_string("hello", 10), "hello");
    assert_eq!(truncate_string("hello world", 8), "hello...");
    assert_eq!(truncate_string("short", 5), "short");
    assert_eq!(truncate_string("exactly ten", 11), "exactly ten");
}

#[test]
fn test_format_timestamp_relative() {
    use std::time::{SystemTime, UNIX_EPOCH};
    
    let now = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis() as u64;
    
    // Current time should be "0 seconds ago"
    assert_eq!(format_timestamp_relative(now), "0 seconds ago");
    
    // 30 seconds ago
    let thirty_secs_ago = now - 30_000;
    assert!(format_timestamp_relative(thirty_secs_ago).contains("seconds ago"));
    
    // 5 minutes ago
    let five_mins_ago = now - 300_000;
    assert!(format_timestamp_relative(five_mins_ago).contains("minutes ago"));
}

#[test]
fn test_status_indicators() {
    // Just verify they don't panic and return colored strings
    let _ = StatusIndicator::success();
    let _ = StatusIndicator::error();
    let _ = StatusIndicator::warning();
    let _ = StatusIndicator::info();
    let _ = StatusIndicator::running();
    let _ = StatusIndicator::stopped();
}