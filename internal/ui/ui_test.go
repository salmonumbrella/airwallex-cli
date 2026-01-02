package ui

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		colorMode  string
		noColorEnv string
		wantColor  bool
	}{
		{
			name:      "never mode disables color",
			colorMode: "never",
			wantColor: false,
		},
		{
			name:      "always mode enables color",
			colorMode: "always",
			wantColor: true,
		},
		{
			name:       "NO_COLOR env disables color even in always mode",
			colorMode:  "always",
			noColorEnv: "1",
			wantColor:  false,
		},
		{
			name:       "NO_COLOR env disables color in auto mode",
			colorMode:  "auto",
			noColorEnv: "1",
			wantColor:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set/unset NO_COLOR
			if tt.noColorEnv != "" {
				os.Setenv("NO_COLOR", tt.noColorEnv)
				defer os.Unsetenv("NO_COLOR")
			} else {
				os.Unsetenv("NO_COLOR")
			}

			u := New(tt.colorMode)
			if u.ColorEnabled() != tt.wantColor {
				t.Errorf("ColorEnabled() = %v, want %v", u.ColorEnabled(), tt.wantColor)
			}
		})
	}
}

func TestNew_OutWriters(t *testing.T) {
	u := New("never")

	if u.Out() == nil {
		t.Error("Out() should not be nil")
	}
	if u.Err() == nil {
		t.Error("Err() should not be nil")
	}
}

func TestWithUI_FromContext(t *testing.T) {
	ctx := context.Background()
	u := New("never")

	ctx = WithUI(ctx, u)
	got := FromContext(ctx)

	if got != u {
		t.Error("FromContext should return the UI set with WithUI")
	}
}

func TestFromContext_Default(t *testing.T) {
	ctx := context.Background()

	got := FromContext(ctx)

	if got == nil {
		t.Error("FromContext should return a default UI when none is set")
	}
}

