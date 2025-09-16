# Advanced Rules Test Scenarios

*Comprehensive testing guide for Step 7 - Advanced Rules functionality*

## 🧪 **Test All Advanced Rules**

### **Test 1: Documentation Rule** (3+ transforms)
```bash
# Create a complex tool with 4+ transforms
port42 declare tool complex-processor --transforms "data,analysis,transform,export"

# Expected: Auto-spawns complex-processor-docs
# Verify: port42 ls /tools | grep complex-processor
```

### **Test 2: Git Tools Rule**
```bash
# Create a git-related tool
port42 declare tool git-branch-manager --transforms "git,branch,management"

# Expected: Auto-spawns git-status-enhanced
# Verify: port42 ls /tools | grep git
```

### **Test 3: Test Suite Rule**
```bash
# Create a test-related tool
port42 declare tool pytest-runner --transforms "test,python,automation"

# Expected: Auto-spawns test-runner-enhanced
# Verify: port42 ls /tools | grep test
```

### **Test 4: Viewer Rule** (analysis tools)
```bash
# Create an analysis tool
port42 declare tool log-analyzer --transforms "logs,analysis,parsing"

# Expected: Auto-spawns log-analyzer-viewer
# Verify: port42 ls /tools | grep analyzer
```

### **Test 5: Documentation Emergence Rule**
```bash
# Test 1: Wiki-based documentation tool
port42 declare tool wiki-manager --transforms "wiki,edit,content"

# Test 2: README focused tool  
port42 declare tool readme-generator --transforms "readme,markdown,create"

# Test 3: Documentation in name
port42 declare tool doc-builder --transforms "build,deploy"

# Expected: Auto-spawns complete documentation infrastructure
# - doc-template-generator (creates templates)
# - doc-validator (quality checking)  
# - doc-site-builder (static site generation)
# Verify: port42 ls /tools | grep -E "(doc-template|doc-validator|doc-site)"
```

## 📊 **Verification Commands**

### **Check Active Rules**
```bash
# View all active rules
port42 watch rules

# Expected output:
# ⚡ Auto-spawn viewer for analysis tools: Status: enabled
# ⚡ Auto-spawn documentation for complex tools: Status: enabled  
# ⚡ Auto-spawn git tools: Status: enabled
# ⚡ Auto-spawn test suite tools: Status: enabled
# ⚡ Documentation Emergence Intelligence: Status: enabled
```

### **View Rule-Spawned Tools**
```bash
# See all auto-spawned tools organized by parent
port42 ls /tools/spawned-by/

# Expected: Directories showing which tools spawned what
# Example: git-quick-commit/, pytest-runner/, etc.
```

### **Inspect Auto-Generated Tools**
```bash
# Check specific auto-spawned tool
port42 cat /tools/test-runner-enhanced

# Expected: Generated test automation tool
port42 cat /tools/git-status-enhanced

# Expected: Enhanced git status tool
```

### **Count Tool Growth**
```bash
# Before creating tools
port42 ls /tools | wc -l

# Create a tool that triggers rules
port42 declare tool integration-tester --transforms "test,integration,validation"

# After - should show +2 tools (original + auto-spawned)
port42 ls /tools | wc -l
```

## 🔍 **Quick Test Sequence**

### **Complete Rule Validation**
```bash
# 1. Check starting state
echo "=== Initial State ==="
port42 watch rules
echo "Tool count: $(port42 ls /tools | wc -l)"

# 2. Test Documentation Rule (3+ transforms)
echo "=== Testing Documentation Rule ==="
port42 declare tool data-pipeline --transforms "data,transform,validate,export"
port42 ls /tools | grep -E "(data-pipeline|docs)"

# 3. Test Git Tools Rule
echo "=== Testing Git Tools Rule ==="
port42 declare tool git-flow-helper --transforms "git,workflow,automation"
port42 ls /tools | grep git-status-enhanced

# 4. Test Test Suite Rule  
echo "=== Testing Test Suite Rule ==="
port42 declare tool unit-validator --transforms "test,unit,validation"
port42 ls /tools | grep test-runner-enhanced

# 5. Test Viewer Rule
echo "=== Testing Viewer Rule ==="
port42 declare tool metrics-analyzer --transforms "metrics,analysis,reporting"
port42 ls /tools | grep viewer

# 6. Final verification
echo "=== Final State ==="
echo "Rule-spawned tools:"
port42 ls /tools/spawned-by/
echo "Total tools: $(port42 ls /tools | wc -l)"
```

