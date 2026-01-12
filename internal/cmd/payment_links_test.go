package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPaymentLinksListCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "with limit",
			args:    []string{"--page-size", "50"},
			wantErr: false,
		},
		{
			name:    "with small limit",
			args:    []string{"--page-size", "5"},
			wantErr: false, // Should be adjusted to minimum 10
		},
		{
			name:    "with minimum limit",
			args:    []string{"--page-size", "10"},
			wantErr: false,
		},
		{
			name:    "with large limit",
			args:    []string{"--page-size", "100"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentLinksCmd := newPaymentLinksCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(paymentLinksCmd)

			fullArgs := append([]string{"payment-links", "list"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestPaymentLinksListCommand_PageSizeMinimum(t *testing.T) {
	cmd := newPaymentLinksListCmd()

	// Verify the help text mentions minimum
	pageSizeFlag := cmd.Flags().Lookup("page-size")
	if pageSizeFlag == nil {
		t.Fatal("page-size flag not found")
	}

	if pageSizeFlag.Deprecated == "" {
		t.Errorf("expected page-size flag to be deprecated")
	}

	// Verify default value
	defaultVal := pageSizeFlag.DefValue
	if defaultVal != "0" {
		t.Errorf("expected default value of 0 for deprecated page-size, got: %s", defaultVal)
	}
}

func TestPaymentLinksGetCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no payment link ID",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg(s)",
		},
		{
			name:        "too many args",
			args:        []string{"pl_123", "pl_456"},
			wantErr:     true,
			errContains: "accepts 1 arg(s)",
		},
		{
			name:    "valid payment link ID",
			args:    []string{"pl_123"},
			wantErr: false,
		},
		{
			name:    "valid payment link ID with prefix",
			args:    []string{"pl_abcd1234efgh5678"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentLinksCmd := newPaymentLinksCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(paymentLinksCmd)

			fullArgs := append([]string{"payment-links", "get"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestPaymentLinksCreateCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing all required flags",
			args:        []string{},
			wantErr:     true,
			errContains: "required flag(s)",
		},
		{
			name: "missing amount",
			args: []string{
				"--currency", "USD",
			},
			wantErr:     true,
			errContains: "required flag(s)",
		},
		{
			name: "missing currency",
			args: []string{
				"--amount", "100",
			},
			wantErr:     true,
			errContains: "required flag(s)",
		},
		{
			name: "valid minimum required flags",
			args: []string{
				"--amount", "100",
				"--currency", "USD",
			},
			wantErr: false,
		},
		{
			name: "with description",
			args: []string{
				"--amount", "50",
				"--currency", "EUR",
				"--description", "Invoice #123",
			},
			wantErr: false,
		},
		{
			name: "with expires-in days",
			args: []string{
				"--amount", "75",
				"--currency", "GBP",
				"--expires-in", "7d",
			},
			wantErr: false,
		},
		{
			name: "with expires-in hours",
			args: []string{
				"--amount", "200",
				"--currency", "CAD",
				"--expires-in", "24h",
			},
			wantErr: false,
		},
		{
			name: "with all optional flags",
			args: []string{
				"--amount", "150.50",
				"--currency", "AUD",
				"--description", "Payment for services",
				"--expires-in", "3d",
			},
			wantErr: false,
		},
		{
			name: "with decimal amount",
			args: []string{
				"--amount", "99.99",
				"--currency", "USD",
			},
			wantErr: false,
		},
		{
			name: "with long description",
			args: []string{
				"--amount", "100",
				"--currency", "USD",
				"--description", "This is a very long description that should still be valid for a payment link creation",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentLinksCmd := newPaymentLinksCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(paymentLinksCmd)

			fullArgs := append([]string{"payment-links", "create"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestPaymentLinksCreateCommand_AmountValidation(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		amount      string
		currency    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "zero amount",
			amount:      "0",
			currency:    "USD",
			wantErr:     true,
			errContains: "amount must be positive",
		},
		{
			name:        "negative amount",
			amount:      "-100",
			currency:    "USD",
			wantErr:     true,
			errContains: "amount must be positive",
		},
		{
			name:     "very large amount",
			amount:   "999999.99",
			currency: "USD",
			wantErr:  false,
		},
		{
			name:     "amount with many decimals",
			amount:   "100.12345",
			currency: "USD",
			wantErr:  false,
		},
		{
			name:     "small decimal amount",
			amount:   "0.01",
			currency: "USD",
			wantErr:  false,
		},
		{
			name:     "standard amount",
			amount:   "100.00",
			currency: "USD",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentLinksCmd := newPaymentLinksCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(paymentLinksCmd)

			fullArgs := []string{"payment-links", "create", "--amount", tt.amount, "--currency", tt.currency}
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestPaymentLinksCreateCommand_RequiredFlags(t *testing.T) {
	cmd := newPaymentLinksCreateCmd()

	// Verify amount flag is marked as required
	amountFlag := cmd.Flags().Lookup("amount")
	if amountFlag == nil {
		t.Fatal("amount flag not found")
	}

	// Verify currency flag is marked as required
	currencyFlag := cmd.Flags().Lookup("currency")
	if currencyFlag == nil {
		t.Fatal("currency flag not found")
	}

	// Verify the flags have the expected types
	if amountFlag.Value.Type() != "float64" {
		t.Errorf("amount flag should be float64, got %s", amountFlag.Value.Type())
	}

	if currencyFlag.Value.Type() != "string" {
		t.Errorf("currency flag should be string, got %s", currencyFlag.Value.Type())
	}

	// Test that command fails without required flags
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	paymentLinksCmd := newPaymentLinksCmd()
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(paymentLinksCmd)

	// Try to run without any flags
	rootCmd.SetArgs([]string{"payment-links", "create"})
	err := rootCmd.Execute()

	if err == nil {
		t.Error("expected error when running create without required flags")
	} else if !strings.Contains(err.Error(), "required flag") {
		t.Errorf("expected 'required flag' error, got: %v", err)
	}
}

func TestPaymentLinksCreateCommand_HelpText(t *testing.T) {
	cmd := newPaymentLinksCreateCmd()

	// Verify help text contains examples
	if !strings.Contains(cmd.Long, "Examples:") {
		t.Error("create command help should contain Examples section")
	}

	// Verify examples show both currencies
	if !strings.Contains(cmd.Long, "USD") {
		t.Error("examples should show USD currency")
	}

	if !strings.Contains(cmd.Long, "EUR") {
		t.Error("examples should show EUR currency")
	}

	// Verify examples show optional parameters
	if !strings.Contains(cmd.Long, "--description") {
		t.Error("examples should demonstrate --description flag")
	}

	if !strings.Contains(cmd.Long, "--expires-in") {
		t.Error("examples should demonstrate --expires-in flag")
	}
}

func TestPaymentLinksCommand_Aliases(t *testing.T) {
	cmd := newPaymentLinksCmd()

	// Verify the command has the "pl" alias
	aliases := cmd.Aliases
	if len(aliases) == 0 {
		t.Error("payment-links command should have aliases")
		return
	}

	hasPlAlias := false
	for _, alias := range aliases {
		if alias == "pl" {
			hasPlAlias = true
			break
		}
	}

	if !hasPlAlias {
		t.Errorf("payment-links command should have 'pl' alias, got: %v", aliases)
	}
}

func TestPaymentLinksCommand_Subcommands(t *testing.T) {
	cmd := newPaymentLinksCmd()

	// Verify all expected subcommands are present
	expectedSubcommands := []string{"list", "get", "create"}
	subcommands := cmd.Commands()

	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, subcmd := range subcommands {
			// Check if the command Use starts with the expected name
			// (e.g., "get <linkId>" starts with "get")
			if strings.HasPrefix(subcmd.Use, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}
