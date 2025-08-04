use port42::protocol::{PossessRequest, PossessResponse, RequestBuilder, ResponseParser};
use serde_json::json;

#[test]
fn test_possess_request_builder() {
    let request = PossessRequest {
        agent: "@ai-engineer".to_string(),
        message: "test message".to_string(),
    };
    
    let daemon_request = request.build_request("test-123".to_string()).unwrap();
    
    assert_eq!(daemon_request.request_type, "possess");
    assert_eq!(daemon_request.id, "test-123");
    assert_eq!(daemon_request.payload["agent"], "@ai-engineer");
    assert_eq!(daemon_request.payload["message"], "test message");
}

#[test]
fn test_possess_response_parser() {
    // Test basic response
    let data = json!({
        "message": "Hello from AI",
        "session_id": "session-123",
        "agent": "@ai-muse",
        "command_generated": false
    });
    
    let response = PossessResponse::parse_response(&data).unwrap();
    
    assert_eq!(response.message, "Hello from AI");
    assert_eq!(response.session_id, "session-123");
    assert_eq!(response.agent, "@ai-muse");
    assert!(!response.command_generated);
    assert!(response.command_spec.is_none());
}

#[test]
fn test_possess_response_with_command() {
    let data = json!({
        "message": "I created a command for you",
        "session_id": "session-456",
        "agent": "@ai-engineer",
        "command_generated": true,
        "command_spec": {
            "name": "hello-world",
            "description": "Prints hello world",
            "language": "bash"
        }
    });
    
    let response = PossessResponse::parse_response(&data).unwrap();
    
    assert!(response.command_generated);
    assert!(response.command_spec.is_some());
    
    let spec = response.command_spec.unwrap();
    assert_eq!(spec.name, "hello-world");
    assert_eq!(spec.description, "Prints hello world");
    assert_eq!(spec.language, "bash");
}

#[test]
fn test_possess_response_with_artifact() {
    let data = json!({
        "message": "I created an artifact",
        "session_id": "session-789",
        "agent": "@ai-muse",
        "artifact_generated": true,
        "artifact_spec": {
            "name": "readme",
            "type": "document",
            "path": "/artifacts/document/readme.md",
            "description": "Project documentation",
            "format": "md"
        }
    });
    
    let response = PossessResponse::parse_response(&data).unwrap();
    
    assert!(response.artifact_generated);
    assert!(response.artifact_spec.is_some());
    
    let spec = response.artifact_spec.unwrap();
    assert_eq!(spec.name, "readme");
    assert_eq!(spec.artifact_type, "document");
    assert_eq!(spec.path, "/artifacts/document/readme.md");
}