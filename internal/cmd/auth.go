package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/salmonumbrella/airwallex-cli/internal/auth"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and account management",
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthAddCmd())
	cmd.AddCommand(newAuthListCmd())
	cmd.AddCommand(newAuthRemoveCmd())
	cmd.AddCommand(newAuthRenameCmd())
	cmd.AddCommand(newAuthTestCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate via browser",
		Long: `Opens a browser window to configure API credentials interactively.

This provides a guided setup experience with:
  - Links to find your API credentials
  - Connection testing before saving
  - Secure credential storage in keychain

Examples:
  airwallex auth login`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())

			store, err := openSecretsStore()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			u.Info("Opening browser for authentication setup...")
			u.Info("Complete the setup in your browser, then return here.")

			// Create context with timeout and cancellation
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
			defer cancel()

			// Handle interrupt
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				cancel()
			}()

			server, err := auth.NewSetupServer(store)
			if err != nil {
				return fmt.Errorf("failed to create setup server: %w", err)
			}
			result, err := server.Start(ctx)
			if err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}

			if result.Error != nil {
				return result.Error
			}

			u.Success(fmt.Sprintf("Account '%s' configured successfully!", result.AccountName))
			return nil
		},
	}
}

func newAuthAddCmd() *cobra.Command {
	var clientID string
	var apiKey string
	var accountID string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add account credentials",
		Long: `Add account credentials for API authentication.

The account-id flag is required when your API key has access to multiple accounts.
It specifies which account the token should be authorized for (sent as x-login-as header).

Examples:
  # Basic authentication (single account API key)
  airwallex auth add production --client-id xxx
  # You'll be prompted securely for API Key

  # Multi-account API key (requires account-id)
  airwallex auth add production --client-id xxx --account-id acct_xxx
  # You'll be prompted securely for API Key`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			name := strings.TrimSpace(args[0])

			// Validate account name
			if err := auth.ValidateAccountName(name); err != nil {
				return fmt.Errorf("invalid account name: %w", err)
			}

			if clientID == "" {
				return fmt.Errorf("--client-id is required")
			}

			clientID = strings.TrimSpace(clientID)
			if err := auth.ValidateClientID(clientID); err != nil {
				return fmt.Errorf("invalid client ID: %w", err)
			}

			if apiKey == "" {
				fmt.Fprint(os.Stderr, "API Key: ")
				key, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					// Fallback for non-terminal
					reader := bufio.NewReader(os.Stdin)
					line, _ := reader.ReadString('\n')
					apiKey = strings.TrimSpace(line)
				} else {
					apiKey = string(key)
					fmt.Fprintln(os.Stderr)
				}
			}

			apiKey = strings.TrimSpace(apiKey)
			if err := auth.ValidateAPIKey(apiKey); err != nil {
				return fmt.Errorf("invalid API key: %w", err)
			}

			store, err := openSecretsStore()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			err = store.Set(name, secrets.Credentials{
				ClientID:  clientID,
				APIKey:    apiKey,
				AccountID: strings.TrimSpace(accountID),
			})
			if err != nil {
				return fmt.Errorf("failed to store credentials: %w", err)
			}

			u.Success(fmt.Sprintf("Added account: %s", name))
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "Airwallex Client ID (required)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Airwallex API Key (omit to prompt)")
	cmd.Flags().StringVar(&accountID, "account-id", "", "Airwallex Account ID for x-login-as (required for multi-account API keys)")
	return cmd
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openSecretsStore()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			creds, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list accounts: %w", err)
			}

			f := outfmt.FromContext(cmd.Context())

			if outfmt.IsJSON(cmd.Context()) {
				return f.Output(map[string]interface{}{
					"accounts": creds,
				})
			}

			if len(creds) == 0 {
				f.Empty("No accounts configured")
				return nil
			}

			f.StartTable([]string{"NAME", "CLIENT_ID", "CREATED"})
			for _, c := range creds {
				f.Row(c.Name, c.ClientID, c.CreatedAt.Format("2006-01-02"))
			}
			return f.EndTable()
		},
	}
}

func newAuthRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove account credentials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			name := args[0]

			store, err := openSecretsStore()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			if err := store.Delete(name); err != nil {
				return fmt.Errorf("failed to remove account: %w", err)
			}

			u.Success(fmt.Sprintf("Removed account: %s", name))
			return nil
		},
	}
}

func newAuthRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename an account",
		Long: `Rename an existing account to a new name.

Examples:
  airwallex auth rename production prod
  airwallex auth rename vlad-local dm`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			oldName := strings.TrimSpace(args[0])
			newName := strings.TrimSpace(args[1])

			// Validate new name
			if err := auth.ValidateAccountName(newName); err != nil {
				return fmt.Errorf("invalid new account name: %w", err)
			}

			store, err := openSecretsStore()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			// Get existing credentials
			creds, err := store.Get(oldName)
			if err != nil {
				return fmt.Errorf("account not found: %s", oldName)
			}

			// Check if new name already exists
			if _, err := store.Get(newName); err == nil {
				return fmt.Errorf("account already exists: %s", newName)
			}

			// Set with new name (preserve CreatedAt)
			err = store.Set(newName, secrets.Credentials{
				ClientID:  creds.ClientID,
				APIKey:    creds.APIKey,
				AccountID: creds.AccountID,
				CreatedAt: creds.CreatedAt,
			})
			if err != nil {
				return fmt.Errorf("failed to create new account: %w", err)
			}

			// Delete old name
			if err := store.Delete(oldName); err != nil {
				// Try to rollback
				_ = store.Delete(newName)
				return fmt.Errorf("failed to remove old account: %w", err)
			}

			u.Success(fmt.Sprintf("Renamed account: %s → %s", oldName, newName))
			return nil
		},
	}
}

func newAuthTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test account credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			account, err := requireAccount(&flags)
			if err != nil {
				return err
			}

			store, err := openSecretsStore()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			creds, err := store.Get(account)
			if err != nil {
				return fmt.Errorf("account not found: %s", account)
			}

			u.Info(fmt.Sprintf("Testing account: %s (client_id: %s)", account, creds.ClientID))

			// Actually test the credentials by fetching a token
			client, err := newClientForCreds(creds)
			if err != nil {
				u.Error(fmt.Sprintf("Failed to create client: %v", err))
				return err
			}
			_, err = client.Get(cmd.Context(), "/api/v1/balances/current")
			if err != nil {
				u.Error(fmt.Sprintf("Authentication failed: %v", err))
				return err
			}

			u.Success("Credentials valid")
			return nil
		},
	}
}
