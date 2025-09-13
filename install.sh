#!/usr/bin/env bash
# Port 42 Universal Installer
# https://port42.ai
# 
# This script installs Port 42 on your system.
# It works in three modes:
#   1. Local mode - When run from a cloned repository
#   2. Remote mode - When curled from the web (downloads binaries)
#   3. Build mode - Forces building from source (--build flag)
#
# Usage:
#   Local:  ./install.sh
#   Remote: curl -fsSL https://port42.ai/install.sh | bash
#   Build:  curl -fsSL https://port42.ai/install.sh | bash -s -- --build

set -euo pipefail

# Debug trap to catch errors
trap 'echo "Error occurred at line $LINENO with exit code $?"' ERR

# Configuration
VERSION="${PORT42_VERSION:-latest}"
REPO="gordonmattey/port42"
PORT42_HOME="$HOME/.port42"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Set installation directory
set_install_dir() {
    # Everything goes in user's .port42 directory - no sudo ever needed!
    INSTALL_DIR="$PORT42_HOME/bin"
    print_info "Will install to: $INSTALL_DIR"
}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
GRAY='\033[0;90m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Section template functions
print_section_divider() {
    local color="$1"
    echo -e "${color}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_section_header() {
    local title="$1"
    local color="$2"
    echo
    print_section_divider "$color"
    echo -e "${BOLD}${title}${NC}"
    echo
}

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

# Check for running Port42 processes
check_running_processes() {
    local running_processes=()
    local has_critical_processes=false
    
    # Check for port42 CLI processes (excluding the installer itself)
    if pgrep -f "port42" | grep -v $$ > /dev/null 2>&1; then
        # Get detailed process list
        local processes=$(ps aux | grep -E "port42" | grep -v grep | grep -v "install.sh")
        
        if [ -n "$processes" ]; then
            # Check for interactive processes like context --watch
            if echo "$processes" | grep -E "context.*--watch|shell|interactive|swim" > /dev/null; then
                has_critical_processes=true
            fi
            
            print_warning "Detected running Port42 processes:"
            echo "$processes" | while read -r line; do
                echo "  $line" | cut -c1-120
            done
            echo
            
            if [ "$has_critical_processes" = true ]; then
                echo -e "${YELLOW}⚠️  Critical Port42 processes detected (context --watch, shell, etc.)${NC}"
                echo -e "${YELLOW}   Installing now will interrupt these processes and may cause data loss.${NC}"
                echo
                echo "Options:"
                echo "  1) Stop all Port42 processes and continue installation"
                echo "  2) Cancel installation (recommended - save your work first)"
                echo
                read -p "Choice [2]: " process_choice
                process_choice=${process_choice:-2}
                
                case "$process_choice" in
                    1)
                        print_info "Stopping all Port42 processes..."
                        # Stop daemon gracefully first
                        if command_exists port42; then
                            port42 daemon stop >/dev/null 2>&1 || true
                        fi
                        sleep 2
                        # Kill remaining processes
                        pkill -f "port42" || true
                        sleep 1
                        print_success "Processes stopped"
                        ;;
                    *)
                        print_info "Installation cancelled. Please save your work and close Port42 processes before installing."
                        exit 0
                        ;;
                esac
            else
                echo "These appear to be non-interactive processes."
                echo "Would you like to:"
                echo "  1) Stop them and continue installation"  
                echo "  2) Continue anyway (may cause issues)"
                echo "  3) Cancel installation"
                echo
                read -p "Choice [1]: " process_choice
                process_choice=${process_choice:-1}
                
                case "$process_choice" in
                    1)
                        # Stopping Port42 processes
                        if command_exists port42; then
                            port42 daemon stop >/dev/null 2>&1 || true
                        fi
                        sleep 2
                        pkill -f "port42" || true
                        sleep 1
                        print_success "Processes stopped"
                        ;;
                    2)
                        print_warning "Continuing with processes running - this may cause issues!"
                        ;;
                    *)
                        print_info "Installation cancelled"
                        exit 0
                        ;;
                esac
            fi
        fi
    fi
    
    # Check if daemon is running on port
    if command_exists port42 && port42 status >/dev/null 2>&1; then
        print_info "Port42 daemon is currently running"
        echo "The daemon will be restarted with new binaries after installation."
        echo
    fi
}

# Detect installation mode
detect_mode() {
    if [ -d "$SCRIPT_DIR/daemon/src" ] && [ -d "$SCRIPT_DIR/cli" ]; then
        INSTALL_MODE="local"
        print_info "Detected local repository installation"
    else
        INSTALL_MODE="remote"
        # Installing from remote source
    fi
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
        x86_64) 
            # Keep x86_64 for compatibility with our release naming
            ARCH="x86_64" 
            ;;
        aarch64|arm64) 
            # Use aarch64 to match our release naming
            ARCH="aarch64" 
            ;;
        *) print_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    # Detected platform: $PLATFORM
}

