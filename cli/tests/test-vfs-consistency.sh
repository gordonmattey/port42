#!/bin/bash

# VFS Consistency Test Suite
# Tests the tool materialization and VFS consistency issue

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0

# Directories
ROOT_DIR="$(dirname "$(dirname "$(pwd)")")"
PORT42_BIN="$ROOT_DIR/bin/port42"

echo -e "${YELLOW}🔍 VFS Consistency Test Suite${NC}"
echo "Testing tool materialization and VFS consistency"
echo ""

# Helper functions
pass_test() {
    echo -e "   ${GREEN}✅ $1${NC}"
    ((PASSED_TESTS++))
}

fail_test() {
    echo -e "   ${RED}❌ $1${NC}"
}

run_test() {
    echo "Test $((++TOTAL_TESTS)): $1"
}

# Test 1: Create a fresh tool and capture materialization process
run_test "Tool Creation and Object Tracking"

TOOL_NAME="vfs-test-$(date +%s)"
echo "Creating tool: $TOOL_NAME"

# Capture object store state before
OBJECTS_BEFORE=$(find ~/.port42/objects -type f | wc -l)

# Create tool with debug enabled
PORT42_DEBUG=1 "$PORT42_BIN" declare tool "$TOOL_NAME" --transforms "test,consistency" > /tmp/tool_creation.log 2>&1 &
TOOL_PID=$!

# Wait for completion or timeout
timeout 120 wait $TOOL_PID || {
    echo "Tool creation timed out"
    kill $TOOL_PID 2>/dev/null || true
    fail_test "Tool creation completed within timeout"
    exit 1
}

if [[ -f ~/.port42/commands/$TOOL_NAME ]]; then
    pass_test "Tool materialized on filesystem"
else
    fail_test "Tool materialized on filesystem"
    exit 1
fi

# Test 2: Compare VFS vs Filesystem content
run_test "VFS vs Filesystem Content Consistency"

# Get VFS object ID
VFS_OBJECT_ID=$("$PORT42_BIN" info "/commands/$TOOL_NAME" | grep "Object ID" | awk '{print $3}' | sed 's/\x1b\[[0-9;]*m//g')
echo "VFS Object ID: $VFS_OBJECT_ID"

# Get filesystem symlink target
if [[ -L ~/.port42/commands/$TOOL_NAME ]]; then
    SYMLINK_TARGET=$(readlink ~/.port42/commands/$TOOL_NAME)
    # Extract object ID from path: /path/to/objects/ab/cd/efgh -> abcdefgh
    FS_OBJECT_ID=$(echo "$SYMLINK_TARGET" | sed 's|.*/objects/||' | tr -d '/')
    echo "Filesystem Object ID: $FS_OBJECT_ID"
    
    if [[ "$VFS_OBJECT_ID" == "$FS_OBJECT_ID" ]]; then
        pass_test "VFS and filesystem point to same object"
    else
        fail_test "VFS and filesystem point to different objects"
        echo "  VFS:        $VFS_OBJECT_ID"
        echo "  Filesystem: $FS_OBJECT_ID"
    fi
else
    fail_test "Filesystem command is not a symlink"
fi

# Test 3: Content comparison
run_test "Content Size and Quality Check"

# Get VFS content
VFS_CONTENT=$("$PORT42_BIN" cat "/commands/$TOOL_NAME")
VFS_SIZE=$(echo "$VFS_CONTENT" | wc -c)

# Get filesystem content  
FS_CONTENT=$(cat ~/.port42/commands/$TOOL_NAME)
FS_SIZE=$(echo "$FS_CONTENT" | wc -c)

echo "VFS Content Size: $VFS_SIZE bytes"
echo "Filesystem Content Size: $FS_SIZE bytes"

if [[ $VFS_SIZE -eq $FS_SIZE ]]; then
    pass_test "Content sizes match"
else
    fail_test "Content sizes differ (VFS: $VFS_SIZE, FS: $FS_SIZE)"
fi

# Check if content is stub vs full implementation
if echo "$VFS_CONTENT" | grep -q "TODO: Implement actual tool logic"; then
    fail_test "VFS content is stub implementation"
    echo "  VFS shows stub content"
else
    pass_test "VFS content is full implementation"
fi

if echo "$FS_CONTENT" | grep -q "TODO: Implement actual tool logic"; then
    fail_test "Filesystem content is stub implementation"
else
    pass_test "Filesystem content is full implementation"
fi

# Test 4: Object Store Analysis
run_test "Object Store Integrity Check"

