#!/bin/bash
# Test Phase 1: Help Infrastructure

set -e

echo "=== Testing Phase 1: Help Infrastructure ==="
echo

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test function
test_help() {
    local description="$1"
    local command="$2"
    local expected="$3"
    
    echo -n "Testing: $description... "
    
    # Create a temporary expect script for interactive testing
    cat > /tmp/test_help_expect.exp << EOF
#!/usr/bin/expect -f
set timeout 5
spawn port42
expect "port42>"
send "$command\r"
expect {
    "$expected" {
        send "exit\r"
        exit 0
    }
    timeout {
        send "exit\r"
        exit 1
    }
}
EOF
    
    chmod +x /tmp/test_help_expect.exp
    
    if /tmp/test_help_expect.exp > /dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Expected to find: $expected"
    fi
    
    rm -f /tmp/test_help_expect.exp
}

# Manual test instructions for non-expect environments
echo -e "${BLUE}=== Manual Test Instructions ===${NC}"
echo
echo "If expect is not available, test manually:"
echo
echo "1. Start interactive shell:"
echo "   $ port42"
echo
echo "2. Test general help:"
echo "   port42> help"
echo "   Should show: Main help screen"
echo
echo "3. Test command-specific help:"
echo "   port42> help possess"
echo "   Should show: Detailed possess command help with agents and examples"
echo
echo "   port42> help memory"
echo "   Should show: Memory command help with actions and examples"
echo
echo "   port42> help ls"
echo "   Should show: Virtual filesystem navigation help"
echo
echo "   port42> help search"
echo "   Should show: Search command with all filter options"
echo
echo "   port42> help cat"
echo "   Should show: Display content help"
echo
echo "   port42> help info"
echo "   Should show: Metadata examination help"
echo
echo "   port42> help reality"
echo "   Should show: Crystallized commands help"
echo
echo "   port42> help status"
echo "   Should show: Daemon status help"
echo
echo "4. Test invalid command help:"
echo "   port42> help invalid"
echo "   Should show: 'No help available' message"
echo
echo "5. Exit:"
echo "   port42> exit"
echo

# Check if expect is available
if command -v expect &> /dev/null; then
    echo -e "${BLUE}=== Running Automated Tests ===${NC}"
    echo
    
    test_help "General help command" "help" "Core Commands:"
    test_help "Possess command help" "help possess" "Channel an AI agent"
    test_help "Memory command help" "help memory" "persistent consciousness"
    test_help "LS command help" "help ls" "multidimensional filesystem"
    test_help "Search command help" "help search" "collective consciousness"
    test_help "Invalid command help" "help invalid" "No help available"
else
    echo -e "${YELLOW}Note: 'expect' not found. Please test manually using instructions above.${NC}"
fi

echo
echo -e "${BLUE}=== Visual Inspection ===${NC}"
echo
echo "Please verify:"
echo "1. Help text uses reality compiler language"
echo "2. Commands have clear examples"
echo "3. Text is colorized appropriately"
echo "4. Help for each command shows detailed usage"
echo "5. Error messages are helpful"