package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Webhook represents a webhook subscription
type Webhook struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

type WebhooksResponse struct {
	Items   []Webhook `json:"items"`
	HasMore bool      `json:"has_more"`
}

// ListWebhooks lists all webhook subscriptions
func (c *Client) ListWebhooks(ctx context.Context, pageNum, pageSize int) (*WebhooksResponse, error) {
	params := url.Values{}
	if pageNum > 0 {
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/webhooks"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var result WebhooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWebhook retrieves a webhook by ID
func (c *Client) GetWebhook(ctx context.Context, webhookID string) (*Webhook, error) {
	if err := ValidateResourceID(webhookID, "webhook"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/webhooks/"+url.PathEscape(webhookID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var wh Webhook
	if err := json.NewDecoder(resp.Body).Decode(&wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// CreateWebhook creates a new webhook subscription
func (c *Client) CreateWebhook(ctx context.Context, webhookURL string, events []string) (*Webhook, error) {
	req := map[string]interface{}{
		"url":    webhookURL,
		"events": events,
	}

	resp, err := c.Post(ctx, "/api/v1/webhooks/create", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var wh Webhook
	if err := json.NewDecoder(resp.Body).Decode(&wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// DeleteWebhook deletes a webhook subscription
func (c *Client) DeleteWebhook(ctx context.Context, webhookID string) error {
	if err := ValidateResourceID(webhookID, "webhook"); err != nil {
		return err
	}

	resp, err := c.Post(ctx, "/api/v1/webhooks/"+url.PathEscape(webhookID)+"/delete", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return ParseAPIError(body)
	}
	return nil
}
