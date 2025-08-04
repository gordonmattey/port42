# AI Tool Integration - Implementation Guide

**Purpose**: Exact implementation steps for adding command execution tools without breaking existing functionality
**Scope**: Where to make changes, what to preserve, minimal implementation

## Current System Analysis

### 1. System Prompt Generation (daemon/agents.go)
```go
func GetAgentPrompt(agentName string) string {
    // Currently builds prompt from:
    // 1. agent.Prompt (base prompt)
    // 2. BaseGuidance.ArtifactGuidance (if has tools)
    // 3. BaseGuidance.FormatTemplate + Implementation (if creates commands)
    // 4. agent.Suffix
    
    // We need to insert available commands AFTER the base prompt
    // but BEFORE the artifact guidance
}
```

### 2. Tool Provisioning (daemon/possession.go)
```go
// Currently in Send() method around line 163:
if agentInfo.NoImplementation {
    tools = []AnthropicTool{
        getArtifactGenerationTool(),
    }
} else {
    tools = []AnthropicTool{
        getCommandGenerationTool(),
        getArtifactGenerationTool(),
    }
}
// We need to ADD command runner tool here for ALL agents
```

## Implementation Changes

### Change 1: Add Command Runner Tool (daemon/possession.go)

Add new tool definition:
```go
// Add after getArtifactGenerationTool()
func getCommandRunnerTool() AnthropicTool {
    return AnthropicTool{
        Name:        "run_command",
        Description: "Run a previously generated Port 42 command",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "command": map[string]interface{}{
                    "type":        "string",
                    "description": "Command name (e.g., rainbow-art, git-haiku)",
                },
                "args": map[string]interface{}{
                    "type":        "array",
                    "items":       map[string]interface{}{"type": "string"},
                    "description": "Command arguments",
                },
                "stdin": map[string]interface{}{
                    "type":        "string",
                    "description": "Optional input to pipe to the command",
                },
            },
            "required": []string{"command"},
        },
    }
}
```

Update tool provisioning in Send():
```go
// Around line 173-186, modify to:
if agentInfo.NoImplementation {
    tools = []AnthropicTool{
        getCommandRunnerTool(),        // ADD THIS
        getArtifactGenerationTool(),
    }
} else {
    tools = []AnthropicTool{
        getCommandRunnerTool(),        // ADD THIS
        getCommandGenerationTool(),
        getArtifactGenerationTool(),
    }
}
```

### Change 2: Update System Prompt (daemon/agents.go)

Modify GetAgentPrompt function:
```go
func GetAgentPrompt(agentName string) string {
    // ... existing code up to prompt.WriteString(agent.Prompt) ...
    
    // ADD THIS SECTION after agent.Prompt:
    // List available commands
    commands := listAvailableCommands()
    if len(commands) > 0 {
        prompt.WriteString("\n\n<available_commands>")
        prompt.WriteString("\nYou have access to these Port 42 commands via the run_command tool:")
        for _, cmd := range commands {
            prompt.WriteString(fmt.Sprintf("\n- %s: %s", cmd.Name, cmd.Description))
        }
        prompt.WriteString("\n</available_commands>")
        prompt.WriteString("\nUse run_command to execute any of these when they would be helpful.")
    }
    
    // ... rest of existing code (artifact guidance, etc.) ...
}
```

Add helper function:
```go
type CommandMetadata struct {
    Name        string
    Description string
}

func listAvailableCommands() []CommandMetadata {
    commandsDir := filepath.Join(os.Getenv("HOME"), ".port42/commands")
    var commands []CommandMetadata
    
    files, err := ioutil.ReadDir(commandsDir)
    if err != nil {
        return commands
    }
    
    for _, file := range files {
        if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
            continue
        }
        
        // For MVP, just use filename as name
        // Later we can read metadata from file header
        commands = append(commands, CommandMetadata{
            Name:        file.Name(),
            Description: "User-generated command",
        })
    }
    
    return commands
}
```

### Change 3: Handle Tool Execution (daemon/possession.go)

Add handler after extractArtifactSpecFromToolCall:
```go
// Add around line 400 in handlePossessWithAI
} else if content.Type == "tool_use" && content.Name == "run_command" {
    // Handle command execution
    if output, err := executeCommand(content.Input); err == nil {
        // Include command output in response
        responseText += fmt.Sprintf("\n\nCommand output:\n%s", output)
    } else {
        responseText += fmt.Sprintf("\n\nCommand error: %v", err)
    }
```

Add execution function:
```go
func executeCommand(input json.RawMessage) (string, error) {
    var params struct {
        Command string   `json:"command"`
        Args    []string `json:"args"`
        Stdin   string   `json:"stdin"`
    }
    
    if err := json.Unmarshal(input, &params); err != nil {
        return "", fmt.Errorf("invalid parameters: %v", err)
    }
    
    // Security: verify command exists
    cmdPath := filepath.Join(os.Getenv("HOME"), ".port42/commands", params.Command)
    if _, err := os.Stat(cmdPath); err != nil {
        return "", fmt.Errorf("command not found: %s", params.Command)
    }
    
    // Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, cmdPath, params.Args...)
    if params.Stdin != "" {
        cmd.Stdin = strings.NewReader(params.Stdin)
    }
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return string(output), fmt.Errorf("execution failed: %v", err)
    }
    
    return string(output), nil
}
```

### Change 4: Add Required Imports

In daemon/possession.go:
```go
import (
    "context"
    "os/exec"
    "path/filepath"
    "io/ioutil"
    // ... existing imports
)
```

## No CLI Changes Required

The CLI doesn't need any changes because:
1. Tool execution happens daemon-side
2. System prompt is built daemon-side
3. Response handling already supports tool outputs

## Testing Plan

1. **Test existing functionality works**:
   - Generate a command normally
   - Generate an artifact normally
   - Verify no regression

2. **Test new functionality**:
   - Generate a simple command (e.g., `echo-test`)
   - In new conversation, ask AI to use it
   - Verify AI can list and execute commands

3. **Test cross-agent**:
   - Create command with @ai-muse
   - Use command with @ai-engineer
   - Verify it works

## Security Considerations

1. **Path validation**: Only run commands from ~/.port42/commands/
2. **Timeout**: 30 second execution limit
3. **No shell expansion**: Direct execution only
4. **Output limits**: Consider truncating large outputs

## Future Enhancements

1. **Command metadata**: Read description from file header
2. **List commands tool**: Let AI discover commands dynamically
3. **Usage tracking**: Log which commands are used
4. **Output formatting**: Better integration of command output

## Summary

This implementation:
- Adds `run_command` tool to ALL agents
- Updates system prompt to list available commands
- Handles command execution securely
- Requires NO CLI changes
- Preserves all existing functionality