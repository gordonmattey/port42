#!/bin/bash
# Final artifact test - should work with updated daemon

echo "🧪 Port 42 Artifact Generation Test"
echo "==================================="
echo

# Function to test and display results
test_artifact() {
    local agent="$1"
    local prompt="$2"
    
    echo "📍 Testing with $agent"
    echo "   Prompt: $prompt"
    echo "$prompt" | ./bin/port42 possess "$agent"
    echo
    sleep 3
    
    echo "📍 Checking artifacts..."
    ./bin/port42 ls /artifacts
    echo
}

# Test 1: Simple document with @ai-engineer
test_artifact "@ai-engineer" \
    "Create a markdown document artifact called 'readme-test' that contains a simple README for Port 42. Use the artifact generation tool."

# Test 2: Document with @ai-founder
test_artifact "@ai-founder" \
    "Create a pitch deck artifact called 'pitch-outline' with sections for Problem, Solution, and Market Size."

# Test 3: Code artifact
test_artifact "@ai-engineer" \
    "Create a code artifact called 'hello-web' with a simple HTML file that displays 'Hello from Port 42'."

echo "📍 Final artifact listing:"
./bin/port42 ls /artifacts
echo
./bin/port42 ls /artifacts/document
echo
./bin/port42 ls /artifacts/code

echo
echo "📍 Checking logs for tool usage:"
tail -30 ~/.port42/daemon.log | grep -E "(generate_artifact|Artifact|will use|tools)" || echo "No artifact logs found"

echo
echo "✅ Test complete!"