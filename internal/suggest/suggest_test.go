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
	if !strings.Contains(output, "•") {
		t.Error("expected bullet point in output")
	}
}

func TestFormatSuggestionsWithHelp(t *testing.T) {
	tests := []struct {
		name     string
		matches  []Match
		helpCmd  string
		contains []string
	}{
		{
			name:    "empty matches",
			matches: nil,
			helpCmd: "airwallex help",
		},
		{
			name:     "with help command",
			matches:  []Match{{Value: "ben_abc123", Label: "John Smith"}},
			helpCmd:  "airwallex beneficiaries list",
			contains: []string{"Did you mean", "ben_abc123", "(John Smith)", "Run 'airwallex beneficiaries list'"},
		},
		{
			name:     "without help command",
			matches:  []Match{{Value: "ben_abc123", Label: "John Smith"}},
			helpCmd:  "",
			contains: []string{"Did you mean", "ben_abc123", "(John Smith)"},
		},
		{
			name:     "multiple matches",
			matches:  []Match{{Value: "ben_abc123"}, {Value: "ben_def456"}},
			helpCmd:  "airwallex help",
			contains: []string{"ben_abc123", "ben_def456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatSuggestionsWithHelp(tt.matches, tt.helpCmd)

			if len(tt.matches) == 0 {
				if output != "" {
					t.Errorf("expected empty output for empty matches, got %q", output)
				}
				return
			}

			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}
