package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newCardholdersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cardholders",
		Aliases: []string{"cardholder", "ch"},
		Short:   "Cardholder management",
	}
	cmd.AddCommand(newCardholdersListCmd())
	cmd.AddCommand(newCardholdersGetCmd())
	cmd.AddCommand(newCardholdersCreateCmd())
	cmd.AddCommand(newCardholdersUpdateCmd())
	return cmd
}

func newCardholdersListCmd() *cobra.Command {
	cmd := NewListCommand(ListConfig[api.Cardholder]{
		Use:          "list",
		Aliases:      []string{"ls", "l"},
		Short:        "List cardholders",
		Headers:      []string{"CARDHOLDER_ID", "TYPE", "NAME", "EMAIL", "STATUS"},
		EmptyMessage: "No cardholders found",
		RowFunc: func(ch api.Cardholder) []string {
			name := fmt.Sprintf("%s %s", ch.FirstName, ch.LastName)
			return []string{ch.CardholderID, ch.Type, name, ch.Email, ch.Status}
		},
		IDFunc: func(ch api.Cardholder) string { return ch.CardholderID },
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Cardholder], error) {
			result, err := client.ListCardholders(ctx, opts.Page, normalizePageSize(opts.Limit))
			if err != nil {
				return ListResult[api.Cardholder]{}, err
			}
			return ListResult[api.Cardholder]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
	return cmd
}

func newCardholdersGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Cardholder]{
		Use:     "get <cardholderId>",
		Aliases: []string{"g"},
		Short:   "Get cardholder details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Cardholder, error) {
			return client.GetCardholder(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, ch *api.Cardholder) error {
			rows := []outfmt.KV{
				{Key: "cardholder_id", Value: ch.CardholderID},
				{Key: "type", Value: ch.Type},
				{Key: "first_name", Value: ch.FirstName},
				{Key: "last_name", Value: ch.LastName},
				{Key: "email", Value: ch.Email},
				{Key: "status", Value: ch.Status},
				{Key: "created_at", Value: ch.CreatedAt},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newCardholdersCreateCmd() *cobra.Command {
	var chType string
	var email string
	var firstName string
	var lastName string

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a new cardholder",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			req := map[string]interface{}{
				"type":       chType,
				"email":      email,
				"first_name": firstName,
				"last_name":  lastName,
			}

			ch, err := client.CreateCardholder(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return writeJSONOutput(cmd, ch)
			}

			u.Success(fmt.Sprintf("Created cardholder: %s", ch.CardholderID))
			return nil
		},
	}

	cmd.Flags().StringVar(&chType, "type", "INDIVIDUAL", "INDIVIDUAL or DELEGATE")
	cmd.Flags().StringVar(&email, "email", "", "Email address (required)")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name (required)")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name (required)")
	mustMarkRequired(cmd, "email")
	mustMarkRequired(cmd, "first-name")
	mustMarkRequired(cmd, "last-name")
	return cmd
}

func newCardholdersUpdateCmd() *cobra.Command {
	var email string

	cmd := &cobra.Command{
		Use:     "update <cardholderId>",
		Aliases: []string{"up", "u"},
		Short:   "Update cardholder",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			update := make(map[string]interface{})
			if cmd.Flags().Changed("email") {
				update["email"] = email
			}

			if len(update) == 0 {
				return fmt.Errorf("no updates specified")
			}

			cardholderID := NormalizeIDArg(args[0])
			ch, err := client.UpdateCardholder(cmd.Context(), cardholderID, update)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return writeJSONOutput(cmd, ch)
			}

			u.Success(fmt.Sprintf("Updated cardholder: %s", ch.CardholderID))
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Email address")
	return cmd
}
