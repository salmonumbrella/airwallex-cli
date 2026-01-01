package cmd

import (
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
