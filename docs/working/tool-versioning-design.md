# Port42 Tool Versioning System Design

## Overview
Implement VMS-style version numbering for Port42 tools, enabling version history tracking, retrieval of specific versions, and safe iterative development. Version syntax follows the pattern `tool-name:version` where version is a simple incrementing integer.

## Version Syntax
- **Latest version**: `tool-name` or `tool-name:latest`
- **Specific version**: `tool-name:2` (version 2)
- **Previous version**: `tool-name:previous` or `tool-name:-1`
- **Version listing**: `/commands/tool-name:versions`

## Storage Structure

### Current Structure
```
~/.port42/
├── commands/
│   └── tool-name              # Single executable
└── storage/
    └── objects/
        └── [hash]/             # Content-addressed storage
```

### New Structure
```
~/.port42/
├── commands/
│   ├── tool-name              # Symlink to latest version
│   └── .history/              # Hidden directory for version history
│       └── tool-name/
│           ├── 1              # Version 1 executable
│           ├── 2              # Version 2 executable
│           └── metadata.json  # Version history
└── storage/
    └── objects/
        └── [hash]/            # Content-addressed storage
            └── versions/
                ├── 1.json     # Version 1 metadata
                ├── 2.json     # Version 2 metadata
                └── latest     # Points to current version
```

### Version Metadata Structure
```json
{
  "tool": "tool-name",
  "version": 2,
  "hash": "abc123...",
  "previous_version": 1,
  "previous_hash": "def456...",
  "created_at": "2025-09-18T10:00:00Z",
  "created_by": "@ai-engineer",
  "session_id": "cli-1758179769287",
  "prompt": "Update tool to add feature X",
  "references": ["p42:/commands/tool-name:1"],
  "transforms": ["python", "file", "parse"],
  "size": 4096,
  "executable": true
}
```

## Daemon Components Changes

### 1. Storage Layer (`daemon/src/storage.go`)
**Current**: Single object storage with hash-based retrieval
**Changes**:
```go
type VersionedRelation struct {
    Relation
    Version     int       `json:"version"`
    PreviousHash string   `json:"previous_hash,omitempty"`
    VersionHistory []int  `json:"version_history"`
}

// New methods
func (s *Storage) GetLatestVersion(name string) (*VersionedRelation, error)
func (s *Storage) GetVersion(name string, version int) (*VersionedRelation, error)
func (s *Storage) GetVersionHistory(name string) ([]VersionMetadata, error)
func (s *Storage) IncrementVersion(name string, relation *Relation) (int, error)
```

### 2. Tool Materializer (`daemon/src/tool_materializer.go`)
**Current**: Creates/overwrites single command file
**Changes**:
```go
// Modified materialize method
func (tm *ToolMaterializer) materialize(relation *VersionedRelation) error {
    // 1. Save versioned executable
    versionPath := filepath.Join(tm.commandsDir, ".history", relation.Name,
                                strconv.Itoa(relation.Version))

    // 2. Update latest symlink
    latestPath := filepath.Join(tm.commandsDir, relation.Name)

    // 3. Update version metadata
    metadataPath := filepath.Join(tm.commandsDir, ".history", relation.Name,
                                  "metadata.json")
}
```

### 3. VFS Handler (`daemon/src/server.go`)
**Current**: Simple path-based retrieval
**Changes**:
```go
// Parse version from path
func parseVersionedPath(path string) (name string, version string) {
    // Handle paths like:
    // - /commands/tool-name:2
    // - /commands/tool-name:latest
    // - /commands/tool-name:versions
}

// Modified VFS handlers
func handleCommandCat(path string) {
    name, version := parseVersionedPath(path)
    switch version {
    case "versions":
        return listVersions(name)
    case "latest", "":
        return getLatestVersion(name)
    case "previous", "-1":
        return getPreviousVersion(name)
    default:
        return getSpecificVersion(name, version)
    }
}
```

