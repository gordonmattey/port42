# Port 42 Agile Architecture: Ship One Command, Evolve Once

## Core Philosophy: Each Command Drives Minimal Architecture

We ship one demo command at a time. Each command teaches us what architecture we actually need. No speculation, no over-engineering.

---

## Level 0: Base Commands (Phase 1) ✅
**Examples**: `git-haiku` ✅, `tweet-storm`, `pitch-writer`  
**What they do**: Transform input into creative output  
**Architecture needed**: NONE - Current system works!  
**Lessons learned**: 
- Fixed shebang duplication bug
- Fixed string escaping for Python
- Command generation works across domains

---

## Level 1: Stateful Commands (Phase 2)
**Examples**: `terminal-pet`, `content-calendar`, `pitch-perfect`, `meeting-maven`  
**What they do**: Remember things between runs  
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

## Level 2: Runtime AI Commands (Phase 3)
**Examples**: `code-roast`, `email-alchemist`, `mood-dj`  
**What they do**: Get AI input during execution  
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

## Level 3: Multi-Context Commands (Phase 4)
**Examples**: `pr-writer`, `competitor-scan`, `project-health`  
**What they do**: Analyze multiple files/sources  
**Architecture needed**: Multi-file context handling

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

## Level 4: Structured Context Commands (Phase 5)
**Examples**: `debug-detective`, `investor-intel`, `expense-categorizer`  
**What they do**: Work with structured data  
**Architecture needed**: Structured JSON context

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

## Level 5: Time-Series Commands (Phase 6)
**Examples**: `standup-writer`, `campaign-tracker`, `expense-wizard`  
**What they do**: Track data over time  
**Architecture needed**: Append-only logs

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

## Level 6: Code Understanding Commands (Phase 7)
**Examples**: `explain-this`, `doc-generator`, `contract-analyzer`  
**What they do**: Parse and understand code/text semantically  
**Architecture needed**: Semantic parsing capabilities

### Minimal Changes Required:
```go
// Add endpoint to list command usage
func (d *Daemon) handleGetCommandStats(w http.ResponseWriter, r *http.Request) {
    stats := d.getCommandUsageStats()
    json.NewEncoder(w).Encode(stats)
}
```

---

## Level 7: Visualization Commands (Phase 8)
**Examples**: `commit-time-travel`, `team-pulse`, `sales-pipeline`  
**What they do**: Create visual representations in the terminal  
**Architecture needed**: ASCII art generation

### Minimal Changes Required:
- No daemon changes needed
- Commands handle visualization internally
- Can leverage existing AI for creative ASCII

---

## Level 8: Cross-Format Commands (Phase 9)
**Examples**: `code-translator`, `recipe-remix`, `doc-migrator`  
**What they do**: Convert between formats/languages  
**Architecture needed**: Format detection and conversion

### Minimal Changes Required:
- No daemon changes needed
- AI handles format understanding
- Commands focus on I/O handling

---

## Level 9: Meta-Awareness Commands (Phase 10)
**Examples**: `terminal-consciousness`, `deal-flow`, `productivity-coach`  
**What they do**: Learn from usage patterns  
**Architecture needed**: Usage analytics and pattern recognition

### Minimal Changes Required:
```go
// Add command usage tracking
func (d *Daemon) trackCommandUsage(command string, context map[string]interface{}) {
    // Append to usage log
}
```

---

## The Pattern: Incremental Architecture

### Phases 1-2: File-Based State
- Commands read/write JSON files
- No daemon changes needed
- Proves the concept

### Phases 3-4: Simple Callbacks
- One endpoint for AI callbacks
- Reuse existing AI client
- Commands stay simple

### Phases 5-6: Richer Context
- Extend callbacks with context
- Add activity logging
- Still just HTTP + JSON

### Phases 7+: Meta-Features
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

## Success Metrics Per Level

### Phase 1 (Level 0 - Base Commands)
- 3+ commands work across domains
- 50+ successful generations
- git-haiku shared 10+ times

### Phase 2 (Level 1 - Stateful)
- 4 stateful commands live
- 100+ daily active users
- Zero state corruption issues

### Phase 3 (Level 2 - Runtime AI)
- 3 AI-callback commands
- <2s response times
- 90% user satisfaction

### Phase 4 (Level 3 - Multi-Context)
- Handles 10+ files gracefully
- pr-writer saves 5+ min/PR
- competitor-scan finds insights

### Phase 5 (Level 4 - Structured)
- Complex JSON contexts work
- debug-detective 80% accurate
- investor-intel saves hours

### Phase 6 (Level 5 - Time-Series)
- Handles 1000+ events/day
- standup-writer adopted by 5 teams
- expense tracking accurate

### Phase 7 (Level 6 - Understanding)
- Parses 5+ languages
- explain-this clarifies in <30s
- 95% explanation accuracy

### Phase 8 (Level 7 - Visualization)
- Beautiful ASCII output
- team-pulse in 10+ standups
- Screenshots shared widely

### Phase 9 (Level 8 - Cross-Format)
- 10+ format conversions
- Zero data loss
- 90% conversion accuracy

### Phase 10 (Level 9 - Meta-Awareness)
- Predicts user needs 70%+
- Measurable productivity gains
- Self-improving system

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
Level 0: Base (command generation works)
  - git-haiku (dev): Git commits as poetry
  - tweet-storm (marketing): Ideas into tweet threads
  - pitch-writer (sales): Quick elevator pitches

Level 1: +State (JSON files)
  - terminal-pet (personal): Virtual pet that remembers you
  - content-calendar (marketing): Track published/scheduled content
  - pitch-perfect (sales): Remembers company/product details
  - meeting-maven (ops): Tracks action items across meetings

Level 2: +Runtime AI (HTTP callback)
  - code-roast (dev): Sassy code reviews
  - email-alchemist (sales): AI personalizes emails per recipient
  - mood-dj (personal): Suggests music based on your day

Level 3: +Multi-file Context (git/filesystem integration)
  - pr-writer (dev): Analyzes all changed files for PR description
  - competitor-scan (marketing): Analyzes multiple competitor sources
  - project-health (ops): Scans multiple project indicators

Level 4: +Structured Context (JSON context)
  - debug-detective (dev): Structured error analysis
  - investor-intel (fundraising): Structures investor research
  - expense-categorizer (ops): Auto-categorizes from receipts

Level 5: +Time-series (append-only logs)
  - standup-writer (dev): Daily activity summaries
  - campaign-tracker (marketing): Performance over time
  - expense-wizard (ops): Expense tracking with trends

Level 6: +Code Understanding (semantic parsing)
  - explain-this (dev): Explains code in context
  - doc-generator (dev): Creates docs from code structure
  - contract-analyzer (legal): Parses legal documents

Level 7: +Visualization (ASCII/terminal graphics)
  - commit-time-travel (dev): Visual git history
  - team-pulse (management): Team morale dashboard
  - sales-pipeline (sales): Visual funnel tracking

Level 8: +Cross-format/Cross-language
  - code-translator (dev): Converts between languages
  - recipe-remix (personal): Converts any recipe format
  - doc-migrator (ops): Converts between doc systems

Level 9: +Meta-awareness (usage patterns)
  - terminal-consciousness (dev): Knows your coding patterns
  - deal-flow (sales): Learns what closes deals
  - productivity-coach (personal): Suggests optimizations
```

### Why This Order Matters

1. **State before AI** - Prove we can persist data before adding complexity
2. **Simple AI before context** - One prompt/response before structured data
3. **Local before distributed** - File system before any network
4. **Read before write** - Analyze code before generating it
5. **Present before past** - Current state before history
6. **Single before meta** - One command before ecosystem

### Universal Architecture, Domain-Specific Magic

Notice how the same technical capabilities serve wildly different needs:

- **Level 1 State**: Pet personality vs. company details vs. meeting notes
- **Level 2 Runtime AI**: Code critique vs. email personalization vs. music moods
- **Level 3 Multi-file**: Git diffs vs. competitor websites vs. project files
- **Level 5 Time-series**: Code activity vs. marketing metrics vs. expense trends
- **Level 7 Visualization**: Git branches vs. team morale vs. sales funnels

The architecture is domain-agnostic. The magic happens when AI personalities understand each domain deeply.

### Architecture Debt We Accept (And When We Pay It)

**Phases 1-3: Hack It (Levels 0-2)**
- Inline everything in commands
- Minimal error handling
- Manual testing only
- *Debt payment: Phase 4 - Extract state patterns*

**Phases 4-6: Extract Patterns (Levels 3-5)**
- Shared modules for common operations
- Basic error handling library
- Integration test suite
- *Debt payment: Phase 7 - Build frameworks*

**Phases 7-9: Build Frameworks (Levels 6-8)**
- Command SDK for each language
- Automated testing harness
- Documentation generation
- *Debt payment: Phase 10 - Production hardening*

**Phase 10: Production Ready (Level 9)**
- Performance optimization
- Security audit
- Monitoring and analytics
- Ready for wide adoption

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

**Each level de-risks the next:**
- Level 0 (Phase 1): Command generation works across domains
- Level 1 (Phase 2): State persistence is reliable
- Level 2 (Phase 3): Runtime AI callbacks are fast
- Level 3 (Phase 4): Multi-file context is manageable
- Level 4 (Phase 5): Structured data improves quality
- Level 5 (Phase 6): Time-series scales well
- Level 6 (Phase 7): Code understanding is valuable
- Level 7 (Phase 8): Visualizations enhance UX
- Level 8 (Phase 9): Format conversion is feasible
- Level 9 (Phase 10): Meta-learning delivers ROI

**If any phase fails, we pivot immediately.**

## Next Action: Ship Level 1 Commands

1. Create state directory handling
2. Generate pet with personality
3. Make it remember users
4. Ship it
5. Learn what breaks
6. Fix only what's broken

The dolphins swim in shallow waters first. 🐬