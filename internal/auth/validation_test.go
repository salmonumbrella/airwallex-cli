package auth

import "testing"

func TestValidateAccountName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "myaccount", false},
		{"valid with dash", "my-account", false},
		{"valid with underscore", "my_account", false},
		{"valid with numbers", "account123", false},
		{"valid single char", "a", false},
		{"valid max length", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false}, // 64 'a's
		{"invalid empty", "", true},
		{"invalid special chars", "account!", true},
		{"invalid at sign", "account@home", true},
		{"invalid spaces", "my account", true},
		{"invalid too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true}, // 65 'a's
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccountName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAccountName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateClientID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "client123", false},
		{"valid with special", "cl1ent_id-test", false},
		{"valid max length", string(make([]byte, 128)), false},
		{"invalid empty", "", true},
		{"invalid too long", string(make([]byte, 129)), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClientID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClientID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "key123", false},
		{"valid long", "test_key_abcdef1234567890abcdef1234567890", false},
		{"valid max length", string(make([]byte, 256)), false},
		{"invalid empty", "", true},
		{"invalid too long", string(make([]byte, 257)), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidationTrimming(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		validator   func(string) error
		shouldError bool
	}{
		// Validators should NOT trim input - they validate as-is
		{"account name with leading space", " myaccount", ValidateAccountName, true},
		{"account name with trailing space", "myaccount ", ValidateAccountName, true},
		{"client ID with spaces", "client 123", ValidateClientID, false},
		{"api key with spaces", "key 123", ValidateAPIKey, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator(tt.input)
			if (err != nil) != tt.shouldError {
				t.Errorf("Validator with input %q error = %v, shouldError %v", tt.input, err, tt.shouldError)
			}
		})
	}
}
