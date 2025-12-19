package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =====================================================
// Beneficiary Tests
// =====================================================

func TestListBeneficiaries_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters are correctly set
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "2" {
			t.Errorf("page_num = %q, want '2'", pageNum)
		}
		if pageSize != "20" {
			t.Errorf("page_size = %q, want '20'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "ben_123",
					"nickname": "Test Company",
					"beneficiary": {
						"entity_type": "COMPANY",
						"company_name": "Test Corp",
						"bank_details": {
							"bank_country_code": "US",
							"bank_name": "Test Bank",
							"account_name": "Test Account"
						}
					},
					"transfer_methods": ["SWIFT", "LOCAL"]
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

	result, err := c.ListBeneficiaries(context.Background(), 2, 20)
	if err != nil {
		t.Fatalf("ListBeneficiaries() error: %v", err)
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
	if result.Items[0].BeneficiaryID != "ben_123" {
		t.Errorf("beneficiary_id = %q, want 'ben_123'", result.Items[0].BeneficiaryID)
	}
	if result.Items[0].Nickname != "Test Company" {
		t.Errorf("nickname = %q, want 'Test Company'", result.Items[0].Nickname)
	}
	if result.Items[0].Beneficiary.EntityType != "COMPANY" {
		t.Errorf("entity_type = %q, want 'COMPANY'", result.Items[0].Beneficiary.EntityType)
	}
	if len(result.Items[0].TransferMethods) != 2 {
		t.Errorf("transfer_methods count = %d, want 2", len(result.Items[0].TransferMethods))
	}
}

func TestListBeneficiaries_WithoutPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries" {
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

	result, err := c.ListBeneficiaries(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListBeneficiaries() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
}

func TestGetBeneficiary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries/ben_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "ben_123",
			"nickname": "Primary Supplier",
			"beneficiary": {
				"entity_type": "INDIVIDUAL",
				"first_name": "John",
				"last_name": "Doe",
				"bank_details": {
					"bank_country_code": "GB",
					"bank_name": "HSBC",
					"account_name": "John Doe"
				}
			},
			"transfer_methods": ["LOCAL"]
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

	beneficiary, err := c.GetBeneficiary(context.Background(), "ben_123")
	if err != nil {
		t.Fatalf("GetBeneficiary() error: %v", err)
	}
	if beneficiary == nil {
		t.Fatal("beneficiary is nil")
	}
	if beneficiary.BeneficiaryID != "ben_123" {
		t.Errorf("beneficiary_id = %q, want 'ben_123'", beneficiary.BeneficiaryID)
	}
	if beneficiary.Beneficiary.EntityType != "INDIVIDUAL" {
		t.Errorf("entity_type = %q, want 'INDIVIDUAL'", beneficiary.Beneficiary.EntityType)
	}
	if beneficiary.Beneficiary.FirstName != "John" {
		t.Errorf("first_name = %q, want 'John'", beneficiary.Beneficiary.FirstName)
	}
}

func TestGetBeneficiary_InvalidID(t *testing.T) {
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

	_, err := c.GetBeneficiary(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty beneficiary ID, got nil")
	}

	_, err = c.GetBeneficiary(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid beneficiary ID, got nil")
	}
}

func TestGetBeneficiaryRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries/ben_456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "ben_456",
			"nickname": "Vendor",
			"custom_field": "custom_value",
			"beneficiary": {
				"entity_type": "COMPANY",
				"company_name": "Acme Corp"
			}
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

	result, err := c.GetBeneficiaryRaw(context.Background(), "ben_456")
	if err != nil {
		t.Fatalf("GetBeneficiaryRaw() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if result["id"] != "ben_456" {
		t.Errorf("id = %q, want 'ben_456'", result["id"])
	}
	if result["custom_field"] != "custom_value" {
		t.Errorf("custom_field = %q, want 'custom_value'", result["custom_field"])
	}
}

func TestGetBeneficiaryRaw_InvalidID(t *testing.T) {
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

	_, err := c.GetBeneficiaryRaw(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty beneficiary ID, got nil")
	}
}

func TestCreateBeneficiary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "ben_new",
			"nickname": "New Beneficiary",
			"beneficiary": {
				"entity_type": "COMPANY",
				"company_name": "New Corp",
				"bank_details": {
					"bank_country_code": "US",
					"bank_name": "Bank of America",
					"account_name": "New Corp"
				}
			},
			"transfer_methods": ["SWIFT"]
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
		"nickname": "New Beneficiary",
		"beneficiary": map[string]interface{}{
			"entity_type":  "COMPANY",
			"company_name": "New Corp",
		},
	}

	beneficiary, err := c.CreateBeneficiary(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBeneficiary() error: %v", err)
	}
	if beneficiary == nil {
		t.Fatal("beneficiary is nil")
	}
	if beneficiary.BeneficiaryID != "ben_new" {
		t.Errorf("beneficiary_id = %q, want 'ben_new'", beneficiary.BeneficiaryID)
	}
	if beneficiary.Nickname != "New Beneficiary" {
		t.Errorf("nickname = %q, want 'New Beneficiary'", beneficiary.Nickname)
	}
}

