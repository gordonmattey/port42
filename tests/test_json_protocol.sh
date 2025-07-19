#!/bin/bash

echo "🐬 Testing Port 42 JSON Protocol..."
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to test JSON request
test_json() {
    local test_name=$1
    local json_request=$2
    local expected_field=$3
    
    echo -n "Test: $test_name ... "
    
    # Send JSON and capture response
    response=$(echo "$json_request" | nc localhost 42 2>/dev/null)
    
    if [ -z "$response" ]; then
        echo -e "${RED}FAILED${NC} - No response"
        return 1
    fi
    
    # Check if response contains expected field
    if echo "$response" | grep -q "$expected_field"; then
        echo -e "${GREEN}PASSED${NC}"
        echo "  Response: $response"
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Expected: $expected_field"
        echo "  Got: $response"
        return 1
    fi
}

# Test 1: Status request
echo "1. Testing status request"
test_json "status" \
    '{"type":"status","id":"test-1"}' \
    '"status":"swimming"'
echo

# Test 2: List request
echo "2. Testing list request"
test_json "list commands" \
    '{"type":"list","id":"test-2"}' \
    '"commands":\[\]'
echo

# Test 3: Possess request
echo "3. Testing possess request"
test_json "possess AI" \
    '{"type":"possess","id":"test-3","payload":{"agent":"muse","message":"Hello"}}' \
    '"message":"Possession mode not yet implemented'
echo

# Test 4: Unknown request type
echo "4. Testing unknown request type"
test_json "unknown type" \
    '{"type":"unknown","id":"test-4"}' \
    '"error":"Unknown request type: unknown"'
echo

# Test 5: Invalid JSON
echo "5. Testing invalid JSON"
echo -n "Test: invalid JSON ... "
response=$(echo "not json" | nc localhost 42 2>/dev/null)
if echo "$response" | grep -q '"error":"Invalid JSON request"'; then
    echo -e "${GREEN}PASSED${NC}"
    echo "  Response: $response"
else
    echo -e "${RED}FAILED${NC}"
    echo "  Expected error for invalid JSON"
    echo "  Got: $response"
fi
echo

# Test 6: Multiple requests (concurrency)
echo "6. Testing concurrent requests"
for i in {1..3}; do
    echo '{"type":"status","id":"concurrent-'$i'"}' | nc localhost 42 &
done
wait
echo

echo "✅ JSON protocol tests complete!"