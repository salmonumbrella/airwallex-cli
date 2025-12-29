package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Payer represents a payout payer.
type Payer struct {
	ID         string `json:"id"`
	PayerID    string `json:"payer_id"`
	EntityType string `json:"entity_type"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

type PayersResponse struct {
	Items   []Payer `json:"items"`
	HasMore bool    `json:"has_more"`
}

// PayerListParams defines filters for listing payers.
type PayerListParams struct {
	EntityType string
	Name       string
	NickName   string
	FromDate   string
	ToDate     string
	PageNum    int
	PageSize   int
}

// ListPayers lists payout payers.
func (c *Client) ListPayers(ctx context.Context, params PayerListParams) (*PayersResponse, error) {
	query := url.Values{}
	if params.EntityType != "" {
		query.Set("entity_type", params.EntityType)
	}
	if params.Name != "" {
		query.Set("name", params.Name)
	}
	if params.NickName != "" {
		query.Set("nick_name", params.NickName)
	}
	if params.FromDate != "" {
		query.Set("from_date", params.FromDate)
	}
	if params.ToDate != "" {
		query.Set("to_date", params.ToDate)
	}
	if params.PageSize > 0 {
		if params.PageNum < 1 {
			params.PageNum = 1
		}
		query.Set("page_num", fmt.Sprintf("%d", params.PageNum))
		query.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}

	path := Endpoints.PayersList.Path
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

	var result PayersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPayer retrieves a payer by ID.
func (c *Client) GetPayer(ctx context.Context, payerID string) (*Payer, error) {
	if err := ValidateResourceID(payerID, "payer"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/payers/"+url.PathEscape(payerID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var payer Payer
	if err := json.NewDecoder(resp.Body).Decode(&payer); err != nil {
		return nil, err
	}
	return &payer, nil
}

// CreatePayer creates a new payer.
func (c *Client) CreatePayer(ctx context.Context, req map[string]interface{}) (*Payer, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.PayersCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var payer Payer
	if err := json.NewDecoder(resp.Body).Decode(&payer); err != nil {
		return nil, err
	}
	return &payer, nil
}

// UpdatePayer updates a payer.
func (c *Client) UpdatePayer(ctx context.Context, payerID string, req map[string]interface{}) (*Payer, error) {
	if err := ValidateResourceID(payerID, "payer"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/payers/update/"+url.PathEscape(payerID), req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var payer Payer
	if err := json.NewDecoder(resp.Body).Decode(&payer); err != nil {
		return nil, err
	}
	return &payer, nil
}

// DeletePayer deletes a payer.
func (c *Client) DeletePayer(ctx context.Context, payerID string) error {
	if err := ValidateResourceID(payerID, "payer"); err != nil {
		return err
	}

	resp, err := c.Post(ctx, "/api/v1/payers/delete/"+url.PathEscape(payerID), nil)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return ParseAPIError(body)
	}
	return nil
}

// ValidatePayer validates payer details without creating.
func (c *Client) ValidatePayer(ctx context.Context, req map[string]interface{}) error {
	resp, err := c.Post(ctx, Endpoints.PayersValidate.Path, req)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return ParseAPIError(body)
	}
	return nil
}
