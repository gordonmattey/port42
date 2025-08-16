package validation

import "fmt"

// ValidationError represents a user-friendly validation error
type ValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Code       string `json:"code"`
	Suggestion string `json:"suggestion,omitempty"`
	Example    string `json:"example,omitempty"`
}

// Validator interface for all validation components
type Validator interface {
	Validate(input interface{}) []ValidationError
}

// ValidationResult aggregates multiple validation errors
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

func (v ValidationResult) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v ValidationResult) FirstError() string {
	if len(v.Errors) > 0 {
		return v.Errors[0].Message
	}
	return ""
}

// IsEmpty returns true if the ValidationError is empty (no error)
func (e ValidationError) IsEmpty() bool {
	return e.Message == "" && e.Code == ""
}

// String returns a formatted error message
func (e ValidationError) String() string {
	if e.IsEmpty() {
		return ""
	}
	result := e.Message
	if e.Suggestion != "" {
		result += fmt.Sprintf("\n    💡 %s", e.Suggestion)
	}
	if e.Example != "" {
		result += fmt.Sprintf("\n    📝 Example: %s", e.Example)
	}
	return result
}