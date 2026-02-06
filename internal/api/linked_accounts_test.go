package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestListLinkedAccounts_WithPagination verifies pagination parameters are correctly set
func TestListLinkedAccounts_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/linked_accounts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters are correctly set
		pageNum := r.URL.Query().Get("page_num")
		pageSize := r.URL.Query().Get("page_size")

		if pageNum != "2" {
			t.Errorf("page_num = %q, want '2'", pageNum)
		}
		if pageSize != "15" {
			t.Errorf("page_size = %q, want '15'", pageSize)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"id": "la_123",
					"type": "bank_account",
					"status": "ACTIVE",
					"account_name": "Business Account",
					"bank_name": "Chase Bank",
					"account_number_last4": "1234",
					"currency": "USD",
					"created_at": "2024-01-01T00:00:00Z"
				},
				{
					"id": "la_456",
					"type": "bank_account",
					"status": "PENDING",
					"account_name": "Savings Account",
					"bank_name": "Wells Fargo",
					"account_number_last4": "5678",
					"currency": "USD",
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

	result, err := c.ListLinkedAccounts(context.Background(), 2, 15)
	if err != nil {
		t.Fatalf("ListLinkedAccounts() error: %v", err)
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
	if result.Items[0].ID != "la_123" {
		t.Errorf("items[0].id = %q, want 'la_123'", result.Items[0].ID)
	}
	if result.Items[0].Currency != "USD" {
		t.Errorf("items[0].currency = %q, want 'USD'", result.Items[0].Currency)
	}
	if result.Items[1].ID != "la_456" {
		t.Errorf("items[1].id = %q, want 'la_456'", result.Items[1].ID)
	}
}

// TestListLinkedAccounts_WithoutPagination verifies behavior with zero/negative values
func TestListLinkedAccounts_WithoutPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/linked_accounts" {
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

	result, err := c.ListLinkedAccounts(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListLinkedAccounts() error: %v", err)
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

// TestListLinkedAccounts_EmptyResults verifies handling of empty results
func TestListLinkedAccounts_EmptyResults(t *testing.T) {
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

	result, err := c.ListLinkedAccounts(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListLinkedAccounts() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
}

// TestListLinkedAccounts_APIError verifies error handling for API errors
func TestListLinkedAccounts_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{
			"code": "unauthorized",
			"message": "Invalid API credentials"
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

	_, err := c.ListLinkedAccounts(context.Background(), 1, 10)
	if err == nil {
		t.Error("expected error for API error response, got nil")
	}
}

// TestGetLinkedAccount_Success verifies successful account retrieval
func TestGetLinkedAccount_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/linked_accounts/la_test123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "la_test123",
			"type": "bank_account",
			"status": "ACTIVE",
			"account_name": "Business Checking",
			"bank_name": "Bank of America",
			"account_number_last4": "9876",
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

	account, err := c.GetLinkedAccount(context.Background(), "la_test123")
	if err != nil {
		t.Fatalf("GetLinkedAccount() error: %v", err)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	if account.ID != "la_test123" {
		t.Errorf("id = %q, want 'la_test123'", account.ID)
	}
	if account.Type != "bank_account" {
		t.Errorf("type = %q, want 'bank_account'", account.Type)
	}
	if account.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", account.Status)
	}
	if account.AccountName != "Business Checking" {
		t.Errorf("account_name = %q, want 'Business Checking'", account.AccountName)
	}
	if account.BankName != "Bank of America" {
		t.Errorf("bank_name = %q, want 'Bank of America'", account.BankName)
	}
	if account.AccountNumber != "9876" {
		t.Errorf("account_number_last4 = %q, want '9876'", account.AccountNumber)
	}
	if account.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", account.Currency)
	}
}

// TestGetLinkedAccount_InvalidID verifies validation of account IDs
func TestGetLinkedAccount_InvalidID(t *testing.T) {
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
	_, err := c.GetLinkedAccount(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty account ID, got nil")
	}

	// Test ID with invalid characters
	_, err = c.GetLinkedAccount(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid account ID, got nil")
	}

	// Test ID with spaces
	_, err = c.GetLinkedAccount(context.Background(), "la 123")
	if err == nil {
		t.Error("expected error for account ID with spaces, got nil")
	}
}

// TestGetLinkedAccount_NotFound verifies handling of 404 errors
func TestGetLinkedAccount_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Linked account not found"
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

	_, err := c.GetLinkedAccount(context.Background(), "la_nonexistent")
	if err == nil {
		t.Error("expected error for not found account, got nil")
	}
}

// TestCreateLinkedAccount_Success verifies successful account creation
func TestCreateLinkedAccount_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/linked_accounts/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "la_new123",
			"type": "bank_account",
			"status": "PENDING",
			"account_name": "New Business Account",
			"bank_name": "Citibank",
			"account_number_last4": "4321",
			"currency": "USD",
			"created_at": "2024-01-20T15:45:00Z"
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
		"account_name":   "New Business Account",
		"bank_name":      "Citibank",
		"account_number": "12344321",
		"routing_number": "111000025",
		"currency":       "USD",
	}

	account, err := c.CreateLinkedAccount(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateLinkedAccount() error: %v", err)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	if account.ID != "la_new123" {
		t.Errorf("id = %q, want 'la_new123'", account.ID)
	}
	if account.Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", account.Status)
	}
	if account.AccountName != "New Business Account" {
		t.Errorf("account_name = %q, want 'New Business Account'", account.AccountName)
	}
}

