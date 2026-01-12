package cmd

import (
	"strings"
	"testing"
)

func TestParseDateRangeRFC3339(t *testing.T) {
	from, to := "2024-01-01", "2024-01-02"
	gotFrom, gotTo, err := parseDateRangeRFC3339(from, to, "--from", "--to", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotFrom != "2024-01-01T00:00:00Z" {
		t.Fatalf("expected from RFC3339, got %q", gotFrom)
	}
	if gotTo != "2024-01-02T23:59:59Z" {
		t.Fatalf("expected to RFC3339, got %q", gotTo)
	}
}

func TestValidateDateRangeFlags_LabelError(t *testing.T) {
	err := validateDateRangeFlags("nope", "", "--from", "--to", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid --from date") {
		t.Fatalf("expected label in error, got %q", err.Error())
	}
}
