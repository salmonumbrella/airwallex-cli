package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source,omitempty"`
}

func (e *APIError) Error() string {
	if e.Source != "" {
		return fmt.Sprintf("%s: %s (source: %s)", e.Code, e.Message, e.Source)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func ParseAPIError(body []byte) *APIError {
	var e APIError
	if err := json.Unmarshal(body, &e); err != nil {
		return &APIError{
			Code:    "unknown_error",
			Message: "An error occurred processing the API response",
		}
	}

	// Sanitize and limit field lengths to prevent information disclosure
	const maxMessageLength = 500
	const maxCodeLength = 100
	const maxSourceLength = 200

	if len(e.Message) > maxMessageLength {
		e.Message = e.Message[:maxMessageLength] + "..."
	}
	if len(e.Code) > maxCodeLength {
		e.Code = e.Code[:maxCodeLength]
	}
	if len(e.Source) > maxSourceLength {
		e.Source = e.Source[:maxSourceLength]
	}

	if e.Code == "" && e.Message == "" {
		return &APIError{
			Code:    "unknown_error",
			Message: "An error occurred but no details were provided",
		}
	}
	return &e
}

// ValidationError represents an input validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// RateLimitError represents a rate limit exceeded error.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %s", e.RetryAfter)
}

// AuthError represents an authentication or authorization error.
type AuthError struct {
	Reason string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication error: %s", e.Reason)
}

// CircuitBreakerError indicates the circuit breaker is open.
type CircuitBreakerError struct{}

func (e *CircuitBreakerError) Error() string {
	return "circuit breaker is open, too many recent failures"
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	var e *RateLimitError
	return errors.As(err, &e)
}

// IsAuthError checks if the error is an authentication error.
func IsAuthError(err error) bool {
	var e *AuthError
	return errors.As(err, &e)
}

// IsCircuitBreakerError checks if the error is a circuit breaker error.
func IsCircuitBreakerError(err error) bool {
	var e *CircuitBreakerError
	return errors.As(err, &e)
}
