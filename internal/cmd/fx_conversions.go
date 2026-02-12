package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newFXConversionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "conversions",
		Aliases: []string{"conversion", "conv", "cv"},
		Short:   "Manage currency conversions",
	}
	cmd.AddCommand(newFXConversionsListCmd())
	cmd.AddCommand(newFXConversionsGetCmd())
	cmd.AddCommand(newFXConversionsCreateCmd())
	return cmd
}

func newFXConversionsListCmd() *cobra.Command {
	var status, fromDate, toDate string
	cmd := NewListCommand(ListConfig[api.Conversion]{
		Use:          "list",
		Aliases:      []string{"ls", "l"},
		Short:        "List conversions",
		Headers:      []string{"CONVERSION_ID", "SELL", "BUY", "RATE", "STATUS"},
		EmptyMessage: "No conversions found",
		RowFunc: func(c api.Conversion) []string {
			return []string{
				c.ID,
				outfmt.FormatMoney(c.SellAmount) + " " + c.SellCurrency,
				outfmt.FormatMoney(c.BuyAmount) + " " + c.BuyCurrency,
				outfmt.FormatRate(c.Rate),
				c.Status,
			}
		},
		MoreHint: "# More results available",
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Conversion], error) {
			if err := validateDateRangeFlags(fromDate, toDate, "--from", "--to", true); err != nil {
				return ListResult[api.Conversion]{}, err
			}

			result, err := client.ListConversions(ctx, status, fromDate, toDate, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.Conversion]{}, err
			}
			return ListResult[api.Conversion]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status")
	cmd.Flags().StringVarP(&fromDate, "from", "f", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toDate, "to", "", "To date (YYYY-MM-DD)")
	return cmd
}

func newFXConversionsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Conversion]{
		Use:     "get <conversionId>",
		Aliases: []string{"g"},
		Short:   "Get conversion details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Conversion, error) {
			return client.GetConversion(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, conv *api.Conversion) error {
			rows := []outfmt.KV{
				{Key: "conversion_id", Value: conv.ID},
				{Key: "sell_currency", Value: conv.SellCurrency},
				{Key: "buy_currency", Value: conv.BuyCurrency},
				{Key: "sell_amount", Value: outfmt.FormatMoney(conv.SellAmount)},
				{Key: "buy_amount", Value: outfmt.FormatMoney(conv.BuyAmount)},
				{Key: "rate", Value: outfmt.FormatRate(conv.Rate)},
				{Key: "status", Value: conv.Status},
				{Key: "created_at", Value: conv.CreatedAt},
			}
			if conv.QuoteID != "" {
				rows = append(rows, outfmt.KV{Key: "quote_id", Value: conv.QuoteID})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newFXConversionsCreateCmd() *cobra.Command {
	var sellCurrency, buyCurrency string
	var sellAmount, buyAmount float64
	var quoteID string

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Execute a currency conversion",
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

			req := map[string]interface{}{
				"request_id": uuid.New().String(),
			}

			if quoteID != "" {
				// Using a quote - just need the quote ID
				req["quote_id"] = quoteID
			} else {
				// Market rate conversion - validate currencies
				if err := validateCurrency(sellCurrency); err != nil {
					return fmt.Errorf("--sell-currency: %w", err)
				}
				if err := validateCurrency(buyCurrency); err != nil {
					return fmt.Errorf("--buy-currency: %w", err)
				}

				hasSellAmount := sellAmount > 0
				hasBuyAmount := buyAmount > 0
				if hasSellAmount == hasBuyAmount {
					if !hasSellAmount {
						return fmt.Errorf("must provide --quote-id OR (--sell-currency, --buy-currency, and one of --sell-amount/--buy-amount)")
					}
					return fmt.Errorf("cannot provide both --sell-amount and --buy-amount")
				}

				// Validate the provided amount
				if hasSellAmount {
					if err := validateAmount(sellAmount); err != nil {
						return fmt.Errorf("--sell-amount: %w", err)
					}
				}
				if hasBuyAmount {
					if err := validateAmount(buyAmount); err != nil {
						return fmt.Errorf("--buy-amount: %w", err)
					}
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
			rows := []outfmt.KV{
				{Key: "conversion_id", Value: conv.ID},
				{Key: "sold", Value: outfmt.FormatMoney(conv.SellAmount) + " " + conv.SellCurrency},
				{Key: "bought", Value: outfmt.FormatMoney(conv.BuyAmount) + " " + conv.BuyCurrency},
				{Key: "rate", Value: outfmt.FormatRate(conv.Rate)},
				{Key: "status", Value: conv.Status},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}

	cmd.Flags().StringVar(&quoteID, "quote-id", "", "Use a locked quote")
	cmd.Flags().StringVar(&sellCurrency, "sell-currency", "", "Currency to sell")
	cmd.Flags().StringVar(&buyCurrency, "buy-currency", "", "Currency to buy")
	cmd.Flags().Float64Var(&sellAmount, "sell-amount", 0, "Amount to sell")
	cmd.Flags().Float64Var(&buyAmount, "buy-amount", 0, "Amount to buy")
	return cmd
}