# Check prerequisites
check_prerequisites() {
    local missing_deps=()
    
    if [ "$INSTALL_MODE" = "remote" ]; then
        # Check for curl or wget
        if ! command_exists curl && ! command_exists wget; then
            missing_deps+=("curl or wget")
        fi
        
        # Check for git if we might need to build
        if [ "${BUILD_FROM_SOURCE:-false}" = "true" ] && ! command_exists git; then
            missing_deps+=("git")
        fi
    fi
    
    # If building from source, check for build tools
    if [ "${BUILD_FROM_SOURCE:-false}" = "true" ] || [ "$INSTALL_MODE" = "local" ]; then
        # Only check if we'll actually need to build
        if [ "$INSTALL_MODE" = "local" ] && [ -f "$SCRIPT_DIR/bin/port42d" ] && [ -f "$SCRIPT_DIR/bin/port42" ]; then
            # Binaries exist, no need for build tools
            :
        else
            if ! command_exists go; then
                missing_deps+=("go (1.21+)")
            fi
            if ! command_exists cargo; then
                missing_deps+=("rust/cargo")
            fi
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

# Build from local repository
build_local() {
    print_section_header "🔨 Building Port42" "$CYAN"
    
    if [ ! -f "$SCRIPT_DIR/build.sh" ]; then
        print_error "build.sh not found. Are you in the Port 42 repository?"
        exit 1
    fi
    
    # Run the build script
    cd "$SCRIPT_DIR"
    ./build.sh
    
    if [ ! -f "$SCRIPT_DIR/bin/port42d" ] || [ ! -f "$SCRIPT_DIR/bin/port42" ]; then
        print_error "Build failed. Please check the error messages above."
        exit 1
    fi
    
    print_success "Build completed successfully"
}

# Build from source (remote)
build_from_source() {
    print_info "Building from source..."
    
    # Create temp directory
    local temp_repo=$(mktemp -d)
    trap "rm -rf $temp_repo" EXIT
    
    print_info "Cloning repository..."
    git clone "https://github.com/$REPO.git" "$temp_repo"
    cd "$temp_repo"
    
    # Build daemon
    print_info "Building daemon..."
    cd daemon/src
    # Run go mod tidy first to ensure dependencies are up to date
    go mod tidy >/dev/null 2>&1 || true
    go build -o "$temp_repo/bin/port42d" .
    cd ../..
    
    # Build CLI
    print_info "Building CLI..."
    cd cli
    cargo build --release
    cp target/release/port42 "$temp_repo/bin/port42"
    cd ..
    
    # Copy binaries to install location
    SCRIPT_DIR="$temp_repo"
}

# Download and install pre-built binaries from GitHub releases or repo
download_and_install_binaries() {
    local platform="${1:-$PLATFORM}"
    
    # Get version from version.txt or default
    local version=$(curl -s "https://raw.githubusercontent.com/$REPO/main/version.txt" 2>/dev/null || echo "0.0.9")
    
    # Try versioned repo file first, then GitHub releases
    local versioned_url="https://raw.githubusercontent.com/$REPO/main/releases/port42-${platform}-v${version}.tar.gz"
    local release_url="https://github.com/$REPO/releases/latest/download/port42-${platform}.tar.gz"
    
    # Check which URL works
    local binary_url=""
    if curl -sI "$versioned_url" 2>/dev/null | head -n 1 | grep -q "200\|302"; then
        binary_url="$versioned_url"
    elif curl -sI "$release_url" 2>/dev/null | head -n 1 | grep -q "200\|302"; then
        binary_url="$release_url"
    else
        binary_url="$versioned_url"  # Default to versioned URL
    fi
    
    # Downloading binaries silently
    # URL: $binary_url
    
    # Create temp directory
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Download tarball
    local temp_file="$temp_dir/port42-binaries.tar.gz"
    if ! download "$binary_url" "$temp_file" 2>/dev/null; then
        print_warning "Pre-built binaries not available for $platform"
        print_info "Falling back to build from source..."
        return 1
    fi
    
    # Extract binaries
    # Extracting binaries
    # Try extraction with better error reporting
    local extract_result
    extract_result=$(tar -xzf "$temp_file" -C "$temp_dir" 2>&1)
    if [ $? -ne 0 ]; then
        print_error "Failed to extract binaries"
        [ -n "$extract_result" ] && print_error "Error details: $extract_result"
        # Check if file exists and is valid
        if [ ! -f "$temp_file" ]; then
            print_error "Downloaded file does not exist"
        else
            local file_type=$(file "$temp_file" | cut -d: -f2)
            print_error "File type: $file_type"
            print_error "File size: $(ls -lh "$temp_file" | awk '{print $5}')"
        fi
        return 1
    fi
    
    # Check if bin directory exists in the extraction
    if [ ! -d "$temp_dir/bin" ]; then
        print_error "Invalid binary package structure"
        return 1
    fi
    
    # Ensure daemon directory exists for agents.json
    mkdir -p "$temp_dir/daemon"
    
    # Check if agents.json is included and copy it
    if [ -f "$temp_dir/agents.json" ]; then
        cp "$temp_dir/agents.json" "$temp_dir/daemon/"
    elif [ -f "$temp_dir/daemon/agents.json" ]; then
        # Already in the right place
        :
    else
        # Create a default agents.json if not included
        echo '{}' > "$temp_dir/daemon/agents.json"
    fi
    
    # Set SCRIPT_DIR to temp for installation
    SCRIPT_DIR="$temp_dir"
    return 0
}

# Install binaries from local tarball file
install_from_local_binaries() {
    local binary_file="$1"
    
    if [ ! -f "$binary_file" ]; then
        print_error "Binary file not found: $binary_file"
        return 1
    fi
    
    # Installing from local binary file
    
    # Create temp directory
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Extract binaries
    # Extracting binaries
    # Try extraction with better error reporting
    local extract_result
    extract_result=$(tar -xzf "$binary_file" -C "$temp_dir" 2>&1)
    if [ $? -ne 0 ]; then
        print_error "Failed to extract binaries"
        [ -n "$extract_result" ] && print_error "Error details: $extract_result"
        # Check if file exists and is valid
        if [ ! -f "$binary_file" ]; then
            print_error "Binary file does not exist: $binary_file"
        else
            local file_type=$(file "$binary_file" | cut -d: -f2)
            print_error "File type: $file_type"
            print_error "File size: $(ls -lh "$binary_file" | awk '{print $5}')"
        fi
        return 1
    fi
    
    # Check if bin directory exists
    if [ ! -d "$temp_dir/bin" ]; then
        print_error "Invalid binary package structure"
        return 1
    fi
    
    # Ensure daemon directory exists for agents.json
    mkdir -p "$temp_dir/daemon"
    
    # Check if agents.json is included and copy it
    if [ -f "$temp_dir/agents.json" ]; then
        cp "$temp_dir/agents.json" "$temp_dir/daemon/"
    elif [ -f "$temp_dir/daemon/agents.json" ]; then
        # Already in the right place
        :
    else
        # Create a default agents.json if not included
        echo '{}' > "$temp_dir/daemon/agents.json"
    fi
    
    # Set SCRIPT_DIR to temp for installation
    SCRIPT_DIR="$temp_dir"
    return 0
}

# Legacy download function for individual files (kept for backward compatibility)
download_binaries() {
    local base_url="https://github.com/$REPO/releases/latest/download"
    
    if [ "$VERSION" != "latest" ]; then
        base_url="https://github.com/$REPO/releases/download/$VERSION"
    fi
    
    # Downloading pre-built binaries
    
    # Create temp directory for downloads
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Download binaries
    local daemon_url="${base_url}/port42d-${PLATFORM}"
    local cli_url="${base_url}/port42-${PLATFORM}"
    local agents_url="${base_url}/agents.json"
    
    # Downloading daemon
    if ! download "$daemon_url" "$temp_dir/port42d" 2>/dev/null; then
        print_warning "Pre-built binaries not available for $PLATFORM"
        return 1
    fi
    
    # Downloading CLI
    if ! download "$cli_url" "$temp_dir/port42" 2>/dev/null; then
        print_warning "Failed to download CLI"
        return 1
    fi
    
    # Downloading configuration
    if ! download "$agents_url" "$temp_dir/agents.json" 2>/dev/null; then
        print_warning "Failed to download agents.json"
        return 1
    fi
    
    # Make binaries executable
    chmod +x "$temp_dir/port42d" "$temp_dir/port42"
    
    # Set SCRIPT_DIR to temp for installation
    mkdir -p "$temp_dir/bin" "$temp_dir/daemon"
    mv "$temp_dir/port42d" "$temp_dir/bin/"
    mv "$temp_dir/port42" "$temp_dir/bin/"
    mv "$temp_dir/agents.json" "$temp_dir/daemon/"
    
    SCRIPT_DIR="$temp_dir"
    return 0
}

# Create Port 42 home directory structure
create_directories() {
    # Creating Port 42 directories
    
    # Check if directory exists and is owned by root
    if [ -d "$PORT42_HOME" ] && [ "$(stat -f %u "$PORT42_HOME" 2>/dev/null || stat -c %u "$PORT42_HOME" 2>/dev/null)" = "0" ]; then
        print_warning "Existing $PORT42_HOME is owned by root"
        print_info "Please run: sudo chown -R \$(whoami):staff $PORT42_HOME"
        print_info "Then run this installer again"
        exit 1
    fi
    
    mkdir -p "$PORT42_HOME"/{bin,commands,memory/sessions,artifacts,metadata,objects,tools}
    
    # Create initial memory index
    if [ ! -f "$PORT42_HOME/memory/index.json" ]; then
        echo '{"sessions":[],"stats":{"total_sessions":0,"total_commands":0}}' > "$PORT42_HOME/memory/index.json"
    fi
    
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
    
    # Directories created
}

# Update PATH in shell configuration
update_path() {
    local shell_name=$(basename "$SHELL")
    local shell_rc=""
    local path_additions=""
    
    # Both binaries and commands are in .port42
    local port42_paths="$PORT42_HOME/bin:$PORT42_HOME/commands"
    
    case "$shell_name" in
        bash)
            if [ -f "$HOME/.bashrc" ]; then
                shell_rc="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                shell_rc="$HOME/.bash_profile"
            fi
            path_additions="export PATH=\"\$PATH:$port42_paths\""
            ;;
        zsh)
            shell_rc="$HOME/.zshrc"
            path_additions="export PATH=\"\$PATH:$port42_paths\""
            ;;
        fish)
            shell_rc="$HOME/.config/fish/config.fish"
            path_additions="set -gx PATH \$PATH $PORT42_HOME/bin $PORT42_HOME/commands"
            ;;
        *)
            print_warning "Unknown shell: $shell_name"
            print_info "Please manually add to your PATH:"
            print_info "  - $PORT42_HOME/bin"
            print_info "  - $PORT42_HOME/commands"
            return
            ;;
    esac
    
    if [ -n "$shell_rc" ]; then
        # Check if PATH update already exists
        if ! grep -q "port42/commands" "$shell_rc" 2>/dev/null && ! grep -q "port42/bin" "$shell_rc" 2>/dev/null; then
            echo "" >> "$shell_rc"
            echo "# Port 42" >> "$shell_rc"
            echo "$path_additions" >> "$shell_rc"
            print_success "Updated PATH in $shell_rc"
            SAVED_TO_PROFILE="$shell_rc"
        fi
        # else PATH already configured
    fi
}

