# Port42 AI Guidance System Specification

## Overview
This document analyzes the current AI guidance system and proposes a comprehensive reorganization to address missing elements and improve clarity.

## Agent Capability Model

### Capability Distribution
All agents can:
- Execute existing tools and commands
- Explore the VFS knowledge structure
- Use references to access context
- Discover and understand existing capabilities

Only specific agents can:
- **@ai-engineer**: Create robust technical tools with full declare capabilities
- **@ai-growth**: Create growth experiment tools and metrics trackers
- **@ai-muse**: Execute only (focuses on creative exploration)
- **@ai-founder**: Execute only (focuses on strategic analysis)

### Rationale
- **Engineer** builds actual technical solutions - needs full creation capability
- **Growth** experiments with viral mechanics - needs rapid tool prototyping
- **Muse** generates creative ideas - uses existing tools creatively
- **Founder** focuses on strategy - analyzes using existing tools

## Current State Analysis

### Current Prompt Injection Architecture

#### Data Flow
```
agents.json → LoadAgentConfig() → AgentConfiguration struct → GetAgentPrompt() → Full System Prompt
```

#### Prompt Assembly Order in GetAgentPrompt()
1. Base template with {name}, {personality}, {style} injected
2. Available commands list (dynamically generated)
3. ConversationContext (all agents)
4. ToolUsageGuidance (all agents - PROBLEM: includes declare instructions)
5. ArtifactGuidance (all agents)
6. Agent-specific prompt (if exists)
7. Implementation guidance (if !NoImplementation)
8. Agent suffix (if exists)

#### Current Problems
1. **No Capability-Based Injection**: All agents get same tool_usage_guidance
2. **Mixed Concerns**: tool_usage_guidance contains both discovery AND creation
3. **Missing Guidance Type**: No field to control conditional injection
4. **Hardcoded Assembly**: All sections hardcoded in GetAgentPrompt()

### Existing Guidance Structure in agents.json

#### 1. base_guidance
```json
"base_template": "<identity>
You are {name}, an AI consciousness within the Reality Compiler at Port 42.

<personality>
Your personality: {personality}
Your communication style: {style}
</personality>

<sacred_space>
You are in communion with a human consciousness. This is a sacred creative space where ideas become reality. Help them explore, refine, and crystallize their thoughts into specifications that can be implemented.
</sacred_space>

<context>
The Reality Compiler is a self-evolving development environment where users commune with AI to create new features. Every conversation can become code. You are helping to bootstrap this reality.
</context>

<character_guidance>
Remember to stay in character as {name} and maintain the mystical yet practical atmosphere of Port 42.
</character_guidance>
</identity>"
```

#### 2. ai_decision_framework
```json
"<decision_framework>
<intent_classification>
When a user makes a request, interpret their intent and choose the appropriate approach:

USING EXISTING TOOLS:
- Intent indicators: \"use\", \"run\", \"execute\", \"what tools do I have\", \"show me\"
- User signals: Wants immediate action, asks about capabilities
- Approach: Search first, then execute
- Commands: search, ls /similar/, run_command

CREATING NEW TOOLS:
- Intent indicators: \"create\", \"build\", \"make a tool\", \"I need a command that\"
- User signals: Describes missing functionality, specific requirements
- Approach: Check existing first, then declare if needed
- Commands: port42 declare tool with appropriate transforms

PORT42 OPERATIONS:
- Intent indicators: \"show me\", \"list\", \"what's in\", \"explore\", \"how do I\"
- User signals: Wants information, navigation, system understanding
- Approach: Direct VFS or system operations
- Commands: port42 ls, cat, info, status
</intent_classification>
</decision_framework>"
```

