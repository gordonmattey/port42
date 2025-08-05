# Shell Command Execution Design

**Purpose**: Enable direct execution of Port 42 commands and system commands from the Port 42 shell
**Scope**: Modify shell to support PATH-like command resolution with Port 42 commands taking precedence

## Current State

The Port 42 shell currently:
- Handles built-in Port 42 commands (possess, memory, reality, etc.)
- Executes Port 42 generated commands from `~/.port42/commands/`
- Executes system commands with full argument support
- Uses `!` prefix to force system commands when names conflict
- Does NOT support shell metacharacters (pipes, redirections, expansions)

## Desired State

Users should be able to:
1. Run any Port 42 command directly: `port42> git-haiku`
2. Run system commands: `port42> ls -la`
3. Have Port 42 commands override system commands when names conflict
4. Use all arguments and pipes normally

## Design Approach

### Command Resolution Order

When a user types a command in the Port 42 shell:

1. **Check built-in commands** (current behavior)
   - possess, memory, reality, status, etc.
   - These always take precedence

2. **Check Port 42 commands** (new)
   - Look in `~/.port42/commands/`
   - If found and executable, run it

3. **Check system commands** (new)
   - Attempt to execute via system PATH
   - Includes all normal shell commands

4. **Show error** if none found

### Handling Command Conflicts

For commands that exist in both Port 42 built-ins and as system commands (like `ls`):

#### Option 1: Built-ins Always Win (Current Behavior)
- `ls` always runs Port 42's virtual filesystem ls
- Use `/bin/ls` to explicitly run system ls

#### Option 2: Context-Aware Resolution
- `ls` runs Port 42 ls when given virtual paths (/memories, /sessions)
- `ls` runs system ls for regular paths
- `!ls` forces system ls
- `@ls` forces Port 42 ls

#### Option 3: Prefix for Port 42 Commands
- `ls` runs system ls
- `p42 ls` or `@ls` runs Port 42 ls
- More explicit but changes current behavior

#### Recommended: Escape Prefix for System Commands
- Keep current behavior (Port 42 built-ins win)
- Add escape prefix `!` to force system command
- `ls` → Port 42 virtual filesystem ls
- `!ls` → System ls
- `!git status` → System git (even if you have a Port 42 git)

### Implementation Details

#### File: `cli/src/shell.rs`

Modify the `execute_command` method's default case (currently lines 294-297):

```rust
_ => {
    // Try to execute as Port 42 command or system command
    if let Err(e) = self.execute_external_command(parts) {
        println!("{}", format_unknown_command(parts[0]).red());
        println!("{}", MSG_SHELL_HELP_HINT.dimmed());
    }
}
```

Add new method:

```rust
fn execute_external_command(&self, parts: &[&str]) -> Result<()> {
    use std::process::Command;
    use std::path::PathBuf;
    
    if parts.is_empty() {
        return Ok(());
    }
    
    let command_name = parts[0];
    let args = &parts[1..];
    
    // Check for escape prefix to force system command
    let (force_system, actual_command) = if command_name.starts_with('!') {
        (true, &command_name[1..])
    } else {
        (false, command_name)
    };
    
    // If not forcing system command, check Port 42 commands first
    if !force_system {
        let port42_cmd_path = dirs::home_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join(".port42")
            .join("commands")
            .join(actual_command);
        
        if port42_cmd_path.exists() && port42_cmd_path.is_file() {
            // Execute Port 42 command
            let status = Command::new(&port42_cmd_path)
                .args(args)
                .status()?;
                
            if !status.success() {
                if let Some(code) = status.code() {
                    return Err(anyhow::anyhow!("Command exited with code {}", code));
                }
            }
            return Ok(());
        }
    }
    
    // Try system command
    match Command::new(actual_command).args(args).status() {
        Ok(status) => {
            if !status.success() {
                if let Some(code) = status.code() {
                    return Err(anyhow::anyhow!("Command exited with code {}", code));
                }
            }
            Ok(())
        }
        Err(e) => {
            // Command not found in system PATH
            Err(anyhow::anyhow!("Command not found: {}", actual_command))
        }
    }
}
```

