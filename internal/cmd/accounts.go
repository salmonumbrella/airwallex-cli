package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Global account management",
	}
	cmd.AddCommand(newAccountsListCmd())
	cmd.AddCommand(newAccountsGetCmd())
	return cmd
}

func newAccountsListCmd() *cobra.Command {
	return NewListCommand(ListConfig[api.GlobalAccount]{
		Use:          "list",
		Short:        "List global accounts",
		Headers:      []string{"ACCOUNT_ID", "NAME", "CURRENCY", "COUNTRY", "STATUS"},
		EmptyMessage: "No global accounts found",
		ColumnTypes: []outfmt.ColumnType{
			outfmt.ColumnPlain,    // ACCOUNT_ID
			outfmt.ColumnPlain,    // NAME
			outfmt.ColumnCurrency, // CURRENCY
			outfmt.ColumnPlain,    // COUNTRY
			outfmt.ColumnStatus,   // STATUS
		},
		RowFunc: func(a api.GlobalAccount) []string {
			return []string{a.AccountID, a.AccountName, a.Currency, a.CountryCode, a.Status}
		},
		MoreHint: "# More results available",
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.GlobalAccount], error) {
			result, err := client.ListGlobalAccounts(ctx, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.GlobalAccount]{}, err
			}
			return ListResult[api.GlobalAccount]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newAccountsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.GlobalAccount]{
		Use:   "get <accountId>",
		Short: "Get global account details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.GlobalAccount, error) {
			return client.GetGlobalAccount(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, a *api.GlobalAccount) error {
			rows := []outfmt.KV{
				{Key: "account_id", Value: a.AccountID},
				{Key: "account_name", Value: a.AccountName},
				{Key: "currency", Value: a.Currency},
				{Key: "country_code", Value: a.CountryCode},
				{Key: "status", Value: a.Status},
			}
			if a.AccountNumber != "" {
				rows = append(rows, outfmt.KV{Key: "account_number", Value: a.AccountNumber})
			}
			if a.RoutingCode != "" {
				rows = append(rows, outfmt.KV{Key: "routing_code", Value: a.RoutingCode})
			}
			if a.IBAN != "" {
				rows = append(rows, outfmt.KV{Key: "iban", Value: a.IBAN})
			}
			if a.SwiftCode != "" {
				rows = append(rows, outfmt.KV{Key: "swift_code", Value: a.SwiftCode})
			}
			rows = append(rows, outfmt.KV{Key: "created_at", Value: a.CreatedAt})
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}
