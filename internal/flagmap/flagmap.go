// Package flagmap provides mappings between CLI flags and Airwallex API schema paths.
package flagmap

// Mapping describes how a CLI flag maps to Airwallex schema fields
type Mapping struct {
	Flag        string // CLI flag name (e.g., "routing-number")
	SchemaPath  string // JSON path in schema (e.g., "beneficiary.bank_details.account_routing_value1")
	RoutingType string // If this is a routing field, the type (e.g., "aba", "sort_code")
	Description string // Human-readable description
}

// mappings defines all CLI flag to schema path mappings
var mappings = map[string]Mapping{
	// Top-level request fields
	"entity-type": {
		Flag:        "entity-type",
		SchemaPath:  "entity_type",
		Description: "Entity type: COMPANY or PERSONAL",
	},
	"bank-country": {
		Flag:        "bank-country",
		SchemaPath:  "bank_country_code",
		Description: "Bank country code (e.g., US, GB, AU)",
	},
	"payment-method": {
		Flag:        "payment-method",
		SchemaPath:  "payment_method",
		Description: "Payment method: LOCAL or SWIFT",
	},
	"nickname": {
		Flag:       "nickname",
		SchemaPath: "nickname",
	},

	// SWIFT/International
	"swift-code": {
		Flag:        "swift-code",
		SchemaPath:  "beneficiary.bank_details.swift_code",
		Description: "SWIFT/BIC code for international transfers",
	},
	"iban": {
		Flag:        "iban",
		SchemaPath:  "beneficiary.bank_details.iban",
		Description: "IBAN for European/international transfers",
	},

	// Country-specific routing
	"routing-number": {
		Flag:        "routing-number",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "aba",
		Description: "US ABA routing number (9 digits)",
	},
	"sort-code": {
		Flag:        "sort-code",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "sort_code",
		Description: "UK sort code (6 digits, format: NN-NN-NN)",
	},
	"bsb": {
		Flag:        "bsb",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "bsb",
		Description: "Australian BSB number (6 digits)",
	},
	// Canada Interac / email + phone
	"email": {
		Flag:        "email",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "email_address",
		Description: "Email for Interac e-Transfer",
	},
	"phone": {
		Flag:        "phone",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "phone_number",
		Description: "Phone for Interac e-Transfer",
	},
	// Australia PayID/NPP
	"payid-phone": {
		Flag:        "payid-phone",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "phone_number",
		Description: "Australia PayID phone (format: +61-nnnnnnnnn)",
	},
	"payid-email": {
		Flag:        "payid-email",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "email_address",
		Description: "Australia PayID email address",
	},
	"payid-abn": {
		Flag:        "payid-abn",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "australian_business_number",
		Description: "Australia PayID ABN (9 or 11 digits)",
	},
	"ifsc": {
		Flag:        "ifsc",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "ifsc",
		Description: "Indian IFSC code",
	},
	"clabe": {
		Flag:        "clabe",
		SchemaPath:  "beneficiary.bank_details.clabe",
		Description: "Mexican CLABE (18 digits)",
	},
	"institution-number": {
		Flag:        "institution-number",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "institution_number",
		Description: "Canadian institution number (3 digits)",
	},
	"transit-number": {
		Flag:        "transit-number",
		SchemaPath:  "beneficiary.bank_details.account_routing_value2",
		RoutingType: "transit_number",
		Description: "Canadian transit/branch number (5 digits)",
	},
	// Japan Zengin (two-tier routing)
	"zengin-bank-code": {
		Flag:        "zengin-bank-code",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "bank_code",
		Description: "Japan Zengin bank code (4 digits)",
	},
	"zengin-branch-code": {
		Flag:        "zengin-branch-code",
		SchemaPath:  "beneficiary.bank_details.account_routing_value2",
		RoutingType: "branch_code",
		Description: "Japan Zengin branch code (3 digits)",
	},
	// China CNAPS
	"cnaps": {
		Flag:        "cnaps",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "cnaps",
		Description: "China CNAPS code (12 digits)",
	},
	// South Korea
	"korea-bank-code": {
		Flag:        "korea-bank-code",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "bank_code",
		Description: "South Korea bank code (3 digits)",
	},
	// Singapore PayNow
	"paynow-vpa": {
		Flag:        "paynow-vpa",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "virtual_payment_address",
		Description: "Singapore PayNow VPA (up to 21 chars)",
	},
	"uen": {
		Flag:        "uen",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "business_registration_number",
		Description: "Singapore UEN for business PayNow (8-13 chars)",
	},
	"nric": {
		Flag:        "nric",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "personal_id_number",
		Description: "Singapore NRIC for personal PayNow (9 chars)",
	},
	"sg-bank-code": {
		Flag:        "sg-bank-code",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "bank_code",
		Description: "Singapore bank code (7 digits: 4 bank + 3 branch)",
	},
	// Sweden
	"clearing-number": {
		Flag:        "clearing-number",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "bank_code",
		Description: "Sweden clearing number (4-5 digits)",
	},
	// Hong Kong FPS
	"hk-bank-code": {
		Flag:        "hk-bank-code",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "bank_code",
		Description: "Hong Kong bank code (3 digits)",
	},
	"fps-id": {
		Flag:        "fps-id",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "fps_identifier",
		Description: "Hong Kong FPS identifier (7-9 digits)",
	},
	"hkid": {
		Flag:        "hkid",
		SchemaPath:  "beneficiary.bank_details.account_routing_value1",
		RoutingType: "personal_id_number",
		Description: "Hong Kong ID for FPS routing",
	},
	// Personal/Business ID fields
	"personal-id-type": {
		Flag:        "personal-id-type",
		SchemaPath:  "beneficiary.personal_id_type",
		Description: "Personal ID type (e.g., INDIVIDUAL_TAX_ID, CHINESE_NATIONAL_ID)",
	},
	"personal-id-number": {
		Flag:        "personal-id-number",
		SchemaPath:  "beneficiary.personal_id_number",
		Description: "Personal ID number (format depends on type/country)",
	},
	"business-registration-number": {
		Flag:        "business-registration-number",
		SchemaPath:  "beneficiary.business_registration_number",
		Description: "Business registration number (e.g., CNPJ for Brazil)",
	},
	// Brazil tax IDs (aliases for personal-id-number/business-registration-number)
	"cpf": {
		Flag:        "cpf",
		SchemaPath:  "beneficiary.personal_id_number",
		Description: "Brazil CPF (individual tax ID, 11 digits)",
	},
	"cnpj": {
		Flag:        "cnpj",
		SchemaPath:  "beneficiary.business_registration_number",
		Description: "Brazil CNPJ (business tax ID, 14 digits)",
	},
	"bank-branch": {
		Flag:        "bank-branch",
		SchemaPath:  "beneficiary.bank_details.bank_branch",
		Description: "Bank branch code (Brazil: 4-7 chars)",
	},
	"bank-code": {
		Flag:        "bank-code",
		SchemaPath:  "beneficiary.bank_details.bank_code",
		Description: "Generic bank code",
	},
	"branch-code": {
		Flag:        "branch-code",
		SchemaPath:  "beneficiary.bank_details.branch_code",
		Description: "Generic branch code",
	},

	// Account details
	"account-number": {
		Flag:       "account-number",
		SchemaPath: "beneficiary.bank_details.account_number",
	},
	"account-name": {
		Flag:       "account-name",
		SchemaPath: "beneficiary.bank_details.account_name",
	},
	"account-currency": {
		Flag:       "account-currency",
		SchemaPath: "beneficiary.bank_details.account_currency",
	},
	"clearing-system": {
		Flag:        "clearing-system",
		SchemaPath:  "beneficiary.bank_details.local_clearing_system",
		Description: "Clearing system: EFT, REGULAR_EFT, INTERAC, etc.",
	},
	"account-category": {
		Flag:        "account-category",
		SchemaPath:  "beneficiary.bank_details.bank_account_category",
		Description: "Bank account category: Checking or Savings",
	},
	"bank-account-category": {
		Flag:        "bank-account-category",
		SchemaPath:  "beneficiary.bank_details.bank_account_category",
		Description: "Bank account category: Checking or Savings (alias for account-category)",
	},

	// China legal representative (for business beneficiaries)
	"legal-rep-first-name": {
		Flag:        "legal-rep-first-name",
		SchemaPath:  "beneficiary.additional_info.legal_rep_first_name",
		Description: "China legal representative first name (Chinese, up to 15 chars)",
	},
	"legal-rep-last-name": {
		Flag:        "legal-rep-last-name",
		SchemaPath:  "beneficiary.additional_info.legal_rep_last_name",
		Description: "China legal representative last name (Chinese, up to 15 chars)",
	},
	"legal-rep-id": {
		Flag:        "legal-rep-id",
		SchemaPath:  "beneficiary.additional_info.legal_rep_id_number",
		Description: "China legal representative ID number (15 or 18 chars)",
	},
	"bank-name": {
		Flag:        "bank-name",
		SchemaPath:  "beneficiary.bank_details.bank_name",
		Description: "Bank name (required for China, up to 200 chars)",
	},

	// Entity details
	"company-name": {
		Flag:       "company-name",
		SchemaPath: "beneficiary.company_name",
	},
	"first-name": {
		Flag:       "first-name",
		SchemaPath: "beneficiary.first_name",
	},
	"last-name": {
		Flag:       "last-name",
		SchemaPath: "beneficiary.last_name",
	},

	// Address
	"address-country": {
		Flag:       "address-country",
		SchemaPath: "beneficiary.address.country_code",
	},
	"address-street": {
		Flag:       "address-street",
		SchemaPath: "beneficiary.address.street_address",
	},
	"address-city": {
		Flag:       "address-city",
		SchemaPath: "beneficiary.address.city",
	},
	"address-state": {
		Flag:       "address-state",
		SchemaPath: "beneficiary.address.state",
	},
	"address-postcode": {
		Flag:       "address-postcode",
		SchemaPath: "beneficiary.address.postcode",
	},
}

// GetMapping returns the mapping for a CLI flag
func GetMapping(flag string) (Mapping, bool) {
	m, ok := mappings[flag]
	return m, ok
}

// AllMappings returns all defined mappings
func AllMappings() map[string]Mapping {
	result := make(map[string]Mapping, len(mappings))
	for k, v := range mappings {
		result[k] = v
	}
	return result
}

// RoutingFlags returns all flags that represent routing information
func RoutingFlags() []string {
	return []string{
		"routing-number", "sort-code", "bsb", "ifsc", "clabe",
		"email", "phone",
		"institution-number", "transit-number", "bank-code", "branch-code",
		"swift-code", "iban",
		"zengin-bank-code", "zengin-branch-code",
		"cnaps",
		"korea-bank-code",
		"paynow-vpa", "uen", "nric", "sg-bank-code",
		"clearing-number",
		"hk-bank-code", "fps-id", "hkid",
		"payid-phone", "payid-email", "payid-abn",
	}
}
