#!/bin/bash

# Test Step 5: Memory-Relation Bridge - Crystallized View
# Tests the /memory/{session}/crystallized virtual view functionality

set -e

echo "🧪 Testing Step 5: Memory-Relation Bridge - Crystallized View"
echo "============================================================"

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
    # Clean up test files
    rm -f ~/.port42/relations/test-*.json 2>/dev/null || true
}
trap cleanup EXIT

# Function to send JSON request to daemon
send_request() {
    echo "$1" | nc -w 2 localhost 42 2>/dev/null || echo "$1" | nc -w 2 localhost 4242
}

echo "📋 Test 1: Declare tool with session context"
# Declare a tool with explicit session context
RESPONSE=$(send_request '{
    "type": "declare_relation",
    "id": "test-001",
    "payload": {
        "relation": {
            "type": "Tool",
            "properties": {
                "name": "test-crystallized-tool",
                "transforms": ["test", "memory"]
            }
        }
    },
    "session_context": {
        "session_id": "memory-test-session-456",
        "agent": "ai-test"
    }
}')

echo "Response: $RESPONSE"
if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ Tool declared with session context"
else
    echo "❌ Tool declaration failed"
    echo "Response: $RESPONSE"
    exit 1
fi

echo "📋 Test 2: Verify relation has memory_session property"
# List relations to verify the memory_session property was added
RESPONSE=$(send_request '{
    "type": "list_relations", 
    "id": "test-002",
    "payload": {}
}')

if echo "$RESPONSE" | grep -q "memory-test-session-456"; then
    echo "✅ Relation contains memory_session property"
else
    echo "❌ memory_session property not found in relation"
    echo "Response: $RESPONSE"
    exit 1
fi

echo "📋 Test 3: Access generated view"
# Test the /memory/{session}/generated virtual view
RESPONSE=$(send_request '{
    "type": "list",
    "id": "test-003", 
    "payload": {
        "path": "/memory/memory-test-session-456/generated"
    }
}')

echo "Generated view response: $RESPONSE"
if echo "$RESPONSE" | grep -q "test-crystallized-tool"; then
    echo "✅ Generated view shows tool from session"
else
    echo "❌ Generated view does not show expected tool"
    echo "Expected: test-crystallized-tool"
    echo "Response: $RESPONSE"
    exit 1
fi

echo "📋 Test 4: Multiple tools from same session"
# Declare another tool from the same session
RESPONSE=$(send_request '{
    "type": "declare_relation",
    "id": "test-004",
    "payload": {
        "relation": {
            "type": "Tool", 
            "properties": {
                "name": "test-second-tool",
                "transforms": ["test", "second"]
            }
        }
    },
    "session_context": {
        "session_id": "memory-test-session-456",
        "agent": "ai-test"
    }
}')

if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ Second tool declared with same session"
else
    echo "❌ Second tool declaration failed"
    exit 1
fi

# Check generated view now shows both tools
RESPONSE=$(send_request '{
    "type": "list",
    "id": "test-005",
    "payload": {
        "path": "/memory/memory-test-session-456/generated"
    }
}')

if echo "$RESPONSE" | grep -q "test-crystallized-tool" && echo "$RESPONSE" | grep -q "test-second-tool"; then
    echo "✅ Generated view shows both tools from session"
else
    echo "❌ Generated view does not show both tools"
    echo "Response: $RESPONSE"
    exit 1
fi

echo "📋 Test 5: Different session isolation"
# Declare tool from different session - should not appear in first session's view
RESPONSE=$(send_request '{
    "type": "declare_relation", 
    "id": "test-006",
    "payload": {
        "relation": {
            "type": "Tool",
            "properties": {
                "name": "test-isolated-tool",
                "transforms": ["test", "isolated"]
            }
        }
    },
    "session_context": {
        "session_id": "memory-test-session-789",
        "agent": "ai-test"
    }
}')

if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ Tool declared in different session"
else
    echo "❌ Different session tool declaration failed"
    exit 1
fi

# Original session view should NOT contain the isolated tool
RESPONSE=$(send_request '{
    "type": "list",
    "id": "test-007",
    "payload": {
        "path": "/memory/memory-test-session-456/generated"
    }
}')

if echo "$RESPONSE" | grep -q "test-isolated-tool"; then
    echo "❌ Session isolation failed - isolated tool appears in wrong session"
    echo "Response: $RESPONSE"
    exit 1
else
    echo "✅ Session isolation working - isolated tool not in original session view"
fi

# New session view should contain the isolated tool
RESPONSE=$(send_request '{
    "type": "list", 
    "id": "test-008",
    "payload": {
        "path": "/memory/memory-test-session-789/generated"
    }
}')

if echo "$RESPONSE" | grep -q "test-isolated-tool"; then
    echo "✅ New session view contains its tool"
else
    echo "❌ New session view missing its tool"
    echo "Response: $RESPONSE" 
    exit 1
fi

echo ""
echo "🎉 All Step 5 Memory-Relation Bridge tests passed!"
echo "✅ Session context capture working"
echo "✅ memory_session property added to relations"
echo "✅ /memory/{session}/generated virtual view functional"
echo "✅ Multi-tool sessions supported"  
echo "✅ Session isolation working correctly"
echo ""
echo "Step 5: Memory-Relation Bridge implementation complete! 🐬"