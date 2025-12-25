package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestIO_DefaultIO(t *testing.T) {
	io := DefaultIO()
	if io == nil {
		t.Fatal("DefaultIO() returned nil")
	}
	if io.Out == nil || io.ErrOut == nil || io.In == nil {
		t.Error("DefaultIO() should have non-nil streams")
	}
}

func TestIO_WithIO(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	customIO := &IO{
		Out:    &outBuf,
		ErrOut: &errBuf,
		In:     strings.NewReader("test input"),
	}

	ctx := WithIO(context.Background(), customIO)
	retrieved := GetIO(ctx)

	if retrieved != customIO {
		t.Error("GetIO() did not return the IO that was set with WithIO()")
	}
}

func TestIO_GetIO_DefaultsWhenNotSet(t *testing.T) {
	ctx := context.Background()
	io := GetIO(ctx)

	if io == nil {
		t.Fatal("GetIO() should never return nil")
	}
	// When IO is not in context, GetIO should return default streams
	// We can't check exact equality with os.Stdout since DefaultIO creates a new struct
	if io.Out == nil || io.ErrOut == nil || io.In == nil {
		t.Error("GetIO() without context should return default streams")
	}
}

// TestVersionCommand_CapturesOutput demonstrates command-level output capture via injected IO.
func TestVersionCommand_CapturesOutput(t *testing.T) {
	// Create custom IO to capture output
	var outBuf, errBuf bytes.Buffer
	customIO := &IO{
		Out:    &outBuf,
		ErrOut: &errBuf,
		In:     strings.NewReader(""),
	}

	// Create context with custom IO
	ctx := WithIO(context.Background(), customIO)

	// Create and execute the version command
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"version"})

	// Execute with custom context
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Verify output was captured
	output := outBuf.String()
	if !strings.Contains(output, "airwallex-cli") {
		t.Errorf("Expected output to contain 'airwallex-cli', got:\n%s", output)
	}
	if !strings.Contains(output, "commit:") {
		t.Errorf("Expected output to contain 'commit:', got:\n%s", output)
	}
	if !strings.Contains(output, "build date:") {
		t.Errorf("Expected output to contain 'build date:', got:\n%s", output)
	}

	// Verify nothing was written to os.Stdout (all output went to our buffer)
	if output == "" {
		t.Error("No output was captured - IO injection may not be working")
	}
}

// TestVersionCommand_JSONOutputCapture demonstrates JSON output capture via injected IO.
func TestVersionCommand_JSONOutputCapture(t *testing.T) {
	// Create custom IO to capture output
	var outBuf, errBuf bytes.Buffer
	customIO := &IO{
		Out:    &outBuf,
		ErrOut: &errBuf,
		In:     strings.NewReader(""),
	}

	// Create context with custom IO
	ctx := WithIO(context.Background(), customIO)

	// Create and execute the version command with JSON output
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"version", "--output", "json"})

	// Execute with custom context
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Verify JSON output was captured
	output := outBuf.String()
	if !strings.Contains(output, `"version"`) {
		t.Errorf("Expected JSON output to contain '\"version\"', got:\n%s", output)
	}
	if !strings.Contains(output, `"commit"`) {
		t.Errorf("Expected JSON output to contain '\"commit\"', got:\n%s", output)
	}
	if !strings.Contains(output, `"build_date"`) {
		t.Errorf("Expected JSON output to contain '\"build_date\"', got:\n%s", output)
	}
}
