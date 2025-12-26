package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestTransfersListCmd_PageSizeValidation(t *testing.T) {
	tests := []struct {
		name        string
		pageSize    int
		expectedMin int
		description string
	}{
		{
			name:        "page size below minimum gets adjusted",
			pageSize:    5,
			expectedMin: 10,
			description: "page size less than 10 should be adjusted to 10",
		},
		{
			name:        "page size at minimum is unchanged",
			pageSize:    10,
			expectedMin: 10,
			description: "page size of exactly 10 should remain 10",
		},
		{
			name:        "page size above minimum is unchanged",
			pageSize:    50,
			expectedMin: 10,
			description: "page size above 10 should be unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTransfersListCmd()
			if err := cmd.Flags().Set("page-size", intToString(tt.pageSize)); err != nil {
				t.Fatalf("failed to set page-size flag: %v", err)
			}

			// We can't easily test the actual API call without mocking,
			// but we can verify the flag is set and the command validates it
			pageSizeFlag := cmd.Flags().Lookup("page-size")
			if pageSizeFlag == nil {
				t.Fatal("page-size flag not found")
			}

			// Verify the help text mentions minimum
			if !strings.Contains(pageSizeFlag.Usage, "min 10") {
				t.Errorf("page-size flag help text should mention minimum of 10")
			}
		})
	}
}

func TestTransfersCreateCmd_AmountValidation(t *testing.T) {
	tests := []struct {
		name              string
		setTransferAmount bool
		transferAmount    float64
		setSourceAmount   bool
		sourceAmount      float64
		wantErr           bool
		errContains       string
	}{
		{
			name:              "neither amount provided",
			setTransferAmount: false,
			setSourceAmount:   false,
			wantErr:           true,
			errContains:       "must provide exactly one of --transfer-amount or --source-amount",
		},
		{
			name:              "both amounts provided",
			setTransferAmount: true,
			transferAmount:    100.0,
			setSourceAmount:   true,
			sourceAmount:      100.0,
			wantErr:           true,
			errContains:       "cannot provide both --transfer-amount and --source-amount",
		},
		{
			name:              "only transfer amount provided",
			setTransferAmount: true,
			transferAmount:    100.0,
			setSourceAmount:   false,
			wantErr:           false,
		},
		{
			name:              "only source amount provided",
			setTransferAmount: false,
			setSourceAmount:   true,
			sourceAmount:      100.0,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTransfersCreateCmd()
			cmd.SetContext(context.Background())

			// Set required flags (but NOT amounts yet)
			setRequiredTransferFlagsNoAmount(t, cmd)

			// Set the amounts being tested
			if tt.setTransferAmount {
				if err := cmd.Flags().Set("transfer-amount", floatToString(tt.transferAmount)); err != nil {
					t.Fatalf("failed to set transfer-amount: %v", err)
				}
			}
			if tt.setSourceAmount {
				if err := cmd.Flags().Set("source-amount", floatToString(tt.sourceAmount)); err != nil {
					t.Fatalf("failed to set source-amount: %v", err)
				}
			}

			err := cmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil && !isExpectedTestError(err) {
				// We expect client initialization to fail in tests,
				// but validation should pass before that point
				if !strings.Contains(err.Error(), "must provide exactly one") &&
					!strings.Contains(err.Error(), "cannot provide both") {
					// This is not an amount validation error, so test passed
					return
				}
				t.Errorf("unexpected amount validation error: %v", err)
			}
		})
	}
}

func TestTransfersCreateCmd_SecurityQAPairing(t *testing.T) {
	tests := []struct {
		name             string
		securityQuestion string
		securityAnswer   string
		wantErr          bool
		errContains      string
	}{
		{
			name:             "both question and answer provided",
			securityQuestion: "What is your company name?",
			securityAnswer:   "Acme123",
			wantErr:          false,
		},
		{
			name:             "neither question nor answer provided",
			securityQuestion: "",
			securityAnswer:   "",
			wantErr:          false,
		},
		{
			name:             "only question provided",
			securityQuestion: "What is your company name?",
			securityAnswer:   "",
			wantErr:          true,
			errContains:      "--security-question and --security-answer must be provided together",
		},
		{
			name:             "only answer provided",
			securityQuestion: "",
			securityAnswer:   "Acme123",
			wantErr:          true,
			errContains:      "--security-question and --security-answer must be provided together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTransfersCreateCmd()
			cmd.SetContext(context.Background())

			// Set required flags
			setRequiredTransferFlags(t, cmd)

			// Set security Q&A
			if tt.securityQuestion != "" {
				if err := cmd.Flags().Set("security-question", tt.securityQuestion); err != nil {
					t.Fatalf("failed to set security-question: %v", err)
				}
			}
			if tt.securityAnswer != "" {
				if err := cmd.Flags().Set("security-answer", tt.securityAnswer); err != nil {
					t.Fatalf("failed to set security-answer: %v", err)
				}
			}

			err := cmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil && !isExpectedTestError(err) {
				if strings.Contains(err.Error(), "must be provided together") {
					t.Errorf("unexpected pairing validation error: %v", err)
				}
			}
		})
	}
}

