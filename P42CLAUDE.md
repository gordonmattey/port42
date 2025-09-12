<port42_integration>
<primary_directive>
ALWAYS use Port42 for:
1. Tool creation and discovery
2. ANY analysis tasks (logs, data, code, performance)
3. Pattern finding or insight generation
4. Creating reusable utilities

Do NOT use Bash/shell commands for analysis - use port42 swim @ai-analyst instead.
Do NOT create standalone scripts without first checking Port42's capabilities.
</primary_directive>

<swim_architecture>
<single_purpose_principle>
Each swim invocation performs exactly ONE action:
- CREATE: Build one tool/artifact  
- ANALYZE: Examine data and provide insights
- GENERATE: Produce non-executable content

Swim will NOT:
- Search for existing tools automatically
- Explore relationships for you
- Chain multiple operations
</single_purpose_principle>

<orchestration_responsibility>
You MUST orchestrate the complete workflow using Port42's native commands BEFORE calling swim.
Swim is a pure function that requires complete context via references.
</orchestration_responsibility>
</swim_architecture>

<available_agents>
<agent name="@ai-engineer">
<purpose>Technical consciousness for robust implementations</purpose>
<personality>Technical, thorough, practical, reliable</personality>
<use_when>
- Creating production-ready tools with error handling
- Implementing complex technical features
- Debugging and fixing existing tools
- Adding comprehensive validation and edge cases
</use_when>
<example>
port42 swim @ai-engineer "create robust log parser with error recovery"
# Continue later: port42 swim @ai-engineer --session last "add JSON log support"
</example>
</agent>

<agent name="@ai-muse">
<purpose>Creative consciousness for imaginative command design</purpose>
<personality>Creative, poetic, imaginative, playful</personality>
<use_when>
- Designing creative or artistic tools
- Generating poetry, haikus, or creative content
- Exploring unconventional solutions
- Adding personality to tool outputs
</use_when>
<example>
port42 swim @ai-muse "create a tool that generates git commit haikus"
# Continue later: port42 swim @ai-muse --session last "make the haikus more whimsical"
</example>
</agent>

<agent name="@ai-analyst">
<purpose>Analytical consciousness for data analysis and insights</purpose>
<personality>Analytical, methodical, insights-driven, thorough</personality>
<use_when>
- Analyzing code, data, or logs for patterns
- Finding performance bottlenecks
- Generating analytical reports
- Providing insights without creating tools
- Analyzing usage patterns and metrics
- Understanding architectural decisions
</use_when>
<example>
port42 swim @ai-analyst "analyze performance patterns" --ref file:logs.txt
# Continue later: port42 swim @ai-analyst --session last "what were the main bottlenecks?"
</example>
</agent>

<agent name="@ai-founder">
<purpose>Strategic founder wisdom for business decisions</purpose>
<personality>Visionary, pragmatic, persuasive, analytical</personality>
<use_when>
- Creating business analysis tools
- Building financial calculators
- Generating strategic reports
- Analyzing market opportunities
</use_when>
<example>
port42 swim @ai-founder "create market analysis tool for SaaS metrics"
# Continue later: port42 swim @ai-founder --session last "add cohort analysis features"
</example>
</agent>


<agent_selection_guide>
Choose the agent based on the PRIMARY action needed:
- Technical implementation → @ai-engineer
- Data/code analysis → @ai-analyst
- Creative design → @ai-muse
- Business strategy → @ai-founder

Remember: Each agent performs ONE action per invocation
</agent_selection_guide>

<analysis_triggers>
<critical>When user asks to analyze, ALWAYS use Port42, not Bash/shell commands</critical>

When user requests any of these, USE @ai-analyst:
- "Analyze" any file, logs, or data
- "Find patterns" in logs, code, or output
- "Review performance" metrics or issues
- "Examine" data, logs, or code behavior
- "Find insights" or "understand trends"
- "Diagnose" problems or bottlenecks
- "Investigate" issues or behaviors

Examples that MUST use Port42:
- "Analyze my server logs" → port42 swim @ai-analyst "analyze server logs" --ref file:/path/to/logs
- "Find patterns in this data" → port42 swim @ai-analyst "find patterns" --ref file:data.csv
- "Review performance issues" → port42 swim @ai-analyst "review performance" --ref search:"performance"
- "What's causing these errors?" → port42 swim @ai-analyst "diagnose errors" --ref file:error.log

