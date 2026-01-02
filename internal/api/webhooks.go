package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
)

const (
	// MaxPageSize is the maximum number of items per page
	MaxPageSize = 1000

	// MaxPageNum is the maximum page number
	MaxPageNum = 100000
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

// validateWebhookURL validates webhook URLs for security and correctness
func validateWebhookURL(webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	// Parse the URL
	u, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Require HTTPS (or HTTP for localhost in development)
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("webhook URL must use HTTPS or HTTP scheme (got %q)", u.Scheme)
	}

	// Only allow HTTP for localhost/127.0.0.1 (development)
	if u.Scheme == "http" {
		host := u.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			return fmt.Errorf("HTTP URLs only allowed for localhost (got %q)", host)
		}
	}

	// SSRF prevention: block localhost/127.0.0.1/private IPs for HTTPS
	if u.Scheme == "https" {
		host := u.Hostname()

		// Block localhost
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return fmt.Errorf("webhook URL cannot use localhost with HTTPS")
		}

		// Block private IP ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
		if ip := net.ParseIP(host); ip != nil {
			if ip.IsLoopback() || ip.IsPrivate() {
				return fmt.Errorf("webhook URL cannot use private IP address: %s", host)
			}
		}

		// Block common internal hostnames
		lowerHost := strings.ToLower(host)
		if strings.HasSuffix(lowerHost, ".local") || strings.HasSuffix(lowerHost, ".internal") {
			return fmt.Errorf("webhook URL cannot use internal domain: %s", host)
		}
	}

	return nil
}

// ListWebhooks lists all webhook subscriptions
func (c *Client) ListWebhooks(ctx context.Context, pageNum, pageSize int) (*WebhooksResponse, error) {
	// Validate pagination bounds
	if pageNum > MaxPageNum {
		return nil, fmt.Errorf("page_num exceeds maximum allowed value of %d", MaxPageNum)
	}
	if pageSize > MaxPageSize {
		return nil, fmt.Errorf("page_size exceeds maximum allowed value of %d", MaxPageSize)
	}

	params := url.Values{}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
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
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
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

	path := "/api/v1/webhooks/" + url.PathEscape(webhookID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
	}

	var wh Webhook
	if err := json.NewDecoder(resp.Body).Decode(&wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// CreateWebhook creates a new webhook subscription
func (c *Client) CreateWebhook(ctx context.Context, webhookURL string, events []string) (*Webhook, error) {
	// Validate webhook URL for security
	if err := validateWebhookURL(webhookURL); err != nil {
		return nil, err
	}

	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	req := map[string]interface{}{
		"url":    webhookURL,
		"events": events,
	}

	path := "/api/v1/webhooks/create"
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
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

	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	path := "/api/v1/webhooks/" + url.PathEscape(webhookID) + "/delete"
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}
	return nil
}
