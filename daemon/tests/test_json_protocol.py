#!/usr/bin/env python3

import json
import socket
import sys

def send_json_request(request_data):
    """Send JSON request to Port 42 daemon and return response"""
    try:
        # Connect to daemon
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect(('localhost', 42))
        
        # Send JSON request
        json_str = json.dumps(request_data)
        sock.sendall(json_str.encode() + b'\n')
        
        # Receive response
        response = sock.recv(4096).decode()
        sock.close()
        
        return json.loads(response)
    except Exception as e:
        return {"error": str(e)}

def test_status():
    """Test status request"""
    print("Testing status request...")
    req = {
        "type": "status",
        "id": "py-test-1"
    }
    resp = send_json_request(req)
    print(f"Response: {json.dumps(resp, indent=2)}")
    assert resp.get("success") == True
    assert "swimming" in resp.get("data", {}).get("status", "")
    print("✅ Status test passed\n")

def test_list():
    """Test list request"""
    print("Testing list request...")
    req = {
        "type": "list",
        "id": "py-test-2"
    }
    resp = send_json_request(req)
    print(f"Response: {json.dumps(resp, indent=2)}")
    assert resp.get("success") == True
    assert "commands" in resp.get("data", {})
    print("✅ List test passed\n")

def test_possess():
    """Test possess request"""
    print("Testing possess request...")
    req = {
        "type": "possess",
        "id": "py-test-3",
        "payload": {
            "agent": "muse",
            "message": "Hello from Python"
        }
    }
    resp = send_json_request(req)
    print(f"Response: {json.dumps(resp, indent=2)}")
    assert resp.get("success") == True
    print("✅ Possess test passed\n")

def test_unknown_type():
    """Test unknown request type"""
    print("Testing unknown request type...")
    req = {
        "type": "unknown",
        "id": "py-test-4"
    }
    resp = send_json_request(req)
    print(f"Response: {json.dumps(resp, indent=2)}")
    assert resp.get("success") == False
    assert "Unknown request type" in resp.get("error", "")
    print("✅ Unknown type test passed\n")

if __name__ == "__main__":
    print("🐬 Python JSON Protocol Tests\n")
    
    try:
        test_status()
        test_list()
        test_possess()
        test_unknown_type()
        
        print("🎉 All tests passed!")
    except AssertionError as e:
        print(f"❌ Test failed: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"❌ Error: {e}")
        print("Make sure the daemon is running on port 42")
        sys.exit(1)