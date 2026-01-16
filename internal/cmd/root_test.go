package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func TestRootCmd_ContextInjection(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantOutputFmt string
		wantColorMode string
	}{
		{
			name:          "default values",
			args:          []string{},
			wantOutputFmt: "text",
			wantColorMode: "auto",
		},
		{
			name:          "json output format",
			args:          []string{"--output", "json"},
			wantOutputFmt: "json",
			wantColorMode: "auto",
		},
		{
			name:          "never color mode",
			args:          []string{"--color", "never"},
			wantOutputFmt: "text",
			wantColorMode: "never",
		},
		{
			name:          "both flags set",
			args:          []string{"--output", "json", "--color", "always"},
			wantOutputFmt: "json",
			wantColorMode: "always",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedCtx context.Context

			// Create root command with a test subcommand that captures context
			cmd := NewRootCmd()
			testCmd := &cobra.Command{
				Use: "test",
				RunE: func(cmd *cobra.Command, args []string) error {
					capturedCtx = cmd.Context()
					return nil
				},
			}
			cmd.AddCommand(testCmd)

			// Set args and execute
			fullArgs := append(tt.args, "test")
			cmd.SetArgs(fullArgs)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Verify output format in context
			gotFormat := outfmt.GetFormat(capturedCtx)
			if gotFormat != tt.wantOutputFmt {
				t.Errorf("output format = %v, want %v", gotFormat, tt.wantOutputFmt)
			}

			// Verify UI is in context (not just default)
			u := ui.FromContext(capturedCtx)
			if u == nil {
				t.Error("UI not found in context")
			}
		})
	}
}

func TestRootCmd_AgentFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantYes    bool
		wantLimit  int
		wantSortBy string
		wantDesc   bool
	}{
		{
			name:       "default values",
			args:       []string{},
			wantYes:    false,
			wantLimit:  0,
			wantSortBy: "",
			wantDesc:   false,
		},
		{
			name:       "yes flag",
			args:       []string{"--yes"},
			wantYes:    true,
			wantLimit:  0,
			wantSortBy: "",
			wantDesc:   false,
		},
		{
			name:       "yes flag short form",
			args:       []string{"-y"},
			wantYes:    true,
			wantLimit:  0,
			wantSortBy: "",
			wantDesc:   false,
		},
		{
			name:       "force flag (alias for yes)",
			args:       []string{"--force"},
			wantYes:    true,
			wantLimit:  0,
			wantSortBy: "",
			wantDesc:   false,
		},
		{
			name:       "limit flag",
			args:       []string{"--output-limit", "10"},
			wantYes:    false,
			wantLimit:  10,
			wantSortBy: "",
			wantDesc:   false,
		},
		{
			name:       "sort-by flag",
			args:       []string{"--sort-by", "created_at"},
			wantYes:    false,
			wantLimit:  0,
			wantSortBy: "created_at",
			wantDesc:   false,
		},
		{
			name:       "sort-by with desc",
			args:       []string{"--sort-by", "created_at", "--desc"},
			wantYes:    false,
			wantLimit:  0,
			wantSortBy: "created_at",
			wantDesc:   true,
		},
		{
			name:       "all agent flags combined",
			args:       []string{"--yes", "--output-limit", "5", "--sort-by", "amount", "--desc"},
			wantYes:    true,
			wantLimit:  5,
			wantSortBy: "amount",
			wantDesc:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedCtx context.Context

			// Create root command with a test subcommand that captures context
			cmd := NewRootCmd()
			testCmd := &cobra.Command{
				Use: "test",
				RunE: func(cmd *cobra.Command, args []string) error {
					capturedCtx = cmd.Context()
					return nil
				},
			}
			cmd.AddCommand(testCmd)

			// Set args and execute
			fullArgs := append(tt.args, "test")
			cmd.SetArgs(fullArgs)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Verify agent flags in context
			if got := outfmt.GetYes(capturedCtx); got != tt.wantYes {
				t.Errorf("yes = %v, want %v", got, tt.wantYes)
			}
			if got := outfmt.GetLimit(capturedCtx); got != tt.wantLimit {
				t.Errorf("limit = %v, want %v", got, tt.wantLimit)
			}
			if got := outfmt.GetSortBy(capturedCtx); got != tt.wantSortBy {
				t.Errorf("sortBy = %v, want %v", got, tt.wantSortBy)
			}
			if got := outfmt.GetDesc(capturedCtx); got != tt.wantDesc {
				t.Errorf("desc = %v, want %v", got, tt.wantDesc)
			}
		})
	}
}

func TestRootCmd_DescRequiresSortBy(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "desc without sort-by fails",
			args:    []string{"--desc"},
			wantErr: "--desc requires --sort-by to be specified",
		},
		{
			name:    "desc with sort-by succeeds",
			args:    []string{"--sort-by", "created_at", "--desc"},
			wantErr: "",
		},
		{
			name:    "sort-by without desc succeeds",
			args:    []string{"--sort-by", "created_at"},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command with a test subcommand
			cmd := NewRootCmd()
			testCmd := &cobra.Command{
				Use:  "test",
				RunE: func(cmd *cobra.Command, args []string) error { return nil },
			}
			cmd.AddCommand(testCmd)

			// Set args and execute
			fullArgs := append(tt.args, "test")
			cmd.SetArgs(fullArgs)

			err := cmd.Execute()

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Execute() unexpected error = %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Execute() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestRootCmd_QueryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "query.jq")
	if err := os.WriteFile(path, []byte(".items[] | .id"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var capturedCtx context.Context

	cmd := NewRootCmd()
	testCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			capturedCtx = cmd.Context()
			return nil
		},
	}
	cmd.AddCommand(testCmd)
	cmd.SetArgs([]string{"--query-file", path, "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := outfmt.GetQuery(capturedCtx)
	if got != ".items[] | .id" {
		t.Errorf("query = %q, want %q", got, ".items[] | .id")
	}
}

func TestRootCmd_QueryFileConflict(t *testing.T) {
	cmd := NewRootCmd()
	testCmd := &cobra.Command{
		Use:  "test",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	cmd.AddCommand(testCmd)
	cmd.SetArgs([]string{"--query", ".items[]", "--query-file", "query.jq", "test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "use only one of --query or --query-file") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "use only one of --query or --query-file")
	}
}
