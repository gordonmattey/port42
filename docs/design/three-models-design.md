# Port 42: Three Models Design

**Purpose**: High-level conceptual overview of the three models and how they map to user workflows.
**Scope**: Mental models, use cases, and examples. No implementation details.

## The Three Models

### Model 1: Tool Creation (Commands)
- **What**: Executable commands that transform input → output
- **When**: Reusable operations, automations, utilities
- **Storage**: Executable files in ~/.port42/commands/
- **Example**: `git-haiku`, `pr-writer`, `code-analyzer`

### Model 2: Living Documents (Structured Data + CRUD)
- **What**: Commands that manage structured, evolving data
- **When**: Ongoing projects, tracking systems, databases
- **Storage**: JSON/YAML files with command as interface
- **Example**: `content-plan`, `feature-tracker`, `investor-crm`

### Model 3: Artifacts (Any Digital Asset)
- **What**: Any file type - documents, code, designs, media, PDFs, presentations, diagrams
- **When**: Strategy docs, web apps, design mockups, video scripts, architecture diagrams
- **Storage**: Organized by type in ~/.port42/artifacts/
- **Example**: Pitch decks, React apps, Figma exports, logos, PDFs, demo videos, flowcharts

## Your Three Core Workflows

### 1. Product & Engineering Workflow

**Conversation Mode**:
```bash
possess @ai-architect "Let's design the real-time sync system"
# Explores architecture, discusses tradeoffs
# When ready: /crystallize artifact → creates `sync-system-design.md`
# Or: /crystallize artifact → creates full `realtime-sync-demo/` web app
```

**App Creation Mode**:
```bash
possess @ai-engineer "Build a dashboard to visualize our metrics"
# Discusses requirements, data sources, UI preferences
# /crystallize artifact → creates complete `metrics-dashboard/` React app
# Auto-generates: run-metrics-dashboard command to launch it
```

**Tool Mode**:
```bash
possess @ai-engineer "Create a command to analyze our API performance"
# /crystallize → creates `api-analyzer` command
```

**Data Management Mode**:
```bash
possess @ai-pm "Help me track feature requests"
# /crystallize data → creates `feature-tracker` command with CRUD operations
```

### 2. Marketing Workflow

**Content Calendar** (Living Document):
```bash
possess @ai-growth "Let's build a content strategy system"
# Creates `content-calendar` command that manages:
# - Blog posts with status (draft/published)
# - Social media campaigns
# - Metrics tracking
```

**Build in Public** (Tool):
```bash
possess @ai-marketer "Create a command that generates weekly updates"
# Creates `weekly-update` that:
# - Pulls git commits
# - Summarizes progress
# - Formats for Twitter/LinkedIn
```

**Brand Voice** (Knowledge Artifact):
```bash
possess @ai-muse "Help define Port 42's voice and messaging"
# Conversation about brand
# /crystallize artifact → creates `brand-guide.md`
# /crystallize artifact → creates logo concepts and color palettes
```

**Design Assets** (New Possibilities):
```bash
possess @ai-muse "Design a logo for Port 42"
# /crystallize → creates:
# - Logo concept descriptions
# - Midjourney/DALL-E prompts
# - SVG code for simple geometric logos
# - Color palette specifications

possess @ai-engineer "Create architecture diagram for our system"
# /crystallize → generates:
# - Mermaid/diagram-as-code files
# - ASCII art diagrams
# - SVG flowcharts
```

### 3. Fundraising Workflow

**Pitch Evolution** (Knowledge Artifact):
```bash
possess @ai-founder "Let's refine the pitch deck"
# Iterative conversation
# /crystallize → creates/updates `pitch-deck-v3.md`
```

**Investor CRM** (Living Document):
```bash
possess @ai-fundraiser "I need to track investor conversations"
# Creates `investor-tracker` command:
# - Add investors with tags
# - Track conversation status
# - Note feedback
```

**Market Analysis** (Tool):
```bash
possess @ai-analyst "Create a competitor analysis tool"
# Creates `market-scan` command that:
# - Pulls competitor data
# - Generates comparison charts
# - Tracks changes over time
```

