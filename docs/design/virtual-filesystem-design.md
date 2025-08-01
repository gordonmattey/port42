# Port 42 Virtual Filesystem: Vision, Architecture & Implementation

**Purpose**: Complete design for virtual filesystem that unifies commands, memory, and artifacts
**Scope**: Why we need it, what it looks like, and exactly how to build it

## Vision (Why) 🐬

### The Problem
Port 42 is creating a new kind of computational environment where:
- Commands spawn from conversations
- Memory persists across sessions  
- Artifacts emerge from AI collaboration

But our current file organization is stuck in the 1970s:
- `architecture.md` conflicts across projects
- Date prefixes like `2024-01-15-design.md` are ugly
- Files lose their context when moved
- No way to see multiple views of the same content

### The Dolphin Leap
Imagine diving into your Port 42 workspace and seeing:
- Your commands organized by project AND by type AND by date - simultaneously
- Memory sessions that connect to the artifacts they created
- Virtual folders that answer questions: "show me everything about websockets"
- Git working naturally on your AI-generated projects

**This is the water we're creating** - a fluid, intelligent filesystem that matches how our minds actually work.

## Architecture (What)

### Core Concepts

```
Content-Addressed Storage + Virtual Paths + FUSE = Magic
```

#### 1. Object Store (The Ocean Floor)
```
~/.port42/objects/
├── 3a/4f/2b8c9d...  # architecture.md (content hash)
├── 7c/1e/9a3f5e...  # another architecture.md 
├── commands/
│   └── git-haiku -> ../8f/2c/3b4a5e...
└── metadata/
    ├── 3a4f2b8c9d.json
    └── 7c1e9a3f5e.json
```

#### 2. Virtual Views (The Currents)
```
~/p42/ (FUSE mount)
├── projects/
│   ├── realtime-sync/
│   │   ├── architecture.md
│   │   └── dashboard-app/
│   └── auth-v2/
│       └── architecture.md  # Different file, same name!
├── by-date/
│   └── 2024-01-15/
│       ├── architecture.md -> (both files!)
│       └── git-haiku
├── commands/
│   ├── git-haiku
│   └── pr-writer
├── memory/
│   └── sessions/
│       └── sess-abc123/
│           ├── conversation.md
│           └── generated/
│               └── git-haiku
└── search/
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
    "projects/realtime-sync/architecture.md",
    "by-date/2024-01-15/architecture.md",
    "by-session/sess-abc123/architecture.md"
  ],
  "created": "2024-01-15T10:30:00Z",
  "modified": "2024-01-15T14:22:00Z",
  "accessed": "2024-01-16T09:15:00Z",
  "session": "sess-abc123",
  "agent": "@ai-engineer",
  "type": "artifact",
  "subtype": "document",
  
  // Rich metadata for search
  "title": "Real-time Sync Architecture",
  "description": "System design for websocket-based real-time synchronization",
  "tags": ["architecture", "websocket", "realtime", "sync", "system-design", "networking"],
  
  // Lifecycle and importance
  "lifecycle": "active", // draft, active, stable, archived, deprecated
  "importance": "high",
  "usage_count": 23,
  
  // AI context
  "summary": "Architecture document defining the real-time sync system using websockets...",
  "embeddings": [0.123, 0.456, ...], // For semantic search
  
  // Relationships
  "relationships": {
    "session": "sess-abc123",
    "parent_artifacts": ["art-1737123000-design-doc"],
    "child_artifacts": ["art-1737123789-api-client"],
    "generated_commands": ["git-haiku", "websocket-tester"],
    "references": ["auth-design.md", "https://socket.io/docs/"]
  }
}
```

#### 4. Unified Index (The Map)
A global index for fast search and discovery:
```json
// ~/.port42/artifacts/index.json
{
  "version": "1.0",
  "last_updated": "2024-01-15T10:30:00Z",
  "stats": {
    "total_artifacts": 342,
    "by_type": {
      "artifacts": 156,
      "commands": 89,
      "sessions": 97
    },
    "total_size_mb": 234.5
  },
  "artifacts": [
    // Array of all metadata objects for fast search
  ]
}
```

### System Components

