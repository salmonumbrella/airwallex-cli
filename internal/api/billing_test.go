package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// =====================================================
// Billing Customer Tests
// =====================================================

func TestListBillingCustomers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/customers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "cust_123",
					"merchant_customer_id": "mc_001",
					"business_name": "Acme Corp",
					"first_name": "John",
					"last_name": "Doe",
					"email": "john@acme.com",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
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

	result, err := c.ListBillingCustomers(context.Background(), BillingCustomerListParams{})
	if err != nil {
		t.Fatalf("ListBillingCustomers() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].ID != "cust_123" {
		t.Errorf("id = %q, want 'cust_123'", result.Items[0].ID)
	}
	if result.Items[0].BusinessName != "Acme Corp" {
		t.Errorf("business_name = %q, want 'Acme Corp'", result.Items[0].BusinessName)
	}
}

func TestListBillingCustomers_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/customers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

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
			"items": [],
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

	result, err := c.ListBillingCustomers(context.Background(), BillingCustomerListParams{
		PageNum:  2,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListBillingCustomers() error: %v", err)
	}
	if !result.HasMore {
		t.Error("has_more = false, want true")
	}
}

func TestListBillingCustomers_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{
			"code": "server_error",
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

	_, err := c.ListBillingCustomers(context.Background(), BillingCustomerListParams{})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetBillingCustomer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/customers/cust_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "cust_123",
			"merchant_customer_id": "mc_001",
			"business_name": "Acme Corp",
			"first_name": "John",
			"last_name": "Doe",
			"email": "john@acme.com"
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

	customer, err := c.GetBillingCustomer(context.Background(), "cust_123")
	if err != nil {
		t.Fatalf("GetBillingCustomer() error: %v", err)
	}
	if customer == nil {
		t.Fatal("customer is nil")
	}
	if customer.ID != "cust_123" {
		t.Errorf("id = %q, want 'cust_123'", customer.ID)
	}
	if customer.Email != "john@acme.com" {
		t.Errorf("email = %q, want 'john@acme.com'", customer.Email)
	}
}

func TestGetBillingCustomer_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Customer not found"
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

	_, err := c.GetBillingCustomer(context.Background(), "cust_nonexistent")
	if err == nil {
		t.Error("expected error for not found customer, got nil")
	}
}

func TestGetBillingCustomer_InvalidID(t *testing.T) {
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

	_, err := c.GetBillingCustomer(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty customer ID, got nil")
	}

	_, err = c.GetBillingCustomer(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid customer ID, got nil")
	}
}

func TestCreateBillingCustomer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/customers/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "cust_new",
			"merchant_customer_id": "mc_new",
			"business_name": "New Corp",
			"email": "new@corp.com"
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
		"merchant_customer_id": "mc_new",
		"business_name":        "New Corp",
		"email":                "new@corp.com",
	}

	customer, err := c.CreateBillingCustomer(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBillingCustomer() error: %v", err)
	}
	if customer == nil {
		t.Fatal("customer is nil")
	}
	if customer.ID != "cust_new" {
		t.Errorf("id = %q, want 'cust_new'", customer.ID)
	}
}

func TestCreateBillingCustomer_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_failed",
			"message": "Email is required"
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
		"business_name": "Missing Email Corp",
	}

	_, err := c.CreateBillingCustomer(context.Background(), req)
	if err == nil {
		t.Error("expected validation error, got nil")
	}
}

func TestUpdateBillingCustomer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/pa/customers/cust_123/update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "cust_123",
			"business_name": "Updated Corp",
			"email": "updated@corp.com"
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
		"business_name": "Updated Corp",
	}

	customer, err := c.UpdateBillingCustomer(context.Background(), "cust_123", req)
	if err != nil {
		t.Fatalf("UpdateBillingCustomer() error: %v", err)
	}
	if customer.BusinessName != "Updated Corp" {
		t.Errorf("business_name = %q, want 'Updated Corp'", customer.BusinessName)
	}
}

func TestUpdateBillingCustomer_InvalidID(t *testing.T) {
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

	_, err := c.UpdateBillingCustomer(context.Background(), "", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for empty customer ID, got nil")
	}
}

// =====================================================
// Billing Product Tests
// =====================================================

