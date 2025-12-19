package cmd

import (
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newTransfersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfers",
		Short: "Transfer/payout operations",
	}
	cmd.AddCommand(newTransfersListCmd())
	cmd.AddCommand(newTransfersGetCmd())
	cmd.AddCommand(newTransfersCreateCmd())
	cmd.AddCommand(newTransfersCancelCmd())
	cmd.AddCommand(newTransfersConfirmationCmd())
	return cmd
}

func newTransfersListCmd() *cobra.Command {
	var status string
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transfers",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate page size (minimum 10)
			if pageSize < 10 {
				pageSize = 10
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := client.ListTransfers(cmd.Context(), status, 0, pageSize)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if len(result.Items) == 0 {
				fmt.Fprintln(os.Stderr, "No transfers found")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "TRANSFER_ID\tAMOUNT\tCURRENCY\tSTATUS\tREFERENCE")
			for _, t := range result.Items {
				fmt.Fprintf(tw, "%s\t%.2f\t%s\t%s\t%s\n",
					t.TransferID, t.TransferAmount, t.TransferCurrency, t.Status, t.Reference)
			}
			tw.Flush()

			if result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().IntVar(&pageSize, "limit", 20, "Max results (min 10)")
	return cmd
}

func newTransfersGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <transferId>",
		Short: "Get transfer details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			t, err := client.GetTransfer(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, t)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintf(tw, "transfer_id\t%s\n", t.TransferID)
			fmt.Fprintf(tw, "beneficiary_id\t%s\n", t.BeneficiaryID)
			fmt.Fprintf(tw, "transfer_amount\t%.2f\n", t.TransferAmount)
			fmt.Fprintf(tw, "transfer_currency\t%s\n", t.TransferCurrency)
			fmt.Fprintf(tw, "source_amount\t%.2f\n", t.SourceAmount)
			fmt.Fprintf(tw, "source_currency\t%s\n", t.SourceCurrency)
			fmt.Fprintf(tw, "status\t%s\n", t.Status)
			fmt.Fprintf(tw, "reference\t%s\n", t.Reference)
			fmt.Fprintf(tw, "reason\t%s\n", t.Reason)
			fmt.Fprintf(tw, "created_at\t%s\n", t.CreatedAt)
			tw.Flush()
			return nil
		},
	}
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

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new transfer",
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

			t, err := client.CreateTransfer(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, t)
			}

			u.Success(fmt.Sprintf("Created transfer: %s", t.TransferID))
			return nil
		},
	}

	cmd.Flags().StringVar(&beneficiaryID, "beneficiary-id", "", "Beneficiary ID (required)")
	cmd.Flags().Float64Var(&transferAmount, "transfer-amount", 0, "Amount beneficiary receives")
	cmd.Flags().StringVar(&transferCurrency, "transfer-currency", "", "Currency of transfer amount (required)")
	cmd.Flags().Float64Var(&sourceAmount, "source-amount", 0, "Amount to send from wallet")
	cmd.Flags().StringVar(&sourceCurrency, "source-currency", "", "Source currency (required)")
	cmd.Flags().StringVar(&transferMethod, "method", "LOCAL", "LOCAL or SWIFT")
	cmd.Flags().StringVar(&localClearingSystem, "clearing-system", "", "Clearing system (CA: EFT/INTERAC, US: ACH/FEDWIRE)")
	cmd.Flags().StringVar(&reference, "reference", "", "Reference text (required)")
	cmd.Flags().StringVar(&reason, "reason", "", "Transfer reason (required)")
	cmd.Flags().StringVar(&securityQuestion, "security-question", "", "Interac security question (1-40 chars)")
	cmd.Flags().StringVar(&securityAnswer, "security-answer", "", "Interac security answer (3-25 alphanumeric)")
	mustMarkRequired(cmd, "beneficiary-id")
	mustMarkRequired(cmd, "transfer-currency")
	mustMarkRequired(cmd, "source-currency")
	mustMarkRequired(cmd, "reference")
	mustMarkRequired(cmd, "reason")
	return cmd
}

func newTransfersCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <transferId>",
		Short: "Cancel a transfer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			t, err := client.CancelTransfer(cmd.Context(), args[0])
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
}

func newTransfersConfirmationCmd() *cobra.Command {
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "confirmation <transferId>",
		Short: "Download transfer confirmation letter as PDF",
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
			transferID := args[0]

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
