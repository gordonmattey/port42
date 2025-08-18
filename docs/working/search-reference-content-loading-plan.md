# Search Reference Content Loading Enhancement Plan

## Problem Statement

Currently, search references (e.g., `--ref search:architecture`) only provide paths and snippets, not the full content from search results. When a search finds memory sessions containing architectural discussions, the AI only gets a list of paths rather than the actual conversation content.

## Current Behavior

**Search Results Structure:**
```json
{
  "results": [
    {
      "path": "/memory/cli-1755534830887",
      "score": 2.40,
      "snippet": "Tag: architecture"
    }
  ]
}
```

**Current Reference Resolution:**
- Returns formatted list of paths and snippets
- AI receives: "Found 20 results: 1. /memory/cli-1755534830887..."
- AI lacks actual conversation content to analyze

## Target Behavior

**Enhanced Reference Resolution:**
- Filter results by score threshold (≥ 2.0) and hard limit (≤ 5 results)
- Load full content from each high-scoring result path
- Provide AI with actual conversation text, not just metadata

**Example Context:**
```
=== Search Results: search:architecture ===
Found 3 high-relevance results:

1. /tools/code-debt-analysis (score: 10.00)
[FULL EXECUTABLE CONTENT]

2. /memory/cli-1755534830887 (score: 2.40)  
[FULL CONVERSATION TRANSCRIPT]

3. /memory/cli-1755534659995 (score: 2.40)
[FULL CONVERSATION TRANSCRIPT]
```

## Architecture Requirements

### Daemon-Side Implementation
Reference resolution must be handled uniformly by the daemon for both:
- **possess mode**: References resolved and injected as AI context
- **declare mode**: References resolved and embedded in tool creation

### Component Integration
- **Search engine**: Already provides scored results with paths
- **VFS resolver**: Already resolves paths to content via `ResolvePath()`
- **Reference system**: Needs enhancement to handle search: type differently

### Implementation Location
**Primary:** `daemon/storage.go` or new `daemon/reference_resolver.go`
- Extend existing reference resolution logic
- Reuse VFS path resolution for content loading
- Apply score filtering and limits

## Technical Design

### Score-Based Filtering
```go
const (
    SearchScoreThreshold = 2.0
    SearchMaxResults     = 5
)

func resolveSearchReference(query string) (string, error) {
    // 1. Execute search
    results := search(query)
    
    // 2. Filter by score and limit
    filtered := filterByScore(results, SearchScoreThreshold)
    limited := limitResults(filtered, SearchMaxResults)
    
    // 3. Load full content for each result
    content := ""
    for _, result := range limited {
        fullContent := ResolvePath(result.Path)
        content += formatResult(result, fullContent)
    }
    
    return content, nil
}
```

### Integration Points
- **possess command**: Send references to daemon, receive resolved content
- **declare command**: Already sends references to daemon for tool creation
- **search command**: Shares same filtering logic for consistency

## Performance Considerations

### Request Volume
- Current: 1 search operation
- Enhanced: 1 search + up to 5 path resolutions
- Mitigation: Hard limit keeps additional requests bounded

### Content Volume  
- Risk: 5 full memory sessions could be very large
- Mitigation: Score filtering ensures only high-relevance content
- Future: Add content truncation if needed

### Latency
- Additional path resolutions add latency
- Daemon-side processing more efficient than client-side
- Can optimize with concurrent resolution

## Testing Strategy

### Pre-Implementation Tests
1. **Current behavior baseline:**
   ```bash
   port42 possess @ai-engineer "analyze architecture" --ref search:architecture
   # Measure: context size, AI knowledge, performance
   ```

2. **Manual content verification:**
   ```bash
   port42 search architecture  # Get result paths
   port42 cat /memory/cli-1755534830887  # Verify content exists and size
   ```

### Post-Implementation Validation
1. **Content loading:** AI should have full conversation details
2. **Score filtering:** Only high-relevance results should be included
3. **Performance:** Reasonable latency increase (< 2x current time)
4. **Error handling:** Graceful handling of missing/invalid paths

### Test Cases
- High-score search (tools/commands with score ~10.0)
- Mixed-score search (architecture: tools + memory sessions)  
- Low-score search (mostly results below threshold)
- Empty search results
- Search with some invalid result paths

## Implementation Phases

### Phase 1: Fix Reference Resolution Architecture
- Move possess mode reference resolution to daemon-side
- Reuse daemon reference resolution components
- Maintain current behavior (paths + snippets)

### Phase 2: Enhanced Search Content Loading  
- Implement score-based filtering in daemon
- Add full content resolution for search results
- Apply to both possess and declare modes

### Phase 3: Optimization
- Add concurrent path resolution
- Implement content truncation if needed
- Performance monitoring and tuning

## Success Criteria

1. **Functional:** AI receives full conversation content from search references
2. **Performance:** < 2x latency increase for typical search references  
3. **Quality:** Score filtering provides only relevant, high-value content
4. **Consistency:** Same behavior in both possess and declare modes
5. **Reliability:** Graceful handling of missing or invalid result paths

## Configuration

### Internal Constants
```go
SearchScoreThreshold = 2.0  // Include results with score >= 2.0
SearchMaxResults     = 5    // Hard cap on content loading
```

### Future User Controls
- Search result limits via command flags
- Score threshold overrides for power users
- Content truncation settings