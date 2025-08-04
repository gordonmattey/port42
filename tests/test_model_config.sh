#!/bin/bash

echo "🔬 Testing model configuration consistency..."
echo ""

# Kill any existing daemon
killall port42d 2>/dev/null || true
sleep 1

# Build the daemon with fixes
echo "1️⃣ Building daemon with model config fixes..."
cd daemon && go build -o daemon . && cd ..
sudo cp daemon/daemon /usr/local/bin/port42d

echo ""
echo "2️⃣ Testing default model (from agents.json)..."
sudo -E /usr/local/bin/port42d > /tmp/model_test_default.log 2>&1 &
DAEMON_PID=$!
sleep 2

# Send a test message
./bin/port42 possess @ai-engineer "test-model-$(date +%s)" "Just say hello"

# Check what model was used
echo ""
echo "Model used (default config):"
grep "Claude API Request: model=" /tmp/model_test_default.log | tail -1

sudo kill $DAEMON_PID 2>/dev/null || true
sleep 1

echo ""
echo "3️⃣ Testing environment variable override..."
export CLAUDE_MODEL="claude-3-5-sonnet-20241022"
sudo -E /usr/local/bin/port42d > /tmp/model_test_env.log 2>&1 &
DAEMON_PID=$!
sleep 2

# Send another test message
./bin/port42 possess @ai-engineer "test-model-env-$(date +%s)" "Just say hello"

# Check what model was used
echo ""
echo "Model used (with CLAUDE_MODEL env):"
grep -E "(Using model from CLAUDE_MODEL|Claude API Request: model=)" /tmp/model_test_env.log | tail -2

sudo kill $DAEMON_PID 2>/dev/null || true
unset CLAUDE_MODEL

echo ""
echo "4️⃣ Checking rate limit logic..."
echo ""
echo "Rate limit detection in logs:"
grep -E "(rateLimitKey|rate limit)" /tmp/model_test_default.log | tail -5

echo ""
echo "5️⃣ Verifying storage uses correct model..."
echo ""
echo "Session metadata model:"
# Check the stored session files
find ~/.port42/metadata -name "*.json" -mmin -5 -exec grep -H '"model":' {} \; | tail -3

echo ""
echo "✅ Model configuration test complete!"