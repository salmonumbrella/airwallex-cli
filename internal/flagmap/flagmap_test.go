// internal/flagmap/flagmap_test.go
package flagmap

import "testing"

func TestFlagToSchemaPath(t *testing.T) {
	tests := []struct {
		flag     string
		wantPath string
		wantType string // routing type if applicable
	}{
		// Top-level request fields
		{"entity-type", "entity_type", ""},
		{"bank-country", "bank_country_code", ""},
		{"payment-method", "transfer_method", ""},
		// Routing fields
		{"swift-code", "beneficiary.bank_details.swift_code", ""},
		{"iban", "beneficiary.bank_details.iban", ""},
		{"routing-number", "beneficiary.bank_details.account_routing_value1", "aba"},
		{"sort-code", "beneficiary.bank_details.account_routing_value1", "sort_code"},
		{"bsb", "beneficiary.bank_details.account_routing_value1", "bsb"},
		{"ifsc", "beneficiary.bank_details.account_routing_value1", "ifsc"},
		{"account-number", "beneficiary.bank_details.account_number", ""},
		{"account-name", "beneficiary.bank_details.account_name", ""},
		{"company-name", "beneficiary.company_name", ""},
		{"first-name", "beneficiary.first_name", ""},
		{"last-name", "beneficiary.last_name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			mapping, ok := GetMapping(tt.flag)
			if !ok {
				t.Fatalf("no mapping for flag %s", tt.flag)
			}
			if mapping.SchemaPath != tt.wantPath {
				t.Errorf("path = %s, want %s", mapping.SchemaPath, tt.wantPath)
			}
			if mapping.RoutingType != tt.wantType {
				t.Errorf("routing type = %s, want %s", mapping.RoutingType, tt.wantType)
			}
		})
	}
}

func TestGetMappingNotFound(t *testing.T) {
	_, ok := GetMapping("nonexistent-flag")
	if ok {
		t.Error("expected ok=false for nonexistent flag")
	}
}

func TestAllMappings(t *testing.T) {
	all := AllMappings()
	if len(all) != 27 {
		t.Errorf("expected 27 mappings, got %d", len(all))
	}
}

func TestRoutingFlags(t *testing.T) {
	flags := RoutingFlags()
	if len(flags) != 11 {
		t.Errorf("expected 11 routing flags, got %d", len(flags))
	}
}

func TestRoutingFlagsConsistency(t *testing.T) {
	// Verify all routing flags exist in mappings
	for _, flag := range RoutingFlags() {
		if _, ok := GetMapping(flag); !ok {
			t.Errorf("routing flag %q not found in mappings", flag)
		}
	}
}

func TestAllMappingsReturnsCopy(t *testing.T) {
	// Get a copy via AllMappings
	all := AllMappings()
	originalLen := len(all)

	// Mutate the returned map
	all["test-key"] = Mapping{Flag: "test-key"}
	delete(all, "swift-code")

	// Get another copy and verify internal state is unchanged
	fresh := AllMappings()
	if len(fresh) != originalLen {
		t.Errorf("internal mappings were mutated: got %d, want %d", len(fresh), originalLen)
	}
	if _, ok := fresh["swift-code"]; !ok {
		t.Error("internal mappings lost 'swift-code' after external deletion")
	}
	if _, ok := fresh["test-key"]; ok {
		t.Error("internal mappings gained 'test-key' after external addition")
	}
}
