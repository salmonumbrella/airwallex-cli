package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSchemasBeneficiaryCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		// Required flags validation
		{
			name:        "missing bank-country",
			args:        []string{"--entity-type", "COMPANY"},
			wantErr:     true,
			errContains: `required flag(s) "bank-country" not set`,
		},
		{
			name:        "missing entity-type",
			args:        []string{"--bank-country", "US"},
			wantErr:     true,
			errContains: `required flag(s) "entity-type" not set`,
		},
		{
			name:        "missing both required flags",
			args:        []string{},
			wantErr:     true,
			errContains: `required flag(s)`,
		},
		// Valid minimal invocations (optional payment-method not provided)
		{
			name: "valid with bank-country and entity-type only",
			args: []string{
				"--bank-country", "US",
				"--entity-type", "COMPANY",
			},
			wantErr: false,
		},
		{
			name: "valid PERSONAL entity type",
			args: []string{
				"--bank-country", "CA",
				"--entity-type", "PERSONAL",
			},
			wantErr: false,
		},
		// Valid with optional payment-method
		{
			name: "valid with payment-method LOCAL",
			args: []string{
				"--bank-country", "US",
				"--entity-type", "COMPANY",
				"--payment-method", "LOCAL",
			},
			wantErr: false,
		},
		{
			name: "valid with payment-method SWIFT",
			args: []string{
				"--bank-country", "GB",
				"--entity-type", "COMPANY",
				"--payment-method", "SWIFT",
			},
			wantErr: false,
		},
		// Edge cases for country codes
		{
			name: "two-letter country code",
			args: []string{
				"--bank-country", "CN",
				"--entity-type", "COMPANY",
			},
			wantErr: false,
		},
		{
			name: "lowercase country code",
			args: []string{
				"--bank-country", "us",
				"--entity-type", "COMPANY",
			},
			wantErr: false,
		},
		// Edge cases for entity types
		{
			name: "lowercase entity type",
			args: []string{
				"--bank-country", "US",
				"--entity-type", "company",
			},
			wantErr: false,
		},
		{
			name: "mixed case entity type",
			args: []string{
				"--bank-country", "US",
				"--entity-type", "Company",
			},
			wantErr: false,
		},
		// Combined valid scenarios
		{
			name: "all parameters provided - COMPANY with LOCAL",
			args: []string{
				"--bank-country", "CA",
				"--entity-type", "COMPANY",
				"--payment-method", "LOCAL",
			},
			wantErr: false,
		},
		{
			name: "all parameters provided - PERSONAL with SWIFT",
			args: []string{
				"--bank-country", "FR",
				"--entity-type", "PERSONAL",
				"--payment-method", "SWIFT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the schemas parent command and add the beneficiary subcommand
			schemasCmd := newSchemasCmd()

			// Set up a root command
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(schemasCmd)

			// Prepend "schemas beneficiary" to the args
			fullArgs := append([]string{"schemas", "beneficiary"}, tt.args...)
			rootCmd.SetArgs(fullArgs)

			// Execute the command
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
				// For non-error cases, we expect to fail on the actual API call
				// since we don't have a mock client set up. This is acceptable
				// as we're only testing validation logic here.
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSchemasTransferCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		// Required flags validation
		{
			name:        "missing source-currency",
			args:        []string{"--dest-currency", "EUR"},
			wantErr:     true,
			errContains: `required flag(s) "source-currency" not set`,
		},
		{
			name:        "missing dest-currency",
			args:        []string{"--source-currency", "USD"},
			wantErr:     true,
			errContains: `required flag(s) "dest-currency" not set`,
		},
		{
			name:        "missing both required flags",
			args:        []string{},
			wantErr:     true,
			errContains: `required flag(s)`,
		},
		// Valid minimal invocations (optional payment-method not provided)
		{
			name: "valid with source and dest currency only",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "EUR",
			},
			wantErr: false,
		},
		{
			name: "valid with same currency",
			args: []string{
				"--source-currency", "CAD",
				"--dest-currency", "CAD",
			},
			wantErr: false,
		},
		// Valid with optional payment-method
		{
			name: "valid with payment-method LOCAL",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "USD",
				"--payment-method", "LOCAL",
			},
			wantErr: false,
		},
		{
			name: "valid with payment-method SWIFT",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "EUR",
				"--payment-method", "SWIFT",
			},
			wantErr: false,
		},
		// Edge cases for currency codes
		{
			name: "three-letter currency codes",
			args: []string{
				"--source-currency", "GBP",
				"--dest-currency", "JPY",
			},
			wantErr: false,
		},
		{
			name: "lowercase currency codes",
			args: []string{
				"--source-currency", "usd",
				"--dest-currency", "eur",
			},
			wantErr: false,
		},
		{
			name: "mixed case currency codes",
			args: []string{
				"--source-currency", "Usd",
				"--dest-currency", "Eur",
			},
			wantErr: false,
		},
		// Various currency pairs
		{
			name: "USD to CAD",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "CAD",
			},
			wantErr: false,
		},
		{
			name: "EUR to GBP",
			args: []string{
				"--source-currency", "EUR",
				"--dest-currency", "GBP",
			},
			wantErr: false,
		},
		{
			name: "AUD to NZD",
			args: []string{
				"--source-currency", "AUD",
				"--dest-currency", "NZD",
			},
			wantErr: false,
		},
		// Combined valid scenarios
		{
			name: "all parameters provided - cross currency with LOCAL",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "EUR",
				"--payment-method", "LOCAL",
			},
			wantErr: false,
		},
		{
			name: "all parameters provided - same currency with SWIFT",
			args: []string{
				"--source-currency", "GBP",
				"--dest-currency", "GBP",
				"--payment-method", "SWIFT",
			},
			wantErr: false,
		},
		// Edge cases for payment methods
		{
			name: "lowercase payment method",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "EUR",
				"--payment-method", "local",
			},
			wantErr: false,
		},
		{
			name: "mixed case payment method",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "EUR",
				"--payment-method", "Swift",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the schemas parent command and add the transfer subcommand
			schemasCmd := newSchemasCmd()

			// Set up a root command
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(schemasCmd)

			// Prepend "schemas transfer" to the args
			fullArgs := append([]string{"schemas", "transfer"}, tt.args...)
			rootCmd.SetArgs(fullArgs)

			// Execute the command
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
				// For non-error cases, we expect to fail on the actual API call
				// since we don't have a mock client set up. This is acceptable
				// as we're only testing validation logic here.
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSchemasBeneficiaryCommand_FlagCombinations(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "US COMPANY no payment method",
			args: []string{
				"--bank-country", "US",
				"--entity-type", "COMPANY",
			},
			wantErr: false,
		},
		{
			name: "CA PERSONAL with LOCAL",
			args: []string{
				"--bank-country", "CA",
				"--entity-type", "PERSONAL",
				"--payment-method", "LOCAL",
			},
			wantErr: false,
		},
		{
			name: "GB COMPANY with SWIFT",
			args: []string{
				"--bank-country", "GB",
				"--entity-type", "COMPANY",
				"--payment-method", "SWIFT",
			},
			wantErr: false,
		},
		{
			name: "AU PERSONAL with SWIFT",
			args: []string{
				"--bank-country", "AU",
				"--entity-type", "PERSONAL",
				"--payment-method", "SWIFT",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemasCmd := newSchemasCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(schemasCmd)
			fullArgs := append([]string{"schemas", "beneficiary"}, tt.args...)
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
				// For non-error cases, we expect to fail on the actual API call
				// since we don't have real credentials. This is acceptable.
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestSchemasTransferCommand_CurrencyPairs(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "USD to EUR no payment method",
			args: []string{
				"--source-currency", "USD",
				"--dest-currency", "EUR",
			},
			wantErr: false,
		},
		{
			name: "CAD to CAD with LOCAL",
			args: []string{
				"--source-currency", "CAD",
				"--dest-currency", "CAD",
				"--payment-method", "LOCAL",
			},
			wantErr: false,
		},
		{
			name: "GBP to USD with SWIFT",
			args: []string{
				"--source-currency", "GBP",
				"--dest-currency", "USD",
				"--payment-method", "SWIFT",
			},
			wantErr: false,
		},
		{
			name: "JPY to EUR with LOCAL",
			args: []string{
				"--source-currency", "JPY",
				"--dest-currency", "EUR",
				"--payment-method", "LOCAL",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemasCmd := newSchemasCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(schemasCmd)
			fullArgs := append([]string{"schemas", "transfer"}, tt.args...)
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
				// For non-error cases, we expect to fail on the actual API call
				// since we don't have real credentials. This is acceptable.
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}
