package schemavalidator

import (
	"strings"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

func TestValidate_AllRequiredPresent(t *testing.T) {
	schema := &api.Schema{
		Fields: []api.SchemaField{
			{Key: "account_name", Path: "beneficiary.bank_details.account_name", Required: true},
			{Key: "account_number", Path: "beneficiary.bank_details.account_number", Required: true},
			{Key: "nickname", Path: "nickname", Required: false},
		},
	}

	provided := map[string]string{
		"beneficiary.bank_details.account_name":   "Test",
		"beneficiary.bank_details.account_number": "123",
	}

	missing, err := Validate(schema, provided)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) > 0 {
		t.Errorf("unexpected missing fields: %v", missing)
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	schema := &api.Schema{
		Fields: []api.SchemaField{
			{Key: "account_name", Path: "beneficiary.bank_details.account_name", Required: true},
			{Key: "swift_code", Path: "beneficiary.bank_details.swift_code", Required: true},
		},
	}

	provided := map[string]string{
		"beneficiary.bank_details.account_name": "Test",
		// swift_code missing
	}

	missing, _ := Validate(schema, provided)
	if len(missing) != 1 {
		t.Fatalf("expected 1 missing, got %d", len(missing))
	}
	if missing[0].Key != "swift_code" {
		t.Errorf("expected swift_code missing, got %s", missing[0].Key)
	}
}

func TestValidate_EmptyValueTreatedAsMissing(t *testing.T) {
	schema := &api.Schema{
		Fields: []api.SchemaField{
			{Key: "account_name", Path: "beneficiary.bank_details.account_name", Required: true},
		},
	}

	provided := map[string]string{
		"beneficiary.bank_details.account_name": "", // empty string
	}

	missing, err := Validate(schema, provided)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) != 1 {
		t.Errorf("expected 1 missing field for empty value, got %d", len(missing))
	}
}

func TestValidate_UsesKeyWhenPathEmpty(t *testing.T) {
	schema := &api.Schema{
		Fields: []api.SchemaField{
			{Key: "nickname", Path: "", Required: true},
		},
	}

	provided := map[string]string{
		"nickname": "MyNickname",
	}

	missing, err := Validate(schema, provided)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) > 0 {
		t.Errorf("unexpected missing fields: %v", missing)
	}
}

func TestValidatePattern_ValidMatch(t *testing.T) {
	err := ValidatePattern("ABC123", "^[A-Z]+[0-9]+$")
	if err != nil {
		t.Errorf("expected no error for valid pattern match, got: %v", err)
	}
}

func TestValidatePattern_InvalidMatch(t *testing.T) {
	err := ValidatePattern("abc123", "^[A-Z]+[0-9]+$")
	if err == nil {
		t.Error("expected error for invalid pattern match")
	}
}

func TestValidatePattern_EmptyPattern(t *testing.T) {
	err := ValidatePattern("anything", "")
	if err != nil {
		t.Errorf("expected no error for empty pattern, got: %v", err)
	}
}

func TestValidatePattern_InvalidRegex(t *testing.T) {
	err := ValidatePattern("test", "[invalid")
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestFormatMissingFields_Empty(t *testing.T) {
	result := FormatMissingFields(nil)
	if result != "" {
		t.Errorf("expected empty string for no missing fields, got: %q", result)
	}
}

func TestFormatMissingFields_WithDescription(t *testing.T) {
	missing := []MissingField{
		{Key: "swift_code", Path: "beneficiary.bank_details.swift_code", Description: "SWIFT/BIC code"},
	}

	result := FormatMissingFields(missing)
	if result == "" {
		t.Error("expected non-empty result")
	}

	// Check it contains key info
	if !strings.Contains(result, "swift_code") {
		t.Errorf("expected result to contain 'swift_code', got: %s", result)
	}
	if !strings.Contains(result, "SWIFT/BIC code") {
		t.Errorf("expected result to contain description, got: %s", result)
	}
}

func TestFormatMissingFields_WithoutDescription(t *testing.T) {
	missing := []MissingField{
		{Key: "account_name", Path: "beneficiary.bank_details.account_name"},
	}

	result := FormatMissingFields(missing)
	if !strings.Contains(result, "account_name") {
		t.Errorf("expected result to contain 'account_name', got: %s", result)
	}
}
