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

## Implementation Steps

### Step 1: Verify Deprecated Code Already Removed ✅
- Remove `lastSessionID string` field if it exists - **Already done**
- Remove all references to `s.lastSessionID` - **Already done**
- Remove loading of `last_session` file - **Already done**
- **KEEP AgentSessions** - it's the current working implementation

### Step 2: Add New Types to daemon/src/types.go
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

### Step 3: Create Migration Script (but don't run yet)
Create `daemon/migrate-session-index.go`:
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

Run with: `go run daemon/migrate-session-index.go` (but not yet!)

### Step 4: Update Storage struct
Change the sessionIndex field type in `daemon/src/storage.go`:
```go
type Storage struct {
    // ... existing fields ...
    sessionIndex  *SessionIndex    // CHANGED from map[string]SessionReference
    agentSessions *AgentSessions   // Keep for now (will remove later)
    
    indexMutex   sync.RWMutex
    // ... rest ...
}
```

### Step 5: Update loadSessionIndex function
Replace the existing loadSessionIndex in `daemon/src/storage.go`:
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
    
    // Try to parse as new format
    var index SessionIndex
    if err := json.Unmarshal(data, &index); err != nil {
        return fmt.Errorf("failed to parse session index - run migration: go run daemon/migrate-session-index.go")
    }
    
    // Verify it's v2.0
    if index.Metadata.Version != "2.0" {
        return fmt.Errorf("session-index.json needs migration - run: go run daemon/migrate-session-index.go")
    }
    
    s.sessionIndex = &index
    return nil
}
```

### Step 6: Update saveSessionIndex function
Replace the existing saveSessionIndex:
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

### Step 7: Update SaveSession
Modify SaveSession to use the new structure:
```go
func (s *Storage) SaveSession(session *Session) error {
    s.indexMutex.Lock()
    defer s.indexMutex.Unlock()
    
    // ... existing save logic ...
    
    // Update session in consolidated index
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
    
    // Update last session for this agent in the consolidated index
    if session.Agent != "" {
        agent := strings.TrimPrefix(session.Agent, "@")
        s.sessionIndex.LastSessions[agent] = session.ID
    }
    
    // Save the consolidated index
    if err := s.saveSessionIndex(); err != nil {
        log.Printf("Warning: Failed to save session index: %v", err)
    }
    
    // TEMPORARILY keep updating AgentSessions until we switch over
    if session.Agent != "" {
        s.UpdateLastSession(session.Agent, session.ID)
    }
    
    return nil
}
```

### Step 8: Run Migration Script
Now run the migration to create v2.0 format:
```bash
go run daemon/migrate-session-index.go
```
This will:
- Backup existing session-index.json
- Create new v2.0 format with sessions, last_sessions, and metadata
- Clean up agent_sessions.json and last_session files

### Step 9: Update GetLastSession
Replace GetLastSession to use the consolidated index:
```go
func (s *Storage) GetLastSession(agent string) (string, error) {
    if agent == "" {
        return "", fmt.Errorf("agent parameter required")
    }
    
    agent = strings.TrimPrefix(agent, "@")
    
    s.indexMutex.RLock()
    defer s.indexMutex.RUnlock()
    
    // Use consolidated index
    sessionID, exists := s.sessionIndex.LastSessions[agent]
    if !exists {
        return "", fmt.Errorf("no sessions found for agent %s", agent)
    }
    
    // Verify session still exists
    if _, exists := s.sessionIndex.Sessions[sessionID]; !exists {
        delete(s.sessionIndex.LastSessions, agent)
        return "", fmt.Errorf("session %s no longer exists", sessionID)
    }
    
    log.Printf("🔍 [STORAGE] Retrieved last session for %s: %s", agent, sessionID)
    return sessionID, nil
}
```

### Step 10: Test Everything Works
- Test `--session last` for each agent
- Create new sessions and verify they're tracked
- Restart daemon and verify persistence

### Step 11: Remove AgentSessions
Once confirmed working, remove:
- `type AgentSessions struct` and all its methods
- `agentSessions` field from Storage
- Remove AgentSessions initialization in NewStorage
- Remove UpdateLastSession calls from SaveSession

### Step 12: Clean up any remaining references
- Remove any remaining references to AgentSessions
- Remove the temporary UpdateLastSession call in SaveSession
- Verify all session tracking goes through the consolidated index

## Summary

Simplified plan with 12 steps:
- Steps 1-3: Preparation (types, migration script)
- Steps 4-7: Update daemon to use new consolidated structure
- Step 8: Run migration to convert data
- Step 9: Update GetLastSession to use consolidated index
- Step 10: Test everything
- Steps 11-12: Remove AgentSessions and cleanup

**Key simplifications:**
- No "V2" labeling - just update the types directly
- No dual system - migrate and switch over
- Clean, direct replacement without backward compatibility code

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