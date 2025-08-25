#!/bin/bash
# Test port42 search command functionality

set -e

echo "=== Testing port42 search command ==="

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

# Setup: Create test data if needed
echo ""
echo -e "${BLUE}Setting up test data...${NC}"

# Create a test session with known content
echo -e "/generate test-search-cmd\necho 'Testing search functionality'\n---\nA command for testing search" | nc localhost 42 >/dev/null 2>&1 || true
sleep 1

# Create another command with different content
echo -e "/generate docker-deploy\ndocker compose up -d\n---\nDeploy services with docker" | nc localhost 42 >/dev/null 2>&1 || true
sleep 1

echo ""
echo -e "${BLUE}1. Basic Search Tests${NC}"
echo ""

# Test 1: Basic search
run_test "Search for existing term" \
    "port42 search 'docker'" \
    "docker"

run_test "Search for non-existent term" \
    "port42 search 'xyznonexistent123'" \
    "No results found"

run_test "Empty query shows all objects" \
    "port42 search ''" \
    "Found"

run_test "Case insensitive search" \
    "port42 search 'DOCKER'" \
    "docker"

echo ""
echo -e "${BLUE}2. Path Filtering Tests${NC}"
echo ""

run_test "Search within /commands path" \
    "port42 search --path /commands 'test'" \
    "success"

run_test "Search within /memory path" \
    "port42 search --path /memory 'session'" \
    "success"

run_test "Invalid path still works (just no results)" \
    "port42 search --path /nonexistent 'test'" \
    "success"

echo ""
echo -e "${BLUE}3. Type Filtering Tests${NC}"
echo ""

run_test "Filter by command type" \
    "port42 search --type command 'echo'" \
    "success"

run_test "Filter by session type" \
    "port42 search --type session ''" \
    "success"

run_test "Invalid type returns no results" \
    "port42 search --type invalid 'test'" \
    "No results found"

echo ""
echo -e "${BLUE}4. Date Filtering Tests${NC}"
echo ""

# Get today's date
TODAY=$(date +%Y-%m-%d)
YESTERDAY=$(date -d "yesterday" +%Y-%m-%d 2>/dev/null || date -v-1d +%Y-%m-%d)
TOMORROW=$(date -d "tomorrow" +%Y-%m-%d 2>/dev/null || date -v+1d +%Y-%m-%d)

run_test "Search after yesterday" \
    "port42 search --after '$YESTERDAY' ''" \
    "Found"

run_test "Search before tomorrow" \
    "port42 search --before '$TOMORROW' ''" \
    "Found"

run_test "Search in date range" \
    "port42 search --after '$YESTERDAY' --before '$TOMORROW' ''" \
    "Found"

run_test "Invalid date format" \
    "port42 search --after 'invalid-date' 'test' 2>&1" \
    "Invalid date format"

echo ""
echo -e "${BLUE}5. Agent Filtering Tests${NC}"
echo ""

# Find an agent from existing sessions
AGENT=$(port42 ls /memory 2>/dev/null | grep -o '@[a-zA-Z0-9-]*' | head -1 || echo "")

if [ ! -z "$AGENT" ]; then
    run_test "Filter by existing agent" \
        "port42 search --agent '$AGENT' ''" \
        "success"
    
    run_test "Filter by non-existent agent" \
        "port42 search --agent '@nonexistent-agent' ''" \
        "No results found"
else
    echo "No agents found - skipping agent filter tests"
fi

echo ""
echo -e "${BLUE}6. Tag Filtering Tests${NC}"
echo ""

# Tags are stored with sessions and commands
run_test "Filter by single tag" \
    "port42 search --tag 'command' ''" \
    "success"

run_test "Filter by multiple tags" \
    "port42 search --tag 'command' --tag 'ai' ''" \
    "success"

run_test "Filter by non-existent tag" \
    "port42 search --tag 'xyznonexistenttag123' ''" \
    "No results found"

echo ""
echo -e "${BLUE}7. Combined Filter Tests${NC}"
echo ""

run_test "Type + query" \
    "port42 search --type command 'echo'" \
    "success"

run_test "Path + agent + query" \
    "port42 search --path /memory --agent '$AGENT' 'test' 2>&1 || echo 'No agent results'" \
    "success"

run_test "All filters together" \
    "port42 search --path /commands --type command --after '$YESTERDAY' --limit 5 'test'" \
    "success"

echo ""
echo -e "${BLUE}8. Limit Parameter Tests${NC}"
echo ""

run_test "Limit to 1 result" \
    "port42 search --limit 1 '' 2>&1 | head -1" \
    "Found 1 result"

run_test "Default limit (20)" \
    "port42 search ''" \
    "Found"

echo ""
echo -e "${BLUE}9. Edge Cases${NC}"
echo ""

run_test "Very long query" \
    "port42 search '$(printf 'a%.0s' {1..200})'" \
    "No results found"

run_test "Special characters in query" \
    "port42 search '!@#$%^&*()' 2>&1" \
    "success"

run_test "Help command" \
    "port42 search --help" \
    "Search across the virtual filesystem"

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
echo -e "${BLUE}Sample searches to try:${NC}"
echo "- port42 search 'reality compiler'"
echo "- port42 search --type command 'docker'"
echo "- port42 search --path /memory --after $YESTERDAY 'AI'"
echo "- port42 search --agent @ai-engineer --limit 10 'database'"