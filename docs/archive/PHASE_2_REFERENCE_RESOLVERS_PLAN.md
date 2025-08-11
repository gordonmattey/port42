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

#### 3. Memory Access (via P42 VFS: `p42:/memory/session-id`)
**Purpose**: Load conversation history from specific sessions
**Implementation**:
- Access memory sessions through P42 VFS unified interface
- Use P42 resolver's storage integration for session lookup
- Extract key conversation points and decisions from storage
- Format: "P42 Memory Session: [conversation summary + metadata]"

#### 4. File Resolver (`file:path`)
**Purpose**: Read local filesystem files with security boundaries
**Implementation**:
- Sanitized local file access with path traversal protection
- Working directory boundary enforcement
- Support for text files, code files, config files (JSON/YAML)
- Size limits (default 1MB) and timeout protection
- Format: "File Content from path: [content + metadata]"

#### 5. P42 Resolver (`p42:path`) 
**Purpose**: Access Port 42 VFS and crystallized knowledge
**Implementation**:
- **VFS Path Resolution**: Convert `/tools/log-parser` to storage object IDs
- **Storage Integration**: Use `Storage.Read(objectID)` for content access
- **Search Fallback**: Use `Storage.SearchObjects()` for path-based lookup
- **Relations Integration**: Access tool definitions via Relations store
- **Access Methods**:
  - Direct symlink resolution (if VFS symlinks exist)
  - Search-based resolution using path patterns
  - Relations-based resolution for `/tools/*` paths
- Format: "P42 Content from path: [content + crystallized metadata]"

#### 6. URL Resolver (`url:https://...`) ✅ COMPLETE
**Purpose**: Fetch and summarize web content with artifact caching
**Implementation**:
- HTTP GET request to URL with intelligent caching
- URL artifact Relations with 24-hour TTL
- Cache-first resolution with graceful fallback
- Extract meaningful content (strip HTML, focus on text)
- Format: "URL Content from url: [content + cache status]"

## Architecture Design

### Core Components

1. **Reference Resolution Engine** (`daemon/resolution/`)
   - `service.go` - Main resolver orchestration ✅ IMPLEMENTED
   - `resolvers.go` - All resolver implementations ✅ IMPLEMENTED
   - `interface.go` - Resolver interfaces and types ✅ IMPLEMENTED
   - Search resolver ✅ IMPLEMENTED
   - Tool resolver ✅ IMPLEMENTED  
   - File resolver (`file:`) ✅ IMPLEMENTED  
   - P42 resolver (`p42:`) ✅ IMPLEMENTED (includes memory via `/memory/` paths)
   - URL resolver ✅ IMPLEMENTED

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

### Step 2.3: File and P42 Resolvers Implementation ✅ COMPLETE
- Implement enhanced file resolver for local filesystem access ✅ COMPLETE
- Implement P42 resolver for VFS/crystallized knowledge access ✅ COMPLETE
- **Files**: Enhanced `daemon/resolution/resolvers.go`, `daemon/server.go` ✅ COMPLETE

#### File Resolver (`file:path`) Implementation Details

**Security Architecture:**
- Path sanitization to prevent `../` directory traversal attacks
- Working directory boundary enforcement (only access files within project)
- File extension whitelist for security (`.txt`, `.md`, `.json`, `.yaml`, `.log`, code files)
- Size limits (default 1MB) to prevent memory exhaustion

**File Access Strategy:**
```go
func (r *fileResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
    // 1. Security: Sanitize and validate path
    cleanPath := filepath.Clean(target)
    if strings.Contains(cleanPath, "..") {
        return errorResult("path traversal not allowed")
    }
    
    // 2. Size check before reading
    fileInfo, err := os.Stat(cleanPath)
    if fileInfo.Size() > maxFileSize {
        return errorResult("file too large")
    }
    
    // 3. Content processing based on file type
    content := processFileContent(cleanPath, fileInfo)
    return successResult(content)
}
```

#### P42 Resolver (`p42:path`) Implementation Details

**VFS Access Architecture:**
```go
func (r *p42Resolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
    // target examples: "/tools/log-parser", "/commands/my-tool"
    
    // Method 1: Direct storage lookup (fastest)
    if objectID := r.resolveDirectPath(target); objectID != "" {
        return r.loadStorageObject(objectID)
    }
    
    // Method 2: Search-based resolution
    results := r.storage.SearchObjects(extractSearchTerm(target), filters)
    if len(results) > 0 {
        return r.loadStorageObject(results[0].ObjectID)
    }
    
    // Method 3: Relations-based resolution (for tools)
    if strings.HasPrefix(target, "/tools/") {
        return r.resolveToolRelation(target)
    }
    
    return notFoundResult()
}
```

**P42 Path Resolution Methods:**

1. **Direct Storage Access**: For crystallized knowledge with known object IDs
2. **Search Integration**: Use existing search infrastructure with path-based queries  
3. **Relations Resolution**: Access tool definitions and generated commands
4. **Metadata Enrichment**: Include creation dates, tags, usage statistics

