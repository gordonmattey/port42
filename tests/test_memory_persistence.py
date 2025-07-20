#!/usr/bin/env python3
"""
Test memory persistence in Port 42
"""

import json
import socket
import time
import os
import sys

def send_request(req):
    """Send JSON request to Port 42 daemon"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect(('localhost', 42))
        sock.send(json.dumps(req).encode() + b'\n')
        response = sock.recv(16384).decode()
        sock.close()
        return json.loads(response)
    except Exception as e:
        return {"error": str(e)}

print("🧪 Testing Port 42 Memory Persistence\n")

# Test 1: Create a session with some messages
print("1. Creating test session...")
session_id = f"memory-test-{int(time.time())}"

req1 = {
    "type": "possess",
    "id": session_id,
    "payload": {
        "agent": "@ai-engineer",
        "message": "Hello, I'm testing memory persistence"
    }
}

resp1 = send_request(req1)
if resp1.get('success'):
    print("✅ First message sent")
else:
    print(f"❌ Failed: {resp1.get('error')}")
    sys.exit(1)

time.sleep(2)

# Test 2: Send another message
print("\n2. Sending follow-up message...")
req2 = {
    "type": "possess",
    "id": session_id,
    "payload": {
        "agent": "@ai-engineer",
        "message": "Can you remember what I said before?"
    }
}

resp2 = send_request(req2)
if resp2.get('success'):
    print("✅ Second message sent")
else:
    print(f"❌ Failed: {resp2.get('error')}")

# Test 3: Check memory endpoint
print("\n3. Checking memory endpoint...")
req3 = {
    "type": "memory",
    "id": "test-memory"
}

resp3 = send_request(req3)
if resp3.get('success'):
    data = resp3.get('data', {})
    active = data.get('active_sessions', [])
    stats = data.get('stats', {})
    
    print(f"✅ Memory response received")
    print(f"   Active sessions: {len(active)}")
    print(f"   Total sessions: {stats.get('total_sessions', 0)}")
    print(f"   Commands generated: {stats.get('commands_generated', 0)}")
    
    # Find our test session
    found = False
    for session in active:
        if session.get('id') == session_id:
            found = True
            print(f"   ✓ Found our test session with {len(session.get('messages', []))} messages")
            break
    
    if not found:
        print("   ⚠️  Test session not found in active sessions")
else:
    print(f"❌ Memory request failed: {resp3.get('error')}")

# Test 4: Check persistence files
print("\n4. Checking persistence files...")
# Give async saves time to complete
time.sleep(2)

home = os.path.expanduser("~")
memory_dir = os.path.join(home, ".port42", "memory")
index_file = os.path.join(memory_dir, "index.json")

if os.path.exists(index_file):
    print(f"✅ Index file exists: {index_file}")
    
    # Read and display index
    with open(index_file, 'r') as f:
        index = json.load(f)
        print(f"   Sessions in index: {len(index.get('sessions', []))}")
        
    # Check for session files
    sessions_dir = os.path.join(memory_dir, "sessions")
    if os.path.exists(sessions_dir):
        # Count session files
        session_count = 0
        for root, dirs, files in os.walk(sessions_dir):
            session_count += len([f for f in files if f.endswith('.json')])
        print(f"   Session files on disk: {session_count}")
else:
    print(f"❌ No index file found at {index_file}")

# Test 5: End session
print("\n5. Ending session...")
req5 = {
    "type": "end",
    "id": session_id
}

resp5 = send_request(req5)
if resp5.get('success'):
    print("✅ Session ended")
else:
    print(f"❌ Failed to end session: {resp5.get('error')}")

print("\n✨ Memory persistence test complete!")
print("\n💡 To verify persistence:")
print("   1. Restart the daemon: sudo -E ./port42d")
print("   2. Check memory endpoint - sessions should be loaded from disk")
print("   3. Look in ~/.port42/memory/sessions/ for JSON files")