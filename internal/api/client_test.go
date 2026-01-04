package api

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	testRateLimitBaseDelay    = 10 * time.Millisecond
	testServerErrorRetryDelay = 10 * time.Millisecond
)

func init() {
	rateLimitBaseDelay = testRateLimitBaseDelay
	serverErrorRetryDelay = testServerErrorRetryDelay
}

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

	if callCount != 3 {
		t.Errorf("expected 3 calls (2 retries), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify exponential backoff: base + 2x base (with jitter could be more)
	minDelay := testRateLimitBaseDelay + (2 * testRateLimitBaseDelay)
	if elapsed < minDelay {
		t.Errorf("elapsed = %v, expected at least %v for exponential backoff", elapsed, minDelay)
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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate limit"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify we waited for the Retry-After duration (1 second)
	if elapsed < 1*time.Second {
		t.Errorf("elapsed = %v, expected at least 1s based on Retry-After header", elapsed)
	}
	// Should be close to 1s, not the exponential backoff of ~10ms
	if elapsed > 2*time.Second {
		t.Errorf("elapsed = %v, expected around 1s based on Retry-After header, not exponential backoff", elapsed)
	}
}

func TestClient_doWithRetry_respectsRetryAfterHeaderDate(t *testing.T) {
	callCount := 0
	var retryAt time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			retryAt = time.Now().Add(2 * time.Second).UTC()
			w.Header().Set("Retry-After", retryAt.Format(http.TimeFormat))
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": "rate limit"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if retryAt.IsZero() {
		t.Fatalf("retryAt not set by handler")
	}
	if elapsed < 1*time.Second {
		t.Errorf("elapsed = %v, expected delay based on Retry-After date header", elapsed)
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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Create a context that we'll cancel during retry
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 5ms (during the retry delay)
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(ctx, req)
	elapsed := time.Since(start)

	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected context.Canceled error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}

	// Should fail quickly, not wait for full retry delay
	if elapsed > 50*time.Millisecond {
		t.Errorf("elapsed = %v, expected cancellation within ~50ms", elapsed)
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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify delay
	if elapsed < testServerErrorRetryDelay {
		t.Errorf("elapsed = %v, expected at least %v delay", elapsed, testServerErrorRetryDelay)
	}
}

func TestClient_doWithRetry_contextCancellationDuring5xxRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	start := time.Now()
	resp, err := c.doWithRetry(ctx, req)
	elapsed := time.Since(start)

	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected context.Canceled error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("elapsed = %v, expected cancellation within ~50ms", elapsed)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call before cancellation, got %d", callCount)
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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

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
	defer closeBody(resp)

	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry on 5xx for OPTIONS), got %d", callCount)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestNewClient_enforcesTLS12(t *testing.T) {
	client, err := NewClient("test-id", "test-key")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

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
	client, err := NewClientWithAccount("test-id", "test-key", "account-id")
	if err != nil {
		t.Fatalf("NewClientWithAccount failed: %v", err)
	}

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

func TestNewClient_verifiesCertificates(t *testing.T) {
	client, err := NewClient("test-id", "test-key")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}

	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify = true, want false (certificates must be verified)")
	}
}

func TestNewClientWithAccount_verifiesCertificates(t *testing.T) {
	client, err := NewClientWithAccount("test-id", "test-key", "account-id")
	if err != nil {
		t.Fatalf("NewClientWithAccount failed: %v", err)
	}

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}

	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify = true, want false (certificates must be verified)")
	}
}

