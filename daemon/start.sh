#!/bin/bash

echo "🐬 Starting Port 42 daemon..."

# Check if we can bind to port 42 without sudo
if nc -z localhost 42 2>/dev/null; then
    echo "⚠️  Port 42 is already in use"
    exit 1
fi

# Try to run without sudo first
go run main.go 2>&1 | grep -q "permission denied"
if [ $? -eq 0 ]; then
    echo "🔐 Port 42 requires elevated permissions"
    echo "🐬 Requesting sudo access to open Port 42..."
    sudo go run main.go
else
    go run main.go
fi