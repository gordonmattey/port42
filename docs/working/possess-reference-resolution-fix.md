# Fix Possess Mode Reference Resolution Architecture

## Problem Statement

Currently, `possess` and `declare` modes handle references inconsistently:

- **declare mode**: ✅ Sends references to daemon for server-side resolution
- **possess mode**: ❌ Resolves references client-side, then sends `references: None` to daemon

This architectural inconsistency prevents possess mode from leveraging daemon reference resolution capabilities and blocks the search content loading enhancement.

## Current Architecture Issues

### Possess Mode Flow (Broken)
```rust
// possess.rs
1. Client receives --ref parameters
2. Client calls resolve_references_for_context()  
3. Client makes individual read_path/search requests to daemon
4. Client formats resolved content as strings
5. Client sends PossessRequest with references: None + memory_context: resolved_strings
6. Daemon processes possess without knowledge of original references
```

### Declare Mode Flow (Correct)  
```rust
// declare.rs  
1. Client receives --ref parameters
2. Client parses references into Reference structs
3. Client sends DeclareRelationRequest with references: parsed_refs
4. Daemon resolves all references server-side
5. Daemon uses resolved content in tool creation
```

### Code Evidence
**Possess mode client-side resolution:**
```rust
// possess.rs:65
let reference_contexts = resolve_references_for_context(&mut client, refs)?;

// possess.rs:33 - references field unused!
references: None,
```

**Declare mode daemon-side resolution:**
```rust
// declare.rs:47
let request = DeclareRelationRequest { relation, references: parsed_refs, user_prompt: prompt };
```

## Target Architecture

### Unified Reference Resolution Flow
```rust
1. Client receives --ref parameters
2. Client parses references into Reference structs (same as declare)
3. Client sends request with references: parsed_refs (same as declare)  
4. Daemon resolves all references server-side (reuse existing logic)
5. Daemon injects resolved content into AI context
6. Both modes use identical reference resolution pipeline
```

### Benefits
- **Consistency**: Same reference behavior in both modes
- **Performance**: Single daemon round-trip instead of N+1 requests  
- **Maintainability**: One reference resolution codebase to maintain
- **Extensibility**: Search content loading enhancement works for both modes
- **Debugging**: Centralized reference resolution logging in daemon

## Implementation Plan

### Phase 1: Update Client Protocol (Combined Client + Protocol Changes)
**Locations**: 
- `cli/src/commands/possess.rs`
- `cli/src/protocol/possess.rs`

**Changes**:
1. **Remove client-side resolution**:
   - Delete `resolve_references_for_context()` function
   - Remove individual daemon requests for reference resolution

2. **Add reference parsing** (reuse from declare.rs):
   ```rust
   // Parse references like declare mode does
   let parsed_refs = if let Some(ref_strings) = references {
       let mut refs = Vec::new();
       for ref_str in ref_strings {
           refs.push(Reference::from_string(&ref_str)?);
       }
       Some(refs)
   } else {
       None
   };
   ```

3. **Update PossessRequest protocol**:
   ```rust
   // protocol/possess.rs - change from:
   references: None,
   // to:
   references: parsed_refs,
   ```

4. **Clean up memory_context**:
   - Remove memory_context field if only used for client-resolved references
   - Stop injecting resolved references as memory_context strings

### Phase 2: Update Daemon Possess Handler
**Location**: `daemon/server.go` (possess request handler)

**Changes**:
1. **Add reference resolution**:
   - Extract references from PossessRequest  
   - Call existing reference resolution logic (same as declare mode)
   - Resolve each reference to content

2. **Inject resolved content**:
   ```go
   // Before sending to AI:
   if request.References != nil {
       resolvedContext := resolveReferences(request.References)
       // Add to AI conversation context
       aiContext = append(aiContext, resolvedContext...)
   }
   ```

3. **Reuse existing components**:
   - Leverage same reference resolver as declare mode
   - Use same VFS path resolution (`ResolvePath`)
   - Apply same error handling patterns


## Testing Strategy

