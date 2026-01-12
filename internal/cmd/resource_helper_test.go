package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

type testResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestNewGetCommand_JSONOutput(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	customIO := &iocontext.IO{
		Out:    &outBuf,
		ErrOut: &errBuf,
		In:     strings.NewReader(""),
	}

	ctx := iocontext.WithIO(outfmt.WithFormat(context.Background(), "json"), customIO)

	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			return &testResource{ID: id, Name: "Test Resource"}, nil
		},
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"res_123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, `"id"`) || !strings.Contains(output, "res_123") {
		t.Errorf("expected JSON output with id, got %q", output)
	}
}

func TestNewGetCommand_TextOutputWithTextOutput(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	customIO := &iocontext.IO{
		Out:    &outBuf,
		ErrOut: &errBuf,
		In:     strings.NewReader(""),
	}

	ctx := iocontext.WithIO(outfmt.WithFormat(context.Background(), "text"), customIO)

	var textOutputCalled bool
	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			return &testResource{ID: id, Name: "Test Resource"}, nil
		},
		TextOutput: func(cmd *cobra.Command, item *testResource) error {
			textOutputCalled = true
			return outfmt.WriteKV(cmd.OutOrStdout(), []outfmt.KV{
				{Key: "id", Value: item.ID},
				{Key: "name", Value: item.Name},
			})
		},
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"res_123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if !textOutputCalled {
		t.Error("expected TextOutput to be called")
	}
}

func TestNewGetCommand_TextOutputWithoutTextOutput(t *testing.T) {
	ctx := outfmt.WithFormat(context.Background(), "text")

	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			return &testResource{ID: id, Name: "Test Resource"}, nil
		},
		// TextOutput is nil
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"res_123"})

	// Should succeed even without TextOutput
	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
}

func TestNewGetCommand_FetchError(t *testing.T) {
	expectedErr := errors.New("fetch failed")

	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			return nil, expectedErr
		},
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"res_123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewGetCommand_ClientError(t *testing.T) {
	expectedErr := errors.New("client creation failed")

	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			return &testResource{ID: id, Name: "Test"}, nil
		},
	}, func(context.Context) (*api.Client, error) {
		return nil, expectedErr
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"res_123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewGetCommand_RequiresExactlyOneArg(t *testing.T) {
	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			return &testResource{ID: id, Name: "Test"}, nil
		},
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"res_123"}, wantErr: false},
		{name: "two args", args: []string{"res_123", "extra"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCmd := NewGetCommand(GetConfig[*testResource]{
				Use:   "get <id>",
				Short: "Get resource",
				Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
					return &testResource{ID: id, Name: "Test"}, nil
				},
			}, func(context.Context) (*api.Client, error) {
				return &api.Client{}, nil
			})

			testCmd.SetContext(ctx)
			testCmd.SetArgs(tt.args)

			err := testCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewGetCommand_PassesCorrectID(t *testing.T) {
	var capturedID string

	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			capturedID = id
			return &testResource{ID: id, Name: "Test"}, nil
		},
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"my_resource_id_456"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if capturedID != "my_resource_id_456" {
		t.Errorf("expected ID 'my_resource_id_456', got %q", capturedID)
	}
}

func TestNewGetCommand_TextOutputError(t *testing.T) {
	expectedErr := errors.New("text output failed")

	cmd := NewGetCommand(GetConfig[*testResource]{
		Use:   "get <id>",
		Short: "Get resource",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*testResource, error) {
			return &testResource{ID: id, Name: "Test"}, nil
		},
		TextOutput: func(cmd *cobra.Command, item *testResource) error {
			return expectedErr
		},
	}, func(context.Context) (*api.Client, error) {
		return &api.Client{}, nil
	})

	ctx := outfmt.WithFormat(context.Background(), "text")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"res_123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
