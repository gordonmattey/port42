package resolution

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"
)

// URLArtifactID generates deterministic IDs for URL artifacts
type URLArtifactID struct {
	url string
}

// NewURLArtifactID creates a new URL artifact ID generator
func NewURLArtifactID(rawURL string) *URLArtifactID {
	// Normalize URL for consistent IDs
	normalized := normalizeURL(rawURL)
	return &URLArtifactID{url: normalized}
}

// Generate creates a deterministic ID for this URL
func (id *URLArtifactID) Generate() string {
	hash := sha256.Sum256([]byte(id.url))
	return fmt.Sprintf("url-artifact-%x", hash[:8])
}

// GetNormalizedURL returns the normalized URL used for ID generation
func (id *URLArtifactID) GetNormalizedURL() string {
	return id.url
}

// normalizeURL standardizes URL format for consistent hashing
func normalizeURL(rawURL string) string {
	// Parse URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// Fallback to raw URL if parsing fails
		return strings.TrimSpace(rawURL)
	}
	
	// Normalize components
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = "" // Remove fragment for consistency
	
	// Sort query parameters for consistent ordering
	if parsed.RawQuery != "" {
		query := parsed.Query()
		parsed.RawQuery = query.Encode() // This sorts parameters
	}
	
	return parsed.String()
}

// IsValidURL performs basic URL validation
func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}
	
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	
	// Must have scheme and host
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	
	// Only support HTTP/HTTPS
	scheme := strings.ToLower(parsed.Scheme)
	return scheme == "http" || scheme == "https"
}