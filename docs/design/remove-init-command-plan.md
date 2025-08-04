# Plan: Remove Init Command and Consolidate Directory Creation

## Background
The `init` command is redundant and can cause permission issues when the installer runs as root but users run `init` later. The installers should be the single source of truth for directory creation.

## Directories Analysis

### Directories Created by Installer:
- `~/.port42/commands/` - Stores generated executable commands
- `~/.port42/memory/sessions/` - Stores conversation history as JSON files
- `~/.port42/artifacts/` - Stores generated artifacts (documents, apps, designs)
- `~/.port42/agents.json` - Agent configuration file (copied by installer)

### Directories Created by Daemon (automatically):
- `~/.port42/metadata/` - Created by daemon's Storage system for object metadata
- `~/.port42/objects/` - Created by daemon's Storage system for content-addressed storage

### Unused Directories (to remove from installer):
- `~/.port42/templates/` - Not implemented, only in docs/ideas
- `~/.port42/entities/` - Not used anywhere

## Decision: Single Source of Truth

**All directories should be created by the installer**, not split between installer and daemon.

## Implementation Plan

### 1. Update Both Installers
**Files:** `install.sh`, `install-local.sh`

Changes needed:
- Line ~201/52: Change from:
  ```bash
  mkdir -p "$PORT42_HOME"/{commands,memory/sessions,templates,entities}
  ```
  To:
  ```bash
  mkdir -p "$PORT42_HOME"/{commands,memory/sessions,artifacts,metadata,objects}
  ```
- Keep the same permission settings (755 for general, 700 for memory)
- Keep the sophisticated ownership checking in install.sh

### 2. Update Daemon Storage
**File:** `daemon/storage.go`

Change the Storage initialization to check if directories exist rather than always creating them:
- Change `os.MkdirAll` to first check if directories exist
- Log a warning if they don't exist (indicates installer wasn't run properly)
- Still create them if missing (for backward compatibility) but warn the user

### 3. Remove Init Command
**Files to modify:**
- `cli/src/main.rs` - Remove Init command from enum and handler
- `cli/src/commands/mod.rs` - Remove init module export
- `cli/src/help_text.rs` - Remove init-related constants

**Files to delete:**
- `cli/src/commands/init.rs`

### 4. Update Documentation
- Remove references to `port42 init` from README
- Update any setup instructions to only mention the installer
- Remove MSG_DIR_TEMPLATES from help_text.rs

### 5. Benefits
1. **Single source of truth** - Installer handles all setup
2. **No permission conflicts** - Avoids root/user ownership issues
3. **Simpler UX** - Users don't need to remember to run init
4. **Cleaner codebase** - Less redundant code
5. **Consistent setup** - All directories created at once with proper permissions

## Testing
1. Run installer as regular user - verify directories created with correct ownership
2. Run installer with sudo - verify ownership warning works
3. Start daemon and verify it creates metadata/ and objects/ directories
4. Verify daemon can write to all directories
5. Test command generation still works
6. Test memory persistence still works
7. Test artifact generation works
8. Test that `port42 info` works (uses metadata system)