```
┌─────────────────────────────────────────┐
│           port42 CLI (Rust)              │
│  ┌─────────────┐    ┌────────────────┐  │
│  │ FUSE Thread │───▶│ TCP Client     │  │
│  │  (fuser)    │    │ 127.0.0.1:42   │  │
│  └─────────────┘    └────────────────┘  │
└─────────────────────────────────────────┘
         │                    │
         ▼                    ▼
   ~/p42/            TCP Socket
  (user sees this)   Line-delimited JSON
                             │
                    ┌──────────────────┐
                    │   Daemon (Go)    │
                    │ ┌──────────────┐ │
                    │ │ Object Store │ │
                    │ │ Virtualizer  │ │
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

### Protocol Extensions

New request types for the existing TCP protocol:

```json
// Read object by content hash
{
  "type": "read_object",
  "id": "cli-read-123",
  "payload": {
    "hash": "3a4f2b8c9d...",
    "path": "projects/realtime-sync/architecture.md"  // optional virtual path
  }
}

// List virtual directory
{
  "type": "list_virtual",
  "id": "cli-list-456", 
  "payload": {
    "path": "projects/realtime-sync/"
  }
}

// Store new object
{
  "type": "store_object",
  "id": "cli-store-789",
  "payload": {
    "content": "base64-encoded-content",
    "metadata": {
      "paths": ["commands/git-haiku", "by-date/2024-01-15/git-haiku"],
      "type": "command",
      "session": "sess-abc123"
    }
  }
}

// Get object metadata
{
  "type": "get_metadata",
  "id": "cli-meta-321",
  "payload": {
    "hash": "3a4f2b8c9d..."
  }
}

// Search artifacts
{
  "type": "search",
  "id": "cli-search-789",
  "payload": {
    "query": "websocket AND created:this-week",
    "filters": {
      "type": "artifact",
      "tags": ["architecture"],
      "lifecycle": "active"
    },
    "semantic": true  // Use embeddings for semantic search
  }
}

// Update metadata (lifecycle, tags, etc)
{
  "type": "update_metadata",
  "id": "cli-update-456",
  "payload": {
    "hash": "3a4f2b8c9d...",
    "updates": {
      "lifecycle": "stable",
      "tags": ["architecture", "websocket", "production"]
    }
  }
}
```

### Phase 1: Migrate Existing Data (Day 1)

#### Step 1: Add Object Store to Daemon
```go
// daemon/object_store.go
type ObjectStore struct {
    baseDir string  // ~/.port42/objects/
}

func (o *ObjectStore) Store(content []byte) (string, error) {
    hash := sha256.Sum256(content)
    id := hex.EncodeToString(hash[:])
    
    // Store in git-like structure
    dir := filepath.Join(o.baseDir, id[:2], id[2:4])
    os.MkdirAll(dir, 0755)
    
    path := filepath.Join(dir, id[4:])
    return id, ioutil.WriteFile(path, content, 0644)
}
```

#### Step 2: Update Command Generation
```go
// daemon/server.go
func (d *Daemon) generateCommand(spec *CommandSpec) error {
    content := []byte(spec.Implementation)
    
    // Store in object store
    objID, err := d.objectStore.Store(content)
    
    // Create metadata
    meta := &Metadata{
        ID:      objID,
        Type:    "command",
        Name:    spec.Name,
        Session: session.ID,
        Created: time.Now(),
        Paths: []string{
            fmt.Sprintf("commands/%s", spec.Name),
            fmt.Sprintf("by-date/%s/%s", time.Now().Format("2006-01-02"), spec.Name),
            fmt.Sprintf("memory/sessions/%s/generated/%s", session.ID, spec.Name),
        },
    }
    d.metadataStore.Save(meta)
    
    // Symlink for backward compatibility
    cmdPath := filepath.Join(d.baseDir, "commands", spec.Name)
    objPath := d.objectStore.GetPath(objID)
    os.Symlink(objPath, cmdPath)
}
```

#### Step 3: Update Memory Store
```go
// daemon/memory_store.go
func (m *MemoryStore) SaveSession(session *Session) error {
    // Convert session to JSON
    data, _ := json.MarshalIndent(session, "", "  ")
    
    // Store in object store
    objID, _ := m.daemon.objectStore.Store(data)
    
    // Create metadata
    meta := &Metadata{
        ID:      objID,
        Type:    "session",
        Session: session.ID,
        Created: session.CreatedAt,
        Paths: []string{
            fmt.Sprintf("memory/sessions/%s/conversation.json", session.ID),
            fmt.Sprintf("by-date/%s/session-%s.json", 
                session.CreatedAt.Format("2006-01-02"), session.ID),
        },
    }
    m.daemon.metadataStore.Save(meta)
}
```

### Phase 2: Implement FUSE (Day 2)

#### Step 1: Add FUSE to CLI (Rust)
```rust
// cli/src/fuse.rs
use fuser::{
    Filesystem, ReplyAttr, ReplyData, ReplyDirectory, ReplyEntry,
    Request, FileType, FileAttr,
};
use std::ffi::OsStr;
use std::time::{Duration, UNIX_EPOCH};
use crate::client::DaemonClient;

