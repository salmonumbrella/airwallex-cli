package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =====================================================
// ListTransactionDisputes Tests
// =====================================================

func TestListTransactionDisputes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issuing/transaction_disputes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"dispute_id": "disp_123",
					"id": "disp_123",
					"transaction_id": "txn_456",
					"status": "PENDING",
					"reason": "FRAUD",
					"amount": 100.50,
					"currency": "USD",
					"created_at": "2024-01-01T00:00:00Z"
				},
				{
					"dispute_id": "disp_789",
					"id": "disp_789",
					"transaction_id": "txn_012",
					"status": "RESOLVED",
					"reason": "MERCHANDISE_NOT_RECEIVED",
					"amount": 250.00,
					"currency": "EUR",
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

	result, err := c.ListTransactionDisputes(context.Background(), TransactionDisputeListParams{})
	if err != nil {
		t.Fatalf("ListTransactionDisputes() error: %v", err)
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

	// Verify first dispute
	d1 := result.Items[0]
	if d1.DisputeID != "disp_123" {
		t.Errorf("dispute_id = %q, want 'disp_123'", d1.DisputeID)
	}
	if d1.TransactionID != "txn_456" {
		t.Errorf("transaction_id = %q, want 'txn_456'", d1.TransactionID)
	}
	if d1.Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", d1.Status)
	}
	if d1.Reason != "FRAUD" {
		t.Errorf("reason = %q, want 'FRAUD'", d1.Reason)
	}
	if d1.Amount != 100.50 {
		t.Errorf("amount = %f, want 100.50", d1.Amount)
	}
	if d1.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", d1.Currency)
	}

	// Verify second dispute
	d2 := result.Items[1]
	if d2.DisputeID != "disp_789" {
		t.Errorf("dispute_id = %q, want 'disp_789'", d2.DisputeID)
	}
	if d2.Status != "RESOLVED" {
		t.Errorf("status = %q, want 'RESOLVED'", d2.Status)
	}
}