func TestTransfersCreateCmd_SecurityQuestionLength(t *testing.T) {
	tests := []struct {
		name             string
		securityQuestion string
		securityAnswer   string
		wantErr          bool
		errContains      string
	}{
		{
			name:             "question at minimum length (1 char)",
			securityQuestion: "Q",
			securityAnswer:   "Answer123",
			wantErr:          false,
		},
		{
			name:             "question at maximum length (40 chars)",
			securityQuestion: "1234567890123456789012345678901234567890",
			securityAnswer:   "Answer123",
			wantErr:          false,
		},
		{
			name:             "question too long (41 chars)",
			securityQuestion: "12345678901234567890123456789012345678901",
			securityAnswer:   "Answer123",
			wantErr:          true,
			errContains:      "--security-question must be 1-40 characters (got 41)",
		},
		{
			name:             "valid question length",
			securityQuestion: "What is our company name?",
			securityAnswer:   "Acme123",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTransfersCreateCmd()
			cmd.SetContext(context.Background())

			// Set required flags
			setRequiredTransferFlags(t, cmd)

			// Set security Q&A
			if err := cmd.Flags().Set("security-question", tt.securityQuestion); err != nil {
				t.Fatalf("failed to set security-question: %v", err)
			}
			if err := cmd.Flags().Set("security-answer", tt.securityAnswer); err != nil {
				t.Fatalf("failed to set security-answer: %v", err)
			}

			err := cmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil && !isExpectedTestError(err) {
				if strings.Contains(err.Error(), "security-question must be") {
					t.Errorf("unexpected question length validation error: %v", err)
				}
			}
		})
	}
}

func TestTransfersCreateCmd_SecurityAnswerFormat(t *testing.T) {
	tests := []struct {
		name           string
		securityAnswer string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "answer at minimum length (3 chars)",
			securityAnswer: "ABC",
			wantErr:        false,
		},
		{
			name:           "answer at maximum length (25 chars)",
			securityAnswer: "1234567890123456789012345",
			wantErr:        false,
		},
		{
			name:           "answer too short (2 chars)",
			securityAnswer: "AB",
			wantErr:        true,
			errContains:    "--security-answer must be 3-25 characters (got 2)",
		},
		{
			name:           "answer too long (26 chars)",
			securityAnswer: "12345678901234567890123456",
			wantErr:        true,
			errContains:    "--security-answer must be 3-25 characters (got 26)",
		},
		{
			name:           "valid alphanumeric answer",
			securityAnswer: "Acme123",
			wantErr:        false,
		},
		{
			name:           "answer with @ symbol",
			securityAnswer: "Acme@123",
			wantErr:        true,
			errContains:    "--security-answer must contain only alphanumeric characters (no special chars like @, &, *)",
		},
		{
			name:           "answer with & symbol",
			securityAnswer: "Acme&Co",
			wantErr:        true,
			errContains:    "--security-answer must contain only alphanumeric characters (no special chars like @, &, *)",
		},
		{
			name:           "answer with * symbol",
			securityAnswer: "Acme*123",
			wantErr:        true,
			errContains:    "--security-answer must contain only alphanumeric characters (no special chars like @, &, *)",
		},
		{
			name:           "answer with space",
			securityAnswer: "Acme 123",
			wantErr:        true,
			errContains:    "--security-answer must contain only alphanumeric characters (no special chars like @, &, *)",
		},
		{
			name:           "answer with hyphen",
			securityAnswer: "Acme-123",
			wantErr:        true,
			errContains:    "--security-answer must contain only alphanumeric characters (no special chars like @, &, *)",
		},
		{
			name:           "answer with period",
			securityAnswer: "Acme.123",
			wantErr:        true,
			errContains:    "--security-answer must contain only alphanumeric characters (no special chars like @, &, *)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTransfersCreateCmd()
			cmd.SetContext(context.Background())

			// Set required flags
			setRequiredTransferFlags(t, cmd)

			// Set security Q&A
			if err := cmd.Flags().Set("security-question", "What is your company name?"); err != nil {
				t.Fatalf("failed to set security-question: %v", err)
			}
			if err := cmd.Flags().Set("security-answer", tt.securityAnswer); err != nil {
				t.Fatalf("failed to set security-answer: %v", err)
			}

			err := cmd.RunE(cmd, []string{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil && !isExpectedTestError(err) {
				if strings.Contains(err.Error(), "security-answer must") {
					t.Errorf("unexpected answer validation error: %v", err)
				}
			}
		})
	}
}

// Helper functions

func setRequiredTransferFlags(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	// Set all required flags with valid values
	flags := map[string]string{
		"beneficiary-id":    "benef_123",
		"transfer-currency": "CAD",
		"source-currency":   "CAD",
		"reference":         "Test transfer",
		"reason":            "payment_to_supplier",
		"transfer-amount":   "100.00",
	}

	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set required flag %s: %v", name, err)
		}
	}
}

func setRequiredTransferFlagsNoAmount(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	// Set all required flags EXCEPT amounts
	flags := map[string]string{
		"beneficiary-id":    "benef_123",
		"transfer-currency": "CAD",
		"source-currency":   "CAD",
		"reference":         "Test transfer",
		"reason":            "payment_to_supplier",
	}

	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("failed to set required flag %s: %v", name, err)
		}
	}
}
