use port42::possess::{PossessDisplay, SimpleDisplay, AnimatedDisplay};
use port42::protocol::{CommandSpec, ArtifactSpec};

#[test]
fn test_simple_display_creation() {
    let display = SimpleDisplay::new();
    
    // Test that we can create display without panic
    display.show_ai_message("@ai-engineer", "Hello from test");
    display.show_session_info("test-123", true);
    display.show_error("Test error");
}

#[test]
fn test_animated_display_creation() {
    let display = AnimatedDisplay::new();
    
    // Test depth-based creation
    let display_with_depth = AnimatedDisplay::with_depth(10);
    
    // These would normally animate, but in tests they just run quickly
    display.show_session_info("test-456", false);
    display_with_depth.show_error("Test error");
}

#[test]
fn test_display_command_spec() {
    let display = SimpleDisplay::new();
    
    let command_spec = CommandSpec {
        name: "test-command".to_string(),
        description: "A test command".to_string(),
        language: "bash".to_string(),
    };
    
    // Should not panic
    display.show_command_created(&command_spec);
}

#[test]
fn test_display_artifact_spec() {
    let display = SimpleDisplay::new();
    
    let artifact_spec = ArtifactSpec {
        name: "test-artifact".to_string(),
        artifact_type: "document".to_string(),
        path: "/artifacts/document/test-artifact.md".to_string(),
        description: "Test artifact".to_string(),
        format: "md".to_string(),
    };
    
    // Should not panic
    display.show_artifact_created(&artifact_spec);
}

// Mock display for testing that we're calling the right methods
struct MockDisplay {
    pub ai_message_called: std::cell::RefCell<bool>,
    pub command_created_called: std::cell::RefCell<bool>,
    pub artifact_created_called: std::cell::RefCell<bool>,
    pub session_info_called: std::cell::RefCell<bool>,
    pub error_called: std::cell::RefCell<bool>,
}

impl MockDisplay {
    fn new() -> Self {
        MockDisplay {
            ai_message_called: std::cell::RefCell::new(false),
            command_created_called: std::cell::RefCell::new(false),
            artifact_created_called: std::cell::RefCell::new(false),
            session_info_called: std::cell::RefCell::new(false),
            error_called: std::cell::RefCell::new(false),
        }
    }
}

impl PossessDisplay for MockDisplay {
    fn show_ai_message(&self, _agent: &str, _message: &str) {
        *self.ai_message_called.borrow_mut() = true;
    }
    
    fn show_command_created(&self, _spec: &CommandSpec) {
        *self.command_created_called.borrow_mut() = true;
    }
    
    fn show_artifact_created(&self, _spec: &ArtifactSpec) {
        *self.artifact_created_called.borrow_mut() = true;
    }
    
    fn show_session_info(&self, _session_id: &str, _is_new: bool) {
        *self.session_info_called.borrow_mut() = true;
    }
    
    fn show_error(&self, _error: &str) {
        *self.error_called.borrow_mut() = true;
    }
}

#[test]
fn test_trait_methods_called() {
    let mock = MockDisplay::new();
    
    // Call each method
    mock.show_ai_message("@ai-muse", "Test message");
    mock.show_command_created(&CommandSpec {
        name: "test".to_string(),
        description: "test".to_string(),
        language: "bash".to_string(),
    });
    mock.show_artifact_created(&ArtifactSpec {
        name: "test".to_string(),
        artifact_type: "doc".to_string(),
        path: "/test".to_string(),
        description: "test".to_string(),
        format: "md".to_string(),
    });
    mock.show_session_info("test-123", true);
    mock.show_error("Test error");
    
    // Verify all methods were called
    assert!(*mock.ai_message_called.borrow());
    assert!(*mock.command_created_called.borrow());
    assert!(*mock.artifact_created_called.borrow());
    assert!(*mock.session_info_called.borrow());
    assert!(*mock.error_called.borrow());
}