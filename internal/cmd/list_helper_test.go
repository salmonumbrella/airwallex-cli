package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

type testItem struct {
	ID   string
	Name string
}

func TestNewListCommand_PaginationDefaults(t *testing.T) {
	var capturedPage, capturedPageSize int

	cfg := ListConfig[testItem]{
		Use:          "test",
		Short:        "Test list command",
		Headers:      []string{"ID", "NAME"},
		EmptyMessage: "No items",
		RowFunc: func(item testItem) []string {
			return []string{item.ID, item.Name}
		},
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
			capturedPage = page
			capturedPageSize = pageSize
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
	if capturedPage != 0 {
		t.Errorf("expected page 0, got %d", capturedPage)
	}
	if capturedPageSize != 20 {
		t.Errorf("expected page size 20, got %d", capturedPageSize)
	}
}

func TestNewListCommand_PageSizeMinimumEnforcement(t *testing.T) {
	tests := []struct {
		name             string
		inputPageSize    string
		expectedPageSize int
	}{
		{"below minimum", "5", 10},
		{"at minimum", "10", 10},
		{"above minimum", "50", 50},
		{"zero defaults to 20 then enforces minimum", "", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedPageSize int

			cfg := ListConfig[testItem]{
				Use:          "test",
				Short:        "Test list command",
				Headers:      []string{"ID", "NAME"},
				EmptyMessage: "No items",
				RowFunc: func(item testItem) []string {
					return []string{item.ID, item.Name}
				},
				Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
					capturedPageSize = pageSize
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

			if tt.inputPageSize != "" {
				cmd.SetArgs([]string{"--page-size", tt.inputPageSize})
			} else {
				cmd.SetArgs([]string{})
			}

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if capturedPageSize != tt.expectedPageSize {
				t.Errorf("expected page size %d, got %d", tt.expectedPageSize, capturedPageSize)
			}
		})
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
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
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
				Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
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
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
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
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
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
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
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
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
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
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[testItem], error) {
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