func TestFormatHeader(t *testing.T) {
	tests := []struct {
		name    string
		color   bool
		header  string
		wantRaw string
	}{
		{
			name:    "no color returns raw header",
			color:   false,
			header:  "ID",
			wantRaw: "ID",
		},
		{
			name:    "with color applies bold",
			color:   true,
			header:  "ID",
			wantRaw: "", // will check it's not just plain text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorMode := "never"
			if tt.color {
				colorMode = "always"
			}
			u := New(colorMode)
			got := u.FormatHeader(tt.header)

			if !tt.color {
				if got != tt.wantRaw {
					t.Errorf("FormatHeader() = %q, want %q", got, tt.wantRaw)
				}
			} else {
				// With color, the result should contain ANSI escape sequences
				if !strings.Contains(got, "\x1b[") {
					t.Errorf("FormatHeader() should contain ANSI escape sequences when color is enabled")
				}
				// Should still contain the original text
				if !strings.Contains(got, tt.header) {
					t.Errorf("FormatHeader() should contain the original header text")
				}
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantColor bool
	}{
		// Success states (green)
		{name: "COMPLETED is success", status: "COMPLETED", wantColor: true},
		{name: "ACTIVE is success", status: "ACTIVE", wantColor: true},
		{name: "SETTLED is success", status: "SETTLED", wantColor: true},
		{name: "SUCCESS is success", status: "SUCCESS", wantColor: true},
		{name: "VERIFIED is success", status: "VERIFIED", wantColor: true},
		{name: "PAID is success", status: "PAID", wantColor: true},

		// Pending states (yellow)
		{name: "PENDING is pending", status: "PENDING", wantColor: true},
		{name: "PROCESSING is pending", status: "PROCESSING", wantColor: true},
		{name: "IN_PROGRESS is pending", status: "IN_PROGRESS", wantColor: true},
		{name: "AWAITING is pending", status: "AWAITING", wantColor: true},
		{name: "INACTIVE is pending", status: "INACTIVE", wantColor: true},
		{name: "CREATED is pending", status: "CREATED", wantColor: true},

		// Failure states (red)
		{name: "FAILED is failure", status: "FAILED", wantColor: true},
		{name: "CANCELLED is failure", status: "CANCELLED", wantColor: true},
		{name: "CANCELED is failure", status: "CANCELED", wantColor: true},
		{name: "REJECTED is failure", status: "REJECTED", wantColor: true},
		{name: "CLOSED is failure", status: "CLOSED", wantColor: true},
		{name: "EXPIRED is failure", status: "EXPIRED", wantColor: true},

		// Unknown status (no color)
		{name: "UNKNOWN has no color", status: "UNKNOWN", wantColor: false},
		{name: "OTHER has no color", status: "OTHER", wantColor: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := New("always")
			got := u.FormatStatus(tt.status)

			hasANSI := strings.Contains(got, "\x1b[")
			if tt.wantColor && !hasANSI {
				t.Errorf("FormatStatus(%q) should have ANSI color codes", tt.status)
			}
			if !tt.wantColor && hasANSI {
				t.Errorf("FormatStatus(%q) should not have ANSI color codes", tt.status)
			}

			// Should always contain the status text
			if !strings.Contains(got, tt.status) {
				t.Errorf("FormatStatus(%q) should contain the status text", tt.status)
			}
		})
	}
}

func TestFormatStatus_NoColor(t *testing.T) {
	u := New("never")

	statuses := []string{"COMPLETED", "PENDING", "FAILED", "UNKNOWN"}
	for _, status := range statuses {
		got := u.FormatStatus(status)
		if got != status {
			t.Errorf("FormatStatus(%q) with color disabled = %q, want %q", status, got, status)
		}
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		wantRed bool // negative amounts are red
		color   bool
	}{
		{
			name:    "positive amount is green",
			amount:  "100.00",
			wantRed: false,
			color:   true,
		},
		{
			name:    "negative amount is red",
			amount:  "-50.00",
			wantRed: true,
			color:   true,
		},
		{
			name:    "negative with leading space",
			amount:  "  -25.00",
			wantRed: true,
			color:   true,
		},
		{
			name:    "no color mode returns raw",
			amount:  "-100.00",
			wantRed: false,
			color:   false,
		},
		{
			name:    "zero is green",
			amount:  "0.00",
			wantRed: false,
			color:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorMode := "never"
			if tt.color {
				colorMode = "always"
			}
			u := New(colorMode)
			got := u.FormatAmount(tt.amount)

			if !tt.color {
				if got != tt.amount {
					t.Errorf("FormatAmount(%q) = %q, want %q", tt.amount, got, tt.amount)
				}
				return
			}

			// With color enabled, check for ANSI codes
			hasANSI := strings.Contains(got, "\x1b[")
			if !hasANSI {
				t.Errorf("FormatAmount(%q) should have ANSI color codes when color is enabled", tt.amount)
			}

			// Red color code is 31
			hasRed := strings.Contains(got, "31")
			if tt.wantRed && !hasRed {
				t.Errorf("FormatAmount(%q) should be red (negative amount)", tt.amount)
			}
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		color    bool
	}{
		{
			name:     "USD with color",
			currency: "USD",
			color:    true,
		},
		{
			name:     "EUR with color",
			currency: "EUR",
			color:    true,
		},
		{
			name:     "USD no color",
			currency: "USD",
			color:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colorMode := "never"
			if tt.color {
				colorMode = "always"
			}
			u := New(colorMode)
			got := u.FormatCurrency(tt.currency)

			if !tt.color {
				if got != tt.currency {
					t.Errorf("FormatCurrency(%q) = %q, want %q", tt.currency, got, tt.currency)
				}
				return
			}

			// With color enabled, should have ANSI codes
			if !strings.Contains(got, "\x1b[") {
				t.Errorf("FormatCurrency(%q) should have ANSI color codes when color is enabled", tt.currency)
			}

			// Should contain the currency code
			if !strings.Contains(got, tt.currency) {
				t.Errorf("FormatCurrency(%q) should contain the currency code", tt.currency)
			}
		})
	}
}
