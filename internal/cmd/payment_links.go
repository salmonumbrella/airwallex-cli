package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newPaymentLinksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "payment-links",
		Aliases: []string{"pl"},
		Short:   "Payment link operations",
		Long:    "Create and manage payment links for collecting payments.",
	}
	cmd.AddCommand(newPaymentLinksListCmd())
	cmd.AddCommand(newPaymentLinksGetCmd())
	cmd.AddCommand(newPaymentLinksCreateCmd())
	return cmd
}

func newPaymentLinksListCmd() *cobra.Command {
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payment links",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageSize < 10 {
				return fmt.Errorf("--page-size must be at least 10 (got %d)", pageSize)
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListPaymentLinks(cmd.Context(), page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No payment links found")
				return nil
			}

			headers := []string{"ID", "AMOUNT", "CURRENCY", "STATUS", "DESCRIPTION"}
			rowFn := func(item any) []string {
				pl := item.(api.PaymentLink)
				desc := pl.Description
				if len(desc) > 30 {
					desc = desc[:27] + "..."
				}
				return []string{pl.ID, fmt.Sprintf("%.2f", pl.Amount), pl.Currency, pl.Status, desc}
			}

			if err := f.OutputList(result.Items, headers, rowFn); err != nil {
				return err
			}

			if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
	return cmd
}

func newPaymentLinksGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <linkId>",
		Short: "Get payment link details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			pl, err := client.GetPaymentLink(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, pl)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "id\t%s\n", pl.ID)
			_, _ = fmt.Fprintf(tw, "url\t%s\n", pl.URL)
			_, _ = fmt.Fprintf(tw, "amount\t%.2f\n", pl.Amount)
			_, _ = fmt.Fprintf(tw, "currency\t%s\n", pl.Currency)
			if pl.Description != "" {
				_, _ = fmt.Fprintf(tw, "description\t%s\n", pl.Description)
			}
			_, _ = fmt.Fprintf(tw, "status\t%s\n", pl.Status)
			if pl.ExpiresAt != "" {
				_, _ = fmt.Fprintf(tw, "expires_at\t%s\n", pl.ExpiresAt)
			}
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", pl.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
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
