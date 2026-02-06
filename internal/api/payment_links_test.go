package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListPaymentLinks_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/payment_links" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters are correctly set
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "1" {
			t.Errorf("page_num = %q, want '1'", pageNum)
		}
		if pageSize != "20" {
			t.Errorf("page_size = %q, want '20'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "pl_123",
					"url": "https://checkout.airwallex.com/pl_123",
					"amount": 100.50,
					"currency": "USD",
					"description": "Test payment link",
					"status": "ACTIVE",
					"expires_at": "2024-12-31T23:59:59Z",
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

	result, err := c.ListPaymentLinks(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("ListPaymentLinks() error: %v", err)
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
	if result.Items[0].ID != "pl_123" {
		t.Errorf("id = %q, want 'pl_123'", result.Items[0].ID)
	}
	if result.Items[0].URL != "https://checkout.airwallex.com/pl_123" {
		t.Errorf("url = %q, want 'https://checkout.airwallex.com/pl_123'", result.Items[0].URL)
	}
	if result.Items[0].Amount != jn("100.50") {
		t.Errorf("amount = %s, want 100.50", result.Items[0].Amount)
	}
	if result.Items[0].Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", result.Items[0].Currency)
	}
	if result.Items[0].Description != "Test payment link" {
		t.Errorf("description = %q, want 'Test payment link'", result.Items[0].Description)
	}
	if result.Items[0].Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", result.Items[0].Status)
	}
}

func TestListPaymentLinks_WithoutPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/payment_links" {
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

	result, err := c.ListPaymentLinks(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListPaymentLinks() error: %v", err)
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

func TestListPaymentLinks_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	result, err := c.ListPaymentLinks(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListPaymentLinks() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
}

func TestListPaymentLinks_PartialPagination(t *testing.T) {
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

	_, err := c.ListPaymentLinks(context.Background(), 0, 5)
	if err != nil {
		t.Fatalf("ListPaymentLinks() error: %v", err)
	}
}

func TestListPaymentLinks_MultipleItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "pl_001",
					"url": "https://checkout.airwallex.com/pl_001",
					"amount": 50.00,
					"currency": "USD",
					"status": "ACTIVE",
					"created_at": "2024-01-01T00:00:00Z"
				},
				{
					"id": "pl_002",
					"url": "https://checkout.airwallex.com/pl_002",
					"amount": 75.25,
					"currency": "EUR",
					"status": "EXPIRED",
					"created_at": "2024-01-02T00:00:00Z"
				},
				{
					"id": "pl_003",
					"url": "https://checkout.airwallex.com/pl_003",
					"amount": 200.00,
					"currency": "GBP",
					"status": "ACTIVE",
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

	result, err := c.ListPaymentLinks(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListPaymentLinks() error: %v", err)
	}
	if len(result.Items) != 3 {
		t.Errorf("items count = %d, want 3", len(result.Items))
	}
	if result.Items[1].Currency != "EUR" {
		t.Errorf("items[1].currency = %q, want 'EUR'", result.Items[1].Currency)
	}
	if result.Items[2].Amount != jn("200.00") {
		t.Errorf("items[2].amount = %s, want 200.00", result.Items[2].Amount)
	}
}

func TestGetPaymentLink_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/payment_links/pl_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "pl_123",
			"url": "https://checkout.airwallex.com/pl_123",
			"amount": 100.50,
			"currency": "USD",
			"description": "Test payment link",
			"status": "ACTIVE",
			"expires_at": "2024-12-31T23:59:59Z",
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

	link, err := c.GetPaymentLink(context.Background(), "pl_123")
	if err != nil {
		t.Fatalf("GetPaymentLink() error: %v", err)
	}
	if link == nil {
		t.Fatal("link is nil")
	}
	if link.ID != "pl_123" {
		t.Errorf("id = %q, want 'pl_123'", link.ID)
	}
	if link.URL != "https://checkout.airwallex.com/pl_123" {
		t.Errorf("url = %q, want 'https://checkout.airwallex.com/pl_123'", link.URL)
	}
	if link.Amount != jn("100.50") {
		t.Errorf("amount = %s, want 100.50", link.Amount)
	}
	if link.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", link.Currency)
	}
	if link.Description != "Test payment link" {
		t.Errorf("description = %q, want 'Test payment link'", link.Description)
	}
	if link.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", link.Status)
	}
}

func TestGetPaymentLink_InvalidID(t *testing.T) {
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

	_, err := c.GetPaymentLink(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty payment link ID, got nil")
	}

	_, err = c.GetPaymentLink(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid payment link ID, got nil")
	}
}

func TestGetPaymentLink_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Payment link not found"
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

	_, err := c.GetPaymentLink(context.Background(), "pl_nonexistent")
	if err == nil {
		t.Error("expected error for not found payment link, got nil")
	}
}

