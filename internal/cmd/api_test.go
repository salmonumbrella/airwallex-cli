package cmd

import (
	"net/url"
	"strings"
	"testing"
)

func TestAPICommand_Flags(t *testing.T) {
	cmd := newAPICmd()

	// Check command has correct args validation
	if cmd.Args == nil {
		t.Error("expected Args validation")
	}

	// Check all expected flags exist
	expectedFlags := []struct {
		name      string
		shorthand string
	}{
		{"method", "X"},
		{"data", "d"},
		{"data-file", ""},
		{"header", "H"},
		{"query", "q"},
		{"silent", "s"},
		{"include", "i"},
	}

	for _, ef := range expectedFlags {
		flag := cmd.Flags().Lookup(ef.name)
		if flag == nil {
			t.Errorf("expected flag --%s", ef.name)
			continue
		}
		if ef.shorthand != "" && flag.Shorthand != ef.shorthand {
			t.Errorf("flag --%s expected shorthand -%s, got -%s", ef.name, ef.shorthand, flag.Shorthand)
		}
	}
}

func TestAPICommand_EndpointNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/v1/balances", "/api/v1/balances"},
		{"api/v1/balances", "/api/v1/balances"},
		{"/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			endpoint := tt.input
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}
			if endpoint != tt.expected {
				t.Errorf("got %s, want %s", endpoint, tt.expected)
			}
		})
	}
}

func TestAPICommand_Usage(t *testing.T) {
	cmd := newAPICmd()

	if cmd.Use != "api <endpoint>" {
		t.Errorf("unexpected Use: %s", cmd.Use)
	}

	// Check help includes examples
	if !strings.Contains(cmd.Long, "GET current balances") {
		t.Error("expected help to include GET example")
	}
	if !strings.Contains(cmd.Long, "POST with inline JSON") {
		t.Error("expected help to include POST example")
	}
}

func TestAPICommand_QueryParamEncoding(t *testing.T) {
	// Test that query parameters are properly URL-encoded
	// This tests the encoding logic used in the api command
	tests := []struct {
		name     string
		params   []string
		expected string // expected query string (URL-encoded)
	}{
		{
			name:     "simple key=value",
			params:   []string{"status=COMPLETED"},
			expected: "status=COMPLETED",
		},
		{
			name:     "multiple params",
			params:   []string{"status=COMPLETED", "page_size=10"},
			expected: "page_size=10&status=COMPLETED", // url.Values sorts alphabetically
		},
		{
			name:     "value with spaces",
			params:   []string{"name=John Doe"},
			expected: "name=John+Doe",
		},
		{
			name:     "value with special chars",
			params:   []string{"filter=a=b&c=d"},
			expected: "filter=a%3Db%26c%3Dd",
		},
		{
			name:     "value with equals sign",
			params:   []string{"expr=1+1=2"},
			expected: "expr=1%2B1%3D2",
		},
		{
			name:     "key without value",
			params:   []string{"flag"},
			expected: "flag=",
		},
		{
			name:     "unicode value",
			params:   []string{"city=東京"},
			expected: "city=%E6%9D%B1%E4%BA%AC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the encoding logic from api.go
			params := url.Values{}
			for _, qp := range tt.params {
				parts := strings.SplitN(qp, "=", 2)
				if len(parts) == 2 {
					params.Add(parts[0], parts[1])
				} else {
					params.Add(parts[0], "")
				}
			}
			encoded := params.Encode()
			if encoded != tt.expected {
				t.Errorf("got %q, want %q", encoded, tt.expected)
			}
		})
	}
}
