package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newFXConversionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conversions",
		Short: "Manage currency conversions",
	}
	cmd.AddCommand(newFXConversionsListCmd())
	cmd.AddCommand(newFXConversionsGetCmd())
	cmd.AddCommand(newFXConversionsCreateCmd())
	return cmd
}

func newFXConversionsListCmd() *cobra.Command {
	var status, fromDate, toDate string
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List conversions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageSize < 10 {
				pageSize = 10
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListConversions(cmd.Context(), status, fromDate, toDate, 0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No conversions found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "CONVERSION_ID\tSELL\tBUY\tRATE\tSTATUS")
			for _, c := range result.Items {
				fmt.Fprintf(tw, "%s\t%.2f %s\t%.2f %s\t%.6f\t%s\n",
					c.ID, c.SellAmount, c.SellCurrency, c.BuyAmount, c.BuyCurrency, c.Rate, c.Status)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&fromDate, "from", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toDate, "to", "", "To date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results (min 10)")
	return cmd
}

func newFXConversionsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <conversionId>",
		Short: "Get conversion details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			conv, err := client.GetConversion(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, conv)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "conversion_id\t%s\n", conv.ID)
			if conv.QuoteID != "" {
				fmt.Fprintf(tw, "quote_id\t%s\n", conv.QuoteID)
			}
			fmt.Fprintf(tw, "sell_currency\t%s\n", conv.SellCurrency)
			fmt.Fprintf(tw, "buy_currency\t%s\n", conv.BuyCurrency)
			fmt.Fprintf(tw, "sell_amount\t%.2f\n", conv.SellAmount)
			fmt.Fprintf(tw, "buy_amount\t%.2f\n", conv.BuyAmount)
			fmt.Fprintf(tw, "rate\t%.6f\n", conv.Rate)
			fmt.Fprintf(tw, "status\t%s\n", conv.Status)
			fmt.Fprintf(tw, "created_at\t%s\n", conv.CreatedAt)
			tw.Flush()
			return nil
		},
	}
}

func newFXConversionsCreateCmd() *cobra.Command {
	var sellCurrency, buyCurrency string
	var sellAmount, buyAmount float64
	var quoteID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Execute a currency conversion",
		Long: `Execute a currency conversion, optionally using a locked quote.

Examples:
  # Convert at market rate
  airwallex fx conversions create --sell-currency USD --buy-currency EUR --sell-amount 10000

  # Convert using a locked quote
  airwallex fx conversions create --quote-id qt_xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			req := map[string]interface{}{}

			if quoteID != "" {
				// Using a quote - just need the quote ID
				req["quote_id"] = quoteID
			} else {
				// Market rate conversion
				hasSellAmount := sellAmount > 0
				hasBuyAmount := buyAmount > 0
				if hasSellAmount == hasBuyAmount {
					if !hasSellAmount {
						return fmt.Errorf("must provide --quote-id OR (--sell-currency, --buy-currency, and one of --sell-amount/--buy-amount)")
					}
					return fmt.Errorf("cannot provide both --sell-amount and --buy-amount")
				}

				req["sell_currency"] = sellCurrency
				req["buy_currency"] = buyCurrency
				if sellAmount > 0 {
					req["sell_amount"] = sellAmount
				}
				if buyAmount > 0 {
					req["buy_amount"] = buyAmount
				}
			}

			conv, err := client.CreateConversion(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, conv)
			}

			u.Success(fmt.Sprintf("Conversion executed: %s", conv.ID))
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "conversion_id\t%s\n", conv.ID)
			fmt.Fprintf(tw, "sold\t%.2f %s\n", conv.SellAmount, conv.SellCurrency)
			fmt.Fprintf(tw, "bought\t%.2f %s\n", conv.BuyAmount, conv.BuyCurrency)
			fmt.Fprintf(tw, "rate\t%.6f\n", conv.Rate)
			fmt.Fprintf(tw, "status\t%s\n", conv.Status)
			tw.Flush()
			return nil
		},
	}

	cmd.Flags().StringVar(&quoteID, "quote-id", "", "Use a locked quote")
	cmd.Flags().StringVar(&sellCurrency, "sell-currency", "", "Currency to sell")
	cmd.Flags().StringVar(&buyCurrency, "buy-currency", "", "Currency to buy")
	cmd.Flags().Float64Var(&sellAmount, "sell-amount", 0, "Amount to sell")
	cmd.Flags().Float64Var(&buyAmount, "buy-amount", 0, "Amount to buy")
	return cmd
}
