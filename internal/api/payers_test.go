package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =====================================================
// ListPayers Tests
// =====================================================

func TestListPayers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "pyr_123",
					"payer_id": "pyr_123",
					"entity_type": "COMPANY",
					"name": "Acme Corp",
					"status": "ACTIVE",
					"created_at": "2024-01-01T00:00:00Z"
				},
				{
					"id": "pyr_456",
					"payer_id": "pyr_456",
					"entity_type": "INDIVIDUAL",
					"name": "John Doe",
					"status": "ACTIVE",
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

	result, err := c.ListPayers(context.Background(), PayerListParams{})
	if err != nil {
		t.Fatalf("ListPayers() error: %v", err)
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
	if result.Items[0].ID != "pyr_123" {
		t.Errorf("items[0].id = %q, want 'pyr_123'", result.Items[0].ID)
	}
	if result.Items[0].EntityType != "COMPANY" {
		t.Errorf("items[0].entity_type = %q, want 'COMPANY'", result.Items[0].EntityType)
	}
}

func TestListPayers_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers" {
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
					"id": "pyr_789",
					"payer_id": "pyr_789",
					"entity_type": "COMPANY",
					"name": "Test Payer",
					"status": "ACTIVE",
					"created_at": "2024-01-15T00:00:00Z"
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

	result, err := c.ListPayers(context.Background(), PayerListParams{
		PageNum:  2,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListPayers() error: %v", err)
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

func TestListPayers_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters for filters
		entityType := r.URL.Query().Get("entity_type")
		name := r.URL.Query().Get("name")
		nickName := r.URL.Query().Get("nick_name")
		fromDate := r.URL.Query().Get("from_date")
		toDate := r.URL.Query().Get("to_date")

		if entityType != "COMPANY" {
			t.Errorf("entity_type = %q, want 'COMPANY'", entityType)
		}
		if name != "Acme" {
			t.Errorf("name = %q, want 'Acme'", name)
		}
		if nickName != "primary" {
			t.Errorf("nick_name = %q, want 'primary'", nickName)
		}
		if fromDate != "2024-01-01" {
			t.Errorf("from_date = %q, want '2024-01-01'", fromDate)
		}
		if toDate != "2024-12-31" {
			t.Errorf("to_date = %q, want '2024-12-31'", toDate)
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

	_, err := c.ListPayers(context.Background(), PayerListParams{
		EntityType: "COMPANY",
		Name:       "Acme",
		NickName:   "primary",
		FromDate:   "2024-01-01",
		ToDate:     "2024-12-31",
	})
	if err != nil {
		t.Fatalf("ListPayers() error: %v", err)
	}
}

func TestListPayers_PageNumDefaultsToOneWhenPageSizeSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// When page_size is set but page_num is 0, it should default to 1
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

	_, err := c.ListPayers(context.Background(), PayerListParams{
		PageSize: 5,
		// PageNum intentionally left as 0
	})
	if err != nil {
		t.Fatalf("ListPayers() error: %v", err)
	}
}

func TestListPayers_Error(t *testing.T) {
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

	_, err := c.ListPayers(context.Background(), PayerListParams{})
	if err == nil {
		t.Error("expected error for server error, got nil")
	}
}

// =====================================================
// GetPayer Tests
// =====================================================

func TestGetPayer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers/pyr_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "pyr_123",
			"payer_id": "pyr_123",
			"entity_type": "COMPANY",
			"name": "Acme Corporation",
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

	payer, err := c.GetPayer(context.Background(), "pyr_123")
	if err != nil {
		t.Fatalf("GetPayer() error: %v", err)
	}
	if payer == nil {
		t.Fatal("payer is nil")
	}
	if payer.ID != "pyr_123" {
		t.Errorf("id = %q, want 'pyr_123'", payer.ID)
	}
	if payer.EntityType != "COMPANY" {
		t.Errorf("entity_type = %q, want 'COMPANY'", payer.EntityType)
	}
	if payer.Name != "Acme Corporation" {
		t.Errorf("name = %q, want 'Acme Corporation'", payer.Name)
	}
	if payer.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", payer.Status)
	}
}

func TestGetPayer_InvalidID(t *testing.T) {
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

	_, err := c.GetPayer(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty payer ID, got nil")
	}

	_, err = c.GetPayer(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid payer ID, got nil")
	}
}

func TestGetPayer_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Payer not found"
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

	_, err := c.GetPayer(context.Background(), "pyr_nonexistent")
	if err == nil {
		t.Error("expected error for not found payer, got nil")
	}
}

// =====================================================
// CreatePayer Tests
// =====================================================

