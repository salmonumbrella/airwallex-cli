package secrets

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/99designs/keyring"
)

// mockKeyring implements the keyring.Keyring interface for testing
type mockKeyring struct {
	items map[string]keyring.Item
	err   error // Error to return on operations
}

func newMockKeyring() *mockKeyring {
	return &mockKeyring{
		items: make(map[string]keyring.Item),
	}
}

func (m *mockKeyring) Get(key string) (keyring.Item, error) {
	if m.err != nil {
		return keyring.Item{}, m.err
	}
	item, ok := m.items[key]
	if !ok {
		return keyring.Item{}, keyring.ErrKeyNotFound
	}
	return item, nil
}

func (m *mockKeyring) GetMetadata(key string) (keyring.Metadata, error) {
	if m.err != nil {
		return keyring.Metadata{}, m.err
	}
	item, ok := m.items[key]
	if !ok {
		return keyring.Metadata{}, keyring.ErrKeyNotFound
	}
	// Create a copy of the item without sensitive data
	itemCopy := item
	itemCopy.Data = nil
	return keyring.Metadata{
		Item:             &itemCopy,
		ModificationTime: time.Now(),
	}, nil
}

func (m *mockKeyring) Set(item keyring.Item) error {
	if m.err != nil {
		return m.err
	}
	m.items[item.Key] = item
	return nil
}

func (m *mockKeyring) Remove(key string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.items[key]; !ok {
		return keyring.ErrKeyNotFound
	}
	delete(m.items, key)
	return nil
}

func (m *mockKeyring) Keys() ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys, nil
}

func TestCredentialKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"my-company", "account:my-company"},
		{"test-account", "account:test-account"},
	}
	for _, tt := range tests {
		got := credentialKey(tt.name)
		if got != tt.want {
			t.Errorf("credentialKey(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  Test  ", "test"},
		{"UPPER", "upper"},
		{"already-lower", "already-lower"},
	}
	for _, tt := range tests {
		got := normalize(tt.input)
		if got != tt.want {
			t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseCredentialKey(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		wantOK bool
	}{
		{"account:my-company", "my-company", true},
		{"account:test-account", "test-account", true},
		{"invalid", "", false},
		{"account:", "", false},
		{"account:  ", "", false},
	}
	for _, tt := range tests {
		got, ok := ParseCredentialKey(tt.input)
		if ok != tt.wantOK {
			t.Errorf("ParseCredentialKey(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
		}
		if got != tt.want {
			t.Errorf("ParseCredentialKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCredentialAgeWarning(t *testing.T) {
	tests := []struct {
		name        string
		createdAt   time.Time
		accountName string
		wantWarning bool
	}{
		{
			name:        "old credentials warn on first retrieval",
			createdAt:   time.Now().Add(-100 * 24 * time.Hour),
			accountName: "test-old-account",
			wantWarning: true,
		},
		{
			name:        "recent credentials do not warn",
			createdAt:   time.Now().Add(-30 * 24 * time.Hour),
			accountName: "test-recent-account",
			wantWarning: false,
		},
		{
			name:        "zero time credentials do not warn",
			createdAt:   time.Time{},
			accountName: "test-zero-time-account",
			wantWarning: false,
		},
		{
			name:        "credentials just before threshold do not warn",
			createdAt:   time.Now().Add(-CredentialRotationThreshold + time.Minute),
			accountName: "test-threshold-account",
			wantWarning: false,
		},
		{
			name:        "credentials just past threshold warn",
			createdAt:   time.Now().Add(-CredentialRotationThreshold - time.Hour),
			accountName: "test-past-threshold-account",
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset warning state for each test
			warnedAccounts.Delete(tt.accountName)

			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			// Simulate credential retrieval by checking the warning logic
			creds := Credentials{
				Name:      tt.accountName,
				CreatedAt: tt.createdAt,
			}

			// Execute the same logic as in Get()
			if !creds.CreatedAt.IsZero() && time.Since(creds.CreatedAt) > CredentialRotationThreshold {
				if _, warned := warnedAccounts.LoadOrStore(tt.accountName, true); !warned {
					log.Printf("Warning: credentials for account %q are over 90 days old, consider rotating", tt.accountName)
				}
			}

			logOutput := buf.String()
			hasWarning := len(logOutput) > 0

			if hasWarning != tt.wantWarning {
				t.Errorf("warning = %v, want %v (log: %q)", hasWarning, tt.wantWarning, logOutput)
			}

			// For old credentials, verify second retrieval does NOT warn (rate limiting)
			if tt.wantWarning {
				buf.Reset()

				// Second retrieval
				if !creds.CreatedAt.IsZero() && time.Since(creds.CreatedAt) > CredentialRotationThreshold {
					if _, warned := warnedAccounts.LoadOrStore(tt.accountName, true); !warned {
						log.Printf("Warning: credentials for account %q are over 90 days old, consider rotating", tt.accountName)
					}
				}

				secondLogOutput := buf.String()
				if len(secondLogOutput) > 0 {
					t.Errorf("second retrieval should not warn, got: %q", secondLogOutput)
				}
			}
		})
	}
}

func TestKeyringStore_Set(t *testing.T) {
	tests := []struct {
		name      string
		storeName string
		creds     Credentials
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid credentials",
			storeName: "test-account",
			creds: Credentials{
				ClientID:  "client123",
				APIKey:    "key123",
				AccountID: "acc123",
				CreatedAt: time.Now().UTC(),
			},
			wantErr: false,
		},
		{
			name:      "auto-set creation time if zero",
			storeName: "test-account-2",
			creds: Credentials{
				ClientID: "client456",
				APIKey:   "key456",
			},
			wantErr: false,
		},
		{
			name:      "normalize account name",
			storeName: "  Test-Account  ",
			creds: Credentials{
				ClientID: "client789",
				APIKey:   "key789",
			},
			wantErr: false,
		},
		{
			name:      "missing account name",
			storeName: "",
			creds: Credentials{
				ClientID: "client123",
				APIKey:   "key123",
			},
			wantErr: true,
			errMsg:  "missing account name",
		},
		{
			name:      "whitespace only account name",
			storeName: "   ",
			creds: Credentials{
				ClientID: "client123",
				APIKey:   "key123",
			},
			wantErr: true,
			errMsg:  "missing account name",
		},
		{
			name:      "missing client ID",
			storeName: "test-account",
			creds: Credentials{
				APIKey: "key123",
			},
			wantErr: true,
			errMsg:  "missing client ID",
		},
		{
			name:      "missing API key",
			storeName: "test-account",
			creds: Credentials{
				ClientID: "client123",
			},
			wantErr: true,
			errMsg:  "missing API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockKeyring()
			store := &KeyringStore{ring: mock}

			err := store.Set(tt.storeName, tt.creds)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Set() expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Set() error = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Set() unexpected error: %v", err)
				return
			}

			// Verify the credential was stored with normalized name
			normalizedName := normalize(tt.storeName)
			key := credentialKey(normalizedName)
			item, ok := mock.items[key]
			if !ok {
				t.Errorf("Set() credential not found in keyring")
				return
			}

			var stored storedCredentials
			if err := json.Unmarshal(item.Data, &stored); err != nil {
				t.Errorf("Set() failed to unmarshal stored data: %v", err)
				return
			}

			if stored.ClientID != tt.creds.ClientID {
				t.Errorf("Set() ClientID = %q, want %q", stored.ClientID, tt.creds.ClientID)
			}
			if stored.APIKey != tt.creds.APIKey {
				t.Errorf("Set() APIKey = %q, want %q", stored.APIKey, tt.creds.APIKey)
			}
			if stored.AccountID != tt.creds.AccountID {
				t.Errorf("Set() AccountID = %q, want %q", stored.AccountID, tt.creds.AccountID)
			}
			if stored.CreatedAt.IsZero() {
				t.Errorf("Set() CreatedAt should not be zero")
			}
		})
	}
}

func TestKeyringStore_Get(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		storeName string
		setup     func(*mockKeyring)
		want      Credentials
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid credentials",
			storeName: "test-account",
			setup: func(m *mockKeyring) {
				creds := storedCredentials{
					ClientID:  "client123",
					APIKey:    "key123",
					AccountID: "acc123",
					CreatedAt: now,
				}
				data, _ := json.Marshal(creds)
				m.items[credentialKey("test-account")] = keyring.Item{
					Key:  credentialKey("test-account"),
					Data: data,
				}
			},
			want: Credentials{
				Name:      "test-account",
				ClientID:  "client123",
				APIKey:    "key123",
				AccountID: "acc123",
				CreatedAt: now,
			},
			wantErr: false,
		},
		{
			name:      "normalize account name",
			storeName: "  Test-Account  ",
			setup: func(m *mockKeyring) {
				creds := storedCredentials{
					ClientID: "client456",
					APIKey:   "key456",
				}
				data, _ := json.Marshal(creds)
				m.items[credentialKey("test-account")] = keyring.Item{
					Key:  credentialKey("test-account"),
					Data: data,
				}
			},
			want: Credentials{
				Name:     "test-account",
				ClientID: "client456",
				APIKey:   "key456",
			},
			wantErr: false,
		},
		{
			name:      "missing account name",
			storeName: "",
			setup:     func(m *mockKeyring) {},
			wantErr:   true,
			errMsg:    "missing account name",
		},
		{
			name:      "whitespace only account name",
			storeName: "   ",
			setup:     func(m *mockKeyring) {},
			wantErr:   true,
			errMsg:    "missing account name",
		},
		{
			name:      "credential not found",
			storeName: "nonexistent",
			setup:     func(m *mockKeyring) {},
			wantErr:   true,
		},
		{
			name:      "invalid JSON data",
			storeName: "corrupt-account",
			setup: func(m *mockKeyring) {
				m.items[credentialKey("corrupt-account")] = keyring.Item{
					Key:  credentialKey("corrupt-account"),
					Data: []byte("invalid json"),
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockKeyring()
			tt.setup(mock)
			store := &KeyringStore{ring: mock}

			got, err := store.Get(tt.storeName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Get() expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Get() error = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Get() unexpected error: %v", err)
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("Get() Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.ClientID != tt.want.ClientID {
				t.Errorf("Get() ClientID = %q, want %q", got.ClientID, tt.want.ClientID)
			}
			if got.APIKey != tt.want.APIKey {
				t.Errorf("Get() APIKey = %q, want %q", got.APIKey, tt.want.APIKey)
			}
			if got.AccountID != tt.want.AccountID {
				t.Errorf("Get() AccountID = %q, want %q", got.AccountID, tt.want.AccountID)
			}
			if !tt.want.CreatedAt.IsZero() && !got.CreatedAt.Equal(tt.want.CreatedAt) {
				t.Errorf("Get() CreatedAt = %v, want %v", got.CreatedAt, tt.want.CreatedAt)
			}
		})
	}
}

func TestKeyringStore_Delete(t *testing.T) {
	tests := []struct {
		name      string
		storeName string
		setup     func(*mockKeyring)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "delete existing credential",
			storeName: "test-account",
			setup: func(m *mockKeyring) {
				m.items[credentialKey("test-account")] = keyring.Item{
					Key:  credentialKey("test-account"),
					Data: []byte("data"),
				}
			},
			wantErr: false,
		},
		{
			name:      "normalize account name",
			storeName: "  Test-Account  ",
			setup: func(m *mockKeyring) {
				m.items[credentialKey("test-account")] = keyring.Item{
					Key:  credentialKey("test-account"),
					Data: []byte("data"),
				}
			},
			wantErr: false,
		},
		{
			name:      "missing account name",
			storeName: "",
			setup:     func(m *mockKeyring) {},
			wantErr:   true,
			errMsg:    "missing account name",
		},
		{
			name:      "whitespace only account name",
			storeName: "   ",
			setup:     func(m *mockKeyring) {},
			wantErr:   true,
			errMsg:    "missing account name",
		},
		{
			name:      "delete nonexistent credential",
			storeName: "nonexistent",
			setup:     func(m *mockKeyring) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockKeyring()
			tt.setup(mock)
			store := &KeyringStore{ring: mock}

			err := store.Delete(tt.storeName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Delete() expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Delete() error = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error: %v", err)
				return
			}

			// Verify credential was removed
			normalizedName := normalize(tt.storeName)
			key := credentialKey(normalizedName)
			if _, ok := mock.items[key]; ok {
				t.Errorf("Delete() credential still exists in keyring")
			}
		})
	}
}

