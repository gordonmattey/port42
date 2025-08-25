#!/usr/bin/env python3
"""Test that the system prompt fix resolves session continuation"""

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

def test_fixed_system_prompt():
    """Test that system prompt fix resolves the issue"""
    
    session_id = f"system-fix-test-{int(time.time())}"
    
    print(f"Testing fixed system prompt with session: {session_id}")
    print("="*60)
    
    # First message - create something memorable
    print("\n1. First message - establishing context")
    req1 = {
        "type": "possess",
        "id": session_id,
        "payload": {
            "agent": "@ai-engineer",
            "message": "I want to create a command called 'rainbow-ls' that lists files with rainbow colors. Remember this name!"
        }
    }
    
    resp1 = send_request(req1)
    print(f"Success: {resp1.get('success')}")
    
    time.sleep(2)
    
    # Second message - test memory with complex prompt
    print("\n2. Second message - testing memory with complex prompt")
    req2 = {
        "type": "possess", 
        "id": session_id,
        "payload": {
            "agent": "@ai-engineer",
            "message": "What did we discuss earlier in this session? What was the command name I mentioned? Just answer directly."
        }
    }
    
    resp2 = send_request(req2)
    if resp2.get('success'):
        message = resp2.get('data', {}).get('message', '').lower()
        print(f"\nAI Response preview: {message[:300]}...")
        
        # Check if AI remembers
        has_rainbow = "rainbow" in message
        has_ls = "ls" in message or "rainbow-ls" in message
        
        print(f"\nMemory check:")
        print(f"  Remembers 'rainbow': {has_rainbow}")
        print(f"  Remembers 'ls/rainbow-ls': {has_ls}")
        
        # Check for contradiction
        has_contradiction = ("first" in message or "no earlier" in message or "don't have access" in message) and ("rainbow" in message or "discussed" in message)
        
        if has_rainbow and has_ls and not has_contradiction:
            print("\n✅ SUCCESS! AI correctly remembered without contradiction!")
        elif has_contradiction:
            print("\n⚠️  PARTIAL: AI shows both memory and denial (contradiction)")
        else:
            print("\n❌ FAILED: AI didn't remember the command name")

if __name__ == "__main__":
    test_fixed_system_prompt()