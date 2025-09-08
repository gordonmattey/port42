# Claude Code Test Scenarios for Port42 Integration

## Purpose
Test scenarios to verify that Claude Code correctly understands and implements the single-purpose possess architecture and orchestration requirements defined in P42CLAUDE.md.

## Prerequisites
**Note:** These tests are designed to work with a fresh Port42 installation. Some tests reference tools or artifacts that may not exist yet. The expected behavior includes:
- Searching for existing tools/artifacts first
- Creating new tools if none exist
- Using `search:` references when specific artifacts aren't available
- Gracefully handling missing references

### Optional: Generate Test Prerequisites
To create sample tools and artifacts for more comprehensive testing:

```bash
# 1. Create a basic JSON validator tool (for test #4)
port42 possess @ai-engineer "create a tool called json-validator that validates JSON syntax and structure" --transforms "stdin,json,validate,error,python"

# 2. Create a log analyzer tool (for tests #3, #6)
port42 possess @ai-engineer "create a tool called log-analyzer that parses and analyzes log files" --transforms "file,text,parse,analyze,error,bash"

# 3. Create a sample marketing spec document (for test #5)
port42 possess @ai-engineer "generate a marketing metrics specification document" --action generate

# 4. Create a git-related tool to test relationships (for test #15)
port42 possess @ai-engineer "create a tool called git-haiku that generates haikus from git commits" --transforms "git,text,poetry,bash"

# 5. Create a deployment tool (for test #9)
port42 possess @ai-engineer "create a simple deploy-tool that shows deployment status" --transforms "bash,status,deploy"
```

After running these, the tests will have richer discovery results and can test reference handling more thoroughly.

## Test Scenarios

### 1. Basic Tool Discovery & Creation
```bash
# Test that Claude Code does discovery BEFORE creation
"Create a tool to format JSON files"

# Expected Claude Code behavior:
# 1. port42 search "json format"
# 2. port42 ls /tools/by-transform/json/
# 3. port42 ls /artifacts/document/ | grep -i "json"
# 4. port42 info /commands/[existing-json-tool]
# 5. port42 possess @ai-engineer "create json formatter" --ref p42:/commands/[existing-tool]
```

### 2. Agent Selection Test
```bash
# Test that Claude Code chooses the right agent
"Analyze the performance patterns in my server logs"

# Should use @ai-analyst, not @ai-engineer:
# port42 possess @ai-analyst "analyze performance patterns" --ref file:/path/to/logs
```

### 3. Single-Purpose Principle Test
```bash
# Test that Claude Code doesn't expect multi-step from possess
"Search for existing log analyzers and create an improved version"

# Should break into steps:
# 1. port42 search "log analyzer"
# 2. port42 ls /tools/by-transform/log/
# 3a. If log-analyzer exists (from prerequisites):
#     port42 possess @ai-engineer "create improved log analyzer" --ref p42:/commands/log-analyzer
# 3b. If no existing analyzer:
#     port42 possess @ai-engineer "create improved log analyzer" --ref search:"log analysis patterns"
# NOT: port42 possess @ai-engineer "search and create improved version"
```

### 4. Reference Usage Test
```bash
# Test that Claude Code provides multiple references
"Create a tool to validate API responses against OpenAPI specs"

# Should gather multiple references (if they exist):
# port42 search "json validator api openapi"
# port42 ls /tools/by-transform/validate/
# If tools exist:
#   port42 possess @ai-engineer "create API validator" \
#     --ref p42:/commands/json-validator \
#     --ref search:"validation patterns"
# If no tools exist:
#   port42 possess @ai-engineer "create API validator" \
#     --ref search:"API validation OpenAPI"
```

### 5. Document/Artifact Discovery Test
```bash
# Test that Claude Code searches for specs and documentation
"Build a tool for processing marketing metrics"

# Should search artifacts:
# 1. port42 search "marketing metrics patterns"
# 2. port42 ls /artifacts/
# 3. port42 ls /tools/by-transform/metrics/ (if exists)
# 4. If artifacts found, include as --ref
# 5. If no artifacts, use search references:
#    port42 possess @ai-engineer "create marketing metrics processor" \
#      --ref search:"marketing metrics"
```

### 6. Tool Enhancement Test
```bash
# Test updating existing tools with proper reference
"Add error handling to the log-analyzer tool"

# Should check if tool exists first:
# 1. port42 info /commands/log-analyzer
# 2a. If exists (from prerequisites):
#     port42 possess @ai-engineer "add error handling to log-analyzer" --ref p42:/commands/log-analyzer
# 2b. If doesn't exist:
#     port42 possess @ai-engineer "create log-analyzer with error handling" --transforms "file,text,parse,analyze,error,bash"
# NOT: port42 cat /commands/log-analyzer followed by possess without reference
```

### 7. Memory/Session Continuation Test
```bash
# Test session continuation with single-purpose principle
"Continue our previous discussion about the testing framework"

# Should handle memory but still do ONE action:
# port42 memory
# port42 possess @ai-engineer cli-session-123 "implement the testing framework we discussed"
```

