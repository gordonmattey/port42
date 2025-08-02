# Session to Memory Terminology Change Analysis

## Executive Summary

The codebase has approximately 400+ occurrences of "session" terminology across 29 files that need to be changed to "memory" terminology. This is a significant refactoring that touches both the CLI (Rust) and daemon (Go) components, as well as documentation.

## Component Breakdown

### 1. CLI Component (Rust) - 9 Files

#### Core Types & Structures
- **src/types.rs**
  - `Session` struct → `Memory` struct
  - Related fields and methods

#### Interactive Module
- **src/interactive.rs** 
  - `InteractiveSession` struct → `InteractiveMemory`
  - `session_id` field → `memory_id`
  - `show_session_memory()` method (naming confusion - needs clarity)
  - Session summary/commands born from session messages

#### Commands
- **src/commands/memory.rs**
  - `list_sessions()` → `list_memories()`
  - `show_session()` → `show_memory()`
  - `print_session_summary()` → `print_memory_summary()`
  - Various "session" labels in UI output

- **src/commands/possess.rs**
  - References to session IDs and continuation

- **src/commands/status.rs**
  - `active_sessions` → `active_memories`
  - `total_sessions` → `total_memories`

- **src/commands/init.rs**
  - Directory structure: `memory/sessions/` → `memory/memories/` or just `memory/`

#### Shell & Boot
- **src/shell.rs**
  - Help text referring to "session" for possess command
  - Session ID references in memory command handling

- **src/boot.rs**
  - Likely session references (needs checking)

- **src/main.rs**
  - Entry point references

### 2. Daemon Component (Go) - 4 Core Files

#### Types & Data Structures
- **daemon/types.go**
  - `SessionState` → `MemoryState`
  - `SessionActive/Idle/Completed/Abandoned` → `MemoryActive/Idle/Completed/Abandoned`
  - `SessionReference` → `MemoryReference`
  - `PersistentSession` → `PersistentMemory`
  - `SessionSummary` → `MemorySummary`
  - Various session-related fields in structs

#### Storage Layer
- **daemon/storage.go**
  - `sessionIndex` → `memoryIndex`
  - `SaveSession()` → `SaveMemory()`
  - `LoadSession()` → `LoadMemory()`
  - `loadSessionIndex()` → `loadMemoryIndex()`
  - `ListSessions()` → `ListMemories()`
  - `GetSessionStats()` → `GetMemoryStats()`
  - File paths: `memory/sessions/` → `memory/memories/`

#### Server & Protocol
- **daemon/server.go**
  - `sessions` map → `memories` map
  - `getOrCreateSession()` → `getOrCreateMemory()`
  - Session management logic

- **daemon/protocol.go**
  - Protocol messages with "sessions" field

#### AI Integration
- **daemon/possession.go**
  - `SessionID` field in CommandSpec
  - Session references in AI handling

### 3. Documentation - 20 Files

Multiple markdown files contain references to sessions that need updating for consistency.

### 4. Tests

Test files will need updates but were excluded from initial analysis.

## Special Considerations

### Cases Where "Session" Should Remain

1. **TCP/Network Sessions** - Low-level network connections
2. **Shell Sessions** - Terminal/shell session contexts
3. **HTTP Sessions** - Web session management (if applicable)
4. **External API Sessions** - Third-party integrations

### Breaking Changes

1. **File System Structure**
   - Current: `.port42/memory/sessions/`
   - New: `.port42/memory/memories/` or `.port42/memory/`
   - Migration needed for existing installations

2. **API/Protocol Changes**
   - JSON field names will change
   - Client-server protocol compatibility

3. **Command Syntax**
   - `possess @agent [session-id]` → `possess @agent [memory-id]`
   - Help text and documentation

### Related Terms to Update

- `SessionID` → `MemoryID`
- `session_id` → `memory_id`
- `sessions` → `memories`
- `Session` → `Memory` (context-dependent)
- `session-related` → `memory-related`

## Implementation Strategy

### Phase 1: Type System & Core
1. Update Go types in daemon/types.go
2. Update Rust types in src/types.rs
3. Update storage interfaces

### Phase 2: Storage & Persistence
1. Update storage.go methods
2. Update file system paths
3. Create migration logic for existing data

### Phase 3: API & Protocol
1. Update protocol definitions
2. Update server handlers
3. Ensure backward compatibility

### Phase 4: CLI & UI
1. Update command handlers
2. Update interactive mode
3. Update help text and prompts

### Phase 5: Documentation
1. Update all markdown files
2. Update code comments
3. Update examples

## Risk Assessment

- **High Risk**: Breaking existing installations without migration
- **Medium Risk**: Client-server protocol incompatibility
- **Low Risk**: Documentation inconsistency

## Recommendations

1. Create a migration tool for existing `.port42` directories
2. Consider supporting both terms temporarily with deprecation warnings
3. Update tests comprehensively
4. Version the protocol to handle compatibility
5. Clear communication in release notes about the breaking change