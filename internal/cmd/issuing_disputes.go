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
	return NewListCommand(ListConfig[api.TransactionDispute]{
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
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.TransactionDispute], error) {
			result, err := client.ListTransactionDisputes(ctx, page, pageSize)
			if err != nil {
				return ListResult[api.TransactionDispute]{}, err
			}
			return ListResult[api.TransactionDispute]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
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
