package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// MockServer provides HTTP mocking for API tests
type MockServer struct {
	Server *httptest.Server
	Mux    *http.ServeMux
	mu     sync.Mutex
	routes map[string]map[string]http.HandlerFunc // method -> path -> handler
}

// NewMockServer creates a test server with a valid auth endpoint
func NewMockServer() *MockServer {
	mux := http.NewServeMux()
	ms := &MockServer{
		Mux:    mux,
		routes: make(map[string]map[string]http.HandlerFunc),
	}

	// Register default auth handler
	ms.Handle("POST", "/api/v1/authentication/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return a token that expires in the far future
		resp := map[string]string{
			"token":      "mock-test-token",
			"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Create server with the multiplexer
	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms.mu.Lock()
		methodRoutes, ok := ms.routes[r.Method]
		ms.mu.Unlock()

		if ok {
			if handler, found := methodRoutes[r.URL.Path]; found {
				handler(w, r)
				return
			}
		}

		// Default 404 response
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":    "not_found",
			"message": "endpoint not found",
		})
	}))

	return ms
}

// Close shuts down the server
func (m *MockServer) Close() {
	m.Server.Close()
}

// URL returns the server URL
func (m *MockServer) URL() string {
	return m.Server.URL
}

// Handle registers a handler for a path and method
func (m *MockServer) Handle(method, path string, handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.routes[method] == nil {
		m.routes[method] = make(map[string]http.HandlerFunc)
	}
	m.routes[method][path] = handler
}

// HandleJSON registers a handler that returns JSON
func (m *MockServer) HandleJSON(method, path string, statusCode int, response any) {
	m.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	})
}

// HandleError registers a handler that returns an error response
func (m *MockServer) HandleError(method, path string, statusCode int, message string) {
	m.Handle(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"code":    http.StatusText(statusCode),
			"message": message,
		})
	})
}

// Example usage in api package tests:
//
// Create a client pointing to the mock server:
//
//   ms := testutil.NewMockServer()
//   defer ms.Close()
//
//   client := &api.Client{
//       baseURL:        ms.URL(),
//       clientID:       "test-id",
//       apiKey:         "test-key",
//       httpClient:     http.DefaultClient,
//       circuitBreaker: &circuitBreaker{},
//       token: &TokenCache{
//           Token:     "test-token",
//           ExpiresAt: time.Now().Add(10 * time.Minute),
//       },
//   }
//
// See mock_integration_test.go for working examples.
