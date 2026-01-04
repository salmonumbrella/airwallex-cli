package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newDepositsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits",
		Short: "Deposit operations",
		Long:  "View and track inbound deposits from bank transfers and linked accounts.",
	}
	cmd.AddCommand(newDepositsListCmd())
	cmd.AddCommand(newDepositsGetCmd())
	return cmd
}

func newDepositsListCmd() *cobra.Command {
	var status, fromDate, toDate string
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deposits",
		Long: `List inbound deposits with optional filters.

Examples:
  airwallex deposits list
  airwallex deposits list --status SETTLED
  airwallex deposits list --from 2024-01-01 --to 2024-01-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate date inputs
			if err := validateDate(fromDate); err != nil {
				return fmt.Errorf("--from: %w", err)
			}
			if err := validateDate(toDate); err != nil {
				return fmt.Errorf("--to: %w", err)
			}
			if err := validateDateRange(fromDate, toDate); err != nil {
				return err
			}

			pageSize = normalizePageSize(pageSize)

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListDeposits(cmd.Context(), status, fromDate, toDate, page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No deposits found")
				return nil
			}

			headers := []string{"DEPOSIT_ID", "AMOUNT", "CURRENCY", "STATUS", "SOURCE", "CREATED"}
			colTypes := []outfmt.ColumnType{
				outfmt.ColumnPlain,    // DEPOSIT_ID
				outfmt.ColumnAmount,   // AMOUNT
				outfmt.ColumnCurrency, // CURRENCY
				outfmt.ColumnStatus,   // STATUS
				outfmt.ColumnPlain,    // SOURCE
				outfmt.ColumnPlain,    // CREATED
			}
			rowFn := func(item any) []string {
				d := item.(api.Deposit)
				return []string{d.ID, fmt.Sprintf("%.2f", d.Amount), d.Currency, d.Status, d.Source, d.CreatedAt}
			}

			if err := f.OutputListWithColors(result.Items, headers, colTypes, rowFn); err != nil {
				return err
			}

			if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (PENDING, SETTLED, FAILED)")
	cmd.Flags().StringVar(&fromDate, "from", "", "From date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toDate, "to", "", "To date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
	return cmd
}

func newDepositsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <depositId>",
		Short: "Get deposit details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			d, err := client.GetDeposit(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, d)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "deposit_id\t%s\n", d.ID)
			_, _ = fmt.Fprintf(tw, "amount\t%.2f\n", d.Amount)
			_, _ = fmt.Fprintf(tw, "currency\t%s\n", d.Currency)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", d.Status)
			_, _ = fmt.Fprintf(tw, "source\t%s\n", d.Source)
			if d.LinkedAccountID != "" {
				_, _ = fmt.Fprintf(tw, "linked_account_id\t%s\n", d.LinkedAccountID)
			}
			if d.GlobalAccountID != "" {
				_, _ = fmt.Fprintf(tw, "global_account_id\t%s\n", d.GlobalAccountID)
			}
			if d.Reference != "" {
				_, _ = fmt.Fprintf(tw, "reference\t%s\n", d.Reference)
			}
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", d.CreatedAt)
			if d.SettledAt != "" {
				_, _ = fmt.Fprintf(tw, "settled_at\t%s\n", d.SettledAt)
			}
			_ = tw.Flush()
			return nil
		},
	}
}
