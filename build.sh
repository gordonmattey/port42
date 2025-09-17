#!/bin/bash
# Build script for Port 42

set -euo pipefail

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Create bin directory if it doesn't exist
mkdir -p bin

# Build daemon
echo -e "${BLUE}Building Go daemon...${NC}"
# Run go mod tidy first to ensure dependencies are up to date
if cd daemon/src && go mod tidy >/dev/null 2>&1; then
    if go build -o ../../bin/port42d .; then
        cd ../..
        echo -e "${GREEN}✅ Daemon built successfully${NC}"
    else
        cd ../..
        echo -e "${RED}❌ Failed to build daemon${NC}"
        exit 1
    fi
else
    # Try to build anyway - go mod tidy might fail but build might work
    if go build -o ../../bin/port42d .; then
        cd ../..
        echo -e "${GREEN}✅ Daemon built successfully${NC}"
    else
        cd ../..
        echo -e "${RED}❌ Failed to build daemon${NC}"
        exit 1
    fi
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

# Detect platform for packaging
PLATFORM=""
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    darwin) 
        case "$ARCH" in
            arm64) PLATFORM="darwin-aarch64" ;;
            x86_64) PLATFORM="darwin-x86_64" ;;
            *) PLATFORM="darwin-$ARCH" ;;
        esac
        ;;
    linux)
        case "$ARCH" in
            x86_64) PLATFORM="linux-x86_64" ;;
            aarch64) PLATFORM="linux-aarch64" ;;
            *) PLATFORM="linux-$ARCH" ;;
        esac
        ;;
    *) 
        PLATFORM="$OS-$ARCH"
        ;;
esac

# Get version from version.txt or default
VERSION=$(cat version.txt 2>/dev/null || echo "0.1.0")

# Update Cargo.toml version to match version.txt
if [ -f "cli/Cargo.toml" ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS sed requires backup extension
        sed -i '' "s/^version = \".*\"/version = \"$VERSION\"/" cli/Cargo.toml
    else
        # Linux sed
        sed -i "s/^version = \".*\"/version = \"$VERSION\"/" cli/Cargo.toml
    fi
    echo -e "${BLUE}Updated Cargo.toml version to ${VERSION}${NC}"
fi

# Package binaries if requested or by default
PACKAGE_BINARIES=${PACKAGE:-true}
RELEASE_MODE=${RELEASE:-false}

if [ "$PACKAGE_BINARIES" = "true" ]; then
    echo
    echo -e "${BLUE}Creating release package (v${VERSION})...${NC}"
    
    # Create releases directory
    mkdir -p releases
    
    # Package name with version
    PACKAGE_NAME="port42-${PLATFORM}-v${VERSION}.tar.gz"
    # Also create a "latest" symlink
    LATEST_NAME="port42-${PLATFORM}.tar.gz"
    
    # Create tarball with binaries, config, Claude integration, and version
    # Include optional files if they exist
    files_to_package=(
        "bin/port42"
        "bin/port42d"
        "daemon/agents.json"
        "daemon/agent_guidance.md"
        "version.txt"
    )
    
    # Add optional documentation files if they exist
    [ -f "P42CLAUDE.md" ] && files_to_package+=("P42CLAUDE.md")
    [ -f "README.md" ] && files_to_package+=("README.md")
    
    if tar -czf "releases/${PACKAGE_NAME}" "${files_to_package[@]}" 2>/dev/null; then
        
        # Create symlink to latest
        cd releases
        ln -sf "${PACKAGE_NAME}" "${LATEST_NAME}"
        cd ..
        
        echo -e "${GREEN}✅ Release package created: releases/${PACKAGE_NAME}${NC}"
        echo -e "   Latest symlink: releases/${LATEST_NAME}"
        echo -e "   Version: ${VERSION}"
        echo -e "   Size: $(ls -lh releases/${PACKAGE_NAME} | awk '{print $5}')"
        
        # Show contents
        echo -e "${BLUE}   Contents:${NC}"
        tar -tzf "releases/${PACKAGE_NAME}" | sed 's/^/     - /'
    else
        echo -e "${RED}❌ Failed to create release package${NC}"
    fi
fi

echo
echo -e "${GREEN}✅ Build complete!${NC}"
echo -e "${BLUE}Binaries created:${NC}"
echo "  - ./bin/port42d (daemon)"
echo "  - ./bin/port42  (CLI)"
if [ "$PACKAGE_BINARIES" = "true" ] && [ -f "releases/${PACKAGE_NAME}" ]; then
    echo "  - ./releases/${PACKAGE_NAME} (release package v${VERSION})"
    echo "  - ./releases/${LATEST_NAME} (latest symlink)"
fi
echo
echo -e "${BLUE}To test locally:${NC}"
echo "  1. Start daemon: sudo -E ./bin/port42d"
echo "  2. Use CLI: ./bin/port42 status"
echo
echo -e "${BLUE}To install system-wide:${NC}"
echo "  ./install.sh"
if [ "$PACKAGE_BINARIES" = "true" ] && [ -f "releases/${PACKAGE_NAME}" ]; then
    echo
    echo -e "${BLUE}To test binary installation:${NC}"
    echo "  ./install.sh --binaries releases/${LATEST_NAME}"
    echo
    echo -e "${BLUE}To create GitHub release:${NC}"
    echo "  gh release create v${VERSION} releases/${PACKAGE_NAME} \\"
    echo "    --title \"Port42 v${VERSION}\" \\"
    echo "    --notes \"Release v${VERSION} with ${PLATFORM} binaries\""
    echo
    echo -e "${BLUE}To bump version:${NC}"
    echo "  echo '0.2.0' > version.txt && ./build.sh"
fi