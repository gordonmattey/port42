#!/bin/bash
# Test all possess syntax variations

echo "=== Testing Port 42 Possess Syntax ==="
echo
echo "Build the latest CLI first..."
cd /Users/gordon/Dropbox/Work/Hacking/workspace/port42/cli
cargo build --release
cp target/release/port42 ../bin/port42

echo
echo "Test cases to verify manually:"
echo
echo "1. TEST: possess @claude"
echo "   EXPECT: Creates new session (Session: cli-TIMESTAMP)"
echo "   RUN: ./bin/port42 possess @claude"
echo
echo "2. TEST: possess @claude x1"  
echo "   EXPECT: Continues session x1 (↻ Continuing session: x1)"
echo "   RUN: ./bin/port42 possess @claude x1"
echo
echo "3. TEST: possess @claude \"help with git\""
echo "   EXPECT: New session with message (Session: cli-TIMESTAMP)"
echo "   RUN: ./bin/port42 possess @claude \"help with git\""
echo
echo "4. TEST: possess @claude x1 \"what did we discuss?\""
echo "   EXPECT: Continues x1 with message (↻ Continuing session: x1)"
echo "   RUN: ./bin/port42 possess @claude x1 \"what did we discuss?\""
echo
echo "5. TEST: Inside shell - same behaviors"
echo "   RUN: ./bin/port42"
echo "   Then test all commands above without './bin/port42' prefix"
echo
echo "=== Key behavior changes ==="
echo "- possess @agent ALWAYS creates new session (no more auto-continue)"
echo "- Must explicitly specify memory ID to continue"
echo "- All syntaxes now supported including memory + message"