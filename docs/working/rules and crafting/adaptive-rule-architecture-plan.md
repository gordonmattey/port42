# Adaptive Rule Architecture Plan

*Moving from hardcoded rules to intelligent, configurable pattern recognition*

## 🚨 **The Hardcoding Problem**

### **Current Issues in Stigmergic Rules:**
```go
// TOO RIGID - What if user says "navigate" instead of "cd"?
directoryNames := []string{"cd", "dir", "log", "track", "nav", "jump"}

// TOO SPECIFIC - Can't adapt to user creativity  
directoryTransforms := []string{"directory", "navigation", "tracking", "logging"}

// TOO PRESCRIPTIVE - Assumes exact ecosystem shape
ecosystemTools := []string{"cd-history", "frequent-dirs", "smart-bookmarks"}
```

### **Problems:**
- ❌ **Brittle**: Breaks when users use different terminology
- ❌ **Inflexible**: Can't adapt to creative tool combinations  
- ❌ **Maintenance Heavy**: Every domain needs manual keyword lists
- ❌ **User-Limiting**: Forces users into predefined patterns

## 🏗️ **Hybrid Adaptive Architecture**

### **Core Principles:**
1. **Semantic Understanding** over keyword matching
2. **AI-Driven Pattern Recognition** over hardcoded rules
3. **Template-Based Generation** over specific tool lists
4. **Configuration-Driven** over code-embedded rules
5. **User Extensible** over system-only rule creation

---

## 🧠 **Component 1: Semantic Pattern Detection**

### **Replace Keyword Lists with Embeddings**
```go
type SemanticMatcher struct {
    embeddingService EmbeddingService
    domainVectors    map[string][]float64
}

func (sm *SemanticMatcher) GetToolDomain(relation Relation) (string, float64) {
    toolEmbedding := sm.getToolEmbedding(relation)
    
    bestMatch := ""
    bestScore := 0.0
    
    for domain, domainVector := range sm.domainVectors {
        similarity := cosineSimilarity(toolEmbedding, domainVector)
        if similarity > bestScore {
            bestMatch = domain
            bestScore = similarity
        }
    }
    
    return bestMatch, bestScore
}

func (sm *SemanticMatcher) getToolEmbedding(relation Relation) []float64 {
    // Combine tool name, transforms, and description into semantic vector
    text := fmt.Sprintf("%s %s %s", 
        getRelationName(relation),
        strings.Join(getTransforms(relation), " "),
        getDescription(relation))
    
    return sm.embeddingService.GetEmbedding(text)
}
```

### **Domain Vector Initialization**
```go
// Seed domains with representative text, not keyword lists
var DomainSeeds = map[string]string{
    "directory_intelligence": "directory navigation path location folder change cd tracking history frequent bookmarks workspace project jumping",
    "git_workflow": "git version control commit push pull branch merge status add staging repository",
    "system_monitoring": "process cpu memory disk network performance monitor resource usage diagnostic system health",
    "file_management": "file copy move backup sync organize duplicate clean sort filter search",
    "development_tools": "build compile test debug deploy package environment setup configuration",
}

func initializeDomainVectors() map[string][]float64 {
    vectors := make(map[string][]float64)
    for domain, seedText := range DomainSeeds {
        vectors[domain] = embeddingService.GetEmbedding(seedText)
    }
    return vectors
}
```

---

## 🤖 **Component 2: AI-Driven Ecosystem Generation**

### **Dynamic Tool Suggestion Instead of Hardcoded Lists**
```go
type EcosystemGenerator struct {
    aiService AIService
}

func (eg *EcosystemGenerator) GenerateEcosystem(tools []Relation, domain string) (*EcosystemPlan, error) {
    prompt := fmt.Sprintf(`
    The user has created these tools in the %s domain:
    %s
    
    What 2-3 complementary tools would complete this ecosystem?
    Consider: missing functionality, common workflows, user pain points.
    
    Respond with JSON:
    {
        "suggested_tools": [
            {
                "name": "tool-name",
                "transforms": ["transform1", "transform2"],
                "description": "what it does",
                "rationale": "why it completes the ecosystem"
            }
        ]
    }
    `, domain, formatToolsForAI(tools))
    
    return eg.aiService.GenerateEcosystem(prompt)
}

type EcosystemPlan struct {
    SuggestedTools []ToolSuggestion `json:"suggested_tools"`
}

type ToolSuggestion struct {
    Name        string   `json:"name"`
    Transforms  []string `json:"transforms"`
    Description string   `json:"description"`
    Rationale   string   `json:"rationale"`
}
```

