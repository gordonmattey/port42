# **Hierarchical Plan: Enhanced Possess Mode + Flexible Declare**

## **🎯 Implementation Progress Tracker**

| Step | Component | Status | Notes |
|------|-----------|---------|-------|
| **1** | Design Fresh Agent Guidance System | ⏳ **Not Started** | XML-structured prompts, AI decision framework, expanded artifacts |
| **2** | Enhance Tool Materializer | ⏳ **Not Started** | Multi-language support, intelligent language selection |
| **3** | Create Comprehensive Examples | ⏳ **Not Started** | Simple/complex tools, multi-language showcase, artifact templates |
| **4** | Fresh Architecture Testing & Validation | ⏳ **Not Started** | Clean slate validation, performance benchmarking |

**Legend**: ⏳ Not Started • 🚧 In Progress • ✅ Complete • ❌ Blocked

---

## **Level 1: Main Components**

### **A. Agent Instruction Architecture** 
Structured guidance system for Claude in possess mode
*Clean slate design - no backward compatibility with legacy prompt formats*

### **B. Tool Materialization Flexibility**
Multi-language support with intelligent language selection
*Clean slate design - no backward compatibility with CommandSpec data structures*

### **C. Integration & Testing**
Validation and rollout strategy

---

## **Level 2: Component Breakdown**

### **A. Agent Instruction Architecture**

#### **A1. Tool Usage Classification**
Clear separation of when/how to use different approaches

#### **A2. Port42 Command Guidelines** 
Best practices for using existing port42 ecosystem

#### **A3. Tool Creation Framework**
Comprehensive declare command construction

#### **A4. Context Integration**
Reference system usage and prompt crafting

### **B. Tool Materialization Flexibility**

#### **B1. Language Selection Logic**
AI-driven language choice based on transforms/requirements  

#### **B2. Template System**
Multi-language code generation templates

#### **B3. Dependency Management**
External tool and library handling

### **C. Integration & Testing**

#### **C1. Clean Slate Validation**
Test new guidance system from scratch (no backward compatibility needed)

#### **C2. Fresh Implementation Testing**
Direct validation of new architecture without legacy constraints

#### **C3. Comprehensive Validation Suite**
Test all new capabilities: multi-language tools, artifact generation, XML-guided workflows

---

## **Level 3: Detailed Design**

### **A1. Tool Usage Classification (AI-Driven)**

#### **A1.1 Decision Framework for AI**
*Guidance provided to Claude for intelligent decision-making*

```
When a user makes a request, interpret their intent and choose the appropriate approach:

USING EXISTING TOOLS:
- Intent indicators: "use", "run", "execute", "what tools do I have", "show me"
- User signals: Wants immediate action, asks about capabilities
- Approach: Search first, then execute
- Commands: search, ls /similar/, run_command

CREATING NEW TOOLS:  
- Intent indicators: "create", "build", "make a tool", "I need a command that"
- User signals: Describes missing functionality, specific requirements
- Approach: Check existing first, then declare if needed
- Commands: port42 declare tool with appropriate transforms

PORT42 OPERATIONS:
- Intent indicators: "show me", "list", "what's in", "explore", "how do I"
- User signals: Wants information, navigation, system understanding
- Approach: Direct VFS or system operations  
- Commands: port42 ls, cat, info, status
```

#### **A1.2 AI Decision Process (XML-Structured Prompts)** 
*Natural language understanding with clear XML organization - clean slate design, no legacy prompt compatibility*

```xml
<decision_workflow>
<understand>
What is the user actually asking for?
- Immediate action with existing tools?
- New capability that doesn't exist yet?
- Information, exploration, or learning?
</understand>

<discover_first>
Before creating anything new, always check what exists:
- run_command('port42', ['search', 'relevant keywords'])
- run_command('port42', ['ls', '/similar/related-tool/'])
- run_command('port42', ['ls', '/tools/by-transform/category/'])
- Ask user: "Found these existing tools [list]. Try these first or create enhanced version?"
</discover_first>

<choose_approach>
Based on discovery results:
- If adequate tools exist: Use them or orchestrate combination
- If no tools exist: Create with port42 declare
- If user wants to explore: Use VFS navigation commands  
- If unclear: Ask clarifying questions
</choose_approach>

<execute_with_context>
- Use existing: Explain what the tools do, show results
- Create new: Use comprehensive declare patterns with references
- Navigate: Guide user through discovery process
- Always explain your reasoning and next steps
</execute_with_context>
</decision_workflow>
```

