# Step 5: Error Handling and Validation - Implementation Plan

## Overview

Transform the universal prompt/reference system from functional to production-ready through comprehensive error handling and validation. Focus on root cause fixes, not symptom patches, while leveraging existing help system architecture.

## Component Decomposition & Architecture

### 1. Validation Layer (Input Sanitization)
**Location**: `daemon/validation/` (NEW package)

**Components**:
- `reference_validator.go` - Validates reference format and existence
- `prompt_validator.go` - Validates prompt content and length  
- `request_validator.go` - Validates complete request structures
- `error_formatter.go` - Converts validation errors to user-friendly messages
- `types.go` - Common validation types and interfaces

**Responsibility**: Early validation before processing, fail-fast principle

**Architecture**:
```go
// Validator interface for all validation components
type Validator interface {
    Validate(input interface{}) []ValidationError
}

// ValidationError with user-friendly messaging
type ValidationError struct {
    Field      string `json:"field"`
    Message    string `json:"message"`
    Code       string `json:"code"`
    Suggestion string `json:"suggestion,omitempty"`
    Example    string `json:"example,omitempty"`
}
```

### 2. Error Handling Enhancement (Existing Components)
**Location**: Enhance existing error handling in:

**Components**:
- `daemon/server.go` - Request handler error wrapping and translation
- `daemon/resolution/service.go` - Reference resolution error handling with graceful degradation
- `daemon/tool_materializer.go` - AI generation error handling with retry logic
- `cli/src/common/errors.rs` - Enhanced CLI error types using existing help system

**Responsibility**: Graceful degradation, helpful error messages, system stability

### 3. Help System Integration (Existing)
**Location**: `cli/src/help_text.rs` (EXTEND existing)

**Components**:
- Error message constants with examples following Reality Compiler language
- Context-aware help suggestions
- Integration with existing color coding and formatting
- Error-to-help mapping for consistent UX

**Responsibility**: User guidance through existing help patterns, not just error reporting

## Root Cause Analysis & Solutions

### Problem 1: Reference Resolution Failures
**Root Cause**: No validation of reference format or existence before processing
**Current Symptom**: Cryptic errors during resolution or silent failures
**Solution**: Early validation with specific error messages per reference type

```go
// Before: Silent failure or cryptic Go error
// After: "Reference 'file:/nonexistent.txt' not found. Check the file path or use 'ls' to verify location."
```

### Problem 2: Poor User Experience on Errors
**Root Cause**: Technical errors bubble up to user without translation
**Current Symptom**: Users see Go stack traces or raw API errors  
**Solution**: Error translation layer with actionable suggestions using help system patterns

```go
// Before: "json: cannot unmarshal string into Go struct field"
// After: "🚫 Invalid reference format. Use: file:path, p42:/tools/name, url:https://..., or search:\"query\""
```

### Problem 3: No Input Sanitization
**Root Cause**: Direct passing of user input to internal systems
**Current Symptom**: Potential crashes or unexpected behavior with malformed input
**Solution**: Validation layer with sanitization and limits

```go
// Before: Unlimited prompt length potentially causing API failures
// After: Prompt validation with length limits and character sanitization
```

### Problem 4: Inconsistent Error Messaging
**Root Cause**: No centralized error formatting following existing patterns
**Current Symptom**: Mix of technical and user-friendly messages
**Solution**: Integrate with existing help_text.rs patterns for consistency

## Implementation Plan

### Phase 1: Validation Infrastructure (30min)

**File**: `daemon/validation/types.go`
```go
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
```

### Phase 2: Reference Validation (45min)

**File**: `daemon/validation/reference_validator.go`
```go
package validation

import (
    "fmt"
    "net/url"
    "os"
    "path/filepath"
    "strings"
)

type ReferenceValidator struct{}

func NewReferenceValidator() *ReferenceValidator {
    return &ReferenceValidator{}
}

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

    // TODO: Check if path exists in VFS (implement in Phase 4)
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
```

### Phase 3: Prompt Validation (30min)

**File**: `daemon/validation/prompt_validator.go`
```go
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
    if strings.Contains(strings.ToLower(prompt), "ignore previous") ||
       strings.Contains(strings.ToLower(prompt), "system:") {
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
```

### Phase 4: Error Translation Layer (30min)

**File**: `daemon/validation/error_formatter.go`
```go
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
```

### Phase 5: Integration Points (45min)

