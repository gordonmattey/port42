#!/usr/bin/env bash
# Port 42 Activation Script
# Source this file to activate Port 42 in your current shell
# Usage: source activate.sh

# Detect shell and source appropriate profile
SHELL_NAME=$(basename "$SHELL")

case "$SHELL_NAME" in
    bash)
        if [ -f "$HOME/.bashrc" ]; then
            source "$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            source "$HOME/.bash_profile"
        fi
        ;;
    zsh)
        if [ -f "$HOME/.zshrc" ]; then
            source "$HOME/.zshrc"
        fi
        ;;
    *)
        echo "⚠️  Unknown shell: $SHELL_NAME"
        echo "Please manually source your shell profile"
        return 1
        ;;
esac

# Check if API key is now available
if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
    echo "✅ API key loaded successfully"
    
    # Restart daemon if installed
    if command -v port42 >/dev/null 2>&1; then
        echo "🔄 Restarting Port 42 daemon with API key..."
        port42 daemon restart
    fi
else
    echo "⚠️  No API key found in shell profile"
    echo "Set it with: export ANTHROPIC_API_KEY='your-key-here'"
fi