#!/bin/bash

# Port 42 Universal Prompt & Reference System - Comprehensive Manual Test Suite
# This script tests every piece of functionality from simple to complex
# with clear debugging and verification steps

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_DIR="$(dirname "$TEST_DIR")"
ROOT_DIR="$(dirname "$CLI_DIR")"
PORT42_BIN="$ROOT_DIR/bin/port42"
DAEMON_BIN="$ROOT_DIR/bin/port42d"

# Test data directory
TEST_DATA_DIR="$TEST_DIR/test-data"
mkdir -p "$TEST_DATA_DIR"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Enable debug mode
export PORT42_DEBUG=1

echo -e "${CYAN}🧪 Port 42 Universal Prompt & Reference System Test Suite${NC}"
echo -e "${CYAN}================================================================${NC}"
echo -e "${BLUE}Test Data Directory: $TEST_DATA_DIR${NC}"
echo -e "${BLUE}Port 42 Binary: $PORT42_BIN${NC}"
echo -e "${BLUE}Debug Mode: ENABLED${NC}"
echo ""

# Utility functions
log_test() {
    echo -e "${YELLOW}🔍 Test $((++TOTAL_TESTS)): $1${NC}"
}

log_step() {
    echo -e "   ${BLUE}→ $1${NC}"
}

log_success() {
    echo -e "   ${GREEN}✅ $1${NC}"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "   ${RED}❌ $1${NC}"
    ((FAILED_TESTS++))
}

log_debug() {
    echo -e "   ${CYAN}🐛 DEBUG: $1${NC}"
}

verify_file_exists() {
    if [[ -f "$1" ]]; then
        log_success "File exists: $1"
        return 0
    else
        log_error "File not found: $1"
        return 1
    fi
}

verify_command_exists() {
    if command -v "$1" >/dev/null 2>&1; then
        log_success "Command available: $1"
        return 0
    else
        log_error "Command not found: $1"
        return 1
    fi
}

show_object_store() {
    local description="$1"
    log_step "Object Store Inspection: $description"
    
    # Show VFS structure
    echo -e "   ${CYAN}VFS Structure:${NC}"
    $PORT42_BIN ls / 2>/dev/null || echo "   (VFS not available)"
    
    echo -e "   ${CYAN}Tools:${NC}"
    $PORT42_BIN ls /tools/ 2>/dev/null || echo "   (No tools yet)"
    
    echo -e "   ${CYAN}Commands:${NC}"
    $PORT42_BIN ls /commands/ 2>/dev/null || echo "   (No commands yet)"
    
    echo -e "   ${CYAN}Memory:${NC}"
    $PORT42_BIN ls /memory/ 2>/dev/null || echo "   (No memory yet)"
    
    echo ""
}

check_debug_output() {
    local test_name="$1"
    local log_file="$2"
    
    if [[ -f "$log_file" ]]; then
        log_step "Debug output for $test_name:"
        echo -e "   ${CYAN}Last 10 lines:${NC}"
        tail -10 "$log_file" | sed 's/^/   /'
        echo ""
    fi
}

# Create test input files
create_test_files() {
    log_test "Setting up test input files"
    
    # Simple config file
    if [[ ! -f "$TEST_DATA_DIR/test-config.json" ]]; then
        cat > "$TEST_DATA_DIR/test-config.json" << 'EOF'
{
  "api_url": "https://api.example.com",
  "timeout": 30,
  "retry_count": 3,
  "security": {
    "require_auth": true,
    "rate_limit": 100
  }
}
EOF
        log_success "Created test-config.json"
    else
        log_success "Test-config.json already exists"
    fi
    
    # Sample data file
    if [[ ! -f "$TEST_DATA_DIR/sample-data.csv" ]]; then
        cat > "$TEST_DATA_DIR/sample-data.csv" << 'EOF'
name,age,city,email
John Doe,30,New York,john@example.com
Jane Smith,25,San Francisco,jane@example.com
Bob Johnson,35,Chicago,bob@example.com
Alice Brown,28,Seattle,alice@example.com
EOF
        log_success "Created sample-data.csv"
    else
        log_success "Sample-data.csv already exists"
    fi
    
    # Documentation file
    if [[ ! -f "$TEST_DATA_DIR/project-readme.md" ]]; then
        cat > "$TEST_DATA_DIR/project-readme.md" << 'EOF'
# Test Project

This is a test project for demonstrating Port 42's universal prompt and reference system.

## Features
- Configuration management
- Data processing
- API integration
- Security validation

## Usage
```bash
./tool --config config.json --input data.csv
```

## Security
- Always validate input data
- Use secure connections
- Implement rate limiting
EOF
        log_success "Created project-readme.md"
    else
        log_success "Project-readme.md already exists"
    fi
    
    # Schema file
    if [[ ! -f "$TEST_DATA_DIR/data-schema.json" ]]; then
        cat > "$TEST_DATA_DIR/data-schema.json" << 'EOF'
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {"type": "string", "minLength": 1},
    "age": {"type": "integer", "minimum": 0, "maximum": 150},
    "city": {"type": "string"},
    "email": {"type": "string", "format": "email"}
  },
  "required": ["name", "age", "email"]
}
EOF
        log_success "Created data-schema.json"
    else
        log_success "Data-schema.json already exists"
    fi
    
    # API spec file
    if [[ ! -f "$TEST_DATA_DIR/api-spec.yaml" ]]; then
        cat > "$TEST_DATA_DIR/api-spec.yaml" << 'EOF'
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id: {type: integer}
        name: {type: string}
        email: {type: string, format: email}
