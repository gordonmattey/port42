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

# Global variable to track if we saved to shell profile
SAVED_TO_PROFILE=""

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
    
    mkdir -p "$PORT42_HOME"/{commands,memory/sessions,artifacts,metadata,objects,tools}
    
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
    
    # Check if agents.json exists
    if [ ! -f "$SCRIPT_DIR/daemon/agents.json" ]; then
        print_error "agents.json not found in daemon directory"
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
    
    # Copy agents.json to user config directory
    mkdir -p "$PORT42_HOME"
    cp "$SCRIPT_DIR/daemon/agents.json" "$PORT42_HOME/"
    print_success "Configuration copied to $PORT42_HOME/agents.json"
    
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
                SAVED_TO_PROFILE="$shell_rc"
            else
                print_info "API key already exists in $shell_rc"
                # Still need to activate in current session
                SAVED_TO_PROFILE="$shell_rc"
            fi
        fi
    fi
    
    return 0
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
    echo -e "${BLUE}🐛 Issues:${NC} https://github.com/yourusername/port42/issues"
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