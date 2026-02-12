package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

func TestCardholdersCreateValidation(t *testing.T) {
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
			name:        "missing email",
			args:        []string{"--first-name", "John", "--last-name", "Doe"},
			wantErr:     true,
			errContains: "required flag",
		},
		{
			name:        "missing first-name",
			args:        []string{"--email", "john@example.com", "--last-name", "Doe"},
			wantErr:     true,
			errContains: "required flag",
		},
		{
			name:        "missing last-name",
			args:        []string{"--email", "john@example.com", "--first-name", "John"},
			wantErr:     true,
			errContains: "required flag",
		},
		{
			name: "valid cardholder creation",
			args: []string{
				"--email", "john@example.com",
				"--first-name", "John",
				"--last-name", "Doe",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issuingCmd := newIssuingCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(issuingCmd)

			fullArgs := append([]string{"issuing", "cardholders", "create"}, tt.args...)
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

func TestCardholdersUpdateValidation(t *testing.T) {
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
			name:        "no updates specified",
			args:        []string{"chld_123"},
			wantErr:     true,
			errContains: "no updates specified",
		},
		{
			name:        "missing cardholder ID",
			args:        []string{"--email", "new@example.com"},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
		{
			name:    "valid email update",
			args:    []string{"chld_123", "--email", "new@example.com"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issuingCmd := newIssuingCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(issuingCmd)

			fullArgs := append([]string{"issuing", "cardholders", "update"}, tt.args...)
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
