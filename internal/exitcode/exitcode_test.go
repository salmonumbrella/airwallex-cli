package exitcode

import (
	"errors"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

func TestFromError_NilReturnsSuccess(t *testing.T) {
	if got := FromError(nil); got != Success {
		t.Errorf("FromError(nil) = %d, want %d", got, Success)
	}
}

func TestFromError_GenericErrorReturnsError(t *testing.T) {
	err := errors.New("something went wrong")
	if got := FromError(err); got != Error {
		t.Errorf("FromError(generic) = %d, want %d", got, Error)
	}
}

func TestFromError_AuthErrorReturnsAuthRequired(t *testing.T) {
	err := &api.AuthError{Reason: "token expired"}
	if got := FromError(err); got != AuthRequired {
		t.Errorf("FromError(AuthError) = %d, want %d", got, AuthRequired)
	}
}

func TestFromError_NotFoundErrorReturnsNotFound(t *testing.T) {
	err := &NotFoundError{Resource: "transfer", ID: "tfr_123"}
	if got := FromError(err); got != NotFound {
		t.Errorf("FromError(NotFoundError) = %d, want %d", got, NotFound)
	}
}

func TestFromError_ValidationErrorReturnsValidation(t *testing.T) {
	err := &api.ValidationError{Field: "amount", Message: "must be positive"}
	if got := FromError(err); got != Validation {
		t.Errorf("FromError(ValidationError) = %d, want %d", got, Validation)
	}
}

func TestFromError_RateLimitErrorReturnsRateLimited(t *testing.T) {
	err := &api.RateLimitError{RetryAfter: 60}
	if got := FromError(err); got != RateLimited {
		t.Errorf("FromError(RateLimitError) = %d, want %d", got, RateLimited)
	}
}

func TestFromError_ConflictErrorReturnsConflict(t *testing.T) {
	err := &ConflictError{Resource: "transfer", Msg: "already exists"}
	if got := FromError(err); got != Conflict {
		t.Errorf("FromError(ConflictError) = %d, want %d", got, Conflict)
	}
}

func TestFromError_ServerErrorReturnsServerError(t *testing.T) {
	err := &ServerError{StatusCode: 500, Msg: "internal server error"}
	if got := FromError(err); got != ServerErr {
		t.Errorf("FromError(ServerError) = %d, want %d", got, ServerErr)
	}
}

func TestFromError_CircuitBreakerErrorReturnsServerError(t *testing.T) {
	err := &api.CircuitBreakerError{}
	if got := FromError(err); got != ServerErr {
		t.Errorf("FromError(CircuitBreakerError) = %d, want %d", got, ServerErr)
	}
}

func TestFromError_WrappedError(t *testing.T) {
	// Test that wrapped errors are properly unwrapped
	innerErr := &api.AuthError{Reason: "expired"}
	wrappedErr := errors.Join(errors.New("context"), innerErr)
	if got := FromError(wrappedErr); got != AuthRequired {
		t.Errorf("FromError(wrapped AuthError) = %d, want %d", got, AuthRequired)
	}
}

func TestNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      NotFoundError
		expected string
	}{
		{
			name:     "with ID",
			err:      NotFoundError{Resource: "transfer", ID: "tfr_123"},
			expected: "transfer not found: tfr_123",
		},
		{
			name:     "without ID",
			err:      NotFoundError{Resource: "transfer"},
			expected: "transfer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConflictError_Error(t *testing.T) {
	err := ConflictError{Resource: "beneficiary", Msg: "already exists"}
	expected := "beneficiary conflict: already exists"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

func TestServerError_Error(t *testing.T) {
	err := ServerError{StatusCode: 503, Msg: "service unavailable"}
	expected := "service unavailable"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

func TestFromError_ContextualError_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   int
	}{
		{"401 unauthorized", 401, AuthRequired},
		{"403 forbidden", 403, AuthRequired},
		{"404 not found", 404, NotFound},
		{"400 bad request", 400, Validation},
		{"422 unprocessable", 422, Validation},
		{"429 rate limited", 429, RateLimited},
		{"409 conflict", 409, Conflict},
		{"500 server error", 500, ServerErr},
		{"502 bad gateway", 502, ServerErr},
		{"503 unavailable", 503, ServerErr},
		{"418 teapot", 418, Error}, // Unknown 4xx
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &api.ContextualError{
				Method:     "GET",
				URL:        "/test",
				StatusCode: tt.statusCode,
				Err:        errors.New("test error"),
			}
			if got := FromError(err); got != tt.expected {
				t.Errorf("FromError(status %d) = %d, want %d", tt.statusCode, got, tt.expected)
			}
		})
	}
}

