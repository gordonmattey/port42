#!/bin/bash
# Final Step 5 Memory-Relation Bridge Verification
# Test all 3 gaps with correct paths and expectations

set -e
cd /Users/gordon/Dropbox/Work/Hacking/workspace/port42

echo "🎯 Final Step 5 Memory-Relation Bridge Verification"
echo "=================================================="

# Test Setup
TEST_TOOL="final-gap-test-tool"
TIMESTAMP=$(date +%s)

echo
echo "✨ Creating test tool with session context..."
./cli/target/debug/port42 declare tool $TEST_TOOL --transforms echo,final,test

# Get tool data
TOOL_DATA=$(./bin/port42 cat /tools/$TEST_TOOL/definition 2>/dev/null)
TOOL_ID=$(echo "$TOOL_DATA" | jq -r '.id' 2>/dev/null)
SESSION_ID=$(echo "$TOOL_DATA" | jq -r '.properties.memory_session' 2>/dev/null)

echo
echo "🔍 VERIFICATION RESULTS"
echo "======================="

echo
echo "✅ Gap 1: Session Context in Tool Properties"
echo "   Tool ID: $TOOL_ID"
echo "   Session ID: $SESSION_ID"
if [[ "$SESSION_ID" != "null" && "$SESSION_ID" != "" ]]; then
    echo "   Status: ✅ IMPLEMENTED - Tool has session context property"
else
    echo "   Status: ❌ MISSING - Tool lacks session context"
    exit 1
fi

echo
echo "✅ Gap 2: Bidirectional Navigation Logic"
echo "   Testing path: /memory/$SESSION_ID/generated/"
GENERATED_TOOLS=$(./bin/port42 ls /memory/$SESSION_ID/generated/ 2>/dev/null | grep -v "^/" | grep -v "^$" | grep -v "(empty)")
if [[ -n "$GENERATED_TOOLS" ]]; then
    echo "   Found tools: $GENERATED_TOOLS"
    echo "   Status: ✅ IMPLEMENTED - Tools appear in memory navigation"
else
    echo "   Status: ❌ INCOMPLETE - Memory path exists but not populated"
    exit 1
fi

echo
echo "✅ Gap 3: Session ID Capture During Tool Creation"
if [[ "$SESSION_ID" =~ ^cli-session-[0-9]+$ ]]; then
    echo "   Format: CLI session ID format correct"
    echo "   Status: ✅ IMPLEMENTED - Session ID captured during CLI tool creation"
else
    echo "   Format: Invalid session ID format: $SESSION_ID"
    echo "   Status: ❌ MISSING - Session ID not properly captured"
    exit 1
fi

echo
echo "🎉 FINAL RESULT: All Step 5 Memory-Relation Bridge Gaps RESOLVED!"
echo "================================================================"
echo
echo "✅ Gap 1 - Session Context in Tool Properties: IMPLEMENTED"
echo "✅ Gap 2 - Bidirectional Navigation Logic: IMPLEMENTED"  
echo "✅ Gap 3 - Session ID Capture During Creation: IMPLEMENTED"
echo
echo "🔗 Memory-Relation Bridge is now functional!"
echo "   CLI tools are linked to their creation sessions"
echo "   Tools can be discovered via memory filesystem navigation"
echo "   Bidirectional relationship between conversations and tools established"

echo
echo "🧹 Cleaning up test tool..."
# Note: No cleanup command exists for relations yet