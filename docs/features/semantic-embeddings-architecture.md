# Port 42 Semantic Embeddings Architecture

## 🎯 **Overview & Architecture**

**Goal**: Implement vector embeddings to enable semantic search, intelligent reference resolution, and content-aware AI context management.

**Core Philosophy**: Transform Port 42 from keyword-based similarity to semantic understanding while maintaining the existing VFS and relation architecture.

---

## 🏗️ **Architecture Design**

### **1. Embedding Service Layer**
```
┌─────────────────────────────────────────────────────────────┐
│                    Port 42 Embedding Architecture           │
├─────────────────┬─────────────────┬─────────────────────────┤
│   CLI Layer     │   Daemon Core   │   Embedding Services    │
├─────────────────┼─────────────────┼─────────────────────────┤
│ • Search UI     │ • VFS Engine    │ • Embedding Generator   │
│ • References    │ • Relations     │ • Vector Store          │
│ • Similarity    │ • Storage       │ • Similarity Engine     │
│   Commands      │   Manager       │ • Content Chunker       │
└─────────────────┴─────────────────┴─────────────────────────┘
```

### **2. Data Flow Architecture**
```
Tool Creation → Content Extraction → Chunking → Embedding → Vector Store
       ↓              ↓              ↓           ↓           ↓
   Relations  →   Raw Content  →  Chunks  →  Vectors  →  Index
       ↓              ↓              ↓           ↓           ↓
   VFS Paths  →   Similarity   →  Search  →  Context  →  AI Input
```

---

## 🔧 **Component Design**

### **A. Embedding Generator Service**
```go
// daemon/embedding/generator.go
type EmbeddingGenerator struct {
    client     EmbeddingClient
    chunker    *ContentChunker
    cache      *EmbeddingCache
    config     EmbeddingConfig
}

type EmbeddingClient interface {
    GenerateEmbeddings(texts []string) ([][]float32, error)
    GetDimensions() int
    GetModel() string
}

type EmbeddingConfig struct {
    Provider    string  // "openai", "anthropic", "local"
    Model       string  // "text-embedding-3-small", "text-embedding-ada-002"
    MaxTokens   int     // 8192 for OpenAI
    BatchSize   int     // 100 embeddings per batch
    Dimensions  int     // 1536 for OpenAI, configurable
}
```

### **B. Content Chunker**
```go
// daemon/embedding/chunker.go
type ContentChunker struct {
    maxChunkSize   int
    overlapSize    int
    preserveCode   bool
    splitStrategy  ChunkStrategy
}

type ChunkStrategy string
const (
    ByLines      ChunkStrategy = "lines"      // Split by line count
    ByTokens     ChunkStrategy = "tokens"     // Split by token count
    BySemantic   ChunkStrategy = "semantic"   // Split by code blocks/functions
    ByContent    ChunkStrategy = "content"    // Smart content-aware splitting
)

type ContentChunk struct {
    ID          string    `json:"id"`
    ParentID    string    `json:"parent_id"`     // Relation ID
    Content     string    `json:"content"`
    StartLine   int       `json:"start_line"`
    EndLine     int       `json:"end_line"`
    ChunkType   string    `json:"chunk_type"`    // "function", "class", "comment", "config"
    Language    string    `json:"language"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

### **C. Vector Store**
```go
// daemon/embedding/vector_store.go
type VectorStore struct {
    storage    Storage
    index      VectorIndex
    dimensions int
}

type VectorIndex interface {
    Add(id string, vector []float32, metadata map[string]interface{}) error
    Search(vector []float32, limit int, threshold float32) ([]VectorMatch, error)
    Delete(id string) error
    Update(id string, vector []float32, metadata map[string]interface{}) error
}

type VectorMatch struct {
    ID         string                 `json:"id"`
    Score      float32               `json:"score"`      // Cosine similarity
    Metadata   map[string]interface{} `json:"metadata"`
    Content    string                `json:"content"`
}

// Storage format: ~/.port42/embeddings/
// - index.json (vector index metadata)
// - vectors/ (binary vector storage)
// - chunks/ (chunk content and metadata)
```

