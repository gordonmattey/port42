# Stigmergic Crafting Implementation Plan

*Building terminal intelligence through incremental tool crafting*

## 🧩 **Core Concept: Infinite Crafting for Terminals**

Like infinite crafting games, users build complex intelligent behaviors by creating simple tools that combine and trigger more sophisticated auto-spawned tools.

## 🎯 **Demo Sequence: Directory Intelligence Crafting**

### **Step 1: Create Basic Building Blocks**

```bash
# Create directory logging tool
$ port42 declare tool logcd --transforms "directory,logging,tracking"

# Create enhanced cd command
$ port42 declare tool smart-cd --transforms "directory,enhancement,wrapper" --ref tool:logcd
```

### **Step 2: Watch System Detect Pattern**

```bash
# Check what tools exist
$ port42 ls /tools

# Watch for auto-spawned tools (rule should trigger)
$ port42 watch rules
```

### **Step 3: Expected Auto-spawned Tools**

The **Directory Intelligence Crafting Rule** should detect the pattern and spawn:
- `cd-history` - Shows directory navigation history
- `frequent-dirs` - Shows most visited directories
- `smart-bookmarks` - Quick directory bookmarking system

### **Step 4: Next Level Crafting**

```bash
# User creates next level tool
$ port42 declare tool project-jumper --transforms "directory,project,navigation" --ref tool:frequent-dirs

# System auto-spawns advanced tools:
# - workspace-detector (recognizes project types)
# - dev-session-manager (manages development environments per directory)
```

## 🤖 **Implementation Plan**

### **Phase 1: Directory Intelligence Crafting Rule**

```go
func directoryIntelligenceCraftingRule() Rule {
    return Rule{
        ID:          "craft-directory-intelligence",
        Name:        "Directory Intelligence Crafting",
        Description: "When user builds directory tracking tools, complete the ecosystem",
        Enabled:     true,
        
        Condition: func(relation Relation) bool {
            // Only process Tool relations
            if relation.Type != "Tool" { return false }
            
            // Skip auto-spawned tools
            if isAutoSpawned(relation) { return false }
            
            // Check if this is a directory-related tool
            if !isDirectoryRelatedTool(relation) { return false }
            
            // Find all existing directory tools
            directoryTools := findDirectoryTools(compiler)
            
            // Trigger when we have 2+ directory tools but missing ecosystem tools
            return len(directoryTools) >= 2 && !hasDirectoryEcosystem(compiler)
        },
        
        Action: func(relation Relation, compiler *RealityCompiler) error {
            return spawnDirectoryEcosystem(compiler)
        },
    }
}
```

### **Phase 2: Helper Functions**