func TestListBillingProducts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/products" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "prod_123",
					"name": "Premium Plan",
					"description": "Our premium subscription",
					"unit": "license",
					"active": true
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

	result, err := c.ListBillingProducts(context.Background(), BillingProductListParams{})
	if err != nil {
		t.Fatalf("ListBillingProducts() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Name != "Premium Plan" {
		t.Errorf("name = %q, want 'Premium Plan'", result.Items[0].Name)
	}
	if !result.Items[0].Active {
		t.Error("active = false, want true")
	}
}

func TestListBillingProducts_WithActiveFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		active := r.URL.Query().Get("active")
		if active != "true" {
			t.Errorf("active = %q, want 'true'", active)
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

	active := true
	_, err := c.ListBillingProducts(context.Background(), BillingProductListParams{
		Active: &active,
	})
	if err != nil {
		t.Fatalf("ListBillingProducts() error: %v", err)
	}
}

func TestGetBillingProduct_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/products/prod_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "prod_123",
			"name": "Enterprise Plan",
			"description": "Full-featured enterprise solution",
			"unit": "seat",
			"active": true
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

	product, err := c.GetBillingProduct(context.Background(), "prod_123")
	if err != nil {
		t.Fatalf("GetBillingProduct() error: %v", err)
	}
	if product.ID != "prod_123" {
		t.Errorf("id = %q, want 'prod_123'", product.ID)
	}
	if product.Unit != "seat" {
		t.Errorf("unit = %q, want 'seat'", product.Unit)
	}
}

func TestGetBillingProduct_InvalidID(t *testing.T) {
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

	_, err := c.GetBillingProduct(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty product ID, got nil")
	}
}

func TestCreateBillingProduct_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/products/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "prod_new",
			"name": "Basic Plan",
			"active": true
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
		"name": "Basic Plan",
	}

	product, err := c.CreateBillingProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBillingProduct() error: %v", err)
	}
	if product.ID != "prod_new" {
		t.Errorf("id = %q, want 'prod_new'", product.ID)
	}
}

func TestUpdateBillingProduct_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/products/prod_123/update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "prod_123",
			"name": "Updated Plan",
			"active": false
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
		"active": false,
	}

	product, err := c.UpdateBillingProduct(context.Background(), "prod_123", req)
	if err != nil {
		t.Fatalf("UpdateBillingProduct() error: %v", err)
	}
	if product.Active {
		t.Error("active = true, want false")
	}
}

// =====================================================
// Billing Price Tests
// =====================================================

func TestListBillingPrices_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/prices" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "price_123",
					"product_id": "prod_456",
					"currency": "USD",
					"unit_amount": 99.99,
					"pricing_model": "per_unit",
					"type": "recurring",
					"active": true,
					"recurring": {
						"period": 1,
						"period_unit": "MONTH"
					}
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

	result, err := c.ListBillingPrices(context.Background(), BillingPriceListParams{})
	if err != nil {
		t.Fatalf("ListBillingPrices() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].UnitAmount != jn("99.99") {
		t.Errorf("unit_amount = %s, want 99.99", result.Items[0].UnitAmount)
	}
	if result.Items[0].Recurring == nil {
		t.Fatal("recurring is nil")
	}
	if result.Items[0].Recurring.Period != 1 {
		t.Errorf("recurring.period = %d, want 1", result.Items[0].Recurring.Period)
	}
}

func TestListBillingPrices_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currency := r.URL.Query().Get("currency")
		productID := r.URL.Query().Get("product_id")

		if currency != "EUR" {
			t.Errorf("currency = %q, want 'EUR'", currency)
		}
		if productID != "prod_abc" {
			t.Errorf("product_id = %q, want 'prod_abc'", productID)
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

	_, err := c.ListBillingPrices(context.Background(), BillingPriceListParams{
		Currency:  "EUR",
		ProductID: "prod_abc",
	})
	if err != nil {
		t.Fatalf("ListBillingPrices() error: %v", err)
	}
}

func TestGetBillingPrice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/prices/price_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "price_123",
			"product_id": "prod_456",
			"currency": "USD",
			"unit_amount": 49.99,
			"type": "one_time",
			"active": true
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

	price, err := c.GetBillingPrice(context.Background(), "price_123")
	if err != nil {
		t.Fatalf("GetBillingPrice() error: %v", err)
	}
	if price.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", price.Currency)
	}
}