### **Smart Template System**
```go
type ToolTemplate struct {
    NamePattern    string            `json:"name_pattern"`     // "{domain}-history"
    Transforms     []string          `json:"transforms"`       // ["{domain}", "history"]
    Description    string            `json:"description"`      // "Shows {domain} history"
    Variables      map[string]string `json:"variables"`        // {"domain": "directory"}
    Conditions     []string          `json:"conditions"`       // ["has_tracking_capability"]
}

func (tt *ToolTemplate) Instantiate(variables map[string]string) Relation {
    name := replacePlaceholders(tt.NamePattern, variables)
    transforms := replaceInSlice(tt.Transforms, variables)
    description := replacePlaceholders(tt.Description, variables)
    
    return Relation{
        Type: "Tool",
        Properties: map[string]interface{}{
            "name":        name,
            "transforms":  transforms,
            "description": description,
            "auto_spawned": true,
            "generated_by": "ecosystem-template",
            "template_id":  tt.NamePattern,
        },
    }
}
```

---

## 📋 **Component 3: Configuration-Driven Rules**

### **YAML Rule Configuration**
```yaml
# crafting_rules.yaml
crafting_rules:
  - id: "directory_intelligence_crafting"
    name: "Directory Intelligence Crafting"
    description: "Completes directory navigation ecosystems"
    
    triggers:
      semantic_domain: "directory_intelligence"
      similarity_threshold: 0.75
      min_tools: 2
      max_tools: 10
      
    conditions:
      - type: "semantic_clustering"
        params:
          cluster_threshold: 0.7
      - type: "capability_gap"
        params:
          required_capabilities: ["tracking", "navigation"]
          
    generation_strategy: "ai_driven"
    
    templates:
      - name_pattern: "{domain}-history"
        transforms: ["{domain}", "history", "analytics"]
        description: "Shows {domain} navigation history"
        conditions: ["has_tracking_tool"]
        
      - name_pattern: "frequent-{items}"
        transforms: ["{domain}", "frequency", "analytics"] 
        description: "Shows most frequently accessed {items}"
        conditions: ["has_usage_data"]
        
      - name_pattern: "{domain}-bookmarks"
        transforms: ["{domain}", "bookmarks", "quick-access"]
        description: "Quick bookmarking for {domain} items"
        conditions: ["has_navigation_tool"]

  - id: "git_workflow_crafting"
    name: "Git Workflow Crafting"
    # ... similar structure for git domain
```

### **Rule Configuration Loader**
```go
type RuleConfig struct {
    ID                  string                 `yaml:"id"`
    Name                string                 `yaml:"name"`
    Description         string                 `yaml:"description"`
    Triggers            TriggerConfig          `yaml:"triggers"`
    Conditions          []ConditionConfig      `yaml:"conditions"`
    GenerationStrategy  string                 `yaml:"generation_strategy"`
    Templates           []ToolTemplate         `yaml:"templates"`
}

type TriggerConfig struct {
    SemanticDomain       string  `yaml:"semantic_domain"`
    SimilarityThreshold  float64 `yaml:"similarity_threshold"`
    MinTools            int     `yaml:"min_tools"`
    MaxTools            int     `yaml:"max_tools"`
}

func LoadCraftingRules(configPath string) ([]Rule, error) {
    var config struct {
        CraftingRules []RuleConfig `yaml:"crafting_rules"`
    }
    
    data, err := ioutil.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    var rules []Rule
    for _, ruleConfig := range config.CraftingRules {
        rule := convertConfigToRule(ruleConfig)
        rules = append(rules, rule)
    }
    
    return rules, nil
}
```

---

## 🎯 **Component 4: Universal Crafting Engine**

