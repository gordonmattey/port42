#!/bin/bash

# Test Tool Handler Implementation
set -e

echo "🔧 Testing Tool Handler Implementation"
echo "====================================="

# Build components
echo "📦 Building..."
go build -o ../bin/port42d
cd ../cli && cargo build --release && cp target/release/port42 ../bin/ && cd ../daemon

# Start daemon (skip if already running)
echo "🚀 Starting daemon (if not already running)..."
if ! ../bin/port42 status >/dev/null 2>&1; then
    ../bin/port42d &
    DAEMON_PID=$!
    sleep 3
    echo "  Daemon started"
else
    echo "  Daemon already running"
fi

cleanup() {
    if [ ! -z "$DAEMON_PID" ]; then
        kill $DAEMON_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT

echo ""
echo "📋 Step 1: Create a test tool to reference later"
../bin/port42 declare tool test-parser --transforms "parsing,logs" > /dev/null

echo ""
echo "📋 Step 2: Create another tool that references the first one"
echo "  Using --ref tool:test-parser to test tool handler..."

# Get current log position before the command
LOG_BEFORE=$(wc -l ~/.port42/daemon.log | awk '{print $1}')

../bin/port42 declare tool advanced-parser --transforms "advanced parsing" \
  --ref tool:test-parser

# Get logs from after the command started
LOG_AFTER=$(wc -l ~/.port42/daemon.log | awk '{print $1}')
LINES_TO_CHECK=$((LOG_AFTER - LOG_BEFORE))

echo ""
echo "🔍 Checking daemon logs for tool handler activity (last $LINES_TO_CHECK lines):"
echo "============================================================================"

# Check for tool handler calls in the new log lines
if tail -n $LINES_TO_CHECK ~/.port42/daemon.log | grep -q "🔧 Tool handler called"; then
    echo "✅ Tool handler was called:"
    tail -n $LINES_TO_CHECK ~/.port42/daemon.log | grep "🔧 Tool handler called"
    
    if tail -n $LINES_TO_CHECK ~/.port42/daemon.log | grep -q "✅ Tool found"; then
        echo "✅ Tool was successfully found:"
        tail -n $LINES_TO_CHECK ~/.port42/daemon.log | grep "✅ Tool found"
        echo ""
        echo "🎯 TOOL HANDLER WORKING CORRECTLY!"
    else
        echo "❌ Tool handler called but tool not found"
        tail -n $LINES_TO_CHECK ~/.port42/daemon.log | grep "⚠️ Tool.*not found" || echo "No 'not found' message"
    fi
else
    echo "❌ Tool handler was not called - check reference parsing"
fi

echo ""
echo "📊 Resolution Statistics:"
if tail -n $LINES_TO_CHECK ~/.port42/daemon.log | grep -q "Resolution stats:"; then
    tail -n $LINES_TO_CHECK ~/.port42/daemon.log | grep "Resolution stats:" | tail -1
else
    echo "No resolution stats found in the new log lines"
fi

echo ""
echo "Done! Tool handler test complete."