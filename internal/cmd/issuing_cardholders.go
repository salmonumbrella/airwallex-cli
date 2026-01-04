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

func newCardholdersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cardholders",
		Short: "Cardholder management",
	}
	cmd.AddCommand(newCardholdersListCmd())
	cmd.AddCommand(newCardholdersGetCmd())
	cmd.AddCommand(newCardholdersCreateCmd())
	cmd.AddCommand(newCardholdersUpdateCmd())
	return cmd
}

func newCardholdersListCmd() *cobra.Command {
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cardholders",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			pageSize = normalizePageSize(pageSize)

			result, err := client.ListCardholders(cmd.Context(), page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No cardholders found")
				return nil
			}

			headers := []string{"CARDHOLDER_ID", "TYPE", "NAME", "EMAIL", "STATUS"}
			rowFn := func(item any) []string {
				ch := item.(api.Cardholder)
				name := fmt.Sprintf("%s %s", ch.FirstName, ch.LastName)
				return []string{ch.CardholderID, ch.Type, name, ch.Email, ch.Status}
			}

			if err := f.OutputList(result.Items, headers, rowFn); err != nil {
				return err
			}

			if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
	return cmd
}

func newCardholdersGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <cardholderId>",
		Short: "Get cardholder details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			ch, err := client.GetCardholder(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, ch)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "cardholder_id\t%s\n", ch.CardholderID)
			_, _ = fmt.Fprintf(tw, "type\t%s\n", ch.Type)
			_, _ = fmt.Fprintf(tw, "first_name\t%s\n", ch.FirstName)
			_, _ = fmt.Fprintf(tw, "last_name\t%s\n", ch.LastName)
			_, _ = fmt.Fprintf(tw, "email\t%s\n", ch.Email)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", ch.Status)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", ch.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}

func newCardholdersCreateCmd() *cobra.Command {
	var chType string
	var email string
	var firstName string
	var lastName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cardholder",
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
				return outfmt.WriteJSON(os.Stdout, ch)
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
		Use:   "update <cardholderId>",
		Short: "Update cardholder",
		Args:  cobra.ExactArgs(1),
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

			ch, err := client.UpdateCardholder(cmd.Context(), args[0], update)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, ch)
			}

			u.Success(fmt.Sprintf("Updated cardholder: %s", ch.CardholderID))
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Email address")
	return cmd
}
