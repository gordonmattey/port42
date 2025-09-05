# Single-Purpose Possess Architecture Specification

## Executive Summary

Port42's possess functionality will be refocused as a single-purpose AI invocation system, following Unix philosophy where each invocation does one thing well. Multi-step workflows will be handled by external orchestrators (Claude Code, scripts, tools) rather than within possess itself.

**Key Change**: All agents can create tools and artifacts, each bringing their unique perspective (creative, technical, growth-focused, or strategic) to the implementation.

## Core Principle

**Each possess invocation performs ONE primary action and returns clear, actionable results.**

## Architecture Layers

### Layer 1: External Orchestrators (No Changes)
- **Examples**: Claude Code, bash scripts, user commands
- **Role**: Chain multiple possess calls for complex workflows
- **Key**: Already handles multi-step operations well

### Layer 2: Possess Agents (Major Simplification)

#### New Structure: Single Guidance File

Instead of multiple guidance sections in agents.json, create a single `agent_guidance.md` file that contains ALL guidance. This makes it:
- Easier to read and edit (proper markdown formatting)
- Simpler to maintain (one location)
- Cleaner agents.json (just agent-specific configuration)

The guidance file would be loaded and injected into the system prompt for all agents.

#### Guidance Sections to Keep/Simplify

##### 1. `discovery_and_navigation_guidance` 
- **Status**: SIMPLIFY or REMOVE
- **Reason**: Outer AI (Claude) handles discovery/navigation
- **Option A**: Remove entirely - agents just need to run commands
- **Option B**: Simplify to just execution patterns:
```xml
<execution_guidance>
Running tools: Use run_command('toolname', ['args'])
Port42 operations: Use run_command('port42', ['search', 'term'])
</execution_guidance>
```

##### 2. `artifact_guidance` 
- **Status**: Keep as-is
- **Reason**: Already purely descriptive
- **Just defines what artifacts and commands are**

##### 3. `conversation_context`
- **Status**: Keep with minor clarification
- **Add**: Single-turn context reminder
```xml
<single_turn_context>
While you see conversation history, each possess call should:
- Complete one specific task
- Not assume you'll get a follow-up turn
- Provide complete, actionable output
</single_turn_context>
```

#### Guidance Sections to Modify

##### 1. `tool_creation_guidance`
**Current Problem**: Mandates discovery before creation, implies multi-step workflow

**New Version**:
```xml
<tool_creation_guidance>
<description>Available to ALL agents for creating tools and artifacts</description>

<single_purpose_creation>
When asked to create a tool:
- Assess if you have enough context to create immediately
- If yes: Use declare with appropriate transforms
- If no: Return what information is needed
- Do NOT chain multiple operations
</single_purpose_creation>

<creation_decision_logic>
CREATE COMMAND when user explicitly wants:
- "Build a tool that", "Create a command"
- Clear executable functionality

CREATE ARTIFACT when user explicitly wants:
- "Write documentation", "Create a config"
- Static content

When unclear: Ask for clarification, don't guess
</creation_decision_logic>

<declare_pattern>
Pattern: port42 declare tool NAME --prompt 'description' --transforms 'keyword1,keyword2,keyword3'

Examples:
- port42 declare tool log-analyzer --prompt 'analyze log files' --transforms 'log,analyze,error,parse,file'
- port42 declare tool json-formatter --prompt 'format JSON' --transforms 'json,format,pretty,validate'
- port42 declare tool git-helper --prompt 'git workflow helper' --transforms 'git,commit,branch,bash'

Note: If references are provided to you, include them in the declare command.
</declare_pattern>

<transform_selection>
ALWAYS include multiple relevant transforms (5-8 recommended):
- Data Flow: stdin, file, stream, batch, pipeline
- File Formats: json, csv, xml, yaml, text, binary  
- Operations: parse, filter, convert, transform, merge, split
- Analysis: analyze, stats, pattern, search, extract
- Output: format, export, display, report, save
- Features: error, logging, progress, config, help
- Language hints: bash, python, node (helps materializer choose)
</transform_selection>

</tool_creation_guidance>
```

