package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

// mockStore is a mock implementation of secrets.Store for testing
type mockStore struct{}

func (m *mockStore) Get(account string) (secrets.Credentials, error) {
	return secrets.Credentials{
		ClientID:  "test-client-id",
		APIKey:    "test-api-key",
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockStore) Set(account string, creds secrets.Credentials) error {
	return nil
}

func (m *mockStore) Delete(account string) error {
	return nil
}

func (m *mockStore) Keys() ([]string, error) {
	return []string{"test-account"}, nil
}

func (m *mockStore) List() ([]secrets.Credentials, error) {
	return []secrets.Credentials{
		{
			ClientID:  "test-client-id",
			APIKey:    "test-api-key",
			CreatedAt: time.Now(),
		},
	}, nil
}

// isExpectedTestError checks if an error is expected in tests.
// When testing validation logic, we expect the command to:
// 1. Pass validation checks (what we're actually testing)
// 2. Fail at API/client initialization (because we use mock credentials)
// This function returns true for errors related to test infrastructure
// (client init, API calls, auth) so tests can distinguish between
// validation failures (unexpected) and infrastructure failures (expected).
func isExpectedTestError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "client") ||
		strings.Contains(msg, "context") ||
		strings.Contains(msg, "config") ||
		strings.Contains(msg, "API") ||
		strings.Contains(msg, "auth") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "Forbidden")
}

// setupTestEnvironment sets up the test environment with mocked secrets store.
// It sets the AWX_ACCOUNT environment variable and mocks the openSecretsStore function.
// Returns a cleanup function that should be deferred to restore the original state.
func setupTestEnvironment(t *testing.T) func() {
	t.Helper()
	t.Setenv("AWX_ACCOUNT", "test-account")
	original := openSecretsStore
	openSecretsStore = func() (secrets.Store, error) {
		return &mockStore{}, nil
	}
	return func() {
		openSecretsStore = original
	}
}

// intToString converts an integer to a string (used for flag values)
func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}

// floatToString converts a float to a string (used for flag values)
func floatToString(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
