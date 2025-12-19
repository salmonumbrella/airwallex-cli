package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetBeneficiarySchema_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiary_api_schemas/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": [
				{
					"name": "account_number",
					"type": "string",
					"required": true,
					"description": "Bank account number",
					"min_length": 8,
					"max_length": 17
				},
				{
					"name": "account_routing_type1",
					"type": "string",
					"required": true,
					"description": "Routing code type",
					"enum": ["aba", "swift"]
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

	schema, err := c.GetBeneficiarySchema(context.Background(), "US", "COMPANY", "")
	if err != nil {
		t.Fatalf("GetBeneficiarySchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 2 {
		t.Errorf("fields count = %d, want 2", len(schema.Fields))
	}

	// Validate first field
	if schema.Fields[0].Name != "account_number" {
		t.Errorf("fields[0].name = %q, want 'account_number'", schema.Fields[0].Name)
	}
	if schema.Fields[0].Type != "string" {
		t.Errorf("fields[0].type = %q, want 'string'", schema.Fields[0].Type)
	}
	if !schema.Fields[0].Required {
		t.Error("fields[0].required = false, want true")
	}
	if schema.Fields[0].MinLength != 8 {
		t.Errorf("fields[0].min_length = %d, want 8", schema.Fields[0].MinLength)
	}
	if schema.Fields[0].MaxLength != 17 {
		t.Errorf("fields[0].max_length = %d, want 17", schema.Fields[0].MaxLength)
	}

	// Validate second field with enum
	if schema.Fields[1].Name != "account_routing_type1" {
		t.Errorf("fields[1].name = %q, want 'account_routing_type1'", schema.Fields[1].Name)
	}
	if len(schema.Fields[1].Enum) != 2 {
		t.Errorf("fields[1].enum count = %d, want 2", len(schema.Fields[1].Enum))
	}
}

func TestGetBeneficiarySchema_WithPaymentMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/beneficiary_api_schemas/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": [
				{
					"name": "swift_code",
					"type": "string",
					"required": true,
					"description": "SWIFT/BIC code",
					"pattern": "^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$"
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

	schema, err := c.GetBeneficiarySchema(context.Background(), "GB", "COMPANY", "SWIFT")
	if err != nil {
		t.Fatalf("GetBeneficiarySchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 1 {
		t.Errorf("fields count = %d, want 1", len(schema.Fields))
	}
	if schema.Fields[0].Pattern == "" {
		t.Error("fields[0].pattern is empty, expected pattern")
	}
}

func TestGetBeneficiarySchema_EmptyFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": []
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

	schema, err := c.GetBeneficiarySchema(context.Background(), "XX", "INDIVIDUAL", "")
	if err != nil {
		t.Fatalf("GetBeneficiarySchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 0 {
		t.Errorf("fields count = %d, want 0", len(schema.Fields))
	}
}

func TestGetBeneficiarySchema_InvalidCountry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"code": "invalid_country_code",
			"message": "Invalid bank country code"
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

	_, err := c.GetBeneficiarySchema(context.Background(), "INVALID", "COMPANY", "")
	if err == nil {
		t.Error("expected error for invalid country code, got nil")
	}
}

func TestGetBeneficiarySchema_NetworkError(t *testing.T) {
	c := &Client{
		baseURL:        "http://nonexistent.invalid",
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetBeneficiarySchema(context.Background(), "US", "COMPANY", "")
	if err == nil {
		t.Error("expected network error, got nil")
	}
}

func TestGetBeneficiarySchema_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json`))
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

	_, err := c.GetBeneficiarySchema(context.Background(), "US", "COMPANY", "")
	if err == nil {
		t.Error("expected JSON decode error, got nil")
	}
}

func TestGetTransferSchema_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transfer_api_schemas/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": [
				{
					"name": "amount",
					"type": "number",
					"required": true,
					"description": "Transfer amount"
				},
				{
					"name": "purpose_code",
					"type": "string",
					"required": false,
					"description": "Purpose of payment",
					"enum": ["SALARY", "GOODS", "SERVICES"]
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

	schema, err := c.GetTransferSchema(context.Background(), "USD", "EUR", "")
	if err != nil {
		t.Fatalf("GetTransferSchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 2 {
		t.Errorf("fields count = %d, want 2", len(schema.Fields))
	}

	// Validate amount field
	if schema.Fields[0].Name != "amount" {
		t.Errorf("fields[0].name = %q, want 'amount'", schema.Fields[0].Name)
	}
	if schema.Fields[0].Type != "number" {
		t.Errorf("fields[0].type = %q, want 'number'", schema.Fields[0].Type)
	}
	if !schema.Fields[0].Required {
		t.Error("fields[0].required = false, want true")
	}

	// Validate purpose_code field with enum
	if schema.Fields[1].Name != "purpose_code" {
		t.Errorf("fields[1].name = %q, want 'purpose_code'", schema.Fields[1].Name)
	}
	if schema.Fields[1].Required {
		t.Error("fields[1].required = true, want false")
	}
	if len(schema.Fields[1].Enum) != 3 {
		t.Errorf("fields[1].enum count = %d, want 3", len(schema.Fields[1].Enum))
	}
}

func TestGetTransferSchema_WithPaymentMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/transfer_api_schemas/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": [
				{
					"name": "reference",
					"type": "string",
					"required": true,
					"description": "Payment reference",
					"max_length": 35
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

	schema, err := c.GetTransferSchema(context.Background(), "GBP", "EUR", "SWIFT")
	if err != nil {
		t.Fatalf("GetTransferSchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 1 {
		t.Errorf("fields count = %d, want 1", len(schema.Fields))
	}
	if schema.Fields[0].MaxLength != 35 {
		t.Errorf("fields[0].max_length = %d, want 35", schema.Fields[0].MaxLength)
	}
}

func TestGetTransferSchema_EmptyFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": []
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

	schema, err := c.GetTransferSchema(context.Background(), "USD", "USD", "")
	if err != nil {
		t.Fatalf("GetTransferSchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 0 {
		t.Errorf("fields count = %d, want 0", len(schema.Fields))
	}
}

func TestGetTransferSchema_InvalidCurrency(t *testing.T) {
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

	_, err := c.GetTransferSchema(context.Background(), "INVALID", "EUR", "")
	if err == nil {
		t.Error("expected error for invalid currency code, got nil")
	}
}

func TestGetTransferSchema_NetworkError(t *testing.T) {
	c := &Client{
		baseURL:        "http://nonexistent.invalid",
		clientID:       "test-id",
		apiKey:         "test-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "test-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetTransferSchema(context.Background(), "USD", "EUR", "")
	if err == nil {
		t.Error("expected network error, got nil")
	}
}

func TestGetTransferSchema_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{incomplete json`))
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

	_, err := c.GetTransferSchema(context.Background(), "USD", "EUR", "")
	if err == nil {
		t.Error("expected JSON decode error, got nil")
	}
}