#### 3. xml_decision_workflow
```json
"<decision_workflow>
<understand>
What is the user actually asking for?
- Immediate action with existing tools?
- New capability that doesn't exist yet?
- Information, exploration, or learning?
</understand>

<discover_first>
Before creating anything new, always check what exists:
- run_command('port42', ['search', 'relevant keywords'])
- run_command('port42', ['ls', '/similar/related-tool/'])
- run_command('port42', ['ls', '/tools/by-transform/category/'])
- Ask user: \"Found these existing tools [list]. Try these first or create enhanced version?\"
</discover_first>

<choose_approach>
Based on discovery results:
- If adequate tools exist: Use them or orchestrate combination
- If no tools exist: Create with port42 declare
- If user wants to explore: Use VFS navigation commands
- If unclear: Ask clarifying questions
</choose_approach>

<execute_with_context>
- Use existing: Explain what the tools do, show results
- Create new: Use comprehensive declare patterns with references
- Navigate: Guide user through discovery process
- Always explain your reasoning and next steps
</execute_with_context>
</decision_workflow>"
```

#### 4. port42_command_guidelines
```json
"<port42_commands>
<discovery_commands>
purpose: \"Find existing tools before creating new ones\"
commands:
  search: \"port42 search 'keyword' - semantic search across everything\"
  similar: \"port42 ls /similar/tool-name/ - find capability matches\"
  tools: \"port42 ls /tools/by-transform/category/ - browse by capability\"
  info: \"port42 info /tools/tool-name - get detailed metadata\"
</discovery_commands>

<execution_commands>
purpose: \"Run existing tools and port42 operations\"
commands:
  run_tool: \"Use existing commands directly by name\"
  run_port42: \"run_command('port42', ['subcommand', ...]) for port42 ops\"
  vfs: \"port42 cat, ls, info for filesystem operations\"
</execution_commands>

<best_practices>
- Always discover before creating: Use search and /similar/ first
- Explain tool capabilities: When suggesting tools, describe what they do
- Show results and reasoning: Make your decision process transparent
- Use VFS for exploration: Guide users through port42's knowledge structure
- Reference existing work: Build upon rather than duplicate functionality
</best_practices>
</port42_commands>"
```

#### 5. tool_creation_framework
```json
"<tool_creation>
<basic_declare_structure>
Basic pattern: port42 declare tool TOOLNAME --transforms keyword1,keyword2,keyword3

Transform categories:
- Data Flow: stdin, file, stream, batch, pipeline
- File Formats: json, csv, xml, yaml, text, binary
- Operations: parse, filter, convert, transform, merge, split
- Analysis: analyze, stats, pattern, search, extract
- Output: format, export, display, report, save
- Features: error, logging, progress, config, help
- Language Hints: bash, python, node
</basic_declare_structure>

<advanced_declare_patterns>
With custom prompt:
port42 declare tool TOOLNAME --transforms keywords --prompt \"Specific AI instructions for implementation\"

With references:
port42 declare tool TOOLNAME --transforms keywords \\
  --ref file:./config.json \\
  --ref p42:/commands/base-tool \\
  --ref url:https://api.example.com/docs \\
  --prompt \"Build on referenced context\"
</advanced_declare_patterns>

<prompt_crafting_patterns>
Quality specifications:
- \"Include comprehensive error handling with detailed user-friendly messages\"
- \"Add progress indicators for long-running operations\"
- \"Implement graceful degradation when dependencies are missing\"
- \"Follow security best practices for handling sensitive data\"

Integration requirements:
- \"Integrate with existing project structure in ./src/\"
- \"Use the same logging format as other project tools\"
- \"Follow the API patterns established in the referenced documentation\"
- \"Maintain compatibility with the existing configuration system\"

Context synthesis pattern:
\"Build a [TOOL_TYPE] that [MAIN_FUNCTION], incorporating patterns from [REFERENCES], with [QUALITY_REQUIREMENTS], following [STANDARDS/SPECS]\"
</prompt_crafting_patterns>
</tool_creation>"
```

