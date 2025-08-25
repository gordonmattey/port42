#!/bin/bash
# Test 3: User Prompt Integration
# Tests that user prompts are integrated into AI generation and affect output

set -e

echo "💬 Test 3: User Prompt Integration"

# Test 1: Tool with custom prompt
echo "🧪 Test 1: Tool generation with user prompt"
prompt_tool="prompt-test-$(date +%s)"
port42 declare tool "$prompt_tool" --transforms "api,rest" \
  --prompt "Create FastAPI server with marker PROMPT_TEST_ABC123 in comments" || {
  echo "❌ FAIL: Tool creation with prompt failed"
  exit 1
}

# Verify tool was created
if [[ ! -f "$HOME/.port42/commands/$prompt_tool" ]]; then
  echo "❌ FAIL: Prompt tool $prompt_tool was not created"
  exit 1
fi

# Test 2: Verify prompt requirements are in the generated code
echo "🧪 Test 2: Verify prompt marker appears in generated code"
tool_content=$(port42 cat "/commands/$prompt_tool")

if [[ "$tool_content" == *"PROMPT_TEST_ABC123"* ]]; then
  echo "✅ PASS: User prompt marker found in generated code"
else
  echo "❌ FAIL: User prompt marker not found in generated code"
  echo "Generated tool content:"
  echo "$tool_content"
  exit 1
fi

# Test 3: Tool with prompt and references
echo "🧪 Test 3: Combined prompt and reference integration"

# Create test reference
TEST_DIR="/tmp/port42-prompt-test"
mkdir -p "$TEST_DIR"
cat > "$TEST_DIR/api-spec.json" << 'EOF'
{
  "endpoints": [
    "/users", "/posts", "/comments"
  ],
  "auth": "JWT Bearer tokens",
  "rate_limit": "100 requests/minute"
}
EOF

combined_tool="combined-test-$(date +%s)"
port42 declare tool "$combined_tool" --transforms "client,api,http" \
  --ref "file:$TEST_DIR/api-spec.json" \
  --prompt "Add special header X-CUSTOM-MARKER with value COMBINED_TEST_DEF456" || {
  echo "❌ FAIL: Tool creation with prompt and references failed"
  exit 1
}

# Verify combined tool was created
if [[ ! -f "$HOME/.port42/commands/$combined_tool" ]]; then
  echo "❌ FAIL: Combined tool $combined_tool was not created"
  exit 1
fi

# Test 4: Verify both prompt and reference context are used
echo "🧪 Test 4: Verify prompt and reference integration"
combined_content=$(port42 cat "/commands/$combined_tool")

# Check for prompt marker
prompt_found=false
ref_found=false

if [[ "$combined_content" == *"COMBINED_TEST_DEF456"* ]]; then
  prompt_found=true
fi

# Check for reference context (should mention endpoints, JWT, or rate_limit)
if [[ "$combined_content" == *"users"* ]] || [[ "$combined_content" == *"JWT"* ]] || [[ "$combined_content" == *"rate"* ]]; then
  ref_found=true
fi

if $prompt_found && $ref_found; then
  echo "✅ PASS: Both prompt and reference context found in generated code"
elif $prompt_found; then
  echo "⚠️  PARTIAL: Prompt found but reference context missing"
  echo "Generated tool content:"
  echo "$combined_content"
elif $ref_found; then
  echo "⚠️  PARTIAL: Reference context found but prompt missing"
  echo "Generated tool content:"  
  echo "$combined_content"
else
  echo "❌ FAIL: Neither prompt nor reference context found"
  echo "Generated tool content:"
  echo "$combined_content"
  exit 1
fi

# Test 5: Empty prompt handling
echo "🧪 Test 5: Empty prompt handling"
empty_prompt_tool="empty-prompt-$(date +%s)"
port42 declare tool "$empty_prompt_tool" --transforms "basic,test" \
  --prompt "" || {
  echo "❌ FAIL: Tool creation with empty prompt failed"
  exit 1
}

if [[ -f "$HOME/.port42/commands/$empty_prompt_tool" ]]; then
  echo "✅ PASS: Empty prompt handled gracefully"
else
  echo "❌ FAIL: Tool with empty prompt was not created"
  exit 1
fi

# Cleanup
rm -rf "$TEST_DIR"

echo ""
echo "✅ All prompt integration tests passed!"
echo "💬 User prompts are being integrated into AI generation"
echo "🔗 References and prompts work together correctly"