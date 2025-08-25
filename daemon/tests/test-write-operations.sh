#!/bin/bash
# Test script for Port 42 write operations

echo "🐬 Port 42 Write Operations Test Suite"
echo "====================================="
echo ""

# Helper function to send requests
send_request() {
    local type=$1
    local payload=$2
    local id="test-$(date +%s)-$$"
    
    echo "{\"type\": \"$type\", \"id\": \"$id\", \"payload\": $payload}" | nc localhost 42
}

# Test 1: Store a document
echo "📝 Test 1: Storing a document at /artifacts/documents/test-doc.md"
CONTENT=$(echo "# Test Document\n\nThis is a test document for the virtual filesystem." | base64)
RESULT=$(send_request "store_path" "{
    \"path\": \"/artifacts/documents/test-doc.md\",
    \"content\": \"$CONTENT\",
    \"metadata\": {
        \"agent\": \"@test-agent\",
        \"crystallization_type\": \"artifact\",
        \"title\": \"Test Document\",
        \"description\": \"A test document for filesystem operations\"
    }
}")
echo "Result: $RESULT"
echo ""

# Test 2: Store a command
echo "🔧 Test 2: Storing a command at /commands/hello-test"
CMD_CONTENT=$(echo "#!/bin/bash\necho 'Hello from test command!'" | base64)
RESULT=$(send_request "store_path" "{
    \"path\": \"/commands/hello-test\",
    \"content\": \"$CMD_CONTENT\",
    \"metadata\": {
        \"memory_id\": \"mem-test-123\",
        \"agent\": \"@ai-engineer\",
        \"crystallization_type\": \"tool\",
        \"title\": \"Hello Test Command\",
        \"description\": \"Test command that says hello\"
    }
}")
echo "Result: $RESULT"
echo ""

# Test 3: Update metadata
echo "🔄 Test 3: Updating metadata for /artifacts/documents/test-doc.md"
RESULT=$(send_request "update_path" "{
    \"path\": \"/artifacts/documents/test-doc.md\",
    \"metadata_updates\": {
        \"lifecycle\": \"stable\",
        \"tags\": [\"test\", \"documentation\", \"virtual-fs\"],
        \"importance\": \"high\"
    }
}")
echo "Result: $RESULT"
echo ""

# Test 4: Create a memory thread
echo "🧠 Test 4: Creating a new memory thread"
RESULT=$(send_request "create_memory" "{
    \"agent\": \"@ai-architect\",
    \"initial_message\": \"Let's test the new filesystem design\"
}")
echo "Result: $RESULT"
echo ""

# Test 5: List virtual paths (using existing list_path if implemented)
echo "📂 Test 5: Checking if command was created"
ls -la ~/.port42/commands/hello-test 2>/dev/null && echo "✅ Command symlink created!" || echo "❌ Command symlink not found"
echo ""

# Test 6: Delete a path
echo "🗑️  Test 6: Deleting /artifacts/documents/test-doc.md"
RESULT=$(send_request "delete_path" "{
    \"path\": \"/artifacts/documents/test-doc.md\"
}")
echo "Result: $RESULT"
echo ""

echo "====================================="
echo "✨ Write operations test complete!"
echo ""
echo "Check the object store:"
echo "ls -la ~/.port42/objects/"
echo ""
echo "Check metadata:"
echo "ls -la ~/.port42/metadata/"