// Integration tests for WrapError → FromError flow

func TestFromError_WrapError_Integration(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		url        string
		statusCode int
		innerErr   error
		wantCode   int
	}{
		{
			name:       "WrapError with 401 maps to AuthRequired",
			method:     "GET",
			url:        "/api/v1/accounts",
			statusCode: 401,
			innerErr:   errors.New("unauthorized"),
			wantCode:   AuthRequired,
		},
		{
			name:       "WrapError with 404 maps to NotFound",
			method:     "GET",
			url:        "/api/v1/transfers/tfr_123",
			statusCode: 404,
			innerErr:   errors.New("not found"),
			wantCode:   NotFound,
		},
		{
			name:       "WrapError with 500 maps to ServerErr",
			method:     "POST",
			url:        "/api/v1/payments",
			statusCode: 500,
			innerErr:   errors.New("internal server error"),
			wantCode:   ServerErr,
		},
		{
			name:       "WrapError with APIError inner error",
			method:     "POST",
			url:        "/api/v1/beneficiaries",
			statusCode: 400,
			innerErr:   &api.APIError{Code: "invalid_request", Message: "bad input"},
			wantCode:   Validation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.WrapError(tt.method, tt.url, tt.statusCode, tt.innerErr)
			if got := FromError(err); got != tt.wantCode {
				t.Errorf("FromError(WrapError(%s, %s, %d)) = %d, want %d",
					tt.method, tt.url, tt.statusCode, got, tt.wantCode)
			}
		})
	}
}

func TestFromError_WrapError_DoubleWrapped(t *testing.T) {
	// Test that double-wrapped errors still map correctly
	innerErr := errors.New("original error")
	firstWrap := api.WrapError("GET", "/api/v1/inner", 404, innerErr)
	// Wrap again with a different status - outer status should win
	doubleWrap := api.WrapError("POST", "/api/v1/outer", 500, firstWrap)

	got := FromError(doubleWrap)
	if got != ServerErr {
		t.Errorf("FromError(double-wrapped) = %d, want %d (ServerErr from outer 500)", got, ServerErr)
	}
}

func TestFromError_WrapError_WithJoin(t *testing.T) {
	// Test WrapError combined with errors.Join
	wrapped := api.WrapError("DELETE", "/api/v1/resource/123", 403, errors.New("forbidden"))
	joined := errors.Join(errors.New("operation failed"), wrapped)

	got := FromError(joined)
	if got != AuthRequired {
		t.Errorf("FromError(joined with WrapError) = %d, want %d (AuthRequired from 403)", got, AuthRequired)
	}
}

func TestWrapError_ErrorMessage_Formatting(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		url      string
		status   int
		innerErr error
		wantMsg  string
	}{
		{
			name:     "basic error message",
			method:   "GET",
			url:      "/api/v1/accounts",
			status:   401,
			innerErr: errors.New("unauthorized"),
			wantMsg:  "GET /api/v1/accounts failed (status 401): unauthorized",
		},
		{
			name:     "POST with APIError",
			method:   "POST",
			url:      "/api/v1/payments",
			status:   400,
			innerErr: &api.APIError{Code: "invalid_amount", Message: "amount must be positive"},
			wantMsg:  "POST /api/v1/payments failed (status 400): invalid_amount: amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.WrapError(tt.method, tt.url, tt.status, tt.innerErr)
			if got := err.Error(); got != tt.wantMsg {
				t.Errorf("WrapError().Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestContextualError_Unwrap(t *testing.T) {
	// Verify that the inner error can be unwrapped
	innerErr := &api.APIError{Code: "test_code", Message: "test message"}
	wrapped := api.WrapError("GET", "/test", 400, innerErr)

	var apiErr *api.APIError
	if !errors.As(wrapped, &apiErr) {
		t.Error("errors.As failed to extract inner APIError from wrapped error")
	}
	if apiErr.Code != "test_code" {
		t.Errorf("unwrapped APIError.Code = %q, want %q", apiErr.Code, "test_code")
	}
}
