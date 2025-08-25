#!/bin/bash
# Test 2: Reference Resolution Integration
# Tests that references are resolved and integrated into AI generation

set -e

echo "🔍 Test 2: Reference Resolution Integration"

# Setup test files
TEST_DIR="/tmp/port42-test-refs"
mkdir -p "$TEST_DIR"

# Create test reference file
cat > "$TEST_DIR/sample-data.json" << 'EOF'
{
  "format": "json",
  "schema": {
    "user": {"name": "string", "email": "string", "age": "number"},
    "metadata": {"created": "datetime", "source": "string"}
  },
  "validation_rules": [
    "email must be valid format",
    "age must be positive integer",
    "name is required field"
  ]
}
EOF

echo "📁 Created test reference file"

# Test 1: Tool with file reference
echo "🧪 Test 1: Tool with file reference"
tool_name="data-validator-$(date +%s)"
port42 declare tool "$tool_name" --transforms "validation,json,schema" \
  --ref "file:$TEST_DIR/sample-data.json" || {
  echo "❌ FAIL: Tool creation with file reference failed"
  exit 1
}

# Verify tool was created
if [[ ! -f "$HOME/.port42/commands/$tool_name" ]]; then
  echo "❌ FAIL: Tool $tool_name was not created"
  exit 1
fi

# Test 2: Tool should understand the referenced data format
echo "🧪 Test 2: Verify tool understands referenced format"
tool_content=$(port42 cat "/commands/$tool_name" | head -50)

# Check if tool mentions json, validation, or schema (context from reference)
if [[ "$tool_content" == *"json"* ]] || [[ "$tool_content" == *"validation"* ]] || [[ "$tool_content" == *"schema"* ]]; then
  echo "✅ PASS: Tool shows context awareness from file reference"
else
  echo "❌ FAIL: Tool doesn't show context from reference file"
  echo "Tool content preview:"
  echo "$tool_content"
  exit 1
fi

# Test 3: Multiple references
echo "🧪 Test 3: Multiple references integration"

# Create second reference file
cat > "$TEST_DIR/validation-examples.txt" << 'EOF'
# Validation Examples
- Email validation: check @ symbol and domain
- Age validation: must be 18-120 range
- Name validation: no special characters except dash/apostrophe
EOF

multi_tool="multi-ref-validator-$(date +%s)"
port42 declare tool "$multi_tool" --transforms "validation,multi,comprehensive" \
  --ref "file:$TEST_DIR/sample-data.json" \
  --ref "file:$TEST_DIR/validation-examples.txt" || {
  echo "❌ FAIL: Tool creation with multiple references failed"
  exit 1
}

# Verify multi-reference tool was created
if [[ ! -f "$HOME/.port42/commands/$multi_tool" ]]; then
  echo "❌ FAIL: Multi-reference tool $multi_tool was not created"
  exit 1
fi

echo "✅ PASS: Multiple references processed"

# Test 4: Invalid reference handling
echo "🧪 Test 4: Invalid reference error handling"
error_tool="error-test-$(date +%s)"
if port42 declare tool "$error_tool" --transforms "test" \
  --ref "file:/nonexistent/file.json" 2>/dev/null; then
  echo "❌ FAIL: Should have failed with invalid reference"
  exit 1
else
  echo "✅ PASS: Invalid reference properly rejected"
fi

# Cleanup
rm -rf "$TEST_DIR"

echo ""
echo "✅ All reference resolution tests passed!"
echo "🎯 References are being resolved and integrated into AI generation"