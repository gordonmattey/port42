# AI Tool Integration - Using Generated Commands as AI Tools

**Purpose**: Design for making Port 42 generated commands available as tools the AI can use
**Scope**: How commands become tools, implementation approach, security considerations

## Vision

Once you generate a command through Port 42, it becomes part of your personal AI toolkit. The AI can then use these commands in ANY future conversation, with ANY agent, creating a feedback loop of capability growth. Your AI gets permanently smarter with every command you create.

## Example Flow

```bash
# Monday: Generate a command with @ai-muse
possess @ai-muse "create a command that makes text into rainbow ASCII art"
# → Creates rainbow-art command

# Tuesday: Use it with @ai-engineer
possess @ai-engineer "can you create a colorful README header for my project?"
# AI responds: "I'll use the rainbow-art command to create that for you..."
# → AI calls rainbow-art and shows the output

# Wednesday: Use it with @ai-growth
possess @ai-growth "make our product announcement more eye-catching"
# AI responds: "I'll use rainbow-art to create attention-grabbing headers..."
# → Same command, different agent, different context
```

## Key Principle: Global Command Registry

Every command generated becomes permanently available to ALL agents in ALL conversations. This creates a true "personal AI operating system" where capabilities accumulate over time.

## Implementation Approaches

### Approach 1: Dynamic Tool Registration

Every generated command becomes a callable tool:

```go
// When a command is generated
func (d *Daemon) registerCommandAsTool(spec *CommandSpec) {
    toolName := fmt.Sprintf("run_%s", spec.Name)
    
    d.availableTools[toolName] = AnthropicTool{
        Name:        toolName,
        Description: fmt.Sprintf("Execute %s: %s", spec.Name, spec.Description),
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "args": map[string]interface{}{
                    "type":        "array",
                    "items":       map[string]interface{}{"type": "string"},
                    "description": "Command line arguments",
                },
                "stdin": map[string]interface{}{
                    "type":        "string",
                    "description": "Optional input to pipe to the command",
                },
            },
        },
    }
}
```

### Approach 2: Generic Command Runner (Simpler)

One tool that can run any Port 42 command:

```go
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
                    "description": "Input to pipe to the command",
                },
            },
            "required": []string{"command"},
        },
    }
}
```

## Tool Discovery

The AI needs to know what commands are available across all conversations:

### Dynamic System Prompt (Recommended)
Update the system prompt dynamically with available commands:

```go
func getAgentPrompt(agent string) string {
    basePrompt := GetAgentPrompt(agent)
    
    // Dynamically append ALL available commands
    commands := listAvailableCommands()
    if len(commands) > 0 {
        basePrompt += "\n\n<available_commands>\n"
        basePrompt += "You have access to these Port 42 commands via the run_command tool:\n"
        for _, cmd := range commands {
            basePrompt += fmt.Sprintf("- %s: %s\n", cmd.Name, cmd.Description)
        }
        basePrompt += "</available_commands>\n"
        basePrompt += "\nUse run_command to execute any of these when they would be helpful."
    }
    
    return basePrompt
}
```

### Command Discovery at Runtime
```go
func listAvailableCommands() []CommandInfo {
    commandsDir := filepath.Join(os.Getenv("HOME"), ".port42/commands")
    var commands []CommandInfo
    
    // Read all commands
    files, _ := ioutil.ReadDir(commandsDir)
    for _, file := range files {
        if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
            continue
        }
        
        // Read command metadata (from comment header or .meta file)
        metadata := readCommandMetadata(filepath.Join(commandsDir, file.Name()))
        commands = append(commands, metadata)
    }
    
    return commands
}
```

### Dynamic Tool Query
Also provide a tool for runtime command discovery:
```go
func getListCommandsTool() AnthropicTool {
    return AnthropicTool{
        Name:        "list_port42_commands",
        Description: "List all available Port 42 commands",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "category": map[string]interface{}{
                    "type":        "string",
                    "description": "Filter by category (optional)",
                },
            },
        },
    }
}
```

## Security Considerations

### 1. Sandboxing
Commands run with limited permissions:
- No network access by default
- Limited file system access
- Timeout after X seconds
- Resource limits (CPU, memory)

### 2. Command Validation
Before running:
- Verify command exists in ~/.port42/commands/
- Check command metadata/signature
- Validate arguments
- Sanitize inputs

### 3. Output Handling
- Capture stdout/stderr
- Limit output size
- Handle errors gracefully
- Format for AI consumption

## Implementation Plan

