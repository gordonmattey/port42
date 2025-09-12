# Port42 Releases

## Version History

### v0.1.0 (Current)
- Initial release
- Core reality compilation functionality
- AI agents (@ai-engineer, @ai-muse, @ai-analyst, @ai-founder)
- Virtual filesystem navigation
- Context tracking and watch mode
- macOS support (darwin-aarch64, darwin-x86_64)

## Platform Support

| Platform | Architecture | Binary Available |
|----------|-------------|-----------------|
| macOS | Apple Silicon (aarch64) | ✅ |
| macOS | Intel (x86_64) | ✅ |
| Linux | x86_64 | 🔜 Coming soon |
| Linux | aarch64 | 🔜 Coming soon |
| Windows | x86_64 | 🔜 Future |

## Release Files

Each release includes:
- `port42-${platform}-v${version}.tar.gz` - Versioned release
- `port42-${platform}.tar.gz` - Symlink to latest version

## Installation

### Latest Version
```bash
curl -fsSL https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh | bash
```

### Specific Version
```bash
curl -fsSL https://raw.githubusercontent.com/gordonmattey/port42/main/install.sh | bash -s -- --binaries releases/port42-darwin-aarch64-v0.1.0.tar.gz
```

## Upgrading

To upgrade to the latest version:
```bash
./install.sh --upgrade
```

## Version Management

The version is stored in `version.txt` at the repository root. To create a new release:

1. Update version: `echo '0.2.0' > version.txt`
2. Build with packaging: `./build.sh`
3. Commit the new tarball: `git add releases/ version.txt && git commit -m "Release v0.2.0"`
4. Push to repository: `git push`

The install script will automatically use the latest version available.