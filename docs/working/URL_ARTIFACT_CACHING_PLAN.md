# URL Artifact Caching Implementation Plan

## Overview
Transform ephemeral URL fetches into permanent knowledge artifacts stored in the object store. This provides persistence, reproducibility, offline access, and audit trails for web content used in tool generation.

## Current State
- URL resolver fetches content and includes in AI context
- Content is discarded after use
- Same URL fetched repeatedly
- No offline access or historical record

## Target State
- First fetch stores content as persistent **URL artifact Relation**
- Subsequent references use cached content from Relations
- Full audit trail of what web content influenced tool generation  
- Offline operation after initial fetch
- **Searchable archive** of fetched URLs via Relations queries
- **Rules integration** - URL artifacts can trigger Rules Engine
- **Rich metadata** in Relations Properties

---

## Key Learnings & Architecture Decisions

### ✅ Relations > Raw Storage
**Decision**: Store URL artifacts as Relations instead of raw storage objects.  
**Rationale**: 
- **Consistency** with existing system architecture (tools, memory use Relations)
- **Rich metadata** support via Properties map
- **Searchability** through Relations queries  
- **Rules integration** - URL artifacts can trigger Rules Engine
- **Type safety** with proper URLArtifact Relations type

### ✅ Content + Metadata Separation  
**Architecture**: Content stored in object storage, metadata in Relations.  
**Benefits**:
- **Efficient queries** - Relations queries don't load large content
- **Flexible storage** - Content in optimized object storage
- **Rich indexing** - Relations Properties fully searchable

---

## Implementation Plan

### Phase 1: Relations Integration ✅ COMPLETE
**Goal**: Connect URL resolver to Relations system (CHANGED from raw storage to Relations)

**KEY LEARNING**: URL artifacts should be Relations, not raw storage objects, for consistency with system architecture and to enable rich metadata, searchability, and rules integration.

#### 1.1 Update Handler Interface ✅
```go
// daemon/resolution/interface.go - IMPLEMENTED
type Handlers struct {
    SearchHandler    func(query string, limit int) ([]SearchResult, error)
    ToolHandler      func(toolName string) (*ToolDefinition, error) 
    MemoryHandler    func(sessionID string) (*MemorySession, error)
    FileHandler      func(path string) (*FileContent, error)
    RelationsHandler func() RelationsManager // NEW: For URL artifact Relations
}

// RelationsManager interface for URL artifact Relations
type RelationsManager interface {
    DeclareRelation(relation *URLArtifactRelation) error
    GetRelationByID(id string) (*URLArtifactRelation, error)
    ListRelationsByType(relationType string) ([]*URLArtifactRelation, error)
}

// URLArtifactRelation represents a URL artifact stored as a Relation
type URLArtifactRelation struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`       // "URLArtifact"
    Properties map[string]interface{} `json:"properties"` // Rich metadata
    CreatedAt  time.Time              `json:"created_at"`
    UpdatedAt  time.Time              `json:"updated_at"`
    ContentID  string                 `json:"content_id"`  // Object ID in storage
    Content    string                 `json:"-"`           // Loaded content
}
```

#### 1.2 Pass Relations to URL Resolver ✅
```go
// daemon/server.go - IMPLEMENTED
handlers := resolution.Handlers{
    // ... existing handlers ...
    RelationsHandler: func() resolution.RelationsManager {
        return &relationsAdapter{
            realityCompiler: d.realityCompiler,
            storage:         d.storage,
        }
    },
}

// relationsAdapter bridges daemon's Relations to resolution interface
type relationsAdapter struct {
    realityCompiler *RealityCompiler
    storage         *Storage
}

// Key methods implemented:
// - DeclareRelation: Stores content in storage + creates Relation
// - GetRelationByID: Loads Relation + content from storage
// - ListRelationsByType: Lists Relations (without loading content)
```

#### 1.3 Update URL Resolver Constructor ✅
```go
// daemon/resolution/resolvers.go - IMPLEMENTED
type urlResolver struct {
    relations RelationsManager // Relations for URL artifact caching
}

