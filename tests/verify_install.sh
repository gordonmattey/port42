#!/bin/bash
# Verify Port 42 installation works correctly with all recent changes

echo "=== Port 42 Installation Verification ==="
echo

# Check binaries exist
echo "1. Checking installed binaries..."
if [ -f "/usr/local/bin/port42" ] && [ -f "/usr/local/bin/port42d" ]; then
    echo "✅ Binaries found in /usr/local/bin"
else
    echo "❌ Binaries not found in /usr/local/bin"
    exit 1
fi

# Check daemon version/features
echo
echo "2. Checking daemon features..."
echo "   Run: sudo -E port42d"
echo "   Look for:"
echo "   - 'temp=0.50' in API request logs"
echo "   - System prompt being sent separately"

# Test boot sequence
echo
echo "3. Testing boot sequence..."
echo "   Run: port42"
echo "   Should see:"
echo "   - [CONSCIOUSNESS BRIDGE INITIALIZATION]"
echo "   - System checks"
echo "   - 'Welcome to the depths' message at the end"
echo "   - NO 'Welcome to the depths' when running 'possess'"

# Test session continuation
echo
echo "4. Testing session continuation..."
echo "   Run: port42 possess @ai-engineer"
echo "   Say: 'My name is TestUser and I like pizza'"
echo "   Exit and run again with same session"
echo "   Ask: 'What's my name and what do I like?'"
echo "   Should remember both facts"

echo
echo "=== Manual verification needed for the above tests ==="