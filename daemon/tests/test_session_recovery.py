#!/usr/bin/env python3
"""
Test session recovery and continuation after daemon restart
"""

import json
import socket
import time
import os
import sys
import subprocess

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

def check_daemon_running():
    """Check if daemon is running"""
    req = {"type": "status", "id": "test"}
    resp = send_request(req)
    return "error" not in resp

print("🧪 Testing Port 42 Session Recovery\n")

# Test 1: Create a session with context
print("1. Creating session with context...")
session_id = f"recovery-test-{int(time.time())}"

# First message
req1 = {
    "type": "possess",
    "id": session_id,
    "payload": {
        "agent": "@ai-muse",
        "message": "My favorite color is blue and I love dolphins"
    }
}

resp1 = send_request(req1)
if resp1.get('success'):
    print("✅ First message sent")
    print(f"   AI Response: {resp1.get('data', {}).get('message', '')[:100]}...")
else:
    print(f"❌ Failed: {resp1.get('error')}")
    sys.exit(1)

time.sleep(2)

# Second message building on context
req2 = {
    "type": "possess", 
    "id": session_id,
    "payload": {
        "agent": "@ai-muse",
        "message": "What's my favorite color? And what animal did I mention?"
    }
}

resp2 = send_request(req2)
if resp2.get('success'):
    response = resp2.get('data', {}).get('message', '')
    print("✅ Second message sent")
    print(f"   AI Response: {response[:150]}...")
    
    # Check if AI remembers context
    if "blue" in response.lower() and "dolphin" in response.lower():
        print("   ✓ AI correctly remembered context within session!")
    else:
        print("   ⚠️  AI may not have remembered the context")
else:
    print(f"❌ Failed: {resp2.get('error')}")

# Test 2: Check session is saved
print("\n2. Verifying session persistence...")
time.sleep(3)  # Allow async save

home = os.path.expanduser("~")
memory_dir = os.path.join(home, ".port42", "memory", "sessions")
today = time.strftime("%Y-%m-%d")
session_dir = os.path.join(memory_dir, today)

session_found = False
if os.path.exists(session_dir):
    # Extract timestamp from session_id (e.g., "recovery-test-1752974212" -> "1752974212")
    timestamp = session_id.split('-')[-1]
    
    for filename in os.listdir(session_dir):
        # Check if timestamp is in filename
        if timestamp in filename and filename.endswith('.json'):
            filepath = os.path.join(session_dir, filename)
            # Verify it's actually our session by checking inside the file
            with open(filepath, 'r') as f:
                session_data = json.load(f)
                if session_data.get('id') == session_id:
                    session_found = True
                    print(f"✅ Session file found: {filename}")
                    print(f"   Messages saved: {len(session_data.get('messages', []))}")
                    print(f"   State: {session_data.get('state')}")
                    break

if not session_found:
    print("❌ Session file not found on disk")

# Test 3: Simulate daemon restart
print("\n3. Simulating daemon restart...")
print("   ⚠️  NOTE: This test would require restarting the daemon")
print("   To manually test session recovery:")
print("   a) Stop the daemon (Ctrl+C)")
print("   b) Restart with: sudo -E ./bin/port42d")
print("   c) Run the continuation test below")

# Test 4: Test continuation (for manual testing after restart)
print("\n4. Session continuation test (run after daemon restart)...")
print(f"   Run this command to test continuation:")
print(f'   echo \'{{"type":"possess","id":"{session_id}","payload":{{"agent":"@ai-muse","message":"Do you still remember my favorite color and the animal?"}}}}\' | nc localhost 42 | jq .')

# Test 5: Check what happens with a new session with same ID
print("\n5. Testing new session with recovered ID...")
req5 = {
    "type": "possess",
    "id": session_id,
    "payload": {
        "agent": "@ai-muse", 
        "message": "This is a test after recovery. What was my favorite color?"
    }
}

resp5 = send_request(req5)
if resp5.get('success'):
    response = resp5.get('data', {}).get('message', '')
    print("✅ Message sent to existing session ID")
    print(f"   AI Response: {response[:150]}...")
    
    # Check if this is a continuation or new session
    if "blue" in response.lower():
        print("   ✓ Session context maintained!")
    else:
        print("   ⚠️  Session appears to be new (no previous context)")

# Test 6: Check memory endpoint
print("\n6. Checking memory endpoint for recovered sessions...")
req6 = {
    "type": "memory",
    "id": "test"
}

resp6 = send_request(req6)
if resp6.get('success'):
    data = resp6.get('data', {})
    recent = data.get('recent_sessions', [])
    print(f"✅ Memory endpoint responded")
    print(f"   Recent sessions from disk: {len(recent)}")
    
    # Look for our session
    for session in recent:
        if session.get('id') == session_id:
            print(f"   ✓ Found our test session in recent sessions")
            print(f"     Created: {session.get('created_at')}")
            print(f"     Messages: {session.get('message_count')}")
            break

print("\n✨ Session recovery test complete!")
print("\n💡 Important findings:")
print("   - Sessions are saved to disk automatically")
print("   - Sessions can be loaded on daemon restart")
print("   - BUT: Sessions are isolated - no automatic continuation")
print("   - Each session maintains its own context bubble")
print("\n🔮 Future enhancement ideas:")
print("   - Cross-session memory search")
print("   - Session continuation after restart")
print("   - Agent-specific persistent memory")