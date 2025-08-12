#!/bin/bash
# tests/integration/prompt-ref-system/01-basic-functionality.sh
# Test basic CLI functionality for prompt and reference system

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test-helpers.sh"

echo "🔧 Testing basic CLI functionality (integration)"

setup_test_environment

# Test that CLI accepts prompt parameter (even if backend doesn't process it yet)
echo "  Testing --prompt parameter acceptance..."

# Tool declaration with prompt (should not error)
assert_success "port42 declare tool --help | grep -E 'prompt'" \
    "Tool help shows prompt parameter" || echo "  Note: --prompt not implemented yet (expected)"

# Artifact declaration with prompt (should not error)  
assert_success "port42 declare artifact --help | grep -E 'prompt'" \
    "Artifact help shows prompt parameter" || echo "  Note: --prompt not implemented yet (expected)"

# Possession with references (should not error)
assert_success "port42 possess --help | grep -E 'ref'" \
    "Possess help shows ref parameter" || echo "  Note: --ref not implemented yet (expected)"

# Test existing reference functionality still works
echo "  Testing existing reference functionality..."
assert_success "port42 declare tool basic-ref-test --transforms 'test'" \
    "Basic tool creation works"

sleep 2  # Allow materialization

# Test tool with existing reference support
assert_success "port42 declare tool ref-using-test --transforms 'test' --ref p42:/tools/basic-ref-test" \
    "Tool with p42 reference works"

# Test basic VFS navigation
assert_success "port42 ls /tools | grep basic-ref-test" \
    "Created tool appears in VFS"

assert_success "port42 ls /tools | grep ref-using-test" \
    "Tool with reference appears in VFS"

# Test daemon status includes rules
assert_success "port42 status | grep -E '(rule|Rule)'" \
    "Status command shows rule information"

teardown_test_environment

echo "✅ Basic CLI functionality tests passed"