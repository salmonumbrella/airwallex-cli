package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestLinkedAccountsListCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "default pagination",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "custom limit above minimum",
			args:    []string{"--page-size", "50"},
			wantErr: false,
		},
		{
			name:    "limit below minimum (should be adjusted to 10)",
			args:    []string{"--page-size", "5"},
			wantErr: false,
		},
		{
			name:    "limit at minimum",
			args:    []string{"--page-size", "10"},
			wantErr: false,
		},
		{
			name:        "invalid limit (non-numeric)",
			args:        []string{"--page-size", "abc"},
			wantErr:     true,
			errContains: "invalid argument",
		},
		{
			name:    "limit with negative value (accepted by cobra)",
			args:    []string{"--page-size", "-10"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			laCmd := newLinkedAccountsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(laCmd)

			fullArgs := append([]string{"linked-accounts", "list"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestLinkedAccountsGetCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

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
			args:    []string{"la_123456"},
			wantErr: false,
		},
		{
			name:        "too many arguments",
			args:        []string{"la_123456", "la_789012"},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			laCmd := newLinkedAccountsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(laCmd)

			fullArgs := append([]string{"linked-accounts", "get"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestLinkedAccountsCreateCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		// Required fields validation
		{
			name: "missing type",
			args: []string{
				"--account-name", "My Account",
				"--currency", "AUD",
				"--account-number", "12345678",
			},
			wantErr:     true,
			errContains: `required flag(s) "type" not set`,
		},
		{
			name: "missing account-name",
			args: []string{
				"--type", "AU_BANK",
				"--currency", "AUD",
				"--account-number", "12345678",
			},
			wantErr:     true,
			errContains: `required flag(s) "account-name" not set`,
		},
		{
			name: "missing currency",
			args: []string{
				"--type", "AU_BANK",
				"--account-name", "My Account",
				"--account-number", "12345678",
			},
			wantErr:     true,
			errContains: `required flag(s) "currency" not set`,
		},
		{
			name: "missing account-number",
			args: []string{
				"--type", "AU_BANK",
				"--account-name", "My Account",
				"--currency", "AUD",
			},
			wantErr:     true,
			errContains: `required flag(s) "account-number" not set`,
		},
		// Valid configurations for different account types
		{
			name: "valid AU_BANK with BSB",
			args: []string{
				"--type", "AU_BANK",
				"--account-name", "My Australian Account",
				"--currency", "AUD",
				"--account-number", "12345678",
				"--bsb", "062000",
			},
			wantErr: false,
		},
		{
			name: "valid AU_BANK without BSB",
			args: []string{
				"--type", "AU_BANK",
				"--account-name", "My Australian Account",
				"--currency", "AUD",
				"--account-number", "12345678",
			},
			wantErr: false,
		},
		{
			name: "valid US_BANK with routing number",
			args: []string{
				"--type", "US_BANK",
				"--account-name", "My US Account",
				"--currency", "USD",
				"--account-number", "12345678",
				"--routing-number", "021000021",
			},
			wantErr: false,
		},
		{
			name: "valid US_BANK without routing number",
			args: []string{
				"--type", "US_BANK",
				"--account-name", "My US Account",
				"--currency", "USD",
				"--account-number", "12345678",
			},
			wantErr: false,
		},
		{
			name: "valid CA_BANK",
			args: []string{
				"--type", "CA_BANK",
				"--account-name", "My Canadian Account",
				"--currency", "CAD",
				"--account-number", "12345678",
			},
			wantErr: false,
		},
		{
			name: "valid GB_BANK",
			args: []string{
				"--type", "GB_BANK",
				"--account-name", "My UK Account",
				"--currency", "GBP",
				"--account-number", "12345678",
			},
			wantErr: false,
		},
		{
			name: "valid SG_BANK",
			args: []string{
				"--type", "SG_BANK",
				"--account-name", "My Singapore Account",
				"--currency", "SGD",
				"--account-number", "12345678",
			},
			wantErr: false,
		},
		{
			name: "valid HK_BANK",
			args: []string{
				"--type", "HK_BANK",
				"--account-name", "My Hong Kong Account",
				"--currency", "HKD",
				"--account-number", "12345678",
			},
			wantErr: false,
		},
		// Test with both BSB and routing number (should be valid)
		{
			name: "with both BSB and routing number",
			args: []string{
				"--type", "AU_BANK",
				"--account-name", "My Account",
				"--currency", "AUD",
				"--account-number", "12345678",
				"--bsb", "062000",
				"--routing-number", "021000021",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			laCmd := newLinkedAccountsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(laCmd)

			fullArgs := append([]string{"linked-accounts", "create"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestLinkedAccountsDepositCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		// Argument validation
		{
			name:        "missing account ID",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
		{
			name: "too many arguments",
			args: []string{
				"la_123456",
				"la_789012",
			},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
		// Required flags validation
		{
			name: "missing amount",
			args: []string{
				"la_123456",
				"--currency", "AUD",
			},
			wantErr:     true,
			errContains: `required flag(s) "amount" not set`,
		},
		{
			name: "missing currency",
			args: []string{
				"la_123456",
				"--amount", "5000",
			},
			wantErr:     true,
			errContains: `required flag(s) "currency" not set`,
		},
		{
			name: "missing both amount and currency",
			args: []string{
				"la_123456",
			},
			wantErr:     true,
			errContains: `required flag(s) "amount"`,
		},
		// Amount validation
		{
			name: "zero amount",
			args: []string{
				"la_123456",
				"--amount", "0",
				"--currency", "AUD",
			},
			wantErr:     true,
			errContains: "amount must be positive",
		},
		{
			name: "negative amount",
			args: []string{
				"la_123456",
				"--amount", "-100",
				"--currency", "AUD",
			},
			wantErr:     true,
			errContains: "amount must be positive",
		},
		{
			name: "positive amount",
			args: []string{
				"la_123456",
				"--amount", "5000",
				"--currency", "AUD",
			},
			wantErr: false,
		},
		{
			name: "decimal amount",
			args: []string{
				"la_123456",
				"--amount", "5000.50",
				"--currency", "AUD",
			},
			wantErr: false,
		},
		{
			name: "very small amount",
			args: []string{
				"la_123456",
				"--amount", "0.01",
				"--currency", "USD",
			},
			wantErr: false,
		},
		{
			name: "large amount",
			args: []string{
				"la_123456",
				"--amount", "999999.99",
				"--currency", "USD",
			},
			wantErr: false,
		},
		{
			name: "invalid amount (non-numeric)",
			args: []string{
				"la_123456",
				"--amount", "abc",
				"--currency", "AUD",
			},
			wantErr:     true,
			errContains: "invalid argument",
		},
		// Valid deposits with different currencies
		{
			name: "valid deposit AUD",
			args: []string{
				"la_123456",
				"--amount", "5000",
				"--currency", "AUD",
			},
			wantErr: false,
		},
		{
			name: "valid deposit USD",
			args: []string{
				"la_123456",
				"--amount", "10000",
				"--currency", "USD",
			},
			wantErr: false,
		},
		{
			name: "valid deposit CAD",
			args: []string{
				"la_123456",
				"--amount", "7500.25",
				"--currency", "CAD",
			},
			wantErr: false,
		},
		{
			name: "valid deposit GBP",
			args: []string{
				"la_123456",
				"--amount", "3000.50",
				"--currency", "GBP",
			},
			wantErr: false,
		},
		{
			name: "valid deposit EUR",
			args: []string{
				"la_123456",
				"--amount", "8000",
				"--currency", "EUR",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			laCmd := newLinkedAccountsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(laCmd)

			fullArgs := append([]string{"linked-accounts", "deposit"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}
