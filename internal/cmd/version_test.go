package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		commit       string
		buildDate    string
		wantContains []string
	}{
		{
			name:      "with all version info",
			version:   "1.2.3",
			commit:    "abc123def",
			buildDate: "2024-01-15T10:30:00Z",
			wantContains: []string{
				"airwallex-cli 1.2.3",
				"commit:     abc123def",
				"build date: 2024-01-15T10:30:00Z",
			},
		},
		{
			name:      "with semantic version",
			version:   "v2.0.0-beta.1",
			commit:    "1234567890abcdef",
			buildDate: "2024-12-19T15:45:30Z",
			wantContains: []string{
				"airwallex-cli v2.0.0-beta.1",
				"commit:     1234567890abcdef",
				"build date: 2024-12-19T15:45:30Z",
			},
		},
		{
			name:      "with short commit hash",
			version:   "0.1.0",
			commit:    "abc123",
			buildDate: "2023-06-01T00:00:00Z",
			wantContains: []string{
				"airwallex-cli 0.1.0",
				"commit:     abc123",
				"build date: 2023-06-01T00:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original values
			origVersion := Version
			origCommit := Commit
			origBuildDate := BuildDate
			defer func() {
				Version = origVersion
				Commit = origCommit
				BuildDate = origBuildDate
			}()

			// Set test values
			Version = tt.version
			Commit = tt.commit
			BuildDate = tt.buildDate

			// Create version command
			cmd := newVersionCmd()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Verify output
			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected string %q\ngot:\n%s", want, output)
				}
			}
		})
	}
}

func TestVersionCommand_DefaultValues(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		commit       string
		buildDate    string
		wantContains []string
	}{
		{
			name:      "all default values",
			version:   "dev",
			commit:    "unknown",
			buildDate: "unknown",
			wantContains: []string{
				"airwallex-cli dev",
				"commit:     unknown",
				"build date: unknown",
			},
		},
		{
			name:      "empty version string",
			version:   "",
			commit:    "unknown",
			buildDate: "unknown",
			wantContains: []string{
				"airwallex-cli ",
				"commit:     unknown",
				"build date: unknown",
			},
		},
		{
			name:      "version set but commit and date unknown",
			version:   "1.0.0",
			commit:    "unknown",
			buildDate: "unknown",
			wantContains: []string{
				"airwallex-cli 1.0.0",
				"commit:     unknown",
				"build date: unknown",
			},
		},
		{
			name:      "commit set but version and date default",
			version:   "dev",
			commit:    "abc123",
			buildDate: "unknown",
			wantContains: []string{
				"airwallex-cli dev",
				"commit:     abc123",
				"build date: unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original values
			origVersion := Version
			origCommit := Commit
			origBuildDate := BuildDate
			defer func() {
				Version = origVersion
				Commit = origCommit
				BuildDate = origBuildDate
			}()

			// Set test values
			Version = tt.version
			Commit = tt.commit
			BuildDate = tt.buildDate

			// Create version command
			cmd := newVersionCmd()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Verify output
			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing expected string %q\ngot:\n%s", want, output)
				}
			}
		})
	}
}

func TestVersionCommand_Properties(t *testing.T) {
	cmd := newVersionCmd()

	// Test command properties
	if cmd.Use != "version" {
		t.Errorf("Use = %q, want %q", cmd.Use, "version")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil, expected a function")
	}
}

func TestVersionCommand_NoArgs(t *testing.T) {
	// Store original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	}()

	// Set test values
	Version = "test"
	Commit = "test-commit"
	BuildDate = "test-date"

	// Create a root command to test args validation
	rootCmd := &cobra.Command{Use: "root"}
	versionCmd := newVersionCmd()
	rootCmd.AddCommand(versionCmd)

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Execute with no args (should work)
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() with no args error = %v", err)
	}

	// Verify output was generated
	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty string")
	}
}