### Phase 1: Basic Runner (MVP)
1. Add `run_command` tool to daemon
2. Update system prompt dynamically with available commands
3. Implement basic command execution
4. Capture and return output
5. Handle errors

### Phase 2: Discovery & Persistence
1. Add `list_commands` tool
2. Store command metadata on generation
3. Load all commands on daemon startup
4. Include in system prompt for ALL agents

### Phase 3: Advanced Features
1. Command composition (pipe commands together)
2. Async execution for long-running commands
3. Progress reporting
4. Result caching
5. Command versioning

## Tool Provisioning Update

```go
func getToolsForRequest(request *Request, agent *AgentInfo) []AnthropicTool {
    tools := []AnthropicTool{}
    
    // Always include command runner and discovery tools for ALL agents
    tools = append(tools, getCommandRunnerTool())
    tools = append(tools, getListCommandsTool())
    
    // Add other tools based on context...
    if request.IsInteractive && strings.Contains(request.Message, "/crystallize") {
        tools = append(tools, getCommandGenerationTool())
        tools = append(tools, getArtifactGenerationTool())
    }
    
    return tools
}
```

## Example Implementation

```go
// Handle command execution tool call
func (d *Daemon) handleRunCommand(input json.RawMessage) (string, error) {
    var params struct {
        Command string   `json:"command"`
        Args    []string `json:"args"`
        Stdin   string   `json:"stdin"`
    }
    
    if err := json.Unmarshal(input, &params); err != nil {
        return "", err
    }
    
    // Verify command exists
    cmdPath := filepath.Join(os.Getenv("HOME"), ".port42/commands", params.Command)
    if _, err := os.Stat(cmdPath); err != nil {
        return "", fmt.Errorf("command not found: %s", params.Command)
    }
    
    // Build command
    cmd := exec.Command(cmdPath, params.Args...)
    if params.Stdin != "" {
        cmd.Stdin = strings.NewReader(params.Stdin)
    }
    
    // Set timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
    
    // Capture output
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Sprintf("Error: %v\nOutput: %s", err, output), nil
    }
    
    return string(output), nil
}
```

## Use Cases

### 1. Cross-Agent Workflows
```
# With @ai-muse (creative agent)
User: "Make my git log more fun"
AI: "I'll use git-haiku to convert your recent commits to haikus..."
[Runs: git-haiku --limit 5]

# Later with @ai-engineer (technical agent)
User: "Add a fun changelog to our release"
AI: "I'll use git-haiku that was created earlier to make the changelog engaging..."
[Same command, different agent]
```

### 2. Accumulating Capabilities
```
# Week 1: Create rainbow-art with @ai-muse
# Week 2: Create pr-analyzer with @ai-engineer  
# Week 3: Create tweet-storm with @ai-growth

# Week 4: All agents can use all commands
User to @ai-founder: "Create a Twitter thread about our latest release"
AI: "I'll combine several tools: pr-analyzer for content, rainbow-art for visuals, and tweet-storm for formatting..."
```

### 3. Personal AI OS Evolution
```
# Your command library grows:
~/.port42/commands/
├── rainbow-art          # Week 1
├── git-haiku           # Week 1
├── pr-analyzer         # Week 2
├── tweet-storm         # Week 2
├── content-calendar    # Week 3
├── investor-tracker    # Week 4
└── market-analyzer     # Week 5

# Every conversation can use ALL of these
```

## Benefits

1. **Compounding Value** - Every command created increases AI capabilities
2. **Personalization** - Your AI has YOUR specific tools
3. **Discoverability** - AI can suggest commands you forgot about
4. **Composition** - AI can chain commands creatively
5. **Learning** - AI learns from command usage patterns

## Future Possibilities

1. **Command Marketplace** - Share commands with others
2. **AI Command Improvement** - AI suggests enhancements to existing commands
3. **Workflow Automation** - AI creates multi-command workflows
4. **Context Awareness** - AI knows when to use which command
5. **Meta Commands** - Commands that generate other commands

## Summary

By making generated commands available as AI tools across ALL conversations, we create a revolutionary system:
- Users create commands through conversation with any agent
- Commands become permanently available tools for ALL agents
- AI capabilities compound over time - every command makes every future conversation more powerful
- Your AI literally gets smarter with use

This transforms Port 42 from a command generator into a true "Personal AI Operating System" where:
- Commands are like installed applications
- Every agent can use every command
- Capabilities accumulate permanently
- Your AI evolves uniquely based on your needs

The key insight: **Commands aren't just outputs, they're permanent capability upgrades for your AI.**