package factory

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

// Factory provides dependencies for commands.
type Factory struct {
	// IO streams
	IO *iocontext.IO

	// UI for colored output
	UI *ui.UI

	// Client creates an API client (lazy)
	Client func(ctx context.Context) (*api.Client, error)

	// Config provides CLI configuration
	Config func() (*Config, error)

	// Secrets provides credential storage
	Secrets func() (secrets.Store, error)

	// AgentMode indicates the CLI is being used by an automated agent
	AgentMode bool
}

// Config holds CLI configuration.
type Config struct {
	Account      string
	OutputFormat string
	Color        string
	Debug        bool
}

// New creates a Factory with default implementations.
func New() *Factory {
	return &Factory{
		IO: iocontext.DefaultIO(),
		UI: ui.New("auto"),
		Client: func(ctx context.Context) (*api.Client, error) {
			return nil, fmt.Errorf("API client not configured; ensure command is run from root")
		},
		Config: func() (*Config, error) {
			return &Config{}, nil
		},
		Secrets: secrets.OpenDefault,
	}
}

// WithIO returns a copy with custom IO streams.
func (f *Factory) WithIO(io *iocontext.IO) *Factory {
	cp := *f
	cp.IO = io
	return &cp
}

// WithAgentMode returns a copy configured for agent usage.
func (f *Factory) WithAgentMode() *Factory {
	cp := *f
	cp.AgentMode = true
	return &cp
}

// WithClient returns a copy with a custom client factory.
func (f *Factory) WithClient(client func(ctx context.Context) (*api.Client, error)) *Factory {
	cp := *f
	cp.Client = client
	return &cp
}

// WithConfig returns a copy with a custom config factory.
func (f *Factory) WithConfig(config func() (*Config, error)) *Factory {
	cp := *f
	cp.Config = config
	return &cp
}

// GetIO returns IO streams with proper fallback chain:
//
//  1. Factory IO - if f != nil && f.IO != nil, use factory's IO streams
//  2. Context IO - otherwise, use iocontext.GetIO(ctx) which returns DefaultIO() if not set
//  3. Cobra override - if out is still os.Stdout, check cmd.OutOrStdout()
//     (same for errOut with os.Stderr via cmd.OutOrStderr())
//
// This allows tests to inject custom writers at any level while ensuring
// production code falls back to sensible defaults. Safe to call with nil factory.
func (f *Factory) GetIO(cmd *cobra.Command) (out, errOut io.Writer) {
	var ioCtx *iocontext.IO
	if f != nil && f.IO != nil {
		ioCtx = f.IO
	} else {
		ioCtx = iocontext.GetIO(cmd.Context())
	}
	out = ioCtx.Out
	errOut = ioCtx.ErrOut

	// If IO is still the default (os.Stdout/os.Stderr), check if cobra has custom writers
	if out == os.Stdout {
		out = cmd.OutOrStdout()
	}
	if errOut == os.Stderr {
		errOut = cmd.OutOrStderr()
	}

	return out, errOut
}
