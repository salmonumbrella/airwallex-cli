package suggest

import (
	"strings"
	"testing"
)

func TestFindSimilar(t *testing.T) {
	items := []Match{
		{Value: "ben_abc123", Label: "John Smith (US)"},
		{Value: "ben_def456", Label: "Jane Doe (UK)"},
		{Value: "ben_xyz789", Label: "Bob Wilson (CA)"},
	}

	// Test partial match
	matches := FindSimilar("ben_abc", items, 3)
	if len(matches) == 0 {
		t.Error("expected matches for 'ben_abc'")
	}
	if matches[0].Value != "ben_abc123" {
		t.Errorf("expected first match to be ben_abc123, got %s", matches[0].Value)
	}

	// Test label match
	matches = FindSimilar("john", items, 3)
	if len(matches) == 0 {
		t.Error("expected matches for 'john'")
	}

	// Test no match
	matches = FindSimilar("zzz", items, 3)
	if len(matches) != 0 {
		t.Errorf("expected no matches for 'zzz', got %d", len(matches))
	}
}

func TestFormatSuggestions(t *testing.T) {
	matches := []Match{
		{Value: "ben_abc123", Label: "John Smith"},
	}

	output := FormatSuggestions(matches)
	if output == "" {
		t.Error("expected formatted output")
	}
	if !strings.Contains(output, "Did you mean") {
		t.Error("expected 'Did you mean' in output")
	}
}
