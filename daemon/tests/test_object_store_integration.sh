#!/bin/bash

echo "🐬 Testing Port 42 Object Store Integration..."
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Change to daemon directory
cd daemon

# Run Go tests
echo -e "${BLUE}Running object store tests...${NC}"
go test -v -run TestObjectStore

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✅ Object store tests passed!${NC}"
else
    echo -e "\n${RED}❌ Object store tests failed${NC}"
    exit 1
fi

# Build daemon to ensure integration compiles
echo -e "\n${BLUE}Building daemon with object store...${NC}"
go build -o ../bin/port42d

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Daemon builds successfully with object store!${NC}"
else
    echo -e "${RED}❌ Failed to build daemon${NC}"
    exit 1
fi

echo -e "\n${GREEN}🎉 Object store implementation complete and integrated!${NC}"
echo
echo "Next steps:"
echo "1. Update command generation to use object store"
echo "2. Update memory sessions to use object store"
echo "3. Implement protocol handlers for object operations"