func TestListTransactionDisputes_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issuing/transaction_disputes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("status") != "PENDING" {
			t.Errorf("status = %q, want 'PENDING'", query.Get("status"))
		}
		if query.Get("detailed_status") != "UNDER_REVIEW" {
			t.Errorf("detailed_status = %q, want 'UNDER_REVIEW'", query.Get("detailed_status"))
		}
		if query.Get("reason") != "FRAUD" {
			t.Errorf("reason = %q, want 'FRAUD'", query.Get("reason"))
		}
		if query.Get("reference") != "REF123" {
			t.Errorf("reference = %q, want 'REF123'", query.Get("reference"))
		}
		if query.Get("transaction_id") != "txn_456" {
			t.Errorf("transaction_id = %q, want 'txn_456'", query.Get("transaction_id"))
		}
		if query.Get("updated_by") != "user@example.com" {
			t.Errorf("updated_by = %q, want 'user@example.com'", query.Get("updated_by"))
		}
		if query.Get("from_created_at") != "2024-01-01" {
			t.Errorf("from_created_at = %q, want '2024-01-01'", query.Get("from_created_at"))
		}
		if query.Get("to_created_at") != "2024-01-31" {
			t.Errorf("to_created_at = %q, want '2024-01-31'", query.Get("to_created_at"))
		}
		if query.Get("from_updated_at") != "2024-01-15" {
			t.Errorf("from_updated_at = %q, want '2024-01-15'", query.Get("from_updated_at"))
		}
		if query.Get("to_updated_at") != "2024-01-20" {
			t.Errorf("to_updated_at = %q, want '2024-01-20'", query.Get("to_updated_at"))
		}
		if query.Get("page") != "2" {
			t.Errorf("page = %q, want '2'", query.Get("page"))
		}
		if query.Get("page_size") != "10" {
			t.Errorf("page_size = %q, want '10'", query.Get("page_size"))
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

	params := TransactionDisputeListParams{
		Status:         "PENDING",
		DetailedStatus: "UNDER_REVIEW",
		Reason:         "FRAUD",
		Reference:      "REF123",
		TransactionID:  "txn_456",
		UpdatedBy:      "user@example.com",
		FromCreatedAt:  "2024-01-01",
		ToCreatedAt:    "2024-01-31",
		FromUpdatedAt:  "2024-01-15",
		ToUpdatedAt:    "2024-01-20",
		Page:           "2",
		PageSize:       10,
	}

	result, err := c.ListTransactionDisputes(context.Background(), params)
	if err != nil {
		t.Fatalf("ListTransactionDisputes() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestListTransactionDisputes_EmptyResponse(t *testing.T) {
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

	result, err := c.ListTransactionDisputes(context.Background(), TransactionDisputeListParams{})
	if err != nil {
		t.Fatalf("ListTransactionDisputes() error: %v", err)
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

func TestListTransactionDisputes_APIError(t *testing.T) {
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

	result, err := c.ListTransactionDisputes(context.Background(), TransactionDisputeListParams{})
	if err == nil {
		t.Error("expected error for unauthorized response, got nil")
	}
	if result != nil {
		t.Error("expected nil result for error response")
	}
}

func TestListTransactionDisputes_InvalidJSON(t *testing.T) {
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

	result, err := c.ListTransactionDisputes(context.Background(), TransactionDisputeListParams{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if result != nil {
		t.Error("expected nil result for invalid JSON")
	}
}

// =====================================================
// GetTransactionDispute Tests
// =====================================================

func TestGetTransactionDispute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issuing/transaction_disputes/disp_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"dispute_id": "disp_123",
			"id": "disp_123",
			"transaction_id": "txn_456",
			"status": "PENDING",
			"reason": "FRAUD",
			"amount": 150.75,
			"currency": "USD",
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

	dispute, err := c.GetTransactionDispute(context.Background(), "disp_123")
	if err != nil {
		t.Fatalf("GetTransactionDispute() error: %v", err)
	}
	if dispute == nil {
		t.Fatal("dispute is nil")
	}
	if dispute.DisputeID != "disp_123" {
		t.Errorf("dispute_id = %q, want 'disp_123'", dispute.DisputeID)
	}
	if dispute.TransactionID != "txn_456" {
		t.Errorf("transaction_id = %q, want 'txn_456'", dispute.TransactionID)
	}
	if dispute.Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", dispute.Status)
	}
	if dispute.Reason != "FRAUD" {
		t.Errorf("reason = %q, want 'FRAUD'", dispute.Reason)
	}
	if dispute.Amount != 150.75 {
		t.Errorf("amount = %f, want 150.75", dispute.Amount)
	}
	if dispute.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", dispute.Currency)
	}
}

func TestGetTransactionDispute_InvalidID(t *testing.T) {
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

	_, err := c.GetTransactionDispute(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty dispute ID, got nil")
	}

	_, err = c.GetTransactionDispute(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid dispute ID, got nil")
	}
}

func TestGetTransactionDispute_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Dispute not found"
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

	_, err := c.GetTransactionDispute(context.Background(), "disp_nonexistent")
	if err == nil {
		t.Error("expected error for not found dispute, got nil")
	}
}

func TestGetTransactionDispute_InvalidJSON(t *testing.T) {
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

	_, err := c.GetTransactionDispute(context.Background(), "disp_123")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// =====================================================
// CreateTransactionDispute Tests
// =====================================================

func TestCreateTransactionDispute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issuing/transaction_disputes/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"dispute_id": "disp_new",
			"id": "disp_new",
			"transaction_id": "txn_789",
			"status": "CREATED",
			"reason": "MERCHANDISE_NOT_RECEIVED",
			"amount": 200.00,
			"currency": "EUR",
			"created_at": "2024-01-15T10:30:00Z"
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
		"transaction_id": "txn_789",
		"reason":         "MERCHANDISE_NOT_RECEIVED",
		"amount":         200.00,
		"currency":       "EUR",
	}

	dispute, err := c.CreateTransactionDispute(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTransactionDispute() error: %v", err)
	}
	if dispute == nil {
		t.Fatal("dispute is nil")
	}
	if dispute.DisputeID != "disp_new" {
		t.Errorf("dispute_id = %q, want 'disp_new'", dispute.DisputeID)
	}
	if dispute.TransactionID != "txn_789" {
		t.Errorf("transaction_id = %q, want 'txn_789'", dispute.TransactionID)
	}
	if dispute.Status != "CREATED" {
		t.Errorf("status = %q, want 'CREATED'", dispute.Status)
	}
	if dispute.Reason != "MERCHANDISE_NOT_RECEIVED" {
		t.Errorf("reason = %q, want 'MERCHANDISE_NOT_RECEIVED'", dispute.Reason)
	}
	if dispute.Amount != 200.00 {
		t.Errorf("amount = %f, want 200.00", dispute.Amount)
	}
}

func TestCreateTransactionDispute_SuccessWithStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"dispute_id": "disp_ok",
			"id": "disp_ok",
			"transaction_id": "txn_111",
			"status": "CREATED",
			"reason": "FRAUD",
			"amount": 50.00,
			"currency": "USD",
			"created_at": "2024-01-15T10:30:00Z"
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
		"transaction_id": "txn_111",
		"reason":         "FRAUD",
	}

	dispute, err := c.CreateTransactionDispute(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTransactionDispute() error: %v", err)
	}
	if dispute == nil {
		t.Fatal("dispute is nil")
	}
	if dispute.DisputeID != "disp_ok" {
		t.Errorf("dispute_id = %q, want 'disp_ok'", dispute.DisputeID)
	}
}

