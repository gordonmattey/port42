#!/bin/bash
# Test script for Phase 4 - Error Messages with Reality Compiler Language

set -e

echo "🧪 Testing Phase 4: Error Messages"
echo "================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PORT42="./target/debug/port42"

# Build first
echo "Building port42..."
cargo build --bin port42 2>/dev/null || {
    echo -e "${RED}Failed to build port42${NC}"
    exit 1
}

echo -e "\n${YELLOW}Test 1: Memory search without query${NC}"
echo "Command: port42 memory search"
$PORT42 memory search 2>&1 || true
echo "Expected: Should show reality compiler language error"

echo -e "\n${YELLOW}Test 2: Invalid agent error${NC}"
echo "Command: port42 possess @unknown-agent"
# This would need agent validation to be implemented
# $PORT42 possess @unknown-agent 2>&1 || true

echo -e "\n${YELLOW}Test 3: Daemon not running${NC}"
echo "Command: port42 status (with daemon stopped)"
# First ensure daemon is stopped
pkill -f port42d 2>/dev/null || true
sleep 1
$PORT42 status 2>&1 || true
echo "Expected: Should show dormant gateway message with helpful start command"

echo -e "\n${YELLOW}Test 4: Help with -help flag${NC}"
echo "Command: port42 possess -help"
$PORT42 possess -help 2>&1 || true
echo "Expected: Should show reality compiler help, not built-in help"

echo -e "\n${GREEN}✅ Error message tests complete${NC}"
echo -e "\n${YELLOW}Manual Testing Instructions:${NC}"
echo "1. Try invalid dates: port42 search 'test' --after 'not-a-date'"
echo "2. Try missing API key: unset ANTHROPIC_API_KEY && port42 daemon start"
echo "3. Try invalid paths: port42 cat /nonexistent/path"
echo "4. Try connection errors: Stop daemon and run possess command"