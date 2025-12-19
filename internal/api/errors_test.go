package api

import (
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
