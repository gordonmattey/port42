# Phase 2: Reference Resolvers - Implementation Plan

## Scope Definition

Phase 2 builds on Phase 1's reference syntax by implementing **context resolution engines** that transform references into actionable context for AI tool generation.

### Core Objective
Convert reference specifications (e.g., `--ref search:"nginx errors"`, `--ref tool:log-parser`) into rich contextual information that enhances AI tool generation with relevant background knowledge.

### Reference Type Resolvers

#### 1. Search Resolver (`search:query`)
**Purpose**: Resolve search queries into relevant knowledge fragments
**Implementation**: 
- Use existing search infrastructure in daemon
- Query crystallized knowledge base with the search term
- Return top N relevant results as context strings
- Format: "Search Results for 'query': [result1, result2, ...]"

#### 2. Tool Resolver (`tool:tool-name`)
**Purpose**: Load existing tool definitions and their properties
**Implementation**:
- Query relations storage for tool-type relations matching name
- Extract tool properties, transforms, and historical usage
- Include generated commands if available
- Format: "Tool Definition: {name: X, transforms: Y, commands: Z}"

#### 3. Memory Resolver (`memory:session-id`)
**Purpose**: Load conversation history from specific sessions
**Implementation**:
- Use existing memory system to fetch session content
- Extract key conversation points and decisions
- Limit to most relevant parts to avoid context overflow
- Format: "Memory from session-id: [conversation summary]"

#### 4. File Resolver (`file:path`)
**Purpose**: Read file contents from virtual or real filesystem
**Implementation**:
- Use existing VFS (virtual file system) infrastructure
- Support both crystallized paths (/tools/X) and real paths
- Handle different file types appropriately
- Format: "File Content from path: [content]"

#### 5. URL Resolver (`url:https://...`)
**Purpose**: Fetch and summarize web content
**Implementation**:
- HTTP GET request to URL
- Extract meaningful content (strip HTML, focus on text)
- Limit content size to prevent context overflow
- Format: "URL Content from url: [summary]"

## Architecture Design

### Core Components

1. **Reference Resolution Engine** (`daemon/resolvers/`)
   - `resolver.go` - Main resolver interface and orchestration
   - `search_resolver.go` - Search query resolution
   - `tool_resolver.go` - Tool definition resolution  
   - `memory_resolver.go` - Memory session resolution
   - `file_resolver.go` - File content resolution
   - `url_resolver.go` - URL content resolution

2. **Context Synthesis** (`daemon/context/`)
   - `synthesizer.go` - Combine resolved references into unified context
   - `formatter.go` - Format context for AI consumption
   - `limiter.go` - Prevent context overflow

3. **Integration Points**
   - Modify `daemon/server.go` handleDeclareRelation() to resolve references
   - Add context to AI generation requests
   - Maintain backward compatibility (no references = no context)

### Resolution Flow

```
References Input: [--ref search:"nginx" --ref tool:log-parser]
    ↓
Phase 1 Validation (existing): Syntax validated, stored in relation
    ↓
Phase 2 Resolution: Each reference type resolved to context
    ↓
Context Synthesis: Multiple contexts combined intelligently
    ↓
AI Generation: Enhanced with resolved context
    ↓
Tool Creation: Better tools with rich background knowledge
```

### Error Handling Strategy

1. **Graceful Degradation**: If a reference fails to resolve, continue with others
2. **Partial Resolution**: Return what can be resolved, log what cannot
3. **Context Overflow Protection**: Limit total context size, prioritize by importance
4. **Resolution Timeouts**: Prevent hanging on slow resolvers (especially URL)

## Implementation Steps

### Step 2.1: Core Resolution Infrastructure
- Create resolver interface and orchestration
- Implement basic context synthesis
- Add integration point in handleDeclareRelation()
- **Files**: `daemon/resolvers/resolver.go`, `daemon/context/synthesizer.go`

### Step 2.2: Search and Tool Resolvers  
- Implement search resolver using existing search infrastructure
- Implement tool resolver using relations storage
- **Files**: `daemon/resolvers/search_resolver.go`, `daemon/resolvers/tool_resolver.go`