# Install binaries
install_binaries() {
    print_section_header "📦 Installing Port42" "$BLUE"
    
    # Check if binaries exist
    if [ ! -f "$SCRIPT_DIR/bin/port42d" ] || [ ! -f "$SCRIPT_DIR/bin/port42" ]; then
        print_error "Binaries not found in $SCRIPT_DIR/bin/"
        exit 1
    fi
    
    # Check if agents.json exists
    if [ ! -f "$SCRIPT_DIR/daemon/agents.json" ]; then
        print_error "agents.json not found in $SCRIPT_DIR/daemon/"
        exit 1
    fi
    
    # Create backup of existing binaries if they exist
    if [ -f "$INSTALL_DIR/port42" ] || [ -f "$INSTALL_DIR/port42d" ]; then
        # Backing up existing binaries
        mkdir -p "$PORT42_HOME/backup"
        [ -f "$INSTALL_DIR/port42" ] && cp "$INSTALL_DIR/port42" "$PORT42_HOME/backup/port42.bak" 2>/dev/null || true
        [ -f "$INSTALL_DIR/port42d" ] && cp "$INSTALL_DIR/port42d" "$PORT42_HOME/backup/port42d.bak" 2>/dev/null || true
    fi
    
    # Use atomic move operations to prevent corruption
    # Installing binaries atomically
    
    # Copy to temp location first
    cp "$SCRIPT_DIR/bin/port42d" "$INSTALL_DIR/port42d.tmp"
    cp "$SCRIPT_DIR/bin/port42" "$INSTALL_DIR/port42.tmp"
    chmod +x "$INSTALL_DIR/port42d.tmp" "$INSTALL_DIR/port42.tmp"
    
    # Atomic move (rename) - this prevents partial writes
    mv -f "$INSTALL_DIR/port42d.tmp" "$INSTALL_DIR/port42d"
    mv -f "$INSTALL_DIR/port42.tmp" "$INSTALL_DIR/port42"
    
    # Copy agents.json and guidance to home directory
    cp "$SCRIPT_DIR/daemon/agents.json" "$PORT42_HOME/"
    # Agent configuration installed
    
    # Copy agent_guidance.md if it exists
    if [ -f "$SCRIPT_DIR/daemon/agent_guidance.md" ]; then
        cp "$SCRIPT_DIR/daemon/agent_guidance.md" "$PORT42_HOME/"
        # Agent guidance installed
    fi
    
    # Verify installation
    if [ -x "$INSTALL_DIR/port42" ] && [ -x "$INSTALL_DIR/port42d" ]; then
        print_success "Port42 installed"
    else
        print_error "Installation verification failed - binaries may be corrupted"
        if [ -f "$PORT42_HOME/backup/port42.bak" ]; then
            # Restoring from backup
            cp "$PORT42_HOME/backup/port42.bak" "$INSTALL_DIR/port42" 2>/dev/null || true
            cp "$PORT42_HOME/backup/port42d.bak" "$INSTALL_DIR/port42d" 2>/dev/null || true
            chmod +x "$INSTALL_DIR/port42" "$INSTALL_DIR/port42d"
        fi
        exit 1
    fi
}

