#!/bin/bash
# Test storage implementation

echo "Running storage tests..."

# Get to tests directory
cd "$(dirname "$0")"

# Copy test file to daemon directory temporarily
cp storage_test.go ../daemon/storage_test.go

# Run tests
cd ../daemon
go test -v -run TestStorage

# Clean up
rm storage_test.go

echo "Storage tests complete"