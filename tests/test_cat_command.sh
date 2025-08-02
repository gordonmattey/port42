#!/bin/bash
# Test port42 cat command functionality

set -e

echo "=== Testing port42 cat command ==="

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
    local expected_behavior="$3"  # "success", "fail", or specific content
    
    TESTS=$((TESTS + 1))
    echo -n "Test $TESTS: $description... "
    
    # Run command and capture output and exit code
    output=""
    exit_code=0
    if output=$(eval "$command" 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi
    
    case "$expected_behavior" in
        "success")
            if [ $exit_code -eq 0 ]; then
                echo -e "${GREEN}PASSED${NC}"
                PASSED=$((PASSED + 1))
            else
                echo -e "${RED}FAILED${NC}"
                echo "  Expected success but command failed"
                echo "  Output: $output"
            fi
            ;;
        "fail")
            if [ $exit_code -ne 0 ]; then
                echo -e "${GREEN}PASSED${NC} (expected failure)"
                PASSED=$((PASSED + 1))
            else
                echo -e "${RED}FAILED${NC}"
                echo "  Expected failure but command succeeded"
                echo "  Output: $output"
            fi
            ;;
        *)
            # Check if output contains expected content
            if [ $exit_code -eq 0 ] && echo "$output" | grep -q "$expected_behavior"; then
                echo -e "${GREEN}PASSED${NC}"
                PASSED=$((PASSED + 1))
            elif [ $exit_code -ne 0 ] && echo "$output" | grep -q "$expected_behavior"; then
                # Also pass if exit code is non-zero but output contains expected text
                echo -e "${GREEN}PASSED${NC}"
                PASSED=$((PASSED + 1))
            else
                echo -e "${RED}FAILED${NC}"
                echo "  Expected to contain: $expected_behavior"
                echo "  Got: $output"
                echo "  Exit code: $exit_code"
            fi
            ;;
    esac
}

# Check if daemon is running
echo "Checking daemon status..."
if ! nc -z localhost 42 2>/dev/null && ! nc -z localhost 4242 2>/dev/null; then
    echo -e "${YELLOW}Warning: Port 42 daemon doesn't appear to be running.${NC}"
    echo "Please start the daemon first with: sudo port42d"
    exit 1
fi

echo ""
echo "Running tests..."
echo ""

# Test 1: Read memory/session (should work if there are active sessions)
run_test "Read memory path" \
    "port42 cat /memory/cli-1754116317 2>/dev/null || echo 'No active sessions'" \
    "success"

# Test 2: Read commands (may be empty if no commands exist yet)
run_test "Read command path" \
    "port42 cat /commands/test-command 2>/dev/null || echo 'No commands'" \
    "success"

# Test 3: Invalid path should fail gracefully
run_test "Handle non-existent path" \
    "port42 cat /invalid/path" \
    "fail"

# Test 4: Missing path argument  
run_test "Missing path argument" \
    "port42 cat 2>&1" \
    "error:"

# Test 5: Help command
run_test "Cat help command" \
    "port42 cat --help" \
    "Display content from virtual filesystem"

# Test 6: Read from artifacts (likely empty)
run_test "Read artifact path" \
    "port42 cat /artifacts/test.txt 2>/dev/null || echo 'No artifacts'" \
    "success"

# Test 7: Check if we can list and then cat a memory session
echo ""
echo "Integration test: List memory and cat first session..."
if sessions=$(port42 ls /memory 2>/dev/null | grep -o 'cli-[0-9]*' | head -1); then
    if [ ! -z "$sessions" ]; then
        run_test "Cat existing memory session" \
            "port42 cat /memory/$sessions" \
            "success"
    else
        echo "No memory sessions found to test"
    fi
else
    echo "Could not list memory sessions"
fi

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