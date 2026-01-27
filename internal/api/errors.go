package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FieldError represents a specific field validation error from the API.
type FieldError struct {
	Source  string                 `json:"source"`
	Code    string                 `json:"code"`
	Message string                 `json:"message,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// APIErrorDetails can be either a string or an object containing field errors.
// Airwallex returns details as a string for some errors (e.g., access_denied)
// and as an object with nested errors for validation errors.
type APIErrorDetails struct {
	String string       // When details is a plain string
	Errors []FieldError // When details is {errors: [...]}
}

func (d *APIErrorDetails) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.String = s
		return nil
	}
	// Try object with errors array
	var obj struct {
		Errors []FieldError `json:"errors,omitempty"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		d.Errors = obj.Errors
		return nil
	}
	// Ignore unparseable details
	return nil
}

type APIError struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Source  string           `json:"source,omitempty"`
	Errors  []FieldError     `json:"errors,omitempty"`
	Details *APIErrorDetails `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Code, e.Message)
	if e.Source != "" {
		msg += fmt.Sprintf(" (source: %s)", e.Source)
	}

	// Include details string if present and different from message
	if e.Details != nil && e.Details.String != "" && e.Details.String != e.Message {
		msg += fmt.Sprintf(" (details: %s)", e.Details.String)
	}

	// Get errors from either top-level or nested in details
	fieldErrors := e.Errors
	if len(fieldErrors) == 0 && e.Details != nil {
		fieldErrors = e.Details.Errors
	}

	if len(fieldErrors) > 0 {
		msg += "\nField errors:"
		for _, fe := range fieldErrors {
			errMsg := fe.Message
			if errMsg == "" && fe.Params != nil {
				if opts, ok := fe.Params["value_options"]; ok {
					errMsg = fmt.Sprintf("must be one of: %v", opts)
				}
			}
			if errMsg == "" {
				errMsg = fmt.Sprintf("error code %s", fe.Code)
			}
			msg += fmt.Sprintf("\n  - %s: %s", fe.Source, errMsg)
		}
	}
	return msg
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
	const maxFieldErrors = 20

	if len(e.Message) > maxMessageLength {
		e.Message = e.Message[:maxMessageLength] + "..."
	}
	if len(e.Code) > maxCodeLength {
		e.Code = e.Code[:maxCodeLength]
	}
	if len(e.Source) > maxSourceLength {
		e.Source = e.Source[:maxSourceLength]
	}

	// Sanitize field errors
	if len(e.Errors) > maxFieldErrors {
		e.Errors = e.Errors[:maxFieldErrors]
	}
	for i := range e.Errors {
		if len(e.Errors[i].Source) > maxSourceLength {
			e.Errors[i].Source = e.Errors[i].Source[:maxSourceLength]
		}
		if len(e.Errors[i].Code) > maxCodeLength {
			e.Errors[i].Code = e.Errors[i].Code[:maxCodeLength]
		}
		if len(e.Errors[i].Message) > maxMessageLength {
			e.Errors[i].Message = e.Errors[i].Message[:maxMessageLength] + "..."
		}
	}

	// Sanitize nested details (string or errors array)
	if e.Details != nil {
		if len(e.Details.String) > maxMessageLength {
			e.Details.String = e.Details.String[:maxMessageLength] + "..."
		}
		if len(e.Details.Errors) > maxFieldErrors {
			e.Details.Errors = e.Details.Errors[:maxFieldErrors]
		}
		for i := range e.Details.Errors {
			if len(e.Details.Errors[i].Source) > maxSourceLength {
				e.Details.Errors[i].Source = e.Details.Errors[i].Source[:maxSourceLength]
			}
			if len(e.Details.Errors[i].Code) > maxCodeLength {
				e.Details.Errors[i].Code = e.Details.Errors[i].Code[:maxCodeLength]
			}
			if len(e.Details.Errors[i].Message) > maxMessageLength {
				e.Details.Errors[i].Message = e.Details.Errors[i].Message[:maxMessageLength] + "..."
			}
		}
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

// IsNotFoundError checks if the error indicates a resource was not found.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == "not_found" ||
			apiErr.Code == "resource_not_found" ||
			strings.Contains(strings.ToLower(apiErr.Message), "not found")
	}
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

// ContextualError wraps an API error with request context
type ContextualError struct {
	Method     string
	URL        string
	StatusCode int
	Err        error
}

func (e *ContextualError) Error() string {
	return fmt.Sprintf("%s %s failed (status %d): %v", e.Method, e.URL, e.StatusCode, e.Err)
}

func (e *ContextualError) Unwrap() error {
	return e.Err
}

// WrapError adds request context to an API error
func WrapError(method, url string, statusCode int, err error) error {
	return &ContextualError{
		Method:     method,
		URL:        url,
		StatusCode: statusCode,
		Err:        err,
	}
}

// NormalizeAPIError maps API errors to typed errors when possible.
func NormalizeAPIError(statusCode int, apiErr *APIError) error {
	if apiErr == nil {
		return apiErr
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return &AuthError{Reason: apiErr.Message}
	}
	return apiErr
}