### **Unified Rule Implementation**
```go
func createAdaptiveCraftingRule(config RuleConfig) Rule {
    return Rule{
        ID:          config.ID,
        Name:        config.Name,
        Description: config.Description,
        Enabled:     true,
        
        Condition: func(relation Relation) bool {
            // Skip auto-spawned tools
            if isAutoSpawned(relation) { return false }
            
            // Check semantic domain match
            domain, similarity := semanticMatcher.GetToolDomain(relation)
            if domain != config.Triggers.SemanticDomain { return false }
            if similarity < config.Triggers.SimilarityThreshold { return false }
            
            // Count similar tools in domain
            similarTools := findSimilarToolsInDomain(relation, domain)
            if len(similarTools) < config.Triggers.MinTools { return false }
            
            // Check if ecosystem already exists
            if hasEcosystemTools(similarTools, config.Templates) { return false }
            
            return true
        },
        
        Action: func(relation Relation, compiler *RealityCompiler) error {
            domain, _ := semanticMatcher.GetToolDomain(relation)
            similarTools := findSimilarToolsInDomain(relation, domain)
            
            switch config.GenerationStrategy {
            case "ai_driven":
                return generateEcosystemAI(similarTools, domain, compiler)
            case "template_based":
                return generateEcosystemTemplates(similarTools, config.Templates, compiler)
            default:
                return fmt.Errorf("unknown generation strategy: %s", config.GenerationStrategy)
            }
        },
    }
}
```

### **AI-Driven Generation Implementation**
```go
func generateEcosystemAI(tools []Relation, domain string, compiler *RealityCompiler) error {
    generator := NewEcosystemGenerator(aiService)
    plan, err := generator.GenerateEcosystem(tools, domain)
    if err != nil {
        return fmt.Errorf("failed to generate ecosystem: %w", err)
    }
    
    for _, suggestion := range plan.SuggestedTools {
        relation := Relation{
            ID:   generateRelationID("Tool", suggestion.Name),
            Type: "Tool",
            Properties: map[string]interface{}{
                "name":         suggestion.Name,
                "transforms":   suggestion.Transforms,
                "description":  suggestion.Description,
                "auto_spawned": true,
                "crafted_by":   domain + "_crafting",
                "rationale":    suggestion.Rationale,
            },
            CreatedAt: time.Now(),
        }
        
        log.Printf("🧩 AI-crafting tool: %s (%s)", suggestion.Name, suggestion.Rationale)
        
        _, err := compiler.DeclareRelation(relation)
        if err != nil {
            log.Printf("❌ Failed to craft %s: %v", suggestion.Name, err)
            continue
        }
        
        log.Printf("✅ Successfully AI-crafted: %s", suggestion.Name)
    }
    
    return nil
}
```

---

## 🚀 **Implementation Roadmap**

### **Phase 1: Semantic Foundation (Week 1)**
- [ ] Implement semantic embedding service integration
- [ ] Create domain vector initialization system
- [ ] Replace hardcoded keywords with similarity matching
- [ ] Test with directory intelligence use case

### **Phase 2: AI-Driven Generation (Week 2)**  
- [ ] Implement EcosystemGenerator with AI service
- [ ] Create tool suggestion prompt engineering
- [ ] Add template instantiation system
- [ ] Test dynamic ecosystem generation

### **Phase 3: Configuration System (Week 3)**
- [ ] Design YAML rule configuration format
- [ ] Implement rule loading and parsing
- [ ] Create adaptive rule factory
- [ ] Add configuration validation

### **Phase 4: User Extensibility (Week 4)**
- [ ] Allow users to define custom domains
- [ ] Enable user-created rule templates
- [ ] Add rule sharing and import system
- [ ] Create rule debugging tools

---

## 🧪 **Testing Strategy**

### **Semantic Accuracy Tests**
```go
func TestSemanticDomainDetection(t *testing.T) {
    tests := []struct {
        toolName    string
        transforms  []string
        expectedDomain string
        minSimilarity float64
    }{
        {"logcd", []string{"directory", "logging"}, "directory_intelligence", 0.7},
        {"git-status", []string{"git", "status"}, "git_workflow", 0.8},
        {"cpu-monitor", []string{"system", "monitoring"}, "system_monitoring", 0.75},
    }
    
    for _, test := range tests {
        relation := createTestRelation(test.toolName, test.transforms)
        domain, similarity := semanticMatcher.GetToolDomain(relation)
        
        assert.Equal(t, test.expectedDomain, domain)
        assert.GreaterOrEqual(t, similarity, test.minSimilarity)
    }
}
```

