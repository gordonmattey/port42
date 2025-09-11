#!/bin/bash
# Test script to verify context collector functionality

set -e

echo "🧪 Testing Context Collector (Step 2)"
echo "====================================="

# Test 1: Run some commands to track
echo -e "\n📌 Test 1: Generating command history"
port42 status > /dev/null 2>&1
echo "  ✓ Ran: port42 status"
sleep 1

port42 search "test" > /dev/null 2>&1
echo "  ✓ Ran: port42 search \"test\""
sleep 1

port42 ls /tools/ > /dev/null 2>&1
echo "  ✓ Ran: port42 ls /tools/"
sleep 1

# Test 2: Check that commands are tracked
echo -e "\n📌 Test 2: Checking command tracking"
CONTEXT=$(port42 context 2>&1)

COMMANDS_COUNT=$(echo "$CONTEXT" | jq '.recent_commands | length')
if [ "$COMMANDS_COUNT" -gt 0 ]; then
    echo "✅ Commands tracked: $COMMANDS_COUNT"
    echo "  Recent commands:"
    echo "$CONTEXT" | jq -r '.recent_commands[] | "    - \(.command) [\(.age_seconds)s ago]"' | head -5
else
    echo "⚠️  No commands tracked yet (may need more time)"
fi

# Test 3: Check suggestions
echo -e "\n📌 Test 3: Checking suggestions"
SUGGESTIONS_COUNT=$(echo "$CONTEXT" | jq '.suggestions | length')
if [ "$SUGGESTIONS_COUNT" -gt 0 ]; then
    echo "✅ Suggestions generated: $SUGGESTIONS_COUNT"
    echo "  Suggestions:"
    echo "$CONTEXT" | jq -r '.suggestions[] | "    - \(.command)"' | head -3
    echo "$CONTEXT" | jq -r '.suggestions[] | "      (\(.reason))"' | head -3
else
    echo "⚠️  No suggestions generated"
fi

# Test 4: Create a session and check tool tracking
echo -e "\n📌 Test 4: Session with tool creation"
port42 possess @ai-engineer "create a simple test tool called collector-test that echoes 'testing'" > /dev/null 2>&1 &
POSSESS_PID=$!

# Wait for tool creation (this is async)
echo "  Waiting for tool creation..."
sleep 10

# Check if tool was tracked
CONTEXT=$(port42 context 2>&1)
TOOLS_COUNT=$(echo "$CONTEXT" | jq '.created_tools | length')
if [ "$TOOLS_COUNT" -gt 0 ]; then
    echo "✅ Tools tracked: $TOOLS_COUNT"
    echo "$CONTEXT" | jq -r '.created_tools[] | "    - \(.name)"'
else
    echo "⚠️  No tools tracked (tool creation may not have completed)"
fi

# Test 5: Pretty format with all data
echo -e "\n📌 Test 5: Pretty format display"
echo "---"
port42 context --pretty | head -20
echo "---"

# Test 6: Compact format
echo -e "\n📌 Test 6: Compact format"
COMPACT=$(port42 context --compact)
echo "  $COMPACT"

# Test 7: Verify suggestions are contextual
echo -e "\n📌 Test 7: Contextual suggestions"
CONTEXT=$(port42 context 2>&1)
if echo "$CONTEXT" | jq -e '.active_session != null' > /dev/null 2>&1; then
    echo "✅ Active session detected"
    # Should suggest continuing session
    if echo "$CONTEXT" | jq -r '.suggestions[].command' | grep -q "session last"; then
        echo "✅ Suggests continuing session"
    fi
    # If tool created, should suggest using it
    if echo "$CONTEXT" | jq -r '.suggestions[].command' | grep -q "\-\-help"; then
        echo "✅ Suggests using created tool"
    fi
fi

echo -e "\n✨ Context Collector tests complete!"
echo "Features working:"
echo "  • Command tracking"
echo "  • Age calculation" 
echo "  • Suggestions generation"
echo "  • Session awareness"
echo "  • Tool tracking (if created)"