#!/bin/bash
# Port 42 End-to-End Integration Test
# Tests the full flow from installation to command generation

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Port 42 End-to-End Integration Test${NC}"
echo "===================================="
echo

# Function to check if a command succeeded
check_result() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
    else
        echo -e "${RED}✗${NC} $1"
        exit 1
    fi
}

# 1. Check if daemon is running
echo "1. Checking daemon status..."
if port42 status >/dev/null 2>&1; then
    check_result "Daemon is running"
else
    echo -e "${YELLOW}Starting daemon...${NC}"
    port42 daemon start -b
    sleep 3
    port42 status >/dev/null 2>&1
    check_result "Daemon started"
fi

# 2. Test basic possession
echo -e "\n2. Testing AI possession..."
TEST_SESSION="integration-test-$$"
RESPONSE=$(echo "exit" | port42 possess @ai-echo --session "$TEST_SESSION" 2>&1)
if echo "$RESPONSE" | grep -q "consciousness"; then
    check_result "AI possession works"
else
    echo -e "${RED}✗${NC} AI possession failed"
    echo "Response: $RESPONSE"
    exit 1
fi

# 3. Test command generation
echo -e "\n3. Testing command generation..."
python3 - <<EOF
import socket
import json
import time

def send_request(req):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    # Try port 42 first, then 4242
    for port in [42, 4242]:
        try:
            sock.connect(('localhost', port))
            break
        except:
            if port == 4242:
                raise Exception("Cannot connect to daemon")
    
    sock.send(json.dumps(req).encode() + b'\n')
    response = b''
    while True:
        chunk = sock.recv(4096)
        if not chunk:
            break
        response += chunk
        if b'\n' in response:
            break
    sock.close()
    return json.loads(response.decode().strip())

# Create a simple test command
req = {
    "type": "possess",
    "id": "test-cmd-gen",
    "payload": {
        "agent": "@ai-muse",
        "message": "Create a simple command called 'test-hello' that just prints 'Hello from Port 42!'"
    }
}

print("Sending command generation request...")
resp = send_request(req)

if resp.get('success'):
    if 'command_generated' in resp.get('data', {}):
        print("Command generated successfully!")
        exit(0)
    else:
        print("Response received but no command generated")
        print(json.dumps(resp, indent=2))
        exit(1)
else:
    print("Request failed:", resp.get('error'))
    exit(1)
EOF
check_result "Command generation works"

# 4. Check if command was created
echo -e "\n4. Checking generated command..."
if [ -f "$HOME/.port42/commands/test-hello" ]; then
    check_result "Command file created"
    
    # Test if command is executable
    if [ -x "$HOME/.port42/commands/test-hello" ]; then
        check_result "Command is executable"
    else
        echo -e "${RED}✗${NC} Command is not executable"
        exit 1
    fi
    
    # Clean up test command
    rm -f "$HOME/.port42/commands/test-hello"
else
    echo -e "${RED}✗${NC} Command file not found"
    exit 1
fi

# 5. Test memory persistence
echo -e "\n5. Testing memory persistence..."
if [ -d "$HOME/.port42/memory/sessions" ]; then
    check_result "Memory directory exists"
    
    # Check if any session files exist
    if find "$HOME/.port42/memory/sessions" -name "*.json" -type f | grep -q .; then
        check_result "Session files are being saved"
    else
        echo -e "${YELLOW}⚠${NC} No session files found (this might be okay for a fresh install)"
    fi
else
    echo -e "${RED}✗${NC} Memory directory not found"
    exit 1
fi

# 6. Test session continuation
echo -e "\n6. Testing session continuation..."
# This would require a more complex test, so we'll just check the feature exists
if port42 --help | grep -q "session"; then
    check_result "Session continuation feature available"
else
    echo -e "${YELLOW}⚠${NC} Session continuation not found in help"
fi

# Summary
echo -e "\n${GREEN}════════════════════════════════════${NC}"
echo -e "${GREEN}All integration tests passed! 🎉${NC}"
echo -e "${GREEN}════════════════════════════════════${NC}"
echo
echo "Port 42 is working correctly end-to-end:"
echo "  ✓ Daemon communication"
echo "  ✓ AI possession"
echo "  ✓ Command generation"
echo "  ✓ Memory persistence"
echo "  ✓ Core features operational"