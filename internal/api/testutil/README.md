# API Test Utilities

This package provides HTTP mocking utilities for testing the Airwallex API client.

## MockServer

`MockServer` is a lightweight HTTP mock server that makes it easy to test API client behavior with mocked responses.

### Features

- Automatic authentication endpoint handling
- Simple JSON response mocking
- Custom error response mocking
- Custom handler registration
- Thread-safe concurrent request handling
- Method-based routing (GET, POST, PUT, DELETE, etc.)

### Quick Start

```go
package api

import (
    "testing"
    "github.com/salmonumbrella/airwallex-cli/internal/api/testutil"
)

func TestYourAPI(t *testing.T) {
    // Create mock server
    ms := testutil.NewMockServer()
    defer ms.Close()

    // Mock an endpoint
    ms.HandleJSON("GET", "/api/v1/balances/current", http.StatusOK, []Balance{
        {Currency: "USD", AvailableAmount: 1000.50},
    })

    // Create test client
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

    // Test your API call
    result, err := client.GetBalances(context.Background())
    // ...assertions...
}
```

### API Reference

#### Creating a Mock Server

```go
ms := testutil.NewMockServer()
defer ms.Close()
```

The server automatically includes a default authentication handler for `/api/v1/authentication/login`.

#### Mocking JSON Responses

```go
ms.HandleJSON("GET", "/api/v1/resource", http.StatusOK, map[string]string{
    "id": "123",
    "name": "test",
})
```

#### Mocking Error Responses

```go
ms.HandleError("GET", "/api/v1/error", http.StatusNotFound, "Resource not found")
```

#### Custom Handlers

For more complex scenarios, register a custom handler:

```go
callCount := 0
ms.Handle("POST", "/api/v1/custom", func(w http.ResponseWriter, r *http.Request) {
    callCount++

    // Read request body
    body, _ := io.ReadAll(r.Body)

    // Validate headers
    if r.Header.Get("Authorization") == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // Return custom response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]int{"count": callCount})
})
```

#### Getting Server URL

```go
url := ms.URL()  // e.g., "http://127.0.0.1:54321"
```

### Example Tests

See the following files for complete examples:

- **`mock_server_test.go`** - Tests demonstrating mock server features
- **`../mock_integration_test.go`** - Full integration tests with API client

### Common Patterns

#### Testing Success Cases

```go
func TestAPI_Success(t *testing.T) {
    ms := testutil.NewMockServer()
    defer ms.Close()

    ms.HandleJSON("GET", "/api/v1/data", http.StatusOK, map[string]string{
        "result": "success",
    })

    client := createTestClient(ms.URL())
    result, err := client.GetData(context.Background())

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // ...assertions...
}
```

#### Testing Error Handling

```go
func TestAPI_NotFound(t *testing.T) {
    ms := testutil.NewMockServer()
    defer ms.Close()

    ms.HandleError("GET", "/api/v1/data", http.StatusNotFound, "Not found")

    client := createTestClient(ms.URL())
    _, err := client.GetData(context.Background())

    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if !api.IsNotFoundError(err) {
        t.Errorf("expected not found error, got: %v", err)
    }
}
```

#### Testing Rate Limit Retry

```go
func TestAPI_RateLimitRetry(t *testing.T) {
    ms := testutil.NewMockServer()
    defer ms.Close()

    callCount := 0
    ms.Handle("GET", "/api/v1/data", func(w http.ResponseWriter, r *http.Request) {
        callCount++
        if callCount == 1 {
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"result": "success"})
    })

    client := createTestClient(ms.URL())
    result, err := client.GetData(context.Background())

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if callCount != 2 {
        t.Errorf("expected 2 calls (1 retry), got %d", callCount)
    }
}
```

#### Testing Authentication

```go
func TestAPI_Authentication(t *testing.T) {
    ms := testutil.NewMockServer()
    defer ms.Close()

    authCalled := false
    ms.Handle("POST", "/api/v1/authentication/login", func(w http.ResponseWriter, r *http.Request) {
        authCalled = true

        // Verify credentials
        if r.Header.Get("x-client-id") != "test-id" {
            t.Error("invalid client ID")
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "token": "test-token",
            "expires_at": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
        })
    })

    // Client without token (will authenticate)
    client := &Client{
        baseURL:    ms.URL(),
        clientID:   "test-id",
        apiKey:     "test-key",
        httpClient: http.DefaultClient,
    }

    // Make request - should trigger auth
    _, err := client.Get(context.Background(), "/api/v1/test")

    if !authCalled {
        t.Error("authentication was not called")
    }
}
```

### Thread Safety

The mock server is thread-safe and can handle concurrent requests:

```go
func TestAPI_Concurrent(t *testing.T) {
    ms := testutil.NewMockServer()
    defer ms.Close()

    ms.HandleJSON("GET", "/api/v1/data", http.StatusOK, map[string]string{
        "result": "ok",
    })

    // Make concurrent requests
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            client := createTestClient(ms.URL())
            _, err := client.GetData(context.Background())
            if err != nil {
                t.Errorf("request failed: %v", err)
            }
        }()
    }
    wg.Wait()
}
```

### Best Practices

1. **Always defer Close()** - Ensure the server is cleaned up after tests
2. **Use table-driven tests** - Test multiple scenarios with the same setup
3. **Validate request data** - Use custom handlers to verify headers, body, etc.
4. **Test error paths** - Don't just test happy paths
5. **Test retry behavior** - Verify rate limits, server errors are handled correctly
6. **Keep tests focused** - One test should verify one behavior

### Limitations

- The mock server runs in-process (not suitable for external client testing)
- No automatic response recording/playback (must manually define responses)
- No built-in request matching beyond method + path

For more complex scenarios, consider using the `Handle()` method with custom logic.