OBJECTS_AFTER=$(find ~/.port42/objects -type f | wc -l)
NEW_OBJECTS=$((OBJECTS_AFTER - OBJECTS_BEFORE))

echo "New objects created: $NEW_OBJECTS"

# Check if both object IDs exist in store
if [[ -n "$VFS_OBJECT_ID" ]] && [[ -f ~/.port42/objects/${VFS_OBJECT_ID:0:2}/${VFS_OBJECT_ID:2:2}/${VFS_OBJECT_ID:4} ]]; then
    pass_test "VFS object exists in store"
    VFS_OBJECT_SIZE=$(stat -f%z ~/.port42/objects/${VFS_OBJECT_ID:0:2}/${VFS_OBJECT_ID:2:2}/${VFS_OBJECT_ID:4})
    echo "  VFS object size: $VFS_OBJECT_SIZE bytes"
else
    fail_test "VFS object missing from store"
fi

if [[ -n "$FS_OBJECT_ID" ]] && [[ -f ~/.port42/objects/${FS_OBJECT_ID:0:2}/${FS_OBJECT_ID:2:2}/${FS_OBJECT_ID:4} ]]; then
    pass_test "Filesystem object exists in store"
    FS_OBJECT_SIZE=$(stat -f%z ~/.port42/objects/${FS_OBJECT_ID:0:2}/${FS_OBJECT_ID:2:2}/${FS_OBJECT_ID:4})
    echo "  Filesystem object size: $FS_OBJECT_SIZE bytes"
else
    fail_test "Filesystem object missing from store"
fi

# Test 5: Relation Analysis
run_test "Relation Store Analysis"

RELATION_CONTENT=$("$PORT42_BIN" cat "/tools/$TOOL_NAME" 2>/dev/null || echo "No relation found")
if echo "$RELATION_CONTENT" | grep -q '"executable"'; then
    fail_test "Relation still contains executable content (should only have executable_id)"
    echo "  Relations should not store executable content directly"
else
    pass_test "Relation does not contain duplicate executable content"
fi

if echo "$RELATION_CONTENT" | grep -q '"executable_id"'; then
    pass_test "Relation contains executable_id reference"
    RELATION_EXEC_ID=$(echo "$RELATION_CONTENT" | grep '"executable_id"' | sed 's/.*"executable_id": *"\([^"]*\)".*/\1/')
    echo "  Relation executable_id: $RELATION_EXEC_ID"
    
    if [[ "$RELATION_EXEC_ID" == "$FS_OBJECT_ID" ]]; then
        pass_test "Relation executable_id matches filesystem object"
    else
        fail_test "Relation executable_id does not match filesystem object"
        echo "    Relation: $RELATION_EXEC_ID"
        echo "    Filesystem: $FS_OBJECT_ID"
    fi
else
    fail_test "Relation missing executable_id reference"
fi

# Test 6: Materialization Log Analysis
run_test "Materialization Process Analysis"

if [[ -f /tmp/tool_creation.log ]]; then
    if grep -q "✅ Updated relation.*with executable_id" /tmp/tool_creation.log; then
        pass_test "Materialization updated relation with executable_id"
    else
        fail_test "Materialization did not update relation with executable_id"
    fi
    
    # Check for timing issues
    STORE_CALLS=$(grep -c "Store called" /tmp/tool_creation.log || echo "0")
    echo "  Total Store() calls during materialization: $STORE_CALLS"
    
    if [[ $STORE_CALLS -gt 2 ]]; then
        fail_test "Excessive Store() calls suggest duplicate storage ($STORE_CALLS calls)"
    else
        pass_test "Reasonable number of Store() calls ($STORE_CALLS)"
    fi
else
    fail_test "Materialization log not found"
fi

# Summary
echo ""
echo -e "${YELLOW}🏁 Test Summary${NC}"
echo "Passed: $PASSED_TESTS/$TOTAL_TESTS tests"

if [[ $PASSED_TESTS -eq $TOTAL_TESTS ]]; then
    echo -e "${GREEN}✅ All tests passed - VFS consistency working correctly${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed - VFS consistency issues detected${NC}"
    echo ""
    echo "Root Cause Analysis:"
    if [[ "$VFS_OBJECT_ID" != "$FS_OBJECT_ID" ]]; then
        echo "- VFS and filesystem point to different objects (CRITICAL)"
    fi
    if echo "$VFS_CONTENT" | grep -q "TODO: Implement actual tool logic"; then
        echo "- VFS serves stub content instead of full implementation (CRITICAL)"
    fi
    exit 1
fi