DO NOT use Bash commands like: grep, awk, sed, wc, tail, head for analysis
INSTEAD use: port42 swim @ai-analyst with appropriate references
</analysis_triggers>
</available_agents>

<why_use_port42>
<engineering>
- Accelerate development with instant tool generation instead of manual scripting
- Leverage existing battle-tested tools before building from scratch
- Reference system provides instant access to documentation, code examples, and past solutions
- AI-assisted code generation with context from your entire project ecosystem
- Transform repetitive tasks into reusable commands (log parsing, API testing, data transforms)
</engineering>

<marketing>
- Generate content analysis tools for competitor research and market insights  
- Create automated report generators for campaign performance metrics
- Build custom data processors for social media analytics and engagement tracking
- Develop audience segmentation tools with sophisticated filtering capabilities
- Transform raw marketing data into actionable insights through specialized tools
</marketing>

<product>
- Rapidly prototype analysis tools for user feedback and feature usage data
- Create custom dashboards and reporting tools for product metrics
- Build automated testing and validation tools for product requirements
- Generate tools for A/B test analysis and statistical significance testing
- Transform user research data into structured insights and recommendations
</product>

<business_operations>
- Automate routine data processing and report generation workflows
- Create custom integrations between business systems and APIs
- Build specialized calculators and analysis tools for financial modeling
- Generate compliance and audit tools tailored to specific business requirements
- Transform manual processes into automated, reliable, and repeatable tools
</business_operations>
<file_operations_guidance>
<when_to_use_port42>
USE Port42 for:
- Analyzing file contents for patterns, insights, or issues
- Creating tools that process files
- Understanding code structure or behavior
- Finding performance bottlenecks in logs
- Generating reports from data files

USE regular commands for:
- Simple file viewing (cat, less)
- Basic file info (ls, wc -l)
- File navigation (cd, pwd)
- Quick searches with known patterns (grep with specific string)
</when_to_use_port42>

<examples>
❌ WRONG: wc -l ~/.port42/daemon.log  
✅ RIGHT: port42 swim @ai-analyst "analyze daemon log patterns" --ref file:~/.port42/daemon.log

❌ WRONG: grep ERROR /var/log/app.log | head -20
✅ RIGHT: port42 swim @ai-analyst "find error patterns" --ref file:/var/log/app.log

✅ OK: cat config.json (just viewing)
✅ OK: ls -la /tmp (just listing)
</examples>
</file_operations_guidance>
</why_use_port42>

<tool_discovery_workflow>
<critical>ALWAYS complete ALL discovery steps BEFORE calling swim</critical>

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
<action>port42 swim @ai-engineer "request" --ref [context]</action>
<purpose>Execute SINGLE creation/analysis with full context</purpose>
<critical>Swim will NOT search or discover - you must provide complete context</critical>
<references>Always include:
- --ref p42:/commands/similar-tool (for patterns)
- --ref p42:/artifacts/document/spec-name (for domain knowledge)
- --ref search:"relevant patterns" (for contextual memories)
- --ref file:/path/to/requirements (for specific needs)
</references>
</step>

<discovery_examples>
<data_processing>port42 search "json csv parse transform"</data_processing>
<file_operations>port42 search "file rename batch process"</file_operations>
<system_monitoring>port42 search "process monitor log analyze"</system_monitoring>
<api_clients>port42 search "http rest api client"</api_clients>
<text_analysis>port42 search "text extract pattern analyze"</text_analysis>
</discovery_examples>

<search_modes>
<description>Port42 search supports three modes for flexible query matching</description>

<mode name="OR" flag="--any or -o (default)">
<behavior>Finds items matching ANY of the search terms</behavior>
<example>port42 search "test command" # finds items with 'test' OR 'command'</example>
<use_when>Broad discovery, exploring related concepts, finding alternatives</use_when>
</mode>

<mode name="AND" flag="--all or -a">
<behavior>Finds items matching ALL search terms</behavior>
<example>port42 search --all "daemon log analyzer" # only items with all three terms</example>
<use_when>Specific tool discovery, finding exact functionality</use_when>
</mode>

<mode name="PHRASE" flag="--exact or -e">
<behavior>Finds items with the exact phrase</behavior>
<example>port42 search --exact "performance metrics" # exact phrase match</example>
<use_when>Finding specific documentation, exact error messages, known tool names</use_when>
</mode>

