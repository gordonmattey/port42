use port42::possess::{determine_session_id, PossessDisplay, SimpleDisplay};
use port42::protocol::{CommandSpec, ArtifactSpec};
use std::cell::RefCell;

// Mock display that tracks calls
struct MockDisplay {
    pub ai_messages: RefCell<Vec<(String, String)>>,
    pub commands: RefCell<Vec<CommandSpec>>,
    pub artifacts: RefCell<Vec<ArtifactSpec>>,
    pub session_infos: RefCell<Vec<(String, bool)>>,
    pub errors: RefCell<Vec<String>>,
}

impl MockDisplay {
    fn new() -> Self {
        MockDisplay {
            ai_messages: RefCell::new(vec![]),
            commands: RefCell::new(vec![]),
            artifacts: RefCell::new(vec![]),
            session_infos: RefCell::new(vec![]),
            errors: RefCell::new(vec![]),
        }
    }
}

impl PossessDisplay for MockDisplay {
    fn show_ai_message(&self, agent: &str, message: &str) {
        self.ai_messages.borrow_mut().push((agent.to_string(), message.to_string()));
    }
    
    fn show_command_created(&self, spec: &CommandSpec) {
        self.commands.borrow_mut().push(spec.clone());
    }
    
    fn show_artifact_created(&self, spec: &ArtifactSpec) {
        self.artifacts.borrow_mut().push(spec.clone());
    }
    
    fn show_session_info(&self, session_id: &str, is_new: bool) {
        self.session_infos.borrow_mut().push((session_id.to_string(), is_new));
    }
    
    fn show_error(&self, error: &str) {
        self.errors.borrow_mut().push(error.to_string());
    }
}

#[test]
fn test_determine_session_id_new() {
    let (id, is_new) = determine_session_id(None);
    assert!(is_new);
    assert!(id.starts_with("cli-"));
    
    // Should contain timestamp
    let timestamp_part = id.strip_prefix("cli-").unwrap();
    assert!(timestamp_part.parse::<u128>().is_ok());
}

#[test]
fn test_determine_session_id_existing() {
    let existing_id = "test-session-123".to_string();
    let (id, is_new) = determine_session_id(Some(existing_id.clone()));
    assert!(!is_new);
    assert_eq!(id, existing_id);
}

#[test]
fn test_session_handler_creation() {
    // This test would require a real daemon client, so we just test that it compiles
    // In a real test, we'd use a mock client
    let _display = Box::new(SimpleDisplay::new());
    // Would create like: SessionHandler::with_display(client, display);
}

#[test]
fn test_session_handler_display_session_info() {
    let mock_display = MockDisplay::new();
    
    // Test new session
    mock_display.show_session_info("test-123", true);
    
    // Test existing session
    mock_display.show_session_info("test-456", false);
    
    // Check results
    let infos = mock_display.session_infos.borrow();
    assert_eq!(infos.len(), 2);
    assert_eq!(infos[0], ("test-123".to_string(), true));
    assert_eq!(infos[1], ("test-456".to_string(), false));
}