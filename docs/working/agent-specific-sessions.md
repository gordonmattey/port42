# Agent-Specific Session Tracking - Implementation Plan

## Overview
Fix the critical bug where `--session last` doesn't properly isolate sessions per agent. Currently, all agents share the same "last session" marker, causing @ai-engineer to resume @ai-analyst's session.

## Problem Statement
- `last_session` stores only a session ID with no agent reference
- Multiple agents can have parallel sessions
- `--session last` picks up wrong agent's session
- No isolation between agent sessions

## Solution: JSON-Based Agent Session Mapping
Use a dedicated JSON file (`agent_sessions.json`) to maintain agent-to-session mappings with proper locking for thread safety.

## Implementation Design

### 1. Storage Layer Changes (`daemon/src/storage.go`)

#### 1.1 New AgentSessions Structure
```go
// AgentSessions manages last session tracking per agent
type AgentSessions struct {
    mu          sync.RWMutex
    sessions    map[string]string // agent -> sessionID
    filePath    string
}

// NewAgentSessions creates a new agent session tracker
func NewAgentSessions(baseDir string) *AgentSessions {
    return &AgentSessions{
        sessions: make(map[string]string),
        filePath: filepath.Join(baseDir, "agent_sessions.json"),
    }
}

// Load reads agent sessions from disk
func (as *AgentSessions) Load() error {
    as.mu.Lock()
    defer as.mu.Unlock()
    
    data, err := os.ReadFile(as.filePath)
    if err != nil {
        if os.IsNotExist(err) {
            // File doesn't exist yet, not an error
            return nil
        }
        return fmt.Errorf("failed to read agent sessions: %w", err)
    }
    
    if err := json.Unmarshal(data, &as.sessions); err != nil {
        return fmt.Errorf("failed to parse agent sessions: %w", err)
    }
    
    return nil
}

// Save writes agent sessions to disk
func (as *AgentSessions) Save() error {
    as.mu.RLock()
    defer as.mu.RUnlock()
    
    data, err := json.MarshalIndent(as.sessions, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal agent sessions: %w", err)
    }
    
    if err := os.WriteFile(as.filePath, data, 0644); err != nil {
        return fmt.Errorf("failed to write agent sessions: %w", err)
    }
    
    return nil
}

// GetLastSession returns the last session for an agent
func (as *AgentSessions) GetLastSession(agent string) (string, bool) {
    as.mu.RLock()
    defer as.mu.RUnlock()
    
    sessionID, exists := as.sessions[agent]
    return sessionID, exists
}

// SetLastSession updates the last session for an agent
func (as *AgentSessions) SetLastSession(agent, sessionID string) error {
    as.mu.Lock()
    defer as.mu.Unlock()
    
    as.sessions[agent] = sessionID
    
    // Save immediately for persistence
    data, err := json.MarshalIndent(as.sessions, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal agent sessions: %w", err)
    }
    
    if err := os.WriteFile(as.filePath, data, 0644); err != nil {
        return fmt.Errorf("failed to write agent sessions: %w", err)
    }
    
    log.Printf("📌 [AGENT_SESSIONS] Updated %s -> %s", agent, sessionID)
    return nil
}
```

#### 1.2 Update Storage Structure
```go
type Storage struct {
    // ... existing fields ...
    agentSessions *AgentSessions  // NEW: Per-agent session tracking
}

// In NewStorage function
func NewStorage(baseDir string) (*Storage, error) {
    // ... existing code ...
    
    // Initialize agent sessions
    agentSessions := NewAgentSessions(baseDir)
    if err := agentSessions.Load(); err != nil {
        log.Printf("⚠️ [STORAGE] Failed to load agent sessions: %v", err)
    }
    
    return &Storage{
        // ... existing fields ...
        agentSessions: agentSessions,
    }, nil
}
```

