package resolution

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Search filtering constants
const (
	SearchScoreThreshold = 2.0   // Minimum score for including results
	SearchMaxResults     = 5     // Maximum number of results to load full content
	SearchContentLimit   = 20000 // Maximum content size before truncation
)

// searchResolver handles search queries
type searchResolver struct {
	handler    func(query string, limit int) ([]SearchResult, error)
	p42Handler func(p42Path string) (*FileContent, error) // For loading full content
}

func (r *searchResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
	// Get search results with higher limit for filtering
	results, err := r.handler(target, 20)
	if err != nil {
		return &ResolvedContext{
			Type:    "search",
			Target:  target,
			Success: false,
			Error:   err.Error(),
		}, nil // Return successful response with error details
	}
	
	// Apply score filtering and content loading
	content := r.formatSearchResultsWithContent(target, results)
	
	return &ResolvedContext{
		Type:    "search",
		Target:  target,
		Content: content,
		Success: true,
	}, nil
}

func (r *searchResolver) getTimeout() time.Duration {
	return 5 * time.Second
}

// toolResolver handles tool lookups
type toolResolver struct {
	handler func(toolName string) (*ToolDefinition, error)
}

func (r *toolResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
	toolDef, err := r.handler(target)
	if err != nil {
		return &ResolvedContext{
			Type:    "tool",
			Target:  target,
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	// Handle graceful degradation: (nil, nil) means tool not found but don't fail resolution
	if toolDef == nil {
		return &ResolvedContext{
			Type:    "tool",
			Target:  target,
			Success: false,
			Error:   "Tool not found",
		}, nil
	}
	
	content := formatToolDefinition(toolDef)
	
	return &ResolvedContext{
		Type:    "tool",
		Target:  target,
		Content: content,
		Success: true,
	}, nil
}

func (r *toolResolver) getTimeout() time.Duration {
	return 3 * time.Second
}


// fileResolver handles file content loading
type fileResolver struct {
	handler func(path string) (*FileContent, error)
}

func (r *fileResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
	fileContent, err := r.handler(target)
	if err != nil {
		return &ResolvedContext{
			Type:    "file",
			Target:  target,
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	content := formatFileContent(fileContent)
	
	return &ResolvedContext{
		Type:    "file",
		Target:  target,
		Content: content,
		Success: true,
	}, nil
}

func (r *fileResolver) getTimeout() time.Duration {
	return 3 * time.Second
}

// p42Resolver handles Port 42 VFS (Virtual File System) access
type p42Resolver struct {
	handler func(p42Path string) (*FileContent, error)
}

func (r *p42Resolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
	fileContent, err := r.handler(target)
	if err != nil {
		return &ResolvedContext{
			Type:    "p42",
			Target:  target,
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	content := formatP42Content(fileContent)
	
	return &ResolvedContext{
		Type:    "p42",
		Target:  target,
		Content: content,
		Success: true,
	}, nil
}

func (r *p42Resolver) getTimeout() time.Duration {
	return 5 * time.Second // Slightly longer for VFS operations
}

// urlResolver handles URL fetching with artifact caching
type urlResolver struct {
	relations       RelationsManager // Relations for URL artifact caching
	artifactManager *ArtifactManager // Artifact lifecycle management
}

func (r *urlResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
	// Validate URL using new validation
	if !IsValidURL(target) {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   "Invalid URL format",
		}, nil
	}
	
	// Generate artifact ID (deterministic for caching)
	artifactID := NewURLArtifactID(target).Generate()
	
	// DEBUG: For now, force fresh artifacts to test cache behavior
	// TODO: Remove this and fix Relations update behavior
	if r.artifactManager != nil {
		log.Printf("🔍 DEBUG: Checking cache for %s", artifactID)
	}
	
	// Phase 3: Enhanced Resolution Flow - Cache-first with proper fallback logic
	if r.artifactManager != nil {
		// Try cache first  
		if cached, err := r.artifactManager.LoadCached(artifactID); err == nil && cached != nil {
			// Cache hit - successful cache-first resolution
			log.Printf("🎯 URL cache HIT: %s -> %s", target, artifactID)
			content := r.formatCachedURLContent(cached.Content, cached.Properties, target)
			return &ResolvedContext{
				Type:    "url",
				Target:  target,
				Content: content,
				Success: true,
			}, nil
		}
		
		// Cache miss - proceed to fetch with caching enabled
		log.Printf("🌐 URL cache MISS: %s -> fetching fresh (will cache)", target)
		return r.fetchAndStore(ctx, target, artifactID)
	} else {
		// No cache manager - direct fetch without caching
		log.Printf("🌐 URL direct fetch: %s (no cache available)", target)
		return r.fetchWithoutCaching(ctx, target)
	}
}

// fetchAndStore fetches URL content and stores as artifact if possible
func (r *urlResolver) fetchAndStore(ctx context.Context, target, artifactID string) (*ResolvedContext, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}
	
	req.Header.Set("User-Agent", "Port42-ReferenceResolver/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("HTTP request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}
	
	// Read with size limit
	limitedReader := io.LimitReader(resp.Body, 50*1024) // 50KB limit
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("Failed to read response: %v", err),
		}, nil
	}
	
	content := string(bodyBytes)
	
	// Try to store as artifact if artifact manager is available
	if r.artifactManager != nil {
		now := time.Now()
		
		// CRITICAL FIX: Always use current timestamp for fresh fetches
		// This ensures Relations updates don't preserve stale metadata
		freshTimestamp := now.Unix()
		
		artifact := &URLArtifactRelation{
			ID:        artifactID,
			Type:      "URLArtifact",
			Content:   content,
			CreatedAt: now,
			UpdatedAt: now,
			Properties: map[string]interface{}{
				"source_url":     target,
				"content_type":   resp.Header.Get("Content-Type"),
				"status_code":    resp.StatusCode,
				"content_length": len(content),
				"fetched_at":     freshTimestamp, // Always current time
				"cache_version":  3,              // Increment for timestamp fix
				"last_updated":   freshTimestamp, // Always current time
				"debug_fetched":  now.Format("2006-01-02 15:04:05"), // Human readable debug
			},
		}
		
		// Store artifact (errors are logged but don't fail resolution)
		r.artifactManager.Store(artifact)
	}
	
	formattedContent := formatURLContent(content, resp.Header.Get("Content-Type"), target)
	formattedContent += "\n[Freshly fetched]"
	
	return &ResolvedContext{
		Type:    "url",
		Target:  target,
		Content: formattedContent,
		Success: true,
	}, nil
}

// formatCachedURLContent formats cached URL content with cache indicator
func (r *urlResolver) formatCachedURLContent(content string, properties map[string]interface{}, url string) string {
	contentType, _ := properties["content_type"].(string)
	fetchedAt, _ := properties["fetched_at"].(int64)
	
	formattedContent := formatURLContent(content, contentType, url)
	
	// Add cache indicator
	if fetchedAt > 0 {
		fetchTime := time.Unix(fetchedAt, 0)
		formattedContent += fmt.Sprintf("\n[Cached from %s]", fetchTime.Format("2006-01-02 15:04:05"))
	} else {
		formattedContent += "\n[From cache]"
	}
	
	return formattedContent
}

// fetchWithoutCaching performs direct HTTP fetch without any caching (graceful degradation)
func (r *urlResolver) fetchWithoutCaching(ctx context.Context, target string) (*ResolvedContext, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}
	
	req.Header.Set("User-Agent", "Port42-ReferenceResolver/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("HTTP request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}
	
	// Read with size limit
	limitedReader := io.LimitReader(resp.Body, 50*1024) // 50KB limit
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return &ResolvedContext{
			Type:    "url",
			Target:  target,
			Success: false,
			Error:   fmt.Sprintf("Failed to read response: %v", err),
		}, nil
	}
	
	content := string(bodyBytes)
	formattedContent := formatURLContent(content, resp.Header.Get("Content-Type"), target)
	formattedContent += "\n[Direct fetch - no caching]"
	
	return &ResolvedContext{
		Type:    "url",
		Target:  target,
		Content: formattedContent,
		Success: true,
	}, nil
}

