package validation

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Reference represents a parsed reference
type Reference struct {
	Type   string
	Target string
}

type ReferenceValidator struct{}

func NewReferenceValidator() *ReferenceValidator {
	return &ReferenceValidator{}
}

// ParseReference parses a reference string into type and target
func (rv *ReferenceValidator) ParseReference(refStr string) (Reference, ValidationError) {
	if refStr == "" {
		return Reference{}, ValidationError{} // Empty references are valid
	}

	// Parse reference format: type:target
	parts := strings.SplitN(refStr, ":", 2)
	if len(parts) != 2 {
		return Reference{}, ValidationError{
			Field:      "reference",
			Message:    fmt.Sprintf("Invalid reference format: %s", refStr),
			Code:       "INVALID_REFERENCE_FORMAT",
			Suggestion: "Use format: type:target (file:, p42:, url:, search:)",
			Example:    "file:./config.json, p42:/tools/analyzer, url:https://api.docs, search:\"patterns\"",
		}
	}

	return Reference{
		Type:   parts[0],
		Target: parts[1],
	}, ValidationError{}
}

// ValidateReference validates a single reference
func (rv *ReferenceValidator) ValidateReference(ref Reference) ValidationError {
	// Validate reference type
	if ref.Type == "" {
		return ValidationError{
			Field:      "reference.type",
			Message:    "Reference type is required",
			Code:       "MISSING_REFERENCE_TYPE",
			Suggestion: "Specify reference type: file:, p42:, url:, or search:",
			Example:    "file:./config.json",
		}
	}

	switch ref.Type {
	case "file":
		return rv.validateFileReference(ref.Target)
	case "p42":
		return rv.validateP42Reference(ref.Target)
	case "url":
		return rv.validateURLReference(ref.Target)
	case "search":
		return rv.validateSearchReference(ref.Target)
	default:
		return ValidationError{
			Field:      "reference.type",
			Message:    fmt.Sprintf("Unknown reference type: %s", ref.Type),
			Code:       "INVALID_REFERENCE_TYPE",
			Suggestion: "Valid types: file, p42, url, search",
			Example:    "file:./data.json, p42:/tools/analyzer, url:https://api.docs, search:\"patterns\"",
		}
	}
}

func (rv *ReferenceValidator) validateFileReference(target string) ValidationError {
	if target == "" {
		return ValidationError{
			Field:      "reference.target",
			Message:    "File path is required",
			Code:       "MISSING_FILE_PATH",
			Suggestion: "Provide a valid file path",
			Example:    "file:./config.json or file:/absolute/path/file.txt",
		}
	}

	// Check if file exists
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return ValidationError{
			Field:      "reference.target",
			Message:    fmt.Sprintf("File not found: %s", target),
			Code:       "FILE_NOT_FOUND",
			Suggestion: "Check the file path or use 'ls' to verify location",
			Example:    "Use absolute path or verify file exists: ls " + filepath.Dir(target),
		}
	}

	return ValidationError{} // No error
}

func (rv *ReferenceValidator) validateP42Reference(target string) ValidationError {
	if target == "" {
		return ValidationError{
			Field:      "reference.target",
			Message:    "Port 42 path is required",
			Code:       "MISSING_P42_PATH",
			Suggestion: "Provide a valid Port 42 VFS path",
			Example:    "p42:/tools/analyzer or p42:/commands/helper",
		}
	}

	// Validate path format
	if !strings.HasPrefix(target, "/") {
		return ValidationError{
			Field:      "reference.target",
			Message:    fmt.Sprintf("Port 42 path must start with '/': %s", target),
			Code:       "INVALID_P42_PATH_FORMAT",
			Suggestion: "Use absolute path starting with '/'",
			Example:    "p42:/tools/name or p42:/commands/tool-name",
		}
	}

	// TODO: Check if path exists in VFS (implement in future phase)
	return ValidationError{} // No error for now
}

func (rv *ReferenceValidator) validateURLReference(target string) ValidationError {
	if target == "" {
		return ValidationError{
			Field:      "reference.target",
			Message:    "URL is required",
			Code:       "MISSING_URL",
			Suggestion: "Provide a valid HTTP/HTTPS URL",
			Example:    "url:https://api.example.com/docs",
		}
	}

	// Parse URL
	parsedURL, err := url.Parse(target)
	if err != nil {
		return ValidationError{
			Field:      "reference.target",
			Message:    fmt.Sprintf("Invalid URL format: %s", target),
			Code:       "INVALID_URL_FORMAT",
			Suggestion: "Use valid HTTP/HTTPS URL format",
			Example:    "url:https://api.example.com/docs",
		}
	}

	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ValidationError{
			Field:      "reference.target",
			Message:    fmt.Sprintf("URL must use HTTP or HTTPS: %s", target),
			Code:       "INVALID_URL_SCHEME",
			Suggestion: "Use http:// or https:// URLs only",
			Example:    "url:https://api.example.com/docs",
		}
	}

	return ValidationError{} // No error
}

func (rv *ReferenceValidator) validateSearchReference(target string) ValidationError {
	if target == "" {
		return ValidationError{
			Field:      "reference.target",
			Message:    "Search query is required",
			Code:       "MISSING_SEARCH_QUERY",
			Suggestion: "Provide a search query string",
			Example:    "search:\"error patterns\" or search:\"api documentation\"",
		}
	}

	// Basic length validation
	if len(target) < 2 {
		return ValidationError{
			Field:      "reference.target",
			Message:    "Search query too short (minimum 2 characters)",
			Code:       "SEARCH_QUERY_TOO_SHORT",
			Suggestion: "Use more descriptive search terms",
			Example:    "search:\"error handling\" or search:\"configuration\"",
		}
	}

	if len(target) > 200 {
		return ValidationError{
			Field:      "reference.target",
			Message:    "Search query too long (maximum 200 characters)",
			Code:       "SEARCH_QUERY_TOO_LONG",
			Suggestion: "Use more concise search terms",
			Example:    "search:\"key concepts\" instead of very long queries",
		}
	}

	return ValidationError{} // No error
}