### **Ecosystem Generation Tests**
```go
func TestEcosystemGeneration(t *testing.T) {
    // Test that logcd + smart-cd generates appropriate ecosystem
    tools := []Relation{
        createTestRelation("logcd", []string{"directory", "logging"}),
        createTestRelation("smart-cd", []string{"directory", "enhancement"}),
    }
    
    plan, err := ecosystemGenerator.GenerateEcosystem(tools, "directory_intelligence")
    require.NoError(t, err)
    
    // Should suggest history/analytics tools
    assert.Contains(t, plan.SuggestedTools, toolWithNameContaining("history"))
    assert.Contains(t, plan.SuggestedTools, toolWithNameContaining("frequent"))
}
```

---

## 🌟 **Benefits of Hybrid Approach**

### **For Users:**
- ✅ **Creative Freedom**: Use any terminology, system adapts
- ✅ **Intelligent Assistance**: System understands intent, not just keywords  
- ✅ **Extensible**: Users can define custom domains and rules
- ✅ **Predictable**: Template system provides consistent patterns

### **For Developers:**
- ✅ **Maintainable**: Configuration-driven, not code-embedded
- ✅ **Scalable**: AI handles edge cases and creativity
- ✅ **Testable**: Clear separation between detection and generation
- ✅ **Debuggable**: Rule execution is traceable and configurable

### **For System:**
- ✅ **Adaptive**: Learns from user patterns and terminology
- ✅ **Robust**: Graceful degradation when AI unavailable
- ✅ **Efficient**: Semantic caching and batch processing
- ✅ **Evolvable**: New domains and patterns emerge organically

This architecture transforms rule creation from **"programming patterns"** to **"teaching the system about domains"** - a much more sustainable and user-friendly approach.

---

## 🌐 **Future: Universal State Perception Engine**

*Beyond local system monitoring to universal perception of any stateful environment*

### **Expanding Beyond Local System Triggers**

The current architecture focuses on **relation triggers** (tool creation) and **local system triggers** (file changes, git status). However, rules should respond to state changes across **any system we're tracking** - transforming Port 42 into a universal intelligent assistant.

### **Universal State Perception Types**

#### **Local Environment Perception:**
```go
type FilesystemPerception struct{}  // File changes, directory structure
type GitPerception struct{}        // Repo status, commits, branches  
type ProcessPerception struct{}    // Running services, resource usage
type NetworkPerception struct{}    // Open ports, connections, traffic
```

#### **Remote System Perception:**
```go
type AWSPerception struct{}        // EC2, RDS, S3 bucket changes
type DockerPerception struct{}     // Container lifecycle, health checks
type KubernetesPerception struct{} // Pod status, service discovery
type CIPipelinePerception struct{} // Build status, deployment events
```

#### **External Service Perception:**
```go
type SlackPerception struct{}      // Mentions, channel activity, status
type JiraPerception struct{}       // Issue updates, sprint changes
type PagerDutyPerception struct{}  // Incident lifecycle, alert status
type CalendarPerception struct{}   // Meeting events, schedule changes
```

### **Perception-Driven Architecture**

#### **Universal State Change Model**
```go
type StateChange struct {
    Source      string                 // "filesystem", "git", "aws", "slack", "ci"
    Type        string                 // "file_modified", "commit_pushed", "instance_started"
    Entity      string                 // "/project/file.js", "main", "i-1234567"
    OldState    map[string]interface{} // Previous state snapshot
    NewState    map[string]interface{} // New state snapshot
    Timestamp   time.Time
    Metadata    map[string]interface{} // Context-specific data
}

type PerceptionRule struct {
    ID          string
    Name        string
    Description string
    
    // What state changes trigger this rule
    Triggers    []PerceptionTrigger
    
    // Conditions across multiple perceptions  
    Condition   func(change StateChange, allStates map[string]PerceptionState) bool
    
    // Actions can affect any system
    Action      func(change StateChange, compiler *RealityCompiler) error
}
```

#### **Cross-System Intelligence Examples**

**Development Flow Intelligence:**
```go
devFlowRule := PerceptionRule{
    Name: "Development Flow Orchestrator",
    Triggers: []PerceptionTrigger{
        {Source: "git", EventTypes: []string{"commit_pushed"}},
        {Source: "ci", EventTypes: []string{"build_completed"}},
        {Source: "aws", EventTypes: []string{"deployment_successful"}},
    },
    
    Action: func(change StateChange, compiler *RealityCompiler) error {
        // Auto-spawn: smoke-test, notify-team, update-docs tools
        return spawnPostDeploymentTools(change, compiler)
    },
}
```

