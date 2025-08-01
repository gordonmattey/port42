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