# Universal Prompt and Reference System Implementation Plan

*Adding `--prompt` and `--ref` support across all Port 42 commands for consistent, powerful context-driven generation*

## Architecture Overview

### **Current State**
- `declare tool`: Has `--ref`, missing `--prompt`
- `declare artifact`: Missing both `--ref` and `--prompt`  
- `possess`: Has message/prompt, missing `--ref`

### **Target State**
- All commands support both `--ref` and `--prompt`
- Unified reference resolution across tools, artifacts, and AI sessions
- Enhanced AI generation with user-provided requirements and context

## Component Map

### **CLI Layer (Rust)**
```
cli/
├── commands/
│   ├── declare.rs          # declare tool/artifact commands
│   ├── possess.rs          # AI possession sessions
│   └── common_args.rs      # NEW: Shared --ref/--prompt arguments
└── main.rs                 # CLI entry point
```

### **Protocol Layer (Go)**
```
daemon/
├── protocol.go            # JSON message structures (CLI ↔ Daemon)
│   ├── DeclareToolRequest     # ENHANCE: Add UserPrompt field
│   ├── DeclareArtifactRequest # ENHANCE: Add References + UserPrompt
│   └── PossessRequest         # ENHANCE: Add References field
└── server.go              # Request handlers
    ├── handleDeclareTool()    # ENHANCE: Process user prompts
    ├── handleDeclareArtifact() # NEW: Add reference resolution
    └── handlePossess()        # ENHANCE: Add reference resolution
```

### **Reference Resolution System (Go)**
```
daemon/resolution/
├── interface.go           # ResolutionService interface
├── service.go            # Resolution orchestration
│   ├── searchResolver        # search:"query" references
│   ├── fileResolver          # file:path references  
│   ├── p42Resolver           # p42:/path references
│   └── urlResolver           # url:https://... references
└── artifact_manager.go   # URL artifact caching
```

### **AI Generation Layer (Go)**
```
daemon/
├── tool_materializer.go  # Tool AI generation
│   └── generateTool()        # ENHANCE: Add prompt building
├── artifact_materializer.go # Artifact AI generation  
│   └── generateArtifact()    # ENHANCE: Add prompt building
├── possession.go         # AI session management
│   └── startSession()       # ENHANCE: Add reference context
└── prompt_builder.go     # NEW: Unified prompt construction
    └── buildEnhancedPrompt() # Combines base + user + references
```

### **Storage Layer (Go)**
```
daemon/
├── storage.go            # Object storage
├── relation_store.go     # Relation management
└── memory/               # Session memory storage
```

### **Data Flow Architecture**

```
┌─────────────┐    ──ref, ──prompt    ┌─────────────┐
│   CLI       │ ──────────────────────→│   Protocol  │
│   (Rust)    │                       │   (Go)      │
└─────────────┘                       └─────────────┘
                                             │
                                             ▼
┌─────────────┐    References Array    ┌─────────────┐
│  Reference  │ ←──────────────────────│   Server    │
│  Resolution │                       │   Handler   │
│  (Go)       │                       │   (Go)      │
└─────────────┘                       └─────────────┘
      │                                      │
      │ Resolved Context                     │ Enhanced Prompt
      ▼                                      ▼
┌─────────────┐    Prompt + Context    ┌─────────────┐
│   Prompt    │ ──────────────────────→│     AI      │
│   Builder   │                       │ Generation  │
│   (Go)      │                       │   (Go)      │
└─────────────┘                       └─────────────┘
                                             │
                                             ▼
                                    ┌─────────────┐
                                    │   Storage   │
                                    │   (Go)      │
                                    └─────────────┘
```

### **Reference Type Handlers**

| Reference Type | Handler | Purpose | Example |
|---------------|---------|---------|---------|
| `file:` | fileResolver | Local file access | `file:./config.json` |
| `url:` | urlResolver | Web resource fetching | `url:https://api.docs.com` |
| `search:` | searchResolver | Query-based context | `search:"error patterns"` |
| `p42:` | p42Resolver | VFS navigation | `p42:/tools/base-processor` |

