package factory

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
)

func TestNew(t *testing.T) {
	f := New()

	if f.IO == nil {
		t.Error("IO should not be nil")
	}
	if f.UI == nil {
		t.Error("UI should not be nil")
	}
	if f.Client == nil {
		t.Error("Client should not be nil")
	}
	if f.Config == nil {
		t.Error("Config should not be nil")
	}
	if f.Secrets == nil {
		t.Error("Secrets should not be nil")
	}
	if f.AgentMode {
		t.Error("AgentMode should default to false")
	}
}

func TestNew_DefaultClientReturnsError(t *testing.T) {
	f := New()

	client, err := f.Client(context.Background())
	if err == nil {
		t.Error("Default Client() should return an error")
	}
	if client != nil {
		t.Error("Default Client() should return nil client")
	}
	if err.Error() != "API client not configured; ensure command is run from root" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFactory_WithIO(t *testing.T) {
	f := New()

	var buf bytes.Buffer
	customIO := &iocontext.IO{Out: &buf}

	f2 := f.WithIO(customIO)

	if f2.IO != customIO {
		t.Error("WithIO did not set IO")
	}
	if f.IO == customIO {
		t.Error("WithIO modified original factory")
	}
}

func TestFactory_WithAgentMode(t *testing.T) {
	f := New()
	if f.AgentMode {
		t.Error("AgentMode should default to false")
	}

	f2 := f.WithAgentMode()
	if !f2.AgentMode {
		t.Error("WithAgentMode did not set AgentMode")
	}
	if f.AgentMode {
		t.Error("WithAgentMode modified original factory")
	}
}

func TestFactory_WithClient(t *testing.T) {
	f := New()

	called := false
	customClient := func(ctx context.Context) (*api.Client, error) {
		called = true
		return nil, nil
	}

	f2 := f.WithClient(customClient)

	// Call the client function to verify it was set
	_, _ = f2.Client(context.Background())
	if !called {
		t.Error("WithClient did not set custom client function")
	}
}

func TestFactory_WithConfig(t *testing.T) {
	f := New()

	customConfig := &Config{
		Account:      "test-account",
		OutputFormat: "json",
		Color:        "never",
		Debug:        true,
	}
	customConfigFn := func() (*Config, error) {
		return customConfig, nil
	}

	f2 := f.WithConfig(customConfigFn)

	cfg, err := f2.Config()
	if err != nil {
		t.Fatalf("Config() returned error: %v", err)
	}
	if cfg.Account != "test-account" {
		t.Errorf("Config().Account = %q, want %q", cfg.Account, "test-account")
	}
	if cfg.Debug != true {
		t.Error("Config().Debug should be true")
	}
}

func TestFactory_GetIO_WithFactoryIO(t *testing.T) {
	// Factory with custom IO should use factory IO
	var outBuf, errBuf bytes.Buffer
	customIO := &iocontext.IO{Out: &outBuf, ErrOut: &errBuf}
	f := New().WithIO(customIO)

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	out, errOut := f.GetIO(cmd)

	if out != &outBuf {
		t.Error("GetIO should return factory's Out writer")
	}
	if errOut != &errBuf {
		t.Error("GetIO should return factory's ErrOut writer")
	}
}

func TestFactory_GetIO_NilFactory(t *testing.T) {
	// Nil factory should fall back to context IO
	var outBuf, errBuf bytes.Buffer
	ctxIO := &iocontext.IO{Out: &outBuf, ErrOut: &errBuf}

	cmd := &cobra.Command{}
	ctx := iocontext.WithIO(context.Background(), ctxIO)
	cmd.SetContext(ctx)

	var f *Factory = nil
	out, errOut := f.GetIO(cmd)

	if out != &outBuf {
		t.Error("GetIO with nil factory should return context's Out writer")
	}
	if errOut != &errBuf {
		t.Error("GetIO with nil factory should return context's ErrOut writer")
	}
}

func TestFactory_GetIO_NilIO(t *testing.T) {
	// Factory with nil IO should fall back to context IO
	var outBuf, errBuf bytes.Buffer
	ctxIO := &iocontext.IO{Out: &outBuf, ErrOut: &errBuf}

	f := &Factory{IO: nil}

	cmd := &cobra.Command{}
	ctx := iocontext.WithIO(context.Background(), ctxIO)
	cmd.SetContext(ctx)

	out, errOut := f.GetIO(cmd)

	if out != &outBuf {
		t.Error("GetIO with nil IO should return context's Out writer")
	}
	if errOut != &errBuf {
		t.Error("GetIO with nil IO should return context's ErrOut writer")
	}
}

func TestFactory_GetIO_CobraOutWriterWhenDefaultIO(t *testing.T) {
	// When factory IO is os.Stdout, cobra Out writer should take precedence
	defaultIO := iocontext.DefaultIO() // os.Stdout, os.Stderr
	f := New().WithIO(defaultIO)

	var cobraOut bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&cobraOut)

	out, _ := f.GetIO(cmd)

	if out != &cobraOut {
		t.Error("GetIO should return cobra's Out writer when factory IO is os.Stdout")
	}
}

