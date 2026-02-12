package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/flagmap"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/reqbuilder"
	"github.com/salmonumbrella/airwallex-cli/internal/schemavalidator"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

// Pre-compiled regexes for beneficiary field validation.
// Grouped by pattern: "digits of length N" share reDigitsN vars.
var (
	reDigits3     = regexp.MustCompile(`^\d{3}$`)          // institution-number, zengin-branch-code, korea-bank-code, hk-bank-code
	reDigits4     = regexp.MustCompile(`^\d{4}$`)          // zengin-bank-code
	reDigits5     = regexp.MustCompile(`^\d{5}$`)          // transit-number
	reDigits6     = regexp.MustCompile(`^\d{6}$`)          // sort-code, bsb
	reDigits7     = regexp.MustCompile(`^\d{7}$`)          // sg-bank-code
	reDigits9     = regexp.MustCompile(`^\d{9}$`)          // routing-number
	reDigits11    = regexp.MustCompile(`^\d{11}$`)         // cpf
	reDigits12    = regexp.MustCompile(`^\d{12}$`)         // cnaps
	reDigits14    = regexp.MustCompile(`^\d{14}$`)         // cnpj
	reDigits18    = regexp.MustCompile(`^\d{18}$`)         // clabe
	reDigits4or5  = regexp.MustCompile(`^\d{4,5}$`)        // clearing-number
	reDigits7to9  = regexp.MustCompile(`^\d{7,9}$`)        // fps-id
	reDigits9or11 = regexp.MustCompile(`^\d{9}$|^\d{11}$`) // payid-abn
	rePhoneCA     = regexp.MustCompile(`^\+1-\d{10}$`)     // Canada phone
	rePhoneAU     = regexp.MustCompile(`^\+61-\d{9}$`)     // Australia PayID phone
	reIFSC        = regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
	reNRIC        = regexp.MustCompile(`^[STFG]\d{7}[A-Z]$`)
	reEmail       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func newBeneficiariesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "beneficiaries",
		Aliases: []string{"benef", "beneficiary", "ben"},
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
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List beneficiaries",
		Long: `List beneficiaries for payouts.

Use --output json with --query for advanced filtering using jq syntax.
Tip: add --items-only to output just the array for jq piping.

Examples:
  # List recent beneficiaries
  airwallex beneficiaries list --page-size 20

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
			result, err := client.ListBeneficiaries(ctx, opts.Page, opts.Limit)
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
	return NewGetCommand(GetConfig[*api.Beneficiary]{
		Use:     "get <beneficiaryId>",
		Aliases: []string{"g"},
		Short:   "Get beneficiary details",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.Beneficiary, error) {
			return client.GetBeneficiary(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, b *api.Beneficiary) error {
			rows := []outfmt.KV{
				{Key: "beneficiary_id", Value: b.BeneficiaryID},
				{Key: "nickname", Value: b.Nickname},
				{Key: "entity_type", Value: b.Beneficiary.EntityType},
			}
			if b.Beneficiary.CompanyName != "" {
				rows = append(rows, outfmt.KV{Key: "company_name", Value: b.Beneficiary.CompanyName})
			}
			if b.Beneficiary.FirstName != "" {
				rows = append(rows,
					outfmt.KV{Key: "first_name", Value: b.Beneficiary.FirstName},
					outfmt.KV{Key: "last_name", Value: b.Beneficiary.LastName},
				)
			}
			rows = append(rows,
				outfmt.KV{Key: "bank_country", Value: b.Beneficiary.BankDetails.BankCountryCode},
				outfmt.KV{Key: "bank_name", Value: b.Beneficiary.BankDetails.BankName},
				outfmt.KV{Key: "account_name", Value: b.Beneficiary.BankDetails.AccountName},
			)
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBeneficiariesCreateCmd() *cobra.Command {
	// Validation mode
	var validateOnly bool
	// Raw field overrides
	var fieldOverrides []string

	mappings := flagmap.AllMappings()
	mappingKeys := sortedMappingKeys(mappings)

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a new beneficiary",
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
    --account-currency CAD --email john@example.com --clearing-system INTERAC \
    --address-country CA --address-street "123 Main St" --address-city Toronto

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

			overrideFields, err := parseFieldOverrides(fieldOverrides)
			if err != nil {
				return err
			}

			flagValues, err := collectFlagValues(cmd, mappingKeys)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("transfer-method") {
				if transferMethod, err := cmd.Flags().GetString("transfer-method"); err == nil && transferMethod != "" {
					flagValues["payment-method"] = transferMethod
				}
			}

			entityType := flagValues["entity-type"]
			entityType = normalizeEnumValue(entityType, []string{"COMPANY", "PERSONAL"})
			bankCountry := flagValues["bank-country"]
			companyName := flagValues["company-name"]
			firstName := flagValues["first-name"]
			lastName := flagValues["last-name"]
			nickname := flagValues["nickname"]
			paymentMethod := flagValues["payment-method"]
			accountCurrency := flagValues["account-currency"]
			accountName := flagValues["account-name"]
			accountNumber := flagValues["account-number"]
			institutionNumber := flagValues["institution-number"]
			transitNumber := flagValues["transit-number"]
			email := flagValues["email"]
			phone := flagValues["phone"]
			localClearingSystem := flagValues["clearing-system"]
			// SWIFT/international routing
			swiftCode := flagValues["swift-code"]
			routingNumber := flagValues["routing-number"]
			iban := flagValues["iban"]
			// Additional international routing flags
			sortCode := flagValues["sort-code"]
			bsb := flagValues["bsb"]
			ifsc := flagValues["ifsc"]
			clabe := flagValues["clabe"]
			bankCode := flagValues["bank-code"]
			branchCode := flagValues["branch-code"]
			// Japan Zengin
			zenginBankCode := flagValues["zengin-bank-code"]
			zenginBranchCode := flagValues["zengin-branch-code"]
			bankAccountCategory := flagValues["bank-account-category"]
			if val := flagValues["account-category"]; val != "" {
				bankAccountCategory = val
			}
			// China
			cnaps := flagValues["cnaps"]
			// South Korea
			koreaBankCode := flagValues["korea-bank-code"]
			// Brazil
			cpf := flagValues["cpf"]
			cnpj := flagValues["cnpj"]
			bankBranch := flagValues["bank-branch"]
			// Singapore PayNow
			paynowVPA := flagValues["paynow-vpa"]
			uen := flagValues["uen"]
			nric := flagValues["nric"]
			sgBankCode := flagValues["sg-bank-code"]
			// Sweden
			clearingNumber := flagValues["clearing-number"]
			// Hong Kong FPS
			hkBankCode := flagValues["hk-bank-code"]
			fpsID := flagValues["fps-id"]
			hkid := flagValues["hkid"]
			// Australia PayID
			payidPhone := flagValues["payid-phone"]
			payidEmail := flagValues["payid-email"]
			payidABN := flagValues["payid-abn"]
			// China legal representative
			legalRepFirstName := flagValues["legal-rep-first-name"]
			legalRepLastName := flagValues["legal-rep-last-name"]
			legalRepID := flagValues["legal-rep-id"]
			bankName := flagValues["bank-name"]
			personalIDType := flagValues["personal-id-type"]
			personalIDNumber := flagValues["personal-id-number"]
			businessRegNumber := flagValues["business-registration-number"]
			// Address fields (required for Interac)
			addressCountry := flagValues["address-country"]
			addressStreet := flagValues["address-street"]
			addressCity := flagValues["address-city"]
			addressState := flagValues["address-state"]
			addressPostcode := flagValues["address-postcode"]

			// Validation: Required fields based on entity type
			accountNameValue := valueOrOverride(overrideFields, "beneficiary.bank_details.account_name", accountName)
			accountCurrencyValue := valueOrOverride(overrideFields, "beneficiary.bank_details.account_currency", accountCurrency)
			firstNameValue := valueOrOverride(overrideFields, "beneficiary.first_name", firstName)
			lastNameValue := valueOrOverride(overrideFields, "beneficiary.last_name", lastName)
			companyNameValue := valueOrOverride(overrideFields, "beneficiary.company_name", companyName)

			if accountNameValue == "" {
				return fmt.Errorf("--account-name is required")
			}
			if accountCurrencyValue == "" {
				return fmt.Errorf("--account-currency is required")
			}

			switch entityType {
			case "COMPANY":
				if companyNameValue == "" {
					return fmt.Errorf("--company-name is required when entity-type is COMPANY")
				}
			case "PERSONAL":
				if firstNameValue == "" {
					return fmt.Errorf("--first-name is required when entity-type is PERSONAL")
				}
				if lastNameValue == "" {
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

			hasRoutingOverride := hasRoutingOverrideField(overrideFields)
			hasAnyRouting := hasEmail || hasPhone || hasEFT || hasSWIFT || hasRouting ||
				hasIBAN || hasSortCode || hasBSB || hasIFSC || hasCLABE || hasBankCode || hasZengin || hasCNAPS || hasKorea || hasPayNow || hasClearing || hasFPS || hasPayID || hasRoutingOverride

			if !hasAnyRouting {
				return fmt.Errorf("must provide at least one routing method (e.g., --swift-code, --iban, --routing-number, --sort-code, --bsb)")
			}

			// Validation: Canada EFT requires both institution and transit numbers
			if institutionNumber != "" && transitNumber == "" {
				return fmt.Errorf("--transit-number is required when --institution-number is provided")
			}

			// Validation: Phone number format
			if phone != "" {
				if !rePhoneCA.MatchString(phone) {
					return fmt.Errorf("--phone must match format +1-nnnnnnnnnn (e.g., +1-4165551234)")
				}
			}

			// Validation: Institution number format
			if institutionNumber != "" {
				if !reDigits3.MatchString(institutionNumber) {
					return fmt.Errorf("--institution-number must be exactly 3 digits")
				}
			}

			// Validation: Transit number format
			if transitNumber != "" {
				if !reDigits5.MatchString(transitNumber) {
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

			// Validation: Interac e-Transfer (Canada)
			localClearingValue := valueOrOverride(overrideFields, "beneficiary.bank_details.local_clearing_system", localClearingSystem)
			isInterac := strings.EqualFold(localClearingValue, "INTERAC")
			if (hasEmail || hasPhone) && strings.EqualFold(bankCountry, "CA") {
				if localClearingValue == "" {
					localClearingValue = "INTERAC"
					isInterac = true
				} else if !isInterac {
					return fmt.Errorf("--clearing-system must be INTERAC when using --email or --phone for CA")
				}
			}
			if isInterac {
				localClearingSystem = localClearingValue
				if !strings.EqualFold(bankCountry, "CA") {
					return fmt.Errorf("--clearing-system INTERAC is only valid with --bank-country CA")
				}
				if !hasEmail && !hasPhone {
					return fmt.Errorf("--email or --phone is required for Interac e-Transfer")
				}
				addressCountryValue := valueOrOverride(overrideFields, "beneficiary.address.country_code", addressCountry)
				addressStreetValue := valueOrOverride(overrideFields, "beneficiary.address.street_address", addressStreet)
				addressCityValue := valueOrOverride(overrideFields, "beneficiary.address.city", addressCity)
				if addressCountryValue == "" || addressStreetValue == "" || addressCityValue == "" {
					return fmt.Errorf("--address-country, --address-street, and --address-city are required for Interac e-Transfer")
				}
			}

			// Validation: Routing number format (US ABA - 9 digits)
			if routingNumber != "" {
				if !reDigits9.MatchString(routingNumber) {
					return fmt.Errorf("--routing-number must be exactly 9 digits")
				}
			}

			// Validation: Sort code format (UK - 6 digits)
			if sortCode != "" {
				if !reDigits6.MatchString(sortCode) {
					return fmt.Errorf("--sort-code must be exactly 6 digits")
				}
			}

			// Validation: BSB format (Australia - 6 digits)
			if bsb != "" {
				if !reDigits6.MatchString(bsb) {
					return fmt.Errorf("--bsb must be exactly 6 digits")
				}
			}

			// Validation: CLABE format (Mexico - 18 digits)
			if clabe != "" {
				if !reDigits18.MatchString(clabe) {
					return fmt.Errorf("--clabe must be exactly 18 digits")
				}
			}

			// Validation: IFSC format (India - 11 chars: 4 letters, 0, 6 alphanumeric)
			if ifsc != "" {
				if !reIFSC.MatchString(strings.ToUpper(ifsc)) {
					return fmt.Errorf("--ifsc must be 11 characters: 4 letters, 0, then 6 alphanumeric (e.g., SBIN0001234)")
				}
			}

			// Validation: Japan Zengin bank code (4 digits)
			if zenginBankCode != "" {
				if !reDigits4.MatchString(zenginBankCode) {
					return fmt.Errorf("--zengin-bank-code must be exactly 4 digits")
				}
				if zenginBranchCode == "" {
					return fmt.Errorf("--zengin-branch-code is required when --zengin-bank-code is provided")
				}
			}

			// Validation: Japan Zengin branch code (3 digits)
			if zenginBranchCode != "" {
				if !reDigits3.MatchString(zenginBranchCode) {
					return fmt.Errorf("--zengin-branch-code must be exactly 3 digits")
				}
				if zenginBankCode == "" {
					return fmt.Errorf("--zengin-bank-code is required when --zengin-branch-code is provided")
				}
			}

			// Validation: China CNAPS (12 digits)
			if cnaps != "" {
				if !reDigits12.MatchString(cnaps) {
					return fmt.Errorf("--cnaps must be exactly 12 digits")
				}
			}

			// Validation: South Korea bank code (3 digits)
			if koreaBankCode != "" {
				if !reDigits3.MatchString(koreaBankCode) {
					return fmt.Errorf("--korea-bank-code must be exactly 3 digits")
				}
			}

			// Validation: Brazil CPF (11 digits)
			if cpf != "" {
				if !reDigits11.MatchString(cpf) {
					return fmt.Errorf("--cpf must be exactly 11 digits")
				}
			}

			// Validation: Brazil CNPJ (14 digits)
			if cnpj != "" {
				if !reDigits14.MatchString(cnpj) {
					return fmt.Errorf("--cnpj must be exactly 14 digits")
				}
			}

			// Validation: Singapore NRIC (9 chars, format SnnnnnnnA)
			if nric != "" {
				if !reNRIC.MatchString(strings.ToUpper(nric)) {
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
				if !reDigits7.MatchString(sgBankCode) {
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
				if !rePhoneAU.MatchString(payidPhone) {
					return fmt.Errorf("--payid-phone must be in format +61-nnnnnnnnn")
				}
			}
			if payidEmail != "" {
				if !reEmail.MatchString(payidEmail) {
					return fmt.Errorf("--payid-email must be a valid email address")
				}
			}
			if payidABN != "" {
				if !reDigits9or11.MatchString(payidABN) {
					return fmt.Errorf("--payid-abn must be 9 or 11 digits")
				}
			}

			// Validation: Sweden clearing number (4-5 digits)
			if clearingNumber != "" {
				if !reDigits4or5.MatchString(clearingNumber) {
					return fmt.Errorf("--clearing-number must be 4-5 digits")
				}
			}

			// Validation: Hong Kong bank code (3 digits)
			if hkBankCode != "" {
				if !reDigits3.MatchString(hkBankCode) {
					return fmt.Errorf("--hk-bank-code must be exactly 3 digits")
				}
			}

			// Validation: Hong Kong FPS ID (7-9 digits)
			if fpsID != "" {
				if !reDigits7to9.MatchString(fpsID) {
					return fmt.Errorf("--fps-id must be 7-9 digits")
				}
			}

			// Validation: China legal representative ID (15 or 18 chars)
			if legalRepID != "" {
				if len(legalRepID) != 15 && len(legalRepID) != 18 {
					return fmt.Errorf("--legal-rep-id must be 15 or 18 characters")
				}
			}

			// Resolve routing unless overridden via --field.
			routingType := ""
			routingValue1 := ""
			routingType2 := ""
			routingValue2 := ""
			hasRoutingOverride1 := overrideFields["beneficiary.bank_details.account_routing_value1"] != "" ||
				overrideFields["beneficiary.bank_details.account_routing_type1"] != ""
			hasRoutingOverride2 := overrideFields["beneficiary.bank_details.account_routing_value2"] != "" ||
				overrideFields["beneficiary.bank_details.account_routing_type2"] != ""

			routingTypeFor := func(flagName string) string {
				if mapping, ok := mappings[flagName]; ok && mapping.RoutingType != "" {
					return mapping.RoutingType
				}
				switch flagName {
				case "bank-code":
					return "bank_code"
				default:
					return ""
				}
			}

			if !hasRoutingOverride1 {
				switch {
				case routingNumber != "":
					routingType = routingTypeFor("routing-number")
					routingValue1 = routingNumber
				case sortCode != "":
					routingType = routingTypeFor("sort-code")
					routingValue1 = sortCode
				case bsb != "":
					routingType = routingTypeFor("bsb")
					routingValue1 = bsb
				case ifsc != "":
					routingType = routingTypeFor("ifsc")
					routingValue1 = ifsc
				case bankCode != "":
					routingType = routingTypeFor("bank-code")
					routingValue1 = bankCode
				case email != "":
					routingType = routingTypeFor("email")
					routingValue1 = email
				case phone != "":
					routingType = routingTypeFor("phone")
					routingValue1 = phone
				case institutionNumber != "":
					routingType = routingTypeFor("institution-number")
					routingValue1 = institutionNumber
					if transitNumber != "" {
						routingType2 = routingTypeFor("transit-number")
						routingValue2 = transitNumber
					}
				case zenginBankCode != "":
					routingType = routingTypeFor("zengin-bank-code")
					routingValue1 = zenginBankCode
					if zenginBranchCode != "" {
						routingType2 = routingTypeFor("zengin-branch-code")
						routingValue2 = zenginBranchCode
					}
				case cnaps != "":
					routingType = routingTypeFor("cnaps")
					routingValue1 = cnaps
				case koreaBankCode != "":
					routingType = routingTypeFor("korea-bank-code")
					routingValue1 = koreaBankCode
				case nric != "":
					routingType = routingTypeFor("nric")
					routingValue1 = strings.ToUpper(nric)
				case uen != "":
					routingType = routingTypeFor("uen")
					routingValue1 = uen
				case paynowVPA != "":
					routingType = routingTypeFor("paynow-vpa")
					routingValue1 = paynowVPA
				case sgBankCode != "":
					routingType = routingTypeFor("sg-bank-code")
					routingValue1 = sgBankCode
				case clearingNumber != "":
					routingType = routingTypeFor("clearing-number")
					routingValue1 = clearingNumber
				case hkBankCode != "":
					routingType = routingTypeFor("hk-bank-code")
					routingValue1 = hkBankCode
				case fpsID != "":
					routingType = routingTypeFor("fps-id")
					routingValue1 = fpsID
				case hkid != "":
					routingType = routingTypeFor("hkid")
					routingValue1 = hkid
				case payidPhone != "":
					routingType = routingTypeFor("payid-phone")
					routingValue1 = payidPhone
				case payidEmail != "":
					routingType = routingTypeFor("payid-email")
					routingValue1 = payidEmail
				case payidABN != "":
					routingType = routingTypeFor("payid-abn")
					routingValue1 = payidABN
				}
			}

			fields := map[string]string{
				"beneficiary.entity_type":                    entityType,
				"beneficiary.bank_details.bank_country_code": bankCountry,
			}
			addMapped := func(flagName, value string) {
				if value == "" {
					return
				}
				if mapping, ok := flagmap.GetMapping(flagName); ok {
					fields[mapping.SchemaPath] = value
				}
			}

			// Basic details
			addMapped("nickname", nickname)
			addMapped("company-name", companyName)
			addMapped("first-name", firstName)
			addMapped("last-name", lastName)

			// Brazil convenience fields
			if cpf != "" {
				fields["beneficiary.personal_id_number"] = cpf
				if personalIDType == "" {
					fields["beneficiary.personal_id_type"] = "INDIVIDUAL_TAX_ID"
				}
			}
			if cnpj != "" {
				fields["beneficiary.business_registration_number"] = cnpj
			}

			// General ID fields (override convenience fields if provided)
			addMapped("personal-id-type", personalIDType)
			addMapped("personal-id-number", personalIDNumber)
			addMapped("business-registration-number", businessRegNumber)

			// China legal representative
			addMapped("legal-rep-first-name", legalRepFirstName)
			addMapped("legal-rep-last-name", legalRepLastName)
			addMapped("legal-rep-id", legalRepID)

			// Account/bank details
			addMapped("account-name", accountName)
			addMapped("account-number", accountNumber)
			addMapped("account-currency", accountCurrency)
			addMapped("account-category", bankAccountCategory)
			addMapped("bank-name", bankName)
			addMapped("bank-branch", bankBranch)
			addMapped("bank-code", bankCode)
			addMapped("branch-code", branchCode)
			addMapped("swift-code", swiftCode)
			addMapped("iban", iban)
			addMapped("clabe", clabe)
			addMapped("clearing-system", localClearingSystem)

			// Address
			addMapped("address-country", addressCountry)
			addMapped("address-street", addressStreet)
			addMapped("address-city", addressCity)
			addMapped("address-state", addressState)
			addMapped("address-postcode", addressPostcode)

			// Routing values
			if routingValue1 != "" && !hasRoutingOverride1 {
				fields["beneficiary.bank_details.account_routing_value1"] = routingValue1
				if routingType != "" {
					fields["beneficiary.bank_details.account_routing_type1"] = routingType
				}
			}
			if routingValue2 != "" && !hasRoutingOverride2 {
				fields["beneficiary.bank_details.account_routing_value2"] = routingValue2
				if routingType2 != "" {
					fields["beneficiary.bank_details.account_routing_type2"] = routingType2
				}
			}

			req := reqbuilder.BuildNestedMap(fields)
			req = reqbuilder.MergeRequest(req, map[string]interface{}{
				"transfer_methods": []string{paymentMethod},
				"payment_methods":  []string{paymentMethod},
			})
			if len(overrideFields) > 0 {
				req = reqbuilder.MergeRequest(req, reqbuilder.BuildNestedMap(overrideFields))
			}

			// Optional: Fetch schema and validate
			if validateOnly {
				schema, err := client.GetBeneficiarySchema(cmd.Context(), bankCountry, entityType, paymentMethod)
				if err != nil {
					return fmt.Errorf("failed to fetch schema: %w", err)
				}

				// Build provided fields map for validation using flagmap + overrides
				provided := make(map[string]string)
				addProvided := func(flagName, value string) {
					if value == "" {
						return
					}
					if m, ok := flagmap.GetMapping(flagName); ok {
						provided[m.SchemaPath] = value
					}
				}

				addProvided("entity-type", entityType)
				addProvided("bank-country", bankCountry)
				addProvided("payment-method", paymentMethod)

				for path, value := range fields {
					if value == "" {
						continue
					}
					provided[path] = value
				}

				for path, value := range overrideFields {
					if value == "" {
						continue
					}
					provided[path] = value
				}

				// Validate using schemavalidator package
				missing, err := schemavalidator.Validate(schema, provided)
				if err != nil {
					return fmt.Errorf("validation error: %w", err)
				}

				if len(missing) > 0 {
					return fmt.Errorf("%s", schemavalidator.FormatMissingFields(missing))
				}

				for _, field := range schema.Fields {
					if field.Rule.Pattern == "" {
						continue
					}
					path := field.Path
					if path == "" {
						path = field.Key
					}
					if value, ok := provided[path]; ok && value != "" {
						if err := schemavalidator.ValidatePattern(value, field.Rule.Pattern); err != nil {
							return fmt.Errorf("field %s: %w", field.Key, err)
						}
					}
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

	registerMappedFlags(cmd, mappingKeys, map[string]string{
		"payment-method": "LOCAL",
	}, map[string]string{
		"entity-type":  "COMPANY or PERSONAL (required)",
		"bank-country": "Bank country code e.g. CA, US (required)",
	})
	cmd.Flags().String("transfer-method", "LOCAL", "Alias for --payment-method (deprecated)")
	_ = cmd.Flags().MarkHidden("transfer-method")
	_ = cmd.Flags().MarkHidden("bank-account-category")

	// Validation mode flag
	cmd.Flags().BoolVar(&validateOnly, "validate", false, "Validate against schema without creating")
	cmd.Flags().StringArrayVar(&fieldOverrides, "field", nil, "Set raw field (path=value)")

	mustMarkRequired(cmd, "entity-type")
	mustMarkRequired(cmd, "bank-country")
	flagAlias(cmd.Flags(), "entity-type", "et")
	flagAlias(cmd.Flags(), "bank-country", "bk")
	flagAlias(cmd.Flags(), "account-name", "an")
	flagAlias(cmd.Flags(), "account-number", "acn")
	flagAlias(cmd.Flags(), "account-currency", "ac")
	flagAlias(cmd.Flags(), "company-name", "cn")
	flagAlias(cmd.Flags(), "swift-code", "sw")
	flagAlias(cmd.Flags(), "routing-number", "rn")
	flagAlias(cmd.Flags(), "payment-method", "pm")
	flagAlias(cmd.Flags(), "clearing-system", "cs")
	flagAlias(cmd.Flags(), "institution-number", "inst")
	flagAlias(cmd.Flags(), "transit-number", "tn")
	return cmd
}

func newBeneficiariesUpdateCmd() *cobra.Command {
	var fieldOverrides []string
	updateFlagKeys := []string{
		"nickname",
		"company-name",
		"first-name",
		"last-name",
		"address-country",
		"address-street",
		"address-city",
		"address-state",
		"address-postcode",
	}

	cmd := &cobra.Command{
		Use:     "update <beneficiaryId>",
		Aliases: []string{"up", "u"},
		Short:   "Update beneficiary (nickname, names, address)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			// Check if any updates were specified
			flagValues, err := collectFlagValues(cmd, updateFlagKeys)
			if err != nil {
				return err
			}

			hasUpdates := len(fieldOverrides) > 0
			for _, flagName := range updateFlagKeys {
				if cmd.Flags().Changed(flagName) {
					hasUpdates = true
					break
				}
			}

			if !hasUpdates {
				return fmt.Errorf("no updates specified")
			}

			overrideFields, err := parseFieldOverrides(fieldOverrides)
			if err != nil {
				return err
			}

			beneficiaryID := NormalizeIDArg(args[0])

			// Fetch existing beneficiary data
			existing, err := client.GetBeneficiaryRaw(cmd.Context(), beneficiaryID)
			if err != nil {
				return fmt.Errorf("failed to fetch existing beneficiary: %w", err)
			}

			// Remove id field - API doesn't want it in update request
			delete(existing, "id")

			updateFields := make(map[string]string)
			for _, flagName := range updateFlagKeys {
				if !cmd.Flags().Changed(flagName) {
					continue
				}
				value := flagValues[flagName]
				if value == "" {
					continue
				}
				if mapping, ok := flagmap.GetMapping(flagName); ok {
					updateFields[mapping.SchemaPath] = value
				}
			}

			updateReq := reqbuilder.BuildNestedMap(updateFields)
			if len(overrideFields) > 0 {
				updateReq = reqbuilder.MergeRequest(updateReq, reqbuilder.BuildNestedMap(overrideFields))
			}
			existing = reqbuilder.MergeRequest(existing, updateReq)

			b, err := client.UpdateBeneficiary(cmd.Context(), beneficiaryID, existing)
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

	registerMappedFlags(cmd, updateFlagKeys, nil, nil)
	cmd.Flags().StringArrayVar(&fieldOverrides, "field", nil, "Set raw field (path=value)")
	return cmd
}

func newBeneficiariesDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <beneficiaryId>",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a beneficiary",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			beneficiaryID := NormalizeIDArg(args[0])

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
		Use:     "validate",
		Aliases: []string{"val", "v"},
		Short:   "Validate beneficiary details",
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

func parseFieldOverrides(entries []string) (map[string]string, error) {
	overrides := make(map[string]string)
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("--field must be in path=value format: %q", entry)
		}
		overrides[parts[0]] = parts[1]
	}
	return overrides, nil
}

func sortedMappingKeys(m map[string]flagmap.Mapping) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func registerMappedFlags(cmd *cobra.Command, keys []string, defaults, overrides map[string]string) {
	mappings := flagmap.AllMappings()
	for _, key := range keys {
		mapping, ok := mappings[key]
		if !ok {
			panic(fmt.Sprintf("unknown flag mapping for %q", key))
		}
		desc := mapping.Description
		if overrides != nil {
			if override, ok := overrides[key]; ok {
				desc = override
			}
		}
		if desc == "" {
			desc = key
		}
		def := ""
		if defaults != nil {
			if d, ok := defaults[key]; ok {
				def = d
			}
		}
		cmd.Flags().String(key, def, desc)
	}
}

func collectFlagValues(cmd *cobra.Command, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		value, err := cmd.Flags().GetString(key)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, nil
}

func valueOrOverride(overrides map[string]string, path, fallback string) string {
	if value, ok := overrides[path]; ok && value != "" {
		return value
	}
	return fallback
}

func hasRoutingOverrideField(overrides map[string]string) bool {
	for path, value := range overrides {
		if value == "" {
			continue
		}
		if strings.HasPrefix(path, "beneficiary.bank_details.account_routing_") {
			return true
		}
		switch path {
		case "beneficiary.bank_details.swift_code",
			"beneficiary.bank_details.iban",
			"beneficiary.bank_details.clabe",
			"beneficiary.bank_details.bank_code",
			"beneficiary.bank_details.branch_code",
			"beneficiary.bank_details.bank_branch":
			return true
		}
	}
	return false
}
