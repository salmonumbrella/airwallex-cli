package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newDepositsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits",
		Short: "Deposit operations",
		Long:  "View and track inbound deposits from bank transfers and linked accounts.",
	}
	cmd.AddCommand(newDepositsListCmd())
	cmd.AddCommand(newDepositsGetCmd())
	return cmd
}

func newDepositsListCmd() *cobra.Command {
	var status, fromDate, toDate string
	cmd := NewListCommand(ListConfig[api.Deposit]{
		Use:   "list",
		Short: "List deposits",
		Long: `List inbound deposits with optional filters.

Examples:
  airwallex deposits list
  airwallex deposits list --status SETTLED
  airwallex deposits list --from 2024-01-01 --to 2024-01-31`,
		Headers:      []string{"DEPOSIT_ID", "AMOUNT", "CURRENCY", "STATUS", "SOURCE", "CREATED"},
		EmptyMessage: "No deposits found",
		ColumnTypes: []outfmt.ColumnType{
			outfmt.ColumnPlain,    // DEPOSIT_ID
			outfmt.ColumnAmount,   // AMOUNT
			outfmt.ColumnCurrency, // CURRENCY
			outfmt.ColumnStatus,   // STATUS
			outfmt.ColumnPlain,    // SOURCE
			outfmt.ColumnPlain,    // CREATED
		},
		RowFunc: func(d api.Deposit) []string {
			return []string{d.ID, fmt.Sprintf("%.2f", d.Amount), d.Currency, d.Status, d.Source, d.CreatedAt}
		},
		MoreHint: "# More results available",
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Deposit], error) {
			if err := validateDateRangeFlags(fromDate, toDate, "--from", "--to", true); err != nil {
				return ListResult[api.Deposit]{}, err
			}

			result, err := client.ListDeposits(ctx, status, fromDate, toDate, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.Deposit]{}, err
			}

			return ListResult[api.Deposit]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (PENDING, SETTLED, FAILED)")
	cmd.Flags().StringVar(&fromDate, "from", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toDate, "to", "", "To date (YYYY-MM-DD)")
	return cmd
}

func newDepositsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Deposit]{
		Use:   "get <depositId>",
		Short: "Get deposit details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Deposit, error) {
			return client.GetDeposit(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, d *api.Deposit) error {
			rows := []outfmt.KV{
				{Key: "deposit_id", Value: d.ID},
				{Key: "amount", Value: fmt.Sprintf("%.2f", d.Amount)},
				{Key: "currency", Value: d.Currency},
				{Key: "status", Value: d.Status},
				{Key: "source", Value: d.Source},
				{Key: "created_at", Value: d.CreatedAt},
			}
			if d.LinkedAccountID != "" {
				rows = append(rows, outfmt.KV{Key: "linked_account_id", Value: d.LinkedAccountID})
			}
			if d.GlobalAccountID != "" {
				rows = append(rows, outfmt.KV{Key: "global_account_id", Value: d.GlobalAccountID})
			}
			if d.Reference != "" {
				rows = append(rows, outfmt.KV{Key: "reference", Value: d.Reference})
			}
			if d.SettledAt != "" {
				rows = append(rows, outfmt.KV{Key: "settled_at", Value: d.SettledAt})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}