### **A2. Port42 Command Guidelines**

#### **A2.1 Discovery Commands**
```yaml
purpose: "Find existing tools before creating new ones"
commands:
  search: "port42 search 'keyword' - semantic search across everything"
  similar: "port42 ls /similar/tool-name/ - find capability matches"  
  tools: "port42 ls /tools/by-transform/category/ - browse by capability"
  info: "port42 info /tools/tool-name - get detailed metadata"
```

#### **A2.2 Execution Commands**  
```yaml
purpose: "Run existing tools and port42 operations"
commands:
  run_tool: "Use existing commands directly by name"
  run_port42: "run_command('port42', ['subcommand', ...]) for port42 ops"
  vfs: "port42 cat, ls, info for filesystem operations"
```

### **A3. Tool Creation Framework**

#### **A3.1 Basic Declare Structure**
```bash
port42 declare tool TOOLNAME --transforms keyword1,keyword2,keyword3
```

#### **A3.2 Advanced Declare Patterns** 
*From documentation analysis*

**With Custom Prompt:**
```bash
port42 declare tool TOOLNAME --transforms keywords \
  --prompt "Specific AI instructions for implementation"
```

**With References:**
```bash
port42 declare tool TOOLNAME --transforms keywords \
  --ref file:./config.json \
  --ref p42:/commands/base-tool \
  --ref url:https://api.example.com/docs \
  --prompt "Build on referenced context"
```

#### **A3.3 Transform Selection Guidelines**
*Based on documentation examples*

**Core Categories:**
- **Data Flow**: `stdin, file, stream, batch, pipeline`
- **File Formats**: `json, csv, xml, yaml, text, binary`  
- **Operations**: `parse, filter, convert, transform, merge, split`
- **Analysis**: `analyze, stats, pattern, search, extract`
- **Output**: `format, export, display, report, save`
- **Features**: `error, logging, progress, config, help`

**Language Hints**: `bash, python, node` (for multi-language support)

#### **A3.4 Prompt Crafting Patterns**

**Quality Requirements:**
- Error handling specifications
- Input/output format requirements  
- Performance considerations
- Integration requirements

**Context Integration:**
- Reference existing patterns: `--ref p42:/commands/similar-tool`
- Use project context: `--ref file:./project-config`
- Leverage web specs: `--ref url:https://standard.org/spec`

### **A4. Context Integration**

#### **A4.1 Reference System Usage**
*How Claude should use different reference types effectively*

```xml
<reference_types>
<file_references>
- Use: --ref file:./config.json, --ref file:./README.md
- Purpose: Project-specific context, existing patterns, configurations
- AI Guidance: "Analyze the referenced files and adapt the tool to work with the existing project structure"
</file_references>

<port42_references>
- Use: --ref p42:/commands/existing-tool, --ref p42:/memory/session-id
- Purpose: Build on existing Port42 knowledge and tools
- AI Guidance: "Reference existing Port42 capabilities and extend/enhance rather than duplicate"
</port42_references>

<web_references>
- Use: --ref url:https://api.example.com/docs
- Purpose: External APIs, standards, specifications
- AI Guidance: "Incorporate external standards and API patterns into the implementation"
</web_references>

<search_references>
- Use: --ref search:"error handling patterns"
- Purpose: Crystallized knowledge from previous sessions
- AI Guidance: "Apply accumulated best practices and patterns from Port42's knowledge base"
</search_references>
</reference_types>
```

#### **A4.2 Prompt Crafting Patterns**
*How to construct effective prompts for complex tools*

```xml
<prompt_patterns>
<quality_specifications>
Examples:
- "Include comprehensive error handling with detailed user-friendly messages"
- "Add progress indicators for long-running operations"
- "Implement graceful degradation when dependencies are missing"
- "Follow security best practices for handling sensitive data"
</quality_specifications>

<integration_requirements>
Examples:
- "Integrate with existing project structure in ./src/"
- "Use the same logging format as other project tools"  
- "Follow the API patterns established in the referenced documentation"
- "Maintain compatibility with the existing configuration system"
</integration_requirements>

<context_synthesis>
Pattern: "Build a [TOOL_TYPE] that [MAIN_FUNCTION], incorporating patterns from [REFERENCES], with [QUALITY_REQUIREMENTS], following [STANDARDS/SPECS]"

Example: "Build a log analyzer that processes nginx access logs, incorporating patterns from the referenced log-parser tool, with comprehensive error handling and progress indicators, following the Common Log Format specification"
</context_synthesis>
</prompt_patterns>
```

