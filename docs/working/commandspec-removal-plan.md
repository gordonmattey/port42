# CommandSpec Legacy Path Removal Plan

## Overview
Remove all legacy CommandSpec generation paths to complete architectural unification. Currently Claude can use two different paths which causes metadata inconsistencies:

- **Legacy Path**: `generate_command` tool → CommandSpec JSON → `d.generateCommand()` → creates tools with `session_id`, `agent`, `source: "possess"`
- **Unified Path**: `run_command("port42", ["declare"])` → CLI execution → creates tools with `memory_session`, `transforms`

## Root Cause
The architectural unification is incomplete - we updated prompts to guide Claude toward `port42 declare` but never removed the old CommandSpec detection and processing logic.

## Implementation Plan

### Phase 1: Remove Old Prompt Configuration *(safest - just prompts)*
**Goal**: Stop Claude from being told to use `generate_command` tool

**Changes to `daemon/agents.json`**:
- [ ] Remove `artifact_guidance` section that mentions `generate_command` tool
- [ ] Remove `format_template` with JSON CommandSpec examples  
- [ ] Remove `implementation` guidelines about JSON format
- [ ] Update `@ai-muse` prompt to remove "use generate_command for executable CLI tools"
- [ ] Update `@ai-engineer` prompt examples to only show `run_command` usage
- [ ] Ensure only `port42_integration` guidance remains (which says use `run_command`)

**Expected Result**: Claude will only be instructed to use `run_command("port42", ["declare"])` path

### **Phase 1 Testing Checkpoint** ✅
**Validate prompt changes work**:
- [ ] Test possess flow with tool creation request
- [ ] Check daemon logs to see which path Claude attempts to use
- [ ] Verify Claude tries `run_command("port42", ["declare"])` instead of `generate_command`
- [ ] If Claude still uses old path, the detection logic will still process it (expected)
- [ ] Confirm prompts are being applied correctly

---

### Phase 2: Remove CommandSpec Detection Logic *(stop the old path)*
**Goal**: Even if Claude tries to use old path, daemon won't process it