EOF
        log_success "Created api-spec.yaml"
    else
        log_success "Api-spec.yaml already exists"
    fi
    
    echo ""
}

# Test 1: Basic Setup and Status
test_basic_setup() {
    log_test "Basic setup and daemon status"
    
    log_step "Checking binaries exist"
    verify_file_exists "$PORT42_BIN" || return 1
    verify_file_exists "$DAEMON_BIN" || return 1
    
    log_step "Starting daemon"
    $PORT42_BIN daemon start -b >/dev/null 2>&1 || true
    sleep 2
    
    log_step "Checking daemon status"
    if $PORT42_BIN status >/dev/null 2>&1; then
        log_success "Daemon is running"
    else
        log_error "Daemon not responding"
        return 1
    fi
    
    log_step "Checking VFS root"
    show_object_store "Initial state"
    
    echo ""
}

# Test 2: Simple tool declaration (no references)
test_simple_tool_declaration() {
    log_test "Simple tool declaration without references"
    
    local tool_name="test-basic-tool"
    
    log_step "Declaring basic tool"
    if $PORT42_BIN declare tool "$tool_name" --transforms "test,basic,demo" 2>&1; then
        log_success "Tool declaration completed"
    else
        log_error "Tool declaration failed"
        return 1
    fi
    
    log_step "Verifying tool exists in VFS"
    if $PORT42_BIN ls /tools/ | grep -q "$tool_name"; then
        log_success "Tool appears in /tools/"
    else
        log_error "Tool not found in /tools/"
    fi
    
    if $PORT42_BIN ls /commands/ | grep -q "$tool_name"; then
        log_success "Tool appears in /commands/"
    else
        log_error "Tool not found in /commands/"
    fi
    
    log_step "Inspecting tool definition"
    $PORT42_BIN cat "/tools/$tool_name/definition" 2>/dev/null || log_error "Cannot read tool definition"
    
    log_step "Inspecting generated executable"
    $PORT42_BIN cat "/commands/$tool_name" 2>/dev/null || log_error "Cannot read tool executable"
    
    show_object_store "After basic tool creation"
    echo ""
}

# Test 3: File reference
test_file_reference() {
    log_test "Tool declaration with file reference"
    
    local tool_name="test-file-ref-tool"
    local config_file="$TEST_DATA_DIR/test-config.json"
    
    log_step "Declaring tool with file reference"
    log_debug "Command: $PORT42_BIN declare tool '$tool_name' --transforms 'config,validate,security' --ref 'file:$config_file'"
    
    if $PORT42_BIN declare tool "$tool_name" --transforms "config,validate,security" --ref "file:$config_file" 2>&1; then
        log_success "Tool with file reference created"
    else
        log_error "Tool with file reference failed"
        return 1
    fi
    
    log_step "Verifying tool includes file content knowledge"
    local tool_def=$($PORT42_BIN cat "/tools/$tool_name/definition" 2>/dev/null)
    if echo "$tool_def" | grep -q "api_url"; then
        log_success "Tool definition contains file content reference"
    else
        log_error "Tool definition missing file content"
    fi
    
    log_step "Checking generated tool code"
    local tool_code=$($PORT42_BIN cat "/commands/$tool_name" 2>/dev/null)
    if echo "$tool_code" | grep -q -i "config\|validate\|security"; then
        log_success "Generated tool reflects config/security focus"
    else
        log_error "Generated tool doesn't reflect expected functionality"
    fi
    
    show_object_store "After file reference tool"
    echo ""
}

