# Port42 Release Process - macOS First Approach

## Overview
Implement a professional release process starting with macOS (developer platform), then expanding to other platforms based on user demand. Web-based installer accessible via `curl -L https://port42.ai/install | bash`.

## Strategy: Start with macOS, Expand Incrementally

### Why macOS First?
- ✅ **Your development platform** - can test thoroughly locally
- ✅ **Target audience** - many developers use Macs  
- ✅ **Easier debugging** - issues happen on your own machine
- ✅ **Proven process** - nail the experience before expanding
- ✅ **Lower risk** - single platform to start, expand based on demand

## Architecture

### 1. Two-Layer Install Process

#### Web Installer Script (hosted at `https://port42.ai/install`)
**Phase 1**: macOS-focused installer that:
- Detects if user is on macOS (Intel vs Apple Silicon)
- Downloads pre-built macOS binaries
- Falls back to source build for non-macOS platforms
- Downloads the main `install.sh` from GitHub
- Calls `install.sh` with appropriate flags

**Phase 1 - macOS Only Web Installer:**
```bash
#!/bin/bash
# Script served at https://port42.ai/install (Phase 1: macOS focus)
set -e

echo "🐬 Port42 Installer"

# Check if macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ Pre-built binaries currently only available for macOS"
    echo "🔨 Building from source for your platform..."
    curl -L https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh | bash -s -- --build
    exit 0
fi

# Detect Mac architecture  
ARCH=$(uname -m)
case $ARCH in
    arm64) PLATFORM="darwin-aarch64" ;;
    x86_64) PLATFORM="darwin-x86_64" ;;
    *) 
        echo "❌ Unsupported Mac architecture: $ARCH"
        echo "🔨 Building from source..."
        curl -L https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh | bash -s -- --build
        exit 0
        ;;
esac

echo "📱 Detected: macOS $ARCH"
echo "📥 Downloading installer..."

# Download the real installer
curl -L https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh -o /tmp/port42-install.sh
chmod +x /tmp/port42-install.sh

echo "🚀 Installing pre-built binaries for $PLATFORM..."
/tmp/port42-install.sh --download-binaries --platform "$PLATFORM"
```

**Future Phase 2+ - Multi-Platform Web Installer:**
```bash
#!/bin/bash
# Future expanded version with full platform support
set -e

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="x86_64" ;;
    arm64|aarch64) ARCH="aarch64" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac
PLATFORM="${OS}-${ARCH}"

echo "🐬 Port42 Universal Installer"
echo "Detected platform: $PLATFORM"

# Check if binaries exist for this platform
BINARY_URL="https://github.com/gordonmattey/port42/releases/latest/download/port42-${PLATFORM}.tar.gz"
if curl -s --head "$BINARY_URL" | head -n 1 | grep -q "200 OK"; then
    HAS_BINARIES=true
    echo "✅ Pre-built binaries available"
else
    HAS_BINARIES=false
    echo "⚠️  No pre-built binaries, will build from source"
fi

# Download and run installer
curl -L https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh -o /tmp/port42-install.sh
chmod +x /tmp/port42-install.sh

if [ "$HAS_BINARIES" = true ]; then
    /tmp/port42-install.sh --download-binaries --platform "$PLATFORM"
else
    /tmp/port42-install.sh --build
fi
```

#### Enhanced `install.sh` (in your repo)
Your existing `install.sh` gets new command-line options:
- `--download-binaries --platform <platform>` - Downloads and installs pre-built binaries
- `--build` - Forces build from source (current behavior)
- `--binaries <path>` - Uses local binary tarball

### 2. Target Platforms (Phased Rollout)

**Phase 1: macOS Focus** 
- `darwin-aarch64` (Apple Silicon Mac) - **PRIMARY**
- `darwin-x86_64` (Intel Mac)

**Phase 2: Add Linux** (Based on user demand)
- `linux-x86_64` (Intel/AMD Linux)
- `linux-aarch64` (ARM64 Linux, e.g., Raspberry Pi)

**Phase 3: Add Windows** (If requested)
- `windows-x86_64` (Windows 64-bit)

### 3. Release Workflow

