package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

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
	Nickname        string             `json:"nickname,omitempty"`
	Beneficiary     BeneficiaryDetails `json:"beneficiary"`
	TransferMethods []string           `json:"transfer_methods"`
}

type BeneficiariesResponse struct {
	Items   []Beneficiary `json:"items"`
	HasMore bool          `json:"has_more"`
}

// ListTransfers lists all transfers
func (c *Client) ListTransfers(status string, pageNum, pageSize int) (*TransfersResponse, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if pageNum > 0 {
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/transfers"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var result TransfersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTransfer retrieves a single transfer by ID
func (c *Client) GetTransfer(transferID string) (*Transfer, error) {
	resp, err := c.Get("/api/v1/transfers/" + url.PathEscape(transferID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var t Transfer
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTransfer creates a new transfer
func (c *Client) CreateTransfer(req map[string]interface{}) (*Transfer, error) {
	resp, err := c.Post("/api/v1/transfers/create", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var t Transfer
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CancelTransfer cancels a transfer
func (c *Client) CancelTransfer(transferID string) (*Transfer, error) {
	resp, err := c.Post("/api/v1/transfers/"+url.PathEscape(transferID)+"/cancel", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var t Transfer
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// ListBeneficiaries lists all beneficiaries
func (c *Client) ListBeneficiaries(pageNum, pageSize int) (*BeneficiariesResponse, error) {
	params := url.Values{}
	if pageNum > 0 {
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/beneficiaries"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var result BeneficiariesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBeneficiary retrieves a single beneficiary by ID
func (c *Client) GetBeneficiary(beneficiaryID string) (*Beneficiary, error) {
	resp, err := c.Get("/api/v1/beneficiaries/" + url.PathEscape(beneficiaryID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var b Beneficiary
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBeneficiary creates a new beneficiary
func (c *Client) CreateBeneficiary(req map[string]interface{}) (*Beneficiary, error) {
	resp, err := c.Post("/api/v1/beneficiaries/create", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var b Beneficiary
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// UpdateBeneficiary updates a beneficiary
func (c *Client) UpdateBeneficiary(beneficiaryID string, update map[string]interface{}) (*Beneficiary, error) {
	resp, err := c.Post("/api/v1/beneficiaries/"+url.PathEscape(beneficiaryID)+"/update", update)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var b Beneficiary
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// DeleteBeneficiary deletes a beneficiary
func (c *Client) DeleteBeneficiary(beneficiaryID string) error {
	resp, err := c.Post("/api/v1/beneficiaries/"+url.PathEscape(beneficiaryID)+"/delete", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return ParseAPIError(body)
	}
	return nil
}

// ValidateBeneficiary validates beneficiary details without creating
func (c *Client) ValidateBeneficiary(req map[string]interface{}) error {
	resp, err := c.Post("/api/v1/beneficiaries/validate", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return ParseAPIError(body)
	}
	return nil
}

// GetConfirmationLetter retrieves a transfer confirmation letter as PDF
func (c *Client) GetConfirmationLetter(transferID string, format string) ([]byte, error) {
	req := map[string]interface{}{
		"transaction_id": transferID,
		"format":         format,
	}

	resp, err := c.Post("/api/v1/confirmation_letters/create", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF response: %w", err)
	}

	return pdfData, nil
}
