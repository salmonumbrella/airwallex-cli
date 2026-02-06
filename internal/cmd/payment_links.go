package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newPaymentLinksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "payment-links",
		Aliases: []string{"pl", "payment_links", "paymentlinks", "paylinks", "paylink"},
		Short:   "Payment link operations",
		Long:    "Create and manage payment links for collecting payments.",
	}
	cmd.AddCommand(newPaymentLinksListCmd())
	cmd.AddCommand(newPaymentLinksGetCmd())
	cmd.AddCommand(newPaymentLinksCreateCmd())
	return cmd
}

func newPaymentLinksListCmd() *cobra.Command {
	cmd := NewListCommand(ListConfig[api.PaymentLink]{
		Use:          "list",
		Short:        "List payment links",
		Headers:      []string{"ID", "AMOUNT", "CURRENCY", "STATUS", "DESCRIPTION"},
		EmptyMessage: "No payment links found",
		RowFunc: func(pl api.PaymentLink) []string {
			desc := pl.Description
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}
			return []string{pl.ID, outfmt.FormatMoney(pl.Amount), pl.Currency, pl.Status, desc}
		},
		MoreHint: "# More results available",
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.PaymentLink], error) {
			result, err := client.ListPaymentLinks(ctx, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.PaymentLink]{}, err
			}
			return ListResult[api.PaymentLink]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
	return cmd
}

func newPaymentLinksGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.PaymentLink]{
		Use:   "get <linkId>",
		Short: "Get payment link details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.PaymentLink, error) {
			return client.GetPaymentLink(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, pl *api.PaymentLink) error {
			rows := []outfmt.KV{
				{Key: "id", Value: pl.ID},
				{Key: "url", Value: pl.URL},
				{Key: "amount", Value: outfmt.FormatMoney(pl.Amount)},
				{Key: "currency", Value: pl.Currency},
				{Key: "status", Value: pl.Status},
				{Key: "created_at", Value: pl.CreatedAt},
			}
			if pl.Description != "" {
				rows = append(rows, outfmt.KV{Key: "description", Value: pl.Description})
			}
			if pl.ExpiresAt != "" {
				rows = append(rows, outfmt.KV{Key: "expires_at", Value: pl.ExpiresAt})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newPaymentLinksCreateCmd() *cobra.Command {
	var amount float64
	var currency, description, title string
	var expiresIn string
	var reusable bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment link",
		Long: `Create a new payment link for collecting payments.

Examples:
  airwallex payment-links create --amount 100 --currency USD
  airwallex payment-links create --amount 50 --currency EUR --description "Invoice #123" --expires-in 7d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if err := validateAmount(amount); err != nil {
				return err
			}
			if err := validateCurrency(currency); err != nil {
				return err
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Use description as title if title not provided
			effectiveTitle := title
			if effectiveTitle == "" && description != "" {
				effectiveTitle = description
			}
			if effectiveTitle == "" {
				effectiveTitle = fmt.Sprintf("Payment of %.2f %s", amount, currency)
			}

			req := map[string]interface{}{
				"amount":   amount,
				"currency": currency,
				"title":    effectiveTitle,
				"reusable": reusable,
			}
			if description != "" {
				req["description"] = description
			}
			if expiresIn != "" {
				req["expires_in"] = expiresIn
			}

			pl, err := client.CreatePaymentLink(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, pl)
			}

			u.Success(fmt.Sprintf("Created payment link: %s", pl.ID))
			_, _ = fmt.Fprintf(os.Stdout, "URL: %s\n", pl.URL)
			return nil
		},
	}

	cmd.Flags().Float64Var(&amount, "amount", 0, "Amount to collect (required)")
	cmd.Flags().StringVar(&currency, "currency", "", "Currency (required)")
	cmd.Flags().StringVar(&title, "title", "", "Title for the payment link (defaults to description or auto-generated)")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().BoolVar(&reusable, "reusable", false, "Allow the link to be used multiple times")
	cmd.Flags().StringVar(&expiresIn, "expires-in", "", "Expiration period (e.g., 7d, 24h)")
	mustMarkRequired(cmd, "amount")
	mustMarkRequired(cmd, "currency")
	return cmd
}
