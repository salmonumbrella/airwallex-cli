package dryrun

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type contextKey string

const dryRunKey contextKey = "dry_run_enabled"

// WithDryRun returns a context with dry-run mode enabled/disabled.
func WithDryRun(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, dryRunKey, enabled)
}

// IsEnabled returns true if dry-run mode is enabled.
func IsEnabled(ctx context.Context) bool {
	if v, ok := ctx.Value(dryRunKey).(bool); ok {
		return v
	}
	return false
}

// Preview represents a dry-run preview of an operation
type Preview struct {
	Operation   string
	Resource    string
	Description string
	Details     map[string]interface{}
	Warnings    []string
}

// Write outputs the preview to the writer
func (p *Preview) Write(w io.Writer) {
	fmt.Fprintf(w, "\n[DRY-RUN] Would %s %s\n", p.Operation, p.Resource)
	fmt.Fprintf(w, "─────────────────────────────────────\n")

	if p.Description != "" {
		fmt.Fprintf(w, "%s\n\n", p.Description)
	}

	if len(p.Details) > 0 {
		for k, v := range p.Details {
			fmt.Fprintf(w, "  %s: %v\n", k, v)
		}
		fmt.Fprintln(w)
	}

	if len(p.Warnings) > 0 {
		fmt.Fprintln(w, "Warnings:")
		for _, warning := range p.Warnings {
			fmt.Fprintf(w, "  ⚠ %s\n", warning)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "─────────────────────────────────────\n")
	fmt.Fprintln(w, "No changes made (dry-run mode)")
}

// FormatAmount formats a currency amount for display
func FormatAmount(amount float64, currency string) string {
	return fmt.Sprintf("%.2f %s", amount, strings.ToUpper(currency))
}