func TestCreateTransactionDispute_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_failed",
			"message": "Invalid dispute data"
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
		"invalid": "data",
	}

	_, err := c.CreateTransactionDispute(context.Background(), req)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestCreateTransactionDispute_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
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

	req := map[string]interface{}{
		"transaction_id": "txn_123",
	}

	_, err := c.CreateTransactionDispute(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// =====================================================
// UpdateTransactionDispute Tests
// =====================================================

func TestUpdateTransactionDispute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issuing/transaction_disputes/disp_123/update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"dispute_id": "disp_123",
			"id": "disp_123",
			"transaction_id": "txn_456",
			"status": "PENDING",
			"reason": "DUPLICATE_CHARGE",
			"amount": 175.00,
			"currency": "USD",
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

	update := map[string]interface{}{
		"reason": "DUPLICATE_CHARGE",
		"amount": 175.00,
	}

	dispute, err := c.UpdateTransactionDispute(context.Background(), "disp_123", update)
	if err != nil {
		t.Fatalf("UpdateTransactionDispute() error: %v", err)
	}
	if dispute == nil {
		t.Fatal("dispute is nil")
	}
	if dispute.DisputeID != "disp_123" {
		t.Errorf("dispute_id = %q, want 'disp_123'", dispute.DisputeID)
	}
	if dispute.Reason != "DUPLICATE_CHARGE" {
		t.Errorf("reason = %q, want 'DUPLICATE_CHARGE'", dispute.Reason)
	}
	if dispute.Amount != 175.00 {
		t.Errorf("amount = %f, want 175.00", dispute.Amount)
	}
}

func TestUpdateTransactionDispute_InvalidID(t *testing.T) {
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

	update := map[string]interface{}{
		"reason": "FRAUD",
	}

	_, err := c.UpdateTransactionDispute(context.Background(), "", update)
	if err == nil {
		t.Error("expected error for empty dispute ID, got nil")
	}

	_, err = c.UpdateTransactionDispute(context.Background(), "invalid/id", update)
	if err == nil {
		t.Error("expected error for invalid dispute ID, got nil")
	}
}

func TestUpdateTransactionDispute_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Dispute not found"
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

	update := map[string]interface{}{
		"reason": "FRAUD",
	}

	_, err := c.UpdateTransactionDispute(context.Background(), "disp_nonexistent", update)
	if err == nil {
		t.Error("expected error for not found dispute, got nil")
	}
}

