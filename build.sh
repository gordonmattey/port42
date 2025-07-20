#!/bin/bash
# Build script for Port 42

echo "🔨 Building Port 42..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build daemon
echo "Building daemon..."
cd daemon && go build -o ../bin/port42d . && cd ..

# TODO: Build Rust CLI when ready
# echo "Building CLI..."
# cd cli && cargo build --release && cp target/release/port42 ../bin/ && cd ..

echo "✅ Build complete! Binaries in ./bin/"
echo ""
echo "To run the daemon:"
echo "  sudo -E ./bin/port42d"