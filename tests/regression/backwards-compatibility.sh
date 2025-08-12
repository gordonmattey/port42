#!/bin/bash
# tests/regression/backwards-compatibility.sh  
# Ensure old command syntax continues to work

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../integration/prompt-ref-system/test-helpers.sh"

echo "🔄 Testing backwards compatibility (regression)"

setup_test_environment

# Test all current command variations still work

# Basic tool declaration (no optional params)
assert_success "port42 declare tool compat-test-1" \
    "Tool declaration without transforms works"

# Tool with transforms only
assert_success "port42 declare tool compat-test-2 --transforms 'test,basic'" \
    "Tool declaration with transforms works"

# Tool with existing reference support
assert_success "port42 declare tool compat-test-3 --transforms 'test' --ref p42:/tools/compat-test-1" \
    "Tool declaration with references works"

# Basic artifact declaration
assert_success "port42 declare artifact compat-artifact-1" \
    "Basic artifact declaration works"

# Artifact with type
assert_success "port42 declare artifact compat-artifact-2 --artifact-type 'documentation'" \
    "Artifact with type works"

# Artifact with file type
assert_success "port42 declare artifact compat-artifact-3 --file-type '.json'" \
    "Artifact with file type works"

# Artifact with both parameters
assert_success "port42 declare artifact compat-artifact-4 --artifact-type 'specification' --file-type '.yaml'" \
    "Artifact with both type and file-type works"

# Basic possession (existing syntax)
assert_success "timeout 10 port42 possess @ai-engineer 'test backwards compatibility'" \
    "Basic possession works"

# Possession with search (existing syntax)
assert_success "timeout 10 port42 possess @ai-engineer --search 'test' 'find compatibility info'" \
    "Possession with search works"

# Test help commands still work
assert_success "port42 --help" \
    "Main help command works"

assert_success "port42 declare tool --help" \
    "Tool help command works"

assert_success "port42 declare artifact --help" \
    "Artifact help command works"

assert_success "port42 possess --help" \
    "Possess help command works"

# Test other existing commands
assert_success "port42 ls /tools" \
    "ls command works"

assert_success "port42 status" \
    "status command works"

# Test that created tools/artifacts are accessible
sleep 2  # Allow materialization
assert_success "ls ~/.port42/commands/compat-test-1" \
    "Created tools are accessible"

teardown_test_environment

echo "✅ All backwards compatibility tests passed"