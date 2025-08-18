#!/bin/bash

# Test script to verify daemon-side reference resolution for possess mode
# This bypasses the CLI and sends a raw JSON request to the daemon

PORT=${1:-4242}
echo "Testing possess reference resolution on port $PORT..."

# Check if daemon is running
if ! nc -z 127.0.0.1 $PORT 2>/dev/null; then
    echo "❌ Daemon not running on port $PORT"
    echo "Start with: ./port42d"
    exit 1
fi

echo "✅ Daemon is running"

# Test 1: Simple file reference
echo "🧪 Test 1: File reference to git-status-enhanced command"

REQUEST1=$(cat <<'EOF'
{
    "type": "possess",
    "id": "test-ref-1",
    "payload": {
        "agent": "@ai-engineer",
        "message": "What do you know about the git-status-enhanced tool I referenced?"
    },
    "references": [
        {
            "type": "p42",
            "target": "/commands/git-status-enhanced",
            "context": "Referenced tool for testing"
        }
    ]
}
EOF
)

echo "Sending request with p42:/commands/git-status-enhanced reference..."
echo "$REQUEST1" | nc 127.0.0.1 $PORT

echo -e "\n" 

# Test 2: Search reference  
echo "🧪 Test 2: Search reference for architecture discussions"

REQUEST2=$(cat <<'EOF'
{
    "type": "possess", 
    "id": "test-ref-2",
    "payload": {
        "agent": "@ai-engineer",
        "message": "What architectural patterns do you see in the referenced search results?"
    },
    "references": [
        {
            "type": "search",
            "target": "architecture", 
            "context": "Architecture discussions for analysis"
        }
    ]
}
EOF
)

echo "Sending request with search:architecture reference..."
echo "$REQUEST2" | nc 127.0.0.1 $PORT

echo -e "\n"

# Test 3: No references (baseline)
echo "🧪 Test 3: No references (control test)"

REQUEST3=$(cat <<'EOF'
{
    "type": "possess",
    "id": "test-ref-3", 
    "payload": {
        "agent": "@ai-engineer",
        "message": "Hello, do you have any reference context available?"
    }
}
EOF
)

echo "Sending request without references..."
echo "$REQUEST3" | nc 127.0.0.1 $PORT

echo -e "\n✅ All tests sent. Check daemon logs for reference resolution debug output."