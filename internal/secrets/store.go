package secrets

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/airwallex-cli/internal/config"
)

const (
	// CredentialRotationThreshold is the age after which credentials should be rotated
	CredentialRotationThreshold = 90 * 24 * time.Hour
)

var warnedAccounts sync.Map

type Store interface {
	Keys() ([]string, error)
	Set(name string, creds Credentials) error
	Get(name string) (Credentials, error)
	Delete(name string) error
	List() ([]Credentials, error)
}

type KeyringStore struct {
	ring keyring.Keyring
}

type Credentials struct {
	Name      string    `json:"name"`
	ClientID  string    `json:"client_id"`
	APIKey    string    `json:"-"`
	AccountID string    `json:"account_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type storedCredentials struct {
	ClientID  string    `json:"client_id"`
	APIKey    string    `json:"api_key"`
	AccountID string    `json:"account_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func OpenDefault() (Store, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: config.AppName,
	})
	if err != nil {
		return nil, err
	}
	return &KeyringStore{ring: ring}, nil
}

func (s *KeyringStore) Keys() ([]string, error) {
	return s.ring.Keys()
}

func (s *KeyringStore) Set(name string, creds Credentials) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing account name")
	}
	if creds.ClientID == "" {
		return fmt.Errorf("missing client ID")
	}
	if creds.APIKey == "" {
		return fmt.Errorf("missing API key")
	}
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = time.Now().UTC()
	}

	payload, err := json.Marshal(storedCredentials{
		ClientID:  creds.ClientID,
		APIKey:    creds.APIKey,
		AccountID: creds.AccountID,
		CreatedAt: creds.CreatedAt,
	})
	if err != nil {
		return err
	}

	return s.ring.Set(keyring.Item{
		Key:  credentialKey(name),
		Data: payload,
	})
}

func (s *KeyringStore) Get(name string) (Credentials, error) {
	name = normalize(name)
	if name == "" {
		return Credentials{}, fmt.Errorf("missing account name")
	}
	item, err := s.ring.Get(credentialKey(name))
	if err != nil {
		return Credentials{}, err
	}
	var stored storedCredentials
	if err := json.Unmarshal(item.Data, &stored); err != nil {
		return Credentials{}, err
	}

	creds := Credentials{
		Name:      name,
		ClientID:  stored.ClientID,
		APIKey:    stored.APIKey,
		AccountID: stored.AccountID,
		CreatedAt: stored.CreatedAt,
	}

	// Warn if credentials are older than 90 days (backwards compatible with zero time)
	// Only warn once per session per account to avoid spam
	if !creds.CreatedAt.IsZero() && time.Since(creds.CreatedAt) > CredentialRotationThreshold {
		if _, warned := warnedAccounts.LoadOrStore(name, true); !warned {
			slog.Warn("credentials over 90 days old, consider rotating", "account", name, "age_days", int(time.Since(creds.CreatedAt).Hours()/24))
		}
	}

	return creds, nil
}

func (s *KeyringStore) Delete(name string) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing account name")
	}
	return s.ring.Remove(credentialKey(name))
}

func (s *KeyringStore) List() ([]Credentials, error) {
	keys, err := s.Keys()
	if err != nil {
		return nil, err
	}
	var out []Credentials
	for _, k := range keys {
		name, ok := ParseCredentialKey(k)
		if !ok {
			continue
		}
		creds, err := s.Get(name)
		if err != nil {
			return nil, err
		}
		out = append(out, creds)
	}
	return out, nil
}

func ParseCredentialKey(k string) (name string, ok bool) {
	const prefix = "account:"
	if !strings.HasPrefix(k, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(k, prefix)
	if strings.TrimSpace(rest) == "" {
		return "", false
	}
	return rest, true
}

func credentialKey(name string) string {
	return fmt.Sprintf("account:%s", name)
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
