package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetRates tests rate retrieval with various currency pair combinations
func TestGetRates_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fx/rates/current" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters
		sellCurrency := r.URL.Query().Get("sell_currency")
		buyCurrency := r.URL.Query().Get("buy_currency")

		if sellCurrency != "USD" {
			t.Errorf("sell_currency = %q, want 'USD'", sellCurrency)
		}
		if buyCurrency != "EUR" {
			t.Errorf("buy_currency = %q, want 'EUR'", buyCurrency)
		}

		// API returns a single rate object, not an array
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sell_currency": "USD",
			"buy_currency": "EUR",
			"rate": 0.85,
			"rate_type": "SPOT"
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

	result, err := c.GetRates(context.Background(), "USD", "EUR")
	if err != nil {
		t.Fatalf("GetRates() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Rates) != 1 {
		t.Errorf("rates count = %d, want 1", len(result.Rates))
	}
	if result.Rates[0].SellCurrency != "USD" {
		t.Errorf("sell_currency = %q, want 'USD'", result.Rates[0].SellCurrency)
	}
	if result.Rates[0].BuyCurrency != "EUR" {
		t.Errorf("buy_currency = %q, want 'EUR'", result.Rates[0].BuyCurrency)
	}
	if result.Rates[0].Rate != 0.85 {
		t.Errorf("rate = %f, want 0.85", result.Rates[0].Rate)
	}
	if result.Rates[0].RateType != "SPOT" {
		t.Errorf("rate_type = %q, want 'SPOT'", result.Rates[0].RateType)
	}
}

func TestGetRates_ArrayResponse(t *testing.T) {
	// Test that the client can handle array response format (fallback)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fx/rates/current" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Return array format (some API versions may return this)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"rates": [
				{
					"sell_currency": "USD",
					"buy_currency": "EUR",
					"rate": 0.85,
					"rate_type": "SPOT"
				},
				{
					"sell_currency": "USD",
					"buy_currency": "GBP",
					"rate": 0.73,
					"rate_type": "SPOT"
				}
			]
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

	result, err := c.GetRates(context.Background(), "USD", "EUR")
	if err != nil {
		t.Fatalf("GetRates() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Rates) != 2 {
		t.Errorf("rates count = %d, want 2", len(result.Rates))
	}
}

func TestGetRates_OnlySellCurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify only sell_currency is set
		sellCurrency := r.URL.Query().Get("sell_currency")
		buyCurrency := r.URL.Query().Get("buy_currency")

		if sellCurrency != "USD" {
			t.Errorf("sell_currency = %q, want 'USD'", sellCurrency)
		}
		if buyCurrency != "" {
			t.Errorf("buy_currency = %q, want empty", buyCurrency)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rates": []}`))
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

	_, err := c.GetRates(context.Background(), "USD", "")
	if err != nil {
		t.Fatalf("GetRates() error: %v", err)
	}
}

func TestGetRates_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_currency",
			"message": "Invalid currency code"
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

	_, err := c.GetRates(context.Background(), "XXX", "YYY")
	if err == nil {
		t.Error("expected error for invalid currency, got nil")
	}
}

// TestCreateQuote tests quote creation with various amount configurations
func TestCreateQuote_WithSellAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fx/quotes/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"quote_id": "quote_123",
			"sell_currency": "USD",
			"buy_currency": "EUR",
			"sell_amount": 1000.00,
			"buy_amount": 850.00,
			"client_rate": 0.85,
			"valid_to_at": "2024-01-01T12:05:00Z",
			"validity": "MIN_5",
			"status": "ACTIVE",
			"valid_from_at": "2024-01-01T12:00:00Z"
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

	req := map[string]interface{}{
		"sell_currency": "USD",
		"buy_currency":  "EUR",
		"sell_amount":   1000.00,
	}

	quote, err := c.CreateQuote(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateQuote() error: %v", err)
	}
	if quote == nil {
		t.Fatal("quote is nil")
	}
	if quote.ID != "quote_123" {
		t.Errorf("id = %q, want 'quote_123'", quote.ID)
	}
	if quote.SellCurrency != "USD" {
		t.Errorf("sell_currency = %q, want 'USD'", quote.SellCurrency)
	}
	if quote.BuyCurrency != "EUR" {
		t.Errorf("buy_currency = %q, want 'EUR'", quote.BuyCurrency)
	}
	if quote.SellAmount != 1000.00 {
		t.Errorf("sell_amount = %f, want 1000.00", quote.SellAmount)
	}
	if quote.BuyAmount != 850.00 {
		t.Errorf("buy_amount = %f, want 850.00", quote.BuyAmount)
	}
	if quote.Rate != 0.85 {
		t.Errorf("rate = %f, want 0.85", quote.Rate)
	}
	if quote.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", quote.Status)
	}
}