### **Enhanced Request Processing Flow**

```
1. CLI Parsing
   ├── Parse --ref flags → References[]
   ├── Parse --prompt flag → UserPrompt
   └── Send to Daemon via Protocol

2. Reference Resolution  
   ├── For each reference in References[]
   ├── Route to appropriate resolver
   ├── Resolve content/context
   └── Aggregate resolved contexts

3. Prompt Enhancement
   ├── Start with base prompt (tool/artifact type)
   ├── Add resolved reference contexts  
   ├── Append user prompt requirements
   └── Generate enhanced AI prompt

4. AI Generation
   ├── Send enhanced prompt to AI
   ├── Receive generated content
   └── Store/materialize result

5. Response
   ├── Return success/failure
   └── Provide path to generated entity
```

---

## Integration Test Suite

### **Test Infrastructure Setup**

**Create comprehensive test suite that runs at every step:**

```bash
# Create test directory structure
mkdir -p tests/integration/prompt-ref-system/
mkdir -p tests/regression/

# Core test files
tests/integration/prompt-ref-system/
├── 01-basic-functionality.sh       # Core command functionality
├── 02-reference-resolution.sh      # All reference types work
├── 03-prompt-integration.sh        # User prompts in generated content
├── 04-cross-command-refs.sh        # Tools ↔ artifacts ↔ possession
├── 05-error-handling.sh            # Error conditions and edge cases
├── 06-performance.sh               # Load and performance testing
└── test-helpers.sh                 # Common test utilities

# Regression test suite
tests/regression/
├── existing-functionality.sh       # Ensure no breakage of current features
├── rule-engine-integration.sh      # Rules still work with new features
└── backwards-compatibility.sh      # Old commands still work
```

**Test Runner Script:**
```bash
#!/bin/bash
# tests/run-prompt-ref-tests.sh

set -e

echo "🧪 Running Universal Prompt & Reference Integration Tests"

# Run regression tests first
echo "📋 Running regression tests..."
./tests/regression/existing-functionality.sh
./tests/regression/rule-engine-integration.sh
./tests/regression/backwards-compatibility.sh

# Run integration tests
echo "🔧 Running integration tests..."
for test in tests/integration/prompt-ref-system/*.sh; do
    if [[ "$test" != *"test-helpers.sh" ]]; then
        echo "Running $(basename "$test")..."
        "$test"
    fi
done

echo "✅ All tests passed!"
```

---

## Implementation Steps

### **Step 1: Test Infrastructure Setup**

**Objective**: Create comprehensive test suite for continuous integration and regression testing

**Changes**:
- Create integration test directory structure
- Build test helper utilities
- Create regression test baseline
- Set up test runner script

**Files to Create**:
- `tests/integration/prompt-ref-system/test-helpers.sh` - Common test utilities
- `tests/regression/existing-functionality.sh` - Baseline functionality tests
- `tests/regression/rule-engine-integration.sh` - Rule engine compatibility
- `tests/regression/backwards-compatibility.sh` - Existing command compatibility
- `tests/run-prompt-ref-tests.sh` - Main test runner

**Test**:
```bash
# Set up test infrastructure
mkdir -p tests/integration/prompt-ref-system/ tests/regression/
chmod +x tests/run-prompt-ref-tests.sh

# Run baseline regression tests to establish current functionality
./tests/run-prompt-ref-tests.sh

# Verify test infrastructure works
echo "✅ Test infrastructure operational"
```

### **Step 2: Protocol and Backend Foundation**

**Objective**: Establish all backend data structures and protocol support for prompts and references

