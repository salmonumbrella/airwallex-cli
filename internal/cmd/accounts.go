package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

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
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List global accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListGlobalAccounts(0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No global accounts found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ACCOUNT_ID\tNAME\tCURRENCY\tCOUNTRY\tSTATUS")
			for _, a := range result.Items {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					a.AccountID, a.AccountName, a.Currency, a.CountryCode, a.Status)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results")
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

			a, err := client.GetGlobalAccount(args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, a)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "account_id\t%s\n", a.AccountID)
			fmt.Fprintf(tw, "account_name\t%s\n", a.AccountName)
			fmt.Fprintf(tw, "currency\t%s\n", a.Currency)
			fmt.Fprintf(tw, "country_code\t%s\n", a.CountryCode)
			fmt.Fprintf(tw, "status\t%s\n", a.Status)
			if a.AccountNumber != "" {
				fmt.Fprintf(tw, "account_number\t%s\n", a.AccountNumber)
			}
			if a.RoutingCode != "" {
				fmt.Fprintf(tw, "routing_code\t%s\n", a.RoutingCode)
			}
			if a.IBAN != "" {
				fmt.Fprintf(tw, "iban\t%s\n", a.IBAN)
			}
			if a.SwiftCode != "" {
				fmt.Fprintf(tw, "swift_code\t%s\n", a.SwiftCode)
			}
			fmt.Fprintf(tw, "created_at\t%s\n", a.CreatedAt)
			tw.Flush()
			return nil
		},
	}
}
