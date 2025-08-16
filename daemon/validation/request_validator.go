package validation

import (
	"fmt"
	"log"
)

// RequestValidator coordinates all validation components
type RequestValidator struct {
	referenceValidator *ReferenceValidator
	promptValidator    *PromptValidator
	errorFormatter     *ErrorFormatter
}

func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		referenceValidator: NewReferenceValidator(),
		promptValidator:    NewPromptValidator(),
		errorFormatter:     NewErrorFormatter(true), // Use color output
	}
}

// ValidateRequest validates a complete request with references and prompt
func (rv *RequestValidator) ValidateRequest(req interface{}) ValidationResult {
	var errors []ValidationError

	// Type assertion to get request fields - handle different request types
	switch r := req.(type) {
	case map[string]interface{}:
		// Handle generic map request (from JSON)
		errors = append(errors, rv.validateMapRequest(r)...)
	default:
		log.Printf("🔍 RequestValidator: Unknown request type, skipping validation")
		return ValidationResult{Valid: true, Errors: []ValidationError{}}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func (rv *RequestValidator) validateMapRequest(req map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Validate user prompt if present
	if userPrompt, exists := req["user_prompt"]; exists {
		if promptStr, ok := userPrompt.(string); ok {
			if err := rv.promptValidator.ValidatePrompt(promptStr); !err.IsEmpty() {
				errors = append(errors, err)
			}
		}
	}

	// Validate references if present
	if references, exists := req["references"]; exists {
		if refSlice, ok := references.([]interface{}); ok {
			for i, ref := range refSlice {
				if refStr, ok := ref.(string); ok {
					// Parse reference
					parsedRef, parseErr := rv.referenceValidator.ParseReference(refStr)
					if !parseErr.IsEmpty() {
						parseErr.Field = fmt.Sprintf("references[%d]", i)
						errors = append(errors, parseErr)
						continue
					}

					// Validate reference
					if validationErr := rv.referenceValidator.ValidateReference(parsedRef); !validationErr.IsEmpty() {
						validationErr.Field = fmt.Sprintf("references[%d]", i)
						errors = append(errors, validationErr)
					}
				}
			}
		}
	}

	return errors
}

// ValidateReferences validates a slice of reference strings
func (rv *RequestValidator) ValidateReferences(references []string) []ValidationError {
	var errors []ValidationError

	for i, refStr := range references {
		// Parse reference
		parsedRef, parseErr := rv.referenceValidator.ParseReference(refStr)
		if !parseErr.IsEmpty() {
			parseErr.Field = fmt.Sprintf("references[%d]", i)
			errors = append(errors, parseErr)
			continue
		}

		// Validate reference
		if validationErr := rv.referenceValidator.ValidateReference(parsedRef); !validationErr.IsEmpty() {
			validationErr.Field = fmt.Sprintf("references[%d]", i)
			errors = append(errors, validationErr)
		}
	}

	return errors
}

// ValidatePromptAndReferences validates prompt and references separately
func (rv *RequestValidator) ValidatePromptAndReferences(prompt string, references []string) ValidationResult {
	var errors []ValidationError

	// Validate prompt
	if promptErr := rv.promptValidator.ValidatePrompt(prompt); !promptErr.IsEmpty() {
		errors = append(errors, promptErr)
	}

	// Validate references
	refErrors := rv.ValidateReferences(references)
	errors = append(errors, refErrors...)

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// FormatErrors formats validation errors for user display
func (rv *RequestValidator) FormatErrors(errors []ValidationError) string {
	return rv.errorFormatter.FormatValidationErrors(errors)
}