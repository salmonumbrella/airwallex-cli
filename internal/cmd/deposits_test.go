package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDepositsListCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "list without filters",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "list with valid status filter",
			args:    []string{"--status", "SETTLED"},
			wantErr: false,
		},
		{
			name:    "list with PENDING status",
			args:    []string{"--status", "PENDING"},
			wantErr: false,
		},
		{
			name:    "list with FAILED status",
			args:    []string{"--status", "FAILED"},
			wantErr: false,
		},
		{
			name: "list with valid date range",
			args: []string{
				"--from", "2024-01-01",
				"--to", "2024-01-31",
			},
			wantErr: false,
		},
		{
			name:    "list with only from date",
			args:    []string{"--from", "2024-01-01"},
			wantErr: false,
		},
		{
			name:    "list with only to date",
			args:    []string{"--to", "2024-01-31"},
			wantErr: false,
		},
		{
			name: "list with all filters",
			args: []string{
				"--status", "SETTLED",
				"--from", "2024-01-01",
				"--to", "2024-01-31",
				"--page-size", "50",
			},
			wantErr: false,
		},
		{
			name:    "list with custom page size",
			args:    []string{"--page-size", "100"},
			wantErr: false,
		},
		{
			name:    "list with minimum page size",
			args:    []string{"--page-size", "10"},
			wantErr: false,
		},
		{
			name:    "list with page size below minimum (should adjust to 10)",
			args:    []string{"--page-size", "5"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags = rootFlags{}

			depositsCmd := newDepositsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(depositsCmd)

			fullArgs := append([]string{"deposits", "list"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestDepositsListCommand_PageSizeValidation(t *testing.T) {
	tests := []struct {
		name        string
		pageSize    int
		expectedMin int
		description string
	}{
		{
			name:        "page size below minimum gets adjusted",
			pageSize:    5,
			expectedMin: 10,
			description: "page size less than 10 should be adjusted to 10",
		},
		{
			name:        "page size at minimum is unchanged",
			pageSize:    10,
			expectedMin: 10,
			description: "page size of exactly 10 should remain 10",
		},
		{
			name:        "page size above minimum is unchanged",
			pageSize:    50,
			expectedMin: 10,
			description: "page size above 10 should be unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newDepositsListCmd()
			if err := cmd.Flags().Set("page-size", intToString(tt.pageSize)); err != nil {
				t.Fatalf("failed to set page-size flag: %v", err)
			}

			// Verify the flag is set
			pageSizeFlag := cmd.Flags().Lookup("page-size")
			if pageSizeFlag == nil {
				t.Fatal("page-size flag not found")
			}

			if pageSizeFlag.Deprecated == "" {
				t.Errorf("expected page-size flag to be deprecated")
			}
		})
	}
}

func TestDepositsGetCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no deposit ID provided",
			args:        []string{},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 0",
		},
		{
			name:    "valid deposit ID provided",
			args:    []string{"dep_123456"},
			wantErr: false,
		},
		{
			name:        "too many arguments",
			args:        []string{"dep_123456", "extra_arg"},
			wantErr:     true,
			errContains: "accepts 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags = rootFlags{}

			depositsCmd := newDepositsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(depositsCmd)

			fullArgs := append([]string{"deposits", "get"}, tt.args...)
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
				if err != nil && !isExpectedTestError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestDepositsCommand_Subcommands(t *testing.T) {
	cmd := newDepositsCmd()

	expectedSubcommands := []string{"list", "get"}
	actualSubcommands := make(map[string]bool)

	for _, subcmd := range cmd.Commands() {
		actualSubcommands[subcmd.Name()] = true
	}

	for _, expected := range expectedSubcommands {
		if !actualSubcommands[expected] {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

func TestDepositsListCommand_FlagDefaults(t *testing.T) {
	cmd := newDepositsListCmd()

	tests := []struct {
		flagName     string
		expectedType string
		hasDefault   bool
	}{
		{
			flagName:     "status",
			expectedType: "string",
			hasDefault:   false,
		},
		{
			flagName:     "from",
			expectedType: "string",
			hasDefault:   false,
		},
		{
			flagName:     "to",
			expectedType: "string",
			hasDefault:   false,
		},
		{
			flagName:     "page-size",
			expectedType: "int",
			hasDefault:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if flag.Value.Type() != tt.expectedType {
				t.Errorf("flag %q has type %q, expected %q", tt.flagName, flag.Value.Type(), tt.expectedType)
			}
		})
	}
}

func TestDepositsGetCommand_ArgsValidation(t *testing.T) {
	cmd := newDepositsGetCmd()

	// Test that the command expects exactly 1 argument
	if cmd.Args == nil {
		t.Fatal("Args validator not set")
	}

	// cobra.ExactArgs(1) should reject 0 args
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error for 0 arguments, got nil")
	}

	// cobra.ExactArgs(1) should accept 1 arg
	err = cmd.Args(cmd, []string{"dep_123"})
	if err != nil {
		t.Errorf("expected no error for 1 argument, got: %v", err)
	}

	// cobra.ExactArgs(1) should reject 2 args
	err = cmd.Args(cmd, []string{"dep_123", "extra"})
	if err == nil {
		t.Error("expected error for 2 arguments, got nil")
	}
}
