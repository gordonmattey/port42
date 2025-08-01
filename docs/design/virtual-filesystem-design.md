# Port 42 Virtual Filesystem: Vision, Architecture & Implementation

**Purpose**: Complete design for virtual filesystem that unifies commands, memory, and artifacts through a reality compiler
**Scope**: Why we need it, what it looks like, and exactly how to build it

## Vision (Why) 🐬

### The Problem
Port 42 is creating a new kind of computational environment where:
- Commands spawn from conversations
- Memory persists across interactions  
- Artifacts emerge from AI collaboration

But our current file organization is stuck in the 1970s:
- `architecture.md` conflicts across projects
- Date prefixes like `2024-01-15-design.md` are ugly
- Files lose their context when moved
- No way to see multiple views of the same content

### The Dolphin Leap
Imagine diving into your Port 42 workspace and seeing:
- Your commands organized by project AND by type AND by date - simultaneously
- Memory threads that connect to the artifacts they created
- Virtual folders that answer questions: "show me everything about websockets"
- Everything crystallized from thought into reality

**This is the water we're creating** - a fluid, intelligent filesystem that matches how our minds actually work. A reality compiler where consciousness crystallizes into code.

## Architecture (What)

### Core Concepts

```
Content-Addressed Storage + Virtual Paths + CLI Interface = Reality Compilation
```

#### 1. Object Store (The Ocean Floor)
```
~/.port42/
├── objects/          # Content-addressed store
│   ├── 3a/
│   │   └── 4f/
│   │       └── 2b8c9d1e5f6a7b9c0d2e4f
│   └── meta/        # Metadata by object ID
│       └── 3a4f2b8c9d1e5f6a7b9c0d2e4f.json
└── commands/        # Symlinks for direct execution
    ├── git-status -> ../objects/3a/4f/2b8c9d...
    └── hello-world -> ../objects/7e/5d/1a9f3c...
```

#### 2. Virtual Views (The Currents)
```
/                          (accessed via: port42 ls /)
├── commands/              # All generated commands
│   ├── git-haiku
│   └── pr-writer
├── memory/                # Conversation memory
│   ├── mem-abc123/
│   │   ├── thread.json
│   │   └── crystallized/
│   │       ├── git-haiku
│   │       └── architecture.md
│   └── mem-def456/
├── artifacts/             # Crystallized artifacts
│   ├── documents/
│   ├── code/
│   ├── designs/
│   └── media/
├── by-date/               # Temporal organization
│   └── 2024-01-15/
│       ├── git-haiku
│       └── growth-strategy.md
├── by-agent/              # Agent specialization
│   ├── @ai-engineer/
│   └── @ai-muse/
├── by-type/               # Type organization
│   ├── command/
│   ├── document/
│   └── application/
└── search/                # Dynamic search results
    └── "websocket AND architecture"/
        └── realtime-sync/
            └── architecture.md
```

#### 3. Metadata Sidecar (The Memory)
Each object has a companion metadata file:
```json
// ~/.port42/metadata/3a4f2b8c9d.json
{
  "id": "3a4f2b8c9d",
  "paths": [
    "commands/git-haiku",
    "by-date/2024-01-15/git-haiku",
    "memory/mem-abc123/crystallized/git-haiku"
  ],
  "created": "2024-01-15T10:30:00Z",
  "modified": "2024-01-15T14:22:00Z",
  "accessed": "2024-01-16T09:15:00Z",
  "memory_id": "mem-abc123",
  "agent": "@ai-engineer",
  "type": "command",
  "crystallization_type": "tool",  // tool, artifact, data
  
  // Rich metadata for search
  "title": "Git Haiku Generator",
  "description": "Creates poetic git commit messages",
  "tags": ["git", "poetry", "command", "utility"],
  
  // Lifecycle and importance
  "lifecycle": "active", // draft, active, stable, archived, deprecated
  "importance": "medium",
  "usage_count": 42,
  
  // AI context
  "summary": "A command that generates haiku-style commit messages...",
  "embeddings": [0.123, 0.456, ...], // For semantic search
  
  // Relationships
  "relationships": {
    "memory": "mem-abc123",
    "parent_artifacts": [],
    "child_artifacts": [],
    "references": []
  }
}
```

