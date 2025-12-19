package api

import (
	"context"
	"crypto/tls"
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
	}

	err := c.ensureValidToken(context.Background())
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

		token: &TokenCache{
			Token:     "existing-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	err := c.ensureValidToken(context.Background())
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

		token: &TokenCache{
			Token:     "old-token",
			ExpiresAt: time.Now().Add(30 * time.Second), // Within 60s threshold
		},
	}

	err := c.ensureValidToken(context.Background())
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(context.Background(), req)
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
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

func TestClient_doWithRetry_respectsRetryAfterHeader(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Retry-After", "2")
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(context.Background(), req)
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

	// Verify we waited for the Retry-After duration (2 seconds)
	if elapsed < 2*time.Second {
		t.Errorf("elapsed = %v, expected at least 2s based on Retry-After header", elapsed)
	}
	// Should be close to 2s, not the exponential backoff of ~1s
	if elapsed > 3*time.Second {
		t.Errorf("elapsed = %v, expected around 2s based on Retry-After header, not exponential backoff", elapsed)
	}
}

func TestClient_doWithRetry_contextCancellationDuringRetry(t *testing.T) {
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Create a context that we'll cancel during retry
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 500ms (during the retry delay)
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(ctx, req)
	elapsed := time.Since(start)

	if err == nil {
		resp.Body.Close()
		t.Fatal("expected context.Canceled error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}

	// Should fail quickly (around 500ms), not wait for full retry delay
	if elapsed > 1*time.Second {
		t.Errorf("elapsed = %v, expected cancellation within ~500ms", elapsed)
	}

	// Should have made 1 call before context was cancelled
	if callCount != 1 {
		t.Errorf("expected 1 call before cancellation, got %d", callCount)
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(context.Background(), req)
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
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

func TestClient_Post_retriesOn429WithBodyReplay(t *testing.T) {
	callCount := 0
	var receivedBodies []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Read and store the body
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		receivedBodies = append(receivedBodies, string(body[:n]))

		if callCount == 1 {
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	requestBody := map[string]string{"test": "data", "field": "value"}
	resp, err := c.Post(context.Background(), "/test", requestBody)
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
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

// TestClient_doWithRetry_GET_retriesOn5xx verifies that GET requests ARE retried on 5xx errors
func TestClient_doWithRetry_GET_retriesOn5xx(t *testing.T) {
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry on 5xx for GET), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestClient_doWithRetry_POST_noRetryOn5xx verifies that POST requests are NOT retried on 5xx errors
func TestClient_doWithRetry_POST_noRetryOn5xx(t *testing.T) {
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("POST", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry on 5xx for POST), got %d", callCount)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// TestClient_doWithRetry_POST_retriesOn429 verifies that POST requests ARE still retried on 429 rate limit
func TestClient_doWithRetry_POST_retriesOn429(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("POST", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry on 429 for POST), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestClient_doWithRetry_PUT_noRetryOn5xx verifies that PUT requests are NOT retried on 5xx errors
func TestClient_doWithRetry_PUT_noRetryOn5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error": "bad gateway"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("PUT", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry on 5xx for PUT), got %d", callCount)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}
}

// TestClient_doWithRetry_DELETE_noRetryOn5xx verifies that DELETE requests are NOT retried on 5xx errors
func TestClient_doWithRetry_DELETE_noRetryOn5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error": "service unavailable"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("DELETE", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry on 5xx for DELETE), got %d", callCount)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}

// TestClient_doWithRetry_PATCH_noRetryOn5xx verifies that PATCH requests are NOT retried on 5xx errors
func TestClient_doWithRetry_PATCH_noRetryOn5xx(t *testing.T) {
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

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("PATCH", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 1 {
		t.Errorf("expected 1 call (no retry on 5xx for PATCH), got %d", callCount)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// TestClient_doWithRetry_HEAD_retriesOn5xx verifies that HEAD requests ARE retried on 5xx errors
func TestClient_doWithRetry_HEAD_retriesOn5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("HEAD", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry on 5xx for HEAD), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestClient_doWithRetry_OPTIONS_retriesOn5xx verifies that OPTIONS requests ARE retried on 5xx errors
func TestClient_doWithRetry_OPTIONS_retriesOn5xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	req, _ := http.NewRequest("OPTIONS", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	defer resp.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry on 5xx for OPTIONS), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestNewClient_enforcesTLS12(t *testing.T) {
	client := NewClient("test-id", "test-key")

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d (TLS 1.2)", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
}

func TestNewClientWithAccount_enforcesTLS12(t *testing.T) {
	client := NewClientWithAccount("test-id", "test-key", "account-id")

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d (TLS 1.2)", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
}

// TestGenerateIdempotencyKey verifies that idempotency keys are unique
func TestGenerateIdempotencyKey(t *testing.T) {
	key1 := generateIdempotencyKey()
	key2 := generateIdempotencyKey()

	if key1 == "" {
		t.Error("expected non-empty key")
	}
	if key2 == "" {
		t.Error("expected non-empty key")
	}
	if key1 == key2 {
		t.Errorf("expected unique keys, got duplicate: %s", key1)
	}

	// Verify key is hex encoded (32 chars for 16 bytes)
	if len(key1) != 32 {
		t.Errorf("expected 32 character hex string, got %d characters", len(key1))
	}
}

// TestIsFinancialOperation verifies financial path detection
func TestIsFinancialOperation(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/transfers/create", true},
		{"/api/v1/issuing/cards/create", true},
		{"/api/v1/beneficiaries/create", true},
		{"/api/v1/accounts/list", false},
		{"/api/v1/balances", false},
		{"/api/v1/transfers/list", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isFinancialOperation(tt.path)
			if result != tt.expected {
				t.Errorf("isFinancialOperation(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestClient_Post_addsIdempotencyKeyForFinancialOperations verifies idempotency header for financial paths
func TestClient_Post_addsIdempotencyKeyForFinancialOperations(t *testing.T) {
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Test financial operation
	resp, err := c.Post(context.Background(), "/api/v1/transfers/create", map[string]string{"amount": "100"})
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	defer resp.Body.Close()

	idempotencyKey := capturedHeaders.Get("x-idempotency-key")
	if idempotencyKey == "" {
		t.Error("expected x-idempotency-key header for financial operation, got empty")
	}
	if len(idempotencyKey) != 32 {
		t.Errorf("expected 32 character idempotency key, got %d characters", len(idempotencyKey))
	}
}

// TestClient_Post_noIdempotencyKeyForNonFinancialOperations verifies no header for non-financial paths
func TestClient_Post_noIdempotencyKeyForNonFinancialOperations(t *testing.T) {
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:    server.URL,
		clientID:   "test-id",
		apiKey:     "test-key",
		httpClient: http.DefaultClient,
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Test non-financial operation
	resp, err := c.Post(context.Background(), "/api/v1/accounts/list", map[string]string{"filter": "active"})
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	defer resp.Body.Close()

	idempotencyKey := capturedHeaders.Get("x-idempotency-key")
	if idempotencyKey != "" {
		t.Errorf("expected no x-idempotency-key header for non-financial operation, got %q", idempotencyKey)
	}
}
