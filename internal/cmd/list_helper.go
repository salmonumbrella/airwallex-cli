package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/pagination"
)

// ListResult represents the result of a paginated list operation
type ListResult[T any] struct {
	Items   []T
	HasMore bool
}

// ListOptions provides cursor-based pagination parameters.
// Cursor uses the same value as the --after flag in the CLI.
type ListOptions struct {
	pagination.Options
}

// ListConfig defines how a list command behaves
type ListConfig[T any] struct {
	// Command metadata
	Use     string
	Short   string
	Long    string
	Example string

	// Fetch function - called with pagination options
	Fetch func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[T], error)

	// Output configuration
	Headers      []string
	RowFunc      func(T) []string
	ColumnTypes  []outfmt.ColumnType // Optional: column types for colorization
	EmptyMessage string

	// IDFunc extracts ID from item for cursor-based pagination
	// If nil, next cursor hint won't be shown
	IDFunc func(T) string
}

// NewListCommand creates a cobra command from ListConfig
func NewListCommand[T any](cfg ListConfig[T], getClient func(context.Context) (*api.Client, error)) *cobra.Command {
	var limit int
	var after string
	var itemsOnly bool

	cmd := &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Enforce limits
			if limit <= 0 {
				limit = 20
			}
			if limit > 100 {
				limit = 100
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			result, err := cfg.Fetch(cmd.Context(), client, ListOptions{
				Options: pagination.Options{
					Limit:  limit,
					Cursor: after,
				},
			})
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			// Handle empty results
			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					if itemsOnly {
						return f.Output(result.Items)
					}
					return f.Output(map[string]interface{}{
						"items":    result.Items,
						"has_more": result.HasMore,
					})
				}
				f.Empty(cfg.EmptyMessage)
				return nil
			}

			// For JSON output, include pagination metadata
			if outfmt.IsJSON(cmd.Context()) {
				if itemsOnly {
					return f.Output(result.Items)
				}
				output := map[string]interface{}{
					"items":    result.Items,
					"has_more": result.HasMore,
				}
				if result.HasMore && len(result.Items) > 0 && cfg.IDFunc != nil {
					lastID := cfg.IDFunc(result.Items[len(result.Items)-1])
					output["next_cursor"] = lastID
				}
				return f.Output(output)
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
			if result.HasMore && cfg.IDFunc != nil {
				lastID := cfg.IDFunc(result.Items[len(result.Items)-1])
				fmt.Fprintf(os.Stderr, "# More results available. Next page: --after %s\n", lastID)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Max items to return (1-100)")
	cmd.Flags().StringVar(&after, "after", "", "Cursor for next page (from previous result)")
	cmd.Flags().BoolVar(&itemsOnly, "items-only", false, "Output items array only (JSON mode)")
	cmd.Flags().BoolVar(&itemsOnly, "results-only", false, "Alias for --items-only")

	// Keep deprecated flags for backwards compatibility
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().Int("page-size", 0, "")
	_ = cmd.Flags().MarkDeprecated("page", "use --after for cursor-based pagination")
	_ = cmd.Flags().MarkDeprecated("page-size", "use --limit instead")

	return cmd
}
