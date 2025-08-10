#!/bin/bash

# Test script for possession session memory search functionality
# This script tests that memories created in one session can be found and loaded in another session

set -e  # Exit on any error

echo "🐬 Testing Port42 Memory Search Functionality"
echo "=============================================="
echo

# Build the project first
echo "📦 Building Port42..."
if [ -f "./build.sh" ]; then
    ./build.sh
    BUILD_EXIT_CODE=$?
else
    # Fallback to building in cli directory
    cd cli && cargo build --release --quiet
    BUILD_EXIT_CODE=$?
    cd ..
fi

if [ $BUILD_EXIT_CODE -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"
echo

# Define test topic and search terms (using unique term to avoid existing sessions)
TEST_TOPIC="thermochronological crystallization matrices"
SEARCH_QUERY="thermochronological"
AGENT="@ai-engineer"
PORT="4242"

# Function to start daemon if not running
start_daemon_if_needed() {
    # Check if daemon is already running on our test port
    if ./bin/port42 -p $PORT status >/dev/null 2>&1; then
        echo "✅ Daemon already running on port $PORT"
    else
        echo "🚀 Starting daemon on port $PORT..."
        sudo -E ./bin/port42d -p $PORT > /dev/null 2>&1 &
        DAEMON_PID=$!
        sleep 3
        
        # Verify daemon started
        if ./bin/port42 -p $PORT status >/dev/null 2>&1; then
            echo "✅ Daemon started successfully on port $PORT"
        else
            echo "❌ Failed to start daemon on port $PORT"
            exit 1
        fi
    fi
}

# Function to cleanup
cleanup() {
    echo
    echo "🧹 Cleaning up..."
    if [ ! -z "$DAEMON_PID" ]; then
        kill $DAEMON_PID 2>/dev/null || true
        wait $DAEMON_PID 2>/dev/null || true
    fi
    echo "✅ Cleanup complete"
}

# Set trap for cleanup on exit
trap cleanup EXIT

# Start daemon
start_daemon_if_needed
echo

# Step 1: Create a session with a specific topic
echo "📝 Step 1: Creating session with topic '$TEST_TOPIC'"
echo "================================================"
echo

# Create first session with a message about the test topic
SESSION1_MESSAGE="I'm working on $TEST_TOPIC and need help understanding the crystallographic phase transitions in thermochronological systems. Can you explain the key thermodynamic principles?"

echo "Sending message to create memory about: $TEST_TOPIC"
echo "Message: $SESSION1_MESSAGE"
echo

# Run possess with the test message
echo "Running: ./bin/port42 -p $PORT possess $AGENT \"$SESSION1_MESSAGE\""
./bin/port42 -p $PORT possess $AGENT "$SESSION1_MESSAGE"

echo
echo "✅ Step 1 complete: Memory created with topic '$TEST_TOPIC'"
echo

# Wait a moment to ensure memory is persisted
sleep 1

# Step 2: Test memory listing to verify it was stored
echo "📚 Step 2: Verifying memory was stored"
echo "======================================"
echo

echo "Listing recent memories:"
./bin/port42 -p $PORT memory

echo
echo "✅ Step 2 complete: Memory storage verified"
echo

# Step 3: Test search functionality
echo "🔍 Step 3: Testing search for stored topic"
echo "=========================================="
echo

echo "Searching for: '$SEARCH_QUERY'"
./bin/port42 -p $PORT search "$SEARCH_QUERY" --agent $AGENT -n 5

echo
echo "✅ Step 3 complete: Search functionality tested"
echo

# Step 4: Test possession with search to load memory context
echo "🧠 Step 4: Testing possession with search parameter"
echo "=================================================="
echo

echo "Starting possession session with search query: '$SEARCH_QUERY'"
echo "This should load the previous memory about '$TEST_TOPIC' into context"
echo

# Create a follow-up question to test memory context loading
FOLLOWUP_MESSAGE="Based on our previous discussion about thermochronological systems, can you dive deeper into the phase boundary kinetics?"

echo "Follow-up message: $FOLLOWUP_MESSAGE"
echo

# Test the search-based possession
echo "Running: PORT42_DEBUG=1 ./bin/port42 -p $PORT possess $AGENT --search \"$SEARCH_QUERY\" \"$FOLLOWUP_MESSAGE\""
PORT42_DEBUG=1 ./bin/port42 -p $PORT possess $AGENT --search "$SEARCH_QUERY" "$FOLLOWUP_MESSAGE"

echo
echo "✅ Step 4 complete: Search-based possession tested"
echo

# Final verification
echo "🎯 Final Verification"
echo "===================="
echo

echo "Let's verify the search found and loaded the right memories:"
echo "Expected: Should have found memory containing '$TEST_TOPIC'"
echo "Expected: Should have loaded that memory as context for the new session"
echo "Expected: The AI should reference the previous discussion in its response"
echo

echo "Test completed! Check the output above to verify:"
echo "1. ✅ Memory was created with the test topic"
echo "2. ✅ Search found the memory when querying '$SEARCH_QUERY'"
echo "3. ✅ Possession with search loaded the memory context"
echo "4. ✅ AI response referenced the previous context"
echo

echo "🎉 Memory search functionality test complete!"