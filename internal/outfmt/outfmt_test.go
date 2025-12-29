package outfmt

import (
	"bytes"
	"context"
	"testing"
)

func TestIsJSON(t *testing.T) {
	ctx := context.Background()
	if IsJSON(ctx) {
		t.Error("IsJSON(empty ctx) = true, want false")
	}

	ctx = WithFormat(ctx, "json")
	if !IsJSON(ctx) {
		t.Error("IsJSON(json ctx) = false, want true")
	}

	ctx = WithFormat(ctx, "text")
	if IsJSON(ctx) {
		t.Error("IsJSON(text ctx) = true, want false")
	}
}

func TestGetFormat(t *testing.T) {
	ctx := context.Background()
	if got := GetFormat(ctx); got != "text" {
		t.Errorf("GetFormat(empty ctx) = %q, want 'text'", got)
	}

	ctx = WithFormat(ctx, "json")
	if got := GetFormat(ctx); got != "json" {
		t.Errorf("GetFormat(json ctx) = %q, want 'json'", got)
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	err := WriteJSON(&buf, data)
	if err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}
	want := `{
  "key": "value"
}`
	if got := buf.String(); got != want+"\n" {
		t.Errorf("WriteJSON = %q, want %q", got, want+"\n")
	}
}

func TestWriteJSONFiltered_WithTypedStruct(t *testing.T) {
	// This test verifies that WriteJSONFiltered works with typed Go structs,
	// not just map[string]interface{}. This was the root cause of the
	// --query flag crash with "invalid type: *api.BalancesResponse"
	type Balance struct {
		Currency        string  `json:"currency"`
		AvailableAmount float64 `json:"available_amount"`
	}
	type BalancesResponse struct {
		Balances []Balance `json:"balances"`
	}

	response := &BalancesResponse{
		Balances: []Balance{
			{Currency: "USD", AvailableAmount: 100.50},
			{Currency: "EUR", AvailableAmount: 200.75},
		},
	}

	var buf bytes.Buffer
	err := WriteJSONFiltered(&buf, response, ".balances[0]")
	if err != nil {
		t.Fatalf("WriteJSONFiltered error: %v", err)
	}

	want := `{
  "available_amount": 100.5,
  "currency": "USD"
}
`
	if got := buf.String(); got != want {
		t.Errorf("WriteJSONFiltered = %q, want %q", got, want)
	}
}

func TestWriteJSONFiltered_WithoutQuery(t *testing.T) {
	type Item struct {
		ID string `json:"id"`
	}
	data := &Item{ID: "test-123"}

	var buf bytes.Buffer
	err := WriteJSONFiltered(&buf, data, "")
	if err != nil {
		t.Fatalf("WriteJSONFiltered error: %v", err)
	}

	want := `{
  "id": "test-123"
}
`
	if got := buf.String(); got != want {
		t.Errorf("WriteJSONFiltered = %q, want %q", got, want)
	}
}