func TestGetBillingPrice_InvalidID(t *testing.T) {
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

	_, err := c.GetBillingPrice(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty price ID, got nil")
	}
}

func TestCreateBillingPrice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/prices/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "price_new",
			"product_id": "prod_123",
			"currency": "GBP",
			"unit_amount": 75.00
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
		"product_id":  "prod_123",
		"currency":    "GBP",
		"unit_amount": 75.00,
	}

	price, err := c.CreateBillingPrice(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBillingPrice() error: %v", err)
	}
	if price.ID != "price_new" {
		t.Errorf("id = %q, want 'price_new'", price.ID)
	}
}

func TestUpdateBillingPrice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/prices/price_123/update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "price_123",
			"active": false
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
		"active": false,
	}

	price, err := c.UpdateBillingPrice(context.Background(), "price_123", req)
	if err != nil {
		t.Fatalf("UpdateBillingPrice() error: %v", err)
	}
	if price.Active {
		t.Error("active = true, want false")
	}
}

// =====================================================
// Billing Invoice Tests
// =====================================================

func TestListBillingInvoices_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/invoices" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "inv_123",
					"customer_id": "cust_456",
					"subscription_id": "sub_789",
					"status": "PAID",
					"currency": "USD",
					"total_amount": 199.99,
					"period_start_at": "2024-01-01T00:00:00Z",
					"period_end_at": "2024-02-01T00:00:00Z",
					"paid_at": "2024-01-05T00:00:00Z"
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

	result, err := c.ListBillingInvoices(context.Background(), BillingInvoiceListParams{})
	if err != nil {
		t.Fatalf("ListBillingInvoices() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Status != "PAID" {
		t.Errorf("status = %q, want 'PAID'", result.Items[0].Status)
	}
	if result.Items[0].TotalAmount != jn("199.99") {
		t.Errorf("total_amount = %s, want 199.99", result.Items[0].TotalAmount)
	}
}

func TestListBillingInvoices_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerID := r.URL.Query().Get("customer_id")
		status := r.URL.Query().Get("status")

		if customerID != "cust_abc" {
			t.Errorf("customer_id = %q, want 'cust_abc'", customerID)
		}
		if status != "PENDING" {
			t.Errorf("status = %q, want 'PENDING'", status)
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

	_, err := c.ListBillingInvoices(context.Background(), BillingInvoiceListParams{
		CustomerID: "cust_abc",
		Status:     "PENDING",
	})
	if err != nil {
		t.Fatalf("ListBillingInvoices() error: %v", err)
	}
}

func TestGetBillingInvoice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/invoices/inv_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "inv_123",
			"customer_id": "cust_456",
			"status": "PAID",
			"currency": "EUR",
			"total_amount": 299.00
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

	invoice, err := c.GetBillingInvoice(context.Background(), "inv_123")
	if err != nil {
		t.Fatalf("GetBillingInvoice() error: %v", err)
	}
	if invoice.ID != "inv_123" {
		t.Errorf("id = %q, want 'inv_123'", invoice.ID)
	}
}

func TestGetBillingInvoice_InvalidID(t *testing.T) {
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

	_, err := c.GetBillingInvoice(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty invoice ID, got nil")
	}
}

func TestCreateBillingInvoice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/invoices/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "inv_new",
			"customer_id": "cust_123",
			"status": "DRAFT",
			"currency": "USD",
			"total_amount": 150.00
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
		"customer_id": "cust_123",
		"currency":    "USD",
	}

	invoice, err := c.CreateBillingInvoice(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBillingInvoice() error: %v", err)
	}
	if invoice.ID != "inv_new" {
		t.Errorf("id = %q, want 'inv_new'", invoice.ID)
	}
}

func TestPreviewBillingInvoice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/invoices/preview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"customer_id": "cust_123",
			"subscription_id": "sub_456",
			"currency": "USD",
			"total_amount": 99.99,
			"items": [
				{
					"id": "item_1",
					"amount": 99.99,
					"quantity": 1.0
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

	req := map[string]interface{}{
		"subscription_id": "sub_456",
	}

	preview, err := c.PreviewBillingInvoice(context.Background(), req)
	if err != nil {
		t.Fatalf("PreviewBillingInvoice() error: %v", err)
	}
	if preview.TotalAmount != jn("99.99") {
		t.Errorf("total_amount = %s, want 99.99", preview.TotalAmount)
	}
	if len(preview.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(preview.Items))
	}
}

