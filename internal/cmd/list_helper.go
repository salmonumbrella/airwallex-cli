package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
	Aliases []string
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
	var itemsOnlyFlag bool
	var fetchAll bool

	cmd := &cobra.Command{
		Use:     cfg.Use,
		Aliases: cfg.Aliases,
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

			// When --all is used, fetch all pages using max page size.
			if fetchAll {
				opts.Limit = 100 // max page size
				if mode == PaginationPage {
					opts.Page = 1
				}
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

			// Auto-paginate when --all is set
			if fetchAll && result.HasMore {
				allItems := make([]T, 0, len(result.Items)*2)
				allItems = append(allItems, result.Items...)
				for result.HasMore {
					switch mode {
					case PaginationPage:
						opts.Page++
					case PaginationCursor:
						if cfg.IDFunc != nil && len(result.Items) > 0 {
							opts.Cursor = cfg.IDFunc(result.Items[len(result.Items)-1])
						} else {
							result.HasMore = false
							continue
						}
					}
					switch {
					case cfg.FetchWithArgs != nil:
						result, err = cfg.FetchWithArgs(cmd.Context(), client, opts, args)
					case cfg.Fetch != nil:
						result, err = cfg.Fetch(cmd.Context(), client, opts)
					}
					if err != nil {
						return err
					}
					allItems = append(allItems, result.Items...)
				}
				result.Items = allItems
				result.HasMore = false
			}

			f := outfmt.FromContext(cmd.Context())
			itemsOnly := itemsOnlyFlag || outfmt.GetItemsOnly(cmd.Context())

			// Handle empty results
			if len(result.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					// Ensure empty slice serializes as [] not null
					empty := make([]T, 0)
					if itemsOnly {
						return f.Output(empty)
					}
					return f.Output(map[string]interface{}{
						"items":    empty,
						"has_more": result.HasMore,
					})
				}
				f.Empty(cfg.EmptyMessage)
				return nil
			}

			// For JSON output, include pagination metadata
			if outfmt.IsJSON(cmd.Context()) {
				// Add per-item self links where possible so agents can directly follow up
				// without reconstructing command paths.
				itemsOut := make([]any, 0, len(result.Items))
				itemGetPath := deriveSiblingGetPath(cmd)
				for _, it := range result.Items {
					links := map[string]string{}
					if cfg.IDFunc != nil && itemGetPath != "" {
						id := cfg.IDFunc(it)
						if id != "" {
							links["self"] = buildItemGetLink(itemGetPath, args, id)
						}
					}
					if len(links) > 0 {
						itemsOut = append(itemsOut, outfmt.AnnotatedOutput{Data: it, Links: links})
					} else {
						itemsOut = append(itemsOut, it)
					}
				}

				if itemsOnly {
					return f.Output(itemsOut)
				}
				output := map[string]interface{}{
					"items":    itemsOut,
					"has_more": result.HasMore,
				}
				selfOverride := ""
				switch mode {
				case PaginationCursor:
					if after != "" || cmd.Flags().Changed("after") {
						selfOverride = "after"
					}
				case PaginationPage:
					if page != 1 || cmd.Flags().Changed("page") {
						selfOverride = "page"
					}
				}
				links := map[string]string{"self": buildCommandLink(cmd, mode, page, pageSize, after, limit, selfOverride)}
				if result.HasMore && len(result.Items) > 0 {
					switch mode {
					case PaginationCursor:
						if cfg.IDFunc != nil {
							lastID := cfg.IDFunc(result.Items[len(result.Items)-1])
							output["next_cursor"] = lastID
							links["next"] = buildCommandLink(cmd, mode, page, pageSize, lastID, limit, "after")
						}
					case PaginationPage:
						output["next_page"] = page + 1
						links["next"] = buildCommandLink(cmd, mode, page+1, pageSize, after, limit, "page")
					}
				}
				if mode == PaginationPage && page > 1 {
					links["prev"] = buildCommandLink(cmd, mode, page-1, pageSize, after, limit, "page")
				}
				// Only emit links that are actionable (avoid empty self if this command
				// isn't rooted under "airwallex" in tests/embedding).
				if len(links) > 0 && cmd.Root() != nil && cmd.Root().Use != "" {
					return f.OutputAnnotated(output, links)
				}
				return f.Output(output)
			}

			// Use OutputListWithColors for consistent sort/limit handling
			// Wrap RowFunc to match OutputList's signature
			rowFn := func(item any) []string {
				t, ok := item.(T)
				if !ok {
					return []string{fmt.Sprintf("<%T>", item)}
				}
				return cfg.RowFunc(t)
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
		cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Max items to return (1-100)")
		cmd.Flags().StringVar(&after, "after", "", "Cursor for next page (from previous result)")
		flagAlias(cmd.Flags(), "after", "af")
	case PaginationPage:
		cmd.Flags().IntVarP(&page, "page", "p", 1, "Page number (1+)")
		cmd.Flags().IntVarP(&pageSize, "page-size", "n", 20, "Page size (1-100)")
		flagAlias(cmd.Flags(), "page-size", "ps")
	default:
		panic(fmt.Sprintf("unsupported pagination mode %q", mode))
	}
	cmd.Flags().BoolVarP(&fetchAll, "all", "a", false, "Fetch all pages (auto-paginate)")
	cmd.Flags().BoolVarP(&itemsOnlyFlag, "items-only", "i", false, "Output only the items/results array when present (JSON output)")
	cmd.Flags().BoolVar(&itemsOnlyFlag, "results-only", false, "Alias for --items-only")
	flagAlias(cmd.Flags(), "items-only", "io")
	flagAlias(cmd.Flags(), "results-only", "ro")

	return cmd
}

