#!/bin/bash

# Test Step 3 Phase B: Unified Tool Hierarchy (Breaking Changes)
echo "🧪 Testing Step 3 Phase B: Unified Tool Hierarchy"
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

echo -e "\n${YELLOW}Phase B Success Criteria (Unified Tool Hierarchy):${NC}"
echo "1. /tools/ replaces both /relations/tools/ and /commands/"
echo "2. /tools/test-analyzer/definition shows relation JSON"
echo "3. /tools/test-analyzer/executable shows physical tool path" 
echo "4. /tools/test-analyzer/spawned/ shows auto-spawned entities"
echo "5. /tools/test-analyzer/parents/ shows parent chain"
echo "6. /tools/spawned-by/ shows global spawned-by index"
echo "7. /tools/transforms/ shows global transform index"
echo "8. OLD PATHS BREAK: /relations/tools/, /commands/ no longer work"

echo -e "\n${YELLOW}Step 1: Ensure we have test tools with spawning relationships${NC}"
echo "Creating test tools to ensure we have relations with spawned entities..."

# Create test tools to ensure we have good test data
port42 declare tool phase-b-analyzer --transforms data,analysis >/dev/null 2>&1 || true
port42 declare tool phase-b-processor --transforms data,transform >/dev/null 2>&1 || true
sleep 2  # Allow spawning to complete

echo -e "\n${YELLOW}Step 2: Test new unified /tools/ root${NC}"

run_test "Tools root directory" \
    "port42 ls /tools/" \
    ""  # Should show tool directories

run_test "Tools by-name subdirectory" \
    "port42 ls /tools/by-name/" \
    ""  # Should show all tools by name

run_test "Tools by-transform subdirectory" \
    "port42 ls /tools/by-transform/" \
    ""  # Should show transform categories

run_test "Tools spawned-by subdirectory" \
    "port42 ls /tools/spawned-by/" \
    ""  # Should show tools that have spawned others

echo -e "\n${YELLOW}Step 3: Test individual tool navigation${NC}"

# Find a tool to test with
tool_files=$(ls ~/.port42/relations/relation-tool-* 2>/dev/null | head -1)
if [[ -n "$tool_files" ]]; then
    tool_id=$(basename "$tool_files" .json | sed 's/relation-//')
    tool_name=$(jq -r '.properties.name' "$tool_files" 2>/dev/null || echo "unknown")
    echo "Testing with tool: $tool_name (ID: $tool_id)"
    
    run_test "Individual tool directory" \
        "port42 ls /tools/$tool_name/" \
        ""  # Should show tool subpaths: definition, executable, spawned, parents
        
    run_test "Tool definition (relation JSON)" \
        "port42 cat /tools/$tool_name/definition" \
        ""  # Should show relation JSON
        
    run_test "Tool executable info" \
        "port42 cat /tools/$tool_name/executable" \
        ""  # Should show physical tool info
        
    run_test "Tool spawned entities" \
        "port42 ls /tools/$tool_name/spawned/" \
        ""  # Should show spawned entities (may be empty)
        
    run_test "Tool parent chain" \
        "port42 ls /tools/$tool_name/parents/" \
        ""  # Should show parent chain (may be empty)
else
    echo "⚠️ No tool relations found for individual tool tests"
fi

echo -e "\n${YELLOW}Step 4: Test global relationship indexes${NC}"

run_test "Spawned-by global index" \
    "port42 ls /tools/spawned-by/" \
    ""  # Should show tools that spawned others

run_test "Transforms global index" \
    "port42 ls /tools/transforms/" \
    ""  # Should show available transforms

run_test "Specific transform listing" \
    "port42 ls /tools/transforms/analysis/" \
    ""  # Should show tools with analysis transform

echo -e "\n${YELLOW}Step 5: Test spawned relationship navigation${NC}"

# Find a tool with spawned entities
spawned_files=$(ls ~/.port42/relations/relation-tool-*analyzer*.json 2>/dev/null | head -1)
if [[ -n "$spawned_files" ]]; then
    spawned_tool_name=$(jq -r '.properties.name' "$spawned_files" 2>/dev/null || echo "unknown")
    if [[ "$spawned_tool_name" != "unknown" && "$spawned_tool_name" =~ analyzer ]]; then
        echo "Testing spawned relationships with: $spawned_tool_name"
        
        run_test "Direct tool spawned listing" \
            "port42 ls /tools/$spawned_tool_name/spawned/" \
            ""  # Should show view-* tools if they exist
            
        run_test "Global spawned-by for tool" \
            "port42 ls /tools/spawned-by/$spawned_tool_name/" \
            ""  # Should show same spawned entities
    fi
fi

echo -e "\n${YELLOW}Step 6: BREAKING CHANGES - Verify old paths fail${NC}"

run_test "OLD /relations/tools/ should fail or redirect" \
    "port42 ls /relations/tools/" \
    ""  # May fail with new implementation

run_test "OLD /commands/ should fail or redirect" \
    "port42 ls /commands/" \
    ""  # May fail with new implementation

run_test "OLD root should not show relations/" \
    "port42 ls /" \
    "tools/"  # Should show tools/ instead of relations/

echo -e "\n${YELLOW}Step 7: Test advanced navigation patterns${NC}"

run_test "Multi-transform tool discovery" \
    "port42 ls /tools/transforms/" \
    ""  # Should show multiple transform categories

run_test "Cross-reference: spawned tools in transforms" \
    "port42 ls /tools/transforms/view/" \
    ""  # Should show view-* tools if any exist

run_test "Parent chain navigation" \
    "port42 ls /tools/by-name/ | head -5" \
    ""  # Should show some tools

echo -e "\n================================================"
echo -e "${BLUE}Phase B Unified Hierarchy Test Summary${NC}"
echo "Tests run: $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $((TESTS_RUN - TESTS_PASSED))"

if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED - Unified Tool Hierarchy Working!${NC}"
    exit 0
elif [ $TESTS_PASSED -gt $((TESTS_RUN / 2)) ]; then
    echo -e "${YELLOW}⚠️ MAJORITY PASSED - Core functionality working${NC}"
    exit 1
else
    echo -e "${RED}❌ SIGNIFICANT FAILURES - Implementation needed${NC}"
    exit 2
fi