#!/bin/bash
# Test Step 5 Memory-Relation Bridge Gaps
# Verify the 3 identified gaps with concrete test code

set -e
cd /Users/gordon/Dropbox/Work/Hacking/workspace/port42

echo "🧪 Testing Step 5 Memory-Relation Bridge Implementation Gaps"
echo "============================================================"

# Test Setup
TEST_TOOL="gap-test-tool"
TIMESTAMP=$(date +%s)

echo
echo "🔍 GAP 1: Session Context in Tool Properties"
echo "--------------------------------------------"

# Create a test tool
echo "Creating test tool: $TEST_TOOL"
./cli/target/debug/port42 declare tool $TEST_TOOL --transforms test,gap,verification

# Check if tool properties contain session context
echo "Checking tool properties for session context..."

# Get tool relation data
TOOL_DATA=$(./cli/target/debug/port42 cat /tools/$TEST_TOOL/definition 2>/dev/null || echo "Tool not found")

# Test for session-related properties
echo "Looking for session context in tool properties:"
echo "$TOOL_DATA" | grep -E "(memory_session|created_by|crystallized_from|session_id)" && echo "✅ Session context FOUND" || echo "❌ Session context MISSING"

# Show what properties actually exist
echo
echo "Actual tool properties structure:"
echo "$TOOL_DATA" | jq '.properties | keys' 2>/dev/null || echo "Could not parse JSON"

echo
echo "🔍 GAP 2: Bidirectional Navigation Logic"
echo "---------------------------------------"

# Check if tool appears in memory system
echo "Looking for tool in memory filesystem..."

# Get tool ID from the relation
TOOL_ID=$(echo "$TOOL_DATA" | jq -r '.id' 2>/dev/null)
echo "Tool ID: $TOOL_ID"

# Check if memory path exists for this tool
echo "Checking /memory/$TOOL_ID/ path:"
MEMORY_PATH_EXISTS=$(./cli/target/debug/port42 ls /memory/$TOOL_ID/ 2>/dev/null && echo "EXISTS" || echo "MISSING")
echo "Memory path status: $MEMORY_PATH_EXISTS"

if [ "$MEMORY_PATH_EXISTS" = "EXISTS" ]; then
    echo "Checking /memory/$TOOL_ID/generated/ contents:"
    GENERATED_CONTENTS=$(./cli/target/debug/port42 ls /memory/$TOOL_ID/generated/ 2>/dev/null)
    if [ -z "$GENERATED_CONTENTS" ] || echo "$GENERATED_CONTENTS" | grep -q "(empty)"; then
        echo "❌ Generated folder is EMPTY - bidirectional link missing"
    else
        echo "✅ Generated folder has contents: $GENERATED_CONTENTS"
    fi
else
    echo "❌ No memory path exists for tool"
fi

echo
echo "🔍 GAP 3: Session ID Capture During Tool Creation"
echo "------------------------------------------------"

# Test if CLI tool creation captures session context
echo "Testing session ID capture mechanism..."

# Check if there's any session tracking in the daemon
echo "Looking for session tracking in handleDeclareRelation..."

# Try to find evidence of session capture in tool properties
echo "Checking if newly created tool has any session identifiers:"
echo "$TOOL_DATA" | jq '.properties' | grep -E "(session|cli|conversation)" && echo "✅ Session capture FOUND" || echo "❌ Session capture MISSING"

echo
echo "Checking for getCurrentSessionID implementation..."
# Look for session detection in the daemon logs or behavior
echo "Testing if tool creation context includes session information..."

# Create another tool and see if it gets linked to the same or different session
TEST_TOOL_2="${TEST_TOOL}-2"
echo "Creating second tool to test session linkage: $TEST_TOOL_2"
./cli/target/debug/port42 declare tool $TEST_TOOL_2 --transforms test,gap,verification

TOOL_2_DATA=$(./cli/target/debug/port42 cat /tools/$TEST_TOOL_2/definition 2>/dev/null)
TOOL_2_ID=$(echo "$TOOL_2_DATA" | jq -r '.id' 2>/dev/null)

echo "Comparing session context between tools:"
echo "Tool 1 created_at: $(echo "$TOOL_DATA" | jq -r '.created_at')"
echo "Tool 2 created_at: $(echo "$TOOL_2_DATA" | jq -r '.created_at')"

# Check if they have any session linkage
echo "Checking if tools created in same session have common session identifier..."
TOOL_1_SESSION=$(echo "$TOOL_DATA" | jq -r '.properties.session_id // "none"')
TOOL_2_SESSION=$(echo "$TOOL_2_DATA" | jq -r '.properties.session_id // "none"')

echo "Tool 1 session ID: $TOOL_1_SESSION"
echo "Tool 2 session ID: $TOOL_2_SESSION"

if [ "$TOOL_1_SESSION" = "none" ] && [ "$TOOL_2_SESSION" = "none" ]; then
    echo "❌ Neither tool has session ID - session capture NOT implemented"
elif [ "$TOOL_1_SESSION" = "$TOOL_2_SESSION" ] && [ "$TOOL_1_SESSION" != "none" ]; then
    echo "✅ Tools share session ID - session capture IS implemented"
else
    echo "⚠️ Tools have different session IDs - partial implementation"
fi

echo
echo "🔍 SUMMARY: Step 5 Memory-Relation Bridge Gaps"
echo "==============================================" 

echo "Gap 1 - Session Context in Tool Properties:"
if echo "$TOOL_DATA" | grep -q -E "(memory_session|created_by|crystallized_from)"; then
    echo "  ✅ IMPLEMENTED - Tools contain session context"
else
    echo "  ❌ MISSING - Tools lack session context properties"
fi

echo "Gap 2 - Bidirectional Navigation Logic:"
if [ "$MEMORY_PATH_EXISTS" = "EXISTS" ] && [ -n "$GENERATED_CONTENTS" ] && ! echo "$GENERATED_CONTENTS" | grep -q "(empty)"; then
    echo "  ✅ IMPLEMENTED - Memory paths populated with generated tools"
else
    echo "  ❌ INCOMPLETE - Memory paths exist but not populated"
fi

echo "Gap 3 - Session ID Capture:"
if [ "$TOOL_1_SESSION" != "none" ] || [ "$TOOL_2_SESSION" != "none" ]; then
    echo "  ✅ IMPLEMENTED - Tools capture session context during creation"
else
    echo "  ❌ MISSING - No session ID capture during tool creation"
fi

echo
echo "🧪 Test completed. Check output above for detailed gap analysis."

# Cleanup
echo
echo "🧹 Cleaning up test tools..."
# Note: No cleanup command exists, tools remain in system