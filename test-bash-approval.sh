#!/bin/bash

# Test script for bash approval flow

set -e

echo "🧪 Testing bash approval flow..."
echo "================================"
echo ""

# Make sure daemon is running
if ! ./bin/port42 status > /dev/null 2>&1; then
    echo "⚠️  Daemon not running. Starting daemon..."
    sudo -E ./bin/port42d -b > /dev/null 2>&1 &
    sleep 2
fi

echo "📝 Test 1: AI requests to analyze shell history"
echo "------------------------------------------------"
echo "When prompted, type 'y' to approve the bash command"
echo ""

./bin/port42 swim @ai-analyst "analyze my shell history for patterns"

echo ""
echo "✅ Test complete!"
echo ""
echo "Expected behavior:"
echo "1. AI should try to run a bash command to read shell history"
echo "2. You should see an approval prompt"
echo "3. After approval, the command should execute"
echo "4. Results should be displayed"