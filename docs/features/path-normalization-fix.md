# P42 Path Normalization Fix

## Problem Statement

Port42's VFS path resolution fails when paths include trailing slashes, causing inconsistent behavior across the system. This affects all P42 path operations, not just references.

### Examples of Broken Behavior
- `--ref p42:/commands/test-vfs-fix/` fails to resolve (returns "No context resolved")
- `--ref p42:/commands/test-vfs-fix` works correctly
- This likely affects: `port42 ls /commands/`, `port42 cat /memory/session/`, and all VFS operations

## Root Cause Analysis

The issue is in the **daemon's VFS path handling** where paths aren't normalized before resolution. The system treats `/commands/test-vfs-fix` and `/commands/test-vfs-fix/` as different paths, but they should resolve to the same resource.

## Scope of Impact

This trailing slash issue affects **all P42 path operations**:
- Reference resolution (`--ref p42:/path/`)
- VFS list operations (`port42 ls /commands/`)
- VFS file operations (`port42 cat /memory/session/`)
- Search operations with path filters
- Memory and storage path resolution
- All daemon VFS handlers

## Implementation Plan

### Phase 1: Locate Path Processing Points

#### CLI Side (`cli/src/`)
- Reference parsing in `cli/src/protocol/relations.rs`
- Command argument parsing and validation
- Path validation logic before sending to daemon

#### Daemon Side (`daemon/`)
- VFS path handlers in `daemon/server.go`
- P42 resolver in `daemon/resolution/resolvers.go`
- File operations, list operations, memory operations
- Any path-based routing or resolution

### Phase 2: Create Central Path Normalization

#### Add Core Normalization Function
```go
// daemon/path_utils.go
func NormalizeP42Path(path string) string {
    // Remove trailing slashes except for root "/"
    // Handle double slashes (//) -> single slash (/)
    // Handle relative path components (./,  ../)
    // Trim whitespace
    // Ensure consistent path format
    return cleanPath
}
```

#### Path Normalization Rules
1. **Trailing Slash Removal**: `/commands/tool/` → `/commands/tool`
2. **Root Path Exception**: `/` remains `/` (don't remove root trailing slash)
3. **Double Slash Cleanup**: `//commands` → `/commands`
4. **Whitespace Trimming**: ` /commands ` → `/commands`
5. **Empty Path Handling**: `""` → `/` or appropriate default

#### Apply Normalization at Entry Points
- When requests are received by daemon (before routing)
- Before any VFS operations
- In P42 resolver before resolution attempts
- During reference parsing on CLI side

### Phase 3: Update All Path Handlers

#### VFS Operations (`daemon/server.go`)
- List operations (`list_path` handler)
- File read operations (`cat` handler) 
- Info operations (`info` handler)
- Any other VFS endpoint handlers

#### Reference Resolution (`daemon/resolution/resolvers.go`)
- P42 resolver path handling
- Reference target normalization
- VFS access path preparation

#### Search Operations
- Path-based search functionality
- Filter path normalization
- Search result path consistency

#### Memory Operations
- Session path resolution
- Storage path handling
- Memory VFS operations

### Phase 4: Add Path Validation

#### Validation Rules
1. **Well-formed paths**: Must start with `/` for absolute paths
2. **Character restrictions**: Invalid characters for filesystem paths
3. **Length limits**: Reasonable path length restrictions
4. **Security checks**: Prevent path traversal attempts

#### Error Handling
- Consistent error messages for invalid paths
- Helpful suggestions for malformed paths
- Clear distinction between "path invalid" vs "path not found"

## Testing Strategy

### Unit Tests
```go
func TestNormalizeP42Path(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"/commands/tool/", "/commands/tool"},
        {"/commands//tool", "/commands/tool"},
        {"//commands", "/commands"},
        {" /commands ", "/commands"},
        {"/", "/"},
        {"", "/"},
    }
    // ... test implementation
}
```

### Integration Tests
1. **VFS Operations**: All operations with trailing slashes
2. **Reference Resolution**: All reference types with various path formats
3. **CLI Commands**: `ls`, `cat`, `info` with normalized paths
4. **Cross-platform**: Ensure consistent behavior across OS

### Regression Tests
- Ensure existing functionality without trailing slashes continues working
- Verify no performance impact from normalization
- Test edge cases and error conditions

## Files to Examine/Modify

### Core Path Handling
- `daemon/resolution/resolvers.go` - P42 resolver path handling
- `daemon/server.go` - VFS operation handlers  
- `cli/src/protocol/relations.rs` - Reference parsing and validation

### New Files to Add
- `daemon/path_utils.go` - Path normalization utilities
- `daemon/path_utils_test.go` - Comprehensive path normalization tests

### Configuration Files
- Update any path validation configuration
- Document new path normalization behavior

## Backwards Compatibility

### Compatibility Strategy
- **Existing paths without trailing slashes**: Continue working unchanged
- **Gradual migration**: No breaking changes to existing functionality
- **Deprecation warnings**: Optional warnings for malformed paths (if needed)

### Migration Path
1. **Phase 1**: Add normalization without breaking existing behavior
2. **Phase 2**: Add optional warnings for non-normalized paths
3. **Phase 3**: Full enforcement of normalized paths (if needed)

## Security Considerations

### Path Traversal Prevention
- Ensure normalization doesn't introduce security vulnerabilities
- Validate normalized paths don't escape intended boundaries
- Block attempts to access system paths outside P42 VFS

### Input Sanitization
- Proper handling of malicious path inputs
- Rate limiting for invalid path requests
- Logging of suspicious path access attempts

## Performance Impact

### Optimization Considerations
- Path normalization should be fast (O(n) where n = path length)
- Cache normalized paths if beneficial
- Avoid repeated normalization of same paths

### Benchmarking
- Measure performance impact of normalization
- Compare before/after performance metrics
- Ensure no significant latency increase

## Success Criteria

### Functional Requirements
1. ✅ `--ref p42:/commands/tool/` works identically to `--ref p42:/commands/tool`
2. ✅ All VFS operations handle trailing slashes consistently
3. ✅ No regression in existing functionality
4. ✅ Comprehensive test coverage for path normalization

### Non-Functional Requirements
1. ✅ Less than 1ms additional latency for path normalization
2. ✅ No breaking changes to existing API
3. ✅ Clear error messages for invalid paths
4. ✅ Consistent behavior across all platforms

## Implementation Priority

### High Priority (Fix Core Issue)
1. Add `NormalizeP42Path()` function
2. Apply to P42 resolver in reference resolution
3. Apply to main VFS handlers (`ls`, `cat`, `info`)

### Medium Priority (System-wide Consistency)
1. Apply to all remaining VFS operations
2. Add comprehensive test suite
3. Update CLI-side path validation

### Low Priority (Polish & Enhancement)
1. Add optional path validation warnings
2. Performance optimization if needed
3. Documentation updates

---

**Status**: Planning Complete
**Next Steps**: Begin implementation with P42 resolver path normalization
**Est. Timeline**: 1-2 days for core fix, 3-5 days for comprehensive solution