func TestGetTransferSchema_SameCurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": [
				{
					"name": "domestic_reference",
					"type": "string",
					"required": true,
					"description": "Domestic payment reference"
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

	schema, err := c.GetTransferSchema(context.Background(), "USD", "USD", "")
	if err != nil {
		t.Fatalf("GetTransferSchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 1 {
		t.Errorf("fields count = %d, want 1", len(schema.Fields))
	}
}

func TestGetBeneficiarySchema_AllFieldTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"fields": [
				{
					"name": "account_number",
					"type": "string",
					"required": true,
					"description": "Account number",
					"min_length": 5,
					"max_length": 20,
					"pattern": "^[0-9]+$"
				},
				{
					"name": "routing_type",
					"type": "string",
					"required": true,
					"enum": ["aba", "swift", "iban"]
				},
				{
					"name": "amount_limit",
					"type": "number",
					"required": false
				},
				{
					"name": "is_active",
					"type": "boolean",
					"required": false
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

	schema, err := c.GetBeneficiarySchema(context.Background(), "US", "COMPANY", "")
	if err != nil {
		t.Fatalf("GetBeneficiarySchema() error: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Fields) != 4 {
		t.Errorf("fields count = %d, want 4", len(schema.Fields))
	}

	// Validate all field attributes are preserved
	stringField := schema.Fields[0]
	if stringField.MinLength != 5 || stringField.MaxLength != 20 {
		t.Errorf("string field length constraints not preserved")
	}
	if stringField.Pattern != "^[0-9]+$" {
		t.Errorf("pattern = %q, want '^[0-9]+$'", stringField.Pattern)
	}

	enumField := schema.Fields[1]
	if len(enumField.Enum) != 3 {
		t.Errorf("enum count = %d, want 3", len(enumField.Enum))
	}

	numberField := schema.Fields[2]
	if numberField.Type != "number" {
		t.Errorf("number field type = %q, want 'number'", numberField.Type)
	}

	boolField := schema.Fields[3]
	if boolField.Type != "boolean" {
		t.Errorf("boolean field type = %q, want 'boolean'", boolField.Type)
	}
}

func TestGetTransferSchema_Unauthorized(t *testing.T) {
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
		clientID:       "invalid-id",
		apiKey:         "invalid-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "invalid-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetTransferSchema(context.Background(), "USD", "EUR", "")
	if err == nil {
		t.Error("expected unauthorized error, got nil")
	}
}

func TestGetBeneficiarySchema_Unauthorized(t *testing.T) {
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
		clientID:       "invalid-id",
		apiKey:         "invalid-key",
		httpClient:     http.DefaultClient,
		circuitBreaker: &circuitBreaker{},

		token: &TokenCache{
			Token:     "invalid-token",
			ExpiresAt: time.Now().Add(10 * time.Minute),
		},
	}

	_, err := c.GetBeneficiarySchema(context.Background(), "US", "COMPANY", "")
	if err == nil {
		t.Error("expected unauthorized error, got nil")
	}
}
