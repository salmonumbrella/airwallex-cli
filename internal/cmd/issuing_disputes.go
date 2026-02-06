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

func newDisputesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disputes",
		Short: "Transaction dispute management",
	}
	cmd.AddCommand(newDisputesListCmd())
	cmd.AddCommand(newDisputesGetCmd())
	cmd.AddCommand(newDisputesCreateCmd())
	cmd.AddCommand(newDisputesUpdateCmd())
	cmd.AddCommand(newDisputesSubmitCmd())
	cmd.AddCommand(newDisputesCancelCmd())
	return cmd
}

func disputeID(d api.TransactionDispute) string {
	if d.DisputeID != "" {
		return d.DisputeID
	}
	return d.ID
}

func newDisputesListCmd() *cobra.Command {
	var status string
	var detailedStatus string
	var reason string
	var reference string
	var transactionID string
	var updatedBy string
	var from string
	var to string
	var fromUpdated string
	var toUpdated string

	cmd := NewListCommand(ListConfig[api.TransactionDispute]{
		Use:          "list",
		Short:        "List disputes",
		Headers:      []string{"DISPUTE_ID", "TRANSACTION_ID", "STATUS", "AMOUNT", "CURRENCY"},
		EmptyMessage: "No disputes found",
		RowFunc: func(d api.TransactionDispute) []string {
			amount := ""
			if d.Amount != 0 {
				amount = fmt.Sprintf("%.2f", d.Amount)
			}
			return []string{disputeID(d), d.TransactionID, d.Status, amount, d.Currency}
		},
		IDFunc: func(d api.TransactionDispute) string {
			return disputeID(d)
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.TransactionDispute], error) {
			fromRFC3339, toRFC3339, err := parseDateRangeRFC3339(from, to, "--from", "--to", false)
			if err != nil {
				return ListResult[api.TransactionDispute]{}, err
			}
			fromUpdatedRFC3339, toUpdatedRFC3339, err := parseDateRangeRFC3339(fromUpdated, toUpdated, "--from-updated", "--to-updated", false)
			if err != nil {
				return ListResult[api.TransactionDispute]{}, err
			}

			result, err := client.ListTransactionDisputes(ctx, api.TransactionDisputeListParams{
				Status:         status,
				DetailedStatus: detailedStatus,
				Reason:         reason,
				Reference:      reference,
				TransactionID:  transactionID,
				UpdatedBy:      updatedBy,
				FromCreatedAt:  fromRFC3339,
				ToCreatedAt:    toRFC3339,
				FromUpdatedAt:  fromUpdatedRFC3339,
				ToUpdatedAt:    toUpdatedRFC3339,
				Page:           "",
				PageSize:       opts.Limit,
			})
			if err != nil {
				return ListResult[api.TransactionDispute]{}, err
			}
			return ListResult[api.TransactionDispute]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&detailedStatus, "detailed-status", "", "Filter by detailed status")
	cmd.Flags().StringVar(&reason, "reason", "", "Filter by reason")
	cmd.Flags().StringVar(&reference, "reference", "", "Filter by reference")
	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "Filter by transaction ID")
	cmd.Flags().StringVar(&updatedBy, "updated-by", "", "Filter by updated by")
	cmd.Flags().StringVar(&from, "from", "", "From created date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To created date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&fromUpdated, "from-updated", "", "From updated date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toUpdated, "to-updated", "", "To updated date (YYYY-MM-DD)")
	return cmd
}

func newDisputesGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.TransactionDispute]{
		Use:   "get <disputeId>",
		Short: "Get dispute details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.TransactionDispute, error) {
			return client.GetTransactionDispute(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, dispute *api.TransactionDispute) error {
			rows := []outfmt.KV{
				{Key: "dispute_id", Value: disputeID(*dispute)},
				{Key: "transaction_id", Value: dispute.TransactionID},
				{Key: "status", Value: dispute.Status},
				{Key: "reason", Value: dispute.Reason},
				{Key: "created_at", Value: dispute.CreatedAt},
			}
			if dispute.Amount != 0 || dispute.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "amount", Value: fmt.Sprintf("%.2f %s", dispute.Amount, dispute.Currency)})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newDisputesCreateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.TransactionDispute]{
		Use:   "create",
		Short: "Create a dispute",
		Long: `Create a dispute using a JSON payload.

Examples:
  airwallex issuing disputes create --data '{"transaction_id":"txn_123","reason":"fraud"}'
  airwallex issuing disputes create --from-file dispute.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.TransactionDispute, error) {
			return client.CreateTransactionDispute(ctx, payload)
		},
		SuccessMessage: func(dispute *api.TransactionDispute) string {
			return fmt.Sprintf("Created dispute: %s", disputeID(*dispute))
		},
	}, getClient)
}

func newDisputesUpdateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.TransactionDispute]{
		Use:   "update <disputeId>",
		Short: "Update a dispute",
		Long: `Update a dispute using a JSON payload.

Examples:
  airwallex issuing disputes update dpt_123 --data '{"reason":"service_not_received"}'
  airwallex issuing disputes update dpt_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.TransactionDispute, error) {
			return client.UpdateTransactionDispute(ctx, NormalizeIDArg(args[0]), payload)
		},
		SuccessMessage: func(dispute *api.TransactionDispute) string {
			return fmt.Sprintf("Updated dispute: %s", disputeID(*dispute))
		},
	}, getClient)
}

func newDisputesSubmitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "submit <disputeId>",
		Short: "Submit a dispute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			disputeIDArg := NormalizeIDArg(args[0])
			dispute, err := client.SubmitTransactionDispute(cmd.Context(), disputeIDArg)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, dispute)
			}

			u.Success(fmt.Sprintf("Submitted dispute: %s", disputeID(*dispute)))
			return nil
		},
	}
}

func newDisputesCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <disputeId>",
		Short: "Cancel a dispute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			disputeIDArg := NormalizeIDArg(args[0])
			dispute, err := client.CancelTransactionDispute(cmd.Context(), disputeIDArg)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, dispute)
			}

			u.Success(fmt.Sprintf("Cancelled dispute: %s", disputeID(*dispute)))
			return nil
		},
	}
}
