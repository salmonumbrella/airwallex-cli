package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

func TestCardsCreateValidation(t *testing.T) {
	t.Setenv("AWX_ACCOUNT", "test-account")

	originalOpenSecretsStore := openSecretsStore
	defer func() { openSecretsStore = originalOpenSecretsStore }()
	openSecretsStore = func() (secrets.Store, error) {
		return &mockStore{}, nil
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "additional-cardholders without company flag",
			args: []string{
				"MyCard",
				"--cardholder-id", "chld_123",
				"--limit", "100",
				"--additional-cardholders", "chld_456",
			},
			wantErr:     true,
			errContains: "--additional-cardholders requires --company flag",
		},
		{
			name: "too many additional cardholders",
			args: []string{
				"MyCard",
				"--cardholder-id", "chld_123",
				"--limit", "100",
				"--company",
				"--additional-cardholders", "chld_1,chld_2,chld_3,chld_4",
			},
			wantErr:     true,
			errContains: "maximum 3 additional cardholders",
		},
		{
			name: "valid company card with additional cardholders",
			args: []string{
				"MyCard",
				"--cardholder-id", "chld_123",
				"--limit", "100",
				"--company",
				"--additional-cardholders", "chld_456,chld_789",
			},
			wantErr: false,
		},
		{
			name: "valid employee card",
			args: []string{
				"MyCard",
				"--cardholder-id", "chld_123",
				"--limit", "100",
			},
			wantErr: false,
		},
		{
			name: "missing required cardholder-id and limit",
			args: []string{
				"MyCard",
			},
			wantErr:     true,
			errContains: "required flag",
		},
		{
			name: "missing nickname argument",
			args: []string{
				"--cardholder-id", "chld_123",
				"--limit", "100",
			},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			issuingCmd := newIssuingCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(issuingCmd)

			fullArgs := append([]string{"issuing", "cards", "create"}, tt.args...)
			rootCmd.SetArgs(fullArgs)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				// For valid cases, expect API/auth errors (no real credentials)
				if err != nil && !isExpectedAPIError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func isExpectedAPIError(err error) bool {
	var contextual *api.ContextualError
	if errors.As(err, &contextual) {
		return true
	}
	var apiErr *api.APIError
	return errors.As(err, &apiErr)
}
