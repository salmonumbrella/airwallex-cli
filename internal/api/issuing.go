package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// Card represents an Airwallex issued card
type Card struct {
	CardID       string `json:"card_id"`
	CardNumber   string `json:"card_number"`
	CardStatus   string `json:"card_status"`
	NickName     string `json:"nick_name"`
	CardholderID string `json:"cardholder_id"`
	Brand        string `json:"brand"`
	FormFactor   string `json:"form_factor"`
	CreatedAt    string `json:"created_at"`
}

type CardsResponse struct {
	Items   []Card `json:"items"`
	HasMore bool   `json:"has_more"`
}

// CardDetails contains HIGHLY SENSITIVE payment card data (PCI DSS Level 1).
// WARNING: Never log, cache, or persist this data.
// Only display to authorized users and immediately discard after use.
type CardDetails struct {
	CardID      string `json:"card_id"`
	CardNumber  string `json:"card_number"`
	Cvv         string `json:"cvv"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

// MaskedPAN returns the card number with all but last 4 digits masked.
func (cd *CardDetails) MaskedPAN() string {
	if len(cd.CardNumber) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(cd.CardNumber)-4) + cd.CardNumber[len(cd.CardNumber)-4:]
}

// Zeroize clears all sensitive card data from memory.
// Call this when done using the card details.
func (cd *CardDetails) Zeroize() {
	cd.CardNumber = ""
	cd.Cvv = ""
	cd.ExpiryMonth = 0
	cd.ExpiryYear = 0
}

// CardLimits contains spending limits for a card
type CardLimits struct {
	Currency string      `json:"currency"`
	Limits   []CardLimit `json:"limits"`
}

// CardLimit represents a single spending limit
type CardLimit struct {
	Amount    float64 `json:"amount"`
	Interval  string  `json:"interval"`
	Remaining float64 `json:"remaining"`
}

// Cardholder represents an Airwallex cardholder
type Cardholder struct {
	CardholderID string `json:"cardholder_id"`
	Type         string `json:"type"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

type CardholdersResponse struct {
	Items   []Cardholder `json:"items"`
	HasMore bool         `json:"has_more"`
}

// Transaction represents an issuing transaction
type Transaction struct {
	TransactionID   string  `json:"transaction_id"`
	CardID          string  `json:"card_id"`
	CardNickname    string  `json:"card_nickname"`
	TransactionType string  `json:"transaction_type"`
	Amount          float64 `json:"transaction_amount"`
	Currency        string  `json:"transaction_currency"`
	BillingAmount   float64 `json:"billing_amount"`
	BillingCurrency string  `json:"billing_currency"`
	Merchant        struct {
		Name string `json:"name"`
	} `json:"merchant"`
	Status          string `json:"status"`
	TransactionDate string `json:"transaction_date"`
}

type TransactionsResponse struct {
	Items   []Transaction `json:"items"`
	HasMore bool          `json:"has_more"`
}

// ListCards lists all cards with optional filters
func (c *Client) ListCards(ctx context.Context, status, cardholderID string, pageNum, pageSize int) (*CardsResponse, error) {
	if cardholderID != "" {
		if err := ValidateResourceID(cardholderID, "cardholder"); err != nil {
			return nil, err
		}
	}
	params := url.Values{}
	if status != "" {
		params.Set("card_status", status)
	}
	if cardholderID != "" {
		params.Set("cardholder_id", cardholderID)
	}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/issuing/cards"
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

	var result CardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCard retrieves a single card by ID
func (c *Client) GetCard(ctx context.Context, cardID string) (*Card, error) {
	if err := ValidateResourceID(cardID, "card"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/issuing/cards/"+url.PathEscape(cardID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}
	return &card, nil
}

// GetCardDetails retrieves sensitive card details (PAN, CVV)
func (c *Client) GetCardDetails(ctx context.Context, cardID string) (*CardDetails, error) {
	if err := ValidateResourceID(cardID, "card"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/issuing/cards/"+url.PathEscape(cardID)+"/details")
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var details CardDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}
	return &details, nil
}

// GetCardLimits retrieves remaining spending limits for a card
func (c *Client) GetCardLimits(ctx context.Context, cardID string) (*CardLimits, error) {
	if err := ValidateResourceID(cardID, "card"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/issuing/cards/"+url.PathEscape(cardID)+"/limits")
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var limits CardLimits
	if err := json.NewDecoder(resp.Body).Decode(&limits); err != nil {
		return nil, err
	}
	return &limits, nil
}

// UpdateCard updates a card
func (c *Client) UpdateCard(ctx context.Context, cardID string, update map[string]interface{}) (*Card, error) {
	if err := ValidateResourceID(cardID, "card"); err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "/api/v1/issuing/cards/"+url.PathEscape(cardID)+"/update", update)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}
	return &card, nil
}

// ActivateCard activates a physical card
func (c *Client) ActivateCard(ctx context.Context, cardID string) (*Card, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	if err := ValidateResourceID(cardID, "card"); err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "/api/v1/issuing/cards/"+url.PathEscape(cardID)+"/activate", nil)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, err
	}
	return &card, nil
}