### Step 2.3: Memory and File Resolvers
- Implement memory resolver using existing memory system  
- Implement file resolver using VFS infrastructure
- **Files**: `daemon/resolvers/memory_resolver.go`, `daemon/resolvers/file_resolver.go`

### Step 2.4: URL Resolver and Context Management
- Implement URL resolver with HTTP client
- Add context size limiting and overflow protection
- **Files**: `daemon/resolvers/url_resolver.go`, `daemon/context/limiter.go`

### Step 2.5: Integration and Optimization
- Complete integration with AI generation pipeline
- Add comprehensive logging and debugging
- Performance optimization for resolution speed

## Testing Strategy

### Unit Tests
- Individual resolver tests for each reference type
- Context synthesis and formatting tests  
- Error handling and edge case tests
- Mock external dependencies (HTTP, filesystem)

### Integration Tests
- End-to-end resolution flow tests
- Multiple reference type combinations
- Context overflow and limiting tests
- Performance and timeout tests

### Test Files Structure
```
daemon/resolvers/
├── resolver_test.go
├── search_resolver_test.go  
├── tool_resolver_test.go
├── memory_resolver_test.go
├── file_resolver_test.go
└── url_resolver_test.go

daemon/test_phase2_reference_resolvers.sh
```

### Test Scenarios

1. **Single Reference Resolution**
   ```bash
   # Test each resolver individually
   port42 declare tool test-search --transforms logs --ref search:"error patterns"
   port42 declare tool test-tool --transforms parser --ref tool:existing-parser
   port42 declare tool test-memory --transforms analysis --ref memory:session-123
   port42 declare tool test-file --transforms config --ref file:/config/app.json
   port42 declare tool test-url --transforms docs --ref url:https://docs.example.com/api
   ```

2. **Multiple Reference Resolution**
   ```bash
   # Test combination resolution and synthesis
   port42 declare tool advanced-analyzer --transforms "logs,metrics,alerts" \
     --ref search:"performance issues" \
     --ref tool:log-parser \
     --ref memory:troubleshooting-session \
     --ref file:/monitoring/config.yaml
   ```

3. **Error Handling**
   ```bash
   # Test graceful degradation
   port42 declare tool robust-tool --transforms analysis \
     --ref search:"nonexistent query" \
     --ref tool:missing-tool \
     --ref memory:invalid-session \
     --ref file:/nonexistent/path \
     --ref url:https://invalid-domain-xyz123.com/
   ```

4. **Context Overflow Protection**
   ```bash
   # Test context size limiting
   port42 declare tool large-context --transforms processing \
     --ref url:https://very-large-document.com/content \
     --ref memory:very-long-session \
     --ref file:/huge/config/file.json
   ```

### Success Criteria

- All reference types resolve correctly when valid
- Invalid references degrade gracefully without breaking tool generation
- Multiple references synthesize into coherent context
- Context size limits prevent overflow
- Resolution performance stays under reasonable timeouts
- Generated tools show improvement with rich context

## Timeline and Dependencies

**Dependencies**: 
- Phase 1 implementation (completed)
- Existing search, memory, and VFS infrastructure
- Relations storage system

**Estimated Effort**: 
- Step 2.1: Core infrastructure (4-6 hours)
- Step 2.2: Search/Tool resolvers (3-4 hours)  
- Step 2.3: Memory/File resolvers (3-4 hours)
- Step 2.4: URL resolver + context management (4-5 hours)
- Step 2.5: Integration + optimization (2-3 hours)
- Testing: (6-8 hours)

**Total**: 22-30 hours over multiple sessions

## Risk Mitigation

1. **Context Explosion**: Implement strict size limits and intelligent truncation
2. **Resolution Performance**: Add timeouts and async resolution where possible  
3. **External Dependencies**: URL resolver may fail - ensure graceful handling
4. **Memory Leaks**: Careful memory management in context synthesis
5. **Backward Compatibility**: Ensure existing declare commands continue working

---

This plan transforms Phase 1's reference syntax into actionable contextual intelligence, enabling the Reality Compiler to generate more sophisticated and contextually aware tools.