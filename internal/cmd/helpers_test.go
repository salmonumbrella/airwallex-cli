package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func TestConvertDateToRFC3339(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid date",
			input:   "2024-01-15",
			want:    "2024-01-15T00:00:00Z",
			wantErr: false,
		},
		{
			name:    "valid date end of month",
			input:   "2024-12-31",
			want:    "2024-12-31T00:00:00Z",
			wantErr: false,
		},
		{
			name:    "invalid format - wrong separator",
			input:   "2024/01/15",
			wantErr: true,
			errMsg:  "expected format YYYY-MM-DD",
		},
		{
			name:    "invalid format - no separators",
			input:   "20240115",
			wantErr: true,
			errMsg:  "expected format YYYY-MM-DD",
		},
		{
			name:    "invalid format - too short",
			input:   "2024-1-5",
			wantErr: true,
			errMsg:  "expected format YYYY-MM-DD",
		},
		{
			name:    "invalid date - month 13",
			input:   "2024-13-01",
			wantErr: true,
			errMsg:  "expected format YYYY-MM-DD",
		},
		{
			name:    "invalid date - day 32",
			input:   "2024-01-32",
			wantErr: true,
			errMsg:  "expected format YYYY-MM-DD",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "expected format YYYY-MM-DD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertDateToRFC3339(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertDateToRFC3339End(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid date",
			input:   "2024-01-15",
			want:    "2024-01-15T23:59:59Z",
			wantErr: false,
		},
		{
			name:    "valid date end of month",
			input:   "2024-12-31",
			want:    "2024-12-31T23:59:59Z",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "2024/01/15",
			wantErr: true,
			errMsg:  "expected format YYYY-MM-DD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertDateToRFC3339End(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfirmOrYes(t *testing.T) {
	// Save original isTerminal and restore after tests
	origIsTerminal := isTerminal
	defer func() { isTerminal = origIsTerminal }()

	tests := []struct {
		name      string
		yesFlagOn bool
		jsonMode  bool
		isTTY     bool
		input     string
		want      bool
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "yes flag set skips prompt",
			yesFlagOn: true,
			isTTY:     false,
			input:     "",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "json mode skips prompt",
			yesFlagOn: false,
			jsonMode:  true,
			isTTY:     false,
			input:     "",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "non-TTY without yes flag returns error",
			yesFlagOn: false,
			isTTY:     false,
			input:     "",
			want:      false,
			wantErr:   true,
			errMsg:    "stdin is not a terminal",
		},
		{
			name:      "TTY with y response confirms",
			yesFlagOn: false,
			isTTY:     true,
			input:     "y\n",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "TTY with Y response confirms",
			yesFlagOn: false,
			isTTY:     true,
			input:     "Y\n",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "TTY with yes response confirms",
			yesFlagOn: false,
			isTTY:     true,
			input:     "yes\n",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "TTY with Yes response confirms",
			yesFlagOn: false,
			isTTY:     true,
			input:     "Yes\n",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "TTY with YES response confirms",
			yesFlagOn: false,
			isTTY:     true,
			input:     "YES\n",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "TTY with n response declines",
			yesFlagOn: false,
			isTTY:     true,
			input:     "n\n",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "TTY with no response declines",
			yesFlagOn: false,
			isTTY:     true,
			input:     "no\n",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "TTY with empty response declines",
			yesFlagOn: false,
			isTTY:     true,
			input:     "\n",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "TTY with random text declines",
			yesFlagOn: false,
			isTTY:     true,
			input:     "maybe\n",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "TTY with whitespace-padded y confirms",
			yesFlagOn: false,
			isTTY:     true,
			input:     "  y  \n",
			want:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock TTY check
			isTerminal = func() bool { return tt.isTTY }

			// Build context with appropriate flags
			ctx := context.Background()
			if tt.yesFlagOn {
				ctx = outfmt.WithYes(ctx, true)
			}
			if tt.jsonMode {
				ctx = outfmt.WithFormat(ctx, "json")
			}

			// Set up IO with test input
			stdin := strings.NewReader(tt.input)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			io := &iocontext.IO{
				In:     stdin,
				Out:    stdout,
				ErrOut: stderr,
			}
			ctx = iocontext.WithIO(ctx, io)

			got, err := ConfirmOrYes(ctx, "Test prompt?")

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfirmOrYes_YesFlagSkipsPrompt(t *testing.T) {
	ctx := context.Background()
	ctx = outfmt.WithYes(ctx, true)

	got, err := ConfirmOrYes(ctx, "Delete everything?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true when --yes flag is set")
	}
}

func TestConfirmOrYes_JSONModeSkipsPrompt(t *testing.T) {
	ctx := context.Background()
	ctx = outfmt.WithFormat(ctx, "json")

	got, err := ConfirmOrYes(ctx, "Delete everything?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true when JSON mode is enabled")
	}
}

func TestConfirmOrYes_NonTTYReturnsError(t *testing.T) {
	// Save and restore original
	origIsTerminal := isTerminal
	defer func() { isTerminal = origIsTerminal }()

	// Mock non-TTY
	isTerminal = func() bool { return false }

	ctx := context.Background()

	_, err := ConfirmOrYes(ctx, "Delete everything?")
	if err == nil {
		t.Fatal("expected error for non-TTY stdin")
	}
	if !strings.Contains(err.Error(), "stdin is not a terminal") {
		t.Errorf("expected error about non-terminal, got: %v", err)
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected error to mention --yes flag, got: %v", err)
	}
}

func TestConfirmOrYes_PromptsToStderr(t *testing.T) {
	// Save and restore original
	origIsTerminal := isTerminal
	defer func() { isTerminal = origIsTerminal }()

	// Mock TTY
	isTerminal = func() bool { return true }

	ctx := context.Background()
	stdin := strings.NewReader("y\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	io := &iocontext.IO{
		In:     stdin,
		Out:    stdout,
		ErrOut: stderr,
	}
	ctx = iocontext.WithIO(ctx, io)

	_, err := ConfirmOrYes(ctx, "Delete everything?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify prompt went to stderr, not stdout
	if stdout.Len() > 0 {
		t.Errorf("expected no output to stdout, got: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Delete everything?") {
		t.Errorf("expected stderr to contain prompt, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "[y/N]") {
		t.Errorf("expected stderr to contain [y/N], got: %s", stderr.String())
	}
}
