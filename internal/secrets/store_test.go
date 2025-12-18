package secrets

import (
	"testing"
)

func TestCredentialKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"my-company", "account:my-company"},
		{"test-account", "account:test-account"},
	}
	for _, tt := range tests {
		got := credentialKey(tt.name)
		if got != tt.want {
			t.Errorf("credentialKey(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  Test  ", "test"},
		{"UPPER", "upper"},
		{"already-lower", "already-lower"},
	}
	for _, tt := range tests {
		got := normalize(tt.input)
		if got != tt.want {
			t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseCredentialKey(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		wantOK bool
	}{
		{"account:my-company", "my-company", true},
		{"account:test-account", "test-account", true},
		{"invalid", "", false},
		{"account:", "", false},
		{"account:  ", "", false},
	}
	for _, tt := range tests {
		got, ok := ParseCredentialKey(tt.input)
		if ok != tt.wantOK {
			t.Errorf("ParseCredentialKey(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
		}
		if got != tt.want {
			t.Errorf("ParseCredentialKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
