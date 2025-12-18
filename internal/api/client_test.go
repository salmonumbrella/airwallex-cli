package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_ensureValidToken_fetchesWhenEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/authentication/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token": "test-token", "expires_at": "2099-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
	}

	err := c.ensureValidToken()
	if err != nil {
		t.Fatalf("ensureValidToken() error: %v", err)
	}
	if c.token == nil {
		t.Fatal("token is nil")
	}
	if c.token.Token != "test-token" {
		t.Errorf("token = %q, want 'test-token'", c.token.Token)
	}
}

func TestClient_ensureValidToken_reusesValidToken(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token": "new-token", "expires_at": "2099-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "existing-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	err := c.ensureValidToken()
	if err != nil {
		t.Fatalf("ensureValidToken() error: %v", err)
	}
	if callCount != 0 {
		t.Errorf("expected no API calls, got %d", callCount)
	}
	if c.token.Token != "existing-token" {
		t.Errorf("token = %q, want 'existing-token'", c.token.Token)
	}
}

func TestClient_ensureValidToken_refreshesExpiredToken(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token": "new-token", "expires_at": "2099-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "old-token",
			ExpiresAt: time.Now().Add(30 * time.Second), // Within 60s threshold
		},
	}

	err := c.ensureValidToken()
	if err != nil {
		t.Fatalf("ensureValidToken() error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call, got %d", callCount)
	}
	if c.token.Token != "new-token" {
		t.Errorf("token = %q, want 'new-token'", c.token.Token)
	}
}

func TestClient_doWithRetry_noRetryOn4xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestClient_doWithRetry_retriesOn429WithExponentialBackoff(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate limit"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 3 {
		t.Errorf("expected 3 calls (2 retries), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify exponential backoff: 1s + 2s = ~3s (with jitter could be more)
	if elapsed < 3*time.Second {
		t.Errorf("elapsed = %v, expected at least 3s for exponential backoff", elapsed)
	}
}

func TestClient_doWithRetry_maxRetriesOn429(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error": "rate limit"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	// Should be 1 initial + 3 retries = 4 total calls
	if callCount != 4 {
		t.Errorf("expected 4 calls (3 retries), got %d", callCount)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
}

func TestClient_doWithRetry_retriesOnce5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "server error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify 1s delay
	if elapsed < 1*time.Second {
		t.Errorf("elapsed = %v, expected at least 1s delay", elapsed)
	}
}

func TestClient_doWithRetry_noSecondRetryOn5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	// Should be 1 initial + 1 retry = 2 total calls (not more)
	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry only), got %d", callCount)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestClient_doWithRetry_successOnFirstAttempt(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Post_retriesWithBodyReplay(t *testing.T) {
	callCount := 0
	var receivedBodies []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Read and store the body
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		receivedBodies = append(receivedBodies, string(body[:n]))

		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "server error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	requestBody := map[string]string{"test": "data", "field": "value"}
	resp, err := c.Post("/test", requestBody)
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify body was sent on both attempts
	if len(receivedBodies) != 2 {
		t.Fatalf("expected 2 bodies received, got %d", len(receivedBodies))
	}
	if receivedBodies[0] != receivedBodies[1] {
		t.Errorf("body mismatch: first=%q, second=%q", receivedBodies[0], receivedBodies[1])
	}
	if receivedBodies[0] == "" {
		t.Error("body was empty on retry")
	}
}

func TestClient_doWithRetry_mixedErrors5xxThen429(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "server error"}`))
		case 2:
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate limit"}`))
		case 3:
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate limit"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data": "success"}`))
		}
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	// Should be: 1 initial 5xx + 1 retry (gets 429) + 2 retries for 429 + 1 success = 4 total
	if callCount != 4 {
		t.Errorf("expected 4 calls (5xx retry + 2x 429 retries), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_doWithRetry_mixedErrors429Then5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate limit"}`))
		case 2:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "server error"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data": "success"}`))
		}
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		ctx:        context.Background(),
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	// Should be: 1 initial 429 + 1 retry (gets 5xx) + 1 retry for 5xx (gets success) = 3 total
	if callCount != 3 {
		t.Errorf("expected 3 calls (429 retry + 5xx retry + success), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
