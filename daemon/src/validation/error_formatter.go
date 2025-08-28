package validation

import (
	"fmt"
	"strings"
)

type ErrorFormatter struct {
	useColorOutput bool
}

func NewErrorFormatter(useColor bool) *ErrorFormatter {
	return &ErrorFormatter{useColorOutput: useColor}
}

func (ef *ErrorFormatter) FormatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var result strings.Builder
	
	if ef.useColorOutput {
		result.WriteString("🚫 Validation failed:\n")
	} else {
		result.WriteString("Validation failed:\n")
	}

	for i, err := range errors {
		if i > 0 {
			result.WriteString("\n")
		}
		
		result.WriteString(fmt.Sprintf("  • %s", err.Message))
		
		if err.Suggestion != "" {
			result.WriteString(fmt.Sprintf("\n    💡 %s", err.Suggestion))
		}
		
		if err.Example != "" {
			result.WriteString(fmt.Sprintf("\n    📝 Example: %s", err.Example))
		}
	}

	return result.String()
}

func (ef *ErrorFormatter) FormatUserError(err error, context string) string {
	// Convert technical errors to user-friendly messages
	errMsg := err.Error()
	
	// Common error patterns and their user-friendly translations
	errorTranslations := map[string]string{
		"connection refused":              "Cannot connect to Port 42 daemon. Run 'port42 daemon start'",
		"no such file or directory":       "File not found. Check the path and try again",
		"permission denied":               "Permission denied. You may need elevated privileges",
		"invalid character":               "Invalid JSON format in request",
		"timeout":                        "Operation timed out. The AI service may be busy, try again",
		"invalid_request_error":          "Invalid request to AI service. Check your API key and parameters",
		"authentication_error":           "AI service authentication failed. Check your API key",
		"rate_limit_exceeded":            "AI service rate limit exceeded. Wait a moment and try again",
		"insufficient_quota":             "AI service quota exceeded. Check your account billing",
		"context deadline exceeded":      "Request timed out. Try again in a moment",
		"bad gateway":                    "AI service temporarily unavailable. Try again shortly",
		"service unavailable":            "AI service temporarily unavailable. Try again shortly",
	}

	// Find matching error pattern
	for pattern, translation := range errorTranslations {
		if strings.Contains(strings.ToLower(errMsg), pattern) {
			if ef.useColorOutput {
				return fmt.Sprintf("🚫 %s", translation)
			}
			return translation
		}
	}

	// Fallback to generic error with context
	if context != "" {
		return fmt.Sprintf("Error in %s: %s", context, errMsg)
	}
	
	return errMsg
}