// daemon/resolution/service.go - IMPLEMENTED
var relations RelationsManager
if handlers.RelationsHandler != nil {
    relations = handlers.RelationsHandler()
}
s.resolvers["url"] = &urlResolver{relations: relations}
```

**PHASE 1 COMPLETE**: URL resolver now has access to Relations system with proper architecture.

### Phase 2: Artifact Management 🔄 NEXT
**Goal**: Generate unique IDs and manage URL artifact Relations lifecycle

#### 2.1 Artifact ID Generation 🔄
```go
// daemon/resolution/resolvers.go
import "crypto/sha256"

func generateURLArtifactID(url string) string {
    hash := sha256.Sum256([]byte(url))
    return fmt.Sprintf("url-fetch-%x", hash[:8])
}

// Note: Relations don't use paths - they're stored in Relations system
// Content stored separately in object storage, referenced by Relations
```

#### 2.2 Cache Check Logic 🔄
```go
func (r *urlResolver) loadCachedContent(artifactID string) (*URLArtifactRelation, error) {
    if r.relations == nil {
        return nil, nil // No caching available
    }
    
    // Try to load existing URL artifact relation
    relation, err := r.relations.GetRelationByID(artifactID)
    if err != nil {
        return nil, nil // Cache miss
    }
    
    // Check if cache is expired (24h default)
    if fetchedAt, exists := relation.Properties["fetched_at"].(int64); exists {
        fetchTime := time.Unix(fetchedAt, 0)
        if time.Since(fetchTime) > 24*time.Hour {
            return nil, nil // Cache expired
        }
    }
    
    // Cache hit - relation includes content
    return relation, nil
}
```

### Phase 3: Core Resolution Logic ✅ COMPLETE 
**Goal**: Implement cache-first resolution with proper fallback logic

#### 3.1 Enhanced Resolution Flow ✅
```go
func (r *urlResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
    // Validate URL
    if !IsValidURL(target) {
        return &ResolvedContext{
            Type:    "url", Target: target, Success: false,
            Error:   "Invalid URL format",
        }, nil
    }
    
    // Generate artifact ID
    artifactID := NewURLArtifactID(target).Generate()
    
    // Phase 3: Enhanced Resolution Flow - Cache-first with proper fallback logic
    if r.artifactManager != nil {
        // Try cache first  
        if cached, err := r.artifactManager.LoadCached(artifactID); err == nil && cached != nil {
            // Cache hit - successful cache-first resolution
            log.Printf("🎯 URL cache HIT: %s -> %s", target, artifactID)
            content := r.formatCachedURLContent(cached.Content, cached.Properties, target)
            return &ResolvedContext{Type: "url", Target: target, Content: content, Success: true}, nil
        }
        
        // Cache miss - proceed to fetch with caching enabled
        log.Printf("🌐 URL cache MISS: %s -> fetching fresh (will cache)", target)
        return r.fetchAndStore(ctx, target, artifactID)
    } else {
        // No cache manager - direct fetch without caching (graceful degradation)
        log.Printf("🌐 URL direct fetch: %s (no cache available)", target)
        return r.fetchWithoutCaching(ctx, target)
    }
}
```

**Key Features Implemented:**
- **Cache-first logic**: Always check cache before HTTP requests
- **Graceful degradation**: Works without caching infrastructure  
- **Clear logging**: Distinct messages for cache HIT/MISS/unavailable
- **Proper fallbacks**: Multiple fallback paths ensure reliability
- **Performance optimization**: Eliminates duplicate HTTP requests

### Phase 4: Relations Storage Implementation 🔄
**Goal**: Store fetched content as URL artifact Relations with **rich reference context**

#### 4.1 Enhanced Fetch and Store Logic
```go
func (r *urlResolver) fetchAndStoreURL(ctx context.Context, url, artifactID string, referenceContext *ReferenceContext) (string, *Metadata, error) {
    // Fetch content (existing HTTP logic)
    client := &http.Client{Timeout: 8 * time.Second}
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return "", nil, err
    }
    
    req.Header.Set("User-Agent", "Port42-ReferenceResolver/1.0")
    resp, err := client.Do(req)
    if err != nil {
        return "", nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return "", nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
    }
    
    // Read content with size limit
    limitedReader := io.LimitReader(resp.Body, 50*1024)
    bodyBytes, err := io.ReadAll(limitedReader)
    if err != nil {
        return "", nil, err
    }
    
    content := string(bodyBytes)
    
    // Store as URL artifact relation with RICH METADATA
    if r.relations != nil {
        now := time.Now()
        properties := map[string]interface{}{
            // Basic HTTP metadata
            "source_url":     url,
            "content_type":   resp.Header.Get("Content-Type"),
            "status_code":    resp.StatusCode,
            "content_length": len(content),
            "fetched_at":     now.Unix(),
            "cache_version":  2, // Updated for rich metadata
            
            // Content analysis
            "title":          extractTitleFromContent(content, resp.Header.Get("Content-Type")),
            "has_html":       strings.Contains(resp.Header.Get("Content-Type"), "html"),
            "is_json":        strings.Contains(resp.Header.Get("Content-Type"), "json"),
        }
        
        // RICH REFERENCE CONTEXT - Link to tool generation session
        if referenceContext != nil {
            if referenceContext.RelationID != "" {
                properties["triggered_by_relation"] = referenceContext.RelationID
                properties["relation_type"] = referenceContext.RelationType
            }
            if referenceContext.SessionID != "" {
                properties["memory_session"] = referenceContext.SessionID
                properties["agent"] = referenceContext.Agent
            }
            if len(referenceContext.AllReferences) > 0 {
                properties["reference_batch_size"] = len(referenceContext.AllReferences)
                
                // Cross-reference context - what other references were resolved together?
                var crossRefs []string
                for _, ref := range referenceContext.AllReferences {
                    if ref.Type != "url" || ref.Target != url { // Exclude self
                        crossRefs = append(crossRefs, fmt.Sprintf("%s:%s", ref.Type, ref.Target))
                    }
                }
                if len(crossRefs) > 0 {
                    properties["cross_references"] = crossRefs
                }
            }
            properties["resolution_timestamp"] = referenceContext.ResolvedAt.Unix()
        }
        
        relation := &URLArtifactRelation{
            ID:        artifactID,
            Type:      "URLArtifact", 
            Content:   content,
            CreatedAt: now,
            UpdatedAt: now,
            Properties: properties,
        }
        
        err = r.relations.DeclareRelation(relation)
        if err != nil {
            log.Printf("⚠️ Failed to cache URL artifact relation: %v", err)
            // Continue without caching - don't fail resolution
        } else {
            log.Printf("💾 URL cached as relation with rich context: %s", artifactID)
            if referenceContext != nil && referenceContext.RelationID != "" {
                log.Printf("🔗 Linked to relation: %s", referenceContext.RelationID)
            }
        }
        
        return content, relation, nil
    }
    
    return content, nil, nil
}
```

#### 4.2 Reference Context Structure
```go
// ReferenceContext captures the rich context of why a URL was fetched
type ReferenceContext struct {
    RelationID     string      `json:"relation_id,omitempty"`     // Tool/entity being created
    RelationType   string      `json:"relation_type,omitempty"`   // "Tool", "Memory", etc.
    SessionID      string      `json:"session_id,omitempty"`      // Memory session
    Agent          string      `json:"agent,omitempty"`           // AI agent name
    AllReferences  []Reference `json:"all_references,omitempty"`  // Full reference batch 
    ResolvedAt     time.Time   `json:"resolved_at"`               // When resolution occurred
}

