package cmd

import "github.com/spf13/cobra"

func newFXCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fx",
		Short: "Foreign exchange operations",
		Long:  "Manage FX rates, quotes, and currency conversions.",
	}
	cmd.AddCommand(newFXRatesCmd())
	cmd.AddCommand(newFXQuotesCmd())
	cmd.AddCommand(newFXConversionsCmd())
	return cmd
}
