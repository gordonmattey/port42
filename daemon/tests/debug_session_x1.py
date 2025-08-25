#!/usr/bin/env python3
"""Debug the x1 session issue"""

import json
import subprocess
import time

def check_session_file():
    """Check if x1 session file exists and its contents"""
    import glob
    import os
    
    print("Looking for x1 session files...")
    pattern = os.path.expanduser("~/.port42/memory/sessions/*/session-*x1*.json")
    files = glob.glob(pattern)
    
    if not files:
        # Check for the specific file mentioned in logs
        specific_file = os.path.expanduser("~/.port42/memory/sessions/2025-07-20/session-1753040510-diskview.json")
        if os.path.exists(specific_file):
            files = [specific_file]
    
    if files:
        for f in files:
            print(f"\nFound: {f}")
            with open(f, 'r') as fp:
                data = json.load(fp)
                print(f"Session ID: {data.get('id')}")
                print(f"Agent: {data.get('agent')}")
                print(f"Messages: {len(data.get('messages', []))}")
                
                # Show message summary
                for i, msg in enumerate(data.get('messages', [])):
                    preview = msg['content'][:100] + "..." if len(msg['content']) > 100 else msg['content']
                    print(f"  [{i}] {msg['role']}: {preview}")
    else:
        print("No x1 session files found")

def test_x1_again():
    """Test the x1 session continuation"""
    print("\n\nTesting x1 session continuation...")
    print("="*60)
    
    # Ask about previous work
    cmd = [
        "port42", "possess", "@ai-engineer",
        "--session", "x1",
        "What command did we create in our previous conversation? Just tell me the name."
    ]
    
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, capture_output=True, text=True, env={**subprocess.os.environ, "PORT42_DEBUG": "1"})
    
    print(f"\nExit code: {result.returncode}")
    print("\nOutput:")
    print(result.stdout)
    
    if "diskview" in result.stdout.lower():
        print("\n✅ AI correctly remembered the diskview command!")
    else:
        print("\n❌ AI doesn't seem to remember diskview")

if __name__ == "__main__":
    check_session_file()
    test_x1_again()