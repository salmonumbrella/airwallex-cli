package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/batch"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

// openSecretsStore is a variable that can be overridden in tests
var openSecretsStore = secrets.OpenDefault

// mustMarkRequired marks a flag as required, panicking on error.
// Use for flags that are definitely defined - panics indicate programmer error.
func mustMarkRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Sprintf("flag %q not defined: %v", name, err))
	}
}

// newClientForCreds is a variable that can be overridden in tests.
var newClientForCreds = func(creds secrets.Credentials) (*api.Client, error) {
	if creds.AccountID != "" {
		return api.NewClientWithAccount(creds.ClientID, creds.APIKey, creds.AccountID)
	}
	return api.NewClient(creds.ClientID, creds.APIKey)
}

// getClient creates an API client from the current account
func getClient(ctx context.Context) (*api.Client, error) {
	account, err := requireAccount(&flags)
	if err != nil {
		return nil, err
	}

	store, err := openSecretsStore()
	if err != nil {
		return nil, err
	}

	creds, err := store.Get(account)
	if err != nil {
		return nil, fmt.Errorf("account not found: %s", account)
	}

	return newClientForCreds(creds)
}

// convertDateToRFC3339 converts a date string in YYYY-MM-DD format to RFC3339 format
// with time set to 00:00:00 UTC
func convertDateToRFC3339(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("expected format YYYY-MM-DD, got %q", dateStr)
	}
	// Convert to RFC3339 format with UTC timezone
	return t.UTC().Format(time.RFC3339), nil
}

// convertDateToRFC3339End converts a date string in YYYY-MM-DD format to RFC3339 format
// with time set to 23:59:59 UTC (inclusive end-of-day).
func convertDateToRFC3339End(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("expected format YYYY-MM-DD, got %q", dateStr)
	}
	endOfDay := t.UTC().Add(24*time.Hour - time.Second)
	return endOfDay.Format(time.RFC3339), nil
}

// validateDate validates that a date string is in YYYY-MM-DD format
func validateDate(dateStr string) error {
	if dateStr == "" {
		return nil // empty is valid (optional parameter)
	}
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: expected YYYY-MM-DD, got %q", dateStr)
	}
	return nil
}

// validateAmount validates that an amount is positive
func validateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive, got %.2f", amount)
	}
	return nil
}

// validateCurrency validates that a currency code is 3 uppercase letters
func validateCurrency(currency string) error {
	if currency == "" {
		return nil // empty is valid (optional parameter)
	}
	if len(currency) != 3 {
		return fmt.Errorf("currency must be 3 letters, got %q", currency)
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return fmt.Errorf("currency must be uppercase letters, got %q", currency)
		}
	}
	return nil
}

// validateDateRange validates that from date is before to date when both are provided
func validateDateRange(from, to string) error {
	if from == "" || to == "" {
		return nil // only validate if both are provided
	}

	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return fmt.Errorf("invalid from date: expected YYYY-MM-DD, got %q", from)
	}

	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		return fmt.Errorf("invalid to date: expected YYYY-MM-DD, got %q", to)
	}

	if fromTime.After(toTime) {
		return fmt.Errorf("from date (%s) must be before or equal to to date (%s)", from, to)
	}

	return nil
}

// isTerminal is a variable that can be overridden in tests
var isTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// ConfirmOrYes prompts for confirmation unless --yes/--force flag is set.
// Returns true if confirmed, false if declined.
// Returns an error if stdin is not a TTY and confirmation is needed.
func ConfirmOrYes(ctx context.Context, prompt string) (bool, error) {
	// If --yes or --force flag is set, skip confirmation
	if outfmt.GetYes(ctx) {
		return true, nil
	}

	// If JSON output mode, skip confirmation (scripts expect non-interactive)
	if outfmt.IsJSON(ctx) {
		return true, nil
	}

	// Check if stdin is a terminal
	if !isTerminal() {
		return false, fmt.Errorf("cannot prompt for confirmation: stdin is not a terminal (use --yes to skip)")
	}

	// Get IO from context
	io := iocontext.GetIO(ctx)

	// Print prompt to stderr (so it doesn't interfere with stdout output)
	_, _ = fmt.Fprint(io.ErrOut, prompt+" [y/N]: ")

	// Read response from stdin
	reader := bufio.NewReader(io.In)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

func readJSONPayload(data, fromFile string) (map[string]interface{}, error) {
	if data != "" && fromFile != "" {
		return nil, fmt.Errorf("use only one of --data or --from-file")
	}

	var reader io.Reader
	switch {
	case fromFile != "":
		if fromFile == "-" {
			reader = os.Stdin
		} else {
			//nolint:gosec // G304: filename comes from user input, intentional
			f, err := os.Open(fromFile)
			if err != nil {
				return nil, fmt.Errorf("failed to open file: %w", err)
			}
			defer func() { _ = f.Close() }()
			reader = f
		}
	case data != "":
		reader = strings.NewReader(data)
	default:
		return nil, fmt.Errorf("provide --data or --from-file")
	}

	limitedReader := io.LimitReader(reader, batch.MaxInputSize+1)
	payload, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON payload: %w", err)
	}
	if len(payload) > batch.MaxInputSize {
		return nil, fmt.Errorf("input too large: exceeds maximum size of %d bytes", batch.MaxInputSize)
	}

	var result map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid JSON object: %w", err)
	}
	return result, nil
}

func readOptionalJSONPayload(data, fromFile string) (map[string]interface{}, error) {
	if data == "" && fromFile == "" {
		return nil, nil
	}
	return readJSONPayload(data, fromFile)
}

// normalizePageSize ensures page size is at least the API minimum (10)
func normalizePageSize(pageSize int) int {
	if pageSize < 10 {
		return 10
	}
	return pageSize
}