# Install Claude Code integration
install_claude_integration() {
    local claude_config="$HOME/.claude/CLAUDE.md"
    local p42_instructions=""
    
    # Configure Claude Code integration silently
    
    # Determine where to get P42CLAUDE.md from
    if [ -f "$SCRIPT_DIR/P42CLAUDE.md" ]; then
        # Found in extraction directory (from tarball) or local repo
        p42_instructions="$SCRIPT_DIR/P42CLAUDE.md"
    elif [ "$INSTALL_MODE" = "local" ] && [ -f "./P42CLAUDE.md" ]; then
        # Local installation - use file from current directory
        p42_instructions="./P42CLAUDE.md"
    else
        # Try to download from GitHub as fallback
        # Downloading Claude Code integration file
        local temp_file=$(mktemp)
        if download "https://raw.githubusercontent.com/$REPO/main/P42CLAUDE.md" "$temp_file" 2>/dev/null; then
            p42_instructions="$temp_file"
        else
            print_warning "Could not find Claude integration file, skipping"
            print_info "You can manually add it later from: https://github.com/$REPO/blob/main/P42CLAUDE.md"
            return
        fi
    fi
    
    # Create .claude directory if it doesn't exist
    mkdir -p "$HOME/.claude"
    
    # Check if Port42 integration exists
    if [ -f "$claude_config" ] && grep -q "<port42_integration>" "$claude_config" 2>/dev/null; then
        # Updating Port 42 Claude integration
        
        # Create backup
        cp "$claude_config" "${claude_config}.backup.$(date +%Y%m%d_%H%M%S)"
        
        # Replace the entire port42_integration section with updated content
        # Create a temp file to handle multi-line replacement properly
        awk '
            /<port42_integration>/ { 
                skip=1
                print ""  # Add blank line before if needed
                while ((getline line < "'"$p42_instructions"'") > 0) print line
            }
            /<\/port42_integration>/ { 
                skip=0
                next
            }
            !skip { print }
        ' "$claude_config" > "${claude_config}.tmp"
        
        # Move temp file to original
        mv "${claude_config}.tmp" "$claude_config"
        
        # Clean up temp file if it exists (only relevant for remote mode)
        if [ "$INSTALL_MODE" = "remote" ]; then
            [ -n "${temp_file:-}" ] && [ -f "${temp_file:-}" ] && rm "$temp_file" 2>/dev/null || true
        fi
        
        # Port 42 integration updated
        # Backup saved
    else
        # No existing integration - add it
        # Will either append to existing or create new Claude config
        
        # Just append the content
        cat "$p42_instructions" >> "$claude_config"
        
        # Clean up temp file if we used one
        [ "$INSTALL_MODE" = "remote" ] && [ -n "${temp_file:-}" ] && [ -f "$temp_file" ] && rm "$temp_file" 2>/dev/null
        
        # Claude Code configured for Port 42
        print_info "Claude will search and create Port 42 tools without being asked"
    fi
    
    # Done with Claude integration
}

