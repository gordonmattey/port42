# Port 42 Debug Guide for Manual Testing

This guide explains how to enable comprehensive debugging for the Universal Prompt & Reference System testing.

## Quick Start

```bash
# Run the complete test suite with debugging
cd cli/tests
export PORT42_DEBUG=1
./manual-test-suite.sh

# Run specific test sections
./run-manual-tests.sh basic      # Basic functionality
./run-manual-tests.sh references # All reference types
./run-manual-tests.sh advanced   # Combined prompts + references
```

## Debug Environment Setup

### 1. Enable Debug Mode

```bash
# Essential debug flag
export PORT42_DEBUG=1

# Optional: More verbose daemon logging
export PORT42_VERBOSE=1

# Optional: Reference resolver debugging
export PORT42_REF_DEBUG=1
```

### 2. Daemon Debug Mode

```bash
# Start daemon in foreground with debug output
port42 daemon stop
PORT42_DEBUG=1 ./bin/port42d

# Or start in background and monitor logs
port42 daemon start -b
port42 daemon logs -f
```

### 3. CLI Debug Output

```bash
# Individual commands with debug
PORT42_DEBUG=1 port42 declare tool test-tool --transforms test --ref file:./config.json

# All commands with verbose output
PORT42_DEBUG=1 port42 --verbose ls /tools/
```

## Debug Output Interpretation

### Tool Declaration Debug Flow

When declaring a tool, look for these debug markers:

```
DEBUG: declare.rs - Starting tool declaration: test-tool
DEBUG: client.rs - Sending declare_tool request
DEBUG: reference_resolver - Processing file reference: ./config.json
DEBUG: reference_resolver - File content loaded: 1234 bytes
DEBUG: daemon - UserPrompt: "Create a tool that..."
DEBUG: daemon - References resolved: 1 items
DEBUG: tool_materializer - Generating executable for: test-tool
DEBUG: vfs - Creating tool paths: /tools/test-tool/
```

### Reference Resolution Debug

Each reference type shows specific debug info:

**File References:**
```
DEBUG: file_resolver - Reading file: ./config.json
DEBUG: file_resolver - File size: 1234 bytes
DEBUG: file_resolver - Content preview: {"api_url": "https://..."}
```

**URL References:**
```
DEBUG: url_resolver - Fetching: https://api.example.com/docs
DEBUG: url_resolver - Response status: 200
DEBUG: url_resolver - Content-Type: application/json
DEBUG: url_resolver - Content size: 5678 bytes
```

**P42 VFS References:**
```
DEBUG: p42_resolver - Resolving: p42:/tools/existing-tool
DEBUG: p42_resolver - Tool definition found: existing-tool
DEBUG: p42_resolver - Including executable and metadata
```

**Search References:**
```
DEBUG: search_resolver - Query: "config validation"
DEBUG: search_resolver - Found 3 matching items
DEBUG: search_resolver - Including: cli-1234, tool-config-validator, artifact-docs
```

### AI Integration Debug

```
DEBUG: ai_integration - Prompt length: 1234 characters
DEBUG: ai_integration - Reference context: 5678 characters
DEBUG: ai_integration - Total context size: 6912 characters
DEBUG: ai_integration - API request sent to Claude
DEBUG: ai_integration - Response received: 200 OK
DEBUG: ai_integration - Generated tool code: 89 lines
```

## Common Debug Scenarios

### 1. Reference Not Loading

**Symptoms:** Tool doesn't reflect referenced content
**Debug:** Look for these patterns:

```bash
PORT42_DEBUG=1 port42 declare tool test --ref file:./missing.json
# Look for:
# ERROR: file_resolver - File not found: ./missing.json
# or
# DEBUG: file_resolver - File empty or unreadable
```

**Solution:** Check file paths, permissions, and content

### 2. Prompt Not Applied

**Symptoms:** Generated tool ignores custom prompt
**Debug:** Check prompt processing:

```bash
PORT42_DEBUG=1 port42 declare tool test --prompt "specific instructions"
# Look for:
# DEBUG: daemon - UserPrompt: "specific instructions"
# DEBUG: ai_integration - Prompt incorporated into context
```

