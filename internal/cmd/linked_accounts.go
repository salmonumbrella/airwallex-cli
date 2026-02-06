package cmd

import (
	"context"
	"fmt"
	"os"

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
	return NewListCommand(ListConfig[api.LinkedAccount]{
		Use:          "list",
		Short:        "List linked accounts",
		Headers:      []string{"ID", "TYPE", "ACCOUNT_NAME", "BANK", "CURRENCY", "STATUS"},
		EmptyMessage: "No linked accounts found",
		RowFunc: func(la api.LinkedAccount) []string {
			return []string{la.ID, la.Type, la.AccountName, la.BankName, la.Currency, la.Status}
		},
		MoreHint: "# More results available",
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.LinkedAccount], error) {
			result, err := client.ListLinkedAccounts(ctx, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.LinkedAccount]{}, err
			}
			return ListResult[api.LinkedAccount]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newLinkedAccountsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.LinkedAccount]{
		Use:   "get <accountId>",
		Short: "Get linked account details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.LinkedAccount, error) {
			return client.GetLinkedAccount(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, la *api.LinkedAccount) error {
			rows := []outfmt.KV{
				{Key: "id", Value: la.ID},
				{Key: "type", Value: la.Type},
				{Key: "status", Value: la.Status},
				{Key: "account_name", Value: la.AccountName},
				{Key: "currency", Value: la.Currency},
				{Key: "created_at", Value: la.CreatedAt},
			}
			if la.BankName != "" {
				rows = append(rows, outfmt.KV{Key: "bank_name", Value: la.BankName})
			}
			if la.AccountNumber != "" {
				rows = append(rows, outfmt.KV{Key: "account_number", Value: "****" + la.AccountNumber})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
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

			accountID := NormalizeIDArg(args[0])
			di, err := client.InitiateDeposit(cmd.Context(), accountID, amount, currency)
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
