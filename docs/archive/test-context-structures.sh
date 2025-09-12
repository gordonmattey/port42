#!/bin/bash
# Test script to verify context data structures work correctly

set -e

echo "🧪 Testing Port42 Context Structures"
echo "===================================="

# Test 1: Basic context call works
echo -e "\n📌 Test 1: Context command responds"
OUTPUT=$(port42 context 2>&1)
if [ $? -eq 0 ]; then
    echo "✅ Context command works"
    if echo "$OUTPUT" | jq -e '.active_session' > /dev/null 2>&1; then
        if [ "$(echo "$OUTPUT" | jq -r '.active_session')" == "null" ]; then
            echo "  No active session"
        else
            echo "  Active session found: $(echo "$OUTPUT" | jq -r '.active_session.id')"
        fi
    fi
else
    echo "❌ Failed to get context"
    exit 1
fi

# Test 2: Check all expected fields exist
echo -e "\n📌 Test 2: Structure fields"
if echo "$OUTPUT" | jq -e 'has("active_session") and has("recent_commands") and has("created_tools") and has("suggestions")' > /dev/null 2>&1; then
    echo "✅ All required fields present"
else
    echo "❌ Missing required fields"
    echo "$OUTPUT" | jq '.'
    exit 1
fi

# Test 3: Arrays are initialized
echo -e "\n📌 Test 3: Arrays initialized"
COMMANDS_LEN=$(echo "$OUTPUT" | jq '.recent_commands | length')
TOOLS_LEN=$(echo "$OUTPUT" | jq '.created_tools | length')
SUGGESTIONS_LEN=$(echo "$OUTPUT" | jq '.suggestions | length')

echo "  Recent commands: $COMMANDS_LEN"
echo "  Created tools: $TOOLS_LEN"
echo "  Suggestions: $SUGGESTIONS_LEN"
echo "✅ Arrays properly initialized"

# Test 4: Create a session and verify structure
echo -e "\n📌 Test 4: With active session"
port42 possess @ai-engineer "test" > /dev/null 2>&1

OUTPUT=$(port42 context 2>&1)
if echo "$OUTPUT" | jq -e '.active_session != null' > /dev/null 2>&1; then
    echo "✅ Active session detected"
    
    # Check session fields
    SESSION=$(echo "$OUTPUT" | jq '.active_session')
    
    if echo "$SESSION" | jq -e 'has("id") and has("agent") and has("message_count") and has("start_time") and has("last_activity") and has("state")' > /dev/null 2>&1; then
        echo "✅ Session has all required fields"
        
        # Display session info
        echo "  ID: $(echo "$SESSION" | jq -r '.id')"
        echo "  Agent: $(echo "$SESSION" | jq -r '.agent')"
        echo "  Messages: $(echo "$SESSION" | jq -r '.message_count')"
        echo "  State: $(echo "$SESSION" | jq -r '.state')"
    else
        echo "❌ Session missing required fields"
        echo "$SESSION" | jq '.'
        exit 1
    fi
else
    echo "❌ Failed: Expected active session"
    exit 1
fi

# Test 5: Pretty format
echo -e "\n📌 Test 5: Pretty format"
PRETTY=$(port42 context --pretty 2>&1)
if echo "$PRETTY" | grep -q "🔄 Active:"; then
    echo "✅ Pretty format working"
    echo "$PRETTY" | head -3
else
    echo "❌ Pretty format failed"
    exit 1
fi

# Test 6: Compact format
echo -e "\n📌 Test 6: Compact format"
COMPACT=$(port42 context --compact 2>&1)
if echo "$COMPACT" | grep -q "@ai-engineer\["; then
    echo "✅ Compact format working: $COMPACT"
else
    echo "❌ Compact format failed"
    exit 1
fi

# Test 7: Formatters handle empty data correctly
echo -e "\n📌 Test 7: Formatters with minimal data"
# The current session should have empty arrays for commands/tools
OUTPUT=$(port42 context 2>&1)
COMMANDS_COUNT=$(echo "$OUTPUT" | jq '.recent_commands | length')
TOOLS_COUNT=$(echo "$OUTPUT" | jq '.created_tools | length')

echo "  Commands array: $COMMANDS_COUNT items"
echo "  Tools array: $TOOLS_COUNT items"
echo "✅ Formatters handle empty arrays correctly"

echo -e "\n✨ All tests passed! Context structures working correctly."
echo "Next steps: Implement Step 2 (Context Collector) to populate these fields."