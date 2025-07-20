#!/bin/bash
# Build script for Port 42

set -euo pipefail

echo "🔨 Building Port 42..."

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Create bin directory if it doesn't exist
mkdir -p bin

# Build daemon
echo -e "${BLUE}Building Go daemon...${NC}"
if cd daemon && go build -o ../bin/port42d .; then
    cd ..
    echo -e "${GREEN}✅ Daemon built successfully${NC}"
else
    cd ..
    echo -e "${RED}❌ Failed to build daemon${NC}"
    exit 1
fi

# Build Rust CLI
echo -e "${BLUE}Building Rust CLI...${NC}"
if cd cli && cargo build --release; then
    cp target/release/port42 ../bin/
    cd ..
    echo -e "${GREEN}✅ CLI built successfully${NC}"
else
    cd ..
    echo -e "${RED}❌ Failed to build CLI${NC}"
    exit 1
fi

echo
echo -e "${GREEN}✅ Build complete!${NC}"
echo -e "${BLUE}Binaries created:${NC}"
echo "  - ./bin/port42d (daemon)"
echo "  - ./bin/port42  (CLI)"
echo
echo -e "${BLUE}To test locally:${NC}"
echo "  1. Start daemon: sudo -E ./bin/port42d"
echo "  2. Use CLI: ./bin/port42 status"
echo
echo -e "${BLUE}To install system-wide:${NC}"
echo "  ./install.sh"