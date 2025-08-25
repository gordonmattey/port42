#!/bin/bash
# Port 42 Test Suite Runner
# Runs all tests and provides a summary

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0
SKIPPED=0

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║      Port 42 Test Suite Runner       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo

# Function to run a test
run_test() {
    local test_name=$1
    local test_file=$2
    
    echo -n "Running $test_name... "
    
    if [ ! -f "$test_file" ]; then
        echo -e "${YELLOW}SKIPPED${NC} (file not found)"
        ((SKIPPED++))
        return
    fi
    
    # Create a temp file for output
    local output_file=$(mktemp)
    
    # Run the test based on file extension
    if [[ "$test_file" == *.sh ]]; then
        if bash "$test_file" > "$output_file" 2>&1; then
            echo -e "${GREEN}PASSED${NC}"
            ((PASSED++))
        else
            echo -e "${RED}FAILED${NC}"
            echo "  Output:"
            sed 's/^/    /' "$output_file"
            ((FAILED++))
        fi
    elif [[ "$test_file" == *.py ]]; then
        if python3 "$test_file" > "$output_file" 2>&1; then
            echo -e "${GREEN}PASSED${NC}"
            ((PASSED++))
        else
            echo -e "${RED}FAILED${NC}"
            echo "  Output:"
            sed 's/^/    /' "$output_file"
            ((FAILED++))
        fi
    else
        echo -e "${YELLOW}SKIPPED${NC} (unknown file type)"
        ((SKIPPED++))
    fi
    
    rm -f "$output_file"
}

# Check prerequisites
echo "Checking prerequisites..."
if ! command -v nc >/dev/null 2>&1; then
    echo -e "${YELLOW}Warning: 'nc' (netcat) not found. Some tests may fail.${NC}"
fi
if ! command -v python3 >/dev/null 2>&1; then
    echo -e "${YELLOW}Warning: 'python3' not found. Python tests will be skipped.${NC}"
fi

# Check if daemon is running
echo "Checking daemon status..."
if ! nc -z localhost 42 2>/dev/null && ! nc -z localhost 4242 2>/dev/null; then
    echo -e "${YELLOW}Warning: Port 42 daemon doesn't appear to be running.${NC}"
    echo -e "${YELLOW}Start it with: port42 daemon start${NC}"
    echo
fi

# Run all tests
echo -e "\n${BLUE}Running tests...${NC}\n"

# Basic connectivity tests
run_test "TCP Connection" "test_tcp.sh"

# Protocol tests
run_test "JSON Protocol (bash)" "test_json_protocol.sh"
run_test "JSON Protocol (python)" "test_json_protocol.py"

# Daemon structure tests
run_test "Daemon Structure" "test_daemon_structure.py"

# AI integration tests
run_test "AI Possession" "test_ai_possession.py"
run_test "AI Possession v2" "test_ai_possession_v2.py"

# Memory and persistence tests
run_test "Memory Persistence" "test_memory_persistence.py"
run_test "Session Recovery" "test_session_recovery.py"
run_test "Restart Continuation" "test_restart_continuation.py"

# Feature tests
run_test "Dependency Handling" "test_dependency_handling.py"

# Print summary
echo
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Test Summary               ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo
echo -e "  ${GREEN}Passed:${NC}  $PASSED"
echo -e "  ${RED}Failed:${NC}  $FAILED"
echo -e "  ${YELLOW}Skipped:${NC} $SKIPPED"
echo -e "  Total:    $((PASSED + FAILED + SKIPPED))"
echo

# Exit with appropriate code
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi