# Port 42 Agile Architecture: Ship One Command, Evolve Once

## Core Philosophy: Each Command Drives Minimal Architecture

We ship one demo command at a time. Each command teaches us what architecture we actually need. No speculation, no over-engineering.

---

## Three Ways to Use Port 42

### 1. Create New Commands: `possess @ai-personality`
```bash
# Have a conversation to create a new command
possess @ai-engineer "I need a command that shows git activity"
possess @ai-muse "help me create something fun with my commits"
possess @ai-strategist "create a command for tracking team morale"
```
This is the **core Port 42 experience** - conversations that birth new commands.

### 2. Enhance Existing Commands: `possess <command>` (Level 10)
```bash
# Make system commands AI-aware
possess git     # Now: git "show me what changed last week"
possess vim     # Now: vim "that file with the bug"
possess cd      # Now: cd "the project I was working on yesterday"
```
This augments existing commands with natural language understanding.

### 3. Direct Intent: `possess @ai-assistant`
```bash
# Use AI assistant for complex orchestration
possess @ai-assistant "prep for standup"
  # -> Checks calendar, opens Zoom, gathers updates, opens notes

possess @ai-assistant "start my day"  
  # -> Opens calendar, triages email, shows Slack summary

possess @ai-assistant "I'm going into focus mode"
  # -> Mutes notifications, starts timer, opens relevant files
```
This orchestrates multiple commands and tools to achieve high-level goals.

---

## Level 0: Base Commands ✅
**Examples**: `git-haiku` ✅, `tweet-storm`, `pitch-writer`, `screen-record`, `sound-bite`  
**What they do**: Transform input into creative output  
**Architecture needed**: NONE - Current system works!  
**Lessons learned**: 
- Fixed shebang duplication bug
- Fixed string escaping for Python
- Command generation works across domains

---

## Level 1: Stateful Commands
**Examples**: `terminal-pet`, `content-calendar`, `meeting-maven`, `email-draft`, `auth-manager`  
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

## Level 2: Runtime AI Commands
**Examples**: `code-roast`, `email-alchemist`, `browse-for-me`, `meeting-scheduler`  
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

## Level 3: Multi-Context Commands
**Examples**: `pr-writer`, `competitor-scan`, `inbox-zero`, `slack-digest`  
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

## Level 4: Structured Context Commands
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

## Level 5: Time-Series Commands
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

## Level 6: Code Understanding Commands
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

## Level 7: Visualization Commands
**Examples**: `commit-time-travel`, `team-pulse`, `sales-pipeline`  
**What they do**: Create visual representations in the terminal  
**Architecture needed**: ASCII art generation

### Minimal Changes Required:
- No daemon changes needed
- Commands handle visualization internally
- Can leverage existing AI for creative ASCII

---

## Level 8: Cross-Format Commands
**Examples**: `code-translator`, `recipe-remix`, `doc-migrator`  
**What they do**: Convert between formats/languages  
**Architecture needed**: Format detection and conversion

### Minimal Changes Required:
- No daemon changes needed
- AI handles format understanding
- Commands focus on I/O handling

---

## Level 9: Meta-Awareness Commands
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

## Level 10: Command Augmentation
**Examples**: `possess vim`, `possess git`, `possess cd`  
**What they do**: AI-enhance existing system commands  
**Architecture needed**: Command interception and intelligent routing

### Minimal Changes Required:
```go
// Add possess endpoint that generates wrapper scripts
func (d *Daemon) handlePossess(cmd string) {
    // Generate wrapper script that intercepts the command
    wrapper := generateIntelligentWrapper(cmd)
    saveToPossessedCommands(cmd, wrapper)
}
```

### How it works:
```bash
# User runs: possess cd
# Creates: ~/.port42/possessed/cd that goes in PATH

#!/bin/bash
if [[ "$1" =~ ^[a-zA-Z] ]] && [[ ! -d "$1" ]]; then
    # Natural language detected
    target=$(port42 callback --context "cd" --query "$@")
    builtin cd "$target"
else
    # Normal cd behavior
    builtin cd "$@"
fi
```

### Example Usage:
```bash
possess cd
cd "that project from last week"      # AI: ~/projects/port42
cd "where I keep my photos"           # AI: ~/Pictures
cd meeting                            # AI: ~/work/meetings/2024-01-15

possess vim  
vim "the file with the bug"           # AI finds the right file
vim "my todo list"                    # AI: ~/.port42/todo.md

possess git
git "undo everything"                 # AI: git reset --hard HEAD
git "show me what changed"            # AI: git diff --stat
```

---

## The Pattern: Incremental Architecture

### Levels 0-1: File-Based State
- Commands read/write JSON files
- No daemon changes needed
- Proves the concept

### Levels 2-3: AI Integration
- One endpoint for AI callbacks
- Multi-file context handling
- Commands stay simple

### Levels 4-5: Rich Data Structures
- Structured JSON contexts
- Time-series append logs
- Still just HTTP + JSON

### Levels 6-10: Advanced Capabilities
- Semantic understanding
- Visualizations
- Cross-format conversions
- Meta-awareness
- Command augmentation

---

## What We DON'T Build Until Proven Necessary

❌ Complex state management systems  
❌ Command versioning  
❌ Sophisticated templating  
❌ Multi-user support  / Shared memories
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

### Level 0 - Base Commands
- 3+ commands work across domains
- 50+ successful generations
- git-haiku shared 10+ times

### Level 1 - Stateful
- 4 stateful commands live
- 100+ daily active users
- Zero state corruption issues

### Level 2 - Runtime AI
- 3 AI-callback commands
- <2s response times
- 90% user satisfaction

### Level 3 - Multi-Context
- Handles 10+ files gracefully
- pr-writer saves 5+ min/PR
- competitor-scan finds insights

### Level 4 - Structured
- Complex JSON contexts work
- debug-detective 80% accurate
- investor-intel saves hours

### Level 5 - Time-Series
- Handles 1000+ events/day
- standup-writer adopted by 5 teams
- expense tracking accurate

### Level 6 - Understanding
- Parses 5+ languages
- explain-this clarifies in <30s
- 95% explanation accuracy

### Level 7 - Visualization
- Beautiful ASCII output
- team-pulse in 10+ standups
- Screenshots shared widely

### Level 8 - Cross-Format
- 10+ format conversions
- Zero data loss
- 90% conversion accuracy

### Level 9 - Meta-Awareness
- Predicts user needs 70%+
- Measurable productivity gains
- Self-improving system

### Level 10 - Command Augmentation
- 10+ system commands enhanced
- Natural language understood 90%+
- Zero impact on normal usage
- "Magic" moments daily

---

## The Meta-Learning

Each command teaches us:
1. What architecture we actually need (not what we think we need)
2. How users actually use Port 42 (not how we imagine)
3. Which features drive adoption (not which seem cool)

**The best architecture emerges from shipping, not planning.**

---

## Command Complexity Levels Summary

| Level | Capability | Key Examples |
|-------|------------|--------------|
| 0 | Base | git-haiku, tweet-storm, screen-record |
| 1 | +State | terminal-pet, email-draft, auth-manager |
| 2 | +Runtime AI | code-roast, browse-for-me, email-writer |
| 3 | +Multi-Context | pr-writer, inbox-zero, slack-digest |
| 4 | +Structured Data | debug-detective, oauth-flow, api-composer |
| 5 | +Time-Series | standup-writer, screen-journal, comm-history |
| 6 | +Semantic Understanding | explain-this, figma-to-code, meeting-insights |
| 7 | +Visualization | commit-time-travel, calendar-map, inbox-flow |
| 8 | +Cross-Format | code-translator, meeting-to-doc, voice-to-task |
| 9 | +Meta-Awareness | terminal-consciousness, focus-mode, auto-responder |
| 10 | +Command Augmentation | possess cd, possess vim, possess chrome |

Each level represents increasing architectural complexity and capability.

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

### The Complete Digital Assistant Vision

With these system integrations, Port 42 evolves from a terminal tool to a comprehensive OS intelligence layer:

```bash
# Morning routine
possess @ai-assistant "start my day"
- Opens calendar with today's visual timeline
- Starts email triage with AI summaries  
- Shows Slack digest across all channels
- Begins screen recording for later review

# Deep work session
possess @ai-assistant "I need to focus on the API redesign"
- Mutes all notifications
- Starts focus timer based on your patterns
- Opens relevant files in vim
- Queues emails for batch response

# Meeting time  
possess @ai-assistant "prep for standup"
- Gathers what you worked on
- Drafts update from screen journal
- Opens Zoom with recording ready
- Pulls up yesterday's action items

# End of day
possess @ai-assistant "wrap up"
- Summarizes screen activity
- Drafts status update email
- Creates tomorrow's task list
- Archives today's contexts
```

Each intent leverages multiple levels of our architecture working together through the AI assistant.

### Architecture Debt We Accept (And When We Pay It)

**Levels 0-2: Hack It**
- Inline everything in commands
- Minimal error handling
- Manual testing only
- *Debt payment: After Level 3 - Extract common patterns*

**Levels 3-5: Extract Patterns**
- Shared modules for common operations
- Basic error handling library
- Integration test suite
- *Debt payment: After Level 6 - Build frameworks*

**Levels 6-8: Build Frameworks**
- Command SDK for each language
- Automated testing harness
- Documentation generation
- *Debt payment: After Level 8 - Production hardening*

**Levels 9-10: Production Ready**
- Performance optimization
- Security audit
- Monitoring and analytics
- Command augmentation framework
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
- Level 0: Command generation works across domains
- Level 1: State persistence is reliable
- Level 2: Runtime AI callbacks are fast
- Level 3: Multi-file context is manageable
- Level 4: Structured data improves quality
- Level 5: Time-series scales well
- Level 6: Code understanding is valuable
- Level 7: Visualizations enhance UX
- Level 8: Format conversion is feasible
- Level 9: Meta-learning delivers ROI
- Level 10: Command augmentation proves magical

**If any level proves unnecessary, we skip it.**

## Next Action: Ship Level 1 Commands

1. Create state directory handling
2. Generate pet with personality
3. Make it remember users
4. Ship it
5. Learn what breaks
6. Fix only what's broken

The dolphins swim in shallow waters first. 🐬