**File**: Enhanced `daemon/server.go`
```go
// Add to imports
import (
    "your-project/daemon/validation"
)

// Add to Daemon struct
type Daemon struct {
    // ... existing fields ...
    validator *validation.RequestValidator
}

// Update NewDaemon
func NewDaemon(config DaemonConfig) (*Daemon, error) {
    // ... existing code ...
    
    validator := validation.NewRequestValidator()
    
    return &Daemon{
        // ... existing fields ...
        validator: validator,
    }, nil
}

// Enhanced handleDeclareRelation
func (d *Daemon) handleDeclareRelation(req Request) Response {
    resp := NewResponse(req.ID, true)
    
    // PHASE 1: Early validation - fail fast with helpful messages
    if validationResult := d.validator.ValidateRequest(req); validationResult.HasErrors() {
        errorFormatter := validation.NewErrorFormatter(true) // Use color output
        userFriendlyError := errorFormatter.FormatValidationErrors(validationResult.Errors)
        resp.SetError(userFriendlyError)
        return resp
    }
    
    // Continue with existing logic...
    if d.realityCompiler == nil {
        resp.SetError("Reality compiler not initialized")
        return resp
    }
    
    // ... rest of existing function with enhanced error handling ...
}
```

**File**: Enhanced CLI error handling `cli/src/common/errors.rs`
```rust
// Add new error variants using existing help_text patterns
use crate::help_text;

#[derive(Debug)]
pub enum ValidationError {
    InvalidReference { 
        ref_type: String, 
        target: String, 
        suggestion: String 
    },
    InvalidPrompt { 
        reason: String, 
        max_length: usize 
    },
    FileNotFound { 
        path: String 
    },
    // Leverage existing error constants
}

impl ValidationError {
    pub fn user_message(&self) -> String {
        match self {
            ValidationError::InvalidReference { ref_type, target, suggestion } => {
                format!("{} Invalid reference '{}:{}'. {}", 
                    help_text::ERR_INVALID_REFERENCE_PREFIX,
                    ref_type, target, suggestion)
            },
            ValidationError::InvalidPrompt { reason, max_length } => {
                format!("{} Prompt validation failed: {}. Maximum length: {} characters", 
                    help_text::ERR_PROMPT_VALIDATION_PREFIX,
                    reason, max_length)
            },
            ValidationError::FileNotFound { path } => {
                format!("{} File not found: {}. Use 'port42 ls' to explore available files", 
                    help_text::ERR_FILE_NOT_FOUND_PREFIX,
                    path)
            },
        }
    }
}
```

### Phase 6: Comprehensive Testing (60min)

**File**: `tests/integration/prompt-ref-system/05-error-handling.sh`
```bash
#!/bin/bash
# Test 5: Error Handling and Validation
# Tests that all error conditions are handled gracefully with helpful messages

set -e

echo "🚫 Test 5: Error Handling and Validation"

# Test 1: Invalid file reference
echo "🧪 Test 1: Invalid file reference handling"
if port42 declare tool error-test-file --transforms "test" \
  --ref "file:/nonexistent/file.txt" 2>&1 | grep -q "File not found"; then
  echo "✅ PASS: File not found error handled gracefully"
else
  echo "❌ FAIL: File not found error not handled properly"
  exit 1
fi

# Test 2: Malformed reference format
echo "🧪 Test 2: Malformed reference format"
if port42 declare tool error-test-format --transforms "test" \
  --ref "invalid-reference-format" 2>&1 | grep -q "Invalid reference"; then
  echo "✅ PASS: Malformed reference handled gracefully"
else
  echo "❌ FAIL: Malformed reference not handled properly"
  exit 1
fi

# Test 3: Invalid P42 path
echo "🧪 Test 3: Invalid P42 path"
if port42 declare tool error-test-p42 --transforms "test" \
  --ref "p42:/nonexistent/tool" 2>&1 | grep -q "not found\|invalid"; then
  echo "✅ PASS: Invalid P42 path handled gracefully"
else
  echo "❌ FAIL: Invalid P42 path not handled properly"
  exit 1
fi

# Test 4: Invalid URL format
echo "🧪 Test 4: Invalid URL format"
if port42 declare tool error-test-url --transforms "test" \
  --ref "url:not-a-valid-url" 2>&1 | grep -q "Invalid URL\|URL"; then
  echo "✅ PASS: Invalid URL handled gracefully"
else
  echo "❌ FAIL: Invalid URL not handled properly"
  exit 1
fi

# Test 5: Overly long prompt
echo "🧪 Test 5: Overly long prompt"
LONG_PROMPT=$(python3 -c "print('A' * 6000)")  # Exceeds 5KB limit
if port42 declare tool error-test-prompt --transforms "test" \
  --prompt "$LONG_PROMPT" 2>&1 | grep -q "too long\|length"; then
  echo "✅ PASS: Long prompt handled gracefully"
else
  echo "❌ FAIL: Long prompt not handled properly"
  exit 1
fi

# Test 6: Empty references (should be valid)
echo "🧪 Test 6: Empty references handling"
if port42 declare tool empty-ref-test --transforms "test" 2>&1; then
  echo "✅ PASS: Empty references handled correctly"
else
  echo "❌ FAIL: Empty references caused unexpected error"
  exit 1
fi

# Test 7: Multiple invalid references
echo "🧪 Test 7: Multiple invalid references"
if port42 declare tool multi-error-test --transforms "test" \
  --ref "file:/nonexistent1.txt" \
  --ref "invalid-format" \
  --ref "url:not-a-url" 2>&1 | grep -q "Validation failed"; then
  echo "✅ PASS: Multiple errors aggregated properly"
else
  echo "❌ FAIL: Multiple errors not handled properly"
  exit 1
fi

# Test 8: System stability under invalid input
echo "🧪 Test 8: System stability test"
# Send multiple invalid requests rapidly
for i in {1..5}; do
  port42 declare tool "rapid-error-$i" --transforms "test" \
    --ref "file:/invalid$i.txt" >/dev/null 2>&1 || true
done

# Check if daemon is still responsive
if port42 status >/dev/null 2>&1; then
  echo "✅ PASS: System remains stable under error conditions"
else
  echo "❌ FAIL: System became unresponsive after error conditions"
  exit 1
fi

echo ""
echo "✅ All error handling tests passed!"
echo "🛡️ System handles errors gracefully with helpful messages"
echo "🎯 Users receive actionable guidance instead of technical errors"
```