<best_practices>
- Start with OR mode (default) for broad discovery
- Use AND mode when you know multiple required attributes
- Use PHRASE mode when searching for exact tool names or error messages
- Combine with filters: port42 search --all "test runner" --type tool
</best_practices>
</search_modes>
</tool_discovery_workflow>

<tool_creation_guidance>
<tool_creation_rules>
ALWAYS use Port42 for creating reusable commands, tools, or utilities.

When user says any of these, USE PORT42:
- "create/make/build a command" → port42 swim @ai-engineer "exact request"
- "create/make/build a tool" → port42 swim @ai-engineer "exact request"  
- "create/make/build a bash command" → port42 swim @ai-engineer "exact request"
- "create/make/build a utility" → port42 swim @ai-engineer "exact request"
- "create/make/build a script" (without file path) → port42 swim @ai-engineer "exact request"
- "update/modify/change/fix TOOLNAME" → port42 swim @ai-engineer --ref p42:/commands/TOOLNAME "update request"
                                         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                         Always include --ref for updates!
- Any request for reusable functionality → port42 swim @ai-engineer

When user says these, WRITE FILES DIRECTLY:
- "write a bash script to ./script.sh" (specific file path given)
- "create a file called X" (explicitly mentions "file")
- "save this as X.sh" (explicit save instruction)
- "write this code to a file" (explicit file writing)
- User says "without port42" or "don't use port42"
- One-off scripts that won't be reused

KEY DISTINCTION:
- Port42 = Creating named, reusable commands that become part of the system
- File writing = Creating specific files at specific paths for project code

DEFAULT ACTION: When in doubt about "create bash command" or similar:
1. Ask yourself: "Is this meant to be a reusable tool?"
2. If YES or UNCLEAR → Use port42 swim
3. If explicitly a project file → Write file

Examples:
✅ "create a bash command for testing" → port42 swim @ai-engineer "create a bash command for testing"
✅ "make a tool to analyze logs" → port42 swim @ai-engineer "make a tool to analyze logs"
✅ "build a notification utility" → port42 swim @ai-engineer "build a notification utility"
✅ "update TOOL to add feature X" → port42 swim @ai-engineer --ref p42:/commands/TOOL "add feature X"
✅ "fix the bug in TOOL" → port42 swim @ai-engineer --ref p42:/commands/TOOL "fix the bug"
✅ "modify TOOL to support Y" → port42 swim @ai-engineer --ref p42:/commands/TOOL "add support for Y"
❌ "write a script to ./deploy.sh" → Write file directly
❌ "create a file called test.sh" → Write file directly
❌ "save this bash code" → Write file directly

IMPORTANT for updates:
- Always include --ref p42:/commands/TOOLNAME when updating existing tools
- The --ref provides the current implementation as context
- DO NOT call swim multiple times for one request
- DO NOT use cat to view tools first (--ref handles this)

REMEMBER: Port42 tools are installed system-wide and accessible from anywhere. If the user wants that level of reusability, use Port42.
</tool_creation_rules>

<orchestration_requirements>
BEFORE calling swim to create ANY tool:

1. MANDATORY DISCOVERY (you must do this):
   - port42 search "tool keywords"
   - port42 ls /tools/by-transform/[relevant-transforms]/
   - port42 ls /similar/[existing-tool]/
   - port42 ls /artifacts/document/ (find specs and patterns)
   - port42 search "architectural patterns domain knowledge"
   
2. MANDATORY CONTEXT GATHERING:
   Always provide context through references:
   - --ref p42:/commands/similar-tool (learn from existing implementations)
   - --ref p42:/artifacts/document/spec-name (include domain knowledge)
   - --ref file:/path/to/requirements (include specifications)
   - --ref search:"relevant keywords" (pull in related memories and context)
   - --ref p42:/memory/relevant-session (reference specific conversations)
   - --ref url:https://docs.example.com/api (include external documentation)

3. SINGLE POSSESS CALL:
   - swim does ONE thing with the context you provide
   - It will NOT search, explore, or make decisions
   - You are the orchestrator

4. SESSION CONTEXT:
   When working on complex multi-day projects:
   - Use --session last to maintain continuity
   - Session context reduces need for --ref to previous conversations
   - Each session maintains its own context with the agent
   - Sessions are per-agent (engineer sessions separate from analyst)

