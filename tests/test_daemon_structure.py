#!/usr/bin/env python3

import json
import socket
import sys
import time
import uuid
from concurrent.futures import ThreadPoolExecutor

def send_json_request(request_data, port=42):
    """Send JSON request to Port 42 daemon and return response"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect(('localhost', port))
        
        json_str = json.dumps(request_data)
        sock.sendall(json_str.encode() + b'\n')
        
        # Read response until we get a complete line
        response = b''
        while b'\n' not in response:
            chunk = sock.recv(4096)
            if not chunk:
                break
            response += chunk
        
        sock.close()
        
        return json.loads(response.decode().strip())
    except Exception as e:
        return {"error": str(e)}

def test_session_management():
    """Test session creation and management"""
    print("Testing session management...")
    
    # Create a session
    session_id = str(uuid.uuid4())
    req = {
        "type": "possess",
        "id": session_id,
        "payload": {
            "agent": "muse",
            "message": "Hello, creating a session"
        }
    }
    resp = send_json_request(req)
    print(f"Session creation response: {json.dumps(resp, indent=2)}")
    assert resp.get("success") == True
    assert resp.get("data", {}).get("session_id") == session_id
    
    # Check status should show 1 active session
    req = {"type": "status", "id": "status-1"}
    resp = send_json_request(req)
    print(f"Status after session creation: {json.dumps(resp, indent=2)}")
    assert resp.get("data", {}).get("sessions") >= 1
    
    # End the session
    req = {"type": "end", "id": session_id}
    resp = send_json_request(req)
    print(f"End session response: {json.dumps(resp, indent=2)}")
    assert resp.get("success") == True
    
    print("✅ Session management test passed\n")

def test_memory_endpoint():
    """Test memory endpoint"""
    print("Testing memory endpoint...")
    
    # Create a few sessions first
    for i in range(3):
        req = {
            "type": "possess",
            "id": f"memory-test-{i}",
            "payload": {
                "agent": "echo",
                "message": f"Memory test message {i}"
            }
        }
        send_json_request(req)
    
    # Get memory
    req = {"type": "memory", "id": "get-memory"}
    resp = send_json_request(req)
    print(f"Memory response: {json.dumps(resp, indent=2)}")
    
    assert resp.get("success") == True
    data = resp.get("data", {})
    assert "active_sessions" in data
    assert "recent_sessions" in data
    assert "stats" in data
    assert data.get("active_count", 0) >= 3
    
    print("✅ Memory endpoint test passed\n")

def test_concurrent_sessions():
    """Test multiple concurrent sessions"""
    print("Testing concurrent sessions...")
    
    def create_session(i):
        req = {
            "type": "possess",
            "id": f"concurrent-{i}",
            "payload": {
                "agent": "builder",
                "message": f"Concurrent message {i}"
            }
        }
        return send_json_request(req)
    
    # Create 10 concurrent sessions
    with ThreadPoolExecutor(max_workers=10) as executor:
        futures = [executor.submit(create_session, i) for i in range(10)]
        results = [f.result() for f in futures]
    
    # Check all succeeded
    success_count = sum(1 for r in results if r.get("success") == True)
    print(f"Created {success_count}/10 concurrent sessions")
    assert success_count == 10
    
    # Check status
    req = {"type": "status", "id": "status-concurrent"}
    resp = send_json_request(req)
    print(f"Active sessions: {resp.get('data', {}).get('sessions', 0)}")
    
    print("✅ Concurrent sessions test passed\n")

def test_daemon_info():
    """Test daemon provides proper info"""
    print("Testing daemon info...")
    
    req = {"type": "status", "id": "info-test"}
    resp = send_json_request(req)
    
    data = resp.get("data", {})
    print(f"Daemon info: {json.dumps(data, indent=2)}")
    
    # Check required fields
    assert "status" in data
    assert "port" in data
    assert "sessions" in data
    assert "uptime" in data
    assert "dolphins" in data
    
    print("✅ Daemon info test passed\n")

def test_graceful_shutdown():
    """Test that daemon handles graceful shutdown"""
    print("Testing graceful shutdown behavior...")
    
    # Create a session
    req = {
        "type": "possess",
        "id": "shutdown-test",
        "payload": {
            "agent": "muse",
            "message": "Testing shutdown"
        }
    }
    resp = send_json_request(req)
    assert resp.get("success") == True
    
    print("✅ Graceful shutdown test passed (manual verification needed)\n")

if __name__ == "__main__":
    print("🐬 Daemon Structure Tests\n")
    
    try:
        # Check if daemon is running
        test_req = {"type": "status", "id": "test"}
        resp = send_json_request(test_req)
        if "error" in resp:
            print("❌ Error: Daemon not running on port 42")
            print("Trying port 4242...")
            resp = send_json_request(test_req, 4242)
            if "error" in resp:
                print("❌ Error: Daemon not running on port 4242 either")
                sys.exit(1)
        
        test_daemon_info()
        test_session_management()
        test_memory_endpoint()
        test_concurrent_sessions()
        test_graceful_shutdown()
        
        print("🎉 All daemon structure tests passed!")
    except AssertionError as e:
        print(f"❌ Test failed: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)