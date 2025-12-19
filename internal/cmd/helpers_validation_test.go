package cmd

import (
	"testing"
)

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"valid date", "2024-01-15", false},
		{"valid leap year date", "2024-02-29", false},
		{"empty string", "", false}, // empty is valid (optional)
		{"invalid format - wrong separator", "2024/01/15", true},
		{"invalid format - no separator", "20240115", true},
		{"invalid month", "2024-13-01", true},
		{"invalid day", "2024-01-32", true},
		{"non-leap year Feb 29", "2023-02-29", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDate(%q) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{"positive amount", 100.50, false},
		{"small positive", 0.01, false},
		{"large amount", 999999.99, false},
		{"zero", 0, true},
		{"negative", -100, true},
		{"small negative", -0.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAmount(%v) error = %v, wantErr %v", tt.amount, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		wantErr  bool
	}{
		{"valid USD", "USD", false},
		{"valid EUR", "EUR", false},
		{"valid CAD", "CAD", false},
		{"empty string", "", false}, // empty is valid (optional)
		{"lowercase", "usd", true},
		{"too short", "US", true},
		{"too long", "USDA", true},
		{"with numbers", "US1", true},
		{"with special chars", "US$", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCurrency(tt.currency)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCurrency(%q) error = %v, wantErr %v", tt.currency, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		{"valid range", "2024-01-01", "2024-01-31", false},
		{"same date", "2024-01-15", "2024-01-15", false},
		{"one day apart", "2024-01-15", "2024-01-16", false},
		{"empty from", "", "2024-01-31", false}, // partial ranges are allowed
		{"empty to", "2024-01-01", "", false},
		{"both empty", "", "", false},
		{"reversed range", "2024-01-31", "2024-01-01", true},
		{"invalid from date", "2024-13-01", "2024-01-31", true},
		{"invalid to date", "2024-01-01", "2024-13-31", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDateRange(%q, %q) error = %v, wantErr %v", tt.from, tt.to, err, tt.wantErr)
			}
		})
	}
}
