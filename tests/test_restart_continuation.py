#!/usr/bin/env python3
"""
Manual test for session continuation after daemon restart
Run this AFTER restarting the daemon to test true session recovery
"""

import json
import socket
import sys

# Use a known session ID from previous run
SESSION_ID = "continuation-test-manual"

def send_request(req, port=42):
    """Send JSON request to Port 42 daemon"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect(('localhost', port))
        sock.send(json.dumps(req).encode() + b'\n')
        response = sock.recv(16384).decode()
        sock.close()
        return json.loads(response)
    except Exception as e:
        return {"error": str(e)}

print("🧪 Testing Session Continuation After Restart\n")

# Test 1: Set up context in a new session
print("1. Setting up initial context...")
req1 = {
    "type": "possess",
    "id": SESSION_ID,
    "payload": {
        "agent": "@ai-engineer",
        "message": "My name is Gordon and I'm working on a project called Port 42. I like the color purple."
    }
}

resp1 = send_request(req1)
if resp1.get('success'):
    print("✅ Initial context set")
    print(f"   Response: {resp1.get('data', {}).get('message', '')[:100]}...")
else:
    print(f"❌ Failed: {resp1.get('error')}")

# Test 2: Add more context
print("\n2. Adding more context...")
req2 = {
    "type": "possess",
    "id": SESSION_ID,
    "payload": {
        "agent": "@ai-engineer",
        "message": "I need help creating a command that shows disk usage in a tree format"
    }
}

resp2 = send_request(req2)
if resp2.get('success'):
    print("✅ Additional context added")
else:
    print(f"❌ Failed: {resp2.get('error')}")

# Test 3: Check session was saved
print("\n3. Checking memory endpoint...")
req3 = {
    "type": "memory",
    "id": "test"
}

resp3 = send_request(req3)
if resp3.get('success'):
    data = resp3.get('data', {})
    active = data.get('active_sessions', [])
    
    found = False
    for session in active:
        if session.get('id') == SESSION_ID:
            found = True
            print(f"✅ Session found with {len(session.get('messages', []))} messages")
            break
    
    if not found:
        print("❌ Session not found in active sessions")

print("\n" + "="*60)
print("🔄 NOW RESTART THE DAEMON")
print("="*60)
print("\nTo test continuation:")
print("1. Stop the daemon (Ctrl+C)")
print("2. Restart: sudo -E ./bin/port42d")
print("3. Run: ./tests/test_restart_continuation_part2.py")
print("\nThe part 2 script will test if the session context is preserved.")