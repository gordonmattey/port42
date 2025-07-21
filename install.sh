#!/usr/bin/env bash
set -euo pipefail

# Port 42 Installer
# https://port42.ai
# 
# This script installs Port 42 on your system.
# It will:
#   - Download pre-built binaries OR build from source
#   - Install the daemon (port42d) and CLI (port42)
#   - Create necessary directories
#   - Update your PATH
#   - Start the daemon

VERSION="${PORT42_VERSION:-latest}"
REPO="yourusername/port42"
INSTALL_DIR="/usr/local/bin"
PORT42_HOME="$HOME/.port42"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Helper functions
print_error() {
    echo -e "${RED}❌ Error: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}🐬 $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect OS and Architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux|darwin) ;;
        *) print_error "Unsupported OS: $OS"; exit 1 ;;
    esac
    
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) print_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    print_info "Detected platform: $PLATFORM"
}

# Check prerequisites
check_prerequisites() {
    local missing_deps=()
    
    # Check for curl or wget
    if ! command_exists curl && ! command_exists wget; then
        missing_deps+=("curl or wget")
    fi
    
    # If building from source, check for build tools
    if [ "${BUILD_FROM_SOURCE:-false}" = "true" ]; then
        if ! command_exists go; then
            missing_deps+=("go (1.21+)")
        fi
        if ! command_exists cargo; then
            missing_deps+=("rust/cargo")
        fi
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        print_info "Please install the missing dependencies and try again."
        exit 1
    fi
}

# Download file with curl or wget
download() {
    local url=$1
    local output=$2
    
    if command_exists curl; then
        curl -fsSL "$url" -o "$output"
    elif command_exists wget; then
        wget -q "$url" -O "$output"
    else
        print_error "Neither curl nor wget found"
        exit 1
    fi
}

# Build from source
build_from_source() {
    print_info "Building from source..."
    
    # Clone repository
    local temp_repo=$(mktemp -d)
    trap "rm -rf $temp_repo" EXIT
    
    print_info "Cloning repository..."
    git clone "https://github.com/$REPO.git" "$temp_repo"
    cd "$temp_repo"
    
    # Build daemon
    print_info "Building daemon..."
    cd daemon
    go build -o "$TEMP_DIR/port42d" .
    cd ..
    
    # Build CLI
    print_info "Building CLI..."
    cd cli
    cargo build --release
    cp target/release/port42 "$TEMP_DIR/port42"
    cd ..
}

# Download pre-built binaries
download_binaries() {
    local base_url="https://github.com/$REPO/releases/latest/download"
    
    if [ "$VERSION" != "latest" ]; then
        base_url="https://github.com/$REPO/releases/download/$VERSION"
    fi
    
    print_info "Downloading pre-built binaries..."
    
    # Try to download pre-built binaries
    if download "$base_url/port42d-$PLATFORM" "$TEMP_DIR/port42d" 2>/dev/null && \
       download "$base_url/port42-$PLATFORM" "$TEMP_DIR/port42" 2>/dev/null; then
        chmod +x "$TEMP_DIR/port42d" "$TEMP_DIR/port42"
        return 0
    else
        print_warning "Pre-built binaries not found for $PLATFORM"
        print_info "Falling back to building from source..."
        BUILD_FROM_SOURCE=true
        build_from_source
    fi
}

# Install binaries
install_binaries() {
    print_info "Installing binaries to $INSTALL_DIR..."
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        cp "$TEMP_DIR/port42d" "$INSTALL_DIR/"
        cp "$TEMP_DIR/port42" "$INSTALL_DIR/"
    else
        print_info "Need sudo access to install to $INSTALL_DIR"
        sudo cp "$TEMP_DIR/port42d" "$INSTALL_DIR/"
        sudo cp "$TEMP_DIR/port42" "$INSTALL_DIR/"
    fi
    
    print_success "Binaries installed"
}

# Create Port 42 home directory structure
create_directories() {
    print_info "Setting up Port 42 directories..."
    
    # Check if directory exists and handle permissions
    if [ -d "$PORT42_HOME" ]; then
        # Check ownership
        local owner=$(stat -f %u "$PORT42_HOME" 2>/dev/null || stat -c %u "$PORT42_HOME" 2>/dev/null || echo "unknown")
        local current_uid=$(id -u)
        
        if [ "$owner" = "0" ] && [ "$current_uid" != "0" ]; then
            print_error "Port 42 directory is owned by root"
            print_info "Please run: sudo chown -R $(whoami):$(id -gn) $PORT42_HOME"
            print_info "Then run the installer again"
            exit 1
        elif [ "$owner" != "$current_uid" ] && [ "$owner" != "unknown" ]; then
            print_warning "Port 42 directory is owned by another user"
            print_info "This may cause permission issues"
        fi
        
        print_info "Using existing Port 42 directory"
    fi
    
    # Create directories (will not fail if they exist)
    mkdir -p "$PORT42_HOME"/{commands,memory/sessions,templates,entities}
    
    # Create initial memory index if missing
    if [ ! -f "$PORT42_HOME/memory/index.json" ]; then
        echo '{"sessions":[],"stats":{"total_sessions":0,"total_commands":0}}' > "$PORT42_HOME/memory/index.json"
        print_success "Initialized memory store"
    fi
    
    # Ensure proper permissions
    chmod 755 "$PORT42_HOME"
    chmod 755 "$PORT42_HOME/commands"
    chmod 700 "$PORT42_HOME/memory"  # Private for user
    
    # Create activation helper
    cat > "$PORT42_HOME/activate.sh" << 'EOF'
#!/bin/bash
# Quick activation script for Port 42
# Usage: source ~/.port42/activate.sh

# Source shell profile based on current shell
case "$(basename "$SHELL")" in
    bash) [ -f ~/.bashrc ] && source ~/.bashrc || source ~/.bash_profile ;;
    zsh) [ -f ~/.zshrc ] && source ~/.zshrc ;;
    *) echo "Please source your shell profile manually" ;;
