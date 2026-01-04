package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

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
			if err := validateDate(from); err != nil {
				return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --from date: %w", err)
			}
			if err := validateDate(to); err != nil {
				return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --to date: %w", err)
			}
			if err := validateDate(fromUpdated); err != nil {
				return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --from-updated date: %w", err)
			}
			if err := validateDate(toUpdated); err != nil {
				return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --to-updated date: %w", err)
			}

			fromRFC3339 := ""
			if from != "" {
				var err error
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --from date: %w", err)
				}
			}
			toRFC3339 := ""
			if to != "" {
				var err error
				toRFC3339, err = convertDateToRFC3339End(to)
				if err != nil {
					return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --to date: %w", err)
				}
			}
			fromUpdatedRFC3339 := ""
			if fromUpdated != "" {
				var err error
				fromUpdatedRFC3339, err = convertDateToRFC3339(fromUpdated)
				if err != nil {
					return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --from-updated date: %w", err)
				}
			}
			toUpdatedRFC3339 := ""
			if toUpdated != "" {
				var err error
				toUpdatedRFC3339, err = convertDateToRFC3339End(toUpdated)
				if err != nil {
					return ListResult[api.TransactionDispute]{}, fmt.Errorf("invalid --to-updated date: %w", err)
				}
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
	return &cobra.Command{
		Use:   "get <disputeId>",
		Short: "Get dispute details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			dispute, err := client.GetTransactionDispute(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, dispute)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "dispute_id\t%s\n", disputeID(*dispute))
			_, _ = fmt.Fprintf(tw, "transaction_id\t%s\n", dispute.TransactionID)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", dispute.Status)
			if dispute.Amount != 0 || dispute.Currency != "" {
				_, _ = fmt.Fprintf(tw, "amount\t%.2f %s\n", dispute.Amount, dispute.Currency)
			}
			_, _ = fmt.Fprintf(tw, "reason\t%s\n", dispute.Reason)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", dispute.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}

func newDisputesCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a dispute",
		Long: `Create a dispute using a JSON payload.

Examples:
  airwallex issuing disputes create --data '{"transaction_id":"txn_123","reason":"fraud"}'
  airwallex issuing disputes create --from-file dispute.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			dispute, err := client.CreateTransactionDispute(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, dispute)
			}

			u.Success(fmt.Sprintf("Created dispute: %s", disputeID(*dispute)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newDisputesUpdateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <disputeId>",
		Short: "Update a dispute",
		Long: `Update a dispute using a JSON payload.

Examples:
  airwallex issuing disputes update dpt_123 --data '{"reason":"service_not_received"}'
  airwallex issuing disputes update dpt_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			dispute, err := client.UpdateTransactionDispute(cmd.Context(), args[0], payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, dispute)
			}

			u.Success(fmt.Sprintf("Updated dispute: %s", disputeID(*dispute)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
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

			dispute, err := client.SubmitTransactionDispute(cmd.Context(), args[0])
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

			dispute, err := client.CancelTransactionDispute(cmd.Context(), args[0])
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
