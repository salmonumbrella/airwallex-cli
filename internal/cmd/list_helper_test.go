package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

type testItem struct {
	ID   string
	Name string
}

func TestNewListCommand_PaginationDefaults(t *testing.T) {
	var capturedOpts ListOptions

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		IDFunc: func(item testItem) string {
			return item.ID
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			capturedOpts = opts
			return ListResult[testItem]{
				Items:   []testItem{{ID: "1", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify default pagination values
	if capturedOpts.Limit != 20 {
		t.Errorf("expected page size 20, got %d", capturedOpts.Limit)
	}
	if capturedOpts.Cursor != "" {
		t.Errorf("expected empty cursor, got %q", capturedOpts.Cursor)
	}
	if capturedOpts.Page != 1 {
		t.Errorf("expected page 1, got %d", capturedOpts.Page)
	}
}

func TestNewListCommand_PageSizeEnforcement(t *testing.T) {
	tests := []struct {
		name          string
		inputPageSize string
		expectedLimit int
	}{
		{"below minimum defaults to 20", "0", 20},
		{"negative defaults to 20", "-5", 20},
		{"at minimum", "1", 1},
		{"normal value", "50", 50},
		{"at maximum", "100", 100},
		{"above maximum capped", "200", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedOpts ListOptions

			cfg := ListConfig[testItem]{
				Use:          "test",
				Short:        "Test list command",
				Headers:      []string{"ID", "NAME"},
				EmptyMessage: "No items",
				RowFunc: func(item testItem) []string {
					return []string{item.ID, item.Name}
				},
				Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
					capturedOpts = opts
					return ListResult[testItem]{
						Items:   []testItem{{ID: "1", Name: "Test"}},
						HasMore: false,
					}, nil
				},
			}

			cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
				return &api.Client{}, nil
			})

			ctx := outfmt.WithFormat(context.Background(), "text")
			cmd.SetContext(ctx)
			cmd.SetArgs([]string{"--page-size", tt.inputPageSize})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if capturedOpts.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, capturedOpts.Limit)
			}
		})
	}
}

