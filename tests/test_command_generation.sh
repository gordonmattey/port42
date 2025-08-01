#!/bin/bash

echo "🐬 Testing Command Generation with Object Store..."
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Change to daemon directory
cd daemon

# Run command generation tests
echo -e "${BLUE}Running command generation tests...${NC}"
go test -v -run TestCommandGeneration

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✅ Command generation tests passed!${NC}"
else
    echo -e "\n${RED}❌ Command generation tests failed${NC}"
    exit 1
fi

# Run extract tags test
echo -e "\n${BLUE}Running tag extraction tests...${NC}"
go test -v -run TestExtractTags

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Tag extraction tests passed!${NC}"
else
    echo -e "${RED}❌ Tag extraction tests failed${NC}"
    exit 1
fi

# Build daemon to ensure everything compiles
echo -e "\n${BLUE}Building daemon with updated command generation...${NC}"
go build -o ../bin/port42d

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Daemon builds successfully!${NC}"
else
    echo -e "${RED}❌ Failed to build daemon${NC}"
    exit 1
fi

echo -e "\n${GREEN}🎉 Command generation now uses object store!${NC}"
echo
echo "What changed:"
echo "✨ Commands are stored in content-addressed object store"
echo "✨ Rich metadata with tags, paths, and relationships"
echo "✨ Multiple virtual paths for each command"
echo "✨ No more filesystem writes (pure object store)"
echo
echo "Note: Commands won't be executable until FUSE is implemented."