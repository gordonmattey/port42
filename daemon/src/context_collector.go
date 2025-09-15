package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ContextCollector collects and manages context data for the daemon
type ContextCollector struct {
	mu               sync.RWMutex
	daemon           *Daemon
	recentCommands   []CommandRecord
	createdTools     []ToolRecord
	accessedMemories map[string]*MemoryAccess // path -> access info
	maxCommands      int
	maxTools         int
	maxMemories      int
}

// NewContextCollector creates a new context collector
func NewContextCollector(daemon *Daemon) *ContextCollector {
	return &ContextCollector{
		daemon:           daemon,
		maxCommands:      30,  // Increased to show more activity history
		maxTools:         10,
		maxMemories:      15,
		recentCommands:   make([]CommandRecord, 0, 30),
		createdTools:     make([]ToolRecord, 0, 10),
		accessedMemories: make(map[string]*MemoryAccess),
	}
}

// TrackCommand records a command execution
func (cc *ContextCollector) TrackCommand(cmd string, exitCode int) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	record := CommandRecord{
		Command:   cmd,
		Timestamp: time.Now(),
		ExitCode:  exitCode,
	}
	
	// Add to front of slice (most recent first)
	cc.recentCommands = append([]CommandRecord{record}, cc.recentCommands...)
	
	// Trim to max size
	if len(cc.recentCommands) > cc.maxCommands {
		cc.recentCommands = cc.recentCommands[:cc.maxCommands]
	}
	
	log.Printf("📝 Tracked command: %s (exit: %d)", cmd, exitCode)
}

// TrackToolCreation records when a tool is created
func (cc *ContextCollector) TrackToolCreation(name string, toolType string, transforms []string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	record := ToolRecord{
		Name:      name,
		Type:      toolType,
		Transforms: transforms,
		CreatedAt: time.Now(),
	}
	
	// Add to front of slice (most recent first)
	cc.createdTools = append([]ToolRecord{record}, cc.createdTools...)
	
	// Trim to max size
	if len(cc.createdTools) > cc.maxTools {
		cc.createdTools = cc.createdTools[:cc.maxTools]
	}
	
	log.Printf("🛠 Tracked tool creation: %s (type: %s)", name, toolType)
}

// TrackMemoryAccess records when a memory or artifact is accessed
func (cc *ContextCollector) TrackMemoryAccess(path string, accessType string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	
	// Update or create memory access record
	if access, exists := cc.accessedMemories[path]; exists {
		access.AccessCount++
	} else {
		// Generate human-readable display name
		displayName := cc.generateDisplayName(path, accessType)
		
		cc.accessedMemories[path] = &MemoryAccess{
			Path:        path,
			Type:        accessType,
			AccessCount: 1,
			DisplayName: displayName,
		}
	}
	
	// Trim if we have too many tracked memories
	if len(cc.accessedMemories) > cc.maxMemories {
		// Find and remove least accessed memory
		var minPath string
		minCount := int(^uint(0) >> 1) // Max int
		for path, access := range cc.accessedMemories {
			if access.AccessCount < minCount {
				minCount = access.AccessCount
				minPath = path
			}
		}
		if minPath != "" {
			delete(cc.accessedMemories, minPath)
		}
	}
}

// Collect gathers all context data
func (cc *ContextCollector) Collect() *ContextData {
	data := &ContextData{
		RecentCommands:   []CommandRecord{},
		CreatedTools:     []ToolRecord{},
		Suggestions:      []ContextSuggestion{},
		AccessedMemories: []MemoryAccess{},
	}
	
	// Get active session
	activeSession := cc.getActiveSession()
	if activeSession != nil {
		data.ActiveSession = &ActiveSessionInfo{
			ID:           activeSession.ID,
			Agent:        activeSession.Agent,
			MessageCount: len(activeSession.Messages),
			StartTime:    activeSession.CreatedAt,
			LastActivity: activeSession.LastActivity,
			State:        string(activeSession.State),
		}
		
		// Add tool created info if present
		if activeSession.CommandGenerated != nil {
			toolName := activeSession.CommandGenerated.Name
			data.ActiveSession.ToolCreated = &toolName
			
			// Also add to created tools list
			data.CreatedTools = append(data.CreatedTools, ToolRecord{
				Name:      toolName,
				Type:      "tool",
				CreatedAt: activeSession.LastActivity,
			})
		}
	}
	
	// Get recent commands with age calculation
	data.RecentCommands = cc.getRecentCommands()
	
	// Get created tools
	cc.mu.RLock()
	data.CreatedTools = append(data.CreatedTools, cc.createdTools...)
	
	// Get accessed memories (convert map to slice)
	for _, access := range cc.accessedMemories {
		data.AccessedMemories = append(data.AccessedMemories, *access)
	}
	cc.mu.RUnlock()
	
	// Sort accessed memories by access count (most accessed first)
	if len(data.AccessedMemories) > 1 {
		for i := 0; i < len(data.AccessedMemories)-1; i++ {
			for j := i + 1; j < len(data.AccessedMemories); j++ {
				if data.AccessedMemories[j].AccessCount > data.AccessedMemories[i].AccessCount {
					data.AccessedMemories[i], data.AccessedMemories[j] = data.AccessedMemories[j], data.AccessedMemories[i]
				}
			}
		}
	}
	
	// Generate contextual suggestions
	data.Suggestions = cc.generateSuggestions(data)
	
	return data
}

