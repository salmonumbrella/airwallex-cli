package api

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
)

// GlobalAccount represents an Airwallex global account
type GlobalAccount struct {
	AccountID     string `json:"id"`
	AccountName   string `json:"account_name"`
	Currency      string `json:"currency"`
	CountryCode   string `json:"country_code"`
	Status        string `json:"status"`
	AccountNumber string `json:"account_number,omitempty"`
	RoutingCode   string `json:"routing_code,omitempty"`
	IBAN          string `json:"iban,omitempty"`
	SwiftCode     string `json:"swift_code,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type GlobalAccountsResponse struct {
	Items   []GlobalAccount `json:"items"`
	HasMore bool            `json:"has_more"`
}

// ListGlobalAccounts lists all global accounts
func (c *Client) ListGlobalAccounts(ctx context.Context, pageNum, pageSize int) (*GlobalAccountsResponse, error) {
	params := url.Values{}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", strconv.Itoa(pageNum))
		params.Set("page_size", strconv.Itoa(pageSize))
	}

	path := "/api/v1/global_accounts"
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

	var result GlobalAccountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGlobalAccount retrieves a single global account by ID
func (c *Client) GetGlobalAccount(ctx context.Context, accountID string) (*GlobalAccount, error) {
	if err := ValidateResourceID(accountID, "account"); err != nil {
		return nil, err
	}
	path := "/api/v1/global_accounts/" + url.PathEscape(accountID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
	}

	var a GlobalAccount
	if err := json.NewDecoder(resp.Body).Decode(&a); err != nil {
		return nil, err
	}
	return &a, nil
}
