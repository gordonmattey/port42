#!/bin/bash

# Focused tests for specific Unified Tool Hierarchy features
echo "🎯 Focused Tests: Unified Tool Hierarchy Key Features"
echo "===================================================="

# Test the core unified structure expectations
echo -e "\n📋 Expected New Structure:"
echo "/tools/"
echo "├── by-name/              # All tools by name"  
echo "├── by-transform/         # Group by capabilities"
echo "├── spawned-by/           # Relationship navigation"
echo "├── ancestry/             # Parent chains"
echo "└── {tool-name}/"
echo "    ├── definition        # Relation JSON"
echo "    ├── executable        # Physical tool"
echo "    ├── spawned/          # What this spawned"
echo "    └── parents/          # Parent chain"

echo -e "\n📋 Expected Breaking Changes:"
echo "❌ /relations/tools/ → replaced by /tools/by-name/"
echo "❌ /commands/ → replaced by /tools/{tool}/executable"  
echo "❌ /relations/ → removed from root"
echo "✅ /tools/ → new unified entry point"

echo -e "\n🧪 Quick Smoke Tests (will fail until implemented):"

echo -e "\n1. Root should show /tools/ not /relations/"
port42 ls / | grep -E "(tools|relations|commands)"

echo -e "\n2. /tools/ should exist and show unified structure"
port42 ls /tools/ 2>&1 | head -10

echo -e "\n3. Individual tool should have subpaths"
tool_name=$(ls ~/.port42/relations/relation-tool-*.json 2>/dev/null | head -1 | xargs basename -s .json | sed 's/relation-tool-[^-]*-[^-]*-//' | cut -d- -f1 2>/dev/null || echo "test-tool")
echo "Testing tool: $tool_name"
port42 ls /tools/$tool_name/ 2>&1

echo -e "\n4. Old paths should be gone or redirected"
echo "Testing old /relations/tools/:"
port42 ls /relations/tools/ 2>&1 | head -3
echo "Testing old /commands/:"  
port42 ls /commands/ 2>&1 | head -3

echo -e "\n5. Transform navigation should work"
port42 ls /tools/by-transform/ 2>&1
port42 ls /tools/by-transform/analysis/ 2>&1 | head -5

echo -e "\n6. Spawned-by navigation should work"  
port42 ls /tools/spawned-by/ 2>&1
# Try a tool that likely spawned something
analyzer_tool=$(ls ~/.port42/relations/relation-tool-*analyzer*.json 2>/dev/null | head -1 | xargs jq -r '.properties.name' 2>/dev/null || echo "")
if [[ -n "$analyzer_tool" ]]; then
    echo "Testing spawned-by for: $analyzer_tool"
    port42 ls /tools/spawned-by/$analyzer_tool/ 2>&1
fi

echo -e "\n📊 Implementation Status:"
echo "Most tests will fail - this shows what needs to be implemented"
echo "Success criteria: All /tools/ paths work, old paths fail gracefully"