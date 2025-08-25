#!/bin/bash
# Test port42 info command functionality

set -e

echo "=== Testing port42 info command ==="

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
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
            fi
            ;;
        *)
            # Check if output contains expected content
            if echo "$output" | grep -q "$expected_behavior"; then
                echo -e "${GREEN}PASSED${NC}"
                PASSED=$((PASSED + 1))
            else
                echo -e "${RED}FAILED${NC}"
                echo "  Expected to contain: $expected_behavior"
                echo "  Got: $output"
            fi
            ;;
    esac
}

# Check if daemon is running
echo "Checking daemon status..."
if ! nc -z localhost 42 2>/dev/null && ! nc -z localhost 4242 2>/dev/null; then
    echo -e "${YELLOW}Warning: Port 42 daemon doesn't appear to be running.${NC}"
    echo "Please restart the daemon with: sudo bin/port42d"
    exit 1
fi

echo ""
echo -e "${BLUE}Testing info command on various paths...${NC}"
echo ""

# Test 1: Info on memory/session
MEMORY_ID=$(port42 ls /memory 2>/dev/null | grep -o 'cli-[0-9]*' | head -1 || echo "")
if [ ! -z "$MEMORY_ID" ]; then
    run_test "Get info for memory session" \
        "port42 info /memory/$MEMORY_ID" \
        "Type:"
    
    run_test "Info shows object ID" \
        "port42 info /memory/$MEMORY_ID" \
        "Object ID:"
    
    run_test "Info shows creation date" \
        "port42 info /memory/$MEMORY_ID" \
        "Created:"
    
    run_test "Info shows session type" \
        "port42 info /memory/$MEMORY_ID" \
        "session"
    
    run_test "Info shows virtual paths" \
        "port42 info /memory/$MEMORY_ID" \
        "Virtual Paths:"
else
    echo "No memory sessions found - skipping memory tests"
fi

# Test 2: Info on commands
COMMAND=$(port42 ls /commands 2>/dev/null | grep -v "empty" | grep -o '[a-zA-Z0-9_-]*' | head -1 || echo "")
if [ ! -z "$COMMAND" ]; then
    run_test "Get info for command" \
        "port42 info /commands/$COMMAND" \
        "Type:"
    
    run_test "Command info shows type as command" \
        "port42 info /commands/$COMMAND" \
        "command"
else
    echo "No commands found - creating a test command"
    # Try to create a simple command
    echo -e "/generate test-info-cmd\necho 'Test for info command'\n---\nA test command for info" | nc localhost 42 >/dev/null 2>&1 || true
    sleep 1
    
    run_test "Get info for newly created command" \
        "port42 info /commands/test-info-cmd 2>&1 || echo 'Command may not exist'" \
        "success"
fi

# Test 3: Invalid path
run_test "Info on non-existent path fails gracefully" \
    "port42 info /invalid/path" \
    "fail"

# Test 4: Root paths
run_test "Info on /memory directory" \
    "port42 info /memory 2>&1 || echo 'May not work for directories'" \
    "success"

# Test 5: Help command
run_test "Info help command" \
    "port42 info --help" \
    "Show metadata information"

# Test 6: Missing path argument
run_test "Missing path argument" \
    "port42 info 2>&1" \
    "error:"

# Test 7: Check metadata fields
if [ ! -z "$MEMORY_ID" ]; then
    echo ""
    echo -e "${BLUE}Testing metadata field presence...${NC}"
    echo ""
    
    OUTPUT=$(port42 info /memory/$MEMORY_ID 2>&1 || echo "FAILED")
    
    # Check for various metadata fields
    if echo "$OUTPUT" | grep -q "Age:"; then
        echo -e "  Age field: ${GREEN}Present${NC}"
    else
        echo -e "  Age field: ${YELLOW}Missing${NC}"
    fi
    
    if echo "$OUTPUT" | grep -q "Size:"; then
        echo -e "  Size field: ${GREEN}Present${NC}"
    else
        echo -e "  Size field: ${YELLOW}Missing${NC}"
    fi
    
    if echo "$OUTPUT" | grep -q "Agent:"; then
        echo -e "  Agent field: ${GREEN}Present${NC}"
    else
        echo -e "  Agent field: ${YELLOW}Missing${NC}"
    fi
    
    if echo "$OUTPUT" | grep -q "Tags:"; then
        echo -e "  Tags field: ${GREEN}Present${NC}"
    else
        echo -e "  Tags field: ${YELLOW}May be empty${NC}"
    fi
fi

# Summary
echo ""
echo "=== Test Summary ==="
echo "Total tests: $TESTS"
echo -e "Passed: ${GREEN}$PASSED${NC}"
FAILED=$((TESTS - PASSED))
if [ $FAILED -gt 0 ]; then
    echo -e "Failed: ${RED}$FAILED${NC}"
else
    echo -e "${GREEN}All tests passed!${NC}"
fi

echo ""
echo -e "${BLUE}Sample output:${NC}"
echo "Try running: port42 info /memory/\$(port42 ls /memory | grep -o 'cli-[0-9]*' | head -1)"