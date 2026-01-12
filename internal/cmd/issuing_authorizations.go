package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newAuthorizationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorizations",
		Short: "Authorization retrieval",
	}
	cmd.AddCommand(newAuthorizationsListCmd())
	cmd.AddCommand(newAuthorizationsGetCmd())
	return cmd
}

func authorizationID(a api.Authorization) string {
	if a.AuthorizationID != "" {
		return a.AuthorizationID
	}
	if a.ID != "" {
		return a.ID
	}
	return a.TransactionID
}

func newAuthorizationsListCmd() *cobra.Command {
	var status string
	var cardID string
	var billingCurrency string
	var digitalWalletTokenID string
	var lifecycleID string
	var retrievalRef string
	var from string
	var to string
	cmd := NewListCommand(ListConfig[api.Authorization]{
		Use:          "list",
		Short:        "List authorizations",
		Headers:      []string{"AUTH_ID", "TRANSACTION_ID", "CARD_ID", "STATUS", "AMOUNT", "CURRENCY", "MERCHANT"},
		EmptyMessage: "No authorizations found",
		RowFunc: func(a api.Authorization) []string {
			amount := ""
			if a.Amount != 0 {
				amount = fmt.Sprintf("%.2f", a.Amount)
			}
			return []string{authorizationID(a), a.TransactionID, a.CardID, a.Status, amount, a.Currency, a.Merchant.Name}
		},
		MoreHint: "# More results available",
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Authorization], error) {
			fromRFC3339, toRFC3339, err := parseDateRangeRFC3339(from, to, "--from", "--to", false)
			if err != nil {
				return ListResult[api.Authorization]{}, err
			}

			result, err := client.ListAuthorizations(ctx, api.AuthorizationListParams{
				Status:               status,
				CardID:               cardID,
				BillingCurrency:      billingCurrency,
				DigitalWalletTokenID: digitalWalletTokenID,
				LifecycleID:          lifecycleID,
				RetrievalRef:         retrievalRef,
				FromCreatedAt:        fromRFC3339,
				ToCreatedAt:          toRFC3339,
				PageNum:              opts.Page,
				PageSize:             normalizePageSize(opts.Limit),
			})
			if err != nil {
				return ListResult[api.Authorization]{}, err
			}
			return ListResult[api.Authorization]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&cardID, "card-id", "", "Filter by card ID")
	cmd.Flags().StringVar(&billingCurrency, "billing-currency", "", "Filter by billing currency")
	cmd.Flags().StringVar(&digitalWalletTokenID, "digital-wallet-token-id", "", "Filter by digital wallet token ID")
	cmd.Flags().StringVar(&lifecycleID, "lifecycle-id", "", "Filter by lifecycle ID")
	cmd.Flags().StringVar(&retrievalRef, "retrieval-ref", "", "Filter by retrieval reference")
	cmd.Flags().StringVar(&from, "from", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To date (YYYY-MM-DD)")
	return cmd
}

func newAuthorizationsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Authorization]{
		Use:   "get <transactionId>",
		Short: "Get authorization details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Authorization, error) {
			return client.GetAuthorization(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, auth *api.Authorization) error {
			rows := []outfmt.KV{
				{Key: "authorization_id", Value: authorizationID(*auth)},
				{Key: "transaction_id", Value: auth.TransactionID},
				{Key: "card_id", Value: auth.CardID},
				{Key: "cardholder_id", Value: auth.CardholderID},
				{Key: "status", Value: auth.Status},
				{Key: "created_at", Value: auth.CreatedAt},
			}
			if auth.Amount != 0 || auth.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "amount", Value: fmt.Sprintf("%.2f %s", auth.Amount, auth.Currency)})
			}
			if auth.Merchant.Name != "" {
				rows = append(rows, outfmt.KV{Key: "merchant", Value: auth.Merchant.Name})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}
