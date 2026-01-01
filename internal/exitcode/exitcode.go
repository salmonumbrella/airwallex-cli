// Package exitcode provides structured exit codes for CLI operations.
// These codes enable agents and scripts to programmatically determine
// the nature of failures and respond appropriately.
package exitcode

import (
	"errors"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

// Exit codes for structured error handling.
// These align with common CLI conventions and enable agent automation.
const (
	Success      = 0 // Command completed successfully
	Error        = 1 // Generic error
	AuthRequired = 4 // Authentication required or expired
	NotFound     = 5 // Resource not found
	Validation   = 6 // Validation error (bad input)
	RateLimited  = 7 // Rate limit exceeded
	Conflict     = 8 // Resource conflict (already exists, etc.)
	ServerErr    = 9 // Server-side error (5xx)
)

// NotFoundError indicates a resource was not found.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return e.Resource + " not found: " + e.ID
	}
	return e.Resource + " not found"
}

// ConflictError indicates a resource conflict.
type ConflictError struct {
	Resource string
	Msg      string
}

func (e *ConflictError) Error() string {
	return e.Resource + " conflict: " + e.Msg
}

// ServerError indicates a server-side error.
type ServerError struct {
	StatusCode int
	Msg        string
}

func (e *ServerError) Error() string {
	return e.Msg
}

// FromError determines the exit code based on error type.
// It checks both exitcode-specific types and api package error types.
func FromError(err error) int {
	if err == nil {
		return Success
	}

	// Check api package error types
	var authErr *api.AuthError
	if errors.As(err, &authErr) {
		return AuthRequired
	}

	var validationErr *api.ValidationError
	if errors.As(err, &validationErr) {
		return Validation
	}

	var rateLimitErr *api.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return RateLimited
	}

	var circuitBreakerErr *api.CircuitBreakerError
	if errors.As(err, &circuitBreakerErr) {
		return ServerErr
	}

	// Check exitcode-specific types
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		return NotFound
	}

	var conflictErr *ConflictError
	if errors.As(err, &conflictErr) {
		return Conflict
	}

	var serverErr *ServerError
	if errors.As(err, &serverErr) {
		return ServerErr
	}

	return Error
}
