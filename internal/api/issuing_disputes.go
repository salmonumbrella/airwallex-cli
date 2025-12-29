package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// TransactionDispute represents a dispute for an issuing transaction.
type TransactionDispute struct {
	DisputeID     string  `json:"dispute_id"`
	ID            string  `json:"id"`
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Reason        string  `json:"reason"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	CreatedAt     string  `json:"created_at"`
}

type TransactionDisputesResponse struct {
	Items   []TransactionDispute `json:"items"`
	HasMore bool                 `json:"has_more"`
}

// TransactionDisputeListParams defines filters for disputes list.
type TransactionDisputeListParams struct {
	Status         string
	DetailedStatus string
	Reason         string
	Reference      string
	TransactionID  string
	UpdatedBy      string
	FromCreatedAt  string
	ToCreatedAt    string
	FromUpdatedAt  string
	ToUpdatedAt    string
	Page           string
	PageSize       int
}

// ListTransactionDisputes lists issuing transaction disputes.
func (c *Client) ListTransactionDisputes(ctx context.Context, params TransactionDisputeListParams) (*TransactionDisputesResponse, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.DetailedStatus != "" {
		query.Set("detailed_status", params.DetailedStatus)
	}
	if params.Reason != "" {
		query.Set("reason", params.Reason)
	}
	if params.Reference != "" {
		query.Set("reference", params.Reference)
	}
	if params.TransactionID != "" {
		query.Set("transaction_id", params.TransactionID)
	}
	if params.UpdatedBy != "" {
		query.Set("updated_by", params.UpdatedBy)
	}
	if params.FromCreatedAt != "" {
		query.Set("from_created_at", params.FromCreatedAt)
	}
	if params.ToCreatedAt != "" {
		query.Set("to_created_at", params.ToCreatedAt)
	}
	if params.FromUpdatedAt != "" {
		query.Set("from_updated_at", params.FromUpdatedAt)
	}
	if params.ToUpdatedAt != "" {
		query.Set("to_updated_at", params.ToUpdatedAt)
	}
	if params.Page != "" {
		query.Set("page", params.Page)
	}
	if params.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}

	path := Endpoints.TransactionDisputesList.Path
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

	var result TransactionDisputesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTransactionDispute retrieves a dispute by ID.
func (c *Client) GetTransactionDispute(ctx context.Context, disputeID string) (*TransactionDispute, error) {
	if err := ValidateResourceID(disputeID, "dispute"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/issuing/transaction_disputes/"+url.PathEscape(disputeID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var dispute TransactionDispute
	if err := json.NewDecoder(resp.Body).Decode(&dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}

// CreateTransactionDispute creates a new dispute.
func (c *Client) CreateTransactionDispute(ctx context.Context, req map[string]interface{}) (*TransactionDispute, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.TransactionDisputesCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var dispute TransactionDispute
	if err := json.NewDecoder(resp.Body).Decode(&dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}

// UpdateTransactionDispute updates a dispute.
func (c *Client) UpdateTransactionDispute(ctx context.Context, disputeID string, req map[string]interface{}) (*TransactionDispute, error) {
	if err := ValidateResourceID(disputeID, "dispute"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/issuing/transaction_disputes/"+url.PathEscape(disputeID)+"/update", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var dispute TransactionDispute
	if err := json.NewDecoder(resp.Body).Decode(&dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}

// SubmitTransactionDispute submits a dispute.
func (c *Client) SubmitTransactionDispute(ctx context.Context, disputeID string) (*TransactionDispute, error) {
	if err := ValidateResourceID(disputeID, "dispute"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/issuing/transaction_disputes/"+url.PathEscape(disputeID)+"/submit", nil)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var dispute TransactionDispute
	if err := json.NewDecoder(resp.Body).Decode(&dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}

// CancelTransactionDispute cancels a dispute.
func (c *Client) CancelTransactionDispute(ctx context.Context, disputeID string) (*TransactionDispute, error) {
	if err := ValidateResourceID(disputeID, "dispute"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/issuing/transaction_disputes/"+url.PathEscape(disputeID)+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var dispute TransactionDispute
	if err := json.NewDecoder(resp.Body).Decode(&dispute); err != nil {
		return nil, err
	}
	return &dispute, nil
}
