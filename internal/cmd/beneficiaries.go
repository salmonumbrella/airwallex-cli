package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/flagmap"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/schemavalidator"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newBeneficiariesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "beneficiaries",
		Aliases: []string{"benef"},
		Short:   "Beneficiary management",
	}
	cmd.AddCommand(newBeneficiariesListCmd())
	cmd.AddCommand(newBeneficiariesGetCmd())
	cmd.AddCommand(newBeneficiariesCreateCmd())
	cmd.AddCommand(newBeneficiariesUpdateCmd())
	cmd.AddCommand(newBeneficiariesDeleteCmd())
	cmd.AddCommand(newBeneficiariesValidateCmd())
	return cmd
}

func newBeneficiariesListCmd() *cobra.Command {
	return NewListCommand(ListConfig[api.Beneficiary]{
		Use:          "list",
		Short:        "List beneficiaries",
		Headers:      []string{"BENEFICIARY_ID", "TYPE", "NAME", "BANK_COUNTRY", "METHODS"},
		EmptyMessage: "No beneficiaries found",
		RowFunc: func(b api.Beneficiary) []string {
			name := b.Nickname
			if name == "" {
				name = b.Beneficiary.BankDetails.AccountName
			}
			methods := ""
			if len(b.TransferMethods) > 0 {
				methods = b.TransferMethods[0]
			}
			return []string{
				b.BeneficiaryID,
				b.Beneficiary.EntityType,
				name,
				b.Beneficiary.BankDetails.BankCountryCode,
				methods,
			}
		},
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.Beneficiary], error) {
			result, err := client.ListBeneficiaries(ctx, page, pageSize)
			if err != nil {
				return ListResult[api.Beneficiary]{}, err
			}
			return ListResult[api.Beneficiary]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newBeneficiariesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <beneficiaryId>",
		Short: "Get beneficiary details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			b, err := client.GetBeneficiary(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if outfmt.IsJSON(cmd.Context()) {
				return f.Output(b)
			}

			// For "get" commands, still use manual tabwriter for key-value format
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "beneficiary_id\t%s\n", b.BeneficiaryID)
			_, _ = fmt.Fprintf(tw, "nickname\t%s\n", b.Nickname)
			_, _ = fmt.Fprintf(tw, "entity_type\t%s\n", b.Beneficiary.EntityType)
			if b.Beneficiary.CompanyName != "" {
				_, _ = fmt.Fprintf(tw, "company_name\t%s\n", b.Beneficiary.CompanyName)
			}
			if b.Beneficiary.FirstName != "" {
				_, _ = fmt.Fprintf(tw, "first_name\t%s\n", b.Beneficiary.FirstName)
				_, _ = fmt.Fprintf(tw, "last_name\t%s\n", b.Beneficiary.LastName)
			}
			_, _ = fmt.Fprintf(tw, "bank_country\t%s\n", b.Beneficiary.BankDetails.BankCountryCode)
			_, _ = fmt.Fprintf(tw, "bank_name\t%s\n", b.Beneficiary.BankDetails.BankName)
			_, _ = fmt.Fprintf(tw, "account_name\t%s\n", b.Beneficiary.BankDetails.AccountName)
			_ = tw.Flush()
			return nil
		},
	}
}

