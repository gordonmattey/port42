#!/bin/bash
# Quick test to verify test infrastructure and current functionality
set -e

echo "🧪 Quick Infrastructure Test"
echo "============================"

# Test daemon is working
echo "📡 Testing daemon connectivity..."
if port42 status >/dev/null 2>&1; then
    echo "✅ Daemon is responding"
else
    echo "❌ Daemon not responding"
    exit 1
fi

# Test basic CLI commands work
echo "🔧 Testing basic CLI commands..."

# Test help commands
if port42 --help >/dev/null 2>&1; then
    echo "✅ Main help works"
else
    echo "❌ Main help failed"
    exit 1
fi

if port42 declare tool --help >/dev/null 2>&1; then
    echo "✅ Tool help works"
else
    echo "❌ Tool help failed"  
    exit 1
fi

if port42 declare artifact --help >/dev/null 2>&1; then
    echo "✅ Artifact help works"
else
    echo "❌ Artifact help failed"
    exit 1
fi

# Test VFS commands
if port42 ls /tools >/dev/null 2>&1; then
    echo "✅ VFS navigation works"
else
    echo "❌ VFS navigation failed"
    exit 1
fi

# Test rule engine is active
if port42 watch rules | head -5 | grep -E "(enabled|Status)" >/dev/null 2>&1; then
    echo "✅ Rule engine is active"
else
    echo "❌ Rule engine not active"
    exit 1
fi

echo ""
echo "🎯 Test Infrastructure Summary:"
echo "✅ Daemon connectivity: Working"
echo "✅ CLI commands: Working"  
echo "✅ VFS navigation: Working"
echo "✅ Rule engine: Active"
echo ""
echo "✅ Test infrastructure is operational!"
echo "Ready to proceed with Step 1 implementation."