#!/bin/bash
# Debug test for artifact generation - shows exactly what's being sent

echo "🔍 Port 42 Artifact Debug Test"
echo "=============================="
echo

# Simple test with minimal prompt
echo "📍 Sending simple artifact request to @ai-engineer"
echo "Create a markdown artifact called test-doc" | ./bin/port42 possess "@ai-engineer"

echo
echo "⏳ Waiting for daemon to process..."
sleep 3

echo
echo "📜 Checking daemon logs for debugging info:"
echo "==========================================="
tail -50 ~/.port42/daemon.log | grep -E "(Tool [0-9]+:|System prompt contains XML|Sending request with|Checking tools for agent|will use|generate_)"

echo
echo "📊 Checking if any artifacts were created:"
./bin/port42 ls /artifacts

echo
echo "💡 Key things to check in logs above:"
echo "  1. How many tools are being sent? (should be 2)"
echo "  2. What are the tool names? (should include generate_artifact)"
echo "  3. Does system prompt contain XML tags? (should be true)"
echo "  4. Which agent tools are assigned? (should show 'will use tool-based generation')"