# Debug Instructions for Stack Overflow

## Current Status
1. Fixed memory endpoint to return summaries (11KB vs 141KB) ✓
2. Added debug logging to possess handler to track response sizes
3. Stack overflow still happens when continuing sessions

## Test Commands

### Test 1: New Session (should work)
```bash
PORT42_DEBUG=1 ./target/debug/port42 possess @claude -s "test-new" "hello"
```

### Test 2: Continue Recent Session (causes stack overflow)
```bash
PORT42_DEBUG=1 ./target/debug/port42 possess @claude "what did we do last time?"
```

## What to Look For in Daemon Logs
```
🔍 Possess response size: XXX bytes
⚠️  Large possess response detected! Keys: [...]
```

## Hypothesis
The possess handler might be accidentally including extra data when:
1. Session has long history
2. AI response references previous context
3. Some serialization issue with continued sessions

## Next Steps
1. Restart daemon: `sudo -E ./bin/port42d`
2. Run Test 2 and check daemon logs
3. Look for "Possess response size" in daemon output
4. If response > 100KB, the issue is in possess handler
5. If response is normal size, issue might be in CLI parsing