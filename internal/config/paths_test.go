package config

import "testing"

func TestAppName(t *testing.T) {
	if AppName != "airwallex-cli" {
		t.Errorf("AppName = %q, want 'airwallex-cli'", AppName)
	}
}
