package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/debug"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

type rootFlags struct {
	Account  string
	Output   string
	Color    string
	Debug    bool
	Query    string
	Template string // Go template for custom output
	// Agent-friendly flags
	Yes    bool   // skip confirmation prompts
	Limit  int    // limit number of results (0 = no limit)
	SortBy string // field name to sort by
	Desc   bool   // sort descending (only valid with --sort-by)
}

var flags rootFlags

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "airwallex",
		Short:        "Airwallex CLI for cards, transfers, and more",
		Long:         "A command-line interface for the Airwallex API.",
		Version:      Version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate flag combinations
			if flags.Desc && flags.SortBy == "" {
				return fmt.Errorf("--desc requires --sort-by to be specified")
			}

			// Setup debug mode
			debug.SetupLogger(flags.Debug)
			ctx := debug.WithDebug(cmd.Context(), flags.Debug)

			// Inject IO streams (only if not already set, to support testing)
			if !iocontext.HasIO(ctx) {
				ctx = iocontext.WithIO(ctx, iocontext.DefaultIO())
			}

			// Inject UI context
			u := ui.New(flags.Color)
			ctx = ui.WithUI(ctx, u)

			// Inject output format context
			ctx = outfmt.WithFormat(ctx, flags.Output)

			// Inject query filter context
			ctx = outfmt.WithQuery(ctx, flags.Query)

			// Inject template format context
			ctx = outfmt.WithTemplate(ctx, flags.Template)

			// Inject agent-friendly flags
			ctx = outfmt.WithYes(ctx, flags.Yes)
			ctx = outfmt.WithLimit(ctx, flags.Limit)
			ctx = outfmt.WithSortBy(ctx, flags.SortBy)
			ctx = outfmt.WithDesc(ctx, flags.Desc)

			cmd.SetContext(ctx)
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&flags.Account, "account", os.Getenv("AWX_ACCOUNT"), "Account name (or AWX_ACCOUNT env)")
	cmd.PersistentFlags().StringVar(&flags.Output, "output", getEnvOrDefault("AWX_OUTPUT", "text"), "Output format: text|json")
	cmd.PersistentFlags().StringVar(&flags.Color, "color", getEnvOrDefault("AWX_COLOR", "auto"), "Color output: auto|always|never")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug output (shows API requests/responses)")
	cmd.PersistentFlags().StringVar(&flags.Query, "query", "", "JQ filter expression for JSON output")
	cmd.PersistentFlags().StringVar(&flags.Template, "format", "", "Go template for custom output (e.g., '{{.ID}}: {{.Status}}')")

	// Agent-friendly flags
	cmd.PersistentFlags().BoolVarP(&flags.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.PersistentFlags().BoolVar(&flags.Yes, "force", false, "Skip confirmation prompts (alias for --yes)")
	cmd.PersistentFlags().IntVar(&flags.Limit, "limit", 0, "Limit number of results (0 = no limit)")
	cmd.PersistentFlags().StringVar(&flags.SortBy, "sort-by", "", "Field name to sort results by")
	cmd.PersistentFlags().BoolVar(&flags.Desc, "desc", false, "Sort descending (requires --sort-by)")

	cmd.AddCommand(newAPICmd())
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
	cmd.AddCommand(newPayersCmd())
	cmd.AddCommand(newBillingCmd())

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
