# Port42 Release Process

## Quick Start - macOS Binary Release

### 1. Test Locally (5 minutes)
```bash
# Build and package
./build.sh
tar -czf port42-darwin-aarch64.tar.gz bin/ daemon/agents.json

# Test installation
./install.sh --binaries port42-darwin-aarch64.tar.gz

# Verify
~/.port42/bin/port42 --version
```

### 2. Create GitHub Release

#### Option A: Manual Release (First Time)
1. Go to https://github.com/gordonmattey/port42/releases
2. Click "Draft a new release"
3. Create tag: `v0.1.0` (or next version)
4. Title: `Port42 v0.1.0 - macOS Release`
5. Upload `port42-darwin-aarch64.tar.gz`
6. Publish release

#### Option B: Automated Release (After Setup)
1. Create and push a tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
2. Go to GitHub and create release from tag
3. GitHub Actions will build and upload binaries automatically

### 3. Test Web Installer
```bash
# Test the web installer locally
./web-installer.sh

# Once hosted at port42.ai:
curl -L https://port42.ai/install | bash
```

## File Structure

### What We Created
- `install.sh` - Enhanced with binary download support
  - `--binaries <file>` - Install from local tarball
  - `--download-binaries --platform <platform>` - Download from GitHub
  - `--build` - Force build from source

- `web-installer.sh` - Simple web installer for port42.ai/install
  - Detects macOS vs other platforms
  - Downloads binaries for macOS
  - Falls back to source build for others

- `.github/workflows/release.yml` - GitHub Actions workflow
  - Builds for Apple Silicon and Intel Macs
  - Uploads to GitHub releases automatically
  - Includes test mode for validation

## Usage Examples

### For Users
```bash
# One-line install (future, when hosted)
curl -L https://port42.ai/install | bash

# Install specific version
curl -L https://port42.ai/install | bash -s -- --version=v0.1.0

# Force build from source
curl -L https://port42.ai/install | bash -s -- --build
```

### For Developers
```bash
# Test with local binaries
./install.sh --binaries port42-darwin-aarch64.tar.gz

# Download specific platform
./install.sh --download-binaries --platform darwin-x86_64

# Force source build
./install.sh --build
```

## Next Steps

### Immediate
1. Test GitHub Actions workflow with `workflow_dispatch`
2. Create first manual release with binaries
3. Test download functionality end-to-end

### Later
1. Host web-installer.sh at https://port42.ai/install
2. Add Intel Mac support (build on different runner)
3. Add Linux when users request it
4. Add Windows if needed

## Platform Expansion

Current: macOS only (darwin-aarch64, darwin-x86_64)

Future platforms (add based on demand):
- `linux-x86_64` - Most common Linux
- `linux-aarch64` - ARM Linux (Raspberry Pi, etc.)
- `windows-x86_64` - Windows 64-bit

## Troubleshooting

### Binary Won't Download
- Check GitHub release exists
- Verify binary name matches: `port42-darwin-aarch64.tar.gz`
- Check URL: `https://github.com/gordonmattey/port42/releases/latest/download/port42-darwin-aarch64.tar.gz`

### Installation Fails
- Ensure tarball contains `bin/` directory
- Check agents.json is included
- Verify binary permissions (should be executable)

### GitHub Actions Fails
- Check Go version (needs 1.21+)
- Verify Rust toolchain installed
- Ensure build.sh is executable

## Success! 🎉

You now have a professional binary release process for Port42:
- ✅ Binary download support in install.sh
- ✅ Web installer for easy installation  
- ✅ GitHub Actions for automated builds
- ✅ Local testing capability
- ✅ Documentation for the process

Total time: ~45 minutes 🚀