// Update URL resolver resolve method to accept reference context
func (r *urlResolver) resolve(ctx context.Context, target string, referenceContext *ReferenceContext) (*ResolvedContext, error) {
    // ... existing validation logic ...
    
    // Generate artifact ID
    artifactID := generateURLArtifactID(target)
    
    // Try cache first (with usage tracking)
    if cached := r.loadCachedContent(artifactID); cached != nil {
        r.updateUsageStats(cached, referenceContext) // Track reuse
        log.Printf("🎯 URL cache hit: %s", target)
        return r.formatCachedResult(target, cached), nil
    }
    
    // Cache miss - fetch fresh content with rich context
    log.Printf("🌐 URL cache miss, fetching: %s", target)
    content, metadata, err := r.fetchAndStoreURL(ctx, target, artifactID, referenceContext)
    // ... rest of logic ...
}
```

#### 4.2 Rich Metadata Creation
```go
func (r *urlResolver) createURLMetadata(url string, resp *http.Response, content string) *Metadata {
    now := time.Now()
    
    // Extract title from HTML content
    title := extractTitleFromContent(content, resp.Header.Get("Content-Type"))
    if title == "" {
        title = "Web Content: " + url
    }
    
    return &Metadata{
        Type:        "url-fetch",
        Title:       title,
        Description: fmt.Sprintf("Web content fetched from %s", url),
        Tags:        []string{"web", "reference", "cached"},
        Created:     now,
        Modified:    now,
        Accessed:    now,
        
        // URL-specific properties
        Properties: map[string]interface{}{
            "source_url":     url,
            "content_type":   resp.Header.Get("Content-Type"),
            "status_code":    resp.StatusCode,
            "content_length": len(content),
            "fetched_at":     now.Unix(),
            "cache_version":  1,
        },
    }
}

