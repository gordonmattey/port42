# Port 42: Three Models for Three Core Workflows

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

### Model 3: Knowledge Artifacts (Documents/Conversations)
- **What**: Markdown documents, conversation logs, unstructured knowledge
- **When**: Strategy, exploration, documentation, learning
- **Storage**: Markdown files in ~/.port42/artifacts/
- **Example**: Pitch decks, strategy docs, architecture decisions

## Your Three Core Workflows

### 1. Product & Engineering Workflow

**Conversation Mode**:
```bash
possess @ai-architect "Let's design the real-time sync system"
# Explores architecture, discusses tradeoffs
# When ready: /crystallize → creates `sync-system-design.md`
```

**Tool Mode**:
```bash
possess @ai-engineer "Create a command to analyze our API performance"
# /crystallize → creates `api-analyzer` command
```

**Living Document Mode**:
```bash
possess @ai-pm "Help me track feature requests"
# /crystallize → creates `feature-tracker` command with CRUD operations
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
possess @ai-brand "Help define Port 42's voice and messaging"
# Conversation about brand
# /crystallize → creates `brand-guide.md`
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
/crystallize document     # Create markdown artifact
/crystallize data         # Create CRUD command
```

### 2. How It Works

**Command Generation (Current)**
- AI detects command specification in conversation
- Generates executable script in `~/.port42/commands/`
- Immediately available in PATH

**Document Generation (New)**
- AI creates structured markdown with metadata
- Saves to `~/.port42/artifacts/{type}/{date}-{title}.md`
- Auto-generates viewing command: `view-{title}`
- Perfect for decisions, strategies, meeting notes

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
├── artifacts/         # Markdown documents
│   ├── decisions/
│   │   └── 2024-01-15-api-architecture.md
│   ├── strategies/
│   │   └── 2024-01-15-growth-plan.md
│   └── index.json
├── data/             # Structured data for CRUD commands
│   ├── content-plan.json
│   ├── feature-tracker.json
│   └── investor-crm.json
└── sessions/         # Conversation history
```

### 4. AI Context for Each Type

**For Commands**: "Generate an executable tool that..."
**For Documents**: "Create a markdown document that captures..."
**For Data**: "Design a data schema and CRUD operations for..."

## Next Steps

1. **Prototype the document creation flow**
   - Extend crystallize to create markdown artifacts
   - Add document browsing commands

2. **Enhance data management commands**
   - Better CRUD templates
   - Data migration tools
   - Export/import capabilities

3. **Use existing agents for all workflows**
   - `@ai-engineer` + `@ai-muse` for product & engineering
   - `@ai-growth` + `@ai-muse` for marketing & content
   - `@ai-founder` + `@ai-growth` for fundraising & strategy

## Key Insights from Design Process

1. **The /crystallize command already exists** - We just need to extend it with options
2. **Three distinct storage locations** make sense:
   - Commands → `~/.port42/commands/` (executable)
   - Documents → `~/.port42/artifacts/` (knowledge)
   - Data → `~/.port42/data/` (structured JSON)

3. **Document viewing** can be handled by auto-generated commands
4. **CRUD templates** can be simple bash/Python scripts that manage JSON
5. **AI context switching** is just different prompts based on type

## The Vision

Port 42 becomes your **extended team**:
- Your architect who remembers every design discussion
- Your marketer who tracks every campaign
- Your fundraising advisor who knows every investor interaction

Not just commands, but a complete **cognitive infrastructure** for building a startup.