func (r *urlResolver) getTimeout() time.Duration {
	return 10 * time.Second
}

// Formatting functions
func formatSearchResults(query string, results []SearchResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No results found for search query: '%s'", query)
	}
	
	var parts []string
	parts = append(parts, fmt.Sprintf("Search results for '%s' (%d results):", query, len(results)))
	
	for i, result := range results {
		if i >= 3 { // Limit to top 3 results
			parts = append(parts, fmt.Sprintf("... and %d more results", len(results)-3))
			break
		}
		
		resultText := fmt.Sprintf("\n%d. %s (Score: %.2f)", i+1, result.Path, result.Score)
		if result.Summary != "" {
			summary := result.Summary
			if len(summary) > 150 {
				summary = summary[:150] + "..."
			}
			resultText += fmt.Sprintf("\n   Summary: %s", summary)
		}
		parts = append(parts, resultText)
	}
	
	return strings.Join(parts, "")
}

// formatSearchResultsWithContent applies score filtering and loads full content
func (r *searchResolver) formatSearchResultsWithContent(query string, results []SearchResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No results found for search query: '%s'", query)
	}
	
	// Filter by score threshold
	var filteredResults []SearchResult
	for _, result := range results {
		if result.Score >= SearchScoreThreshold {
			filteredResults = append(filteredResults, result)
		}
	}
	
	// Apply limit
	if len(filteredResults) > SearchMaxResults {
		filteredResults = filteredResults[:SearchMaxResults]
	}
	
	if len(filteredResults) == 0 {
		return fmt.Sprintf("Search results for '%s': Found %d results, but none met the relevance threshold (score ≥ %.1f)", 
			query, len(results), SearchScoreThreshold)
	}
	
	var parts []string
	parts = append(parts, fmt.Sprintf("=== Search Results: search:%s ===", query))
	parts = append(parts, fmt.Sprintf("Found %d high-relevance results:", len(filteredResults)))
	parts = append(parts, "")
	
	for i, result := range filteredResults {
		// Add result header
		parts = append(parts, fmt.Sprintf("%d. %s (score: %.2f)", i+1, result.Path, result.Score))
		
		// Load full content if P42Handler is available
		if r.p42Handler != nil {
			if fileContent, err := r.p42Handler(result.Path); err == nil {
				// Add full content with appropriate formatting
				if result.Type == "tool" {
					parts = append(parts, "[FULL TOOL DEFINITION]")
				} else if result.Type == "memory" || result.Type == "session" {
					parts = append(parts, "[FULL CONVERSATION TRANSCRIPT]")
				} else {
					parts = append(parts, "[FULL CONTENT]")
				}
				
				// Add the actual content (truncate if too large)
				content := fileContent.Content
				if len(content) > SearchContentLimit {
					content = content[:SearchContentLimit] + fmt.Sprintf("\n\n[Content truncated - full content exceeds %d characters]", SearchContentLimit)
				}
				parts = append(parts, content)
			} else {
				// Fallback to summary if content loading fails
				parts = append(parts, "[Content loading failed, showing summary]")
				if result.Summary != "" {
					summary := result.Summary
					if len(summary) > 150 {
						summary = summary[:150] + "..."
					}
					parts = append(parts, fmt.Sprintf("Summary: %s", summary))
				}
			}
		} else {
			// Fallback to summary if no P42Handler available
			if result.Summary != "" {
				summary := result.Summary
				if len(summary) > 150 {
					summary = summary[:150] + "..."
				}
				parts = append(parts, fmt.Sprintf("Summary: %s", summary))
			}
		}
		
		parts = append(parts, "") // Add blank line between results
	}
	
	return strings.Join(parts, "\n")
}