func newBeneficiariesCreateCmd() *cobra.Command {
	var entityType string
	var bankCountry string
	var companyName string
	var firstName string
	var lastName string
	var nickname string
	var transferMethod string
	var accountCurrency string
	var accountName string
	var accountNumber string
	var institutionNumber string
	var transitNumber string
	var email string
	var phone string
	var localClearingSystem string
	// SWIFT/international routing
	var swiftCode string
	var routingNumber string
	var iban string
	// Additional international routing flags
	var sortCode string
	var bsb string
	var ifsc string
	var clabe string
	var bankCode string
	var branchCode string
	// Address fields (required for Interac)
	var addressCountry string
	var addressStreet string
	var addressCity string
	var addressState string
	var addressPostcode string
	// Validation mode
	var validateOnly bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new beneficiary",
		Long: `Create a new beneficiary for payouts.

Examples:
  # US SWIFT (international wire)
  airwallex beneficiaries create --entity-type COMPANY --bank-country US \
    --company-name "Acme Corp" --account-name "Acme Corp" \
    --account-currency USD --account-number 123456789 \
    --swift-code CHASUS33 --transfer-method SWIFT

  # US ACH (domestic)
  airwallex beneficiaries create --entity-type COMPANY --bank-country US \
    --company-name "Acme Corp" --account-name "Acme Corp" \
    --account-currency USD --account-number 123456789 \
    --routing-number 021000021

  # Europe IBAN/SWIFT
  airwallex beneficiaries create --entity-type COMPANY --bank-country DE \
    --company-name "GmbH Corp" --account-name "GmbH Corp" \
    --account-currency EUR --iban DE89370400440532013000 \
    --swift-code COBADEFFXXX --transfer-method SWIFT

  # UK with Sort Code
  airwallex beneficiaries create --entity-type COMPANY --bank-country GB \
    --company-name "UK Ltd" --account-name "UK Ltd" \
    --account-currency GBP --account-number 12345678 \
    --sort-code 123456

  # Australia with BSB
  airwallex beneficiaries create --entity-type PERSONAL --bank-country AU \
    --first-name Jane --last-name Smith --account-name "Jane Smith" \
    --account-currency AUD --account-number 123456789 \
    --bsb 123456

  # India with IFSC
  airwallex beneficiaries create --entity-type PERSONAL --bank-country IN \
    --first-name Raj --last-name Kumar --account-name "Raj Kumar" \
    --account-currency INR --account-number 1234567890 \
    --ifsc HDFC0001234

  # Mexico with CLABE
  airwallex beneficiaries create --entity-type COMPANY --bank-country MX \
    --company-name "Mexico SA" --account-name "Mexico SA" \
    --account-currency MXN --clabe 012345678901234567

  # Canada EFT (bank transfer)
  airwallex beneficiaries create --entity-type PERSONAL --bank-country CA \
    --first-name John --last-name Doe --account-name "John Doe" \
    --account-currency CAD --account-number 1234567 \
    --institution-number 001 --transit-number 12345

  # Canada Interac e-Transfer (email)
  airwallex beneficiaries create --entity-type PERSONAL --bank-country CA \
    --first-name John --last-name Doe --account-name "John Doe" \
    --account-currency CAD --email john@example.com --clearing-system INTERAC`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Validation: Required fields based on entity type
			if accountName == "" {
				return fmt.Errorf("--account-name is required")
			}
			if accountCurrency == "" {
				return fmt.Errorf("--account-currency is required")
			}

			switch entityType {
			case "COMPANY":
				if companyName == "" {
					return fmt.Errorf("--company-name is required when entity-type is COMPANY")
				}
			case "PERSONAL":
				if firstName == "" {
					return fmt.Errorf("--first-name is required when entity-type is PERSONAL")
				}
				if lastName == "" {
					return fmt.Errorf("--last-name is required when entity-type is PERSONAL")
				}
			}

			// Validation: Must provide at least one routing method
			hasEmail := email != ""
			hasPhone := phone != ""
			hasEFT := institutionNumber != ""
			hasSWIFT := swiftCode != ""
			hasRouting := routingNumber != ""
			hasIBAN := iban != ""
			hasSortCode := sortCode != ""
			hasBSB := bsb != ""
			hasIFSC := ifsc != ""
			hasCLABE := clabe != ""
			hasBankCode := bankCode != ""

			hasAnyRouting := hasEmail || hasPhone || hasEFT || hasSWIFT || hasRouting ||
				hasIBAN || hasSortCode || hasBSB || hasIFSC || hasCLABE || hasBankCode

			if !hasAnyRouting {
				return fmt.Errorf("must provide at least one routing method (e.g., --swift-code, --iban, --routing-number, --sort-code, --bsb)")
			}

			// Validation: Canada EFT requires both institution and transit numbers
			if institutionNumber != "" && transitNumber == "" {
				return fmt.Errorf("--transit-number is required when --institution-number is provided")
			}

			// Validation: Phone number format
			if phone != "" {
				phoneRegex := regexp.MustCompile(`^\+1-\d{10}$`)
				if !phoneRegex.MatchString(phone) {
					return fmt.Errorf("--phone must match format +1-nnnnnnnnnn (e.g., +1-4165551234)")
				}
			}

			// Validation: Institution number format
			if institutionNumber != "" {
				instRegex := regexp.MustCompile(`^\d{3}$`)
				if !instRegex.MatchString(institutionNumber) {
					return fmt.Errorf("--institution-number must be exactly 3 digits")
				}
			}

			// Validation: Transit number format
			if transitNumber != "" {
				transitRegex := regexp.MustCompile(`^\d{5}$`)
				if !transitRegex.MatchString(transitNumber) {
					return fmt.Errorf("--transit-number must be exactly 5 digits")
				}
			}

			// Validation: Email format
			if email != "" {
				if !strings.Contains(email, "@") {
					return fmt.Errorf("--email must be a valid email address")
				}
				parts := strings.Split(email, "@")
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return fmt.Errorf("--email must be a valid email address")
				}
			}

			// Validation: Routing number format (US ABA - 9 digits)
			if routingNumber != "" {
				abaRegex := regexp.MustCompile(`^\d{9}$`)
				if !abaRegex.MatchString(routingNumber) {
					return fmt.Errorf("--routing-number must be exactly 9 digits")
				}
			}

			// Validation: Sort code format (UK - 6 digits)
			if sortCode != "" {
				sortCodeRegex := regexp.MustCompile(`^\d{6}$`)
				if !sortCodeRegex.MatchString(sortCode) {
					return fmt.Errorf("--sort-code must be exactly 6 digits")
				}
			}

			// Validation: BSB format (Australia - 6 digits)
			if bsb != "" {
				bsbRegex := regexp.MustCompile(`^\d{6}$`)
				if !bsbRegex.MatchString(bsb) {
					return fmt.Errorf("--bsb must be exactly 6 digits")
				}
			}

			// Validation: CLABE format (Mexico - 18 digits)
			if clabe != "" {
				clabeRegex := regexp.MustCompile(`^\d{18}$`)
				if !clabeRegex.MatchString(clabe) {
					return fmt.Errorf("--clabe must be exactly 18 digits")
				}
			}

			// Validation: IFSC format (India - 11 chars: 4 letters, 0, 6 alphanumeric)
			if ifsc != "" {
				ifscRegex := regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
				if !ifscRegex.MatchString(strings.ToUpper(ifsc)) {
					return fmt.Errorf("--ifsc must be 11 characters: 4 letters, 0, then 6 alphanumeric (e.g., SBIN0001234)")
				}
			}

			// Build beneficiary object
			beneficiary := map[string]interface{}{
				"entity_type": entityType,
			}
			if companyName != "" {
				beneficiary["company_name"] = companyName
			}
			if firstName != "" {
				beneficiary["first_name"] = firstName
			}
			if lastName != "" {
				beneficiary["last_name"] = lastName
			}

			// Build address (required for Interac)
			if addressCountry != "" || addressStreet != "" || addressCity != "" {
				address := map[string]interface{}{}
				if addressCountry != "" {
					address["country_code"] = addressCountry
				}
				if addressStreet != "" {
					address["street_address"] = addressStreet
				}
				if addressCity != "" {
					address["city"] = addressCity
				}
				if addressState != "" {
					address["state"] = addressState
				}
				if addressPostcode != "" {
					address["postcode"] = addressPostcode
				}
				beneficiary["address"] = address
			}

			// Build bank_details
			bankDetails := map[string]interface{}{
				"bank_country_code": bankCountry,
			}
			if accountCurrency != "" {
				bankDetails["account_currency"] = accountCurrency
			}
			if accountName != "" {
				bankDetails["account_name"] = accountName
			}
			if accountNumber != "" {
				bankDetails["account_number"] = accountNumber
			}
			if localClearingSystem != "" {
				bankDetails["local_clearing_system"] = localClearingSystem
			}

			// Handle routing - SWIFT/international first
			if swiftCode != "" {
				bankDetails["swift_code"] = swiftCode
			}
			if iban != "" {
				bankDetails["iban"] = iban
			}
			if clabe != "" {
				bankDetails["clabe"] = clabe
			}

			// Set account_routing_type1/value1 based on provided flag
			if routingNumber != "" {
				bankDetails["account_routing_type1"] = "aba"
				bankDetails["account_routing_value1"] = routingNumber
			} else if sortCode != "" {
				bankDetails["account_routing_type1"] = "sort_code"
				bankDetails["account_routing_value1"] = sortCode
			} else if bsb != "" {
				bankDetails["account_routing_type1"] = "bsb"
				bankDetails["account_routing_value1"] = bsb
			} else if ifsc != "" {
				bankDetails["account_routing_type1"] = "ifsc"
				bankDetails["account_routing_value1"] = ifsc
			} else if bankCode != "" {
				bankDetails["account_routing_type1"] = "bank_code"
				bankDetails["account_routing_value1"] = bankCode
			} else if email != "" {
				bankDetails["account_routing_type1"] = "email_address"
				bankDetails["account_routing_value1"] = email
			} else if phone != "" {
				bankDetails["account_routing_type1"] = "phone_number"
				bankDetails["account_routing_value1"] = phone
			} else if institutionNumber != "" {
				bankDetails["account_routing_type1"] = "institution_number"
				bankDetails["account_routing_value1"] = institutionNumber
				if transitNumber != "" {
					bankDetails["account_routing_type2"] = "transit_number"
					bankDetails["account_routing_value2"] = transitNumber
				}
			}

			if branchCode != "" {
				bankDetails["branch_code"] = branchCode
			}

			beneficiary["bank_details"] = bankDetails

			// Build request
			req := map[string]interface{}{
				"beneficiary":      beneficiary,
				"transfer_methods": []string{transferMethod},
				"payment_methods":  []string{transferMethod},
			}
			if nickname != "" {
				req["nickname"] = nickname
			}

			// Optional: Fetch schema and validate
			if validateOnly {
				schema, err := client.GetBeneficiarySchema(cmd.Context(), bankCountry, entityType, transferMethod)
				if err != nil {
					return fmt.Errorf("failed to fetch schema: %w", err)
				}

				// Build provided fields map for validation using flagmap
				provided := make(map[string]string)

				// Helper to add a field using flagmap
				addField := func(flagName, value string) {
					if value != "" {
						if m, ok := flagmap.GetMapping(flagName); ok {
							provided[m.SchemaPath] = value
						}
					}
				}

				// Account details
				addField("account-name", accountName)
				addField("account-number", accountNumber)
				addField("account-currency", accountCurrency)

				// SWIFT/International routing
				addField("swift-code", swiftCode)
				addField("iban", iban)
				addField("clabe", clabe)

				// Country-specific routing
				addField("routing-number", routingNumber)
				addField("sort-code", sortCode)
				addField("bsb", bsb)
				addField("ifsc", ifsc)
				addField("bank-code", bankCode)
				addField("branch-code", branchCode)

				// Canada EFT
				addField("institution-number", institutionNumber)
				addField("transit-number", transitNumber)

				// Email/phone routing (not in flagmap, use direct paths)
				if email != "" {
					provided["beneficiary.bank_details.account_routing_value1"] = email
				}
				if phone != "" {
					provided["beneficiary.bank_details.account_routing_value1"] = phone
				}

				// Entity details
				addField("company-name", companyName)
				addField("first-name", firstName)
				addField("last-name", lastName)

				// Address fields
				addField("address-country", addressCountry)
				addField("address-street", addressStreet)
				addField("address-city", addressCity)
				addField("address-state", addressState)
				addField("address-postcode", addressPostcode)

				// Validate using schemavalidator package
				missing, err := schemavalidator.Validate(schema, provided)
				if err != nil {
					return fmt.Errorf("validation error: %w", err)
				}

				if len(missing) > 0 {
					return fmt.Errorf("%s", schemavalidator.FormatMissingFields(missing))
				}

				// Show what would be sent
				u.Success("Schema validation passed")
				if outfmt.IsJSON(cmd.Context()) {
					return outfmt.WriteJSON(os.Stdout, req)
				}
				u.Info(fmt.Sprintf("Would create beneficiary in %s with %s routing", bankCountry, transferMethod))
				return nil
			}

			b, err := client.CreateBeneficiary(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, b)
			}

			u.Success(fmt.Sprintf("Created beneficiary: %s", b.BeneficiaryID))
			return nil
		},
	}

	// Required flags
	cmd.Flags().StringVar(&entityType, "entity-type", "", "COMPANY or PERSONAL (required)")
	cmd.Flags().StringVar(&bankCountry, "bank-country", "", "Bank country code e.g. CA, US (required)")
	cmd.Flags().StringVar(&transferMethod, "transfer-method", "LOCAL", "Transfer method: LOCAL or SWIFT")
	cmd.Flags().StringVar(&accountCurrency, "account-currency", "", "Currency e.g. CAD, USD (required)")
	cmd.Flags().StringVar(&accountName, "account-name", "", "Account holder name (required)")

	// Name flags
	cmd.Flags().StringVar(&companyName, "company-name", "", "Company name (for COMPANY entity)")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name (for PERSONAL entity)")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name (for PERSONAL entity)")
	cmd.Flags().StringVar(&nickname, "nickname", "", "Nickname for the beneficiary")

	// Bank account flags (EFT)
	cmd.Flags().StringVar(&accountNumber, "account-number", "", "Bank account number")
	cmd.Flags().StringVar(&institutionNumber, "institution-number", "", "Institution number (Canada: 3 digits)")
	cmd.Flags().StringVar(&transitNumber, "transit-number", "", "Transit/branch number (Canada: 5 digits)")

	// Interac e-Transfer flags
	cmd.Flags().StringVar(&email, "email", "", "Email for Interac e-Transfer")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone for Interac e-Transfer (format: +1-nnnnnnnnnn)")
	cmd.Flags().StringVar(&localClearingSystem, "clearing-system", "", "Clearing system: EFT, REGULAR_EFT, INTERAC, etc.")

	// SWIFT/international routing flags
	cmd.Flags().StringVar(&swiftCode, "swift-code", "", "SWIFT/BIC code for international transfers")
	cmd.Flags().StringVar(&routingNumber, "routing-number", "", "US ABA routing number (9 digits)")
	cmd.Flags().StringVar(&iban, "iban", "", "IBAN for European/international transfers")

	// Additional international routing flags
	cmd.Flags().StringVar(&sortCode, "sort-code", "", "UK sort code (6 digits)")
	cmd.Flags().StringVar(&bsb, "bsb", "", "Australian BSB number (6 digits)")
	cmd.Flags().StringVar(&ifsc, "ifsc", "", "Indian IFSC code")
	cmd.Flags().StringVar(&clabe, "clabe", "", "Mexican CLABE (18 digits)")
	cmd.Flags().StringVar(&bankCode, "bank-code", "", "Generic bank code")
	cmd.Flags().StringVar(&branchCode, "branch-code", "", "Generic branch code")

	// Address flags (required for Interac)
	cmd.Flags().StringVar(&addressCountry, "address-country", "", "Beneficiary country code (e.g. CA)")
	cmd.Flags().StringVar(&addressStreet, "address-street", "", "Beneficiary street address")
	cmd.Flags().StringVar(&addressCity, "address-city", "", "Beneficiary city")
	cmd.Flags().StringVar(&addressState, "address-state", "", "Beneficiary state/province")
	cmd.Flags().StringVar(&addressPostcode, "address-postcode", "", "Beneficiary postal code")

	// Validation mode flag
	cmd.Flags().BoolVar(&validateOnly, "validate", false, "Validate against schema without creating")

	mustMarkRequired(cmd, "entity-type")
	mustMarkRequired(cmd, "bank-country")
	return cmd
}