**Solution:** Verify prompt is being sent to AI service

### 3. VFS Reference Issues

**Symptoms:** P42 references fail to resolve
**Debug:** Check VFS state:

```bash
PORT42_DEBUG=1 port42 ls /tools/
PORT42_DEBUG=1 port42 declare tool test --ref p42:/tools/nonexistent
# Look for:
# ERROR: p42_resolver - Path not found: /tools/nonexistent
```

**Solution:** Verify referenced tools exist in VFS

### 4. Network/URL Issues

**Symptoms:** URL references timeout or fail
**Debug:** Check network resolution:

```bash
PORT42_DEBUG=1 port42 declare tool test --ref url:https://slow-api.com
# Look for:
# DEBUG: url_resolver - Request timeout after 30s
# ERROR: url_resolver - Failed to fetch: connection timeout
```

**Solution:** Check network connectivity, URL validity

## Test Data Inspection

The test suite creates these files for debugging:

```
cli/tests/test-data/
├── test-config.json      # Simple configuration
├── sample-data.csv       # CSV data for processing
├── project-readme.md     # Documentation reference
├── data-schema.json      # JSON Schema
└── api-spec.yaml         # OpenAPI specification
```

You can inspect these files and modify them to test different scenarios.

## VFS Object Store Inspection

### View Tool Definitions

```bash
# See raw tool definition (includes all references)
port42 cat /tools/test-tool/definition

# See generated executable code
port42 cat /commands/test-tool

# See tool metadata
port42 info /tools/test-tool
```

### Trace Tool Creation

```bash
# List all tools to see what was created
port42 ls /tools/

# Check tool relationships
port42 ls /tools/test-tool/spawned/    # Auto-spawned tools
port42 ls /tools/test-tool/parents/    # Parent relationships

# Search for related tools
port42 search "test" --type tool
```

### Memory and Session Inspection

```bash
# View all sessions
port42 memory

# Inspect specific session
port42 memory cli-1234567890

# Search conversation history
port42 memory search "tool creation"
```

## Advanced Debugging Techniques

### 1. Daemon Process Monitoring

```bash
# Monitor daemon in real-time
watch -n 1 'ps aux | grep port42d'

# Check daemon port binding
netstat -an | grep :42
lsof -i :42

# Monitor file system activity
sudo fs_usage -w -f pathname | grep port42
```

### 2. Network Request Tracing

```bash
# Monitor HTTP requests for URL references
sudo tcpdump -i any port 80 or port 443

# Or use a proxy like mitmproxy
mitmproxy -s debug_proxy.py
```

### 3. File System Monitoring

```bash
# Watch for file access during reference resolution
sudo fs_usage -w -f pathname | grep -E "(test-data|\.port42)"
```

## Error Pattern Recognition

### Common Error Patterns and Solutions

| Error Pattern | Likely Cause | Solution |
|---------------|--------------|----------|
| `reference_resolver - Failed to resolve` | Bad reference format | Check reference syntax |
| `daemon - Connection refused` | Daemon not running | `port42 daemon start` |
| `ai_integration - API key missing` | No ANTHROPIC_API_KEY | Set API key |
| `tool_materializer - Generation failed` | AI service error | Check API limits, network |
| `vfs - Path not found` | Invalid VFS path | Verify path exists |
| `file_resolver - Permission denied` | File access issue | Check file permissions |

## Test Suite Debug Features

The manual test suite includes these debug features:

1. **Automatic Debug Enable:** Sets `PORT42_DEBUG=1`
2. **Object Store Inspection:** Shows VFS state after each test
3. **Debug Output Capture:** Captures and displays debug logs
4. **Verification Steps:** Checks that references were properly loaded
5. **Error Testing:** Validates error handling with invalid inputs

## Getting Help

If debugging reveals issues:

1. **Check the debug output patterns above**
2. **Run individual test sections:** `./run-manual-tests.sh basic`
3. **Inspect the object store:** Use `port42 ls`, `port42 cat`, `port42 info`
4. **Check daemon logs:** `port42 daemon logs -f`
5. **Test with minimal examples first**

The test suite is designed to progressively test functionality, so start with basic tests and work up to advanced scenarios.