func TestCreateQuote_WithBuyAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"quote_id": "quote_456",
			"sell_currency": "GBP",
			"buy_currency": "USD",
			"sell_amount": 730.00,
			"buy_amount": 1000.00,
			"client_rate": 1.37,
			"valid_to_at": "2024-01-01T12:05:00Z",
			"validity": "MIN_5",
			"status": "ACTIVE",
			"valid_from_at": "2024-01-01T12:00:00Z"
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

	req := map[string]interface{}{
		"sell_currency": "GBP",
		"buy_currency":  "USD",
		"buy_amount":    1000.00,
	}

	quote, err := c.CreateQuote(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateQuote() error: %v", err)
	}
	if quote == nil {
		t.Fatal("quote is nil")
	}
	if quote.ID != "quote_456" {
		t.Errorf("id = %q, want 'quote_456'", quote.ID)
	}
	if quote.BuyAmount != 1000.00 {
		t.Errorf("buy_amount = %f, want 1000.00", quote.BuyAmount)
	}
}

func TestCreateQuote_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_amount",
			"message": "Amount must be greater than zero"
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

	req := map[string]interface{}{
		"sell_currency": "USD",
		"buy_currency":  "EUR",
		"sell_amount":   0,
	}

	_, err := c.CreateQuote(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid amount, got nil")
	}
}

// TestGetQuote tests quote retrieval
func TestGetQuote_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fx/quotes/quote_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"quote_id": "quote_123",
			"sell_currency": "USD",
			"buy_currency": "EUR",
			"sell_amount": 1000.00,
			"buy_amount": 850.00,
			"client_rate": 0.85,
			"valid_to_at": "2024-01-01T12:05:00Z",
			"validity": "MIN_5",
			"status": "ACTIVE",
			"valid_from_at": "2024-01-01T12:00:00Z"
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

	quote, err := c.GetQuote(context.Background(), "quote_123")
	if err != nil {
		t.Fatalf("GetQuote() error: %v", err)
	}
	if quote == nil {
		t.Fatal("quote is nil")
	}
	if quote.ID != "quote_123" {
		t.Errorf("id = %q, want 'quote_123'", quote.ID)
	}
	if quote.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", quote.Status)
	}
}

func TestGetQuote_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Quote not found"
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

	_, err := c.GetQuote(context.Background(), "quote_nonexistent")
	if err == nil {
		t.Error("expected error for not found quote, got nil")
	}
}

func TestGetQuote_InvalidID(t *testing.T) {
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

	_, err := c.GetQuote(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty quote ID, got nil")
	}

	_, err = c.GetQuote(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid quote ID, got nil")
	}
}

// TestListConversions tests conversion listing with pagination and filters
func TestListConversions_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fx/conversions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters
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
					"id": "conv_123",
					"quote_id": "quote_123",
					"sell_currency": "USD",
					"buy_currency": "EUR",
					"sell_amount": 1000.00,
					"buy_amount": 850.00,
					"rate": 0.85,
					"status": "COMPLETED",
					"created_at": "2024-01-01T12:00:00Z"
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

	result, err := c.ListConversions(context.Background(), "", "", "", 2, 10)
	if err != nil {
		t.Fatalf("ListConversions() error: %v", err)
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
	if result.Items[0].ID != "conv_123" {
		t.Errorf("id = %q, want 'conv_123'", result.Items[0].ID)
	}
}