### **D. Semantic Search Engine**
```go
// daemon/embedding/search.go
type SemanticSearchEngine struct {
    vectorStore    *VectorStore
    generator      *EmbeddingGenerator
    resultRanker   *ResultRanker
}

type SearchRequest struct {
    Query         string            `json:"query"`
    Filters       map[string]string `json:"filters"`      // language, type, etc.
    Limit         int              `json:"limit"`
    Threshold     float32          `json:"threshold"`    // Minimum similarity
    ContextType   string           `json:"context_type"` // "reference", "search", "similar"
}

type SearchResult struct {
    Chunks        []ContentChunk   `json:"chunks"`
    Aggregated    string          `json:"aggregated"`    // Combined relevant content
    Summary       string          `json:"summary"`       // AI-generated summary
    Confidence    float32         `json:"confidence"`
    TotalMatches  int             `json:"total_matches"`
}
```

---

## 🔌 **Integration Points**

### **1. VFS Integration** 
```go
// Extend existing VFS paths:
// /embeddings/                    - Root embeddings namespace
// /embeddings/by-similarity/      - Similarity-based discovery
// /embeddings/search/{query}/     - Semantic search results
// /embeddings/chunks/{relation}/  - Chunks for specific relation
// /embeddings/vectors/{chunk}/    - Vector data for chunk

// daemon/storage.go - Extend existing resolveVirtualPath
func (s *Storage) resolveVirtualPath(path string) []map[string]interface{} {
    // ... existing code ...
    
    if strings.HasPrefix(path, "/embeddings") {
        return s.handleEmbeddingsView(path)
    }
}
```

### **2. Relation System Integration**
```go
// Extend existing Relation type:
type Relation struct {
    // ... existing fields ...
    
    // Embedding metadata
    EmbeddingStatus  string    `json:"embedding_status"`  // "pending", "processing", "ready", "failed"
    ChunkCount      int       `json:"chunk_count"`
    LastEmbedded    time.Time `json:"last_embedded"`
    EmbeddingModel  string    `json:"embedding_model"`
    ContentHash     string    `json:"content_hash"`      // Detect content changes
}
```

### **3. Reference Resolution Integration**
```go
// cli/src/commands/possess.rs - Enhanced reference resolution
fn resolve_references_for_context(
    client: &mut DaemonClient, 
    references: Vec<String>
) -> Result<Vec<String>> {
    let mut contexts = Vec::new();
    
    for reference in references {
        // Enhanced resolution with semantic chunking
        let context = match reference {
            ref_str if ref_str.starts_with("p42:/commands/") => {
                // Use semantic search to find relevant chunks
                resolve_semantic_tool_reference(client, ref_str)?
            }
            _ => resolve_traditional_reference(client, reference)?
        };
        contexts.push(context);
    }
    
    Ok(contexts)
}

fn resolve_semantic_tool_reference(
    client: &mut DaemonClient,
    tool_ref: &str
) -> Result<String> {
    // Instead of loading full 333-line tool,
    // get semantically relevant chunks for the current task
    let request = DaemonRequest {
        request_type: "semantic_search".to_string(),
        payload: serde_json::json!({
            "target": tool_ref,
            "context_type": "reference",
            "max_chunks": 5,
            "relevance_threshold": 0.7
        }),
        // ...
    };
    
    // Returns summarized, relevant content instead of full tool
}
```

### **4. Search Command Integration**
```go
// Enhanced search with semantic capabilities
func (d *Daemon) handleSearch(req Request) Response {
    // ... existing code ...
    
    // Add semantic search option
    if useSemanticSearch {
        semanticResults := d.embeddingEngine.Search(SearchRequest{
            Query: query,
            Limit: limit,
            Threshold: 0.6,
        })
        
        // Merge with traditional keyword results
        results = mergeSearchResults(keywordResults, semanticResults)
    }
}
```

---

## 📦 **Dependencies & External Services**

### **A. Embedding Providers**

**Option 1: OpenAI Embeddings**
```go
// daemon/embedding/providers/openai.go
type OpenAIEmbeddingClient struct {
    apiKey    string
    baseURL   string
    model     string    // "text-embedding-3-small" (1536 dims, cheap)
    timeout   time.Duration
    rateLimit *RateLimiter
}
```

**Option 2: Local Embeddings (Offline)**
```go
// daemon/embedding/providers/local.go  
type LocalEmbeddingClient struct {
    modelPath   string    // Path to ONNX/TensorFlow model
    processor   *ModelProcessor
    dimensions  int
}
// Use sentence-transformers models, ONNX runtime
```

