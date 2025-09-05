# Auto-Spawn Exit Status Bug Investigation

## Date: 2025-01-05

## Issue Description
When using `port42 possess @ai-engineer` to create tools, the command intermittently fails with "Command error: port42 command failed: exit status 1" even though the tool is successfully created.

## Symptoms
- Tool creation succeeds (tool is created and usable)
- Daemon logs show successful completion
- CLI reports exit status 1 error
- Issue appears intermittent (some tools succeed, some fail)
- Error occurs AFTER all operations complete successfully

## Root Cause Analysis

### Architecture Flow
1. User runs `port42 possess @ai-engineer "create a tool"`
2. CLI connects to daemon, sends possess request
3. Daemon calls Claude API
4. Claude responds with `run_command` tool use
5. Daemon executes `executePort42Command()` which:
   - Spawns a NEW port42 CLI process
   - Runs `port42 declare tool NAME ...`
6. NEW CLI process connects back to daemon
7. Daemon processes declare request, creates tool
8. When auto-spawn rule triggers, daemon spawns documentation
9. Daemon sends success response with spawned relation data
10. NEW CLI process attempts to deserialize response
11. CLI exits with status 1 if deserialization fails
12. Original daemon's `executePort42Command()` sees exit status 1

### The Circular Architecture
```
Original CLI → Daemon → Claude → Daemon spawns new CLI → New CLI connects back to Daemon
```

### Investigation Steps

#### 1. Log Analysis
Found pattern in daemon logs:
```
✅ Rule 'Auto-spawn documentation for complex tools' executed successfully
◊ Response sent [cli-xxxxx] success: true
◊ Consciousness disconnected
❌ Command execution failed: port42 command failed: exit status 1
```

#### 2. Compared Working vs Failing Cases

**Working (loc2):**
```
✅ Successfully spawned documentation: loc2-docs
✅ Rule 'Auto-spawn documentation for complex tools' executed successfully
◊ Response sent success: true
✅ [PORT42_CLI] Command completed successfully
```

**Failing (loc3, line-counter):**
```
✅ Successfully spawned documentation: loc3-docs
✅ Rule 'Auto-spawn documentation for complex tools' executed successfully
◊ Response sent success: true
❌ Command execution failed: port42 command failed: exit status 1
```

#### 3. Code Analysis

Examined `/daemon/src/possession.go`:
```go
output, err := cmd.CombinedOutput()
if err != nil {
    return string(output), fmt.Errorf("port42 command failed: %v", err)
}
```

Examined `/cli/src/commands/declare.rs`:
```rust
let declare_response = DeclareRelationResponse::parse_response(&data)?;
declare_response.display(OutputFormat::Plain)?;
```

The `DeclareRelationResponse` struct:
```rust
pub struct DeclareRelationResponse {
    pub relation_id: String,
    pub relation_type: String,
    pub materialized: bool,
    pub physical_path: String,
    pub status: String,
}
```

**No field for spawned relations!**

#### 4. Hypothesis Testing

Disabled auto-spawn rule in `/daemon/src/rules.go`:
```go
Enabled: false, // TEMPORARILY DISABLED FOR TESTING
```

Ran 5 tests:
```bash
for i in {1..5}; do
  port42 possess @ai-engineer "create a tool called loc-test-$i"
  echo "Exit code: $?"
done
```

Results: **All 5 succeeded with exit code 0**

## Root Cause
When the auto-spawn documentation rule triggers, the daemon includes additional fields about spawned relations in the response. The CLI's `DeclareRelationResponse` struct doesn't have fields for this data, causing deserialization to fail. Since Rust's main returns `Result<()>`, any error causes exit status 1.

## Why It's Intermittent
The issue depends on:
- Whether the tool has 3+ transforms (triggers auto-spawn)
- Possibly the complexity/size of spawned documentation
- Timing of response serialization/deserialization

## Temporary Fix
Disable the auto-spawn documentation rule in `rules.go`.

## Proper Fixes (TODO)

### Option 1: Update CLI Response Handling
Add spawned relations field to `DeclareRelationResponse`:
```rust
pub struct DeclareRelationResponse {
    pub relation_id: String,
    pub relation_type: String,
    pub materialized: bool,
    pub physical_path: String,
    pub status: String,
    pub spawned_relations: Option<Vec<SpawnedRelation>>, // New field
}
```

### Option 2: Separate Spawned Data
Have daemon send spawned relation data separately, not in the main declare response.

### Option 3: Make CLI More Tolerant
Configure serde to ignore unknown fields:
```rust
#[derive(Deserialize)]
#[serde(deny_unknown_fields = false)] // Allow unknown fields
pub struct DeclareRelationResponse { ... }
```

## Lessons Learned
1. The circular architecture (daemon spawning CLI that connects back to daemon) creates complex failure modes
2. Response schema mismatches between daemon and CLI cause silent failures
3. Testing with different rule configurations helps isolate issues
4. Exit status propagation through multiple layers obscures root causes

## Next Steps
1. Choose permanent fix approach
2. Add integration tests for auto-spawn scenarios
3. Consider refactoring to avoid daemon spawning CLI processes
4. Add better error reporting for deserialization failures