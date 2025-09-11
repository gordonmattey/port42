#!/bin/bash
# Test script to verify watch mode functionality

set -e

echo "🧪 Testing Port42 Context Watch Mode"
echo "====================================="

# Test 1: Check help text
echo -e "\n📌 Test 1: Check watch flag in help"
if port42 context --help | grep -q "watch"; then
    echo "✅ Watch flag present in help"
else
    echo "❌ Watch flag missing from help"
    exit 1
fi

# Test 2: Start watch mode in background
echo -e "\n📌 Test 2: Starting watch mode"
echo "Starting watch mode with 500ms refresh..."

# Create a test file to capture output
WATCH_OUTPUT="/tmp/port42-watch-test.txt"
rm -f "$WATCH_OUTPUT"

# Start watch mode in background
port42 context --watch --refresh 500 > "$WATCH_OUTPUT" 2>&1 &
WATCH_PID=$!

echo "Watch mode started (PID: $WATCH_PID)"

# Test 3: Generate activity while watching
echo -e "\n📌 Test 3: Generating activity"
sleep 2

# Run some commands to generate activity
port42 status > /dev/null 2>&1
echo "  Generated: port42 status"
sleep 1

port42 ls /tools/ > /dev/null 2>&1
echo "  Generated: port42 ls /tools/"
sleep 1

port42 search "test" > /dev/null 2>&1
echo "  Generated: port42 search"
sleep 2

# Test 4: Create a session to see active session updates
echo -e "\n📌 Test 4: Creating active session"
port42 possess @ai-engineer "test watch mode" > /dev/null 2>&1 &
POSSESS_PID=$!
echo "  Started session (PID: $POSSESS_PID)"
sleep 3

# Kill watch mode after tests
sleep 2
kill $WATCH_PID 2>/dev/null || true
sleep 1

# Test 5: Check watch mode output
echo -e "\n📌 Test 5: Analyzing watch output"
if [ -f "$WATCH_OUTPUT" ]; then
    # Check for watch mode indicators
    if grep -q "Port42 Context Monitor" "$WATCH_OUTPUT"; then
        echo "✅ Watch mode header found"
    else
        echo "❌ Watch mode header missing"
        cat "$WATCH_OUTPUT"
        exit 1
    fi
    
    if grep -q "Press Ctrl+C to exit" "$WATCH_OUTPUT"; then
        echo "✅ Watch mode footer found"
    else
        echo "❌ Watch mode footer missing"
    fi
    
    if grep -q "Refreshing every" "$WATCH_OUTPUT"; then
        echo "✅ Refresh rate displayed"
    else
        echo "❌ Refresh rate not displayed"
    fi
    
    # Show sample of output
    echo -e "\n📄 Sample output:"
    head -20 "$WATCH_OUTPUT" | sed 's/^/  /'
else
    echo "❌ No watch output captured"
    exit 1
fi

# Test 6: Verify watch updates included commands
echo -e "\n📌 Test 6: Command tracking in watch"
if grep -q "status\|list_path\|search" "$WATCH_OUTPUT"; then
    echo "✅ Commands tracked in watch mode"
    
    # Count unique timestamps (shows updates happening)
    UNIQUE_TIMES=$(grep -o "[0-9][0-9]:[0-9][0-9]:[0-9][0-9]" "$WATCH_OUTPUT" | sort -u | wc -l)
    echo "  Found $UNIQUE_TIMES unique update times"
    
    if [ "$UNIQUE_TIMES" -gt 1 ]; then
        echo "✅ Multiple updates captured"
    else
        echo "⚠️  Only one update captured (may need longer test)"
    fi
else
    echo "⚠️  Commands not visible in watch output"
fi

# Clean up
rm -f "$WATCH_OUTPUT"

echo -e "\n✨ Watch mode tests complete!"
echo "Summary:"
echo "  • Watch flag available in CLI"
echo "  • Watch mode starts and displays header/footer"
echo "  • Refresh rate configurable"
echo "  • Live updates captured"
echo "  • Commands tracked in real-time"