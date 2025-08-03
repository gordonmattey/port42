#!/bin/bash

echo "🔬 Testing artifact generation with async fix..."
echo ""


# Create multiple artifacts to test for OS error 35
echo "2️⃣ Creating artifacts to test for OS error 35..."
echo ""

SESSION_ID="test-async-$(date +%s)"

# Test 1: Single file artifact
echo "Test 1: Single file artifact"
./bin/port42 possess @ai-engineer "$SESSION_ID" "Create an artifact called 'test-doc' containing a markdown document about async programming. Use generate_artifact tool."
echo ""

# Test 2: Multi-file artifact (more complex)
echo "Test 2: Multi-file artifact"
./bin/port42 possess @ai-engineer "$SESSION_ID-2" "Create a multi-file artifact called 'web-app' with an HTML file, CSS file, and JavaScript file for a simple counter app. Use generate_artifact tool."
echo ""

# Test 3: Large artifact
echo "Test 3: Large artifact"
./bin/port42 possess @ai-engineer "$SESSION_ID-3" "Create an artifact called 'large-doc' with a very detailed markdown guide (at least 100 lines) about the benefits of asynchronous programming. Use generate_artifact tool."
echo ""

# Give time for async operations to complete
sleep 3

echo "3️⃣ Checking results..."
echo ""

# Check if artifacts were created
echo "📂 Artifacts created:"
./bin/port42 ls /artifacts
echo ""

# Check for any errors in daemon log
echo "🔍 Checking for errors in daemon log:"
grep -E "(OS error|error 35|Failed to generate artifact)" /Users/gordon/.port42/daemon.log || echo "✅ No OS errors found!"
echo ""

# Check for successful artifact creation
echo "✅ Successful artifact generations:"
grep -E "(Artifact stored|Artifact generation completed)" /Users/gordon/.port42/daemon.log | tail -10

echo ""
echo "4️⃣ Cleanup..."
sudo kill $DAEMON_PID 2>/dev/null || true

echo ""
echo "✅ Test complete!"