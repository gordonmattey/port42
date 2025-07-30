# /crystallize Extension Implementation Plan

## Summary

Extend the `/crystallize` command in Port 42 to support three distinct output types:
1. **Commands** - Executable tools (current behavior)
2. **Documents** - Markdown knowledge artifacts
3. **Data** - CRUD-based data management commands

## Implementation Steps (2-Day Sprint)

### Day 1: Core Changes (8 hours)

**Morning (4 hours)**
1. Update CLI (`interactive.rs`) - 1 hour:
   ```rust
   enum CrystallizeType { Auto, Command, Document, Data }
   ```
   - Parse `/crystallize [type]` syntax
   - Pass type in request payload

2. Update Daemon types - 1 hour:
   - Add `artifact_type` field to request
   - Add `DocumentSpec` and `DataCommandSpec` types
   - Extend session to track any artifact type

3. Update AI prompts - 2 hours:
   - Create type-specific prompts
   - Add JSON examples for each type
   - Test prompt effectiveness

**Afternoon (4 hours)**
4. Document generation - 2 hours:
   - File generation logic
   - Markdown formatting with metadata
   - Directory structure creation

5. Data command generation - 2 hours:
   - Simple bash template implementation
   - JSON data file initialization
   - Basic CRUD operations

### Day 2: Integration & Polish (8 hours)

**Morning (4 hours)**
1. Wire everything together - 2 hours:
   - Connect CLI to daemon changes
   - Test all three crystallization types
   - Fix integration issues

2. AI response handling - 2 hours:
   - Parse different artifact types
   - Route to correct generator
   - Handle errors gracefully

**Afternoon (4 hours)**
3. Testing & refinement - 2 hours:
   - Test each workflow type
   - Fix edge cases
   - Ensure backward compatibility

4. Documentation & ship - 2 hours:
   - Update user docs
   - Add examples
   - Create announcement

## File Changes Required

### CLI Side
- `cli/src/interactive.rs` - Add crystallize type handling
- `cli/src/types.rs` - Add new request types

### Daemon Side
- `daemon/possession.go` - Handle different artifact types
- `daemon/document_store.go` - New file for document management
- `daemon/data_templates.go` - Templates for data commands
- `daemon/server.go` - Route generation based on type

### Documentation
- Update user documentation
- Add examples for each crystallize type
- Create migration guide

## Benefits

1. **Flexibility** - Users can explicitly choose output format
2. **Organization** - Different artifact types stored separately
3. **Extensibility** - Easy to add new crystallization types
4. **Intelligence** - AI can suggest appropriate type

## Next Steps

1. Start with updating `interactive.rs` to parse options
2. Extend daemon to handle multiple artifact types
3. Implement document generation first (simpler)
4. Add data command generation
5. Update documentation and examples

## Timeline Estimate

- Day 1: Core implementation (8 hours)
  - Morning: CLI updates, daemon types, AI prompts
  - Afternoon: Document & data generation logic
  
- Day 2: Integration & ship (8 hours)
  - Morning: Wire together, handle responses
  - Afternoon: Test, document, deploy

Total: 2 days (16 hours focused work)