#!/bin/bash
# tests/run-prompt-ref-tests.sh
# Universal Prompt & Reference System Test Runner

set -e

echo "🧪 Running Universal Prompt & Reference Integration Tests"
echo "========================================================"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test result tracking
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

run_test() {
    local test_file="$1"
    local test_name="$(basename "$test_file" .sh)"
    
    echo -e "\n${YELLOW}Running: $test_name${NC}"
    ((TESTS_RUN++))
    
    if [[ -x "$test_file" ]]; then
        if "$test_file"; then
            echo -e "${GREEN}✅ PASSED: $test_name${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}❌ FAILED: $test_name${NC}"
            ((TESTS_FAILED++))
        fi
    else
        echo -e "${YELLOW}⏭️  SKIPPED: $test_name (not executable or doesn't exist)${NC}"
    fi
}

# Run regression tests first
echo -e "\n📋 Running Regression Tests..."
echo "==============================="

run_test "tests/regression/existing-functionality.sh"
run_test "tests/regression/rule-engine-integration.sh" 
run_test "tests/regression/backwards-compatibility.sh"

# Run integration tests
echo -e "\n🔧 Running Integration Tests..."
echo "================================"

for test in tests/integration/prompt-ref-system/*.sh; do
    if [[ "$test" != *"test-helpers.sh" && -f "$test" ]]; then
        run_test "$test"
    fi
done

# Summary
echo -e "\n📊 Test Summary"
echo "==============="
echo -e "Tests Run: $TESTS_RUN"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
else
    echo -e "Failed: $TESTS_FAILED"
fi

if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "\n${RED}❌ Some tests failed!${NC}"
    exit 1
else
    echo -e "\n${GREEN}✅ All tests passed!${NC}"
    exit 0
fi