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

func newPayersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payers",
		Short: "Payer management for payouts",
	}
	cmd.AddCommand(newPayersListCmd())
	cmd.AddCommand(newPayersGetCmd())
	cmd.AddCommand(newPayersCreateCmd())
	cmd.AddCommand(newPayersUpdateCmd())
	cmd.AddCommand(newPayersDeleteCmd())
	cmd.AddCommand(newPayersValidateCmd())
	return cmd
}

func payerID(p api.Payer) string {
	if p.ID != "" {
		return p.ID
	}
	return p.PayerID
}

func newPayersListCmd() *cobra.Command {
	var entityType string
	var name string
	var nickName string
	var from string
	var to string
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateDateRangeFlags(from, to, "--from", "--to", true); err != nil {
				return err
			}

			pageSize = normalizePageSize(pageSize)

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListPayers(cmd.Context(), api.PayerListParams{
				EntityType: entityType,
				Name:       name,
				NickName:   nickName,
				FromDate:   from,
				ToDate:     to,
				PageNum:    page,
				PageSize:   pageSize,
			})
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())
			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No payers found")
				return nil
			}

			headers := []string{"PAYER_ID", "ENTITY_TYPE", "NAME", "STATUS"}
			rowFn := func(item any) []string {
				p := item.(api.Payer)
				return []string{payerID(p), p.EntityType, p.Name, p.Status}
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

	cmd.Flags().StringVar(&entityType, "entity-type", "", "Filter by entity type")
	cmd.Flags().StringVar(&name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&nickName, "nick-name", "", "Filter by nickname")
	cmd.Flags().StringVar(&from, "from", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
	return cmd
}

func newPayersGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Payer]{
		Use:   "get <payerId>",
		Short: "Get payer details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Payer, error) {
			return client.GetPayer(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, payer *api.Payer) error {
			rows := []outfmt.KV{
				{Key: "payer_id", Value: payerID(*payer)},
				{Key: "entity_type", Value: payer.EntityType},
				{Key: "name", Value: payer.Name},
				{Key: "status", Value: payer.Status},
				{Key: "created_at", Value: payer.CreatedAt},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newPayersCreateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.Payer]{
		Use:   "create",
		Short: "Create a payer",
		Long: `Create a payer using a JSON payload.

Examples:
  airwallex payers create --data '{"entity_type":"COMPANY","name":"Acme Corp"}'
  airwallex payers create --from-file payer.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.Payer, error) {
			return client.CreatePayer(ctx, payload)
		},
		SuccessMessage: func(payer *api.Payer) string {
			return fmt.Sprintf("Created payer: %s", payerID(*payer))
		},
	}, getClient)
}

func newPayersUpdateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.Payer]{
		Use:   "update <payerId>",
		Short: "Update a payer",
		Long: `Update a payer using a JSON payload.

Examples:
  airwallex payers update payer_123 --data '{"name":"Updated Name"}'
  airwallex payers update payer_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.Payer, error) {
			return client.UpdatePayer(ctx, args[0], payload)
		},
		SuccessMessage: func(payer *api.Payer) string {
			return fmt.Sprintf("Updated payer: %s", payerID(*payer))
		},
	}, getClient)
}

func newPayersDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <payerId>",
		Short: "Delete a payer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			ok, err := ConfirmOrYes(cmd.Context(), fmt.Sprintf("Delete payer %s?", args[0]))
			if err != nil {
				return err
			}
			if !ok {
				u.Info("Cancelled")
				return nil
			}

			if err := client.DeletePayer(cmd.Context(), args[0]); err != nil {
				return err
			}

			u.Success("Payer deleted")
			return nil
		},
	}
}

func newPayersValidateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[map[string]any]{
		Use:   "validate",
		Short: "Validate payer details",
		Long: `Validate payer details using a JSON payload.

Examples:
  airwallex payers validate --data '{"entity_type":"COMPANY","name":"Acme Corp"}'
  airwallex payers validate --from-file payer.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (map[string]any, error) {
			if err := client.ValidatePayer(ctx, payload); err != nil {
				return nil, err
			}
			return map[string]any{"valid": true}, nil
		},
		SuccessMessage: func(_ map[string]any) string {
			return "Payer details are valid"
		},
	}, getClient)
}
