#!/usr/bin/env python3
"""Test that temperature configuration is working"""

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

def test_temperature_setting():
    """Test that temperature is being set in API calls"""
    
    print("Testing temperature configuration...")
    print("="*60)
    print("\nNote: Check daemon logs for temperature value")
    print("Should see: 'Claude API Request: ... temp=0.50'")
    print("\n" + "="*60)
    
    session_id = f"temp-test-{int(time.time())}"
    
    req = {
        "type": "possess",
        "id": session_id,
        "payload": {
            "agent": "@ai-engineer",
            "message": "Just say 'Temperature test successful' and nothing else."
        }
    }
    
    resp = send_request(req)
    
    if resp.get('success'):
        print("\n✅ Request successful")
        print("Check daemon logs to verify temperature=0.50")
    else:
        print(f"\n❌ Request failed: {resp.get('error', 'Unknown error')}")

if __name__ == "__main__":
    test_temperature_setting()
    print("\n🔍 Remember to check daemon logs for: 'temp=0.50'")