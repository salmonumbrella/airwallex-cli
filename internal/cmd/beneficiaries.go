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
		Use:   "list",
		Short: "List beneficiaries",
		Long: `List beneficiaries for payouts.

Use --output json with --query for advanced filtering using jq syntax.

Examples:
  # List recent beneficiaries
  airwallex beneficiaries list --limit 20

  # Filter by nickname (case-insensitive) and show key fields
  airwallex beneficiaries list --output json --query \
    '.items[] | select((.nickname // "") | test("Jason|Jing Sen|Huang"; "i")) | {id: .id, nickname: .nickname, account_name: .beneficiary.bank_details.account_name}'`,
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
		IDFunc: func(b api.Beneficiary) string {
			return b.BeneficiaryID
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.Beneficiary], error) {
			result, err := client.ListBeneficiaries(ctx, 0, opts.Limit)
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
	var paymentMethod string
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
	// Japan Zengin
	var zenginBankCode string
	var zenginBranchCode string
	var bankAccountCategory string
	// China
	var cnaps string
	// South Korea
	var koreaBankCode string
	// Brazil
	var cpf string
	var cnpj string
	var bankBranch string
	// Singapore PayNow
	var paynowVPA string
	var uen string
	var nric string
	var sgBankCode string
	// Sweden
	var clearingNumber string
	// Hong Kong FPS
	var hkBankCode string
	var fpsID string
	var hkid string
	// Australia PayID
	var payidPhone string
	var payidEmail string
	var payidABN string
	// China legal representative
	var legalRepFirstName string
	var legalRepLastName string
	var legalRepID string
	var bankName string
	var personalIDType string
	var personalIDNumber string
	var businessRegNumber string
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
    --swift-code CHASUS33 --payment-method SWIFT

  # US ACH (domestic)
  airwallex beneficiaries create --entity-type COMPANY --bank-country US \
    --company-name "Acme Corp" --account-name "Acme Corp" \
    --account-currency USD --account-number 123456789 \
    --routing-number 021000021

  # Europe IBAN/SWIFT
  airwallex beneficiaries create --entity-type COMPANY --bank-country DE \
    --company-name "GmbH Corp" --account-name "GmbH Corp" \
    --account-currency EUR --iban DE89370400440532013000 \
    --swift-code COBADEFFXXX --payment-method SWIFT

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
    --account-currency CAD --email john@example.com --clearing-system INTERAC

  # Japan with Zengin routing
  airwallex beneficiaries create --entity-type PERSONAL --bank-country JP \
    --first-name Taro --last-name Yamada --account-name "Yamada Taro" \
    --account-currency JPY --account-number 1234567 \
    --zengin-bank-code 0001 --zengin-branch-code 001 \
    --account-category Savings

  # China with CNAPS
  airwallex beneficiaries create --entity-type PERSONAL --bank-country CN \
    --first-name Wei --last-name Zhang --account-name "Zhang Wei" \
    --account-currency CNY --account-number 6222021234567890123 \
    --cnaps 102100099996 --bank-name "Industrial and Commercial Bank" \
    --personal-id-type CHINESE_NATIONAL_ID --personal-id-number 310101199001011234

  # Brazil with CPF
  airwallex beneficiaries create --entity-type PERSONAL --bank-country BR \
    --first-name João --last-name Silva --account-name "João Silva" \
    --account-currency BRL --account-number 123456789 \
    --swift-code BRASBRRJ --cpf 12345678901 --bank-branch 1234

  # South Korea
  airwallex beneficiaries create --entity-type PERSONAL --bank-country KR \
    --first-name Min --last-name Kim --account-name "Kim Min" \
    --account-currency KRW --account-number 1234567890123 \
    --korea-bank-code 004

  # Singapore with PayNow NRIC
  airwallex beneficiaries create --entity-type PERSONAL --bank-country SG \
    --first-name Wei --last-name Tan --account-name "Tan Wei" \
    --account-currency SGD --nric S1234567A

  # Hong Kong with FPS
  airwallex beneficiaries create --entity-type PERSONAL --bank-country HK \
    --first-name Wing --last-name Chan --account-name "Chan Wing" \
    --account-currency HKD --account-number 12345678901234 \
    --hk-bank-code 004

  # Australia PayID (phone)
  airwallex beneficiaries create --entity-type PERSONAL \
    --bank-country AU --account-currency AUD \
    --payid-phone "+61-412345678" --account-name "Jane Smith" \
    --first-name Jane --last-name Smith

  # Australia PayID (email)
  airwallex beneficiaries create --entity-type PERSONAL \
    --bank-country AU --account-currency AUD \
    --payid-email "jane@example.com" --account-name "Jane Smith" \
    --first-name Jane --last-name Smith

  # Australia PayID (ABN for business)
  airwallex beneficiaries create --entity-type COMPANY \
    --bank-country AU --account-currency AUD \
    --payid-abn "12345678901" --account-name "Acme Pty Ltd" \
    --company-name "Acme Pty Ltd"

  # Sweden with clearing number
  airwallex beneficiaries create --entity-type PERSONAL --bank-country SE \
    --first-name Erik --last-name Svensson --account-name "Erik Svensson" \
    --account-currency SEK --account-number 123456789012345 \
    --clearing-number 1234`,
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
			hasZengin := zenginBankCode != ""
			hasCNAPS := cnaps != ""
			hasKorea := koreaBankCode != ""
			hasPayNow := paynowVPA != "" || uen != "" || nric != "" || sgBankCode != ""
			hasClearing := clearingNumber != ""
			hasFPS := hkBankCode != "" || fpsID != "" || hkid != ""
			hasPayID := payidPhone != "" || payidEmail != "" || payidABN != ""

			hasAnyRouting := hasEmail || hasPhone || hasEFT || hasSWIFT || hasRouting ||
				hasIBAN || hasSortCode || hasBSB || hasIFSC || hasCLABE || hasBankCode || hasZengin || hasCNAPS || hasKorea || hasPayNow || hasClearing || hasFPS || hasPayID

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

			// Validation: Japan Zengin bank code (4 digits)
			if zenginBankCode != "" {
				zenginBankRegex := regexp.MustCompile(`^\d{4}$`)
				if !zenginBankRegex.MatchString(zenginBankCode) {
					return fmt.Errorf("--zengin-bank-code must be exactly 4 digits")
				}
				if zenginBranchCode == "" {
					return fmt.Errorf("--zengin-branch-code is required when --zengin-bank-code is provided")
				}
			}

			// Validation: Japan Zengin branch code (3 digits)
			if zenginBranchCode != "" {
				zenginBranchRegex := regexp.MustCompile(`^\d{3}$`)
				if !zenginBranchRegex.MatchString(zenginBranchCode) {
					return fmt.Errorf("--zengin-branch-code must be exactly 3 digits")
				}
				if zenginBankCode == "" {
					return fmt.Errorf("--zengin-bank-code is required when --zengin-branch-code is provided")
				}
			}

			// Validation: China CNAPS (12 digits)
			if cnaps != "" {
				cnapsRegex := regexp.MustCompile(`^\d{12}$`)
				if !cnapsRegex.MatchString(cnaps) {
					return fmt.Errorf("--cnaps must be exactly 12 digits")
				}
			}

			// Validation: South Korea bank code (3 digits)
			if koreaBankCode != "" {
				koreaBankRegex := regexp.MustCompile(`^\d{3}$`)
				if !koreaBankRegex.MatchString(koreaBankCode) {
					return fmt.Errorf("--korea-bank-code must be exactly 3 digits")
				}
			}

			// Validation: Brazil CPF (11 digits)
			if cpf != "" {
				cpfRegex := regexp.MustCompile(`^\d{11}$`)
				if !cpfRegex.MatchString(cpf) {
					return fmt.Errorf("--cpf must be exactly 11 digits")
				}
			}

			// Validation: Brazil CNPJ (14 digits)
			if cnpj != "" {
				cnpjRegex := regexp.MustCompile(`^\d{14}$`)
				if !cnpjRegex.MatchString(cnpj) {
					return fmt.Errorf("--cnpj must be exactly 14 digits")
				}
			}

			// Validation: Singapore NRIC (9 chars, format SnnnnnnnA)
			if nric != "" {
				nricRegex := regexp.MustCompile(`^[STFG]\d{7}[A-Z]$`)
				if !nricRegex.MatchString(strings.ToUpper(nric)) {
					return fmt.Errorf("--nric must be 9 characters in format SnnnnnnnA (e.g., S1234567A)")
				}
			}

			// Validation: Singapore UEN (8-13 chars)
			if uen != "" {
				if len(uen) < 8 || len(uen) > 13 {
					return fmt.Errorf("--uen must be 8-13 characters")
				}
			}

			// Validation: Singapore bank code (7 digits)
			if sgBankCode != "" {
				sgBankRegex := regexp.MustCompile(`^\d{7}$`)
				if !sgBankRegex.MatchString(sgBankCode) {
					return fmt.Errorf("--sg-bank-code must be exactly 7 digits")
				}
			}

			// Validation: Singapore PayNow VPA (up to 21 chars)
			if paynowVPA != "" {
				if len(paynowVPA) > 21 {
					return fmt.Errorf("--paynow-vpa must be 21 characters or fewer")
				}
			}

			// Australia PayID validation
			if payidPhone != "" {
				payidPhoneRegex := regexp.MustCompile(`^\+61-\d{9}$`)
				if !payidPhoneRegex.MatchString(payidPhone) {
					return fmt.Errorf("--payid-phone must be in format +61-nnnnnnnnn")
				}
			}
			if payidEmail != "" {
				// Basic email validation
				emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
				if !emailRegex.MatchString(payidEmail) {
					return fmt.Errorf("--payid-email must be a valid email address")
				}
			}
			if payidABN != "" {
				abnRegex := regexp.MustCompile(`^\d{9}$|^\d{11}$`)
				if !abnRegex.MatchString(payidABN) {
					return fmt.Errorf("--payid-abn must be 9 or 11 digits")
				}
			}

			// Validation: Sweden clearing number (4-5 digits)
			if clearingNumber != "" {
				clearingRegex := regexp.MustCompile(`^\d{4,5}$`)
				if !clearingRegex.MatchString(clearingNumber) {
					return fmt.Errorf("--clearing-number must be 4-5 digits")
				}
			}

			// Validation: Hong Kong bank code (3 digits)
			if hkBankCode != "" {
				hkBankRegex := regexp.MustCompile(`^\d{3}$`)
				if !hkBankRegex.MatchString(hkBankCode) {
					return fmt.Errorf("--hk-bank-code must be exactly 3 digits")
				}
			}

			// Validation: Hong Kong FPS ID (7-9 digits)
			if fpsID != "" {
				fpsIDRegex := regexp.MustCompile(`^\d{7,9}$`)
				if !fpsIDRegex.MatchString(fpsID) {
					return fmt.Errorf("--fps-id must be 7-9 digits")
				}
			}

			// Validation: China legal representative ID (15 or 18 chars)
			if legalRepID != "" {
				if len(legalRepID) != 15 && len(legalRepID) != 18 {
					return fmt.Errorf("--legal-rep-id must be 15 or 18 characters")
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
			if cpf != "" {
				beneficiary["personal_id_type"] = "INDIVIDUAL_TAX_ID"
				beneficiary["personal_id_number"] = cpf
			}
			if cnpj != "" {
				beneficiary["business_registration_number"] = cnpj
			}

			// China legal representative additional_info
			if legalRepFirstName != "" || legalRepLastName != "" || legalRepID != "" {
				additionalInfo := map[string]interface{}{}
				if legalRepFirstName != "" {
					additionalInfo["legal_rep_first_name"] = legalRepFirstName
				}
				if legalRepLastName != "" {
					additionalInfo["legal_rep_last_name"] = legalRepLastName
				}
				if legalRepID != "" {
					additionalInfo["legal_rep_id_number"] = legalRepID
				}
				beneficiary["additional_info"] = additionalInfo
			}

			// General ID fields (China and others)
			if personalIDType != "" {
				beneficiary["personal_id_type"] = personalIDType
			}
			if personalIDNumber != "" {
				beneficiary["personal_id_number"] = personalIDNumber
			}
			if businessRegNumber != "" {
				beneficiary["business_registration_number"] = businessRegNumber
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
			if bankAccountCategory != "" {
				bankDetails["bank_account_category"] = bankAccountCategory
			}
			if localClearingSystem != "" {
				bankDetails["local_clearing_system"] = localClearingSystem
			}
			if bankName != "" {
				bankDetails["bank_name"] = bankName
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
			} else if zenginBankCode != "" {
				bankDetails["account_routing_type1"] = "bank_code"
				bankDetails["account_routing_value1"] = zenginBankCode
				if zenginBranchCode != "" {
					bankDetails["account_routing_type2"] = "branch_code"
					bankDetails["account_routing_value2"] = zenginBranchCode
				}
			} else if cnaps != "" {
				bankDetails["account_routing_type1"] = "cnaps"
				bankDetails["account_routing_value1"] = cnaps
			} else if koreaBankCode != "" {
				bankDetails["account_routing_type1"] = "bank_code"
				bankDetails["account_routing_value1"] = koreaBankCode
			} else if nric != "" {
				bankDetails["account_routing_type1"] = "personal_id_number"
				bankDetails["account_routing_value1"] = strings.ToUpper(nric)
			} else if uen != "" {
				bankDetails["account_routing_type1"] = "business_registration_number"
				bankDetails["account_routing_value1"] = uen
			} else if paynowVPA != "" {
				bankDetails["account_routing_type1"] = "virtual_payment_address"
				bankDetails["account_routing_value1"] = paynowVPA
			} else if sgBankCode != "" {
				bankDetails["account_routing_type1"] = "bank_code"
				bankDetails["account_routing_value1"] = sgBankCode
			} else if clearingNumber != "" {
				bankDetails["account_routing_type1"] = "bank_code"
				bankDetails["account_routing_value1"] = clearingNumber
			} else if hkBankCode != "" {
				bankDetails["account_routing_type1"] = "bank_code"
				bankDetails["account_routing_value1"] = hkBankCode
			} else if fpsID != "" {
				bankDetails["account_routing_type1"] = "fps_identifier"
				bankDetails["account_routing_value1"] = fpsID
			} else if hkid != "" {
				bankDetails["account_routing_type1"] = "personal_id_number"
				bankDetails["account_routing_value1"] = hkid
			} else if payidPhone != "" {
				bankDetails["account_routing_type1"] = "phone_number"
				bankDetails["account_routing_value1"] = payidPhone
			} else if payidEmail != "" {
				bankDetails["account_routing_type1"] = "email_address"
				bankDetails["account_routing_value1"] = payidEmail
			} else if payidABN != "" {
				bankDetails["account_routing_type1"] = "australian_business_number"
				bankDetails["account_routing_value1"] = payidABN
			}

			if branchCode != "" {
				bankDetails["branch_code"] = branchCode
			}
			if bankBranch != "" {
				bankDetails["bank_branch"] = bankBranch
			}

			beneficiary["bank_details"] = bankDetails

			// Build request
			req := map[string]interface{}{
				"beneficiary":      beneficiary,
				"transfer_methods": []string{paymentMethod},
				"payment_methods":  []string{paymentMethod},
			}
			if nickname != "" {
				req["nickname"] = nickname
			}

			// Optional: Fetch schema and validate
			if validateOnly {
				schema, err := client.GetBeneficiarySchema(cmd.Context(), bankCountry, entityType, paymentMethod)
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

				// Top-level request fields
				addField("entity-type", entityType)
				addField("bank-country", bankCountry)
				addField("payment-method", paymentMethod)

				// Account details
				addField("account-name", accountName)
				addField("account-number", accountNumber)
				addField("account-currency", accountCurrency)
				addField("account-category", bankAccountCategory)

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

				// Japan Zengin
				addField("zengin-bank-code", zenginBankCode)
				addField("zengin-branch-code", zenginBranchCode)

				// China CNAPS
				addField("cnaps", cnaps)

				// South Korea
				addField("korea-bank-code", koreaBankCode)

				// Brazil
				addField("cpf", cpf)
				addField("cnpj", cnpj)
				addField("bank-branch", bankBranch)

				// Singapore PayNow
				addField("paynow-vpa", paynowVPA)
				addField("uen", uen)
				addField("nric", nric)
				addField("sg-bank-code", sgBankCode)

				// Canada EFT
				addField("institution-number", institutionNumber)
				addField("transit-number", transitNumber)

				// Sweden
				addField("clearing-number", clearingNumber)

				// Hong Kong FPS
				addField("hk-bank-code", hkBankCode)
				addField("fps-id", fpsID)
				addField("hkid", hkid)

				// China legal representative and general ID fields
				addField("legal-rep-first-name", legalRepFirstName)
				addField("legal-rep-last-name", legalRepLastName)
				addField("legal-rep-id", legalRepID)
				addField("bank-name", bankName)
				addField("personal-id-type", personalIDType)
				addField("personal-id-number", personalIDNumber)
				addField("business-registration-number", businessRegNumber)

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
				u.Info(fmt.Sprintf("Would create beneficiary in %s with %s routing", bankCountry, paymentMethod))
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
	cmd.Flags().StringVar(&paymentMethod, "payment-method", "LOCAL", "Payment method: LOCAL or SWIFT")
	// Alias for backwards compatibility
	cmd.Flags().StringVar(&paymentMethod, "transfer-method", "LOCAL", "Alias for --payment-method (deprecated)")
	_ = cmd.Flags().MarkHidden("transfer-method")
	cmd.Flags().StringVar(&accountCurrency, "account-currency", "", "Currency e.g. CAD, USD (required)")
	cmd.Flags().StringVar(&accountName, "account-name", "", "Account holder name (required)")

	// Name flags
	cmd.Flags().StringVar(&companyName, "company-name", "", "Company name (for COMPANY entity)")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name (for PERSONAL entity)")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name (for PERSONAL entity)")
	cmd.Flags().StringVar(&nickname, "nickname", "", "Nickname for the beneficiary")

	// Bank account flags (EFT)
	cmd.Flags().StringVar(&accountNumber, "account-number", "", "Bank account number")
	cmd.Flags().StringVar(&bankAccountCategory, "bank-account-category", "", "Account category: Checking or Savings (required for US)")
	cmd.Flags().StringVar(&bankAccountCategory, "account-category", "", "Alias for --bank-account-category")
	_ = cmd.Flags().MarkHidden("bank-account-category")
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

	// Japan Zengin routing
	cmd.Flags().StringVar(&zenginBankCode, "zengin-bank-code", "", "Japan Zengin bank code (4 digits)")
	cmd.Flags().StringVar(&zenginBranchCode, "zengin-branch-code", "", "Japan Zengin branch code (3 digits)")

	// China CNAPS routing
	cmd.Flags().StringVar(&cnaps, "cnaps", "", "China CNAPS code (12 digits)")

	// South Korea routing
	cmd.Flags().StringVar(&koreaBankCode, "korea-bank-code", "", "South Korea bank code (3 digits)")

	// Brazil routing
	cmd.Flags().StringVar(&cpf, "cpf", "", "Brazil CPF individual tax ID (11 digits)")
	cmd.Flags().StringVar(&cnpj, "cnpj", "", "Brazil CNPJ business tax ID (14 digits)")
	cmd.Flags().StringVar(&bankBranch, "bank-branch", "", "Bank branch code (Brazil: 4-7 chars)")

	// Singapore PayNow routing
	cmd.Flags().StringVar(&paynowVPA, "paynow-vpa", "", "Singapore PayNow VPA (up to 21 chars)")
	cmd.Flags().StringVar(&uen, "uen", "", "Singapore UEN for business PayNow (8-13 chars)")
	cmd.Flags().StringVar(&nric, "nric", "", "Singapore NRIC for personal PayNow (9 chars, format: SnnnnnnnA)")
	cmd.Flags().StringVar(&sgBankCode, "sg-bank-code", "", "Singapore bank code (7 digits)")

	// Sweden routing
	cmd.Flags().StringVar(&clearingNumber, "clearing-number", "", "Sweden clearing number (4-5 digits)")

	// Hong Kong FPS routing
	cmd.Flags().StringVar(&hkBankCode, "hk-bank-code", "", "Hong Kong bank code (3 digits)")
	cmd.Flags().StringVar(&fpsID, "fps-id", "", "Hong Kong FPS identifier (7-9 digits)")
	cmd.Flags().StringVar(&hkid, "hkid", "", "Hong Kong ID for FPS routing")

	// Australia PayID flags
	cmd.Flags().StringVar(&payidPhone, "payid-phone", "", "Australia PayID phone (format: +61-nnnnnnnnn)")
	cmd.Flags().StringVar(&payidEmail, "payid-email", "", "Australia PayID email address")
	cmd.Flags().StringVar(&payidABN, "payid-abn", "", "Australia PayID ABN (9 or 11 digits)")

	// China special fields
	cmd.Flags().StringVar(&legalRepFirstName, "legal-rep-first-name", "", "China legal representative first name (Chinese)")
	cmd.Flags().StringVar(&legalRepLastName, "legal-rep-last-name", "", "China legal representative last name (Chinese)")
	cmd.Flags().StringVar(&legalRepID, "legal-rep-id", "", "China legal representative ID number (15 or 18 chars)")
	cmd.Flags().StringVar(&bankName, "bank-name", "", "Bank name (required for China)")
	cmd.Flags().StringVar(&personalIDType, "personal-id-type", "", "Personal ID type (e.g., INDIVIDUAL_TAX_ID, CHINESE_NATIONAL_ID)")
	cmd.Flags().StringVar(&personalIDNumber, "personal-id-number", "", "Personal ID number")
	cmd.Flags().StringVar(&businessRegNumber, "business-registration-number", "", "Business registration number")

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
				u.Info("Deletion cancelled.")
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
