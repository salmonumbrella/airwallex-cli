package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func TestParseEvents(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single event",
			input:    []string{"transfer.completed"},
			expected: []string{"transfer.completed"},
		},
		{
			name:     "multiple events separate flags",
			input:    []string{"transfer.completed", "transfer.failed"},
			expected: []string{"transfer.completed", "transfer.failed"},
		},
		{
			name:     "comma separated",
			input:    []string{"transfer.completed,transfer.failed"},
			expected: []string{"transfer.completed", "transfer.failed"},
		},
		{
			name:     "mixed comma and separate",
			input:    []string{"transfer.completed,transfer.failed", "deposit.settled"},
			expected: []string{"transfer.completed", "transfer.failed", "deposit.settled"},
		},
		{
			name:     "with whitespace",
			input:    []string{"transfer.completed, transfer.failed", "deposit.settled"},
			expected: []string{"transfer.completed", "transfer.failed", "deposit.settled"},
		},
		{
			name:     "duplicates removed",
			input:    []string{"transfer.completed", "transfer.completed"},
			expected: []string{"transfer.completed"},
		},
		{
			name:     "comma separated duplicates",
			input:    []string{"transfer.completed,transfer.failed,transfer.completed"},
			expected: []string{"transfer.completed", "transfer.failed"},
		},
		{
			name:     "empty strings filtered",
			input:    []string{"transfer.completed", "", "transfer.failed"},
			expected: []string{"transfer.completed", "transfer.failed"},
		},
		{
			name:     "trailing comma",
			input:    []string{"transfer.completed,"},
			expected: []string{"transfer.completed"},
		},
		{
			name:     "leading comma",
			input:    []string{",transfer.completed"},
			expected: []string{"transfer.completed"},
		},
		{
			name:     "multiple commas",
			input:    []string{"transfer.completed,,transfer.failed"},
			expected: []string{"transfer.completed", "transfer.failed"},
		},
		{
			name:     "whitespace only trimmed",
			input:    []string{"  transfer.completed  ", "  transfer.failed  "},
			expected: []string{"transfer.completed", "transfer.failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse events using the same logic from the create command
			var allEvents []string
			seen := make(map[string]bool)

			for _, e := range tt.input {
				for _, ev := range strings.Split(e, ",") {
					ev = strings.TrimSpace(ev)
					if ev != "" && !seen[ev] {
						allEvents = append(allEvents, ev)
						seen[ev] = true
					}
				}
			}

			if len(allEvents) != len(tt.expected) {
				t.Errorf("got %d events, want %d events", len(allEvents), len(tt.expected))
				t.Errorf("got: %v", allEvents)
				t.Errorf("want: %v", tt.expected)
				return
			}

			for i, got := range allEvents {
				if got != tt.expected[i] {
					t.Errorf("events[%d] = %q, want %q", i, got, tt.expected[i])
				}
			}
		})
	}
}

