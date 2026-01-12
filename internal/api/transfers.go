package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/salmonumbrella/airwallex-cli/internal/wait"
)

// TransferFinalStatuses are statuses that indicate the transfer is complete
var TransferFinalStatuses = map[string]bool{
	"COMPLETED": true,
	"FAILED":    true,
	"CANCELLED": true,
	"RETURNED":  true,
}

// Transfer represents a transfer/payout
type Transfer struct {
	TransferID       string  `json:"id"`
	BeneficiaryID    string  `json:"beneficiary_id"`
	TransferAmount   float64 `json:"transfer_amount"`
	TransferCurrency string  `json:"transfer_currency"`
	SourceAmount     float64 `json:"source_amount"`
	SourceCurrency   string  `json:"source_currency"`
	Status           string  `json:"status"`
	Reference        string  `json:"reference"`
	Reason           string  `json:"reason"`
	CreatedAt        string  `json:"created_at"`
}

type TransfersResponse struct {
	Items   []Transfer `json:"items"`
	HasMore bool       `json:"has_more"`
}

// BeneficiaryDetails contains the nested beneficiary information
type BeneficiaryDetails struct {
	EntityType  string `json:"entity_type"`
	CompanyName string `json:"company_name,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	BankDetails struct {
		BankCountryCode string `json:"bank_country_code"`
		BankName        string `json:"bank_name"`
		AccountName     string `json:"account_name"`
	} `json:"bank_details"`
}

// Beneficiary represents a transfer beneficiary
type Beneficiary struct {
	BeneficiaryID   string             `json:"id"`
	Nickname        string             `json:"nickname"`
	Beneficiary     BeneficiaryDetails `json:"beneficiary"`
	TransferMethods []string           `json:"transfer_methods"`
}

type BeneficiariesResponse struct {
	Items   []Beneficiary `json:"items"`
	HasMore bool          `json:"has_more"`
}

// ListTransfers lists all transfers
func (c *Client) ListTransfers(ctx context.Context, status string, pageNum, pageSize int) (*TransfersResponse, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/transfers"
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

	var result TransfersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTransfer retrieves a single transfer by ID
func (c *Client) GetTransfer(ctx context.Context, transferID string) (*Transfer, error) {
	if err := ValidateResourceID(transferID, "transfer"); err != nil {
		return nil, err
	}
	path := "/api/v1/transfers/" + url.PathEscape(transferID)
	var t Transfer
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTransfer creates a new transfer
func (c *Client) CreateTransfer(ctx context.Context, req map[string]interface{}) (*Transfer, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.TransfersCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", Endpoints.TransfersCreate.Path, resp.StatusCode, ParseAPIError(body))
	}

	var t Transfer
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CancelTransfer cancels a transfer
func (c *Client) CancelTransfer(ctx context.Context, transferID string) (*Transfer, error) {
	if err := ValidateResourceID(transferID, "transfer"); err != nil {
		return nil, err
	}
	path := "/api/v1/transfers/" + url.PathEscape(transferID) + "/cancel"
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}

	var t Transfer
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// WaitForTransfer polls until the transfer reaches a final status.
// Uses the unified wait pattern for consistent polling behavior.
func (c *Client) WaitForTransfer(ctx context.Context, transferID string, timeout time.Duration) (*Transfer, error) {
	if err := ValidateResourceID(transferID, "transfer"); err != nil {
		return nil, err
	}

	cfg := wait.Config{
		Timeout:       timeout,
		PollInterval:  2 * time.Second,
		SuccessStates: []string{"COMPLETED"},
		FailureStates: []string{"FAILED", "CANCELLED", "RETURNED"},
	}

	var transfer *Transfer
	_, err := wait.For(ctx, cfg, func() (string, error) {
		t, err := c.GetTransfer(ctx, transferID)
		if err != nil {
			return "", err
		}
		transfer = t
		return t.Status, nil
	})

	return transfer, err
}

// ListBeneficiaries lists all beneficiaries
func (c *Client) ListBeneficiaries(ctx context.Context, pageNum, pageSize int) (*BeneficiariesResponse, error) {
	params := url.Values{}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/beneficiaries"
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

	var result BeneficiariesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBeneficiary retrieves a single beneficiary by ID
func (c *Client) GetBeneficiary(ctx context.Context, beneficiaryID string) (*Beneficiary, error) {
	if err := ValidateResourceID(beneficiaryID, "beneficiary"); err != nil {
		return nil, err
	}
	path := "/api/v1/beneficiaries/" + url.PathEscape(beneficiaryID)
	var b Beneficiary
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBeneficiaryRaw returns the full beneficiary data as a map for merging with updates
func (c *Client) GetBeneficiaryRaw(ctx context.Context, beneficiaryID string) (map[string]interface{}, error) {
	if err := ValidateResourceID(beneficiaryID, "beneficiary"); err != nil {
		return nil, err
	}
	path := "/api/v1/beneficiaries/" + url.PathEscape(beneficiaryID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateBeneficiary creates a new beneficiary
func (c *Client) CreateBeneficiary(ctx context.Context, req map[string]interface{}) (*Beneficiary, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	path := "/api/v1/beneficiaries/create"
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}

	var b Beneficiary
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// UpdateBeneficiary updates a beneficiary
func (c *Client) UpdateBeneficiary(ctx context.Context, beneficiaryID string, update map[string]interface{}) (*Beneficiary, error) {
	if err := ValidateResourceID(beneficiaryID, "beneficiary"); err != nil {
		return nil, err
	}

	path := "/api/v1/beneficiaries/" + url.PathEscape(beneficiaryID) + "/update"
	resp, err := c.Post(ctx, path, update)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}

	var b Beneficiary
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// DeleteBeneficiary deletes a beneficiary
func (c *Client) DeleteBeneficiary(ctx context.Context, beneficiaryID string) error {
	if err := ValidateResourceID(beneficiaryID, "beneficiary"); err != nil {
		return err
	}
	path := "/api/v1/beneficiaries/" + url.PathEscape(beneficiaryID) + "/delete"
	resp, err := c.Post(ctx, path, nil)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}
	return nil
}

// ValidateBeneficiary validates beneficiary details without creating
func (c *Client) ValidateBeneficiary(ctx context.Context, req map[string]interface{}) error {
	path := "/api/v1/beneficiaries/validate"
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}
	return nil
}

// GetConfirmationLetter retrieves a transfer confirmation letter as PDF
func (c *Client) GetConfirmationLetter(ctx context.Context, transferID string, format string) ([]byte, error) {
	if err := ValidateResourceID(transferID, "transfer"); err != nil {
		return nil, err
	}
	req := map[string]interface{}{
		"transaction_id": transferID,
		"format":         format,
	}

	path := "/api/v1/confirmation_letters/create"
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", path, resp.StatusCode, ParseAPIError(body))
	}

	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF response: %w", err)
	}

	return pdfData, nil
}