### 8. Orchestration Pattern Test
```bash
# Test complex workflow orchestration
"Analyze our codebase structure and create a documentation generator based on the patterns found"

# Should orchestrate multiple steps:
# 1. port42 possess @ai-analyst "analyze codebase structure" --ref file:./src
# 2. Save or note the analysis output
# 3. port42 possess @ai-engineer "create documentation generator based on codebase patterns" \
#      --ref search:"documentation generator"
# Note: Since possess doesn't persist artifacts between calls, 
# the analysis insights would need to be included in the prompt
```

### 9. Wrong Agent Correction Test
```bash
# Test if Claude Code avoids using wrong agents
"Create a haiku about our deployment process"

# Should use @ai-muse, not @ai-engineer:
# port42 possess @ai-muse "create deployment haiku" --ref p42:/commands/deploy-tool
```

### 10. No Context Failure Test
```bash
# Test that Claude Code always provides context
"Make it better"

# Should ask for clarification or search for context:
# "What would you like me to improve? Please specify the tool or provide more context."
# OR search for recent tools/memory to understand context
```

### 11. Tool Spawning Discovery Test
```bash
# Test discovery of spawned relationships
"Find tools that were created by the log-analyzer"

# Should explore relationships:
# port42 ls /tools/spawned-by/log-analyzer/
# port42 ls /tools/ancestry/
```

### 12. Transform-Based Discovery Test
```bash
# Test capability-based tool discovery
"I need something to parse CSV files"

# Should search by transform:
# 1. port42 search "csv parse"
# 2. port42 ls /tools/by-transform/csv/
# 3. port42 ls /tools/by-transform/parse/
```

### 13. Common Mistakes Test
Test each mistake pattern from P42CLAUDE.md:

```bash
# Mistake 1: Skipping discovery
"Create a json formatter" 
# Should search first, not go straight to possess

# Mistake 2: Expecting multi-step
"Find similar tools and create an improved version"
# Should break into separate discovery and creation steps

# Mistake 3: No references
"Improve the performance"
# Should ask for clarification or find context

# Mistake 4: Too much in one call
"Analyze the logs, create a tool, and document it"
# Should break into three separate possess calls
```

### 14. @ai-analyst Agent Test
```bash
# Test the new @ai-analyst agent specifically
"Analyze the usage patterns of our Port42 tools"

# Should correctly use:
# port42 possess @ai-analyst "analyze usage patterns" --ref search:"tool usage"
```

### 15. VFS Navigation Test
```bash
# Test understanding of VFS structure
"Show me the relationship between the git-haiku tool and its parent tools"

# Should navigate:
# 1. port42 ls /tools/git-haiku/
# 2. port42 ls /tools/git-haiku/parents/
# 3. port42 info /commands/git-haiku
```

## How to Run These Tests

1. **Copy each scenario as a prompt to Claude Code**
2. **Observe if Claude Code:**
   - Does discovery before creation
   - Uses correct agents
   - Provides proper references
   - Breaks complex tasks into steps
   - Searches for documents/specs in VFS

3. **Success Criteria:**
   - ✅ Always searches before creating
   - ✅ Uses --ref with context
   - ✅ Selects appropriate agent
   - ✅ Breaks multi-step requests
   - ✅ Discovers artifacts/documents
   - ✅ One possess action per call

## Expected Failures (What NOT to do)

### ❌ Direct Creation Without Discovery
```bash
# BAD: Goes straight to creation
port42 possess @ai-engineer "create log analyzer"

# GOOD: Discovers first
port42 search "log analyzer"
port42 ls /tools/by-transform/log/
port42 possess @ai-engineer "create log analyzer" --ref p42:/commands/existing-analyzer
```

### ❌ Multi-Step in Single Possess
```bash
# BAD: Expects possess to do multiple things
port42 possess @ai-engineer "search for tools then create improved version"

# GOOD: Orchestrates steps
port42 search "relevant tools"
port42 possess @ai-engineer "create improved version" --ref p42:/commands/found-tool
```

### ❌ Missing References
```bash
# BAD: No context provided
port42 possess @ai-engineer "make it better"

# GOOD: Provides specific references
port42 possess @ai-engineer "improve performance" --ref p42:/commands/specific-tool
```

## Test Results Tracking

| Test # | Scenario | Expected Behavior | Pass/Fail | Notes |
|--------|----------|-------------------|-----------|-------|
| 1 | Basic Discovery | Search → Explore → Create | | |
| 2 | Agent Selection | Uses @ai-analyst for analysis | | |
| 3 | Single-Purpose | Breaks into steps | | |
| 4 | Multiple References | Provides 3+ refs | | |
| 5 | Document Discovery | Searches /artifacts/document/ | | |
| 6 | Tool Enhancement | Uses --ref to existing | | |
| 7 | Memory Continuation | Handles session correctly | | |
| 8 | Complex Orchestration | Multiple possess calls | | |
| 9 | Agent Correction | Uses appropriate agent | | |
| 10 | Context Handling | Asks for clarification | | |
| 11 | Spawned Discovery | Explores relationships | | |
| 12 | Transform Discovery | Uses /tools/by-transform/ | | |
| 13 | Common Mistakes | Avoids all 4 patterns | | |
| 14 | @ai-analyst Test | Correctly uses new agent | | |
| 15 | VFS Navigation | Understands structure | | |

## Notes

- These tests verify the changes made to P42CLAUDE.md on 2025-01-05
- Focus is on single-purpose possess architecture
- Claude Code must orchestrate discovery → decision → action
- Each possess invocation performs exactly ONE action