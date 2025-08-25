#!/usr/bin/env python3

import json
import socket
import time

def send_request(req):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(('localhost', 42))
    sock.sendall(json.dumps(req).encode() + b'\n')
    response = sock.recv(8192).decode()
    sock.close()
    return json.loads(response)

print("🐬 Natural conversation with @ai-muse\n")

session_id = f"natural-{int(time.time())}"

# Simple, natural request - just like the user would type
req = {
    "type": "possess",
    "id": session_id,
    "payload": {
        "agent": "@ai-muse",
        "message": "I need a command that shows my git commits as haikus"
    }
}

resp = send_request(req)
print("Muse responds:")
print(resp.get('data', {}).get('message', ''))
print()

# Check if command was generated
if resp.get('data', {}).get('command_generated'):
    print("✨ A command crystallized!")
    spec = resp.get('data', {}).get('command_spec', {})
    print(f"Name: {spec.get('name')}")
    print(f"Purpose: {spec.get('description')}")
else:
    print("Continue the conversation to refine your vision...")