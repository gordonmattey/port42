#!/bin/bash

# Simple test for ViewerRule spawning
echo "Testing ViewerRule spawning..."

# Test 1: Create analysis tool and check for viewer spawning
echo "Creating analysis tool quick-analyzer..."
./cli/target/release/port42 declare tool quick-analyzer --transforms quick,analysis &

# Wait for it to complete (give it 60 seconds max)
wait

# Check results using which command
echo ""
echo "Checking results:"

if which quick-analyzer >/dev/null 2>&1; then
    echo "✅ Main tool quick-analyzer created"
else
    echo "❌ Main tool quick-analyzer NOT created"
fi

if which view-quick-analyzer >/dev/null 2>&1; then
    echo "✅ Viewer tool view-quick-analyzer AUTO-SPAWNED!"
    echo "🎉 ViewerRule is working!"
else
    echo "❌ Viewer tool view-quick-analyzer NOT spawned"
    echo "❓ ViewerRule may not be working"
fi

echo ""
echo "Test complete."