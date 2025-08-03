#!/bin/bash

echo "🔬 Testing large response handling..."
echo ""

# 1. Check response size for commands
echo "1️⃣ Checking daemon response sizes in logs..."
echo ""
sudo grep -E "Possess response size|Large possess response" /var/log/port42d.log | tail -20

echo ""
echo "2️⃣ Testing with debug enabled..."
export PORT42_DEBUG=1
export PORT42_VERBOSE=1

# Kill any existing daemon
killall port42d 2>/dev/null || true
sleep 1

# Start daemon
echo ""
echo "3️⃣ Starting daemon with fresh logs..."
sudo -E /usr/local/bin/port42d > /tmp/daemon_debug.log 2>&1 &
DAEMON_PID=$!
sleep 2

# Create a command with controlled size
echo ""
echo "4️⃣ Creating a command that generates predictable response size..."
SESSION_ID="test-size-$(date +%s)"

./bin/port42 possess @ai-engineer "Create a simple command called 'test-small' that just prints 'hello'. Keep the implementation very short, under 100 lines." --session "$SESSION_ID" 2>&1 | tee /tmp/client_debug.log

echo ""
echo "5️⃣ Checking daemon debug output..."
sudo cat /tmp/daemon_debug.log | grep -E "(response size|Large possess|Error encoding)"

echo ""
echo "6️⃣ Checking client debug output..."
grep -E "(DEBUG|bytes|timeout)" /tmp/client_debug.log

echo ""
echo "7️⃣ Cleanup..."
sudo kill $DAEMON_PID 2>/dev/null || true
unset PORT42_DEBUG PORT42_VERBOSE

echo ""
echo "✅ Test complete!"