**Incident Response Intelligence:**
```go
incidentRule := PerceptionRule{
    Name: "Incident Response Orchestrator",
    Triggers: []PerceptionTrigger{
        {Source: "pagerduty", EventTypes: []string{"incident_triggered"}},
        {Source: "aws", EventTypes: []string{"service_unhealthy"}},
        {Source: "monitoring", EventTypes: []string{"alert_fired"}},
    },
    
    Action: func(change StateChange, compiler *RealityCompiler) error {
        // Auto-spawn: gather-logs, check-dependencies, escalation-helper tools
        return spawnIncidentResponseTools(change, compiler)
    },
}
```

**Team Collaboration Intelligence:**
```go
teamSyncRule := PerceptionRule{
    Name: "Team Synchronization Intelligence", 
    Triggers: []PerceptionTrigger{
        {Source: "slack", EventTypes: []string{"mention_received", "urgent_message"}},
        {Source: "jira", EventTypes: []string{"ticket_assigned", "sprint_started"}},
        {Source: "calendar", EventTypes: []string{"meeting_starting"}},
    },
    
    Action: func(change StateChange, compiler *RealityCompiler) error {
        // Auto-spawn: context-gatherer, status-updater, meeting-prep tools
        return spawnCollaborationTools(change, compiler)
    },
}
```

### **Configuration-Driven Universal Perception**

```yaml
# perceptions.yaml
perceptions:
  - type: "git"
    config:
      repositories: [".", "../other-repo"]
      watch_events: ["commit", "push", "merge", "branch_created"]
      
  - type: "aws"
    config:
      regions: ["us-east-1", "us-west-2"] 
      services: ["ec2", "rds", "lambda", "s3"]
      watch_events: ["instance_state_change", "deployment_complete"]
      
  - type: "slack"
    config:
      channels: ["#engineering", "#alerts", "#deployments"]
      watch_events: ["mention", "urgent_keyword", "bot_alert"]
      keywords: ["@gordon", "production", "incident", "urgent"]

perception_rules:
  - name: "Full Stack Development Flow"
    triggers:
      - source: "filesystem"
        events: ["file_saved"]
        filter: "*.js,*.py,*.go"
      - source: "git"
        events: ["commit_created"]
      - source: "ci"
        events: ["build_started", "tests_passed"]
        
    spawns:
      - name: "dev-flow-status"
        triggers_on: "any_event"
        description: "Shows current development flow status"
      - name: "auto-deployer"
        triggers_on: "tests_passed"
        description: "Automated deployment pipeline"
```

### **Universal Intelligence Benefits**

#### **Cross-System Awareness:**
- **DevOps Intelligence**: Code changes → CI status → deployment health → incident response
- **Team Intelligence**: Slack activity → Jira updates → calendar events → collaboration tools  
- **Infrastructure Intelligence**: Resource usage → scaling events → cost optimization tools

#### **Contextual Tool Spawning:**
- **Environment-Aware**: Different tools for dev/staging/prod environments
- **Team-Aware**: Different tools based on who's online, team status, workload
- **Time-Aware**: Different tools for business hours vs off-hours incidents

#### **Proactive Assistance:**
- **Predictive**: Spawn tools before problems become critical
- **Adaptive**: Learn patterns across all monitored systems
- **Intelligent**: Understand relationships between different system states

### **Implementation Vision**

```go
type PerceptionManager struct {
    perceptions map[string]StatePerception
    ruleEngine  *EnhancedRuleEngine
    stateStore  PerceptionStateStore
}

func (pm *PerceptionManager) RegisterPerception(p StatePerception) error {
    source := p.GetPerceptionType()
    pm.perceptions[source] = p
    
    // Subscribe to state changes
    return p.Subscribe(func(change StateChange) {
        // Store state change
        pm.stateStore.RecordChange(change)
        
        // Trigger rules across all perceptions
        pm.ruleEngine.ProcessStateChange(change, pm.getAllCurrentStates())
    })
}
```

This vision transforms Port 42 from a **local tool generator** into a **universal intelligent assistant** that perceives and responds to your entire digital environment - creating truly ambient intelligence that anticipates needs across all systems you interact with.