func TestListBillingInvoiceItems_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/invoices/inv_123/items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "item_1",
					"invoice_id": "inv_123",
					"amount": 50.00,
					"currency": "USD",
					"quantity": 2.0
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

	result, err := c.ListBillingInvoiceItems(context.Background(), "inv_123", 0, 0)
	if err != nil {
		t.Fatalf("ListBillingInvoiceItems() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Amount != jn("50.00") {
		t.Errorf("amount = %s, want 50.00", result.Items[0].Amount)
	}
}

func TestListBillingInvoiceItems_InvalidID(t *testing.T) {
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

	_, err := c.ListBillingInvoiceItems(context.Background(), "", 0, 0)
	if err == nil {
		t.Error("expected error for empty invoice ID, got nil")
	}
}

func TestGetBillingInvoiceItem_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/invoices/inv_123/items/item_456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "item_456",
			"invoice_id": "inv_123",
			"amount": 25.00,
			"quantity": 1.0
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

	item, err := c.GetBillingInvoiceItem(context.Background(), "inv_123", "item_456")
	if err != nil {
		t.Fatalf("GetBillingInvoiceItem() error: %v", err)
	}
	if item.ID != "item_456" {
		t.Errorf("id = %q, want 'item_456'", item.ID)
	}
}

func TestGetBillingInvoiceItem_InvalidIDs(t *testing.T) {
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

	_, err := c.GetBillingInvoiceItem(context.Background(), "", "item_456")
	if err == nil {
		t.Error("expected error for empty invoice ID, got nil")
	}

	_, err = c.GetBillingInvoiceItem(context.Background(), "inv_123", "")
	if err == nil {
		t.Error("expected error for empty item ID, got nil")
	}
}

// =====================================================
// Billing Subscription Tests
// =====================================================

func TestListBillingSubscriptions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscriptions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "sub_123",
					"customer_id": "cust_456",
					"status": "ACTIVE",
					"current_period_start_at": "2024-01-01T00:00:00Z",
					"current_period_end_at": "2024-02-01T00:00:00Z",
					"next_billing_at": "2024-02-01T00:00:00Z",
					"cancel_at_period_end": false,
					"remaining_billing_cycles": 12,
					"total_billing_cycles": 12
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

	result, err := c.ListBillingSubscriptions(context.Background(), BillingSubscriptionListParams{})
	if err != nil {
		t.Fatalf("ListBillingSubscriptions() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", result.Items[0].Status)
	}
	if result.Items[0].CancelAtPeriodEnd {
		t.Error("cancel_at_period_end = true, want false")
	}
}

func TestListBillingSubscriptions_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerID := r.URL.Query().Get("customer_id")
		status := r.URL.Query().Get("status")
		recurringPeriod := r.URL.Query().Get("recurring_period")

		if customerID != "cust_abc" {
			t.Errorf("customer_id = %q, want 'cust_abc'", customerID)
		}
		if status != "ACTIVE" {
			t.Errorf("status = %q, want 'ACTIVE'", status)
		}
		if recurringPeriod != "1" {
			t.Errorf("recurring_period = %q, want '1'", recurringPeriod)
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

	_, err := c.ListBillingSubscriptions(context.Background(), BillingSubscriptionListParams{
		CustomerID:      "cust_abc",
		Status:          "ACTIVE",
		RecurringPeriod: 1,
	})
	if err != nil {
		t.Fatalf("ListBillingSubscriptions() error: %v", err)
	}
}

func TestGetBillingSubscription_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscriptions/sub_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "sub_123",
			"customer_id": "cust_456",
			"status": "ACTIVE",
			"latest_invoice_id": "inv_789"
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

	sub, err := c.GetBillingSubscription(context.Background(), "sub_123")
	if err != nil {
		t.Fatalf("GetBillingSubscription() error: %v", err)
	}
	if sub.ID != "sub_123" {
		t.Errorf("id = %q, want 'sub_123'", sub.ID)
	}
	if sub.LatestInvoiceID != "inv_789" {
		t.Errorf("latest_invoice_id = %q, want 'inv_789'", sub.LatestInvoiceID)
	}
}