### Pre-Implementation Baseline
1. **Current possess behavior**:
   ```bash
   PORT42_DEBUG=1 port42 possess @ai-engineer "test message" --ref p42:/commands/git-status
   # Document: network requests, context received, AI knowledge
   ```

2. **Current declare behavior** (working correctly):
   ```bash
   PORT42_DEBUG=1 port42 declare tool test-ref --transforms test --ref p42:/commands/git-status  
   # Document: how daemon resolves references
   ```

### Implementation Testing
1. **Reference loading verification**:
   - AI should have access to reference content
   - Simple test: Ask AI what was loaded from the reference
   - Debug logs should show daemon-side reference resolution

3. **Error handling**:
   - Invalid references should fail gracefully  
   - Missing reference targets should be handled consistently

### Test Cases
1. **Single file reference**: `--ref p42:/commands/git-status`
2. **Multiple references**: `--ref p42:/commands/tool1 --ref file:./doc.md`
3. **Search reference**: `--ref search:architecture`
4. **Invalid reference**: `--ref p42:/nonexistent`
5. **No references**: Ensure normal possess still works

## Implementation Details

### Reference Parsing (Reuse from declare.rs)
```rust
fn parse_references(ref_strings: Vec<String>) -> Result<Vec<Reference>> {
    let mut refs = Vec::new();
    for ref_str in ref_strings {
        match Reference::from_string(&ref_str) {
            Ok(reference) => refs.push(reference),
            Err(e) => bail!("Invalid reference {}: {}", ref_str, e),
        }
    }
    Ok(refs)
}
```

### Daemon Reference Resolution (Extend existing)
```go
// daemon/reference_resolver.go (if separate) or daemon/server.go
func (s *Server) resolvePossessReferences(refs []Reference) ([]string, error) {
    var contexts []string
    
    for _, ref := range refs {
        switch ref.Type {
        case "search":
            content, err := s.resolveSearchReference(ref.Target)
        case "p42", "file":
            content, err := s.ResolvePath(ref.Target)
        default:
            return nil, fmt.Errorf("unsupported reference type: %s", ref.Type)
        }
        
        if err != nil {
            log.Printf("Failed to resolve reference %s: %v", ref.Target, err)
            continue // or fail fast, depending on requirements
        }
        
        contexts = append(contexts, formatReferenceContext(ref, content))
    }
    
    return contexts, nil
}
```

## Risk Mitigation

### Backward Compatibility
- **Risk**: Breaking existing possess users
- **Mitigation**: Thorough testing with existing possess patterns
- **Fallback**: Keep current behavior behind feature flag during transition

### Performance Impact  
- **Risk**: Daemon processing adds latency
- **Mitigation**: Daemon processing is more efficient than N+1 client requests
- **Monitoring**: Measure before/after performance

### Reference Resolution Failures
- **Risk**: Daemon failures affect possess mode
- **Mitigation**: Reuse proven declare mode error handling
- **Graceful degradation**: Continue with partial context if some references fail

## Success Criteria

1. **Architectural Consistency**: Possess and declare use identical reference resolution
2. **Performance**: Single possess request instead of N+1 requests  
3. **Functionality**: AI receives same reference content as before
4. **Error Handling**: Graceful handling of invalid/missing references
5. **Debuggability**: Clear daemon-side logging of reference resolution
6. **Foundation**: Ready for search content loading enhancement

## Future Enhancements Enabled

Once this architectural fix is complete:

1. **Search content loading**: Apply score-based filtering to both modes
2. **Reference caching**: Cache resolved content in daemon  
3. **Reference preprocessing**: Optimize/transform content before AI injection
4. **Reference analytics**: Track reference usage patterns
5. **Advanced references**: Support new reference types (url:, tool:, etc.)

## Implementation Phases

### Phase 1: Daemon Reference Resolution
- Add reference resolution to daemon possess handler
- Reuse existing declare mode reference resolution logic
- Test by manually crafting possess request with references
- Verify with simple "what was loaded?" validation

### Phase 2: Client Integration
- Update possess mode to parse and send references to daemon
- Clean up client-side resolution code and protocol
- Update PossessRequest to include references field