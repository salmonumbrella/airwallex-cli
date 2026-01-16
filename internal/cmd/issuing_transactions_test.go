package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

func TestTransactionsListValidation(t *testing.T) {
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
			name:        "invalid from date format",
			args:        []string{"--from", "01-01-2024"},
			wantErr:     true,
			errContains: "invalid --from date",
		},
		{
			name:        "invalid to date format",
			args:        []string{"--to", "2024/12/31"},
			wantErr:     true,
			errContains: "invalid --to date",
		},
		{
			name:    "valid date filters",
			args:    []string{"--from", "2024-01-01", "--to", "2024-01-31"},
			wantErr: false,
		},
		{
			name:    "valid card-id filter",
			args:    []string{"--card-id", "card_123"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			issuingCmd := newIssuingCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(issuingCmd)

			fullArgs := append([]string{"issuing", "transactions", "list"}, tt.args...)
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