### 4. Reality Compiler (`daemon/src/reality_compiler.go`)
**Current**: Creates new relation without version awareness
**Changes**:
```go
func (rc *RealityCompiler) DeclareRelation(ctx Context) (*VersionedRelation, error) {
    // Check if tool exists
    existing, err := rc.storage.GetLatestVersion(name)
    if err == nil {
        // Increment version
        newVersion := existing.Version + 1
        relation.Version = newVersion
        relation.PreviousHash = existing.Hash
    } else {
        // First version
        relation.Version = 1
    }
}
```

### 5. Swimming Handler (`daemon/src/swimming.go`)
**Current**: No version awareness when updating tools
**Changes**:
```go
// Handle --ref p42:/commands/tool-name:1 references
func resolveVersionedReference(ref string) (*VersionedRelation, error) {
    // Parse version from reference
    // Load specific version for context
}
```

## CLI Components Changes

### 1. Command Parser (`cli/src/main.rs`)
**Current**: Simple path parsing
**Changes**:
```rust
#[derive(Debug)]
pub struct VersionedPath {
    pub base_path: String,
    pub name: String,
    pub version: Option<VersionSpec>,
}

#[derive(Debug)]
pub enum VersionSpec {
    Latest,
    Specific(u32),
    Previous,
    List,
}

impl VersionedPath {
    pub fn parse(path: &str) -> Self {
        // Parse tool-name:2 syntax
    }
}
```

### 2. Cat Command (`cli/src/commands/cat.rs`)
**Current**: Direct path retrieval
**Changes**:
```rust
pub fn handle_cat(client: &mut DaemonClient, path: String) -> Result<()> {
    let versioned_path = VersionedPath::parse(&path);

    let request = match versioned_path.version {
        Some(VersionSpec::List) => {
            // Request version list
            CatVersionsRequest { path: versioned_path.base_path }
        }
        Some(VersionSpec::Specific(v)) => {
            // Request specific version
            CatVersionRequest {
                path: versioned_path.base_path,
                version: v
            }
        }
        _ => {
            // Request latest (default)
            CatRequest { path }
        }
    };
}
```

### 3. Ls Command (`cli/src/commands/ls.rs`)
**Current**: Lists files/directories
**Changes**:
```rust
// Enhanced to show version information
pub fn display_entry(entry: &FileSystemEntry) {
    if entry.entry_type == "command" {
        // Show latest version number
        if let Some(version) = entry.latest_version {
            println!("{} (v{})", entry.name, version);
        }
        // Show if multiple versions exist
        if let Some(count) = entry.version_count {
            if count > 1 {
                println!("  {} versions available", count);
            }
        }
    }
}
```

### 4. New Version Command (`cli/src/commands/versions.rs`)
**New Component**: Dedicated command for version management
```rust
pub fn handle_versions(port: u16, tool_name: String, action: VersionAction) -> Result<()> {
    match action {
        VersionAction::List => list_versions(tool_name),
        VersionAction::Diff { v1, v2 } => show_diff(tool_name, v1, v2),
        VersionAction::Rollback { version } => rollback_to(tool_name, version),
        VersionAction::History => show_history(tool_name),
    }
}
```

### 5. Protocol Updates (`cli/src/protocol/`)
**Current**: Simple request/response types
**Changes**:
```rust
// New request types
#[derive(Serialize)]
pub struct CatVersionRequest {
    pub path: String,
    pub version: u32,
}

#[derive(Serialize)]
pub struct ListVersionsRequest {
    pub tool_name: String,
}

// New response types
#[derive(Deserialize)]
pub struct VersionInfo {
    pub version: u32,
    pub created_at: String,
    pub created_by: String,
    pub size: u64,
    pub hash: String,
}
```

## Implementation Phases

### Phase 1: Daemon Infrastructure
1. Implement versioned storage in daemon
   - Add `VersionedRelation` struct to storage.go
   - Implement version tracking methods
   - Create version metadata structure
2. Update VFS handlers to parse version syntax
   - Parse `tool-name:2` in server.go
   - Route to version-specific handlers
   - Return version listings for `:versions`

