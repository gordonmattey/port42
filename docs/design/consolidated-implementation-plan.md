# Consolidated Implementation Plan

**Purpose**: Step-by-step implementation plan combining AI tool integration and data commands
**Scope**: Ordered steps with clear priorities

## Phase 1: AI Tool Integration (HIGH PRIORITY)
*Estimated time: 2-3 hours*
*Impact: Makes all existing commands available to AI*

### Step 1: Add Command Runner Tool
**File**: `daemon/possession.go`
- Add `getCommandRunnerTool()` function after `getArtifactGenerationTool()`
- Update tool provisioning in `Send()` method (around line 173-186)
- Add to both NoImplementation and full implementation agents

### Step 2: Update System Prompt  
**File**: `daemon/agents.go`
- Add `CommandMetadata` struct
- Add `listAvailableCommands()` helper function
- Modify `GetAgentPrompt()` to insert available commands after base prompt

### Step 3: Handle Command Execution
**File**: `daemon/possession.go`
- Add required imports (context, os/exec, path/filepath, io/ioutil)
- Add `executeCommand()` function with security checks
- Add handler for "run_command" tool use in `handlePossessWithAI()`

### Step 4: Test AI Tool Integration
- Generate a test command: `echo-test`
- Test AI can list available commands
- Test AI can execute the command
- Test cross-agent usage (create with @ai-muse, use with @ai-engineer)

## Phase 2: Data Commands (LOWER PRIORITY)
*Estimated time: 4-6 hours*
*Impact: Adds third crystallization type for structured data*

### Step 5: Update CLI for Data Type
**File**: `cli/src/interactive.rs`
- Add `Data` variant to `CrystallizeType` enum
- Update `handle_special_command()` to handle `/crystallize data`
- Update `request_crystallization()` with data-specific prompt

### Step 6: Add Data Generation Tool
**File**: `daemon/possession.go`
- Add `getDataGenerationTool()` function
- Add `DataCommandSpec` struct
- Update tool provisioning logic to include data tool

### Step 7: Implement Data Command Generation
**File**: `daemon/data_generator.go` (new file)
- Create bash template for CRUD operations
- Implement JSON storage with jq
- Add basic operations: add, list, show, update, delete, search
- Generate executable script in ~/.port42/commands/

### Step 8: Test Data Commands
- Test `/crystallize data` in interactive mode
- Create a content-calendar command
- Test all CRUD operations
- Verify JSON storage works correctly

## Success Criteria

### Phase 1 Success:
- [ ] AI can list all available commands in system prompt
- [ ] AI can execute any previously generated command
- [ ] Commands work across all agents
- [ ] No regression in existing functionality

### Phase 2 Success:
- [ ] `/crystallize data` creates working CRUD commands
- [ ] Generated commands handle JSON data correctly
- [ ] Schema is discovered through conversation
- [ ] Basic operations work (add, list, update, delete)

## Implementation Order Rationale

1. **AI Tool Integration First** because:
   - Simpler implementation (3-4 hours)
   - Immediate value for all existing commands
   - No CLI changes required
   - Foundation for testing data commands later

2. **Data Commands Second** because:
   - More complex (new generation logic)
   - Requires CLI changes
   - Can leverage AI tool integration for testing
   - Separate feature that can be deferred

## Risk Mitigation

1. **Test after each step** - Don't wait until the end
2. **Commit after Phase 1** - Get value deployed quickly
3. **Keep changes minimal** - Preserve existing functionality
4. **Security first** - Path validation, timeouts, no shell expansion