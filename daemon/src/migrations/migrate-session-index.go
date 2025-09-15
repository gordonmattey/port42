// +build ignore

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SessionReference matches the existing type in daemon
type SessionReference struct {
	ObjectID         string    `json:"object_id"`
	SessionID        string    `json:"session_id"`
	Agent            string    `json:"agent"`
	CreatedAt        time.Time `json:"created_at"`
	LastUpdated      time.Time `json:"last_updated"`
	CommandGenerated bool      `json:"command_generated"`
	State            string    `json:"state"`
	MessageCount     int       `json:"message_count"`
}

// SessionIndex represents the complete session storage
type SessionIndex struct {
	Sessions     map[string]SessionReference `json:"sessions"`
	LastSessions map[string]string           `json:"last_sessions"`
	Metadata     SessionIndexMetadata        `json:"metadata"`
}

// SessionIndexMetadata contains index-level metadata
type SessionIndexMetadata struct {
	Version       string    `json:"version"`
	LastUpdated   time.Time `json:"last_updated"`
	TotalSessions int       `json:"total_sessions"`
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get home directory:", err)
	}
	port42Dir := filepath.Join(homeDir, ".port42")

	// Paths
	sessionIndexPath := filepath.Join(port42Dir, "session-index.json")
	agentSessionsPath := filepath.Join(port42Dir, "agent_sessions.json")
	backupPath := filepath.Join(port42Dir, "session-index.json.backup")

	// Check if session-index.json exists
	if _, err := os.Stat(sessionIndexPath); os.IsNotExist(err) {
		log.Fatal("session-index.json does not exist")
	}

	// Check if already migrated
	data, err := ioutil.ReadFile(sessionIndexPath)
	if err != nil {
		log.Fatal("Failed to read session-index.json:", err)
	}

	// Try parsing as new format first
	var testIndex map[string]interface{}
	if err := json.Unmarshal(data, &testIndex); err == nil {
		if _, hasMetadata := testIndex["metadata"]; hasMetadata {
			log.Println("✅ session-index.json is already in v2.0 format")
			return
		}
	}

	// Parse old format
	var oldIndex map[string]SessionReference
	if err := json.Unmarshal(data, &oldIndex); err != nil {
		log.Fatal("Failed to parse old session index:", err)
	}
	log.Printf("📚 Found %d sessions in old format", len(oldIndex))

	// Load agent sessions if exists
	agentLastSessions := make(map[string]string)
	if agentData, err := ioutil.ReadFile(agentSessionsPath); err == nil {
		if err := json.Unmarshal(agentData, &agentLastSessions); err == nil {
			log.Printf("📦 Loaded %d agent last sessions from agent_sessions.json", len(agentLastSessions))
		}
	} else {
		// Build from session index if no agent_sessions.json
		log.Printf("📦 No agent_sessions.json found, building from session history...")
		agentLatest := make(map[string]time.Time)
		for sessionID, ref := range oldIndex {
			agent := strings.TrimPrefix(ref.Agent, "@")
			if ref.State != "abandoned" && ref.Agent != "" {
				if ref.LastUpdated.After(agentLatest[agent]) {
					agentLatest[agent] = ref.LastUpdated
					agentLastSessions[agent] = sessionID
				}
			}
		}
		log.Printf("📦 Built last sessions for %d agents from session history", len(agentLastSessions))
	}

	// Normalize agent names in sessions (remove @ prefix)
	normalizedSessions := make(map[string]SessionReference)
	for sessionID, ref := range oldIndex {
		ref.Agent = strings.TrimPrefix(ref.Agent, "@")
		normalizedSessions[sessionID] = ref
	}
	
	// Create new format
	newIndex := SessionIndex{
		Sessions:     normalizedSessions,
		LastSessions: agentLastSessions,
		Metadata: SessionIndexMetadata{
			Version:       "2.0",
			LastUpdated:   time.Now(),
			TotalSessions: len(normalizedSessions),
		},
	}

	// Backup old file
	if err := os.Rename(sessionIndexPath, backupPath); err != nil {
		log.Fatal("Failed to backup old session index:", err)
	}
	log.Printf("💾 Backed up old index to %s", backupPath)

	// Write new format
	newData, err := json.MarshalIndent(newIndex, "", "  ")
	if err != nil {
		// Restore backup on failure
		os.Rename(backupPath, sessionIndexPath)
		log.Fatal("Failed to marshal new session index:", err)
	}

	if err := ioutil.WriteFile(sessionIndexPath, newData, 0644); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, sessionIndexPath)
		log.Fatal("Failed to write new session index:", err)
	}

	log.Printf("✅ Successfully migrated session-index.json to v2.0 format")
	log.Printf("   - %d sessions migrated", len(oldIndex))
	log.Printf("   - %d agent last sessions tracked", len(agentLastSessions))

	// Clean up obsolete files since they're now integrated
	obsoleteFiles := []string{
		agentSessionsPath,
		filepath.Join(port42Dir, "last_session"),
	}

	for _, file := range obsoleteFiles {
		if _, err := os.Stat(file); err == nil {
			if err := os.Remove(file); err != nil {
				log.Printf("⚠️  Failed to remove %s: %v", file, err)
			} else {
				log.Printf("🧹 Removed obsolete file: %s", filepath.Base(file))
			}
		}
	}

	log.Println("")
	log.Println("Next steps:")
	log.Println("1. Update daemon code to use new format")
	log.Println("2. Rebuild and restart daemon")
	log.Println("3. Test session functionality")
}