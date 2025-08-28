package main

import (
	"strings"
)

// extractSessionTags extracts relevant tags from a session
func extractSessionTags(session *Session) []string {
	tags := []string{"conversation", "ai", session.Agent}
	
	// Add state tag
	tags = append(tags, string(session.State))
	
	// Add command tag if generated
	if session.CommandGenerated != nil {
		tags = append(tags, "command-generated")
	}
	
	// Extract keywords from messages
	for _, msg := range session.Messages {
		words := strings.Fields(msg.Content)
		for _, word := range words {
			word = strings.ToLower(word)
			if len(word) > 4 && !isCommonWord(word) {
				tags = append(tags, word)
			}
		}
	}
	
	return tags
}

// mapStateToLifecycle maps session state to lifecycle status
func mapStateToLifecycle(state SessionState) string {
	switch state {
	case SessionActive, SessionIdle:
		return "active"
	case SessionCompleted:
		return "stable"
	case SessionAbandoned:
		return "archived"
	default:
		return "draft"
	}
}

// cleanAgentName removes special characters from agent names
func cleanAgentName(agent string) string {
	// Remove @ prefix and any special characters
	cleaned := strings.TrimPrefix(agent, "@")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	return cleaned
}

// isCommonWord checks if a word is too common to be a useful tag
func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true, "but": true,
		"not": true, "you": true, "all": true, "can": true, "her": true,
		"was": true, "one": true, "our": true, "out": true, "day": true,
		"had": true, "has": true, "his": true, "how": true, "its": true,
		"may": true, "new": true, "now": true, "old": true, "see": true,
		"two": true, "way": true, "who": true, "boy": true, "did": true,
		"get": true, "got": true, "him": true, "let": true, "put": true,
		"say": true, "she": true, "too": true, "use": true, "will": true,
		"with": true, "have": true, "this": true, "that": true, "from": true,
		"what": true, "when": true, "where": true, "which": true, "some": true,
		"would": true, "there": true, "their": true, "about": true, "after": true,
		"before": true, "could": true, "should": true, "other": true, "because": true,
	}
	return commonWords[word]
}