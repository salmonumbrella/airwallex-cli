package api

import (
	"context"
	"encoding/json"
	"io"
)

// SchemaFieldRule contains validation rules for a schema field
type SchemaFieldRule struct {
	Type      string   `json:"type,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	MinLength int      `json:"minLength,omitempty"`
	MaxLength int      `json:"maxLength,omitempty"`
}

// SchemaField represents a field in a dynamic schema
type SchemaField struct {
	Key         string          `json:"key"`
	Path        string          `json:"path,omitempty"`
	Required    bool            `json:"required"`
	Description string          `json:"description,omitempty"`
	Rule        SchemaFieldRule `json:"rule,omitempty"`
}

// Name returns the field name (from Key for API compatibility)
func (f SchemaField) Name() string {
	return f.Key
}

// Type returns the field type from the rule
func (f SchemaField) Type() string {
	return f.Rule.Type
}

// Enum returns the enum values from the rule
func (f SchemaField) Enum() []string {
	return f.Rule.Enum
}

// Pattern returns the validation pattern from the rule
func (f SchemaField) Pattern() string {
	return f.Rule.Pattern
}

// MinLength returns the minimum length from the rule
func (f SchemaField) MinLength() int {
	return f.Rule.MinLength
}

// MaxLength returns the maximum length from the rule
func (f SchemaField) MaxLength() int {
	return f.Rule.MaxLength
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
