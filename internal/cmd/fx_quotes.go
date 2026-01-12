package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newFXQuotesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quotes",
		Short: "Manage FX quotes",
	}
	cmd.AddCommand(newFXQuotesCreateCmd())
	cmd.AddCommand(newFXQuotesGetCmd())
	return cmd
}

func newFXQuotesCreateCmd() *cobra.Command {
	var sellCurrency, buyCurrency string
	var sellAmount, buyAmount float64
	var validity string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a quote to lock in an exchange rate",
		Long: `Create an FX quote to lock in an exchange rate for a period of time.

Examples:
  # Lock rate for 1 hour, specifying sell amount
  airwallex fx quotes create --sell-currency USD --buy-currency EUR --sell-amount 10000 --validity 1h

  # Lock rate for 24 hours, specifying buy amount
  airwallex fx quotes create --sell-currency USD --buy-currency EUR --buy-amount 9000 --validity 24h

Validity periods: 1m, 5m, 15m, 30m, 1h, 2h, 4h, 12h, 24h`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check required currencies (supports both --sell-currency and --sell aliases)
			if sellCurrency == "" {
				return fmt.Errorf("--sell-currency is required (or use --sell)")
			}
			if buyCurrency == "" {
				return fmt.Errorf("--buy-currency is required (or use --buy)")
			}
			// Validate currency format
			if err := validateCurrency(sellCurrency); err != nil {
				return fmt.Errorf("--sell-currency: %w", err)
			}
			if err := validateCurrency(buyCurrency); err != nil {
				return fmt.Errorf("--buy-currency: %w", err)
			}

			// Validate: exactly one of sell_amount or buy_amount
			hasSellAmount := sellAmount > 0
			hasBuyAmount := buyAmount > 0
			if hasSellAmount == hasBuyAmount {
				if !hasSellAmount {
					return fmt.Errorf("must provide exactly one of --sell-amount or --buy-amount")
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

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Convert user-friendly validity to API format (MIN_X, HR_X)
			// Note: Airwallex API uses HR_X (not HOUR_X) for hour durations
			validityMap := map[string]string{
				"1m":  "MIN_1",
				"5m":  "MIN_5",
				"15m": "MIN_15",
				"30m": "MIN_30",
				"1h":  "HR_1",
				"2h":  "HR_2",
				"4h":  "HR_4",
				"8h":  "HR_8",
				"12h": "HR_12",
				"24h": "HR_24",
			}
			apiValidity, ok := validityMap[validity]
			if !ok {
				// Allow pass-through for raw API formats like MIN_15, HR_1
				apiValidity = validity
			}

			req := map[string]interface{}{
				"request_id":    uuid.New().String(),
				"sell_currency": sellCurrency,
				"buy_currency":  buyCurrency,
				"validity":      apiValidity,
			}
			if sellAmount > 0 {
				req["sell_amount"] = sellAmount
			}
			if buyAmount > 0 {
				req["buy_amount"] = buyAmount
			}

			quote, err := client.CreateQuote(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, quote)
			}

			u.Success(fmt.Sprintf("Created quote: %s (expires: %s)", quote.ID, quote.RateExpiry))
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "quote_id\t%s\n", quote.ID)
			_, _ = fmt.Fprintf(tw, "sell_currency\t%s\n", quote.SellCurrency)
			_, _ = fmt.Fprintf(tw, "buy_currency\t%s\n", quote.BuyCurrency)
			_, _ = fmt.Fprintf(tw, "sell_amount\t%.2f\n", quote.SellAmount)
			_, _ = fmt.Fprintf(tw, "buy_amount\t%.2f\n", quote.BuyAmount)
			_, _ = fmt.Fprintf(tw, "rate\t%.6f\n", quote.Rate)
			_, _ = fmt.Fprintf(tw, "expires\t%s\n", quote.RateExpiry)
			_ = tw.Flush()
			return nil
		},
	}

	cmd.Flags().StringVar(&sellCurrency, "sell-currency", "", "Currency to sell (required)")
	cmd.Flags().StringVar(&buyCurrency, "buy-currency", "", "Currency to buy (required)")
	// Add short aliases for consistency with fx rates command
	cmd.Flags().StringVar(&sellCurrency, "sell", "", "Currency to sell (alias for --sell-currency)")
	cmd.Flags().StringVar(&buyCurrency, "buy", "", "Currency to buy (alias for --buy-currency)")
	_ = cmd.Flags().MarkHidden("sell") // Hide from help to avoid confusion
	_ = cmd.Flags().MarkHidden("buy")  // Primary flags are --sell-currency/--buy-currency
	cmd.Flags().Float64Var(&sellAmount, "sell-amount", 0, "Amount to sell")
	cmd.Flags().Float64Var(&buyAmount, "buy-amount", 0, "Amount to buy")
	cmd.Flags().StringVar(&validity, "validity", "1h", "Quote validity period (1m, 5m, 1h, 24h)")
	return cmd
}

func newFXQuotesGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Quote]{
		Use:   "get <quoteId>",
		Short: "Get quote details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Quote, error) {
			return client.GetQuote(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, quote *api.Quote) error {
			rows := []outfmt.KV{
				{Key: "quote_id", Value: quote.ID},
				{Key: "sell_currency", Value: quote.SellCurrency},
				{Key: "buy_currency", Value: quote.BuyCurrency},
				{Key: "sell_amount", Value: fmt.Sprintf("%.2f", quote.SellAmount)},
				{Key: "buy_amount", Value: fmt.Sprintf("%.2f", quote.BuyAmount)},
				{Key: "rate", Value: fmt.Sprintf("%.6f", quote.Rate)},
				{Key: "status", Value: quote.Status},
				{Key: "expires", Value: quote.RateExpiry},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}