func extractTitleFromContent(content, contentType string) string {
    if strings.Contains(contentType, "html") {
        // Extract <title> tag
        titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
        if matches := titleRegex.FindStringSubmatch(content); len(matches) > 1 {
            return strings.TrimSpace(matches[1])
        }
    }
    return ""
}
```

### Phase 5: Result Formatting
**Goal**: Consistent formatting for cached vs fresh content

#### 5.1 Format Functions
```go
func (r *urlResolver) formatCachedResult(url string, cached *URLArtifact) *ResolvedContext {
    content := formatURLContent(cached.Content, 
        cached.Metadata.Properties["content_type"].(string), url)
    
    // Add cache indicator
    content += fmt.Sprintf("\n[Cached from %s]", 
        time.Unix(cached.Metadata.Properties["fetched_at"].(int64), 0).Format("2006-01-02 15:04:05"))
    
    return &ResolvedContext{
        Type:    "url",
        Target:  url, 
        Content: content,
        Success: true,
    }
}

func (r *urlResolver) formatFreshResult(url, content string, metadata *Metadata) *ResolvedContext {
    var contentType string
    if metadata != nil {
        contentType = metadata.Properties["content_type"].(string)
    }
    
    formattedContent := formatURLContent(content, contentType, url)
    formattedContent += "\n[Freshly fetched]"
    
    return &ResolvedContext{
        Type:    "url",
        Target:  url,
        Content: formattedContent, 
        Success: true,
    }
}
```

### Phase 6: Enhanced Logging & Stats
**Goal**: Visibility into caching behavior

#### 6.1 Resolution Stats Enhancement
```go
// Add to resolution/service.go
type Stats struct {
    TotalReferences  int            `json:"total_references"`
    ResolvedCount    int            `json:"resolved_count"`
    FailedCount      int            `json:"failed_count"`
    TotalContentSize int            `json:"total_content_size"`
    TypeBreakdown    map[string]int `json:"type_breakdown"`
    SuccessRate      float64        `json:"success_rate_percent"`
    
    // NEW: Cache statistics
    CacheStats       CacheStats     `json:"cache_stats,omitempty"`
}

type CacheStats struct {
    CacheHits   int `json:"cache_hits"`
    CacheMisses int `json:"cache_misses"`
    CacheRate   float64 `json:"cache_hit_rate_percent"`
}
```

#### 6.2 Enhanced Logging
```go
// In resolution logs:
log.Printf("📊 Resolution stats: %d/%d successful (%.1f%%) [%d cached, %d fresh]", 
    stats.ResolvedCount, stats.TotalReferences, stats.SuccessRate,
    stats.CacheStats.CacheHits, stats.CacheStats.CacheMisses)
```

---

## Testing Strategy

### Unit Tests
```go
// daemon/resolution/resolvers_test.go
func TestURLResolver_CacheHitMiss(t *testing.T)
func TestURLResolver_CacheExpiration(t *testing.T) 
func TestURLResolver_StorageFailure(t *testing.T)
func TestURLResolver_GracefulDegradation(t *testing.T)
```

### Integration Tests
```bash
# Test cache behavior
../bin/port42 declare tool test-url-cache --ref url:https://httpbin.org/json
# Should see "Freshly fetched" in logs