#### **A4.3 Reference Resolution Integration**
*How references get resolved and integrated into AI prompts*

```
RESOLUTION FLOW:
1. User provides: --ref file:./config.json --ref p42:/commands/base-tool
2. System resolves: file content + existing tool definition  
3. AI prompt enhanced: Original prompt + "Additional Context from References: [resolved content]"
4. Tool generation: AI uses both original requirements AND resolved context
5. Result: Context-aware tool that integrates with existing ecosystem
```

### **B1. Language Selection Logic**

#### **B1.1 Transform → Language Mapping**
*Clean slate design - new mapping system without CommandSpec constraints*
```yaml
bash_preferred:
  - git, commit, branch, status
  - file, directory, filesystem  
  - pipe, stream, filter
  - system, process, service

python_preferred:
  - json, xml, yaml, data
  - analyze, stats, calculate
  - http, api, client
  - parse, validate, transform

node_preferred:
  - web, server, api
  - json, rest, graphql
  - frontend, ui, interactive
```

#### **B1.2 AI Language Selection Prompt**
```
Based on transforms [X,Y,Z], choose the most appropriate language:
- bash: System operations, git, pipes, file processing
- python: Data analysis, APIs, complex parsing, scientific computing  
- node: Web services, JSON APIs, interactive tools, modern JS features

Consider: Performance, ecosystem, complexity, maintenance
Default: bash for simple operations, python for data/analysis, node for web/API
```

### **B2. Template System**
*Clean slate design - new template formats without preserving CommandSpec JSON structure*

#### **B2.1 Bash Template**
```bash
#!/bin/bash
# Tool: {name}  
# Description: {description}
# Transforms: {transforms}

set -euo pipefail  # Error handling
{dependency_checks}
{argument_parsing}
{main_implementation}
{error_handling}
```

#### **B2.2 Python Template**  
```python
#!/usr/bin/env python3
"""
{name}: {description}
Transforms: {transforms}
"""

import argparse
import sys
{additional_imports}

{dependency_checks}
{main_implementation}
{error_handling}

if __name__ == "__main__":
    main()
```

#### **B2.3 Node Template**
```javascript
#!/usr/bin/env node
// {name}: {description}  
// Transforms: {transforms}

{dependency_checks}
{argument_parsing}
{main_implementation}
{error_handling}
```

### **B3. Dependency Management**

#### **B3.1 External Tool Detection**
```yaml
dependency_categories:
  git_tools: "git, gh, hub"
  text_processing: "jq, yq, ripgrep, sed, awk"  
  visualization: "lolcat, figlet, boxes, cowsay"
  system: "curl, wget, nc, ps, lsof"
  development: "npm, pip, cargo, go"
```

#### **B3.2 Dependency Validation Templates**
**Bash Dependency Check:**
```bash
check_dependencies() {
    local deps=("$@")
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            missing+=("$dep")
        fi
    done
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "Missing dependencies: ${missing[*]}" >&2
        echo "Install with: brew install ${missing[*]}" >&2  # or apt-get, etc
        exit 1
    fi
}
```

**Python Dependency Check:**
```python
def check_dependencies(deps):
    import shutil
    missing = [dep for dep in deps if not shutil.which(dep)]
    if missing:
        print(f"Missing dependencies: {', '.join(missing)}", file=sys.stderr)
        print(f"Install with: pip install {' '.join(missing)}", file=sys.stderr)
        sys.exit(1)
```

#### **B3.3 Smart Dependency Inference**
Based on transforms, automatically suggest likely dependencies:
- `git` transforms → require `git` command
- `json` transforms → suggest `jq` for bash, `json` module for python
- `http` transforms → suggest `curl` for bash, `requests` for python
- `visualization` transforms → suggest `lolcat`, `figlet`

---

## **Level 4: Implementation Steps**

### **Step 1: Design Fresh Agent Guidance System**

#### **1.1 Create AI-Driven Decision Framework (A1.1, A1.2)**
- Design XML-structured decision workflow for Claude
- Implement natural language intent interpretation patterns
- Build discovery-first workflow guidance (search before create)

#### **1.2 Build Port42 Command Guidelines (A2.1, A2.2)**  
- Define discovery command patterns (search, ls /similar/, info)
- Define execution command patterns (run_command, VFS operations)
- Create command usage examples and best practices

#### **1.3 Design Tool Creation Framework (A3.1-A3.4)**
- Define declare command structure and advanced patterns
- Create transform selection guidelines with keyword mapping
- Build prompt crafting patterns for quality and integration requirements