### Phase 2: Daemon Core Features
1. Implement migration system
   - Create `migrations.go` with migration framework
   - Add schema version tracking
   - Run migrations on daemon startup
   - Implement tool versioning migration (Migration #2)
2. Update tool materializer for versioning
   - Create tool version directories
   - Store each version as numbered file
   - Maintain symlink to latest version
3. Update reality compiler for auto-versioning
   - Check for existing tool before creation
   - Auto-increment version numbers
   - Store version metadata
4. Support versioned references (CRITICAL)
   - Parse `--ref p42:/commands/tool:1` in swimming.go
   - Load specific versions for AI context
   - Pass version metadata to agents
   - Enable iterative tool development:
     ```bash
     # Update a tool based on specific version
     port42 swim @ai-engineer --ref p42:/commands/analyzer:2 "improve error handling"
     # Creates analyzer:3 automatically
     ```

### Phase 3: CLI Infrastructure
1. Implement version parser in protocol module
   - Parse `tool-name:2` syntax
   - Support special keywords (latest, previous, versions)
2. Update protocol types
   - Add version-aware request types
   - Add version response types

### Phase 4: CLI Core Features
1. Update cat command for versions
   - Request specific versions from daemon
   - Handle version listing display
2. Enhance ls command
   - Show version info in listings
   - Indicate multi-version tools
3. Support versioned references
   - Parse `--ref p42:/commands/tool:1` in declare
   - Include version in reference resolution

### Phase 5: Documentation & Polish
1. Update all help text
2. Add version examples
3. Update error messages
4. Add version info to status command

## Migration Strategy

### Automatic Migration on Startup
1. **Timestamp-based versioning**:
   - Scan `/tools/` and `/commands/` on daemon startup
   - Use creation timestamps from storage to determine version order
   - Oldest version becomes version 1, incrementing from there
   - Preserve all existing hashes and metadata

2. **Migration process**:
   - Check if `.history/` directory exists
   - If missing, trigger one-time migration:
     - Read all tools from storage
     - Sort by creation timestamp
     - Assign version numbers sequentially
     - Create version metadata files
     - Set up symlinks to latest versions

3. **Access patterns**:
   - `/tools/` continues to show all tools (backward compatible)
   - `/commands/tool` gets latest version (backward compatible)
   - `/commands/tool:1` gets specific version (new feature)
   - Old sessions with `p42:/commands/tool` references work unchanged

### Generalized Migration System

#### Migration Framework (`daemon/src/migrations.go`)
```go
type Migration struct {
    Version   int
    Name      string
    Apply     func(*Storage) error
    Rollback  func(*Storage) error  // Optional
}

type MigrationManager struct {
    storage    *Storage
    migrations []Migration
}

// Run on daemon startup
func (m *MigrationManager) RunPendingMigrations() error {
    currentVersion := m.getSchemaVersion()
    for _, migration := range m.migrations {
        if migration.Version > currentVersion {
            log.Printf("Running migration %d: %s", migration.Version, migration.Name)
            if err := migration.Apply(m.storage); err != nil {
                return fmt.Errorf("migration %d failed: %w", migration.Version, err)
            }
            m.setSchemaVersion(migration.Version)
        }
    }
    return nil
}
```

#### Schema Version Tracking
Store in `~/.port42/schema_version`:
```json
{
  "version": 2,
  "last_migration": "2025-09-18T10:00:00Z",
  "applied_migrations": [
    {"version": 1, "name": "initial_schema", "applied": "2025-09-17T08:00:00Z"},
    {"version": 2, "name": "tool_versioning", "applied": "2025-09-18T10:00:00Z"}
  ]
}
```

#### Tool Versioning Migration (Migration #2)
```go
var toolVersioningMigration = Migration{
    Version: 2,
    Name: "tool_versioning",
    Apply: func(s *Storage) error {
        // 1. Create .history directory structure
        // 2. Scan all existing tools
        // 3. Group by name, sort by timestamp
        // 4. Assign version numbers
        // 5. Create metadata files
        // 6. Set up symlinks
        return migrateToolVersions(s)
    },
}
```

#### Daemon Startup Sequence
```go
func main() {
    // ... initialization ...

    // Run migrations before starting server
    migrationManager := NewMigrationManager(storage)
    if err := migrationManager.RunPendingMigrations(); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    // Continue with normal startup
    server.Start()
}
```

#### Benefits of Generalized Migration System
1. **Future-proof**: Any schema changes can be added as new migrations
2. **Automatic**: No user intervention required
3. **Traceable**: Full history of applied migrations
4. **Safe**: Migrations run before server starts, preventing corruption
5. **Reusable**: Framework works for any future structural changes

#### Future Migrations Could Include
- Migration 3: Add tool dependencies tracking
- Migration 4: Implement semantic versioning
- Migration 5: Add tool signing/verification
- Migration 6: Reorganize memory structure
- Migration 7: Add cross-tool version constraints

### Handling Duplicate Tools
When multiple tools with the same name exist (e.g., window-focus-tracker):
1. **Detection**: Group tools by base name during migration
2. **Version assignment**: Order by creation timestamp
   - Oldest becomes version 1
   - Newer versions increment sequentially
3. **Hash preservation**: Each version keeps its original hash
4. **Latest determination**: Most recent timestamp becomes the symlink target
5. **Conflict resolution**: If timestamps are identical, use hash ordering

Example migration:
```
window-focus-tracker (created: 2025-09-17 10:00) → window-focus-tracker:1
window-focus-tracker-enhanced (created: 2025-09-17 14:00) → window-focus-tracker:2
window-focus-tracker (created: 2025-09-18 09:00) → window-focus-tracker:3
/commands/window-focus-tracker → symlink to version 3
```

## Testing Plan

1. **Unit tests**:
   - Version parsing logic
   - Storage version management
   - VFS version resolution

2. **Integration tests**:
   - Create tool → update → verify versions
   - Reference specific versions in swim
   - Rollback and verify functionality

3. **Edge cases**:
   - Version overflow (>999)
   - Missing version files
   - Corrupted metadata
   - Concurrent version updates

## Security Considerations

1. **Version tampering**:
   - Versions are immutable once created
   - Hash verification for each version
   - Audit log of version changes

2. **Access control**:
   - Same permissions as current commands
   - No additional security surface

## Performance Impact

1. **Storage overhead**:
   - ~5-10% increase in storage for metadata
   - Negligible for command files (already small)

2. **Retrieval performance**:
   - Latest version: No change (direct symlink)
   - Specific version: One additional file read
   - Version listing: New operation, but fast

## Success Metrics

1. **Functional success**:
   - Can retrieve any version of any tool
   - No loss of existing tools
   - Version history preserved

2. **Performance success**:
   - Latest version retrieval <5ms overhead
   - Version listing <50ms for 100 versions
   - No increase in memory usage

3. **User success**:
   - Intuitive version syntax
   - Clear error messages
   - Helpful version discovery

## Risk Mitigation

1. **Risk**: Storage corruption
   - **Mitigation**: Keep original files, version files are copies

2. **Risk**: Performance degradation
   - **Mitigation**: Symlink for latest, lazy loading

3. **Risk**: User confusion
   - **Mitigation**: Clear documentation, helpful error messages

4. **Risk**: Breaking changes
   - **Mitigation**: Extensive backward compatibility

## Implementation Checklist

### Daemon Changes
- [ ] Update storage.go with version methods
- [ ] Modify tool_materializer.go for versioned output
- [ ] Enhance server.go VFS handlers
- [ ] Update reality_compiler.go for auto-versioning
- [ ] Modify swimming.go for versioned references
- [ ] Create version_manager.go for version operations
- [ ] Update protocol types in daemon
- [ ] Add version metadata to relation structure

### CLI Changes
- [ ] Create version parser in protocol module
- [ ] Update cat command for versions
- [ ] Enhance ls command display
- [ ] Create versions command
- [ ] Update help text for version syntax
- [ ] Add version info to shell completion
- [ ] Update error messages for version issues

### Testing & Documentation
- [ ] Write unit tests for version logic
- [ ] Create integration tests
- [ ] Update user documentation
- [ ] Create migration guide
- [ ] Add version examples to README

### Deployment
- [ ] Create migration script
- [ ] Update installation process
- [ ] Version the versioning system itself
- [ ] Monitor for issues post-deployment

## Conclusion

This versioning system provides a robust, intuitive way to manage tool versions in Port42 while maintaining backward compatibility and preparing for future enhancements. The VMS-style syntax (tool:2) is simple, memorable, and aligns with Port42's philosophy of making complex systems accessible.