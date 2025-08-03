#!/bin/bash

echo "🔬 Testing new model configuration architecture..."
echo ""

# Kill any existing daemon
killall port42d 2>/dev/null || true
sleep 1

# Copy and start the new daemon
echo "1️⃣ Installing updated daemon..."
sudo cp daemon/daemon /usr/local/bin/port42d

echo ""
echo "2️⃣ Checking agents.json structure..."
echo ""
echo "Models defined:"
grep -A5 '"models"' daemon/agents.json | grep '"id"'
echo ""
echo "Agent model assignments:"
grep -B1 -A1 '"model"' daemon/agents.json | grep -E '("name"|"model")'

echo ""
echo "3️⃣ Starting daemon and testing each agent..."
sudo -E /usr/local/bin/port42d > /tmp/daemon_model_test.log 2>&1 &
DAEMON_PID=$!
sleep 2

# Test each agent
echo ""
echo "Testing @ai-engineer (should use opus-4):"
./bin/port42 possess @ai-engineer "test-engineer-$(date +%s)" "Just say hello and tell me what model you are"
sleep 1

echo ""
echo "Testing @ai-muse (should use opus-4 with temp override):"
./bin/port42 possess @ai-muse "test-muse-$(date +%s)" "Just say hello and tell me what model you are"
sleep 1

echo ""
echo "4️⃣ Checking daemon logs for model usage..."
echo ""
echo "Model loading logs:"
grep -E "(Model for agent|Claude API Request: model=)" /tmp/daemon_model_test.log | tail -10

echo ""
echo "Rate limiting logs:"
grep -E "(Rate limiting|minDelay)" /tmp/daemon_model_test.log | tail -5

echo ""
echo "5️⃣ Testing undefined agent (should fail gracefully):"
./bin/port42 possess @ai-unknown "test-unknown-$(date +%s)" "Hello" 2>&1 | grep -E "(Failed|Error|not found)"

echo ""
echo "6️⃣ Cleanup..."
sudo kill $DAEMON_PID 2>/dev/null || true

echo ""
echo "✅ Model configuration test complete!"