#### **1.4 Implement Context Integration System (A4.1-A4.3)**
- Define reference system usage for file, p42, web, search references
- Create prompt enhancement patterns for context synthesis
- Design reference resolution integration flow

#### **1.5 Expand Artifact Creation Capabilities**
- **Commands**: `port42 declare tool` with multi-language support
- **Artifacts**: Static content with specific categories:
  - **Documentation**: READMEs, API docs, guides, tutorials
  - **Web Applications**: Dashboards, sites, interactive tools  
  - **Configuration Files**: Docker, Kubernetes, CI/CD configs
  - **Diagrams**: Architecture diagrams, flowcharts, system designs
  - **Reports**: Analysis summaries, data reports, presentations
  - **Scripts**: Utility scripts, automation, deployment scripts
  - **Templates**: Code templates, document templates, boilerplate
- Design AI decision logic for artifact type selection based on user intent

#### **1.6 Update agents.json Configuration**
- Replace legacy sections with new XML-structured guidance
- Integrate all frameworks (A1-A4) into agent prompts
- Remove deprecated CommandSpec references

#### **1.7 Unit Testing**
- Test XML prompt parsing and structure validation
- Test decision framework logic with sample user requests  
- Test transform selection accuracy and keyword mapping

### **Step 2: Enhance Tool Materializer**

#### **2.1 Implement Language Selection Logic (B1.1, B1.2)**
- Replace hardcoded Python requirement in `buildToolPrompt()` function
- Implement transform → language mapping algorithm based on keyword analysis
- Add AI-driven language selection prompt integration with tool materializer
- Update prompt generation to include language selection instructions

#### **2.2 Create Multi-Language Template System (B2.1-B2.3)**
- Design and implement Bash template with proper error handling and argument parsing
- Design and implement Python template with argparse and comprehensive imports
- Design and implement Node.js template with modern JavaScript patterns
- Create template selection logic that matches chosen language from step 2.1

#### **2.3 Build Dependency Management System (B3.1-B3.3)**
- Implement external tool detection and validation for each language
- Create dependency validation templates for bash, python, and node
- Build smart dependency inference based on transform keywords
- Add dependency installation guidance and error messages

#### **2.4 Integration and Error Handling**
- Integrate language selection (2.1) with template system (2.2) and dependency management (2.3)
- Implement comprehensive error handling for template generation failures
- Add validation for generated code syntax and structure
- Create fallback mechanisms when language selection or template generation fails

#### **2.5 Unit Testing**
- Test language selection logic with various transform combinations
- Test template generation for all three languages with edge cases
- Test dependency inference accuracy and validation
- Test integrated materializer flow end-to-end

### **Step 3: Create Comprehensive Examples**

#### **3.1 Simple Tool Creation Examples**
- Create basic bash tool examples (git, file processing, text manipulation)
- Create basic python tool examples (data parsing, API calls, analysis)
- Create basic node tool examples (JSON processing, web utilities, modern JS)
- Document transform selection patterns and decision rationale for each

#### **3.2 Complex Tool with References Examples**
- Create tools using `--ref file:` with project configuration context
- Create tools using `--ref p42:/commands/` with existing tool integration
- Create tools using `--ref url:` with external API/standard specifications
- Create tools using `--ref search:` with accumulated knowledge patterns
- Document advanced declare command construction with multiple references

#### **3.3 Multi-Language Showcase Examples**
- Create equivalent tools in bash, python, and node to demonstrate language selection
- Document transform → language mapping decisions with examples
- Show dependency management across different languages
- Demonstrate when to choose each language based on requirements

#### **3.4 Error Handling and Edge Cases**
- Create examples showing graceful dependency failure handling
- Document invalid transform combinations and error recovery
- Show prompt crafting failures and fallback strategies
- Demonstrate reference resolution failures and alternatives

#### **3.5 Documentation and Validation**
- Create comprehensive example documentation with usage patterns
- Validate all examples work end-to-end with current implementation
- Test example accuracy and reference resolution functionality
- Build example test suite for regression testing

### **Step 4: Fresh Architecture Testing & Validation**

#### **4.1 Clean Slate Validation (C1)**
- Test new XML-structured agent prompts (A1.2) without legacy CommandSpec compatibility
- Validate AI decision framework (A1.1) with fresh prompt structures
- Test clean slate data structures in tool materializer (B1-B3) without CommandSpec constraints
- Verify no backward compatibility dependencies remain in codebase

