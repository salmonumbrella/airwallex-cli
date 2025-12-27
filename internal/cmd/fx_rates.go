package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newFXRatesCmd() *cobra.Command {
	var sellCurrency, buyCurrency string

	cmd := &cobra.Command{
		Use:   "rates",
		Short: "Get current exchange rates",
		Long: `Get current exchange rate between a currency pair.

Both --sell and --buy currencies are required.

Examples:
  airwallex fx rates --sell USD --buy EUR
  airwallex fx rates --sell CAD --buy USD`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Both currencies are required
			if sellCurrency == "" || buyCurrency == "" {
				return fmt.Errorf("both --sell and --buy currencies are required")
			}
			// Validate currencies
			if err := validateCurrency(sellCurrency); err != nil {
				return fmt.Errorf("--sell: %w", err)
			}
			if err := validateCurrency(buyCurrency); err != nil {
				return fmt.Errorf("--buy: %w", err)
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.GetRates(cmd.Context(), sellCurrency, buyCurrency)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Rates) == 0 {
				f.Empty("No rates found")
				return nil
			}

			f.StartTable([]string{"SELL", "BUY", "RATE", "TYPE"})
			for _, r := range result.Rates {
				f.Row(r.SellCurrency, r.BuyCurrency, fmt.Sprintf("%.6f", r.Rate), r.RateType)
			}
			return f.EndTable()
		},
	}

	cmd.Flags().StringVar(&sellCurrency, "sell", "", "Sell currency (e.g., USD)")
	cmd.Flags().StringVar(&buyCurrency, "buy", "", "Buy currency (e.g., EUR)")
	return cmd
}
