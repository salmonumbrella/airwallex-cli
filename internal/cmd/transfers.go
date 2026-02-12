package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/batch"
	"github.com/salmonumbrella/airwallex-cli/internal/dryrun"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/suggest"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newTransfersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfers",
		Aliases: []string{"transfer", "tfr", "tr", "payout", "payouts"},
		Short:   "Transfer/payout operations",
	}
	cmd.AddCommand(newTransfersListCmd())
	cmd.AddCommand(newTransfersGetCmd())
	cmd.AddCommand(newTransfersCreateCmd())
	cmd.AddCommand(newTransfersBatchCreateCmd())
	cmd.AddCommand(newTransfersCancelCmd())
	cmd.AddCommand(newTransfersConfirmationCmd())
	return cmd
}

// suggestBeneficiaries fetches beneficiaries and returns suggestions
func suggestBeneficiaries(ctx context.Context, client *api.Client, query string) string {
	result, err := client.ListBeneficiaries(ctx, 0, 50)
	if err != nil || len(result.Items) == 0 {
		return ""
	}

	var items []suggest.Match
	for _, b := range result.Items {
		label := ""
		if b.Beneficiary.CompanyName != "" {
			label = b.Beneficiary.CompanyName
		} else if b.Beneficiary.FirstName != "" {
			label = b.Beneficiary.FirstName + " " + b.Beneficiary.LastName
		}
		if b.Beneficiary.BankDetails.BankCountryCode != "" {
			label += " (" + b.Beneficiary.BankDetails.BankCountryCode + ")"
		}
		items = append(items, suggest.Match{Value: b.BeneficiaryID, Label: label})
	}

	matches := suggest.FindSimilar(query, items, 3)
	return suggest.FormatSuggestions(matches)
}

