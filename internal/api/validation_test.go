package api

import (
	"strings"
	"testing"
)

func TestValidateResourceID(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		resourceType string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid alphanumeric ID",
			id:           "acc_123456",
			resourceType: "account",
			wantErr:      false,
		},
		{
			name:         "valid with underscores",
			id:           "card_holder_12345",
			resourceType: "cardholder",
			wantErr:      false,
		},
		{
			name:         "valid with dashes",
			id:           "txn-2024-01-15-abc",
			resourceType: "transaction",
			wantErr:      false,
		},
		{
			name:         "valid mixed case",
			id:           "RePoRt_ABC123xyz",
			resourceType: "report",
			wantErr:      false,
		},
		{
			name:         "empty ID",
			id:           "",
			resourceType: "account",
			wantErr:      true,
			errContains:  "cannot be empty",
		},
		{
			name:         "ID too long",
			id:           strings.Repeat("a", 129),
			resourceType: "transfer",
			wantErr:      true,
			errContains:  "too long",
		},
		{
			name:         "invalid characters - space",
			id:           "acc 123",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "invalid characters - slash",
			id:           "acc/123",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "invalid characters - backslash",
			id:           "acc\\123",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "invalid characters - dot",
			id:           "acc.123",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "invalid characters - special chars",
			id:           "acc@123!",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "path traversal attempt",
			id:           "../../../etc/passwd",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "SQL injection attempt",
			id:           "'; DROP TABLE accounts; --",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "single character valid",
			id:           "a",
			resourceType: "account",
			wantErr:      false,
		},
		{
			name:         "exactly 128 characters",
			id:           strings.Repeat("a", 128),
			resourceType: "account",
			wantErr:      false,
		},
		{
			name:         "URL encoding attempt",
			id:           "%2e%2e%2f",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "null byte injection",
			id:           "acc\x00123",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
		{
			name:         "newline injection",
			id:           "acc\n123",
			resourceType: "account",
			wantErr:      true,
			errContains:  "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceID(tt.id, tt.resourceType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateResourceID() expected error but got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateResourceID() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateResourceID() unexpected error = %v", err)
				}
			}
		})
	}
}
