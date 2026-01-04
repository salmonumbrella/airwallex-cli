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

func TestGetGlobalAccount_InvalidID(t *testing.T) {
	c := &Client{
		baseURL:        "http://test.example.com",
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},
		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetGlobalAccount(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty account ID, got nil")
	}

	_, err = c.GetGlobalAccount(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid account ID, got nil")
	}
}

func TestGetGlobalAccount_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Account not found"
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

	_, err := c.GetGlobalAccount(context.Background(), "ga_nonexistent")
	if err == nil {
		t.Error("expected error for not found account, got nil")
	}
}

func TestGetGlobalAccount_AllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/global_accounts/ga_456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "ga_456",
			"account_name": "EUR Account",
			"currency": "EUR",
			"country_code": "DE",
			"status": "ACTIVE",
			"account_number": "DE89370400440532013000",
			"routing_code": "COBADEFFXXX",
			"iban": "DE89370400440532013000",
			"swift_code": "COBADEFFXXX",
			"created_at": "2024-03-15T10:30:00Z"
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

	account, err := c.GetGlobalAccount(context.Background(), "ga_456")
	if err != nil {
		t.Fatalf("GetGlobalAccount() error: %v", err)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	if account.AccountID != "ga_456" {
		t.Errorf("account_id = %q, want 'ga_456'", account.AccountID)
	}
	if account.AccountName != "EUR Account" {
		t.Errorf("account_name = %q, want 'EUR Account'", account.AccountName)
	}
	if account.Currency != "EUR" {
		t.Errorf("currency = %q, want 'EUR'", account.Currency)
	}
	if account.CountryCode != "DE" {
		t.Errorf("country_code = %q, want 'DE'", account.CountryCode)
	}
	if account.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", account.Status)
	}
	if account.IBAN != "DE89370400440532013000" {
		t.Errorf("iban = %q, want 'DE89370400440532013000'", account.IBAN)
	}
	if account.SwiftCode != "COBADEFFXXX" {
		t.Errorf("swift_code = %q, want 'COBADEFFXXX'", account.SwiftCode)
	}
}

func TestListGlobalAccounts_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{
			"code": "internal_error",
			"message": "Internal server error"
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

	_, err := c.ListGlobalAccounts(context.Background(), 1, 10)
	if err == nil {
		t.Error("expected error for server error, got nil")
	}
}

func TestListGlobalAccounts_MultipleItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "ga_001",
					"account_name": "USD Account",
					"currency": "USD",
					"country_code": "US",
					"status": "ACTIVE",
					"created_at": "2024-01-01T00:00:00Z"
				},
				{
					"id": "ga_002",
					"account_name": "EUR Account",
					"currency": "EUR",
					"country_code": "DE",
					"status": "ACTIVE",
					"created_at": "2024-01-02T00:00:00Z"
				},
				{
					"id": "ga_003",
					"account_name": "GBP Account",
					"currency": "GBP",
					"country_code": "GB",
					"status": "INACTIVE",
					"created_at": "2024-01-03T00:00:00Z"
				}
			],
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
	if len(result.Items) != 3 {
		t.Errorf("items count = %d, want 3", len(result.Items))
	}

	// Verify each account
	expectedCurrencies := []string{"USD", "EUR", "GBP"}
	for i, currency := range expectedCurrencies {
		if result.Items[i].Currency != currency {
			t.Errorf("items[%d].currency = %q, want %q", i, result.Items[i].Currency, currency)
		}
	}

	// Verify status
	if result.Items[2].Status != "INACTIVE" {
		t.Errorf("items[2].status = %q, want 'INACTIVE'", result.Items[2].Status)
	}
}
