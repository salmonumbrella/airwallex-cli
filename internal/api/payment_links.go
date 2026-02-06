package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// PaymentLink represents a payment link
type PaymentLink struct {
	ID          string      `json:"id"`
	URL         string      `json:"url"`
	Amount      json.Number `json:"amount"`
	Currency    string      `json:"currency"`
	Description string      `json:"description,omitempty"`
	Status      string      `json:"status"`
	ExpiresAt   string      `json:"expires_at,omitempty"`
	CreatedAt   string      `json:"created_at"`
}

type PaymentLinksResponse struct {
	Items   []PaymentLink `json:"items"`
	HasMore bool          `json:"has_more"`
}

// ListPaymentLinks lists all payment links
func (c *Client) ListPaymentLinks(ctx context.Context, pageNum, pageSize int) (*PaymentLinksResponse, error) {
	params := url.Values{}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/pa/payment_links"
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

	var result PaymentLinksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPaymentLink retrieves a payment link by ID
func (c *Client) GetPaymentLink(ctx context.Context, linkID string) (*PaymentLink, error) {
	if err := ValidateResourceID(linkID, "payment link"); err != nil {
		return nil, err
	}

	path := "/api/v1/pa/payment_links/" + url.PathEscape(linkID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
	}

	var pl PaymentLink
	if err := json.NewDecoder(resp.Body).Decode(&pl); err != nil {
		return nil, err
	}
	return &pl, nil
}

// CreatePaymentLink creates a new payment link
func (c *Client) CreatePaymentLink(ctx context.Context, req map[string]interface{}) (*PaymentLink, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	path := "/api/v1/pa/payment_links/create"
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}

	var pl PaymentLink
	if err := json.NewDecoder(resp.Body).Decode(&pl); err != nil {
		return nil, err
	}
	return &pl, nil
}