## Implementation Steps

1. **Add the `execute_external_command` method** to `Port42Shell`
   - Handles both Port 42 and system command execution
   - Returns proper errors for better error handling

2. **Update the match statement** in `execute_command`
   - Replace the default case to call `execute_external_command`
   - Preserve error formatting

3. **Test command execution**
   - Port 42 commands work with arguments
   - System commands work normally
   - Error messages are helpful

4. **Handle edge cases**
   - Commands with spaces in arguments
   - Commands that need stdin/stdout
   - Signal handling (Ctrl+C)

## Usage Examples

```bash
# Port 42 built-in commands (always win)
port42> ls                    # Port 42 virtual filesystem ls
port42> ls /memories          # Port 42 virtual filesystem ls
port42> cat /sessions/abc123  # Port 42 virtual cat

# Force system commands with !
port42> !ls                   # System ls
port42> !ls -la              # System ls with flags
port42> !cat /etc/hosts      # System cat

# Port 42 generated commands (from ~/.port42/commands/)
port42> git-haiku            # Runs ~/.port42/commands/git-haiku
port42> rainbow-art "Hello"  # Runs ~/.port42/commands/rainbow-art

# System commands (when no Port 42 equivalent exists)
port42> grep "pattern" file.txt
port42> ps aux | grep port42

# Mixed usage
port42> !ls | grep ".txt"    # System ls piped to system grep
port42> ls /memories | !grep "2024"  # Port 42 ls piped to system grep

# If you possessed a system command
port42> cd                   # Port 42 AI-enhanced cd (if exists)
port42> !cd                  # Force system cd
```

## Security Considerations

1. **Path validation**: Only execute from `~/.port42/commands/`, not arbitrary paths
2. **No shell expansion**: Don't interpret shell metacharacters in Port 42 paths
3. **Permission checks**: Ensure commands are executable before running
4. **Command injection**: Use `Command::new()` API which handles escaping

## Future Enhancements

1. **Shell Metacharacter Support** (HIGH PRIORITY)
   - Current limitation: Commands with pipes, redirections, expansions fail
   - Root cause: Direct `Command::new()` execution vs shell interpretation
   - Solution: Hybrid approach with metacharacter detection
   ```rust
   fn needs_shell_interpretation(input: &str) -> bool {
       input.contains('|') || input.contains('~') || input.contains('>') 
           || input.contains('<') || input.contains('&') || input.contains('$')
           || input.contains('*') || input.contains('?')
   }
   ```
   - If metacharacters detected, delegate to `/bin/sh -c "command"`
   - Otherwise use direct execution for performance/control

2. **Tab completion** for Port 42 commands
3. **Command aliases** support
4. **Environment variable** support for Port 42 commands
5. **Background execution** with `&`

## Testing Plan

1. Create test Port 42 command: `~/.port42/commands/test-cmd`
2. Test execution with various arguments
3. Test system command execution
4. Test error cases (non-existent commands)
5. Test commands with same name as system commands
6. Test signal handling during command execution

## Implementation Status

✅ **COMPLETED** (v1.0)
- Full command execution (Port 42 commands, system commands)
- `!` prefix for forcing system commands
- Command arguments work perfectly
- Error handling and command resolution

⚠️ **KNOWN LIMITATIONS**
- Shell metacharacters not supported (`|`, `~`, `>`, `<`, `&`, `$`, `*`, `?`)
- Complex shell operations like `!ls | grep port` fail because metacharacters are treated as literal arguments
- Simple commands with arguments work great: `!ls -la`, `!grep -n "pattern" file.txt`
- Workaround: Use system shell directly for piping/redirection

## Impact

- **No breaking changes**: All existing shell commands continue to work
- **Added functionality**: Can now run any command from the shell with full argument support
- **Natural usage**: No special syntax or prefixes needed