**Option 3: Anthropic (Future)**
```go
// When Anthropic releases embedding API
type AnthropicEmbeddingClient struct {
    apiKey  string
    model   string
}
```

### **B. Vector Index Libraries**

**Option 1: In-Memory with Persistence**
```go
// Simple flat index with JSON persistence
type FlatVectorIndex struct {
    vectors   map[string][]float32
    metadata  map[string]map[string]interface{}
    filePath  string
}
```

**Option 2: Faiss Integration (Advanced)**
```bash
# CGO dependency for high-performance vector search
go get github.com/DataIntelligenceCrew/go-faiss
```

**Option 3: SQLite with Vector Extension**
```bash
# Use sqlite-vss extension for vector similarity
go get github.com/mattn/go-sqlite3
```

### **C. Content Processing Dependencies**
```bash
# Text tokenization and chunking
go get github.com/tiktoken-go/tokenizer

# Programming language detection
go get github.com/go-enry/go-enry/v2

# Code parsing for semantic chunking
go get github.com/tree-sitter/tree-sitter-go
```

---

## 🚀 **Implementation Phases**

### **Phase 1: Foundation**
1. **Embedding service architecture** 
2. **Content chunker** with basic line-based splitting
3. **Simple vector store** with in-memory index
4. **OpenAI embedding client** integration

### **Phase 2: VFS Integration**
1. **Extend VFS** with `/embeddings/` namespace
2. **Relation embedding status** tracking
3. **Basic semantic search** endpoint
4. **CLI embedding commands** (`port42 embed`, `port42 search --semantic`)

### **Phase 3: Reference Enhancement**
1. **Semantic reference resolution** in possession mode
2. **Intelligent content summarization** for large tools
3. **Context-aware chunking** based on user queries
4. **Performance optimization** and caching

### **Phase 4: Advanced Features**
1. **Code-aware chunking** (functions, classes, modules)
2. **Multiple embedding providers** (local models)
3. **Advanced similarity algorithms** (hybrid keyword + semantic)
4. **Embedding-based tool discovery** and recommendations

---

## ⚖️ **Design Considerations**

### **A. Performance**
- **Lazy embedding generation** - only embed when needed
- **Incremental updates** - re-embed only changed content
- **Batch processing** - generate embeddings in batches
- **Caching strategy** - cache embeddings and search results

### **B. Storage Efficiency**
- **Quantized embeddings** - reduce from float32 to int8 when possible
- **Chunk deduplication** - avoid embedding identical code snippets
- **Compression** - compress vector storage files
- **Cleanup policies** - remove embeddings for deleted relations

### **C. Privacy & Security**
- **Local-first option** - support offline embedding models
- **Content filtering** - avoid embedding sensitive information
- **API key management** - secure storage of embedding service keys
- **Rate limiting** - respect embedding service limits

### **D. User Experience**
- **Transparent fallback** - fall back to keyword search if embeddings fail
- **Progressive enhancement** - work without embeddings, enhance with them
- **Clear feedback** - show when embeddings are being generated/used
- **Configuration options** - let users choose embedding providers

---

## 📊 **Success Metrics**

1. **Reference Resolution**: Reduce AI context size by 70% while maintaining relevance
2. **Search Quality**: Improve semantic search accuracy by 50% vs keyword-only
3. **Performance**: < 100ms for semantic search queries
4. **Storage**: < 10MB embedding storage per 1000 tools
5. **User Adoption**: 80% of power users use semantic features within first month

---

## 🎯 **Immediate Problem Resolution**

**Current Issue**: Tool references in possession mode overwhelm AI with 333+ lines of technical content, causing malformed command specs.

**Embedding Solution**:
1. **Chunk large tools** into semantic sections (functions, classes, configs)
2. **Vector search** for relevant chunks based on user's possession request
3. **Summarize context** - provide 3-5 most relevant chunks instead of full tool
4. **Maintain full access** - AI can still request full content if needed

**Example**: Instead of passing 333 lines of `apple-mail-sync`, provide:
- Function that handles permissions (30 lines)
- Error handling patterns (20 lines) 
- Configuration parsing logic (25 lines)
- **Total**: 75 lines of highly relevant context vs 333 lines of everything

This architecture provides a solid foundation for adding semantic intelligence to Port 42 while maintaining compatibility with existing systems and enabling powerful new capabilities for AI-assisted development.