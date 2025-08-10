#!/bin/bash

# Test Phase 1: Reference Protocol Foundation
# Tests the basic reference syntax acceptance and validation

set -e

echo "🧪 Testing Phase 1: Reference Protocol Foundation"
echo "================================================="

# Build daemon and CLI
echo "📦 Building components..."
go build -o ../bin/port42d
cd ../cli && cargo build --release && cp target/release/port42 ../bin/ && cd ../daemon

# Start daemon in background
echo "🚀 Starting daemon..."
../bin/port42d &
DAEMON_PID=$!
sleep 2

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up..."
    if [ ! -z "$DAEMON_PID" ]; then
        kill $DAEMON_PID 2>/dev/null || true
    fi
    # Clean up test files
    rm -f ~/.port42/relations/test-*.json 2>/dev/null || true
}
trap cleanup EXIT

echo "📋 Test 1: Basic reference syntax acceptance"
echo "  Testing CLI --ref argument parsing and daemon acceptance..."

# Test single reference
../bin/port42 declare tool test-ref-single --transforms test --ref search:"test query"
if [ $? -eq 0 ]; then
    echo "  ✅ Single reference accepted"
else
    echo "  ❌ Single reference failed"
    exit 1
fi

echo "📋 Test 2: Multiple references"
echo "  Testing multiple --ref arguments..."

# Test multiple references
../bin/port42 declare tool test-ref-multiple --transforms test --ref search:"nginx errors" --ref tool:log-parser --ref file:error.log
if [ $? -eq 0 ]; then
    echo "  ✅ Multiple references accepted"
else
    echo "  ❌ Multiple references failed"
    exit 1
fi

echo "📋 Test 3: Invalid reference format"
echo "  Testing error handling for invalid reference syntax..."

# Test invalid reference format (should fail)
../bin/port42 declare tool test-ref-invalid --transforms test --ref "invalid-format" 2>/dev/null
if [ $? -ne 0 ]; then
    echo "  ✅ Invalid reference format properly rejected"
else
    echo "  ❌ Invalid reference format should have been rejected"
    exit 1
fi

echo "📋 Test 4: Backward compatibility"
echo "  Testing that declarations work without references..."

# Test backward compatibility (no references)
../bin/port42 declare tool test-no-refs --transforms test
if [ $? -eq 0 ]; then
    echo "  ✅ Backward compatibility maintained"
else
    echo "  ❌ Backward compatibility broken"
    exit 1
fi

echo "📋 Test 5: Reference types validation"
echo "  Testing all valid reference types..."

# Test all valid reference types
../bin/port42 declare tool test-all-types --transforms test \
  --ref search:"query" \
  --ref tool:some-tool \
  --ref memory:session-123 \
  --ref file:config.json \
  --ref url:https://example.com

if [ $? -eq 0 ]; then
    echo "  ✅ All reference types accepted"
else
    echo "  ❌ Some reference types failed"
    exit 1
fi

echo "📋 Test 6: CLI help text includes references"
echo "  Checking that --ref appears in help..."

../bin/port42 declare tool --help | grep -q "ref"
if [ $? -eq 0 ]; then
    echo "  ✅ --ref documented in help text"
else
    echo "  ❌ --ref not documented in help text"
    exit 1
fi

echo ""
echo "🎉 All Phase 1 Reference Protocol Foundation tests passed!"
echo "✅ Reference syntax parsing working"
echo "✅ Multiple references supported"
echo "✅ Error handling for invalid formats"
echo "✅ Backward compatibility maintained"
echo "✅ All reference types accepted"
echo "✅ CLI help documentation updated"
echo ""
echo "Phase 1: Reference Protocol Foundation complete! 🐬"
echo ""
echo "Next steps:"
echo "  - Phase 2: Reference Resolvers (search, tool, file, memory, url)"
echo "  - Phase 3: Context synthesis and AI integration"