**Changes**:
- Add `UserPrompt` field to all request structures (`DeclareToolRequest`, `DeclareArtifactRequest`, `PossessRequest`)
- Add `References` field to `DeclareArtifactRequest` and `PossessRequest`
- Update JSON protocol message handling in server
- Create shared prompt building infrastructure

**Files to Modify**:
- `daemon/protocol.go` - Add UserPrompt and References fields to all request types
- `daemon/server.go` - Update all message handlers to accept new fields
- `daemon/prompt_builder.go` - NEW: Create unified prompt construction

**Test**:
```bash
# Verify structure changes compile and protocol accepts new parameters
cd daemon && go build .

# Run regression tests to ensure no breakage
./tests/run-prompt-ref-tests.sh

# Test new protocol fields (will fail gracefully until CLI updated)
port42 declare tool test-prompt --transforms "test" --prompt "test prompt" || echo "Expected failure - CLI not updated yet"
```

### **Step 3: CLI Parameter Addition**

**Objective**: Add `--prompt` and missing `--ref` parameters to CLI commands

**Changes**:
- Add `--prompt` flag to `declare tool` command
- Add `--ref` and `--prompt` flags to `declare artifact` command  
- Add `--ref` flag to `possess` command
- Create shared CLI argument structures
- Add integration tests for basic CLI functionality

**Files to Modify**:
- CLI argument parsing files (find with `find . -name "*.rs" -path "*/cli/*"`)
- Help text and command definitions
- `tests/integration/prompt-ref-system/01-basic-functionality.sh` - NEW

**Test**:
```bash
# Verify new parameters appear in help and basic protocol works
port42 declare tool --help | grep -E "(prompt|ref)"
port42 declare artifact --help | grep -E "(prompt|ref)"
port42 possess --help | grep -E "(prompt|ref)"

# Test basic functionality works
port42 declare tool test-prompt --transforms "test" --prompt "test prompt"

# Run full test suite including regression
./tests/run-prompt-ref-tests.sh
```

### **Step 4: Unified AI Generation Enhancement**

**Objective**: Integrate reference resolution and prompt building across all AI generation (tools, artifacts, possession)

**Changes**:
- Add reference resolution to artifact and possession flows
- Integrate unified prompt builder with all AI generation paths
- Update tool, artifact, and possession handlers to use enhanced prompts
- Handle reference resolution errors gracefully
- Add comprehensive reference resolution tests

**Files to Modify**:
- `daemon/server.go` - Update `handleDeclareTool`, `handleDeclareArtifact`, `handlePossess`
- `daemon/tool_materializer.go` - Use enhanced prompt building
- `daemon/possession.go` - Add reference resolution and prompt enhancement
- Artifact materialization code - Integrate prompt building
- `tests/integration/prompt-ref-system/02-reference-resolution.sh` - NEW
- `tests/integration/prompt-ref-system/03-prompt-integration.sh` - NEW

**Test**:
```bash
# Test all three command types with references and prompts
port42 declare tool base-tool --transforms "base"

# Test tool with prompt
port42 declare tool prompt-tool \
  --transforms "api,rest" \
  --ref p42:/tools/base-tool \
  --prompt "Create FastAPI server with marker ABC123"
port42 cat /tools/prompt-tool | grep "ABC123"

# Test artifact with references and prompt  
port42 declare artifact ref-doc \
  --artifact-type "documentation" \
  --ref p42:/tools/base-tool \
  --prompt "Documentation with marker DEF456"
port42 cat /artifacts/ref-doc | grep "DEF456"

# Test possession with references
port42 possess @ai-engineer \
  --ref p42:/tools/base-tool \
  "describe this tool with marker GHI789" | grep "GHI789"

# Run full test suite including regression
./tests/run-prompt-ref-tests.sh
```

### **Step 5: Error Handling and Validation**

**Objective**: Robust error handling for new prompt and reference combinations

**Changes**:
- Handle missing references gracefully
- Validate prompt length and content
- Provide helpful error messages for malformed references
- Handle empty prompts and reference resolution failures
- Add comprehensive error handling tests

