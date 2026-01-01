package api

import (
	"errors"
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:    "not_found",
		Message: "Resource not found",
	}
	got := err.Error()
	if !strings.Contains(got, "not_found") {
		t.Errorf("Error() = %q, want to contain 'not_found'", got)
	}
	if !strings.Contains(got, "Resource not found") {
		t.Errorf("Error() = %q, want to contain message", got)
	}
}

func TestAPIError_ErrorWithSource(t *testing.T) {
	err := &APIError{
		Code:    "validation_error",
		Message: "Invalid field",
		Source:  "email",
	}
	got := err.Error()
	if !strings.Contains(got, "source: email") {
		t.Errorf("Error() = %q, want to contain source", got)
	}
}

func TestParseAPIError(t *testing.T) {
	body := []byte(`{"code": "invalid_argument", "message": "Bad request"}`)
	err := ParseAPIError(body)
	if err.Code != "invalid_argument" {
		t.Errorf("Code = %q, want 'invalid_argument'", err.Code)
	}
	if err.Message != "Bad request" {
		t.Errorf("Message = %q, want 'Bad request'", err.Message)
	}
}

func TestParseAPIError_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	err := ParseAPIError(body)
	if err.Code != "unknown_error" {
		t.Errorf("Code = %q, want 'unknown_error'", err.Code)
	}
	if err.Message != "An error occurred processing the API response" {
		t.Errorf("Message = %q, want generic error message", err.Message)
	}
}

func TestParseAPIError_EmptyFields(t *testing.T) {
	body := []byte(`{"code": "", "message": ""}`)
	err := ParseAPIError(body)
	if err.Code != "unknown_error" {
		t.Errorf("Code = %q, want 'unknown_error'", err.Code)
	}
	if err.Message != "An error occurred but no details were provided" {
		t.Errorf("Message = %q, want generic error message", err.Message)
	}
}

func TestAPIError_ErrorWithFieldErrors(t *testing.T) {
	err := &APIError{
		Code:    "validation_failed",
		Message: "The request failed our schema validation.",
		Errors: []FieldError{
			{
				Source:  "beneficiary.bank_details.swift_code",
				Code:    "field_required",
				Message: "This field is required",
			},
			{
				Source:  "beneficiary.bank_details.account_number",
				Code:    "field_invalid",
				Message: "Invalid account number format",
			},
		},
	}
	got := err.Error()

	// Check that the main message is present
	if !strings.Contains(got, "validation_failed") {
		t.Errorf("Error() = %q, want to contain 'validation_failed'", got)
	}

	// Check that field errors header is present
	if !strings.Contains(got, "Field errors:") {
		t.Errorf("Error() = %q, want to contain 'Field errors:'", got)
	}

	// Check that individual field errors are included
	if !strings.Contains(got, "beneficiary.bank_details.swift_code: This field is required") {
		t.Errorf("Error() = %q, want to contain first field error", got)
	}
	if !strings.Contains(got, "beneficiary.bank_details.account_number: Invalid account number format") {
		t.Errorf("Error() = %q, want to contain second field error", got)
	}
}

func TestParseAPIError_WithFieldErrors(t *testing.T) {
	body := []byte(`{
		"code": "validation_failed",
		"message": "The request failed our schema validation.",
		"errors": [
			{
				"source": "beneficiary.bank_details.swift_code",
				"code": "field_required",
				"message": "This field is required"
			}
		]
	}`)
	err := ParseAPIError(body)

	if err.Code != "validation_failed" {
		t.Errorf("Code = %q, want 'validation_failed'", err.Code)
	}
	if len(err.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(err.Errors))
	}
	if err.Errors[0].Source != "beneficiary.bank_details.swift_code" {
		t.Errorf("Errors[0].Source = %q, want 'beneficiary.bank_details.swift_code'", err.Errors[0].Source)
	}
	if err.Errors[0].Code != "field_required" {
		t.Errorf("Errors[0].Code = %q, want 'field_required'", err.Errors[0].Code)
	}
	if err.Errors[0].Message != "This field is required" {
		t.Errorf("Errors[0].Message = %q, want 'This field is required'", err.Errors[0].Message)
	}
}

func TestParseAPIError_WithNestedDetailsErrors(t *testing.T) {
	body := []byte(`{
		"code": "validation_failed",
		"message": "Validation error",
		"details": {
			"errors": [
				{
					"source": "beneficiary.bank_details.account_name",
					"code": "field_required",
					"message": "Account name is required"
				}
			]
		}
	}`)
	err := ParseAPIError(body)

	if err.Code != "validation_failed" {
		t.Errorf("Code = %q, want 'validation_failed'", err.Code)
	}
	if err.Details == nil {
		t.Fatal("Details should not be nil")
	}
	if len(err.Details.Errors) != 1 {
		t.Fatalf("len(Details.Errors) = %d, want 1", len(err.Details.Errors))
	}
	if err.Details.Errors[0].Source != "beneficiary.bank_details.account_name" {
		t.Errorf("Details.Errors[0].Source = %q, want 'beneficiary.bank_details.account_name'", err.Details.Errors[0].Source)
	}

	// Verify Error() output includes nested details.errors
	got := err.Error()
	if !strings.Contains(got, "beneficiary.bank_details.account_name") {
		t.Errorf("Error() = %q, want to contain nested details.errors field", got)
	}
	if !strings.Contains(got, "Account name is required") {
		t.Errorf("Error() = %q, want to contain nested error message", got)
	}
}

func TestContextualError(t *testing.T) {
	inner := &APIError{Code: "not_found", Message: "Transfer not found"}
	err := WrapError("GET", "/api/v1/transfers/123", 404, inner)

	// Check error message format
	expected := "GET /api/v1/transfers/123 failed (status 404): not_found: Transfer not found"
	if err.Error() != expected {
		t.Errorf("error = %q, want %q", err.Error(), expected)
	}

	// Check unwrap
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Error("expected to unwrap to APIError")
	}
}
