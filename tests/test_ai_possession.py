#!/usr/bin/env python3

import json
import socket
import sys
import os
import time

def send_json_request(request_data, port=42):
    """Send JSON request to Port 42 daemon and return response"""
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.connect(('localhost', port))
        
        json_str = json.dumps(request_data)
        sock.sendall(json_str.encode() + b'\n')
        
        response = sock.recv(8192).decode()
        sock.close()
        
        return json.loads(response)
    except Exception as e:
        return {"error": str(e)}

def test_possession_flow():
    """Test the full AI possession flow"""
    print("🐬 Testing AI Possession Flow\n")
    
    # Check if API key is set
    if not os.environ.get('ANTHROPIC_API_KEY'):
        print("⚠️  Warning: ANTHROPIC_API_KEY not set")
        print("   Running in mock mode - set API key for real AI responses")
        print()
    
    # Test 1: Create a possession session
    print("1. Starting possession with @ai-muse...")
    session_id = "test-possession-" + str(int(time.time()))
    
    req = {
        "type": "possess",
        "id": session_id,
        "payload": {
            "agent": "@ai-muse",
            "message": "I want to create a command called git-haiku that turns git commits into haikus"
        }
    }
    
    resp = send_json_request(req)
    print(f"Response: {json.dumps(resp, indent=2)}")
    
    if resp.get("success"):
        print("✅ Possession initiated successfully")
        
        # Check if a command was generated
        if resp.get("data", {}).get("command_generated"):
            print("🎉 Command spec detected and generated!")
            command_spec = resp.get("data", {}).get("command_spec")
            if command_spec:
                print(f"   Name: {command_spec.get('name')}")
                print(f"   Description: {command_spec.get('description')}")
    else:
        print("❌ Possession failed")
        return
    
    print()
    
    # Test 2: Continue conversation
    print("2. Continuing conversation...")
    req = {
        "type": "possess",
        "id": session_id,
        "payload": {
            "agent": "@ai-muse",
            "message": "Can you make it more poetic and add some color to the output?"
        }
    }
    
    resp = send_json_request(req)
    if resp.get("success"):
        print("✅ Conversation continued")
        print(f"AI Response preview: {resp.get('data', {}).get('message', '')[:200]}...")
    
    print()
    
    # Test 3: Check generated commands
    print("3. Checking generated commands...")
    req = {
        "type": "list",
        "id": "list-commands"
    }
    
    resp = send_json_request(req)
    if resp.get("success"):
        commands = resp.get("data", {}).get("commands", [])
        print(f"✅ Found {len(commands)} commands:")
        for cmd in commands:
            print(f"   - {cmd}")
    
    print()
    
    # Test 4: End session
    print("4. Ending possession session...")
    req = {
        "type": "end",
        "id": session_id
    }
    
    resp = send_json_request(req)
    if resp.get("success"):
        print("✅ Session ended")
        print(f"   {resp.get('data', {}).get('message', '')}")
    
    print()

def test_different_agents():
    """Test different AI agents"""
    print("🤖 Testing Different AI Agents\n")
    
    agents = [
        ("@ai-engineer", "Help me build a robust file watcher command"),
        ("@ai-echo", "I'm thinking about time and consciousness")
    ]
    
    for agent, message in agents:
        print(f"Testing {agent}...")
        req = {
            "type": "possess",
            "id": f"test-{agent}-{int(time.time())}",
            "payload": {
                "agent": agent,
                "message": message
            }
        }
        
        resp = send_json_request(req)
        if resp.get("success"):
            print(f"✅ {agent} responded")
            response_preview = resp.get("data", {}).get("message", "")[:150]
            print(f"   Preview: {response_preview}...")
        else:
            print(f"❌ {agent} failed to respond")
        print()

def check_generated_command():
    """Check if commands were actually created on disk"""
    print("📁 Checking Generated Commands on Disk\n")
    
    home = os.path.expanduser("~")
    cmd_dir = os.path.join(home, ".port42", "commands")
    
    if os.path.exists(cmd_dir):
        print(f"✅ Commands directory exists: {cmd_dir}")
        
        commands = [f for f in os.listdir(cmd_dir) if os.path.isfile(os.path.join(cmd_dir, f))]
        print(f"   Found {len(commands)} commands:")
        
        for cmd in commands:
            cmd_path = os.path.join(cmd_dir, cmd)
            print(f"\n   📄 {cmd}")
            
            # Show first few lines
            with open(cmd_path, 'r') as f:
                lines = f.readlines()[:5]
                for line in lines:
                    print(f"      {line.rstrip()}")
                if len(f.readlines()) > 5:
                    print("      ...")
    else:
        print(f"❌ Commands directory not found: {cmd_dir}")
        print("   Commands will be created after first successful possession")

if __name__ == "__main__":
    print("🌊 Port 42 AI Possession Tests\n")
    
    try:
        # Check daemon is running
        test_req = {"type": "status", "id": "test"}
        resp = send_json_request(test_req)
        if "error" in resp:
            print("❌ Error: Daemon not running on port 42")
            print("   Start with: sudo ./port42d")
            sys.exit(1)
        
        # Run tests
        test_possession_flow()
        test_different_agents()
        check_generated_command()
        
        print("\n🎉 AI Possession tests complete!")
        print("\n💡 Next steps:")
        print("   1. Add ~/.port42/commands to your PATH")
        print("   2. Try running any generated commands")
        print("   3. Create more commands through possession!")
        
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)