../bin/port42 declare tool test-url-cache2 --ref url:https://httpbin.org/json  
# Should see "URL cache hit" and "Cached from" in content
```

### Manual Testing
1. First URL reference → Fresh fetch + storage
2. Second URL reference → Cache hit 
3. Same URL after 25 hours → Cache miss (expired)
4. Storage failure → Graceful fallback to direct fetch

#### 4.3 Integration with Resolution Pipeline
The rich context flows from tool generation through to URL artifacts:

```
1. Tool Declaration (server.go:1098-1152)
   ├── References: [{type: "url", target: "https://api.example.com/data"}]
   ├── SessionContext: {session: "ai-session-123", agent: "@ai-engineer"}
   └── RelationID: "tool-data-analyzer-abc123"

2. Reference Resolution (resolution/service.go:98-133)
   ├── Creates ReferenceContext from tool declaration
   ├── Passes context to URL resolver 
   └── URL resolver stores rich metadata in URLArtifact relation

3. URL Artifact Properties (what gets stored):
   {
     // HTTP metadata
     "source_url": "https://api.example.com/data",
     "content_type": "application/json",
     "status_code": 200,
     
     // Rich reference context
     "triggered_by_relation": "tool-data-analyzer-abc123",
     "relation_type": "Tool", 
     "memory_session": "ai-session-123",
     "agent": "@ai-engineer",
     "cross_references": ["search:data analysis", "memory:previous-results"],
     "reference_batch_size": 3,
     "resolution_timestamp": 1691234567
   }