# Test 4: Multiple file references
test_multiple_file_references() {
    log_test "Tool declaration with multiple file references"
    
    local tool_name="test-multi-ref-tool"
    
    log_step "Declaring tool with multiple file references"
    log_debug "References: config, schema, and readme files"
    
    if $PORT42_BIN declare tool "$tool_name" \
        --transforms "data,process,validate" \
        --ref "file:$TEST_DATA_DIR/test-config.json" \
        --ref "file:$TEST_DATA_DIR/data-schema.json" \
        --ref "file:$TEST_DATA_DIR/project-readme.md" 2>&1; then
        log_success "Tool with multiple file references created"
    else
        log_error "Tool with multiple file references failed"
        return 1
    fi
    
    log_step "Verifying tool incorporates all references"
    local tool_def=$($PORT42_BIN cat "/tools/$tool_name/definition" 2>/dev/null)
    
    local ref_count=0
    echo "$tool_def" | grep -q "api_url" && ((ref_count++)) && log_success "Config reference found"
    echo "$tool_def" | grep -q "json-schema" && ((ref_count++)) && log_success "Schema reference found"  
    echo "$tool_def" | grep -q "Test Project" && ((ref_count++)) && log_success "README reference found"
    
    if [[ $ref_count -eq 3 ]]; then
        log_success "All file references properly incorporated"
    else
        log_error "Only $ref_count/3 references found in tool definition"
    fi
    
    show_object_store "After multiple file references"
    echo ""
}

# Test 5: URL reference (using a reliable public API)
test_url_reference() {
    log_test "Tool declaration with URL reference"
    
    local tool_name="test-url-ref-tool"
    local test_url="https://httpbin.org/json"
    
    log_step "Declaring tool with URL reference"
    log_debug "URL: $test_url"
    
    if $PORT42_BIN declare tool "$tool_name" \
        --transforms "api,http,client" \
        --ref "url:$test_url" 2>&1; then
        log_success "Tool with URL reference created"
    else
        log_error "Tool with URL reference failed"
        return 1
    fi
    
    log_step "Verifying URL content was fetched"
    local tool_def=$($PORT42_BIN cat "/tools/$tool_name/definition" 2>/dev/null)
    if echo "$tool_def" | grep -q -i "slideshow\|author"; then
        log_success "URL reference content found in tool definition"
    else
        log_error "URL reference content not found"
    fi
    
    show_object_store "After URL reference tool"
    echo ""
}

# Test 6: P42 VFS reference (referencing previously created tool)
test_p42_vfs_reference() {
    log_test "Tool declaration with P42 VFS reference"
    
    local base_tool="test-basic-tool"
    local enhanced_tool="test-enhanced-tool"
    
    log_step "Declaring tool that references existing tool"
    log_debug "Referencing: p42:/tools/$base_tool"
    
    if $PORT42_BIN declare tool "$enhanced_tool" \
        --transforms "enhanced,analysis,reporting" \
        --ref "p42:/tools/$base_tool" 2>&1; then
        log_success "Tool with P42 VFS reference created"
    else
        log_error "Tool with P42 VFS reference failed"
        return 1
    fi
    
    log_step "Verifying VFS reference incorporation"
    local tool_def=$($PORT42_BIN cat "/tools/$enhanced_tool/definition" 2>/dev/null)
    if echo "$tool_def" | grep -q "$base_tool"; then
        log_success "P42 VFS reference found in tool definition"
    else
        log_error "P42 VFS reference not found"
    fi
    
    show_object_store "After P42 VFS reference"
    echo ""
}

# Test 7: Search reference
test_search_reference() {
    log_test "Tool declaration with search reference"
    
    local tool_name="test-search-ref-tool"
    
    log_step "Declaring tool with search reference"
    log_debug "Search query: 'config validation'"
    
    if $PORT42_BIN declare tool "$tool_name" \
        --transforms "search,intelligent,discovery" \
        --ref "search:config validation" 2>&1; then
        log_success "Tool with search reference created"
    else
        log_error "Tool with search reference failed"
        return 1
    fi
    
    log_step "Verifying search reference"
    local tool_def=$($PORT42_BIN cat "/tools/$tool_name/definition" 2>/dev/null)
    if echo "$tool_def" | grep -q -i "search"; then
        log_success "Search reference incorporated"
    else
        log_error "Search reference not found"
    fi
    
    show_object_store "After search reference"
    echo ""
}