**Files to Modify**:
- Error handling in reference resolution
- Input validation in request handlers
- User-facing error messages
- `tests/integration/prompt-ref-system/05-error-handling.sh` - NEW

**Test**:
```bash
# Test error conditions
port42 declare tool error-test --ref p42:/tools/nonexistent --prompt "test"
port42 declare artifact error-doc --ref file:/nonexistent/file.txt
port42 possess @ai-engineer --ref invalid:reference "test"

# Verify helpful error messages, not crashes
# Run full test suite including regression
./tests/run-prompt-ref-tests.sh
```

### **Step 6: Documentation and Help Updates**

**Objective**: Update all documentation to reflect new capabilities

**Changes**:
- Update command help text with examples
- Add reference type documentation
- Create usage examples for complex scenarios
- Update README with new capabilities

**Files to Modify**:
- CLI help text
- `README.md`
- Command documentation
- Example usage files

**Test**:
```bash
# Verify help text is complete and accurate
port42 declare tool --help | grep -A5 -B5 "prompt"
port42 declare artifact --help | grep -A5 -B5 "ref"
port42 possess --help | grep -A5 -B5 "ref"

# Run full test suite including regression
./tests/run-prompt-ref-tests.sh
```

### **Step 7: Integration and Performance Testing**

**Objective**: Verify system works end-to-end with complex scenarios and acceptable performance

**Changes**:
- Test tools referencing artifacts and vice versa
- Test possession sessions with mixed reference types
- Test with large numbers of references and long prompts
- Integration test with rule engine
- Verify reference chains work correctly
- Add comprehensive cross-command and performance tests

**Files to Modify**:
- `tests/integration/prompt-ref-system/04-cross-command-refs.sh` - NEW
- `tests/integration/prompt-ref-system/06-performance.sh` - NEW
- Performance test configurations
- Reference resolution edge cases

**Test**:
```bash
# Test complex reference scenarios
port42 declare tool base --transforms "base"
port42 declare artifact spec --ref p42:/tools/base --prompt "create spec"
port42 declare tool enhanced --ref p42:/artifacts/spec --prompt "implement spec"
port42 possess @ai-engineer \
  --ref p42:/tools/enhanced \
  --ref p42:/artifacts/spec \
  --ref file:./config.json \
  "help optimize this implementation"

# Test system performance and stability
for i in {1..10}; do
  port42 declare tool perf-test-$i \
    --transforms "test,performance" \
    --ref p42:/tools/base \
    --prompt "performance test iteration $i with long prompt text to test memory usage and parsing performance with extended context and multiple reference resolution" &
done
wait
port42 ls /tools | grep perf-test | wc -l  # Should be 10

# Run comprehensive test suite including all regression tests
./tests/run-prompt-ref-tests.sh

# Final validation - ensure rule engine still works with new features
port42 declare tool test-rule-integration --transforms "test,analysis"
# Should trigger viewer rule and spawn test-rule-integration-viewer
port42 ls /tools | grep test-rule-integration-viewer
```

---

## Success Criteria

- ✅ All commands support `--prompt` and `--ref` parameters
- ✅ Reference resolution works across tools, artifacts, and possession
- ✅ User prompts enhance AI generation quality
- ✅ Complex multi-reference scenarios work correctly
- ✅ Error handling is robust and user-friendly
- ✅ Documentation is complete and accurate
- ✅ Performance remains acceptable with new features
- ✅ Integration with existing rule engine works correctly

## Architecture Benefits

1. **Unified Interface**: Consistent parameters across all commands
2. **Rich Context**: Every operation can leverage full reference resolution
3. **User Control**: Direct specification of requirements beyond structured parameters
4. **Composability**: Natural chaining between tools, artifacts, and AI sessions
5. **Extensibility**: Easy to add new reference types and prompt enhancements