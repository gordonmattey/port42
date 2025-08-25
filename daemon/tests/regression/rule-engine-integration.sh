#!/bin/bash
# tests/regression/rule-engine-integration.sh
# Ensure rule engine continues to work correctly

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../integration/prompt-ref-system/test-helpers.sh"

echo "⚡ Testing rule engine integration (regression)"

setup_test_environment

# Test that rules are active
assert_success "port42 watch rules | grep -E '(enabled|Status: enabled)'" \
    "Rules are active and enabled"

# Test Documentation Rule (3+ transforms triggers docs artifact)
assert_success "port42 declare tool rule-test-complex --transforms 'data,analysis,transform,export'" \
    "Complex tool declaration triggers documentation rule"

# Wait for rule processing
sleep 3

# Check if documentation artifact was created
assert_success "port42 ls /artifacts | grep rule-test-complex-docs" \
    "Documentation rule created artifact"

# Test Git Tools Rule
assert_success "port42 declare tool rule-test-git-tool --transforms 'git,workflow'" \
    "Git tool declaration triggers git tools rule"

# Wait for rule processing  
sleep 3

# Check if git-status-enhanced was created
assert_success "port42 ls /tools | grep git-status-enhanced" \
    "Git tools rule created enhanced tool"

# Test Test Suite Rule
assert_success "port42 declare tool rule-test-testing --transforms 'test,validation'" \
    "Test tool declaration triggers test suite rule"

# Wait for rule processing
sleep 3

# Check if test-runner-enhanced was created
assert_success "port42 ls /tools | grep test-runner-enhanced" \
    "Test suite rule created enhanced tool"

# Test Viewer Rule
assert_success "port42 declare tool rule-test-analyzer --transforms 'analysis,data'" \
    "Analysis tool declaration triggers viewer rule"

# Wait for rule processing
sleep 3

# Check if viewer was created
assert_success "port42 ls /tools | grep view-rule-test-analyzer" \
    "Viewer rule created viewer tool"

# Test Documentation Emergence Rule
assert_success "port42 declare tool rule-test-wiki --transforms 'wiki,content'" \
    "Wiki tool declaration triggers documentation emergence rule"

# Wait for rule processing
sleep 5

# Check if documentation infrastructure was created
assert_success "port42 ls /tools | grep -E '(doc-template-generator|doc-validator|doc-site-builder)'" \
    "Documentation emergence rule created infrastructure tools"

teardown_test_environment

echo "✅ All rule engine integration tests passed"