# Test 8: Custom prompt without references
test_custom_prompt() {
    log_test "Tool declaration with custom prompt"
    
    local tool_name="test-prompt-tool"
    local prompt="Create a tool that analyzes CSV data for anomalies and generates security alerts when suspicious patterns are detected"
    
    log_step "Declaring tool with custom prompt"
    log_debug "Prompt: $prompt"
    
    if $PORT42_BIN declare tool "$tool_name" \
        --transforms "analyze,security,alerts" \
        --prompt "$prompt" 2>&1; then
        log_success "Tool with custom prompt created"
    else
        log_error "Tool with custom prompt failed"
        return 1
    fi
    
    log_step "Verifying prompt influence on generated code"
    local tool_code=$($PORT42_BIN cat "/commands/$tool_name" 2>/dev/null)
    local prompt_influence=0
    
    echo "$tool_code" | grep -q -i "csv\|anomal\|security\|alert" && ((prompt_influence++))
    echo "$tool_code" | grep -q -i "suspicious\|pattern" && ((prompt_influence++))
    
    if [[ $prompt_influence -ge 1 ]]; then
        log_success "Custom prompt influenced tool generation"
    else
        log_error "Custom prompt not reflected in generated tool"
    fi
    
    show_object_store "After custom prompt tool"
    echo ""
}

# Test 9: Combined prompt + references (the ultimate test)
test_combined_prompt_references() {
    log_test "Tool declaration with custom prompt AND multiple references"
    
    local tool_name="test-ultimate-tool"
    local prompt="Build a comprehensive data processor that validates CSV against the schema, follows security best practices from the documentation, and integrates with the API specification for external validation"
    
    log_step "Declaring ultimate tool with prompt + multiple references"
    log_debug "Combining custom prompt with file, URL, and P42 references"
    
    if $PORT42_BIN declare tool "$tool_name" \
        --transforms "process,validate,integrate,secure" \
        --ref "file:$TEST_DATA_DIR/data-schema.json" \
        --ref "file:$TEST_DATA_DIR/project-readme.md" \
        --ref "file:$TEST_DATA_DIR/api-spec.yaml" \
        --ref "p42:/tools/test-file-ref-tool" \
        --prompt "$prompt" 2>&1; then
        log_success "Ultimate tool with prompt + references created"
    else
        log_error "Ultimate tool creation failed"
        return 1
    fi
    
    log_step "Verifying comprehensive integration"
    local tool_def=$($PORT42_BIN cat "/tools/$tool_name/definition" 2>/dev/null)
    local tool_code=$($PORT42_BIN cat "/commands/$tool_name" 2>/dev/null)
    
    local integration_score=0
    
    # Check for file references
    echo "$tool_def" | grep -q "json-schema" && ((integration_score++)) && log_success "Schema reference integrated"
    echo "$tool_def" | grep -q "Security" && ((integration_score++)) && log_success "Security docs integrated"
    echo "$tool_def" | grep -q "openapi\|API" && ((integration_score++)) && log_success "API spec integrated"
    
    # Check for prompt influence
    echo "$tool_code" | grep -q -i "validate\|schema\|security" && ((integration_score++)) && log_success "Prompt guidance reflected"
    
    if [[ $integration_score -ge 3 ]]; then
        log_success "Excellent integration of prompt + references"
    else
        log_error "Poor integration: only $integration_score/4 elements found"
    fi
    
    show_object_store "After ultimate combined tool"
    echo ""
}

# Test 10: Artifact creation with prompt
test_artifact_creation() {
    log_test "Artifact creation with custom prompt"
    
    local artifact_name="test-documentation"
    local prompt="Generate comprehensive API documentation with authentication examples, error codes, and integration guides based on the provided specifications"
    
    log_step "Creating artifact with custom prompt"
    
    if $PORT42_BIN declare artifact "$artifact_name" \
        --artifact-type "documentation" \
        --file-type ".md" \
        --prompt "$prompt" 2>&1; then
        log_success "Artifact with custom prompt created"
    else
        log_error "Artifact creation failed"
        return 1
    fi
    
    log_step "Verifying artifact in VFS"
    if $PORT42_BIN ls /artifacts/ 2>/dev/null | grep -q "$artifact_name"; then
        log_success "Artifact appears in VFS"
    else
        log_error "Artifact not found in VFS"
    fi
    
    show_object_store "After artifact creation"
    echo ""
}