func TestGetBillingSubscription_InvalidID(t *testing.T) {
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

	_, err := c.GetBillingSubscription(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty subscription ID, got nil")
	}
}

func TestCreateBillingSubscription_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscriptions/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "sub_new",
			"customer_id": "cust_123",
			"status": "ACTIVE"
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
		"customer_id": "cust_123",
		"items": []map[string]interface{}{
			{"price_id": "price_456", "quantity": 1},
		},
	}

	sub, err := c.CreateBillingSubscription(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateBillingSubscription() error: %v", err)
	}
	if sub.ID != "sub_new" {
		t.Errorf("id = %q, want 'sub_new'", sub.ID)
	}
}

func TestUpdateBillingSubscription_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscriptions/sub_123/update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "sub_123",
			"cancel_at_period_end": true
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
		"cancel_at_period_end": true,
	}

	sub, err := c.UpdateBillingSubscription(context.Background(), "sub_123", req)
	if err != nil {
		t.Fatalf("UpdateBillingSubscription() error: %v", err)
	}
	if !sub.CancelAtPeriodEnd {
		t.Error("cancel_at_period_end = false, want true")
	}
}

func TestUpdateBillingSubscription_InvalidID(t *testing.T) {
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

	_, err := c.UpdateBillingSubscription(context.Background(), "", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for empty subscription ID, got nil")
	}
}

func TestCancelBillingSubscription_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscriptions/sub_123/cancel" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "sub_123",
			"status": "CANCELLED",
			"cancel_at": "2024-02-01T00:00:00Z"
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
		"cancel_at_period_end": true,
	}

	sub, err := c.CancelBillingSubscription(context.Background(), "sub_123", req)
	if err != nil {
		t.Fatalf("CancelBillingSubscription() error: %v", err)
	}
	if sub.Status != "CANCELLED" {
		t.Errorf("status = %q, want 'CANCELLED'", sub.Status)
	}
}

func TestCancelBillingSubscription_InvalidID(t *testing.T) {
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

	_, err := c.CancelBillingSubscription(context.Background(), "", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for empty subscription ID, got nil")
	}
}

func TestListBillingSubscriptionItems_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscriptions/sub_123/items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "si_123",
					"subscription_id": "sub_123",
					"quantity": 5.0,
					"price": {
						"id": "price_456",
						"product_id": "prod_789",
						"currency": "USD",
						"unit_amount": 10.00
					}
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

	result, err := c.ListBillingSubscriptionItems(context.Background(), "sub_123", 0, 0)
	if err != nil {
		t.Fatalf("ListBillingSubscriptionItems() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Quantity != jn("5.0") {
		t.Errorf("quantity = %s, want 5.0", result.Items[0].Quantity)
	}
	if result.Items[0].Price == nil {
		t.Fatal("price is nil")
	}
	if result.Items[0].Price.UnitAmount != jn("10.00") {
		t.Errorf("price.unit_amount = %s, want 10.00", result.Items[0].Price.UnitAmount)
	}
}

func TestListBillingSubscriptionItems_InvalidID(t *testing.T) {
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

	_, err := c.ListBillingSubscriptionItems(context.Background(), "", 0, 0)
	if err == nil {
		t.Error("expected error for empty subscription ID, got nil")
	}
}

func TestGetBillingSubscriptionItem_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subscriptions/sub_123/items/si_456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "si_456",
			"subscription_id": "sub_123",
			"quantity": 2.0
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

	item, err := c.GetBillingSubscriptionItem(context.Background(), "sub_123", "si_456")
	if err != nil {
		t.Fatalf("GetBillingSubscriptionItem() error: %v", err)
	}
	if item.ID != "si_456" {
		t.Errorf("id = %q, want 'si_456'", item.ID)
	}
}

func TestGetBillingSubscriptionItem_InvalidIDs(t *testing.T) {
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

	_, err := c.GetBillingSubscriptionItem(context.Background(), "", "si_456")
	if err == nil {
		t.Error("expected error for empty subscription ID, got nil")
	}

	_, err = c.GetBillingSubscriptionItem(context.Background(), "sub_123", "")
	if err == nil {
		t.Error("expected error for empty item ID, got nil")
	}
}
