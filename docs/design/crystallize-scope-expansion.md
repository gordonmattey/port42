# Extended /crystallize Command Design

**Purpose**: Feature specification for extending /crystallize to support multiple output types.
**Scope**: Command syntax, API design, tool provisioning logic. Links to other docs for details.

## Overview

Extend the `/crystallize` command to support three different output types:
1. Commands (executable CLI tools)
2. Artifacts (any file: docs, code, designs, media)
3. Data (CRUD management tools)

**Key Change**: Tools are no longer provided by default to prevent aggressive auto-generation.

## Command Syntax

### Interactive Mode (Conversational)
```bash
possess @ai-muse
> Let's explore ideas...        # No tools available
> /crystallize                  # Now tools are provided, AI chooses type
> /crystallize command          # Force command creation
> /crystallize artifact         # Force artifact creation
> /crystallize data             # Force data tool creation
```

### Non-Interactive Mode (One-Shot)
```bash
# Auto-detects intent and provides appropriate tool
possess @ai-engineer "create a command that shows disk usage"
possess @ai-muse "create a design for our architecture"
possess @ai-growth "build a data tracker for metrics"

# Just conversation, no tools
possess @ai-founder "what's our pricing strategy?"
```

## Implementation Plan

### 1. Update Interactive Session (cli/src/interactive.rs)

```rust
fn handle_special_command(&mut self, input: &str) -> Result<bool> {
    match input {
        "/crystallize" => {
            self.request_crystallization(CrystallizeType::Auto)?;
            Ok(true)
        }
        "/crystallize command" => {
            self.request_crystallization(CrystallizeType::Command)?;
            Ok(true)
        }
        "/crystallize artifact" => {
            self.request_crystallization(CrystallizeType::Artifact)?;
            Ok(true)
        }
        "/crystallize data" => {
            self.request_crystallization(CrystallizeType::Data)?;
            Ok(true)
        }
        // ... other commands
    }
}

enum CrystallizeType {
    Auto,
    Command,
    Artifact,
    Data,
}
```

### 2. Context-Aware Tool Provisioning (daemon/possession.go)

```go
func getToolsForRequest(request *Request, agent *AgentInfo) []AnthropicTool {
    // Interactive mode - no tools by default
    if request.IsInteractive {
        // Only provide tools when explicitly requested
        if strings.Contains(request.Message, "/crystallize") {
            return getRequestedTools(request.Message)
        }
        return []AnthropicTool{} // No tools during conversation
    }
    
    // Non-interactive mode - detect intent
    if looksLikeGenerationRequest(request.Message) {
        return []AnthropicTool{detectAppropriateTools(request.Message)}
    }
    
    // Default: no tools
    return []AnthropicTool{}
}

func looksLikeGenerationRequest(message string) bool {
    patterns := []string{
        "create a command", "build a tool", "make a command",
        "write a document", "create a dashboard", "build an app",
        "design a", "create a diagram", "generate a mockup",
    }
    // Check if message indicates generation intent
}
```

### 3. Update Daemon Types (daemon/possession.go)

```go
// ArtifactSpec for artifacts (docs, code, designs, etc)
type ArtifactSpec struct {
    Type        string `json:"type"`        // "document", "code", "design", "media"
    Name        string `json:"name"`
    Description string `json:"description"`
    Content     string `json:"content"`     // File content or description
    FileType    string `json:"file_type"`   // Extension: .md, .js, .svg, etc
    Files       []FileSpec `json:"files"`   // For multi-file artifacts
}

// Update session to support both
type Session struct {
    // ... existing fields
    CommandGenerated  *CommandSpec  `json:"command_generated,omitempty"`
    ArtifactGenerated *ArtifactSpec `json:"artifact_generated,omitempty"`
}
```

### 4. Different Prompts for Different Types

```go
func getCrystallizationPrompt(crystallizeType string) string {
    baseGuidance := "You now have access to generation tools. Only use them if there's a clear artifact to create based on our conversation. "
    
    switch crystallizeType {
    case "command":
        return baseGuidance + "If appropriate, CREATE A COMMAND based on our conversation..."
    case "artifact":
        return baseGuidance + "If appropriate, CREATE AN ARTIFACT based on our conversation. This could be a document, code, design, diagram, or any other file that captures our discussion."
    case "data":
        return baseGuidance + "If appropriate, CREATE A DATA MANAGEMENT COMMAND based on our conversation. This should be a command that manages structured data with CRUD operations."
    default:
        return baseGuidance + "Based on our conversation, create the most appropriate output if one is needed: either a command (executable tool), an artifact (any file type), or a data management tool (CRUD operations)."
    }
}
```

### 5. Artifact Storage Structure

Artifacts are saved to `~/.port42/artifacts/` with organization:

```
~/.port42/artifacts/
├── decisions/
│   └── 2024-01-15-api-architecture.md
├── strategies/
│   └── 2024-01-15-growth-plan.md
├── knowledge/
│   └── 2024-01-15-competitor-analysis.md
├── code/
│   └── 2024-01-15-dashboard-app/
└── index.json  # Rich metadata index (see artifact-metadata-system.md)
```

Each artifact is indexed with:
- Tags, categories, and keywords for search
- Lifecycle state (draft → active → archived)
- Relationships to other artifacts
- Usage statistics and importance scoring
- Embeddings for semantic search


## Benefits

1. **User Control**: Explicitly choose output type when needed
2. **AI Intelligence**: Let AI decide when type not specified  
3. **Clear Organization**: Different storage for different artifact types
4. **Extensibility**: Easy to add new crystallization types
5. **Scalability**: Metadata system prevents artifact sprawl
6. **Intentional Generation**: No more accidental artifacts - tools only when requested
7. **Context Awareness**: One-shot CLI still works intuitively

## Migration Path

1. Keep existing `/crystallize` behavior as default
2. Add new options incrementally
3. Update documentation and examples
4. Gather user feedback and iterate