func TestListConversions_WithStatusFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify status filter
		status := r.URL.Query().Get("status")
		if status != "COMPLETED" {
			t.Errorf("status = %q, want 'COMPLETED'", status)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "conv_456",
					"sell_currency": "USD",
					"buy_currency": "EUR",
					"sell_amount": 500.00,
					"buy_amount": 425.00,
					"rate": 0.85,
					"status": "COMPLETED",
					"created_at": "2024-01-01T12:00:00Z"
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

	result, err := c.ListConversions(context.Background(), "COMPLETED", "", "", 0, 0)
	if err != nil {
		t.Fatalf("ListConversions() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Status != "COMPLETED" {
		t.Errorf("status = %q, want 'COMPLETED'", result.Items[0].Status)
	}
}

func TestListConversions_WithDateFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify date filters
		fromDate := r.URL.Query().Get("from_created_at")
		toDate := r.URL.Query().Get("to_created_at")

		if fromDate != "2024-01-01T00:00:00Z" {
			t.Errorf("from_created_at = %q, want '2024-01-01T00:00:00Z'", fromDate)
		}
		if toDate != "2024-01-31T23:59:59Z" {
			t.Errorf("to_created_at = %q, want '2024-01-31T23:59:59Z'", toDate)
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

	_, err := c.ListConversions(context.Background(), "", "2024-01-01T00:00:00Z", "2024-01-31T23:59:59Z", 0, 0)
	if err != nil {
		t.Fatalf("ListConversions() error: %v", err)
	}
}

func TestListConversions_NoFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no query parameters when all filters are empty/zero
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

	result, err := c.ListConversions(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("ListConversions() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestListConversions_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_date_format",
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

	_, err := c.ListConversions(context.Background(), "", "invalid-date", "", 0, 0)
	if err == nil {
		t.Error("expected error for invalid date format, got nil")
	}
}

// TestGetConversion tests conversion retrieval
func TestGetConversion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fx/conversions/conv_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "conv_123",
			"quote_id": "quote_123",
			"sell_currency": "USD",
			"buy_currency": "EUR",
			"sell_amount": 1000.00,
			"buy_amount": 850.00,
			"rate": 0.85,
			"status": "COMPLETED",
			"created_at": "2024-01-01T12:00:00Z"
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

	conv, err := c.GetConversion(context.Background(), "conv_123")
	if err != nil {
		t.Fatalf("GetConversion() error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversion is nil")
	}
	if conv.ID != "conv_123" {
		t.Errorf("id = %q, want 'conv_123'", conv.ID)
	}
	if conv.QuoteID != "quote_123" {
		t.Errorf("quote_id = %q, want 'quote_123'", conv.QuoteID)
	}
	if conv.SellCurrency != "USD" {
		t.Errorf("sell_currency = %q, want 'USD'", conv.SellCurrency)
	}
	if conv.BuyCurrency != "EUR" {
		t.Errorf("buy_currency = %q, want 'EUR'", conv.BuyCurrency)
	}
	if conv.SellAmount != 1000.00 {
		t.Errorf("sell_amount = %f, want 1000.00", conv.SellAmount)
	}
	if conv.BuyAmount != 850.00 {
		t.Errorf("buy_amount = %f, want 850.00", conv.BuyAmount)
	}
	if conv.Rate != 0.85 {
		t.Errorf("rate = %f, want 0.85", conv.Rate)
	}
	if conv.Status != "COMPLETED" {
		t.Errorf("status = %q, want 'COMPLETED'", conv.Status)
	}
}

func TestGetConversion_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Conversion not found"
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

	_, err := c.GetConversion(context.Background(), "conv_nonexistent")
	if err == nil {
		t.Error("expected error for not found conversion, got nil")
	}
}

func TestGetConversion_InvalidID(t *testing.T) {
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

	_, err := c.GetConversion(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty conversion ID, got nil")
	}

	_, err = c.GetConversion(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid conversion ID, got nil")
	}
}

// TestCreateConversion tests conversion creation
func TestCreateConversion_FromQuote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fx/conversions/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "conv_new",
			"quote_id": "quote_123",
			"sell_currency": "USD",
			"buy_currency": "EUR",
			"sell_amount": 1000.00,
			"buy_amount": 850.00,
			"rate": 0.85,
			"status": "COMPLETED",
			"created_at": "2024-01-01T12:00:00Z"
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

	req := map[string]interface{}{
		"quote_id": "quote_123",
	}

	conv, err := c.CreateConversion(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateConversion() error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversion is nil")
	}
	if conv.ID != "conv_new" {
		t.Errorf("id = %q, want 'conv_new'", conv.ID)
	}
	if conv.QuoteID != "quote_123" {
		t.Errorf("quote_id = %q, want 'quote_123'", conv.QuoteID)
	}
	if conv.Status != "COMPLETED" {
		t.Errorf("status = %q, want 'COMPLETED'", conv.Status)
	}
}

func TestCreateConversion_DirectAmounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "conv_direct",
			"sell_currency": "GBP",
			"buy_currency": "USD",
			"sell_amount": 730.00,
			"buy_amount": 1000.00,
			"rate": 1.37,
			"status": "COMPLETED",
			"created_at": "2024-01-01T12:00:00Z"
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

	req := map[string]interface{}{
		"sell_currency": "GBP",
		"buy_currency":  "USD",
		"sell_amount":   730.00,
	}

	conv, err := c.CreateConversion(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateConversion() error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversion is nil")
	}
	if conv.ID != "conv_direct" {
		t.Errorf("id = %q, want 'conv_direct'", conv.ID)
	}
	if conv.SellCurrency != "GBP" {
		t.Errorf("sell_currency = %q, want 'GBP'", conv.SellCurrency)
	}
	if conv.BuyCurrency != "USD" {
		t.Errorf("buy_currency = %q, want 'USD'", conv.BuyCurrency)
	}
}

func TestCreateConversion_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "quote_expired",
			"message": "Quote has expired"
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

	req := map[string]interface{}{
		"quote_id": "quote_expired",
	}

	_, err := c.CreateConversion(context.Background(), req)
	if err == nil {
		t.Error("expected error for expired quote, got nil")
	}
}

func TestCreateConversion_200Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "conv_200",
			"quote_id": "quote_123",
			"sell_currency": "USD",
			"buy_currency": "EUR",
			"sell_amount": 1000.00,
			"buy_amount": 850.00,
			"rate": 0.85,
			"status": "COMPLETED",
			"created_at": "2024-01-01T12:00:00Z"
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

	req := map[string]interface{}{
		"quote_id": "quote_123",
	}

	conv, err := c.CreateConversion(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateConversion() error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversion is nil")
	}
	if conv.ID != "conv_200" {
		t.Errorf("id = %q, want 'conv_200'", conv.ID)
	}
}
