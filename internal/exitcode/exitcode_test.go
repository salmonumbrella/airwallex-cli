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
