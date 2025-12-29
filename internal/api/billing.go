package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// BillingCustomer represents a billing customer.
type BillingCustomer struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

type BillingCustomersResponse struct {
	Items   []BillingCustomer `json:"items"`
	HasMore bool              `json:"has_more"`
}

// BillingProduct represents a billing product.
type BillingProduct struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type BillingProductsResponse struct {
	Items   []BillingProduct `json:"items"`
	HasMore bool             `json:"has_more"`
}

// BillingPrice represents a billing price.
type BillingPrice struct {
	ID         string  `json:"id"`
	PriceID    string  `json:"price_id"`
	ProductID  string  `json:"product_id"`
	Currency   string  `json:"currency"`
	UnitAmount float64 `json:"unit_amount"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
}

type BillingPricesResponse struct {
	Items   []BillingPrice `json:"items"`
	HasMore bool           `json:"has_more"`
}

// BillingInvoice represents a billing invoice.
type BillingInvoice struct {
	ID            string  `json:"id"`
	InvoiceID     string  `json:"invoice_id"`
	InvoiceNumber string  `json:"invoice_number"`
	Status        string  `json:"status"`
	Currency      string  `json:"currency"`
	Amount        float64 `json:"amount"`
	DueDate       string  `json:"due_date"`
	CreatedAt     string  `json:"created_at"`
}

type BillingInvoicesResponse struct {
	Items   []BillingInvoice `json:"items"`
	HasMore bool             `json:"has_more"`
}

// BillingSubscription represents a billing subscription.
type BillingSubscription struct {
	ID             string `json:"id"`
	SubscriptionID string `json:"subscription_id"`
	CustomerID     string `json:"customer_id"`
	PriceID        string `json:"price_id"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

type BillingSubscriptionsResponse struct {
	Items   []BillingSubscription `json:"items"`
	HasMore bool                  `json:"has_more"`
}

func withPaging(path string, pageNum, pageSize int) string {
	params := url.Values{}
	if pageSize > 0 {
		if pageNum < 1 {
			pageNum = 1
		}
		params.Set("page_num", fmt.Sprintf("%d", pageNum))
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	if len(params) == 0 {
		return path
	}
	return path + "?" + params.Encode()
}

// ListBillingCustomers lists billing customers.
func (c *Client) ListBillingCustomers(ctx context.Context, pageNum, pageSize int) (*BillingCustomersResponse, error) {
	resp, err := c.Get(ctx, withPaging(Endpoints.BillingCustomersList.Path, pageNum, pageSize))
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

	resp, err := c.Get(ctx, "/api/v1/billing_customers/"+url.PathEscape(customerID))
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

	resp, err := c.Post(ctx, "/api/v1/billing_customers/"+url.PathEscape(customerID)+"/update", req)
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
func (c *Client) ListBillingProducts(ctx context.Context, pageNum, pageSize int) (*BillingProductsResponse, error) {
	resp, err := c.Get(ctx, withPaging(Endpoints.BillingProductsList.Path, pageNum, pageSize))
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
func (c *Client) ListBillingPrices(ctx context.Context, pageNum, pageSize int) (*BillingPricesResponse, error) {
	resp, err := c.Get(ctx, withPaging(Endpoints.BillingPricesList.Path, pageNum, pageSize))
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
func (c *Client) ListBillingInvoices(ctx context.Context, pageNum, pageSize int) (*BillingInvoicesResponse, error) {
	resp, err := c.Get(ctx, withPaging(Endpoints.BillingInvoicesList.Path, pageNum, pageSize))
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

// ListBillingSubscriptions lists billing subscriptions.
func (c *Client) ListBillingSubscriptions(ctx context.Context, pageNum, pageSize int) (*BillingSubscriptionsResponse, error) {
	resp, err := c.Get(ctx, withPaging(Endpoints.BillingSubscriptionsList.Path, pageNum, pageSize))
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
