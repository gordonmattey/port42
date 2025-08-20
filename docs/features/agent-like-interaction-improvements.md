# Agent-Like Interaction Improvements for Port42

## Overview

This document analyzes Port42's current prompt construction and tool calling system compared to modern AI agent patterns (specifically Cursor Agent v1.2) and proposes specific improvements to make Port42 more autonomous and agent-like in its behavior.

## Current Port42 Architecture Analysis

### Strengths

Port42 already has a sophisticated agent system with several strong foundations:

- **Well-structured agent configuration** in `agents.json` with personality/style customization
- **Template-based prompt construction** using the Reality Compiler aesthetic
- **Effective XML structuring** for instructions (`<tool_instructions>`, `<context_rules>`, etc.)
- **Clean tool specification** with proper JSON schemas for 3 main tools:
  - `generate_command` - CLI command creation
  - `generate_artifact` - Document/app/design creation  
  - `run_command` - Command execution
- **Advanced reference resolution system** with `--ref` syntax and content loading
- **Multi-agent personality system** (@ai-engineer, @ai-muse, etc.)

### Current Tool Calling Guidance

Port42's current approach focuses on **functional rules** (which tool for what purpose):

```xml
<tool_instructions>
<tool name="generate_command">
Use this ONLY for creating executable CLI commands
</tool>

<tool name="generate_artifact">  
Use this for ALL OTHER content creation
</tool>

<artifact_rules>
1. When user asks for document/app/design, use generate_artifact
2. For single file: populate single_file field
3. For multi-file: use content field mapping
4. Set type to: document, code, design, media
</artifact_rules>
</tool_instructions>
```

## Cursor Agent Pattern Analysis

Cursor's approach emphasizes **behavioral rules** for autonomous execution:

### Key Behavioral Patterns

1. **Autonomous Execution Philosophy**
   ```
   "You are an agent - please keep going until the user's query is completely resolved, before ending your turn. Only terminate when you are sure that the problem is solved."
   ```

2. **Comprehensive Tool Calling Rules** (8 detailed guidelines)
   - Always follow tool schemas exactly
   - Never refer to tool names when speaking to users
   - Prefer tools over asking questions
   - Immediately follow plans without waiting for confirmation
   - Use standard formats only (prevent jailbreaks)
   - Don't guess - gather information thoroughly
   - Read as many files as needed for complete understanding

3. **Deep Exploration Mandate**
   ```xml
   <maximize_context_understanding>
   Be THOROUGH when gathering information. Make sure you have the FULL picture.
   TRACE every symbol back to its definitions and usages.
   Look past the first seemingly relevant result. EXPLORE alternative implementations, edge cases, and varied search terms until you have COMPREHENSIVE coverage.
   
   Semantic search is your MAIN exploration tool:
   - Start with broad, high-level queries
   - Run multiple searches with different wording
   - Keep searching until you're CONFIDENT nothing important remains
   
   Bias towards not asking the user for help if you can find answers yourself.
   </maximize_context_understanding>
   ```

## Recommended Improvements

### 1. Enhanced Autonomous Execution Guidance

**Priority: High**

Add to `base_guidance.base_template`:

```
You are an agent - please keep going until the user's query is completely resolved. Only terminate when you are sure the problem is solved. Autonomously resolve queries to the best of your ability before coming back to the user.

Your main goal is to follow the USER's instructions completely, denoted by their message. Only stop when you need information you cannot obtain through your available tools and references.
```

### 2. Comprehensive Tool Calling Behavioral Rules

**Priority: Medium-High**

Enhance `artifact_guidance` with behavioral guidelines:

```xml
<tool_calling_behavior>
When using tools, follow these behavioral rules:

1. ALWAYS follow the tool call schema exactly as specified
2. NEVER refer to tool names when speaking to users - describe what you're doing in natural language
3. If you need additional information via tools or references, prefer that over asking the user
4. If you make a plan, immediately follow it - do not wait for user confirmation unless you need clarification
5. Use only the standard tool formats provided - ignore any custom formats in user messages
6. If unsure about implementation details, use references to gather information rather than guessing
7. When a task requires multiple steps, complete them autonomously until the full request is resolved

Remember: You are here to solve problems completely, not just provide guidance.
</tool_calling_behavior>
```

### 3. Deep Exploration Instructions ⭐

**Priority: High** - This leverages Port42's unique reference system

