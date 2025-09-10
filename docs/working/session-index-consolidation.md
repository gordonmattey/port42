# Session Index Consolidation Plan - Option 2

## Overview
Consolidate all session tracking into a single enhanced `session-index.json` file, eliminating redundancy and improving architecture.

## Current Problems
1. **Three separate session tracking systems**:
   - `session-index.json` (106 sessions with full metadata)
   - `last_session` file (deprecated, global last session)
   - `agent_sessions.json` (new, agent-specific last sessions)

2. **Redundancy**: Session index already contains agent information
3. **Multiple mutexes**: Potential race conditions
4. **Storage overhead**: Three files where one could suffice

## Solution: Enhanced session-index.json Structure

### New Unified Structure
```json
{
  "sessions": {
    "cli-1756381923296": {
      "object_id": "b767d82f1dcf478e6708d42b14efe8aa954a8c16e2955e0fffa9a76537fdca0b",
      "session_id": "cli-1756381923296",
      "agent": "ai-engineer",
      "created_at": "2025-08-28T04:52:03.297238-07:00",
      "last_updated": "2025-08-28T06:07:32.547971-07:00",
      "command_generated": false,
      "state": "abandoned",
      "message_count": 1
    },
    // ... more sessions
  },
  "last_sessions": {
    "ai-engineer": "cli-1757473233669",
    "ai-analyst": "cli-1757471420832",
    "ai-muse": "cli-1757471474253"
  },
  "metadata": {
    "version": "2.0",
    "last_updated": "2025-01-09T20:30:00Z",
    "total_sessions": 106
  }
}
```

## Implementation Phases

### Phase 1: Remove All Deprecated Code

#### 1.1 Remove from Storage struct (`daemon/src/storage.go`)
- Remove `lastSessionID string` field (line ~113)
- Remove all references to `s.lastSessionID`

#### 1.2 Remove deprecated file operations
From `NewStorage()`:
- Remove loading of `last_session` file (lines ~179-184)
- Remove `lastSessionFile := filepath.Join(s.baseDir, "last_session")`
- Remove reading and logging of deprecated last session

From file system: User will manyally remove this `~/.port42/last_session` file

#### 1.3 Remove AgentSessions struct entirely
Remove completely:
- `type AgentSessions struct` (lines 18-100)
- `NewAgentSessions` function
- `Load()` method
- `Save()` method  
- `GetLastSession()` method
- `SetLastSession()` method
- `agentSessions` field from Storage struct (line ~116)
- `agentSessions` initialization in NewStorage (lines ~156-161)

### Phase 2: Restructure session-index.json

#### 2.1 Create new types (`daemon/src/types.go`)
```go
// SessionIndex represents the complete session storage
type SessionIndex struct {
    Sessions     map[string]SessionReference `json:"sessions"`
    LastSessions map[string]string          `json:"last_sessions"`
    Metadata     SessionIndexMetadata       `json:"metadata"`
}

// SessionIndexMetadata contains index-level metadata
type SessionIndexMetadata struct {
    Version       string    `json:"version"`
    LastUpdated   time.Time `json:"last_updated"`
    TotalSessions int       `json:"total_sessions"`
}
```

#### 2.2 Update Storage struct (`daemon/src/storage.go`)
```go
type Storage struct {
    baseDir     string
    objectsDir  string
    metadataDir string
    
    // Single unified session index
    sessionIndex *SessionIndex  // Changed from map to struct
    indexMutex   sync.RWMutex
    
    // Relations integration
    relationStore RelationStore
    
    // Stats
    stats StorageStats
}
```

### Phase 2.5: Create Migration Script