func TestCreateBeneficiary_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_failed",
			"message": "Invalid beneficiary data"
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

	_, err := c.CreateBeneficiary(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid beneficiary data, got nil")
	}
}

func TestUpdateBeneficiary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries/ben_123/update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "ben_123",
			"nickname": "Updated Nickname",
			"beneficiary": {
				"entity_type": "COMPANY",
				"company_name": "Updated Corp",
				"bank_details": {
					"bank_country_code": "US",
					"bank_name": "Test Bank",
					"account_name": "Updated Corp"
				}
			},
			"transfer_methods": ["SWIFT"]
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
		"nickname": "Updated Nickname",
	}

	beneficiary, err := c.UpdateBeneficiary(context.Background(), "ben_123", update)
	if err != nil {
		t.Fatalf("UpdateBeneficiary() error: %v", err)
	}
	if beneficiary == nil {
		t.Fatal("beneficiary is nil")
	}
	if beneficiary.BeneficiaryID != "ben_123" {
		t.Errorf("beneficiary_id = %q, want 'ben_123'", beneficiary.BeneficiaryID)
	}
	if beneficiary.Nickname != "Updated Nickname" {
		t.Errorf("nickname = %q, want 'Updated Nickname'", beneficiary.Nickname)
	}
}

func TestUpdateBeneficiary_InvalidID(t *testing.T) {
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
		"nickname": "New Name",
	}

	_, err := c.UpdateBeneficiary(context.Background(), "", update)
	if err == nil {
		t.Error("expected error for empty beneficiary ID, got nil")
	}

	_, err = c.UpdateBeneficiary(context.Background(), "invalid/id", update)
	if err == nil {
		t.Error("expected error for invalid beneficiary ID, got nil")
	}
}

func TestUpdateBeneficiary_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Beneficiary not found"
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
		"nickname": "New Name",
	}

	_, err := c.UpdateBeneficiary(context.Background(), "ben_nonexistent", update)
	if err == nil {
		t.Error("expected error for not found beneficiary, got nil")
	}
}

func TestDeleteBeneficiary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries/ben_123/delete" {
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

	err := c.DeleteBeneficiary(context.Background(), "ben_123")
	if err != nil {
		t.Fatalf("DeleteBeneficiary() error: %v", err)
	}
}

func TestDeleteBeneficiary_SuccessWithStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries/ben_123/delete" {
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

	err := c.DeleteBeneficiary(context.Background(), "ben_123")
	if err != nil {
		t.Fatalf("DeleteBeneficiary() error: %v", err)
	}
}

func TestDeleteBeneficiary_InvalidID(t *testing.T) {
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

	err := c.DeleteBeneficiary(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty beneficiary ID, got nil")
	}

	err = c.DeleteBeneficiary(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid beneficiary ID, got nil")
	}
}

func TestDeleteBeneficiary_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Beneficiary not found"
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

	err := c.DeleteBeneficiary(context.Background(), "ben_nonexistent")
	if err == nil {
		t.Error("expected error for not found beneficiary, got nil")
	}
}

func TestValidateBeneficiary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiaries/validate" {
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
		"beneficiary": map[string]interface{}{
			"entity_type":  "COMPANY",
			"company_name": "Test Corp",
		},
	}

	err := c.ValidateBeneficiary(context.Background(), req)
	if err != nil {
		t.Fatalf("ValidateBeneficiary() error: %v", err)
	}
}

func TestValidateBeneficiary_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_failed",
			"message": "Invalid bank details"
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
		"beneficiary": map[string]interface{}{
			"entity_type": "INVALID",
		},
	}

	err := c.ValidateBeneficiary(context.Background(), req)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

// =====================================================
// Transfer Tests
// =====================================================

