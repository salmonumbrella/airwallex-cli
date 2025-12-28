package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListGlobalAccounts_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/global_accounts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters are correctly set
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "2" {
			t.Errorf("page_num = %q, want '2'", pageNum)
		}
		if pageSize != "10" {
			t.Errorf("page_size = %q, want '10'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "ga_123",
					"account_name": "Test Account",
					"currency": "USD",
					"country_code": "US",
					"status": "ACTIVE",
					"created_at": "2024-01-01T00:00:00Z"
				}
			],
			"has_more": true
		}`))
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

	result, err := c.ListGlobalAccounts(context.Background(), 2, 10)
	if err != nil {
		t.Fatalf("ListGlobalAccounts() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if !result.HasMore {
		t.Error("has_more = false, want true")
	}
	if result.Items[0].AccountID != "ga_123" {
		t.Errorf("account_id = %q, want 'ga_123'", result.Items[0].AccountID)
	}
}

func TestListGlobalAccounts_WithoutPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/global_accounts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify no query parameters are set when values are 0 or negative
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query parameters, got: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [],
			"has_more": false
		}`))
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

	result, err := c.ListGlobalAccounts(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListGlobalAccounts() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestListGlobalAccounts_PartialPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify both page_num and page_size are set (API requires both)
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "1" {
			t.Errorf("page_num = %q, want '1'", pageNum)
		}
		if pageSize != "5" {
			t.Errorf("page_size = %q, want '5'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items": [], "has_more": false}`))
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

	_, err := c.ListGlobalAccounts(context.Background(), 0, 5)
	if err != nil {
		t.Fatalf("ListGlobalAccounts() error: %v", err)
	}
}

func TestGetGlobalAccount_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/global_accounts/ga_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "ga_123",
			"account_name": "Test Account",
			"currency": "USD",
			"country_code": "US",
			"status": "ACTIVE",
			"account_number": "1234567890",
			"routing_code": "987654321",
			"created_at": "2024-01-01T00:00:00Z"
		}`))
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

	account, err := c.GetGlobalAccount(context.Background(), "ga_123")
	if err != nil {
		t.Fatalf("GetGlobalAccount() error: %v", err)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	if account.AccountID != "ga_123" {
		t.Errorf("account_id = %q, want 'ga_123'", account.AccountID)
	}
	if account.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", account.Currency)
	}
}
