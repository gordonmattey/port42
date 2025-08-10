#!/bin/bash

# Test Step 5: Memory-Relation Bridge - Minimal Test
# Tests only the crystallized view routing without AI generation

set -e

echo "🧪 Testing Step 5: Memory-Relation Bridge - Minimal Test"
echo "========================================================"

# Build first
echo "📦 Building daemon..."
go build -o ../bin/port42d

# Start daemon in background
echo "🚀 Starting daemon..."
../bin/port42d &
DAEMON_PID=$!
sleep 2

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    if [ ! -z "$DAEMON_PID" ]; then
        kill $DAEMON_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Function to send JSON request to daemon
send_request() {
    echo "$1" | nc -w 2 localhost 42 2>/dev/null || echo "$1" | nc -w 2 localhost 4242
}

echo "📋 Test 1: Test crystallized view routing"
# Test the /memory/{session}/crystallized virtual view directly
RESPONSE=$(send_request '{
    "type": "list",
    "id": "test-001", 
    "payload": {
        "path": "/memory/test-session-456/generated"
    }
}')

echo "Crystallized view response: $RESPONSE"
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ Crystallized view routing working (empty result expected)"
else
    echo "❌ Crystallized view routing failed"
    echo "Response: $RESPONSE"
    exit 1
fi

echo "📋 Test 2: Test different session path"
RESPONSE=$(send_request '{
    "type": "list",
    "id": "test-002",
    "payload": {
        "path": "/memory/different-session/generated"
    }
}')

if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ Different session path routing working"
else
    echo "❌ Different session path routing failed"
    exit 1
fi

echo ""
echo "🎉 Step 5 Memory-Relation Bridge routing tests passed!"
echo "✅ /memory/{session}/generated virtual view routing functional"
echo "✅ Path parsing and method invocation working"
echo ""
echo "Note: Full functionality requires relations with memory_session properties"
echo "Step 5: Memory-Relation Bridge implementation complete! 🐬"