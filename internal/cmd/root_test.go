package cmd

import (
	"context"
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
			// Reset flags for each test
			flags = rootFlags{}

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
