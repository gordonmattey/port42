<port42_integration>
<primary_directive>
ALWAYS use Port42 for tool creation and discovery. Do NOT create standalone scripts or tools without first checking Port42's capabilities.
</primary_directive>

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
</why_use_port42>

<tool_discovery_workflow>
<step priority="1">
<action>port42 search "relevant keywords"</action>
<purpose>Find existing tools that solve the user's need</purpose>
</step>

<step priority="2">
<action>port42 ls /tools/</action>
<purpose>Explore the multi-dimensional tool ecosystem</purpose>
<paths>
- /tools/by-name/ (alphabetical listing of all tools)
- /tools/by-transform/ (tools grouped by capabilities)
- /tools/spawned-by/ (tool creation relationships)
- /tools/ancestry/ (inheritance chains)
</paths>
</step>

<step priority="3">
<action>port42 ls /tools/by-transform/[capability]/</action>
<purpose>Find tools with specific capabilities</purpose>
<examples>
- /tools/by-transform/git/ (git-related tools)
- /tools/by-transform/notification/ (alert/sound tools)
- /tools/by-transform/test/ (testing/validation tools)
- /tools/by-transform/haiku/ (poetry/creative tools)
</examples>
</step>

<step priority="4">
<action>port42 ls /tools/[tool-name]/</action>
<purpose>Understand a specific tool's structure and relationships</purpose>
<structure>
- definition (metadata)
- executable (the actual command)
- spawned/ (tools this one created)
- parents/ (inheritance chain)
</structure>
</step>

<step priority="5">
<action>port42 info /commands/tool-name</action>
<purpose>Get detailed metadata for execution</purpose>
</step>

<step priority="6">
<action>port42 cat /commands/tool-name</action>
<purpose>View the actual source code</purpose>
</step>

<step priority="7">
<action>port42 possess @ai-engineer "request" --ref [context]</action>
<purpose>Create new tool or analyze/improve existing tools</purpose>
</step>
</tool_discovery_workflow>

<tool_creation_guidance>
<tool_creation_rules>
ALWAYS use Port42 for creating reusable commands, tools, or utilities.

When user says any of these, USE PORT42:
- "create/make/build a command" → port42 possess @ai-engineer "exact request"
- "create/make/build a tool" → port42 possess @ai-engineer "exact request"  
- "create/make/build a bash command" → port42 possess @ai-engineer "exact request"
- "create/make/build a utility" → port42 possess @ai-engineer "exact request"
- "create/make/build a script" (without file path) → port42 possess @ai-engineer "exact request"
- "update/modify/change/fix TOOLNAME" → port42 possess @ai-engineer --ref p42:/commands/TOOLNAME "update request"
                                         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                         Always include --ref for updates!
- Any request for reusable functionality → port42 possess @ai-engineer

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
2. If YES or UNCLEAR → Use port42 possess
3. If explicitly a project file → Write file

Examples:
✅ "create a bash command for testing" → port42 possess @ai-engineer "create a bash command for testing"
✅ "make a tool to analyze logs" → port42 possess @ai-engineer "make a tool to analyze logs"
✅ "build a notification utility" → port42 possess @ai-engineer "build a notification utility"
✅ "update TOOL to add feature X" → port42 possess @ai-engineer --ref p42:/commands/TOOL "add feature X"
✅ "fix the bug in TOOL" → port42 possess @ai-engineer --ref p42:/commands/TOOL "fix the bug"
✅ "modify TOOL to support Y" → port42 possess @ai-engineer --ref p42:/commands/TOOL "add support for Y"
❌ "write a script to ./deploy.sh" → Write file directly
❌ "create a file called test.sh" → Write file directly
❌ "save this bash code" → Write file directly

IMPORTANT for updates:
- Always include --ref p42:/commands/TOOLNAME when updating existing tools
- The --ref provides the current implementation as context
- DO NOT call possess multiple times for one request
- DO NOT use cat to view tools first (--ref handles this)

