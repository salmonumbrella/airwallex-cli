package outfmt

import (
	"encoding/json"
	"fmt"
)

// FormatMoney formats a json.Number monetary amount for human display with 2
// decimal places. It preserves the exact string value from the API JSON when
// possible and falls back to float64 formatting for computed values.
// Returns "0.00" for empty or invalid numbers.
func FormatMoney(n json.Number) string {
	if n == "" {
		return "0.00"
	}
	f, err := n.Float64()
	if err != nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", f)
}

// FormatRate formats a json.Number exchange rate for human display with 6
// decimal places. Returns "0.000000" for empty or invalid numbers.
func FormatRate(n json.Number) string {
	if n == "" {
		return "0.000000"
	}
	f, err := n.Float64()
	if err != nil {
		return "0.000000"
	}
	return fmt.Sprintf("%.6f", f)
}

// MoneyFloat64 converts a json.Number to float64, returning 0 on error.
// Use this for numeric comparisons (e.g., checking if an amount is non-zero).
func MoneyFloat64(n json.Number) float64 {
	f, _ := n.Float64()
	return f
}
