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
