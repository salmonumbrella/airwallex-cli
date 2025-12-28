package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetBalances_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/current" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"currency": "USD",
				"available_amount": 1000.50,
				"pending_amount": 50.25,
				"reserved_amount": 25.00,
				"total_amount": 1075.75
			},
			{
				"currency": "EUR",
				"available_amount": 500.00,
				"pending_amount": 0.00,
				"reserved_amount": 0.00,
				"total_amount": 500.00
			}
		]`))
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

	result, err := c.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Balances) != 2 {
		t.Errorf("balances count = %d, want 2", len(result.Balances))
	}

	// Verify first balance (USD)
	usd := result.Balances[0]
	if usd.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", usd.Currency)
	}
	if usd.AvailableAmount != 1000.50 {
		t.Errorf("available_amount = %f, want 1000.50", usd.AvailableAmount)
	}
	if usd.PendingAmount != 50.25 {
		t.Errorf("pending_amount = %f, want 50.25", usd.PendingAmount)
	}
	if usd.ReservedAmount != 25.00 {
		t.Errorf("reserved_amount = %f, want 25.00", usd.ReservedAmount)
	}
	if usd.TotalAmount != 1075.75 {
		t.Errorf("total_amount = %f, want 1075.75", usd.TotalAmount)
	}

	// Verify second balance (EUR)
	eur := result.Balances[1]
	if eur.Currency != "EUR" {
		t.Errorf("currency = %q, want 'EUR'", eur.Currency)
	}
	if eur.AvailableAmount != 500.00 {
		t.Errorf("available_amount = %f, want 500.00", eur.AvailableAmount)
	}
	if eur.TotalAmount != 500.00 {
		t.Errorf("total_amount = %f, want 500.00", eur.TotalAmount)
	}
}

func TestGetBalances_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/current" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
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

	result, err := c.GetBalances(context.Background())
	if err != nil {
		t.Fatalf("GetBalances() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Balances) != 0 {
		t.Errorf("balances count = %d, want 0", len(result.Balances))
	}
}

func TestGetBalances_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{
			"code": "unauthorized",
			"message": "Invalid authentication credentials"
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

	result, err := c.GetBalances(context.Background())
	if err == nil {
		t.Error("expected error for unauthorized response, got nil")
	}
	if result != nil {
		t.Error("expected nil result for error response")
	}
}

func TestGetBalances_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json}`))
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

	result, err := c.GetBalances(context.Background())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid JSON")
	}
}

func TestGetBalanceHistory_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}

		// Verify no query parameters for default call
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query parameters, got: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "txn_123",
					"currency": "USD",
					"amount": 100.00,
					"balance": 1000.00,
					"transaction_type": "DEPOSIT",
					"created_at": "2024-01-01T00:00:00Z",
					"description": "Test deposit"
				},
				{
					"id": "txn_456",
					"currency": "USD",
					"amount": -50.00,
					"balance": 950.00,
					"transaction_type": "WITHDRAWAL",
					"created_at": "2024-01-02T00:00:00Z"
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

	result, err := c.GetBalanceHistory(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 2 {
		t.Errorf("items count = %d, want 2", len(result.Items))
	}
	if !result.HasMore {
		t.Error("has_more = false, want true")
	}

	// Verify first item
	item1 := result.Items[0]
	if item1.ID != "txn_123" {
		t.Errorf("id = %q, want 'txn_123'", item1.ID)
	}
	if item1.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", item1.Currency)
	}
	if item1.Amount != 100.00 {
		t.Errorf("amount = %f, want 100.00", item1.Amount)
	}
	if item1.Balance != 1000.00 {
		t.Errorf("balance = %f, want 1000.00", item1.Balance)
	}
	if item1.TransactionType != "DEPOSIT" {
		t.Errorf("transaction_type = %q, want 'DEPOSIT'", item1.TransactionType)
	}
	if item1.Description != "Test deposit" {
		t.Errorf("description = %q, want 'Test deposit'", item1.Description)
	}

	// Verify second item
	item2 := result.Items[1]
	if item2.ID != "txn_456" {
		t.Errorf("id = %q, want 'txn_456'", item2.ID)
	}
	if item2.Amount != -50.00 {
		t.Errorf("amount = %f, want -50.00", item2.Amount)
	}
	if item2.TransactionType != "WITHDRAWAL" {
		t.Errorf("transaction_type = %q, want 'WITHDRAWAL'", item2.TransactionType)
	}
}

