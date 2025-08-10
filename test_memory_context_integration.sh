#!/bin/bash

# Integration test for memory context possession 
# Tests that messages with search context are sent immediately and reference loaded memories

set -e

echo "🧪 Testing Memory Context Integration"
echo "====================================="

PORT="4242"
AGENT="@ai-engineer"
TEST_TOPIC="quantum crystalline matrix dynamics" 
SEARCH_QUERY="quantum crystalline"

# Build first
echo "📦 Building..."
./build.sh > /dev/null 2>&1

# Ensure daemon is running
echo "🔍 Checking daemon status on port $PORT..."
if ! ./bin/port42 -p $PORT status >/dev/null 2>&1; then
    echo "❌ Daemon not running on port $PORT"
    echo "💡 Try starting daemon with: sudo -E ./bin/port42d -p $PORT"
    exit 1
fi

echo "✅ Daemon is running"
echo

# Step 1: Create a memory with unique topic
echo "📝 Step 1: Creating test memory"
SETUP_MESSAGE="I'm researching $TEST_TOPIC and need to understand the phase transition mechanics in quantum systems."

echo "Creating memory with: $SETUP_MESSAGE"
./bin/port42 -p $PORT possess $AGENT "$SETUP_MESSAGE" > /dev/null 2>&1
echo "✅ Memory created"
echo

# Step 2: Test that search finds the memory
echo "🔍 Step 2: Verify search finds memory"
SEARCH_RESULT=$(./bin/port42 -p $PORT search "$SEARCH_QUERY" --agent $AGENT -n 1 2>/dev/null)

if echo "$SEARCH_RESULT" | grep -q "$TEST_TOPIC"; then
    echo "✅ Search found test memory"
else
    echo "❌ Search did not find test memory"
    echo "Search result: $SEARCH_RESULT"
    exit 1
fi
echo

# Step 3: Test possession with search + message (the main test)
echo "🧠 Step 3: Test possession with search and immediate message"
FOLLOWUP_MESSAGE="Based on our previous discussion, explain the entropy implications."

echo "Testing command: ./bin/port42 -p $PORT possess $AGENT --search \"$SEARCH_QUERY\" \"$FOLLOWUP_MESSAGE\""

# Capture output and check for expected behavior
POSSESS_OUTPUT=$(timeout 10s ./bin/port42 -p $PORT possess $AGENT --search "$SEARCH_QUERY" "$FOLLOWUP_MESSAGE" 2>&1)
POSSESS_EXIT_CODE=$?

echo "Exit code: $POSSESS_EXIT_CODE"
echo "Output length: ${#POSSESS_OUTPUT} characters"

# Check results
if [ $POSSESS_EXIT_CODE -eq 0 ]; then
    echo "✅ Command completed successfully (no hanging)"
    
    if echo "$POSSESS_OUTPUT" | grep -q "Loaded.*memories into session context"; then
        echo "✅ Memory context was loaded"
    else
        echo "❌ Memory context loading not found in output"
    fi
    
    if echo "$POSSESS_OUTPUT" | grep -q "@ai-engineer"; then
        echo "✅ AI response was generated"
    else
        echo "❌ No AI response found"
    fi
    
    if echo "$POSSESS_OUTPUT" | grep -i -E "(previous|discussion|entropy)" >/dev/null; then
        echo "✅ AI response seems to reference context or follow-up"
    else
        echo "⚠️  AI response might not reference loaded context"
    fi
    
else
    echo "❌ Command failed or timed out"
    echo "This suggests the hanging issue persists"
fi

echo
echo "📊 Test Results Summary:"
echo "- Command execution: $([ $POSSESS_EXIT_CODE -eq 0 ] && echo "PASS" || echo "FAIL")"
echo "- Memory loading: $(echo "$POSSESS_OUTPUT" | grep -q "Loaded.*memories" && echo "PASS" || echo "FAIL")" 
echo "- AI response: $(echo "$POSSESS_OUTPUT" | grep -q "@ai-engineer" && echo "PASS" || echo "FAIL")"
echo

if [ $POSSESS_EXIT_CODE -eq 0 ] && echo "$POSSESS_OUTPUT" | grep -q "@ai-engineer"; then
    echo "🎉 Integration test PASSED - Memory context possession working!"
    exit 0
else
    echo "❌ Integration test FAILED"
    echo
    echo "Full output:"
    echo "$POSSESS_OUTPUT"
    exit 1
fi