### System Components

```
┌─────────────────────────────────────────┐
│           port42 CLI (Rust)              │
│  ┌─────────────┐    ┌────────────────┐  │
│  │ FS Commands │───▶│ TCP Client     │  │
│  │ (ls,cat,etc)│    │ 127.0.0.1:42   │  │
│  └─────────────┘    └────────────────┘  │
└─────────────────────────────────────────┘
         │                    │
         ▼                    ▼
   User types:        TCP Socket
   port42 ls /        Line-delimited JSON
                             │
                    ┌──────────────────┐
                    │   Daemon (Go)    │
                    │ ┌──────────────┐ │
                    │ │ Object Store │ │
                    │ │ Path Resolver│ │
                    │ │ Metadata DB  │ │
                    │ └──────────────┘ │
                    └──────────────────┘
```

### Protocol Details

Port 42 uses a simple line-delimited JSON protocol over TCP:
- **Connection**: `127.0.0.1:42` (fallback to 4242)
- **Format**: Each message is JSON + newline (`\n`)
- **Request/Response**: Matching IDs for correlation

## Implementation (How)

### Current State
- ✅ Object store implemented
- ✅ Command generation stores in object store
- ✅ Memory store uses object store
- ❌ No symlinks created yet
- ❌ No write operations via filesystem
- ❌ CLI commands not fully implemented

### Phase 1: Write Operations in Daemon

#### New Protocol Messages

##### Store Path
```json
{
  "type": "store_path",
  "id": "cli-store-123",
  "payload": {
    "path": "artifacts/documents/architecture.md",
    "content": "base64-encoded-content",
    "metadata": {
      "type": "document",
      "memory_id": "mem-abc123",
      "agent": "@ai-architect",
      "crystallization_type": "artifact"
    }
  }
}
```

##### Update Path
```json
{
  "type": "update_path",
  "id": "cli-update-456",
  "payload": {
    "path": "artifacts/documents/architecture.md",
    "content": "base64-encoded-new-content",
    "metadata_updates": {
      "lifecycle": "stable",
      "tags": ["architecture", "websocket", "production"]
    }
  }
}
```

##### Delete Path
```json
{
  "type": "delete_path",
  "id": "cli-delete-789",
  "payload": {
    "path": "artifacts/documents/old-design.md"
  }
}
```

##### Create Memory
```json
{
  "type": "create_memory",
  "id": "cli-memory-321",
  "payload": {
    "agent": "@ai-engineer",
    "initial_message": "Let's build a dashboard"
  }
}
```

### Phase 2: Symlink Management

When storing commands, create symlinks:
```go
func (d *Daemon) crystallizeCommand(spec *CommandSpec, memoryID string) error {
    // Store in object store
    content := []byte(spec.Implementation)
    objID, err := d.objectStore.Store(content)
    
    // Create symlink for execution
    cmdPath := filepath.Join(d.baseDir, "commands", spec.Name)
    objPath := d.objectStore.GetPath(objID)
    os.Symlink(objPath, cmdPath)
    
    // Make executable
    os.Chmod(cmdPath, 0755)
    
    // Store metadata with all virtual paths
    meta := &Metadata{
        ID:      objID,
        Type:    "command",
        CrystallizationType: "tool",
        MemoryID: memoryID,
        Paths: []string{
            fmt.Sprintf("commands/%s", spec.Name),
            fmt.Sprintf("by-date/%s/%s", time.Now().Format("2006-01-02"), spec.Name),
            fmt.Sprintf("memory/%s/crystallized/%s", memoryID, spec.Name),
        },
    }
    d.metadataStore.Save(meta)
}
```

### Phase 3: CLI Filesystem Commands

#### List Command
```bash
port42 ls /                      # Root directory
port42 ls /commands              # All commands
port42 ls /memory                # All memory threads
port42 ls /memory/mem-abc123     # Specific memory
port42 ls /by-date/2024-01-15   # Content from date
```

