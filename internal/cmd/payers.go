package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
			if err := validateDate(from); err != nil {
				return fmt.Errorf("invalid --from date: %w", err)
			}
			if err := validateDate(to); err != nil {
				return fmt.Errorf("invalid --to date: %w", err)
			}
			if err := validateDateRange(from, to); err != nil {
				return err
			}

			if pageSize < 10 {
				pageSize = 10
			}

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
	return &cobra.Command{
		Use:   "get <payerId>",
		Short: "Get payer details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payer, err := client.GetPayer(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, payer)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "payer_id\t%s\n", payerID(*payer))
			_, _ = fmt.Fprintf(tw, "entity_type\t%s\n", payer.EntityType)
			_, _ = fmt.Fprintf(tw, "name\t%s\n", payer.Name)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", payer.Status)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", payer.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}

func newPayersCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payer",
		Long: `Create a payer using a JSON payload.

Examples:
  airwallex payers create --data '{"entity_type":"COMPANY","name":"Acme Corp"}'
  airwallex payers create --from-file payer.json`,
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

			payer, err := client.CreatePayer(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, payer)
			}

			u.Success(fmt.Sprintf("Created payer: %s", payerID(*payer)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newPayersUpdateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <payerId>",
		Short: "Update a payer",
		Long: `Update a payer using a JSON payload.

Examples:
  airwallex payers update payer_123 --data '{"name":"Updated Name"}'
  airwallex payers update payer_123 --from-file update.json`,
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

			payer, err := client.UpdatePayer(cmd.Context(), args[0], payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, payer)
			}

			u.Success(fmt.Sprintf("Updated payer: %s", payerID(*payer)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
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
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate payer details",
		Long: `Validate payer details using a JSON payload.

Examples:
  airwallex payers validate --data '{"entity_type":"COMPANY","name":"Acme Corp"}'
  airwallex payers validate --from-file payer.json`,
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

			if err := client.ValidatePayer(cmd.Context(), payload); err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, map[string]any{"valid": true})
			}

			u.Success("Payer details are valid")
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}
