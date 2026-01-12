package cmd

import (
	"context"
	"fmt"
	"os"

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

			pageSize = normalizePageSize(pageSize)

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
	return NewGetCommand(GetConfig[*api.Transaction]{
		Use:   "get <transactionId>",
		Short: "Get transaction details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Transaction, error) {
			return client.GetTransaction(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, txn *api.Transaction) error {
			rows := []outfmt.KV{
				{Key: "transaction_id", Value: txn.TransactionID},
				{Key: "card_id", Value: txn.CardID},
				{Key: "card_nickname", Value: txn.CardNickname},
				{Key: "type", Value: txn.TransactionType},
				{Key: "amount", Value: fmt.Sprintf("%.2f %s", txn.Amount, txn.Currency)},
				{Key: "billing", Value: fmt.Sprintf("%.2f %s", txn.BillingAmount, txn.BillingCurrency)},
				{Key: "merchant", Value: txn.Merchant.Name},
				{Key: "status", Value: txn.Status},
				{Key: "date", Value: txn.TransactionDate},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}