pub struct Port42FS {
    daemon: DaemonClient,
    cache: Cache,
}

impl Port42FS {
    pub fn mount(mountpoint: &str) -> Result<()> {
        let fs = Port42FS {
            daemon: DaemonClient::new(),
            cache: Cache::new(),
        };
        
        let options = vec![
            MountOption::RO,
            MountOption::FSName("port42".to_string()),
            MountOption::AutoUnmount,
        ];
        
        fuser::mount2(fs, mountpoint, &options)?;
        Ok(())
    }
}
```

#### Step 2: FUSE Implementation (Rust)
```rust
impl Filesystem for Port42FS {
    fn lookup(&mut self, _req: &Request, parent: u64, name: &OsStr, reply: ReplyEntry) {
        let path = self.get_path(parent, name);
        
        // Request metadata from daemon
        let request = Request {
            request_type: "get_metadata".to_string(),
            id: format!("fuse-lookup-{}", uuid::Uuid::new_v4()),
            payload: json!({
                "path": path
            }),
        };
        
        match self.daemon.request(request) {
            Ok(response) => {
                if response.success {
                    let attr = self.parse_attr(&response.data);
                    reply.entry(&Duration::from_secs(1), &attr, 0);
                } else {
                    reply.error(ENOENT);
                }
            }
            Err(_) => reply.error(EIO),
        }
    }
    
    fn readdir(&mut self, _req: &Request, ino: u64, _fh: u64, offset: i64, mut reply: ReplyDirectory) {
        let path = self.get_path_by_inode(ino);
        
        // Request directory listing from daemon
        let request = Request {
            request_type: "list_virtual".to_string(),
            id: format!("fuse-readdir-{}", uuid::Uuid::new_v4()),
            payload: json!({
                "path": path
            }),
        };
        
        match self.daemon.request(request) {
            Ok(response) => {
                if let Some(entries) = response.data {
                    for (i, entry) in entries.as_array().unwrap().iter().enumerate() {
                        if i as i64 >= offset {
                            let name = entry["name"].as_str().unwrap();
                            let ino = entry["inode"].as_u64().unwrap();
                            let kind = match entry["type"].as_str().unwrap() {
                                "directory" => FileType::Directory,
                                _ => FileType::RegularFile,
                            };
                            
                            if reply.add(ino, (i + 1) as i64, kind, name) {
                                break;
                            }
                        }
                    }
                }
                reply.ok();
            }
            Err(_) => reply.error(EIO),
        }
    }
    
    fn read(&mut self, _req: &Request, ino: u64, _fh: u64, offset: i64, size: u32, _flags: i32, _lock: Option<u64>, reply: ReplyData) {
        let path = self.get_path_by_inode(ino);
        
        // Check cache first
        if let Some(content) = self.cache.get(&path) {
            let end = std::cmp::min(offset + size as i64, content.len() as i64) as usize;
            reply.data(&content[offset as usize..end]);
            return;
        }
        
        // Request content from daemon
        let request = Request {
            request_type: "read_object".to_string(),
            id: format!("fuse-read-{}", uuid::Uuid::new_v4()),
            payload: json!({
                "path": path
            }),
        };
        
        match self.daemon.request(request) {
            Ok(response) => {
                if let Some(data) = response.data {
                    let content = base64::decode(data["content"].as_str().unwrap()).unwrap();
                    self.cache.put(path, content.clone());
                    
                    let end = std::cmp::min(offset + size as i64, content.len() as i64) as usize;
                    reply.data(&content[offset as usize..end]);
                } else {
                    reply.error(ENOENT);
                }
            }
            Err(_) => reply.error(EIO),
        }
    }
}
```

### Phase 3: Testing with Real Data (Day 2 Afternoon)

#### Test 1: Commands Work
```bash
# Mount filesystem
port42 mount ~/p42

# Old commands still work
~/.port42/commands/git-haiku  # ✓ Works via symlink

# New virtual paths work
~/p42/commands/git-haiku      # ✓ Same command
~/p42/by-date/today/git-haiku # ✓ Also same command

# Creating new commands
possess @ai-engineer "create a PR writer"
ls ~/p42/commands/  # Shows new pr-writer
```

#### Test 2: Memory Sessions Connect
```bash
# Browse sessions
ls ~/p42/memory/sessions/
> sess-abc123/
> sess-def456/