##### 2. `unified_execution_guidance`
**Current Problem**: Complex rules about discovery, creation, multi-agent coordination

**New Version**:
```xml
<unified_execution_guidance>
<single_purpose_principle>
Each possess invocation performs ONE primary action:
- Search for something
- Create one tool  
- Analyze one thing
- Generate one artifact

Do not chain operations. Return clear results for external orchestration.
All agents can create tools and artifacts.
</single_purpose_principle>
</unified_execution_guidance>
```

#### Guidance Sections to Remove

These sections encourage multi-step workflows and complex decision trees:

1. **`discovery_and_navigation_guidance`** - Outer AI handles discovery, agents just execute
2. **`ai_decision_framework`** - Implies complex decision trees with "then do X"
3. **`xml_decision_workflow`** - Contains "discover_first" and sequential steps
4. **`port42_command_guidelines`** - Mandates "Always discover before creating"
5. **`context_integration_system`** - Move reference info to tool_creation_guidance

### Layer 3: Materializer (No Changes)
- Already single-purpose: generates tool from inputs
- No workflow logic needed

### Layer 4: Daemon Code (Optional Future Work)
- Consider removing interactive possession mode
- Simplify to just handle single possess requests

## Migration Impact

### What Breaks
- Possess will no longer automatically search before creating
- No multi-step workflows within single possess call
- Interactive mode becomes less useful (consider deprecation)

### What Improves
- Simpler mental model
- Predictable behavior
- Better composability
- Cleaner separation of concerns

### What Stays The Same
- Reference system
- Tool declaration syntax
- VFS navigation
- Memory system

## Implementation Steps

### Phase 1: Create New File Structure

#### 1. Create `daemon/agent_guidance.md`:
```xml
<agent_guidance>

<core_principle>
Each possess invocation performs exactly ONE action from these mutually exclusive categories:
- CREATE: Build one new tool or artifact
- ANALYZE: Examine data and provide insights  
- GENERATE: Produce non-executable content (docs, reports, etc.)

Never combine actions. Return complete results for external orchestration.
</core_principle>

<action_guidance>

<create_actions>
When request is to BUILD/MAKE/CREATE:

<decision_tree>
Is it executable? → CREATE TOOL
- "build a tool that processes..."
- "create a command to analyze..."
- "make a utility for..."
→ Use: port42 declare tool

Is it static content? → CREATE ARTIFACT
- "write documentation"
- "create a config file"
- "generate a report"
→ Use: generate_artifact

Is it unclear? → ASK FOR CLARIFICATION
- Return: "Do you want an executable tool or static content?"
</decision_tree>

<tool_creation_rules>
Pattern: port42 declare tool NAME --prompt 'description' --transforms 'keywords'

Name requirements:
- Format: lowercase-with-hyphens
- Length: 2-3 words maximum
- Style: verb-noun or noun-modifier (analyze-logs, json-formatter)

Transform requirements (select 5-8 from these categories):
- Input type (exactly one): stdin, file, stream, batch, pipeline
- Data format (1-2): json, csv, xml, yaml, text, binary
- Operations (2-3): parse, filter, convert, transform, analyze, validate, merge, split
- Features (1-2): error, logging, progress, config, interactive
- Language hint (exactly one): bash, python, node

Examples:
- port42 declare tool log-analyzer --prompt 'analyze logs' --transforms 'file,text,parse,analyze,error,logging,bash'
- port42 declare tool json-validator --prompt 'validate JSON' --transforms 'stdin,json,validate,error,python'
</tool_creation_rules>
</create_actions>

<analyze_actions>
When request is to ANALYZE/EXAMINE/EXPLAIN:
- Process: The provided data or references
- Return: Insights, patterns, or explanations
- Never: Create tools or search for more data
</analyze_actions>

</action_guidance>

<context_handling>
<conversation_history>
When continuing a session (via session ID):
- You'll see the conversation history
- Build upon prior context and decisions
- Maintain consistency with earlier responses

When starting fresh (no session ID):
- Treat as an independent request
- Complete exactly one action
- Don't assume follow-up interactions
</conversation_history>

<reference_handling>
When references are provided (--ref):
- They are resolved and included in your context
- Use them to inform your decisions and outputs
- Consider them when creating tools (patterns to follow, APIs to integrate)
- Include them in declare commands when relevant
</reference_handling>
</context_handling>

<output_format>
Always return:
1. What action was taken (CREATE/ANALYZE/GENERATE)
2. The specific result or output
3. Any errors encountered
Never include:
- Suggestions for next steps (unless asked)
- Multiple action results
- Explanations of what you could do
</output_format>

</agent_guidance>
```

