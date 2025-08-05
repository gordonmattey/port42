# Artifact Metadata & Indexing System

**Purpose**: Design for metadata storage, indexing, search, and lifecycle management.
**Scope**: Index structure, search queries, cleanup policies, tool gating strategy.

## The Problem

As Port 42 generates more artifacts, we need:
- Fast search and filtering across all artifacts
- Rich metadata for context and relationships
- Version tracking and evolution history
- Relevance scoring for AI context
- Cleanup of outdated content
- **Prevention of artifact sprawl** due to aggressive tool usage

## Proposed Solution

### 1. Unified Artifact Index

```json
// ~/.port42/artifacts/index.json
{
  "version": "1.0",
  "last_updated": "2024-01-15T10:30:00Z",
  "stats": {
    "total_artifacts": 342,
    "by_type": {
      "documents": 156,
      "code": 89,
      "designs": 45,
      "media": 12
    },
    "total_size_mb": 234.5
  },
  "artifacts": [
    {
      "id": "art-1737123456-dashboard",
      "type": "code",
      "subtype": "react-app",
      "title": "Metrics Dashboard",
      "description": "Real-time metrics visualization dashboard",
      "path": "code/2024-01-15-metrics-dashboard/",
      "created": "2024-01-15T10:30:00Z",
      "modified": "2024-01-15T14:22:00Z",
      "accessed": "2024-01-16T09:15:00Z",
      "size_bytes": 4567890,
      
      // Rich metadata
      "tags": ["dashboard", "metrics", "react", "realtime", "analytics", "monitoring"],
      "session_id": "sess-abc123",
      "agent": "@ai-engineer",
      "parent_artifacts": ["art-1737123000-design-doc"],
      "child_artifacts": ["art-1737123789-api-client"],
      
      // Relevance scoring
      "usage_count": 23,
      "last_used": "2024-01-16T09:15:00Z",
      "importance": "high",
      "lifecycle": "active", // draft, active, archived, deprecated
      
      // Technical metadata
      "language": "javascript",
      "framework": "react",
      "dependencies": ["react", "d3", "websocket"],
      "executable": true,
      "run_command": "run-metrics-dashboard",
      
      // AI context
      "summary": "A React dashboard that displays real-time metrics...",
      "embeddings": [0.123, 0.456, ...], // For semantic search
      
      // Relationships
      "related_commands": ["api-metrics", "export-data"],
      "related_sessions": ["sess-abc123", "sess-def456"],
      "references": ["https://d3js.org", "internal:api-docs"]
    }
  ]
}
```

### 2. Artifact Lifecycle Management

```bash
# States
draft → active → stable → archived → deprecated

# Automatic transitions
- draft → active: After first successful use
- active → stable: No changes for 30 days + high usage
- stable → archived: No access for 90 days
- any → deprecated: Manual or via new version
```

### 3. Search & Query System

```bash
# Subcommand: port42 search
port42 search --type code --tags dashboard --recent 7d
port42 search --tags analytics --importance high
port42 search "real-time metrics" --semantic  # Uses embeddings
port42 search --session sess-abc123  # All artifacts from session
port42 search --unused 30d  # Cleanup candidates
```

### 4. Relationship Graph

```
┌─────────────────┐     references    ┌──────────────────┐
│ Design Document │ ←───────────────── │  Code Prototype  │
└────────┬────────┘                    └────────┬─────────┘
         │                                      │
     inspired                              implements
         │                                      │
         ▼                                      ▼
┌─────────────────┐     uses data     ┌──────────────────┐
│  Data Command   │ ←───────────────── │   Dashboard UI   │
└─────────────────┘                    └──────────────────┘
```

### 5. Auto-Tagging & Categorization

AI automatically extracts:
- **Tags**: Technologies, concepts, tools, and high-level groupings
- **Summary**: One-line description  
- **Embeddings**: For semantic similarity

### 6. Cleanup & Maintenance

```bash
# Auto-cleanup rules
- Archived + unused 180 days → Move to cold storage
- Deprecated + no dependencies → Safe to delete
- Failed/incomplete artifacts → Clean after 7 days

# Manual cleanup
port42 clean --dry-run  # Show what would be removed
port42 clean --archived --older-than 6m
port42 clean --size-limit 1GB  # Keep most recent/important
```

### 7. Context Window Optimization

For AI conversations, automatically include:
1. Recently accessed artifacts (last 24h)
2. Frequently used artifacts (top 10%)
3. Related to current session topic
4. Parent/child of mentioned artifacts
5. High importance + relevant tags

### 8. Version Tracking

```json
{
  "id": "art-dashboard-v3",
  "version": 3,
  "previous_version": "art-dashboard-v2",
  "changelog": "Added real-time updates, fixed memory leak",
  "breaking_changes": false,
  "migration_notes": "No changes needed"
}
```

### 9. Implementation Plan

1. **Index Structure** (1 hour)
   - Define JSON schema
   - Create index manager in daemon

2. **Metadata Extraction** (2 hours)
   - Auto-tag based on content
   - Generate summaries and keywords
   - Calculate importance scores

3. **Search Interface** (2 hours)
   - Add `search` subcommand to port42 CLI
   - Implement filters and sorting
   - Add semantic search option

4. **Lifecycle Management** (1 hour)
   - State transitions
   - Cleanup policies
   - Archive system

5. **Relationship Tracking** (1 hour)
   - Parent/child links
   - Cross-references
   - Dependency graph

## Tool Gating Strategy

To prevent artifact sprawl from aggressive AI tool usage:

1. **Interactive Sessions**: No generation tools by default
   - Pure conversation until `/crystallize` is used
   - Tools only provided when explicitly requested
   
2. **Non-Interactive CLI**: Context-aware tool provisioning
   - Detect generation intent ("create a command", "build a tool")
   - Provide appropriate tools automatically
   - Preserve current one-shot convenience

3. **Tool Usage Guidelines**: Updated prompts emphasize restraint
   - "Only use tools when explicitly appropriate"
   - "Don't create artifacts for every idea discussed"

## Benefits

1. **Find anything instantly** - Rich search across all artifacts
2. **AI has better context** - Knows what's relevant
3. **Storage stays clean** - Automatic lifecycle management
4. **See relationships** - Understand how artifacts connect
5. **Track evolution** - Version history and changes
6. **Intentional creation** - Artifacts only when truly needed

## Example Queries

```bash
# Find all React dashboards created this month
port42 search --type code --subtype react-app --created "2024-01-*"

# Find artifacts related to authentication
port42 search --semantic "user authentication oauth jwt"

# Find unused code artifacts
port42 search --type code --accessed-before "30 days ago"

# Find all artifacts from a specific conversation
port42 search --session sess-abc123

# Find high-importance artifacts about architecture  
port42 search --importance high --tags architecture
```

This system scales elegantly from dozens to thousands of artifacts while keeping everything discoverable and maintaining relevant context for AI interactions.