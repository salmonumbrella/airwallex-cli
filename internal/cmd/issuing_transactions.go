package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transactions",
		Aliases: []string{"transaction", "tx"},
		Short:   "Transaction management",
	}
	cmd.AddCommand(newTransactionsListCmd())
	cmd.AddCommand(newTransactionsGetCmd())
	return cmd
}

func newTransactionsListCmd() *cobra.Command {
	var cardID string
	var from string
	var to string
	cmd := NewListCommand(ListConfig[api.Transaction]{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List transactions",
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
		Headers:      []string{"TRANSACTION_ID", "TYPE", "AMOUNT", "CURRENCY", "MERCHANT", "STATUS"},
		EmptyMessage: "No transactions found",
		RowFunc: func(txn api.Transaction) []string {
			return []string{txn.TransactionID, txn.TransactionType, outfmt.FormatMoney(txn.Amount), txn.Currency, txn.Merchant.Name, txn.Status}
		},
		IDFunc: func(txn api.Transaction) string {
			return txn.TransactionID
		},
		LightFunc: func(txn api.Transaction) any { return toLightTransaction(txn) },
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Transaction], error) {
			fromRFC3339, toRFC3339, err := parseDateRangeRFC3339(from, to, "--from", "--to", false)
			if err != nil {
				return ListResult[api.Transaction]{}, err
			}

			result, err := client.ListTransactions(ctx, cardID, fromRFC3339, toRFC3339, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.Transaction]{}, err
			}

			return ListResult[api.Transaction]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&cardID, "card-id", "", "Filter by card ID")
	cmd.Flags().StringVarP(&from, "from", "f", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To date (YYYY-MM-DD)")
	flagAlias(cmd.Flags(), "card-id", "cid")
	flagAlias(cmd.Flags(), "from", "fr")
	return cmd
}

func newTransactionsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Transaction]{
		Use:     "get <transactionId>",
		Aliases: []string{"g"},
		Short:   "Get transaction details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Transaction, error) {
			return client.GetTransaction(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, txn *api.Transaction) error {
			rows := []outfmt.KV{
				{Key: "transaction_id", Value: txn.TransactionID},
				{Key: "card_id", Value: txn.CardID},
				{Key: "card_nickname", Value: txn.CardNickname},
				{Key: "type", Value: txn.TransactionType},
				{Key: "amount", Value: outfmt.FormatMoney(txn.Amount) + " " + txn.Currency},
				{Key: "billing", Value: outfmt.FormatMoney(txn.BillingAmount) + " " + txn.BillingCurrency},
				{Key: "merchant", Value: txn.Merchant.Name},
				{Key: "status", Value: txn.Status},
				{Key: "date", Value: txn.TransactionDate},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}
