package validation

import (
	"fmt"
	"unicode/utf8"
)

type PromptValidator struct {
	MaxLength int
}

func NewPromptValidator() *PromptValidator {
	return &PromptValidator{
		MaxLength: 5000, // 5KB reasonable limit for prompts
	}
}

func (pv *PromptValidator) ValidatePrompt(prompt string) ValidationError {
	if prompt == "" {
		return ValidationError{} // Empty prompts are valid (optional)
	}

	// Check length
	if len(prompt) > pv.MaxLength {
		return ValidationError{
			Field:      "user_prompt",
			Message:    fmt.Sprintf("Prompt too long (%d characters, maximum %d)", len(prompt), pv.MaxLength),
			Code:       "PROMPT_TOO_LONG",
			Suggestion: "Shorten your prompt or split into multiple commands",
			Example:    "Use concise, specific instructions for best results",
		}
	}

	// Check UTF-8 validity
	if !utf8.ValidString(prompt) {
		return ValidationError{
			Field:      "user_prompt",
			Message:    "Prompt contains invalid UTF-8 characters",
			Code:       "INVALID_PROMPT_ENCODING",
			Suggestion: "Ensure prompt uses valid text characters",
			Example:    "Use standard text without special binary characters",
		}
	}

	// Note: Removed suspicious content detection to avoid false positives
	// Users should be able to create tools with any legitimate requirements

	return ValidationError{} // No error
}