func TestFactory_GetIO_CobraErrWriterBehavior(t *testing.T) {
	// Note: The implementation uses cmd.OutOrStderr() not cmd.ErrOrStderr()
	// This means cmd.SetErr() is not respected - only cmd.SetOut() affects error output
	// when the factory IO is os.Stderr
	defaultIO := iocontext.DefaultIO() // os.Stdout, os.Stderr
	f := New().WithIO(defaultIO)

	var cobraOut bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&cobraOut)
	// Note: cmd.SetErr is NOT used by GetIO - it uses OutOrStderr which checks outWriter

	_, errOut := f.GetIO(cmd)

	// OutOrStderr returns the out writer if set, otherwise stderr
	// So with SetOut(&cobraOut), OutOrStderr returns &cobraOut
	if errOut != &cobraOut {
		t.Error("GetIO uses cmd.OutOrStderr() which returns Out writer when set")
	}
}

func TestFactory_GetIO_FactoryIOTakesPrecedenceOverCobra(t *testing.T) {
	// Custom factory IO should take precedence over cobra writers
	var factoryOut, factoryErr bytes.Buffer
	customIO := &iocontext.IO{Out: &factoryOut, ErrOut: &factoryErr}
	f := New().WithIO(customIO)

	var cobraOut, cobraErr bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&cobraOut)
	cmd.SetErr(&cobraErr)

	out, errOut := f.GetIO(cmd)

	if out != &factoryOut {
		t.Error("Custom factory IO should take precedence over cobra writers for Out")
	}
	if errOut != &factoryErr {
		t.Error("Custom factory IO should take precedence over cobra writers for ErrOut")
	}
}

func TestFactory_GetIO_DefaultBehavior(t *testing.T) {
	// When no custom IO anywhere, should return os.Stdout/os.Stderr
	f := &Factory{IO: nil}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	out, errOut := f.GetIO(cmd)

	if out != os.Stdout {
		t.Error("GetIO should default to os.Stdout when no custom IO")
	}
	if errOut != os.Stderr {
		t.Error("GetIO should default to os.Stderr when no custom IO")
	}
}

func TestFactory_GetIO_PartialCobraOverride(t *testing.T) {
	// Test when only cobra Out writer is set
	// Note: Because GetIO uses OutOrStderr() for errOut, setting Out affects errOut too
	defaultIO := iocontext.DefaultIO()
	f := New().WithIO(defaultIO)

	var cobraOut bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&cobraOut)

	out, errOut := f.GetIO(cmd)

	if out != &cobraOut {
		t.Error("GetIO should return cobra's Out writer")
	}
	// OutOrStderr() returns Out if set, so errOut also becomes cobraOut
	if errOut != &cobraOut {
		t.Error("GetIO uses OutOrStderr which returns Out writer when set")
	}
}

func TestFactory_GetIO_NoCobraWritersSet(t *testing.T) {
	// When no cobra writers are set and using default IO, falls back to os.Stdout/os.Stderr
	defaultIO := iocontext.DefaultIO()
	f := New().WithIO(defaultIO)

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	// Neither SetOut nor SetErr called

	out, errOut := f.GetIO(cmd)

	if out != os.Stdout {
		t.Error("GetIO should return os.Stdout when no custom writers")
	}
	if errOut != os.Stderr {
		t.Error("GetIO should return os.Stderr when no custom writers")
	}
}

func TestFactory_GetIO_ContextIOWithCobraWriters(t *testing.T) {
	// Context with default IO should still allow cobra override
	// Note: Only Out writer is respected; SetErr is not used (OutOrStderr checks Out)
	ctxIO := iocontext.DefaultIO() // os.Stdout, os.Stderr

	var f *Factory = nil

	var cobraOut bytes.Buffer
	cmd := &cobra.Command{}
	ctx := iocontext.WithIO(context.Background(), ctxIO)
	cmd.SetContext(ctx)
	cmd.SetOut(&cobraOut)

	out, errOut := f.GetIO(cmd)

	if out != &cobraOut {
		t.Error("Cobra Out writer should override default context IO Out")
	}
	// OutOrStderr returns Out if set
	if errOut != &cobraOut {
		t.Error("OutOrStderr returns Out writer when set")
	}
}

func TestFactory_GetIO_CustomContextIONotOverriddenByCobra(t *testing.T) {
	// Custom context IO should take precedence over cobra writers
	var ctxOut, ctxErr bytes.Buffer
	ctxIO := &iocontext.IO{Out: &ctxOut, ErrOut: &ctxErr}

	var f *Factory = nil

	var cobraOut, cobraErr bytes.Buffer
	cmd := &cobra.Command{}
	ctx := iocontext.WithIO(context.Background(), ctxIO)
	cmd.SetContext(ctx)
	cmd.SetOut(&cobraOut)
	cmd.SetErr(&cobraErr)

	out, errOut := f.GetIO(cmd)

	if out != &ctxOut {
		t.Error("Custom context IO should take precedence over cobra writers for Out")
	}
	if errOut != &ctxErr {
		t.Error("Custom context IO should take precedence over cobra writers for ErrOut")
	}
}
