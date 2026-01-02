package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Rate represents current exchange rate
type Rate struct {
	SellCurrency string  `json:"sell_currency"`
	BuyCurrency  string  `json:"buy_currency"`
	Rate         float64 `json:"rate"`
	RateType     string  `json:"rate_type"`
}

type RatesResponse struct {
	Rates []Rate `json:"rates"`
}

// Quote represents a locked FX rate quote
type Quote struct {
	ID             string  `json:"quote_id"`
	SellCurrency   string  `json:"sell_currency"`
	BuyCurrency    string  `json:"buy_currency"`
	SellAmount     float64 `json:"sell_amount,omitempty"`
	BuyAmount      float64 `json:"buy_amount,omitempty"`
	Rate           float64 `json:"client_rate"`
	RateExpiry     string  `json:"valid_to_at"`
	ValidityPeriod string  `json:"validity"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"valid_from_at"`
}

// Conversion represents an executed currency conversion
type Conversion struct {
	ID           string  `json:"id"`
	QuoteID      string  `json:"quote_id,omitempty"`
	SellCurrency string  `json:"sell_currency"`
	BuyCurrency  string  `json:"buy_currency"`
	SellAmount   float64 `json:"sell_amount"`
	BuyAmount    float64 `json:"buy_amount"`
	Rate         float64 `json:"rate"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

type ConversionsResponse struct {
	Items   []Conversion `json:"items"`
	HasMore bool         `json:"has_more"`
}

// GetRates retrieves current exchange rates
func (c *Client) GetRates(ctx context.Context, sellCurrency, buyCurrency string) (*RatesResponse, error) {
	params := url.Values{}
	if sellCurrency != "" {
		params.Set("sell_currency", sellCurrency)
	}
	if buyCurrency != "" {
		params.Set("buy_currency", buyCurrency)
	}

	path := "/api/v1/fx/rates/current"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
	}

	// Try parsing as array response first (RatesResponse)
	var result RatesResponse
	if err := json.Unmarshal(body, &result); err == nil && len(result.Rates) > 0 {
		return &result, nil
	}

	// The API returns a single rate object for /rates/current
	var singleRate Rate
	if err := json.Unmarshal(body, &singleRate); err != nil {
		return nil, fmt.Errorf("failed to parse rates response: %w", err)
	}
	// Only wrap if we got actual data
	if singleRate.SellCurrency != "" {
		return &RatesResponse{Rates: []Rate{singleRate}}, nil
	}
	return &result, nil
}

// CreateQuote creates a new FX quote to lock in a rate
func (c *Client) CreateQuote(ctx context.Context, req map[string]interface{}) (*Quote, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.FXQuotesCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	// Accept both 200 and 201 for backward compatibility
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", Endpoints.FXQuotesCreate.Path, resp.StatusCode, ParseAPIError(body))
	}

	var q Quote
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

// GetQuote retrieves a quote by ID
func (c *Client) GetQuote(ctx context.Context, quoteID string) (*Quote, error) {
	if err := ValidateResourceID(quoteID, "quote"); err != nil {
		return nil, err
	}

	path := "/api/v1/fx/quotes/" + url.PathEscape(quoteID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
	}

	var q Quote
	if err := json.NewDecoder(resp.Body).Decode(&q); err != nil {
		return nil, err
	}
	return &q, nil
}

// ListConversions lists all conversions with optional filters
func (c *Client) ListConversions(ctx context.Context, status string, fromDate, toDate string, pageNum, pageSize int) (*ConversionsResponse, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if fromDate != "" {
		params.Set("from_created_at", fromDate)
	}
	if toDate != "" {
		params.Set("to_created_at", toDate)
	}
	// Airwallex API requires both page_num and page_size together
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1 // API uses 1-based page numbering
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}

	path := "/api/v1/fx/conversions"
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

	var result ConversionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConversion retrieves a conversion by ID
func (c *Client) GetConversion(ctx context.Context, conversionID string) (*Conversion, error) {
	if err := ValidateResourceID(conversionID, "conversion"); err != nil {
		return nil, err
	}

	path := "/api/v1/fx/conversions/" + url.PathEscape(conversionID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("GET", path, resp.StatusCode, ParseAPIError(body))
	}

	var conv Conversion
	if err := json.NewDecoder(resp.Body).Decode(&conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

// CreateConversion executes a currency conversion
func (c *Client) CreateConversion(ctx context.Context, req map[string]interface{}) (*Conversion, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.FXConversionsCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	// Accept both 200 and 201 for backward compatibility
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, WrapError("POST", Endpoints.FXConversionsCreate.Path, resp.StatusCode, ParseAPIError(body))
	}

	var conv Conversion
	if err := json.NewDecoder(resp.Body).Decode(&conv); err != nil {
		return nil, err
	}
	return &conv, nil
}
