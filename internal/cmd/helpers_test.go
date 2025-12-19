package cmd

import (
	"strings"
	"testing"
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
