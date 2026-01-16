package cmd

import (
	"context"
	"strings"
	"testing"
)

// TestFXRatesCommand tests the FX rates command flag validation
func TestFXRatesCommand(t *testing.T) {
	tests := []struct {
		name        string
		sellCur     string
		buyCur      string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid with both currencies",
			sellCur: "USD",
			buyCur:  "EUR",
			wantErr: false,
		},
		{
			name:        "error with only sell currency",
			sellCur:     "USD",
			buyCur:      "",
			wantErr:     true,
			errContains: "both --sell and --buy currencies are required",
		},
		{
			name:        "error with only buy currency",
			sellCur:     "",
			buyCur:      "EUR",
			wantErr:     true,
			errContains: "both --sell and --buy currencies are required",
		},
		{
			name:        "error with no currencies",
			sellCur:     "",
			buyCur:      "",
			wantErr:     true,
			errContains: "both --sell and --buy currencies are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			cmd := newFXRatesCmd()
			cmd.SetContext(context.Background())

			if tt.sellCur != "" {
				if err := cmd.Flags().Set("sell", tt.sellCur); err != nil {
					t.Fatalf("failed to set sell flag: %v", err)
				}
			}
			if tt.buyCur != "" {
				if err := cmd.Flags().Set("buy", tt.buyCur); err != nil {
					t.Fatalf("failed to set buy flag: %v", err)
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
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestFXQuotesCreateCommand tests the FX quotes create command validation
func TestFXQuotesCreateCommand(t *testing.T) {
	tests := []struct {
		name        string
		sellCur     string
		buyCur      string
		sellAmount  float64
		buyAmount   float64
		validity    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "both sell and buy amounts provided",
			sellCur:     "USD",
			buyCur:      "EUR",
			sellAmount:  1000.0,
			buyAmount:   900.0,
			validity:    "1h",
			wantErr:     true,
			errContains: "cannot provide both --sell-amount and --buy-amount",
		},
		{
			name:        "neither sell nor buy amount provided",
			sellCur:     "USD",
			buyCur:      "EUR",
			sellAmount:  0,
			buyAmount:   0,
			validity:    "1h",
			wantErr:     true,
			errContains: "must provide exactly one of --sell-amount or --buy-amount",
		},
		{
			name:       "valid with sell amount",
			sellCur:    "USD",
			buyCur:     "EUR",
			sellAmount: 1000.0,
			buyAmount:  0,
			validity:   "1h",
			wantErr:    false,
		},
		{
			name:       "valid with buy amount",
			sellCur:    "USD",
			buyCur:     "EUR",
			sellAmount: 0,
			buyAmount:  900.0,
			validity:   "24h",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			cmd := newFXQuotesCreateCmd()
			cmd.SetContext(context.Background())

			if tt.sellCur != "" {
				if err := cmd.Flags().Set("sell-currency", tt.sellCur); err != nil {
					t.Fatalf("failed to set sell-currency flag: %v", err)
				}
			}
			if tt.buyCur != "" {
				if err := cmd.Flags().Set("buy-currency", tt.buyCur); err != nil {
					t.Fatalf("failed to set buy-currency flag: %v", err)
				}
			}
			if tt.sellAmount > 0 {
				if err := cmd.Flags().Set("sell-amount", floatToString(tt.sellAmount)); err != nil {
					t.Fatalf("failed to set sell-amount flag: %v", err)
				}
			}
			if tt.buyAmount > 0 {
				if err := cmd.Flags().Set("buy-amount", floatToString(tt.buyAmount)); err != nil {
					t.Fatalf("failed to set buy-amount flag: %v", err)
				}
			}
			if tt.validity != "" {
				if err := cmd.Flags().Set("validity", tt.validity); err != nil {
					t.Fatalf("failed to set validity flag: %v", err)
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
				// Check if it's one of the expected validation errors
				if strings.Contains(err.Error(), "must provide exactly one") ||
					strings.Contains(err.Error(), "cannot provide both") {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

// TestFXQuotesGetCommand tests the FX quotes get command argument validation
func TestFXQuotesGetCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no quote ID provided",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 0",
		},
		{
			name:        "too many arguments",
			args:        []string{"quote_123", "extra_arg"},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 2",
		},
		{
			name:    "valid quote ID",
			args:    []string{"quote_123"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			cmd := newFXQuotesGetCmd()
			cmd.SetContext(context.Background())

			// Test Args validation first
			if cmd.Args != nil {
				if err := cmd.Args(cmd, tt.args); err != nil {
					if tt.wantErr {
						if !strings.Contains(err.Error(), tt.errContains) {
							t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
						}
						return
					}
					t.Errorf("unexpected Args validation error: %v", err)
					return
				}
			}

			// If Args validation passed and we have valid args, test RunE
			if !tt.wantErr && len(tt.args) > 0 {
				err := cmd.RunE(cmd, tt.args)
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			} else if tt.wantErr && cmd.Args == nil {
				t.Errorf("expected Args validator to be set, but it was nil")
			}
		})
	}
}

// TestFXConversionsListCommand tests the FX conversions list command validation
func TestFXConversionsListCommand(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		fromDate    string
		toDate      string
		limit       int
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid with no filters",
			status:   "",
			fromDate: "",
			toDate:   "",
			limit:    20,
			wantErr:  false,
		},
		{
			name:     "valid with status filter",
			status:   "COMPLETED",
			fromDate: "",
			toDate:   "",
			limit:    20,
			wantErr:  false,
		},
		{
			name:     "valid with date range",
			status:   "",
			fromDate: "2024-01-01",
			toDate:   "2024-01-31",
			limit:    20,
			wantErr:  false,
		},
		{
			name:     "limit below minimum gets adjusted to 10",
			status:   "",
			fromDate: "",
			toDate:   "",
			limit:    5,
			wantErr:  false,
		},
		{
			name:     "valid custom limit",
			status:   "",
			fromDate: "",
			toDate:   "",
			limit:    50,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			cmd := newFXConversionsListCmd()
			cmd.SetContext(context.Background())

			if tt.status != "" {
				if err := cmd.Flags().Set("status", tt.status); err != nil {
					t.Fatalf("failed to set status flag: %v", err)
				}
			}
			if tt.fromDate != "" {
				if err := cmd.Flags().Set("from", tt.fromDate); err != nil {
					t.Fatalf("failed to set from flag: %v", err)
				}
			}
			if tt.toDate != "" {
				if err := cmd.Flags().Set("to", tt.toDate); err != nil {
					t.Fatalf("failed to set to flag: %v", err)
				}
			}
			if tt.limit > 0 {
				if err := cmd.Flags().Set("page-size", intToString(tt.limit)); err != nil {
					t.Fatalf("failed to set page-size flag: %v", err)
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
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestFXConversionsGetCommand tests the FX conversions get command argument validation
func TestFXConversionsGetCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no conversion ID provided",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 0",
		},
		{
			name:        "too many arguments",
			args:        []string{"conv_123", "extra_arg"},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 2",
		},
		{
			name:    "valid conversion ID",
			args:    []string{"conv_123"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			cmd := newFXConversionsGetCmd()
			cmd.SetContext(context.Background())

			// Test Args validation first
			if cmd.Args != nil {
				if err := cmd.Args(cmd, tt.args); err != nil {
					if tt.wantErr {
						if !strings.Contains(err.Error(), tt.errContains) {
							t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
						}
						return
					}
					t.Errorf("unexpected Args validation error: %v", err)
					return
				}
			}

			// If Args validation passed and we have valid args, test RunE
			if !tt.wantErr && len(tt.args) > 0 {
				err := cmd.RunE(cmd, tt.args)
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			} else if tt.wantErr && cmd.Args == nil {
				t.Errorf("expected Args validator to be set, but it was nil")
			}
		})
	}
}

// TestFXConversionsCreateCommand tests the FX conversions create command validation
func TestFXConversionsCreateCommand(t *testing.T) {
	tests := []struct {
		name        string
		quoteID     string
		sellCur     string
		buyCur      string
		sellAmount  float64
		buyAmount   float64
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid with quote ID",
			quoteID:    "quote_123",
			sellCur:    "",
			buyCur:     "",
			sellAmount: 0,
			buyAmount:  0,
			wantErr:    false,
		},
		{
			name:       "valid market rate with sell amount",
			quoteID:    "",
			sellCur:    "USD",
			buyCur:     "EUR",
			sellAmount: 1000.0,
			buyAmount:  0,
			wantErr:    false,
		},
		{
			name:       "valid market rate with buy amount",
			quoteID:    "",
			sellCur:    "USD",
			buyCur:     "EUR",
			sellAmount: 0,
			buyAmount:  900.0,
			wantErr:    false,
		},
		// Note: The case where market rate conversion is attempted without currencies
		// but with an amount is not currently validated and would result in an API error.
		// This could be improved in the implementation.
		{
			name:        "market rate both amounts provided",
			quoteID:     "",
			sellCur:     "USD",
			buyCur:      "EUR",
			sellAmount:  1000.0,
			buyAmount:   900.0,
			wantErr:     true,
			errContains: "cannot provide both --sell-amount and --buy-amount",
		},
		{
			name:        "market rate neither amount provided",
			quoteID:     "",
			sellCur:     "USD",
			buyCur:      "EUR",
			sellAmount:  0,
			buyAmount:   0,
			wantErr:     true,
			errContains: "must provide --quote-id OR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			cmd := newFXConversionsCreateCmd()
			cmd.SetContext(context.Background())

			if tt.quoteID != "" {
				if err := cmd.Flags().Set("quote-id", tt.quoteID); err != nil {
					t.Fatalf("failed to set quote-id flag: %v", err)
				}
			}
			if tt.sellCur != "" {
				if err := cmd.Flags().Set("sell-currency", tt.sellCur); err != nil {
					t.Fatalf("failed to set sell-currency flag: %v", err)
				}
			}
			if tt.buyCur != "" {
				if err := cmd.Flags().Set("buy-currency", tt.buyCur); err != nil {
					t.Fatalf("failed to set buy-currency flag: %v", err)
				}
			}
			if tt.sellAmount > 0 {
				if err := cmd.Flags().Set("sell-amount", floatToString(tt.sellAmount)); err != nil {
					t.Fatalf("failed to set sell-amount flag: %v", err)
				}
			}
			if tt.buyAmount > 0 {
				if err := cmd.Flags().Set("buy-amount", floatToString(tt.buyAmount)); err != nil {
					t.Fatalf("failed to set buy-amount flag: %v", err)
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
				// Check if it's one of the expected validation errors
				if strings.Contains(err.Error(), "must provide --quote-id") ||
					strings.Contains(err.Error(), "cannot provide both") {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

// TestFXConversionsListCommand_PageSizeValidation tests page size validation
func TestFXConversionsListCommand_PageSizeFlag(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	cmd := newFXConversionsListCmd()

	pageSizeFlag := cmd.Flags().Lookup("page-size")
	if pageSizeFlag == nil {
		t.Fatal("page-size flag not found")
	}

	if pageSizeFlag.Deprecated != "" {
		t.Errorf("expected page-size flag to be active, got deprecated: %s", pageSizeFlag.Deprecated)
	}

	if pageSizeFlag.DefValue != "20" {
		t.Errorf("expected default page-size to be 20, got: %s", pageSizeFlag.DefValue)
	}
}
