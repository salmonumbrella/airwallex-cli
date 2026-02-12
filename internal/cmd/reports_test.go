package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

func TestReportsSettlementValidation(t *testing.T) {
	t.Setenv("AWX_ACCOUNT", "test-account")

	originalOpenSecretsStore := openSecretsStore
	defer func() { openSecretsStore = originalOpenSecretsStore }()
	openSecretsStore = func() (secrets.Store, error) {
		return &mockStore{}, nil
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid format",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
				"--format", "PDF",
			},
			wantErr:     true,
			errContains: "--format must be CSV or EXCEL",
		},
		{
			name: "missing from-date",
			args: []string{
				"--to-date", "2024-01-31",
			},
			wantErr:     true,
			errContains: "required flag",
		},
		{
			name: "missing to-date",
			args: []string{
				"--from-date", "2024-01-01",
			},
			wantErr:     true,
			errContains: "required flag",
		},
		{
			name: "valid CSV format",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
				"--format", "CSV",
			},
			wantErr: false,
		},
		{
			name: "valid EXCEL format",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
				"--format", "EXCEL",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportsCmd := newReportsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(reportsCmd)

			fullArgs := append([]string{"reports", "settlement"}, tt.args...)
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
				if err != nil && !isExpectedAPIError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestReportsBalanceActivityValidation(t *testing.T) {
	t.Setenv("AWX_ACCOUNT", "test-account")

	originalOpenSecretsStore := openSecretsStore
	defer func() { openSecretsStore = originalOpenSecretsStore }()
	openSecretsStore = func() (secrets.Store, error) {
		return &mockStore{}, nil
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid format",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
				"--format", "XML",
			},
			wantErr:     true,
			errContains: "--format must be CSV, EXCEL, or PDF",
		},
		{
			name: "valid PDF format",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
				"--format", "PDF",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportsCmd := newReportsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(reportsCmd)

			fullArgs := append([]string{"reports", "balance-activity"}, tt.args...)
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
				if err != nil && !isExpectedAPIError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestReportsTransactionReconValidation(t *testing.T) {
	t.Setenv("AWX_ACCOUNT", "test-account")

	originalOpenSecretsStore := openSecretsStore
	defer func() { openSecretsStore = originalOpenSecretsStore }()
	openSecretsStore = func() (secrets.Store, error) {
		return &mockStore{}, nil
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "PDF not supported",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
				"--format", "PDF",
			},
			wantErr:     true,
			errContains: "PDF not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportsCmd := newReportsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(reportsCmd)

			fullArgs := append([]string{"reports", "transaction-recon"}, tt.args...)
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
				if err != nil && !isExpectedAPIError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestReportsAccountStatementValidation(t *testing.T) {
	t.Setenv("AWX_ACCOUNT", "test-account")

	originalOpenSecretsStore := openSecretsStore
	defer func() { openSecretsStore = originalOpenSecretsStore }()
	openSecretsStore = func() (secrets.Store, error) {
		return &mockStore{}, nil
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "missing currencies",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
			},
			wantErr:     true,
			errContains: "required flag",
		},
		{
			name: "valid with currencies",
			args: []string{
				"--from-date", "2024-01-01",
				"--to-date", "2024-01-31",
				"--currencies", "CAD,USD",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportsCmd := newReportsCmd()
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(reportsCmd)

			fullArgs := append([]string{"reports", "account-statement"}, tt.args...)
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
				if err != nil && !isExpectedAPIError(err) {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}
