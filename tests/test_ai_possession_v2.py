#!/usr/bin/env python3
"""
Port 42 AI Possession Test Suite
Easily add new tests by updating the TEST_CASES list
"""

import json
import socket
import sys
import os
import time

# Define all test cases here
TEST_CASES = [
    {
        "name": "Git Haiku (Simple)",
        "agent": "@ai-engineer",
        "message": "I want to create a command called git-haiku that turns git commits into haikus.",
        "expect_command": True
    },
    {
        "name": "Disk Usage Tree",
        "agent": "@ai-engineer", 
        "message": "I need a command that shows disk usage as a tree.",
        "expect_command": True
    },
    {
        "name": "System Info Dashboard",
        "agent": "@ai-muse",
        "message": "I want a command called 'sys-dash' that shows a beautiful dashboard with system info like CPU, memory, disk usage",
        "expect_command": True
    },
    {
        "name": "Philosophy Discussion",
        "agent": "@ai-echo",
        "message": "What is the nature of consciousness in digital systems?",
        "expect_command": False  # This is just conversation
    }
]

def send_json_request(request_data, port=42):
    """Send JSON request to Port 42 daemon and return response"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect(('localhost', port))
        
        json_str = json.dumps(request_data)
        sock.sendall(json_str.encode() + b'\n')
        
        response = sock.recv(16384).decode()  # Increased buffer size
        sock.close()
        
        return json.loads(response)
    except Exception as e:
        return {"error": str(e)}

def run_test_case(test_case):
    """Run a single test case"""
    print(f"\n{'='*60}")
    print(f"🧪 Test: {test_case['name']}")
    print(f"   Agent: {test_case['agent']}")
    print(f"   Expect command: {test_case['expect_command']}")
    print(f"{'='*60}\n")
    
    session_id = f"test-{int(time.time())}-{test_case['name'].lower().replace(' ', '-')}"
    
    req = {
        "type": "possess",
        "id": session_id,
        "payload": {
            "agent": test_case['agent'],
            "message": test_case['message']
        }
    }
    
    print(f"📤 Sending: {test_case['message'][:80]}...")
    resp = send_json_request(req)
    
    if resp.get("success"):
        print("✅ AI responded successfully")
        
        # Show response preview
        ai_message = resp.get('data', {}).get('message', '')
        print(f"\n📝 AI says: {ai_message[:200]}...")
        
        # Check if command was generated
        if resp.get("data", {}).get("command_generated"):
            spec = resp.get("data", {}).get("command_spec", {})
            print(f"\n🎉 Command generated!")
            print(f"   Name: {spec.get('name')}")
            print(f"   Description: {spec.get('description')}")
            print(f"   Language: {spec.get('language')}")
            
            if not test_case['expect_command']:
                print("   ⚠️  Unexpected: Command was generated but not expected")
        else:
            if test_case['expect_command']:
                print("\n❌ FAIL: Expected command generation but none occurred")
                print("   Hint: AI might need clearer instructions or JSON format reminder")
            else:
                print("\n✅ PASS: No command expected, none generated")
    else:
        print(f"❌ Request failed: {resp.get('error', 'Unknown error')}")
    
    # Small delay between tests
    time.sleep(1)
    return resp.get("data", {}).get("command_generated", False)

def check_generated_commands():
    """Check what commands exist on disk"""
    print(f"\n{'='*60}")
    print("📁 Generated Commands on Disk")
    print(f"{'='*60}\n")
    
    home = os.path.expanduser("~")
    cmd_dir = os.path.join(home, ".port42", "commands")
    
    if os.path.exists(cmd_dir):
        commands = [f for f in os.listdir(cmd_dir) if os.path.isfile(os.path.join(cmd_dir, f))]
        print(f"Found {len(commands)} commands in {cmd_dir}:\n")
        
        for cmd in sorted(commands):
            cmd_path = os.path.join(cmd_dir, cmd)
            size = os.path.getsize(cmd_path)
            print(f"  📄 {cmd:<20} ({size} bytes)")
            
            # Show shebang line
            with open(cmd_path, 'r') as f:
                first_line = f.readline().strip()
                print(f"     {first_line}")
    else:
        print(f"❌ Commands directory not found: {cmd_dir}")

def main():
    print("🌊 Port 42 AI Possession Test Suite v2\n")
    
    # Check daemon is running
    test_req = {"type": "status", "id": "test"}
    resp = send_json_request(test_req)
    if "error" in resp:
        print("❌ Error: Daemon not running on port 42")
        print("   Start with: sudo -E ./port42d")
        sys.exit(1)
    
    print(f"✅ Daemon running on port {resp.get('data', {}).get('port', '42')}")
    print(f"   Uptime: {resp.get('data', {}).get('uptime', 'unknown')}")
    
    # Check API key
    if not os.environ.get('ANTHROPIC_API_KEY'):
        print("\n⚠️  Warning: ANTHROPIC_API_KEY not set - using mock mode")
    
    # Run all test cases
    generated_count = 0
    for test_case in TEST_CASES:
        if run_test_case(test_case):
            generated_count += 1
    
    # Summary
    print(f"\n{'='*60}")
    print(f"📊 Test Summary")
    print(f"{'='*60}")
    print(f"   Total tests: {len(TEST_CASES)}")
    print(f"   Commands generated: {generated_count}")
    print(f"   Tests expecting commands: {sum(1 for t in TEST_CASES if t['expect_command'])}")
    
    # Check disk
    check_generated_commands()
    
    print("\n✨ Test suite complete!")
    print("\n💡 Tips:")
    print("   - Add new tests to TEST_CASES at the top of this script")
    print("   - Set expect_command=True if the test should generate a command")
    print("   - Be specific in prompts to get reliable command generation")

if __name__ == "__main__":
    main()