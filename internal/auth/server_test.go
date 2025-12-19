package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/secrets"
)

// mockStore implements secrets.Store for testing
type mockStore struct {
	creds map[string]secrets.Credentials
}

func newMockStore() *mockStore {
	return &mockStore{
		creds: make(map[string]secrets.Credentials),
	}
}

func (m *mockStore) Keys() ([]string, error) {
	accounts := make([]string, 0, len(m.creds))
	for account := range m.creds {
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (m *mockStore) Get(account string) (secrets.Credentials, error) {
	if creds, ok := m.creds[account]; ok {
		return creds, nil
	}
	return secrets.Credentials{}, fmt.Errorf("account not found")
}

func (m *mockStore) Set(account string, creds secrets.Credentials) error {
	m.creds[account] = creds
	return nil
}

func (m *mockStore) Delete(account string) error {
	delete(m.creds, account)
	return nil
}

func (m *mockStore) List() ([]secrets.Credentials, error) {
	credsList := make([]secrets.Credentials, 0, len(m.creds))
	for _, creds := range m.creds {
		credsList = append(credsList, creds)
	}
	return credsList, nil
}

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
		attempts    int
		endpoint    string
		wantErr     bool
	}{
		{
			name:        "within limit",
			maxAttempts: 10,
			attempts:    5,
			endpoint:    "/validate",
			wantErr:     false,
		},
		{
			name:        "at limit",
			maxAttempts: 10,
			attempts:    10,
			endpoint:    "/validate",
			wantErr:     false,
		},
		{
			name:        "exceeds limit by 1",
			maxAttempts: 10,
			attempts:    11,
			endpoint:    "/validate",
			wantErr:     true,
		},
		{
			name:        "exceeds limit by many",
			maxAttempts: 10,
			attempts:    20,
			endpoint:    "/validate",
			wantErr:     true,
		},
		{
			name:        "different endpoints tracked separately",
			maxAttempts: 2,
			attempts:    2,
			endpoint:    "/submit",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := newRateLimiter(tt.maxAttempts)

			var err error
			for i := 0; i < tt.attempts; i++ {
				err = rl.check(tt.endpoint)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("rateLimiter.check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRateLimiterSeparateEndpoints(t *testing.T) {
	rl := newRateLimiter(2)

	// Use up limit for /validate
	if err := rl.check("/validate"); err != nil {
		t.Fatalf("unexpected error on first /validate: %v", err)
	}
	if err := rl.check("/validate"); err != nil {
		t.Fatalf("unexpected error on second /validate: %v", err)
	}

	// Third attempt on /validate should fail
	if err := rl.check("/validate"); err == nil {
		t.Error("expected error on third /validate, got nil")
	}

	// But /submit should still work (separate counter)
	if err := rl.check("/submit"); err != nil {
		t.Errorf("unexpected error on first /submit: %v", err)
	}
	if err := rl.check("/submit"); err != nil {
		t.Errorf("unexpected error on second /submit: %v", err)
	}

	// Third attempt on /submit should fail
	if err := rl.check("/submit"); err == nil {
		t.Error("expected error on third /submit, got nil")
	}
}

func TestHandleValidateRateLimit(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Override rate limiter with lower limit for testing
	server.limiter = newRateLimiter(3)

	reqBody := map[string]string{
		"account_name": "test",
		"client_id":    "test_client",
		"api_key":      "test_key",
	}
	body, _ := json.Marshal(reqBody)

	// Make requests up to the limit
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleValidate(w, req)

		// These should succeed (though validation may fail due to mock API)
		if w.Code == http.StatusTooManyRequests {
			t.Errorf("request %d got 429, expected to be within limit", i)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != false {
		t.Error("expected success=false in response")
	}

	if resp["error"] == nil {
		t.Error("expected error message in response")
	}
}

func TestHandleSubmitRateLimit(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Override rate limiter with lower limit for testing
	server.limiter = newRateLimiter(3)

	reqBody := map[string]string{
		"account_name": "test",
		"client_id":    "test_client",
		"api_key":      "test_key",
	}
	body, _ := json.Marshal(reqBody)

	// Make requests up to the limit
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(body))
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleSubmit(w, req)

		// These should succeed (though validation may fail due to mock API)
		if w.Code == http.StatusTooManyRequests {
			t.Errorf("request %d got 429, expected to be within limit", i)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != false {
		t.Error("expected success=false in response")
	}

	if resp["error"] == nil {
		t.Error("expected error message in response")
	}
}

func TestRateLimitEndpointSeparation(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Override rate limiter with lower limit for testing
	server.limiter = newRateLimiter(2)

	reqBody := map[string]string{
		"account_name": "test",
		"client_id":    "test_client",
		"api_key":      "test_key",
	}
	body, _ := json.Marshal(reqBody)

	// Exhaust /validate limit
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.handleValidate(w, req)
	}

	// Verify /validate is rate limited
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected /validate to be rate limited, got %d", w.Code)
	}

	// But /submit should still work
	req = httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	server.handleSubmit(w, req)

	if w.Code == http.StatusTooManyRequests {
		t.Error("expected /submit to still work after /validate exhausted")
	}
}

func TestNewSetupServer(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("NewSetupServer() error = %v", err)
	}

	if server.limiter == nil {
		t.Error("expected rate limiter to be initialized")
	}

	if server.limiter.maxAttempts != 10 {
		t.Errorf("expected maxAttempts=10, got %d", server.limiter.maxAttempts)
	}

	if server.csrfToken == "" {
		t.Error("expected CSRF token to be generated")
	}

	if server.store == nil {
		t.Error("expected store to be set")
	}
}

func TestHandleValidateCSRFProtection(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	reqBody := map[string]string{
		"account_name": "test",
		"client_id":    "test_client",
		"api_key":      "test_key",
	}
	body, _ := json.Marshal(reqBody)

	tests := []struct {
		name       string
		csrfToken  string
		wantStatus int
	}{
		{
			name:       "valid CSRF token",
			csrfToken:  server.csrfToken,
			wantStatus: http.StatusOK, // or other non-403 status
		},
		{
			name:       "invalid CSRF token",
			csrfToken:  "invalid",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "empty CSRF token",
			csrfToken:  "",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
			req.Header.Set("X-CSRF-Token", tt.csrfToken)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.handleValidate(w, req)

			if w.Code == http.StatusForbidden && tt.wantStatus != http.StatusForbidden {
				t.Errorf("got status %d, want not %d", w.Code, http.StatusForbidden)
			}
			if w.Code != http.StatusForbidden && tt.wantStatus == http.StatusForbidden {
				t.Errorf("got status %d, want %d", w.Code, http.StatusForbidden)
			}
		})
	}
}

func TestHandleSubmitCSRFProtection(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	reqBody := map[string]string{
		"account_name": "test",
		"client_id":    "test_client",
		"api_key":      "test_key",
	}
	body, _ := json.Marshal(reqBody)

	tests := []struct {
		name       string
		csrfToken  string
		wantStatus int
	}{
		{
			name:       "valid CSRF token",
			csrfToken:  server.csrfToken,
			wantStatus: http.StatusOK, // or other non-403 status
		},
		{
			name:       "invalid CSRF token",
			csrfToken:  "invalid",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "empty CSRF token",
			csrfToken:  "",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(body))
			req.Header.Set("X-CSRF-Token", tt.csrfToken)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.handleSubmit(w, req)

			if w.Code == http.StatusForbidden && tt.wantStatus != http.StatusForbidden {
				t.Errorf("got status %d, want not %d", w.Code, http.StatusForbidden)
			}
			if w.Code != http.StatusForbidden && tt.wantStatus == http.StatusForbidden {
				t.Errorf("got status %d, want %d", w.Code, http.StatusForbidden)
			}
		})
	}
}

// Additional test to verify rate limiting happens before other checks
func TestRateLimitBeforeValidation(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Set very low limit
	server.limiter = newRateLimiter(1)

	reqBody := map[string]string{
		"account_name": "test",
		"client_id":    "test_client",
		"api_key":      "test_key",
	}
	body, _ := json.Marshal(reqBody)

	// First request
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	// Second request should be rate limited even with invalid CSRF
	req = httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", "invalid_token")
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.handleValidate(w, req)

	// Should get 429, not 403 (rate limit checked first)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected rate limit (429) to be checked before CSRF, got %d", w.Code)
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := newRateLimiter(100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run concurrent checks
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					done <- true
					return
				default:
					_ = rl.check("/validate")
				}
			}
		}()
	}

	// Let it run briefly
	// time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify the map isn't corrupted
	if err := rl.check("/test"); err != nil {
		// Just checking it doesn't panic
	}
}