## Error Categories & Handling Strategy

### 1. Validation Errors (Fail Fast)
**Strategy**: Immediate rejection with specific guidance
- Invalid reference format → Show valid formats with examples
- File not found → Suggest path verification commands
- Prompt too long → Show length limit and current length
- Malformed URL → Provide URL format examples

### 2. Runtime Errors (Graceful Degradation)
**Strategy**: Continue operation where possible, warn user clearly
- Reference resolution failure → Continue without context, show warning
- AI API failure → Retry once, then graceful failure with clear next steps
- Network timeout → Clear timeout message with retry suggestion
- Partial reference resolution → Use available context, note failures

### 3. System Errors (Transparent Reporting)
**Strategy**: Log technical details, show user-friendly message with actions
- Internal server errors → "Service temporarily unavailable, try again"
- Resource exhaustion → Clear explanation of limits and resolution steps
- Configuration issues → Guide user to fix configuration

### 4. User Guidance Integration
**Strategy**: Leverage existing help system patterns
- Use Reality Compiler language and metaphors
- Include examples in error messages
- Suggest specific commands to resolve issues
- Maintain color coding and formatting consistency

## Integration with Existing Help System

### Leverage Existing Patterns
- Use `help_text.rs` constants for consistent messaging
- Follow Reality Compiler language ("consciousness", "manifestation", etc.)
- Maintain existing color coding and emoji usage
- Integrate with existing command help structure

### Extend with Error Context
- Map error codes to relevant help sections
- Provide context-aware suggestions
- Include examples specific to current command
- Link errors to broader help topics when relevant

### Consistent User Experience
- Error messages feel like natural extension of help system
- Same tone and language throughout
- Progressive disclosure: brief message + detailed help available
- Actionable next steps rather than just problem description

## Success Criteria

### 1. System Stability
- ✅ No crashes with any combination of malformed input
- ✅ Daemon remains responsive during error conditions
- ✅ Memory usage stable under error scenarios
- ✅ Performance impact <10ms for validation

### 2. User Experience
- ✅ Clear, actionable error messages in Reality Compiler language
- ✅ Examples provided for common error resolution
- ✅ Progressive help: quick fix → detailed guidance
- ✅ Consistent formatting with existing help system

### 3. Developer Experience
- ✅ Comprehensive error coverage in tests
- ✅ Easy to add new validation rules
- ✅ Clear separation between validation and business logic
- ✅ Error messages easily maintainable and translatable

### 4. Integration Quality
- ✅ Seamless integration with existing architecture
- ✅ No breaking changes to existing functionality
- ✅ Maintains performance characteristics
- ✅ Error handling follows established patterns

## Implementation Timeline

**Total Estimated Time: 4 hours**

1. **Phase 1** (30min): Validation infrastructure and types
2. **Phase 2** (45min): Reference validation with all types
3. **Phase 3** (30min): Prompt validation and sanitization
4. **Phase 4** (30min): Error translation and formatting
5. **Phase 5** (45min): Integration with server and CLI
6. **Phase 6** (60min): Comprehensive testing and edge cases

This plan ensures robust error handling while maintaining the existing architecture and user experience patterns established in the Port 42 system.