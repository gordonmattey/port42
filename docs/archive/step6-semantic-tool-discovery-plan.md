# Step 6: Semantic Tool Discovery - Implementation Plan

**Goal**: Implement automatic tool similarity detection and relationship creation based on transforms and purpose analysis.

**Value Delivered**: Users can discover connections between tools they didn't realize existed, enabling better tool reuse and ecosystem understanding.

---

## **Current State Analysis**

### ✅ **What Already Exists**
1. **Transform-based Organization**: `/tools/by-transform/{transform}/` already groups tools by capabilities
2. **Virtual Path Infrastructure**: `/similar/` path exists but returns empty (placeholder)
3. **Relationship System**: Relations support properties like `spawned_by`, `parent`
4. **Semantic Search**: Full-text search across tools, memory, artifacts with scoring
5. **Relation Storage**: `relationStore` can store and query tool relationships

### ❌ **What's Missing for Step 6**
1. **Similarity Algorithm**: `calculateSimilarity()` function to compare tool transforms
2. **Similar Path Resolution**: Logic to populate `/similar/{tool-name}` with related tools
3. **Similarity Relationship Type**: `similar_to` relationships between tools
4. **Automatic Discovery**: Background similarity detection when tools are created
5. **Similarity Scoring**: Threshold-based similarity matching (50%+ similarity)

---

## **Architecture Design**

### **1. Similarity Calculation Algorithm**
```go
// daemon/similarity.go - New file
type SimilarityCalculator struct {
    relationStore RelationStore
}

func (sc *SimilarityCalculator) calculateTransformSimilarity(transforms1, transforms2 []string) float64 {
    // Jaccard similarity: intersection / union
    // Also consider semantic similarity of transform names
    // Weight common analysis patterns higher
}

func (sc *SimilarityCalculator) findSimilarTools(targetTool Relation) []SimilarTool {
    // Load all tools, calculate similarities, return above threshold
}

type SimilarTool struct {
    Tool       Relation `json:"tool"`
    Similarity float64  `json:"similarity"`
    Reason     []string `json:"reason"` // ["shared transforms: analysis, process", "similar naming pattern"]
}
```

### **2. Virtual Path Resolution Enhancement**
```go
// daemon/storage.go - Extend existing handleVirtualPath
func (s *Storage) handleSimilarView(path string) []map[string]interface{} {
    // Extract tool name from path: /similar/log-analyzer -> "log-analyzer"
    // Find the target tool relation
    // Calculate similarities with all other tools
    // Return similar tools as virtual nodes
}
```

### **3. Relationship Integration**
```go
// Store similarity relationships in relations system
type SimilarityRelationship struct {
    From       string  `json:"from"`        // source tool ID
    To         string  `json:"to"`          // similar tool ID
    Type       string  `json:"type"`        // "similar_to"
    Similarity float64 `json:"similarity"`  // 0.0 - 1.0 score
    Reasons    []string `json:"reasons"`    // explanation
}
```

---

## **Implementation Phases**

### **Phase A: Core Similarity Algorithm**
**Files to Create/Modify:**
- `daemon/similarity.go` (new file)
- Unit tests for similarity calculation

**Key Functions:**
1. `calculateTransformSimilarity(transforms1, transforms2 []string) float64`
2. `findSimilarTools(targetTool Relation, threshold float64) []SimilarTool`
3. `semanticTransformDistance(transform1, transform2 string) float64`

**Algorithm Details:**
```go
func calculateTransformSimilarity(transforms1, transforms2 []string) float64 {
    if len(transforms1) == 0 || len(transforms2) == 0 {
        return 0.0
    }
    
    // Convert to sets for intersection/union calculation
    set1 := toSet(transforms1)
    set2 := toSet(transforms2)
    
    intersection := setIntersection(set1, set2)
    union := setUnion(set1, set2)
    
    if len(union) == 0 {
        return 0.0
    }
    
    // Jaccard similarity coefficient
    baseSimilarity := float64(len(intersection)) / float64(len(union))
    
    // Semantic boost for related transforms
    semanticBoost := calculateSemanticBoost(transforms1, transforms2)
    
    return math.Min(1.0, baseSimilarity + semanticBoost)
}

func calculateSemanticBoost(transforms1, transforms2 []string) float64 {
    // Boost similarity for semantically related transforms
    // e.g., "parse" and "parsing", "analyze" and "analysis"
    semanticPairs := map[string][]string{
        "analyze": {"analysis", "analyzer", "inspect"},
        "parse":   {"parsing", "parser", "process"},
        "format":  {"formatting", "formatter", "display"},
        "test":    {"testing", "verify", "validation"},
    }
    
    boost := 0.0
    for _, t1 := range transforms1 {
        for _, t2 := range transforms2 {
            if areSemanticallySimilar(t1, t2, semanticPairs) {
                boost += 0.1 // 10% boost per semantic match
            }
        }
    }
    
    return math.Min(0.3, boost) // Cap semantic boost at 30%
}
```

