package suggest

import (
	"fmt"
	"sort"
	"strings"
)

// Match represents a suggested match with its score
type Match struct {
	Value string
	Label string
	Score int
}

// FindSimilar finds items similar to the query using simple substring matching.
// Returns up to maxResults matches, sorted by relevance.
func FindSimilar(query string, items []Match, maxResults int) []Match {
	if query == "" || len(items) == 0 {
		return nil
	}

	query = strings.ToLower(query)
	var matches []Match

	for _, item := range items {
		score := calculateScore(query, strings.ToLower(item.Value), strings.ToLower(item.Label))
		if score > 0 {
			item.Score = score
			matches = append(matches, item)
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return matches
}

// calculateScore returns a relevance score (higher = better match)
func calculateScore(query, value, label string) int {
	score := 0

	// Exact match
	if value == query {
		return 100
	}

	// Starts with query
	if strings.HasPrefix(value, query) {
		score += 50
	}

	// Contains query
	if strings.Contains(value, query) {
		score += 30
	}

	// Label contains query
	if strings.Contains(label, query) {
		score += 20
	}

	// Partial ID match (last part after underscore)
	if idx := strings.LastIndex(query, "_"); idx >= 0 {
		suffix := query[idx+1:]
		if strings.Contains(value, suffix) {
			score += 15
		}
	}

	return score
}

// FormatSuggestions formats matches for display
func FormatSuggestions(matches []Match) string {
	if len(matches) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\nDid you mean one of these?\n")
	for _, m := range matches {
		if m.Label != "" {
			sb.WriteString(fmt.Sprintf("  %s  %s\n", m.Value, m.Label))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", m.Value))
		}
	}
	return sb.String()
}
