package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

// MigrateMemoryPaths updates existing metadata to include new path formats
func (s *Storage) MigrateMemoryPaths() error {
	log.Println("🔄 Starting memory path migration...")
	
	metadataFiles, err := filepath.Glob(filepath.Join(s.metadataDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list metadata files: %w", err)
	}
	
	updated := 0
	skipped := 0
	
	for _, metaPath := range metadataFiles {
		data, err := ioutil.ReadFile(metaPath)
		if err != nil {
			log.Printf("⚠️  Failed to read %s: %v", metaPath, err)
			continue
		}
		
		var meta Metadata
		if err := json.Unmarshal(data, &meta); err != nil {
			log.Printf("⚠️  Failed to parse %s: %v", metaPath, err)
			continue
		}
		
		// Only process session types
		if meta.Type != "session" {
			skipped++
			continue
		}
		
		// Check if already has new format (has leading slashes)
		hasNewFormat := false
		for _, path := range meta.Paths {
			if strings.HasPrefix(path, "/") {
				hasNewFormat = true
				break
			}
		}
		
		if hasNewFormat {
			skipped++
			continue
		}
		
		// Update paths
		newPaths := []string{}
		
		// Add direct memory path
		if meta.Session != "" {
			newPaths = append(newPaths, fmt.Sprintf("/memory/%s", meta.Session))
		}
		
		// Update existing paths with leading slashes
		for _, oldPath := range meta.Paths {
			if !strings.HasPrefix(oldPath, "/") {
				newPaths = append(newPaths, "/"+oldPath)
			} else {
				newPaths = append(newPaths, oldPath)
			}
		}
		
		// Add additional global paths if they don't exist
		if meta.Session != "" && meta.Created.IsZero() == false {
			dateStr := meta.Created.Format("2006-01-02")
			globalDatePath := fmt.Sprintf("/by-date/%s/memory/%s", dateStr, meta.Session)
			
			hasGlobalDate := false
			for _, p := range newPaths {
				if p == globalDatePath {
					hasGlobalDate = true
					break
				}
			}
			if !hasGlobalDate {
				newPaths = append(newPaths, globalDatePath)
			}
			
			// Add global agent path
			if meta.Agent != "" {
				globalAgentPath := fmt.Sprintf("/by-agent/%s/memory/%s", 
					cleanAgentName(meta.Agent), meta.Session)
				hasGlobalAgent := false
				for _, p := range newPaths {
					if p == globalAgentPath {
						hasGlobalAgent = true
						break
					}
				}
				if !hasGlobalAgent {
					newPaths = append(newPaths, globalAgentPath)
				}
			}
		}
		
		// Update metadata
		meta.Paths = newPaths
		
		// Save updated metadata
		updatedData, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			log.Printf("⚠️  Failed to marshal updated metadata for %s: %v", meta.ID, err)
			continue
		}
		
		if err := ioutil.WriteFile(metaPath, updatedData, 0644); err != nil {
			log.Printf("⚠️  Failed to write updated metadata for %s: %v", meta.ID, err)
			continue
		}
		
		updated++
		log.Printf("✅ Updated paths for session %s", meta.Session)
	}
	
	log.Printf("🎉 Migration complete: %d updated, %d skipped", updated, skipped)
	return nil
}