```go
func isDirectoryRelatedTool(relation Relation) bool {
    name := getRelationName(relation)
    transforms := getTransforms(relation)
    
    // Check name patterns
    directoryNames := []string{"cd", "dir", "log", "track", "nav", "jump"}
    for _, pattern := range directoryNames {
        if strings.Contains(strings.ToLower(name), pattern) {
            return true
        }
    }
    
    // Check transform patterns
    directoryTransforms := []string{"directory", "navigation", "tracking", "logging"}
    for _, transform := range transforms {
        for _, pattern := range directoryTransforms {
            if strings.Contains(strings.ToLower(transform), pattern) {
                return true
            }
        }
    }
    
    return false
}

func findDirectoryTools(compiler *RealityCompiler) []Relation {
    allTools, _ := compiler.relationStore.LoadByType("Tool")
    var directoryTools []Relation
    
    for _, tool := range allTools {
        if isDirectoryRelatedTool(tool) {
            directoryTools = append(directoryTools, tool)
        }
    }
    
    return directoryTools
}

func hasDirectoryEcosystem(compiler *RealityCompiler) bool {
    // Check if ecosystem tools already exist
    ecosystemTools := []string{"cd-history", "frequent-dirs", "smart-bookmarks"}
    
    for _, toolName := range ecosystemTools {
        if _, err := compiler.relationStore.LoadByProperty("name", toolName); err == nil {
            return true // At least one ecosystem tool exists
        }
    }
    
    return false
}

func spawnDirectoryEcosystem(compiler *RealityCompiler) error {
    ecosystemTools := []struct {
        name       string
        transforms []string
        description string
    }{
        {
            name:        "cd-history",
            transforms:  []string{"directory", "history", "navigation"},
            description: "Shows directory navigation history from logcd data",
        },
        {
            name:        "frequent-dirs", 
            transforms:  []string{"directory", "analytics", "frequency"},
            description: "Analyzes and shows most frequently visited directories",
        },
        {
            name:        "smart-bookmarks",
            transforms:  []string{"directory", "bookmarks", "quick-access"},
            description: "Quick bookmarking system for important directories",
        },
    }
    
    for _, tool := range ecosystemTools {
        relation := Relation{
            ID:   generateRelationID("Tool", tool.name),
            Type: "Tool",
            Properties: map[string]interface{}{
                "name":        tool.name,
                "transforms":  tool.transforms,
                "description": tool.description,
                "auto_spawned": true,
                "crafted_by":  "directory-intelligence-crafting",
            },
            CreatedAt: time.Now(),
        }
        
        log.Printf("🧩 Crafting ecosystem tool: %s", tool.name)
        
        _, err := compiler.DeclareRelation(relation)
        if err != nil {
            log.Printf("❌ Failed to craft %s: %v", tool.name, err)
            return fmt.Errorf("failed to craft %s: %w", tool.name, err)
        }
        
        log.Printf("✅ Successfully crafted: %s", tool.name)
    }
    
    log.Printf("🌟 Directory intelligence ecosystem crafted successfully!")
    return nil
}
```

## 📋 **Testing Steps**

### **Manual Test Sequence**

1. **Start with clean system:**
   ```bash
   $ port42 ls /tools  # Should be minimal
   $ port42 watch rules  # Start monitoring
   ```

2. **Create first building block:**
   ```bash
   $ port42 declare tool logcd --transforms "directory,logging,tracking"
   $ port42 ls /tools  # Should show logcd
   ```

3. **Create second building block:**
   ```bash
   $ port42 declare tool smart-cd --transforms "directory,enhancement,wrapper" --ref tool:logcd
   ```

4. **Watch for auto-spawning:**
   ```bash
   $ port42 ls /tools  # Should now show ecosystem tools
   $ port42 cat /tools/cd-history  # Check auto-spawned content
   ```

5. **Verify rule triggered:**
   ```bash
   $ port42 watch rules  # Should show crafting rule activated
   ```

## 🎮 **Extended Crafting Chains**

### **Git Workflow Crafting**
```
git-status → git-add → git-commit 
    ↓ (triggers Git Workflow Crafting Rule)
Auto-spawns: git-flow, git-sync, branch-manager
```

### **System Monitoring Crafting** 
```
ps-monitor → cpu-check → memory-usage
    ↓ (triggers System Intelligence Crafting Rule)  
Auto-spawns: resource-manager, process-killer, system-optimizer
```

### **File Management Crafting**
```
file-watcher → backup-tool → sync-checker
    ↓ (triggers File Intelligence Crafting Rule)
Auto-spawns: smart-backup, file-organizer, duplicate-finder
```

## 🌟 **Success Metrics**

- ✅ Users can build complex intelligence through simple tool creation
- ✅ System detects crafting patterns and completes ecosystems
- ✅ Each individual tool is useful standalone
- ✅ Auto-spawned tools enhance and complete user's vision
- ✅ Demonstrates true stigmergic intelligence emergence

## 🚀 **Future Expansions**

1. **Cross-Domain Crafting**: Tools from different domains combine to create meta-tools
2. **User Learning**: System learns individual user's crafting patterns
3. **Collaborative Crafting**: Users share crafting patterns with community
4. **Advanced Ecosystems**: Multi-level crafting trees with complex dependencies

This creates an **"infinite crafting terminal"** where every user builds their own unique intelligent environment through natural tool creation!