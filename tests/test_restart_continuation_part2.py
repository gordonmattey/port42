#!/usr/bin/env python3
"""
Part 2: Run this AFTER restarting daemon to test session recovery
"""

import json
import socket
import sys
import time

# Use the same session ID from part 1
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

print("🧪 Testing Session Recovery After Restart (Part 2)\n")

# Test 1: Check if daemon is running
print("1. Checking daemon status...")
status_req = {"type": "status", "id": "test"}
status_resp = send_request(status_req)
if "error" in status_resp:
    print("❌ Daemon not running. Please start it first.")
    sys.exit(1)
else:
    print("✅ Daemon is running")

# Test 2: Try to continue the previous session
print(f"\n2. Attempting to continue session '{SESSION_ID}'...")
req = {
    "type": "possess",
    "id": SESSION_ID,
    "payload": {
        "agent": "@ai-engineer",
        "message": "What's my name and what project am I working on? Also, what's my favorite color?"
    }
}

resp = send_request(req)
if resp.get('success'):
    response_text = resp.get('data', {}).get('message', '')
    print("✅ Got response from AI")
    print(f"\n📝 AI Response:\n{response_text}\n")
    
    # Check if context was preserved
    context_preserved = False
    if "gordon" in response_text.lower():
        print("✓ AI remembered the name!")
        context_preserved = True
    else:
        print("✗ AI didn't remember the name")
        
    if "port 42" in response_text.lower():
        print("✓ AI remembered the project!")
        context_preserved = True
    else:
        print("✗ AI didn't remember the project")
        
    if "purple" in response_text.lower():
        print("✓ AI remembered the favorite color!")
        context_preserved = True
    else:
        print("✗ AI didn't remember the favorite color")
    
    print(f"\n{'='*60}")
    if context_preserved:
        print("🎉 SUCCESS! Session context was preserved across restart!")
        print("The session continuation feature is working correctly.")
    else:
        print("❌ FAILED: Session context was not preserved.")
        print("The AI started a new session instead of continuing the old one.")
else:
    print(f"❌ Failed to get response: {resp.get('error')}")

# Test 3: Check memory endpoint to see session details
print(f"\n3. Checking session details...")
mem_req = {"type": "memory", "id": "test"}
mem_resp = send_request(mem_req)

if mem_resp.get('success'):
    sessions = mem_resp.get('data', {}).get('active_sessions', [])
    recent = mem_resp.get('data', {}).get('recent_sessions', [])
    
    # Check active sessions
    for session in sessions:
        if session.get('id') == SESSION_ID:
            print(f"✅ Found in active sessions:")
            print(f"   Messages: {len(session.get('messages', []))}")
            print(f"   Created: {session.get('created_at')}")
            print(f"   State: {session.get('state')}")
            break
    
    # Check recent sessions (loaded from disk)
    for session in recent:
        if session.get('id') == SESSION_ID:
            print(f"✅ Found in recent sessions (loaded from disk):")
            print(f"   Messages: {session.get('message_count')}")
            break

print("\n✨ Test complete!")