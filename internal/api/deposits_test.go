package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestListDeposits_WithAllFilters tests listing deposits with all filter options
func TestListDeposits_WithAllFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deposits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters are correctly set
		status := r.URL.Query().Get("status")
		fromDate := r.URL.Query().Get("from_created_at")
		toDate := r.URL.Query().Get("to_created_at")
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if status != "SETTLED" {
			t.Errorf("status = %q, want 'SETTLED'", status)
		}
		if fromDate != "2024-01-01T00:00:00Z" {
			t.Errorf("from_created_at = %q, want '2024-01-01T00:00:00Z'", fromDate)
		}
		if toDate != "2024-12-31T23:59:59Z" {
			t.Errorf("to_created_at = %q, want '2024-12-31T23:59:59Z'", toDate)
		}
		if pageNum != "2" {
			t.Errorf("page_num = %q, want '2'", pageNum)
		}
		if pageSize != "50" {
			t.Errorf("page_size = %q, want '50'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "dep_123",
					"amount": 1000.50,
					"currency": "USD",
					"status": "SETTLED",
					"source": "WIRE_TRANSFER",
					"linked_account_id": "la_456",
					"global_account_id": "ga_789",
					"reference": "REF-001",
					"created_at": "2024-01-15T10:30:00Z",
					"settled_at": "2024-01-16T14:20:00Z"
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

	result, err := c.ListDeposits(context.Background(), "SETTLED", "2024-01-01T00:00:00Z", "2024-12-31T23:59:59Z", 2, 50)
	if err != nil {
		t.Fatalf("ListDeposits() error: %v", err)
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

	// Verify deposit details
	deposit := result.Items[0]
	if deposit.ID != "dep_123" {
		t.Errorf("id = %q, want 'dep_123'", deposit.ID)
	}
	if deposit.Amount != jn("1000.50") {
		t.Errorf("amount = %s, want 1000.50", deposit.Amount)
	}
	if deposit.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", deposit.Currency)
	}
	if deposit.Status != "SETTLED" {
		t.Errorf("status = %q, want 'SETTLED'", deposit.Status)
	}
	if deposit.Source != "WIRE_TRANSFER" {
		t.Errorf("source = %q, want 'WIRE_TRANSFER'", deposit.Source)
	}
	if deposit.LinkedAccountID != "la_456" {
		t.Errorf("linked_account_id = %q, want 'la_456'", deposit.LinkedAccountID)
	}
	if deposit.GlobalAccountID != "ga_789" {
		t.Errorf("global_account_id = %q, want 'ga_789'", deposit.GlobalAccountID)
	}
	if deposit.Reference != "REF-001" {
		t.Errorf("reference = %q, want 'REF-001'", deposit.Reference)
	}
}

// TestListDeposits_WithPagination tests listing deposits with only pagination parameters
func TestListDeposits_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deposits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify only pagination parameters are set
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")
		status := r.URL.Query().Get("status")
		fromDate := r.URL.Query().Get("from_created_at")
		toDate := r.URL.Query().Get("to_created_at")

		if pageNum != "1" {
			t.Errorf("page_num = %q, want '1'", pageNum)
		}
		if pageSize != "20" {
			t.Errorf("page_size = %q, want '20'", pageSize)
		}
		if status != "" {
			t.Errorf("status should be empty, got %q", status)
		}
		if fromDate != "" {
			t.Errorf("from_created_at should be empty, got %q", fromDate)
		}
		if toDate != "" {
			t.Errorf("to_created_at should be empty, got %q", toDate)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "dep_page1",
					"amount": 500.00,
					"currency": "EUR",
					"status": "PENDING",
					"source": "ACH",
					"created_at": "2024-01-10T09:00:00Z"
				},
				{
					"id": "dep_page2",
					"amount": 750.25,
					"currency": "GBP",
					"status": "SETTLED",
					"source": "SWIFT",
					"created_at": "2024-01-11T11:30:00Z",
					"settled_at": "2024-01-12T10:00:00Z"
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

	result, err := c.ListDeposits(context.Background(), "", "", "", 0, 20)
	if err != nil {
		t.Fatalf("ListDeposits() error: %v", err)
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
}

// TestListDeposits_WithStatusFilter tests filtering deposits by status only
func TestListDeposits_WithStatusFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deposits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		status := r.URL.Query().Get("status")
		if status != "PENDING" {
			t.Errorf("status = %q, want 'PENDING'", status)
		}

		// Verify other filters are not set
		if r.URL.Query().Get("from_created_at") != "" {
			t.Error("from_created_at should be empty")
		}
		if r.URL.Query().Get("to_created_at") != "" {
			t.Error("to_created_at should be empty")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "dep_pending",
					"amount": 250.00,
					"currency": "USD",
					"status": "PENDING",
					"source": "WIRE_TRANSFER",
					"created_at": "2024-01-20T08:00:00Z"
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

	result, err := c.ListDeposits(context.Background(), "PENDING", "", "", 0, 0)
	if err != nil {
		t.Fatalf("ListDeposits() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", result.Items[0].Status)
	}
}

// TestListDeposits_WithDateFilters tests filtering deposits by date range
func TestListDeposits_WithDateFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deposits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		fromDate := r.URL.Query().Get("from_created_at")
		toDate := r.URL.Query().Get("to_created_at")

		if fromDate != "2024-01-01T00:00:00Z" {
			t.Errorf("from_created_at = %q, want '2024-01-01T00:00:00Z'", fromDate)
		}
		if toDate != "2024-01-31T23:59:59Z" {
			t.Errorf("to_created_at = %q, want '2024-01-31T23:59:59Z'", toDate)
		}

		// Verify status filter is not set
		if r.URL.Query().Get("status") != "" {
			t.Error("status should be empty")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "dep_jan",
					"amount": 1500.00,
					"currency": "USD",
					"status": "SETTLED",
					"source": "ACH",
					"created_at": "2024-01-15T12:00:00Z",
					"settled_at": "2024-01-16T09:00:00Z"
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

	result, err := c.ListDeposits(context.Background(), "", "2024-01-01T00:00:00Z", "2024-01-31T23:59:59Z", 0, 0)
	if err != nil {
		t.Fatalf("ListDeposits() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
}

// TestListDeposits_NoFilters tests listing deposits without any filters
func TestListDeposits_NoFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deposits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify no query parameters are set when values are empty or zero
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query parameters, got: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "dep_all1",
					"amount": 100.00,
					"currency": "USD",
					"status": "SETTLED",
					"source": "WIRE_TRANSFER",
					"created_at": "2024-01-01T00:00:00Z"
				},
				{
					"id": "dep_all2",
					"amount": 200.00,
					"currency": "EUR",
					"status": "PENDING",
					"source": "ACH",
					"created_at": "2024-01-02T00:00:00Z"
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

	result, err := c.ListDeposits(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("ListDeposits() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 2 {
		t.Errorf("items count = %d, want 2", len(result.Items))
	}
	if result.HasMore {
		t.Error("has_more = true, want false")
	}
}

// TestListDeposits_EmptyResults tests handling of empty deposit list
func TestListDeposits_EmptyResults(t *testing.T) {
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

	result, err := c.ListDeposits(context.Background(), "SETTLED", "", "", 1, 10)
	if err != nil {
		t.Fatalf("ListDeposits() error: %v", err)
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

// TestListDeposits_APIError tests handling of API errors
func TestListDeposits_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_request",
			"message": "Invalid status parameter"
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

	_, err := c.ListDeposits(context.Background(), "INVALID_STATUS", "", "", 0, 0)
	if err == nil {
		t.Error("expected error for invalid request, got nil")
	}
}

// TestGetDeposit_Success tests successful deposit retrieval
func TestGetDeposit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deposits/dep_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "dep_123",
			"amount": 2500.75,
			"currency": "USD",
			"status": "SETTLED",
			"source": "WIRE_TRANSFER",
			"linked_account_id": "la_abc",
			"global_account_id": "ga_xyz",
			"reference": "WIRE-2024-001",
			"created_at": "2024-01-15T10:30:00Z",
			"settled_at": "2024-01-16T14:20:00Z"
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

	deposit, err := c.GetDeposit(context.Background(), "dep_123")
	if err != nil {
		t.Fatalf("GetDeposit() error: %v", err)
	}
	if deposit == nil {
		t.Fatal("deposit is nil")
	}

	// Verify all deposit fields
	if deposit.ID != "dep_123" {
		t.Errorf("id = %q, want 'dep_123'", deposit.ID)
	}
	if deposit.Amount != jn("2500.75") {
		t.Errorf("amount = %s, want 2500.75", deposit.Amount)
	}
	if deposit.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", deposit.Currency)
	}
	if deposit.Status != "SETTLED" {
		t.Errorf("status = %q, want 'SETTLED'", deposit.Status)
	}
	if deposit.Source != "WIRE_TRANSFER" {
		t.Errorf("source = %q, want 'WIRE_TRANSFER'", deposit.Source)
	}
	if deposit.LinkedAccountID != "la_abc" {
		t.Errorf("linked_account_id = %q, want 'la_abc'", deposit.LinkedAccountID)
	}
	if deposit.GlobalAccountID != "ga_xyz" {
		t.Errorf("global_account_id = %q, want 'ga_xyz'", deposit.GlobalAccountID)
	}
	if deposit.Reference != "WIRE-2024-001" {
		t.Errorf("reference = %q, want 'WIRE-2024-001'", deposit.Reference)
	}
	if deposit.CreatedAt != "2024-01-15T10:30:00Z" {
		t.Errorf("created_at = %q, want '2024-01-15T10:30:00Z'", deposit.CreatedAt)
	}
	if deposit.SettledAt != "2024-01-16T14:20:00Z" {
		t.Errorf("settled_at = %q, want '2024-01-16T14:20:00Z'", deposit.SettledAt)
	}
}

// TestGetDeposit_MinimalFields tests deposit with only required fields
func TestGetDeposit_MinimalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/deposits/dep_minimal" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "dep_minimal",
			"amount": 100.00,
			"currency": "EUR",
			"status": "PENDING",
			"source": "ACH",
			"created_at": "2024-01-20T08:00:00Z"
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

	deposit, err := c.GetDeposit(context.Background(), "dep_minimal")
	if err != nil {
		t.Fatalf("GetDeposit() error: %v", err)
	}
	if deposit == nil {
		t.Fatal("deposit is nil")
	}

	// Verify required fields are present
	if deposit.ID != "dep_minimal" {
		t.Errorf("id = %q, want 'dep_minimal'", deposit.ID)
	}
	if deposit.Amount != jn("100.00") {
		t.Errorf("amount = %s, want 100.00", deposit.Amount)
	}
	if deposit.Currency != "EUR" {
		t.Errorf("currency = %q, want 'EUR'", deposit.Currency)
	}
	if deposit.Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", deposit.Status)
	}

	// Verify optional fields are empty
	if deposit.LinkedAccountID != "" {
		t.Errorf("linked_account_id = %q, want empty string", deposit.LinkedAccountID)
	}
	if deposit.GlobalAccountID != "" {
		t.Errorf("global_account_id = %q, want empty string", deposit.GlobalAccountID)
	}
	if deposit.Reference != "" {
		t.Errorf("reference = %q, want empty string", deposit.Reference)
	}
	if deposit.SettledAt != "" {
		t.Errorf("settled_at = %q, want empty string", deposit.SettledAt)
	}
}

// TestGetDeposit_InvalidID tests error handling for invalid deposit IDs
func TestGetDeposit_InvalidID(t *testing.T) {
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

	// Test empty ID
	_, err := c.GetDeposit(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty deposit ID, got nil")
	}

	// Test ID with invalid characters
	_, err = c.GetDeposit(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid deposit ID with slash, got nil")
	}

	// Test ID with special characters
	_, err = c.GetDeposit(context.Background(), "dep@123")
	if err == nil {
		t.Error("expected error for deposit ID with @ character, got nil")
	}

	// Test ID with spaces
	_, err = c.GetDeposit(context.Background(), "dep 123")
	if err == nil {
		t.Error("expected error for deposit ID with space, got nil")
	}
}

// TestGetDeposit_NotFound tests handling of not found errors
func TestGetDeposit_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Deposit not found"
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

	_, err := c.GetDeposit(context.Background(), "dep_nonexistent")
	if err == nil {
		t.Error("expected error for not found deposit, got nil")
	}
}

// TestGetDeposit_UnauthorizedError tests handling of unauthorized errors
func TestGetDeposit_UnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{
			"code": "unauthorized",
			"message": "Invalid credentials"
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
			Token:     "invalid-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetDeposit(context.Background(), "dep_123")
	if err == nil {
		t.Error("expected error for unauthorized request, got nil")
	}
}

// TestGetDeposit_InternalServerError tests handling of server errors
func TestGetDeposit_InternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{
			"code": "internal_error",
			"message": "Something went wrong"
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

	_, err := c.GetDeposit(context.Background(), "dep_123")
	if err == nil {
		t.Error("expected error for internal server error, got nil")
	}
}
