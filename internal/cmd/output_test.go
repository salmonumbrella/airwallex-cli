package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func TestWriteJSONOutput_JSONLArrayOneLinePerItem(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	ctx := outfmt.WithFormat(context.Background(), "jsonl")
	cmd.SetContext(ctx)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	data := []struct {
		ID int `json:"id"`
	}{
		{ID: 1},
		{ID: 2},
	}

	if err := writeJSONOutput(cmd, data); err != nil {
		t.Fatalf("writeJSONOutput error: %v", err)
	}

	want := "{\"id\":1}\n{\"id\":2}\n"
	if got := buf.String(); got != want {
		t.Errorf("writeJSONOutput(jsonl array) = %q, want %q", got, want)
	}
}

func TestWriteJSONOutput_JSONLQueryOneLinePerResult(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	ctx := outfmt.WithFormat(context.Background(), "jsonl")
	ctx = outfmt.WithQuery(ctx, ".items[] | .id")
	cmd.SetContext(ctx)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": "a"},
			{"id": "b"},
		},
	}

	if err := writeJSONOutput(cmd, data); err != nil {
		t.Fatalf("writeJSONOutput error: %v", err)
	}

	want := "\"a\"\n\"b\"\n"
	if got := buf.String(); got != want {
		t.Errorf("writeJSONOutput(jsonl query) = %q, want %q", got, want)
	}
}
