package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

// GetConfig defines how a get command behaves.
type GetConfig[T any] struct {
	Use        string
	Short      string
	Long       string
	Example    string
	Fetch      func(ctx context.Context, client *api.Client, id string) (T, error)
	TextOutput func(cmd *cobra.Command, item T) error
}

// NewGetCommand creates a get command with consistent JSON/template handling.
func NewGetCommand[T any](cfg GetConfig[T], getClient func(context.Context) (*api.Client, error)) *cobra.Command {
	return &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			item, err := cfg.Fetch(cmd.Context(), client, args[0])
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())
			if outfmt.GetTemplate(cmd.Context()) != "" || outfmt.IsJSON(cmd.Context()) {
				return f.Output(item)
			}

			if cfg.TextOutput == nil {
				return nil
			}
			return cfg.TextOutput(cmd, item)
		},
	}
}