func TestGetBalanceHistory_WithCurrencyFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify currency parameter
		currency := r.URL.Query().Get("currency")
		if currency != "EUR" {
			t.Errorf("currency = %q, want 'EUR'", currency)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "txn_eur",
					"currency": "EUR",
					"amount": 200.00,
					"balance": 500.00,
					"transaction_type": "DEPOSIT",
					"created_at": "2024-01-01T00:00:00Z"
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

	result, err := c.GetBalanceHistory(context.Background(), "EUR", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Currency != "EUR" {
		t.Errorf("currency = %q, want 'EUR'", result.Items[0].Currency)
	}
}

func TestGetBalanceHistory_WithDateFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify date parameters
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		if from != "2024-01-01T00:00:00Z" {
			t.Errorf("from = %q, want '2024-01-01T00:00:00Z'", from)
		}
		if to != "2024-01-31T23:59:59Z" {
			t.Errorf("to = %q, want '2024-01-31T23:59:59Z'", to)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "txn_filtered",
					"currency": "USD",
					"amount": 75.00,
					"balance": 800.00,
					"transaction_type": "DEPOSIT",
					"created_at": "2024-01-15T12:00:00Z"
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

	result, err := c.GetBalanceHistory(context.Background(), "", "2024-01-01T00:00:00Z", "2024-01-31T23:59:59Z", 0, 0)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
}

func TestGetBalanceHistory_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify pagination parameters
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
					"id": "txn_page2",
					"currency": "USD",
					"amount": 25.00,
					"balance": 700.00,
					"transaction_type": "DEPOSIT",
					"created_at": "2024-01-20T00:00:00Z"
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

	result, err := c.GetBalanceHistory(context.Background(), "", "", "", 2, 10)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
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
}

func TestGetBalanceHistory_AllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify all parameters are set
		query := r.URL.Query()
		if query.Get("currency") != "GBP" {
			t.Errorf("currency = %q, want 'GBP'", query.Get("currency"))
		}
		if query.Get("from") != "2024-01-01T00:00:00Z" {
			t.Errorf("from = %q, want '2024-01-01T00:00:00Z'", query.Get("from"))
		}
		if query.Get("to") != "2024-12-31T23:59:59Z" {
			t.Errorf("to = %q, want '2024-12-31T23:59:59Z'", query.Get("to"))
		}
		if query.Get("page_num") != "1" {
			t.Errorf("page_num = %q, want '1'", query.Get("page_num"))
		}
		if query.Get("page_size") != "50" {
			t.Errorf("page_size = %q, want '50'", query.Get("page_size"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "txn_gbp",
					"currency": "GBP",
					"amount": 150.00,
					"balance": 600.00,
					"transaction_type": "DEPOSIT",
					"created_at": "2024-06-15T12:00:00Z"
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

	result, err := c.GetBalanceHistory(context.Background(), "GBP", "2024-01-01T00:00:00Z", "2024-12-31T23:59:59Z", 1, 50)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Currency != "GBP" {
		t.Errorf("currency = %q, want 'GBP'", result.Items[0].Currency)
	}
}

func TestGetBalanceHistory_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/balances/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
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

	result, err := c.GetBalanceHistory(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
	if result.HasMore {
		t.Error("has_more = true, want false")
	}
}

func TestGetBalanceHistory_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_request",
			"message": "Invalid date format"
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

	result, err := c.GetBalanceHistory(context.Background(), "", "invalid-date", "", 0, 0)
	if err == nil {
		t.Error("expected error for bad request, got nil")
	}
	if result != nil {
		t.Error("expected nil result for error response")
	}
}

func TestGetBalanceHistory_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json}`))
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

	result, err := c.GetBalanceHistory(context.Background(), "", "", "", 0, 0)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid JSON")
	}
}

func TestGetBalanceHistory_PartialPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify both page_num and page_size are set (API requires both)
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "1" {
			t.Errorf("page_num = %q, want '1'", pageNum)
		}
		if pageSize != "20" {
			t.Errorf("page_size = %q, want '20'", pageSize)
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

	_, err := c.GetBalanceHistory(context.Background(), "", "", "", 0, 20)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
}

func TestGetBalanceHistory_ZeroAmounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "txn_zero",
					"currency": "USD",
					"amount": 0.00,
					"balance": 0.00,
					"transaction_type": "ADJUSTMENT",
					"created_at": "2024-01-01T00:00:00Z"
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

	result, err := c.GetBalanceHistory(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetBalanceHistory() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Amount != 0.00 {
		t.Errorf("amount = %f, want 0.00", result.Items[0].Amount)
	}
	if result.Items[0].Balance != 0.00 {
		t.Errorf("balance = %f, want 0.00", result.Items[0].Balance)
	}
}