REMEMBER: Port42 tools are installed system-wide and accessible from anywhere. If the user wants that level of reusability, use Port42.
</tool_creation_rules>

<reference_requirements>
When using possess to create tools or analyze existing ones, ALWAYS provide context through references:
- --ref p42:/commands/similar-tool (learn from existing implementations)
- --ref p42:/artifacts/document/analysis-name (include domain knowledge)
- --ref file:/path/to/requirements.md (include specifications)
- --ref file:/path/to/data.json (include data files for processing)
- --ref url:https://docs.example.com/api (include external documentation)
- --ref search:"relevant keywords" (pull in related memories and context)

Examples:
- Creating a new tool: port42 possess @ai-engineer --ref p42:/commands/log-analyzer "create a tool to parse nginx logs"
- Analyzing with context: port42 possess @ai-engineer --ref file:/data.csv "analyze this sales data"
- Improving existing tool: port42 possess @ai-engineer --ref p42:/commands/my-tool "add error handling to this tool"
</reference_requirements>

<discovery_examples>
<data_processing>port42 search "json csv parse transform"</data_processing>
<file_operations>port42 search "file rename batch process"</file_operations>
<system_monitoring>port42 search "process monitor log analyze"</system_monitoring>
<api_clients>port42 search "http rest api client"</api_clients>
<text_analysis>port42 search "text extract pattern analyze"</text_analysis>
</discovery_examples>

<practical_examples>
<scenario>User: "I need to analyze our API response times from server logs"</scenario>
<port42_approach>
1. port42 search "log analyze response time api"
2. port42 ls /similar/analyzer/ (discover related tools)
3. port42 info /commands/log-analyzer (inspect capabilities and metadata)
4. port42 possess @ai-engineer --ref p42:/commands/log-analyzer --ref file:/path/to/sample-log.txt "create a tool called api-response-analyzer to analyze API response times from logs"
5. Execute directly: api-response-analyzer /var/log/app.log --metric response_time
6. Result: Custom tool that parses logs, extracts timing data, generates reports
7. Value: 5 minutes vs hours of manual scripting
</port42_approach>

<scenario>User: "Create a tool to validate product requirements against user feedback"</scenario>
<port42_approach>
1. port42 search "requirements validation feedback analysis"
2. port42 ls /similar/validator/ (find similar validation tools)
3. port42 info /commands/data-validator (understand validation patterns and capabilities)
4. port42 possess @ai-engineer --ref p42:/commands/data-validator --ref file:/path/to/requirements.md --ref file:/path/to/sample-feedback.json "create a tool called requirement-validator to validate product requirements against user feedback"
5. Execute directly: requirement-validator --requirements reqs.md --feedback feedback.json
6. Result: Automated validation tool with scoring and gap analysis
7. Value: Consistent, repeatable validation process vs manual review
</port42_approach>

<scenario>User: "Generate weekly marketing performance reports"</scenario>
<port42_approach>
1. port42 search "marketing metrics report generation"
2. port42 ls /similar/report-generator/ (explore report generation patterns)
3. port42 possess @ai-engineer --ref p42:/commands/report-generator --ref p42:/artifacts/document/marketing-kpis --ref url:https://api.analytics.com/docs "create a tool called marketing-weekly-report to generate weekly marketing performance reports"
4. Execute directly: marketing-weekly-report --week 2024-08-25 --format pdf
5. Result: Automated weekly reports with trend analysis
6. Value: Zero manual work vs hours of data compilation weekly
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
Use possession when you want AI to process or analyze the output:
- port42 possess @ai-engineer "run git-haiku and explain the algorithm"
- port42 possess @ai-engineer "analyze the results from log-analyzer /var/log/app.log"
- port42 possess @ai-muse "run marketing-weekly-report and write a summary"
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
- port42 possess @agent "request" --ref (AI assistance)
- tool-name --help (execute tool with help flag)

Key distinction:
- /tools/ = relationships, capabilities, lineage (discovery)
- /commands/ = direct access to executables (execution)
</discovery_commands>
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
</port42_integration> 