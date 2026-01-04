package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		clientIP    string
		endpoint    string
		wantErr     bool
	}{
		{
			name:        "within limit",
			maxAttempts: 10,
			attempts:    5,
			clientIP:    "127.0.0.1",
			endpoint:    "/validate",
			wantErr:     false,
		},
		{
			name:        "at limit",
			maxAttempts: 10,
			attempts:    10,
			clientIP:    "127.0.0.1",
			endpoint:    "/validate",
			wantErr:     false,
		},
		{
			name:        "exceeds limit by 1",
			maxAttempts: 10,
			attempts:    11,
			clientIP:    "127.0.0.1",
			endpoint:    "/validate",
			wantErr:     true,
		},
		{
			name:        "exceeds limit by many",
			maxAttempts: 10,
			attempts:    20,
			clientIP:    "127.0.0.1",
			endpoint:    "/validate",
			wantErr:     true,
		},
		{
			name:        "different endpoints tracked separately",
			maxAttempts: 2,
			attempts:    2,
			clientIP:    "127.0.0.1",
			endpoint:    "/submit",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := newRateLimiter(tt.maxAttempts, 15*time.Minute)

			var err error
			for i := 0; i < tt.attempts; i++ {
				err = rl.check(tt.clientIP, tt.endpoint)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("rateLimiter.check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRateLimiterSeparateEndpoints(t *testing.T) {
	rl := newRateLimiter(2, 15*time.Minute)
	clientIP := "127.0.0.1"

	// Use up limit for /validate
	if err := rl.check(clientIP, "/validate"); err != nil {
		t.Fatalf("unexpected error on first /validate: %v", err)
	}
	if err := rl.check(clientIP, "/validate"); err != nil {
		t.Fatalf("unexpected error on second /validate: %v", err)
	}

	// Third attempt on /validate should fail
	if err := rl.check(clientIP, "/validate"); err == nil {
		t.Error("expected error on third /validate, got nil")
	}

	// But /submit should still work (separate counter)
	if err := rl.check(clientIP, "/submit"); err != nil {
		t.Errorf("unexpected error on first /submit: %v", err)
	}
	if err := rl.check(clientIP, "/submit"); err != nil {
		t.Errorf("unexpected error on second /submit: %v", err)
	}

	// Third attempt on /submit should fail
	if err := rl.check(clientIP, "/submit"); err == nil {
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
	server.limiter = newRateLimiter(3, 15*time.Minute)

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
	server.limiter = newRateLimiter(3, 15*time.Minute)

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
	server.limiter = newRateLimiter(2, 15*time.Minute)

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

	// CSRF token should be 64 characters (32 bytes hex encoded)
	if len(server.csrfToken) != 64 {
		t.Errorf("expected CSRF token length 64, got %d", len(server.csrfToken))
	}

	if server.store == nil {
		t.Error("expected store to be set")
	}

	if server.stopCleanup == nil {
		t.Error("expected stopCleanup channel to be initialized")
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

// Test to verify CSRF validation happens before rate limiting
func TestCSRFBeforeRateLimit(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Set very low limit
	server.limiter = newRateLimiter(1, 15*time.Minute)

	reqBody := map[string]string{
		"account_name": "test",
		"client_id":    "test_client",
		"api_key":      "test_key",
	}
	body, _ := json.Marshal(reqBody)

	// First request with valid CSRF
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	// Second request with invalid CSRF should be rejected with 403
	req = httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("X-CSRF-Token", "invalid_token")
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.handleValidate(w, req)

	// Should get 403, not 429 (CSRF checked first)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected CSRF rejection (403) to be checked before rate limit, got %d", w.Code)
	}
}

func TestRateLimiterPerClientIP(t *testing.T) {
	rl := newRateLimiter(2, 15*time.Minute)

	// Client 1 exhausts their limit
	if err := rl.check("127.0.0.1", "/validate"); err != nil {
		t.Fatalf("unexpected error on first request from client 1: %v", err)
	}
	if err := rl.check("127.0.0.1", "/validate"); err != nil {
		t.Fatalf("unexpected error on second request from client 1: %v", err)
	}
	if err := rl.check("127.0.0.1", "/validate"); err == nil {
		t.Error("expected error on third request from client 1, got nil")
	}

	// Client 2 should still have their own limit
	if err := rl.check("127.0.0.2", "/validate"); err != nil {
		t.Errorf("unexpected error on first request from client 2: %v", err)
	}
	if err := rl.check("127.0.0.2", "/validate"); err != nil {
		t.Errorf("unexpected error on second request from client 2: %v", err)
	}
	if err := rl.check("127.0.0.2", "/validate"); err == nil {
		t.Error("expected error on third request from client 2, got nil")
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := newRateLimiter(2, 100*time.Millisecond)
	clientIP := "127.0.0.1"

	// Exhaust the limit
	if err := rl.check(clientIP, "/validate"); err != nil {
		t.Fatalf("unexpected error on first request: %v", err)
	}
	if err := rl.check(clientIP, "/validate"); err != nil {
		t.Fatalf("unexpected error on second request: %v", err)
	}
	if err := rl.check(clientIP, "/validate"); err == nil {
		t.Error("expected error on third request, got nil")
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be able to make requests again
	if err := rl.check(clientIP, "/validate"); err != nil {
		t.Errorf("unexpected error after window reset: %v", err)
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := newRateLimiter(100, 15*time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run concurrent checks
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		clientIP := fmt.Sprintf("127.0.0.%d", i)
		go func(ip string) {
			for {
				select {
				case <-ctx.Done():
					done <- true
					return
				default:
					//nolint:errcheck,gosec // intentionally ignoring error in concurrent stress test
					rl.check(ip, "/validate")
				}
			}
		}(clientIP)
	}

	// Let it run briefly
	// time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify the map isn't corrupted - just checking it doesn't panic
	_ = rl.check("127.0.0.1", "/test")
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := newRateLimiter(10, 100*time.Millisecond)

	// Add some entries
	_ = rl.check("1.1.1.1", "/test")
	_ = rl.check("2.2.2.2", "/test")
	_ = rl.check("3.3.3.3", "/test")

	if rl.size() != 3 {
		t.Errorf("size = %d, want 3", rl.size())
	}

	// Wait for entries to expire
	time.Sleep(150 * time.Millisecond)

	// Run cleanup
	rl.cleanup()

	if rl.size() != 0 {
		t.Errorf("size after cleanup = %d, want 0", rl.size())
	}
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       any
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success response",
			status:     http.StatusOK,
			data:       map[string]any{"success": true, "message": "ok"},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"ok","success":true}`,
		},
		{
			name:       "error response with 400",
			status:     http.StatusBadRequest,
			data:       map[string]any{"success": false, "error": "bad request"},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"bad request","success":false}`,
		},
		{
			name:       "rate limit response",
			status:     http.StatusTooManyRequests,
			data:       map[string]any{"success": false, "error": "too many attempts"},
			wantStatus: http.StatusTooManyRequests,
			wantBody:   `{"error":"too many attempts","success":false}`,
		},
		{
			name:       "simple string map",
			status:     http.StatusOK,
			data:       map[string]string{"key": "value"},
			wantStatus: http.StatusOK,
			wantBody:   `{"key":"value"}`,
		},
		{
			name:       "nested object",
			status:     http.StatusOK,
			data:       map[string]any{"outer": map[string]string{"inner": "value"}},
			wantStatus: http.StatusOK,
			wantBody:   `{"outer":{"inner":"value"}}`,
		},
		{
			name:       "array response",
			status:     http.StatusOK,
			data:       []string{"a", "b", "c"},
			wantStatus: http.StatusOK,
			wantBody:   `["a","b","c"]`,
		},
		{
			name:       "empty object",
			status:     http.StatusOK,
			data:       map[string]any{},
			wantStatus: http.StatusOK,
			wantBody:   `{}`,
		},
		{
			name:       "null value",
			status:     http.StatusOK,
			data:       nil,
			wantStatus: http.StatusOK,
			wantBody:   `null`,
		},
		{
			name:       "boolean response",
			status:     http.StatusOK,
			data:       true,
			wantStatus: http.StatusOK,
			wantBody:   `true`,
		},
		{
			name:       "integer response",
			status:     http.StatusCreated,
			data:       42,
			wantStatus: http.StatusCreated,
			wantBody:   `42`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSON(w, tt.status, tt.data)

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			// Check Content-Type header
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}

			// Check body (trim trailing newline from json.Encoder)
			got := w.Body.String()
			got = got[:len(got)-1] // Remove trailing newline
			if got != tt.wantBody {
				t.Errorf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestWriteJSON_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{
			name:       "localhost with port",
			remoteAddr: "127.0.0.1:12345",
			want:       "127.0.0.1",
		},
		{
			name:       "localhost without port returns empty",
			remoteAddr: "127.0.0.1",
			want:       "", // net.SplitHostPort fails without port
		},
		{
			name:       "ipv4 with port",
			remoteAddr: "192.168.1.100:54321",
			want:       "192.168.1.100",
		},
		{
			name:       "ipv6 localhost with port",
			remoteAddr: "[::1]:12345",
			want:       "::1",
		},
		{
			name:       "ipv6 address with port",
			remoteAddr: "[2001:db8::1]:8080",
			want:       "2001:db8::1",
		},
		{
			name:       "empty remote addr",
			remoteAddr: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr

			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetClientIP_RealRequests(t *testing.T) {
	// Test with actual httptest requests to ensure RemoteAddr is parsed correctly
	req := httptest.NewRequest(http.MethodPost, "/validate", nil)

	// httptest.NewRequest sets RemoteAddr to "192.0.2.1:1234" by default
	ip := getClientIP(req)
	if ip != "192.0.2.1" {
		t.Errorf("getClientIP() with default httptest request = %q, want %q", ip, "192.0.2.1")
	}
}

func TestHandleSetup(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantContains []string
		wantNotFound bool
	}{
		{
			name:       "root path returns setup page",
			path:       "/",
			wantStatus: http.StatusOK,
			wantContains: []string{
				server.csrfToken,
			},
		},
		{
			name:         "non-root path returns 404",
			path:         "/unknown",
			wantStatus:   http.StatusNotFound,
			wantNotFound: true,
		},
		{
			name:         "nested path returns 404",
			path:         "/some/nested/path",
			wantStatus:   http.StatusNotFound,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			server.handleSetup(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if !tt.wantNotFound {
				// Check security headers
				if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
					t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
				}
				if xfo := w.Header().Get("X-Frame-Options"); xfo != "DENY" {
					t.Errorf("X-Frame-Options = %q, want %q", xfo, "DENY")
				}
				if xcto := w.Header().Get("X-Content-Type-Options"); xcto != "nosniff" {
					t.Errorf("X-Content-Type-Options = %q, want %q", xcto, "nosniff")
				}
				if csp := w.Header().Get("Content-Security-Policy"); csp == "" {
					t.Error("Content-Security-Policy header not set")
				}

				// Check body contains CSRF token
				body := w.Body.String()
				for _, want := range tt.wantContains {
					if !bytes.Contains([]byte(body), []byte(want)) {
						t.Errorf("body does not contain %q", want)
					}
				}
			}
		})
	}
}

func TestHandleSuccess(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	t.Run("without pending result", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/success", nil)
		w := httptest.NewRecorder()
		server.handleSuccess(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		// Check security headers
		if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Errorf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
		}
		if xfo := w.Header().Get("X-Frame-Options"); xfo != "DENY" {
			t.Errorf("X-Frame-Options = %q, want %q", xfo, "DENY")
		}
	})

	t.Run("with pending result", func(t *testing.T) {
		server.pendingMu.Lock()
		server.pendingResult = &SetupResult{
			AccountName: "test-account",
			ClientID:    "test-client",
		}
		server.pendingMu.Unlock()

		req := httptest.NewRequest(http.MethodGet, "/success", nil)
		w := httptest.NewRecorder()
		server.handleSuccess(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		body := w.Body.String()
		if !bytes.Contains([]byte(body), []byte("test-account")) {
			t.Error("body should contain account name")
		}
	})
}

func TestHandleComplete(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		store := newMockStore()
		server, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/complete", nil)
		w := httptest.NewRecorder()
		server.handleComplete(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("invalid CSRF token", func(t *testing.T) {
		store := newMockStore()
		server, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/complete", nil)
		req.Header.Set("X-CSRF-Token", "invalid")
		w := httptest.NewRecorder()
		server.handleComplete(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("valid complete without pending result", func(t *testing.T) {
		store := newMockStore()
		server, err := NewSetupServer(store)
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/complete", nil)
		req.Header.Set("X-CSRF-Token", server.csrfToken)
		w := httptest.NewRecorder()

		// Run in goroutine since handleComplete closes shutdown channel
		done := make(chan bool)
		go func() {
			server.handleComplete(w, req)
			done <- true
		}()

		<-done

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]any
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["success"] != true {
			t.Error("expected success=true")
		}
	})
}

func TestHandleValidate_MethodNotAllowed(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleSubmit_MethodNotAllowed(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/submit", nil)
	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleValidate_InvalidJSON(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader([]byte("not json")))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleValidate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != false {
		t.Error("expected success=false")
	}
	if resp["error"] != "Invalid request body" {
		t.Errorf("error = %q, want %q", resp["error"], "Invalid request body")
	}
}

func TestHandleSubmit_InvalidJSON(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader([]byte("not json")))
	req.Header.Set("X-CSRF-Token", server.csrfToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["success"] != false {
		t.Error("expected success=false")
	}
}

func TestHandleValidate_ValidationErrors(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		name    string
		body    map[string]string
		wantErr string
	}{
		{
			name:    "empty account name",
			body:    map[string]string{"account_name": "", "client_id": "test", "api_key": "key"},
			wantErr: "account name cannot be empty",
		},
		{
			name:    "invalid account name chars",
			body:    map[string]string{"account_name": "test@invalid", "client_id": "test", "api_key": "key"},
			wantErr: "invalid characters",
		},
		{
			name:    "empty client id",
			body:    map[string]string{"account_name": "test", "client_id": "", "api_key": "key"},
			wantErr: "client ID cannot be empty",
		},
		{
			name:    "empty api key",
			body:    map[string]string{"account_name": "test", "client_id": "client", "api_key": ""},
			wantErr: "API key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
			req.Header.Set("X-CSRF-Token", server.csrfToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.handleValidate(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}

			var resp map[string]any
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp["success"] != false {
				t.Error("expected success=false")
			}

			errStr, ok := resp["error"].(string)
			if !ok {
				t.Fatal("error is not a string")
			}
			if !bytes.Contains([]byte(errStr), []byte(tt.wantErr)) {
				t.Errorf("error = %q, want to contain %q", errStr, tt.wantErr)
			}
		})
	}
}

func TestHandleSubmit_ValidationErrors(t *testing.T) {
	store := newMockStore()
	server, err := NewSetupServer(store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		name    string
		body    map[string]string
		wantErr string
	}{
		{
			name:    "empty account name",
			body:    map[string]string{"account_name": "", "client_id": "test", "api_key": "key"},
			wantErr: "account name cannot be empty",
		},
		{
			name:    "invalid account name chars",
			body:    map[string]string{"account_name": "test@invalid", "client_id": "test", "api_key": "key"},
			wantErr: "invalid characters",
		},
		{
			name:    "empty client id",
			body:    map[string]string{"account_name": "test", "client_id": "", "api_key": "key"},
			wantErr: "client ID cannot be empty",
		},
		{
			name:    "empty api key",
			body:    map[string]string{"account_name": "test", "client_id": "client", "api_key": ""},
			wantErr: "API key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(body))
			req.Header.Set("X-CSRF-Token", server.csrfToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.handleSubmit(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}

			var resp map[string]any
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp["success"] != false {
				t.Error("expected success=false")
			}

			errStr, ok := resp["error"].(string)
			if !ok {
				t.Fatal("error is not a string")
			}
			if !bytes.Contains([]byte(errStr), []byte(tt.wantErr)) {
				t.Errorf("error = %q, want to contain %q", errStr, tt.wantErr)
			}
		})
	}
}

func TestRateLimiter_StartCleanup(t *testing.T) {
	rl := newRateLimiter(10, 50*time.Millisecond)
	stop := make(chan struct{})

	// Add some entries
	_ = rl.check("1.1.1.1", "/test")
	_ = rl.check("2.2.2.2", "/test")

	if rl.size() != 2 {
		t.Errorf("size = %d, want 2", rl.size())
	}

	// Start cleanup with very short interval
	rl.startCleanup(30*time.Millisecond, stop)

	// Wait for entries to expire and cleanup to run
	time.Sleep(150 * time.Millisecond)

	// Size should be 0 after cleanup
	if rl.size() != 0 {
		t.Errorf("size after cleanup = %d, want 0", rl.size())
	}

	// Stop the cleanup goroutine
	close(stop)

	// Give some time for goroutine to exit
	time.Sleep(50 * time.Millisecond)
}