func formatToolDefinition(tool *ToolDefinition) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Tool Definition: %s", tool.Name))
	parts = append(parts, fmt.Sprintf("ID: %s", tool.ID))
	
	if len(tool.Transforms) > 0 {
		parts = append(parts, fmt.Sprintf("Transforms: %s", strings.Join(tool.Transforms, ", ")))
	}
	
	if len(tool.Commands) > 0 {
		parts = append(parts, fmt.Sprintf("Generated Commands: %d available", len(tool.Commands)))
		// Show first command as example
		if len(tool.Commands[0]) < 200 {
			parts = append(parts, fmt.Sprintf("Example: %s", tool.Commands[0]))
		}
	}
	
	return strings.Join(parts, "\n")
}


func formatFileContent(file *FileContent) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Local File: %s", file.Path))
	parts = append(parts, fmt.Sprintf("Type: %s", file.Type))
	parts = append(parts, fmt.Sprintf("Size: %d bytes", file.Size))
	
	// Add file metadata if available
	if file.Metadata != nil {
		if modified, exists := file.Metadata["modified"]; exists {
			parts = append(parts, fmt.Sprintf("Modified: %v", modified))
		}
	}
	
	content := strings.TrimSpace(file.Content)
	if len(content) > 1000 {
		content = content[:1000] + "\n[Content truncated - showing first 1000 chars]"
	}
	
	parts = append(parts, fmt.Sprintf("Content:\n%s", content))
	
	return strings.Join(parts, "\n")
}

