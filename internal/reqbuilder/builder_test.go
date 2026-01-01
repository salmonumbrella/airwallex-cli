// internal/reqbuilder/builder_test.go
package reqbuilder

import (
	"testing"
)

func TestBuildNestedMap(t *testing.T) {
	fields := map[string]string{
		"beneficiary.company_name":                "Acme Corp",
		"beneficiary.bank_details.account_name":   "Acme Corp",
		"beneficiary.bank_details.account_number": "123456789",
		"beneficiary.bank_details.swift_code":     "CHASUS33",
		"beneficiary.address.city":                "New York",
	}

	result := BuildNestedMap(fields)

	// Check nested structure
	beneficiary, ok := result["beneficiary"].(map[string]interface{})
	if !ok {
		t.Fatal("beneficiary should be a map")
	}

	if beneficiary["company_name"] != "Acme Corp" {
		t.Errorf("company_name = %v", beneficiary["company_name"])
	}

	bankDetails, ok := beneficiary["bank_details"].(map[string]interface{})
	if !ok {
		t.Fatal("bank_details should be a map")
	}

	if bankDetails["swift_code"] != "CHASUS33" {
		t.Errorf("swift_code = %v", bankDetails["swift_code"])
	}
}

func TestBuildRoutingFields(t *testing.T) {
	fields := map[string]string{
		"beneficiary.bank_details.account_routing_value1": "021000021",
	}

	result := BuildNestedMap(fields)
	AddRoutingType(result, 1, "aba")

	beneficiary := result["beneficiary"].(map[string]interface{})
	bankDetails := beneficiary["bank_details"].(map[string]interface{})

	if bankDetails["account_routing_type1"] != "aba" {
		t.Errorf("routing type = %v", bankDetails["account_routing_type1"])
	}
}
