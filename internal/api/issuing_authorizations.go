package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Authorization represents a card authorization.
type Authorization struct {
	AuthorizationID string  `json:"authorization_id"`
	ID              string  `json:"id"`
	TransactionID   string  `json:"transaction_id"`
	CardID          string  `json:"card_id"`
	CardholderID    string  `json:"cardholder_id"`
	Status          string  `json:"status"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	CreatedAt       string  `json:"created_at"`
	Merchant        struct {
		Name string `json:"name"`
	} `json:"merchant"`
}

type AuthorizationsResponse struct {
	Items   []Authorization `json:"items"`
	HasMore bool            `json:"has_more"`
}

// ListAuthorizations lists issuing authorizations with optional filters.
func (c *Client) ListAuthorizations(ctx context.Context, status, cardID, cardholderID, fromCreatedAt, toCreatedAt string, pageNum, pageSize int) (*AuthorizationsResponse, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if cardID != "" {
		params.Set("card_id", cardID)
	}
	if cardholderID != "" {
		params.Set("cardholder_id", cardholderID)
	}
	if fromCreatedAt != "" {
		params.Set("from_created_at", fromCreatedAt)
	}
	if toCreatedAt != "" {
		params.Set("to_created_at", toCreatedAt)
	}
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := Endpoints.AuthorizationsList.Path
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

	var result AuthorizationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAuthorization retrieves an authorization by transaction ID.
func (c *Client) GetAuthorization(ctx context.Context, transactionID string) (*Authorization, error) {
	if err := ValidateResourceID(transactionID, "transaction"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/issuing/authorizations/"+url.PathEscape(transactionID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var auth Authorization
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return nil, err
	}
	return &auth, nil
}