func buildCommandLink(cmd *cobra.Command, mode PaginationMode, page, pageSize int, after string, limit int, override string) string {
	omit := map[string]bool{
		"help":         true,
		"items-only":   true,
		"results-only": true,
		// Pagination flags (we re-add them via overrides).
		"page":      true,
		"page-size": true,
		"after":     true,
		"limit":     true,
	}

	overrides := map[string]string{
		// Ensure consistent machine-readable output for follow-up calls.
		"output": "json",
	}
	switch mode {
	case PaginationCursor:
		if override == "after" {
			overrides["after"] = after
		}
		// Keep the current limit if the user set it; otherwise omit.
		// (If we include it unconditionally, we'd end up hard-coding defaults.)
		if cmd.Flags().Changed("limit") {
			overrides["limit"] = fmt.Sprintf("%d", limit)
		}
	case PaginationPage:
		if override == "page" {
			overrides["page"] = fmt.Sprintf("%d", page)
		}
		if cmd.Flags().Changed("page-size") {
			overrides["page-size"] = fmt.Sprintf("%d", pageSize)
		}
	default:
		// Unknown mode: don't include pagination overrides.
	}

	return renderCommand(cmd, omit, overrides)
}

func renderCommand(cmd *cobra.Command, omit map[string]bool, overrides map[string]string) string {
	base := strings.TrimSpace(cmd.CommandPath())
	if base == "" {
		// Fallback for commands constructed outside a root; still useful as a hint.
		base = cmd.Use
	}

	parts := []string{base}

	seen := map[string]bool{}
	appendChangedFlags := func(fs *pflag.FlagSet) {
		if fs == nil {
			return
		}
		fs.VisitAll(func(f *pflag.Flag) {
			if f == nil || seen[f.Name] || omit[f.Name] || overrides[f.Name] != "" {
				return
			}
			seen[f.Name] = true
			if !f.Changed {
				return
			}

			val := f.Value.String()
			// Booleans: emit --flag or --flag=false
			if f.Value.Type() == "bool" {
				if val == "true" {
					parts = append(parts, "--"+f.Name)
				} else {
					parts = append(parts, "--"+f.Name+"="+val)
				}
				return
			}

			parts = append(parts, "--"+f.Name, shellQuote(val))
		})
	}

	// Include both local flags (filters) and inherited flags (global output knobs like --query).
	appendChangedFlags(cmd.Flags())
	appendChangedFlags(cmd.InheritedFlags())

	// Apply overrides last.
	for k, v := range overrides {
		if v == "" {
			continue
		}
		switch v {
		case "true":
			parts = append(parts, "--"+k)
		case "false":
			parts = append(parts, "--"+k+"=false")
		default:
			parts = append(parts, "--"+k, shellQuote(v))
		}
	}

	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	// Single-quote for POSIX shells; escape embedded single-quotes.
	if s == "" {
		return "''"
	}
	needs := strings.ContainsAny(s, " \t\r\n'\"\\$`")
	if !needs {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func deriveSiblingGetPath(cmd *cobra.Command) string {
	path := strings.TrimSpace(cmd.CommandPath())
	if !strings.HasSuffix(path, " list") {
		return ""
	}
	return strings.TrimSuffix(path, " list") + " get"
}

func buildItemGetLink(getPath string, parentArgs []string, id string) string {
	parts := []string{getPath}
	for _, a := range parentArgs {
		parts = append(parts, shellQuote(NormalizeIDArg(a)))
	}
	parts = append(parts, shellQuote(id), "--output", "json")
	return strings.Join(parts, " ")
}