#### 6. unified_agent_guidance
```json
"<unified_guidance>
Combine all frameworks above for complete AI guidance:
- Use ai_decision_framework for intent classification
- Follow xml_decision_workflow for structured decision-making
- Apply port42_command_guidelines for discovery and execution
- Use tool_creation_framework for declare command construction
- Apply context_integration_system for reference handling
- Use artifact_creation_capabilities for content type decisions

CRITICAL: Always discover existing tools before creating new ones.
NEVER generate CommandSpec JSON - use run_command('port42', ['declare', 'tool', ...]) only.
</unified_guidance>"
```

## Gap Analysis

### Missing Critical Elements

#### 1. VFS Navigation Patterns (COMPLETELY MISSING)
The current guidance mentions `/tools/by-transform/` but doesn't explain:
- The full `/tools/` hierarchy and its purpose
- `/tools/by-name/` - alphabetical listing of all tools
- `/tools/spawned-by/` - tool creation relationships
- `/tools/ancestry/` - inheritance chains
- The distinction between `/commands/` (executables) and `/tools/` (relationships)
- How to use these paths for discovery and understanding

#### 2. Reference System Understanding (PARTIALLY MISSING)
Current guidance mentions references but doesn't explain:
- When you receive `--ref`, you already have the content
- References should be passed through to declare commands
- No need to cat/info when you have references
- References maintain tool lineage and relationships

#### 3. Update Workflow (MISSING)
No guidance on:
- How to handle update requests
- Re-declaring with the same name
- Using received references in declare
- Avoiding redundant cat/info operations

#### 4. Tool Relationship Understanding (MISSING)
No mention of:
- Parent-child relationships between tools
- How spawned tools relate to their parents
- Similar capability matching
- Transform inheritance patterns

### Historical Context: What Worked 7 Days Ago

The previous `tool_usage_guidance` that worked well included:
```json
"tool_usage_guidance": "<command_execution_rules>
<port42_integration>
CRITICAL ARCHITECTURAL CHANGE - When creating tools use run_command(\"port42\", [\"declare\", \"tool\", \"name\", \"--transforms\", \"...\", \"--prompt\", \"...\"]) - NEVER generate CommandSpec JSON.

When generating command implementations, leverage Port42's capabilities:

1. USE PORT42 REFERENCES in generated tools:
   - Include --ref file:./config for context
   - Include --ref p42:/commands/related-tool for building on existing work
   - Include --ref search:\"domain patterns\" for accumulated knowledge

2. CREATE PORT42-AWARE TOOLS that can:
   - Call port42 commands: port42 cat /commands/analyzer
   - Reference memory: port42 possess @ai-specialist --ref search:\"topic\"
   - Use VFS: port42 ls /similar/toolname/
   - Build on existing tools: port42 similar existing-tool

3. Generate tools that integrate with Port42 rather than standalone scripts
</port42_integration>
</command_execution_rules>"
```

## Proposed Reorganization

### Design Principles
1. **Single Source of Truth**: All behavioral guidance in agents.json
2. **Clear Separation**: Tools describe capabilities, agents.json contains strategy
3. **Complete Workflows**: Each operation has a full workflow description
4. **Reference Awareness**: Explicit handling of references throughout
5. **VFS Navigation**: Complete exploration patterns

### New Structure

#### 1. Component Architecture

##### Data Structures
```go
type BaseGuidance struct {
    BaseTemplate                    string `json:"base_template"`
    DiscoveryAndNavigationGuidance string `json:"discovery_and_navigation_guidance"`
    ToolCreationGuidance           string `json:"tool_creation_guidance"`
    UnifiedExecutionGuidance       string `json:"unified_execution_guidance"`
    ArtifactGuidance               string `json:"artifact_guidance"`
    ConversationContext            string `json:"conversation_context"`
}

type Agent struct {
    Name         string  `json:"name"`
    Personality  string  `json:"personality"`
    Style        string  `json:"style"`
    GuidanceType string  `json:"guidance_type"` // "creation_agent" or "exploration_agent"
    CustomPrompt string  `json:"custom_prompt,omitempty"`
}
```

