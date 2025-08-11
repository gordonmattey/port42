package resolution

import (
	"log"
	"time"
)

// ArtifactManager orchestrates URL artifact lifecycle with Relations
type ArtifactManager struct {
	relations RelationsManager
	policy    CachePolicy
}

// NewArtifactManager creates a new artifact manager
func NewArtifactManager(relations RelationsManager) *ArtifactManager {
	return &ArtifactManager{
		relations: relations,
		policy:    DefaultCachePolicy(),
	}
}

// LoadCached attempts to load a cached URL artifact
func (am *ArtifactManager) LoadCached(artifactID string) (*URLArtifactRelation, error) {
	if am.relations == nil {
		return nil, nil // No caching available - graceful degradation
	}
	
	// Try to load from Relations
	relation, err := am.relations.GetRelationByID(artifactID)
	if err != nil {
		// Cache miss - return nil without error for graceful degradation
		return nil, nil
	}
	
	// Check if cache is expired
	if am.IsExpired(relation) {
		fetchedAt, _ := relation.Properties["fetched_at"].(int64)
		age := time.Since(time.Unix(fetchedAt, 0))
		log.Printf("🕐 Cache EXPIRED: %s (age: %v, TTL: %v)", artifactID, age.Truncate(time.Second), am.policy.DefaultTTL)
		return nil, nil // Expired = cache miss
	}
	
	// Update access time for usage tracking
	am.updateLastAccessed(relation)
	
	// Log successful cache hit with age info
	fetchedAt, _ := relation.Properties["fetched_at"].(int64)
	age := time.Since(time.Unix(fetchedAt, 0))
	log.Printf("✅ Cache VALID: %s (age: %v, TTL: %v)", artifactID, age.Truncate(time.Second), am.policy.DefaultTTL)
	return relation, nil
}

// Store saves a URL artifact to Relations if it should be cached
func (am *ArtifactManager) Store(artifact *URLArtifactRelation) error {
	if am.relations == nil {
		return nil // No caching available - graceful degradation
	}
	
	// Check if we should cache this artifact
	if !am.shouldStore(artifact) {
		log.Printf("⚠️ Skipping cache storage for artifact %s (policy rejection)", artifact.ID)
		return nil
	}
	
	// Store in Relations
	err := am.relations.DeclareRelation(artifact)
	if err != nil {
		log.Printf("⚠️ Failed to store artifact %s: %v", artifact.ID, err)
		return err
	}
	
	sourceURL, _ := artifact.Properties["source_url"].(string)
	contentLength, _ := artifact.Properties["content_length"].(int)
	log.Printf("💾 Cached URL artifact: %s (%s, %d bytes)", artifact.ID, sourceURL, contentLength)
	return nil
}

// IsExpired checks if an artifact is expired according to cache policy
func (am *ArtifactManager) IsExpired(artifact *URLArtifactRelation) bool {
	fetchedAt, exists := artifact.Properties["fetched_at"].(int64)
	if !exists {
		return true // No timestamp means expired
	}
	
	return am.policy.IsExpired(time.Unix(fetchedAt, 0))
}

// UpdateUsage tracks usage of a cached artifact for analytics
func (am *ArtifactManager) UpdateUsage(artifactID string, referenceContext *ReferenceContext) error {
	if am.relations == nil {
		return nil // No tracking without Relations
	}
	
	// Load current artifact
	artifact, err := am.relations.GetRelationByID(artifactID)
	if err != nil {
		return nil // Artifact doesn't exist - don't error
	}
	
	// Update usage stats
	now := time.Now()
	artifact.Properties["last_accessed"] = now.Unix()
	
	// Increment access count
	if count, exists := artifact.Properties["access_count"].(int); exists {
		artifact.Properties["access_count"] = count + 1
	} else {
		artifact.Properties["access_count"] = 1
	}
	
	// Track reference context if provided
	if referenceContext != nil && referenceContext.RelationID != "" {
		// Add to list of relations that used this URL
		var usedBy []string
		if existing, exists := artifact.Properties["used_by_relations"].([]string); exists {
			usedBy = existing
		}
		
		// Add relation if not already present
		found := false
		for _, relation := range usedBy {
			if relation == referenceContext.RelationID {
				found = true
				break
			}
		}
		if !found {
			usedBy = append(usedBy, referenceContext.RelationID)
			artifact.Properties["used_by_relations"] = usedBy
		}
	}
	
	artifact.UpdatedAt = now
	
	// Save updated artifact
	return am.relations.DeclareRelation(artifact)
}

// shouldStore determines if an artifact should be cached based on policy
func (am *ArtifactManager) shouldStore(artifact *URLArtifactRelation) bool {
	sourceURL, _ := artifact.Properties["source_url"].(string)
	statusCode, _ := artifact.Properties["status_code"].(int)
	contentLength, _ := artifact.Properties["content_length"].(int)
	contentType, _ := artifact.Properties["content_type"].(string)
	
	return am.policy.ShouldCache(sourceURL, statusCode, int64(contentLength), contentType)
}

// updateLastAccessed updates the last accessed timestamp (internal helper)
func (am *ArtifactManager) updateLastAccessed(artifact *URLArtifactRelation) {
	if artifact.Properties == nil {
		artifact.Properties = make(map[string]interface{})
	}
	artifact.Properties["last_accessed"] = time.Now().Unix()
	
	// Update in background to avoid blocking resolution
	go func() {
		if err := am.relations.DeclareRelation(artifact); err != nil {
			log.Printf("⚠️ Failed to update last_accessed for %s: %v", artifact.ID, err)
		}
	}()
}