package cmd

import "github.com/spf13/cobra"

func newIssuingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issuing",
		Short: "Card issuing operations",
	}
	cmd.AddCommand(newCardsCmd())
	cmd.AddCommand(newCardholdersCmd())
	cmd.AddCommand(newTransactionsCmd())
	cmd.AddCommand(newAuthorizationsCmd())
	cmd.AddCommand(newDisputesCmd())
	return cmd
}