// formatP42Content formats Port 42 VFS content
func formatP42Content(file *FileContent) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Port 42 VFS: %s", file.Path))
	parts = append(parts, fmt.Sprintf("Type: %s", file.Type))
	parts = append(parts, fmt.Sprintf("Size: %d bytes", file.Size))
	
	// Add P42-specific metadata if available
	if file.Metadata != nil {
		if relationID, exists := file.Metadata["relation_id"]; exists {
			parts = append(parts, fmt.Sprintf("Relation ID: %v", relationID))
		}
		if storagePath, exists := file.Metadata["storage_path"]; exists {
			parts = append(parts, fmt.Sprintf("Storage Path: %v", storagePath))
		}
		if score, exists := file.Metadata["score"]; exists {
			parts = append(parts, fmt.Sprintf("Match Score: %.2f", score))
		}
		if created, exists := file.Metadata["created"]; exists {
			parts = append(parts, fmt.Sprintf("Created: %v", created))
		}
	}
	
	content := strings.TrimSpace(file.Content)
	if len(content) > 800 {
		content = content[:800] + "\n[Content truncated - showing first 800 chars]"
	}
	
	parts = append(parts, fmt.Sprintf("Content:\n%s", content))
	
	return strings.Join(parts, "\n")
}

func formatURLContent(body, contentType, url string) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("URL: %s", url))
	parts = append(parts, fmt.Sprintf("Content-Type: %s", contentType))
	
	content := body
	if strings.Contains(contentType, "html") {
		content = extractTextFromHTML(content)
	}
	
	if len(content) > 800 {
		content = content[:800] + "\n[Content truncated]"
	}
	
	parts = append(parts, fmt.Sprintf("Content:\n%s", content))
	
	return strings.Join(parts, "\n")
}

// Utility functions

func extractTextFromHTML(html string) string {
	// Simple HTML text extraction
	content := html
	
	// Remove script and style tags
	scriptRegex := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")
	
	styleRegex := regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	content = styleRegex.ReplaceAllString(content, "")
	
	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	content = tagRegex.ReplaceAllString(content, " ")
	
	// Clean up whitespace
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	
	return strings.TrimSpace(content)
}