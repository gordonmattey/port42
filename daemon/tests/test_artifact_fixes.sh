#!/bin/bash

# Test script to verify artifact fixes

echo "🔧 Testing artifact generation fixes..."
echo ""

# Kill any existing daemon
echo "1️⃣ Stopping any existing daemon..."
killall port42d 2>/dev/null || true
sleep 1

# Start the daemon with the updated binary
echo "2️⃣ Starting daemon with updated binary..."
sudo cp daemon/daemon /usr/local/bin/port42d
sudo -E /usr/local/bin/port42d &
DAEMON_PID=$!
sleep 2

# Test artifact generation with a simple request
echo "3️⃣ Testing artifact generation..."
echo ""

SESSION_ID="test-artifacts-$(date +%s)"

# Send a message that should generate an artifact
./bin/port42 possess @ai-engineer "Please create an artifact containing a simple markdown document about the benefits of Port 42. Use the generate_artifact tool to create this as a document type artifact." --session "$SESSION_ID"

echo ""
echo "4️⃣ Checking if artifact was created..."
sleep 2

# List artifacts directory
echo ""
echo "📂 Listing /artifacts directory:"
./bin/port42 ls /artifacts

echo ""
echo "📂 Listing /artifacts/document directory (if exists):"
./bin/port42 ls /artifacts/document

echo ""
echo "5️⃣ Checking daemon logs for artifact creation..."
echo ""
echo "Recent daemon logs:"
sudo tail -n 50 /var/log/port42d.log | grep -E "(artifact|Artifact)"

echo ""
echo "6️⃣ Cleaning up..."
sudo kill $DAEMON_PID 2>/dev/null || true

echo ""
echo "✅ Test complete!"