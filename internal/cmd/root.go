package cmd

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/debug"
	"github.com/salmonumbrella/airwallex-cli/internal/exitcode"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

//go:embed help.txt
var helpText string

type rootFlags struct {
	Account   string
	Output    string
	Color     string
	Debug     bool
	Query     string
	QueryFile string
	Template  string // Go template for custom output
	JSON      bool   // shorthand for --output json
	NoColor   bool   // shorthand for --color never
	Agent     bool   // agent mode: stable JSON, no colors, no prompts, structured errors
	// Agent-friendly flags
	Yes         bool   // skip confirmation prompts
	NoInput     bool   // disable interactive prompts
	ItemsOnly   bool   // output items/results array only when present
	OutputLimit int    // limit number of results in output (0 = no limit)
	SortBy      string // field name to sort by
	Desc        bool   // sort descending (only valid with --sort-by)
}

type rootFlagsKey struct{}

func withRootFlags(ctx context.Context, f *rootFlags) context.Context {
	return context.WithValue(ctx, rootFlagsKey{}, f)
}

func rootFlagsFromContext(ctx context.Context) (*rootFlags, bool) {
	f, ok := ctx.Value(rootFlagsKey{}).(*rootFlags)
	return f, ok
}

func binaryName() string {
	if len(os.Args) > 0 {
		return filepath.Base(os.Args[0])
	}
	return "awx"
}