func TestUpdateTransactionDispute_InvalidJSON(t *testing.T) {
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

	update := map[string]interface{}{
		"reason": "FRAUD",
	}

	_, err := c.UpdateTransactionDispute(context.Background(), "disp_123", update)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// =====================================================
// SubmitTransactionDispute Tests
// =====================================================

func TestSubmitTransactionDispute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issuing/transaction_disputes/disp_123/submit" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"dispute_id": "disp_123",
			"id": "disp_123",
			"transaction_id": "txn_456",
			"status": "SUBMITTED",
			"reason": "FRAUD",
			"amount": 100.00,
			"currency": "USD",
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

	dispute, err := c.SubmitTransactionDispute(context.Background(), "disp_123")
	if err != nil {
		t.Fatalf("SubmitTransactionDispute() error: %v", err)
	}
	if dispute == nil {
		t.Fatal("dispute is nil")
	}
	if dispute.DisputeID != "disp_123" {
		t.Errorf("dispute_id = %q, want 'disp_123'", dispute.DisputeID)
	}
	if dispute.Status != "SUBMITTED" {
		t.Errorf("status = %q, want 'SUBMITTED'", dispute.Status)
	}
}

func TestSubmitTransactionDispute_InvalidID(t *testing.T) {
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

	_, err := c.SubmitTransactionDispute(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty dispute ID, got nil")
	}

	_, err = c.SubmitTransactionDispute(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid dispute ID, got nil")
	}
}

func TestSubmitTransactionDispute_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Dispute not found"
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

	_, err := c.SubmitTransactionDispute(context.Background(), "disp_nonexistent")
	if err == nil {
		t.Error("expected error for not found dispute, got nil")
	}
}

func TestSubmitTransactionDispute_InvalidState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_operation",
			"message": "Dispute already submitted"
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

	_, err := c.SubmitTransactionDispute(context.Background(), "disp_already_submitted")
	if err == nil {
		t.Error("expected error for already submitted dispute, got nil")
	}
}

func TestSubmitTransactionDispute_InvalidJSON(t *testing.T) {
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

	_, err := c.SubmitTransactionDispute(context.Background(), "disp_123")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// =====================================================
// CancelTransactionDispute Tests
// =====================================================

func TestCancelTransactionDispute_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issuing/transaction_disputes/disp_123/cancel" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"dispute_id": "disp_123",
			"id": "disp_123",
			"transaction_id": "txn_456",
			"status": "CANCELLED",
			"reason": "FRAUD",
			"amount": 100.00,
			"currency": "USD",
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

	dispute, err := c.CancelTransactionDispute(context.Background(), "disp_123")
	if err != nil {
		t.Fatalf("CancelTransactionDispute() error: %v", err)
	}
	if dispute == nil {
		t.Fatal("dispute is nil")
	}
	if dispute.DisputeID != "disp_123" {
		t.Errorf("dispute_id = %q, want 'disp_123'", dispute.DisputeID)
	}
	if dispute.Status != "CANCELLED" {
		t.Errorf("status = %q, want 'CANCELLED'", dispute.Status)
	}
}

func TestCancelTransactionDispute_InvalidID(t *testing.T) {
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

	_, err := c.CancelTransactionDispute(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty dispute ID, got nil")
	}

	_, err = c.CancelTransactionDispute(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid dispute ID, got nil")
	}
}

func TestCancelTransactionDispute_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Dispute not found"
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

	_, err := c.CancelTransactionDispute(context.Background(), "disp_nonexistent")
	if err == nil {
		t.Error("expected error for not found dispute, got nil")
	}
}

func TestCancelTransactionDispute_AlreadyCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_operation",
			"message": "Dispute already cancelled"
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

	_, err := c.CancelTransactionDispute(context.Background(), "disp_already_cancelled")
	if err == nil {
		t.Error("expected error for already cancelled dispute, got nil")
	}
}

func TestCancelTransactionDispute_InvalidJSON(t *testing.T) {
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

	_, err := c.CancelTransactionDispute(context.Background(), "disp_123")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestCancelTransactionDispute_InvalidState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_operation",
			"message": "Cannot cancel dispute in RESOLVED status"
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

	_, err := c.CancelTransactionDispute(context.Background(), "disp_resolved")
	if err == nil {
		t.Error("expected error for resolved dispute, got nil")
	}
}
