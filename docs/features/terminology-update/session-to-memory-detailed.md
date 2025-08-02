# Detailed Session to Memory Change List

## Rust Files (CLI)

### 1. src/types.rs
- Line 29: `pub struct Session` → `pub struct Memory`
- All references to this struct throughout the codebase

### 2. src/interactive.rs
- Line 9: `pub struct InteractiveSession` → `pub struct InteractiveMemory`
- Line 12: `session_id: String` → `memory_id: String`
- Line 18: `impl InteractiveSession` → `impl InteractiveMemory`
- Line 19: `session_id: String` parameter → `memory_id: String`
- Line 23: `session_id,` → `memory_id,`
- Line 115: `self.show_session_memory()?` → `self.show_memory_contents()?` (rename for clarity)
- Line 137-138: Debug messages with `session_id`
- Line 143: `id: self.session_id.clone()` → `id: self.memory_id.clone()`
- Line 244: `fn show_session_memory(&self)` → `fn show_memory_contents(&self)`
- Line 245: `"Session Memory:"` → `"Memory Contents:"`
- Line 264: `"Commands born from this session:"` → `"Commands born from this memory:"`
- Line 307: `"Session Summary:"` → `"Memory Summary:"`

### 3. src/commands/memory.rs
- Line 13: `list_sessions(&mut client)?` → `list_memories(&mut client)?`
- Line 19: Comment about sessions
- Line 22-23: `session_id` parameter → `memory_id`
- Line 30: `fn list_sessions` → `fn list_memories`
- Line 45-46: `active_sessions` → `active_memories`
- Line 48: `"Active Sessions:"` → `"Active Memories:"`
- Line 49-50: `session` variable → `memory`
- Line 56-58: `recent_sessions` → `recent_memories`
- Line 63-65: `session` variable → `memory`
- Line 72-73: `sessions_on_date` → `memories_on_date`
- Line 79: `fn show_session` → `fn show_memory`
- Line 95-97: Session-related response handling
- Line 102: `fn print_session_summary` → `fn print_memory_summary`
- Multiple UI strings throughout

### 4. src/commands/possess.rs
- Session ID references in possess handling
- Need to check actual usage patterns

### 5. src/commands/status.rs
- Line 58: `active_sessions` → `active_memories`
- Line 65: `"Sessions:"` → `"Memories:"`
- Line 73-74: `total_sessions` → `total_memories`
- Line 84: Comment about sessions

### 6. src/commands/init.rs
- Line 31: `memory/sessions` → `memory/memories` or just `memory`
- Line 40: Documentation comment

### 7. src/shell.rs
- Line 123: `session` variable usage
- Line 163: Comment about session
- Line 165: `session` parameter
- Line 184-186: Session ID handling
- Line 251-254: Help text about sessions
- Line 261-263: Memory command help referring to sessions

### 8. src/boot.rs
- Need to check for session references

### 9. src/main.rs
- Need to check for session references

## Go Files (Daemon)

### 1. daemon/types.go
- Lines 7-8: `SessionState` type and comment
- Lines 11-14: State constants (SessionActive, etc.)
- Line 26: `Session` field in struct
- Line 46: `Session` field in Relationships
- Lines 54-55: `SessionReference` struct
- Line 57: `SessionID` field
- Lines 66-67: `PersistentSession` struct
- Line 70: `SessionState` field
- Line 77: `SessionID` field in CommandSpec
- Line 88: `TotalSessions` field
- Lines 90-91: `ActiveSessions`, `LastSessionTime`
- Lines 94-95: `SessionSummary` struct

### 2. daemon/storage.go
- Lines 23-24: `sessionIndex` field and comment
- Lines 33-35: Session-related stats fields
- Line 58: `sessionIndex` initialization
- Lines 62-64: Loading session index
- Line 213: Section comment
- Lines 215-218: `SaveSession` method
- Lines 223-226: Session existence check
- Lines 229-237: Creating persistent session
- Lines 244-249: Command info with session
- Lines 252-256: Session serialization
- Lines 260-272: Metadata creation with session paths
- Lines 279-280: Session storage error
- Lines 283-291: Session index update
- Lines 298-300: Save session index
- Line 302: Success log
- Lines 306-307: `LoadSession` method
- Lines 309: Session index lookup
- Lines 313: Session not found error
- Lines 319-320: Session object read error
- Line 322: Comment about persistent session
- Many more session-related methods throughout the file

### 3. daemon/server.go
- Session map field
- `getOrCreateSession` method
- Session management throughout

### 4. daemon/protocol.go
- Line 31: Comment about possession
- Line 41: `Sessions` field

### 5. daemon/possession.go
- Line 77: `SessionID` field in CommandSpec
- Line 274: Get or create session
- Line 275: Session loaded log
- Multiple session references throughout AI handling

## Migration Considerations

1. **Backward Compatibility**
   - Need to handle old "session" paths
   - Protocol version negotiation
   - Deprecation warnings

2. **Data Migration**
   - Move `.port42/memory/sessions/` → `.port42/memory/memories/`
   - Update metadata files
   - Preserve existing data

3. **User Communication**
   - Clear upgrade instructions
   - Migration tool or automatic migration
   - Documentation updates