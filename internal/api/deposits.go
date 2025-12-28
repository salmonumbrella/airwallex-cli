package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Deposit represents an inbound deposit
type Deposit struct {
	ID              string  `json:"id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
	Source          string  `json:"source"`
	LinkedAccountID string  `json:"linked_account_id,omitempty"`
	GlobalAccountID string  `json:"global_account_id,omitempty"`
	Reference       string  `json:"reference,omitempty"`
	CreatedAt       string  `json:"created_at"`
	SettledAt       string  `json:"settled_at,omitempty"`
}

type DepositsResponse struct {
	Items   []Deposit `json:"items"`
	HasMore bool      `json:"has_more"`
}

// ListDeposits lists all deposits with optional filters
func (c *Client) ListDeposits(ctx context.Context, status, fromDate, toDate string, pageNum, pageSize int) (*DepositsResponse, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if fromDate != "" {
		params.Set("from_created_at", fromDate)
	}
	if toDate != "" {
		params.Set("to_created_at", toDate)
	}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/deposits"
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
		return nil, ParseAPIError(body)
	}

	var result DepositsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDeposit retrieves a deposit by ID
func (c *Client) GetDeposit(ctx context.Context, depositID string) (*Deposit, error) {
	if err := ValidateResourceID(depositID, "deposit"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/deposits/"+url.PathEscape(depositID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var d Deposit
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}
