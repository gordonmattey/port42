#!/bin/bash
# tests/regression/existing-functionality.sh
# Ensure existing Port 42 functionality continues to work

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../integration/prompt-ref-system/test-helpers.sh"

echo "🔍 Testing existing functionality (regression)"

setup_test_environment

# Test basic tool declaration (current functionality)
assert_success "port42 declare tool regression-test-basic --transforms 'test,basic'" \
    "Basic tool declaration works"

# Test tool execution
sleep 2  # Allow time for materialization
assert_success "ls ~/.port42/commands/regression-test-basic" \
    "Tool file was created"

# Test basic artifact declaration (current functionality) 
assert_success "port42 declare artifact regression-test-doc --artifact-type 'documentation'" \
    "Basic artifact declaration works"

# Test VFS navigation
assert_success "port42 ls /tools | grep regression-test-basic" \
    "Tool appears in VFS"

assert_success "port42 ls /artifacts/document/ | grep regression-test-doc || port42 ls /artifacts | grep -E '(document|regression-test-doc)'" \
    "Artifact appears in VFS"

# Test possession mode (current functionality)
assert_success "timeout 10 port42 possess @ai-engineer 'hello' | head -1" \
    "Possession mode works"

# Test status command
assert_success "port42 status" \
    "Status command works"

# Test existing reference functionality (tools already support --ref)
assert_success "port42 declare tool regression-ref-test --transforms 'test' --ref p42:/tools/regression-test-basic" \
    "Existing tool reference functionality works"

teardown_test_environment

echo "✅ All existing functionality tests passed"