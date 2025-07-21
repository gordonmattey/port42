# Port 42 Agile Architecture: Ship One Command, Evolve Once

## Core Philosophy: Each Command Drives Minimal Architecture

We ship one demo command at a time. Each command teaches us what architecture we actually need. No speculation, no over-engineering.

---

## Command 1: `git-haiku` ✅ (Week 1)
**What it does**: Transforms git commits into haikus  
**Architecture needed**: NONE - Current system works!  
**Lessons learned**: 
- Fixed shebang duplication bug
- Fixed string escaping for Python
- Command generation works for simple cases

---

## Command 2: `terminal-pet` (Week 2)
**What it does**: ASCII pet that remembers you between runs  
**Architecture needed**: Simple state persistence

### Minimal Changes Required:
```go
// Just add a state directory for commands
~/.port42/command-state/terminal-pet/pet.json
```

### Implementation:
```python
# In the generated command
import json
from pathlib import Path

STATE_FILE = Path.home() / '.port42' / 'command-state' / 'terminal-pet' / 'pet.json'
STATE_FILE.parent.mkdir(parents=True, exist_ok=True)

def load_pet():
    if STATE_FILE.exists():
        return json.loads(STATE_FILE.read_text())
    return {"name": "Whiskers", "mood": "happy", "last_seen": None}

def save_pet(pet):
    STATE_FILE.write_text(json.dumps(pet))
```

**That's it!** No complex state management system. Just JSON files.

---

## Command 3: `code-roast` (Week 3)
**What it does**: Sarcastic code reviews  
**Architecture needed**: Runtime AI callbacks

### Minimal Changes Required:
```go
// Add ONE endpoint to daemon
func (d *Daemon) handleCommandCallback(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Prompt string `json:"prompt"`
        Agent  string `json:"agent"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    // Reuse existing AI client
    response := d.callAI(req.Agent, req.Prompt)
    json.NewEncoder(w).Encode(map[string]string{"response": response})
}
```

### In Commands:
```bash
# Simple curl to daemon
ROAST=$(curl -s localhost:42/callback -d "{
    \"agent\": \"@ai-muse\",
    \"prompt\": \"Roast this code: $CODE\"
}")
```

**That's it!** Reuse existing AI infrastructure.

---

## Command 4: `pr-writer` (Week 4)
**What it does**: Analyzes changes and writes PR descriptions  
**Architecture needed**: Better git integration

### Minimal Changes Required:
- Add git diff parsing to command generation prompts
- No daemon changes needed!

### Implementation:
```bash
# The command just pipes git data to AI
git diff --staged | curl -s localhost:42/callback -d "{
    \"agent\": \"@ai-engineer\",
    \"prompt\": \"Write a PR description for these changes: $(cat)\"
}"
```

---

## Command 5: `debug-detective` (Week 5)
**What it does**: Investigates bugs using context  
**Architecture needed**: Multi-file context

### Minimal Changes Required:
```go
// Extend callback to accept more context
type CallbackRequest struct {
    Prompt  string                 `json:"prompt"`
    Agent   string                 `json:"agent"`
    Context map[string]interface{} `json:"context"` // NEW
}
```

### In Commands:
```python
context = {
    "error": error_msg,
    "recent_commits": get_recent_commits(),
    "changed_files": get_changed_files()
}
response = requests.post("http://localhost:42/callback", json={
    "agent": "@ai-engineer",
    "prompt": "Investigate this bug",
    "context": context
})
```

---

## Command 6: `standup-writer` (Week 6)
**What it does**: Tracks your daily activity  
**Architecture needed**: Activity collection

### Minimal Changes Required:
```go
// Add activity log endpoint
func (d *Daemon) handleLogActivity(w http.ResponseWriter, r *http.Request) {
    // Append to ~/.port42/activity.jsonl
    // That's it!
}
```

### Commands can read the log:
```python
# Read activity log
with open(Path.home() / '.port42' / 'activity.jsonl') as f:
    activities = [json.loads(line) for line in f]
