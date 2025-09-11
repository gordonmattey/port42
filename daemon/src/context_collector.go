package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ContextCollector collects and manages context data for the daemon
type ContextCollector struct {
	mu             sync.RWMutex
	daemon         *Daemon
	recentCommands []CommandRecord
	createdTools   []ToolRecord
	maxCommands    int
	maxTools       int
}

// NewContextCollector creates a new context collector
func NewContextCollector(daemon *Daemon) *ContextCollector {
	return &ContextCollector{
		daemon:         daemon,
		maxCommands:    20,
		maxTools:       10,
		recentCommands: make([]CommandRecord, 0, 20),
		createdTools:   make([]ToolRecord, 0, 10),
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
	cc.mu.RUnlock()
	
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
			Command:    fmt.Sprintf("port42 possess %s --session last", data.ActiveSession.Agent),
			Reason:     "Continue your conversation",
			Confidence: 0.85,
		})
	} else {
		// No active session - suggest starting one
		suggestions = append(suggestions, ContextSuggestion{
			Command:    "port42 possess @ai-engineer \"How can I help?\"",
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