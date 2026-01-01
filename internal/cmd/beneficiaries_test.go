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
		{
			name: "invalid IFSC format - too short",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "IN",
				"--first-name", "Raj",
				"--last-name", "Patel",
				"--account-name", "Raj Patel",
				"--account-currency", "INR",
				"--account-number", "1234567890",
				"--ifsc", "SBIN012345",
			},
			wantErr:     true,
			errContains: "--ifsc must be 11 characters",
		},
		{
			name: "invalid IFSC format - too long",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "IN",
				"--first-name", "Raj",
				"--last-name", "Patel",
				"--account-name", "Raj Patel",
				"--account-currency", "INR",
				"--account-number", "1234567890",
				"--ifsc", "SBIN00012345",
			},
			wantErr:     true,
			errContains: "--ifsc must be 11 characters",
		},
		{
			name: "invalid IFSC format - missing 0 in 5th position",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "IN",
				"--first-name", "Raj",
				"--last-name", "Patel",
				"--account-name", "Raj Patel",
				"--account-currency", "INR",
				"--account-number", "1234567890",
				"--ifsc", "SBIN1001234",
			},
			wantErr:     true,
			errContains: "--ifsc must be 11 characters",
		},
		{
			name: "invalid IFSC format - numbers in first 4 chars",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "IN",
				"--first-name", "Raj",
				"--last-name", "Patel",
				"--account-name", "Raj Patel",
				"--account-currency", "INR",
				"--account-number", "1234567890",
				"--ifsc", "SB1N0001234",
			},
			wantErr:     true,
			errContains: "--ifsc must be 11 characters",
		},
		{
			name: "valid IFSC format - lowercase input",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "IN",
				"--first-name", "Raj",
				"--last-name", "Patel",
				"--account-name", "Raj Patel",
				"--account-currency", "INR",
				"--account-number", "1234567890",
				"--ifsc", "sbin0001234",
			},
			wantErr: false,
		},
		// Japan Zengin validation
		{
			name: "zengin-bank-code invalid - not 4 digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "JP",
				"--first-name", "Taro",
				"--last-name", "Yamada",
				"--account-name", "Yamada Taro",
				"--account-currency", "JPY",
				"--account-number", "1234567",
				"--zengin-bank-code", "123",
				"--zengin-branch-code", "001",
			},
			wantErr:     true,
			errContains: "--zengin-bank-code must be exactly 4 digits",
		},
		{
			name: "zengin-branch-code invalid - not 3 digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "JP",
				"--first-name", "Taro",
				"--last-name", "Yamada",
				"--account-name", "Yamada Taro",
				"--account-currency", "JPY",
				"--account-number", "1234567",
				"--zengin-bank-code", "0001",
				"--zengin-branch-code", "01",
			},
			wantErr:     true,
			errContains: "--zengin-branch-code must be exactly 3 digits",
		},
		{
			name: "zengin requires both bank and branch codes",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "JP",
				"--first-name", "Taro",
				"--last-name", "Yamada",
				"--account-name", "Yamada Taro",
				"--account-currency", "JPY",
				"--account-number", "1234567",
				"--zengin-bank-code", "0001",
			},
			wantErr:     true,
			errContains: "--zengin-branch-code is required when --zengin-bank-code is provided",
		},
		{
			name: "valid Japan Zengin",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "JP",
				"--first-name", "Taro",
				"--last-name", "Yamada",
				"--account-name", "Yamada Taro",
				"--account-currency", "JPY",
				"--account-number", "1234567",
				"--zengin-bank-code", "0001",
				"--zengin-branch-code", "001",
			},
			wantErr: false,
		},
		// China CNAPS validation
		{
			name: "cnaps invalid - not 12 digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CN",
				"--first-name", "Wei",
				"--last-name", "Zhang",
				"--account-name", "Zhang Wei",
				"--account-currency", "CNY",
				"--account-number", "12345678901234",
				"--cnaps", "12345678901",
			},
			wantErr:     true,
			errContains: "--cnaps must be exactly 12 digits",
		},
		{
			name: "valid China CNAPS",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "CN",
				"--first-name", "Wei",
				"--last-name", "Zhang",
				"--account-name", "Zhang Wei",
				"--account-currency", "CNY",
				"--account-number", "12345678901234",
				"--cnaps", "102100099996",
			},
			wantErr: false,
		},
		// Brazil CPF/CNPJ validation
		{
			name: "cpf invalid - not 11 digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "BR",
				"--first-name", "João",
				"--last-name", "Silva",
				"--account-name", "João Silva",
				"--account-currency", "BRL",
				"--account-number", "123456789",
				"--swift-code", "BRASBRRJ",
				"--cpf", "1234567890",
			},
			wantErr:     true,
			errContains: "--cpf must be exactly 11 digits",
		},
		{
			name: "cnpj invalid - not 14 digits",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "BR",
				"--company-name", "Empresa LTDA",
				"--account-name", "Empresa LTDA",
				"--account-currency", "BRL",
				"--account-number", "123456789",
				"--swift-code", "BRASBRRJ",
				"--cnpj", "1234567890123",
			},
			wantErr:     true,
			errContains: "--cnpj must be exactly 14 digits",
		},
		{
			name: "valid Brazil PERSONAL with CPF",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "BR",
				"--first-name", "João",
				"--last-name", "Silva",
				"--account-name", "João Silva",
				"--account-currency", "BRL",
				"--account-number", "123456789",
				"--swift-code", "BRASBRRJ",
				"--cpf", "12345678901",
			},
			wantErr: false,
		},
		{
			name: "valid Brazil COMPANY with CNPJ",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "BR",
				"--company-name", "Empresa LTDA",
				"--account-name", "Empresa LTDA",
				"--account-currency", "BRL",
				"--account-number", "123456789",
				"--swift-code", "BRASBRRJ",
				"--cnpj", "12345678901234",
			},
			wantErr: false,
		},
		{
			name: "valid Brazil with bank-branch",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "BR",
				"--first-name", "João",
				"--last-name", "Silva",
				"--account-name", "João Silva",
				"--account-currency", "BRL",
				"--account-number", "123456789",
				"--swift-code", "BRASBRRJ",
				"--cpf", "12345678901",
				"--bank-branch", "1234",
			},
			wantErr: false,
		},
		// Singapore NRIC validation
		{
			name: "nric invalid format",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "SG",
				"--first-name", "Wei",
				"--last-name", "Tan",
				"--account-name", "Tan Wei",
				"--account-currency", "SGD",
				"--nric", "12345678A",
			},
			wantErr:     true,
			errContains: "--nric must be 9 characters in format SnnnnnnnA",
		},
		{
			name: "valid Singapore PayNow with NRIC",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "SG",
				"--first-name", "Wei",
				"--last-name", "Tan",
				"--account-name", "Tan Wei",
				"--account-currency", "SGD",
				"--nric", "S1234567A",
			},
			wantErr: false,
		},
		{
			name: "uen invalid - wrong length",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "SG",
				"--company-name", "SG Corp Pte Ltd",
				"--account-name", "SG Corp Pte Ltd",
				"--account-currency", "SGD",
				"--uen", "1234567",
			},
			wantErr:     true,
			errContains: "--uen must be 8-13 characters",
		},
		{
			name: "valid Singapore PayNow with UEN",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "SG",
				"--company-name", "SG Corp Pte Ltd",
				"--account-name", "SG Corp Pte Ltd",
				"--account-currency", "SGD",
				"--uen", "196800306E",
			},
			wantErr: false,
		},
		// South Korea bank code validation
		{
			name: "korea-bank-code invalid - not 3 digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "KR",
				"--first-name", "Min",
				"--last-name", "Kim",
				"--account-name", "Kim Min",
				"--account-currency", "KRW",
				"--account-number", "1234567890123",
				"--korea-bank-code", "12",
			},
			wantErr:     true,
			errContains: "--korea-bank-code must be exactly 3 digits",
		},
		{
			name: "valid South Korea",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "KR",
				"--first-name", "Min",
				"--last-name", "Kim",
				"--account-name", "Kim Min",
				"--account-currency", "KRW",
				"--account-number", "1234567890123",
				"--korea-bank-code", "004",
			},
			wantErr: false,
		},
		// Sweden clearing number validation
		{
			name: "clearing-number invalid - wrong length",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "SE",
				"--first-name", "Erik",
				"--last-name", "Svensson",
				"--account-name", "Erik Svensson",
				"--account-currency", "SEK",
				"--account-number", "123456789012345",
				"--clearing-number", "123",
			},
			wantErr:     true,
			errContains: "--clearing-number must be 4-5 digits",
		},
		{
			name: "valid Sweden clearing number",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "SE",
				"--first-name", "Erik",
				"--last-name", "Svensson",
				"--account-name", "Erik Svensson",
				"--account-currency", "SEK",
				"--account-number", "123456789012345",
				"--clearing-number", "1234",
			},
			wantErr: false,
		},
		// Hong Kong FPS validation
		{
			name: "hk-bank-code invalid - not 3 digits",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "HK",
				"--first-name", "Wing",
				"--last-name", "Chan",
				"--account-name", "Chan Wing",
				"--account-currency", "HKD",
				"--account-number", "12345678901234",
				"--hk-bank-code", "12",
			},
			wantErr:     true,
			errContains: "--hk-bank-code must be exactly 3 digits",
		},
		{
			name: "fps-id invalid - wrong length",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "HK",
				"--first-name", "Wing",
				"--last-name", "Chan",
				"--account-name", "Chan Wing",
				"--account-currency", "HKD",
				"--fps-id", "123456",
			},
			wantErr:     true,
			errContains: "--fps-id must be 7-9 digits",
		},
		{
			name: "valid Hong Kong FPS with bank code",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "HK",
				"--first-name", "Wing",
				"--last-name", "Chan",
				"--account-name", "Chan Wing",
				"--account-currency", "HKD",
				"--account-number", "12345678901234",
				"--hk-bank-code", "004",
			},
			wantErr: false,
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
		// Australia PayID validation
		{
			name: "payid-phone valid format",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--payid-phone", "+61-412345678",
				"--account-name", "Test",
				"--account-currency", "AUD",
				"--first-name", "John",
				"--last-name", "Doe",
			},
			wantErr: false,
		},
		{
			name: "payid-phone invalid format missing plus",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--payid-phone", "61-412345678",
				"--account-name", "Test",
				"--account-currency", "AUD",
				"--first-name", "John",
				"--last-name", "Doe",
			},
			wantErr:     true,
			errContains: "--payid-phone must be in format +61-nnnnnnnnn",
		},
		{
			name: "payid-phone invalid format wrong country code",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--payid-phone", "+1-412345678",
				"--account-name", "Test",
				"--account-currency", "AUD",
				"--first-name", "John",
				"--last-name", "Doe",
			},
			wantErr:     true,
			errContains: "--payid-phone must be in format +61-nnnnnnnnn",
		},
		{
			name: "payid-email valid",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--payid-email", "test@example.com",
				"--account-name", "Test",
				"--account-currency", "AUD",
				"--first-name", "John",
				"--last-name", "Doe",
			},
			wantErr: false,
		},
		{
			name: "payid-email invalid format",
			args: []string{
				"--entity-type", "PERSONAL",
				"--bank-country", "AU",
				"--payid-email", "notanemail",
				"--account-name", "Test",
				"--account-currency", "AUD",
				"--first-name", "John",
				"--last-name", "Doe",
			},
			wantErr:     true,
			errContains: "--payid-email must be a valid email address",
		},
		{
			name: "payid-abn valid 11 digits",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "AU",
				"--payid-abn", "12345678901",
				"--account-name", "Test Corp",
				"--account-currency", "AUD",
				"--company-name", "Test Corp",
			},
			wantErr: false,
		},
		{
			name: "payid-abn valid 9 digits",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "AU",
				"--payid-abn", "123456789",
				"--account-name", "Test Corp",
				"--account-currency", "AUD",
				"--company-name", "Test Corp",
			},
			wantErr: false,
		},
		{
			name: "payid-abn invalid length",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "AU",
				"--payid-abn", "1234567",
				"--account-name", "Test Corp",
				"--account-currency", "AUD",
				"--company-name", "Test Corp",
			},
			wantErr:     true,
			errContains: "--payid-abn must be 9 or 11 digits",
		},
		{
			name: "payid-abn invalid non-numeric",
			args: []string{
				"--entity-type", "COMPANY",
				"--bank-country", "AU",
				"--payid-abn", "1234567890A",
				"--account-name", "Test Corp",
				"--account-currency", "AUD",
				"--company-name", "Test Corp",
			},
			wantErr:     true,
			errContains: "--payid-abn must be 9 or 11 digits",
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
