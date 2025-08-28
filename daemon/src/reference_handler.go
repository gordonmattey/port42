package main

import (
	"fmt"
	"log"
	"port42/daemon/resolution"
)

// ReferenceResolutionResult contains the outcome of reference resolution
type ReferenceResolutionResult struct {
	Success      bool
	ResolvedText string
	Contexts     []*resolution.ResolvedContext
	Error        error
}

// ReferenceHandler provides common reference resolution functionality
type ReferenceHandler struct {
	resolutionService resolution.ResolutionService
}

// NewReferenceHandler creates a new reference handler
func NewReferenceHandler(resolutionService resolution.ResolutionService) *ReferenceHandler {
	return &ReferenceHandler{
		resolutionService: resolutionService,
	}
}

// ResolveReferences performs the common reference resolution logic
func (rh *ReferenceHandler) ResolveReferences(references []Reference, mode string) *ReferenceResolutionResult {
	result := &ReferenceResolutionResult{
		Success: false,
	}

	if len(references) == 0 {
		result.Success = true // No references is success
		return result
	}

	log.Printf("📎 Processing %d references for %s mode", len(references), mode)

	// Phase 1: Validate references
	if err := ValidateReferences(references); err != nil {
		log.Printf("⚠️ Invalid references in %s request: %v", mode, err)
		result.Error = fmt.Errorf("invalid references: %w", err)
		return result
	}

	// Log each reference for debugging
	for i, ref := range references {
		log.Printf("  Reference %d: %s:%s", i, ref.Type, ref.Target)
	}

	// Phase 2: Check resolution service availability
	if rh.resolutionService == nil {
		log.Printf("⚠️ No resolution service available - skipping reference resolution")
		result.Error = fmt.Errorf("resolution service not available")
		return result
	}

	log.Printf("🔍 Resolving references for %s context...", mode)

	// Phase 3: Convert protocol references to resolution references
	var resolutionRefs []resolution.Reference
	for _, ref := range references {
		resolutionRefs = append(resolutionRefs, resolution.Reference{
			Type:    ref.Type,
			Target:  ref.Target,
			Context: ref.Context,
		})
	}

	// Phase 4: Resolve references
	contextStr, contexts, err := rh.resolutionService.ResolveForAI(resolutionRefs)
	if err != nil {
		log.Printf("⚠️ Reference resolution failed: %v", err)
		result.Error = fmt.Errorf("resolution failed: %w", err)
		return result
	}

	// Phase 5: Process results
	result.Contexts = contexts
	result.ResolvedText = contextStr

	if contextStr != "" {
		log.Printf("✨ Resolved reference context (%d chars)", len(contextStr))
		
		// Compute and log resolution stats
		stats := rh.resolutionService.ComputeStatsFromContexts(resolutionRefs, contexts)
		log.Printf("📊 Reference resolution: %d/%d successful (%.1f%%)", 
			stats.ResolvedCount, stats.TotalReferences, stats.SuccessRate)
		
		result.Success = true
	} else {
		log.Printf("⚠️ No context resolved from references")
		result.Error = fmt.Errorf("no context resolved")
	}

	return result
}

// FormatForPossess formats resolved reference context for possess mode (inject into prompt)
func (rh *ReferenceHandler) FormatForPossess(resolvedText string) string {
	if resolvedText == "" {
		return ""
	}

	referenceSection := "\n\n--- REFERENCE CONTEXTS ---\n"
	referenceSection += "The following reference materials are provided for your use:\n\n"
	referenceSection += resolvedText
	referenceSection += "\n--- END REFERENCE CONTEXTS ---\n"
	
	return referenceSection
}

// FormatForDeclare formats resolved reference context for declare mode (store in properties)
func (rh *ReferenceHandler) FormatForDeclare(resolvedText string) string {
	// For declare mode, we store the raw resolved text
	return resolvedText
}