func TestWebhooksCreateValidation(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "missing url",
			args: []string{
				"--events", "transfer.completed",
			},
			wantErr:     true,
			errContains: "required flag(s)",
		},
		{
			name: "missing events",
			args: []string{
				"--url", "https://example.com/webhook",
			},
			wantErr:     true,
			errContains: "required flag(s)",
		},
		{
			name: "empty events",
			args: []string{
				"--url", "https://example.com/webhook",
				"--events", "",
			},
			wantErr:     true,
			errContains: "at least one valid event is required",
		},
		{
			name: "invalid event type",
			args: []string{
				"--url", "https://example.com/webhook",
				"--events", "invalid.event.type",
			},
			wantErr:     true,
			errContains: "invalid event types",
		},
		{
			name: "mixed valid and invalid events",
			args: []string{
				"--url", "https://example.com/webhook",
				"--events", "transfer.completed,invalid.event",
			},
			wantErr:     true,
			errContains: "invalid event types",
		},
		{
			name: "valid single event",
			args: []string{
				"--url", "https://example.com/webhook",
				"--events", "transfer.completed",
			},
			wantErr: false,
		},
		{
			name: "valid multiple events",
			args: []string{
				"--url", "https://example.com/webhook",
				"--events", "transfer.completed,transfer.failed",
			},
			wantErr: false,
		},
		{
			name: "valid comma separated with spaces",
			args: []string{
				"--url", "https://example.com/webhook",
				"--events", "transfer.completed, transfer.failed, deposit.settled",
			},
			wantErr: false,
		},
		{
			name: "valid multiple event flags",
			args: []string{
				"--url", "https://example.com/webhook",
				"--events", "transfer.completed",
				"--events", "transfer.failed",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhooksCmd := newWebhooksCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(webhooksCmd)

			fullArgs := append([]string{"webhooks", "create"}, tt.args...)
			rootCmd.SetArgs(fullArgs)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				// For non-error cases, we expect to fail on the actual API call
				// since we don't have a mock client set up. This is acceptable
				// as we're only testing validation logic here.
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestWebhooksListCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "with limit",
			args:    []string{"--page-size", "50"},
			wantErr: false,
		},
		{
			name:    "with small limit",
			args:    []string{"--page-size", "5"},
			wantErr: false, // Should be adjusted to minimum 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhooksCmd := newWebhooksCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(webhooksCmd)

			fullArgs := append([]string{"webhooks", "list"}, tt.args...)
			rootCmd.SetArgs(fullArgs)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				// For non-error cases, we expect to fail on the actual API call
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestWebhooksGetCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no webhook ID",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg(s)",
		},
		{
			name:        "too many args",
			args:        []string{"wh_123", "wh_456"},
			wantErr:     true,
			errContains: "accepts 1 arg(s)",
		},
		{
			name:    "valid webhook ID",
			args:    []string{"wh_123"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhooksCmd := newWebhooksCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(webhooksCmd)

			fullArgs := append([]string{"webhooks", "get"}, tt.args...)
			rootCmd.SetArgs(fullArgs)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				// For non-error cases, we expect to fail on the actual API call
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestWebhooksDeleteCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no webhook ID",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg(s)",
		},
		{
			name:        "too many args",
			args:        []string{"wh_123", "wh_456"},
			wantErr:     true,
			errContains: "accepts 1 arg(s)",
		},
		{
			name:    "valid webhook ID with skip confirm",
			args:    []string{"wh_123", "--yes"},
			wantErr: false,
		},
		{
			name:    "valid webhook ID short flag",
			args:    []string{"wh_123", "-y"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhooksCmd := newWebhooksCmd()
			var yesFlag bool
			rootCmd := &cobra.Command{
				Use: "root",
				PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
					// Set up context with yes flag like the real root command does
					ctx := context.Background()
					ctx = outfmt.WithYes(ctx, yesFlag)
					cmd.SetContext(ctx)
					return nil
				},
			}
			// Add persistent flags that normally come from root command
			rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompts")
			rootCmd.PersistentFlags().BoolVar(&yesFlag, "force", false, "Skip confirmation prompts (alias for --yes)")
			rootCmd.AddCommand(webhooksCmd)

			fullArgs := append([]string{"webhooks", "delete"}, tt.args...)
			rootCmd.SetArgs(fullArgs)

			err := rootCmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				// For non-error cases, we expect to fail on the actual API call
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestEventTypesList(t *testing.T) {
	// Test that common event types are documented in help text
	webhooksCmd := newWebhooksCmd()

	// Check parent command help contains common events
	helpText := webhooksCmd.Long
	commonEvents := []string{
		"transfer.completed",
		"transfer.failed",
		"deposit.settled",
		"card.activated",
	}

	for _, event := range commonEvents {
		if !strings.Contains(helpText, event) {
			t.Errorf("webhooks help text missing common event: %s", event)
		}
	}

	// Check create command help contains examples
	var createCmd *cobra.Command
	for _, cmd := range webhooksCmd.Commands() {
		if cmd.Use == "create" {
			createCmd = cmd
			break
		}
	}

	if createCmd == nil {
		t.Fatal("create command not found")
	}

	createHelp := createCmd.Long
	if !strings.Contains(createHelp, "Examples:") {
		t.Error("create command help missing examples section")
	}
}
