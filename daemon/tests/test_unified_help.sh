#!/bin/bash
# Test Unified Help System

set -e

echo "=== Testing Unified Help System ==="
echo

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to compare help output
compare_help() {
    local command="$1"
    local description="$2"
    
    echo -e "${BLUE}Testing: $description${NC}"
    echo
    
    # Capture interactive help
    echo "help $command" | port42 2>/dev/null | sed -n '/Help/,/^$/p' > /tmp/interactive_help.txt
    
    # Capture CLI help
    port42 help $command > /tmp/cli_help.txt 2>&1
    
    echo "Interactive help preview:"
    head -5 /tmp/interactive_help.txt
    echo "..."
    echo
    echo "CLI help preview:"
    head -5 /tmp/cli_help.txt
    echo "..."
    echo
    
    # Check if they contain similar content
    if grep -q "multidimensional filesystem" /tmp/cli_help.txt 2>/dev/null && \
       grep -q "multidimensional filesystem" /tmp/interactive_help.txt 2>/dev/null; then
        echo -e "${GREEN}✓ Both contain reality compiler language${NC}"
    else
        echo -e "${YELLOW}⚠ Content differs between interactive and CLI${NC}"
    fi
    
    echo
    echo "---"
    echo
}

echo -e "${BLUE}=== Manual Test Instructions ===${NC}"
echo
echo "1. Compare help outputs for consistency:"
echo
echo "   Interactive mode:"
echo "   $ port42"
echo "   port42> help ls"
echo
echo "   CLI mode:"
echo "   $ port42 help ls"
echo "   $ port42 ls --help"
echo
echo "   Both should show the SAME rich help with:"
echo "   - 'Navigate the multidimensional filesystem...'"
echo "   - Virtual paths explanation"
echo "   - Examples section"
echo "   - Reality compiler language"
echo

echo -e "${BLUE}=== Testing Main Help ===${NC}"
echo

echo "Testing: port42 --help"
port42 --help | head -20
echo "..."
echo

echo "Testing: port42 help"
port42 help | head -20
echo "..."
echo

echo -e "${BLUE}=== Testing Command Help ===${NC}"
echo

# Test each command
compare_help "ls" "ls command help"
compare_help "search" "search command help"
compare_help "possess" "possess command help"
compare_help "memory" "memory command help"

echo -e "${BLUE}=== What to Verify ===${NC}"
echo
echo "1. Main help (--help and help) should show:"
echo "   [ ] Reality compiler header"
echo "   [ ] CONSCIOUSNESS OPERATIONS section"
echo "   [ ] REALITY NAVIGATION section"
echo "   [ ] SYSTEM section"
echo "   [ ] 'The dolphins are listening' footer"
echo
echo "2. Command help should be IDENTICAL between:"
echo "   [ ] 'port42 help ls' and interactive 'help ls'"
echo "   [ ] 'port42 ls --help' and interactive 'help ls'"
echo "   [ ] Rich examples and virtual paths explanation"
echo "   [ ] Reality compiler language throughout"
echo
echo "3. No Clap default help should appear:"
echo "   [ ] No 'Usage: port42 ls [OPTIONS]' format"
echo "   [ ] No plain technical descriptions"
echo "   [ ] Everything uses our custom help"

echo
echo -e "${YELLOW}=== Known Issues to Check ===${NC}"
echo
echo "1. Does 'port42 ls --help' show our custom help?"
echo "2. Does 'port42 --help' show reality compiler theme?"
echo "3. Are examples formatted correctly?"
echo "4. Is the help identical between modes?"