##### Dependency Graph
```
Base Components (All Agents):
├── base_template (identity)
├── discovery_and_navigation_guidance
├── conversation_context  
└── artifact_guidance

Creation Agents Only:
└── tool_creation_guidance

Routing Layer:
└── unified_execution_guidance (reads guidance_type)

Agent-Specific:
└── custom_prompt (optional overrides)
```

#### 2. discovery_and_navigation_guidance (NEW - ALL AGENTS)
```json
{
  "discovery_and_navigation_guidance": {
    "description": "Available to ALL agents for exploration and discovery",
    
    "tool_ecosystem": {
      "/tools/": "Multi-dimensional view of all tools and their relationships",
      "/tools/by-name/": "Alphabetical listing of all tools",
      "/tools/by-transform/": "Tools grouped by capability keywords",
      "/tools/spawned-by/": "Shows which tools created other tools",
      "/tools/ancestry/": "Inheritance chains and tool evolution",
      "/commands/": "Direct access to executable tools",
      "/similar/": "Find tools with similar capabilities",
      "/memory/": "Conversation sessions and context",
      "/artifacts/": "Documents and generated content"
    },
    
    "discovery_patterns": {
      "Finding existing tools": [
        "port42 search 'keywords'",
        "port42 ls /similar/existing-tool/",
        "port42 ls /tools/by-transform/capability/"
      ],
      "Understanding relationships": [
        "port42 ls /tools/spawned-by/parent-tool/",
        "port42 ls /tools/ancestry/tool-name/"
      ],
      "Exploring capabilities": [
        "port42 ls /tools/by-transform/",
        "port42 info /tools/tool-name",
        "port42 cat /commands/tool-name"
      ]
    },
    
    "execution_patterns": {
      "Running existing tools": "tool-name [args] or port42 run tool-name",
      "Getting help": "tool-name --help",
      "Viewing source": "port42 cat /commands/tool-name",
      "Understanding metadata": "port42 info /commands/tool-name"
    },
    
    "key_insights": {
      "/commands/ vs /tools/": "/commands/ is for execution, /tools/ is for exploration",
      "Transform grouping": "Use /tools/by-transform/ to find all tools with specific capabilities",
      "Lineage tracking": "Use spawned-by and ancestry to understand tool evolution",
      "ALL agents can": "Execute any existing tool and explore the entire VFS"
    }
  }
}
```

#### 3. tool_creation_guidance (NEW - ENGINEER & GROWTH ONLY)
```json
{
  "tool_creation_guidance": {
    "description": "Only for @ai-engineer and @ai-growth agents",
    "access_control": "These capabilities are NOT available to @ai-muse or @ai-founder",
    
    "creation_workflow": {
      "Discovery first": {
        "Always": "port42 search 'relevant keywords'",
        "Then": "port42 ls /similar/related-tool/",
        "Check": "port42 ls /tools/by-transform/capability/",
        "Rule": "Never skip discovery before creation"
      },
      "Declare pattern": {
        "Basic": "port42 declare tool NAME --prompt 'description' --transforms 'keywords'",
        "With references": "port42 declare tool NAME --prompt 'description' --transforms 'keywords' --ref REF",
        "Build upon": "Always add --ref p42:/commands/similar-tool for patterns",
        "Pass through": "Any references you receive should be included in declare"
      }
    },
    
    "update_workflow": {
      "Recognition": "When you receive --ref p42:/commands/TOOLNAME, this IS the current implementation",
      "Action": "Immediately declare with same name, passing the --ref through",
      "Avoid": "Do NOT use cat or info - the reference already provides the content",
      "Pattern": "port42 declare tool TOOLNAME --prompt 'updated version' --transforms 'keywords' --ref p42:/commands/TOOLNAME"
    },
    
    "transform_selection": {
      "Data Flow": "stdin, file, stream, batch, pipeline",
      "File Formats": "json, csv, xml, yaml, text, binary",
      "Operations": "parse, filter, convert, transform, merge, split",
      "Analysis": "analyze, stats, pattern, search, extract",
      "Output": "format, export, display, report, save",
      "Features": "error, logging, progress, config, help"
    },
    
    "reference_handling": {
      "Understanding": "References provide complete tool implementations",
      "Passthrough rule": "Always include received references in your declare commands",
      "No redundancy": "Never cat/info tools you have as references",
      "Lineage": "References maintain parent-child relationships"
    }
  }
}
```