func newBeneficiariesUpdateCmd() *cobra.Command {
	var nickname string
	var companyName string
	var firstName string
	var lastName string
	// Address fields
	var addressCountry string
	var addressStreet string
	var addressCity string
	var addressState string
	var addressPostcode string

	cmd := &cobra.Command{
		Use:   "update <beneficiaryId>",
		Short: "Update beneficiary (nickname, names, address)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Check if any updates were specified
			hasUpdates := cmd.Flags().Changed("nickname") ||
				cmd.Flags().Changed("company-name") ||
				cmd.Flags().Changed("first-name") ||
				cmd.Flags().Changed("last-name") ||
				cmd.Flags().Changed("address-country") ||
				cmd.Flags().Changed("address-street") ||
				cmd.Flags().Changed("address-city") ||
				cmd.Flags().Changed("address-state") ||
				cmd.Flags().Changed("address-postcode")

			if !hasUpdates {
				return fmt.Errorf("no updates specified")
			}

			// Fetch existing beneficiary data
			existing, err := client.GetBeneficiaryRaw(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to fetch existing beneficiary: %w", err)
			}

			// Remove id field - API doesn't want it in update request
			delete(existing, "id")

			// Apply updates to top-level fields
			if cmd.Flags().Changed("nickname") {
				existing["nickname"] = nickname
			}

			// Get or create beneficiary object
			beneficiaryObj, ok := existing["beneficiary"].(map[string]interface{})
			if !ok {
				beneficiaryObj = make(map[string]interface{})
			}

			// Apply name updates
			if cmd.Flags().Changed("company-name") {
				beneficiaryObj["company_name"] = companyName
			}
			if cmd.Flags().Changed("first-name") {
				beneficiaryObj["first_name"] = firstName
			}
			if cmd.Flags().Changed("last-name") {
				beneficiaryObj["last_name"] = lastName
			}

			// Apply address updates
			addressObj, ok := beneficiaryObj["address"].(map[string]interface{})
			if !ok {
				addressObj = make(map[string]interface{})
			}
			if cmd.Flags().Changed("address-country") {
				addressObj["country_code"] = addressCountry
			}
			if cmd.Flags().Changed("address-street") {
				addressObj["street_address"] = addressStreet
			}
			if cmd.Flags().Changed("address-city") {
				addressObj["city"] = addressCity
			}
			if cmd.Flags().Changed("address-state") {
				addressObj["state"] = addressState
			}
			if cmd.Flags().Changed("address-postcode") {
				addressObj["postcode"] = addressPostcode
			}
			beneficiaryObj["address"] = addressObj
			existing["beneficiary"] = beneficiaryObj

			b, err := client.UpdateBeneficiary(cmd.Context(), args[0], existing)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, b)
			}

			u.Success(fmt.Sprintf("Updated beneficiary: %s", b.BeneficiaryID))
			return nil
		},
	}

	cmd.Flags().StringVar(&nickname, "nickname", "", "Beneficiary nickname")
	cmd.Flags().StringVar(&companyName, "company-name", "", "Company name")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	// Address flags
	cmd.Flags().StringVar(&addressCountry, "address-country", "", "Beneficiary country code (e.g. CA)")
	cmd.Flags().StringVar(&addressStreet, "address-street", "", "Beneficiary street address")
	cmd.Flags().StringVar(&addressCity, "address-city", "", "Beneficiary city")
	cmd.Flags().StringVar(&addressState, "address-state", "", "Beneficiary state/province")
	cmd.Flags().StringVar(&addressPostcode, "address-postcode", "", "Beneficiary postal code")
	return cmd
}

func newBeneficiariesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <beneficiaryId>",
		Short: "Delete a beneficiary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			beneficiaryID := args[0]

			// Prompt for confirmation (respects --yes flag and TTY detection)
			prompt := fmt.Sprintf("Are you sure you want to delete beneficiary %s?", beneficiaryID)
			confirmed, err := ConfirmOrYes(cmd.Context(), prompt)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Deletion cancelled.")
				return nil
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			if err := client.DeleteBeneficiary(cmd.Context(), beneficiaryID); err != nil {
				return err
			}

			u.Success(fmt.Sprintf("Deleted beneficiary: %s", beneficiaryID))
			return nil
		},
	}

	return cmd
}

func newBeneficiariesValidateCmd() *cobra.Command {
	var entityType string
	var bankCountry string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate beneficiary details",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			req := map[string]interface{}{
				"entity_type":       entityType,
				"bank_country_code": bankCountry,
			}

			if err := client.ValidateBeneficiary(cmd.Context(), req); err != nil {
				return err
			}

			u.Success("Beneficiary details are valid")
			return nil
		},
	}

	cmd.Flags().StringVar(&entityType, "entity-type", "", "COMPANY or PERSONAL (required)")
	cmd.Flags().StringVar(&bankCountry, "bank-country", "", "Bank country code (required)")
	mustMarkRequired(cmd, "entity-type")
	mustMarkRequired(cmd, "bank-country")
	return cmd
}
