package api

import (
	"context"
	"encoding/json"
	"io"
)

// SchemaField represents a field in a dynamic schema
type SchemaField struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	MinLength   int      `json:"min_length,omitempty"`
	MaxLength   int      `json:"max_length,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
}

// Schema represents a dynamic API schema response
type Schema struct {
	Fields []SchemaField `json:"fields"`
}

// GetBeneficiarySchema retrieves the schema for creating a beneficiary
func (c *Client) GetBeneficiarySchema(ctx context.Context, bankCountry, entityType, paymentMethod string) (*Schema, error) {
	req := map[string]interface{}{
		"bank_country_code": bankCountry,
		"entity_type":       entityType,
	}
	if paymentMethod != "" {
		req["payment_method"] = paymentMethod
	}

	resp, err := c.Post(ctx, "/api/v1/beneficiary_api_schemas/generate", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var schema Schema
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// GetTransferSchema retrieves the schema for creating a transfer
func (c *Client) GetTransferSchema(ctx context.Context, sourceCurrency, destCurrency, paymentMethod string) (*Schema, error) {
	req := map[string]interface{}{
		"source_currency":      sourceCurrency,
		"destination_currency": destCurrency,
	}
	if paymentMethod != "" {
		req["payment_method"] = paymentMethod
	}

	resp, err := c.Post(ctx, "/api/v1/transfer_api_schemas/generate", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var schema Schema
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}
