# TUI Redesign - Synchronous Architecture

## Problem Statement

The initial Ratatui TUI implementation suffered from terminal corruption due to:
- Multiple async event loops competing for terminal control
- Unsafe concurrent access to stdout
- No synchronization between render and event threads
- Terminal state not properly restored on panic/exit

## Core Design Principles

### 1. Single-Threaded Event Loop
**No async, no tokio, no spawns.** One thread owns the terminal.

```rust
// The entire TUI runs in a single synchronous loop
fn run_tui() -> Result<()> {
    let mut terminal = SafeTerminal::new()?;
    let mut app = App::new();
    let mut last_tick = Instant::now();
    
    loop {
        // Poll for events with timeout
        if crossterm::event::poll(Duration::from_millis(50))? {
            match crossterm::event::read()? {
                Event::Key(key) => {
                    if app.handle_key(key)? {
                        break; // Quit requested
                    }
                }
                Event::Resize(w, h) => app.resize(w, h),
                _ => {}
            }
        }
        
        // Timer-based refresh (no async!)
        if last_tick.elapsed() >= Duration::from_secs(1) {
            app.refresh_data()?;
            last_tick = Instant::now();
        }
        
        // Single atomic render
        terminal.draw(|f| app.render(f))?;
    }
    
    Ok(())
}
```

### 2. SafeTerminal Wrapper

Guarantees terminal restoration even on panic:

```rust
struct SafeTerminal {
    inner: Terminal<CrosstermBackend<Stdout>>,
    _guard: TerminalGuard,
}

struct TerminalGuard;

impl TerminalGuard {
    fn new() -> Result<Self> {
        enable_raw_mode()?;
        execute!(stdout(), EnterAlternateScreen, EnableMouseCapture)?;
        
        // Install panic hook
        let original = std::panic::take_hook();
        std::panic::set_hook(Box::new(move |info| {
            let _ = Self::restore();
            original(info);
        }));
        
        Ok(Self)
    }
    
    fn restore() {
        let _ = disable_raw_mode();
        let _ = execute!(stdout(), LeaveAlternateScreen, DisableMouseCapture);
    }
}

impl Drop for TerminalGuard {
    fn drop(&mut self) {
        Self::restore();
    }
}
```

### 3. Rate Limiting

Prevent terminal overwhelm:

```rust
struct RateLimiter {
    last_update: Instant,
    min_interval: Duration,
    pending_update: bool,
}

impl RateLimiter {
    fn should_update(&mut self) -> bool {
        if self.last_update.elapsed() >= self.min_interval {
            self.last_update = Instant::now();
            self.pending_update = false;
            true
        } else {
            self.pending_update = true;
            false
        }
    }
}
```

### 4. Simplified Data Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Daemon     │────▶│  App State   │────▶│   Render     │
│              │     │              │     │              │
│ /context API │     │ - Activities │     │ - Header     │
│              │     │ - Filters    │     │ - Table      │
│              │     │ - Selection  │     │ - Footer     │
└──────────────┘     └──────────────┘     └──────────────┘
     ▲                                            │
     │                                            ▼
     └────────────── 1 second poll ──────── Terminal
```

## Implementation Phases

### Phase 1: SafeTerminal Infrastructure
- Create TerminalGuard with guaranteed cleanup
- Implement panic hook for restoration
- Test with forced panics and Ctrl+C

### Phase 2: Synchronous Event Loop
- Single loop with crossterm::event::poll()
- No threads, no async runtime
- Timer-based refresh using Instant::elapsed()

### Phase 3: Rate Limiting
- Max 10 updates per second
- Debounce keyboard input
- Batch activity updates

### Phase 4: Simplified Rendering
- Single draw call per loop iteration
- Immutable render from app state
- No concurrent modifications

## Alternative: Simple Watch Mode

If TUI proves too complex, implement a simpler alternative:

```bash
# Clear screen and redraw every second
watch -n 1 'port42 context --format=watch'
```

Or built-in:

```rust
fn simple_watch(client: &mut DaemonClient, refresh_secs: u64) -> Result<()> {
    loop {
        // Clear screen
        print!("\x1B[2J\x1B[1;1H");
        
        // Get and display context
        let context = client.get_context()?;
        println!("{}", format_context_watch(&context));
        
        // Wait or exit on Ctrl+C
        std::thread::sleep(Duration::from_secs(refresh_secs));
    }
}
```

## Testing Strategy

1. **Stress Test**: Rapid key presses shouldn't corrupt terminal
2. **Panic Test**: Force panic and verify terminal restored
3. **Kill Test**: Kill -9 and verify terminal recoverable
4. **Resize Test**: Terminal resize during operation
5. **Disconnect Test**: Daemon stops while TUI running

## Success Criteria

- No terminal corruption under any exit condition
- Responsive UI with <100ms key press latency
- Stable operation for hours without glitches
- Clean exit with 'q' or Ctrl+C
- Terminal fully restored after exit

## Lessons Learned

1. **Async TUIs are hard** - Terminal is inherently single-threaded resource
2. **Always use Drop guards** - Cleanup must be guaranteed
3. **Polling > Async for TUIs** - Simpler, more predictable
4. **Rate limit everything** - Terminals can't handle unlimited updates
5. **Test the unhappy paths** - Panics, kills, disconnects

## References

- [Ratatui async issues](https://github.com/ratatui-org/ratatui/issues)
- [Crossterm event handling](https://docs.rs/crossterm/latest/crossterm/event/)
- [Terminal restoration patterns](https://github.com/crossterm-rs/crossterm/tree/master/examples)