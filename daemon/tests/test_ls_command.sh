#!/bin/bash
# Test port42 ls command functionality

set -e

echo "=== Testing port42 ls command ==="

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Test counter
TESTS=0
PASSED=0

# Helper function to run a test
run_test() {
    local description="$1"
    local command="$2"
    local expected_contains="$3"
    
    TESTS=$((TESTS + 1))
    echo -n "Test $TESTS: $description... "
    
    # Run command and capture output
    if output=$(eval "$command" 2>&1); then
        if [ -z "$expected_contains" ] || echo "$output" | grep -q "$expected_contains"; then
            echo -e "${GREEN}PASSED${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "${RED}FAILED${NC}"
            echo "  Expected to contain: $expected_contains"
            echo "  Got: $output"
        fi
    else
        if [ "$expected_contains" = "SHOULD_FAIL" ]; then
            echo -e "${GREEN}PASSED${NC} (expected failure)"
            PASSED=$((PASSED + 1))
        else
            echo -e "${RED}FAILED${NC}"
            echo "  Command failed with: $output"
        fi
    fi
}

# Check if daemon is running
echo "Checking daemon status..."
if ! nc -z localhost 42 2>/dev/null && ! nc -z localhost 4242 2>/dev/null; then
    echo -e "${YELLOW}Warning: Port 42 daemon doesn't appear to be running.${NC}"
    echo "Please start the daemon first with: port42 daemon start"
    exit 1
fi

echo ""
echo "Running tests..."
echo ""

# Test 1: Root listing
run_test "List root directory" \
    "port42 ls /" \
    "commands"

# Test 2: Default to root
run_test "Default to root when no path given" \
    "port42 ls" \
    "commands"

# Test 3: List commands directory
run_test "List /commands directory" \
    "port42 ls /commands" \
    ""

# Test 4: List memory directory
run_test "List /memory directory" \
    "port42 ls /memory" \
    ""

# Test 5: List by-date directory
run_test "List /by-date directory" \
    "port42 ls /by-date" \
    ""

# Test 6: List artifacts directory
run_test "List /artifacts directory" \
    "port42 ls /artifacts" \
    ""

# Test 7: Invalid path (should fail gracefully)
run_test "Handle invalid path" \
    "port42 ls /invalid/path" \
    "(empty)"

# Test 8: Nested path - today's date
TODAY=$(date +%Y-%m-%d)
run_test "List today's entries" \
    "port42 ls /by-date/$TODAY" \
    ""

# Summary
echo ""
echo "=== Test Summary ==="
echo "Total tests: $TESTS"
echo -e "Passed: ${GREEN}$PASSED${NC}"
FAILED=$((TESTS - PASSED))
if [ $FAILED -gt 0 ]; then
    echo -e "Failed: ${RED}$FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
fi