esac

# Check if API key is available
if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
    echo "✅ API key loaded"
    echo "Run 'port42 daemon start' to start the daemon with AI features"
else
    echo "⚠️  No API key found"
    echo "Set ANTHROPIC_API_KEY to enable AI features"
fi
EOF
    chmod +x "$PORT42_HOME/activate.sh"
    
    print_success "Port 42 directories ready at $PORT42_HOME"
}

# Update PATH in shell configuration
update_path() {
    local shell_name=$(basename "$SHELL")
    local shell_rc=""
    local path_line="export PATH=\"\$PATH:$PORT42_HOME/commands\""
    
    case "$shell_name" in
        bash)
            if [ -f "$HOME/.bashrc" ]; then
                shell_rc="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                shell_rc="$HOME/.bash_profile"
            fi
            ;;
        zsh)
            shell_rc="$HOME/.zshrc"
            ;;
        fish)
            shell_rc="$HOME/.config/fish/config.fish"
            path_line="set -gx PATH \$PATH $PORT42_HOME/commands"
            ;;
        *)
            print_warning "Unknown shell: $shell_name"
            print_info "Please manually add $PORT42_HOME/commands to your PATH"
            return
            ;;
    esac
    
    if [ -n "$shell_rc" ]; then
        # Check if PATH update already exists
        if ! grep -q "port42/commands" "$shell_rc" 2>/dev/null; then
            echo "" >> "$shell_rc"
            echo "# Port 42" >> "$shell_rc"
            echo "$path_line" >> "$shell_rc"
            print_success "Updated PATH in $shell_rc"
            print_info "Restart your shell or run: source $shell_rc"
        else
            print_info "PATH already configured"
        fi
    fi
}

# Global variable to track if we saved to shell profile
SAVED_TO_PROFILE=""

# Configure API key
configure_api_key() {
    # Check if already set in environment
    if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
        print_success "Using API key from environment"
        return 0
    fi
    
    # Check if running non-interactively
    if [ ! -t 0 ] || [ "${PORT42_SKIP_API_KEY:-}" = "yes" ]; then
        print_warning "No API key configured (non-interactive mode)"
        print_info "AI features will be disabled until you set ANTHROPIC_API_KEY"
        return 1
    fi
    
    # Prompt for API key
    echo
    print_info "Port 42 uses Anthropic's Claude for AI features"
    echo "Get your API key from: https://console.anthropic.com/api-keys"
    echo
    read -r -p "Enter your Anthropic API key (or press Enter to skip): " api_key
    
    if [ -z "$api_key" ]; then
        print_warning "Skipping API key configuration"
        print_info "AI features will be disabled until you set ANTHROPIC_API_KEY"
        return 1
    fi
    
    # Validate key format (basic check)
    if [[ ! "$api_key" =~ ^sk-ant-api[0-9]{2}-[a-zA-Z0-9_-]{48,}$ ]]; then
        print_warning "API key format looks incorrect (should start with sk-ant-api)"
        read -r -p "Continue anyway? [y/N] " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            print_info "Skipping API key configuration"
            return 1
        fi
    fi
    
    # Export for current session AND any subshells
    export ANTHROPIC_API_KEY="$api_key"
    
    # IMPORTANT: The export above only affects this script's process
    # We need to tell the user to source their profile or provide the key to the daemon
    
    # Offer to save to shell profile
    echo
    read -r -p "Save API key to your shell profile for future sessions? [Y/n] " save_key
    if [[ ! "$save_key" =~ ^[Nn]$ ]]; then
        local shell_name=$(basename "$SHELL")
        local shell_rc=""
        
        case "$shell_name" in
            bash)
                if [ -f "$HOME/.bashrc" ]; then
                    shell_rc="$HOME/.bashrc"
                elif [ -f "$HOME/.bash_profile" ]; then
                    shell_rc="$HOME/.bash_profile"
                fi
                ;;
            zsh)
                shell_rc="$HOME/.zshrc"
                ;;
            fish)
                shell_rc="$HOME/.config/fish/config.fish"
                ;;
        esac
        
        if [ -n "$shell_rc" ]; then
            # Check if already exists
            if ! grep -q "ANTHROPIC_API_KEY" "$shell_rc" 2>/dev/null; then
                echo "" >> "$shell_rc"
                echo "# Port 42 - Anthropic API Key" >> "$shell_rc"
                echo "export ANTHROPIC_API_KEY='$api_key'" >> "$shell_rc"
                print_success "API key saved to $shell_rc"
                SAVED_TO_PROFILE="$shell_rc"
            else
                print_info "API key already exists in $shell_rc"
                # Still need to activate in current session
                SAVED_TO_PROFILE="$shell_rc"
            fi
        fi
    fi
    
    # Make sure the key is available for the daemon start
    print_success "API key configured for this installation session"
    
    return 0
}


