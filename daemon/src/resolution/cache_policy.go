package resolution

import (
	"strings"
	"time"
)

// CachePolicy defines cache behavior for URL artifacts
type CachePolicy struct {
	DefaultTTL     time.Duration
	MaxContentSize int64
}

// DefaultCachePolicy returns sensible default cache settings
func DefaultCachePolicy() CachePolicy {
	return CachePolicy{
		DefaultTTL:     24 * time.Hour, // 24 hour cache by default
		MaxContentSize: 50 * 1024,      // 50KB max content size
	}
}

// IsExpired checks if content fetched at given time is expired
func (cp CachePolicy) IsExpired(fetchedAt time.Time) bool {
	return time.Since(fetchedAt) > cp.DefaultTTL
}

// ShouldCache determines if content should be cached based on response
func (cp CachePolicy) ShouldCache(url string, statusCode int, contentLength int64, contentType string) bool {
	// Don't cache HTTP errors
	if statusCode >= 400 {
		return false
	}
	
	// Don't cache oversized content
	if contentLength > cp.MaxContentSize {
		return false
	}
	
	// Don't cache certain content types that change frequently
	if strings.Contains(strings.ToLower(contentType), "text/event-stream") {
		return false
	}
	
	return true
}

// GetTTLForContentType returns TTL based on content type (future enhancement)
func (cp CachePolicy) GetTTLForContentType(contentType string) time.Duration {
	contentType = strings.ToLower(contentType)
	
	switch {
	case strings.Contains(contentType, "application/json"):
		return 4 * time.Hour // API responses expire faster
	case strings.Contains(contentType, "text/html"):
		return 12 * time.Hour // Web pages expire faster
	default:
		return cp.DefaultTTL
	}
}