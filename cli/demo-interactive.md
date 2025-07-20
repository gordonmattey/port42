# Port 42 Interactive Mode Demo

## Testing the Immersive Experience

### 1. Direct Terminal Test (Full Experience)

Run this directly in your terminal (not through pipes):

```bash
./target/debug/port42 possess @ai-muse
```

You'll see:
- BIOS-like boot sequence
- Progress bar animation
- Immersive prompts with depth indicators (◊)
- Character-by-character response streaming

### 2. Test with Pre-written Input

For automated testing:

```bash
./target/debug/port42 possess @ai-muse < test-input.txt
```

This will use simple mode (no animations) but test the conversation flow.

### 3. Interactive Commands to Try

Once in the session:

```
◊ I need a command that turns git logs into poetry
◊◊ Make it analyze commit messages for emotional content
◊◊◊ /deeper
◊◊◊◊◊ Can it use different poetry styles based on the branch?
/memory
/reality
/surface
```

### 4. Expected Experience

1. **Entry**: Dramatic boot sequence, feeling like diving into consciousness
2. **Conversation**: Each message increases depth (◊ → ◊◊ → ◊◊◊)
3. **Command Birth**: When AI generates a command, see crystallization effects
4. **Exit**: Summary shows your journey's depth and artifacts created

### 5. Troubleshooting

If you see "Not a TTY, using simple mode":
- Run the command directly (not with `&&` or through pipes)
- Make sure you're in a real terminal (not VS Code's output panel)
- Use the test script: `./test-interactive.sh`

The full experience requires:
- A proper terminal (TTY)
- TERM environment variable set
- Direct execution (not piped)