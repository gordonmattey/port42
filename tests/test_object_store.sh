#!/bin/bash

echo "🐬 Testing Port 42 Object Store Implementation..."
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Build test program
echo "Building test program..."
cd tests
go build -o test_object_store test_object_store.go ../daemon/object_store.go
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to build test program${NC}"
    exit 1
fi

# Run tests
echo "Running object store tests..."
./test_object_store

# Cleanup
rm -f test_object_store

echo
echo -e "${GREEN}✅ Object store implementation complete!${NC}"