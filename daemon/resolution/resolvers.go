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

// searchResolver handles search queries
type searchResolver struct {
	handler func(query string, limit int) ([]SearchResult, error)
}

func (r *searchResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
	results, err := r.handler(target, 5)
	if err != nil {
		return &ResolvedContext{
			Type:    "search",
			Target:  target,
			Success: false,
			Error:   err.Error(),
		}, nil // Return successful response with error details
	}
	
	content := formatSearchResults(target, results)
	
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

// memoryResolver handles memory session lookups
type memoryResolver struct {
	handler func(sessionID string) (*MemorySession, error)
}

func (r *memoryResolver) resolve(ctx context.Context, target string) (*ResolvedContext, error) {
	session, err := r.handler(target)
	if err != nil {
		return &ResolvedContext{
			Type:    "memory",
			Target:  target,
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	content := formatMemorySession(session)
	
	return &ResolvedContext{
		Type:    "memory",
		Target:  target,
		Content: content,
		Success: true,
	}, nil
}

func (r *memoryResolver) getTimeout() time.Duration {
	return 4 * time.Second
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
	
	// Generate artifact ID
	artifactID := NewURLArtifactID(target).Generate()
	
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
				"fetched_at":     now.Unix(),
				"cache_version":  2,
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

func formatMemorySession(session *MemorySession) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Memory Session: %s", session.SessionID))
	
	if session.Agent != "" {
		parts = append(parts, fmt.Sprintf("Agent: %s", session.Agent))
	}
	
	if len(session.Messages) > 0 {
		parts = append(parts, fmt.Sprintf("Messages: %d", len(session.Messages)))
		
		// Show last 2 messages as context
		start := len(session.Messages) - 2
		if start < 0 {
			start = 0
		}
		
		for i := start; i < len(session.Messages) && i < start+2; i++ {
			msg := session.Messages[i]
			content := msg.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			parts = append(parts, fmt.Sprintf("[%s] %s", strings.ToUpper(msg.Role), content))
		}
	}
	
	return strings.Join(parts, "\n")
}

func formatFileContent(file *FileContent) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("File: %s", file.Path))
	parts = append(parts, fmt.Sprintf("Size: %d bytes", file.Size))
	
	content := strings.TrimSpace(file.Content)
	if len(content) > 1000 {
		content = content[:1000] + "\n[Content truncated]"
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