func TestListTransfers_WithAllParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transfers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters are correctly set
		status := r.URL.Query().Get("status")
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if status != "COMPLETED" {
			t.Errorf("status = %q, want 'COMPLETED'", status)
		}
		if pageNum != "1" {
			t.Errorf("page_num = %q, want '1'", pageNum)
		}
		if pageSize != "10" {
			t.Errorf("page_size = %q, want '10'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "tfr_123",
					"beneficiary_id": "ben_456",
					"transfer_amount": 1000.00,
					"transfer_currency": "USD",
					"source_amount": 1000.00,
					"source_currency": "USD",
					"status": "COMPLETED",
					"reference": "Invoice 123",
					"reason": "Payment for services",
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

	result, err := c.ListTransfers(context.Background(), "COMPLETED", 1, 10)
	if err != nil {
		t.Fatalf("ListTransfers() error: %v", err)
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
	if result.Items[0].TransferID != "tfr_123" {
		t.Errorf("transfer_id = %q, want 'tfr_123'", result.Items[0].TransferID)
	}
	if result.Items[0].Status != "COMPLETED" {
		t.Errorf("status = %q, want 'COMPLETED'", result.Items[0].Status)
	}
	if result.Items[0].TransferAmount != 1000.00 {
		t.Errorf("transfer_amount = %f, want 1000.00", result.Items[0].TransferAmount)
	}
}

func TestListTransfers_WithoutParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transfers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify no query parameters are set
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

	result, err := c.ListTransfers(context.Background(), "", 0, 0)
	if err != nil {
		t.Fatalf("ListTransfers() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
}

func TestListTransfers_WithStatusOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify only status is set
		status := r.URL.Query().Get("status")
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if status != "PENDING" {
			t.Errorf("status = %q, want 'PENDING'", status)
		}
		if pageNum != "" {
			t.Errorf("page_num = %q, want empty", pageNum)
		}
		if pageSize != "" {
			t.Errorf("page_size = %q, want empty", pageSize)
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

	_, err := c.ListTransfers(context.Background(), "PENDING", 0, 0)
	if err != nil {
		t.Fatalf("ListTransfers() error: %v", err)
	}
}

func TestGetTransfer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transfers/tfr_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "tfr_123",
			"beneficiary_id": "ben_456",
			"transfer_amount": 2500.50,
			"transfer_currency": "GBP",
			"source_amount": 3000.00,
			"source_currency": "USD",
			"status": "PROCESSING",
			"reference": "REF-001",
			"reason": "Supplier payment",
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

	transfer, err := c.GetTransfer(context.Background(), "tfr_123")
	if err != nil {
		t.Fatalf("GetTransfer() error: %v", err)
	}
	if transfer == nil {
		t.Fatal("transfer is nil")
	}
	if transfer.TransferID != "tfr_123" {
		t.Errorf("transfer_id = %q, want 'tfr_123'", transfer.TransferID)
	}
	if transfer.BeneficiaryID != "ben_456" {
		t.Errorf("beneficiary_id = %q, want 'ben_456'", transfer.BeneficiaryID)
	}
	if transfer.TransferAmount != 2500.50 {
		t.Errorf("transfer_amount = %f, want 2500.50", transfer.TransferAmount)
	}
	if transfer.TransferCurrency != "GBP" {
		t.Errorf("transfer_currency = %q, want 'GBP'", transfer.TransferCurrency)
	}
	if transfer.Status != "PROCESSING" {
		t.Errorf("status = %q, want 'PROCESSING'", transfer.Status)
	}
}

func TestGetTransfer_InvalidID(t *testing.T) {
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

	_, err := c.GetTransfer(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty transfer ID, got nil")
	}

	_, err = c.GetTransfer(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid transfer ID, got nil")
	}
}

func TestGetTransfer_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Transfer not found"
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

	_, err := c.GetTransfer(context.Background(), "tfr_nonexistent")
	if err == nil {
		t.Error("expected error for not found transfer, got nil")
	}
}

func TestCreateTransfer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transfers/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "tfr_new",
			"beneficiary_id": "ben_789",
			"transfer_amount": 5000.00,
			"transfer_currency": "EUR",
			"source_amount": 5000.00,
			"source_currency": "EUR",
			"status": "PENDING",
			"reference": "NEW-REF",
			"reason": "New payment",
			"created_at": "2024-01-20T12:00:00Z"
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
		"beneficiary_id":    "ben_789",
		"transfer_amount":   5000.00,
		"transfer_currency": "EUR",
		"reference":         "NEW-REF",
		"reason":            "New payment",
	}

	transfer, err := c.CreateTransfer(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTransfer() error: %v", err)
	}
	if transfer == nil {
		t.Fatal("transfer is nil")
	}
	if transfer.TransferID != "tfr_new" {
		t.Errorf("transfer_id = %q, want 'tfr_new'", transfer.TransferID)
	}
	if transfer.BeneficiaryID != "ben_789" {
		t.Errorf("beneficiary_id = %q, want 'ben_789'", transfer.BeneficiaryID)
	}
	if transfer.Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", transfer.Status)
	}
	if transfer.TransferAmount != 5000.00 {
		t.Errorf("transfer_amount = %f, want 5000.00", transfer.TransferAmount)
	}
}

