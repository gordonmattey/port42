package validation

import (
	"fmt"
	"strings"
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

	// Check for potentially problematic content
	lowerPrompt := strings.ToLower(prompt)
	if strings.Contains(lowerPrompt, "ignore previous") ||
		strings.Contains(lowerPrompt, "system:") ||
		strings.Contains(lowerPrompt, "disregard") ||
		strings.Contains(lowerPrompt, "override") {
		return ValidationError{
			Field:      "user_prompt",
			Message:    "Prompt contains potentially problematic content",
			Code:       "SUSPICIOUS_PROMPT_CONTENT",
			Suggestion: "Use straightforward instructions for tool generation",
			Example:    "Focus on what the tool should do, not system instructions",
		}
	}

	return ValidationError{} // No error
}