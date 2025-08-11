package resolution

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// service implements ResolutionService
type service struct {
	resolvers map[string]resolver
}

// resolver interface for individual reference type resolvers
type resolver interface {
	resolve(ctx context.Context, target string) (*ResolvedContext, error)
	getTimeout() time.Duration
}

// newService creates a service with all resolvers
func newService(handlers Handlers) *service {
	s := &service{
		resolvers: make(map[string]resolver),
	}
	
	// Register resolvers with handlers
	if handlers.SearchHandler != nil {
		s.resolvers["search"] = &searchResolver{handler: handlers.SearchHandler}
	}
	if handlers.ToolHandler != nil {
		s.resolvers["tool"] = &toolResolver{handler: handlers.ToolHandler}
	}
	if handlers.FileHandler != nil {
		s.resolvers["file"] = &fileResolver{handler: handlers.FileHandler}
	}
	if handlers.P42Handler != nil {
		s.resolvers["p42"] = &p42Resolver{handler: handlers.P42Handler}
	}
	
	// URL resolver with artifact management
	var relations RelationsManager
	if handlers.RelationsHandler != nil {
		relations = handlers.RelationsHandler()
	}
	artifactManager := NewArtifactManager(relations)
	s.resolvers["url"] = &urlResolver{
		relations:       relations,
		artifactManager: artifactManager,
	}
	
	log.Printf("🔗 Resolution service initialized with %d resolvers", len(s.resolvers))
	return s
}

// ResolveForAI resolves references and formats for AI, returning both formatted string and contexts
func (s *service) ResolveForAI(references []Reference) (string, []*ResolvedContext, error) {
	if len(references) == 0 {
		return "", nil, nil
	}
	
	log.Printf("🔍 Resolving %d references for AI", len(references))
	
	// Resolve all references
	contexts := s.resolveAll(references)
	
	// Format for AI consumption
	formatted := s.formatForAI(contexts)
	
	return formatted, contexts, nil
}

// GetResolutionStats returns resolution statistics (DEPRECATED: use ComputeStatsFromContexts)
func (s *service) GetResolutionStats(references []Reference) (*Stats, error) {
	if len(references) == 0 {
		return &Stats{}, nil
	}
	
	contexts := s.resolveAll(references)
	return s.ComputeStatsFromContexts(references, contexts), nil
}

// ComputeStatsFromContexts computes stats from already resolved contexts (avoids duplicate resolution)
func (s *service) ComputeStatsFromContexts(references []Reference, contexts []*ResolvedContext) *Stats {
	stats := &Stats{
		TotalReferences: len(references),
		TypeBreakdown:   make(map[string]int),
	}
	
	for _, ctx := range contexts {
		stats.TypeBreakdown[ctx.Type]++
		if ctx.Success {
			stats.ResolvedCount++
			stats.TotalContentSize += len(ctx.Content)
		} else {
			stats.FailedCount++
		}
	}
	
	if stats.TotalReferences > 0 {
		stats.SuccessRate = float64(stats.ResolvedCount) / float64(stats.TotalReferences) * 100
	}
	
	return stats
}

// resolveAll resolves all references with timeouts and error handling
func (s *service) resolveAll(references []Reference) []*ResolvedContext {
	var results []*ResolvedContext
	
	for _, ref := range references {
		resolver, exists := s.resolvers[ref.Type]
		if !exists {
			results = append(results, &ResolvedContext{
				Type:    ref.Type,
				Target:  ref.Target,
				Success: false,
				Error:   fmt.Sprintf("No resolver for type: %s", ref.Type),
			})
			continue
		}
		
		// Create timeout context
		ctx, cancel := context.WithTimeout(context.Background(), resolver.getTimeout())
		
		resolved, err := resolver.resolve(ctx, ref.Target)
		cancel()
		
		if err != nil {
			results = append(results, &ResolvedContext{
				Type:    ref.Type,
				Target:  ref.Target,
				Success: false,
				Error:   err.Error(),
			})
		} else {
			results = append(results, resolved)
		}
	}
	
	return results
}

// formatForAI formats resolved contexts for AI consumption
func (s *service) formatForAI(contexts []*ResolvedContext) string {
	var parts []string
	
	// Only include successful resolutions
	var successful []*ResolvedContext
	for _, ctx := range contexts {
		if ctx.Success && len(ctx.Content) > 0 {
			successful = append(successful, ctx)
		}
	}
	
	if len(successful) == 0 {
		return ""
	}
	
	parts = append(parts, "CONTEXTUAL INFORMATION:")
	
	// Apply simple size limiting (8KB total)
	totalSize := 0
	maxSize := 8 * 1024
	
	for _, ctx := range successful {
		content := ctx.Content
		
		// Limit individual context size
		if len(content) > 2000 {
			content = content[:2000] + "\n[Content truncated for size]"
		}
		
		if totalSize+len(content) > maxSize {
			parts = append(parts, "\n[Additional references omitted due to size limit]")
			break
		}
		
		contextBlock := fmt.Sprintf("\n%s Reference (%s):\n%s\n", 
			strings.Title(ctx.Type), ctx.Target, content)
		
		parts = append(parts, contextBlock)
		totalSize += len(contextBlock)
	}
	
	if len(parts) == 1 {
		return "" // Only header
	}
	
	parts = append(parts, "\nUse this contextual information to generate more relevant tools.\n")
	
	result := strings.Join(parts, "")
	log.Printf("✨ AI context formatted: %d chars from %d successful references", 
		len(result), len(successful))
	
	return result
}