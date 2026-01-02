package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListWebhooks_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/webhooks" {
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
					"id": "wh_123",
					"url": "https://example.com/webhook",
					"events": ["transfer.completed", "transfer.failed"],
					"status": "ACTIVE",
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

	result, err := c.ListWebhooks(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("ListWebhooks() error: %v", err)
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
	if result.Items[0].ID != "wh_123" {
		t.Errorf("id = %q, want 'wh_123'", result.Items[0].ID)
	}
	if result.Items[0].URL != "https://example.com/webhook" {
		t.Errorf("url = %q, want 'https://example.com/webhook'", result.Items[0].URL)
	}
	if len(result.Items[0].Events) != 2 {
		t.Errorf("events count = %d, want 2", len(result.Items[0].Events))
	}
}

func TestListWebhooks_WithoutPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/webhooks" {
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

	result, err := c.ListWebhooks(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListWebhooks() error: %v", err)
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

func TestListWebhooks_EmptyResults(t *testing.T) {
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

	result, err := c.ListWebhooks(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListWebhooks() error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Items) != 0 {
		t.Errorf("items count = %d, want 0", len(result.Items))
	}
}

func TestGetWebhook_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/webhooks/wh_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "wh_123",
			"url": "https://example.com/webhook",
			"events": ["transfer.completed", "deposit.settled"],
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

	webhook, err := c.GetWebhook(context.Background(), "wh_123")
	if err != nil {
		t.Fatalf("GetWebhook() error: %v", err)
	}
	if webhook == nil {
		t.Fatal("webhook is nil")
	}
	if webhook.ID != "wh_123" {
		t.Errorf("id = %q, want 'wh_123'", webhook.ID)
	}
	if webhook.URL != "https://example.com/webhook" {
		t.Errorf("url = %q, want 'https://example.com/webhook'", webhook.URL)
	}
	if webhook.Status != "ACTIVE" {
		t.Errorf("status = %q, want 'ACTIVE'", webhook.Status)
	}
}

func TestGetWebhook_InvalidID(t *testing.T) {
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

	_, err := c.GetWebhook(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty webhook ID, got nil")
	}

	_, err = c.GetWebhook(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid webhook ID, got nil")
	}
}

func TestCreateWebhook_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/webhooks/create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s, want POST", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "wh_new",
			"url": "https://example.com/webhook",
			"events": ["transfer.completed"],
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

	webhook, err := c.CreateWebhook(context.Background(), "https://example.com/webhook", []string{"transfer.completed"})
	if err != nil {
		t.Fatalf("CreateWebhook() error: %v", err)
	}
	if webhook == nil {
		t.Fatal("webhook is nil")
	}
	if webhook.ID != "wh_new" {
		t.Errorf("id = %q, want 'wh_new'", webhook.ID)
	}
	if webhook.URL != "https://example.com/webhook" {
		t.Errorf("url = %q, want 'https://example.com/webhook'", webhook.URL)
	}
	if len(webhook.Events) != 1 {
		t.Errorf("events count = %d, want 1", len(webhook.Events))
	}
	if webhook.Events[0] != "transfer.completed" {
		t.Errorf("events[0] = %q, want 'transfer.completed'", webhook.Events[0])
	}
}

func TestCreateWebhook_MultipleEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"id": "wh_multi",
			"url": "https://example.com/webhook",
			"events": ["transfer.completed", "transfer.failed", "deposit.settled"],
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

	events := []string{"transfer.completed", "transfer.failed", "deposit.settled"}
	webhook, err := c.CreateWebhook(context.Background(), "https://example.com/webhook", events)
	if err != nil {
		t.Fatalf("CreateWebhook() error: %v", err)
	}
	if webhook == nil {
		t.Fatal("webhook is nil")
	}
	if len(webhook.Events) != 3 {
		t.Errorf("events count = %d, want 3", len(webhook.Events))
	}
}

func TestDeleteWebhook_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/webhooks/wh_123/delete" {
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

	err := c.DeleteWebhook(context.Background(), "wh_123")
	if err != nil {
		t.Fatalf("DeleteWebhook() error: %v", err)
	}
}

func TestDeleteWebhook_InvalidID(t *testing.T) {
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

	err := c.DeleteWebhook(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty webhook ID, got nil")
	}

	err = c.DeleteWebhook(context.Background(), "invalid/id")
	if err == nil {
		t.Error("expected error for invalid webhook ID, got nil")
	}
}

func TestDeleteWebhook_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{
			"code": "resource_not_found",
			"message": "Webhook not found"
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

	err := c.DeleteWebhook(context.Background(), "wh_nonexistent")
	if err == nil {
		t.Error("expected error for not found webhook, got nil")
	}
}
