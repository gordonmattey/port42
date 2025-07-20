#!/usr/bin/env bash
# Local installation script for Port 42 (for testing)
set -euo pipefail

# Port 42 Local Installer
# For testing installation from local build

INSTALL_DIR="/usr/local/bin"
PORT42_HOME="$HOME/.port42"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

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

# Create Port 42 home directory structure
create_directories() {
    print_info "Creating Port 42 directories..."
    
    # Check if directory exists and is owned by root
    if [ -d "$PORT42_HOME" ] && [ "$(stat -f %u "$PORT42_HOME" 2>/dev/null || stat -c %u "$PORT42_HOME" 2>/dev/null)" = "0" ]; then
        print_warning "Existing $PORT42_HOME is owned by root"
        print_info "Please run: sudo chown -R \$(whoami):staff $PORT42_HOME"
        print_info "Then run this installer again"
        exit 1
    fi
    
    mkdir -p "$PORT42_HOME"/{commands,memory/sessions,templates,entities}
    
    # Create initial memory index
    if [ ! -f "$PORT42_HOME/memory/index.json" ]; then
        echo '{"sessions":[],"stats":{"total_sessions":0,"total_commands":0}}' > "$PORT42_HOME/memory/index.json"
    fi
    
    print_success "Directories created at $PORT42_HOME"
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

# Install binaries
install_binaries() {
    print_info "Installing binaries to $INSTALL_DIR..."
    
    # Check if binaries exist
    if [ ! -f "$SCRIPT_DIR/bin/port42d" ] || [ ! -f "$SCRIPT_DIR/bin/port42" ]; then
        print_error "Binaries not found. Please run ./build.sh first"
        exit 1
    fi
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        cp "$SCRIPT_DIR/bin/port42d" "$INSTALL_DIR/"
        cp "$SCRIPT_DIR/bin/port42" "$INSTALL_DIR/"
    else
        print_info "Need sudo access to install to $INSTALL_DIR"
        sudo cp "$SCRIPT_DIR/bin/port42d" "$INSTALL_DIR/"
        sudo cp "$SCRIPT_DIR/bin/port42" "$INSTALL_DIR/"
    fi
    
    print_success "Binaries installed"
}

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
    
    # Export for current session
    export ANTHROPIC_API_KEY="$api_key"
    
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
            else
                print_info "API key already exists in $shell_rc"
            fi
        fi
    fi
    
    return 0
}

# Start the daemon
start_daemon() {
    print_info "Starting Port 42 daemon..."
    
    # Check if daemon is already running by looking for actual process
    if pgrep -f "port42d" >/dev/null 2>&1; then
        print_success "Daemon already running"
        return
    fi
    
    # Try to start daemon
    if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
        # Start with API key from environment
        env ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" nohup "$INSTALL_DIR/port42d" >/dev/null 2>&1 &
        sleep 2
        
        if "$INSTALL_DIR/port42" status >/dev/null 2>&1; then
            print_success "Daemon started successfully"
        else
            print_warning "Daemon failed to start on port 42, trying port 4242..."
            env ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" PORT=4242 nohup "$INSTALL_DIR/port42d" >/dev/null 2>&1 &
            sleep 2
            if "$INSTALL_DIR/port42" status >/dev/null 2>&1; then
                print_success "Daemon started on port 4242"
            else
                print_warning "Daemon failed to start. Try running manually with: sudo port42d"
            fi
        fi
    else
        print_warning "No ANTHROPIC_API_KEY found"
        print_info "Daemon started without AI features"
        print_info "To enable AI, set your key and restart:"
        print_info "  export ANTHROPIC_API_KEY='your-key-here'"
        print_info "  ./port42-daemon restart"
        
        # Start anyway for non-AI features
        nohup "$INSTALL_DIR/port42d" >/dev/null 2>&1 &
        sleep 2
    fi
}

# Main installation flow
main() {
    echo
    echo -e "${BOLD}${BLUE}Port 42 Local Installer${NC}"
    echo -e "${BLUE}═══════════════════════${NC}"
    echo
    
    # Run installation steps
    create_directories
    install_binaries
    update_path
    configure_api_key
    start_daemon
    
    # Success message
    echo
    echo -e "${GREEN}${BOLD}✅ Port 42 installed successfully!${NC}"
    echo
    echo -e "${BLUE}🐬 Getting started:${NC}"
    echo -e "   ${BOLD}port42${NC}              - Enter the Port 42 shell"
    echo -e "   ${BOLD}port42 possess${NC}      - Start an AI conversation"
    echo -e "   ${BOLD}port42 status${NC}       - Check daemon status"
    echo -e "   ${BOLD}port42 list${NC}         - List your commands"
    echo
    echo -e "${BLUE}📚 Documentation:${NC} https://port42.ai/docs"
    echo -e "${BLUE}🐛 Issues:${NC} https://github.com/yourusername/port42/issues"
    echo
    
    if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
        echo -e "${YELLOW}${BOLD}⚠️  Remember to set your API key:${NC}"
        echo -e "   export ANTHROPIC_API_KEY='your-key-here'"
        echo -e "   port42d  # Start the daemon"
        echo
    fi
}

# Run main installation
main "$@"