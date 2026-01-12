package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func TestNewPayloadCommand_JSONUsesContextIO(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	customIO := &iocontext.IO{
		Out:    &outBuf,
		ErrOut: &errBuf,
		In:     strings.NewReader(""),
	}

	ctx := iocontext.WithIO(outfmt.WithFormat(context.Background(), "json"), customIO)

	cmd := NewPayloadCommand(PayloadCommandConfig[map[string]any]{
		Use:   "test",
		Short: "Test",
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (map[string]any, error) {
			return map[string]any{"ok": true}, nil
		},
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--data", "{}"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "\"ok\"") {
		t.Fatalf("expected JSON output, got %q", output)
	}
}
