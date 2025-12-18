package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and account management",
	}
	cmd.AddCommand(newAuthAddCmd())
	cmd.AddCommand(newAuthListCmd())
	cmd.AddCommand(newAuthRemoveCmd())
	cmd.AddCommand(newAuthTestCmd())
	return cmd
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
  airwallex auth add production --client-id xxx --api-key yyy

  # Multi-account API key (requires account-id)
  airwallex auth add production --client-id xxx --api-key yyy --account-id acct_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			name := args[0]

			if clientID == "" {
				return fmt.Errorf("--client-id is required")
			}

			if apiKey == "" {
				fmt.Fprint(os.Stderr, "API Key: ")
				key, err := term.ReadPassword(int(syscall.Stdin))
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

			store, err := openSecretsStore()
			if err != nil {
				return fmt.Errorf("failed to open keyring: %w", err)
			}

			err = store.Set(name, secrets.Credentials{
				ClientID:  clientID,
				APIKey:    apiKey,
				AccountID: accountID,
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

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, map[string]interface{}{
					"accounts": creds,
				})
			}

			if len(creds) == 0 {
				fmt.Fprintln(os.Stderr, "No accounts configured")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tCLIENT_ID\tCREATED")
			for _, c := range creds {
				fmt.Fprintf(tw, "%s\t%s\t%s\n", c.Name, c.ClientID, c.CreatedAt.Format("2006-01-02"))
			}
			tw.Flush()
			return nil
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
			var client *api.Client
			if creds.AccountID != "" {
				client = api.NewClientWithAccount(cmd.Context(), creds.ClientID, creds.APIKey, creds.AccountID)
			} else {
				client = api.NewClient(cmd.Context(), creds.ClientID, creds.APIKey)
			}
			_, err = client.Get("/api/v1/balances/current")
			if err != nil {
				u.Error(fmt.Sprintf("Authentication failed: %v", err))
				return err
			}

			u.Success("Credentials valid")
			return nil
		},
	}
}