**VFS Path Examples:**
- `/tools/log-parser` → Relations store tool definition + generated commands
- `/commands/my-script` → Direct storage object via symlink resolution
- `/memory/cli-1234` → Memory session access via storage integration
- `/knowledge/nginx-config` → Search-based content resolution

### Step 2.4: URL Resolver and Context Management ✅ COMPLETE
- URL resolver with HTTP client and intelligent caching ✅ IMPLEMENTED  
- URL artifact Relations with 24-hour TTL ✅ IMPLEMENTED
- Cache-first resolution with graceful fallback ✅ IMPLEMENTED
- JSON type conversion bug fixes ✅ IMPLEMENTED
- Context size limiting and overflow protection ⚠️ PENDING

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
   port42 declare tool test-search --ref search:"error patterns"           # ✅ IMPLEMENTED
   port42 declare tool test-tool --ref tool:existing-parser                # ✅ IMPLEMENTED  
   port42 declare tool test-memory --ref p42:/memory/session-123           # ✅ IMPLEMENTED (via P42 VFS)
   port42 declare tool test-local-file --ref file:./config/app.json        # ✅ IMPLEMENTED
   port42 declare tool test-p42-tool --ref p42:/tools/log-parser           # ✅ IMPLEMENTED
   port42 declare tool test-p42-cmd --ref p42:/commands/existing-tool      # ✅ IMPLEMENTED
   port42 declare tool test-url --ref url:https://docs.example.com/api     # ✅ IMPLEMENTED
   ```

2. **Multiple Reference Resolution**
   ```bash
   # Test combination resolution and synthesis
   port42 declare tool advanced-analyzer \
     --ref search:"performance issues" \
     --ref tool:log-parser \
     --ref file:./monitoring/config.yaml \
     --ref p42:/tools/base-analyzer \
     --ref url:https://docs.monitoring.com/api
   ```

3. **Error Handling**
   ```bash
   # Test graceful degradation
   port42 declare tool robust-tool --transforms analysis \
     --ref search:"nonexistent query" \
     --ref tool:missing-tool \
     --ref p42:/memory/invalid-session \
     --ref file:/nonexistent/path \
     --ref url:https://invalid-domain-xyz123.com/
   ```

4. **Context Overflow Protection**
   ```bash
   # Test context size limiting
   port42 declare tool large-context --transforms processing \
     --ref url:https://very-large-document.com/content \
     --ref p42:/memory/very-long-session \
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

## Current Implementation Status & Next Steps

### ✅ Completed Resolvers (All Working!)
1. **Search Resolver**: Crystallized knowledge search ✅ IMPLEMENTED & TESTED
2. **Tool Resolver**: Relations-based tool definition lookup ✅ IMPLEMENTED & TESTED
3. **URL Resolver**: HTTP content with intelligent caching ✅ IMPLEMENTED & TESTED
4. **File Resolver (`file:path`)**: Local filesystem access ✅ IMPLEMENTED & TESTED
5. **P42 Resolver (`p42:path`)**: VFS/crystallized knowledge + memory access ✅ IMPLEMENTED & TESTED

### 🎉 File Resolvers Implementation Complete!

#### **File Resolver (`file:path`)** ✅ FULLY IMPLEMENTED
- **Security**: Path sanitization, directory traversal protection ✅
- **Boundaries**: Working directory enforcement, size limits ✅
- **File Types**: Text, code, config files (JSON/YAML/MD) ✅
- **Use Cases**: `--ref file:./config.json`, `--ref file:./main.go` ✅
- **Testing**: JSON, Markdown, Go source files all working ✅

#### **P42 Resolver (`p42:path`)** ✅ FULLY IMPLEMENTED  
- **VFS Integration**: Access `/tools/*`, `/commands/*`, `/memory/*`, `/knowledge/*` paths ✅
- **Storage Bridge**: Convert VFS paths to storage object IDs ✅
- **Resolution Methods**: Relations store, storage search, general search ✅
- **Memory Integration**: Access sessions via `/memory/session-id` paths ✅
- **Use Cases**: `--ref p42:/tools/parser`, `--ref p42:/memory/cli-1234` ✅
- **Testing**: Tool paths, command paths, memory paths all working ✅

### Architecture Benefits
- **Unified VFS Interface**: Single `p42:` resolver handles tools, commands, memory, knowledge
- **Five-tier access**: Local files (`file:`), crystallized knowledge (`p42:`), web content (`url:`), search (`search:`), tools (`tool:`)
- **Security by design**: Each resolver has appropriate access boundaries
- **Performance optimization**: Caching where appropriate (URL artifacts)
- **Graceful degradation**: Missing references don't break tool generation
- **Simplified mental model**: Memory via P42 VFS eliminates redundant resolver types

This architecture provides comprehensive contextual intelligence for the Reality Compiler, enabling sophisticated tool generation with rich background knowledge from multiple sources through a clean, unified interface.