func TestNewListCommand_CursorMode(t *testing.T) {
	var capturedOpts ListOptions

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Pagination:   PaginationCursor,
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			capturedOpts = opts
			return ListResult[testItem]{
				Items:   []testItem{{ID: "2", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--after", "cursor_abc123", "--limit", "30"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedOpts.Cursor != "cursor_abc123" {
		t.Errorf("expected cursor 'cursor_abc123', got %q", capturedOpts.Cursor)
	}
	if capturedOpts.Limit != 30 {
		t.Errorf("expected limit 30, got %d", capturedOpts.Limit)
	}
}

func TestNewListCommand_PageSizeFlag(t *testing.T) {
	var capturedOpts ListOptions

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			capturedOpts = opts
			return ListResult[testItem]{
				Items:   []testItem{{ID: "1", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--page-size", "30"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedOpts.Limit != 30 {
		t.Errorf("expected limit 30, got %d", capturedOpts.Limit)
	}
}

func TestNewListCommand_PageFlag(t *testing.T) {
	var capturedOpts ListOptions

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			capturedOpts = opts
			return ListResult[testItem]{
				Items:   []testItem{{ID: "1", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--page", "2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedOpts.Page != 2 {
		t.Errorf("expected page 2, got %d", capturedOpts.Page)
	}
}

func TestNewListCommand_EmptyResults(t *testing.T) {
	var emptyCalled bool

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items found",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			emptyCalled = true
			return ListResult[testItem]{
				Items:   []testItem{},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !emptyCalled {
		t.Error("expected Fetch to be called for empty results")
	}
}

func TestNewListCommand_FetchWithArgs(t *testing.T) {
	var capturedArg string

	cfg := ListConfig[testItem]{
		Use:          "test <id>",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		Args:         cobra.ExactArgs(1),
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		FetchWithArgs: func(ctx context.Context, client *api.Client, opts ListOptions, args []string) (ListResult[testItem], error) {
			capturedArg = args[0]
			return ListResult[testItem]{
				Items:   []testItem{{ID: "1", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"cust_123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedArg != "cust_123" {
		t.Fatalf("expected arg cust_123, got %q", capturedArg)
	}
}

func TestNewListCommand_MoreResultsMessage(t *testing.T) {
	tests := []struct {
		name    string
		hasMore bool
	}{
		{name: "has more results", hasMore: true},
		{name: "no more results", hasMore: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hasMoreReturned bool

			cfg := ListConfig[testItem]{
				Use:          "test",
				Short:        "Test list command",
				Headers:      []string{"ID", "NAME"},
				EmptyMessage: "No items",
				RowFunc: func(item testItem) []string {
					return []string{item.ID, item.Name}
				},
				IDFunc: func(item testItem) string {
					return item.ID
				},
				Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
					hasMoreReturned = tt.hasMore
					return ListResult[testItem]{
						Items:   []testItem{{ID: "1", Name: "Test"}},
						HasMore: tt.hasMore,
					}, nil
				},
			}

			cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
				return &api.Client{}, nil
			})

			ctx := outfmt.WithFormat(context.Background(), "text")
			cmd.SetContext(ctx)
			cmd.SetArgs([]string{})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if hasMoreReturned != tt.hasMore {
				t.Errorf("expected hasMore=%v", tt.hasMore)
			}
		})
	}
}

func TestNewListCommand_JSONOutput(t *testing.T) {
	var jsonFormatDetected bool

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		IDFunc: func(item testItem) string {
			return item.ID
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			jsonFormatDetected = outfmt.IsJSON(ctx)
			return ListResult[testItem]{
				Items: []testItem{
					{ID: "1", Name: "Item 1"},
					{ID: "2", Name: "Item 2"},
				},
				HasMore: true,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "json")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !jsonFormatDetected {
		t.Error("expected JSON format to be detected in context")
	}
}

func TestNewListCommand_JSONItemsOnly(t *testing.T) {
	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			return ListResult[testItem]{
				Items: []testItem{
					{ID: "1", Name: "Item 1"},
				},
				HasMore: true,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "json")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--items-only"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewListCommand_JSONItemsOnly_FromGlobalContext(t *testing.T) {
	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			return ListResult[testItem]{
				Items: []testItem{
					{ID: "1", Name: "Item 1"},
				},
				HasMore: true,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	var out bytes.Buffer
	var errOut bytes.Buffer
	ctx := outfmt.WithFormat(context.Background(), "json")
	ctx = outfmt.WithItemsOnly(ctx, true)
	ctx = iocontext.WithIO(ctx, &iocontext.IO{
		Out:    &out,
		ErrOut: &errOut,
		In:     bytes.NewBuffer(nil),
	})
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON output: %v", err)
	}
	if _, ok := decoded.([]any); !ok {
		t.Fatalf("expected array output when global selector is enabled, got %T", decoded)
	}
}

func TestNewListCommand_JSONItemsOnlyEmpty(t *testing.T) {
	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			return ListResult[testItem]{
				Items:   []testItem{},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "json")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--items-only"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewListCommand_JSONLinksAndItemLinks(t *testing.T) {
	cfg := ListConfig[testItem]{
		Use:          "list",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		IDFunc: func(item testItem) string { return item.ID },
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			return ListResult[testItem]{
				Items:   []testItem{{ID: "tfr_123", Name: "One"}},
				HasMore: true,
			}, nil
		},
	}
	listCmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) { return &api.Client{}, nil })

	// Build a command path that ends with " list" so per-item links can be derived.
	root := &cobra.Command{Use: "airwallex"}
	res := &cobra.Command{Use: "transfers"}
	res.AddCommand(listCmd)
	root.AddCommand(res)

	var out bytes.Buffer
	var errOut bytes.Buffer
	ctx := outfmt.WithFormat(context.Background(), "json")
	ctx = iocontext.WithIO(ctx, &iocontext.IO{Out: &out, ErrOut: &errOut, In: bytes.NewBuffer(nil)})

	root.SetContext(ctx)
	root.SetArgs([]string{"transfers", "list", "--page", "2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	links, ok := decoded["_links"].(map[string]any)
	if !ok {
		t.Fatalf("expected _links object in output, got: %T", decoded["_links"])
	}
	self, _ := links["self"].(string)
	if self == "" {
		t.Fatalf("expected _links.self to be set, got: %v", links["self"])
	}

	items, ok := decoded["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected items array length 1, got: %T len=%d", decoded["items"], len(items))
	}
	item0, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected first item to be an object, got: %T", items[0])
	}
	itemLinks, ok := item0["_links"].(map[string]any)
	if !ok {
		t.Fatalf("expected items[0]._links object, got: %T", item0["_links"])
	}
	itemSelf, _ := itemLinks["self"].(string)
	if itemSelf == "" {
		t.Fatalf("expected items[0]._links.self to be set, got: %v", itemLinks["self"])
	}
}

func TestNewListCommand_FetchError(t *testing.T) {
	expectedErr := errors.New("fetch failed")

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			return ListResult[testItem]{}, expectedErr
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewListCommand_ClientError(t *testing.T) {
	expectedErr := errors.New("client creation failed")

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			return ListResult[testItem]{
				Items:   []testItem{{ID: "1", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return nil, expectedErr
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewListCommand_TextTableOutput(t *testing.T) {
	var itemCount int

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			items := []testItem{
				{ID: "1", Name: "Item One"},
				{ID: "2", Name: "Item Two"},
			}
			itemCount = len(items)
			return ListResult[testItem]{
				Items:   items,
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if itemCount != 2 {
		t.Errorf("expected 2 items to be returned, got %d", itemCount)
	}
}

func TestNewListCommand_CustomFlagsCapture(t *testing.T) {
	// Simulate the pattern used in deposits.go and other migrated commands
	// where custom flags are captured by the Fetch closure
	var customStatus string
	var capturedStatus string

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[testItem], error) {
			// Capture the custom flag value inside the closure
			capturedStatus = customStatus
			return ListResult[testItem]{
				Items:   []testItem{{ID: "1", Name: "Test"}},
				HasMore: false,
			}, nil
		},
	}

	cmd := NewListCommand(cfg, func(ctx context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	// Add custom flag that captures into the closure variable
	cmd.Flags().StringVar(&customStatus, "status", "", "Filter by status")

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"--status", "SETTLED"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the custom flag value was correctly captured inside Fetch
	if capturedStatus != "SETTLED" {
		t.Errorf("expected captured status 'SETTLED', got '%s'", capturedStatus)
	}
}