# Configure Claude Code settings.json for Port42 commands
configure_claude_settings() {
    local settings_file="$HOME/.claude/settings.json"
    
    # Configure Claude Code command permissions silently
    
    # Ask for permission
    echo
    echo -e "${YELLOW}Port 42 needs to update Claude Code settings to:${NC}"
    echo "  • Allow Port42 commands without approval prompts"
    echo "  • Set appropriate timeout values for long-running operations"
    echo
    read -p "$(echo -e ${BOLD}"Update Claude Code settings? (y/n): "${NC})" -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Skipping Claude Code settings configuration"
        print_info "You can manually add Port42 commands to: $settings_file"
        return
    fi
    
    # Create .claude directory if it doesn't exist
    mkdir -p "$HOME/.claude"
    
    # Check if settings.json exists
    if [ -f "$settings_file" ]; then
        # Backup existing settings
        cp "$settings_file" "${settings_file}.backup.$(date +%Y%m%d_%H%M%S)"
        # Backed up existing settings
        
        # Don't skip - we'll merge/update existing settings
        
        # Check if jq is available (required for JSON manipulation)
        if ! command -v jq &> /dev/null; then
            print_warning "jq is not installed. Cannot automatically update Claude Code settings."
            print_info "Please install jq or manually add these to $settings_file:"
            echo "  allowedTools: [\"Bash(port42:*)\", \"Bash(port42)\", \"Bash(port42 search:*)\", etc.]"
            echo "  env.BASH_DEFAULT_TIMEOUT_MS: \"1800000\""
            echo "  env.BASH_MAX_TIMEOUT_MS: \"7200000\""
            return
        fi
        
        # Update with jq - merge and deduplicate
        # Merging Port42 commands into settings
        if jq '
            # Ensure allowedTools array exists
            if .allowedTools == null then .allowedTools = [] else . end |
            # Add Port42 commands with Bash() wrapper - put broadest match first
            .allowedTools += [
                "Bash(port42:*)",
                "Bash(port42)",
                "Bash(port42 cat:*)",
                "Bash(port42 cat /*)",
                "Bash(port42 info:*)",
                "Bash(port42 info /*)",
                "Bash(port42 swim:*)",
                "Bash(port42 search:*)",
                "Bash(port42 ls:*)",
                "Bash(port42 ls /*)",
                "Bash(port42 memory:*)",
                "Bash(port42 status:*)",
                "Bash(port42 daemon:*)",
                "Bash(port42 declare:*)",
                "Bash(port42 reality:*)",
                "Bash(port42 help:*)"
            ] |
            # Remove duplicates
            .allowedTools |= unique |
            # Ensure env object exists
            if .env == null then .env = {} else . end |
            # Set timeout values
            .env.BASH_DEFAULT_TIMEOUT_MS = "1800000" |
            .env.BASH_MAX_TIMEOUT_MS = "7200000"' "$settings_file" > "${settings_file}.tmp"; then
            # Move temp file to actual file
            if ! mv "${settings_file}.tmp" "$settings_file"; then
                print_error "Failed to move temporary file to $settings_file"
                print_info "Temporary file saved at ${settings_file}.tmp"
                return
            fi
        else
            print_error "Failed to process JSON with jq"
            return
        fi
    else
        # Create new settings file with new allowedTools format
        cat > "$settings_file" << 'EOF'
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "env": {
    "BASH_DEFAULT_TIMEOUT_MS": "1800000",
    "BASH_MAX_TIMEOUT_MS": "7200000"
  },
  "allowedTools": [
    "Bash(port42:*)",
    "Bash(port42)",
    "Bash(port42 search:*)",
    "Bash(port42 ls:*)",
    "Bash(port42 ls /*)",
    "Bash(port42 cat:*)",
    "Bash(port42 cat /*)",
    "Bash(port42 info:*)",
    "Bash(port42 info /*)",
    "Bash(port42 swim:*)",
    "Bash(port42 memory:*)",
    "Bash(port42 status:*)",
    "Bash(port42 daemon:*)",
    "Bash(port42 declare:*)",
    "Bash(port42 reality:*)",
    "Bash(port42 help:*)"
  ]
}
EOF
    fi
    
    # Verify the update worked
    if [ -f "$settings_file" ]; then
        if command -v jq &> /dev/null; then
            # Use jq to verify JSON structure - check for allowedTools with Bash wrapper
            if ! (jq -e '.allowedTools | map(select(. == "Bash(port42)" or . == "Bash(port42:*)")) | length > 0' "$settings_file" >/dev/null 2>&1 && \
               jq -e '.env.BASH_DEFAULT_TIMEOUT_MS' "$settings_file" >/dev/null 2>&1); then
                print_warning "Settings file was created but may be incomplete"
                print_info "Please verify $settings_file manually"
            fi
        else
            # Without jq, just check file exists
            # Settings file created
            print_info "Please verify it contains Port42 commands and timeout settings"
        fi
    else
        print_error "Failed to create settings file"
    fi
}

# Consolidated Claude Code setup
setup_claude_code() {
    print_section_header "🤖 Setting up Claude Code Integration" "$MAGENTA"
    
    # Check if Claude Code is installed
    local claude_installed=false
    if [ -d "$HOME/.claude" ]; then
        claude_installed=true
    fi
    
    # Configure API key (always needed)
    configure_api_key
    
    # If Claude Code is installed, configure it
    if [ "$claude_installed" = true ]; then
        install_claude_integration
        configure_claude_settings
        print_success "Claude Code configured"
    else
        print_success "Configuration complete"
    fi
}

