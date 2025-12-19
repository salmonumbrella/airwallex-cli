package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

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
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payment links",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageSize < 10 {
				pageSize = 10
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListPaymentLinks(cmd.Context(), 0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No payment links found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tAMOUNT\tCURRENCY\tSTATUS\tDESCRIPTION")
			for _, pl := range result.Items {
				desc := pl.Description
				if len(desc) > 30 {
					desc = desc[:27] + "..."
				}
				fmt.Fprintf(tw, "%s\t%.2f\t%s\t%s\t%s\n",
					pl.ID, pl.Amount, pl.Currency, pl.Status, desc)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results (min 10)")
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
			fmt.Fprintf(tw, "id\t%s\n", pl.ID)
			fmt.Fprintf(tw, "url\t%s\n", pl.URL)
			fmt.Fprintf(tw, "amount\t%.2f\n", pl.Amount)
			fmt.Fprintf(tw, "currency\t%s\n", pl.Currency)
			if pl.Description != "" {
				fmt.Fprintf(tw, "description\t%s\n", pl.Description)
			}
			fmt.Fprintf(tw, "status\t%s\n", pl.Status)
			if pl.ExpiresAt != "" {
				fmt.Fprintf(tw, "expires_at\t%s\n", pl.ExpiresAt)
			}
			fmt.Fprintf(tw, "created_at\t%s\n", pl.CreatedAt)
			tw.Flush()
			return nil
		},
	}
}

func newPaymentLinksCreateCmd() *cobra.Command {
	var amount float64
	var currency, description string
	var expiresIn string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment link",
		Long: `Create a new payment link for collecting payments.

Examples:
  airwallex payment-links create --amount 100 --currency USD
  airwallex payment-links create --amount 50 --currency EUR --description "Invoice #123" --expires-in 7d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			req := map[string]interface{}{
				"amount":   amount,
				"currency": currency,
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
			fmt.Printf("URL: %s\n", pl.URL)
			return nil
		},
	}

	cmd.Flags().Float64Var(&amount, "amount", 0, "Amount to collect (required)")
	cmd.Flags().StringVar(&currency, "currency", "", "Currency (required)")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().StringVar(&expiresIn, "expires-in", "", "Expiration period (e.g., 7d, 24h)")
	mustMarkRequired(cmd, "amount")
	mustMarkRequired(cmd, "currency")
	return cmd
}
