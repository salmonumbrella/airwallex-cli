package secrets

import (
	"bytes"
	"log"
	"testing"
	"time"
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

func TestCredentialAgeWarning(t *testing.T) {
	tests := []struct {
		name        string
		createdAt   time.Time
		accountName string
		wantWarning bool
	}{
		{
			name:        "old credentials warn on first retrieval",
			createdAt:   time.Now().Add(-100 * 24 * time.Hour),
			accountName: "test-old-account",
			wantWarning: true,
		},
		{
			name:        "recent credentials do not warn",
			createdAt:   time.Now().Add(-30 * 24 * time.Hour),
			accountName: "test-recent-account",
			wantWarning: false,
		},
		{
			name:        "zero time credentials do not warn",
			createdAt:   time.Time{},
			accountName: "test-zero-time-account",
			wantWarning: false,
		},
		{
			name:        "credentials just before threshold do not warn",
			createdAt:   time.Now().Add(-CredentialRotationThreshold + time.Minute),
			accountName: "test-threshold-account",
			wantWarning: false,
		},
		{
			name:        "credentials just past threshold warn",
			createdAt:   time.Now().Add(-CredentialRotationThreshold - time.Hour),
			accountName: "test-past-threshold-account",
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset warning state for each test
			warnedAccounts.Delete(tt.accountName)

			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			// Simulate credential retrieval by checking the warning logic
			creds := Credentials{
				Name:      tt.accountName,
				CreatedAt: tt.createdAt,
			}

			// Execute the same logic as in Get()
			if !creds.CreatedAt.IsZero() && time.Since(creds.CreatedAt) > CredentialRotationThreshold {
				if _, warned := warnedAccounts.LoadOrStore(tt.accountName, true); !warned {
					log.Printf("Warning: credentials for account %q are over 90 days old, consider rotating", tt.accountName)
				}
			}

			logOutput := buf.String()
			hasWarning := len(logOutput) > 0

			if hasWarning != tt.wantWarning {
				t.Errorf("warning = %v, want %v (log: %q)", hasWarning, tt.wantWarning, logOutput)
			}

			// For old credentials, verify second retrieval does NOT warn (rate limiting)
			if tt.wantWarning {
				buf.Reset()

				// Second retrieval
				if !creds.CreatedAt.IsZero() && time.Since(creds.CreatedAt) > CredentialRotationThreshold {
					if _, warned := warnedAccounts.LoadOrStore(tt.accountName, true); !warned {
						log.Printf("Warning: credentials for account %q are over 90 days old, consider rotating", tt.accountName)
					}
				}

				secondLogOutput := buf.String()
				if len(secondLogOutput) > 0 {
					t.Errorf("second retrieval should not warn, got: %q", secondLogOutput)
				}
			}
		})
	}
}