**Test Success Criteria:**
- `log-analyzer` and `quick-analyzer` show high similarity (shared "analysis")
- `basic-parser` and `enhanced-parser` show high similarity (shared "parse")
- Tools with no common transforms show 0 similarity

---

### **Phase B: Virtual Path Resolution**
**Files to Modify:**
- `daemon/storage.go` - Add similar path handling

**Key Changes:**
1. Extend `handleVirtualPath()` to detect `/similar/` patterns
2. Add `handleSimilarView()` function
3. Integrate with existing virtual filesystem

**Implementation:**
```go
// In daemon/storage.go, add to handleVirtualPath()
if strings.HasPrefix(path, "/similar/") {
    return s.handleSimilarView(path)
}

func (s *Storage) handleSimilarView(path string) []map[string]interface{} {
    entries := []map[string]interface{}{}
    
    // Extract tool name from path: /similar/log-analyzer -> "log-analyzer"  
    pathParts := strings.Split(strings.Trim(path, "/"), "/")
    if len(pathParts) < 2 {
        return entries
    }
    
    toolName := pathParts[1]
    
    // Find the target tool
    targetTool := s.findToolByName(toolName)
    if targetTool == nil {
        return entries
    }
    
    // Calculate similarities
    calculator := NewSimilarityCalculator(s.relationStore)
    similarTools := calculator.findSimilarTools(*targetTool, 0.3) // 30% threshold
    
    // Convert to virtual nodes
    for _, simTool := range similarTools {
        entry := map[string]interface{}{
            "name":        simTool.Tool.Properties["name"],
            "type":        "file",
            "similarity":  simTool.Similarity,
            "reasons":     simTool.Reason,
            "transforms":  simTool.Tool.Properties["transforms"],
            "created":     simTool.Tool.CreatedAt,
        }
        entries = append(entries, entry)
    }
    
    return entries
}
```

**Test Success Criteria:**
- `port42 ls /similar/log-analyzer` shows related analysis tools
- Each similar tool shows similarity score and reasons
- Empty result for tools with no similar matches

---

### **Phase C: Relationship Storage Integration** 
**Files to Modify:**
- `daemon/similarity.go` - Add relationship creation
- `daemon/server.go` - Hook similarity detection into tool creation

**Key Features:**
1. Store `similar_to` relationships in relation system
2. Automatic similarity detection when tools are created
3. Bidirectional similarity relationships

**Implementation:**
```go
// In similarity.go
func (sc *SimilarityCalculator) createSimilarityRelationships(tool Relation) error {
    similarTools := sc.findSimilarTools(tool, 0.5) // 50% threshold for relationships
    
    for _, simTool := range similarTools {
        // Create bidirectional relationship
        relationship := Relation{
            ID:   fmt.Sprintf("similarity-%s-%s", tool.ID, simTool.Tool.ID),
            Type: "Relationship",
            Properties: map[string]interface{}{
                "relationship_type": "similar_to",
                "from":             tool.ID,
                "to":               simTool.Tool.ID,
                "similarity_score": simTool.Similarity,
                "reasons":          simTool.Reason,
            },
            CreatedAt: time.Now(),
        }
        
        err := sc.relationStore.Store(relationship)
        if err != nil {
            return err
        }
    }
    
    return nil
}

// In server.go handleDeclareRelation, add after materialization:
if d.similarityCalculator != nil {
    go func() {
        err := d.similarityCalculator.createSimilarityRelationships(payload.Relation)
        if err != nil {
            log.Printf("Failed to create similarity relationships: %v", err)
        }
    }()
}
```

**Test Success Criteria:**
- New tools automatically get similarity relationships
- Relationships are bidirectional (A similar to B, B similar to A)
- Background processing doesn't block tool creation

---

### **Phase D: Advanced Discovery Features**
**Optional enhancements for comprehensive tool discovery:**

1. **Similarity Threshold Configuration**
   - Configurable similarity thresholds
   - Different thresholds for path resolution vs relationship creation

2. **Multi-dimensional Similarity**
   - Consider tool naming patterns
   - Factor in creation timestamps (recent tools more similar)
   - Include file extension patterns from executables

3. **Similarity Explanation**
   - Detailed reasoning for why tools are similar
   - Visual similarity scores in CLI output

4. **Discovery Recommendations**
   - `port42 discover` command to find underutilized similar tools
   - Recommendations when creating new tools

---

## **Testing Strategy**

