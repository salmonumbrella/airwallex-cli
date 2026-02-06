package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

// validWebhookEvents contains all valid Airwallex webhook event types
var validWebhookEvents = map[string]bool{
	// Transfer events
	"transfer.completed":  true,
	"transfer.failed":     true,
	"transfer.cancelled":  true,
	"transfer.created":    true,
	"transfer.updated":    true,
	"transfer.processing": true,

	// Payment events
	"payment.completed":            true,
	"payment.failed":               true,
	"payment.created":              true,
	"payment.updated":              true,
	"payment.cancelled":            true,
	"payment.authorization_failed": true,
	"payment.capture_failed":       true,
	"payment.refund_completed":     true,
	"payment.refund_failed":        true,
	"payment.chargeback_received":  true,
	"payment.chargeback_reversed":  true,
	"payment.dispute_opened":       true,
	"payment.dispute_resolved":     true,

	// Deposit events
	"deposit.settled":    true,
	"deposit.failed":     true,
	"deposit.created":    true,
	"deposit.processing": true,

	// Beneficiary events
	"beneficiary.created":  true,
	"beneficiary.updated":  true,
	"beneficiary.deleted":  true,
	"beneficiary.verified": true,

	// Card events
	"card.activated":             true,
	"card.deactivated":           true,
	"card.transaction.completed": true,
	"card.transaction.declined":  true,
	"card.transaction.reversed":  true,
	"card.created":               true,
	"card.updated":               true,

	// Payout events
	"payout.completed":  true,
	"payout.failed":     true,
	"payout.created":    true,
	"payout.processing": true,
	"payout.cancelled":  true,

	// Account events
	"account.updated":         true,
	"account.balance.updated": true,

	// Invoice events
	"invoice.created": true,
	"invoice.paid":    true,
	"invoice.overdue": true,
	"invoice.voided":  true,

	// Verification events
	"verification.completed": true,
	"verification.failed":    true,
	"verification.required":  true,
}

func newWebhooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhooks",
		Aliases: []string{"webhook", "wh"},
		Short:   "Webhook operations",
		Long: `Manage webhook subscriptions for receiving event notifications.

Common events:
  transfer.completed, transfer.failed
  deposit.settled, deposit.failed
  card.activated, card.transaction.completed
  beneficiary.created, beneficiary.updated`,
	}
	cmd.AddCommand(newWebhooksListCmd())
	cmd.AddCommand(newWebhooksGetCmd())
	cmd.AddCommand(newWebhooksCreateCmd())
	cmd.AddCommand(newWebhooksDeleteCmd())
	return cmd
}

func newWebhooksListCmd() *cobra.Command {
	return NewListCommand(ListConfig[api.Webhook]{
		Use:          "list",
		Short:        "List webhook subscriptions",
		Headers:      []string{"ID", "URL", "EVENTS", "STATUS"},
		EmptyMessage: "No webhooks found",
		RowFunc: func(wh api.Webhook) []string {
			events := strings.Join(wh.Events, ", ")
			if len(events) > 40 {
				events = events[:37] + "..."
			}
			return []string{wh.ID, wh.URL, events, wh.Status}
		},
		MoreHint: "# More results available",
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Webhook], error) {
			result, err := client.ListWebhooks(ctx, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.Webhook]{}, err
			}
			return ListResult[api.Webhook]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newWebhooksGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Webhook]{
		Use:   "get <webhookId>",
		Short: "Get webhook details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Webhook, error) {
			return client.GetWebhook(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, wh *api.Webhook) error {
			rows := []outfmt.KV{
				{Key: "id", Value: wh.ID},
				{Key: "url", Value: wh.URL},
				{Key: "events", Value: strings.Join(wh.Events, ", ")},
				{Key: "status", Value: wh.Status},
				{Key: "created_at", Value: wh.CreatedAt},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newWebhooksCreateCmd() *cobra.Command {
	var webhookURL string
	var events []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook subscription",
		Long: `Create a new webhook subscription to receive event notifications.

Examples:
  airwallex webhooks create --url https://example.com/hook --events transfer.completed,deposit.settled
  airwallex webhooks create --url https://example.com/hook --events transfer.completed --events transfer.failed

Common events:
  transfer.completed, transfer.failed, transfer.cancelled
  deposit.settled, deposit.failed
  card.activated, card.transaction.completed
  beneficiary.created, beneficiary.updated`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Parse comma-separated events with deduplication
			seen := make(map[string]bool)
			var allEvents []string
			var invalidEvents []string

			for _, e := range events {
				for _, ev := range strings.Split(e, ",") {
					ev = strings.TrimSpace(ev)
					if ev == "" {
						continue
					}

					// Check for duplicates
					if seen[ev] {
						continue
					}
					seen[ev] = true

					// Validate event type
					if !validWebhookEvents[ev] {
						invalidEvents = append(invalidEvents, ev)
						continue
					}

					allEvents = append(allEvents, ev)
				}
			}

			if len(invalidEvents) > 0 {
				return fmt.Errorf("invalid event types: %s", strings.Join(invalidEvents, ", "))
			}

			if len(allEvents) == 0 {
				return fmt.Errorf("at least one valid event is required")
			}

			wh, err := client.CreateWebhook(cmd.Context(), webhookURL, allEvents)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, wh)
			}

			u.Success(fmt.Sprintf("Created webhook: %s", wh.ID))
			_, _ = fmt.Fprintf(os.Stdout, "URL: %s\n", wh.URL)
			_, _ = fmt.Fprintf(os.Stdout, "Events: %s\n", strings.Join(wh.Events, ", "))
			return nil
		},
	}

	cmd.Flags().StringVar(&webhookURL, "url", "", "Webhook URL (required)")
	cmd.Flags().StringArrayVar(&events, "events", nil, "Events to subscribe to (comma-separated or repeated)")
	mustMarkRequired(cmd, "url")
	mustMarkRequired(cmd, "events")
	return cmd
}

func newWebhooksDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <webhookId>",
		Short: "Delete a webhook subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			webhookID := NormalizeIDArg(args[0])

			// Prompt for confirmation (respects --yes flag, JSON mode, and TTY detection)
			prompt := fmt.Sprintf("Are you sure you want to delete webhook %s?", webhookID)
			confirmed, err := ConfirmOrYes(cmd.Context(), prompt)
			if err != nil {
				return err
			}
			if !confirmed {
				u.Info("Deletion cancelled.")
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.DeleteWebhook(cmd.Context(), webhookID); err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, map[string]string{
					"id":     webhookID,
					"status": "deleted",
				})
			}

			u.Success(fmt.Sprintf("Deleted webhook: %s", webhookID))
			return nil
		},
	}

	return cmd
}