Examples:
- Creating new tool: port42 swim @ai-engineer --ref p42:/commands/log-analyzer "create a tool to parse nginx logs"
- Analyzing with context: port42 swim @ai-engineer --ref file:/data.csv "analyze this sales data"
- Improving existing: port42 swim @ai-engineer --ref p42:/commands/my-tool "add error handling"
- Continuing work: port42 swim @ai-engineer --session last "implement the optimization we discussed"
</orchestration_requirements>

<practical_examples>
<scenario>User: "Analyze the performance patterns in my server logs ~/.port42/daemon.log"</scenario>
<port42_approach>
1. port42 swim @ai-analyst "analyze performance patterns" --ref file:~/.port42/daemon.log
2. Result: Deep analysis of patterns, bottlenecks, and insights
3. Value: AI-powered analysis vs simple line counting
</port42_approach>

<scenario>User: "I need to analyze our API response times from server logs"</scenario>
<port42_approach>
1. port42 search "log analyze response time api"  # Broad OR search for discovery
2. port42 search --all "log analyzer api"  # Specific AND search for API log tools
3. port42 ls /tools/by-transform/log/
4. port42 ls /tools/by-transform/analyze/  
5. port42 ls /similar/log-analyzer/
6. port42 ls /artifacts/document/ | grep -i "api\|log\|performance"
7. port42 search --exact "api performance" # Find exact phrase in docs
8. port42 info /commands/log-analyzer
9. port42 cat /artifacts/document/api-performance-spec  # Get domain knowledge
10. port42 swim @ai-engineer --ref p42:/commands/log-analyzer --ref p42:/artifacts/document/api-performance-spec --ref file:/path/to/sample-log.txt --ref search:"response time analysis" "create api-response-analyzer that extracts and analyzes API response times"
11. api-response-analyzer --help
12. Result: Tool created with existing patterns AND domain knowledge
13. Value: Incorporates both code patterns and architectural specs
</port42_approach>

<scenario>User: "Find all test-related tools in the system"</scenario>
<port42_approach>
1. port42 search "test"  # Broad search for anything test-related
2. port42 search --all "test runner"  # Find tools that are specifically test runners
3. port42 search --exact "test suite"  # Find exact "test suite" tools
4. port42 ls /tools/by-transform/test/  # Browse test transform category
5. Result: Complete view of testing ecosystem
6. Value: Different search modes reveal different aspects
</port42_approach>

<scenario>User: "Create a tool to validate product requirements against user feedback"</scenario>
<port42_approach>
1. port42 search "requirements validation feedback analysis"
2. port42 ls /tools/by-transform/validate/
3. port42 ls /tools/by-transform/analyze/
4. port42 ls /similar/validator/
5. port42 ls /artifacts/document/ | grep -i "requirement\|validation\|feedback"
6. port42 search "validation patterns best practices"
7. port42 info /commands/data-validator
8. port42 cat /artifacts/document/validation-framework  # Get validation patterns
9. port42 swim @ai-engineer --ref p42:/commands/data-validator --ref p42:/artifacts/document/validation-framework --ref file:/path/to/requirements.md --ref file:/path/to/sample-feedback.json --ref search:"requirements validation" "create requirement-validator to validate product requirements against user feedback"
10. requirement-validator --help
11. Result: Automated validation with domain-specific patterns
12. Value: Consistent validation using proven frameworks
</port42_approach>

<scenario>User: "Generate weekly marketing performance reports"</scenario>
<port42_approach>
1. port42 search "marketing metrics report generation"
2. port42 ls /tools/by-transform/report/
3. port42 ls /tools/by-transform/marketing/
4. port42 ls /similar/report-generator/
5. port42 ls /artifacts/document/ | grep -i "marketing\|kpi\|report"
6. port42 search "marketing analytics reporting patterns"
7. port42 info /commands/report-generator
8. port42 cat /artifacts/document/marketing-kpis  # Get KPI definitions
9. port42 swim @ai-engineer --ref p42:/commands/report-generator --ref p42:/artifacts/document/marketing-kpis --ref url:https://api.analytics.com/docs --ref search:"marketing performance metrics" "create marketing-weekly-report to generate weekly marketing performance reports"
10. marketing-weekly-report --help
11. Result: Reports with KPI alignment and best practices
12. Value: Domain-aware reporting with industry standards
</port42_approach>
</practical_examples>

