package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
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
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		Long: `List card transactions with optional filters.

Use --output json with --query for advanced filtering using jq syntax.

Examples:
  # List recent transactions
  airwallex issuing transactions list --page-size 20

  # Filter by merchant name (case-insensitive)
  airwallex issuing transactions list --output json --query \
    '[.[] | select(.merchant.name | test("COACH"; "i"))]'

  # Last 10 transactions sorted by date
  airwallex issuing transactions list --output json --query \
    'sort_by(.transaction_date) | reverse | .[0:10]'

  # Top 5 highest spend transactions
  airwallex issuing transactions list --output json --page-size 100 --query \
    'sort_by(.transaction_amount) | .[0:5]'

  # Transactions over $500
  airwallex issuing transactions list --output json --query \
    '[.[] | select(.transaction_amount < -500)]'

  # Declined/failed transactions
  airwallex issuing transactions list --output json --query \
    '[.[] | select(.status != "APPROVED")]'

  # Spend by card (which cards are spending most)
  airwallex issuing transactions list --output json --page-size 100 --query \
    'group_by(.card_nickname) | map({card: .[0].card_nickname, total: (map(.transaction_amount) | add)}) | sort_by(.total)'

  # Top vendors by total spend
  airwallex issuing transactions list --output json --page-size 100 --query \
    'group_by(.merchant.name) | map({vendor: .[0].merchant.name, total: (map(.transaction_amount) | add), count: length}) | sort_by(.total) | .[0:10]'

  # Compact view with selected fields
  airwallex issuing transactions list --output json --query \
    '.[] | {date: .transaction_date[0:10], merchant: .merchant.name, amount: .transaction_amount}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Convert YYYY-MM-DD dates to RFC3339 format
			var fromRFC3339, toRFC3339 string

			if from != "" {
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return fmt.Errorf("invalid --from date: %w", err)
				}
			}

			if to != "" {
				toRFC3339, err = convertDateToRFC3339End(to)
				if err != nil {
					return fmt.Errorf("invalid --to date: %w", err)
				}
			}

			if pageSize < 10 {
				pageSize = 10
			}

			result, err := client.ListTransactions(cmd.Context(), cardID, fromRFC3339, toRFC3339, page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No transactions found")
				return nil
			}

			headers := []string{"TRANSACTION_ID", "TYPE", "AMOUNT", "CURRENCY", "MERCHANT", "STATUS"}
			rowFn := func(item any) []string {
				txn := item.(api.Transaction)
				return []string{txn.TransactionID, txn.TransactionType, fmt.Sprintf("%.2f", txn.Amount), txn.Currency, txn.Merchant.Name, txn.Status}
			}

			if err := f.OutputList(result.Items, headers, rowFn); err != nil {
				return err
			}

			if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
				_, _ = fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&cardID, "card-id", "", "Filter by card ID")
	cmd.Flags().StringVar(&from, "from", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
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

			txn, err := client.GetTransaction(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, txn)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "transaction_id\t%s\n", txn.TransactionID)
			_, _ = fmt.Fprintf(tw, "card_id\t%s\n", txn.CardID)
			_, _ = fmt.Fprintf(tw, "card_nickname\t%s\n", txn.CardNickname)
			_, _ = fmt.Fprintf(tw, "type\t%s\n", txn.TransactionType)
			_, _ = fmt.Fprintf(tw, "amount\t%.2f %s\n", txn.Amount, txn.Currency)
			_, _ = fmt.Fprintf(tw, "billing\t%.2f %s\n", txn.BillingAmount, txn.BillingCurrency)
			_, _ = fmt.Fprintf(tw, "merchant\t%s\n", txn.Merchant.Name)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", txn.Status)
			_, _ = fmt.Fprintf(tw, "date\t%s\n", txn.TransactionDate)
			_ = tw.Flush()
			return nil
		},
	}
}
