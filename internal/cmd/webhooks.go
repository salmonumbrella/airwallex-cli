package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newWebhooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhooks",
		Short: "Webhook operations",
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
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhook subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageSize < 10 {
				pageSize = 10
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListWebhooks(cmd.Context(), 0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No webhooks found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tURL\tEVENTS\tSTATUS")
			for _, wh := range result.Items {
				events := strings.Join(wh.Events, ", ")
				if len(events) > 40 {
					events = events[:37] + "..."
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					wh.ID, wh.URL, events, wh.Status)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results (min 10)")
	return cmd
}

func newWebhooksGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <webhookId>",
		Short: "Get webhook details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			wh, err := client.GetWebhook(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, wh)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "id\t%s\n", wh.ID)
			fmt.Fprintf(tw, "url\t%s\n", wh.URL)
			fmt.Fprintf(tw, "events\t%s\n", strings.Join(wh.Events, ", "))
			fmt.Fprintf(tw, "status\t%s\n", wh.Status)
			fmt.Fprintf(tw, "created_at\t%s\n", wh.CreatedAt)
			tw.Flush()
			return nil
		},
	}
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

			// Parse comma-separated events
			var allEvents []string
			for _, e := range events {
				for _, ev := range strings.Split(e, ",") {
					ev = strings.TrimSpace(ev)
					if ev != "" {
						allEvents = append(allEvents, ev)
					}
				}
			}

			if len(allEvents) == 0 {
				return fmt.Errorf("at least one event is required")
			}

			wh, err := client.CreateWebhook(cmd.Context(), webhookURL, allEvents)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, wh)
			}

			u.Success(fmt.Sprintf("Created webhook: %s", wh.ID))
			fmt.Printf("URL: %s\n", wh.URL)
			fmt.Printf("Events: %s\n", strings.Join(wh.Events, ", "))
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
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "delete <webhookId>",
		Short: "Delete a webhook subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			webhookID := args[0]

			if !skipConfirm {
				fmt.Printf("Are you sure you want to delete webhook %s? [y/N]: ", webhookID)
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				if response != "y" && response != "yes" {
					fmt.Println("Deletion aborted.")
					return nil
				}
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.DeleteWebhook(cmd.Context(), webhookID); err != nil {
				return err
			}

			u.Success(fmt.Sprintf("Deleted webhook: %s", webhookID))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}