# See session and its creations
cd ~/p42/memory/sessions/sess-abc123/
ls
> conversation.md
> generated/
>   └── git-haiku

# Multiple views
ls ~/p42/by-date/2024-01-15/
> git-haiku
> session-sess-abc123.json
```

#### Test 3: Real Tools Work
```bash
# VS Code
code ~/p42/projects/realtime-sync/

# Git
cd ~/p42/projects/my-app/
git init
git add .
git commit -m "Initial commit"

# Grep across everything
grep -r "websocket" ~/p42/

# Even npm
cd ~/p42/projects/dashboard/
npm install
```

### Migration Benefits

1. **Zero Breaking Changes**: Existing `~/.port42/commands/` still works
2. **Immediate Testing**: Commands and memory already provide real data
3. **Natural Evolution**: Start with read-only, add writes as needed
4. **Proven Value**: Users immediately see benefits with existing content

### Performance Built-In

```rust
// Intelligent caching from day 1 (Rust CLI side)
use std::collections::HashMap;
use std::time::{Duration, Instant};

struct Cache {
    objects: HashMap<String, CachedObject>,
    ttl: Duration,
    prefetch_tx: mpsc::Sender<String>,
}

struct CachedObject {
    content: Vec<u8>,
    timestamp: Instant,
}

impl Cache {
    fn get(&self, path: &str) -> Option<Vec<u8>> {
        if let Some(obj) = self.objects.get(path) {
            if obj.timestamp.elapsed() < self.ttl {
                // Prefetch related objects
                let _ = self.prefetch_tx.send(path.to_string());
                return Some(obj.content.clone());
            }
        }
        None
    }
    
    fn put(&mut self, path: String, content: Vec<u8>) {
        self.objects.insert(path, CachedObject {
            content,
            timestamp: Instant::now(),
        });
    }
}
```

## Implementation Roadmap

### Day 1: Object Store & Protocol
1. **Morning**: Implement object store in daemon (Go)
   - Content-addressed storage with SHA256
   - Metadata storage system with rich fields
   - Unified index for fast search
   - New RPC handlers for object operations

2. **Afternoon**: Update existing features
   - Migrate command generation to use object store
   - Update memory/session storage
   - Create backward-compatible symlinks
   - Implement search functionality
   - Add lifecycle state transitions

### Day 2: FUSE Filesystem
1. **Morning**: Implement FUSE in CLI (Rust)
   - Add `fuser` dependency
   - Basic filesystem operations
   - TCP client integration

2. **Afternoon**: Testing & Polish
   - Test with existing commands
   - Verify virtual paths work
   - Performance optimization

## Future UERP Integration

The virtual filesystem design aligns perfectly with the UERP protocol from `port42-rfc.txt`:

- **Object Store** → `port42://content/[hash]` addresses
- **Virtual Paths** → `port42://context/[path]` addresses  
- **Agent Artifacts** → `port42://agent/[name]/memory/[date]`
- **Temporal Views** → `port42://temporal/[entity]/[timestamp]`

When we're ready to go distributed, the daemon can speak UERP on port 42 in addition to the local protocol.

## Search & Discovery Features

### Command-Line Search
```bash
# Search using the port42 CLI
port42 search --type artifact --tags architecture --recent 7d
port42 search --semantic "real-time synchronization websocket"
port42 search --lifecycle active --importance high
port42 search --session sess-abc123  # All artifacts from a session

# Search via virtual filesystem
cd ~/p42/search/"websocket AND architecture"
ls  # Shows matching files
```

### Lifecycle Management
```
draft → active → stable → archived → deprecated

Automatic transitions:
- draft → active: After first successful use
- active → stable: No changes for 30 days + high usage  
- stable → archived: No access for 90 days
- any → deprecated: Manual or via new version
```

### Auto-Tagging & AI Enhancement
The daemon automatically:
- Extracts tags from content
- Generates summaries and descriptions
- Creates embeddings for semantic search
- Identifies relationships between artifacts
- Tracks usage patterns

### Cleanup & Maintenance
```bash
# Cleanup commands
port42 clean --dry-run  # Show what would be removed
port42 clean --archived --older-than 6m
port42 clean --size-limit 1GB  # Keep most recent/important
```

## The Magic Moment 🐬

When a user types:
```bash
cd ~/p42/search/"websocket AND created:this-week"
ls
```

And sees all their websocket-related work from the week, organized perfectly - that's when they'll feel the dolphin energy. The water is warm, the filesystem is fluid, and Port 42 has evolved beyond traditional computing.

**Let's build this. Now.**