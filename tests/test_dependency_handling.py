#!/usr/bin/env python3
"""
Test dependency handling in Port 42 command generation
"""

import json
import socket
import time
import os

def send_request(req):
    sock = socket.socket()
    sock.connect(('localhost', 42))
    sock.send(json.dumps(req).encode() + b'\n')
    resp = sock.recv(16384)
    return json.loads(resp)

print("🐬 Testing Port 42 Dependency Handling\n")

# Test 1: Command with dependencies (lolcat)
print("1. Creating command with lolcat dependency...")
req = {
    "type": "possess",
    "id": f"test-deps-{int(time.time())}",
    "payload": {
        "agent": "@ai-engineer",
        "message": "Create a command called 'rainbow-logs' that shows system logs with rainbow colors using lolcat. Make it show the last 20 lines of /var/log/system.log"
    }
}

resp = send_request(req)
if resp.get('success'):
    print("✅ AI responded")
    if resp.get('data', {}).get('command_generated'):
        spec = resp.get('data', {}).get('command_spec', {})
        print(f"   Command: {spec.get('name')}")
        print(f"   Dependencies: {spec.get('dependencies', [])}")
    else:
        print("   No command generated")
else:
    print("❌ Request failed")

print()

# Test 2: Command without dependencies
print("2. Creating command without external dependencies...")
req = {
    "type": "possess",
    "id": f"test-nodeps-{int(time.time())}",
    "payload": {
        "agent": "@ai-engineer",
        "message": "Create a command called 'disk-alert' that checks disk usage and shows a warning if any disk is over 80% full. Use only bash built-ins and standard unix commands like df."
    }
}

resp = send_request(req)
if resp.get('success'):
    print("✅ AI responded")
    if resp.get('data', {}).get('command_generated'):
        spec = resp.get('data', {}).get('command_spec', {})
        print(f"   Command: {spec.get('name')}")
        print(f"   Dependencies: {spec.get('dependencies', [])}")
    else:
        print("   No command generated")

print()

# Check install script
print("3. Checking dependency installer...")
installer_path = os.path.expanduser("~/.port42/install-deps.sh")
if os.path.exists(installer_path):
    print(f"✅ Installer exists: {installer_path}")
    print("   Run it with: ~/.port42/install-deps.sh <dependency>")
else:
    print("❌ No installer found")

print("\n✨ Dependency handling test complete!")