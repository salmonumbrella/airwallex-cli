package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAuthAddCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing account name",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
		{
			name: "missing client-id flag",
			args: []string{
				"production",
			},
			wantErr:     true,
			errContains: "--client-id is required",
		},
		{
			name: "invalid account name - empty",
			args: []string{
				"",
				"--client-id", "test-client-id",
				"--api-key", "test-api-key",
			},
			wantErr:     true,
			errContains: "invalid account name",
		},
		{
			name: "invalid account name - with spaces",
			args: []string{
				"production env",
				"--client-id", "test-client-id",
				"--api-key", "test-api-key",
			},
			wantErr:     true,
			errContains: "invalid account name",
		},
		{
			name: "invalid client-id - empty",
			args: []string{
				"production",
				"--client-id", "",
				"--api-key", "test-api-key",
			},
			wantErr:     true,
			errContains: "--client-id is required",
		},
		{
			name: "invalid api-key - empty",
			args: []string{
				"production",
				"--client-id", "test-client-id-123456789",
				"--api-key", "",
			},
			wantErr:     true,
			errContains: "invalid API key",
		},
		{
			name: "valid credentials without account-id",
			args: []string{
				"production",
				"--client-id", "test-client-id-123456789",
				"--api-key", "test-api-key-with-sufficient-length",
			},
			wantErr: false,
		},
		{
			name: "valid credentials with account-id",
			args: []string{
				"production",
				"--client-id", "test-client-id-123456789",
				"--api-key", "test-api-key-with-sufficient-length",
				"--account-id", "acct_123",
			},
			wantErr: false,
		},
		{
			name: "valid with whitespace trimmed",
			args: []string{
				"  production  ",
				"--client-id", "  test-client-id-123456789  ",
				"--api-key", "  test-api-key-with-sufficient-length  ",
				"--account-id", "  acct_123  ",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCmd := newAuthCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(authCmd)

			fullArgs := append([]string{"auth", "add"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestAuthListCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCmd := newAuthCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(authCmd)

			fullArgs := append([]string{"auth", "list"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestAuthRemoveCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing account name",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
		{
			name:        "too many arguments",
			args:        []string{"production", "staging"},
			wantErr:     true,
			errContains: "accepts 1 arg",
		},
		{
			name:    "valid account name",
			args:    []string{"production"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCmd := newAuthCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(authCmd)

			fullArgs := append([]string{"auth", "remove"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestAuthTestCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "no arguments with env account",
			args:    []string{},
			wantErr: false, // Should use AWX_ACCOUNT from environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCmd := newAuthCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(authCmd)

			fullArgs := append([]string{"auth", "test"}, tt.args...)
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
				// Auth test will fail on API call, which is expected
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestAuthLoginCommand(t *testing.T) {
	t.Skip("Skipping login test as it starts an actual HTTP server and waits for browser interaction")

	// Note: The login command:
	// 1. Starts an HTTP server on localhost
	// 2. Opens the browser for interactive authentication
	// 3. Waits up to 10 minutes for completion
	// This cannot be properly tested in a unit test without mocking the entire setup server
}

func TestAuthCommandStructure(t *testing.T) {
	authCmd := newAuthCmd()

	if authCmd.Use != "auth" {
		t.Errorf("expected Use to be 'auth', got %q", authCmd.Use)
	}

	if authCmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	expectedSubcommands := []string{"login", "add", "list", "remove", "test"}
	subcommands := authCmd.Commands()

	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range subcommands {
			// cmd.Use can be "add <name>" so we need to check if it starts with the expected name
			if strings.HasPrefix(cmd.Use, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

func TestAuthAddCommandHelp(t *testing.T) {
	authCmd := newAuthCmd()

	var addCmd *cobra.Command
	for _, cmd := range authCmd.Commands() {
		if cmd.Use == "add <name>" {
			addCmd = cmd
			break
		}
	}

	if addCmd == nil {
		t.Fatal("add command not found")
	}

	if !strings.Contains(addCmd.Long, "Examples:") {
		t.Error("add command help missing examples section")
	}

	if !strings.Contains(addCmd.Long, "--client-id") {
		t.Error("add command help missing --client-id flag documentation")
	}

	if !strings.Contains(addCmd.Long, "--account-id") {
		t.Error("add command help missing --account-id flag documentation")
	}

	// Check that flags are properly registered
	clientIDFlag := addCmd.Flags().Lookup("client-id")
	if clientIDFlag == nil {
		t.Error("--client-id flag not registered")
	}

	apiKeyFlag := addCmd.Flags().Lookup("api-key")
	if apiKeyFlag == nil {
		t.Error("--api-key flag not registered")
	}

	accountIDFlag := addCmd.Flags().Lookup("account-id")
	if accountIDFlag == nil {
		t.Error("--account-id flag not registered")
	}
}

func TestAuthLoginCommandHelp(t *testing.T) {
	authCmd := newAuthCmd()

	var loginCmd *cobra.Command
	for _, cmd := range authCmd.Commands() {
		if cmd.Use == "login" {
			loginCmd = cmd
			break
		}
	}

	if loginCmd == nil {
		t.Fatal("login command not found")
	}

	if !strings.Contains(loginCmd.Long, "browser") {
		t.Error("login command help should mention browser authentication")
	}

	if !strings.Contains(loginCmd.Long, "Examples:") {
		t.Error("login command help missing examples section")
	}
}

func TestAuthTestCommandHelp(t *testing.T) {
	authCmd := newAuthCmd()

	var testCmd *cobra.Command
	for _, cmd := range authCmd.Commands() {
		if cmd.Use == "test" {
			testCmd = cmd
			break
		}
	}

	if testCmd == nil {
		t.Fatal("test command not found")
	}

	if testCmd.Short == "" {
		t.Error("test command missing short description")
	}
}

func TestAuthValidationEdgeCases(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "account name with special characters",
			args: []string{
				"production-v2",
				"--client-id", "test-client-id-123456789",
				"--api-key", "test-api-key-with-sufficient-length",
			},
			wantErr: false,
		},
		{
			name: "account name with underscore",
			args: []string{
				"production_env",
				"--client-id", "test-client-id-123456789",
				"--api-key", "test-api-key-with-sufficient-length",
			},
			wantErr: false,
		},
		{
			name: "account name with numbers",
			args: []string{
				"production123",
				"--client-id", "test-client-id-123456789",
				"--api-key", "test-api-key-with-sufficient-length",
			},
			wantErr: false,
		},
		{
			name: "client-id with various formats",
			args: []string{
				"production",
				"--client-id", "20240101_test_client_id",
				"--api-key", "test-api-key-with-sufficient-length",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCmd := newAuthCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(authCmd)

			fullArgs := append([]string{"auth", "add"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

// TestAuthAddMultipleAccounts verifies that multiple accounts can be added
func TestAuthAddMultipleAccounts(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	accounts := []struct {
		name     string
		clientID string
		apiKey   string
	}{
		{"production", "prod-client-id-123456789", "prod-api-key-with-sufficient-length"},
		{"staging", "stag-client-id-123456789", "stag-api-key-with-sufficient-length"},
		{"development", "dev-client-id-123456789", "dev-api-key-with-sufficient-length"},
	}

	for _, acc := range accounts {
		t.Run(fmt.Sprintf("add-%s", acc.name), func(t *testing.T) {
			authCmd := newAuthCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(authCmd)

			args := []string{
				"auth", "add", acc.name,
				"--client-id", acc.clientID,
				"--api-key", acc.apiKey,
			}
			rootCmd.SetArgs(args)

			err := rootCmd.Execute()
			if err != nil && !isExpectedTestError(err) {
				t.Errorf("failed to add account %s: %v", acc.name, err)
			}
		})
	}
}

// TestAuthRemoveNonExistentAccount tests removing an account that doesn't exist
func TestAuthRemoveNonExistentAccount(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	authCmd := newAuthCmd()
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(authCmd)

	args := []string{"auth", "remove", "nonexistent-account"}
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	// The mock store doesn't return errors for delete, but a real store would
	if err != nil && !isExpectedTestError(err) && !strings.Contains(err.Error(), "failed to remove") {
		t.Errorf("unexpected error: %v", err)
	}
}
