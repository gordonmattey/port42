#!/usr/bin/env python3
"""Test session continuation from CLI perspective"""

import subprocess
import time
import json

def run_cli_command(args):
    """Run a CLI command and return output"""
    cmd = ["port42"] + args
    print(f"Running: {' '.join(cmd)}")
    
    result = subprocess.run(cmd, capture_output=True, text=True)
    print(f"Exit code: {result.returncode}")
    print(f"Output preview: {result.stdout[:200]}...")
    return result

def test_session_continuation():
    """Test that session continuation works from CLI"""
    
    session_id = f"cli-test-{int(time.time())}"
    
    print(f"\n1. Creating initial session: {session_id}")
    print("="*50)
    
    # First command - introduce ourselves
    result1 = run_cli_command([
        "possess", "@ai-engineer",
        "--session", session_id,
        "My name is TestUser and I'm testing session continuation. My favorite color is blue."
    ])
    
    if result1.returncode != 0:
        print(f"❌ First command failed: {result1.stderr}")
        return False
    
    time.sleep(2)  # Give daemon time to save
    
    print(f"\n2. Continuing session: {session_id}")
    print("="*50)
    
    # Second command - ask about previous info
    result2 = run_cli_command([
        "possess", "@ai-engineer", 
        "--session", session_id,
        "What's my name and favorite color? Don't create any commands."
    ])
    
    if result2.returncode != 0:
        print(f"❌ Second command failed: {result2.stderr}")
        return False
    
    # Check if AI remembered
    output = result2.stdout.lower()
    
    print("\n3. Checking AI response for memory")
    print("="*50)
    
    if "testuser" in output and "blue" in output:
        print("✅ AI correctly remembered the name and color!")
        return True
    elif "don't have" in output or "not able" in output or "don't know" in output:
        print("❌ AI doesn't seem to remember the conversation")
        print(f"Full output:\n{result2.stdout}")
        return False
    else:
        print("⚠️  Unclear if AI remembered. Full output:")
        print(result2.stdout)
        return False

if __name__ == "__main__":
    print("🐬 Testing CLI Session Continuation")
    print("="*50)
    
    success = test_session_continuation()
    
    if success:
        print("\n✅ Session continuation is working correctly!")
    else:
        print("\n❌ Session continuation has issues")
        print("\nDebugging tips:")
        print("1. Check daemon logs: tail -f ~/.port42/daemon.log")
        print("2. Check session files: ls -la ~/.port42/memory/sessions/*/")
        print("3. Run with debug: PORT42_DEBUG=1 port42 possess ...")
    
    exit(0 if success else 1)