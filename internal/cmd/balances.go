package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances",
		Short: "Balance operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior: show current balances
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			balances, err := client.GetBalances()
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, balances)
			}

			if len(balances.Balances) == 0 {
				fmt.Fprintln(os.Stderr, "No balances")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "CURRENCY\tAVAILABLE\tPENDING\tRESERVED\tTOTAL")
			for _, b := range balances.Balances {
				fmt.Fprintf(tw, "%s\t%.2f\t%.2f\t%.2f\t%.2f\n",
					b.Currency, b.AvailableAmount, b.PendingAmount, b.ReservedAmount, b.TotalAmount)
			}
			tw.Flush()
			return nil
		},
	}
	cmd.AddCommand(newBalancesHistoryCmd())
	return cmd
}

func newBalancesHistoryCmd() *cobra.Command {
	var currency string
	var from string
	var to string
	var pageSize int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show balance transaction history",
		Long: `Show balance transaction history.

Date range is limited to 7 days maximum per query.
Dates should be in YYYY-MM-DD format and will be converted to RFC3339.

Examples:
  # Show all balance history
  airwallex balances history

  # Filter by currency
  airwallex balances history --currency CAD

  # Filter by date range (max 7 days)
  airwallex balances history --from 2024-01-01 --to 2024-01-07

  # Combine filters with custom limit
  airwallex balances history --currency USD --from 2024-01-01 --to 2024-01-07 --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate page size (minimum 10)
			if pageSize < 10 {
				pageSize = 10
			}

			// Convert YYYY-MM-DD dates to RFC3339 format
			var fromRFC3339, toRFC3339 string
			var err error

			if from != "" {
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return fmt.Errorf("invalid --from date: %w", err)
				}
			}

			if to != "" {
				toRFC3339, err = convertDateToRFC3339(to)
				if err != nil {
					return fmt.Errorf("invalid --to date: %w", err)
				}
			}

			// Validate date range (max 7 days)
			if fromRFC3339 != "" && toRFC3339 != "" {
				fromTime, _ := time.Parse(time.RFC3339, fromRFC3339)
				toTime, _ := time.Parse(time.RFC3339, toRFC3339)
				if toTime.Sub(fromTime) > 7*24*time.Hour {
					return fmt.Errorf("date range exceeds 7 days maximum")
				}
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.GetBalanceHistory(currency, fromRFC3339, toRFC3339, 0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No balance history found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tCURRENCY\tAMOUNT\tBALANCE\tTYPE\tCREATED_AT\tDESCRIPTION")
			for _, item := range result.Items {
				fmt.Fprintf(tw, "%s\t%s\t%.2f\t%.2f\t%s\t%s\t%s\n",
					item.ID, item.Currency, item.Amount, item.Balance,
					item.TransactionType, item.CreatedAt, item.Description)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency (e.g., CAD, USD)")
	cmd.Flags().StringVar(&from, "from", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "End date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results per page (min 10)")

	return cmd
}

// convertDateToRFC3339 converts a date string in YYYY-MM-DD format to RFC3339 format
// with time set to 00:00:00 UTC
func convertDateToRFC3339(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("expected format YYYY-MM-DD, got %q", dateStr)
	}
	// Convert to RFC3339 format with UTC timezone
	return t.UTC().Format(time.RFC3339), nil
}
