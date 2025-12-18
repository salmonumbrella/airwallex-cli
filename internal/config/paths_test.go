package config

import (
	"strings"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error: %v", err)
	}
	if !strings.HasSuffix(dir, "airwallex-cli") {
		t.Errorf("ConfigDir() = %q, want suffix 'airwallex-cli'", dir)
	}
}

func TestAppName(t *testing.T) {
	if AppName != "airwallex-cli" {
		t.Errorf("AppName = %q, want 'airwallex-cli'", AppName)
	}
}
