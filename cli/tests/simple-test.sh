#!/bin/bash

# Simplified test to isolate the hanging issue

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test variables
TOTAL_TESTS=0
PASSED_TESTS=0

# Directories
TEST_DIR="$(pwd)"
TEST_DATA_DIR="$TEST_DIR/test-data"
ROOT_DIR="$(dirname "$(dirname "$TEST_DIR")")"
PORT42_BIN="$ROOT_DIR/bin/port42"

echo -e "${YELLOW}🔍 Simple Test - File Creation${NC}"

# Test 1: Check test-config.json
echo "Checking test-config.json..."
if [[ ! -f "$TEST_DATA_DIR/test-config.json" ]]; then
    echo "File does not exist"
else
    echo -e "   ${GREEN}✅ File exists${NC}"
    ((PASSED_TESTS++))
fi

# Test 2: Check sample-data.csv  
echo "Checking sample-data.csv..."
if [[ ! -f "$TEST_DATA_DIR/sample-data.csv" ]]; then
    echo "File does not exist"
else
    echo -e "   ${GREEN}✅ File exists${NC}"
    ((PASSED_TESTS++))
fi

# Test 3: Basic daemon check
echo "Testing basic daemon status..."
if "$PORT42_BIN" status >/dev/null 2>&1; then
    echo -e "   ${GREEN}✅ Daemon running${NC}"
    ((PASSED_TESTS++))
else
    echo "❌ Daemon not running"
fi

echo ""
echo "Completed: $PASSED_TESTS tests passed"