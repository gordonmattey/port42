#!/bin/bash
# Port42 Web Installer - macOS First
# This script is served at https://port42.ai/install
set -e

# When piped through curl, we need to handle stdin differently
# Save the script and re-execute it with proper terminal access
if [ -z "${BASH_SOURCE[0]}" ] || [ "${BASH_SOURCE[0]}" = "-" ]; then
    # We're being piped - save and re-run
    TEMP_SCRIPT=$(mktemp /tmp/port42-installer.XXXXXX)
    cat > "$TEMP_SCRIPT"
    chmod +x "$TEMP_SCRIPT"
    bash "$TEMP_SCRIPT"
    rm -f "$TEMP_SCRIPT"
    exit $?
fi

echo "🐬 Port42 Installer"
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ Pre-built binaries currently only available for macOS"
    echo "🔨 Building from source for your platform..."
    echo ""
    echo "Downloading installer..."
    curl -L https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh | bash -s -- --build
    exit 0
fi

# Detect Mac architecture
ARCH=$(uname -m)
case $ARCH in
    arm64) 
        PLATFORM="darwin-aarch64"
        echo "📱 Detected: macOS Apple Silicon (M1/M2/M3)"
        ;;
    x86_64) 
        PLATFORM="darwin-x86_64"
        echo "📱 Detected: macOS Intel"
        ;;
    *) 
        echo "❌ Unsupported Mac architecture: $ARCH"
        echo "🔨 Building from source..."
        curl -L https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh | bash -s -- --build
        exit 0
        ;;
esac

# Check if binaries exist for this platform
# Get version from version.txt
VERSION=$(curl -s "https://raw.githubusercontent.com/gordonmattey/port42/main/version.txt" 2>/dev/null || echo "0.0.9")

# Try versioned repo file first, then GitHub releases
VERSIONED_BINARY_URL="https://raw.githubusercontent.com/gordonmattey/port42/main/releases/port42-${PLATFORM}-v${VERSION}.tar.gz"
RELEASE_BINARY_URL="https://github.com/gordonmattey/port42/releases/latest/download/port42-${PLATFORM}.tar.gz"

echo "🔍 Checking for pre-built binaries (v${VERSION})..."

# Check versioned file first
if curl -sI "$VERSIONED_BINARY_URL" | head -n 1 | grep -q "200\|302"; then
    echo "✅ Pre-built binaries available for $PLATFORM (v${VERSION})"
    INSTALL_METHOD="binary"
    BINARY_URL="$VERSIONED_BINARY_URL"
# Then check GitHub releases
elif curl -sI "$RELEASE_BINARY_URL" | head -n 1 | grep -q "200\|302"; then
    echo "✅ Pre-built binaries available for $PLATFORM"
    INSTALL_METHOD="binary"
    BINARY_URL="$RELEASE_BINARY_URL"
else
    echo "⚠️  No pre-built binaries available yet for $PLATFORM"
    echo "🔨 Will build from source instead..."
    INSTALL_METHOD="source"
fi

echo ""
echo "📥 Downloading Port42 installer..."

# Download the main installer
curl -fsSL https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh -o /tmp/port42-install.sh
chmod +x /tmp/port42-install.sh

# Run installer interactively
if [ "$INSTALL_METHOD" = "binary" ]; then
    echo "🚀 Pre-built binaries are available for $PLATFORM"
    echo ""
    /tmp/port42-install.sh
else
    echo "🔨 Building Port42 from source..."
    echo ""
    /tmp/port42-install.sh --build
fi

# Clean up
rm -f /tmp/port42-install.sh

echo ""
echo "🎉 Installation complete!"
echo ""
echo "Next steps:"
echo "1. Reload your shell: source ~/.zshrc (or ~/.bashrc)"
echo "2. Start the daemon: port42 daemon start -b"
echo "3. Test it: port42 status"
echo ""
echo "🐬 Welcome to Port42!"