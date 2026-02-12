package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balances",
		Aliases: []string{"balance", "bal", "b"},
		Short:   "Balance operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior: show current balances
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			balances, err := client.GetBalances(cmd.Context())
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if outfmt.IsJSON(cmd.Context()) {
				return f.Output(balances)
			}

			if len(balances.Balances) == 0 {
				f.Empty("No balances")
				return nil
			}

			f.StartTable([]string{"CURRENCY", "AVAILABLE", "PENDING", "RESERVED", "TOTAL"})
			// Column types: currency, amount, amount, amount, amount
			colTypes := []outfmt.ColumnType{
				outfmt.ColumnCurrency,
				outfmt.ColumnAmount,
				outfmt.ColumnAmount,
				outfmt.ColumnAmount,
				outfmt.ColumnAmount,
			}
			for _, b := range balances.Balances {
				f.ColorRow(colTypes,
					b.Currency,
					outfmt.FormatMoney(b.AvailableAmount),
					outfmt.FormatMoney(b.PendingAmount),
					outfmt.FormatMoney(b.ReservedAmount),
					outfmt.FormatMoney(b.TotalAmount))
			}
			return f.EndTable()
		},
	}
	cmd.AddCommand(newBalancesHistoryCmd())
	return cmd
}

func newBalancesHistoryCmd() *cobra.Command {
	var currency string
	var from string
	var to string

	cmd := NewListCommand(ListConfig[api.BalanceHistoryItem]{
		Use:     "history",
		Aliases: []string{"hist", "h"},
		Short:   "Show balance transaction history",
		Long: `Show balance transaction history.

Date range is limited to 7 days maximum per query.
Dates should be in YYYY-MM-DD format and will be converted to RFC3339.

Examples:
  # Show all balance history
  airwallex balances history

  # Filter by currency
  airwallex balances history --currency CAD

  # Filter by date range (max 7 days)
  airwallex balances history --from 2024-01-01 --to 2024-01-07

  # Combine filters with custom limit
  airwallex balances history --currency USD --from 2024-01-01 --to 2024-01-07 --page-size 50`,
		Headers:      []string{"ID", "CURRENCY", "AMOUNT", "BALANCE", "TYPE", "POSTED_AT", "DESCRIPTION"},
		EmptyMessage: "No balance history found",
		RowFunc: func(item api.BalanceHistoryItem) []string {
			return []string{
				item.ID,
				item.Currency,
				outfmt.FormatMoney(item.Amount),
				outfmt.FormatMoney(item.Balance),
				item.TransactionType,
				item.PostedAt,
				item.Description,
			}
		},
		IDFunc: func(item api.BalanceHistoryItem) string {
			return item.ID
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BalanceHistoryItem], error) {
			// Validate date inputs
			if err := validateDate(from); err != nil {
				return ListResult[api.BalanceHistoryItem]{}, fmt.Errorf("--from: %w", err)
			}
			if err := validateDate(to); err != nil {
				return ListResult[api.BalanceHistoryItem]{}, fmt.Errorf("--to: %w", err)
			}
			if err := validateDateRange(from, to); err != nil {
				return ListResult[api.BalanceHistoryItem]{}, err
			}

			// Convert YYYY-MM-DD dates to RFC3339 format
			var fromRFC3339, toRFC3339 string
			var err error

			if from != "" {
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return ListResult[api.BalanceHistoryItem]{}, fmt.Errorf("invalid --from date: %w", err)
				}
			}

			if to != "" {
				toRFC3339, err = convertDateToRFC3339End(to)
				if err != nil {
					return ListResult[api.BalanceHistoryItem]{}, fmt.Errorf("invalid --to date: %w", err)
				}
			}

			// Validate date range (max 7 days)
			if fromRFC3339 != "" && toRFC3339 != "" {
				fromTime, _ := time.Parse(time.RFC3339, fromRFC3339)
				toTime, _ := time.Parse(time.RFC3339, toRFC3339)
				if toTime.Sub(fromTime) > 7*24*time.Hour {
					return ListResult[api.BalanceHistoryItem]{}, fmt.Errorf("date range exceeds 7 days maximum")
				}
			}

			result, err := client.GetBalanceHistory(ctx, currency, fromRFC3339, toRFC3339, 0, opts.Limit)
			if err != nil {
				return ListResult[api.BalanceHistoryItem]{}, err
			}

			return ListResult[api.BalanceHistoryItem]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVarP(&currency, "currency", "c", "", "Filter by currency (e.g., CAD, USD)")
	cmd.Flags().StringVarP(&from, "from", "f", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "End date (YYYY-MM-DD)")

	return cmd
}
