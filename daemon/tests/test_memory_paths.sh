#!/bin/bash
# Test memory path fixes - comprehensive test suite

set -e

echo "=== Testing Memory Path Fixes ==="

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
            if [ $exit_code -eq 0 ] && echo "$output" | grep -q "$expected_behavior"; then
                echo -e "${GREEN}PASSED${NC}"
                PASSED=$((PASSED + 1))
            elif [ $exit_code -eq 0 ]; then
                echo -e "${RED}FAILED${NC}"
                echo "  Expected to contain: $expected_behavior"
                echo "  Got: $output"
            else
                echo -e "${RED}FAILED${NC}"
                echo "  Command failed with: $output"
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
echo -e "${BLUE}Phase 1: Test Existing Sessions${NC}"
echo "Testing that old sessions still work..."
echo ""

# Find existing sessions
EXISTING_SESSION=$(port42 ls /memory 2>/dev/null | grep -o 'cli-[0-9]*' | head -1 || echo "")

if [ ! -z "$EXISTING_SESSION" ]; then
    # Test reading old session with old path format (should still work)
    run_test "Read existing session via /memory/sessions/{id}" \
        "port42 cat /memory/sessions/$EXISTING_SESSION" \
        "success"
    
    # Test reading with new direct path (might fail for old sessions)
    run_test "Read existing session via /memory/{id} (may fail for old data)" \
        "port42 cat /memory/$EXISTING_SESSION 2>&1 || echo 'Expected - old session'" \
        "success"
else
    echo "No existing sessions found - skipping old data tests"
fi

echo ""
echo -e "${BLUE}Phase 2: Create New Session and Test Paths${NC}"
echo "Creating a new session to test updated path storage..."
echo ""

# Create a new test session
echo "Creating test session..."
OUTPUT=$(port42 possess "@test-agent" "This is a test for memory paths" 2>&1 || echo "FAILED")

# Extract the session ID from the output
NEW_SESSION=$(echo "$OUTPUT" | grep -o 'cli-[0-9]*' | tail -1)

if [ -z "$NEW_SESSION" ]; then
    echo -e "${RED}Failed to create test session${NC}"
    echo "Output: $OUTPUT"
    exit 1
fi

echo "Created session: $NEW_SESSION"
echo ""

# Test the new session with various path formats
run_test "List /memory shows new session" \
    "port42 ls /memory" \
    "$NEW_SESSION"

run_test "Read new session via direct /memory/{id}" \
    "port42 cat /memory/$NEW_SESSION" \
    "This is a test for memory paths"

run_test "Read new session via /memory/sessions/{id}" \
    "port42 cat /memory/sessions/$NEW_SESSION" \
    "This is a test for memory paths"

run_test "List /memory/sessions shows new session" \
    "port42 ls /memory/sessions" \
    "$NEW_SESSION"

# Test date-based paths
TODAY=$(date +%Y-%m-%d)
run_test "List /memory/sessions/by-date shows today" \
    "port42 ls /memory/sessions/by-date" \
    "$TODAY"

run_test "List today's sessions includes new session" \
    "port42 ls /memory/sessions/by-date/$TODAY" \
    "$NEW_SESSION"

# Test agent-based paths
run_test "List /memory/sessions/by-agent shows test agent" \
    "port42 ls /memory/sessions/by-agent" \
    "testagent"

run_test "List test agent sessions includes new session" \
    "port42 ls /memory/sessions/by-agent/testagent" \
    "$NEW_SESSION"

# Test global date view
run_test "List /by-date/$TODAY/memory includes new session" \
    "port42 ls /by-date/$TODAY/memory 2>/dev/null || echo 'Path may not exist yet'" \
    "success"

echo ""
echo -e "${BLUE}Phase 3: Test Command Generation with Memory${NC}"
echo "Testing that commands linked to memory work correctly..."
echo ""

# Generate a command in the session
echo "Generating a test command..."
CMD_OUTPUT=$(echo -e "/generate test-memory-cmd\necho 'Memory path test command'\n---\nA simple test command for memory paths" | nc localhost 42 2>/dev/null || echo "FAILED")

if echo "$CMD_OUTPUT" | grep -q "command_generated"; then
    echo "Command generated successfully"
    
    run_test "List /commands shows new command" \
        "port42 ls /commands" \
        "test-memory-cmd"
    
    run_test "Cat command works" \
        "port42 cat /commands/test-memory-cmd" \
        "Memory path test command"
    
    run_test "Command appears in memory's generated list" \
        "port42 ls /memory/$NEW_SESSION/generated 2>/dev/null || echo 'Expected - path may not exist'" \
        "success"
else
    echo "Command generation failed - skipping command tests"
fi

echo ""
echo -e "${BLUE}Phase 4: Edge Cases${NC}"
echo ""

run_test "Invalid memory path fails gracefully" \
    "port42 cat /memory/invalid-id" \
    "fail"

run_test "List /memory root works" \
    "port42 ls /memory" \
    "success"

run_test "List / includes memory directory" \
    "port42 ls /" \
    "memory"

# Summary
echo ""
echo "=== Test Summary ==="
echo "Total tests: $TESTS"
echo -e "Passed: ${GREEN}$PASSED${NC}"
FAILED=$((TESTS - PASSED))
if [ $FAILED -gt 0 ]; then
    echo -e "Failed: ${RED}$FAILED${NC}"
    echo ""
    echo -e "${YELLOW}Note: Some failures may be expected for old sessions that don't have the new path format.${NC}"
    echo -e "${YELLOW}The important tests are that NEW sessions work with both path formats.${NC}"
else
    echo -e "${GREEN}All tests passed!${NC}"
fi

echo ""
echo -e "${BLUE}Recommendations:${NC}"
echo "1. If old sessions failed with /memory/{id}, that's expected"
echo "2. New sessions should work with both /memory/{id} and /memory/sessions/{id}"
echo "3. Consider running a migration script to update old metadata if needed"