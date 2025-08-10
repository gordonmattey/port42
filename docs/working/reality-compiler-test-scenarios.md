# Reality Compiler Test Scenarios

**Purpose**: Collection of interesting test cases to validate each step of the incremental reality compiler implementation.

---

## **Step 1: Basic Relation Storage** ✅

### **Basic Tool Creation**
```bash
# Simple tools
port42 declare tool git-haiku --transforms git-log,haiku
port42 declare tool csv-validator --transforms csv,validate
port42 declare tool json-prettier --transforms json,format

# Test the tools work
git-haiku --commits 5
echo "name,age\nBob,25" | csv-validator
echo '{"a":1,"b":2}' | json-prettier
```

### **Storage Verification**
```bash
# Check relation files exist
ls ~/.port42/relations/relation-tool-*

# Verify object store integration  
ls -la ~/.port42/commands/git-haiku  # Should be symlink

# Check materialization tracking
port42 declare get tool-git-haiku-*
```

---

## **Step 2: Auto-Spawning Rules**

### **Analysis Tool Spawning**
```bash
# Should create BOTH analyzer AND viewer
port42 declare tool log-analyzer --transforms logs,analysis
ls ~/.port42/commands/ | grep analyzer
# Expected: log-analyzer, view-log-analyzer

# Test both tools work
log-analyzer /var/log/system.log > analysis.json
view-log-analyzer analysis.json
```

### **Non-Analysis Tools (No Spawning)**
```bash
# Should create ONLY the main tool
port42 declare tool simple-parser --transforms parse,clean
ls ~/.port42/commands/ | grep parser
# Expected: simple-parser (no view-simple-parser)
```

### **Relationship Tracking**
```bash
# Check spawning relationships
port42 declare get <log-analyzer-relation-id>
# Should show: spawned_by metadata in view-log-analyzer

# Test relationship queries (future)
port42 relationships log-analyzer
# Expected: "spawned: view-log-analyzer"
```

---

## **Step 3: Virtual Views - Commands**

### **Multiple View Access**
```bash
# Same tool accessible multiple ways
port42 ls /commands/
port42 ls /by-date/today/
port42 ls /by-type/analysis/

# Test path resolution
port42 cat /commands/git-haiku  # Should work same as regular path
port42 info /commands/git-haiku  # Should show metadata
```

### **Dynamic Filtering**
```bash
# View analysis tools only
port42 ls /by-type/analysis/

# View tools by creation date
port42 ls /by-date/2024-01-15/

# View spawned tools
port42 ls /spawned/
```

---

## **Step 4: Relationship Tracking**

### **Spawning Relationships**
```bash
# Create tool that spawns others
port42 declare tool data-processor --transforms data,analysis
port42 relationships data-processor
# Expected: Shows view-data-processor spawned

# Reverse relationships  
port42 relationships view-data-processor
# Expected: Shows spawned_by data-processor
```

### **Chain Relationships**
```bash
# Tool that spawns tool that spawns tool (future complex rules)
port42 declare tool mega-analyzer --transforms analysis,report,dashboard
port42 relationships mega-analyzer --recursive
# Expected: Show full spawning chain
```

---

## **Step 5: Memory-Relation Bridge**

### **Session-Tool Connection**
```bash
# Create tool via possession, then declaratively 
port42 possess @ai-engineer "create a CSV processor"  # Creates via session
port42 declare tool csv-processor --transforms csv,process  # Creates declaratively

# Both should connect to memory
port42 ls /memory/sessions/<session-id>/tools/
# Expected: Both tools visible in memory view
```

### **Memory-Triggered Spawning**
```bash
# Tools remember their creation context
port42 declare tool context-aware --transforms analysis
port42 info /memory/tools/context-aware
# Expected: Show creation session, conversation context
```

---

## **Step 6: Tool Discovery & Similarity**

### **Similar Tool Detection**
```bash
# Create tools with overlapping transforms
port42 declare tool csv-cleaner --transforms csv,clean
port42 declare tool data-cleaner --transforms data,clean
port42 declare tool json-cleaner --transforms json,clean

# Should detect similarity
port42 discover similar csv-cleaner
# Expected: Suggest data-cleaner, json-cleaner
```

### **Transform Clustering**
```bash
# See all tools by transform type
port42 ls /by-transform/analysis/
port42 ls /by-transform/clean/

# Discover transform combinations
port42 transforms popular
# Expected: Show most common transform patterns
```

---

## **Step 7: Documentation Auto-Generation**

### **Complex Scenario Documentation**
```bash
# Create complex analysis pipeline
port42 declare tool log-ingester --transforms logs,ingest
port42 declare tool log-processor --transforms logs,analysis  
port42 declare tool log-reporter --transforms logs,report

# Should auto-generate pipeline docs
port42 ls /docs/pipelines/
port42 cat /docs/pipelines/log-analysis.md
# Expected: Auto-generated workflow documentation
```

### **Usage Example Generation**
```bash
# Tools should have auto-generated examples
port42 help log-analyzer --examples
# Expected: Generated usage examples based on transforms
```

---

## **Step 8: Rich Ecosystem Exploration**

### **Ecosystem Overview**
```bash
# Full system visualization
port42 reality map
# Expected: Visual representation of all relations + spawning

# Tool interdependencies
port42 reality graph --format=dot > ecosystem.dot
dot -Tpng ecosystem.dot -o ecosystem.png
```

### **Discovery Workflows**
```bash
# Start with problem, discover tools
port42 solve "I need to analyze CSV logs"
# Expected: Suggest existing tools or auto-create pipeline

# Explore tool evolution
port42 timeline git-haiku
# Expected: Show creation → spawning → usage history
```

---

## **Advanced Test Scenarios**

### **Stress Testing**
```bash
# Create many tools rapidly
for i in {1..20}; do
  port42 declare tool test-$i --transforms test,process
done

# Check system performance
time port42 ls /commands/  # Should be fast
time port42 reality map    # Should handle 100+ relations
```

### **Edge Cases**
```bash
# Circular dependencies (should be prevented)
port42 declare tool circular-a --spawns circular-b
port42 declare tool circular-b --spawns circular-a

# Rule conflicts
port42 declare tool conflict-test --transforms analysis,format
# Multiple rules might match - test priority handling

# Missing dependencies
port42 declare tool needs-python --requires python3.9
# Should gracefully handle missing system dependencies
```

### **Integration Testing**
```bash
# Mix declarative + imperative
port42 possess @ai-engineer "create a tool that works with csv-analyzer"
# Should integrate with existing declared tools

# Cross-step functionality
port42 declare tool full-test --transforms analysis
port42 ls /by-date/today/         # Step 3
port42 relationships full-test    # Step 4  
port42 discover similar full-test # Step 6
```

---

## **Success Criteria Checklist**

- [ ] **Step 1**: All basic declarations create working executables
- [ ] **Step 2**: Analysis tools auto-spawn viewers  
- [ ] **Step 3**: Same tools accessible via multiple virtual paths
- [ ] **Step 4**: Clear relationship graphs between all entities
- [ ] **Step 5**: Memory threads connect to declared tools
- [ ] **Step 6**: System suggests similar tools and optimizations
- [ ] **Step 7**: Complex scenarios auto-generate documentation
- [ ] **Step 8**: Rich ecosystem exploration and discovery tools

**Ultimate Test**: Create a complex data analysis workflow entirely through declarations, then explore and understand it through the virtual filesystem and relationship system. The magic should feel natural and discoverable.