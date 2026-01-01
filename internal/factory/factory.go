package factory

import (
	"context"

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
			// Default implementation - will be set properly by root command
			return nil, nil
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
