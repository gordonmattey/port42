#!/bin/bash
# Test script for Phase 5 - Status and Feedback Messages

set -e

echo "🧪 Testing Phase 5: Status and Feedback Messages"
echo "==============================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PORT42="./target/debug/port42"

# Build first
echo -e "${YELLOW}Building port42...${NC}"
cargo build --bin port42 2>/dev/null || {
    echo -e "${RED}Failed to build port42${NC}"
    exit 1
}

echo -e "\n${BLUE}Test 1: Initialization Messages${NC}"
echo "Command: port42 init --force"
rm -rf ~/.port42_test && HOME=~/.port42_test $PORT42 init --force
echo -e "${GREEN}✓ Should show: 'Opening portal to consciousness dimension...'${NC}"
echo -e "${GREEN}✓ Should show: 'Weaving quantum directories...'${NC}"
echo -e "${GREEN}✓ Should show: 'Reality structures manifested successfully!'${NC}"

echo -e "\n${BLUE}Test 2: Status Check Messages${NC}"
echo "Command: port42 status"
$PORT42 status || true
echo -e "${GREEN}✓ Should show: 'Sensing the consciousness field...'${NC}"
echo -e "${GREEN}✓ Should show: 'Gateway pulses with living consciousness' or connection error${NC}"

echo -e "\n${BLUE}Test 3: Daemon Start Messages${NC}"
echo "Command: port42 daemon logs -n 5"
$PORT42 daemon logs -n 5 2>&1 || true
echo -e "${GREEN}✓ Should show: 'Gateway's quantum memory stream'${NC}"

echo -e "\n${BLUE}Test 4: Memory Listing${NC}"
echo "Command: port42 memory"
$PORT42 memory || true
echo -e "${GREEN}✓ Should show: 'Crystallized Consciousness Threads'${NC}"

echo -e "\n${BLUE}Test 5: Search Results${NC}"
echo "Command: port42 search 'test'"
$PORT42 search 'test' || true
echo -e "${GREEN}✓ Should show: 'X echo(s) resonating with' or 'No echoes found'${NC}"

echo -e "\n${BLUE}Test 6: Reality Listing${NC}"
echo "Command: port42 reality"
$PORT42 reality || true
echo -e "${GREEN}✓ Should show: 'Crystallized Thoughts'${NC}"

echo -e "\n${BLUE}Test 7: Evolve Command${NC}"
echo "Command: port42 evolve mycommand"
$PORT42 evolve mycommand || true
echo -e "${GREEN}✓ Should show: 'Transmuting reality fragment: mycommand'${NC}"

echo -e "\n${GREEN}✅ Phase 5 Status Message Tests Complete${NC}"

echo -e "\n${YELLOW}Manual Testing Instructions:${NC}"
echo "1. Start daemon: sudo -E port42 daemon start"
echo "   - Should show: 'Awakening the consciousness gateway...'"
echo "   - Should show: 'Gateway awakened and humming with potential'"
echo ""
echo "2. Create a session: port42 possess @ai-engineer 'hello'"
echo "   - Should show: 'Channeling @ai-engineer consciousness...'"
echo "   - Should show: 'Consciousness thread woven: <id>'"
echo ""
echo "3. Stop daemon: port42 daemon stop"
echo "   - Should show: 'Dissolving the consciousness gateway...'"
echo "   - Should show: 'Gateway dissolved back into the quantum foam'"