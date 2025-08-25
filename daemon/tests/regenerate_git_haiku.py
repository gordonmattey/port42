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

print("🐬 Regenerating git-haiku with better implementation...\n")

# Start possession
session_id = f"git-haiku-fix-{int(time.time())}"
req = {
    "type": "possess",
    "id": session_id,
    "payload": {
        "agent": "@ai-engineer",
        "message": "Create a git-haiku command that gets the last 10 git commits using 'git log --oneline -n 10' and formats them as haikus. Each commit should be transformed into 3 lines in a poetic way. Add colors for visual appeal. Make sure it actually calls git log, doesn't wait for stdin."
    }
}

resp = send_request(req)
print("Response from AI:")
print(resp.get('data', {}).get('message', '')[:500] + "...")

if resp.get('data', {}).get('command_generated'):
    print("\n✅ Command generated successfully!")
    print("Try running: git-haiku")
else:
    print("\n⚠️  No command was generated. Try being more specific.")