# Configure API key
configure_api_key() {
    # Configure API key
    
    local current_key=""
    local key_source=""
    
    # Check for existing keys
    if [ -n "${PORT42_ANTHROPIC_API_KEY:-}" ]; then
        current_key="$PORT42_ANTHROPIC_API_KEY"
        key_source="PORT42_ANTHROPIC_API_KEY"
    elif [ -n "${ANTHROPIC_API_KEY:-}" ]; then
        current_key="$ANTHROPIC_API_KEY"
        key_source="ANTHROPIC_API_KEY"
    fi
    
    if [ -n "$current_key" ]; then
        # Mask the key for display (show first 8 and last 4 chars)
        local masked_key=""
        if [ ${#current_key} -gt 12 ]; then
            masked_key="${current_key:0:8}...${current_key: -4}"
        else
            masked_key="***hidden***"
        fi
        
        echo -e "${GREEN}Found existing API key:${NC} $masked_key (from $key_source)"
        echo
        echo "Would you like to:"
        echo "  1) Use this key"
        echo "  2) Enter a different key"
        echo "  3) Skip (configure later)"
        echo
        read -p "Choice [1]: " choice
        choice=${choice:-1}
        
        case "$choice" in
            2)
                read -p "Enter your Anthropic API key: " new_key
                if [ -n "$new_key" ]; then
                    API_KEY_TO_SAVE="$new_key"
                    SAVE_API_KEY=true
                fi
                ;;
            3)
                print_info "Skipping API key configuration"
                return
                ;;
            *)
                # Use existing key, but save as PORT42_ANTHROPIC_API_KEY if it was from ANTHROPIC_API_KEY
                if [ "$key_source" = "ANTHROPIC_API_KEY" ]; then
                    API_KEY_TO_SAVE="$current_key"
                    SAVE_API_KEY=true
                fi
                ;;
        esac
    else
        echo "No API key found. Port 42 requires an Anthropic API key for AI features."
        echo
        echo "Would you like to:"
        echo "  1) Enter your API key now"
        echo "  2) Skip (configure later)"
        echo
        read -p "Choice [1]: " choice
        choice=${choice:-1}
        
        if [ "$choice" = "1" ] || [ -z "$choice" ]; then
            read -p "Enter your Anthropic API key: " new_key
            if [ -n "$new_key" ]; then
                API_KEY_TO_SAVE="$new_key"
                SAVE_API_KEY=true
            fi
        else
            print_warning "No API key configured. You'll need to set PORT42_ANTHROPIC_API_KEY or ANTHROPIC_API_KEY to use AI features."
        fi
    fi
    
    # Save the key to shell profile if we have one to save
    if [ "${SAVE_API_KEY:-false}" = true ] && [ -n "${API_KEY_TO_SAVE:-}" ]; then
        local shell_rc=""
        local shell_name=$(basename "$SHELL")
        
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
        esac
        
        if [ -n "$shell_rc" ]; then
            # Check if PORT42_ANTHROPIC_API_KEY already exists
            if ! grep -q "PORT42_ANTHROPIC_API_KEY" "$shell_rc" 2>/dev/null; then
                echo "" >> "$shell_rc"
                echo "# Port 42 API Key" >> "$shell_rc"
                echo "export PORT42_ANTHROPIC_API_KEY='$API_KEY_TO_SAVE'" >> "$shell_rc"
                print_success "API key saved to $shell_rc"
                export PORT42_ANTHROPIC_API_KEY="$API_KEY_TO_SAVE"
            else
                print_info "Updating existing PORT42_ANTHROPIC_API_KEY in $shell_rc"
                # Use a temp file for safe replacement
                sed -i.bak "s/export PORT42_ANTHROPIC_API_KEY=.*/export PORT42_ANTHROPIC_API_KEY='$API_KEY_TO_SAVE'/" "$shell_rc"
                rm "${shell_rc}.bak"
                export PORT42_ANTHROPIC_API_KEY="$API_KEY_TO_SAVE"
            fi
        fi
    fi
}

# Start daemon for general use
start_daemon_for_use() {
    print_section_header "🚀 Starting Port42 Server" "$GREEN"
    
    # Check if daemon is running
    
    # Export path so we can use port42 command
    export PATH="$HOME/.port42/bin:$PATH"
    
    # Check if daemon is already running
    if "$HOME/.port42/bin/port42" status >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Daemon is already running${NC}"
        
        # Ask if they want to restart with new binaries
        echo
        echo "The daemon is currently running. Would you like to restart it with the new binaries?"
        echo "  1) Yes, restart daemon (recommended for binary installs)"
        echo "  2) No, keep current daemon running"
        echo
        echo -n "Choice [1]: "
        read -r restart_choice
        restart_choice=${restart_choice:-1}
        
        if [ "$restart_choice" = "1" ]; then
            echo -e "${BLUE}Stopping daemon gracefully...${NC}"
            
            # First try graceful stop
            if "$HOME/.port42/bin/port42" daemon stop >/dev/null 2>&1; then
                echo -e "${BLUE}Waiting for daemon to shut down...${NC}"
                sleep 3
            else
                echo -e "${YELLOW}⚠️  Daemon stop command failed, checking processes...${NC}"
                # If graceful stop fails, check if it's actually running
                if pgrep -f "port42d" > /dev/null 2>&1; then
                    echo -e "${YELLOW}Forcing daemon shutdown...${NC}"
                    pkill -f "port42d" || true
                    sleep 2
                fi
            fi
            
            echo -e "${BLUE}Starting daemon with new binaries...${NC}"
            if "$HOME/.port42/bin/port42" daemon start -b >/dev/null 2>&1; then
                sleep 2
                # Verify it's actually running
                if "$HOME/.port42/bin/port42" status >/dev/null 2>&1; then
                    print_success "Daemon started with new binaries"
                else
                    echo -e "${YELLOW}⚠️  Daemon started but may not be fully ready yet${NC}"
                    echo -e "${YELLOW}   Try: port42 daemon start -b${NC}"
                fi
            else
                echo -e "${YELLOW}⚠️  Could not start daemon automatically${NC}"
                echo -e "${YELLOW}   Please start manually: port42 daemon start -b${NC}"
            fi
        else
            echo -e "${BLUE}Keeping existing daemon running${NC}"
            echo -e "${YELLOW}Note: You're still using the old version until you restart${NC}"
        fi
    else
        # Daemon not running, start it
        echo -e "${BLUE}Starting Port42 daemon...${NC}"
        if "$HOME/.port42/bin/port42" daemon start -b >/dev/null 2>&1; then
            # Wait for daemon to be ready
            sleep 3
            
            # Verify daemon is running
            if "$HOME/.port42/bin/port42" status >/dev/null 2>&1; then
                print_success "Daemon started"
            else
                echo -e "${YELLOW}⚠️  Daemon started but may not be fully ready yet${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️  Could not start daemon (you can start it later with: port42 daemon start -b)${NC}"
        fi
    fi
}

