package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newReportsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Financial report operations",
		Long: `Generate, list, and download financial reports.

Report types:
  account-statement  - Official PDF account statements
  balance-activity   - Detailed balance activity (CSV/EXCEL/PDF)
  transaction-recon  - Transaction reconciliation (CSV/EXCEL)
  settlement         - Settlement reports (CSV/EXCEL)`,
	}
	cmd.AddCommand(newReportsListCmd())
	cmd.AddCommand(newReportsGetCmd())
	cmd.AddCommand(newReportsAccountStatementCmd())
	cmd.AddCommand(newReportsBalanceActivityCmd())
	cmd.AddCommand(newReportsTransactionReconCmd())
	cmd.AddCommand(newReportsSettlementCmd())
	return cmd
}

func newReportsListCmd() *cobra.Command {
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all financial reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate page size (minimum 10)
			if pageSize < 10 {
				pageSize = 10
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListFinancialReports(cmd.Context(), page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(result)
				}
				f.Empty("No reports found")
				return nil
			}

			headers := []string{"ID", "TYPE", "STATUS", "DATE_RANGE", "FORMAT", "EXPIRES_AT"}
			rowFn := func(item any) []string {
				r := item.(api.FinancialReport)
				dateRange := fmt.Sprintf("%s to %s", r.FromDate, r.ToDate)
				expiresAt := r.ReportExpiresAt
				if expiresAt == "" {
					expiresAt = "N/A"
				}
				return []string{r.ID, r.Type, r.Status, dateRange, r.FileFormat, expiresAt}
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

func newReportsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <reportId>",
		Short: "Get report details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			r, err := client.GetFinancialReport(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, r)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "id\t%s\n", r.ID)
			_, _ = fmt.Fprintf(tw, "type\t%s\n", r.Type)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", r.Status)
			_, _ = fmt.Fprintf(tw, "file_format\t%s\n", r.FileFormat)
			_, _ = fmt.Fprintf(tw, "from_date\t%s\n", r.FromDate)
			_, _ = fmt.Fprintf(tw, "to_date\t%s\n", r.ToDate)
			if len(r.Currencies) > 0 {
				_, _ = fmt.Fprintf(tw, "currencies\t%v\n", r.Currencies)
			}
			if len(r.TransactionTypes) > 0 {
				_, _ = fmt.Fprintf(tw, "transaction_types\t%v\n", r.TransactionTypes)
			}
			if r.ReportVersion != "" {
				_, _ = fmt.Fprintf(tw, "report_version\t%s\n", r.ReportVersion)
			}
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", r.CreatedAt)
			if r.ReportExpiresAt != "" {
				_, _ = fmt.Fprintf(tw, "report_expires_at\t%s\n", r.ReportExpiresAt)
			}
			if r.ErrorMessage != "" {
				_, _ = fmt.Fprintf(tw, "error_message\t%s\n", r.ErrorMessage)
			}
			_ = tw.Flush()
			return nil
		},
	}
}

