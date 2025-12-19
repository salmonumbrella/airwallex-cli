package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newFXRatesCmd() *cobra.Command {
	var sellCurrency, buyCurrency string

	cmd := &cobra.Command{
		Use:   "rates",
		Short: "Get current exchange rates",
		Long: `Get current exchange rates between currency pairs.

Examples:
  airwallex fx rates --sell USD --buy EUR
  airwallex fx rates --sell USD`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if len(result.Rates) == 0 {
				fmt.Fprintln(os.Stderr, "No rates found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "SELL\tBUY\tRATE\tTYPE")
			for _, r := range result.Rates {
				fmt.Fprintf(tw, "%s\t%s\t%.6f\t%s\n",
					r.SellCurrency, r.BuyCurrency, r.Rate, r.RateType)
			}
			tw.Flush()
			return nil
		},
	}

	cmd.Flags().StringVar(&sellCurrency, "sell", "", "Sell currency (e.g., USD)")
	cmd.Flags().StringVar(&buyCurrency, "buy", "", "Buy currency (e.g., EUR)")
	return cmd
}
