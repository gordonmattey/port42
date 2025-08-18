# Enhanced Error Handling Test Guide

## Tests for improved error classification and messaging

### Test 1: Claude API Overloaded (Simulated)
This would happen when Claude's API returns an overload error. The new system should show:
- `🤖 Claude API is currently experiencing issues. Please try again in a moment.`
- Error type: `ClaudeApi` instead of generic `Daemon`

### Test 2: API Key Missing  
Remove or invalid API key should show:
- `🔑 API key issue. Please check your ANTHROPIC_API_KEY configuration.`
- Error type: `ApiKey` instead of generic `Daemon`

### Test 3: Network Issues
Network connectivity problems should show:
- `🌐 Network connection issue. Please check your internet connection.`
- Error type: `Network` instead of generic `Daemon`

### Test 4: Actual Daemon Issues
Real daemon problems should still show:
- `❌ AI connection failed: [error details]`
- Error type: `Daemon` (unchanged for real daemon issues)

## Changes Made

**Daemon Side (`daemon/possession.go`):**
- Added error classification logic that prefixes errors with source type
- `CLAUDE_API_ERROR:`, `API_KEY_ERROR:`, `NETWORK_ERROR:`, `AI_CONNECTION_ERROR:`

**CLI Side (`cli/src/possess/session.rs`):**
- Added `classify_error()` function that parses prefixed errors
- Enhanced error display with appropriate icons and user-friendly messages
- New error types: `ClaudeApi`, `ApiKey`, `Network`, `ExternalService`

## Before vs After

**Before:**
```
❌ AI connection failed: API error: api_error - Overloaded
Error: Daemon error: API error: api_error - Overloaded
```

**After:**
```
🤖 Claude API is currently experiencing issues. Please try again in a moment.
Error: Claude API error: API error: api_error - Overloaded
```

The user now understands it's a Claude issue, not a daemon problem, and gets actionable guidance.