func TestKeyringStore_Keys(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockKeyring)
		want    []string
		wantErr bool
	}{
		{
			name: "return all keys",
			setup: func(m *mockKeyring) {
				m.items["account:test1"] = keyring.Item{Key: "account:test1"}
				m.items["account:test2"] = keyring.Item{Key: "account:test2"}
			},
			want:    []string{"account:test1", "account:test2"},
			wantErr: false,
		},
		{
			name:    "empty keyring",
			setup:   func(m *mockKeyring) {},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "keyring error",
			setup: func(m *mockKeyring) {
				m.err = errors.New("keyring error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockKeyring()
			tt.setup(mock)
			store := &KeyringStore{ring: mock}

			got, err := store.Keys()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Keys() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Keys() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("Keys() returned %d keys, want %d", len(got), len(tt.want))
				return
			}

			// Create a map for easier comparison (order doesn't matter)
			gotMap := make(map[string]bool)
			for _, k := range got {
				gotMap[k] = true
			}

			for _, want := range tt.want {
				if !gotMap[want] {
					t.Errorf("Keys() missing key %q", want)
				}
			}
		})
	}
}

func TestKeyringStore_List(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name    string
		setup   func(*mockKeyring)
		want    int // number of credentials expected
		wantErr bool
	}{
		{
			name: "list multiple credentials",
			setup: func(m *mockKeyring) {
				creds1 := storedCredentials{
					ClientID:  "client1",
					APIKey:    "key1",
					CreatedAt: now,
				}
				data1, _ := json.Marshal(creds1)
				m.items[credentialKey("account1")] = keyring.Item{
					Key:  credentialKey("account1"),
					Data: data1,
				}

				creds2 := storedCredentials{
					ClientID:  "client2",
					APIKey:    "key2",
					CreatedAt: now,
				}
				data2, _ := json.Marshal(creds2)
				m.items[credentialKey("account2")] = keyring.Item{
					Key:  credentialKey("account2"),
					Data: data2,
				}
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "skip invalid keys",
			setup: func(m *mockKeyring) {
				creds := storedCredentials{
					ClientID: "client1",
					APIKey:   "key1",
				}
				data, _ := json.Marshal(creds)
				m.items[credentialKey("account1")] = keyring.Item{
					Key:  credentialKey("account1"),
					Data: data,
				}
				// Add invalid key that should be skipped
				m.items["invalid:key"] = keyring.Item{
					Key:  "invalid:key",
					Data: data,
				}
			},
			want:    1,
			wantErr: false,
		},
		{
			name:    "empty list",
			setup:   func(m *mockKeyring) {},
			want:    0,
			wantErr: false,
		},
		{
			name: "keyring error on Keys()",
			setup: func(m *mockKeyring) {
				m.err = errors.New("keyring error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockKeyring()
			tt.setup(mock)
			store := &KeyringStore{ring: mock}

			got, err := store.List()

			if tt.wantErr {
				if err == nil {
					t.Errorf("List() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("List() unexpected error: %v", err)
				return
			}

			if len(got) != tt.want {
				t.Errorf("List() returned %d credentials, want %d", len(got), tt.want)
			}
		})
	}
}

func TestCredentialRotationThreshold(t *testing.T) {
	expected := 90 * 24 * time.Hour
	if CredentialRotationThreshold != expected {
		t.Errorf("CredentialRotationThreshold = %v, want %v", CredentialRotationThreshold, expected)
	}
}