# Show next steps
show_next_steps() {
    echo
    echo -e "${GREEN}${BOLD}🐬 Port 42 installation complete!${NC}"
    
    # Check if daemon is running from bootstrap
    local daemon_running=false
    if "$HOME/.port42/bin/port42" status >/dev/null 2>&1; then
        daemon_running=true
    fi
    
    # Check if port42-restart was created
    local has_restart_cmd=false
    if [ -f "$HOME/.port42/commands/port42-restart" ]; then
        has_restart_cmd=true
    fi
    
    # Getting Started section
    print_section_header "Getting Started:" "$YELLOW"
    
    # Check Claude Code installation
    local claude_code_installed=false
    if [ -d "$HOME/.claude" ]; then
        claude_code_installed=true
    fi
        
    if [ "$claude_code_installed" = true ]; then
        echo -e "${BLUE}${BOLD}🚀 Using Port42 Inside Claude Code${NC}"
        echo
        echo -e "   ${GREEN}✨ Just ask Claude to create any tool you need!${NC}"
        echo -e "   Examples:"
        echo -e "   ${GRAY}• \"Help me escape the 47-tab chaos\"${NC}"
        echo -e "   ${GRAY}• \"Create a tool to monitor my system performance\"${NC}"
        echo -e "   ${GRAY}• \"Build a command that organizes my downloads\"${NC}"
        echo
        echo -e "   Claude will automatically use Port42 to install tools ${GREEN}system-wide${NC}"
    fi
    
    print_section_divider "$YELLOW"
    echo -e "${BLUE}${BOLD}🐬 Using Port42 Outside Claude Code${NC}"
    echo
    echo -e "   1. ${BLUE}Your First Swim:${NC}"
    echo -e "      ${BOLD}port42 swim @ai-engineer 'help me escape the 47-tab chaos'${NC}"
    echo
    echo -e "   2. ${BLUE}Choose Your Agent:${NC}"
    echo -e "      ${GRAY}@ai-engineer${NC} - Technical implementations"
    echo -e "      ${GRAY}@ai-analyst${NC}  - Data analysis & insights"
    echo -e "      ${GRAY}@ai-muse${NC}     - Creative & artistic tools"
    echo -e "      ${GRAY}@ai-founder${NC}  - Business strategy & decisions"
    
    print_section_divider "$YELLOW"
    echo -e "${BLUE}${BOLD}👁️  Monitor Port42 Learning${NC}"
    echo
    echo -e "   ${BOLD}port42 context --watch${NC}"
    echo -e "   ${GRAY}See Port42 learn your patterns in real-time${NC}"
    echo -e "   ${GRAY}We like to run this in a split terminal with the Claude Code session${NC}"
    
    print_section_divider "$YELLOW"
    echo
    echo -e "${GREEN}${BOLD}🐬 Welcome to Port42 - Your Reality Compiler!${NC}"
    echo -e "Documentation: ${BOLD}https://port42.ai${NC}"
    echo -e "Help: ${BOLD}port42 help"
}