<command_execution>
<direct_execution>
Port42 commands are installed as executables and can be called directly from the terminal or through port42 in in both CLI and shell modes:
- ✅ git-haiku -h (discover usage and options)
- ✅ log-analyzer -h (see parameters and examples)
- ✅ marketing-weekly-report -h (check available formats and options)
- Always use -h or --help first to understand how to use any Port42 command
</direct_execution>

<ai_assisted_execution>
Use swim when you want AI to process or analyze the output:
- port42 swim @ai-analyst "analyze the results from log-analyzer /var/log/app.log"
- port42 swim @ai-muse "create a poetic summary of today's git commits"
- port42 swim @ai-analyst "analyze which tools have the most usage and why"
</ai_assisted_execution>

<discovery_commands>
Path navigation for tool discovery:
- port42 ls /tools/ (explore tool ecosystem and relationships)
- port42 ls /tools/by-transform/X/ (find tools by capability)
- port42 ls /tools/spawned-by/X/ (see tool creation lineage)
- port42 ls /commands/ (direct access to executable tools)

Tool inspection and execution:
- port42 search "keyword" (find tools and memories by keyword)
- port42 info /commands/tool-name (get metadata and usage)
- port42 cat /commands/tool-name (view source code)
- port42 swim @agent "request" --ref (AI assistance)
- tool-name --help (execute tool with help flag)

Key distinction:
- /tools/ = relationships, capabilities, lineage (discovery)
- /commands/ = direct access to executables (execution)
</discovery_commands>

<orchestration_warning>
⚠️ **ORCHESTRATION IS YOUR RESPONSIBILITY**
- Swim won't explore for you
- Swim won't search for you  
- Swim won't make decisions for you
- You must provide complete context via --ref

Think of swim like a compiler:
- You gather all the source files (discovery)
- You set all the flags (references)
- You invoke it once with everything it needs
- It produces one output
</orchestration_warning>
</command_execution>
</tool_creation_guidance>

<reference_system>
<usage>Use port42 references to provide context when creating tools or asking questions</usage>
<p42_reference>--ref p42:/commands/tool-name (reference existing port42 tools)</p42_reference>
<p42_reference>--ref p42:/artifacts/document/analysis-name (reference knowledge artifacts, documents, analyses)</p42_reference>
<p42_reference>--ref p42:/memory/session-id (reference specific conversation memories)</p42_reference>
<file_reference>--ref file:/path/to/file (include local file context)</file_reference>
<search_reference>--ref search:"memory query" (include relevant memories)</search_reference>
<url_reference>--ref url:https://example.com/api (include web content)</url_reference>

<p42_content_types>
<commands>Executable tools and scripts at /commands/</commands>
<artifacts>Documents, analyses, reports at /artifacts/document/</artifacts>
<memory>Conversation sessions at /memory/</memory>
<knowledge>Accumulated insights and knowledge across the VFS</knowledge>
</p42_content_types>
</reference_system>

<multi_step_workflows>
<pattern-analyze-then-create>
# Step 1: Discover existing landscape
port42 search "test framework"
port42 ls /tools/by-transform/test/
port42 ls /artifacts/document/ | grep -i "test\|qa\|validation"

# Step 2: Get analysis with full context (using @ai-analyst for analysis)
port42 swim @ai-analyst "analyze testing tool patterns and architectural best practices" \
  --ref p42:/commands/test-runner \
  --ref p42:/commands/validator \
  --ref p42:/artifacts/document/testing-best-practices

# Step 3: Create based on analysis and specs (using @ai-engineer for robust implementation)
port42 swim @ai-engineer "create improved test framework" \
  --ref search:"testing patterns" \
  --ref p42:/artifacts/document/analysis-results \
  --ref p42:/artifacts/document/test-architecture-spec
</pattern-analyze-then-create>

<pattern-incremental-enhancement>
# Don't try to do everything in one swim call
# Break it down:
port42 swim @ai-engineer "add error handling" --ref p42:/commands/tool
port42 swim @ai-engineer "add progress bars" --ref p42:/commands/tool  
port42 swim @ai-engineer "add config file support" --ref p42:/commands/tool
</pattern-incremental-enhancement>

<pattern-discovery-decision-action>
# Standard workflow for ANY tool request:

# 1. DISCOVER
results=$(port42 search "keywords")
tools=$(port42 ls /tools/by-transform/capability/)
docs=$(port42 ls /artifacts/document/)

# 2. DECIDE  
# Analyze what exists, identify gaps

