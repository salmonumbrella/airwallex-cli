package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newListRouterCmd() *cobra.Command {
	aliases := []string{"ls"}
	cmd := &cobra.Command{
		Use:     "list <resource> [flags]",
		Aliases: aliases,
		Short:   "Desire path: list <resource> (router)",
		Long: `A verb-first desire path for agents.

Examples:
  airwallex list transfers --page-size 5
  airwallex list cards --status ACTIVE
  airwallex list invoices --customer-id cus_123`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := strings.ToLower(args[0])
			rest := args[1:]

			target, ok := listResourceMap()[resource]
			if !ok {
				return fmt.Errorf("unknown resource %q. Try: transfers, beneficiaries, cards, invoices, subscriptions", resource)
			}

			forward := forwardedGlobalArgs(cmd)
			newArgs := append(append(forward, target...), rest...)
			return ExecuteContext(cmd.Context(), newArgs)
		},
	}

	return cmd
}

func newCreateRouterCmd() *cobra.Command {
	aliases := []string{"new", "add"}
	cmd := &cobra.Command{
		Use:     "create <resource> [flags]",
		Aliases: aliases,
		Short:   "Desire path: create <resource> (router)",
		Long: `A verb-first desire path for agents.

Examples:
  airwallex create transfer --beneficiary-id ben_123 --transfer-amount 10 --transfer-currency USD --source-currency USD
  airwallex create webhook --url https://example.com/hook --events transfer.completed
  airwallex create invoice --data '{"customer_id":"cus_123","currency":"USD"}'`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resource := strings.ToLower(args[0])
			rest := args[1:]

			target, ok := createResourceMap()[resource]
			if !ok {
				return fmt.Errorf("unknown resource %q. Try: transfer, beneficiary, card, webhook, invoice, subscription", resource)
			}

			forward := forwardedGlobalArgs(cmd)
			newArgs := append(append(forward, target...), rest...)
			return ExecuteContext(cmd.Context(), newArgs)
		},
	}
	return cmd
}

func newCancelRouterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <id> [flags]",
		Short: "Desire path: cancel by ID (router)",
		Long: `Cancel an operation using only its ID.

This is intentionally opinionated and only routes to commands that are actually
cancelable in the CLI today.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := NormalizeIDArg(args[0])
			rest := args[1:]

			target, err := cancelTargetForID(id)
			if err != nil {
				return err
			}

			forward := forwardedGlobalArgs(cmd)
			newArgs := append(append(forward, target...), append([]string{id}, rest...)...)
			return ExecuteContext(cmd.Context(), newArgs)
		},
	}
	return cmd
}

func listResourceMap() map[string][]string {
	return map[string][]string{
		"transfers":       {"transfers", "list"},
		"transfer":        {"transfers", "list"},
		"payouts":         {"transfers", "list"},
		"payout":          {"transfers", "list"},
		"beneficiaries":   {"beneficiaries", "list"},
		"beneficiary":     {"beneficiaries", "list"},
		"ben":             {"beneficiaries", "list"},
		"linked-accounts": {"linked-accounts", "list"},
		"linked_accounts": {"linked-accounts", "list"},
		"la":              {"linked-accounts", "list"},
		"payment-links":   {"payment-links", "list"},
		"payment_links":   {"payment-links", "list"},
		"pl":              {"payment-links", "list"},
		"webhooks":        {"webhooks", "list"},
		"webhook":         {"webhooks", "list"},
		"wh":              {"webhooks", "list"},
		"deposits":        {"deposits", "list"},
		"deposit":         {"deposits", "list"},
		"dep":             {"deposits", "list"},
		"accounts":        {"accounts", "list"},
		"cards":           {"cards", "list"},
		"cardholders":     {"cardholders", "list"},
		"transactions":    {"transactions", "list"},
		"authorizations":  {"authorizations", "list"},
		"disputes":        {"disputes", "list"},
		"payers":          {"payers", "list"},
		"reports":         {"reports", "list"},

		// Billing convenience (noun-only routes to billing namespace).
		"customers":     {"billing", "customers", "list"},
		"products":      {"billing", "products", "list"},
		"prices":        {"billing", "prices", "list"},
		"invoices":      {"billing", "invoices", "list"},
		"subscriptions": {"billing", "subscriptions", "list"},
	}
}

func createResourceMap() map[string][]string {
	return map[string][]string{
		"transfer":        {"transfers", "create"},
		"transfers":       {"transfers", "create"},
		"beneficiary":     {"beneficiaries", "create"},
		"beneficiaries":   {"beneficiaries", "create"},
		"linked-account":  {"linked-accounts", "create"},
		"linked-accounts": {"linked-accounts", "create"},
		"payment-link":    {"payment-links", "create"},
		"payment-links":   {"payment-links", "create"},
		"webhook":         {"webhooks", "create"},
		"webhooks":        {"webhooks", "create"},
		"card":            {"cards", "create"},
		"cards":           {"cards", "create"},
		"cardholder":      {"cardholders", "create"},
		"cardholders":     {"cardholders", "create"},
		"dispute":         {"disputes", "create"},
		"disputes":        {"disputes", "create"},
		"payer":           {"payers", "create"},
		"payers":          {"payers", "create"},

		// Billing convenience.
		"customer":     {"billing", "customers", "create"},
		"product":      {"billing", "products", "create"},
		"price":        {"billing", "prices", "create"},
		"invoice":      {"billing", "invoices", "create"},
		"subscription": {"billing", "subscriptions", "create"},

		// FX convenience.
		"quote":      {"fx", "quotes", "create"},
		"conversion": {"fx", "conversions", "create"},
	}
}

func cancelTargetForID(id string) ([]string, error) {
	switch {
	case strings.HasPrefix(id, "tfr_"):
		return []string{"transfers", "cancel"}, nil
	case strings.HasPrefix(id, "disp_"):
		return []string{"disputes", "cancel"}, nil
	case strings.HasPrefix(id, "sub_"):
		return []string{"billing", "subscriptions", "cancel"}, nil
	default:
		return nil, fmt.Errorf("don't know how to cancel %q (supported: tfr_, disp_, sub_)", id)
	}
}

func forwardedGlobalArgs(cmd *cobra.Command) []string {
	// Forward only flags that were explicitly set by the user on the outer
	// invocation (especially --account/--json/--agent) so router calls behave
	// identically.
	var out []string
	fs := cmd.InheritedFlags()
	if fs == nil {
		return out
	}

	fs.Visit(func(f *pflag.Flag) {
		if f == nil {
			return
		}
		if f.Name == "help" {
			return
		}
		val := f.Value.String()
		if f.Value.Type() == "bool" {
			if val == "true" {
				out = append(out, "--"+f.Name)
			} else {
				out = append(out, "--"+f.Name+"="+val)
			}
			return
		}
		out = append(out, "--"+f.Name, val)
	})

	// Preserve environment-provided agent mode when present.
	if os.Getenv("AWX_AGENT") != "" {
		// Only add if user didn't already provide --agent (it will be in inherited flags).
		if !containsArg(out, "--agent") {
			out = append(out, "--agent")
		}
	}

	return out
}

func containsArg(args []string, prefix string) bool {
	for _, a := range args {
		if a == prefix || strings.HasPrefix(a, prefix+"=") {
			return true
		}
	}
	return false
}
