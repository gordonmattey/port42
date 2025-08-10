# Universal Reference System for Port 42 Declarative Mode

## Why: The Context Problem

**Current State**: When you declare a tool, the AI generates it in isolation:
```bash
port42 declare tool log-analyzer --transforms logs,analysis
# AI has only: name + transforms. No context about your environment, existing tools, or intent.
```

**The Problem**: 
- AI doesn't know what files you want to process
- AI can't see your existing tools to extend or complement them
- AI can't reference your previous searches or conversations
- AI generates generic tools instead of contextually relevant ones

**The Vision**: Transform declaration from isolated creation to **contextual composition**:
```bash
port42 declare tool smart-analyzer --ref search:"nginx errors" --ref tool:log-parser --ref file:error.log
# AI now has: search context + existing tool patterns + actual file structure
```

## What: Universal Reference Architecture

### Reference Types

**1. Search References** - Include discovery context
```
search:query → Include actual search results and matches
search:nginx → All tools/memory/files matching "nginx"
search:"error analysis" → Semantic matches for error analysis
```

**2. Tool References** - Extend/relate to existing tools
```  
tool:git-haiku → Include tool definition, code, and patterns
tool:csv-parser → Reference existing tool for extension/composition
relation:abc123 → Direct relation ID reference
```

**3. Memory References** - Leverage conversation context
```
memory:cli-1234 → Include specific conversation thread
memory:session-456 → Reference entire session context
memory:recent → Last N conversations with relevant context
```

**4. File References** - Ground in actual data
```
file:./config.json → Include file contents and structure
file:*.log → Pattern matching for multiple files
file:/path/data/ → Directory context and file listing
```

**5. URL References** - External context integration
```
url:https://api.example.com → Fetch and include API documentation
url:github.com/user/repo → Repository context and README
url:docs.service.com/guide → External documentation context
```

**6. Hybrid References** - Multi-dimensional context
```
combo:search="docker errors" + tool:container-manager + file:docker-compose.yml
→ Creates tool that knows about docker errors, extends container-manager, processes your compose file
```

### Reference Resolution Pipeline

**Step 1: Parse References**
```rust
ref search:"nginx errors" → SearchReference { query: "nginx errors" }
ref tool:git-haiku → ToolReference { id: "git-haiku" }  
ref file:error.log → FileReference { path: "error.log" }
```

**Step 2: Resolve Context**
- Search refs → Execute search, return top matches + metadata
- Tool refs → Load relation definition + generated code + usage patterns  
- Memory refs → Load conversation + artifacts + session context
- File refs → Read contents + infer structure + extract metadata
- URL refs → Fetch + parse + extract relevant sections

**Step 3: Context Synthesis**  
Combine all resolved contexts into rich AI prompt:
```
"Create tool 'smart-analyzer' with transforms [logs, analysis].

Context from search 'nginx errors':
- Found 3 existing tools: error-parser, nginx-monitor, log-cleaner
- Common patterns: regex parsing, timestamp extraction, severity classification

Context from tool 'log-parser':
- Existing implementation handles structured logs
- Uses Python with regex + pandas
- Outputs JSON format with metadata

Context from file 'error.log':  
- 2MB file with nginx error entries
- Format: timestamp + level + message + request_id
- Common errors: 404s, timeout, upstream failures
```

## How: Implementation Strategy

### Phase 1: Reference Protocol Foundation
**Goal**: Establish universal reference syntax and parsing

**Protocol Extension** (`protocol.go`):
```go
type Reference struct {
    Type    string `json:"type"`    // "search", "tool", "memory", "file", "url"
    Target  string `json:"target"`  // The thing being referenced  
    Context string `json:"context,omitempty"` // Additional context/query
}

type Request struct {
    // ...existing fields
    References []Reference `json:"references,omitempty"` // Universal references
}
```

**CLI Parsing** (`cli/src/commands/declare.rs`):
```rust
// Parse --ref arguments
port42 declare tool analyzer --ref search:"nginx errors" --ref tool:log-parser

fn parse_references(ref_args: Vec<String>) -> Vec<Reference> {
    // search:"nginx errors" → Reference { type: "search", target: "nginx errors" }
    // tool:log-parser → Reference { type: "tool", target: "log-parser" }
}
```