# Check for existing installation
check_existing_installation() {
    if [ -f "$INSTALL_DIR/port42" ] && [ -f "$INSTALL_DIR/port42d" ]; then
        print_info "Existing Port 42 installation detected"
        
        # Try to get version
        if "$INSTALL_DIR/port42" --version >/dev/null 2>&1; then
            local current_version=$("$INSTALL_DIR/port42" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
            print_info "Current version: $current_version"
        fi
        
        # Check if daemon is running
        local daemon_running=false
        if pgrep -f "port42d" >/dev/null 2>&1; then
            daemon_running=true
            print_warning "Port 42 daemon is currently running"
        fi
        
        # Non-interactive mode (for CI/CD)
        if [ "${PORT42_UPGRADE:-}" = "yes" ] || [ "${CI:-}" = "true" ]; then
            print_info "Auto-upgrade mode: proceeding with installation"
            if [ "$daemon_running" = true ]; then
                print_info "Stopping daemon..."
                pkill -f port42d || true
                sleep 2
            fi
            return
        fi
        
        # Interactive mode
        if [ -t 0 ]; then  # Check if stdin is a terminal
            echo
            echo "What would you like to do?"
            echo "  1) Upgrade/reinstall Port 42 (recommended)"
            echo "  2) Cancel installation"
            echo
            read -r -p "Enter choice [1-2]: " choice
            
            case "$choice" in
                1)
                    print_info "Proceeding with upgrade..."
                    if [ "$daemon_running" = true ]; then
                        print_info "Stopping daemon..."
                        pkill -f port42d || true
                        sleep 2
                    fi
                    ;;
                2)
                    print_info "Installation cancelled"
                    exit 0
                    ;;
                *)
                    print_error "Invalid choice"
                    exit 1
                    ;;
            esac
        else
            # Non-interactive without explicit upgrade flag
            print_error "Existing installation found. Use PORT42_UPGRADE=yes to upgrade non-interactively"
            exit 1
        fi
    fi
}

# Main installation flow
main() {
    echo
    echo -e "${BOLD}${BLUE}Port 42 Installer${NC}"
    echo -e "${BLUE}═══════════════════${NC}"
    echo
    
    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Run installation steps
    detect_platform
    check_prerequisites
    check_existing_installation
    
    # Download or build binaries
    if [ "${BUILD_FROM_SOURCE:-false}" = "true" ]; then
        build_from_source
    else
        download_binaries
    fi
    
    # Install everything
    install_binaries
    create_directories
    update_path
    configure_api_key
    
    # Success message
    echo
    echo -e "${GREEN}${BOLD}✅ Port 42 installed successfully!${NC}"
    echo
    echo -e "${BLUE}🐬 Getting started:${NC}"
    echo -e "   ${BOLD}port42 daemon start${NC} - Start the daemon"
    echo -e "   ${BOLD}port42${NC}              - Enter the Port 42 shell"
    echo -e "   ${BOLD}port42 possess${NC}      - Start an AI conversation"
    echo -e "   ${BOLD}port42 status${NC}       - Check daemon status"
    echo -e "   ${BOLD}port42 list${NC}         - List your commands"
    echo
    echo -e "${BLUE}📚 Documentation:${NC} https://port42.ai/docs"
    echo -e "${BLUE}🐛 Issues:${NC} https://github.com/$REPO/issues"
    echo
    
    if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
        echo -e "${YELLOW}${BOLD}⚠️  No API key was configured${NC}"
        echo -e "   To enable AI features:"
        echo -e "   export ANTHROPIC_API_KEY='your-key-here'"
        echo -e "   port42 daemon start"
        echo
    else
        # Check if the key was just configured but shell needs sourcing
        if [ -n "$SAVED_TO_PROFILE" ]; then
            echo -e "${YELLOW}${BOLD}⚠️  To activate your API key:${NC}"
            echo
            echo -e "${GREEN}${BOLD}Run this command:${NC}"
            echo -e "   ${BOLD}source $SAVED_TO_PROFILE${NC}"
            echo
            echo -e "${BLUE}Then start the daemon:${NC}"
            echo -e "   ${BOLD}port42 daemon start${NC}"
            echo
        else
            echo -e "${GREEN}${BOLD}Start the daemon:${NC}"
            echo -e "   ${BOLD}port42 daemon start${NC}"
            echo
        fi
    fi
}

# Run main installation
main "$@"