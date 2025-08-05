# /crystallize Extension Implementation Plan

**Purpose**: Technical implementation roadmap with time estimates and file changes.
**Scope**: Day-by-day tasks, specific files to modify, testing approach.

## Summary

Extend the `/crystallize` command in Port 42 to support three distinct output types:
1. **Commands** - Executable CLI tools (current behavior)
2. **Artifacts** - Any file type (docs, code, designs, media)
3. **Data** - CRUD-based data management commands

## Implementation Steps (2-Day Sprint)

### Day 1: Core Changes (8 hours)

**Morning (4 hours)**
1. Update CLI (`interactive.rs`) - 1 hour:
   ```rust
   enum CrystallizeType { Auto, Command, Artifact, Data }
   ```
   - Parse `/crystallize [type]` syntax
   - Pass type in request payload
   - Add `is_interactive` flag to requests

2. Context-Aware Tool Provisioning - 1.5 hours:
   - Implement tool gating logic
   - No tools by default in interactive mode
   - Auto-detect generation intent for one-shot CLI
   - Only provide tools when explicitly requested

3. Update Daemon types - 0.5 hours:
   - Add `artifact_type` field to request
   - Add `ArtifactSpec` and `DataCommandSpec` types
   - Extend session to track any output type

4. Update AI prompts - 1 hour:
   - Create type-specific prompts
   - Add explicit tool usage guidelines
   - Emphasize tools should only be used when appropriate

**Afternoon (4 hours)**
4. Artifact generation - 2 hours:
   - File generation logic
   - Support multiple file types
   - Directory structure creation
   - Update artifact index

5. Data command generation - 2 hours:
   - Simple bash template implementation
   - JSON data file initialization
   - Basic CRUD operations
   - Add metadata to index

### Day 2: Integration & Polish (8 hours)

**Morning (4 hours)**
1. Wire everything together - 1.5 hours:
   - Connect CLI to daemon changes
   - Test all crystallization types
   - Fix integration issues

2. Metadata system - 1.5 hours:
   - Create artifact index structure
   - Auto-extract tags and keywords
   - Link artifacts to sessions

3. AI response handling - 1 hour:
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
- `daemon/artifact_store.go` - New file for artifact management & indexing
- `daemon/artifact_generator.go` - Artifact generation logic
- `daemon/data_templates.go` - Templates for data commands
- `daemon/metadata_extractor.go` - Extract tags, keywords, embeddings
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