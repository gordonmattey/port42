#!/bin/bash

# Test Step 2 Phase 2: Auto-spawning ViewerRule
# This script validates that analysis tools auto-spawn viewer tools

set -e

echo "🧪 Testing Step 2 Phase 2: Auto-spawning ViewerRule"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test tracking
TESTS_PASSED=0
TESTS_FAILED=0
ERRORS_FOUND=()

# Helper functions
log_test() {
    echo -e "${BLUE}🔍 TEST: $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
    ((TESTS_PASSED++))
}

log_failure() {
    echo -e "${RED}❌ $1${NC}"
    ((TESTS_FAILED++))
    ERRORS_FOUND+=("$1")
}

log_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Test 1: Analysis tool should spawn viewer
test_analysis_tool_spawning() {
    log_test "Analysis tool should auto-spawn viewer tool"
    
    # Clean up any existing tools
    rm -f ~/.port42/commands/log-analyzer ~/.port42/commands/view-log-analyzer 2>/dev/null || true
    
    # Declare analysis tool
    log_info "Declaring analysis tool..."
    OUTPUT=$(./cli/target/release/port42 declare tool log-analyzer --transforms logs,analysis 2>&1)
    
    if [[ $? -ne 0 ]]; then
        log_failure "Failed to declare analysis tool: $OUTPUT"
        return 1
    fi
    
    # Check that main tool was created
    if [[ ! -f ~/.port42/commands/log-analyzer ]]; then
        log_failure "Main tool log-analyzer was not created"
        return 1
    fi
    
    # Check that viewer tool was auto-spawned
    if [[ ! -f ~/.port42/commands/view-log-analyzer ]]; then
        log_failure "Viewer tool view-log-analyzer was not auto-spawned"
        return 1
    fi
    
    # Test that both tools are executable
    if [[ ! -x ~/.port42/commands/log-analyzer ]]; then
        log_failure "Main tool log-analyzer is not executable"
        return 1
    fi
    
    if [[ ! -x ~/.port42/commands/view-log-analyzer ]]; then
        log_failure "Viewer tool view-log-analyzer is not executable"
        return 1
    fi
    
    log_success "Analysis tool successfully spawned viewer tool"
    return 0
}

# Test 2: Non-analysis tool should NOT spawn viewer
test_non_analysis_no_spawning() {
    log_test "Non-analysis tool should NOT spawn viewer"
    
    # Clean up any existing tools
    rm -f ~/.port42/commands/simple-parser ~/.port42/commands/view-simple-parser 2>/dev/null || true
    
    # Declare non-analysis tool
    log_info "Declaring non-analysis tool..."
    OUTPUT=$(./cli/target/release/port42 declare tool simple-parser --transforms parse,clean 2>&1)
    
    if [[ $? -ne 0 ]]; then
        log_failure "Failed to declare non-analysis tool: $OUTPUT"
        return 1
    fi
    
    # Check that main tool was created
    if [[ ! -f ~/.port42/commands/simple-parser ]]; then
        log_failure "Main tool simple-parser was not created"
        return 1
    fi
    
    # Check that viewer tool was NOT created
    if [[ -f ~/.port42/commands/view-simple-parser ]]; then
        log_failure "Viewer tool was incorrectly spawned for non-analysis tool"
        return 1
    fi
    
    log_success "Non-analysis tool correctly did not spawn viewer"
    return 0
}

# Test 3: Both tools should function independently
test_tool_functionality() {
    log_test "Both spawned tools should function independently"
    
    # Test main tool
    log_info "Testing main tool functionality..."
    MAIN_OUTPUT=$(log-analyzer --help 2>&1)
    if [[ $? -ne 0 ]]; then
        log_failure "Main tool log-analyzer --help failed: $MAIN_OUTPUT"
        return 1
    fi
    
    # Test viewer tool
    log_info "Testing viewer tool functionality..."
    VIEWER_OUTPUT=$(view-log-analyzer --help 2>&1)
    if [[ $? -ne 0 ]]; then
        log_failure "Viewer tool view-log-analyzer --help failed: $VIEWER_OUTPUT"
        return 1
    fi
    
    log_success "Both main and viewer tools function correctly"
    return 0
}

