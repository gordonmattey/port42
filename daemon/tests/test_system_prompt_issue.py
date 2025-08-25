#!/usr/bin/env python3
"""Test to validate system prompt override hypothesis"""

import json
import socket
import time

def send_request(req):
    """Send request to daemon and get response"""
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(('localhost', 42))
    
    sock.send(json.dumps(req).encode() + b'\n')
    
    response = b''
    while b'\n' not in response:
        chunk = sock.recv(4096)
        if not chunk:
            break
        response += chunk
    
    sock.close()
    return json.loads(response.decode().strip())

def test_session_with_history():
    """Test that demonstrates the system prompt issue"""
    
    session_id = f"prompt-test-{int(time.time())}"
    
    print(f"Testing session: {session_id}")
    print("="*60)
    
    # First message - introduce ourselves
    print("\n1. First message - setting context")
    req1 = {
        "type": "possess",
        "id": session_id,
        "payload": {
            "agent": "@ai-engineer",
            "message": "Help me create a command that shows disk usage beautifully"
        }
    }
    
    resp1 = send_request(req1)
    print(f"Success: {resp1.get('success')}")
    if resp1.get('success'):
        print(f"AI acknowledged: {len(resp1.get('data', {}).get('message', '')) > 0}")
    
    time.sleep(2)
    
    # Second message - test memory
    print("\n2. Second message - testing memory")
    req2 = {
        "type": "possess", 
        "id": session_id,
        "payload": {
            "agent": "@ai-engineer",
            "message": "What did we earlier in this session? Just answer directly, don't create any commands."
        }
    }
    
    resp2 = send_request(req2)
    if resp2.get('success'):
        message = resp2.get('data', {}).get('message', '').lower()
        print(f"\nAI Response preview: {message[:200]}...")
        
        # Check if AI remembers
        has_name = "testbot" in message
        has_number = "42" in message
        
        print(f"\nMemory check:")
        print(f"  Remembers name (TestBot): {has_name}")
        print(f"  Remembers number (42): {has_number}")
        
        if has_name and has_number:
            print("\n✅ AI correctly remembered the information!")
        else:
            print("\n❌ AI failed to remember the information")
            
            # Check for telltale signs of system prompt override
            if "don't have" in message or "can't access" in message or "fresh" in message:
                print("\n🔍 DIAGNOSIS: System prompt is likely overriding context")
                print("   The AI is responding as if this is a fresh conversation")

if __name__ == "__main__":
    test_session_with_history()