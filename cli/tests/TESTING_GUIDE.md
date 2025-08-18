# Manual Testing Guide

This guide walks you through comprehensive manual testing of the Universal Prompt & Reference System.

## Quick Start

```bash
# 1. Check environment readiness
cd cli/tests
./pre-test-check.sh

# 2. Run basic functionality tests
./run-manual-tests.sh basic

# 3. Run all tests
./manual-test-suite.sh
```

## Test Sections

### 1. Basic Functionality Tests
Tests core tool declaration without references.

```bash
./run-manual-tests.sh basic
```

**What it tests:**
- Daemon startup and status
- Simple tool declaration
- VFS object store creation
- Basic command generation

**Expected results:**
- Tool appears in `/tools/` and `/commands/`
- Executable code is generated
- VFS shows tool relationships

### 2. Reference System Tests
Tests all reference types: file, URL, p42, and search.

```bash
./run-manual-tests.sh references
```

**What it tests:**
- File references (`file:./config.json`)
- URL references (`url:https://...`)
- P42 VFS references (`p42:/tools/...`)
- Search references (`search:"query"`)
- Multiple references on one tool

**Expected results:**
- Referenced content appears in tool definitions
- Generated tools reflect referenced knowledge
- Debug output shows reference resolution

### 3. Custom Prompt Tests
Tests AI guidance with custom prompts.

```bash
./run-manual-tests.sh prompts
```

**What it tests:**
- Custom prompt processing
- Prompt influence on generated code
- Artifact creation with prompts

**Expected results:**
- Generated tools follow prompt instructions
- Code reflects specific requirements
- Debug shows prompt integration

### 4. Advanced Integration Tests
Tests combined prompts + references.

```bash
./run-manual-tests.sh advanced
```

**What it tests:**
- Custom prompts combined with multiple references
- Complex context-aware tool generation
- Integration of all system components

**Expected results:**
- Tools combine prompt guidance with reference knowledge
- High-quality, contextually appropriate code generation
- Complete integration verification

### 5. Debug and VFS Tests
Tests debugging capabilities and virtual filesystem.

```bash
./run-manual-tests.sh debug
./run-manual-tests.sh vfs
```

**What it tests:**
- Debug output capture and analysis
- VFS navigation and inspection
- Object store verification
- Search functionality

### 6. Error Handling Tests
Tests system robustness with invalid inputs.

```bash
./run-manual-tests.sh errors
```

**What it tests:**
- Invalid file references
- Bad URL references
- Empty prompts
- Graceful error handling

## Manual Verification Steps

### After Each Test Section

1. **Inspect the Object Store:**
   ```bash
   port42 ls /
   port42 ls /tools/
   port42 ls /commands/
   ```

2. **Examine Tool Definitions:**
   ```bash
   # See what references were loaded
   port42 cat /tools/test-tool/definition
   
   # Check generated code
   port42 cat /commands/test-tool
   
   # View metadata
   port42 info /tools/test-tool
   ```

3. **Verify Reference Loading:**
   Look for reference content in tool definitions:
   ```bash
   # Should contain referenced file content
   port42 cat /tools/test-file-ref-tool/definition | grep -i "api_url"
   
   # Should contain URL content
   port42 cat /tools/test-url-ref-tool/definition | grep -i "slideshow"
   ```

4. **Check Debug Output:**
   ```bash
   # Enable debug and run a command
   PORT42_DEBUG=1 port42 declare tool debug-test --ref file:./test.json
   
   # Look for debug lines like:
   # DEBUG: reference_resolver - Processing file reference
   # DEBUG: file_resolver - File content loaded: 123 bytes
   ```

## Debug Mode Deep Dive

### Enable Full Debug Output

```bash
export PORT42_DEBUG=1
export PORT42_VERBOSE=1
export PORT42_REF_DEBUG=1
```

### Key Debug Patterns to Look For

**Reference Resolution:**
```
DEBUG: reference_resolver - Processing file reference: ./config.json
DEBUG: file_resolver - Reading file: ./config.json
DEBUG: file_resolver - File size: 456 bytes
DEBUG: file_resolver - Content preview: {"api_url": "https://..."}
```

**AI Integration:**
```
DEBUG: ai_integration - Prompt length: 234 characters
DEBUG: ai_integration - Reference context: 1234 characters
DEBUG: ai_integration - Total context size: 1468 characters
DEBUG: ai_integration - API request sent to Claude
DEBUG: ai_integration - Response received: 200 OK
```

**Tool Materialization:**
```
DEBUG: tool_materializer - Generating executable for: test-tool
DEBUG: tool_materializer - Template: python_tool.py
DEBUG: tool_materializer - Code generated: 67 lines
DEBUG: vfs - Creating tool paths: /tools/test-tool/
```

### Common Issues and Debug Solutions

**Issue: References not loading**
```bash
# Check file exists and is readable
ls -la ./test-config.json
cat ./test-config.json

# Check debug output for file resolution
PORT42_DEBUG=1 port42 declare tool test --ref file:./test-config.json 2>&1 | grep -i "file_resolver"
```

**Issue: Prompts not applied**
```bash
# Verify prompt is being processed
PORT42_DEBUG=1 port42 declare tool test --prompt "specific instructions" 2>&1 | grep -i "prompt"
```

**Issue: Tools not appearing in VFS**
```bash
# Check daemon logs
port42 daemon logs -n 20

# Verify VFS is working
port42 ls /
port42 status
```

## Expected Test Results

### Successful Test Run Output

```
🧪 Port 42 Universal Prompt & Reference System Test Suite
================================================================

✅ Test 1: Setting up test input files
✅ Test 2: Basic setup and daemon status
✅ Test 3: Simple tool declaration without references
✅ Test 4: Tool declaration with file reference
✅ Test 5: Tool declaration with multiple file references
✅ Test 6: Tool declaration with URL reference
✅ Test 7: Tool declaration with P42 VFS reference
✅ Test 8: Tool declaration with search reference
✅ Test 9: Tool declaration with custom prompt
✅ Test 10: Tool declaration with custom prompt AND multiple references
✅ Test 11: Artifact creation with custom prompt
✅ Test 12: Debug mode functionality verification
✅ Test 13: Virtual filesystem navigation and inspection
✅ Test 14: Error handling and edge cases

================================================================
Test Suite Complete
✅ Passed: 14
❌ Failed: 0
📊 Total: 14

🎉 All tests passed! Universal Prompt & Reference System is working correctly.
```

### What Success Looks Like

1. **All tests pass** with green checkmarks
2. **Object store populated** with tools, commands, and relationships
3. **References properly loaded** into tool definitions
4. **Generated code quality** reflects prompts and references
5. **Debug output available** for troubleshooting
6. **VFS navigation working** for inspection
7. **Error handling graceful** for invalid inputs

## Troubleshooting

### If Tests Fail

1. **Run pre-test check:** `./pre-test-check.sh`
2. **Check daemon status:** `port42 status`
3. **Enable debug mode:** `export PORT42_DEBUG=1`
4. **Run individual sections:** `./run-manual-tests.sh basic`
5. **Inspect logs:** `port42 daemon logs -f`
6. **Check API key:** `echo $ANTHROPIC_API_KEY`

### Common Solutions

- **Build binaries:** `cd ../.. && ./build.sh`
- **Start daemon:** `port42 daemon start`
- **Set API key:** `export ANTHROPIC_API_KEY='your-key'`
- **Check network:** `ping httpbin.org`
- **Clear state:** `rm -rf ~/.port42/` (warning: removes all data)

See `DEBUG_GUIDE.md` for detailed debugging information.