#### GitHub Actions Release Pipeline

**Phase 1: macOS-Only Release**
```yaml
# .github/workflows/release.yml
name: Release macOS Binaries

on:
  release:
    types: [published]

jobs:
  build-macos:
    strategy:
      matrix:
        include:
          - os: macos-latest      # Apple Silicon runner
            target: darwin-aarch64
            go_os: darwin
            go_arch: arm64
          - os: macos-13          # Intel runner
            target: darwin-x86_64
            go_os: darwin
            go_arch: amd64
    
    runs-on: ${{ matrix.os }}
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Setup Rust
        uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
          profile: minimal
          override: true
          
      - name: Build binaries
        env:
          GOOS: ${{ matrix.go_os }}
          GOARCH: ${{ matrix.go_arch }}
        run: |
          ./build.sh --release
          
      - name: Package binaries
        run: |
          mkdir -p release
          tar -czf release/port42-${{ matrix.target }}.tar.gz bin/
          
      - name: Upload release assets
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ github.event.release.upload_url }}
          asset_path: release/port42-${{ matrix.target }}.tar.gz
          asset_name: port42-${{ matrix.target }}.tar.gz
          asset_content_type: application/gzip
```

**Future Phase 2+: Multi-Platform Release**
```yaml
# Future expanded version
name: Release Multi-Platform Binaries

on:
  release:
    types: [published]

jobs:
  build-cross-platform:
    strategy:
      matrix:
        include:
          # macOS
          - os: macos-latest
            target: darwin-aarch64
            go_os: darwin
            go_arch: arm64
          - os: macos-13
            target: darwin-x86_64
            go_os: darwin
            go_arch: amd64
          # Linux
          - os: ubuntu-latest
            target: linux-x86_64
            go_os: linux
            go_arch: amd64
          - os: ubuntu-latest  
            target: linux-aarch64
            go_os: linux
            go_arch: arm64
          # Windows
          - os: windows-latest
            target: windows-x86_64
            go_os: windows
            go_arch: amd64
    
    runs-on: ${{ matrix.os }}
    # ... (similar steps as Phase 1)
```

### 4. Enhanced Build Script

Update `build.sh` to support cross-compilation:

```bash
#!/bin/bash
# Add to existing build.sh

RELEASE_MODE=false
TARGET_PLATFORM=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --release)
            RELEASE_MODE=true
            shift
            ;;
        --target)
            TARGET_PLATFORM="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

if [ "$RELEASE_MODE" = true ] && [ -n "$TARGET_PLATFORM" ]; then
    echo "🚀 Building release binaries for $TARGET_PLATFORM"
    
    # Set cross-compilation environment
    case $TARGET_PLATFORM in
        linux-x86_64)
            export GOOS=linux
            export GOARCH=amd64
            export CARGO_TARGET=x86_64-unknown-linux-gnu
            ;;
        linux-aarch64)
            export GOOS=linux
            export GOARCH=arm64
            export CARGO_TARGET=aarch64-unknown-linux-gnu
            ;;
        darwin-x86_64)
            export GOOS=darwin
            export GOARCH=amd64
            export CARGO_TARGET=x86_64-apple-darwin
            ;;
        darwin-aarch64)
            export GOOS=darwin
            export GOARCH=arm64
            export CARGO_TARGET=aarch64-apple-darwin
            ;;
        windows-x86_64)
            export GOOS=windows
            export GOARCH=amd64
            export CARGO_TARGET=x86_64-pc-windows-gnu
            ;;
        *)
            echo "❌ Unsupported target: $TARGET_PLATFORM"
            exit 1
            ;;
    esac
fi

# Rest of your existing build script...
```

### 5. Binary Package Structure

Each `port42-<platform>.tar.gz` should contain:
```
bin/
├── port42      (CLI executable)
└── port42d     (Daemon executable)
```

### 6. Updated install.sh Integration

Add these functions to your existing `install.sh`:

```bash
# New command line options
DOWNLOAD_BINARIES=false
BUILD_FROM_SOURCE=false
PLATFORM=""
BINARY_PATH=""

# Parse new arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --download-binaries)
            DOWNLOAD_BINARIES=true
            shift
            ;;
        --build)
            BUILD_FROM_SOURCE=true
            shift
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --binaries)
            BINARY_PATH="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

download_and_install_binaries() {
    local platform="$1"
    local binary_url="https://github.com/gordonmattey/port42/releases/latest/download/port42-${platform}.tar.gz"
    
    echo "📥 Downloading binaries for $platform..."
    
    # Download to temp location
    local temp_file="/tmp/port42-binaries.tar.gz"
    if ! curl -L "$binary_url" -o "$temp_file"; then
        echo "❌ Failed to download binaries"
        exit 1
    fi
    
    # Extract binaries
    echo "📦 Extracting binaries..."
    tar -xzf "$temp_file" -C /tmp/
    
    # Install binaries (reuse existing install_binaries function)
    SCRIPT_DIR="/tmp"
    install_binaries
    
    # Cleanup
    rm -f "$temp_file"
    rm -rf /tmp/bin/
}

# Update main installation flow
if [ "$DOWNLOAD_BINARIES" = true ] && [ -n "$PLATFORM" ]; then
    download_and_install_binaries "$PLATFORM"
elif [ -n "$BINARY_PATH" ]; then
    # Use provided binary file
    tar -xzf "$BINARY_PATH" -C /tmp/
    SCRIPT_DIR="/tmp"
    install_binaries
elif [ "$BUILD_FROM_SOURCE" = true ]; then
    # Force build from source (current behavior)
    build_from_source
else
    # Default behavior (current logic)
    # ... existing installation logic
fi
```

## User Experience

### Simple Installation
```bash
# One-line install (detects platform, downloads binaries or builds)
curl -L https://port42.ai/install | bash

# Or save and inspect first
curl -L https://port42.ai/install -o install-port42.sh
chmod +x install-port42.sh
./install-port42.sh
```

### Advanced Options
```bash
# Force build from source
curl -L https://port42.ai/install | bash -s -- --build

# Specify platform manually
curl -L https://port42.ai/install | bash -s -- --platform linux-aarch64
```

## Implementation Plan - One Hour Sprint 🚀

### Rapid Implementation Checklist
**Goal**: Get macOS binary releases working end-to-end in one hour

#### Step 1: Enhance install.sh (15 minutes)
- [ ] Add `--download-binaries --platform <platform>` flag
- [ ] Add `--binaries <path>` flag for local testing
- [ ] Keep existing functionality intact

#### Step 2: Local Testing (15 minutes)
- [ ] Build: `./build.sh && tar -czf port42-darwin-aarch64.tar.gz bin/`
- [ ] Test: `./install.sh --binaries port42-darwin-aarch64.tar.gz`
- [ ] Verify daemon starts and CLI works

#### Step 3: GitHub Actions (15 minutes)
- [ ] Create `.github/workflows/release.yml` (macOS-only)
- [ ] Test with a draft release
- [ ] Verify binary builds and uploads

#### Step 4: Web Installer (10 minutes)
- [ ] Create simple macOS-focused web installer script
- [ ] Test locally by serving the script

#### Step 5: End-to-End Test (5 minutes)
- [ ] Manual release with binary upload
- [ ] Test web installer downloads the binary
- [ ] Verify complete installation flow

### Post-Sprint Expansion (Later)
- Add Intel Mac support
- Create automated CI/CD
- Add other platforms based on demand
- Host web installer at production URL

## Benefits

1. **Professional UX**: Single `curl | bash` command for installation
2. **Fast Installation**: Pre-built binaries for common platforms
3. **Universal Fallback**: Builds from source when binaries unavailable
4. **Cross-Platform**: Supports major operating systems and architectures
5. **Automated Releases**: GitHub Actions handles the build pipeline
6. **Maintainable**: Minimal web installer, main logic stays in repo

## Repository URLs

- **GitHub Repo**: `https://github.com/gordonmattey/port42`
- **Raw Install Script**: `https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh`
- **Release Binaries**: `https://github.com/gordonmattey/port42/releases/latest/download/port42-<platform>.tar.gz`
- **Web Installer**: `https://port42.ai/install` (to be hosted)