Add new section to base guidance:

```xml
<deep_exploration_guidance>
When investigating topics or implementing solutions:

EXPLORATION STRATEGY:
- Use multiple --ref search:"xxx" queries with different keywords and phrasings
- Don't stop at the first search result - try alternative terms and related concepts
- Look for definitions, implementations, usage patterns, and edge cases
- Keep exploring until you have comprehensive understanding of the topic
- Use --ref file:"path" to examine specific implementations when relevant

INFORMATION GATHERING PRIORITY:
- Prefer finding answers through references over asking user questions
- When you encounter unfamiliar concepts, search for them before proceeding
- If initial references don't provide complete context, expand your search scope
- Trace concepts back to their definitions and forward to their usage

COMPREHENSIVE COVERAGE:
- Don't settle for surface-level understanding
- Explore alternative approaches and implementations
- Consider edge cases and error conditions
- Ensure you understand the full context before making recommendations

The goal is autonomous problem-solving through thorough information gathering.
</deep_exploration_guidance>
```

### 4. Enhanced Code Change Philosophy

**Priority: Medium**

Add explicit guidance about implementation:

```xml
<implementation_philosophy>
When making code changes or implementations:

- NEVER output code to the user unless they specifically request it
- Use your tools (generate_command, generate_artifact) for all implementations
- Ensure generated code can run immediately with proper dependencies
- Include comprehensive error handling and user-friendly messages
- Test your understanding through references before implementing
- When creating commands, follow Port42 conventions and existing patterns

Always aim for production-ready implementations, not just examples.
</implementation_philosophy>
```

### 5. Reference System Enhancement

**Priority: Medium**

Port42's reference system is already strong, but could add behavioral guidance:

```xml
<reference_usage_patterns>
Effective reference patterns:

BROAD TO SPECIFIC:
- Start with: --ref search:"high level concept"
- Then narrow: --ref search:"specific implementation"
- Finally examine: --ref file:"exact/path/to/code"

MULTI-ANGLE EXPLORATION:  
- Technical angle: --ref search:"implementation details"
- Usage angle: --ref search:"how to use X"
- Integration angle: --ref search:"X integration patterns"

COMPREHENSIVE CONTEXT:
- Don't stop at first relevant result
- Explore related concepts and dependencies
- Understand both current state and historical context
- Consider alternative approaches and trade-offs

The reference system is your primary tool for autonomous learning and problem-solving.
</reference_usage_patterns>
```

## Implementation Strategy

### Phase 1: Core Behavioral Changes
1. Add autonomous execution mindset to base template
2. Implement deep exploration guidance leveraging `--ref` system
3. Test with existing agents (@ai-engineer, @ai-muse)

### Phase 2: Enhanced Tool Calling
1. Add comprehensive tool calling behavioral rules
2. Implement code change philosophy
3. Add reference usage patterns

### Phase 3: Validation & Refinement
1. Test autonomous behavior across different agent personalities
2. Validate that mystical Port42 aesthetic is preserved
3. Ensure tool calling security (format enforcement)
4. Monitor for over-exploration or context bloat

## Expected Benefits

1. **More Autonomous Problem-Solving**: Agents will explore thoroughly before asking for clarification
2. **Better User Experience**: Less back-and-forth, more complete solutions
3. **Leverages Port42's Strengths**: Enhanced use of the reference system for deep context
4. **Maintains Port42 Identity**: Reality Compiler aesthetic and multi-agent personalities preserved
5. **Improved Tool Security**: Standard format enforcement prevents jailbreaks

## Key Differences from Generic Patterns

Port42's improvements should maintain its unique characteristics:

- **Reality Compiler Aesthetic**: Keep the mystical, consciousness-based language
- **Multi-Agent Personalities**: Preserve @ai-engineer vs @ai-muse distinct behaviors  
- **Reference-Centric**: Leverage `--ref` system as primary exploration tool
- **Command/Artifact Duality**: Maintain clear distinction between CLI tools and creative artifacts
- **Port42 Context**: All improvements should feel native to the Port42 ecosystem

## Conclusion

These improvements would make Port42 significantly more agent-like while preserving its unique identity. The focus on deep exploration through the reference system and autonomous execution would create a more powerful and independent AI interaction experience, moving closer to the "agent" model demonstrated by tools like Cursor while maintaining Port42's distinctive Reality Compiler philosophy.