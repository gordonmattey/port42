#!/bin/bash

echo "🔍 Checking model configuration consistency..."
echo ""

# Check agents.json
echo "1️⃣ Model in agents.json:"
grep -A2 '"model_config"' /Users/gordon/Dropbox/Work/Hacking/workspace/port42/daemon/agents.json | grep '"default"'

# Check embedded fallback in agents.go
echo ""
echo "2️⃣ Embedded fallback in agents.go:"
grep -B1 -A3 "ModelConfig: ModelConfig" /Users/gordon/Dropbox/Work/Hacking/workspace/port42/daemon/agents.go | grep -E "(Default:|Opus:)"

# Check for any other hardcoded models
echo ""
echo "3️⃣ Searching for any other hardcoded model references:"
echo ""
echo "In daemon directory:"
grep -r "claude-[0-9]" /Users/gordon/Dropbox/Work/Hacking/workspace/port42/daemon/ --include="*.go" | grep -v "agents.json" | grep -v "Binary file"

echo ""
echo "In test scripts:"
grep -r "claude-[0-9]" /Users/gordon/Dropbox/Work/Hacking/workspace/port42/tests/ --include="*.sh" 2>/dev/null | head -5

echo ""
echo "4️⃣ Summary:"
echo "- Configuration should use agents.json as single source of truth"
echo "- Embedded fallback should match agents.json"
echo "- No hardcoded models elsewhere in code"

echo ""
echo "✅ Consistency check complete!"