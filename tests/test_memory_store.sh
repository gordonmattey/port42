#!/bin/bash

# Test script for memory store with object storage

set -e

# Get the script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

echo "🧪 Testing Memory Store with Object Storage"
echo "=========================================="

# Run Go tests for memory store
echo ""
echo "🧪 Running Go unit tests for memory store..."
cd "$PROJECT_ROOT/daemon"

# Run the tests
go test -v -run TestMemoryStoreWithObjectStore

# Also run tests for helper functions
echo ""
echo "🧪 Running additional memory store tests..."
go test -v -run TestCleanAgentName
go test -v -run TestMapStateToLifecycle

echo ""
echo "✨ All memory store tests completed successfully!"
echo ""
echo "💡 The memory store now uses content-addressed object storage with:"
echo "   - SHA256 content addressing"
echo "   - Rich metadata with tags and lifecycle states"
echo "   - Virtual paths for organization (by-date, by-agent, etc.)"
echo "   - Backward compatibility for legacy sessions"