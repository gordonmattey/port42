# Command Crystallization in Port 42

## How Commands Are Born

In Port 42, commands crystallize from your conversations with AI in two ways:

### 1. Natural Crystallization (Automatic)

When the AI recognizes a clear command specification during conversation:

```
◊ I need a command that shows disk usage as ocean waves
◊◊ @ai-engineer: I understand! Let me create a command that visualizes disk usage...
◊◊ [AI provides implementation details]

✨ REALITY SHIFT DETECTED ✨
A new command has materialized: disk-waves
```

This happens when:
- You clearly describe what you want
- The AI understands and can implement it
- The conversation naturally leads to a concrete specification

### 2. Forced Crystallization (/crystallize)

When you want to explicitly request command generation:

```
◊ We've been talking about git commits and poetry
◊◊ I like the haiku idea best
◊◊◊ /crystallize
◊◊◊ Focusing intention to crystallize a command...
◊◊◊ Tell me what command you wish to manifest:
◊◊◊◊ [AI generates command based on conversation context]
```

Use `/crystallize` when:
- You've discussed ideas and want to manifest one
- The AI hasn't automatically generated a command yet
- You want to explicitly trigger the creation process

### 3. Command Generation Protocol

Behind the scenes, the AI generates a JSON specification:
```json
{
  "command": "git-haiku",
  "description": "Transform git commits into haikus",
  "language": "bash",
  "implementation": "#!/bin/bash\n# Implementation here..."
}
```

The daemon's forge then creates the actual executable in `~/.port42/commands/`

### Tips for Better Crystallization

1. **Be Specific**: "I need a command that..." works better than vague requests
2. **Provide Context**: Explain what problem you're solving
3. **Iterate**: Refine the idea through conversation before crystallizing
4. **Use Examples**: Show the AI what output you expect

### Special Commands During Possession

- `/crystallize` - Explicitly request command generation
- `/deeper` - Dive deeper into the conversation
- `/memory` - Check session statistics
- `/reality` - See commands already created
- `/surface` - End the session

The magic happens when your intention becomes clear enough for the AI to manifest it as code!