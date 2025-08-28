# Phase 2 Architecture Fix

## Root Cause: Poor Package Structure

### Issues Identified:
1. **Circular Dependencies**: resolvers ↔ context ↔ types
2. **Duplicate Types**: Reference defined in multiple places
3. **Mixed Concerns**: Protocol types mixed with resolution logic
4. **Unclear Boundaries**: No interface separation

## Clean Architecture Solution:

```
daemon/
├── protocol.go          # Core protocol types (Reference, Request, Response)
├── server.go            # Main daemon logic
├── resolution/          # Resolution subsystem (self-contained)
│   ├── interface.go     # Public interface for daemon
│   ├── engine.go        # Resolution orchestration
│   ├── resolvers/       # Individual resolver implementations
│   │   ├── search.go
│   │   ├── tool.go
│   │   ├── memory.go
│   │   ├── file.go
│   │   └── url.go
│   └── context/         # Context synthesis (internal)
│       ├── synthesizer.go
│       └── limiter.go
└── types/               # Shared types only (no logic)
    └── resolution.go
```

### Key Principles:
1. **Single Source of Truth**: Reference type only in protocol.go
2. **Clear Interface**: resolution/ package exposes simple interface to daemon
3. **No Circular Dependencies**: resolution/ is self-contained
4. **Separation of Concerns**: Protocol ≠ Resolution ≠ Context