func TestCreateTransfer_SuccessWithStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "tfr_ok",
			"beneficiary_id": "ben_111",
			"transfer_amount": 100.00,
			"transfer_currency": "USD",
			"source_amount": 100.00,
			"source_currency": "USD",
			"status": "PENDING",
			"reference": "OK-REF",
			"reason": "Test",
			"created_at": "2024-01-20T12:00:00Z"
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
		"beneficiary_id":    "ben_111",
		"transfer_amount":   100.00,
		"transfer_currency": "USD",
	}

	transfer, err := c.CreateTransfer(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTransfer() error: %v", err)
	}
	if transfer == nil {
		t.Fatal("transfer is nil")
	}
	if transfer.TransferID != "tfr_ok" {
		t.Errorf("transfer_id = %q, want 'tfr_ok'", transfer.TransferID)
	}
}

func TestCreateTransfer_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_failed",
			"message": "Invalid transfer data"
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

	_, err := c.CreateTransfer(context.Background(), req)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestCancelTransfer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transfers/tfr_123/cancel" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "tfr_123",
			"beneficiary_id": "ben_456",
			"transfer_amount": 1000.00,
			"transfer_currency": "USD",
			"source_amount": 1000.00,
			"source_currency": "USD",
			"status": "CANCELLED",
			"reference": "REF-001",
			"reason": "Payment cancelled",
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

	transfer, err := c.CancelTransfer(context.Background(), "tfr_123")
	if err != nil {
		t.Fatalf("CancelTransfer() error: %v", err)
	}
	if transfer == nil {
		t.Fatal("transfer is nil")
	}
	if transfer.TransferID != "tfr_123" {
		t.Errorf("transfer_id = %q, want 'tfr_123'", transfer.TransferID)
	}
	if transfer.Status != "CANCELLED" {
		t.Errorf("status = %q, want 'CANCELLED'", transfer.Status)
	}
}

func TestCancelTransfer_InvalidID(t *testing.T) {
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

	_, err := c.CancelTransfer(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty transfer ID, got nil")
	}

	_, err = c.CancelTransfer(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid transfer ID, got nil")
	}
}

func TestCancelTransfer_AlreadyCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_operation",
			"message": "Transfer already cancelled"
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

	_, err := c.CancelTransfer(context.Background(), "tfr_cancelled")
	if err == nil {
		t.Error("expected error for already cancelled transfer, got nil")
	}
}

// =====================================================
// Confirmation Letter Tests
// =====================================================

func TestGetConfirmationLetter_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/confirmation_letters/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("PDF-CONTENT-HERE"))
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

	pdfData, err := c.GetConfirmationLetter(context.Background(), "tfr_123", "pdf")
	if err != nil {
		t.Fatalf("GetConfirmationLetter() error: %v", err)
	}
	if pdfData == nil {
		t.Fatal("pdfData is nil")
	}
	if string(pdfData) != "PDF-CONTENT-HERE" {
		t.Errorf("pdf content = %q, want 'PDF-CONTENT-HERE'", string(pdfData))
	}
}

func TestGetConfirmationLetter_SuccessWithStatus200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PDF-CONTENT"))
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

	pdfData, err := c.GetConfirmationLetter(context.Background(), "tfr_456", "pdf")
	if err != nil {
		t.Fatalf("GetConfirmationLetter() error: %v", err)
	}
	if len(pdfData) == 0 {
		t.Fatal("pdfData is empty")
	}
}

func TestGetConfirmationLetter_InvalidID(t *testing.T) {
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

	_, err := c.GetConfirmationLetter(context.Background(), "", "pdf")
	if err == nil {
		t.Error("expected error for empty transfer ID, got nil")
	}

	_, err = c.GetConfirmationLetter(context.Background(), "invalid/id", "pdf")
	if err == nil {
		t.Error("expected error for invalid transfer ID, got nil")
	}
}

func TestGetConfirmationLetter_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Transfer not found"
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

	_, err := c.GetConfirmationLetter(context.Background(), "tfr_nonexistent", "pdf")
	if err == nil {
		t.Error("expected error for not found transfer, got nil")
	}
}