### Phase 2: Reference Resolvers
**Goal**: Build context resolution engines

**Search Resolver**:
```go
func ResolveSearchReference(ref Reference) SearchContext {
    // Execute search using existing search infrastructure
    // Return matches + metadata + relationship info
}
```

**Tool Resolver**:
```go  
func ResolveToolReference(ref Reference) ToolContext {
    // Load relation definition from relation store
    // Load generated code from filesystem
    // Extract usage patterns and metadata
}
```

**Memory Resolver**:
```go
func ResolveMemoryReference(ref Reference) MemoryContext {
    // Load conversation from memory store
    // Include artifacts and session metadata
    // Extract relevant conversation segments
}
```

**File Resolver**:
```go
func ResolveFileReference(ref Reference) FileContext {
    // Read file contents (with size limits)
    // Infer structure and format
    // Extract key patterns and metadata
}
```

### Phase 3: Context Synthesis Engine
**Goal**: Combine resolved contexts into AI-ready prompts

**Context Combiner**:
```go
func SynthesizeContext(relation Relation, references []ResolvedReference) AIContext {
    // Combine all contexts into structured prompt
    // Handle context size limits intelligently  
    // Prioritize most relevant information
    // Generate context-aware tool specifications
}
```

**Smart Prompting**:
- **Tool Extensions**: "Extend tool X by adding capabilities Y based on pattern Z"
- **Data-Driven**: "Process files matching format X with techniques learned from search Y"  
- **Conversation-Aware**: "Build on discussion Z to solve problem Y with approach X"

### Phase 4: Advanced Reference Intelligence
**Goal**: Smart reference resolution and suggestion

**Reference Suggestion Engine**:
```bash
port42 declare tool analyzer --transforms logs
# AI suggests: --ref search:"log analysis" --ref tool:existing-log-parser --ref file:sample.log
```

**Automatic Reference Discovery**:
- Detect files in current directory that match tool purpose
- Suggest related tools based on transforms similarity  
- Recommend relevant memory sessions based on context
- Identify useful search queries based on tool name/transforms

**Reference Validation**:
- Verify referenced tools/files/sessions exist
- Warn about potential conflicts or redundancies
- Suggest alternatives for missing references

### Phase 5: Reference-Aware Virtual Filesystem
**Goal**: Expose reference relationships through filesystem

**New Virtual Paths**:
```bash
port42 ls /tools/analyzer/references/     # Show what this tool references
port42 ls /tools/by-reference/search/     # Tools that reference searches  
port42 ls /references/tool:git-haiku/     # Everything that references git-haiku
port42 ls /references/file:config.json/   # Tools that reference this file
```

**Reference Metadata**:
- Track what each tool references at declaration time
- Enable reverse lookups (what references this?)
- Support reference-based discovery and navigation

## Implementation Priority

### Step 1: Foundation
- [ ] Extend protocol with Reference types
- [ ] Add CLI --ref argument parsing
- [ ] Build basic reference resolution framework

### Step 2: Core Resolvers  
- [ ] Implement search reference resolver
- [ ] Implement tool reference resolver  
- [ ] Implement file reference resolver
- [ ] Build context synthesis engine

### Step 3: Memory & URL Resolvers
- [ ] Implement memory reference resolver
- [ ] Implement URL reference resolver  
- [ ] Add smart context combination logic

### Step 4: Intelligence Layer
- [ ] Build reference suggestion engine
- [ ] Add reference validation
- [ ] Implement automatic reference discovery

### Step 5: Virtual Filesystem Integration
- [ ] Add reference-aware virtual paths
- [ ] Enable reference-based tool discovery
- [ ] Build reference metadata tracking

## Success Metrics

**Contextual Tool Generation**: Tools generated with references should be significantly more relevant and useful than tools generated without context.

**Reference Utilization**: Users should naturally adopt reference syntax for complex tool declarations.

**Discovery Enhancement**: Reference-based navigation should reveal tool relationships that weren't obvious before.

**Composition Intelligence**: Tools should naturally extend/complement each other when declared with appropriate references.

---

**The Universal Reference System transforms Port 42 from isolated tool generation into contextual reality compilation - where every declaration builds on the rich context of your entire digital ecosystem.** 🐬