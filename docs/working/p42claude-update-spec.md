# P42CLAUDE.md Update Specification

## Purpose
Update P42CLAUDE.md to reflect the single-purpose possess architecture implemented on 2025-01-05, ensuring Claude Code understands it must orchestrate discovery workflows before calling possess.

## Background
- Possess now performs exactly ONE action per invocation (CREATE, ANALYZE, or GENERATE)
- It no longer does internal discovery or exploration
- External orchestrators (like Claude Code) must handle the complete workflow
- This change broke existing patterns where Claude Code expected possess to handle discovery

## Proposed Changes to P42CLAUDE.md

### 1. Add Possess Architecture to XML Structure

**Location**: After the opening "Port42 Integration" section, as part of XML structure  
**Purpose**: Define possess architecture as structural system knowledge

```xml
<possess_architecture>
<single_purpose_principle>
Each possess invocation performs exactly ONE action:
- CREATE: Build one tool/artifact  
- ANALYZE: Examine data and provide insights
- GENERATE: Produce non-executable content

Possess will NOT:
- Search for existing tools automatically
- Explore relationships for you
- Chain multiple operations
</single_purpose_principle>

<orchestration_responsibility>
You MUST orchestrate the complete workflow using Port42's native commands BEFORE calling possess.
Possess is a pure function that requires complete context via references.
</orchestration_responsibility>
</possess_architecture>
```

### 2. Replace Tool Discovery Workflow Section

**Location**: Replace entire `<tool_discovery_workflow>` XML block  
**Purpose**: Make discovery steps mandatory and explicit

```xml
<tool_discovery_workflow>
<critical>ALWAYS complete ALL discovery steps BEFORE calling possess</critical>

<step priority="1" required="true">
<action>port42 search "relevant keywords"</action>
<purpose>Find existing tools that solve the user's need</purpose>
<when>EVERY TIME before creating any tool</when>
</step>

<step priority="2" required="true">
<action>port42 ls /tools/</action>
<purpose>Explore the multi-dimensional tool ecosystem</purpose>
<note>Check /tools/by-transform/, /tools/spawned-by/, /tools/ancestry/</note>
</step>

<step priority="3" required="true">
<action>port42 ls /commands/ | port42 info /commands/similar-tool</action>
<purpose>Understand existing implementations</purpose>
<note>Gather context for references</note>
</step>

<step priority="4" required="true">
<action>port42 ls /artifacts/document/ | port42 search "specifications patterns architecture"</action>
<purpose>Find relevant documentation, specs, and architectural patterns</purpose>
<note>Discover domain knowledge and design patterns to reference</note>
</step>

<step priority="5" required="false">
<action>port42 cat /commands/similar-tool | port42 cat /artifacts/document/relevant-spec</action>
<purpose>View source code and documentation if needed</purpose>
</step>

<step priority="6" required="true">
<decision>Based on discovery, decide whether to:
- Use existing tool as-is
- Enhance existing tool with --ref
- Create new tool with --ref context
</decision>
</step>

<step priority="7">
<action>port42 possess @ai-engineer "request" --ref [context]</action>
<purpose>Execute SINGLE creation/analysis with full context</purpose>
<critical>Possess will NOT search or discover - you must provide complete context</critical>
<references>Always include:
- --ref p42:/commands/similar-tool (for patterns)
- --ref p42:/artifacts/document/spec-name (for domain knowledge)
- --ref search:"relevant patterns" (for contextual memories)
- --ref file:/path/to/requirements (for specific needs)
</references>
</step>
</tool_discovery_workflow>
```

### 3. Add Orchestration Requirements to Tool Creation Guidance

**Location**: Inside `<tool_creation_guidance>` section, add new sub-section  
**Purpose**: Enforce discovery before creation pattern

```xml
<orchestration_requirements>
BEFORE calling possess to create ANY tool:

1. MANDATORY DISCOVERY (you must do this):
   - port42 search "tool keywords"
   - port42 ls /tools/by-transform/[relevant-transforms]/
   - port42 ls /similar/[existing-tool]/
   - port42 ls /artifacts/document/ (find specs and patterns)
   - port42 search "architectural patterns domain knowledge"
   
2. MANDATORY CONTEXT GATHERING:
   - Similar tools: --ref p42:/commands/existing-tool
   - Domain specs: --ref p42:/artifacts/document/relevant-spec  
   - Patterns: --ref search:"patterns architecture"
   - Requirements: --ref file:/path/to/requirements
   - Memory: --ref p42:/memory/relevant-session

3. SINGLE POSSESS CALL:
   - possess does ONE thing with the context you provide
   - It will NOT search, explore, or make decisions
   - You are the orchestrator

WRONG:
❌ port42 possess @ai-engineer "create a log analyzer"

RIGHT:
✅ port42 search "log analyzer"
✅ port42 ls /tools/by-transform/log/
✅ port42 possess @ai-engineer "create log analyzer" --ref p42:/commands/log-parser
</orchestration_requirements>
```

### 4. Update ALL Practical Examples

**Location**: Replace each example in `<practical_examples>` section  
**Purpose**: Show full orchestration workflow

Example format:
```markdown
<scenario>User: "I need to analyze our API response times from server logs"</scenario>
<port42_approach>
1. port42 search "log analyze response time api"
2. port42 ls /tools/by-transform/log/
3. port42 ls /tools/by-transform/analyze/  
4. port42 ls /similar/log-analyzer/
5. port42 ls /artifacts/document/ | grep -i "api\|log\|performance"
6. port42 search "api performance analysis patterns"
7. port42 info /commands/log-analyzer
8. port42 cat /artifacts/document/api-performance-spec  # Get domain knowledge
9. port42 possess @ai-engineer --ref p42:/commands/log-analyzer --ref p42:/artifacts/document/api-performance-spec --ref file:/path/to/sample-log.txt --ref search:"response time analysis" "create api-response-analyzer that extracts and analyzes API response times"
10. api-response-analyzer --help
11. Result: Tool created with existing patterns AND domain knowledge
12. Value: Incorporates both code patterns and architectural specs
</port42_approach>
```

