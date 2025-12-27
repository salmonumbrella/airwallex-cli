package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

func TestAccountsGetValidation(t *testing.T) {
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
			name:        "missing account ID",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
		{
			name:    "valid account ID",
			args:    []string{"acct_123"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags = rootFlags{}

			accountsCmd := newAccountsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(accountsCmd)

			fullArgs := append([]string{"accounts", "get"}, tt.args...)
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