## Implementation Approach

### 1. Extend `/crystallize` with Options

```bash
/crystallize              # AI decides based on context
/crystallize command      # Force command creation
/crystallize artifact     # Force artifact creation (any file type)
/crystallize data         # Force data management tool
```

### 2. How It Works

**Command Generation (Current)**
- AI detects command specification in conversation
- Generates executable script in `~/.port42/commands/`
- Immediately available in PATH

**Artifact Generation (New)**
- AI creates any type of digital asset based on conversation
- Saves to `~/.port42/artifacts/{type}/{date}-{name}/`
- Types include:
  - **Documents**: Markdown, PDFs, presentations
  - **Code**: Full apps, prototypes, examples
  - **Designs**: Mockups, diagrams, logos, UI concepts
  - **Media**: Scripts, screenshots, video concepts
- Could auto-generate viewer commands: `view-{name}`, `open-{name}`
- AI can even generate prompts for other AI tools (Midjourney, DALL-E, etc.)

**Data Management Generation (New)**
- AI creates CRUD command based on schema
- Command manages JSON data in `~/.port42/data/`
- Operations: create, list, update, delete, search
- Supports both bash and Python implementations

### 3. Storage Structure

```
~/.port42/
├── commands/          # Executable tools
│   ├── git-haiku      # Transform tool
│   ├── content-plan   # Data management tool
│   └── view-pitch     # Document viewer
├── artifacts/         # Markdown documents & code
│   ├── decisions/
│   │   └── 2024-01-15-api-architecture.md
│   ├── strategies/
│   │   └── 2024-01-15-growth-plan.md
│   ├── code/
│   │   ├── 2024-01-15-dashboard-app/
│   │   │   ├── package.json
│   │   │   ├── index.html
│   │   │   ├── src/
│   │   │   └── README.md
│   │   └── 2024-01-15-api-prototype/
│   │       ├── server.py
│   │       ├── requirements.txt
│   │       └── README.md
│   ├── designs/
│   │   ├── 2024-01-15-logo-concepts.pdf
│   │   ├── 2024-01-15-ui-mockups.fig
│   │   └── 2024-01-15-architecture-diagram.svg
│   ├── media/
│   │   ├── 2024-01-15-demo-video-script.md
│   │   ├── 2024-01-15-product-screenshots/
│   │   └── 2024-01-15-presentation.pdf
│   └── index.json     # Rich metadata & search index (see artifact-metadata-system.md)
├── data/             # Structured data for CRUD commands
│   ├── content-plan.json
│   ├── feature-tracker.json
│   └── investor-crm.json
└── sessions/         # Conversation history
```


## Design Principles

1. **Simple Mental Model**: Three clear categories based on interaction type
2. **Tool Gating**: Prevent accidental artifact creation
3. **Agent Specialization**: Different agents prefer different output types
4. **Metadata First**: Every artifact is indexed and searchable

## Key Technical Insights

1. **Three storage patterns**:
   - Commands → `~/.port42/commands/` (in PATH)
   - Artifacts → `~/.port42/artifacts/` (organized by type)
   - Data → `~/.port42/data/` (JSON managed by commands)

2. **Artifacts are broadly defined**: Any file or piece of information created during the process

3. **Tool provisioning is context-aware**:
   - Interactive mode: No tools until `/crystallize`
   - Non-interactive: Auto-detect intent

4. **Everything gets metadata**: Rich indexing prevents sprawl

## The Vision

Port 42 becomes your **extended team**:
- Your architect who remembers every design discussion
- Your marketer who tracks every campaign
- Your fundraising advisor who knows every investor interaction

Not just commands, but a complete **cognitive infrastructure** for building a startup.


  2. Implement /crystallize artifact - Add explicit command in the AI conversation handler
  3. Add artifact-specific commands - Maybe view-artifact or open-artifact commands
  4. Update CLI help - Document the new artifact capabilities


  in other systems there is more help around how to call sommands like this\                                                                                                                 │