#### Create standalone migration script (`daemon/migrate-session-index.go`)
```go
// +build ignore

package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"
)

func main() {
    homeDir, _ := os.UserHomeDir()
    port42Dir := filepath.Join(homeDir, ".port42")
    
    // Paths
    sessionIndexPath := filepath.Join(port42Dir, "session-index.json")
    agentSessionsPath := filepath.Join(port42Dir, "agent_sessions.json")
    backupPath := filepath.Join(port42Dir, "session-index.json.backup")
    
    // Check if already migrated
    data, err := ioutil.ReadFile(sessionIndexPath)
    if err != nil {
        log.Fatal("Failed to read session-index.json:", err)
    }
    
    // Try parsing as new format first
    var testIndex map[string]interface{}
    json.Unmarshal(data, &testIndex)
    if _, hasMetadata := testIndex["metadata"]; hasMetadata {
        log.Println("✅ session-index.json is already in v2.0 format")
        return
    }
    
    // Parse old format
    var oldIndex map[string]SessionReference
    if err := json.Unmarshal(data, &oldIndex); err != nil {
        log.Fatal("Failed to parse old session index:", err)
    }
    
    // Load agent sessions if exists
    agentLastSessions := make(map[string]string)
    if agentData, err := ioutil.ReadFile(agentSessionsPath); err == nil {
        json.Unmarshal(agentData, &agentLastSessions)
        log.Printf("📦 Loaded agent sessions from agent_sessions.json")
    } else {
        // Build from session index if no agent_sessions.json
        agentLatest := make(map[string]time.Time)
        for sessionID, ref := range oldIndex {
            agent := strings.TrimPrefix(ref.Agent, "@")
            if ref.State != "abandoned" && ref.Agent != "" {
                if ref.LastUpdated.After(agentLatest[agent]) {
                    agentLatest[agent] = ref.LastUpdated
                    agentLastSessions[agent] = sessionID
                }
            }
        }
        log.Printf("📦 Built last sessions from session history")
    }
    
    // Create new format
    newIndex := SessionIndex{
        Sessions:     oldIndex,
        LastSessions: agentLastSessions,
        Metadata: SessionIndexMetadata{
            Version:       "2.0",
            LastUpdated:   time.Now(),
            TotalSessions: len(oldIndex),
        },
    }
    
    // Backup old file
    if err := os.Rename(sessionIndexPath, backupPath); err != nil {
        log.Fatal("Failed to backup old session index:", err)
    }
    log.Printf("💾 Backed up old index to %s", backupPath)
    
    // Write new format
    newData, _ := json.MarshalIndent(newIndex, "", "  ")
    if err := ioutil.WriteFile(sessionIndexPath, newData, 0644); err != nil {
        // Restore backup on failure
        os.Rename(backupPath, sessionIndexPath)
        log.Fatal("Failed to write new session index:", err)
    }
    
    log.Printf("✅ Successfully migrated session-index.json to v2.0 format")
    log.Printf("   - %d sessions migrated", len(oldIndex))
    log.Printf("   - %d agent last sessions tracked", len(agentLastSessions))
    
    // Clean up obsolete files since they're now integrated
    obsoleteFiles := []string{
        agentSessionsPath,
        filepath.Join(port42Dir, "last_session"),
    }
    
    for _, file := range obsoleteFiles {
        if _, err := os.Stat(file); err == nil {
            if err := os.Remove(file); err != nil {
                log.Printf("⚠️  Failed to remove %s: %v", file, err)
            } else {
                log.Printf("🧹 Removed obsolete file: %s", filepath.Base(file))
            }
        }
    }
}
```

Run with: `go run daemon/migrate-session-index.go`

### Phase 3: Reimplement Core Functions

#### 3.1 Update loadSessionIndex (`daemon/src/storage.go`)
```go
func (s *Storage) loadSessionIndex() error {
    indexPath := filepath.Join(s.baseDir, "session-index.json")
    
    data, err := os.ReadFile(indexPath)
    if err != nil {
        if os.IsNotExist(err) {
            // Initialize empty index with new structure
            s.sessionIndex = &SessionIndex{
                Sessions:     make(map[string]SessionReference),
                LastSessions: make(map[string]string),
                Metadata: SessionIndexMetadata{
                    Version:     "2.0",
                    LastUpdated: time.Now(),
                },
            }
            return nil
        }
        return err
    }
    
    // Parse as v2.0 format
    var index SessionIndex
    if err := json.Unmarshal(data, &index); err != nil {
        // Check if it's old format
        var oldTest map[string]SessionReference
        if err2 := json.Unmarshal(data, &oldTest); err2 == nil {
            return fmt.Errorf("session-index.json is in old format - run: go run daemon/migrate-session-index.go")
        }
        return fmt.Errorf("failed to parse session index: %w", err)
    }
    
    // Verify it's v2.0
    if index.Metadata.Version != "2.0" {
        return fmt.Errorf("unsupported session index version: %s", index.Metadata.Version)
    }
    
    s.sessionIndex = &index
    return nil
}
```

#### 3.3 Update saveSessionIndex
```go
func (s *Storage) saveSessionIndex() error {
    indexPath := filepath.Join(s.baseDir, "session-index.json")
    
    // Update metadata
    s.sessionIndex.Metadata.LastUpdated = time.Now()
    s.sessionIndex.Metadata.TotalSessions = len(s.sessionIndex.Sessions)
    
    data, err := json.MarshalIndent(s.sessionIndex, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal session index: %w", err)
    }
    
    return os.WriteFile(indexPath, data, 0644)
}
```