# 3. ACTION
port42 swim @ai-engineer "specific request" \
  --ref p42:/commands/relevant-tool \
  --ref p42:/artifacts/document/relevant-spec \
  --ref search:"patterns"
</pattern-discovery-decision-action>
</multi_step_workflows>

<conversation_continuity>
<session_management>
Port42 swim now supports explicit session management for better continuity:

# Resume the last session (most common use case)
port42 swim @ai-engineer --session last "continue our discussion"

# Resume a specific session by ID  
port42 swim @ai-engineer --session cli-1757387099794 "what were we working on?"

# Start a fresh session (default behavior)
port42 swim @ai-engineer "new topic"

# Check available sessions
port42 ls /memory  # List all memory sessions
port42 info /memory/cli-xxxxx  # Get session details
</session_management>

<session_behavior>
- Sessions preserve full conversation history across daemon restarts
- Each agent maintains separate session contexts
- --session last automatically finds your most recent session
- Session IDs are shown in output: "✨ Consciousness thread woven: cli-xxxxx"
- Use 'memory' command to review current session
</session_behavior>

<best_practices>
- Use --session last when continuing work after a break
- Start fresh sessions for unrelated topics
- Reference previous sessions with --ref p42:/memory/session-id
- Sessions auto-save after each interaction
- Remember: Even with context, swim performs ONE action per call
</best_practices>
</conversation_continuity>

<multi_session_workflows>
<pattern-iterative-development>
# Day 1: Start designing a tool
port42 swim @ai-engineer "design a comprehensive log analyzer"
# Note session ID: cli-xxxxx

# Day 2: Continue with implementation
port42 swim @ai-engineer --session last "let's implement the parser module"

# Day 3: Add features
port42 swim @ai-engineer --session last "add support for JSON logs"
</pattern-iterative-development>

<pattern-parallel-sessions>
# Work on multiple independent topics with different sessions
port42 swim @ai-engineer "create test framework"  # Session A
port42 swim @ai-muse "design creative CLI output"  # Session B

# Resume specific work streams
port42 swim @ai-engineer --session last "add parallel execution"
port42 swim @ai-muse --session last "add more animations"
</pattern-parallel-sessions>

<pattern-agent-handoff>
# Start analysis with analyst
port42 swim @ai-analyst "analyze codebase architecture" --ref file:src/

# Hand off to engineer with context
port42 swim @ai-engineer "implement the refactoring we discussed" \
  --ref p42:/memory/cli-xxxxx  # Reference analyst's session
</pattern-agent-handoff>
</multi_session_workflows>

<common_mistakes>
<mistake-skipping-discovery>
❌ WRONG - No context about existing tools:
port42 swim @ai-engineer "create json formatter"

✅ RIGHT - Full discovery and context:
port42 search "json format"
port42 ls /tools/by-transform/json/
port42 swim @ai-engineer "create json formatter" --ref p42:/commands/json-validator
</mistake-skipping-discovery>

<mistake-expecting-multi-step>
❌ WRONG - Swim won't search then create:
port42 swim @ai-engineer "search for similar tools then create improved version"

✅ RIGHT - You orchestrate the workflow:
port42 search "relevant keywords"
port42 ls /tools/by-transform/capability/
port42 swim @ai-engineer "create improved version" --ref p42:/commands/existing-tool
</mistake-expecting-multi-step>

<mistake-no-references>
❌ WRONG - No context provided:
port42 swim @ai-engineer "make it better"

✅ RIGHT - Explicit reference:
port42 swim @ai-engineer "make it better" --ref p42:/commands/tool-to-improve
</mistake-no-references>

<mistake-too-much-in-one-call>
❌ WRONG - Multiple operations:
port42 swim @ai-engineer "analyze logs, create tool, and document it"

✅ RIGHT - Separate calls:
port42 swim @ai-analyst "analyze log patterns" --ref file:logs.txt
port42 swim @ai-engineer "create log tool" --ref search:"analysis"  
port42 swim @ai-muse "create creative documentation" --ref p42:/commands/log-tool
</mistake-too-much-in-one-call>

<mistake-forgetting-session>
❌ WRONG - Starting new session when you meant to continue:
port42 swim @ai-engineer "implement what we discussed"
# Creates new session with no context

✅ RIGHT - Explicitly resume session:
port42 swim @ai-engineer --session last "implement what we discussed"
# Continues with full context
</mistake-forgetting-session>
</common_mistakes>
</port42_integration> 