func TestGetPaymentLink_SpecialCharactersInID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should escape special characters in the URL path
		expectedPath := "/api/v1/pa/payment_links/pl_test-123"
		if r.URL.Path != expectedPath {
			t.Errorf("path = %q, want %q", r.URL.Path, expectedPath)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "pl_test-123",
			"url": "https://checkout.airwallex.com/pl_test-123",
			"amount": 50.00,
			"currency": "USD",
			"status": "ACTIVE",
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

	link, err := c.GetPaymentLink(context.Background(), "pl_test-123")
	if err != nil {
		t.Fatalf("GetPaymentLink() error: %v", err)
	}
	if link.ID != "pl_test-123" {
		t.Errorf("id = %q, want 'pl_test-123'", link.ID)
	}
}

func TestCreatePaymentLink_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/payment_links/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "pl_new",
			"url": "https://checkout.airwallex.com/pl_new",
			"amount": 100.50,
			"currency": "USD",
			"description": "Test payment link",
			"status": "ACTIVE",
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

	req := map[string]interface{}{
		"amount":      100.50,
		"currency":    "USD",
		"description": "Test payment link",
	}

	link, err := c.CreatePaymentLink(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePaymentLink() error: %v", err)
	}
	if link == nil {
		t.Fatal("link is nil")
	}
	if link.ID != "pl_new" {
		t.Errorf("id = %q, want 'pl_new'", link.ID)
	}
	if link.URL != "https://checkout.airwallex.com/pl_new" {
		t.Errorf("url = %q, want 'https://checkout.airwallex.com/pl_new'", link.URL)
	}
	if link.Amount != jn("100.50") {
		t.Errorf("amount = %s, want 100.50", link.Amount)
	}
	if link.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", link.Currency)
	}
	if link.Description != "Test payment link" {
		t.Errorf("description = %q, want 'Test payment link'", link.Description)
	}
	if link.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", link.Status)
	}
}

func TestCreatePaymentLink_WithStatusOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test that 200 OK is also accepted (not just 201 Created)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "pl_ok",
			"url": "https://checkout.airwallex.com/pl_ok",
			"amount": 50.00,
			"currency": "EUR",
			"status": "ACTIVE",
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

	req := map[string]interface{}{
		"amount":   50.00,
		"currency": "EUR",
	}

	link, err := c.CreatePaymentLink(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePaymentLink() error: %v", err)
	}
	if link.ID != "pl_ok" {
		t.Errorf("id = %q, want 'pl_ok'", link.ID)
	}
}

func TestCreatePaymentLink_MinimalRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "pl_minimal",
			"url": "https://checkout.airwallex.com/pl_minimal",
			"amount": 10.00,
			"currency": "USD",
			"status": "ACTIVE",
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

	// Minimal request with only required fields
	req := map[string]interface{}{
		"amount":   10.00,
		"currency": "USD",
	}

	link, err := c.CreatePaymentLink(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePaymentLink() error: %v", err)
	}
	if link.Amount != jn("10.00") {
		t.Errorf("amount = %s, want 10.00", link.Amount)
	}
	if link.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", link.Currency)
	}
}

func TestCreatePaymentLink_WithExpiration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "pl_expiring",
			"url": "https://checkout.airwallex.com/pl_expiring",
			"amount": 75.00,
			"currency": "GBP",
			"description": "Payment link with expiration",
			"status": "ACTIVE",
			"expires_at": "2024-12-31T23:59:59Z",
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

	req := map[string]interface{}{
		"amount":      75.00,
		"currency":    "GBP",
		"description": "Payment link with expiration",
		"expires_at":  "2024-12-31T23:59:59Z",
	}

	link, err := c.CreatePaymentLink(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePaymentLink() error: %v", err)
	}
	if link.ExpiresAt != "2024-12-31T23:59:59Z" {
		t.Errorf("expires_at = %q, want '2024-12-31T23:59:59Z'", link.ExpiresAt)
	}
}

func TestCreatePaymentLink_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_request",
			"message": "Amount must be positive"
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
		"amount":   -10.00,
		"currency": "USD",
	}

	_, err := c.CreatePaymentLink(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid request, got nil")
	}
}

func TestCreatePaymentLink_DifferentCurrencies(t *testing.T) {
	currencies := []string{"USD", "EUR", "GBP", "JPY", "AUD"}

	for _, currency := range currencies {
		t.Run(currency, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{
					"id": "pl_` + currency + `",
					"url": "https://checkout.airwallex.com/pl_` + currency + `",
					"amount": 100.00,
					"currency": "` + currency + `",
					"status": "ACTIVE",
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

			req := map[string]interface{}{
				"amount":   100.00,
				"currency": currency,
			}

			link, err := c.CreatePaymentLink(context.Background(), req)
			if err != nil {
				t.Fatalf("CreatePaymentLink() error: %v", err)
			}
			if link.Currency != currency {
				t.Errorf("currency = %q, want %q", link.Currency, currency)
			}
		})
	}
}
