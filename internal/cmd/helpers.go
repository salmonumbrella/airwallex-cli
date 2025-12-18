package cmd

import (
	"context"
	"fmt"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

// openSecretsStore is a variable that can be overridden in tests
var openSecretsStore = secrets.OpenDefault

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
		return api.NewClientWithAccount(ctx, creds.ClientID, creds.APIKey, creds.AccountID), nil
	}
	return api.NewClient(ctx, creds.ClientID, creds.APIKey), nil
}
