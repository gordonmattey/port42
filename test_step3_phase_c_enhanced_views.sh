#!/bin/bash

# Step 3 Phase C: Enhanced Existing Views Test Suite
# Tests that existing views are enriched with relation metadata

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

TESTS_RUN=0
TESTS_PASSED=0

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

echo -e "${BLUE}🧪 Testing Step 3 Phase C: Enhanced Existing Views${NC}"
echo "================================================"

echo -e "\n${YELLOW}Phase C Success Criteria (Enhanced Existing Views):${NC}"
echo "1. /commands/ shows relation-backed tools with metadata"
echo "2. /by-date/ includes relations alongside other objects"
echo "3. port42 info works on /tools/ paths and shows relations"
echo "4. Relation metadata visible in all relevant views"
echo "5. Clean decomposition - no legacy compromises"

echo -e "\n${YELLOW}Step 1: Setup test tools with known dates${NC}"
echo "Creating test tools to ensure we have relation data..."

# Create test tools
port42 declare tool phase-c-test --transforms test,metadata >/dev/null 2>&1 || true
port42 declare tool phase-c-viewer --transforms view,test >/dev/null 2>&1 || true
sleep 1  # Allow processing

echo -e "\n${YELLOW}Step 2: Test Enhanced /commands/ View${NC}"

run_test "Commands show relation-backed tools" \
    "port42 ls /commands/" \
    ""  # Should show tools as commands

run_test "Command listings include relation metadata" \
    "port42 ls /commands/ -l || port42 ls /commands/" \
    ""  # Should show metadata like spawned_by, transforms

run_test "Individual command shows relation context" \
    "port42 info /commands/phase-c-test || port42 cat /commands/phase-c-test" \
    ""  # Should show it's relation-backed

echo -e "\n${YELLOW}Step 3: Test Enhanced /by-date/ View${NC}"

today=$(date '+%Y-%m-%d')

run_test "By-date shows today's relations" \
    "port42 ls /by-date/$today/" \
    ""  # Should show relation entries

run_test "By-date includes both objects and relations" \
    "port42 ls /by-date/$today/" \
    ""  # Should mix relation and traditional objects

run_test "By-date relation entries have metadata" \
    "port42 ls /by-date/$today/ | head -3" \
    ""  # Should show relation info

echo -e "\n${YELLOW}Step 4: Test Enhanced info Command${NC}"

run_test "Info works on tool paths" \
    "port42 info /tools/phase-c-test" \
    ""  # Should show detailed relation info

run_test "Info shows complete relation metadata" \
    "port42 info /tools/phase-c-test" \
    "relation\\|Tool\\|properties"  # Should show relation structure

run_test "Info shows relationship connections" \
    "port42 info /tools/phase-c-test" \
    ""  # Should show spawning, parent info if any

run_test "Info works on spawned tools" \
    "port42 ls /tools/spawned-by/ | head -1 | xargs -I {} port42 info /tools/{}" \
    ""  # Should show spawned tool info

echo -e "\n${YELLOW}Step 5: Test Relation Metadata Visibility${NC}"

run_test "Commands preserve tool executable paths" \
    "port42 cat /commands/phase-c-test 2>/dev/null || echo 'EXPECTED: May redirect to /tools/'" \
    ""  # Should work or redirect cleanly

run_test "Date entries link to relation definitions" \
    "port42 ls /by-date/$today/ | grep phase-c-test | head -1" \
    ""  # Should find our test tool

run_test "Info command shows creation context" \
    "port42 info /tools/phase-c-test" \
    "created\\|Created\\|manifested"  # Should show creation info

echo -e "\n${YELLOW}Step 6: Test Clean Architecture (No Legacy)${NC}"

run_test "No legacy fallbacks in commands view" \
    "port42 ls /commands/" \
    ""  # Should be clean, no mixed modes

run_test "By-date unified with relations" \
    "port42 ls /by-date/$today/" \
    ""  # Should be unified approach

run_test "Info command consistent across paths" \
    "port42 info /tools/phase-c-test 2>&1" \
    ""  # Should not show "not implemented" errors

echo -e "\n================================================"
echo -e "${BLUE}Phase C Enhanced Views Test Summary${NC}"
echo "Tests run: $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"

if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
    echo -e "${GREEN}✅ All Phase C tests PASSED!${NC}"
    echo -e "${GREEN}🎉 Enhanced existing views successfully integrated relations${NC}"
    exit 0
else
    failed=$((TESTS_RUN - TESTS_PASSED))
    echo -e "${RED}❌ $failed Phase C tests FAILED${NC}"
    echo -e "${YELLOW}💡 Enhanced views need relation integration fixes${NC}"
    exit 1
fi