### **Unit Tests**
```go
// daemon/similarity_test.go
func TestTransformSimilarity(t *testing.T) {
    // Test exact matches
    assert.Equal(t, 1.0, calculateTransformSimilarity([]string{"analyze", "test"}, []string{"analyze", "test"}))
    
    // Test partial matches
    assert.Equal(t, 0.5, calculateTransformSimilarity([]string{"analyze", "test"}, []string{"analyze", "format"}))
    
    // Test semantic matches
    assert.Greater(t, calculateTransformSimilarity([]string{"analyze"}, []string{"analysis"}), 0.5)
    
    // Test no matches
    assert.Equal(t, 0.0, calculateTransformSimilarity([]string{"parse"}, []string{"music"}))
}

func TestSimilarToolDetection(t *testing.T) {
    // Create test tools with various transform combinations
    // Verify similarity detection accuracy
    // Test threshold filtering
}
```

### **Integration Tests**
```bash
#!/bin/bash
# test_semantic_discovery.sh

# Create test tools with known similarity patterns
port42 declare tool test-log-analyzer --transforms log,analyze,pattern
port42 declare tool test-data-analyzer --transforms data,analyze,inspect  
port42 declare tool test-quick-parser --transforms parse,quick,text
port42 declare tool test-smart-parser --transforms parse,smart,advanced

# Test similarity detection
echo "Testing log analyzer similarity..."
SIMILAR_LOGS=$(port42 ls /similar/test-log-analyzer | grep analyzer | wc -l)
if [ $SIMILAR_LOGS -gt 0 ]; then
    echo "✅ Log analyzer found similar tools"
else
    echo "❌ No similar tools found for log analyzer"
fi

echo "Testing parser similarity..."  
SIMILAR_PARSERS=$(port42 ls /similar/test-quick-parser | grep parser | wc -l)
if [ $SIMILAR_PARSERS -gt 0 ]; then
    echo "✅ Parser found similar tools"
else
    echo "❌ No similar tools found for parser"
fi

echo "Testing cross-category similarity..."
MIXED_SIMILAR=$(port42 ls /similar/test-log-analyzer | grep parser | wc -l)
if [ $MIXED_SIMILAR -eq 0 ]; then
    echo "✅ Analyzers and parsers correctly NOT similar"
else
    echo "❌ False positive: analyzer showing parser similarity"
fi
```

### **Performance Tests**
- Similarity calculation with 100+ tools
- Virtual path resolution response time
- Background relationship creation efficiency

---

## **Success Metrics**

### **User Experience Goals**
1. **Discovery**: Find 3+ similar tools when looking at `/similar/{popular-tool}`
2. **Accuracy**: >80% of similar tools are genuinely useful/related
3. **Performance**: Similarity queries return within 200ms
4. **Comprehensiveness**: Major tool categories (analyzers, parsers, formatters) show interconnections

### **Technical Goals** 
1. **Algorithm Accuracy**: Similarity scores match human intuition for tool relationships
2. **Coverage**: >70% of tools have at least one similar match above 30% threshold
3. **Integration**: Seamless virtual filesystem integration with existing paths
4. **Scalability**: Handle 500+ tools without performance degradation

---

## **Implementation Priority**

**Priority 1 (MVP)**:
- Phase A: Core similarity algorithm
- Phase B: Virtual path resolution  
- Basic testing and validation

**Priority 2 (Full Step 6)**:
- Phase C: Relationship storage
- Integration tests
- Performance optimization

**Priority 3 (Advanced Features)**:  
- Phase D: Discovery enhancements
- Rich CLI output formatting
- Recommendation system

---

## **Files to Create/Modify**

### **New Files**
- `daemon/similarity.go` - Core similarity calculation logic
- `daemon/similarity_test.go` - Unit tests
- `test_semantic_discovery.sh` - Integration test suite

### **Modified Files**
- `daemon/storage.go` - Add similar path resolution
- `daemon/server.go` - Hook similarity detection into tool creation
- `docs/working/incremental-reality-compiler-plan.md` - Update Step 6 status

---

## **Risk Mitigation**

### **Performance Risks**
- **Risk**: Similarity calculation slow with many tools
- **Mitigation**: Cache similarity results, use background processing
- **Monitoring**: Add timing logs for similarity operations

### **Accuracy Risks**
- **Risk**: False positive similar tools confuse users
- **Mitigation**: Tunable thresholds, semantic transform analysis
- **Monitoring**: User feedback on similarity relevance

### **Integration Risks**
- **Risk**: New similarity code breaks existing virtual filesystem
- **Mitigation**: Comprehensive testing, gradual rollout
- **Monitoring**: Virtual path resolution performance metrics

---

This plan provides a comprehensive roadmap for implementing Step 6 Semantic Tool Discovery while building on the existing Port 42 architecture and maintaining system performance and reliability.