// TestGenerateIdempotencyKey verifies that idempotency keys are unique
func TestGenerateIdempotencyKey(t *testing.T) {
	key1, err := generateIdempotencyKey()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	key2, err := generateIdempotencyKey()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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
		// Should match
		{"/api/v1/transfers/create", true},
		{"/api/v1/issuing/cards/create", true},
		{"/api/v1/beneficiaries/create", true},
		{"/api/v1/fx/conversions/create", true},
		{"/api/v1/linked_accounts/create", true},
		{"/api/v1/pa/payment_links/create", true},

		// Should match - query parameters should be ignored
		{"/api/v1/transfers/create?foo=bar", true},
		{"/api/v1/issuing/cards/create?id=123&type=test", true},

		// Should NOT match - false positive cases with HasSuffix
		{"/prefix/api/v1/transfers/create", false},
		{"/custom/api/v1/issuing/cards/create", false},

		// Should NOT match - similar but different paths
		{"/api/v1/transfers/create-preview", false},
		{"/api/v1/custom/transfers/create", false},
		{"/api/v1/issuing/cards/create/something", false},

		// Should NOT match - different endpoints
		{"/api/v1/transfers/list", false},
		{"/api/v1/transfers", false},
		{"/api/v1/balances/current", false},
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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
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
	defer closeBody(resp)

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
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
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
	defer closeBody(resp)

	idempotencyKey := capturedHeaders.Get("x-idempotency-key")
	if idempotencyKey != "" {
		t.Errorf("expected no x-idempotency-key header for non-financial operation, got %q", idempotencyKey)
	}
}

// TestCircuitBreaker_opensAfterThreshold tests that the circuit breaker opens after consecutive failures
func TestCircuitBreaker_opensAfterThreshold(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make requests until circuit opens - each GET gets 1 retry, so 2 failures per request
	// We need 5 failures total, which means 3 requests (3*2=6, but circuit opens at 5)
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", server.URL+"/test", nil)
		resp, err := c.doWithRetry(context.Background(), req)
		// Once circuit opens, we'll get an error
		if err != nil {
			if err.Error() == "circuit breaker open: API experiencing issues, retry later" {
				break
			}
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	// Circuit should now be open
	if !c.circuitBreaker.isOpen() {
		t.Error("expected circuit breaker to be open after threshold failures")
	}

	// Next request should fail immediately without hitting the server
	beforeCallCount := callCount
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected circuit breaker error, got nil")
	}
	if err.Error() != "circuit breaker open: API experiencing issues, retry later" {
		t.Errorf("expected circuit breaker error, got: %v", err)
	}
	if callCount != beforeCallCount {
		t.Errorf("expected no additional calls when circuit is open, got %d new calls", callCount-beforeCallCount)
	}
}

// TestCircuitBreaker_resetsAfterTimeout tests that the circuit breaker closes after reset time
func TestCircuitBreaker_resetsAfterTimeout(t *testing.T) {
	cb := &circuitBreaker{}

	// Open the circuit
	for i := 0; i < CircuitBreakerThreshold; i++ {
		cb.recordFailure()
	}

	if !cb.isOpen() {
		t.Fatal("expected circuit to be open")
	}

	// Manually set lastFailure to past the reset time
	cb.mu.Lock()
	cb.lastFailure = time.Now().Add(-CircuitBreakerResetTime - 1*time.Second)
	cb.mu.Unlock()

	// Circuit should now be closed
	if cb.isOpen() {
		t.Error("expected circuit to be closed after reset time")
	}

	// Verify failures were reset
	cb.mu.Lock()
	if cb.failures != 0 {
		t.Errorf("expected failures to be reset to 0, got %d", cb.failures)
	}
	cb.mu.Unlock()
}

// TestCircuitBreaker_resetsOnSuccess tests that successful requests reset the circuit breaker
func TestCircuitBreaker_resetsOnSuccess(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "server error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make 2 failing requests
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", server.URL+"/test", nil)
		resp, err := c.doWithRetry(context.Background(), req)
		if err != nil {
			t.Fatalf("request %d: doWithRetry() error: %v", i+1, err)
		}
		_ = resp.Body.Close()
	}

	// Verify circuit is not yet open (need 5 failures)
	if c.circuitBreaker.isOpen() {
		t.Error("circuit should not be open after only 2 failures")
	}

	// Make a successful request
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("doWithRetry() error: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify circuit breaker was reset
	c.circuitBreaker.mu.Lock()
	failures := c.circuitBreaker.failures
	c.circuitBreaker.mu.Unlock()

	if failures != 0 {
		t.Errorf("expected failures to be reset to 0 after success, got %d", failures)
	}
}

