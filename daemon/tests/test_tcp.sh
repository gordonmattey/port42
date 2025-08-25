#!/bin/bash

echo "🐬 Testing Port 42 TCP Server..."
echo

# Test 1: Simple echo
echo "Test 1: Basic echo"
echo "Hello dolphins" | nc localhost 42
echo

# Test 2: Multiple connections
echo "Test 2: Multiple connections"
for i in {1..3}; do
    echo "Connection $i" | nc localhost 42 &
done
wait
echo

# Test 3: Empty message
echo "Test 3: Empty message"
echo "" | nc localhost 42
echo

echo "✅ TCP tests complete!"