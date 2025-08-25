#!/bin/bash
# Test artifact generation with Claude Opus 4 and improved prompts

echo "🧪 Port 42 Artifact Test Suite (Claude Opus 4)"
echo "============================================="
echo

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test function
test_artifact() {
    local test_name="$1"
    local agent="$2"
    local prompt="$3"
    
    echo -e "${BLUE}📍 TEST: $test_name${NC}"
    echo "   Agent: $agent"
    echo "   Prompt: $prompt"
    echo
    
    # Send the prompt
    echo "$prompt" | ./bin/port42 possess "$agent"
    
    # Wait for processing
    sleep 4
    echo
}

# Pre-test: Show current state
echo -e "${YELLOW}📊 Initial State:${NC}"
./bin/port42 ls /artifacts
echo

# Test 1: Simple document
test_artifact "Simple Document" "@ai-engineer" \
    "Create a README document artifact for Port 42. Name it 'port42-readme' with type 'document' and format 'md'. Include sections: Overview, Installation, Usage."

# Check results
echo -e "${GREEN}✓ Checking /artifacts after Test 1:${NC}"
./bin/port42 ls /artifacts
./bin/port42 ls /artifacts/document
echo

# Test 2: Pitch deck with @ai-founder
test_artifact "Pitch Deck" "@ai-founder" \
    "Create a pitch deck artifact called 'investor-deck' as a markdown document. Include: Problem (developers need AI pair programming), Solution (Port 42), Market Size, Traction."

# Test 3: Multi-file web app
test_artifact "Web App" "@ai-engineer" \
    "Create a code artifact called 'port42-dashboard'. It should be a simple web dashboard with index.html, style.css, and app.js files. Make it show Port 42 system status."

# Test 4: Design artifact
test_artifact "Logo Design" "@ai-muse" \
    "Create a design artifact called 'logo-concepts' that describes 3 different logo concepts for Port 42, incorporating dolphins and digital waves."

echo -e "${YELLOW}📊 Final Results:${NC}"
echo

echo -e "${BLUE}All Artifacts:${NC}"
./bin/port42 ls /artifacts
echo

echo -e "${BLUE}Document Artifacts:${NC}"
./bin/port42 ls /artifacts/document
echo

echo -e "${BLUE}Code Artifacts:${NC}"
./bin/port42 ls /artifacts/code
echo

echo -e "${BLUE}Design Artifacts:${NC}"
./bin/port42 ls /artifacts/design
echo

# Try to read an artifact
echo -e "${YELLOW}📖 Reading first document artifact:${NC}"
FIRST_DOC=$(./bin/port42 ls /artifacts/document 2>/dev/null | grep -v "empty" | grep -v "/artifacts" | head -1 | awk '{print $1}')
if [ -n "$FIRST_DOC" ]; then
    echo "Reading: /artifacts/document/$FIRST_DOC"
    ./bin/port42 cat "/artifacts/document/$FIRST_DOC" | head -20
else
    echo "No documents found to read"
fi
echo

# Check metadata
if [ -n "$FIRST_DOC" ]; then
    echo -e "${YELLOW}📋 Metadata for $FIRST_DOC:${NC}"
    ./bin/port42 info "/artifacts/document/$FIRST_DOC"
fi
echo

# Search test
echo -e "${YELLOW}🔍 Search Test:${NC}"
./bin/port42 search "port42" --type artifact
echo

# Check logs
echo -e "${YELLOW}📜 Recent Daemon Logs:${NC}"
tail -30 ~/.port42/daemon.log | grep -E "(artifact|generate_artifact|tools|will use)" | tail -15 || echo "No artifact logs found"

echo
echo -e "${GREEN}✅ Test suite complete!${NC}"
echo
echo "If artifacts were created successfully, you should see them listed above."
echo "If not, check:"
echo "  - Is the daemon running? (ps aux | grep port42d)"
echo "  - Are there errors in the logs? (tail -50 ~/.port42/daemon.log)"
echo "  - Did the AI use the tools? (look for 'tool_use' in logs)"