### 5. Add New Section: Multi-Step Workflows

**Location**: New section after practical examples  
**Purpose**: Teach complex orchestration patterns

```markdown
## Multi-Step Workflows with Possess

Since possess does ONE thing, complex tasks require orchestration:

### Pattern: Analyze Then Create
```bash
# Step 1: Discover existing landscape
port42 search "test framework"
port42 ls /tools/by-transform/test/
port42 ls /artifacts/document/ | grep -i "test\|qa\|validation"

# Step 2: Get analysis with full context
port42 possess @ai-analyst "analyze testing tool patterns and architectural best practices" \
  --ref p42:/commands/test-runner \
  --ref p42:/commands/validator \
  --ref p42:/artifacts/document/testing-best-practices

# Step 3: Create based on analysis and specs
port42 possess @ai-engineer "create improved test framework" \
  --ref search:"testing patterns" \
  --ref p42:/artifacts/document/analysis-results \
  --ref p42:/artifacts/document/test-architecture-spec
```

### Pattern: Incremental Enhancement
```bash
# Don't try to do everything in one possess call
# Break it down:

port42 possess @ai-engineer "add error handling" --ref p42:/commands/tool
port42 possess @ai-engineer "add progress bars" --ref p42:/commands/tool  
port42 possess @ai-engineer "add config file support" --ref p42:/commands/tool
```

### Pattern: Discovery → Decision → Action
```bash
# Standard workflow for ANY tool request:

# 1. DISCOVER
results=$(port42 search "keywords")
tools=$(port42 ls /tools/by-transform/capability/)

# 2. DECIDE  
# Analyze what exists, identify gaps

# 3. ACTION
port42 possess @ai-engineer "specific request" \
  --ref p42:/commands/relevant-tool \
  --ref search:"patterns"
```
```

### 6. Add Memory/Context Section

**Location**: New subsection in reference system  
**Purpose**: Clarify session continuation

```markdown
### Conversation Continuity with Single-Purpose Possess

Possess supports session continuation, but still does ONE thing:

```bash
# Continue previous discussion - but possess still does ONE action
port42 memory  # Review what was discussed
port42 possess @ai-engineer cli-session-123 "implement what we discussed"

# With explicit memory reference
port42 possess @ai-engineer "enhance the tool" --ref p42:/memory/session-123
```

Remember: Even with context, possess performs ONE action per call.
```

### 7. Update Command Execution Section

**Location**: In `<command_execution>` section  
**Purpose**: Emphasize orchestration responsibility

Add warning box:
```markdown
⚠️ **ORCHESTRATION IS YOUR RESPONSIBILITY**
- Possess won't explore for you
- Possess won't search for you  
- Possess won't make decisions for you
- You must provide complete context via --ref

Think of possess like a compiler:
- You gather all the source files (discovery)
- You set all the flags (references)
- You invoke it once with everything it needs
- It produces one output
```

### 8. Add Common Mistakes Section

**Location**: New section near the end  
**Purpose**: Prevent regression to old patterns

```markdown
## Common Mistakes with Single-Purpose Possess

### ❌ Mistake 1: Skipping Discovery
```bash
# WRONG - No context about existing tools
port42 possess @ai-engineer "create json formatter"
```

### ❌ Mistake 2: Expecting Multi-Step from Possess
```bash
# WRONG - Possess won't search then create
port42 possess @ai-engineer "search for similar tools then create improved version"
```

### ❌ Mistake 3: Not Using References
```bash
# WRONG - No context provided
port42 possess @ai-engineer "make it better"

# RIGHT - Explicit reference
port42 possess @ai-engineer "make it better" --ref p42:/commands/tool-to-improve
```

### ❌ Mistake 4: Trying to Do Too Much in One Call
```bash
# WRONG - Multiple operations
port42 possess @ai-engineer "analyze logs, create tool, and document it"

# RIGHT - Separate calls
port42 possess @ai-analyst "analyze log patterns" --ref file:logs.txt
port42 possess @ai-engineer "create log tool" --ref search:"analysis"  
port42 possess @ai-engineer "create documentation" --ref p42:/commands/log-tool
```
```

## Implementation Notes

1. **Preserve Existing Knowledge**: Don't remove Claude Code's existing Port42 knowledge, just clarify that possess won't do discovery internally

2. **Emphasize Orchestration**: Make it clear that Claude Code is the orchestrator and must handle the complete workflow

3. **Show Patterns**: Provide clear patterns for common workflows that previously worked in one possess call

4. **Use Visual Cues**: Add warning boxes, ❌/✅ examples, and clear section headers

5. **Maintain Backwards Compatibility**: Existing --ref patterns still work, they're just mandatory now

## Testing the Changes

After updating P42CLAUDE.md, test with these prompts to Claude Code:

1. "Create a tool to count lines of code" - Should trigger search first
2. "Improve the existing log analyzer" - Should use --ref to existing tool
3. "Build on what we discussed earlier" - Should use memory references
4. "Create a testing framework" - Should explore /tools/by-transform/test/ first

## Success Criteria

- Claude Code always searches before creating tools
- Claude Code uses --ref for context in every possess call
- Claude Code breaks complex requests into multiple possess calls
- No more "possess failed with exit status 1" from missing discovery