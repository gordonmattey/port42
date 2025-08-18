#!/bin/bash

set -e

TEST_DATA_DIR="test-data"

echo "Testing file creation logic..."

# Test 1 - this works
echo "Test 1: First file check"
if [[ ! -f "$TEST_DATA_DIR/test-config.json" ]]; then
    echo "Would create test-config.json"
else
    echo "test-config.json already exists"
fi

# Test 2 - this might hang
echo "Test 2: Second file check" 
if [[ ! -f "$TEST_DATA_DIR/sample-data.csv" ]]; then
    echo "Creating sample-data.csv..."
    cat > "$TEST_DATA_DIR/sample-data.csv" << 'EOF'
name,age,city,email
John Doe,30,New York,john@example.com
EOF
    echo "Created sample-data.csv"
else
    echo "sample-data.csv already exists"
fi

echo "Test complete"