# Main installation flow
main() {
    # Ensure show_next_steps always runs even if script is interrupted
    trap 'show_next_steps; exit 1' INT TERM
    
    # Port 42 Universal Installer
    print_section_header "🌊 Welcome to Port42 Installation" "$CYAN"
    echo -e "${BOLD}Reality Compiler for Personal Computing${NC}"
    echo
    
    # Parse arguments
    BUILD_FROM_SOURCE=false
    DOWNLOAD_BINARIES=false
    BINARY_PLATFORM=""
    BINARY_PATH=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --build)
                BUILD_FROM_SOURCE=true
                # Force building from source
                shift
                ;;
            --download-binaries)
                DOWNLOAD_BINARIES=true
                shift
                ;;
            --platform)
                BINARY_PLATFORM="$2"
                shift 2
                ;;
            --binaries)
                BINARY_PATH="$2"
                print_info "Using local binary file: $BINARY_PATH"
                shift 2
                ;;
            --version=*)
                VERSION="${1#*=}"
                # Installing version: $VERSION
                shift
                ;;
            *)
                shift
                ;;
        esac
    done
    
    # Detect mode and platform
    detect_mode
    detect_platform
    set_install_dir
    
    # Check for running processes before proceeding
    check_running_processes
    
    # Check prerequisites
    check_prerequisites
    
    # Handle different installation paths
    if [ -n "$BINARY_PATH" ]; then
        # User provided a local binary file via command line
        if ! install_from_local_binaries "$BINARY_PATH"; then
            print_error "Failed to install from local binaries"
            exit 1
        fi
    elif [ "$DOWNLOAD_BINARIES" = true ] && [ -n "$BINARY_PLATFORM" ]; then
        # User explicitly wants to download binaries via command line
        if ! download_and_install_binaries "$BINARY_PLATFORM"; then
            print_info "Falling back to building from source..."
            BUILD_FROM_SOURCE=true
            check_prerequisites  # Re-check for build tools
            build_from_source
        fi
    elif [ "$BUILD_FROM_SOURCE" = true ]; then
        # User explicitly wants to build from source via command line
        if [ "$INSTALL_MODE" = "local" ]; then
            build_local
        else
            build_from_source
        fi
    else
        # No explicit flags - ask the user what they prefer
        echo
        echo -e "${BLUE}${BOLD}Installation Method${NC}"
        echo
        
        # Check if pre-built binaries are available
        local binary_available=false
        # Get version from version.txt or default
        local version=$(curl -s "https://raw.githubusercontent.com/$REPO/main/version.txt" 2>/dev/null || echo "0.0.9")
        
        # Try versioned file first (actual file), then GitHub releases
        local versioned_binary_url="https://raw.githubusercontent.com/$REPO/main/releases/port42-${PLATFORM}-v${version}.tar.gz"
        local release_binary_url="https://github.com/$REPO/releases/latest/download/port42-${PLATFORM}.tar.gz"
        
        # Check versioned file first (not the symlink)
        if curl -sI "$versioned_binary_url" 2>/dev/null | head -n 1 | grep -q "200\|302"; then
            binary_available=true
            binary_url="$versioned_binary_url"
        # Fall back to GitHub releases if available
        elif curl -sI "$release_binary_url" 2>/dev/null | head -n 1 | grep -q "200\|302"; then
            binary_available=true
            binary_url="$release_binary_url"
        fi
        
        if [ "$INSTALL_MODE" = "local" ]; then
            # We're in the repo already
            echo "You're running from the Port42 repository."
            echo
            
            # Also check for local release files
            local local_release_available=false
            if [ -f "$SCRIPT_DIR/releases/port42-${PLATFORM}.tar.gz" ]; then
                local_release_available=true
                print_info "Found local release package for $PLATFORM"
            fi
            
            echo "How would you like to install?"
            echo "  1) Build and install from this local repository"
            echo "  2) Use existing binaries in ./bin/ (skip build)"
            if [ "$local_release_available" = true ]; then
                echo "  3) Install from local release package (./releases/port42-${PLATFORM}.tar.gz)"
            elif [ "$binary_available" = true ]; then
                echo "  3) Download and install pre-built binaries from GitHub"
            else
                echo "  3) Clone fresh from GitHub and build (clean install)"
            fi
            echo
            read -p "Choice [1]: " install_choice
            install_choice=${install_choice:-1}
            
            case "$install_choice" in
                1)
                    print_info "Building from local repository..."
                    build_local
                    ;;
                2)
                    if [ -f "$SCRIPT_DIR/bin/port42d" ] && [ -f "$SCRIPT_DIR/bin/port42" ]; then
                        print_info "Using existing binaries from ./bin/"
                    else
                        print_warning "No binaries found in ./bin/, building instead..."
                        build_local
                    fi
                    ;;
                3)
                    if [ "$local_release_available" = true ]; then
                        # Installing from local release package
                        if ! install_from_local_binaries "$SCRIPT_DIR/releases/port42-${PLATFORM}.tar.gz"; then
                            print_info "Failed to install from release package, building instead..."
                            build_local
                        fi
                    elif [ "$binary_available" = true ]; then
                        # Downloading pre-built binaries from GitHub
                        if ! download_and_install_binaries "$PLATFORM"; then
                            print_info "Download failed, building from local source..."
                            build_local
                        fi
                    else
                        # Option 3 when no binaries: Clone fresh from GitHub
                        print_info "Cloning fresh copy from GitHub and building..."
                        BUILD_FROM_SOURCE=true
                        check_prerequisites
                        build_from_source
                    fi
                    ;;
                *)
                    build_local
                    ;;
            esac
        else
            # Remote installation - offer choice
            echo "How would you like to install Port42?"
            echo
            if [ "$binary_available" = true ]; then
                echo "  1) Download and install pre-built binaries from GitHub (recommended)"
                echo "  2) Clone and build from source (requires Go 1.21+ and Rust)"
            else
                echo "  1) Clone and build from source (pre-built binaries not yet available for $PLATFORM)"
            fi
            echo
            
            if [ "$binary_available" = true ]; then
                read -p "Choice [1]: " install_choice
                install_choice=${install_choice:-1}
            else
                # Even with one option, wait for user confirmation
                read -p "Press Enter to continue: " -r
                install_choice=1  # Only option is build from source
            fi
            
            case "$install_choice" in
                1)
                    if [ "$binary_available" = true ]; then
                        # Installing pre-built binaries
                        if ! download_and_install_binaries "$PLATFORM"; then
                            print_info "Binary download failed, falling back to source build..."
                            BUILD_FROM_SOURCE=true
                            check_prerequisites
                            build_from_source
                        fi
                    else
                        BUILD_FROM_SOURCE=true
                        check_prerequisites
                        build_from_source
                    fi
                    ;;
                2)
                    BUILD_FROM_SOURCE=true
                    check_prerequisites
                    build_from_source
                    ;;
                *)
                    # Default to binaries if available
                    if [ "$binary_available" = true ]; then
                        # Installing pre-built binaries
                        if ! download_and_install_binaries "$PLATFORM"; then
                            print_info "Binary download failed, falling back to source build..."
                            BUILD_FROM_SOURCE=true
                            check_prerequisites
                            build_from_source
                        fi
                    else
                        BUILD_FROM_SOURCE=true
                        check_prerequisites
                        build_from_source
                    fi
                    ;;
            esac
        fi
    fi
    
    # Common installation steps
    create_directories
    install_binaries
    update_path
    
    # Setup Claude Code (API key + Claude integration if installed)
    setup_claude_code
    
    # Start the daemon after all configuration
    start_daemon_for_use
    
    # Show completion message
    show_next_steps
    
    # Clear trap since we're done
    trap - INT TERM
}

# Global variable to track if we saved to shell profile
SAVED_TO_PROFILE=""

# Run main installation
main "$@"