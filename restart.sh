#!/bin/bash
# Port42 daemon restart script - can be installed as a system command

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Restarting Port42 daemon...${NC}"

# Step 1: Clean up hanging CLI processes
echo -e "${YELLOW}  Cleaning up CLI processes...${NC}"
pkill -f "port42 possess" 2>/dev/null
pkill -f "port42 shell" 2>/dev/null
pkill -f "port42 context --watch" 2>/dev/null

# Step 2: Stop daemon gracefully
echo -e "${YELLOW}  Stopping daemon...${NC}"
pkill -TERM port42d 2>/dev/null
sleep 1

# Step 3: Force kill if still running
if pgrep -f port42d > /dev/null; then
    echo -e "${YELLOW}  Force stopping daemon...${NC}"
    pkill -KILL port42d 2>/dev/null
    sleep 1
fi

# Step 4: Determine if we need to build (if in dev directory)
if [ -f "./build.sh" ]; then
    echo -e "${YELLOW}  Building from source...${NC}"
    ./build.sh
    
    # Install built binaries
    echo -e "${YELLOW}  Installing binaries...${NC}"
    cp bin/port42d ~/.port42/bin/ 2>/dev/null || true
    cp bin/port42 ~/.port42/bin/ 2>/dev/null || true
fi

# Step 5: Start daemon
echo -e "${YELLOW}  Starting daemon...${NC}"
if [ -x "$HOME/.port42/bin/port42d" ]; then
    $HOME/.port42/bin/port42d &
    sleep 2
    
    # Step 6: Verify it's running
    if port42 status &>/dev/null; then
        echo -e "${GREEN}✅ Port42 daemon restarted successfully${NC}"
        port42 status
    else
        echo -e "${RED}❌ Failed to start daemon${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Port42 daemon not found. Please run install.sh first${NC}"
    exit 1
fi