func newTransfersListCmd() *cobra.Command {
	var status string

	cmd := NewListCommand(ListConfig[api.Transfer]{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List transfers",
		Long: `List payment transfers with optional filters.

Use --output json with --query for advanced filtering using jq syntax.

Examples:
  # List recent transfers
  airwallex transfers list --page-size 20

  # Filter by status
  airwallex transfers list --status PAID

  # Sort by amount (highest first)
  airwallex transfers list --output json --query \
    '.items | sort_by(.transfer_amount) | reverse | .[0:10]'

  # Transfers over $1000
  airwallex transfers list --output json --query \
    '[.items[] | select(.transfer_amount > 1000)]'

  # Failed/pending transfers (not PAID)
  airwallex transfers list --output json --query \
    '[.items[] | select(.status != "PAID")]'

  # Total amount transferred
  airwallex transfers list --output json --query \
    '.items | map(.transfer_amount) | add'

  # Total by currency
  airwallex transfers list --output json --query \
    '.items | group_by(.transfer_currency) | map({currency: .[0].transfer_currency, total: (map(.transfer_amount) | add)})'

  # Filter by reference pattern
  airwallex transfers list --output json --query \
    '[.items[] | select(.reference | test("Invoice"; "i"))]'

  # Compact view with selected fields
  airwallex transfers list --output json --query \
    '.items[] | {ref: .reference, amount: .transfer_amount, currency: .transfer_currency, status: .status}'`,
		Headers:      []string{"TRANSFER_ID", "AMOUNT", "CURRENCY", "STATUS", "REFERENCE"},
		EmptyMessage: "No transfers found",
		ColumnTypes: []outfmt.ColumnType{
			outfmt.ColumnPlain,    // TRANSFER_ID
			outfmt.ColumnAmount,   // AMOUNT
			outfmt.ColumnCurrency, // CURRENCY
			outfmt.ColumnStatus,   // STATUS
			outfmt.ColumnPlain,    // REFERENCE
		},
		RowFunc: func(t api.Transfer) []string {
			return []string{
				t.TransferID,
				outfmt.FormatMoney(t.TransferAmount),
				t.TransferCurrency,
				t.Status,
				t.Reference,
			}
		},
		IDFunc: func(t api.Transfer) string {
			return t.TransferID
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Transfer], error) {
			// Note: API uses page-based pagination internally
			// We pass limit as page_size, page 0 for cursor-based iteration
			result, err := client.ListTransfers(ctx, status, 0, opts.Limit)
			if err != nil {
				return ListResult[api.Transfer]{}, err
			}
			return ListResult[api.Transfer]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status")
	return cmd
}

func newTransfersGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.Transfer]{
		Use:     "get <transferId>",
		Aliases: []string{"g"},
		Short:   "Get transfer details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Transfer, error) {
			return client.GetTransfer(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, t *api.Transfer) error {
			rows := []outfmt.KV{
				{Key: "transfer_id", Value: t.TransferID},
				{Key: "beneficiary_id", Value: t.BeneficiaryID},
				{Key: "transfer_amount", Value: outfmt.FormatMoney(t.TransferAmount)},
				{Key: "transfer_currency", Value: t.TransferCurrency},
				{Key: "source_amount", Value: outfmt.FormatMoney(t.SourceAmount)},
				{Key: "source_currency", Value: t.SourceCurrency},
				{Key: "status", Value: t.Status},
				{Key: "reference", Value: t.Reference},
				{Key: "reason", Value: t.Reason},
				{Key: "created_at", Value: t.CreatedAt},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newTransfersCreateCmd() *cobra.Command {
	var beneficiaryID string
	var transferAmount float64
	var transferCurrency string
	var sourceAmount float64
	var sourceCurrency string
	var transferMethod string
	var localClearingSystem string
	var reference string
	var reason string
	var securityQuestion string
	var securityAnswer string
	var dryRun bool
	var wait bool
	var waitTimeout int

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a new transfer",
		Long: `Create a new transfer/payout.

Examples:
  # Canada EFT (bank transfer)
  airwallex transfers create --beneficiary-id xxx --transfer-amount 100 \
    --transfer-currency CAD --source-currency CAD --method LOCAL \
    --reference "Invoice 123" --reason "payment_to_supplier"

  # Canada Interac e-Transfer (autodeposit enabled)
  airwallex transfers create --beneficiary-id xxx --transfer-amount 100 \
    --transfer-currency CAD --source-currency CAD --method LOCAL \
    --clearing-system INTERAC --reference "Invoice 123" --reason "payment_to_supplier"

  # Canada Interac e-Transfer (no autodeposit - requires security Q&A)
  airwallex transfers create --beneficiary-id xxx --transfer-amount 100 \
    --transfer-currency CAD --source-currency CAD --method LOCAL \
    --clearing-system INTERAC --reference "Invoice 123" --reason "payment_to_supplier" \
    --security-question "What is our company name?" --security-answer "Acme123"

  # USA ACH
  airwallex transfers create --beneficiary-id xxx --transfer-amount 100 \
    --transfer-currency USD --source-currency USD --method LOCAL \
    --clearing-system ACH --reference "Invoice 123" --reason "payment_to_supplier"

Clearing systems by country:
  Canada: EFT (default), REGULAR_EFT, INTERAC, BILL_PAYMENT
  USA:    ACH, NEXT_DAY_ACH, FEDNOW, FEDWIRE

Interac e-Transfer notes:
  If the recipient email is NOT registered with Interac autodeposit, you must
  provide --security-question and --security-answer. Share these with the
  recipient so they can claim the transfer.
  - Question: 1-40 characters
  - Answer: 3-25 alphanumeric characters (no special chars like @, &, *)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate amount fields: exactly one of transfer_amount or source_amount
			hasTransferAmount := transferAmount > 0
			hasSourceAmount := sourceAmount > 0
			if hasTransferAmount == hasSourceAmount {
				if !hasTransferAmount {
					return fmt.Errorf("must provide exactly one of --transfer-amount or --source-amount")
				}
				return fmt.Errorf("cannot provide both --transfer-amount and --source-amount")
			}

			// Validate security Q&A pairing
			hasQuestion := securityQuestion != ""
			hasAnswer := securityAnswer != ""
			if hasQuestion != hasAnswer {
				return fmt.Errorf("--security-question and --security-answer must be provided together")
			}

			// Validate security question length (1-40 characters)
			if hasQuestion && (len(securityQuestion) < 1 || len(securityQuestion) > 40) {
				return fmt.Errorf("--security-question must be 1-40 characters (got %d)", len(securityQuestion))
			}

			// Validate security answer format (3-25 alphanumeric characters only)
			if hasAnswer {
				if len(securityAnswer) < 3 || len(securityAnswer) > 25 {
					return fmt.Errorf("--security-answer must be 3-25 characters (got %d)", len(securityAnswer))
				}
				// Check for alphanumeric only (no special characters)
				alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
				if !alphanumericRegex.MatchString(securityAnswer) {
					return fmt.Errorf("--security-answer must contain only alphanumeric characters (no special chars like @, &, *)")
				}
			}

			// Normalize --method: if set to a clearing system name, convert to LOCAL + clearing system
			clearingSystems := map[string]bool{
				"INTERAC": true, "EFT": true, "REGULAR_EFT": true, "BILL_PAYMENT": true,
				"ACH": true, "NEXT_DAY_ACH": true, "FEDNOW": true, "FEDWIRE": true,
			}
			upper := strings.ToUpper(transferMethod)
			if clearingSystems[upper] {
				if localClearingSystem == "" {
					localClearingSystem = upper
				}
				transferMethod = "LOCAL"
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			req := map[string]interface{}{
				"request_id":        uuid.New().String(),
				"beneficiary_id":    beneficiaryID,
				"source_currency":   sourceCurrency,
				"transfer_currency": transferCurrency,
				"transfer_method":   transferMethod,
				"reference":         reference,
				"reason":            reason,
			}

			if transferAmount > 0 {
				req["transfer_amount"] = transferAmount
			}
			if sourceAmount > 0 {
				req["source_amount"] = sourceAmount
			}
			if localClearingSystem != "" {
				req["local_clearing_system"] = localClearingSystem
			}
			if securityQuestion != "" {
				req["security_question"] = securityQuestion
			}
			if securityAnswer != "" {
				req["security_answer"] = securityAnswer
			}

			if dryRun {
				// Fetch beneficiary details for preview
				beneficiary, err := client.GetBeneficiary(cmd.Context(), beneficiaryID)
				if err != nil {
					return fmt.Errorf("failed to fetch beneficiary for preview: %w", err)
				}

				beneficiaryName := beneficiary.Beneficiary.CompanyName
				if beneficiaryName == "" {
					beneficiaryName = strings.TrimSpace(beneficiary.Beneficiary.FirstName + " " + beneficiary.Beneficiary.LastName)
				}
				if beneficiaryName == "" {
					beneficiaryName = beneficiary.Beneficiary.BankDetails.AccountName
				}
				if beneficiaryName == "" {
					beneficiaryName = beneficiary.Nickname
				}

				// Determine which amount to show in preview
				previewAmount := transferAmount
				previewCurrency := transferCurrency
				if transferAmount == 0 && sourceAmount > 0 {
					previewAmount = sourceAmount
					previewCurrency = sourceCurrency
				}

				preview := &dryrun.Preview{
					Operation:   "create",
					Resource:    "transfer",
					Description: fmt.Sprintf("Send %s to %s", dryrun.FormatAmount(previewAmount, previewCurrency), beneficiaryName),
					Details: map[string]interface{}{
						"Beneficiary":     beneficiaryName,
						"Beneficiary ID":  beneficiaryID,
						"Amount":          dryrun.FormatAmount(previewAmount, previewCurrency),
						"Source Currency": sourceCurrency,
						"Transfer Method": transferMethod,
						"Reference":       reference,
					},
				}

				preview.Write(os.Stderr) //nolint:errcheck // preview output to stderr is best-effort
				return nil
			}

			t, err := client.CreateTransfer(cmd.Context(), req)
			if err != nil {
				if api.IsNotFoundError(err) && strings.Contains(err.Error(), "beneficiary") {
					suggestions := suggestBeneficiaries(cmd.Context(), client, beneficiaryID)
					if suggestions != "" {
						return fmt.Errorf("%w%s\nRun 'airwallex beneficiaries list' to see all beneficiaries", err, suggestions)
					}
				}
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, t)
			}

			u.Success(fmt.Sprintf("Created transfer: %s", t.TransferID))

			if wait {
				u.Info(fmt.Sprintf("Waiting for transfer %s to complete...", t.TransferID))
				t, err = client.WaitForTransfer(cmd.Context(), t.TransferID, time.Duration(waitTimeout)*time.Second)
				if err != nil {
					return err
				}

				if t.Status == "FAILED" {
					return fmt.Errorf("transfer failed")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&beneficiaryID, "beneficiary-id", "b", "", "Beneficiary ID (required)")
	cmd.Flags().Float64Var(&transferAmount, "transfer-amount", 0, "Amount beneficiary receives")
	cmd.Flags().StringVar(&transferCurrency, "transfer-currency", "", "Currency of transfer amount (required)")
	cmd.Flags().Float64Var(&sourceAmount, "source-amount", 0, "Amount to send from wallet")
	cmd.Flags().StringVar(&sourceCurrency, "source-currency", "", "Source currency (required)")
	cmd.Flags().StringVarP(&transferMethod, "method", "m", "LOCAL", "LOCAL, SWIFT, or a clearing system (INTERAC, ACH, FEDWIRE, etc.)")
	cmd.Flags().StringVar(&localClearingSystem, "clearing-system", "", "Clearing system (CA: EFT/INTERAC, US: ACH/FEDWIRE)")
	cmd.Flags().StringVarP(&reference, "reference", "r", "", "Reference text (required)")
	cmd.Flags().StringVar(&reason, "reason", "", "Transfer reason (required)")
	cmd.Flags().StringVar(&securityQuestion, "security-question", "", "Interac security question (1-40 chars)")
	cmd.Flags().StringVar(&securityAnswer, "security-answer", "", "Interac security answer (3-25 alphanumeric)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview the transfer without executing")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "Wait for transfer to complete")
	cmd.Flags().IntVar(&waitTimeout, "timeout", 300, "Timeout in seconds when waiting")
	mustMarkRequired(cmd, "beneficiary-id")
	mustMarkRequired(cmd, "transfer-currency")
	mustMarkRequired(cmd, "source-currency")
	mustMarkRequired(cmd, "reference")
	mustMarkRequired(cmd, "reason")
	return cmd
}

func newTransfersBatchCreateCmd() *cobra.Command {
	var fromFile string
	var continueOnError bool

	cmd := &cobra.Command{
		Use:     "batch-create",
		Aliases: []string{"bc"},
		Short:   "Create multiple transfers from file or stdin",
		Long: `Create multiple transfers from a JSON file or stdin.

Input format (JSON array or newline-delimited JSON):
[
  {
    "beneficiary_id": "ben_xxx",
    "transfer_amount": 100.00,
    "transfer_currency": "USD",
    "source_currency": "USD",
    "transfer_method": "LOCAL",
    "reference": "INV-001",
    "reason": "payment_to_supplier"
  }
]

Examples:
  airwallex transfers batch-create --from-file transfers.json
  cat transfers.json | airwallex transfers batch-create
  airwallex transfers batch-create --from-file transfers.json --continue-on-error`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			items, err := batch.ReadItems(fromFile)
			if err != nil {
				return err
			}

			u.Info(fmt.Sprintf("Processing %d transfers...", len(items)))

			var results []batch.Result
			var summary batch.Summary
			summary.Total = len(items)

			for i, item := range items {
				if _, ok := item["request_id"]; !ok {
					item["request_id"] = uuid.New().String()
				}

				t, err := client.CreateTransfer(cmd.Context(), item)
				if err != nil {
					results = append(results, batch.Result{
						Index:   i,
						Success: false,
						Error:   err.Error(),
						Input:   item,
					})
					summary.Failed++

					if !continueOnError {
						break
					}
					continue
				}

				results = append(results, batch.Result{
					Index:   i,
					Success: true,
					ID:      t.TransferID,
				})
				summary.Success++
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, map[string]interface{}{
					"results": results,
					"summary": summary,
				})
			}

			u.Info(fmt.Sprintf("Completed: %d success, %d failed", summary.Success, summary.Failed))
			for _, r := range results {
				if r.Success {
					u.Success(fmt.Sprintf("[%d] Created: %s", r.Index, r.ID))
				} else {
					u.Error(fmt.Sprintf("[%d] Failed: %s", r.Index, r.Error))
				}
			}

			if summary.Failed > 0 {
				return fmt.Errorf("%d transfers failed", summary.Failed)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&fromFile, "from-file", "F", "", "JSON file with transfers (- for stdin)")
	cmd.Flags().BoolVar(&continueOnError, "continue-on-error", false, "Continue processing on errors")

	return cmd
}

func newTransfersCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel <transferId>",
		Aliases: []string{"x"},
		Short:   "Cancel a transfer",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			transferID := NormalizeIDArg(args[0])

			// Prompt for confirmation (respects --yes flag and TTY detection)
			prompt := fmt.Sprintf("Are you sure you want to cancel transfer %s?", transferID)
			confirmed, err := ConfirmOrYes(cmd.Context(), prompt)
			if err != nil {
				return err
			}
			if !confirmed {
				u.Info("Operation cancelled.")
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			t, err := client.CancelTransfer(cmd.Context(), transferID)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, t)
			}

			u.Success(fmt.Sprintf("Cancelled transfer: %s", t.TransferID))
			return nil
		},
	}

	return cmd
}

func newTransfersConfirmationCmd() *cobra.Command {
	var format string
	var output string

	cmd := &cobra.Command{
		Use:     "confirmation <transferId>",
		Aliases: []string{"conf"},
		Short:   "Download transfer confirmation letter as PDF",
		Long: `Download a PDF confirmation letter for a completed transfer.

Examples:
  # Download with standard format (includes fees)
  airwallex transfers confirmation tfr_xxx --output confirmation.pdf

  # Download without fee display
  airwallex transfers confirmation tfr_xxx --output confirmation.pdf --format NO_FEE_DISPLAY

Format options:
  STANDARD         - Includes transfer fees in the confirmation letter (default)
  NO_FEE_DISPLAY   - Excludes transfer fees from the confirmation letter`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			transferID := NormalizeIDArg(args[0])

			if format != "STANDARD" && format != "NO_FEE_DISPLAY" {
				return fmt.Errorf("invalid format: %s (must be STANDARD or NO_FEE_DISPLAY)", format)
			}

			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			pdfData, err := client.GetConfirmationLetter(cmd.Context(), transferID, format)
			if err != nil {
				return err
			}

			if err := os.WriteFile(output, pdfData, 0o600); err != nil {
				return fmt.Errorf("failed to write PDF file: %w", err)
			}

			u.Success(fmt.Sprintf("Downloaded confirmation letter to: %s", output))
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "STANDARD", "Format type (STANDARD or NO_FEE_DISPLAY)")
	cmd.Flags().StringVar(&output, "output", "", "Output filename (required)")
	mustMarkRequired(cmd, "output")
	return cmd
}
