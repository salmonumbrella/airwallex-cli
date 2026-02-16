package cmd

import (
	"strings"
	"testing"
)

func TestRegressionUserPath_BeneficiaryCreateNicknameAlias(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	root := NewRootCmd()
	root.SetArgs([]string{"ben", "cr", "--nn", "Clo Wang", "--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected --nn to parse on ben cr help, got: %v", err)
	}
}

func TestRegressionUserPath_TransferCreateAliases(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	root := NewRootCmd()
	root.SetArgs([]string{
		"tr", "cr",
		"--acc", "dm",
		"--bid", "3e1ff19b-b345-4dc8-96e8-2a3c5eba0e2e",
		"--ta", "1",
		"--tc", "CAD",
		"--sc", "CAD",
		"--reference", "Invoice 123",
		"--rsn", "payment_to_supplier",
		"--method", "LOCAL",
		"--dry-run",
	})

	err := root.Execute()
	if err == nil {
		return
	}
	if strings.Contains(err.Error(), `required flag(s) "beneficiary-id", "reason", "source-currency", "transfer-currency" not set`) {
		t.Fatalf("expected alias flags to satisfy required checks, got: %v", err)
	}
	if strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("expected user path aliases to parse, got: %v", err)
	}
	if !isExpectedTestError(err) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegressionUserPath_TransferConfirmationFileShortFlag(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	root := NewRootCmd()
	root.SetArgs([]string{
		"tr", "conf", "54d05727-ae79-4a1a-88e8-9a33de04bb4e",
		"--account", "dm",
		"-f", "/tmp/wire-confirmation-clo-wang.pdf",
		"--help",
	})

	if err := root.Execute(); err != nil {
		if strings.Contains(err.Error(), "unknown shorthand flag: 'f'") {
			t.Fatalf("expected -f shorthand to be accepted, got: %v", err)
		}
		t.Fatalf("unexpected error: %v", err)
	}
}