#### 3.4 Reimplement GetLastSession
```go
func (s *Storage) GetLastSession(agent string) (string, error) {
    if agent == "" {
        return "", fmt.Errorf("agent parameter required")
    }
    
    // Normalize agent name
    agent = strings.TrimPrefix(agent, "@")
    
    s.indexMutex.RLock()
    defer s.indexMutex.RUnlock()
    
    // O(1) lookup from last_sessions
    sessionID, exists := s.sessionIndex.LastSessions[agent]
    if !exists {
        return "", fmt.Errorf("no sessions found for agent %s", agent)
    }
    
    // Verify session still exists
    if _, exists := s.sessionIndex.Sessions[sessionID]; !exists {
        // Clean up stale reference
        delete(s.sessionIndex.LastSessions, agent)
        return "", fmt.Errorf("session %s no longer exists", sessionID)
    }
    
    log.Printf("🔍 [STORAGE] Retrieved last session for %s: %s", agent, sessionID)
    return sessionID, nil
}
```

#### 3.5 Reimplement UpdateLastSession
```go
func (s *Storage) UpdateLastSession(agent, sessionID string) error {
    if agent == "" {
        return fmt.Errorf("agent parameter required")
    }
    
    // Normalize agent name
    agent = strings.TrimPrefix(agent, "@")
    
    // Note: Caller should already hold the lock
    // Update in-memory
    s.sessionIndex.LastSessions[agent] = sessionID
    
    // Save will happen when SaveSession completes
    log.Printf("📌 [STORAGE] Updated last session for %s -> %s", agent, sessionID)
    return nil
}
```

#### 3.6 Update SaveSession
```go
func (s *Storage) SaveSession(session *Session) error {
    s.indexMutex.Lock()
    defer s.indexMutex.Unlock()
    
    // ... existing save logic ...
    
    // Update session in index
    s.sessionIndex.Sessions[session.ID] = SessionReference{
        ObjectID:         objectID,
        SessionID:        session.ID,
        Agent:            session.Agent,
        CreatedAt:        session.CreatedAt,
        LastUpdated:      time.Now(),
        CommandGenerated: session.CommandGenerated != nil,
        State:            session.State,
        MessageCount:     len(session.Messages),
    }
    
    // Update last session for this agent
    if session.Agent != "" {
        agent := strings.TrimPrefix(session.Agent, "@")
        s.sessionIndex.LastSessions[agent] = session.ID
    }
    
    // Save the updated index
    if err := s.saveSessionIndex(); err != nil {
        log.Printf("Warning: Failed to save session index: %v", err)
    }
    
    return nil
}
```

### Phase 4: No Daemon Cleanup Needed
The migration script handles all file cleanup - removing `agent_sessions.json` and `last_session` files after successful migration. The daemon doesn't need any cleanup code.

### Phase 5: No CLI Changes Needed
The CLI already uses `get_last_session` with agent parameter, so it will work seamlessly with the new implementation.

## Implementation Order

1. **Step 1**: Create and run migration script
   - Run `go run daemon/migrate-session-index.go`
   - Verify session-index.json is in v2.0 format
   - Verify backup was created
   
2. **Step 2**: Add new types (SessionIndex, SessionIndexMetadata) to daemon
3. **Step 3**: Update Storage struct to use new SessionIndex type
4. **Step 4**: Update load/save functions to work with v2.0 format only
5. **Step 5**: Update GetLastSession/UpdateLastSession to use new structure
6. **Step 6**: Update SaveSession to maintain last_sessions
7. **Step 7**: Remove AgentSessions struct completely
8. **Step 8**: Remove deprecated lastSessionID code  
9. **Step 9**: Test with migrated data
10. **Step 10**: Test fresh installations

## Benefits

1. **Single Source of Truth**: One file, one mutex, one index
2. **O(1) Performance**: Direct lookup for last sessions
3. **Atomic Operations**: All session data updated together
4. **Clean Migration**: Automatic upgrade from old format
5. **Extensibility**: Metadata section for future enhancements
6. **No User Action Required**: Automatic migration on first load
7. **Reduced Complexity**: Single mutex, no synchronization issues

## Testing Strategy

### 1. Migration Test
- Start with old format session-index.json
- Verify automatic migration to new format
- Verify last_sessions populated correctly from most recent sessions
- Verify metadata section created

### 2. Fresh Install Test
- Remove all session files
- Verify new format created from scratch
- Verify empty last_sessions map

### 3. Session Operations
- Create sessions for multiple agents
- Verify last_sessions updates correctly
- Test `--session last` for each agent
- Test resuming specific session IDs
- Verify session cleanup updates last_sessions

### 4. Edge Cases
- Abandoned sessions excluded from last_sessions
- Session deletion removes from last_sessions if it was last
- Agent with no sessions returns appropriate error
- Cross-agent session resumption still works by ID
- Non-existent session handling

### 5. File Cleanup
- Verify obsolete files removed on startup
- Verify no recreation of old files

## Success Criteria

- [ ] All session tracking in single `session-index.json` file
- [ ] O(1) lookup performance for last sessions
- [ ] Automatic migration from old format
- [ ] No deprecated code remains
- [ ] No obsolete files remain
- [ ] All existing functionality preserved
- [ ] Clean error messages for edge cases
- [ ] Session persistence across daemon restarts