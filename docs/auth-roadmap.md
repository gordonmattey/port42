# Port 42 Authentication & Multi-Model Roadmap

## Current State (MVP)
- Single API key from environment: `ANTHROPIC_API_KEY`
- Single model hardcoded: `claude-3-5-sonnet-20241022`
- No user accounts or authentication
- No credential storage (just env vars)
- No model switching capability

## Phase 2 Vision (4 weeks)
A proper authentication and multi-model system that's secure, flexible, and user-friendly.

### Quick Comparison

| Feature | MVP (Now) | Phase 2 |
|---------|-----------|----------|
| **Credential Storage** | Environment variable | OS Keychain |
| **Supported Providers** | Anthropic only | Claude, GPT, Ollama, Groq |
| **Model Selection** | Hardcoded | Dynamic per-request |
| **Account Management** | None | Multiple accounts |
| **Usage Tracking** | None | Per-account limits |
| **Configuration** | Environment only | TOML + Keychain |

### User Experience Evolution

**MVP (Current)**:
```bash
export ANTHROPIC_API_KEY=sk-ant-...
port42d  # Uses the one key for everything
```

**Phase 2**:
```bash
# First time setup
port42 account add anthropic --name personal
Enter API key: [secure input]
✓ Key stored in macOS Keychain

# Add work account
port42 account add openai --name work
Enter API key: [secure input]

# Switch between accounts
port42 account use work
port42 possess @ai-engineer  # Uses work account's GPT-4

# Use specific model
port42 possess @ai-muse --model claude-3-opus

# Check usage
port42 account usage
Personal (Anthropic): $12.43 this month
Work (OpenAI): $8.91 this month
```

### Technical Architecture

```
┌─────────────────────────────────────────┐
│            Port 42 CLI                  │
├─────────────────────────────────────────┤
│  Commands:                              │
│  - account add/list/use/remove          │
│  - model list/set/info                  │
│  - possess --model <model>              │
└──────────────┬──────────────────────────┘
               │ TCP
┌──────────────▼──────────────────────────┐
│         Port 42 Daemon                  │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐   │
│  │     Account Manager              │   │
│  │  - Multiple accounts             │   │
│  │  - Usage tracking                │   │
│  │  - Rate limiting                 │   │
│  └──────────┬──────────────────────┘   │
│             │                           │
│  ┌──────────▼──────────────────────┐   │
│  │     Model Registry               │   │
│  │  - Provider plugins              │   │
│  │  - Model capabilities            │   │
│  │  - Dynamic selection             │   │
│  └──────────┬──────────────────────┘   │
│             │                           │
│  ┌──────────▼──────────────────────┐   │
│  │     Secure KeyStore              │   │
│  │  - OS Keychain integration       │   │
│  │  - Encrypted fallback            │   │
│  │  - Never in memory/logs          │   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Security Improvements

**MVP Risks**:
- API key in environment (visible in `ps`, logs)
- No rotation capability
- Single point of failure
- No usage limits

**Phase 2 Security**:
- Keys in OS keychain (encrypted at rest)
- Per-account isolation
- Usage limits and alerts
- Audit trail
- Key rotation support

### Implementation Priority

1. **Week 1**: Keychain integration (security first)
2. **Week 2**: Multi-model support (flexibility)
3. **Week 3**: Account management (usability)
4. **Week 4**: Configuration system (polish)

### Why This Architecture?

1. **Security**: OS keychains are battle-tested
2. **Flexibility**: Easy to add new AI providers
3. **User Control**: Switch models based on task
4. **Cost Management**: Track usage per account
5. **Future Proof**: Ready for team features

### Next Immediate Steps

1. Fix current timeout issues (DONE)
2. Document Phase 2 plan (DONE)
3. Get user feedback on priorities
4. Start keychain research/POC
5. Design provider plugin interface