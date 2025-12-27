package testutil_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/api/testutil"
)

// TestMockServer_BasicUsage demonstrates basic mock server usage
func TestMockServer_BasicUsage(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Verify server is running
	if ms.URL() == "" {
		t.Fatal("expected non-empty server URL")
	}

	// Verify we can make requests to it
	resp, err := http.Get(ms.URL() + "/nonexistent")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// TestMockServer_HandleJSON demonstrates JSON response mocking
func TestMockServer_HandleJSON(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Register a JSON handler
	ms.HandleJSON("GET", "/api/v1/test", http.StatusOK, map[string]string{
		"message": "success",
		"data":    "test-value",
	})

	// Make request
	resp, err := http.Get(ms.URL() + "/api/v1/test")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// TestMockServer_HandleError demonstrates error response mocking
func TestMockServer_HandleError(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Register an error handler
	ms.HandleError("GET", "/api/v1/error", http.StatusBadRequest, "Invalid request")

	// Make request
	resp, err := http.Get(ms.URL() + "/api/v1/error")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

// TestMockServer_CustomHandler demonstrates custom handler registration
func TestMockServer_CustomHandler(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	callCount := 0

	// Register custom handler that counts calls
	ms.Handle("POST", "/api/v1/custom", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"count": ` + string(rune(callCount)) + `}`))
	})

	// Make first request
	resp1, err := http.Post(ms.URL()+"/api/v1/custom", "application/json", nil)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	_ = resp1.Body.Close()

	// Make second request
	resp2, err := http.Post(ms.URL()+"/api/v1/custom", "application/json", nil)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	_ = resp2.Body.Close()

	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestMockServer_DefaultAuthHandler demonstrates the built-in auth handler
func TestMockServer_DefaultAuthHandler(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// The mock server automatically handles authentication
	resp, err := http.Post(ms.URL()+"/api/v1/authentication/login", "application/json", nil)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// Note: The following test demonstrates the pattern for creating test clients.
// This test is in the testutil_test package so it doesn't have access to
// unexported api.Client fields. See the example tests in api package for
// the actual working implementation.

// TestMockServer_WithAPIClient demonstrates how to use mock server with API client
func TestMockServer_WithAPIClient(t *testing.T) {
	t.Skip("This is a documentation example - see api package tests for working implementation")

	ms := testutil.NewMockServer()
	defer ms.Close()

	// Register mock endpoint
	ms.HandleJSON("GET", "/api/v1/balances/current", http.StatusOK, []map[string]any{
		{
			"currency":         "USD",
			"available_amount": 1000.50,
			"pending_amount":   50.25,
			"reserved_amount":  25.00,
			"total_amount":     1075.75,
		},
	})

	// In api package tests, you would create a client like this:
	//
	// client := &api.Client{
	//     baseURL:        ms.URL(),
	//     clientID:       "test-id",
	//     apiKey:         "test-key",
	//     httpClient:     http.DefaultClient,
	//     circuitBreaker: &circuitBreaker{},
	//     token: &TokenCache{
	//         Token:     "test-token",
	//         ExpiresAt: time.Now().Add(10 * time.Minute),
	//     },
	// }
	//
	// result, err := client.GetBalances(context.Background())
	// ...

	// For this example, we'll just create a regular client
	// (which won't work with our mock server due to baseURL)
	client, err := api.NewClient("test-id", "test-key")
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	// This will fail because client.baseURL points to real API
	_, _ = client.GetBalances(context.Background())

	// See api/mock_integration_test.go for working examples
}

// TestMockServer_MultipleEndpoints demonstrates handling multiple endpoints
func TestMockServer_MultipleEndpoints(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Register multiple endpoints
	ms.HandleJSON("GET", "/api/v1/users", http.StatusOK, []string{"user1", "user2"})
	ms.HandleJSON("GET", "/api/v1/accounts", http.StatusOK, []string{"acc1", "acc2"})
	ms.HandleError("GET", "/api/v1/forbidden", http.StatusForbidden, "Access denied")

	// Test each endpoint
	tests := []struct {
		path       string
		wantStatus int
	}{
		{"/api/v1/users", http.StatusOK},
		{"/api/v1/accounts", http.StatusOK},
		{"/api/v1/forbidden", http.StatusForbidden},
		{"/api/v1/notfound", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := http.Get(ms.URL() + tt.path)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

// TestMockServer_ConcurrentRequests tests thread safety
func TestMockServer_ConcurrentRequests(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.HandleJSON("GET", "/api/v1/concurrent", http.StatusOK, map[string]string{
		"status": "ok",
	})

	// Make concurrent requests
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			resp, err := http.Get(ms.URL() + "/api/v1/concurrent")
			if err != nil {
				t.Errorf("request failed: %v", err)
			} else {
				_ = resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("status = %d, want 200", resp.StatusCode)
				}
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMockServer_MethodRouting tests that different methods are routed correctly
func TestMockServer_MethodRouting(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Register same path with different methods
	ms.HandleJSON("GET", "/api/v1/resource", http.StatusOK, map[string]string{"action": "read"})
	ms.HandleJSON("POST", "/api/v1/resource", http.StatusCreated, map[string]string{"action": "create"})
	ms.HandleJSON("PUT", "/api/v1/resource", http.StatusOK, map[string]string{"action": "update"})
	ms.HandleJSON("DELETE", "/api/v1/resource", http.StatusNoContent, nil)

	tests := []struct {
		method     string
		wantStatus int
	}{
		{"GET", http.StatusOK},
		{"POST", http.StatusCreated},
		{"PUT", http.StatusOK},
		{"DELETE", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ms.URL()+"/api/v1/resource", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}
