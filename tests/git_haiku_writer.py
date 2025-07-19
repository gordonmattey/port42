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

print("🐬 Let's create a git commit haiku writer!\n")

session_id = f"haiku-writer-{int(time.time())}"

req = {
    "type": "possess",
    "id": session_id,
    "payload": {
        "agent": "@ai-muse",
        "message": "I want a command called 'git-haiku-commit' that helps me write git commits AS haikus. It should open an editor with a template showing 5-7-5 syllable format, then use that as my commit message. Maybe with syllable counting?"
    }
}

resp = send_request(req)
print("Muse responds:")
print(resp.get('data', {}).get('message', ''))

if resp.get('data', {}).get('command_generated'):
    print("\n✨ Command crystallized!")
    spec = resp.get('data', {}).get('command_spec', {})
    print(f"Created: {spec.get('name')}")
else:
    print("\nNo command generated yet...")