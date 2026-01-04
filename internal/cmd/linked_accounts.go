package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newLinkedAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "linked-accounts",
		Aliases: []string{"la"},
		Short:   "Linked account operations",
		Long:    "Manage linked external bank accounts for direct debits.",
	}
	cmd.AddCommand(newLinkedAccountsListCmd())
	cmd.AddCommand(newLinkedAccountsGetCmd())
	cmd.AddCommand(newLinkedAccountsCreateCmd())
	cmd.AddCommand(newLinkedAccountsDepositCmd())
	return cmd
}

func newLinkedAccountsListCmd() *cobra.Command {
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List linked accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			pageSize = normalizePageSize(pageSize)

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListLinkedAccounts(cmd.Context(), page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No linked accounts found")
				return nil
			}

			headers := []string{"ID", "TYPE", "ACCOUNT_NAME", "BANK", "CURRENCY", "STATUS"}
			rowFn := func(item any) []string {
				la := item.(api.LinkedAccount)
				return []string{la.ID, la.Type, la.AccountName, la.BankName, la.Currency, la.Status}
			}

			if err := f.OutputList(result.Items, headers, rowFn); err != nil {
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

func newLinkedAccountsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <accountId>",
		Short: "Get linked account details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			la, err := client.GetLinkedAccount(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, la)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "id\t%s\n", la.ID)
			_, _ = fmt.Fprintf(tw, "type\t%s\n", la.Type)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", la.Status)
			_, _ = fmt.Fprintf(tw, "account_name\t%s\n", la.AccountName)
			if la.BankName != "" {
				_, _ = fmt.Fprintf(tw, "bank_name\t%s\n", la.BankName)
			}
			if la.AccountNumber != "" {
				_, _ = fmt.Fprintf(tw, "account_number\t****%s\n", la.AccountNumber)
			}
			_, _ = fmt.Fprintf(tw, "currency\t%s\n", la.Currency)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", la.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}

func newLinkedAccountsCreateCmd() *cobra.Command {
	var accountType, accountName, currency string
	var bsb, routingNumber, accountNumber string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a linked account",
		Long: `Link an external bank account for direct debits.

Examples:
  # Australian bank account
  airwallex linked-accounts create --type AU_BANK --account-name "My Account" \
    --currency AUD --bsb 062000 --account-number 12345678

  # US bank account
  airwallex linked-accounts create --type US_BANK --account-name "My Account" \
    --currency USD --routing-number 021000021 --account-number 12345678

Account types: AU_BANK, US_BANK, CA_BANK, GB_BANK, SG_BANK, HK_BANK`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate currency
			if err := validateCurrency(currency); err != nil {
				return err
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			req := map[string]interface{}{
				"type":           accountType,
				"account_name":   accountName,
				"currency":       currency,
				"account_number": accountNumber,
			}

			// Add type-specific fields
			if bsb != "" {
				req["bsb"] = bsb
			}
			if routingNumber != "" {
				req["routing_number"] = routingNumber
			}

			la, err := client.CreateLinkedAccount(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, la)
			}

			u.Success(fmt.Sprintf("Created linked account: %s", la.ID))
			return nil
		},
	}

	cmd.Flags().StringVar(&accountType, "type", "", "Account type (AU_BANK, US_BANK, etc.) (required)")
	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name (required)")
	cmd.Flags().StringVar(&currency, "currency", "", "Currency (required)")
	cmd.Flags().StringVar(&accountNumber, "account-number", "", "Account number (required)")
	cmd.Flags().StringVar(&bsb, "bsb", "", "BSB (Australian accounts)")
	cmd.Flags().StringVar(&routingNumber, "routing-number", "", "Routing number (US accounts)")
	mustMarkRequired(cmd, "type")
	mustMarkRequired(cmd, "account-name")
	mustMarkRequired(cmd, "currency")
	mustMarkRequired(cmd, "account-number")
	return cmd
}

func newLinkedAccountsDepositCmd() *cobra.Command {
	var amount float64
	var currency string

	cmd := &cobra.Command{
		Use:   "deposit <accountId>",
		Short: "Initiate a deposit from a linked account",
		Long: `Pull funds from a linked external bank account into your Airwallex wallet.

Examples:
  airwallex linked-accounts deposit la_xxx --amount 5000 --currency AUD`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if err := validateAmount(amount); err != nil {
				return err
			}
			if err := validateCurrency(currency); err != nil {
				return err
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			di, err := client.InitiateDeposit(cmd.Context(), args[0], amount, currency)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, di)
			}

			u.Success(fmt.Sprintf("Deposit initiated: %s (%.2f %s)", di.ID, di.Amount, di.Currency))
			return nil
		},
	}

	cmd.Flags().Float64Var(&amount, "amount", 0, "Amount to deposit (required)")
	cmd.Flags().StringVar(&currency, "currency", "", "Currency (required)")
	mustMarkRequired(cmd, "amount")
	mustMarkRequired(cmd, "currency")
	return cmd
}
