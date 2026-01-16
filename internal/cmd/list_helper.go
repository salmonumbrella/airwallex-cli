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
	Page int
}

// PaginationMode indicates which pagination model a list command uses.
type PaginationMode string

const (
	// PaginationPage uses page_num/page_size style pagination.
	PaginationPage PaginationMode = "page"
	// PaginationCursor uses after_id/limit style pagination.
	PaginationCursor PaginationMode = "cursor"
)

// ListConfig defines how a list command behaves
type ListConfig[T any] struct {
	// Command metadata
	Use     string
	Short   string
	Long    string
	Example string

	// Fetch function - called with pagination options
	Fetch func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[T], error)
	// FetchWithArgs is an optional variant that also receives positional args.
	FetchWithArgs func(ctx context.Context, client *api.Client, opts ListOptions, args []string) (ListResult[T], error)

	// Args configures cobra positional args validation.
	Args cobra.PositionalArgs

	// Output configuration
	Headers      []string
	RowFunc      func(T) []string
	ColumnTypes  []outfmt.ColumnType // Optional: column types for colorization
	EmptyMessage string

	// IDFunc extracts ID from item for cursor-based pagination
	// If nil, next cursor hint won't be shown
	IDFunc func(T) string

	// MoreHint overrides the default "next page" hint when HasMore is true.
	MoreHint string

	// Pagination configures which pagination model the endpoint uses.
	// Defaults to PaginationPage.
	Pagination PaginationMode
}

// NewListCommand creates a cobra command from ListConfig
func NewListCommand[T any](cfg ListConfig[T], getClient func(context.Context) (*api.Client, error)) *cobra.Command {
	var limit int
	var after string
	var page int
	var pageSize int
	var itemsOnly bool

	cmd := &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := cfg.Pagination
			if mode == "" {
				mode = PaginationPage
			}

			// Enforce limits and pagination defaults
			switch mode {
			case PaginationCursor:
				if limit <= 0 {
					limit = 20
				}
				if limit > 100 {
					limit = 100
				}
			case PaginationPage:
				if pageSize <= 0 {
					pageSize = 20
				}
				if pageSize > 100 {
					pageSize = 100
				}
				if page <= 0 {
					page = 1
				}
			default:
				return fmt.Errorf("unknown pagination mode %q", mode)
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			opts := ListOptions{
				Options: pagination.Options{
					Limit:  limit,
					Cursor: after,
				},
				Page: page,
			}
			if mode == PaginationPage {
				opts.Limit = pageSize
			}
			var result ListResult[T]
			switch {
			case cfg.FetchWithArgs != nil:
				result, err = cfg.FetchWithArgs(cmd.Context(), client, opts, args)
			case cfg.Fetch != nil:
				result, err = cfg.Fetch(cmd.Context(), client, opts)
			default:
				return fmt.Errorf("list command missing Fetch")
			}
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
				if result.HasMore && len(result.Items) > 0 {
					switch mode {
					case PaginationCursor:
						if cfg.IDFunc != nil {
							lastID := cfg.IDFunc(result.Items[len(result.Items)-1])
							output["next_cursor"] = lastID
						}
					case PaginationPage:
						output["next_page"] = page + 1
					}
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
			if result.HasMore {
				if cfg.MoreHint != "" {
					fmt.Fprintln(os.Stderr, cfg.MoreHint)
				} else {
					switch mode {
					case PaginationCursor:
						if cfg.IDFunc != nil {
							lastID := cfg.IDFunc(result.Items[len(result.Items)-1])
							fmt.Fprintf(os.Stderr, "# More results available. Next page: --after %s\n", lastID)
						}
					case PaginationPage:
						fmt.Fprintf(os.Stderr, "# More results available. Next page: --page %d\n", page+1)
					}
				}
			}
			return nil
		},
	}
	if cfg.Args != nil {
		cmd.Args = cfg.Args
	}

	mode := cfg.Pagination
	if mode == "" {
		mode = PaginationPage
	}
	switch mode {
	case PaginationCursor:
		cmd.Flags().IntVar(&limit, "limit", 20, "Max items to return (1-100)")
		cmd.Flags().StringVar(&after, "after", "", "Cursor for next page (from previous result)")
	case PaginationPage:
		cmd.Flags().IntVar(&page, "page", 1, "Page number (1+)")
		cmd.Flags().IntVar(&pageSize, "page-size", 20, "Page size (1-100)")
	default:
		panic(fmt.Sprintf("unsupported pagination mode %q", mode))
	}
	cmd.Flags().BoolVar(&itemsOnly, "items-only", false, "Output items array only (JSON mode)")
	cmd.Flags().BoolVar(&itemsOnly, "results-only", false, "Alias for --items-only")

	return cmd
}
