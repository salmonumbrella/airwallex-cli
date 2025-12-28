package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// LinkedAccount represents an external bank account
type LinkedAccount struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	AccountName   string `json:"account_name"`
	BankName      string `json:"bank_name,omitempty"`
	AccountNumber string `json:"account_number_last4,omitempty"`
	Currency      string `json:"currency"`
	CreatedAt     string `json:"created_at"`
}

type LinkedAccountsResponse struct {
	Items   []LinkedAccount `json:"items"`
	HasMore bool            `json:"has_more"`
}

// DepositInitiation represents the response from initiating a deposit
type DepositInitiation struct {
	ID       string  `json:"id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
}

// ListLinkedAccounts lists all linked accounts
func (c *Client) ListLinkedAccounts(ctx context.Context, pageNum, pageSize int) (*LinkedAccountsResponse, error) {
	params := url.Values{}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/linked_accounts"
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

	var result LinkedAccountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLinkedAccount retrieves a linked account by ID
func (c *Client) GetLinkedAccount(ctx context.Context, accountID string) (*LinkedAccount, error) {
	if err := ValidateResourceID(accountID, "linked account"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/linked_accounts/"+url.PathEscape(accountID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var la LinkedAccount
	if err := json.NewDecoder(resp.Body).Decode(&la); err != nil {
		return nil, err
	}
	return &la, nil
}

// CreateLinkedAccount creates a new linked account
func (c *Client) CreateLinkedAccount(ctx context.Context, req map[string]interface{}) (*LinkedAccount, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, "/api/v1/linked_accounts/create", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var la LinkedAccount
	if err := json.NewDecoder(resp.Body).Decode(&la); err != nil {
		return nil, err
	}
	return &la, nil
}

// InitiateDeposit initiates a deposit from a linked account
func (c *Client) InitiateDeposit(ctx context.Context, accountID string, amount float64, currency string) (*DepositInitiation, error) {
	if err := ValidateResourceID(accountID, "linked account"); err != nil {
		return nil, err
	}

	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	req := map[string]interface{}{
		"amount":   amount,
		"currency": currency,
	}

	resp, err := c.Post(ctx, "/api/v1/linked_accounts/"+url.PathEscape(accountID)+"/initiate_deposit", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var di DepositInitiation
	if err := json.NewDecoder(resp.Body).Decode(&di); err != nil {
		return nil, err
	}
	return &di, nil
}
