package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/debug"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

type rootFlags struct {
	Account string
	Output  string
	Color   string
	Debug   bool
	Query   string
}

var flags rootFlags

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "airwallex",
		Short:        "Airwallex CLI for cards, transfers, and more",
		Long:         "A command-line interface for the Airwallex API.",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Setup debug mode
			debug.SetupLogger(flags.Debug)
			ctx := debug.WithDebug(cmd.Context(), flags.Debug)

			// Inject UI context
			u := ui.New(flags.Color)
			ctx = ui.WithUI(ctx, u)

			// Inject output format context
			ctx = outfmt.WithFormat(ctx, flags.Output)

			// Inject query filter context
			ctx = outfmt.WithQuery(ctx, flags.Query)

			cmd.SetContext(ctx)
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&flags.Account, "account", os.Getenv("AWX_ACCOUNT"), "Account name (or AWX_ACCOUNT env)")
	cmd.PersistentFlags().StringVar(&flags.Output, "output", getEnvOrDefault("AWX_OUTPUT", "text"), "Output format: text|json")
	cmd.PersistentFlags().StringVar(&flags.Color, "color", getEnvOrDefault("AWX_COLOR", "auto"), "Color output: auto|always|never")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug output (shows API requests/responses)")
	cmd.PersistentFlags().StringVar(&flags.Query, "query", "", "JQ filter expression for JSON output")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newBalancesCmd())
	cmd.AddCommand(newIssuingCmd())
	cmd.AddCommand(newTransfersCmd())
	cmd.AddCommand(newBeneficiariesCmd())
	cmd.AddCommand(newAccountsCmd())
	cmd.AddCommand(newReportsCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newUpgradeCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newFXCmd())
	cmd.AddCommand(newDepositsCmd())
	cmd.AddCommand(newLinkedAccountsCmd())
	cmd.AddCommand(newSchemasCmd())
	cmd.AddCommand(newPaymentLinksCmd())
	cmd.AddCommand(newWebhooksCmd())

	return cmd
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func requireAccount(f *rootFlags) (string, error) {
	// Check explicit flag or env var first
	if f.Account != "" {
		return f.Account, nil
	}

	// Try to auto-select if only one account is configured
	store, err := openSecretsStore()
	if err != nil {
		return "", fmt.Errorf("failed to access keyring: %w. Use --account or set AWX_ACCOUNT", err)
	}

	accounts, err := store.List()
	if err != nil {
		return "", fmt.Errorf("failed to list accounts: %w. Use --account or set AWX_ACCOUNT", err)
	}

	switch len(accounts) {
	case 0:
		return "", fmt.Errorf("no accounts configured. Run: airwallex auth login OR airwallex auth add <name> --client-id <id>")
	case 1:
		// Auto-select the only account
		return accounts[0].Name, nil
	default:
		// Multiple accounts - list them
		names := make([]string, len(accounts))
		for i, a := range accounts {
			names[i] = a.Name
		}
		return "", fmt.Errorf("multiple accounts configured: %s\nSpecify with --account <name> or set AWX_ACCOUNT", strings.Join(names, ", "))
	}
}

func Execute(args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func ExecuteContext(ctx context.Context, args []string) error {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}