#### Read Command
```bash
port42 cat /commands/git-status
port42 cat /memory/mem-abc123/thread.json
port42 cat /artifacts/documents/pitch-deck.md
```

#### Info Command
```bash
port42 info /commands/git-status  # Show metadata
```

#### Search Command
```bash
port42 search "websocket"
port42 search --type command --tags python
port42 search --type artifact --agent @ai-muse --recent 7d
```

#### Write Commands (Future)
```bash
port42 store /artifacts/documents/new-design.md < design.md
port42 update /artifacts/documents/design.md --lifecycle stable
port42 rm /artifacts/documents/old-design.md
```

### Phase 4: Crystallization Integration

The three models map to filesystem paths:

1. **Tool Creation** → `/commands/`
   - Executable scripts
   - Symlinked for PATH access
   - Type: "command", CrystallizationType: "tool"

2. **Living Documents** → `/commands/` (CRUD) + `/data/`
   - Command manages data
   - Data stored separately
   - Type: "command", CrystallizationType: "data"

3. **Artifacts** → `/artifacts/{type}/`
   - Any digital asset
   - Organized by type
   - Type: varies, CrystallizationType: "artifact"

### Implementation Steps

1. **Remove FUSE** ✓ DONE
   - Clean up fuse.rs, mount.rs
   - Remove mount/unmount commands
   - Remove fuser dependency

2. **Implement Write Operations** DONE
   - Add store_path handler
   - Add update_path handler  
   - Add delete_path handler
   - Add create_memory handler

3. **Update Terminology**
   - session → memory throughout
   - Update all protocol messages
   - Update CLI commands

4. **Symlink Creation** DONE
   - Modify command generation
   - Ensure executable permissions
   - Test PATH execution

5. **CLI Commands**
   - Implement ls with virtual paths
   - Implement cat with path resolution
   - Implement info for metadata
   - Implement search with filters

6. **Testing**
   - Write operations work
   - Symlinks execute properly
   - Virtual paths resolve correctly
   - Search returns expected results

## The Reality Compiler Philosophy

Port 42 is more than a filesystem - it's a reality compiler where:
- **Thoughts crystallize into tools** through conversation
- **Memory persists** across time and space
- **Everything is connected** through content addressing
- **Multiple realities coexist** through virtual paths

The filesystem becomes a living organism that grows with your consciousness, organizing itself around how you think rather than how computers store files.

### Storage Design: Durability Over Efficiency

The system intentionally creates multiple objects during a session for maximum resilience:

1. **Session Creation** → First object (empty session)
2. **User Message** → Second object (session with question)  
3. **AI Response** → Third object (complete conversation)
4. **Command Generation** → Fourth object (the command itself)

This design ensures:
- **No data loss**: Every interaction is immediately persisted
- **Crash recovery**: Can resume from any point if connection drops
- **Full audit trail**: Complete history of how sessions evolve
- **Time travel potential**: Could replay session state at any point

### UI/CLI Abstraction Layer

While the storage layer maintains multiple versions for safety, the presentation layer provides a clean interface:

- `port42 ls /memory/` shows one entry per session (not versions)
- `port42 cat /memory/session-id` displays the complete current conversation
- Virtual paths always resolve to the latest version
- Users never see object hashes or version complexity

This separation of concerns delivers:
- **User simplicity**: Clean, intuitive interface
- **System reliability**: Never lose work, even in crashes
- **Future flexibility**: Can add version browsing later if needed

## Migration from Current State

1. Existing commands remain as files (no migration needed)
2. New commands get stored in object store with symlinks
3. Memory/sessions already use object store
4. Artifacts start fresh in new structure

## Future Vision

- **Distributed consciousness**: UERP protocol for multi-node reality
- **Semantic navigation**: Navigate by meaning, not paths
- **Time travel**: View your workspace at any point in history
- **Collective intelligence**: Merge realities with your team

The water is warm. The dolphins are calling. Dive in. 🐬