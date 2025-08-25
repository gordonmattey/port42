#!/usr/bin/env python3

import json
import socket
import time

def send_request(req):
    sock = socket.socket()
    sock.connect(('localhost', 42))
    sock.send(json.dumps(req).encode() + b'\n')
    return json.loads(sock.recv(8192))

print("🐬 Let's fix git-haiku with proper colors!\n")

req = {
    "type": "possess",
    "id": f"fix-haiku-{int(time.time())}",
    "payload": {
        "agent": "@ai-engineer",
        "message": "Create a command called 'git-haiku-v2' that shows git commits as haikus with colors. Use printf or echo -e for color codes. Make it simpler - just show the commit hash and message in 3 lines with nice formatting."
    }
}

resp = send_request(req)
print("Engineer responds:")
print(resp.get('data', {}).get('message', '')[:500] + "...")

if resp.get('data', {}).get('command_generated'):
    print("\n✨ New git-haiku command created!")
else:
    print("\n⚠️  Try being more specific about the implementation")