func TestVersionCommand_WithArgs(t *testing.T) {
	// Store original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	}()

	// Set test values
	Version = "test"
	Commit = "test-commit"
	BuildDate = "test-date"

	// Create a root command to test args validation
	rootCmd := &cobra.Command{Use: "root"}
	versionCmd := newVersionCmd()
	rootCmd.AddCommand(versionCmd)

	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Execute with args (version command accepts args, they're just ignored)
	rootCmd.SetArgs([]string{"version", "extra", "args"})
	err := rootCmd.Execute()

	// Version command doesn't validate args, so this should succeed
	// The args parameter is simply unused in the RunE function
	if err != nil {
		t.Errorf("Execute() with args error = %v", err)
	}
}

func TestVersionCommand_OutputFormat(t *testing.T) {
	// Store original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	}()

	// Set specific test values
	Version = "1.2.3"
	Commit = "abc123"
	BuildDate = "2024-01-15"

	// Create version command
	cmd := newVersionCmd()

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify output format
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Errorf("expected 3 lines of output, got %d:\n%s", len(lines), output)
	}

	// Verify first line starts with "airwallex-cli "
	if !strings.HasPrefix(lines[0], "airwallex-cli ") {
		t.Errorf("first line should start with 'airwallex-cli ', got: %q", lines[0])
	}

	// Verify second line has commit
	if !strings.Contains(lines[1], "commit:") {
		t.Errorf("second line should contain 'commit:', got: %q", lines[1])
	}

	// Verify third line has build date
	if !strings.Contains(lines[2], "build date:") {
		t.Errorf("third line should contain 'build date:', got: %q", lines[2])
	}
}

func TestVersionCommand_JSONOutput(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		buildDate string
	}{
		{
			name:      "standard version info",
			version:   "1.2.3",
			commit:    "abc123def",
			buildDate: "2024-01-15T10:30:00Z",
		},
		{
			name:      "dev version",
			version:   "dev",
			commit:    "unknown",
			buildDate: "unknown",
		},
		{
			name:      "semantic version with prerelease",
			version:   "v2.0.0-beta.1",
			commit:    "1234567890abcdef",
			buildDate: "2024-12-19T15:45:30Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original values
			origVersion := Version
			origCommit := Commit
			origBuildDate := BuildDate
			defer func() {
				Version = origVersion
				Commit = origCommit
				BuildDate = origBuildDate
			}()

			// Set test values
			Version = tt.version
			Commit = tt.commit
			BuildDate = tt.buildDate

			// Create version command with JSON context
			cmd := newVersionCmd()
			ctx := outfmt.WithFormat(context.Background(), "json")
			cmd.SetContext(ctx)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Parse JSON output
			var result versionInfo
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, buf.String())
			}

			// Verify JSON fields
			if result.Version != tt.version {
				t.Errorf("version = %q, want %q", result.Version, tt.version)
			}
			if result.Commit != tt.commit {
				t.Errorf("commit = %q, want %q", result.Commit, tt.commit)
			}
			if result.BuildDate != tt.buildDate {
				t.Errorf("build_date = %q, want %q", result.BuildDate, tt.buildDate)
			}
		})
	}
}

func TestVersionCommand_TextVsJSONOutput(t *testing.T) {
	// Store original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	}()

	// Set test values
	Version = "1.0.0"
	Commit = "abc123"
	BuildDate = "2024-01-01"

	t.Run("text output", func(t *testing.T) {
		cmd := newVersionCmd()
		ctx := outfmt.WithFormat(context.Background(), "text")
		cmd.SetContext(ctx)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		// Text output should contain human-readable format
		if !strings.Contains(output, "airwallex-cli 1.0.0") {
			t.Errorf("text output should contain 'airwallex-cli 1.0.0', got: %s", output)
		}
		if !strings.Contains(output, "commit:") {
			t.Errorf("text output should contain 'commit:', got: %s", output)
		}
	})

	t.Run("json output", func(t *testing.T) {
		cmd := newVersionCmd()
		ctx := outfmt.WithFormat(context.Background(), "json")
		cmd.SetContext(ctx)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		// JSON output should be valid JSON
		var result versionInfo
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		// JSON should not contain text formatting
		output := buf.String()
		if strings.Contains(output, "airwallex-cli 1.0.0") {
			t.Errorf("JSON output should not contain text format, got: %s", output)
		}
	})
}
