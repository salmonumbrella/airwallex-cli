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
					"key": "account_number",
					"path": "beneficiary.account_number",
					"required": true,
					"description": "Bank account number",
					"rule": {"type": "string", "minLength": 8, "maxLength": 17}
				},
				{
					"key": "account_routing_type1",
					"path": "beneficiary.account_routing_type1",
					"required": true,
					"description": "Routing code type",
					"rule": {"type": "string", "enum": ["aba", "swift"]}
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
	if schema.Fields[0].Name() != "account_number" {
		t.Errorf("fields[0].Name() = %q, want 'account_number'", schema.Fields[0].Name())
	}
	if schema.Fields[0].Type() != "string" {
		t.Errorf("fields[0].Type() = %q, want 'string'", schema.Fields[0].Type())
	}
	if !schema.Fields[0].Required {
		t.Error("fields[0].required = false, want true")
	}
	if schema.Fields[0].MinLength() != 8 {
		t.Errorf("fields[0].MinLength() = %d, want 8", schema.Fields[0].MinLength())
	}
	if schema.Fields[0].MaxLength() != 17 {
		t.Errorf("fields[0].MaxLength() = %d, want 17", schema.Fields[0].MaxLength())
	}

	// Validate second field with enum
	if schema.Fields[1].Name() != "account_routing_type1" {
		t.Errorf("fields[1].Name() = %q, want 'account_routing_type1'", schema.Fields[1].Name())
	}
	if len(schema.Fields[1].Enum()) != 2 {
		t.Errorf("fields[1].enum count = %d, want 2", len(schema.Fields[1].Enum()))
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
					"key": "swift_code",
					"path": "beneficiary.swift_code",
					"required": true,
					"description": "SWIFT/BIC code",
					"rule": {"type": "string", "pattern": "^[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?$"}
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
	if schema.Fields[0].Pattern() == "" {
		t.Error("fields[0].Pattern() is empty, expected pattern")
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
					"key": "amount",
					"path": "transfer.amount",
					"required": true,
					"description": "Transfer amount",
					"rule": {"type": "number"}
				},
				{
					"key": "purpose_code",
					"path": "transfer.purpose_code",
					"required": false,
					"description": "Purpose of payment",
					"rule": {"type": "string", "enum": ["SALARY", "GOODS", "SERVICES"]}
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
	if schema.Fields[0].Name() != "amount" {
		t.Errorf("fields[0].Name() = %q, want 'amount'", schema.Fields[0].Name())
	}
	if schema.Fields[0].Type() != "number" {
		t.Errorf("fields[0].Type() = %q, want 'number'", schema.Fields[0].Type())
	}
	if !schema.Fields[0].Required {
		t.Error("fields[0].required = false, want true")
	}

	// Validate purpose_code field with enum
	if schema.Fields[1].Name() != "purpose_code" {
		t.Errorf("fields[1].Name() = %q, want 'purpose_code'", schema.Fields[1].Name())
	}
	if schema.Fields[1].Required {
		t.Error("fields[1].required = true, want false")
	}
	if len(schema.Fields[1].Enum()) != 3 {
		t.Errorf("fields[1].enum count = %d, want 3", len(schema.Fields[1].Enum()))
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
					"key": "reference",
					"path": "transfer.reference",
					"required": true,
					"description": "Payment reference",
					"rule": {"type": "string", "maxLength": 35}
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
	if schema.Fields[0].MaxLength() != 35 {
		t.Errorf("fields[0].MaxLength() = %d, want 35", schema.Fields[0].MaxLength())
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
					"key": "domestic_reference",
					"path": "transfer.domestic_reference",
					"required": true,
					"description": "Domestic payment reference",
					"rule": {"type": "string"}
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
					"key": "account_number",
					"path": "beneficiary.account_number",
					"required": true,
					"description": "Account number",
					"rule": {"type": "string", "minLength": 5, "maxLength": 20, "pattern": "^[0-9]+$"}
				},
				{
					"key": "routing_type",
					"path": "beneficiary.routing_type",
					"required": true,
					"rule": {"type": "string", "enum": ["aba", "swift", "iban"]}
				},
				{
					"key": "amount_limit",
					"path": "beneficiary.amount_limit",
					"required": false,
					"rule": {"type": "number"}
				},
				{
					"key": "is_active",
					"path": "beneficiary.is_active",
					"required": false,
					"rule": {"type": "boolean"}
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
	if stringField.MinLength() != 5 || stringField.MaxLength() != 20 {
		t.Errorf("string field length constraints not preserved")
	}
	if stringField.Pattern() != "^[0-9]+$" {
		t.Errorf("pattern = %q, want '^[0-9]+$'", stringField.Pattern())
	}

	enumField := schema.Fields[1]
	if len(enumField.Enum()) != 3 {
		t.Errorf("enum count = %d, want 3", len(enumField.Enum()))
	}

	numberField := schema.Fields[2]
	if numberField.Type() != "number" {
		t.Errorf("number field type = %q, want 'number'", numberField.Type())
	}

	boolField := schema.Fields[3]
	if boolField.Type() != "boolean" {
		t.Errorf("boolean field type = %q, want 'boolean'", boolField.Type())
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