#### 1.3 Update Session Methods
```go
// GetLastSession now requires agent parameter
func (s *Storage) GetLastSession(agent string) (string, error) {
    if agent == "" {
        return "", fmt.Errorf("agent parameter required")
    }
    
    sessionID, exists := s.agentSessions.GetLastSession(agent)
    if !exists {
        return "", fmt.Errorf("no sessions found for agent %s", agent)
    }
    
    // Verify session still exists
    sessionPath := filepath.Join(s.baseDir, "sessions", sessionID+".json")
    if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
        return "", fmt.Errorf("session %s no longer exists", sessionID)
    }
    
    log.Printf("🔍 [STORAGE] Retrieved last session for %s: %s", agent, sessionID)
    return sessionID, nil
}

// UpdateLastSession now requires agent parameter
func (s *Storage) UpdateLastSession(agent, sessionID string) error {
    if agent == "" {
        return fmt.Errorf("agent parameter required")
    }
    
    return s.agentSessions.SetLastSession(agent, sessionID)
}

// SaveSession needs to extract agent from session
func (s *Storage) SaveSession(session *Session) error {
    s.indexMutex.Lock()
    defer s.indexMutex.Unlock()
    
    // ... existing save logic ...
    
    // Update last session for this agent
    if session.Agent != "" {
        if err := s.UpdateLastSession(session.Agent, session.ID); err != nil {
            log.Printf("⚠️ [STORAGE] Failed to update last session: %v", err)
        }
    }
    
    return nil
}
```

### 2. Protocol Updates

#### 2.1 Update GetLastSession Request (`cli/src/protocol/mod.rs`)
```rust
#[derive(Debug, Serialize, Deserialize)]
pub struct GetLastSessionRequest {
    pub agent: String,  // Required: which agent's last session
}
```

#### 2.2 Update Client Method (`cli/src/client.rs`)
```rust
impl DaemonClient {
    pub fn get_last_session(&self, agent: &str) -> Result<String> {
        let request = Request {
            id: generate_request_id(),
            request_type: "get_last_session".to_string(),
            payload: json!({
                "agent": agent
            }),
            references: vec![],
        };
        
        let response = self.send_request(request)?;
        
        if !response.success {
            bail!("Failed to get last session for {}: {}", 
                  agent, response.error.unwrap_or_default());
        }
        
        let data = response.data
            .ok_or_else(|| anyhow!("No data in response"))?;
        
        data["session_id"]
            .as_str()
            .ok_or_else(|| anyhow!("No session_id in response"))
            .map(|s| s.to_string())
    }
}
```

### 3. Daemon Handler Updates (`daemon/src/server.go`)

```go
func (d *Daemon) handleGetLastSession(req Request) Response {
    resp := NewResponse(req.ID, true)
    
    // Extract agent from request
    agent, ok := req.Payload["agent"].(string)
    if !ok || agent == "" {
        resp.SetError("agent parameter required")
        return resp
    }
    
    // Normalize agent name (remove @ if present)
    agent = strings.TrimPrefix(agent, "@")
    
    if d.storage == nil {
        resp.SetError("Storage not initialized")
        return resp
    }
    
    sessionID, err := d.storage.GetLastSession(agent)
    if err != nil {
        resp.SetError(fmt.Sprintf("No sessions found for %s: %v", agent, err))
        return resp
    }
    
    // Load session to get metadata
    session, err := d.storage.LoadSession(sessionID)
    if err != nil {
        resp.SetError(fmt.Sprintf("Failed to load session: %v", err))
        return resp
    }
    
    data := map[string]interface{}{
        "session_id":    session.ID,
        "agent":         session.Agent,
        "last_activity": session.LastActivity.Format(time.RFC3339),
        "message_count": len(session.Messages),
    }
    
    resp.SetData(data)
    return resp
}
```

### 4. CLI Updates (`cli/src/main.rs`)