```

This creates a **complete audit trail** from conversation → tool → URL fetch, enabling powerful queries like:
- "What URLs did @ai-engineer access when building the data analyzer?"
- "Which tools were influenced by api.example.com content?" 
- "Show me all URLs fetched during session ai-session-123"

---

## Benefits Delivered

### ✅ Persistence
- URL content becomes permanent artifacts
- Available for future tool generations
- Survives system restarts

### ✅ Performance  
- Cached responses avoid repeated HTTP calls
- Faster resolution for repeated URLs
- Reduced bandwidth usage

### ✅ Reliability
- Works offline after first fetch
- No dependency on external site availability
- Consistent content for reproducible builds

### ✅ Searchability
- URL artifacts discoverable via search system
- Can find "What web content did we use?"
- Browse cached content archive

### ✅ Auditability
- Full history of what influenced tool generation
- Metadata tracks when/how content was fetched
- Cache vs fresh indicators in logs

---

## Future Enhancements

### Rich Reference Context Integration
**Pass ReferenceContext through resolution pipeline**
- **Benefits**: Contextual Intelligence - each URL fetch knows "who asked, when, why, and what else was being researched"
- **Implementation**: Modify resolution pipeline to carry ReferenceContext from tool generation through to URL artifact storage

**Link URL artifacts to tool generation sessions**
- **Benefits**: Complete provenance chain from conversation → tool → URLs → generated code. Reproducibility by restoring exact tool generation environment.
- **Implementation**: Store tool relation IDs, session IDs, and agent context in URL artifact Properties

**Cross-reference tracking (what other refs were resolved together)**
- **Benefits**: Pattern Recognition - discover common URL combinations. "URLs A, B, and C are often used together for authentication tools"
- **Implementation**: Store reference batch information in URL artifacts, build co-access analytics

**Memory session and AI agent context preservation**
- **Benefits**: Conversation continuity - URLs remain available even if external sites go down. Agent specialization tracking.
- **Implementation**: Link URL artifacts to memory sessions and track agent attribution patterns

### Advanced Metadata Enhancement
**Tool genealogy: Which tools used this URL content**
- **Benefits**: Impact Analysis - "This URL influenced 12 different tools - changes here have broad impact"
- **Implementation**: Reverse indexing from URLs to tools, tool family clustering

**Session correlation: URLs accessed together in conversations**  
- **Benefits**: Research Pattern Mining - understand how complex technical problems require multiple information sources
- **Implementation**: Conversation clustering by URL access patterns, expert path discovery

**Content analysis: Title extraction, content classification**
- **Benefits**: Semantic Search - find URLs by content topic, not just URL text. Auto-tagging by technical domain.
- **Implementation**: HTML title parsing, content classification ML, semantic indexing

**Usage patterns: Access frequency, last used timestamps**
- **Benefits**: Predictive Caching - pre-fetch URLs likely to be needed. Cache lifecycle management based on actual usage.
- **Implementation**: Usage analytics, predictive models, automated cache optimization

### Result Formatting Enhancements  
**Rich context display for debugging**
- **Benefits**: Deep Diagnostics - see complete resolution chain when tools behave unexpectedly
- **Implementation**: Detailed resolution logs, reference forensics, context reconstruction tools

**Enhanced cache indicators**
- **Benefits**: Cache confidence - users understand freshness and can make decisions about regeneration
- **Implementation**: Detailed age/TTL displays, freshness warnings, quality indicators

### Enhanced Metadata & Content Analysis
- **Title extraction from HTML `<title>` tag**
- **Content summary or key topics** using AI analysis  
- **Content classification** (API response, documentation, blog post, etc.)
- **Derived metadata** from content structure and semantics
- **Content change detection** with diff tracking between fetches

### Advanced Usage Tracking
- **Reference frequency**: How many times each URL was referenced
- **Last accessed time** with automatic cache warming
- **Tool genealogy**: Which specific tools were generated using this content
- **Cross-reference patterns**: Common URL combinations in tool generation
- **Session correlation**: URLs frequently accessed together in conversations

### Cache Management
- Manual cache refresh with `--refresh` flag
- Bulk cache cleanup commands  
- Cache size limits and LRU eviction

### Advanced Features
- Conditional requests (If-Modified-Since)
- Content-based cache invalidation
- URL redirect handling
- Authentication support

### Analytics & Intelligence
- Most referenced URLs dashboard
- Cache hit rate trends over time
- **Content influence mapping**: Which URLs lead to successful tools
- **Dead link detection** and automatic cleanup
- **Recommendation Engine**: "Users who accessed this URL also found these resources helpful"
- **Pattern Recognition Dashboard**: Common URL access sequences and clusters
- **Predictive Analytics**: Automatic cache warming based on usage patterns
- **Quality Metrics**: URLs that consistently produce successful tools
- **Impact Analysis Tools**: Dependency mapping for URL changes

---

## Implementation Checklist

- [x] **Phase 1**: Relations integration ✅ COMPLETE
  - [x] Update handler interface with RelationsHandler
  - [x] Create relationsAdapter bridging daemon Relations
  - [x] Update URL resolver constructor with Relations
  - [x] **BUILDS SUCCESSFULLY** and daemon starts

- [ ] **Phase 2**: Artifact management 🔄 NEXT  
  - [ ] Implement ID generation with URL hashing
  - [ ] Add cache check logic via Relations queries
  - [ ] Handle cache expiration (24h default)

- [x] **Phase 3**: Core resolution ✅ COMPLETE
  - [x] Cache-first resolution flow
  - [x] Graceful fallback handling (no cache available)  
  - [x] Error handling with Relations
  - [x] Enhanced logging with clear HIT/MISS/direct indicators

- [ ] **Phase 4**: Relations storage implementation 🔄
  - [ ] Fetch and store as URLArtifact Relations
  - [ ] Rich metadata in Properties
  - [ ] Content storage with References

- [ ] **Phase 5**: Result formatting 🔄
  - [ ] Consistent formatting functions
  - [ ] Cache indicators (cached vs fresh)
  - [ ] Content processing

- [ ] **Phase 6**: Logging & stats 🔄
  - [ ] Enhanced resolution stats
  - [ ] Cache hit/miss tracking
  - [ ] Debug logging

- [ ] **Testing**: Unit + integration tests 🔄
- [ ] **Documentation**: Update reference resolution docs 🔄

---

*This plan transforms URL references from ephemeral fetches into a persistent knowledge artifact system, providing the foundation for reproducible, offline-capable tool generation with full audit trails.*