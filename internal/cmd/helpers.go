package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
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

	if creds.AccountID != "" {
		return api.NewClientWithAccount(creds.ClientID, creds.APIKey, creds.AccountID)
	}
	return api.NewClient(creds.ClientID, creds.APIKey)
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