```

---

## Command 7: `terminal-consciousness` (Week 7)
**What it does**: Meta-awareness of your terminal state  
**Architecture needed**: Command introspection

### Minimal Changes Required:
```go
// Add endpoint to list command usage
func (d *Daemon) handleGetCommandStats(w http.ResponseWriter, r *http.Request) {
    stats := d.getCommandUsageStats()
    json.NewEncoder(w).Encode(stats)
}
```

---

## The Pattern: Incremental Architecture

### Week 1-2: File-Based State
- Commands read/write JSON files
- No daemon changes needed
- Proves the concept

### Week 3-4: Simple Callbacks
- One endpoint for AI callbacks
- Reuse existing AI client
- Commands stay simple

### Week 5-6: Richer Context
- Extend callbacks with context
- Add activity logging
- Still just HTTP + JSON

### Week 7+: Meta-Features
- Command introspection
- Usage analytics
- Collective intelligence

---

## What We DON'T Build Until Proven Necessary

❌ Complex state management systems  
❌ Command versioning  
❌ Sophisticated templating  
❌ Multi-user support  
❌ Command composition frameworks  
❌ Abstract capability systems  

**Build only what the next command needs.**

---

## Migration Strategy

As patterns emerge, we refactor:

1. **Multiple commands use state?** → Extract state library
2. **Complex callbacks emerge?** → Build callback framework  
3. **Commands need composition?** → Add composition support

But NOT before we have 3+ real examples proving the need.

---

## Success Metrics Per Week

### Week 1 (git-haiku)
- Works on first try for 10 users
- Gets shared on Twitter

### Week 2 (terminal-pet)
- Users run it daily
- Pet state persists correctly

### Week 3 (code-roast)
- Generates funny roasts
- Screenshots go viral

### Week 4 (pr-writer)
- Saves developers 5 minutes
- PRs actually merge

### Week 5 (debug-detective)
- Finds real bugs
- "How did it know?!" moments

### Week 6 (standup-writer)
- Teams adopt it
- Managers love it

### Week 7 (terminal-consciousness)
- Mind = Blown
- "The future is here"

---

## The Meta-Learning

Each command teaches us:
1. What architecture we actually need (not what we think we need)
2. How users actually use Port 42 (not how we imagine)
3. Which features drive adoption (not which seem cool)

**The best architecture emerges from shipping, not planning.**

---

## Deep Dive: Architectural Evolution Pattern

### The Staircase of Complexity

Each command adds EXACTLY ONE new architectural capability:

```
Level 0: git-haiku         → Base (command generation works)
Level 1: terminal-pet      → +State (JSON files)
Level 2: code-roast        → +Runtime AI (HTTP callback)
Level 3: pr-writer         → +Multi-file Context (git integration)
Level 4: debug-detective   → +Structured Context (JSON context)
Level 5: standup-writer    → +Time-series (append-only logs)
Level 6: explain-this      → +Code Understanding (line mapping)
Level 7: commit-time-travel → +Visualization (ASCII art)
Level 8: code-translator   → +Cross-language (AI handles it)
Level 9: terminal-consciousness → +Meta-awareness (usage stats)
```

### Why This Order Matters

1. **State before AI** - Prove we can persist data before adding complexity
2. **Simple AI before context** - One prompt/response before structured data
3. **Local before distributed** - File system before any network
4. **Read before write** - Analyze code before generating it
5. **Present before past** - Current state before history
6. **Single before meta** - One command before ecosystem

### Architecture Debt We Accept (And When We Pay It)

**Week 1-3: Hack It**
- Inline everything in commands
- No error handling
- No tests
- *Debt payment: Week 4 - Extract common patterns*

**Week 4-6: Extract Patterns**
- Shared Python module for state
- Bash function library  
- Common error handling
- *Debt payment: Week 7 - Build framework*

**Week 7-9: Build Framework**
- Command SDK
- Testing harness
- Documentation generator
- *Debt payment: Week 10 - Production ready*

### The "You Aren't Gonna Need It" List

Based on similar projects, here's what we WON'T build:

1. **GraphQL API** - HTTP + JSON is enough forever
2. **Command marketplace** - GitHub is the marketplace  
3. **Web UI** - Terminal-first or death
4. **Plugins** - Commands ARE the plugins
5. **Kubernetes** - It's a localhost daemon, period
6. **Microservices** - Monolith until 1M users
7. **Event sourcing** - Append-only logs are enough
8. **OAuth** - Local only, no cloud

### Risk Mitigation Through Incremental Shipping

**Each week de-risks the next:**
- Week 1: Proves command generation works
- Week 2: Proves state management pattern  
- Week 3: Proves AI integration pattern
- Week 4: Proves we can handle real workflows
- Week 5: Proves complex reasoning works
- Week 6: Proves we can track usage
- Week 7: Proves the meta-concept

**If any week fails, we pivot immediately.**

## Next Action: Ship terminal-pet This Week

1. Create state directory handling
2. Generate pet with personality
3. Make it remember users
4. Ship it
5. Learn what breaks
6. Fix only what's broken

The dolphins swim in shallow waters first. 🐬