package dryrun

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestWithDryRun(t *testing.T) {
	ctx := context.Background()

	if IsEnabled(ctx) {
		t.Error("expected dry-run disabled by default")
	}

	ctx = WithDryRun(ctx, true)
	if !IsEnabled(ctx) {
		t.Error("expected dry-run enabled")
	}
}

func TestPreviewWrite(t *testing.T) {
	p := &Preview{
		Operation:   "create",
		Resource:    "transfer",
		Description: "Send money to John Smith",
		Details: map[string]interface{}{
			"Amount": "1000.00 USD",
			"To":     "John Smith",
		},
	}

	var buf bytes.Buffer
	p.Write(&buf)

	output := buf.String()
	if !strings.Contains(output, "DRY-RUN") {
		t.Error("expected DRY-RUN in output")
	}
	if !strings.Contains(output, "create") {
		t.Error("expected operation in output")
	}
}

func TestPreviewWriteSortsDetails(t *testing.T) {
	p := &Preview{
		Operation: "create",
		Resource:  "transfer",
		Details: map[string]interface{}{
			"Zebra":       "z-value",
			"Amount":      "500.00 USD",
			"Currency":    "USD",
			"Beneficiary": "Alice",
		},
	}

	var buf bytes.Buffer
	p.Write(&buf)
	output := buf.String()

	// Verify keys appear in sorted order
	amountIdx := strings.Index(output, "  Amount:")
	beneficiaryIdx := strings.Index(output, "  Beneficiary:")
	currencyIdx := strings.Index(output, "  Currency:")
	zebraIdx := strings.Index(output, "  Zebra:")

	if amountIdx == -1 || beneficiaryIdx == -1 || currencyIdx == -1 || zebraIdx == -1 {
		t.Fatalf("expected all detail keys in output, got:\n%s", output)
	}

	if amountIdx >= beneficiaryIdx || beneficiaryIdx >= currencyIdx || currencyIdx >= zebraIdx {
		t.Errorf("expected details in sorted order (Amount < Beneficiary < Currency < Zebra), got:\n%s", output)
	}
}

func TestFormatAmount(t *testing.T) {
	result := FormatAmount(1000.50, "usd")
	if result != "1000.50 USD" {
		t.Errorf("expected '1000.50 USD', got %q", result)
	}
}