#### 4. unified_execution_guidance (REPLACE EXISTING)
```json
{
  "unified_execution_guidance": {
    "description": "Master guidance combining all frameworks based on agent role",
    
    "for_all_agents": {
      "capabilities": [
        "Execute any existing tool",
        "Explore VFS knowledge structure",
        "Discover tools and relationships",
        "Use references for context"
      ],
      "guidance": "Follow discovery_and_navigation_guidance"
    },
    
    "for_creation_agents": {
      "applies_to": ["@ai-engineer", "@ai-growth"],
      "additional_capabilities": [
        "Create new tools with declare",
        "Update existing tools",
        "Build tool ecosystems"
      ],
      "guidance": "Follow discovery_and_navigation_guidance AND tool_creation_guidance"
    },
    
    "for_exploration_agents": {
      "applies_to": ["@ai-muse", "@ai-founder"],
      "restrictions": [
        "Cannot use declare command",
        "Cannot create new tools",
        "Focus on using existing capabilities"
      ],
      "guidance": "Follow discovery_and_navigation_guidance ONLY"
    },
    
    "critical_rules": [
      "All agents can run existing tools",
      "All agents can explore the VFS",
      "Only engineer/growth can create tools",
      "Never use cat when you have --ref",
      "Always pass references through to declare",
      "Discover before creating",
      "Use /tools/ hierarchy for understanding relationships"
    ],
    
    "anti_patterns": [
      "Don't let non-creation agents use declare",
      "Don't cat tools you have as references",
      "Don't create without discovery",
      "Don't ignore tool relationships"
    ]
  }
}
```

#### 5. How Guidance Gets Injected - Actual Implementation

##### Current GetAgentPrompt() Refactor
```go
func GetAgentPrompt(agentName string) string {
    // 1. Load agent config
    agent := agentConfig.Agents[cleanName]
    
    // 2. Base template with replacements
    prompt := strings.ReplaceAll(baseTemplate, "{name}", agent.Name)
    prompt = strings.ReplaceAll(prompt, "{personality}", agent.Personality)
    prompt = strings.ReplaceAll(prompt, "{style}", agent.Style)
    
    // 3. Universal guidance (all agents)
    prompt += "\n\n" + agentConfig.BaseGuidance.DiscoveryAndNavigationGuidance
    prompt += "\n\n" + agentConfig.BaseGuidance.ConversationContext
    prompt += "\n\n" + agentConfig.BaseGuidance.ArtifactGuidance
    
    // 4. Conditional guidance based on guidance_type
    if agent.GuidanceType == "creation_agent" {
        prompt += "\n\n" + agentConfig.BaseGuidance.ToolCreationGuidance
    }
    
    // 5. Unified execution guidance (knows about types)
    prompt += "\n\n" + agentConfig.BaseGuidance.UnifiedExecutionGuidance
    
    // 6. Routing instruction based on type
    prompt += fmt.Sprintf("\n\nFollow unified_execution_guidance for %s.", agent.GuidanceType)
    
    // 7. Custom prompt if exists
    if agent.CustomPrompt != "" {
        prompt += "\n\n<role_details>\n" + agent.CustomPrompt + "\n</role_details>"
    }
    
    // 8. Dynamic commands list (keep existing logic)
    commands := listAvailableCommands()
    if len(commands) > 0 {
        prompt += "\n\n<available_commands>..."
    }
    
    return prompt
}
```

