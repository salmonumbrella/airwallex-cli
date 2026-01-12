package cmd

import (
	"context"
	"fmt"
	"os"

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
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List global accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			pageSize = normalizePageSize(pageSize)

			result, err := client.ListGlobalAccounts(cmd.Context(), page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No global accounts found")
				return nil
			}

			headers := []string{"ACCOUNT_ID", "NAME", "CURRENCY", "COUNTRY", "STATUS"}
			colTypes := []outfmt.ColumnType{
				outfmt.ColumnPlain,    // ACCOUNT_ID
				outfmt.ColumnPlain,    // NAME
				outfmt.ColumnCurrency, // CURRENCY
				outfmt.ColumnPlain,    // COUNTRY
				outfmt.ColumnStatus,   // STATUS
			}
			rowFn := func(item any) []string {
				a := item.(api.GlobalAccount)
				return []string{a.AccountID, a.AccountName, a.Currency, a.CountryCode, a.Status}
			}

			if err := f.OutputListWithColors(result.Items, headers, colTypes, rowFn); err != nil {
				return err
			}

			if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
	return cmd
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
