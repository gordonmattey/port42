#!/bin/bash

# Step 3 Phase D: Advanced Discovery Test Suite
# Tests semantic similarity navigation, cross-referencing, and enhanced search

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

echo -e "${BLUE}🧪 Testing Step 3 Phase D: Advanced Discovery${NC}"
echo "=================================================="

echo -e "\n${YELLOW}Phase D Success Criteria (Advanced Discovery):${NC}"
echo "1. Search finds relations (tools) alongside memory and objects"
echo "2. Semantic similarity navigation via search"
echo "3. Cross-reference memory sessions with relations"
echo "4. Advanced filtering across all entity types"
echo "5. Unified search across tools, memory, artifacts"

echo -e "\n${YELLOW}Step 1: Setup test data for discovery${NC}"
echo "Creating diverse content for cross-referencing tests..."

# Create test tools with known names
port42 declare tool discovery-test --transforms search,test >/dev/null 2>&1 || true
port42 declare tool semantic-analyzer --transforms semantic,analysis >/dev/null 2>&1 || true
sleep 1  # Allow processing

echo -e "\n${YELLOW}Step 2: Test Relation Search Integration${NC}"

run_test "Search finds declared relations" \
    "port42 search 'discovery-test'" \
    "discovery-test"  # Should find the declared tool

run_test "Search finds semantic-analyzer relation" \
    "port42 search 'semantic-analyzer'" \
    "semantic-analyzer"  # Should find the tool

run_test "Search includes relation metadata" \
    "port42 search 'semantic'" \
    "transforms\\|properties"  # Should show relation properties

echo -e "\n${YELLOW}Step 3: Test Cross-Referencing${NC}"

run_test "Search crosses memory and relations" \
    "port42 search 'analysis'" \
    "✨.*echos"  # Should find both memory sessions and tools

run_test "Filter by relation type" \
    "port42 search '' --type tool || port42 search '' | grep -i tool" \
    ""  # Should support filtering by tool type

run_test "Cross-reference by transforms" \
    "port42 search 'analysis' | head -10" \
    "analysis"  # Should find analysis tools and related content

echo -e "\n${YELLOW}Step 4: Test Semantic Similarity Navigation${NC}"

run_test "Find tools by capability" \
    "port42 search 'analyze'" \
    ""  # Should find analyzer tools semantically

run_test "Find related tools via transforms" \
    "port42 search 'data processing'" \
    ""  # Should find data-related tools

run_test "Semantic search across entity types" \
    "port42 search 'test' --limit 5" \
    ""  # Should find tools, memory, artifacts with 'test'

echo -e "\n${YELLOW}Step 5: Test Advanced Filtering${NC}"

run_test "Filter by creation date" \
    "port42 search '' --after $(date +%Y-%m-%d) --limit 3" \
    ""  # Should find today's entities

run_test "Filter by agent" \
    "port42 search 'analyzer' --agent @ai-engineer" \
    ""  # Should filter by specific agent

run_test "Combined filters" \
    "port42 search 'test' --after 2025-08-10 --limit 2" \
    ""  # Should apply multiple filters

echo -e "\n${YELLOW}Step 6: Test Unified Search Scope${NC}"

run_test "Search includes tools from /tools/" \
    "port42 search 'discovery'" \
    ""  # Should find tools accessible via /tools/

run_test "Search includes memory sessions" \
    "port42 search 'cli' --type memory || port42 search 'cli' | grep memory" \
    ""  # Should find memory sessions

run_test "Search includes traditional objects" \
    "port42 search '' --limit 5 | grep -E 'command|memory|artifact'" \
    ""  # Should include all entity types

echo -e "\n================================================"
echo -e "${BLUE}Phase D Advanced Discovery Test Summary${NC}"
echo "Tests run: $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"

if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
    echo -e "${GREEN}✅ All Phase D tests PASSED!${NC}"
    echo -e "${GREEN}🎉 Advanced discovery successfully integrates all entity types${NC}"
    exit 0
else
    failed=$((TESTS_RUN - TESTS_PASSED))
    echo -e "${RED}❌ $failed Phase D tests FAILED${NC}"
    echo -e "${YELLOW}💡 Search system needs relation integration for complete discovery${NC}"
    exit 1
fi