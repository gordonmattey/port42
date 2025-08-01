#!/bin/bash

echo "🐬 Testing Port 42 Object Store Implementation..."
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Create a temporary directory for the test
TEMP_DIR=$(mktemp -d)
echo "Working in temp directory: $TEMP_DIR"

# Copy necessary files to temp directory
cp tests/test_object_store.go "$TEMP_DIR/"
cp daemon/object_store.go "$TEMP_DIR/"

# Build test program in temp directory
echo "Building test program..."
cd "$TEMP_DIR"
# First, update the import in test file to not use external package
sed -i '' 's/daemon "github.com\/port42\/port42\/daemon"//' test_object_store.go
sed -i '' 's/daemon\.//' test_object_store.go

go build -o test_object_store test_object_store.go object_store.go
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to build test program${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Run tests
echo "Running object store tests..."
./test_object_store

# Capture exit code
EXIT_CODE=$?

# Cleanup
cd - > /dev/null
rm -rf "$TEMP_DIR"

if [ $EXIT_CODE -eq 0 ]; then
    echo
    echo -e "${GREEN}✅ Object store implementation complete!${NC}"
else
    echo
    echo -e "${RED}❌ Tests failed${NC}"
    exit $EXIT_CODE
fi