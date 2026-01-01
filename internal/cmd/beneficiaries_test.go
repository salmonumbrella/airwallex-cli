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

func TestBeneficiariesCreate_InternationalRouting(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		// Valid international routing configurations
		{
			name: "US with routing number",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "US",
				"--company-name", "Test Corp",
				"--account-name", "Test Corp",
				"--account-currency", "USD",
				"--account-number", "123456789",
				"--routing-number", "021000021",
			},
			wantErr: false,
		},
		{
			name: "UK with sort code",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "GB",
				"--company-name", "UK Ltd",
				"--account-name", "UK Ltd",
				"--account-currency", "GBP",
				"--account-number", "12345678",
				"--sort-code", "123456",
			},
			wantErr: false,
		},
		{
			name: "Australia with BSB",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--first-name", "John",
				"--last-name", "Smith",
				"--account-name", "John Smith",
				"--account-currency", "AUD",
				"--account-number", "123456789",
				"--bsb", "062000",
			},
			wantErr: false,
		},
		{
			name: "Europe with IBAN and SWIFT",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "DE",
				"--company-name", "GmbH",
				"--account-name", "GmbH",
				"--account-currency", "EUR",
				"--iban", "DE89370400440532013000",
				"--swift-code", "COBADEFFXXX",
				"--transfer-method", "SWIFT",
			},
			wantErr: false,
		},
		{
			name: "India with IFSC",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "IN",
				"--first-name", "Raj",
				"--last-name", "Patel",
				"--account-name", "Raj Patel",
				"--account-currency", "INR",
				"--account-number", "1234567890",
				"--ifsc", "SBIN0001234",
			},
			wantErr: false,
		},
		{
			name: "Mexico with CLABE",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "MX",
				"--company-name", "Mexico SA",
				"--account-name", "Mexico SA",
				"--account-currency", "MXN",
				"--clabe", "123456789012345678",
			},
			wantErr: false,
		},
		{
			name: "SWIFT only without account number",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "JP",
				"--company-name", "Japan Corp",
				"--account-name", "Japan Corp",
				"--account-currency", "JPY",
				"--swift-code", "MABORJPJXXX",
				"--transfer-method", "SWIFT",
			},
			wantErr: false,
		},
		// Invalid routing configurations
		{
			name: "no routing method fails",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "US",
				"--company-name", "Test",
				"--account-name", "Test",
				"--account-currency", "USD",
				"--account-number", "123",
			},
			wantErr:     true,
			errContains: "must provide at least one routing method",
		},
		{
			name: "invalid routing number format - too short",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "US",
				"--company-name", "Test Corp",
				"--account-name", "Test Corp",
				"--account-currency", "USD",
				"--account-number", "123456789",
				"--routing-number", "12345",
			},
			wantErr:     true,
			errContains: "--routing-number must be exactly 9 digits",
		},
		{
			name: "invalid routing number format - too long",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "US",
				"--company-name", "Test Corp",
				"--account-name", "Test Corp",
				"--account-currency", "USD",
				"--account-number", "123456789",
				"--routing-number", "0210000210",
			},
			wantErr:     true,
			errContains: "--routing-number must be exactly 9 digits",
		},
		{
			name: "invalid routing number format - non-numeric",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "US",
				"--company-name", "Test Corp",
				"--account-name", "Test Corp",
				"--account-currency", "USD",
				"--account-number", "123456789",
				"--routing-number", "02100002X",
			},
			wantErr:     true,
			errContains: "--routing-number must be exactly 9 digits",
		},
		{
			name: "invalid sort code format - too short",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "GB",
				"--company-name", "UK Ltd",
				"--account-name", "UK Ltd",
				"--account-currency", "GBP",
				"--account-number", "12345678",
				"--sort-code", "12345",
			},
			wantErr:     true,
			errContains: "--sort-code must be exactly 6 digits",
		},
		{
			name: "invalid sort code format - too long",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "GB",
				"--company-name", "UK Ltd",
				"--account-name", "UK Ltd",
				"--account-currency", "GBP",
				"--account-number", "12345678",
				"--sort-code", "1234567",
			},
			wantErr:     true,
			errContains: "--sort-code must be exactly 6 digits",
		},
		{
			name: "invalid sort code format - non-numeric",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "GB",
				"--company-name", "UK Ltd",
				"--account-name", "UK Ltd",
				"--account-currency", "GBP",
				"--account-number", "12345678",
				"--sort-code", "12345A",
			},
			wantErr:     true,
			errContains: "--sort-code must be exactly 6 digits",
		},
		{
			name: "invalid BSB format - too short",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--first-name", "John",
				"--last-name", "Smith",
				"--account-name", "John Smith",
				"--account-currency", "AUD",
				"--account-number", "123456789",
				"--bsb", "12345",
			},
			wantErr:     true,
			errContains: "--bsb must be exactly 6 digits",
		},
		{
			name: "invalid BSB format - too long",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--first-name", "John",
				"--last-name", "Smith",
				"--account-name", "John Smith",
				"--account-currency", "AUD",
				"--account-number", "123456789",
				"--bsb", "0620001",
			},
			wantErr:     true,
			errContains: "--bsb must be exactly 6 digits",
		},
		{
			name: "invalid BSB format - non-numeric",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--first-name", "John",
				"--last-name", "Smith",
				"--account-name", "John Smith",
				"--account-currency", "AUD",
				"--account-number", "123456789",
				"--bsb", "06200A",
			},
			wantErr:     true,
			errContains: "--bsb must be exactly 6 digits",
		},
		{
			name: "invalid CLABE format - too short",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "MX",
				"--company-name", "Mexico SA",
				"--account-name", "Mexico SA",
				"--account-currency", "MXN",
				"--clabe", "12345678901234567",
			},
			wantErr:     true,
			errContains: "--clabe must be exactly 18 digits",
		},
		{
			name: "invalid CLABE format - too long",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "MX",
				"--company-name", "Mexico SA",
				"--account-name", "Mexico SA",
				"--account-currency", "MXN",
				"--clabe", "1234567890123456789",
			},
			wantErr:     true,
			errContains: "--clabe must be exactly 18 digits",
		},
		{
			name: "invalid CLABE format - non-numeric",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "MX",
				"--company-name", "Mexico SA",
				"--account-name", "Mexico SA",
				"--account-currency", "MXN",
				"--clabe", "12345678901234567A",
			},
			wantErr:     true,
			errContains: "--clabe must be exactly 18 digits",
		},
		// Multiple routing methods (should be valid)
		{
			name: "multiple routing methods - IBAN and SWIFT",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "FR",
				"--company-name", "French SA",
				"--account-name", "French SA",
				"--account-currency", "EUR",
				"--iban", "FR7630006000011234567890189",
				"--swift-code", "BNPAFRPPXXX",
			},
			wantErr: false,
		},
		{
			name: "multiple routing methods - routing number and SWIFT",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "US",
				"--company-name", "US Corp",
				"--account-name", "US Corp",
				"--account-currency", "USD",
				"--account-number", "123456789",
				"--routing-number", "021000021",
				"--swift-code", "CHASUS33XXX",
				"--transfer-method", "SWIFT",
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
