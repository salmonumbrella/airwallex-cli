package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "Transaction management",
	}
	cmd.AddCommand(newTransactionsListCmd())
	cmd.AddCommand(newTransactionsGetCmd())
	return cmd
}

func newTransactionsListCmd() *cobra.Command {
	var cardID string
	var from string
	var to string
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListTransactions(cardID, from, to, 0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No transactions found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "TRANSACTION_ID\tTYPE\tAMOUNT\tCURRENCY\tMERCHANT\tSTATUS")
			for _, txn := range result.Items {
				fmt.Fprintf(tw, "%s\t%s\t%.2f\t%s\t%s\t%s\n",
					txn.TransactionID, txn.TransactionType, txn.Amount, txn.Currency, txn.Merchant.Name, txn.Status)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&cardID, "card-id", "", "Filter by card ID")
	cmd.Flags().StringVar(&from, "from", "", "From date (RFC3339)")
	cmd.Flags().StringVar(&to, "to", "", "To date (RFC3339)")
	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results")
	return cmd
}

func newTransactionsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <transactionId>",
		Short: "Get transaction details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			txn, err := client.GetTransaction(args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, txn)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "transaction_id\t%s\n", txn.TransactionID)
			fmt.Fprintf(tw, "card_id\t%s\n", txn.CardID)
			fmt.Fprintf(tw, "card_nickname\t%s\n", txn.CardNickname)
			fmt.Fprintf(tw, "type\t%s\n", txn.TransactionType)
			fmt.Fprintf(tw, "amount\t%.2f %s\n", txn.Amount, txn.Currency)
			fmt.Fprintf(tw, "billing\t%.2f %s\n", txn.BillingAmount, txn.BillingCurrency)
			fmt.Fprintf(tw, "merchant\t%s\n", txn.Merchant.Name)
			fmt.Fprintf(tw, "status\t%s\n", txn.Status)
			fmt.Fprintf(tw, "date\t%s\n", txn.TransactionDate)
			tw.Flush()
			return nil
		},
	}
}
