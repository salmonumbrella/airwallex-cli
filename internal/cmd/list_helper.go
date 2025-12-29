package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

// ListResult represents the result of a paginated list operation
type ListResult[T any] struct {
	Items   []T
	HasMore bool
}

// ListConfig defines how a list command behaves
type ListConfig[T any] struct {
	// Command metadata
	Use     string
	Short   string
	Long    string
	Example string

	// Fetch function - called with page/pageSize, returns items and hasMore
	Fetch func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[T], error)

	// Output configuration
	Headers      []string
	RowFunc      func(T) []string
	ColumnTypes  []outfmt.ColumnType // Optional: column types for colorization
	EmptyMessage string
}

// NewListCommand creates a cobra command from ListConfig
func NewListCommand[T any](cfg ListConfig[T], getClient func(context.Context) (*api.Client, error)) *cobra.Command {
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Enforce minimum page size of 10
			if pageSize < 10 {
				pageSize = 10
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := cfg.Fetch(cmd.Context(), client, page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			// Handle empty results
			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(map[string]interface{}{
						"items":    result.Items,
						"has_more": result.HasMore,
					})
				}
				f.Empty(cfg.EmptyMessage)
				return nil
			}

			// Use OutputListWithColors for consistent sort/limit handling
			// Wrap RowFunc to match OutputList's signature
			rowFn := func(item any) []string {
				return cfg.RowFunc(item.(T))
			}

			if err := f.OutputListWithColors(result.Items, cfg.Headers, cfg.ColumnTypes, rowFn); err != nil {
				return err
			}

			// Show pagination hint for text output
			if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
				fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size (min 10)")
	return cmd
}
