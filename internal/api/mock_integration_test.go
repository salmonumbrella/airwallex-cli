package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/salmonumbrella/airwallex-cli/internal/api/testutil"
)

// TestMockIntegration_GetBalances_Success demonstrates successful API call mocking
func TestMockIntegration_GetBalances_Success(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Mock the balances endpoint
	ms.HandleJSON("GET", "/api/v1/balances/current", http.StatusOK, []map[string]any{
		{
			"currency":         "USD",
			"available_amount": 1000.50,
			"pending_amount":   50.25,
			"reserved_amount":  25.00,
			"total_amount":     1075.75,
		},
		{
			"currency":         "EUR",
			"available_amount": 500.00,
			"pending_amount":   0.00,
			"reserved_amount":  0.00,
			"total_amount":     500.00,
		},
	})

	// Create test client pointing to mock server
	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make the API call
	result, err := client.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}

	// Verify results
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Balances) != 2 {
		t.Errorf("balances count = %d, want 2", len(result.Balances))
	}

	// Check USD balance
	usd := result.Balances[0]
	if usd.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", usd.Currency)
	}
	if usd.AvailableAmount != jn("1000.5") {
		t.Errorf("available_amount = %s, want 1000.50", usd.AvailableAmount)
	}
}

// TestMockIntegration_GetBalances_NotFound demonstrates 404 error handling
func TestMockIntegration_GetBalances_NotFound(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Mock a 404 error
	ms.HandleError("GET", "/api/v1/balances/current", http.StatusNotFound, "Account not found")

	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make the API call
	result, err := client.GetBalances(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}

	// Check that it's an API error with the expected message
	if !IsNotFoundError(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
}

// TestMockIntegration_GetBalances_ServerError demonstrates 500 error handling
func TestMockIntegration_GetBalances_ServerError(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Mock a 500 error
	ms.HandleError("GET", "/api/v1/balances/current", http.StatusInternalServerError, "Internal server error")

	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make the API call - should get retried once (GET is idempotent)
	result, err := client.GetBalances(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// TestMockIntegration_RateLimitRetry demonstrates rate limit retry behavior
func TestMockIntegration_RateLimitRetry(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	callCount := 0

	// Mock endpoint that returns 429 first, then succeeds
	ms.Handle("GET", "/api/v1/balances/current", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code": "rate_limit", "message": "Rate limit exceeded"}`))
			return
		}

		// Success on retry
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"currency": "USD", "available_amount": 100, "pending_amount": 0, "reserved_amount": 0, "total_amount": 100}]`))
	})

	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make the API call - should succeed after retry
	start := time.Now()
	result, err := client.GetBalances(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	// Verify it was called twice (1 initial + 1 retry)
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}

	// Verify there was a delay for the retry (uses test delay from init())
	if elapsed < testRateLimitBaseDelay {
		t.Errorf("elapsed = %v, expected at least %v for retry delay", elapsed, testRateLimitBaseDelay)
	}
}

// TestMockIntegration_Authentication demonstrates authentication flow
func TestMockIntegration_Authentication(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// The mock server has a default auth handler, but we can override it
	authCalled := false
	ms.Handle("POST", "/api/v1/authentication/login", func(w http.ResponseWriter, r *http.Request) {
		authCalled = true

		// Check headers
		if r.Header.Get("x-client-id") == "" {
			t.Error("x-client-id header missing")
		}
		if r.Header.Get("x-api-key") == "" {
			t.Error("x-api-key header missing")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token": "custom-token", "expires_at": "2099-01-01T00:00:00Z"}`))
	})

	// Mock the API endpoint
	ms.HandleJSON("GET", "/api/v1/balances/current", http.StatusOK, []map[string]any{
		{
			"currency":         "USD",
			"available_amount": 100.0,
			"pending_amount":   0.0,
			"reserved_amount":  0.0,
			"total_amount":     100.0,
		},
	})

	// Create client without a token (will trigger authentication)
	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		// No token - will authenticate on first request
	}

	// Make API call - should trigger authentication first
	result, err := client.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}

	// Verify authentication was called
	if !authCalled {
		t.Error("authentication was not called")
	}

	// Verify token was set
	client.tokenMu.RLock()
	hasToken := client.token != nil
	client.tokenMu.RUnlock()

	if !hasToken {
		t.Error("client token was not set after authentication")
	}
}

// TestMockIntegration_MultipleEndpoints demonstrates testing multiple endpoints
func TestMockIntegration_MultipleEndpoints(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Mock multiple endpoints
	ms.HandleJSON("GET", "/api/v1/balances/current", http.StatusOK, []map[string]any{
		{"currency": "USD", "available_amount": 100.0, "pending_amount": 0.0, "reserved_amount": 0.0, "total_amount": 100.0},
	})

	ms.HandleJSON("GET", "/api/v1/balances/history", http.StatusOK, map[string]any{
		"items":    []any{},
		"has_more": false,
	})

	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Test balances endpoint
	balances, err := client.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}
	if len(balances.Balances) != 1 {
		t.Errorf("balances count = %d, want 1", len(balances.Balances))
	}

	// Test history endpoint
	history, err := client.GetBalanceHistory(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
	if len(history.Items) != 0 {
		t.Errorf("history items = %d, want 0", len(history.Items))
	}
	if history.HasMore {
		t.Error("has_more = true, want false")
	}
}

// TestMockIntegration_CustomResponseHeaders demonstrates testing response headers
func TestMockIntegration_CustomResponseHeaders(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Custom handler that sets headers
	ms.Handle("GET", "/api/v1/balances/current", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit-Remaining", "100")
		w.Header().Set("X-Rate-Limit-Reset", "1234567890")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Get the raw response to check headers
	resp, err := client.Get(context.Background(), "/api/v1/balances/current")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer closeBody(resp)

	// Check custom headers
	if rateLimitRemaining := resp.Header.Get("X-Rate-Limit-Remaining"); rateLimitRemaining != "100" {
		t.Errorf("X-Rate-Limit-Remaining = %q, want '100'", rateLimitRemaining)
	}
	if rateLimitReset := resp.Header.Get("X-Rate-Limit-Reset"); rateLimitReset != "1234567890" {
		t.Errorf("X-Rate-Limit-Reset = %q, want '1234567890'", rateLimitReset)
	}
}

// TestMockIntegration_RequestValidation demonstrates validating request data
func TestMockIntegration_RequestValidation(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Custom handler that validates request
	ms.Handle("GET", "/api/v1/balances/current", func(w http.ResponseWriter, r *http.Request) {
		// Check authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Authorization = %q, want 'Bearer test-token'", authHeader)
		}

		// Check API version header
		apiVersion := r.Header.Get("x-api-version")
		if apiVersion != APIVersion {
			t.Errorf("x-api-version = %q, want %q", apiVersion, APIVersion)
		}

		// Check Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q, want 'application/json'", contentType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	client := &Client{
		baseURL:        ms.URL(),
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	// Make the call - handler will verify headers
	_, err := client.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}
}