# Test 4: Relationship metadata should be tracked
test_relationship_tracking() {
    log_test "Spawning relationships should be tracked in metadata"
    
    # Get the relation ID for log-analyzer
    RELATION_FILES=(~/.port42/relations/relation-tool-log-analyzer-*.json)
    if [[ ! -f "${RELATION_FILES[0]}" ]]; then
        log_failure "Could not find relation file for log-analyzer"
        return 1
    fi
    
    MAIN_RELATION_FILE="${RELATION_FILES[0]}"
    log_info "Checking main relation file: $MAIN_RELATION_FILE"
    
    # Check for viewer relation file
    VIEWER_RELATION_FILES=(~/.port42/relations/relation-tool-view-log-analyzer-*.json)
    if [[ ! -f "${VIEWER_RELATION_FILES[0]}" ]]; then
        log_failure "Could not find relation file for view-log-analyzer"
        return 1
    fi
    
    VIEWER_RELATION_FILE="${VIEWER_RELATION_FILES[0]}"
    log_info "Checking viewer relation file: $VIEWER_RELATION_FILE"
    
    # Check that viewer relation has parent metadata
    if ! grep -q "parent" "$VIEWER_RELATION_FILE"; then
        log_failure "Viewer relation missing parent metadata"
        return 1
    fi
    
    if ! grep -q "spawned_by" "$VIEWER_RELATION_FILE"; then
        log_failure "Viewer relation missing spawned_by metadata"
        return 1
    fi
    
    if ! grep -q "auto_spawned" "$VIEWER_RELATION_FILE"; then
        log_failure "Viewer relation missing auto_spawned metadata"
        return 1
    fi
    
    log_success "Relationship metadata correctly tracked"
    return 0
}

# Test 5: Rule engine logs should show activity
test_rule_engine_logging() {
    log_test "Rule engine should log spawning activity"
    
    # This test would require access to daemon logs
    # For now, we'll test that the system doesn't crash
    log_info "Testing rule engine doesn't crash system..."
    
    # Declare another analysis tool to trigger rules again
    OUTPUT=$(./cli/target/release/port42 declare tool data-analyzer --transforms data,analysis 2>&1)
    
    if [[ $? -ne 0 ]]; then
        log_failure "Rule engine caused system failure: $OUTPUT"
        return 1
    fi
    
    # Check tools were created
    if [[ ! -f ~/.port42/commands/data-analyzer || ! -f ~/.port42/commands/view-data-analyzer ]]; then
        log_failure "Second analysis tool did not spawn correctly"
        return 1
    fi
    
    log_success "Rule engine continues to work for multiple tools"
    return 0
}

# Root cause analysis for common issues
analyze_root_causes() {
    if [[ ${#ERRORS_FOUND[@]} -eq 0 ]]; then
        return 0
    fi
    
    echo ""
    echo -e "${YELLOW}🔍 ROOT CAUSE ANALYSIS${NC}"
    echo "======================="
    
    for error in "${ERRORS_FOUND[@]}"; do
        echo -e "${YELLOW}Error: $error${NC}"
        
        # Analyze potential root causes
        case "$error" in
            *"was not auto-spawned"*)
                echo "  🔍 Potential causes:"
                echo "     - Rule condition not matching (check transforms logic)"
                echo "     - Rule engine not attached to RealityCompiler"
                echo "     - Rule action failing silently"
                echo "     - ViewerRule not properly added to defaultRules()"
                ;;
            *"not executable"*)
                echo "  🔍 Potential causes:"
                echo "     - File permissions not set correctly in Materialize()"
                echo "     - Object store symlink issue"
                echo "     - Tool materializer not calling storage system properly"
                ;;
            *"missing parent metadata"*)
                echo "  🔍 Potential causes:"
                echo "     - ViewerRule not setting relationship properties"
                echo "     - Relation properties not being serialized"
                echo "     - File storage not preserving metadata"
                ;;
            *"Rule engine caused system failure"*)
                echo "  🔍 Potential causes:"
                echo "     - Infinite recursion in rule processing"
                echo "     - Rule action creating malformed relations"
                echo "     - Error handling not preventing cascade failures"
                ;;
        esac
        echo ""
    done
}

# Main test execution
main() {
    echo "Starting Phase 2 tests..."
    echo ""
    
    # Ensure daemon is running (assumption - user should start it)
    if ! pgrep -f "port42d" > /dev/null; then
        echo -e "${RED}❌ Daemon not running. Please start with: ./daemon/port42d${NC}"
        exit 1
    fi
    
    # Run all tests
    test_analysis_tool_spawning
    test_non_analysis_no_spawning
    test_tool_functionality
    test_relationship_tracking
    test_rule_engine_logging
    
    # Summary
    echo ""
    echo "=========================================="
    echo -e "${GREEN}✅ Tests passed: $TESTS_PASSED${NC}"
    echo -e "${RED}❌ Tests failed: $TESTS_FAILED${NC}"
    
    if [[ $TESTS_FAILED -gt 0 ]]; then
        analyze_root_causes
        echo ""
        echo -e "${RED}❌ Phase 2 tests FAILED${NC}"
        exit 1
    else
        echo ""
        echo -e "${GREEN}🎉 Phase 2 tests PASSED - ViewerRule working correctly!${NC}"
        exit 0
    fi
}

main "$@"