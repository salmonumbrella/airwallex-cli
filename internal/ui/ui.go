package ui

import (
	"context"
	"io"
	"os"

	"github.com/muesli/termenv"
)

type contextKey string

const uiKey contextKey = "ui"

type UI struct {
	out   *termenv.Output
	err   *termenv.Output
	color bool
}

func New(colorMode string) *UI {
	out := termenv.NewOutput(os.Stdout)
	errOut := termenv.NewOutput(os.Stderr)

	var color bool
	switch colorMode {
	case "never":
		color = false
	case "always":
		color = true
	default: // auto
		color = out.ColorProfile() != termenv.Ascii
	}

	if os.Getenv("NO_COLOR") != "" {
		color = false
	}

	return &UI{
		out:   out,
		err:   errOut,
		color: color,
	}
}

func WithUI(ctx context.Context, u *UI) context.Context {
	return context.WithValue(ctx, uiKey, u)
}

func FromContext(ctx context.Context) *UI {
	if u, ok := ctx.Value(uiKey).(*UI); ok {
		return u
	}
	return New("auto")
}

func (u *UI) Out() io.Writer {
	return u.out
}

func (u *UI) Err() io.Writer {
	return u.err
}

func (u *UI) Success(msg string) {
	if u.color {
		msg = termenv.String(msg).Foreground(termenv.ANSIGreen).String()
	}
	_, _ = u.err.WriteString(msg + "\n")
}

func (u *UI) Error(msg string) {
	if u.color {
		msg = termenv.String(msg).Foreground(termenv.ANSIRed).String()
	}
	_, _ = u.err.WriteString(msg + "\n")
}

func (u *UI) Info(msg string) {
	_, _ = u.err.WriteString(msg + "\n")
}

// ColorEnabled returns whether color output is enabled.
func (u *UI) ColorEnabled() bool {
	return u.color
}

// FormatStatus colorizes status values based on their meaning.
// Green for success states, yellow for pending, red for failed/cancelled.
func (u *UI) FormatStatus(status string) string {
	if !u.color {
		return status
	}

	switch status {
	// Success states
	case "COMPLETED", "ACTIVE", "SETTLED", "SUCCESS", "VERIFIED", "PAID":
		return termenv.String(status).Foreground(termenv.ANSIGreen).String()
	// Pending/in-progress states
	case "PENDING", "PROCESSING", "IN_PROGRESS", "AWAITING", "INACTIVE", "CREATED":
		return termenv.String(status).Foreground(termenv.ANSIYellow).String()
	// Failure states
	case "FAILED", "CANCELLED", "CANCELED", "REJECTED", "CLOSED", "EXPIRED":
		return termenv.String(status).Foreground(termenv.ANSIRed).String()
	default:
		return status
	}
}

// FormatAmount colorizes currency amounts.
// Positive amounts are shown in green, negative in red.
func (u *UI) FormatAmount(amount string) string {
	if !u.color {
		return amount
	}

	// Check if it's a negative amount (starts with - after optional whitespace)
	trimmed := amount
	for len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t') {
		trimmed = trimmed[1:]
	}
	if len(trimmed) > 0 && trimmed[0] == '-' {
		return termenv.String(amount).Foreground(termenv.ANSIRed).String()
	}

	return termenv.String(amount).Foreground(termenv.ANSIGreen).String()
}

// FormatCurrency colorizes currency codes.
func (u *UI) FormatCurrency(currency string) string {
	if !u.color {
		return currency
	}
	return termenv.String(currency).Foreground(termenv.ANSICyan).String()
}

// FormatHeader colorizes table headers.
func (u *UI) FormatHeader(header string) string {
	if !u.color {
		return header
	}
	return termenv.String(header).Bold().String()
}