```rust
// In possess command handler
Some(Commands::Possess { agent, session, references, message }) => {
    let message_text = if message.is_empty() { 
        None 
    } else { 
        Some(message.join(" ")) 
    };
    
    // Handle special "last" value with agent context
    let session_id = match session.as_deref() {
        Some("last") => {
            let client = DaemonClient::new(port);
            // Pass the agent to get_last_session
            match client.get_last_session(&agent) {
                Ok(id) => Some(id),
                Err(_) => {
                    eprintln!("No previous sessions found for {}", agent);
                    std::process::exit(1);
                }
            }
        },
        Some(id) => Some(id.to_string()),
        None => None,
    };
    
    if std::env::var("PORT42_DEBUG").is_ok() {
        eprintln!("DEBUG possess: agent={}, session={:?}, message={:?}", 
                 agent, session_id, message_text);
    }
    
    let show_boot = message_text.is_none();
    commands::possess::handle_possess_with_references(
        port, agent, message_text, session_id, references, show_boot
    )?;
}
```

## File Structure

```
~/.port42/
├── sessions/           # Session JSON files
│   ├── cli-123.json
│   └── cli-456.json
└── agent_sessions.json # Agent -> Session mapping
```

Example `agent_sessions.json`:
```json
{
  "ai-engineer": "cli-1757350466684",
  "ai-analyst": "cli-1757350123456",
  "ai-muse": "cli-1757340987654",
  "ai-founder": "cli-1757330555123"
}
```

## Implementation Steps

### Phase 1: Daemon Storage Layer
1. **Create AgentSessions component** (`daemon/src/storage.go`)
   - Implement thread-safe JSON-based mapping
   - Add Load/Save methods for persistence
   - Add Get/Set methods for agent lookups

2. **Update Storage integration** (`daemon/src/storage.go`)
   - Add agentSessions field to Storage
   - Initialize in NewStorage
   - Update GetLastSession to require agent parameter
   - Update UpdateLastSession to require agent parameter
   - Update SaveSession to extract agent and call UpdateLastSession

### Phase 2: Daemon Request Handler
3. **Update Daemon handler** (`daemon/src/server.go`)
   - Modify handleGetLastSession to extract agent from request
   - Pass agent to storage.GetLastSession
   - Return agent-specific session data

### Phase 3: CLI Protocol & Client
4. **Update Protocol** (`cli/src/protocol/mod.rs`)
   - Add agent field to GetLastSessionRequest
   - Ensure request structure matches daemon expectations

5. **Update Client** (`cli/src/client.rs`)
   - Modify get_last_session to accept agent parameter
   - Include agent in request payload

### Phase 4: CLI Command Handler
6. **Update CLI** (`cli/src/main.rs`)
   - Pass agent when requesting last session
   - Better error messages for agent-specific failures

## Testing

```bash
# Test agent isolation
port42 possess @ai-engineer "test message 1"
port42 possess @ai-analyst "test message 2"
port42 possess @ai-engineer --session last  # Should resume ai-engineer session
port42 possess @ai-analyst --session last   # Should resume ai-analyst session

# Test persistence
# Restart daemon
port42 possess @ai-engineer --session last  # Should still work

# Test error cases
port42 possess @ai-muse --session last      # Should error: no sessions for ai-muse
```

## Benefits

1. **Proper Isolation**: Each agent maintains its own session history
2. **Thread Safety**: Dedicated mutex for agent sessions prevents races
3. **Persistence**: JSON file survives daemon restarts
4. **Simplicity**: Clean separation of concerns
5. **Extensibility**: Easy to add features like session limits per agent

## Example Usage After Implementation

```bash
# Each agent maintains its own "last" session
$ port42 possess @ai-engineer "build a parser"
Session: cli-1757350466684

$ port42 possess @ai-analyst "analyze this data"
Session: cli-1757350123456

$ port42 possess @ai-engineer --session last
🔄 Resuming session cli-1757350466684 for @ai-engineer (1 message)

$ port42 possess @ai-analyst --session last  
🔄 Resuming session cli-1757350123456 for @ai-analyst (1 message)

# Check agent sessions
$ cat ~/.port42/agent_sessions.json
{
  "ai-engineer": "cli-1757350466684",
  "ai-analyst": "cli-1757350123456"
}
```