func TestCreatePayer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "pyr_new",
			"payer_id": "pyr_new",
			"entity_type": "COMPANY",
			"name": "New Company Ltd",
			"status": "PENDING",
			"created_at": "2024-02-01T00:00:00Z"
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
		"entity_type": "COMPANY",
		"name":        "New Company Ltd",
	}

	payer, err := c.CreatePayer(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePayer() error: %v", err)
	}
	if payer == nil {
		t.Fatal("payer is nil")
	}
	if payer.ID != "pyr_new" {
		t.Errorf("id = %q, want 'pyr_new'", payer.ID)
	}
	if payer.Name != "New Company Ltd" {
		t.Errorf("name = %q, want 'New Company Ltd'", payer.Name)
	}
	if payer.Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", payer.Status)
	}
}

func TestCreatePayer_SuccessWithStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "pyr_ok",
			"payer_id": "pyr_ok",
			"entity_type": "INDIVIDUAL",
			"name": "Jane Doe",
			"status": "ACTIVE",
			"created_at": "2024-02-01T00:00:00Z"
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
		"entity_type": "INDIVIDUAL",
		"name":        "Jane Doe",
	}

	payer, err := c.CreatePayer(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePayer() error: %v", err)
	}
	if payer == nil {
		t.Fatal("payer is nil")
	}
	if payer.ID != "pyr_ok" {
		t.Errorf("id = %q, want 'pyr_ok'", payer.ID)
	}
}

func TestCreatePayer_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_failed",
			"message": "Invalid payer data"
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

	_, err := c.CreatePayer(context.Background(), req)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

// =====================================================
// UpdatePayer Tests
// =====================================================

func TestUpdatePayer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers/update/pyr_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "pyr_123",
			"payer_id": "pyr_123",
			"entity_type": "COMPANY",
			"name": "Updated Company Name",
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

	update := map[string]interface{}{
		"name": "Updated Company Name",
	}

	payer, err := c.UpdatePayer(context.Background(), "pyr_123", update)
	if err != nil {
		t.Fatalf("UpdatePayer() error: %v", err)
	}
	if payer == nil {
		t.Fatal("payer is nil")
	}
	if payer.ID != "pyr_123" {
		t.Errorf("id = %q, want 'pyr_123'", payer.ID)
	}
	if payer.Name != "Updated Company Name" {
		t.Errorf("name = %q, want 'Updated Company Name'", payer.Name)
	}
}

func TestUpdatePayer_InvalidID(t *testing.T) {
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
		"name": "New Name",
	}

	_, err := c.UpdatePayer(context.Background(), "", update)
	if err == nil {
		t.Error("expected error for empty payer ID, got nil")
	}

	_, err = c.UpdatePayer(context.Background(), "invalid/id", update)
	if err == nil {
		t.Error("expected error for invalid payer ID, got nil")
	}
}

func TestUpdatePayer_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Payer not found"
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
		"name": "New Name",
	}

	_, err := c.UpdatePayer(context.Background(), "pyr_nonexistent", update)
	if err == nil {
		t.Error("expected error for not found payer, got nil")
	}
}

// =====================================================
// DeletePayer Tests
// =====================================================

func TestDeletePayer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers/delete/pyr_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
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

	err := c.DeletePayer(context.Background(), "pyr_123")
	if err != nil {
		t.Fatalf("DeletePayer() error: %v", err)
	}
}

func TestDeletePayer_SuccessWithStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers/delete/pyr_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
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

	err := c.DeletePayer(context.Background(), "pyr_123")
	if err != nil {
		t.Fatalf("DeletePayer() error: %v", err)
	}
}

func TestDeletePayer_InvalidID(t *testing.T) {
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

	err := c.DeletePayer(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty payer ID, got nil")
	}

	err = c.DeletePayer(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid payer ID, got nil")
	}
}

func TestDeletePayer_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Payer not found"
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

	err := c.DeletePayer(context.Background(), "pyr_nonexistent")
	if err == nil {
		t.Error("expected error for not found payer, got nil")
	}
}

// =====================================================
// ValidatePayer Tests
// =====================================================

func TestValidatePayer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payers/validate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"valid": true}`))
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
		"entity_type": "COMPANY",
		"name":        "Test Company",
	}

	err := c.ValidatePayer(context.Background(), req)
	if err != nil {
		t.Fatalf("ValidatePayer() error: %v", err)
	}
}

func TestValidatePayer_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_failed",
			"message": "Invalid payer data: missing required field 'entity_type'"
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
		"name": "Test Company",
		// Missing entity_type
	}

	err := c.ValidatePayer(context.Background(), req)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestValidatePayer_ServerError(t *testing.T) {
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

	req := map[string]interface{}{
		"entity_type": "COMPANY",
		"name":        "Test Company",
	}

	err := c.ValidatePayer(context.Background(), req)
	if err == nil {
		t.Error("expected server error, got nil")
	}
}