#### **4.2 Fresh Implementation Testing (C2)**
- Validate Agent Instruction Architecture (A1-A4) components work independently
- Test Tool Materialization Flexibility (B1-B3) components work independently  
- Verify new language selection logic works without CommandSpec JSON format
- Test new template system generates valid code for all three languages

#### **4.3 Comprehensive Validation Suite (C3)**
- **Multi-language tool validation**: Test bash, python, node tool generation end-to-end
- **Artifact generation validation**: Test all artifact categories from Step 1.5
- **XML workflow validation**: Test XML-structured decision workflows from A1.2
- **Reference system validation**: Test all reference types (file, p42, url, search) from A4.1

#### **4.4 Integration Testing**
- **Declare → Materialize workflow**: Test tool creation through reality compiler
- **Possess → Declare workflow**: Test full AI agent calling port42 declare command
- **Reference resolution integration**: Test prompt enhancement with resolved context (A4.3)
- **Language selection integration**: Test transform analysis → language choice → template selection

#### **4.5 System Testing**
- **AI decision framework**: Test A1 (tool usage classification) with A3 (declare patterns) integration
- **Tool materializer system**: Test B1 (language selection) + B2 (templates) + B3 (dependencies) working together
- **End-to-end scenarios**: Test complex user requests through complete system
- **Error handling system**: Test failure modes and recovery across all components

#### **4.6 Performance & Acceptance Testing**
- **Performance benchmarking**: Measure new unified architecture performance characteristics
- **Memory usage validation**: Test clean slate data structures for efficiency and resource usage
- **User acceptance scenarios**: Test real-world complex tool creation with references and prompts
- **Regression testing**: Ensure examples from Step 3 continue working with full system

---

## **Current Issues Addressed**

### **Problem 1: Missing Agent Guidance**
**Issue**: Claude in possess mode lacks structured guidance for tool usage vs creation vs port42 operations

**Solution**: Hierarchical instruction architecture (A1-A4) with clear decision trees and examples

### **Problem 2: Hardcoded Python in Declare**
**Issue**: Tool materializer forces Python, losing bash/node flexibility from CommandSpec era  

**Solution**: Multi-language template system (B1-B2) with intelligent language selection

### **Problem 3: Complex Tool Creation Failures**
**Issue**: Claude struggles with sophisticated declare commands and transform selection

**Solution**: Comprehensive frameworks (A3) with transform guidelines and prompt patterns

### **Problem 4: Lost Implementation Quality**
**Issue**: Removed `implementation`, `format_template`, `artifact_guidance` sections

**Solution**: Restore and enhance with modern unified architecture patterns

---

## **Success Metrics**

### **Functionality Restoration**
- ✅ Multi-language tool generation (bash, python, node)
- ✅ Complex tool creation with references and prompts
- ✅ Intelligent transform → language mapping
- ✅ Comprehensive error handling and dependencies

### **User Experience**
- ✅ Clear guidance for when to use existing vs create new tools
- ✅ Sophisticated declare command construction
- ✅ Rich examples and patterns
- ✅ Consistent behavior across simple and complex scenarios

### **Technical Quality**
- ✅ Clean, optimized architecture without legacy constraints
- ✅ Performance optimized for new unified design
- ✅ Comprehensive testing coverage for fresh implementation
- ✅ Clear separation of concerns: commands, documents, living data

**This hierarchical plan addresses all the gaps: proper possess mode guidance, flexible declare implementation, and maintains the unified architecture benefits while restoring lost capabilities.**

---

## **🔮 Future Features**

### **Living Documents System**
*Dynamic data structures for evolving content - prototyped in previous projects*

**Concept**: Living documents are dynamic, self-updating data structures that can evolve and adapt based on context, usage patterns, and user feedback. Unlike static artifacts, they maintain state and can respond to changes in their environment.

**Examples**:
- **Dynamic Configurations**: Config files that adapt based on system usage patterns
- **Evolving Schemas**: Database schemas that self-optimize based on data patterns  
- **Adaptive Datasets**: Training data that incorporates feedback loops for ML models
- **Context-Aware Documentation**: Docs that update based on codebase changes
- **Interactive Dashboards**: Data visualizations that respond to real-time inputs

**Technical Considerations**:
- State persistence and versioning for dynamic content
- Change detection and adaptation triggers
- Conflict resolution for concurrent modifications
- Performance optimization for real-time updates
- Integration with existing Port42 VFS and relation system

**Implementation Complexity**: High - requires sophisticated state management, real-time processing, and complex integration patterns.

**Priority**: Future release after core unified architecture is stable and proven.