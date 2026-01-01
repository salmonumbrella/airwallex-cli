package factory

import (
	"bytes"
	"context"
	"testing"

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