func newReportsSettlementCmd() *cobra.Command {
	var fromDate, toDate string
	var fileFormat string
	var output string
	var wait bool
	var timeout int

	cmd := &cobra.Command{
		Use:   "settlement",
		Short: "Generate settlement report",
		Long: `Generate settlement reports for payment settlement batches.

Available formats: CSV, EXCEL

Examples:
  # Generate CSV settlement report
  airwallex reports settlement --from-date 2024-01-01 --to-date 2024-01-31 \
    --format CSV --output settlement.csv --wait

  # Generate Excel settlement report
  airwallex reports settlement --from-date 2024-01-01 --to-date 2024-01-31 \
    --format EXCEL --output settlement.xlsx --wait`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate date inputs
			if err := validateDate(fromDate); err != nil {
				return fmt.Errorf("--from-date: %w", err)
			}
			if err := validateDate(toDate); err != nil {
				return fmt.Errorf("--to-date: %w", err)
			}
			if err := validateDateRange(fromDate, toDate); err != nil {
				return err
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Validate format
			validFormats := map[string]bool{"CSV": true, "EXCEL": true}
			if !validFormats[fileFormat] {
				return fmt.Errorf("--format must be CSV or EXCEL")
			}

			req := &api.CreateReportRequest{
				Type:       api.ReportTypeSettlement,
				FromDate:   fromDate,
				ToDate:     toDate,
				FileFormat: fileFormat,
			}

			u.Info("Creating settlement report...")

			report, err := client.CreateFinancialReport(cmd.Context(), req)
			if err != nil {
				return err
			}

			u.Info(fmt.Sprintf("Report created: %s (status: %s)", report.ID, report.Status))

			if !wait {
				u.Success(fmt.Sprintf("Report ID: %s - Use 'airwallex reports get %s' to check status", report.ID, report.ID))
				return nil
			}

			// Wait
			u.Info("Waiting for report to complete...")
			report, err = client.WaitForReport(cmd.Context(), report.ID, time.Duration(timeout)*time.Second)
			if err != nil {
				return err
			}

			if report.Status == api.ReportStatusFailed {
				return fmt.Errorf("report failed: %s", report.ErrorMessage)
			}

			// Download
			u.Info("Downloading report...")
			content, contentType, err := client.DownloadFinancialReport(cmd.Context(), report.ID)
			if err != nil {
				return err
			}

			if output == "" {
				ext := map[string]string{"CSV": ".csv", "EXCEL": ".xlsx"}
				output = "settlement" + ext[fileFormat]
			}

			if err := os.WriteFile(output, content, 0o600); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			u.Success(fmt.Sprintf("Downloaded %s (%d bytes, %s)", output, len(content), contentType))
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDate, "from-date", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&fileFormat, "format", "CSV", "File format: CSV or EXCEL")
	cmd.Flags().StringVar(&output, "output", "", "Output filename (default: auto-generated)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for report and download")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds when waiting")

	mustMarkRequired(cmd, "from-date")
	mustMarkRequired(cmd, "to-date")

	return cmd
}

func newReportsAccountStatementCmd() *cobra.Command {
	var fromDate, toDate string
	var currencies []string
	var output string
	var wait bool
	var timeout int

	cmd := &cobra.Command{
		Use:   "account-statement",
		Short: "Generate account statement (PDF)",
		Long: `Generate official PDF account statements for specified currencies.

Examples:
  # Generate single currency statement
  airwallex reports account-statement --from-date 2024-01-01 --to-date 2024-01-31 \
    --currencies CAD --output statement.pdf --wait

  # Generate multi-currency statement (returns ZIP)
  airwallex reports account-statement --from-date 2024-01-01 --to-date 2024-01-31 \
    --currencies CAD,USD,EUR --output statements.zip --wait

Note: Multi-currency requests return a ZIP file containing individual PDF statements.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate date inputs
			if err := validateDate(fromDate); err != nil {
				return fmt.Errorf("--from-date: %w", err)
			}
			if err := validateDate(toDate); err != nil {
				return fmt.Errorf("--to-date: %w", err)
			}
			if err := validateDateRange(fromDate, toDate); err != nil {
				return err
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Validate currencies
			if len(currencies) == 0 {
				return fmt.Errorf("--currencies is required")
			}
			for _, curr := range currencies {
				if err := validateCurrency(curr); err != nil {
					return fmt.Errorf("invalid currency %q: %w", curr, err)
				}
			}

			req := &api.CreateReportRequest{
				Type:       api.ReportTypeAccountStatement,
				FromDate:   fromDate,
				ToDate:     toDate,
				FileFormat: api.FileFormatPDF,
				Currencies: currencies,
			}

			u.Info(fmt.Sprintf("Creating account statement report for %v...", currencies))

			report, err := client.CreateFinancialReport(cmd.Context(), req)
			if err != nil {
				return err
			}

			u.Info(fmt.Sprintf("Report created: %s (status: %s)", report.ID, report.Status))

			if !wait {
				u.Success(fmt.Sprintf("Report ID: %s - Use 'airwallex reports get %s' to check status", report.ID, report.ID))
				return nil
			}

			// Wait for completion
			u.Info("Waiting for report to complete...")
			report, err = client.WaitForReport(cmd.Context(), report.ID, time.Duration(timeout)*time.Second)
			if err != nil {
				return err
			}

			if report.Status == api.ReportStatusFailed {
				return fmt.Errorf("report failed: %s", report.ErrorMessage)
			}

			// Download the report
			u.Info("Downloading report...")
			content, contentType, err := client.DownloadFinancialReport(cmd.Context(), report.ID)
			if err != nil {
				return err
			}

			// Determine output filename extension
			if output == "" {
				if len(currencies) > 1 {
					output = "account-statement.zip"
				} else {
					output = "account-statement.pdf"
				}
			}

			if err := os.WriteFile(output, content, 0o600); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			u.Success(fmt.Sprintf("Downloaded %s (%d bytes, %s)", output, len(content), contentType))
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDate, "from-date", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().StringSliceVar(&currencies, "currencies", nil, "Currencies (e.g., CAD,USD, required)")
	cmd.Flags().StringVar(&output, "output", "", "Output filename (default: auto-generated)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for report and download")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds when waiting")

	mustMarkRequired(cmd, "from-date")
	mustMarkRequired(cmd, "to-date")
	mustMarkRequired(cmd, "currencies")

	return cmd
}

func newReportsBalanceActivityCmd() *cobra.Command {
	var fromDate, toDate string
	var fileFormat string
	var transactionTypes []string
	var output string
	var wait bool
	var timeout int

	cmd := &cobra.Command{
		Use:   "balance-activity",
		Short: "Generate balance activity report",
		Long: `Generate detailed balance activity reports showing all settled transactions.

Available formats: CSV, EXCEL, PDF

Examples:
  # Generate CSV report
  airwallex reports balance-activity --from-date 2024-01-01 --to-date 2024-03-31 \
    --format CSV --output activity.csv --wait

  # Generate Excel report with transaction type filter
  airwallex reports balance-activity --from-date 2024-01-01 --to-date 2024-03-31 \
    --format EXCEL --transaction-types PAYOUT,DEPOSIT --output activity.xlsx --wait`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate date inputs
			if err := validateDate(fromDate); err != nil {
				return fmt.Errorf("--from-date: %w", err)
			}
			if err := validateDate(toDate); err != nil {
				return fmt.Errorf("--to-date: %w", err)
			}
			if err := validateDateRange(fromDate, toDate); err != nil {
				return err
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Validate format
			validFormats := map[string]bool{"CSV": true, "EXCEL": true, "PDF": true}
			if !validFormats[fileFormat] {
				return fmt.Errorf("--format must be CSV, EXCEL, or PDF")
			}

			req := &api.CreateReportRequest{
				Type:             api.ReportTypeBalanceActivity,
				FromDate:         fromDate,
				ToDate:           toDate,
				FileFormat:       fileFormat,
				TransactionTypes: transactionTypes,
				ReportVersion:    "1.1.0",
			}

			u.Info("Creating balance activity report...")

			report, err := client.CreateFinancialReport(cmd.Context(), req)
			if err != nil {
				return err
			}

			u.Info(fmt.Sprintf("Report created: %s (status: %s)", report.ID, report.Status))

			if !wait {
				u.Success(fmt.Sprintf("Report ID: %s - Use 'airwallex reports get %s' to check status", report.ID, report.ID))
				return nil
			}

			// Wait for completion
			u.Info("Waiting for report to complete...")
			report, err = client.WaitForReport(cmd.Context(), report.ID, time.Duration(timeout)*time.Second)
			if err != nil {
				return err
			}

			if report.Status == api.ReportStatusFailed {
				return fmt.Errorf("report failed: %s", report.ErrorMessage)
			}

			// Download
			u.Info("Downloading report...")
			content, contentType, err := client.DownloadFinancialReport(cmd.Context(), report.ID)
			if err != nil {
				return err
			}

			// Auto-generate output name if not provided
			if output == "" {
				ext := map[string]string{"CSV": ".csv", "EXCEL": ".xlsx", "PDF": ".pdf"}
				output = "balance-activity" + ext[fileFormat]
			}

			if err := os.WriteFile(output, content, 0o600); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			u.Success(fmt.Sprintf("Downloaded %s (%d bytes, %s)", output, len(content), contentType))
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDate, "from-date", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&fileFormat, "format", "CSV", "File format: CSV, EXCEL, or PDF")
	cmd.Flags().StringSliceVar(&transactionTypes, "transaction-types", nil, "Filter by types (e.g., PAYOUT,DEPOSIT)")
	cmd.Flags().StringVar(&output, "output", "", "Output filename (default: auto-generated)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for report and download")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds when waiting")

	mustMarkRequired(cmd, "from-date")
	mustMarkRequired(cmd, "to-date")

	return cmd
}

func newReportsTransactionReconCmd() *cobra.Command {
	var fromDate, toDate string
	var fileFormat string
	var transactionTypes []string
	var output string
	var wait bool
	var timeout int

	cmd := &cobra.Command{
		Use:   "transaction-recon",
		Short: "Generate transaction reconciliation report",
		Long: `Generate transaction reconciliation reports for accounting and audit.

Available formats: CSV, EXCEL

Examples:
  # Generate CSV reconciliation report
  airwallex reports transaction-recon --from-date 2024-01-01 --to-date 2024-01-31 \
    --format CSV --output recon.csv --wait

  # Generate Excel report filtered by transaction types
  airwallex reports transaction-recon --from-date 2024-01-01 --to-date 2024-01-31 \
    --format EXCEL --transaction-types PAYOUT --output recon.xlsx --wait`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate date inputs
			if err := validateDate(fromDate); err != nil {
				return fmt.Errorf("--from-date: %w", err)
			}
			if err := validateDate(toDate); err != nil {
				return fmt.Errorf("--to-date: %w", err)
			}
			if err := validateDateRange(fromDate, toDate); err != nil {
				return err
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Validate format (PDF not supported for this report type)
			validFormats := map[string]bool{"CSV": true, "EXCEL": true}
			if !validFormats[fileFormat] {
				return fmt.Errorf("--format must be CSV or EXCEL (PDF not supported for this report type)")
			}

			req := &api.CreateReportRequest{
				Type:             api.ReportTypeTransactionRecon,
				FromDate:         fromDate,
				ToDate:           toDate,
				FileFormat:       fileFormat,
				TransactionTypes: transactionTypes,
				ReportVersion:    "1.1.0",
			}

			u.Info("Creating transaction reconciliation report...")

			report, err := client.CreateFinancialReport(cmd.Context(), req)
			if err != nil {
				return err
			}

			u.Info(fmt.Sprintf("Report created: %s (status: %s)", report.ID, report.Status))

			if !wait {
				u.Success(fmt.Sprintf("Report ID: %s - Use 'airwallex reports get %s' to check status", report.ID, report.ID))
				return nil
			}

			// Wait
			u.Info("Waiting for report to complete...")
			report, err = client.WaitForReport(cmd.Context(), report.ID, time.Duration(timeout)*time.Second)
			if err != nil {
				return err
			}

			if report.Status == api.ReportStatusFailed {
				return fmt.Errorf("report failed: %s", report.ErrorMessage)
			}

			// Download
			u.Info("Downloading report...")
			content, contentType, err := client.DownloadFinancialReport(cmd.Context(), report.ID)
			if err != nil {
				return err
			}

			if output == "" {
				ext := map[string]string{"CSV": ".csv", "EXCEL": ".xlsx"}
				output = "transaction-recon" + ext[fileFormat]
			}

			if err := os.WriteFile(output, content, 0o600); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			u.Success(fmt.Sprintf("Downloaded %s (%d bytes, %s)", output, len(content), contentType))
			return nil
		},
	}

	cmd.Flags().StringVar(&fromDate, "from-date", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&fileFormat, "format", "CSV", "File format: CSV or EXCEL")
	cmd.Flags().StringSliceVar(&transactionTypes, "transaction-types", nil, "Filter by types")
	cmd.Flags().StringVar(&output, "output", "", "Output filename (default: auto-generated)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for report and download")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds when waiting")

	mustMarkRequired(cmd, "from-date")
	mustMarkRequired(cmd, "to-date")

	return cmd
}
