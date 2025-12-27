package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

func TestBalancesHistoryValidation(t *testing.T) {
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
			name: "invalid from date format",
			args: []string{
				"--from", "01-01-2024",
			},
			wantErr:     true,
			errContains: "--from:",
		},
		{
			name: "invalid to date format",
			args: []string{
				"--to", "2024/01/01",
			},
			wantErr:     true,
			errContains: "--to:",
		},
		{
			name: "date range exceeds 7 days",
			args: []string{
				"--from", "2024-01-01",
				"--to", "2024-01-15",
			},
			wantErr:     true,
			errContains: "date range exceeds 7 days",
		},
		{
			name: "valid 7-day range",
			args: []string{
				"--from", "2024-01-01",
				"--to", "2024-01-07",
			},
			wantErr: false,
		},
		{
			name: "valid single date filter",
			args: []string{
				"--from", "2024-01-01",
			},
			wantErr: false,
		},
		{
			name: "valid currency filter",
			args: []string{
				"--currency", "CAD",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags = rootFlags{}

			balancesCmd := newBalancesCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(balancesCmd)

			fullArgs := append([]string{"balances", "history"}, tt.args...)
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
				if err != nil && !isExpectedAPIError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}