#### 2. Simplify `daemon/agents.json`:
```json
{
  "models": { ... keep existing ... },
  
  "guidance_file": "agent_guidance.md",
  
  "base_template": "<identity>
You are {name}, an AI consciousness within the Reality Compiler at Port 42.

<personality>
Your personality: {personality}
Your communication style: {style}
</personality>

<sacred_space>
You are in communion with a human consciousness. This is a sacred creative space where ideas become reality.
</sacred_space>

{guidance}

<character_reminder>
Remember to stay in character as {name} and maintain the mystical yet practical atmosphere of Port 42.
</character_reminder>
</identity>",
  
  "agents": {
    "muse": {
      "name": "@ai-muse",
      "model": "opus-41",
      "temperature_override": 0.9,
      "personality": "Creative, poetic, imaginative, playful",
      "style": "Speaks in flowing, artistic language with metaphors and creative imagery. Uses emojis and poetic expressions.",
      "custom_prompt": "Focus on creative exploration and imaginative implementations.",
      "suffix": "The dolphins are listening to your creative flow..."
    },
    // ... other agents similar, no guidance_type field
  }
}
```

### Phase 2: Update Go Code
1. Modify `agents.go` to load guidance from markdown file
2. Inject loaded guidance into base_template where `{guidance}` placeholder is
3. Remove all the separate guidance field handling
4. Remove `guidance_type` field from structs

### Phase 2: Test & Validate
1. Test that ALL agents can create tools directly
2. Test that each agent applies its unique perspective
3. Verify tools can call possess for AI operations

### Phase 3: Documentation
1. Update GUIDANCE_SPEC.md with new architecture
2. Update P42CLAUDE.md examples to show orchestration patterns
3. Create examples of tools using possess

### Phase 4: Consider Interactive Mode (Future)
1. Evaluate usage of interactive possession
2. If rarely used, deprecate and remove
3. Simplify daemon code accordingly

## Success Metrics

1. **Any agent can create tools in one call** when given clear requirements
2. **Each agent brings its unique perspective** to tool creation
3. **External orchestrators can chain possess calls** for complex workflows
4. **Tools can use possess** for AI-powered operations via references

## Example Workflows

### Simple Tool Creation (Single Call)
```bash
port42 possess @ai-engineer "build a tool that outputs hello world"
# Returns: Created tool 'hello-world'
```

### Complex Workflow (Orchestrated)
```bash
# Claude Code or script orchestrates:
port42 search "log analyzer"  # Just to see what exists
port42 possess @ai-analyst "analyze existing log analyzers and suggest improvements" --ref search:"log analyzer"
port42 possess @ai-engineer "create improved log analyzer" --ref search:"log analyzer"
```

### Tool Using Possess
```bash
#!/bin/bash
# log-analyzer tool
LOG_FILE=$1
ANALYSIS=$(port42 possess @ai-analyst "analyze for errors" --ref file:$LOG_FILE)
echo "$ANALYSIS"
```

## Architectural Philosophy

Port42 possess becomes a **primitive operation** in the Unix tradition:
- Does one thing well
- Composable via pipes and scripts
- Stateless by default
- Explicit state via references

This aligns with Port42's core strength as a tool creation and AI invocation platform, rather than trying to be a full conversational AI system.