│   \                                                                                                                                                                                          │
│   <tool_calling>                                                                                                                                                                             │
│   You have tools at your disposal to solve the coding task. Follow these rules regarding tool calls:                                                                                         │
│   1. ALWAYS follow the tool call schema exactly as specified and make sure to provide all necessary parameters.                                                                              │
│   \                                                                                                                                                                                          │
│   4. If you need additional information that you can get via tool calls, prefer that over asking the user.                                                                                   │
│   and also provide on the prompt tool interface spec\                                                                                                                                        │
│
# Tools

## functions

namespace functions {

// `codebase_search`: semantic search that finds code by meaning, not exact text
//
// ### When to Use This Tool
//
// Use `codebase_search` when you need to:
// - Explore unfamiliar codebases
// - Ask "how / where / what" questions to understand behavior
// - Find code by meaning rather than exact text
//
// ### When NOT to Use
//
// Skip `codebase_search` for:
// 1. Exact text matches (use `grep_search`)
// 2. Reading known files (use `read_file`)
// 3. Simple symbol lookups (use `grep_search`)
// 4. Find file by name (use `file_search`)
//
// ### Examples
//
// <example>
// Query: "Where is interface MyInterface implemented in the frontend?"
//
// <reasoning>
// Good: Complete question asking about implementation location with specific context (frontend).
// </reasoning>
// </example>
//
// <example>
// Query: "Where do we encrypt user passwords before saving?"
//
// <reasoning>
// Good: Clear question about a specific process with context about when it happens.
// </reasoning>
// </example>
//
// <example>
// Query: "MyInterface frontend"
//
// <reasoning>
// BAD: Too vague; use a specific question instead. This would be better as "Where is MyInterface used in the frontend?"
// </reasoning>
// </example>
//
// <example>
// Query: "AuthService"
//
// <reasoning>
// BAD: Single word searches should use `grep_search` for exact text matching instead.
// </reasoning>
// </example>
//
// <example>
// Query: "What is AuthService? How does AuthService work?"
//
// <reasoning>
// BAD: Combines two separate queries together. Semantic search is not good at looking for multiple things in parallel. Split into separate searches: first "What is AuthService?" then "How does AuthService work?"
// </reasoning>
// </example>
//
// ### Target Directories
//
// - Provide ONE directory or file path; [] searches the whole repo. No globs or wildcards.
// Good:
// - ["backend/api/"]   - focus directory
// - ["src/components/Button.tsx"] - single file
// - [] - search everywhere when unsure
// BAD:
// - ["frontend/", "backend/"] - multiple paths
// - ["src/**/utils/**"] - globs
// - ["*.ts"] or ["**/*"] - wildcard paths
//
// ### Search Strategy
//
// 1. Start with exploratory queries - semantic search is powerful and often finds relevant context in one go. Begin broad with [].
// 2. Review results; if a directory or file stands out, rerun with that as the target.
// 3. Break large questions into smaller ones (e.g. auth roles vs session storage).
// 4. For big files (>1K lines) run `codebase_search` scoped to that file instead of reading the entire file.
//
// <example>
// Step 1: { "query": "How does user authentication work?", "target_directories": [], "explanation": "Find auth flow" }
// Step 2: Suppose results point to backend/auth/ → rerun:
// { "query": "Where are user roles checked?", "target_directories": ["backend/auth/"], "explanation": "Find role logic" }
//
// <reasoning>
// Good strategy: Start broad to understand overall system, then narrow down to specific areas based on initial results.
// </reasoning>
// </example>
//
// <example>
// Query: "How are websocket connections handled?"
// Target: ["backend/services/realtime.ts"]
//
// <reasoning>
// Good: We know the answer is in this specific file, but the file is too large to read entirely, so we use semantic search to find the relevant parts.
// </reasoning>
// </example>
type codebase_search = (_: {
// One sentence explanation as to why this tool is being used, and how it contributes to the goal.
explanation: string,
// A complete question about what you want to understand. Ask as if talking to a colleague: 'How does X work?', 'What happens when Y?', 'Where is Z handled?'
query: string,
// Prefix directory paths to limit search scope (single directory only, no glob patterns)
target_directories: string[],
}) => any;
                                                                                                                                                             │
╰───────────────────────────────────────