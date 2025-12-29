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

// AuthorizationListParams defines filters for listing authorizations.
type AuthorizationListParams struct {
	Status               string
	CardID               string
	BillingCurrency      string
	DigitalWalletTokenID string
	LifecycleID          string
	RetrievalRef         string
	FromCreatedAt        string
	ToCreatedAt          string
	PageNum              int
	PageSize             int
}

// ListAuthorizations lists issuing authorizations with optional filters.
func (c *Client) ListAuthorizations(ctx context.Context, params AuthorizationListParams) (*AuthorizationsResponse, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.CardID != "" {
		query.Set("card_id", params.CardID)
	}
	if params.BillingCurrency != "" {
		query.Set("billing_currency", params.BillingCurrency)
	}
	if params.DigitalWalletTokenID != "" {
		query.Set("digital_wallet_token_id", params.DigitalWalletTokenID)
	}
	if params.LifecycleID != "" {
		query.Set("lifecycle_id", params.LifecycleID)
	}
	if params.RetrievalRef != "" {
		query.Set("retrieval_ref", params.RetrievalRef)
	}
	if params.FromCreatedAt != "" {
		query.Set("from_created_at", params.FromCreatedAt)
	}
	if params.ToCreatedAt != "" {
		query.Set("to_created_at", params.ToCreatedAt)
	}
	if params.PageSize > 0 {
		if params.PageNum < 1 {
			params.PageNum = 1
		}
		query.Set("page_num", fmt.Sprintf("%d", params.PageNum))
		query.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}

	path := Endpoints.AuthorizationsList.Path
	if len(query) > 0 {
		path += "?" + query.Encode()
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