// getActiveSession finds the most recent active session
func (cc *ContextCollector) getActiveSession() *Session {
	cc.daemon.mu.RLock()
	defer cc.daemon.mu.RUnlock()
	
	var activeSession *Session
	var latestTime time.Time
	
	for _, session := range cc.daemon.sessions {
		if session.State == SessionActive && session.LastActivity.After(latestTime) {
			activeSession = session
			latestTime = session.LastActivity
		}
	}
	
	return activeSession
}

// getRecentCommands returns recent commands with age calculation
func (cc *ContextCollector) getRecentCommands() []CommandRecord {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	
	now := time.Now()
	result := make([]CommandRecord, 0, len(cc.recentCommands))
	
	for _, cmd := range cc.recentCommands {
		cmdCopy := cmd
		cmdCopy.AgeSeconds = int(now.Sub(cmd.Timestamp).Seconds())
		result = append(result, cmdCopy)
	}
	
	return result
}

// generateSuggestions creates contextual command suggestions
func (cc *ContextCollector) generateSuggestions(data *ContextData) []ContextSuggestion {
	suggestions := []ContextSuggestion{}
	
	if data.ActiveSession != nil {
		// Suggest viewing session details
		suggestions = append(suggestions, ContextSuggestion{
			Command:    fmt.Sprintf("port42 info /memory/%s", data.ActiveSession.ID),
			Reason:     "View current session details",
			Confidence: 0.9,
		})
		
		// If tool was created, suggest using it
		if data.ActiveSession.ToolCreated != nil {
			suggestions = append(suggestions, ContextSuggestion{
				Command:    fmt.Sprintf("%s --help", *data.ActiveSession.ToolCreated),
				Reason:     "Learn about your new tool",
				Confidence: 0.95,
			})
		}
		
		// Suggest continuing session
		suggestions = append(suggestions, ContextSuggestion{
			Command:    fmt.Sprintf("port42 swim %s --session last", data.ActiveSession.Agent),
			Reason:     "Continue your conversation",
			Confidence: 0.85,
		})
	} else {
		// No active session - suggest starting one
		suggestions = append(suggestions, ContextSuggestion{
			Command:    "port42 swim @ai-engineer \"How can I help?\"",
			Reason:     "Start a new AI session",
			Confidence: 0.8,
		})
	}
	
	// Based on recent commands, suggest related actions
	if len(data.RecentCommands) > 0 {
		lastCmd := data.RecentCommands[0]
		
		// If last command was search, suggest exploring results
		if lastCmd.Command == "search" || lastCmd.Command == "ls" {
			suggestions = append(suggestions, ContextSuggestion{
				Command:    "port42 ls /tools/",
				Reason:     "Explore available tools",
				Confidence: 0.7,
			})
		}
		
		// If context was checked, suggest watch mode
		if lastCmd.Command == "context" {
			suggestions = append(suggestions, ContextSuggestion{
				Command:    "port42 context --watch",
				Reason:     "Monitor context in real-time",
				Confidence: 0.75,
			})
		}
	}
	
	// Limit to top 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	
	return suggestions
}

// generateDisplayName creates a human-readable name for a path
func (cc *ContextCollector) generateDisplayName(path string, accessType string) string {
	// Handle memory/session paths
	if strings.HasPrefix(path, "/memory/") {
		sessionID := strings.TrimPrefix(path, "/memory/")
		
		// Try to get session info for better display name
		cc.daemon.mu.RLock()
		if session, exists := cc.daemon.sessions[sessionID]; exists {
			agent := session.Agent
			cc.daemon.mu.RUnlock()
			
			// For created memories, show agent and action
			if accessType == "created" {
				return fmt.Sprintf("New %s session", agent)
			}
			
			// For accessed memories, show agent and first message snippet
			if len(session.Messages) > 0 {
				firstMsg := session.Messages[0].Content
				if len(firstMsg) > 30 {
					firstMsg = firstMsg[:30] + "..."
				}
				return fmt.Sprintf("%s: %s", agent, firstMsg)
			}
			
			return fmt.Sprintf("%s session", agent)
		}
		cc.daemon.mu.RUnlock()
		
		// Fallback to shorter session ID
		if len(sessionID) > 20 {
			return fmt.Sprintf("Session ...%s", sessionID[len(sessionID)-8:])
		}
		return sessionID
	}
	
	// Handle command paths
	if strings.HasPrefix(path, "/commands/") {
		cmdName := strings.TrimPrefix(path, "/commands/")
		return cmdName
	}
	
	// Handle tool paths
	if strings.HasPrefix(path, "/tools/") {
		toolPath := strings.TrimPrefix(path, "/tools/")
		// Remove trailing slashes
		toolPath = strings.TrimSuffix(toolPath, "/")
		
		// If it's just browsing /tools/, show that
		if toolPath == "" {
			return "Tools directory"
		}
		
		// For specific tools, show the tool name
		parts := strings.Split(toolPath, "/")
		if len(parts) > 0 {
			return parts[0]
		}
		return toolPath
	}
	
	// Handle artifact paths
	if strings.HasPrefix(path, "/artifacts/") {
		artifactPath := strings.TrimPrefix(path, "/artifacts/")
		parts := strings.Split(artifactPath, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return artifactPath
	}
	
	// Default: return the last part of the path
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if lastPart != "" {
			return lastPart
		}
	}
	
	return path
}