**Changes to `daemon/possession.go`**:
- [ ] Remove `generate_command` tool detection: `content.Name == "generate_command"`  
- [ ] Remove `extractCommandSpecFromToolCall()` function calls
- [ ] Remove `extractCommandSpec()` text parsing calls
- [ ] Remove `commandSpec` variable and all CommandSpec assignment logic
- [ ] Remove `d.generateCommand(commandSpec)` calls
- [ ] Remove CommandSpec data in response payload
- [ ] Keep only `run_command` and `generate_artifact` tool handling
- [ ] Keep CommandSpec functions/types for now (don't call them)

**Expected Result**: Only `run_command("port42", ["declare"])` path will work

### **Phase 2 Testing Checkpoint** ✅
**Critical validation before proceeding**:
- [ ] Test possess flow with tool creation request
- [ ] Verify Claude ONLY uses `run_command("port42", ["declare"])` path
- [ ] Verify tools get created with unified metadata (`memory_session`, `transforms`)
- [ ] Verify no `session_id`/`agent` metadata from old path
- [ ] Test error handling if Claude somehow tries old path
- [ ] Verify no broken functionality in possess sessions

---

### Phase 3: Update Session Management *(clean data structures)*
**Goal**: Remove CommandSpec references from session data structures

**Changes**:
- [ ] Remove `CommandGenerated *CommandSpec` field from Session struct in `daemon/server.go`
- [ ] Update session serialization in memory store (`daemon/storage.go`)
- [ ] Update session restoration logic to not load CommandSpec data
- [ ] Update session state management (remove CommandSpec-based state transitions)
- [ ] Update any session display/info endpoints

**Expected Result**: Sessions no longer track CommandSpec, only relation creation via memory linkage

---

### Phase 4: Test and Verify Unified Flow *(critical validation point)*
**Goal**: Comprehensive testing of unified architecture

**Test Cases**:
- [ ] **Basic Tool Creation**: Possess → Claude creates tool → verify metadata consistency
- [ ] **CLI Declare**: Direct `port42 declare tool` → verify same metadata structure
- [ ] **Session Continuity**: Sessions restore correctly without CommandSpec references  
- [ ] **Memory Linkage**: Tools created in possess show up in `/memory/{session}/generated`
- [ ] **VFS Consistency**: `/tools/toolname/definition` shows consistent metadata
- [ ] **Info Display**: `port42 info /tools/toolname` shows correct Type, Modified, Agent
- [ ] **Cross-comparison**: Compare possess-created vs CLI-created tool metadata side-by-side

**Success Criteria**:
- ✅ All tool creation uses unified `memory_session` metadata approach
- ✅ No more `session_id`/`agent` vs `memory_session` inconsistencies  
- ✅ Both flows create identical metadata structure
- ✅ Session management works without CommandSpec references
- ✅ VFS info display shows consistent data

---

### Phase 5: Remove Unused Code *(cleanup after verification)*
**Goal**: Remove dead code that's no longer called

**Remove from `daemon/possession.go`**:
- [ ] `CommandSpec` struct definition
- [ ] `ArtifactSpec` struct definition  
- [ ] `extractCommandSpec()` function
- [ ] `extractCommandSpecFromToolCall()` function

**Remove from `daemon/server.go`**:
- [ ] `generateCommand()` method
- [ ] `generateDependencyCheck()` method
- [ ] Any remaining CommandSpec references

**Remove from `daemon/protocol.go`**:
- [ ] `CommandSpec` type (if defined there)
- [ ] Any CommandSpec-related protocol types

**Remove from `daemon/storage.go`**:
- [ ] `StoreCommand()` method (if it exists and unused)
- [ ] Any CommandSpec file I/O operations

---

### Phase 6: Final Cleanup and Validation *(polish)*
**Goal**: Ensure complete removal and document changes

**Verification Steps**:
- [ ] `grep -r "CommandSpec" daemon/` should return no results
- [ ] `grep -r "generate_command" daemon/` should return no results  
- [ ] `grep -r "generateCommand" daemon/` should return no results
- [ ] All imports still valid, no unused imports
- [ ] All tests pass
- [ ] Documentation updated

**Final Test Suite**:
- [ ] Possess flow end-to-end test
- [ ] CLI declare flow end-to-end test  
- [ ] Metadata consistency verification
- [ ] Session management test
- [ ] Error handling test (malformed requests)

---

## Risk Mitigation

### Backup Strategy
- [ ] Create git branch for this work: `remove-commandspec-legacy`
- [ ] Commit after each phase for easy rollback
- [ ] Keep daemon binary backup before starting

### Rollback Plan
If issues arise:
1. **After Phase 1/2**: Revert agents.json and possession.go changes
2. **After Phase 3**: Restore Session struct, may need to clear session files  
3. **After Phase 4**: Full git revert to branch point
4. **After Phase 5/6**: Git revert specific commits

### Testing Strategy
- [ ] Test with real Claude API calls, not just mocks
- [ ] Test both @ai-muse and @ai-engineer agents
- [ ] Test various tool types (bash, python, etc.)
- [ ] Test error scenarios and edge cases

---

## Expected Outcomes

### Before (Current State)
```json
// Possess-created tool
{
  "id": "tool-x-test-1755904138", 
  "properties": {
    "agent": "@ai-engineer",
    "session_id": "cli-1755904081344", 
    "source": "possess"
  }
}

// CLI-created tool  
{
  "id": "tool-log-test-bf0fd6202cd497cf",
  "properties": {
    "memory_session": "cli-session-1755906155322",
    "transforms": ["logs", "analysis"]
  }
}
```

### After (Unified State)
```json
// Both flows create consistent metadata
{
  "id": "tool-example-abc123",
  "properties": {
    "memory_session": "cli-1755904081344", 
    "crystallized_agent": "@ai-engineer", // from original session
    "transforms": ["analysis", "patterns"],
    "original_session_context": "cli-1755904081344" // preserved context
  }
}
```

---

## Future Considerations

### Session Context Preservation
After unification, we should consider preserving original session context when Claude calls `port42 declare`, so tools maintain connection to their originating conversation.

### Metadata Schema Standardization  
Define canonical tool metadata schema and ensure both CLI and possess flows populate it identically.

### Command Creation Provenance
Consider storing the full `port42 declare` command that created each tool for debugging and recreation purposes.

---

*This plan ensures we methodically remove legacy paths while maintaining system stability through careful testing at each phase.*