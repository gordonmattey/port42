# Consolidated Implementation Plan

**Purpose**: Step-by-step implementation plan combining AI tool integration and data commands
**Scope**: Ordered steps with clear priorities

## Phase 1: AI Tool Integration (COMPLETED ✅)
*Estimated time: 2-3 hours*
*Actual time: ~1 hour*
*Impact: Makes all existing commands available to AI*

### Step 1: Add Command Runner Tool ✅
**File**: `daemon/possession.go`
- ✅ Added `getCommandRunnerTool()` function after `getArtifactGenerationTool()`
- ✅ Updated tool provisioning in `Send()` method
- ✅ Added to both NoImplementation and full implementation agents
**Commit**: 3753dfd

### Step 2: Update System Prompt ✅
**File**: `daemon/agents.go`
- ✅ Added `CommandMetadata` struct
- ✅ Added `listAvailableCommands()` helper function
- ✅ Modified `GetAgentPrompt()` to insert available commands after base prompt
**Commit**: 71af38c

### Step 3: Handle Command Execution ✅
**File**: `daemon/possession.go`
- ✅ Added required imports (context, os/exec, path/filepath)
- ✅ Added `executeCommand()` function with security checks
- ✅ Added handler for "run_command" tool use in `handlePossessWithAI()`
**Commit**: fd010fc

### Step 4: Test AI Tool Integration ✅
- ✅ Generated test command: `echo-test`
- ✅ Tested AI can list available commands
- ✅ Tested AI can execute the command
- ✅ Cross-agent usage confirmed working

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

### Phase 1 Success: ✅ ACHIEVED
- [x] AI can list all available commands in system prompt
- [x] AI can execute any previously generated command
- [x] Commands work across all agents
- [x] No regression in existing functionality

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