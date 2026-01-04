package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List authorizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			var fromRFC3339, toRFC3339 string
			if from != "" {
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return fmt.Errorf("invalid --from date: %w", err)
				}
			}
			if to != "" {
				toRFC3339, err = convertDateToRFC3339End(to)
				if err != nil {
					return fmt.Errorf("invalid --to date: %w", err)
				}
			}

			pageSize = normalizePageSize(pageSize)

			result, err := client.ListAuthorizations(cmd.Context(), api.AuthorizationListParams{
				Status:               status,
				CardID:               cardID,
				BillingCurrency:      billingCurrency,
				DigitalWalletTokenID: digitalWalletTokenID,
				LifecycleID:          lifecycleID,
				RetrievalRef:         retrievalRef,
				FromCreatedAt:        fromRFC3339,
				ToCreatedAt:          toRFC3339,
				PageNum:              page,
				PageSize:             pageSize,
			})
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())
			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No authorizations found")
				return nil
			}

			headers := []string{"AUTH_ID", "TRANSACTION_ID", "CARD_ID", "STATUS", "AMOUNT", "CURRENCY", "MERCHANT"}
			rowFn := func(item any) []string {
				a := item.(api.Authorization)
				amount := ""
				if a.Amount != 0 {
					amount = fmt.Sprintf("%.2f", a.Amount)
				}
				return []string{authorizationID(a), a.TransactionID, a.CardID, a.Status, amount, a.Currency, a.Merchant.Name}
			}

			if err := f.OutputList(result.Items, headers, rowFn); err != nil {
				return err
			}

			if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
				_, _ = fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&cardID, "card-id", "", "Filter by card ID")
	cmd.Flags().StringVar(&billingCurrency, "billing-currency", "", "Filter by billing currency")
	cmd.Flags().StringVar(&digitalWalletTokenID, "digital-wallet-token-id", "", "Filter by digital wallet token ID")
	cmd.Flags().StringVar(&lifecycleID, "lifecycle-id", "", "Filter by lifecycle ID")
	cmd.Flags().StringVar(&retrievalRef, "retrieval-ref", "", "Filter by retrieval reference")
	cmd.Flags().StringVar(&from, "from", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
	return cmd
}

func newAuthorizationsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <transactionId>",
		Short: "Get authorization details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			auth, err := client.GetAuthorization(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, auth)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "authorization_id\t%s\n", authorizationID(*auth))
			_, _ = fmt.Fprintf(tw, "transaction_id\t%s\n", auth.TransactionID)
			_, _ = fmt.Fprintf(tw, "card_id\t%s\n", auth.CardID)
			_, _ = fmt.Fprintf(tw, "cardholder_id\t%s\n", auth.CardholderID)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", auth.Status)
			if auth.Amount != 0 || auth.Currency != "" {
				_, _ = fmt.Fprintf(tw, "amount\t%.2f %s\n", auth.Amount, auth.Currency)
			}
			if auth.Merchant.Name != "" {
				_, _ = fmt.Fprintf(tw, "merchant\t%s\n", auth.Merchant.Name)
			}
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", auth.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}