## 🎯 **Expected Results**

### **Rule Triggering Patterns**

| Tool Pattern | Triggers Rule | Auto-Spawns |
|-------------|---------------|-------------|
| **3+ transforms** | Documentation Rule | `{tool-name}-docs` |
| **git/commit/branch** | Git Tools Rule | `git-status-enhanced` |
| **test/spec/unit** | Test Suite Rule | `test-runner-enhanced` |
| **analysis transform** | Viewer Rule | `{tool-name}-viewer` |
| **docs/wiki/readme/manual** | Documentation Emergence | Infrastructure ecosystem |

### **Rule Interaction Examples**

```bash
# Complex git tool triggers BOTH rules
port42 declare tool git-test-automation --transforms "git,test,automation,validation"

# Expected auto-spawns:
# 1. git-test-automation-docs (Documentation Rule - 4 transforms)
# 2. git-status-enhanced (Git Tools Rule - "git" in name)
# 3. test-runner-enhanced (Test Suite Rule - "test" in transforms)
```

### **Anti-Patterns (Should NOT trigger)**

```bash
# Auto-spawned tools don't trigger more rules
# These should exist but not spawn additional tools:
port42 cat /tools/git-status-enhanced      # Has auto_spawned: true
port42 cat /tools/test-runner-enhanced     # Has auto_spawned: true
port42 cat /tools/data-pipeline-docs       # Has auto_spawned: true
```

## 🔧 **Troubleshooting**

### **Common Issues**

1. **Rule not triggering:**
   ```bash
   # Check if rule is active
   port42 watch rules
   
   # Check daemon logs for rule processing
   tail -f ~/.port42/daemon.log | grep -E "(Rule|spawn)"
   ```

2. **Tool not auto-spawning:**
   ```bash
   # Verify pattern matching - these should trigger:
   port42 declare tool my-git-tool --transforms "git,helper"     # Git Rule
   port42 declare tool my-test-suite --transforms "test,runner"  # Test Rule
   port42 declare tool complex-tool --transforms "a,b,c,d"      # Docs Rule (4 transforms)
   ```

3. **Multiple rules triggering:**
   ```bash
   # This should trigger multiple rules:
   port42 declare tool git-test-analyzer --transforms "git,test,analysis,reporting"
   
   # Expected: 3 auto-spawned tools:
   # - git-test-analyzer-docs (4 transforms)
   # - git-status-enhanced (git pattern)  
   # - test-runner-enhanced (test pattern)
   ```

## 📈 **Success Metrics**

- ✅ **All 5 rules show as active** in `port42 watch rules`
- ✅ **Pattern detection works** for name and transform patterns
- ✅ **Auto-spawning succeeds** without errors in daemon logs
- ✅ **Recursion prevention** works (auto-spawned tools don't spawn more)
- ✅ **Multiple rule triggering** works for tools matching multiple patterns
- ✅ **VFS integration** shows auto-spawned tools in `/tools/spawned-by/`
- ✅ **Documentation emergence** creates complete infrastructure ecosystem

## 💡 **Advanced Test Cases**

### **Edge Case Testing**

```bash
# Test mixed case sensitivity
port42 declare tool GIT-Manager --transforms "GIT,WORKFLOW"

# Test partial matches
port42 declare tool my-pytest-config --transforms "testing,configuration"

# Test multiple git tools (should reuse git-status-enhanced)
port42 declare tool git-commit-helper --transforms "git,commit"
port42 declare tool git-merge-tool --transforms "git,merge"

# Verify only ONE git-status-enhanced exists
port42 ls /tools | grep git-status-enhanced | wc -l  # Should be 1

# Test Documentation Emergence with different patterns
port42 declare tool WIKI-Editor --transforms "WIKI,CONTENT"  # Case insensitive
port42 declare tool user-manual-creator --transforms "help,assistance"  # Name pattern
port42 declare tool note-taker --description "documentation and notes system"  # Description pattern

# Verify documentation infrastructure spawned
port42 ls /tools | grep -E "(doc-template|doc-validator|doc-site)" | wc -l  # Should be 3 per trigger
```

This comprehensive test suite validates that **Step 7 - Advanced Rules** is working correctly and demonstrates the intelligent auto-spawning behavior in action!