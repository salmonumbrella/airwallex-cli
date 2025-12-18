package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

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
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cardholders",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListCardholders(0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No cardholders found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "CARDHOLDER_ID\tTYPE\tNAME\tEMAIL\tSTATUS")
			for _, ch := range result.Items {
				name := fmt.Sprintf("%s %s", ch.FirstName, ch.LastName)
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					ch.CardholderID, ch.Type, name, ch.Email, ch.Status)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results")
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

			ch, err := client.GetCardholder(args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, ch)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "cardholder_id\t%s\n", ch.CardholderID)
			fmt.Fprintf(tw, "type\t%s\n", ch.Type)
			fmt.Fprintf(tw, "first_name\t%s\n", ch.FirstName)
			fmt.Fprintf(tw, "last_name\t%s\n", ch.LastName)
			fmt.Fprintf(tw, "email\t%s\n", ch.Email)
			fmt.Fprintf(tw, "status\t%s\n", ch.Status)
			fmt.Fprintf(tw, "created_at\t%s\n", ch.CreatedAt)
			tw.Flush()
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

			ch, err := client.CreateCardholder(req)
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
	if err := cmd.MarkFlagRequired("email"); err != nil {
		panic(fmt.Sprintf("failed to mark email as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("first-name"); err != nil {
		panic(fmt.Sprintf("failed to mark first-name as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("last-name"); err != nil {
		panic(fmt.Sprintf("failed to mark last-name as required: %v", err))
	}
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

			ch, err := client.UpdateCardholder(args[0], update)
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
