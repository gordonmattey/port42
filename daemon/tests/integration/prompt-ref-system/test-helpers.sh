#!/bin/bash
# tests/integration/prompt-ref-system/test-helpers.sh
# Common utilities for prompt and reference system tests

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test utilities
assert_success() {
    local command="$1"
    local description="$2"
    
    echo -e "  ${BLUE}Testing: $description${NC}"
    if eval "$command" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✅ SUCCESS: $description${NC}"
        return 0
    else
        echo -e "  ${RED}❌ FAILED: $description${NC}"
        echo -e "  ${RED}Command: $command${NC}"
        return 1
    fi
}

assert_failure() {
    local command="$1"
    local description="$2"
    
    echo -e "  ${BLUE}Testing: $description${NC}"
    if ! eval "$command" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✅ SUCCESS: $description (correctly failed)${NC}"
        return 0
    else
        echo -e "  ${RED}❌ FAILED: $description (should have failed)${NC}"
        echo -e "  ${RED}Command: $command${NC}"
        return 1
    fi
}

assert_contains() {
    local content="$1"
    local pattern="$2"
    local description="$3"
    
    echo -e "  ${BLUE}Testing: $description${NC}"
    if echo "$content" | grep -q "$pattern"; then
        echo -e "  ${GREEN}✅ SUCCESS: Found '$pattern' in content${NC}"
        return 0
    else
        echo -e "  ${RED}❌ FAILED: Pattern '$pattern' not found${NC}"
        echo -e "  ${RED}Content: $content${NC}"
        return 1
    fi
}

assert_file_exists() {
    local file_path="$1"
    local description="$2"
    
    echo -e "  ${BLUE}Testing: $description${NC}"
    if [[ -f "$file_path" ]]; then
        echo -e "  ${GREEN}✅ SUCCESS: File exists: $file_path${NC}"
        return 0
    else
        echo -e "  ${RED}❌ FAILED: File does not exist: $file_path${NC}"
        return 1
    fi
}

cleanup_test_tools() {
    echo -e "  ${YELLOW}🧹 Cleaning up test tools...${NC}"
    
    # Remove test tools (be careful not to remove real tools)
    local test_patterns=("test-*" "*-test" "prompt-*" "ref-*" "perf-test-*")
    
    for pattern in "${test_patterns[@]}"; do
        if ls ~/.port42/commands/$pattern 1> /dev/null 2>&1; then
            rm -f ~/.port42/commands/$pattern
        fi
    done
    
    # Clean up test artifacts
    if ls ~/.port42/artifacts/*test* 1> /dev/null 2>&1; then
        rm -f ~/.port42/artifacts/*test*
    fi
    
    echo -e "  ${GREEN}✅ Cleanup complete${NC}"
}

wait_for_daemon() {
    echo -e "  ${BLUE}Waiting for daemon to be ready...${NC}"
    
    local max_attempts=10
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if port42 status >/dev/null 2>&1; then
            echo -e "  ${GREEN}✅ Daemon is ready${NC}"
            return 0
        fi
        
        echo -e "  ${YELLOW}Attempt $attempt/$max_attempts - waiting...${NC}"
        sleep 2
        ((attempt++))
    done
    
    echo -e "  ${RED}❌ Daemon not ready after $max_attempts attempts${NC}"
    return 1
}

setup_test_environment() {
    echo -e "${BLUE}🔧 Setting up test environment${NC}"
    
    # Ensure daemon is running
    wait_for_daemon
    
    # Clean up any existing test data
    cleanup_test_tools
    
    echo -e "${GREEN}✅ Test environment ready${NC}"
}

teardown_test_environment() {
    echo -e "${BLUE}🧹 Tearing down test environment${NC}"
    
    # Clean up test data
    cleanup_test_tools
    
    echo -e "${GREEN}✅ Test environment cleaned${NC}"
}

# Export functions for use in test scripts
export -f assert_success
export -f assert_failure  
export -f assert_contains
export -f assert_file_exists
export -f cleanup_test_tools
export -f wait_for_daemon
export -f setup_test_environment
export -f teardown_test_environment