# Test 11: Debug mode verification
test_debug_mode() {
    log_test "Debug mode functionality verification"
    
    log_step "Checking debug environment"
    if [[ -n "$PORT42_DEBUG" ]]; then
        log_success "PORT42_DEBUG is set: $PORT42_DEBUG"
    else
        log_error "PORT42_DEBUG not set"
    fi
    
    log_step "Testing debug output capture"
    local debug_tool="test-debug-tool"
    local debug_output=$(PORT42_DEBUG=1 $PORT42_BIN declare tool "$debug_tool" --transforms "debug,test" 2>&1)
    
    if echo "$debug_output" | grep -q -i "debug"; then
        log_success "Debug output detected in tool creation"
        log_debug "Sample debug output captured"
    else
        log_error "No debug output found"
    fi
    
    log_step "Checking daemon logs"
    if $PORT42_BIN daemon logs -n 5 >/dev/null 2>&1; then
        log_success "Daemon logs accessible"
        echo -e "   ${CYAN}Recent daemon log entries:${NC}"
        $PORT42_BIN daemon logs -n 5 | sed 's/^/   /'
    else
        log_error "Cannot access daemon logs"
    fi
    
    echo ""
}

# Test 12: VFS navigation and inspection
test_vfs_navigation() {
    log_test "Virtual filesystem navigation and inspection"
    
    log_step "Testing VFS root navigation"
    local vfs_root=$($PORT42_BIN ls / 2>/dev/null)
    if echo "$vfs_root" | grep -q "tools\|commands\|memory"; then
        log_success "VFS root structure accessible"
    else
        log_error "VFS root structure incomplete"
    fi
    
    log_step "Testing tools directory"
    local tools_count=$($PORT42_BIN ls /tools/ 2>/dev/null | wc -l)
    log_success "Found $tools_count tools in /tools/"
    
    log_step "Testing commands directory"
    local commands_count=$($PORT42_BIN ls /commands/ 2>/dev/null | wc -l)
    log_success "Found $commands_count commands in /commands/"
    
    log_step "Testing tool info command"
    local first_tool=$($PORT42_BIN ls /tools/ 2>/dev/null | head -1)
    if [[ -n "$first_tool" ]]; then
        if $PORT42_BIN info "/tools/$first_tool" >/dev/null 2>&1; then
            log_success "Tool info command working"
        else
            log_error "Tool info command failed"
        fi
    fi
    
    log_step "Testing search functionality"
    if $PORT42_BIN search "test" --limit 5 >/dev/null 2>&1; then
        log_success "Search command working"
    else
        log_error "Search command failed"
    fi
    
    show_object_store "Final VFS state"
    echo ""
}

# Test 13: Error handling and edge cases
test_error_handling() {
    log_test "Error handling and edge cases"
    
    log_step "Testing invalid file reference"
    if ! $PORT42_BIN declare tool "test-invalid-file" --ref "file:/nonexistent/file.txt" >/dev/null 2>&1; then
        log_success "Invalid file reference properly rejected"
    else
        log_error "Invalid file reference not caught"
    fi
    
    log_step "Testing invalid URL reference"
    if ! $PORT42_BIN declare tool "test-invalid-url" --ref "url:invalid-url" >/dev/null 2>&1; then
        log_success "Invalid URL reference properly rejected"
    else
        log_error "Invalid URL reference not caught"
    fi
    
    log_step "Testing empty prompt"
    if $PORT42_BIN declare tool "test-empty-prompt" --prompt "" >/dev/null 2>&1; then
        log_success "Empty prompt handled gracefully"
    else
        log_error "Empty prompt caused failure"
    fi
    
    echo ""
}

# Main test execution
main() {
    echo -e "${CYAN}Starting comprehensive test suite...${NC}"
    echo ""
    
    # Setup
    create_test_files
    test_basic_setup
    
    # Core functionality tests
    test_simple_tool_declaration
    test_file_reference
    test_multiple_file_references
    test_url_reference
    test_p42_vfs_reference
    test_search_reference
    
    # Advanced functionality tests
    test_custom_prompt
    test_combined_prompt_references
    test_artifact_creation
    
    # System tests
    test_debug_mode
    test_vfs_navigation
    test_error_handling
    
    # Summary
    echo -e "${CYAN}================================================================${NC}"
    echo -e "${CYAN}Test Suite Complete${NC}"
    echo -e "${GREEN}✅ Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}❌ Failed: $FAILED_TESTS${NC}"
    echo -e "${YELLOW}📊 Total: $TOTAL_TESTS${NC}"
    echo ""
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "${GREEN}🎉 All tests passed! Universal Prompt & Reference System is working correctly.${NC}"
        exit 0
    else
        echo -e "${RED}💥 Some tests failed. Check the output above for details.${NC}"
        exit 1
    fi
}

# Check if running directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi