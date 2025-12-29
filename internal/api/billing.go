package api

import (
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strconv"
)

// BillingCustomer represents a billing customer (payment acceptance customer).
type BillingCustomer struct {
	ID                 string `json:"id"`
	MerchantCustomerID string `json:"merchant_customer_id"`
	BusinessName       string `json:"business_name"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	Email              string `json:"email"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type BillingCustomersResponse struct {
	Items   []BillingCustomer `json:"items"`
	HasMore bool              `json:"has_more"`
}

// BillingProduct represents a billing product.
type BillingProduct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
	Active      bool   `json:"active"`
}

type BillingProductsResponse struct {
	Items   []BillingProduct `json:"items"`
	HasMore bool             `json:"has_more"`
}

// BillingPriceRecurring represents recurring price details.
type BillingPriceRecurring struct {
	Period     int    `json:"period"`
	PeriodUnit string `json:"period_unit"`
}

// BillingPrice represents a billing price.
type BillingPrice struct {
	ID           string                 `json:"id"`
	ProductID    string                 `json:"product_id"`
	Currency     string                 `json:"currency"`
	UnitAmount   float64                `json:"unit_amount"`
	FlatAmount   float64                `json:"flat_amount"`
	PricingModel string                 `json:"pricing_model"`
	Type         string                 `json:"type"`
	Active       bool                   `json:"active"`
	Recurring    *BillingPriceRecurring `json:"recurring"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
}

type BillingPricesResponse struct {
	Items   []BillingPrice `json:"items"`
	HasMore bool           `json:"has_more"`
}

// BillingInvoice represents a billing invoice.
type BillingInvoice struct {
	ID                           string  `json:"id"`
	CustomerID                   string  `json:"customer_id"`
	SubscriptionID               string  `json:"subscription_id"`
	Status                       string  `json:"status"`
	Currency                     string  `json:"currency"`
	TotalAmount                  float64 `json:"total_amount"`
	PeriodStartAt                string  `json:"period_start_at"`
	PeriodEndAt                  string  `json:"period_end_at"`
	PaidAt                       string  `json:"paid_at"`
	CreatedAt                    string  `json:"created_at"`
	UpdatedAt                    string  `json:"updated_at"`
	PaymentIntentID              string  `json:"payment_intent_id"`
	LastPaymentAttemptAt         string  `json:"last_payment_attempt_at"`
	NextPaymentAttemptAt         string  `json:"next_payment_attempt_at"`
	PastPaymentAttemptCount      int     `json:"past_payment_attempt_count"`
	RemainingPaymentAttemptCount int     `json:"remaining_payment_attempt_count"`
}

type BillingInvoicesResponse struct {
	Items   []BillingInvoice `json:"items"`
	HasMore bool             `json:"has_more"`
}

// BillingInvoicePreview represents invoice preview response.
type BillingInvoicePreview struct {
	CreatedAt      string               `json:"created_at"`
	Currency       string               `json:"currency"`
	CustomerID     string               `json:"customer_id"`
	SubscriptionID string               `json:"subscription_id"`
	TotalAmount    float64              `json:"total_amount"`
	Items          []BillingInvoiceItem `json:"items"`
}

// BillingInvoiceItem represents an invoice line item.
type BillingInvoiceItem struct {
	ID            string        `json:"id"`
	InvoiceID     string        `json:"invoice_id"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Quantity      float64       `json:"quantity"`
	PeriodStartAt string        `json:"period_start_at"`
	PeriodEndAt   string        `json:"period_end_at"`
	Price         *BillingPrice `json:"price"`
}

type BillingInvoiceItemsResponse struct {
	Items   []BillingInvoiceItem `json:"items"`
	HasMore bool                 `json:"has_more"`
}

// BillingSubscription represents a billing subscription.
type BillingSubscription struct {
	ID                     string `json:"id"`
	CustomerID             string `json:"customer_id"`
	Status                 string `json:"status"`
	CurrentPeriodStartAt   string `json:"current_period_start_at"`
	CurrentPeriodEndAt     string `json:"current_period_end_at"`
	NextBillingAt          string `json:"next_billing_at"`
	TrialStartAt           string `json:"trial_start_at"`
	TrialEndAt             string `json:"trial_end_at"`
	CancelAt               string `json:"cancel_at"`
	CancelAtPeriodEnd      bool   `json:"cancel_at_period_end"`
	CancelRequestedAt      string `json:"cancel_requested_at"`
	LatestInvoiceID        string `json:"latest_invoice_id"`
	RemainingBillingCycles int    `json:"remaining_billing_cycles"`
	TotalBillingCycles     int    `json:"total_billing_cycles"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

type BillingSubscriptionsResponse struct {
	Items   []BillingSubscription `json:"items"`
	HasMore bool                  `json:"has_more"`
}

// BillingSubscriptionItem represents a subscription line item.
type BillingSubscriptionItem struct {
	ID             string        `json:"id"`
	SubscriptionID string        `json:"subscription_id"`
	Quantity       float64       `json:"quantity"`
	Price          *BillingPrice `json:"price"`
}

type BillingSubscriptionItemsResponse struct {
	Items   []BillingSubscriptionItem `json:"items"`
	HasMore bool                      `json:"has_more"`
}

// Billing list params

type BillingCustomerListParams struct {
	MerchantCustomerID string
	FromCreatedAt      string
	ToCreatedAt        string
	PageNum            int
	PageSize           int
}

type BillingProductListParams struct {
	Active   *bool
	PageNum  int
	PageSize int
}

type BillingPriceListParams struct {
	Active              *bool
	Currency            string
	ProductID           string
	RecurringPeriod     int
	RecurringPeriodUnit string
	PageNum             int
	PageSize            int
}

type BillingInvoiceListParams struct {
	CustomerID     string
	SubscriptionID string
	Status         string
	FromCreatedAt  string
	ToCreatedAt    string
	PageNum        int
	PageSize       int
}

type BillingSubscriptionListParams struct {
	CustomerID          string
	Status              string
	RecurringPeriod     int
	RecurringPeriodUnit string
	FromCreatedAt       string
	ToCreatedAt         string
	PageNum             int
	PageSize            int
}

func addPagination(params url.Values, pageNum, pageSize int) {
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1
		}
		params.Set("page_num", strconv.Itoa(pageNum))
		params.Set("page_size", strconv.Itoa(pageSize))
	}
}

func addBool(params url.Values, key string, value *bool) {
	if value != nil {
		params.Set(key, strconv.FormatBool(*value))
	}
}

// ListBillingCustomers lists billing customers.
func (c *Client) ListBillingCustomers(ctx context.Context, params BillingCustomerListParams) (*BillingCustomersResponse, error) {
	query := url.Values{}
	if params.MerchantCustomerID != "" {
		query.Set("merchant_customer_id", params.MerchantCustomerID)
	}
	if params.FromCreatedAt != "" {
		query.Set("from_created_at", params.FromCreatedAt)
	}
	if params.ToCreatedAt != "" {
		query.Set("to_created_at", params.ToCreatedAt)
	}
	addPagination(query, params.PageNum, params.PageSize)

	path := Endpoints.BillingCustomersList.Path
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

	var result BillingCustomersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBillingCustomer retrieves a billing customer by ID.
func (c *Client) GetBillingCustomer(ctx context.Context, customerID string) (*BillingCustomer, error) {
	if err := ValidateResourceID(customerID, "customer"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/pa/customers/"+url.PathEscape(customerID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var customer BillingCustomer
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// CreateBillingCustomer creates a billing customer.
func (c *Client) CreateBillingCustomer(ctx context.Context, req map[string]interface{}) (*BillingCustomer, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.BillingCustomersCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var customer BillingCustomer
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// UpdateBillingCustomer updates a billing customer.
func (c *Client) UpdateBillingCustomer(ctx context.Context, customerID string, req map[string]interface{}) (*BillingCustomer, error) {
	if err := ValidateResourceID(customerID, "customer"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/pa/customers/"+url.PathEscape(customerID)+"/update", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var customer BillingCustomer
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return nil, err
	}
	return &customer, nil
}

// ListBillingProducts lists billing products.
func (c *Client) ListBillingProducts(ctx context.Context, params BillingProductListParams) (*BillingProductsResponse, error) {
	query := url.Values{}
	addBool(query, "active", params.Active)
	addPagination(query, params.PageNum, params.PageSize)

	path := Endpoints.BillingProductsList.Path
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

	var result BillingProductsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBillingProduct retrieves a billing product by ID.
func (c *Client) GetBillingProduct(ctx context.Context, productID string) (*BillingProduct, error) {
	if err := ValidateResourceID(productID, "product"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/products/"+url.PathEscape(productID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var product BillingProduct
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}
	return &product, nil
}

// CreateBillingProduct creates a billing product.
func (c *Client) CreateBillingProduct(ctx context.Context, req map[string]interface{}) (*BillingProduct, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.BillingProductsCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var product BillingProduct
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}
	return &product, nil
}

// UpdateBillingProduct updates a billing product.
func (c *Client) UpdateBillingProduct(ctx context.Context, productID string, req map[string]interface{}) (*BillingProduct, error) {
	if err := ValidateResourceID(productID, "product"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/products/"+url.PathEscape(productID)+"/update", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var product BillingProduct
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}
	return &product, nil
}

// ListBillingPrices lists billing prices.
func (c *Client) ListBillingPrices(ctx context.Context, params BillingPriceListParams) (*BillingPricesResponse, error) {
	query := url.Values{}
	addBool(query, "active", params.Active)
	if params.Currency != "" {
		query.Set("currency", params.Currency)
	}
	if params.ProductID != "" {
		query.Set("product_id", params.ProductID)
	}
	if params.RecurringPeriod > 0 {
		query.Set("recurring_period", strconv.Itoa(params.RecurringPeriod))
	}
	if params.RecurringPeriodUnit != "" {
		query.Set("recurring_period_unit", params.RecurringPeriodUnit)
	}
	addPagination(query, params.PageNum, params.PageSize)

	path := Endpoints.BillingPricesList.Path
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

	var result BillingPricesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBillingPrice retrieves a billing price by ID.
func (c *Client) GetBillingPrice(ctx context.Context, priceID string) (*BillingPrice, error) {
	if err := ValidateResourceID(priceID, "price"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/prices/"+url.PathEscape(priceID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var price BillingPrice
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return nil, err
	}
	return &price, nil
}

// CreateBillingPrice creates a billing price.
func (c *Client) CreateBillingPrice(ctx context.Context, req map[string]interface{}) (*BillingPrice, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.BillingPricesCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var price BillingPrice
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return nil, err
	}
	return &price, nil
}

// UpdateBillingPrice updates a billing price.
func (c *Client) UpdateBillingPrice(ctx context.Context, priceID string, req map[string]interface{}) (*BillingPrice, error) {
	if err := ValidateResourceID(priceID, "price"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/prices/"+url.PathEscape(priceID)+"/update", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var price BillingPrice
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return nil, err
	}
	return &price, nil
}

// ListBillingInvoices lists billing invoices.
func (c *Client) ListBillingInvoices(ctx context.Context, params BillingInvoiceListParams) (*BillingInvoicesResponse, error) {
	query := url.Values{}
	if params.CustomerID != "" {
		query.Set("customer_id", params.CustomerID)
	}
	if params.SubscriptionID != "" {
		query.Set("subscription_id", params.SubscriptionID)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.FromCreatedAt != "" {
		query.Set("from_created_at", params.FromCreatedAt)
	}
	if params.ToCreatedAt != "" {
		query.Set("to_created_at", params.ToCreatedAt)
	}
	addPagination(query, params.PageNum, params.PageSize)

	path := Endpoints.BillingInvoicesList.Path
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

	var result BillingInvoicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBillingInvoice retrieves a billing invoice by ID.
func (c *Client) GetBillingInvoice(ctx context.Context, invoiceID string) (*BillingInvoice, error) {
	if err := ValidateResourceID(invoiceID, "invoice"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/invoices/"+url.PathEscape(invoiceID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var invoice BillingInvoice
	if err := json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// CreateBillingInvoice creates a billing invoice.
func (c *Client) CreateBillingInvoice(ctx context.Context, req map[string]interface{}) (*BillingInvoice, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.BillingInvoicesCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var invoice BillingInvoice
	if err := json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// PreviewBillingInvoice previews a billing invoice.
func (c *Client) PreviewBillingInvoice(ctx context.Context, req map[string]interface{}) (*BillingInvoicePreview, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.BillingInvoicesPreview.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var preview BillingInvoicePreview
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		return nil, err
	}
	return &preview, nil
}

// ListBillingInvoiceItems lists invoice items for an invoice.
func (c *Client) ListBillingInvoiceItems(ctx context.Context, invoiceID string, pageNum, pageSize int) (*BillingInvoiceItemsResponse, error) {
	if err := ValidateResourceID(invoiceID, "invoice"); err != nil {
		return nil, err
	}
	query := url.Values{}
	addPagination(query, pageNum, pageSize)

	path := "/api/v1/invoices/" + url.PathEscape(invoiceID) + "/items"
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

	var result BillingInvoiceItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBillingInvoiceItem retrieves an invoice item by ID.
func (c *Client) GetBillingInvoiceItem(ctx context.Context, invoiceID, itemID string) (*BillingInvoiceItem, error) {
	if err := ValidateResourceID(invoiceID, "invoice"); err != nil {
		return nil, err
	}
	if err := ValidateResourceID(itemID, "item"); err != nil {
		return nil, err
	}

	path := "/api/v1/invoices/" + url.PathEscape(invoiceID) + "/items/" + url.PathEscape(itemID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var item BillingInvoiceItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

// ListBillingSubscriptions lists billing subscriptions.
func (c *Client) ListBillingSubscriptions(ctx context.Context, params BillingSubscriptionListParams) (*BillingSubscriptionsResponse, error) {
	query := url.Values{}
	if params.CustomerID != "" {
		query.Set("customer_id", params.CustomerID)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.RecurringPeriod > 0 {
		query.Set("recurring_period", strconv.Itoa(params.RecurringPeriod))
	}
	if params.RecurringPeriodUnit != "" {
		query.Set("recurring_period_unit", params.RecurringPeriodUnit)
	}
	if params.FromCreatedAt != "" {
		query.Set("from_created_at", params.FromCreatedAt)
	}
	if params.ToCreatedAt != "" {
		query.Set("to_created_at", params.ToCreatedAt)
	}
	addPagination(query, params.PageNum, params.PageSize)

	path := Endpoints.BillingSubscriptionsList.Path
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

	var result BillingSubscriptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBillingSubscription retrieves a billing subscription by ID.
func (c *Client) GetBillingSubscription(ctx context.Context, subscriptionID string) (*BillingSubscription, error) {
	if err := ValidateResourceID(subscriptionID, "subscription"); err != nil {
		return nil, err
	}

	resp, err := c.Get(ctx, "/api/v1/subscriptions/"+url.PathEscape(subscriptionID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var sub BillingSubscription
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// CreateBillingSubscription creates a billing subscription.
func (c *Client) CreateBillingSubscription(ctx context.Context, req map[string]interface{}) (*BillingSubscription, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	resp, err := c.Post(ctx, Endpoints.BillingSubscriptionsCreate.Path, req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var sub BillingSubscription
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// UpdateBillingSubscription updates a billing subscription.
func (c *Client) UpdateBillingSubscription(ctx context.Context, subscriptionID string, req map[string]interface{}) (*BillingSubscription, error) {
	if err := ValidateResourceID(subscriptionID, "subscription"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/subscriptions/"+url.PathEscape(subscriptionID)+"/update", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var sub BillingSubscription
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// CancelBillingSubscription cancels a billing subscription.
func (c *Client) CancelBillingSubscription(ctx context.Context, subscriptionID string, req map[string]interface{}) (*BillingSubscription, error) {
	if err := ValidateResourceID(subscriptionID, "subscription"); err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v1/subscriptions/"+url.PathEscape(subscriptionID)+"/cancel", req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var sub BillingSubscription
	if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListBillingSubscriptionItems lists subscription items.
func (c *Client) ListBillingSubscriptionItems(ctx context.Context, subscriptionID string, pageNum, pageSize int) (*BillingSubscriptionItemsResponse, error) {
	if err := ValidateResourceID(subscriptionID, "subscription"); err != nil {
		return nil, err
	}
	query := url.Values{}
	addPagination(query, pageNum, pageSize)

	path := "/api/v1/subscriptions/" + url.PathEscape(subscriptionID) + "/items"
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

	var result BillingSubscriptionItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBillingSubscriptionItem retrieves a subscription item by ID.
func (c *Client) GetBillingSubscriptionItem(ctx context.Context, subscriptionID, itemID string) (*BillingSubscriptionItem, error) {
	if err := ValidateResourceID(subscriptionID, "subscription"); err != nil {
		return nil, err
	}
	if err := ValidateResourceID(itemID, "item"); err != nil {
		return nil, err
	}

	path := "/api/v1/subscriptions/" + url.PathEscape(subscriptionID) + "/items/" + url.PathEscape(itemID)
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(body)
	}

	var item BillingSubscriptionItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}
