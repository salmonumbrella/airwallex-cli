package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestBeneficiariesCreateValidation(t *testing.T) {
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
			name: "missing account-name",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-currency", "CAD",
				"--email", "john@example.com",
			},
			wantErr:     true,
			errContains: "--account-name is required",
		},
		{
			name: "missing account-currency",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--email", "john@example.com",
			},
			wantErr:     true,
			errContains: "--account-currency is required",
		},
		// Entity type specific validation
		{
			name: "COMPANY missing company-name",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "CA",
				"--account-name", "ACME Corp",
				"--account-currency", "CAD",
				"--email", "finance@acme.com",
			},
			wantErr:     true,
			errContains: "--company-name is required when entity-type is COMPANY",
		},
		{
			name: "PERSONAL missing first-name",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "john@example.com",
			},
			wantErr:     true,
			errContains: "--first-name is required when entity-type is PERSONAL",
		},
		{
			name: "PERSONAL missing last-name",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "john@example.com",
			},
			wantErr:     true,
			errContains: "--last-name is required when entity-type is PERSONAL",
		},
		// Routing method validation
		{
			name: "missing routing method",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
			},
			wantErr:     true,
			errContains: "must provide at least one routing method",
		},
		// Canada EFT validation
		{
			name: "institution-number without transit-number",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "001",
			},
			wantErr:     true,
			errContains: "--transit-number is required when --institution-number is provided",
		},
		// Phone number format validation
		{
			name: "invalid phone format - missing +1",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--phone", "4165551234",
			},
			wantErr:     true,
			errContains: "--phone must match format +1-nnnnnnnnnn",
		},
		{
			name: "invalid phone format - missing dash",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--phone", "+14165551234",
			},
			wantErr:     true,
			errContains: "--phone must match format +1-nnnnnnnnnn",
		},
		{
			name: "invalid phone format - too few digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--phone", "+1-416555123",
			},
			wantErr:     true,
			errContains: "--phone must match format +1-nnnnnnnnnn",
		},
		{
			name: "invalid phone format - too many digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--phone", "+1-41655512345",
			},
			wantErr:     true,
			errContains: "--phone must match format +1-nnnnnnnnnn",
		},
		{
			name: "valid phone format",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--phone", "+1-4165551234",
			},
			wantErr: false,
		},
		// Institution number format validation
		{
			name: "invalid institution-number - too few digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "01",
				"--transit-number", "12345",
			},
			wantErr:     true,
			errContains: "--institution-number must be exactly 3 digits",
		},
		{
			name: "invalid institution-number - too many digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "0001",
				"--transit-number", "12345",
			},
			wantErr:     true,
			errContains: "--institution-number must be exactly 3 digits",
		},
		{
			name: "invalid institution-number - non-numeric",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "00A",
				"--transit-number", "12345",
			},
			wantErr:     true,
			errContains: "--institution-number must be exactly 3 digits",
		},
		// Transit number format validation
		{
			name: "invalid transit-number - too few digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "001",
				"--transit-number", "1234",
			},
			wantErr:     true,
			errContains: "--transit-number must be exactly 5 digits",
		},
		{
			name: "invalid transit-number - too many digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "001",
				"--transit-number", "123456",
			},
			wantErr:     true,
			errContains: "--transit-number must be exactly 5 digits",
		},
		{
			name: "invalid transit-number - non-numeric",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "001",
				"--transit-number", "1234A",
			},
			wantErr:     true,
			errContains: "--transit-number must be exactly 5 digits",
		},
		{
			name: "valid EFT routing",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--institution-number", "001",
				"--transit-number", "12345",
			},
			wantErr: false,
		},
		// Email format validation
		{
			name: "invalid email - missing @",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "johnexample.com",
			},
			wantErr:     true,
			errContains: "--email must be a valid email address",
		},
		{
			name: "invalid email - missing local part",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "@example.com",
			},
			wantErr:     true,
			errContains: "--email must be a valid email address",
		},
		{
			name: "invalid email - missing domain",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "john@",
			},
			wantErr:     true,
			errContains: "--email must be a valid email address",
		},
		{
			name: "valid email",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "john@example.com",
			},
			wantErr: false,
		},
		// Valid COMPANY entity
		{
			name: "valid COMPANY with email",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "CA",
				"--company-name", "ACME Corp",
				"--account-name", "ACME Corp",
				"--account-currency", "CAD",
				"--email", "finance@acme.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the beneficiaries parent command and add the create subcommand
			benefCmd := newBeneficiariesCmd()

			// Set up a root command
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(benefCmd)

			// Prepend "beneficiaries create" to the args
			fullArgs := append([]string{"beneficiaries", "create"}, tt.args...)
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

func TestBeneficiariesCreateValidationCombinations(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "multiple routing methods provided - email and phone",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "john@example.com",
				"--phone", "+1-4165551234",
			},
			wantErr: false,
		},
		{
			name: "multiple routing methods - email and EFT",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "john@example.com",
				"--institution-number", "001",
				"--transit-number", "12345",
			},
			wantErr: false,
		},
		{
			name: "all routing methods provided",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CA",
				"--first-name", "John",
				"--last-name", "Doe",
				"--account-name", "John Doe",
				"--account-currency", "CAD",
				"--email", "john@example.com",
				"--phone", "+1-4165551234",
				"--institution-number", "001",
				"--transit-number", "12345",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			benefCmd := newBeneficiariesCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(benefCmd)
			fullArgs := append([]string{"beneficiaries", "create"}, tt.args...)
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
