#[cfg(test)]
mod tests {
    use super::*;
    use crate::protocol::possess::{PossessRequest};
    use serde_json::json;

    #[test]
    fn test_possess_request_without_memory_context() {
        let request = PossessRequest {
            agent: "@ai-engineer".to_string(),
            message: "Hello".to_string(),
            memory_context: None,
        };
        
        let daemon_request = request.build_request("test-id".to_string()).unwrap();
        let payload = daemon_request.payload;
        
        assert_eq!(payload["agent"], "@ai-engineer");
        assert_eq!(payload["message"], "Hello");
        assert!(payload.get("memory_context").is_none());
    }

    #[test] 
    fn test_possess_request_with_memory_context() {
        let memory_context = vec![
            "=== Memory cli-123 ===\nuser: Previous question\nassistant: Previous response".to_string(),
            "=== Memory cli-456 ===\nuser: Another question\nassistant: Another response".to_string(),
        ];
        
        let request = PossessRequest {
            agent: "@ai-engineer".to_string(), 
            message: "Follow up question".to_string(),
            memory_context: Some(memory_context.clone()),
        };
        
        let daemon_request = request.build_request("test-id".to_string()).unwrap();
        let payload = daemon_request.payload;
        
        assert_eq!(payload["agent"], "@ai-engineer");
        assert_eq!(payload["message"], "Follow up question");
        assert_eq!(payload["memory_context"], json!(memory_context));
    }

    #[test]
    fn test_memory_context_serialization() {
        let contexts = vec!["context1".to_string(), "context2".to_string()];
        let serialized = serde_json::to_string(&contexts).unwrap();
        let deserialized: Vec<String> = serde_json::from_str(&serialized).unwrap();
        
        assert_eq!(contexts, deserialized);
    }
}