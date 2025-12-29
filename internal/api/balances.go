package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

type Balance struct {
	Currency        string  `json:"currency"`
	AvailableAmount float64 `json:"available_amount"`
	PendingAmount   float64 `json:"pending_amount"`
	ReservedAmount  float64 `json:"reserved_amount"`
	TotalAmount     float64 `json:"total_amount"`
}

type BalancesResponse struct {
	Balances []Balance
}

type BalanceHistoryItem struct {
	ID              string  `json:"id"`
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	Balance         float64 `json:"balance"`
	TransactionType string  `json:"transaction_type"`
	PostedAt        string  `json:"posted_at"`
	Description     string  `json:"description,omitempty"`
}

type BalanceHistoryResponse struct {
	Items   []BalanceHistoryItem `json:"items"`
	HasMore bool                 `json:"has_more"`
}

func (c *Client) GetBalances(ctx context.Context) (*BalancesResponse, error) {
	resp, err := c.Get(ctx, Endpoints.BalancesCurrent.Path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != Endpoints.BalancesCurrent.ExpectedStatus {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	// API returns an array directly, not wrapped in an object
	var balances []Balance
	if err := json.NewDecoder(resp.Body).Decode(&balances); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &BalancesResponse{Balances: balances}, nil
}

func (c *Client) GetBalanceHistory(ctx context.Context, currency string, from, to string, pageNum, pageSize int) (*BalanceHistoryResponse, error) {
	params := url.Values{}
	if currency != "" {
		params.Set("currency", currency)
	}
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := Endpoints.BalancesHistory.Path
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != Endpoints.BalancesHistory.ExpectedStatus {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var result BalanceHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