// TestCircuitBreaker_tracksOnlyConsecutiveFailures tests that non-consecutive failures don't open circuit
func TestCircuitBreaker_tracksOnlyConsecutiveFailures(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Fail on odd requests, succeed on even
		if callCount%2 == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "server error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make 10 requests (5 failures, 5 successes, alternating)
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", server.URL+"/test", nil)
		resp, err := c.doWithRetry(context.Background(), req)
		if err != nil {
			t.Fatalf("request %d: doWithRetry() error: %v", i+1, err)
		}
		_ = resp.Body.Close()
	}

	// Circuit should still be closed (no consecutive failures)
	if c.circuitBreaker.isOpen() {
		t.Error("expected circuit to remain closed with non-consecutive failures")
	}
}

// TestCircuitBreaker_onlyTracksServerErrors tests that circuit breaker only tracks 5xx errors
func TestCircuitBreaker_onlyTracksServerErrors(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 5 {
			w.WriteHeader(http.StatusNotFound) // 4xx error
			_, _ = w.Write([]byte(`{"error": "not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make 5 requests that return 404 (4xx errors should not increment circuit breaker)
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", server.URL+"/test", nil)
		resp, err := c.doWithRetry(context.Background(), req)
		if err != nil {
			t.Fatalf("request %d: doWithRetry() error: %v", i+1, err)
		}
		_ = resp.Body.Close()
	}

	// Circuit should still be closed (4xx errors don't count)
	if c.circuitBreaker.isOpen() {
		t.Error("expected circuit to remain closed for 4xx errors")
	}

	// Verify failure count is 0
	c.circuitBreaker.mu.Lock()
	failures := c.circuitBreaker.failures
	c.circuitBreaker.mu.Unlock()

	if failures != 0 {
		t.Errorf("expected 0 failures for 4xx errors, got %d", failures)
	}
}

// TestNewClient_configuresConnectionPooling verifies that connection pooling settings are configured
func TestNewClient_configuresConnectionPooling(t *testing.T) {
	client, err := NewClient("test-id", "test-key")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.MaxIdleConns != MaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", transport.MaxIdleConns, MaxIdleConns)
	}

	if transport.MaxConnsPerHost != MaxConnsPerHost {
		t.Errorf("MaxConnsPerHost = %d, want %d", transport.MaxConnsPerHost, MaxConnsPerHost)
	}

	if transport.IdleConnTimeout != IdleConnTimeout {
		t.Errorf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, IdleConnTimeout)
	}
}

// TestNewClientWithAccount_configuresConnectionPooling verifies that connection pooling settings are configured
func TestNewClientWithAccount_configuresConnectionPooling(t *testing.T) {
	client, err := NewClientWithAccount("test-id", "test-key", "account-id")
	if err != nil {
		t.Fatalf("NewClientWithAccount failed: %v", err)
	}

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.MaxIdleConns != MaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", transport.MaxIdleConns, MaxIdleConns)
	}

	if transport.MaxConnsPerHost != MaxConnsPerHost {
		t.Errorf("MaxConnsPerHost = %d, want %d", transport.MaxConnsPerHost, MaxConnsPerHost)
	}

	if transport.IdleConnTimeout != IdleConnTimeout {
		t.Errorf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, IdleConnTimeout)
	}
}

// TestClient_fetchToken_wrapsErrorWithHTTPContext verifies that auth errors include HTTP context
func TestClient_fetchToken_wrapsErrorWithHTTPContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code": "invalid_credentials", "message": "Invalid API key"}`))
	}))
	defer server.Close()

	c := &Client{
		baseURL:        server.URL,
		clientID:       "test-id",
		apiKey:         "invalid-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
	}

	err := c.fetchToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify error contains HTTP context
	var contextualErr *ContextualError
	if !errors.As(err, &contextualErr) {
		t.Fatalf("expected ContextualError, got %T: %v", err, err)
	}

	if contextualErr.Method != "POST" {
		t.Errorf("Method = %q, want POST", contextualErr.Method)
	}

	if !strings.Contains(contextualErr.URL, "/api/v1/authentication/login") {
		t.Errorf("URL = %q, want to contain /api/v1/authentication/login", contextualErr.URL)
	}

	if contextualErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", contextualErr.StatusCode, http.StatusUnauthorized)
	}

	// Verify wrapped error message contains auth failure details
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("error message %q should contain 'authentication failed'", err.Error())
	}

	if !strings.Contains(err.Error(), "POST") {
		t.Errorf("error message %q should contain HTTP method", err.Error())
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error message %q should contain status code", err.Error())
	}
}