// TestCreateLinkedAccount_MinimalParams verifies creation with minimal parameters
func TestCreateLinkedAccount_MinimalParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "la_minimal",
			"type": "bank_account",
			"status": "PENDING",
			"account_name": "Minimal Account",
			"currency": "EUR",
			"created_at": "2024-01-20T16:00:00Z"
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
		"account_name": "Minimal Account",
		"currency":     "EUR",
	}

	account, err := c.CreateLinkedAccount(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateLinkedAccount() error: %v", err)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	if account.ID != "la_minimal" {
		t.Errorf("id = %q, want 'la_minimal'", account.ID)
	}
	if account.Currency != "EUR" {
		t.Errorf("currency = %q, want 'EUR'", account.Currency)
	}
}

// TestCreateLinkedAccount_ValidationError verifies handling of validation errors
func TestCreateLinkedAccount_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_error",
			"message": "Invalid account number format",
			"source": "account_number"
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
		"account_number": "invalid",
	}

	_, err := c.CreateLinkedAccount(context.Background(), req)
	if err == nil {
		t.Error("expected error for validation failure, got nil")
	}
}

// TestInitiateDeposit_Success verifies successful deposit initiation
func TestInitiateDeposit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/linked_accounts/la_123/initiate_deposit" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "dep_abc123",
			"amount": 1000.50,
			"currency": "USD",
			"status": "PENDING"
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

	deposit, err := c.InitiateDeposit(context.Background(), "la_123", 1000.50, "USD")
	if err != nil {
		t.Fatalf("InitiateDeposit() error: %v", err)
	}
	if deposit == nil {
		t.Fatal("deposit is nil")
	}
	if deposit.ID != "dep_abc123" {
		t.Errorf("id = %q, want 'dep_abc123'", deposit.ID)
	}
	if deposit.Amount != jn("1000.50") {
		t.Errorf("amount = %s, want 1000.50", deposit.Amount)
	}
	if deposit.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", deposit.Currency)
	}
	if deposit.Status != "PENDING" {
		t.Errorf("status = %q, want 'PENDING'", deposit.Status)
	}
}

// TestInitiateDeposit_DifferentCurrencies verifies deposit with various currencies
func TestInitiateDeposit_DifferentCurrencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "dep_test",
			"amount": 500.00,
			"currency": "EUR",
			"status": "PENDING"
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

	tests := []struct {
		name     string
		currency string
		amount   float64
	}{
		{
			name:     "EUR deposit",
			currency: "EUR",
			amount:   500.25,
		},
		{
			name:     "GBP deposit",
			currency: "GBP",
			amount:   250.75,
		},
		{
			name:     "JPY deposit",
			currency: "JPY",
			amount:   100000.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.InitiateDeposit(context.Background(), "la_123", tt.amount, tt.currency)
			if err != nil {
				t.Fatalf("InitiateDeposit() error: %v", err)
			}
		})
	}
}

// TestInitiateDeposit_InvalidAccountID verifies validation of account ID
func TestInitiateDeposit_InvalidAccountID(t *testing.T) {
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
	_, err := c.InitiateDeposit(context.Background(), "", 100.00, "USD")
	if err == nil {
		t.Error("expected error for empty account ID, got nil")
	}

	// Test ID with invalid characters
	_, err = c.InitiateDeposit(context.Background(), "invalid/id", 100.00, "USD")
	if err == nil {
		t.Error("expected error for invalid account ID, got nil")
	}

	// Test path traversal attempt
	_, err = c.InitiateDeposit(context.Background(), "../../../etc", 100.00, "USD")
	if err == nil {
		t.Error("expected error for path traversal attempt, got nil")
	}
}

// TestInitiateDeposit_ValidationErrors verifies handling of various validation errors
func TestInitiateDeposit_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
	}{
		{
			name:       "insufficient funds",
			statusCode: http.StatusBadRequest,
			response: `{
				"code": "insufficient_funds",
				"message": "Linked account has insufficient funds"
			}`,
		},
		{
			name:       "invalid amount",
			statusCode: http.StatusBadRequest,
			response: `{
				"code": "validation_error",
				"message": "Amount must be greater than zero",
				"source": "amount"
			}`,
		},
		{
			name:       "unsupported currency",
			statusCode: http.StatusBadRequest,
			response: `{
				"code": "validation_error",
				"message": "Currency not supported for this account",
				"source": "currency"
			}`,
		},
		{
			name:       "account not active",
			statusCode: http.StatusBadRequest,
			response: `{
				"code": "invalid_state",
				"message": "Linked account must be ACTIVE to initiate deposits"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
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

			_, err := c.InitiateDeposit(context.Background(), "la_123", 100.00, "USD")
			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
			}
		})
	}
}

// TestInitiateDeposit_ZeroAmount verifies handling of zero amount
func TestInitiateDeposit_ZeroAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_error",
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

	_, err := c.InitiateDeposit(context.Background(), "la_123", 0, "USD")
	if err == nil {
		t.Error("expected error for zero amount, got nil")
	}
}

// TestInitiateDeposit_NegativeAmount verifies handling of negative amount
func TestInitiateDeposit_NegativeAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "validation_error",
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

	_, err := c.InitiateDeposit(context.Background(), "la_123", -100.00, "USD")
	if err == nil {
		t.Error("expected error for negative amount, got nil")
	}
}
