package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size")
	return cmd
}

func newAccountsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <accountId>",
		Short: "Get global account details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			a, err := client.GetGlobalAccount(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, a)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "account_id\t%s\n", a.AccountID)
			_, _ = fmt.Fprintf(tw, "account_name\t%s\n", a.AccountName)
			_, _ = fmt.Fprintf(tw, "currency\t%s\n", a.Currency)
			_, _ = fmt.Fprintf(tw, "country_code\t%s\n", a.CountryCode)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", a.Status)
			if a.AccountNumber != "" {
				_, _ = fmt.Fprintf(tw, "account_number\t%s\n", a.AccountNumber)
			}
			if a.RoutingCode != "" {
				_, _ = fmt.Fprintf(tw, "routing_code\t%s\n", a.RoutingCode)
			}
			if a.IBAN != "" {
				_, _ = fmt.Fprintf(tw, "iban\t%s\n", a.IBAN)
			}
			if a.SwiftCode != "" {
				_, _ = fmt.Fprintf(tw, "swift_code\t%s\n", a.SwiftCode)
			}
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", a.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}
