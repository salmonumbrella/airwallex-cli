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
		SchemaPath:  "transfer_method",
		Description: "Payment method: LOCAL or SWIFT",
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
	// Brazil tax IDs
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
		"institution-number", "transit-number", "bank-code", "branch-code",
		"swift-code", "iban",
		"zengin-bank-code", "zengin-branch-code",
		"cnaps",
	}
}