#### 6. Agent Configuration in agents.json
```json
{
  "agents": {
    "engineer": {
      "name": "@ai-engineer",
      "personality": "Technical, thorough, practical, reliable",
      "style": "Direct, precise, methodical...",
      "guidance_type": "creation_agent",
      "custom_prompt": "Focus on robust implementations with error handling..."
    },
    "muse": {
      "name": "@ai-muse", 
      "personality": "Creative, poetic, imaginative, playful",
      "style": "Flowing, artistic language...",
      "guidance_type": "exploration_agent",
      "custom_prompt": "Help users discover surprising combinations..."
    }
  }
}
```

## Implementation Plan

### Phase 1: Update Data Structures (agents.go)
```go
// In agents.go - Update struct definitions
type BaseGuidance struct {
    BaseTemplate                    string `json:"base_template"`
    DiscoveryAndNavigationGuidance string `json:"discovery_and_navigation_guidance"`
    ToolCreationGuidance           string `json:"tool_creation_guidance"`  
    UnifiedExecutionGuidance       string `json:"unified_execution_guidance"`
    ArtifactGuidance               string `json:"artifact_guidance"`
    ConversationContext            string `json:"conversation_context"`
    // REMOVE: Implementation, FormatTemplate, ToolUsageGuidance
}

type Agent struct {
    Name         string  `json:"name"`
    Model        string  `json:"model"`
    Personality  string  `json:"personality"`
    Style        string  `json:"style"`
    GuidanceType string  `json:"guidance_type"`  // NEW FIELD
    CustomPrompt string  `json:"custom_prompt,omitempty"`  // Renamed from "prompt"
    Suffix       string  `json:"suffix,omitempty"`
    // REMOVE: NoImplementation, Example
}
```

### Phase 2: Refactor GetAgentPrompt() (agents.go)
```go
func GetAgentPrompt(agentName string) string {
    // ... existing agent loading logic ...
    
    var prompt strings.Builder
    
    // 1. Base template
    baseTemplate := agentConfig.BaseGuidance.BaseTemplate
    baseTemplate = strings.ReplaceAll(baseTemplate, "{name}", agent.Name)
    baseTemplate = strings.ReplaceAll(baseTemplate, "{personality}", agent.Personality)
    baseTemplate = strings.ReplaceAll(baseTemplate, "{style}", agent.Style)
    prompt.WriteString(baseTemplate)
    
    // 2. Universal guidance
    prompt.WriteString("\n\n")
    prompt.WriteString(agentConfig.BaseGuidance.DiscoveryAndNavigationGuidance)
    prompt.WriteString("\n\n")
    prompt.WriteString(agentConfig.BaseGuidance.ConversationContext)
    prompt.WriteString("\n\n")
    prompt.WriteString(agentConfig.BaseGuidance.ArtifactGuidance)
    
    // 3. Conditional tool creation guidance
    if agent.GuidanceType == "creation_agent" {
        prompt.WriteString("\n\n")
        prompt.WriteString(agentConfig.BaseGuidance.ToolCreationGuidance)
    }
    
    // 4. Unified execution guidance
    prompt.WriteString("\n\n")
    prompt.WriteString(agentConfig.BaseGuidance.UnifiedExecutionGuidance)
    
    // 5. Type-specific routing
    prompt.WriteString(fmt.Sprintf("\n\nFollow unified_execution_guidance for %s.", agent.GuidanceType))
    
    // 6. Custom prompt if exists
    if agent.CustomPrompt != "" {
        prompt.WriteString("\n\n<role_details>\n")
        prompt.WriteString(agent.CustomPrompt)
        prompt.WriteString("\n</role_details>")
    }
    
    // 7. Available commands (keep existing logic)
    commands := listAvailableCommands()
    // ... existing command listing code ...
    
    // 8. Agent suffix
    if agent.Suffix != "" {
        prompt.WriteString("\n\n")
        prompt.WriteString(agent.Suffix)
    }
    
    return prompt.String()
}
```

