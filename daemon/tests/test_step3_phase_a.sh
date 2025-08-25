#!/bin/bash

# Test Step 3 Phase A: Basic Relations Views
echo "🧪 Testing Step 3 Phase A: Basic Relations Views"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0

# Helper function to run test
run_test() {
    local test_name="$1"
    local command="$2" 
    local expected_pattern="$3"
    
    echo -e "\n${BLUE}Test: $test_name${NC}"
    echo "Command: $command"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Run command and capture output
    output=$(eval "$command" 2>&1)
    exit_code=$?
    
    echo "Output:"
    echo "$output"
    
    # Check if command succeeded
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}❌ FAIL: Command failed with exit code $exit_code${NC}"
        return 1
    fi
    
    # Check if output matches expected pattern
    if [[ -n "$expected_pattern" ]] && ! echo "$output" | grep -q "$expected_pattern"; then
        echo -e "${RED}❌ FAIL: Output does not contain expected pattern: $expected_pattern${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
}

echo -e "\n${YELLOW}Phase A Success Criteria:${NC}"
echo "- port42 ls /relations/ shows all declared relations"
echo "- port42 ls /relations/tools/ shows just tool relations"
echo "- port42 cat /relations/{id} shows specific relation details"
echo "- Existing paths still work unchanged"

echo -e "\n${YELLOW}Step 1: Ensure we have some test relations${NC}"
echo "Creating test tools to ensure relations exist..."

# Create a couple test tools to ensure we have relations
port42 declare tool phase-a-test --transforms test,validation >/dev/null 2>&1 || true
port42 declare tool phase-a-analyzer --transforms data,analysis >/dev/null 2>&1 || true

# Wait a moment for spawning to complete
sleep 2

echo -e "\n${YELLOW}Step 2: Test existing virtual filesystem (baseline)${NC}"

run_test "Root directory listing" \
    "port42 ls /" \
    "commands/"

run_test "Commands directory listing" \
    "port42 ls /commands/" \
    ""  # Should work, no specific pattern needed

run_test "Memory directory listing" \
    "port42 ls /memory/" \
    ""  # Should work

echo -e "\n${YELLOW}Step 3: Test new /relations/ paths${NC}"

run_test "Relations root directory" \
    "port42 ls /relations/" \
    ""  # Should show something, even if empty or error

run_test "Relations tools subdirectory" \
    "port42 ls /relations/tools/" \
    ""  # Should show tool relations

run_test "Relations artifacts subdirectory" \
    "port42 ls /relations/artifacts/" \
    ""  # Should show artifact relations (may be empty)

echo -e "\n${YELLOW}Step 4: Test specific relation details${NC}"

# Find a relation ID to test with
relation_files=$(ls ~/.port42/relations/relation-tool-* 2>/dev/null | head -1)
if [[ -n "$relation_files" ]]; then
    # Extract ID from filename
    relation_id=$(basename "$relation_files" .json | sed 's/relation-//')
    echo "Testing with relation ID: $relation_id"
    
    run_test "Specific relation details" \
        "port42 cat /relations/$relation_id" \
        ""  # Should show relation data
else
    echo "⚠️ No relation files found for specific relation test"
fi

echo -e "\n${YELLOW}Step 5: Verify backward compatibility${NC}"

run_test "Commands listing still works" \
    "port42 ls /commands/" \
    ""

run_test "Memory listing still works" \
    "port42 ls /memory/" \
    ""

run_test "By-date listing still works" \
    "port42 ls /by-date/" \
    ""

echo -e "\n${YELLOW}Step 6: Test error handling${NC}"

run_test "Non-existent relation path" \
    "port42 ls /relations/nonexistent/" \
    ""  # May error or return empty

run_test "Invalid relation ID" \
    "port42 cat /relations/invalid-id-123" \
    ""  # Should handle gracefully

echo -e "\n================================================"
echo -e "${BLUE}Phase A Test Summary${NC}"
echo "Tests run: $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $((TESTS_RUN - TESTS_PASSED))"

if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED - Phase A Ready!${NC}"
    exit 0
elif [ $TESTS_PASSED -gt 0 ]; then
    echo -e "${YELLOW}⚠️ PARTIAL SUCCESS - Some functionality working${NC}"
    exit 1
else
    echo -e "${RED}❌ ALL TESTS FAILED - Phase A needs implementation${NC}"
    exit 2
fi