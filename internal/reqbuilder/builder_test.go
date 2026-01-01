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

func TestMergeRequest(t *testing.T) {
	// Test simple key merge
	base := map[string]interface{}{"a": "1"}
	additional := map[string]interface{}{"b": "2"}
	result := MergeRequest(base, additional)

	if result["a"] != "1" || result["b"] != "2" {
		t.Errorf("simple merge failed: %v", result)
	}

	// Test nested map merge
	base = map[string]interface{}{
		"beneficiary": map[string]interface{}{"name": "John"},
	}
	additional = map[string]interface{}{
		"beneficiary": map[string]interface{}{"account": "123"},
	}
	result = MergeRequest(base, additional)

	beneficiary := result["beneficiary"].(map[string]interface{})
	if beneficiary["name"] != "John" || beneficiary["account"] != "123" {
		t.Errorf("nested merge failed: %v", beneficiary)
	}

	// Test value overwrite
	base = map[string]interface{}{"key": "old"}
	additional = map[string]interface{}{"key": "new"}
	result = MergeRequest(base, additional)

	if result["key"] != "new" {
		t.Errorf("overwrite failed: got %v", result["key"])
	}
}