func NewRootCmd() *cobra.Command {
	flags := &rootFlags{}
	cmd := &cobra.Command{
		Use:          binaryName(),
		Short:        "Airwallex CLI for cards, transfers, and more",
		Long:         "A command-line interface for the Airwallex API.",
		Version:      Version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Agent mode defaults: stable JSON output, no colors, and no interactive prompts.
			// Respect explicit user choices.
			if flags.Agent {
				if !cmd.Flags().Changed("output") && !flags.JSON && flags.Template == "" {
					flags.Output = "json"
				}
				if !cmd.Flags().Changed("color") && !flags.NoColor {
					flags.Color = "never"
				}
				if !cmd.Flags().Changed("yes") && !cmd.Flags().Changed("force") {
					flags.Yes = true
				}
				// Prevent Cobra from printing unstructured errors.
				cmd.SilenceErrors = true
				if r := cmd.Root(); r != nil {
					r.SilenceErrors = true
				}
			}

			// Desire-path shorthands. Respect explicit --output/--color if set.
			if flags.JSON && !cmd.Flags().Changed("output") {
				flags.Output = "json"
			}
			if flags.NoColor && !cmd.Flags().Changed("color") {
				flags.Color = "never"
			}

			if flags.Yes {
				flags.NoInput = true
			}

			flags.Output = outfmt.NormalizeFormat(flags.Output)

			// Auto-enable JSON output when --query/--jq/--query-file is set.
			// JQ filtering only makes sense with JSON output.
			if (flagOrAliasChanged(cmd, "query") || flagOrAliasChanged(cmd, "query-file")) && !flagOrAliasChanged(cmd, "output") && !flags.JSON {
				flags.Output = "json"
			}

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

			query, err := readQueryInput(flags.Query, flags.QueryFile)
			if err != nil {
				return err
			}

			// Inject query filter context
			ctx = outfmt.WithQuery(ctx, query)

			// Inject template format context
			ctx = outfmt.WithTemplate(ctx, flags.Template)

			// Inject agent-friendly flags
			ctx = outfmt.WithYes(ctx, flags.Yes)
			ctx = outfmt.WithNoInput(ctx, flags.NoInput)
			ctx = outfmt.WithItemsOnly(ctx, flags.ItemsOnly)
			ctx = outfmt.WithLimit(ctx, flags.OutputLimit)
			ctx = outfmt.WithSortBy(ctx, flags.SortBy)
			ctx = outfmt.WithDesc(ctx, flags.Desc)

			ctx = withRootFlags(ctx, flags)
			cmd.SetContext(ctx)
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&flags.Account, "account", os.Getenv("AWX_ACCOUNT"), "Account name (or AWX_ACCOUNT env)")
	cmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", getEnvOrDefault("AWX_OUTPUT", "text"), "Output format: text|json|jsonl|ndjson (env AWX_OUTPUT)")
	cmd.PersistentFlags().BoolVarP(&flags.JSON, "json", "j", false, "Shorthand for --output json")
	cmd.PersistentFlags().StringVar(&flags.Color, "color", getEnvOrDefault("AWX_COLOR", "auto"), "Color output: auto|always|never")
	cmd.PersistentFlags().BoolVar(&flags.NoColor, "no-color", false, "Shorthand for --color never")
	cmd.PersistentFlags().BoolVar(&flags.Agent, "agent", os.Getenv("AWX_AGENT") != "", "Agent mode: stable JSON, no color, no prompts (or AWX_AGENT env)")
	cmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Enable debug output (shows API requests/responses)")
	cmd.PersistentFlags().StringVarP(&flags.Query, "query", "q", "", "JQ expression to filter JSON output")
	cmd.PersistentFlags().StringVar(&flags.QueryFile, "query-file", "", "Read JQ expression from file ('-' for stdin)")
	// Prefer --template (keep --format for backwards compatibility, but hide it to avoid
	// ambiguity with subcommands that use --format for file formats).
	cmd.PersistentFlags().StringVarP(&flags.Template, "template", "t", "", "Go template for custom output (e.g., '{{.ID}}: {{.Status}}')")
	cmd.PersistentFlags().StringVar(&flags.Template, "format", "", "DEPRECATED: use --template")
	_ = cmd.PersistentFlags().MarkDeprecated("format", "use --template instead")
	_ = cmd.PersistentFlags().MarkHidden("format")

	// Agent-friendly flags
	cmd.PersistentFlags().BoolVarP(&flags.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.PersistentFlags().BoolVar(&flags.NoInput, "no-input", false, "Disable interactive prompts")
	cmd.PersistentFlags().BoolVar(&flags.Yes, "force", false, "Alias for --yes")
	cmd.PersistentFlags().BoolVar(&flags.ItemsOnly, "items-only", false, "Output only the items/results array when present (JSON output)")
	cmd.PersistentFlags().BoolVar(&flags.ItemsOnly, "results-only", false, "Alias for --items-only")
	cmd.PersistentFlags().IntVar(&flags.OutputLimit, "output-limit", 0, "Limit number of results in output (0 = no limit)")
	cmd.PersistentFlags().StringVar(&flags.SortBy, "sort-by", "", "Sort results by field")
	cmd.PersistentFlags().BoolVar(&flags.Desc, "desc", false, "Sort in descending order")

	// Multi-letter hidden flag aliases.
	flagAlias(cmd.PersistentFlags(), "output", "out")
	flagAlias(cmd.PersistentFlags(), "query", "qr")
	flagAlias(cmd.PersistentFlags(), "query-file", "qf")
	flagAlias(cmd.PersistentFlags(), "template", "tmpl")
	flagAlias(cmd.PersistentFlags(), "no-color", "nc")
	flagAlias(cmd.PersistentFlags(), "output-limit", "ol")
	flagAlias(cmd.PersistentFlags(), "sort-by", "sb")
	flagAlias(cmd.PersistentFlags(), "account", "acc")
	flagAlias(cmd.PersistentFlags(), "json", "j")
	flagAlias(cmd.PersistentFlags(), "query", "jq")
	flagAlias(cmd.PersistentFlags(), "items-only", "io")
	flagAlias(cmd.PersistentFlags(), "results-only", "ro")

	cmd.AddCommand(newAPICmd())
	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newBalancesCmd())
	cmd.AddCommand(newIssuingCmd())
	// Desire paths: top-level shortcuts to commonly used issuing commands.
	cmd.AddCommand(newCardsCmd())
	cmd.AddCommand(newCardholdersCmd())
	cmd.AddCommand(newTransactionsCmd())
	cmd.AddCommand(newAuthorizationsCmd())
	cmd.AddCommand(newDisputesCmd())
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
	// Desire path: direct resource access by ID.
	cmd.AddCommand(newGetByIDCmd(getClient))
	// Desire paths: verb-first routers.
	cmd.AddCommand(newListRouterCmd())
	cmd.AddCommand(newCreateRouterCmd())
	cmd.AddCommand(newCancelRouterCmd())
	addCanonicalVerbAliases(cmd)

	defaultHelp := cmd.HelpFunc()
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c.Name() == cmd.Name() && !c.HasParent() {
			_, _ = fmt.Fprint(c.OutOrStdout(), helpText)
			return
		}
		defaultHelp(c, args)
	})

	return cmd
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func requireAccount(ctx context.Context) (string, error) {
	f, ok := rootFlagsFromContext(ctx)
	if !ok || f == nil {
		// Allow subcommands to run without root context (tests, embedding).
		f = &rootFlags{Account: os.Getenv("AWX_ACCOUNT")}
	}
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
	agent := isAgentInvocation(args)
	if agent {
		// Avoid Cobra printing raw errors (including flag parse errors) and emit JSON instead.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(ctx)
	if err != nil && agent {
		writeAgentError(ctx, err)
	}
	return err
}

func isAgentInvocation(args []string) bool {
	// Env opt-in for embedded agent runtimes.
	if os.Getenv("AWX_AGENT") != "" {
		return true
	}
	for _, a := range args {
		// Treat explicit false as an override.
		if a == "--agent=false" {
			return false
		}
		if a == "--agent" || a == "--agent=true" {
			return true
		}
	}
	return false
}

func writeAgentError(ctx context.Context, err error) {
	type errObj struct {
		Message    string `json:"message"`
		ExitCode   int    `json:"exit_code"`
		HTTPStatus int    `json:"http_status,omitempty"`
		Request    string `json:"request,omitempty"`
		APIError   string `json:"api_error,omitempty"`
		APISource  string `json:"api_source,omitempty"`
	}

	out := struct {
		Error errObj `json:"error"`
	}{
		Error: errObj{
			Message:  err.Error(),
			ExitCode: exitcode.FromError(err),
		},
	}

	// Best-effort enrichment: keep it stable, small, and machine-readable.
	var ctxErr *api.ContextualError
	if errors.As(err, &ctxErr) && ctxErr != nil {
		out.Error.HTTPStatus = ctxErr.StatusCode
		out.Error.Request = fmt.Sprintf("%s %s", ctxErr.Method, ctxErr.URL)
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		out.Error.APIError = apiErr.Code
		out.Error.APISource = apiErr.Source
	} else if ctxErr != nil && errors.As(ctxErr.Err, &apiErr) && apiErr != nil {
		out.Error.APIError = apiErr.Code
		out.Error.APISource = apiErr.Source
	}

	io := iocontext.GetIO(ctx)
	enc := json.NewEncoder(io.ErrOut)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(out)
}