// CreateCard creates a new card
func (c *Client) CreateCard(ctx context.Context, req map[string]interface{}) (*Card, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, "/api/v1/issuing/cards/create", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	body, _ := io.ReadAll(resp.Body)

	// Accept 200, 201, and 202 (Accepted) as success
	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 202 {
		return nil, ParseAPIError(body)
	}

	var card Card
	if err := json.Unmarshal(body, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

// ListCardholders lists all cardholders
func (c *Client) ListCardholders(ctx context.Context, pageNum, pageSize int) (*CardholdersResponse, error) {
	params := url.Values{}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/issuing/cardholders"
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

	var result CardholdersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetCardholder retrieves a single cardholder by ID
func (c *Client) GetCardholder(ctx context.Context, cardholderID string) (*Cardholder, error) {
	if err := ValidateResourceID(cardholderID, "cardholder"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/issuing/cardholders/"+url.PathEscape(cardholderID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var ch Cardholder
	if err := json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// CreateCardholder creates a new cardholder
func (c *Client) CreateCardholder(ctx context.Context, req map[string]interface{}) (*Cardholder, error) {
	resp, err := c.Post(ctx, "/api/v1/issuing/cardholders/create", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var ch Cardholder
	if err := json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// UpdateCardholder updates a cardholder
func (c *Client) UpdateCardholder(ctx context.Context, cardholderID string, update map[string]interface{}) (*Cardholder, error) {
	if err := ValidateResourceID(cardholderID, "cardholder"); err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "/api/v1/issuing/cardholders/"+url.PathEscape(cardholderID)+"/update", update)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var ch Cardholder
	if err := json.NewDecoder(resp.Body).Decode(&ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// ListTransactions lists issuing transactions
func (c *Client) ListTransactions(ctx context.Context, cardID string, from, to string, pageNum, pageSize int) (*TransactionsResponse, error) {
	if cardID != "" {
		if err := ValidateResourceID(cardID, "card"); err != nil {
			return nil, err
		}
	}
	params := url.Values{}
	if cardID != "" {
		params.Set("card_id", cardID)
	}
	if from != "" {
		params.Set("from_created_at", from)
	}
	if to != "" {
		params.Set("to_created_at", to)
	}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/issuing/transactions"
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

	var result TransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTransaction retrieves a single transaction by ID
func (c *Client) GetTransaction(ctx context.Context, transactionID string) (*Transaction, error) {
	if err := ValidateResourceID(transactionID, "transaction"); err != nil {
		return nil, err
	}
	resp, err := c.Get(ctx, "/api/v1/issuing/transactions/"+url.PathEscape(transactionID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var txn Transaction
	if err := json.NewDecoder(resp.Body).Decode(&txn); err != nil {
		return nil, err
	}
	return &txn, nil
}
