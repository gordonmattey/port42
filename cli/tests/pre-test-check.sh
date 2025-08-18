#!/bin/bash

# Pre-Test Environment Check
# Validates that everything is ready for manual testing

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔍 Port 42 Pre-Test Environment Check${NC}"
echo "=================================="

# Find project structure
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_DIR="$(dirname "$TEST_DIR")"
ROOT_DIR="$(dirname "$CLI_DIR")"

echo -e "${BLUE}📁 Project Structure:${NC}"
echo "   Root: $ROOT_DIR"
echo "   CLI:  $CLI_DIR"
echo "   Test: $TEST_DIR"
echo ""

# Check binaries
echo -e "${BLUE}🔧 Binary Check:${NC}"
PORT42_BIN="$ROOT_DIR/bin/port42"
DAEMON_BIN="$ROOT_DIR/bin/port42d"

if [[ -f "$PORT42_BIN" ]]; then
    echo -e "   ${GREEN}✅ CLI binary found: $PORT42_BIN${NC}"
else
    echo -e "   ${RED}❌ CLI binary missing: $PORT42_BIN${NC}"
    echo -e "   ${YELLOW}   Build with: cd $ROOT_DIR && ./build.sh${NC}"
fi

if [[ -f "$DAEMON_BIN" ]]; then
    echo -e "   ${GREEN}✅ Daemon binary found: $DAEMON_BIN${NC}"
else
    echo -e "   ${RED}❌ Daemon binary missing: $DAEMON_BIN${NC}"
    echo -e "   ${YELLOW}   Build with: cd $ROOT_DIR && ./build.sh${NC}"
fi

echo ""

# Check test files
echo -e "${BLUE}📋 Test Suite Check:${NC}"
TEST_SUITE="$TEST_DIR/manual-test-suite.sh"
TEST_RUNNER="$TEST_DIR/run-manual-tests.sh"

if [[ -f "$TEST_SUITE" ]]; then
    echo -e "   ${GREEN}✅ Manual test suite: $TEST_SUITE${NC}"
else
    echo -e "   ${RED}❌ Manual test suite missing${NC}"
fi

if [[ -f "$TEST_RUNNER" ]]; then
    echo -e "   ${GREEN}✅ Test runner: $TEST_RUNNER${NC}"
else
    echo -e "   ${RED}❌ Test runner missing${NC}"
fi

if [[ -x "$TEST_SUITE" ]]; then
    echo -e "   ${GREEN}✅ Test suite is executable${NC}"
else
    echo -e "   ${YELLOW}⚠️  Making test suite executable${NC}"
    chmod +x "$TEST_SUITE" 2>/dev/null || echo -e "   ${RED}❌ Failed to make executable${NC}"
fi

echo ""

# Check environment
echo -e "${BLUE}🌍 Environment Check:${NC}"

if [[ -n "$ANTHROPIC_API_KEY" ]]; then
    echo -e "   ${GREEN}✅ ANTHROPIC_API_KEY is set${NC}"
else
    echo -e "   ${YELLOW}⚠️  ANTHROPIC_API_KEY not set (AI features will be limited)${NC}"
fi

if [[ -n "$PORT42_DEBUG" ]]; then
    echo -e "   ${GREEN}✅ PORT42_DEBUG is set: $PORT42_DEBUG${NC}"
else
    echo -e "   ${YELLOW}ℹ️  PORT42_DEBUG not set (will be enabled by test suite)${NC}"
fi

echo ""

# Check daemon status
echo -e "${BLUE}🐬 Daemon Status:${NC}"
if command -v "$PORT42_BIN" >/dev/null 2>&1; then
    if "$PORT42_BIN" status >/dev/null 2>&1; then
        echo -e "   ${GREEN}✅ Daemon is running${NC}"
        DAEMON_INFO=$("$PORT42_BIN" status 2>/dev/null)
        echo "$DAEMON_INFO" | sed 's/^/   /'
    else
        echo -e "   ${YELLOW}⚠️  Daemon is not running (will be started by tests)${NC}"
        echo -e "   ${BLUE}   To start manually: $PORT42_BIN daemon start${NC}"
    fi
else
    echo -e "   ${RED}❌ Cannot check daemon (CLI not available)${NC}"
fi

echo ""

# Check network connectivity (for URL reference tests)
echo -e "${BLUE}🌐 Network Check:${NC}"
if ping -c 1 httpbin.org >/dev/null 2>&1; then
    echo -e "   ${GREEN}✅ Network connectivity (httpbin.org reachable)${NC}"
else
    echo -e "   ${YELLOW}⚠️  Network connectivity limited (URL reference tests may fail)${NC}"
fi

echo ""

# Check test data directory
echo -e "${BLUE}📂 Test Data Directory:${NC}"
TEST_DATA_DIR="$TEST_DIR/test-data"
if [[ -d "$TEST_DATA_DIR" ]]; then
    echo -e "   ${GREEN}✅ Test data directory exists: $TEST_DATA_DIR${NC}"
    FILE_COUNT=$(find "$TEST_DATA_DIR" -type f 2>/dev/null | wc -l)
    echo -e "   ${BLUE}   Contains $FILE_COUNT files${NC}"
else
    echo -e "   ${YELLOW}ℹ️  Test data directory will be created by test suite${NC}"
fi

echo ""

# Summary and recommendations
echo -e "${BLUE}📝 Summary:${NC}"

READY=true

if [[ ! -f "$PORT42_BIN" ]] || [[ ! -f "$DAEMON_BIN" ]]; then
    echo -e "   ${RED}❌ Binaries missing - build required${NC}"
    READY=false
fi

if [[ ! -f "$TEST_SUITE" ]]; then
    echo -e "   ${RED}❌ Test suite missing${NC}"
    READY=false
fi

if [[ "$READY" == "true" ]]; then
    echo -e "   ${GREEN}🎉 Environment ready for testing!${NC}"
    echo ""
    echo -e "${BLUE}🚀 To run tests:${NC}"
    echo -e "   ${GREEN}./run-manual-tests.sh${NC}           # Run all tests"
    echo -e "   ${GREEN}./run-manual-tests.sh basic${NC}     # Basic functionality"
    echo -e "   ${GREEN}./run-manual-tests.sh references${NC} # Reference system"
    echo -e "   ${GREEN}./run-manual-tests.sh advanced${NC}   # Full integration"
    echo ""
    echo -e "${BLUE}📖 For debugging help:${NC}"
    echo -e "   ${GREEN}cat DEBUG_GUIDE.md${NC}"
else
    echo -e "   ${RED}⚠️  Environment not ready - fix issues above first${NC}"
    echo ""
    echo -e "${BLUE}🔧 Common fixes:${NC}"
    echo -e "   ${YELLOW}cd $ROOT_DIR && ./build.sh${NC}  # Build binaries"
    echo -e "   ${YELLOW}export ANTHROPIC_API_KEY='your-key'${NC}  # Set API key"
fi

echo ""