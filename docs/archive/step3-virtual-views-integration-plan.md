# Step 3: Virtual Views Integration Plan
*Connect relations system to existing virtual filesystem*

## **Component Architecture**

### **1. Virtual Path Resolver (Extend Existing)**
```go
// daemon/virtual_resolver.go - Extend existing resolver
type VirtualResolver struct {
    relationStore RelationStore  // NEW - inject relations
    // existing fields...
}

// Add relations-aware resolution
func (vr *VirtualResolver) resolveWithRelations(path string) ([]VirtualNode, error) {
    switch {
    case strings.HasPrefix(path, "/commands"):
        return vr.resolveCommandsWithRelations(path)  // Enhanced
    case strings.HasPrefix(path, "/relations"):       // NEW
        return vr.resolveRelationsView(path)
    case strings.HasPrefix(path, "/spawned-by"):      // NEW  
        return vr.resolveSpawnedByView(path)
    // existing cases...
    }
}
```

### **2. Relations View Provider (New Component)**
```go
// daemon/relations_view.go - New focused component
type RelationsViewProvider struct {
    store RelationStore
}

func (rvp *RelationsViewProvider) ListRelations(filters map[string]string) ([]VirtualNode, error)
func (rvp *RelationsViewProvider) GetRelation(id string) (*VirtualNode, error)  
func (rvp *RelationsViewProvider) GetSpawnedBy(parentID string) ([]VirtualNode, error)
func (rvp *RelationsViewProvider) GetRelationsByTransform(transform string) ([]VirtualNode, error)
```

### **3. Enhanced Commands View (Extend Existing)**
```go
// daemon/commands_view.go - Enhance existing
type CommandsViewProvider struct {
    relationStore RelationStore  // NEW injection
    // existing fields...
}

// Enhanced to show relation metadata
func (cvp *CommandsViewProvider) ListCommandsWithMetadata(path string) ([]VirtualNode, error) {
    // Get existing command list
    // Enrich with relation metadata (parent, spawned_by, transforms)
}
```

## **New Virtual Paths to Support**

### **Core Relations Paths**
- `/relations/` - All relations (tools, artifacts, etc.)
- `/relations/tools/` - Just tool relations  
- `/relations/artifacts/` - Just artifact relations
- `/relations/{id}` - Specific relation details

### **Relationship Navigation Paths**
- `/spawned-by/{tool-name}` - Everything spawned by this tool
- `/parents/{tool-name}` - Parent chain for this tool  
- `/transforms/{transform}/` - All tools with this transform
- `/auto-spawned/` - All auto-spawned entities

### **Enhanced Existing Paths**
- `/commands/` - Now shows relation metadata (spawned_by, transforms)
- `/by-date/` - Now includes relations, not just commands
- `/by-type/` - Split into tools/artifacts/auto-spawned

## **Integration Points**

### **1. Daemon Protocol Extensions**
```go
// daemon/protocol.go - Extend existing handlers
func handleListPath(req ListPathRequest) ListPathResponse {
    // Inject relationStore into existing resolver
    resolver := NewVirtualResolver(existingStore, relationStore)
    // existing logic...
}
```

### **2. CLI Integration (No Changes Needed)**
- Existing `port42 ls`, `port42 cat`, `port42 info` work automatically
- New paths available immediately through existing commands

### **3. Storage Integration**
```go
// daemon/reality_compiler.go - Hook into existing flow
func (rc *RealityCompiler) DeclareRelation(relation Relation) (*MaterializedEntity, error) {
    // Existing materialization...
    
    // NEW: Update virtual filesystem cache
    rc.virtualResolver.InvalidateCache(relation.Type)
    
    return entity, nil
}
```

## **Implementation Phases**

### **Phase A: Basic Relations Views** 
- Add `/relations/` paths
- Inject RelationStore into existing VirtualResolver
- Show basic relation listing

### **Phase B: Relationship Navigation**
- Add `/spawned-by/`, `/transforms/` paths  
- Implement relationship traversal
- Show parent-child chains

### **Phase C: Enhanced Existing Views**
- Enrich `/commands/` with relation metadata
- Update `/by-date/` to include relations
- Add relation info to `port42 info` command

### **Phase D: Advanced Discovery**
- Add semantic similarity navigation
- Cross-reference memory sessions with relations
- Advanced filtering and search

## **Clean Encapsulation Strategy**

### **Separation of Concerns**
- **RelationsViewProvider**: Pure relations → virtual nodes conversion
- **VirtualResolver**: Path routing and caching (existing + enhanced)  
- **RealityCompiler**: Orchestration and change notifications
- **Storage layers**: Isolated, no direct virtual filesystem coupling

### **Dependency Injection Pattern**
```go
// Clean initialization in main.go
relationStore := NewFileRelationStore()
virtualResolver := NewVirtualResolver(commandStore, relationStore) // Enhanced constructor
realityCompiler := NewRealityCompiler(relationStore, virtualResolver)
```

### **Interface-Based Design**
```go
type RelationViewProvider interface {
    ListRelations(filters map[string]string) ([]VirtualNode, error)
    GetRelationsByParent(parentID string) ([]VirtualNode, error)
}

type VirtualPathResolver interface {  // Existing interface extended
    Resolve(path string) ([]VirtualNode, error)
    InvalidateCache(entityType string) error  // NEW
}
```

## **Success Criteria**
- **Phase A**: `port42 ls /relations/` shows all declared relations
- **Phase B**: `port42 ls /spawned-by/test-analyzer` shows `view-test-analyzer`  
- **Phase C**: `port42 ls /commands/` shows relation metadata alongside tools
- **Phase D**: `port42 ls /transforms/analysis` shows all analysis tools

## **Key Benefits**
- **Zero CLI changes** - existing commands work with new paths
- **Clean separation** - relations logic isolated from virtual filesystem 
- **Incremental** - each phase adds value independently
- **Backward compatible** - existing virtual paths unchanged

## **Current State Analysis**
**Existing Virtual Filesystem Paths:**
```
/                   # Root
├── commands/       # Materialized tools
├── memory/         # Conversation sessions  
├── artifacts/      # Stored artifacts
├── by-date/        # Temporal organization
├── by-agent/       # Agent-based organization
└── by-type/        # Type-based organization
```

**What We're Adding:**
```
/                   # Root (existing)
├── relations/      # NEW - Relation definitions
│   ├── tools/      # Just tool relations
│   └── artifacts/  # Just artifact relations
├── spawned-by/     # NEW - Relationship navigation
├── parents/        # NEW - Parent chain navigation
├── transforms/     # NEW - Transform-based organization
└── auto-spawned/   # NEW - Auto-spawned entities
```

This plan leverages your existing virtual filesystem architecture while cleanly integrating the relations system for rich navigation and discovery.