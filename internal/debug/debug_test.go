package debug

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestWithDebug(t *testing.T) {
	ctx := context.Background()

	// Default should be false
	if IsEnabled(ctx) {
		t.Error("expected debug disabled by default")
	}

	// Enable debug
	ctx = WithDebug(ctx, true)
	if !IsEnabled(ctx) {
		t.Error("expected debug enabled after WithDebug(true)")
	}

	// Disable debug
	ctx = WithDebug(ctx, false)
	if IsEnabled(ctx) {
		t.Error("expected debug disabled after WithDebug(false)")
	}
}

func TestSetupLogger_DebugEnabled(t *testing.T) {
	SetupLogger(true)

	handler := slog.Default().Handler()

	// When debug is enabled, Debug level should be enabled
	if !handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected Debug level to be enabled when debugEnabled=true")
	}
	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected Info level to be enabled when debugEnabled=true")
	}
	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected Warn level to be enabled when debugEnabled=true")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Error("expected Error level to be enabled when debugEnabled=true")
	}
}

func TestSetupLogger_DebugDisabled(t *testing.T) {
	SetupLogger(false)

	handler := slog.Default().Handler()

	// When debug is disabled, level should be Warn (Debug and Info disabled)
	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected Debug level to be disabled when debugEnabled=false")
	}
	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected Info level to be disabled when debugEnabled=false")
	}
	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected Warn level to be enabled when debugEnabled=false")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Error("expected Error level to be enabled when debugEnabled=false")
	}
}

func TestSetupLogger_WritesToStderr(t *testing.T) {
	// Save original stderr
	origStderr := os.Stderr

	// Create a pipe to capture stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	// Setup logger and log a message
	SetupLogger(true)
	slog.Error("test message")

	// Close writer and restore stderr
	w.Close()
	os.Stderr = origStderr

	// Read captured output
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	r.Close()

	output := buf.String()
	if output == "" {
		t.Error("expected logger to write to stderr, but got no output")
	}
	if !bytes.Contains([]byte(output), []byte("test message")) {
		t.Errorf("expected output to contain 'test message', got: %s", output)
	}
}
