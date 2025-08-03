#!/bin/bash
# Test script for Port 42 artifact generation and virtual filesystem

set -e  # Exit on error

echo "🧪 Port 42 Artifact Test Suite"
echo "=============================="
echo

# Function to run command and show result
run_test() {
    local desc="$1"
    local cmd="$2"
    echo "📍 TEST: $desc"
    echo "   CMD: $cmd"
    echo "   ---"
    eval "$cmd"
    echo
    echo
}

# Start daemon if needed
if ! pgrep -f "port42d" > /dev/null; then
    echo "🚀 Starting daemon..."
    ./bin/port42 daemon start &
    sleep 3
    echo
fi

# Test 1: Create a simple document artifact
echo "=== PHASE 1: Create Document Artifact ==="
echo "Use the generate_artifact tool to create a markdown document artifact named 'port42-quickstart' with type 'document'. The document should have a title '# Port 42 Quick Start', a section '## What is Port 42?' explaining it's a reality compiler, and a section '## Getting Started' with basic usage." | ./bin/port42 possess @ai-engineer

sleep 3
echo

# Test 2: List artifacts directory
run_test "List /artifacts directory" \
    "./bin/port42 ls /artifacts"

# Test 3: List document artifacts
run_test "List /artifacts/document directory" \
    "./bin/port42 ls /artifacts/document"

# Test 4: Create a code artifact (multi-file app)
echo "=== PHASE 2: Create Code Artifact ==="
echo "Use the generate_artifact tool to create a code artifact named 'hello-dashboard' with type 'code'. Create a simple HTML dashboard with index.html that says 'Port 42 Dashboard' and style.css with basic styling. Use the content field to create multiple files." | ./bin/port42 possess @ai-engineer

sleep 3
echo

# Test 5: List all artifacts again
run_test "List /artifacts to see both types" \
    "./bin/port42 ls /artifacts"

# Test 6: List code artifacts
run_test "List /artifacts/code directory" \
    "./bin/port42 ls /artifacts/code"

# Test 7: Try to read a document artifact (if it exists)
echo "=== PHASE 3: Read Artifacts ==="
# First check what documents exist
DOCS=$(./bin/port42 ls /artifacts/document 2>/dev/null | grep -E "\.md" | head -1 | awk '{print $1}' || echo "")
if [ -n "$DOCS" ]; then
    run_test "Read first document artifact" \
        "./bin/port42 cat /artifacts/document/$DOCS"
else
    echo "⚠️  No document artifacts found to read"
fi

# Test 8: Get metadata info
if [ -n "$DOCS" ]; then
    run_test "Get metadata for document artifact" \
        "./bin/port42 info /artifacts/document/$DOCS"
fi

# Test 9: Search for artifacts
run_test "Search for artifacts with 'port42' keyword" \
    "./bin/port42 search port42"

# Test 10: List by date
run_test "List today's artifacts" \
    "./bin/port42 ls /by-date/$(date +%Y-%m-%d)"

# Test 11: Create artifact with @ai-founder
echo "=== PHASE 4: Test @ai-founder Artifact ==="
echo "Use the generate_artifact tool to create a document artifact named 'investor-pitch' with type 'document'. Create a brief pitch deck outline for Port 42 with sections: Problem, Solution, Market Size, and Ask." | ./bin/port42 possess @ai-founder

sleep 3
echo

# Test 12: List all artifacts one more time
run_test "Final artifact listing" \
    "./bin/port42 ls /artifacts"

echo "✅ Test suite complete!"
echo
echo "📊 Summary:"
echo "- Created document artifacts with @ai-engineer and @ai-founder"
echo "- Created multi-file code artifact"
echo "- Listed artifacts in virtual filesystem"
echo "- Read artifact content (if successful)"
echo "- Retrieved artifact metadata"
echo "- Searched for artifacts"
echo
echo "🔍 Check ~/.port42/daemon.log for detailed logs"
echo "   grep 'artifact' ~/.port42/daemon.log | tail -20"