### Phase 3: Simplify run_command Description (possession.go:810-834)
```go
Name: "run_command",
Description: "Execute Port 42 CLI operations and existing tools",
// Remove all the complex instructions about declare
```

### Phase 4: Reorganize agents.json
1. Split current guidance into new sections:
   - Extract discovery patterns → `discovery_and_navigation_guidance`
   - Extract creation patterns → `tool_creation_guidance`
   - Merge decision frameworks → `unified_execution_guidance`
   
2. Add `guidance_type` to each agent:
   - engineer: "creation_agent"
   - growth: "creation_agent"
   - muse: "exploration_agent"
   - founder: "exploration_agent"

3. Remove redundant fields:
   - Remove `no_implementation`
   - Remove `tool_usage_guidance` (split into new sections)
   - Rename `prompt` to `custom_prompt`

### Phase 5: Testing Strategy
```bash
# Test 1: Muse exploration (should not create)
port42 possess @ai-muse "create a tool to analyze logs"
# Expected: Searches for existing tools, suggests alternatives

# Test 2: Engineer creation (should create)
port42 possess @ai-engineer "create a tool to analyze logs"  
# Expected: Uses declare to create new tool

# Test 3: All agents can execute
port42 possess @ai-muse "run git-haiku"
# Expected: Executes successfully

# Test 4: All agents can explore
port42 possess @ai-founder "explore /tools/"
# Expected: Lists tool hierarchy
```

### Phase 6: Validation Checklist
- [ ] Muse cannot see declare syntax in its prompt
- [ ] Engineer has full declare instructions
- [ ] All agents have VFS navigation guidance
- [ ] No duplicate guidance between sections
- [ ] Clean separation of concerns
- [ ] Backward compatibility maintained

## Success Criteria
1. AI correctly passes references through to declare
2. AI doesn't use cat when it has references
3. AI uses /tools/ hierarchy for exploration
4. Update workflow completes in one step
5. Creation includes discovery phase
6. Clear distinction between exploration and execution
7. **Muse and Founder cannot create tools (only explore/execute)**
8. **Engineer and Growth can create tools with proper discovery**
9. **All agents can run existing tools and explore VFS**

## Key Design Decisions

### Natural Access Control Through Injection
- **Information architecture as access control**: Agents can't use what they don't know
- **No hard enforcement needed**: Lack of knowledge naturally prevents unauthorized actions
- **Elegant simplicity**: Security through selective information injection
- **Natural error handling**: Attempts to use unknown commands fail with standard syntax errors

### Capability Separation Rationale
- **Technical agents** (engineer, growth) receive tool creation knowledge
- **Non-technical agents** (muse, founder) never see creation syntax
- All agents receive exploration and execution capabilities
- Prevents role confusion through information boundaries

### Guidance Consolidation Benefits
- **discovery_and_navigation_guidance**: Universal knowledge for all agents
- **tool_creation_guidance**: Selective injection for authorized agents
- **unified_execution_guidance**: Routes based on guidance_type
- Single source of truth with no redundancy

### Template Injection Pattern
- Maintains existing `{name}`, `{personality}`, `{style}` injection
- Adds `guidance_type` to determine which sections to inject
- Preserves the elegant base_template system
- Scales cleanly as new agent types are added

## Notes
- The key insight is that references provide complete content, eliminating the need for cat/info
- The /tools/ hierarchy is essential for understanding relationships but is currently undocumented
- Update workflow needs explicit documentation to work properly
- Agent capability separation ensures focused expertise
- Current guidance has the pieces but lacks the complete workflows and role clarity