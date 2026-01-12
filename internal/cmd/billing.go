package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Billing operations",
	}
	cmd.AddCommand(newBillingCustomersCmd())
	cmd.AddCommand(newBillingProductsCmd())
	cmd.AddCommand(newBillingPricesCmd())
	cmd.AddCommand(newBillingInvoicesCmd())
	cmd.AddCommand(newBillingSubscriptionsCmd())
	return cmd
}

func billingCustomerID(c api.BillingCustomer) string {
	return c.ID
}

func billingCustomerName(c api.BillingCustomer) string {
	if c.BusinessName != "" {
		return c.BusinessName
	}
	name := strings.TrimSpace(c.FirstName + " " + c.LastName)
	return name
}

func billingProductID(p api.BillingProduct) string {
	return p.ID
}

func billingPriceID(p api.BillingPrice) string {
	return p.ID
}

func billingInvoiceID(i api.BillingInvoice) string {
	return i.ID
}

func billingSubscriptionID(s api.BillingSubscription) string {
	return s.ID
}

func parseOptionalBool(value string) (*bool, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("expected true or false, got %q", value)
	}
	return &parsed, nil
}

func newBillingCustomersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customers",
		Short: "Billing customer management",
	}
	cmd.AddCommand(newBillingCustomersListCmd())
	cmd.AddCommand(newBillingCustomersGetCmd())
	cmd.AddCommand(newBillingCustomersCreateCmd())
	cmd.AddCommand(newBillingCustomersUpdateCmd())
	return cmd
}

func newBillingCustomersListCmd() *cobra.Command {
	var merchantCustomerID string
	var from string
	var to string

	cmd := NewListCommand(ListConfig[api.BillingCustomer]{
		Use:          "list",
		Short:        "List billing customers",
		Headers:      []string{"CUSTOMER_ID", "NAME", "EMAIL", "MERCHANT_ID"},
		EmptyMessage: "No billing customers found",
		RowFunc: func(c api.BillingCustomer) []string {
			return []string{billingCustomerID(c), billingCustomerName(c), c.Email, c.MerchantCustomerID}
		},
		IDFunc: func(c api.BillingCustomer) string {
			return billingCustomerID(c)
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingCustomer], error) {
			if err := validateDate(from); err != nil {
				return ListResult[api.BillingCustomer]{}, fmt.Errorf("invalid --from date: %w", err)
			}
			if err := validateDate(to); err != nil {
				return ListResult[api.BillingCustomer]{}, fmt.Errorf("invalid --to date: %w", err)
			}
			if err := validateDateRange(from, to); err != nil {
				return ListResult[api.BillingCustomer]{}, err
			}

			fromRFC3339 := ""
			if from != "" {
				var err error
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return ListResult[api.BillingCustomer]{}, fmt.Errorf("invalid --from date: %w", err)
				}
			}
			toRFC3339 := ""
			if to != "" {
				var err error
				toRFC3339, err = convertDateToRFC3339End(to)
				if err != nil {
					return ListResult[api.BillingCustomer]{}, fmt.Errorf("invalid --to date: %w", err)
				}
			}

			result, err := client.ListBillingCustomers(ctx, api.BillingCustomerListParams{
				MerchantCustomerID: merchantCustomerID,
				FromCreatedAt:      fromRFC3339,
				ToCreatedAt:        toRFC3339,
				PageNum:            0,
				PageSize:           opts.Limit,
			})
			if err != nil {
				return ListResult[api.BillingCustomer]{}, err
			}
			return ListResult[api.BillingCustomer]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&merchantCustomerID, "merchant-customer-id", "", "Filter by merchant customer ID")
	cmd.Flags().StringVar(&from, "from", "", "From created date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To created date (YYYY-MM-DD)")
	return cmd
}

func newBillingCustomersGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.BillingCustomer]{
		Use:   "get <customerId>",
		Short: "Get billing customer",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.BillingCustomer, error) {
			return client.GetBillingCustomer(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, customer *api.BillingCustomer) error {
			rows := []outfmt.KV{
				{Key: "customer_id", Value: billingCustomerID(*customer)},
				{Key: "name", Value: billingCustomerName(*customer)},
				{Key: "email", Value: customer.Email},
				{Key: "merchant_customer_id", Value: customer.MerchantCustomerID},
				{Key: "created_at", Value: customer.CreatedAt},
				{Key: "updated_at", Value: customer.UpdatedAt},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBillingCustomersCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a billing customer",
		Long: `Create a billing customer using a JSON payload.

Examples:
  airwallex billing customers create --data '{"business_name":"Acme Corp","email":"billing@example.com"}'
  airwallex billing customers create --from-file customer.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			customer, err := client.CreateBillingCustomer(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, customer)
			}

			u.Success(fmt.Sprintf("Created billing customer: %s", billingCustomerID(*customer)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingCustomersUpdateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <customerId>",
		Short: "Update a billing customer",
		Long: `Update a billing customer using a JSON payload.

Examples:
  airwallex billing customers update cus_123 --data '{"business_name":"Updated"}'
  airwallex billing customers update cus_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			customer, err := client.UpdateBillingCustomer(cmd.Context(), args[0], payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, customer)
			}

			u.Success(fmt.Sprintf("Updated billing customer: %s", billingCustomerID(*customer)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingProductsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "products",
		Short: "Billing product management",
	}
	cmd.AddCommand(newBillingProductsListCmd())
	cmd.AddCommand(newBillingProductsGetCmd())
	cmd.AddCommand(newBillingProductsCreateCmd())
	cmd.AddCommand(newBillingProductsUpdateCmd())
	return cmd
}

func newBillingProductsListCmd() *cobra.Command {
	var active string

	cmd := NewListCommand(ListConfig[api.BillingProduct]{
		Use:          "list",
		Short:        "List billing products",
		Headers:      []string{"PRODUCT_ID", "NAME", "ACTIVE"},
		EmptyMessage: "No products found",
		RowFunc: func(p api.BillingProduct) []string {
			return []string{billingProductID(p), p.Name, fmt.Sprintf("%t", p.Active)}
		},
		IDFunc: func(p api.BillingProduct) string {
			return billingProductID(p)
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingProduct], error) {
			activeVal, err := parseOptionalBool(active)
			if err != nil {
				return ListResult[api.BillingProduct]{}, fmt.Errorf("invalid --active: %w", err)
			}

			result, err := client.ListBillingProducts(ctx, api.BillingProductListParams{
				Active:   activeVal,
				PageNum:  0,
				PageSize: opts.Limit,
			})
			if err != nil {
				return ListResult[api.BillingProduct]{}, err
			}
			return ListResult[api.BillingProduct]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&active, "active", "", "Filter by active status (true|false)")
	return cmd
}

func newBillingProductsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.BillingProduct]{
		Use:   "get <productId>",
		Short: "Get billing product",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.BillingProduct, error) {
			return client.GetBillingProduct(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, product *api.BillingProduct) error {
			rows := []outfmt.KV{
				{Key: "product_id", Value: billingProductID(*product)},
				{Key: "name", Value: product.Name},
				{Key: "description", Value: product.Description},
				{Key: "unit", Value: product.Unit},
				{Key: "active", Value: fmt.Sprintf("%t", product.Active)},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBillingProductsCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a billing product",
		Long: `Create a billing product using a JSON payload.

Examples:
  airwallex billing products create --data '{"name":"Starter"}'
  airwallex billing products create --from-file product.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			product, err := client.CreateBillingProduct(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, product)
			}

			u.Success(fmt.Sprintf("Created billing product: %s", billingProductID(*product)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingProductsUpdateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <productId>",
		Short: "Update a billing product",
		Long: `Update a billing product using a JSON payload.

Examples:
  airwallex billing products update prod_123 --data '{"name":"Updated"}'
  airwallex billing products update prod_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			product, err := client.UpdateBillingProduct(cmd.Context(), args[0], payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, product)
			}

			u.Success(fmt.Sprintf("Updated billing product: %s", billingProductID(*product)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingPricesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prices",
		Short: "Billing price management",
	}
	cmd.AddCommand(newBillingPricesListCmd())
	cmd.AddCommand(newBillingPricesGetCmd())
	cmd.AddCommand(newBillingPricesCreateCmd())
	cmd.AddCommand(newBillingPricesUpdateCmd())
	return cmd
}

func newBillingPricesListCmd() *cobra.Command {
	var active string
	var currency string
	var productID string
	var recurringPeriod int
	var recurringPeriodUnit string

	cmd := NewListCommand(ListConfig[api.BillingPrice]{
		Use:          "list",
		Short:        "List billing prices",
		Headers:      []string{"PRICE_ID", "PRODUCT_ID", "AMOUNT", "CURRENCY", "ACTIVE"},
		EmptyMessage: "No prices found",
		RowFunc: func(p api.BillingPrice) []string {
			amount := p.UnitAmount
			if amount == 0 {
				amount = p.FlatAmount
			}
			amountText := ""
			if amount != 0 {
				amountText = fmt.Sprintf("%.2f", amount)
			}
			return []string{billingPriceID(p), p.ProductID, amountText, p.Currency, fmt.Sprintf("%t", p.Active)}
		},
		IDFunc: func(p api.BillingPrice) string {
			return billingPriceID(p)
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingPrice], error) {
			activeVal, err := parseOptionalBool(active)
			if err != nil {
				return ListResult[api.BillingPrice]{}, fmt.Errorf("invalid --active: %w", err)
			}
			if err := validateCurrency(currency); err != nil {
				return ListResult[api.BillingPrice]{}, err
			}

			result, err := client.ListBillingPrices(ctx, api.BillingPriceListParams{
				Active:              activeVal,
				Currency:            currency,
				ProductID:           productID,
				RecurringPeriod:     recurringPeriod,
				RecurringPeriodUnit: recurringPeriodUnit,
				PageNum:             0,
				PageSize:            opts.Limit,
			})
			if err != nil {
				return ListResult[api.BillingPrice]{}, err
			}
			return ListResult[api.BillingPrice]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&active, "active", "", "Filter by active status (true|false)")
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency")
	cmd.Flags().StringVar(&productID, "product-id", "", "Filter by product ID")
	cmd.Flags().IntVar(&recurringPeriod, "recurring-period", 0, "Filter by recurring period")
	cmd.Flags().StringVar(&recurringPeriodUnit, "recurring-period-unit", "", "Filter by recurring period unit")
	return cmd
}

func newBillingPricesGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.BillingPrice]{
		Use:   "get <priceId>",
		Short: "Get billing price",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.BillingPrice, error) {
			return client.GetBillingPrice(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, price *api.BillingPrice) error {
			rows := []outfmt.KV{
				{Key: "price_id", Value: billingPriceID(*price)},
				{Key: "product_id", Value: price.ProductID},
				{Key: "active", Value: fmt.Sprintf("%t", price.Active)},
			}
			if price.UnitAmount != 0 || price.FlatAmount != 0 || price.Currency != "" {
				amount := price.UnitAmount
				if amount == 0 {
					amount = price.FlatAmount
				}
				rows = append(rows, outfmt.KV{Key: "amount", Value: fmt.Sprintf("%.2f %s", amount, price.Currency)})
			}
			if price.Recurring != nil {
				rows = append(rows, outfmt.KV{Key: "recurring", Value: fmt.Sprintf("%d %s", price.Recurring.Period, price.Recurring.PeriodUnit)})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBillingPricesCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a billing price",
		Long: `Create a billing price using a JSON payload.

Examples:
  airwallex billing prices create --data '{"product_id":"prod_123","currency":"USD","unit_amount":100}'
  airwallex billing prices create --from-file price.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			price, err := client.CreateBillingPrice(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, price)
			}

			u.Success(fmt.Sprintf("Created billing price: %s", billingPriceID(*price)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingPricesUpdateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <priceId>",
		Short: "Update a billing price",
		Long: `Update a billing price using a JSON payload.

Examples:
  airwallex billing prices update price_123 --data '{"unit_amount":120}'
  airwallex billing prices update price_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			price, err := client.UpdateBillingPrice(cmd.Context(), args[0], payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, price)
			}

			u.Success(fmt.Sprintf("Updated billing price: %s", billingPriceID(*price)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingInvoicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "Billing invoice management",
	}
	cmd.AddCommand(newBillingInvoicesListCmd())
	cmd.AddCommand(newBillingInvoicesGetCmd())
	cmd.AddCommand(newBillingInvoicesCreateCmd())
	cmd.AddCommand(newBillingInvoicesPreviewCmd())
	cmd.AddCommand(newBillingInvoiceItemsCmd())
	return cmd
}

func newBillingInvoicesListCmd() *cobra.Command {
	var customerID string
	var subscriptionID string
	var status string
	var from string
	var to string

	cmd := NewListCommand(ListConfig[api.BillingInvoice]{
		Use:          "list",
		Short:        "List billing invoices",
		Headers:      []string{"INVOICE_ID", "STATUS", "TOTAL", "CURRENCY", "CUSTOMER_ID"},
		EmptyMessage: "No invoices found",
		RowFunc: func(i api.BillingInvoice) []string {
			amount := ""
			if i.TotalAmount != 0 {
				amount = fmt.Sprintf("%.2f", i.TotalAmount)
			}
			return []string{billingInvoiceID(i), i.Status, amount, i.Currency, i.CustomerID}
		},
		IDFunc: func(i api.BillingInvoice) string {
			return billingInvoiceID(i)
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingInvoice], error) {
			if err := validateDate(from); err != nil {
				return ListResult[api.BillingInvoice]{}, fmt.Errorf("invalid --from date: %w", err)
			}
			if err := validateDate(to); err != nil {
				return ListResult[api.BillingInvoice]{}, fmt.Errorf("invalid --to date: %w", err)
			}
			if err := validateDateRange(from, to); err != nil {
				return ListResult[api.BillingInvoice]{}, err
			}

			fromRFC3339 := ""
			if from != "" {
				var err error
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return ListResult[api.BillingInvoice]{}, fmt.Errorf("invalid --from date: %w", err)
				}
			}
			toRFC3339 := ""
			if to != "" {
				var err error
				toRFC3339, err = convertDateToRFC3339End(to)
				if err != nil {
					return ListResult[api.BillingInvoice]{}, fmt.Errorf("invalid --to date: %w", err)
				}
			}

			result, err := client.ListBillingInvoices(ctx, api.BillingInvoiceListParams{
				CustomerID:     customerID,
				SubscriptionID: subscriptionID,
				Status:         status,
				FromCreatedAt:  fromRFC3339,
				ToCreatedAt:    toRFC3339,
				PageNum:        0,
				PageSize:       opts.Limit,
			})
			if err != nil {
				return ListResult[api.BillingInvoice]{}, err
			}
			return ListResult[api.BillingInvoice]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&customerID, "customer-id", "", "Filter by customer ID")
	cmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Filter by subscription ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&from, "from", "", "From created date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To created date (YYYY-MM-DD)")
	return cmd
}

func newBillingInvoicesGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.BillingInvoice]{
		Use:   "get <invoiceId>",
		Short: "Get billing invoice",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.BillingInvoice, error) {
			return client.GetBillingInvoice(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, invoice *api.BillingInvoice) error {
			rows := []outfmt.KV{
				{Key: "invoice_id", Value: billingInvoiceID(*invoice)},
				{Key: "customer_id", Value: invoice.CustomerID},
				{Key: "subscription_id", Value: invoice.SubscriptionID},
				{Key: "status", Value: invoice.Status},
				{Key: "period_start_at", Value: invoice.PeriodStartAt},
				{Key: "period_end_at", Value: invoice.PeriodEndAt},
				{Key: "created_at", Value: invoice.CreatedAt},
				{Key: "updated_at", Value: invoice.UpdatedAt},
				{Key: "paid_at", Value: invoice.PaidAt},
			}
			if invoice.TotalAmount != 0 || invoice.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "total", Value: fmt.Sprintf("%.2f %s", invoice.TotalAmount, invoice.Currency)})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBillingInvoicesCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a billing invoice",
		Long: `Create a billing invoice using a JSON payload.

Examples:
  airwallex billing invoices create --data '{"customer_id":"cus_123","currency":"USD"}'
  airwallex billing invoices create --from-file invoice.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			invoice, err := client.CreateBillingInvoice(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, invoice)
			}

			u.Success(fmt.Sprintf("Created billing invoice: %s", billingInvoiceID(*invoice)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingInvoicesPreviewCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview a billing invoice",
		Long: `Preview a billing invoice using a JSON payload.

Examples:
  airwallex billing invoices preview --data '{"customer_id":"cus_123","currency":"USD"}'
  airwallex billing invoices preview --from-file invoice.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			preview, err := client.PreviewBillingInvoice(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, preview)
			}

			rows := []outfmt.KV{
				{Key: "customer_id", Value: preview.CustomerID},
				{Key: "subscription_id", Value: preview.SubscriptionID},
				{Key: "created_at", Value: preview.CreatedAt},
			}
			if preview.TotalAmount != 0 || preview.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "total", Value: fmt.Sprintf("%.2f %s", preview.TotalAmount, preview.Currency)})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingInvoiceItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "Invoice items",
	}
	cmd.AddCommand(newBillingInvoiceItemsListCmd())
	cmd.AddCommand(newBillingInvoiceItemsGetCmd())
	return cmd
}

func newBillingInvoiceItemsListCmd() *cobra.Command {
	cmd := NewListCommand(ListConfig[api.BillingInvoiceItem]{
		Use:          "list <invoiceId>",
		Short:        "List invoice items",
		Headers:      []string{"ITEM_ID", "INVOICE_ID", "AMOUNT", "CURRENCY", "QTY"},
		EmptyMessage: "No invoice items found",
		RowFunc: func(i api.BillingInvoiceItem) []string {
			amount := ""
			if i.Amount != 0 {
				amount = fmt.Sprintf("%.2f", i.Amount)
			}
			return []string{i.ID, i.InvoiceID, amount, i.Currency, fmt.Sprintf("%.2f", i.Quantity)}
		},
		IDFunc: func(i api.BillingInvoiceItem) string {
			return i.ID
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingInvoiceItem], error) {
			return ListResult[api.BillingInvoiceItem]{}, fmt.Errorf("invoice ID is required")
		},
	}, getClient)

	cmd.Args = cobra.ExactArgs(1)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd.Context())
		if err != nil {
			return err
		}

		limit, _ := cmd.Flags().GetInt("limit")
		if limit <= 0 {
			limit = 20
		}

		result, err := client.ListBillingInvoiceItems(cmd.Context(), args[0], 0, limit)
		if err != nil {
			return err
		}

		f := outfmt.FromContext(cmd.Context())
		if len(result.Items) == 0 {
			if outfmt.IsJSON(cmd.Context()) {
				return f.Output(result)
			}
			f.Empty("No invoice items found")
			return nil
		}

		headers := []string{"ITEM_ID", "INVOICE_ID", "AMOUNT", "CURRENCY", "QTY"}
		rowFn := func(item any) []string {
			it := item.(api.BillingInvoiceItem)
			amount := ""
			if it.Amount != 0 {
				amount = fmt.Sprintf("%.2f", it.Amount)
			}
			return []string{it.ID, it.InvoiceID, amount, it.Currency, fmt.Sprintf("%.2f", it.Quantity)}
		}

		if err := f.OutputList(result.Items, headers, rowFn); err != nil {
			return err
		}
		if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
			_, _ = fmt.Fprintln(os.Stderr, "# More results available")
		}
		return nil
	}

	return cmd
}

func newBillingInvoiceItemsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <invoiceId> <itemId>",
		Short: "Get invoice item",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			item, err := client.GetBillingInvoiceItem(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, item)
			}

			rows := []outfmt.KV{
				{Key: "item_id", Value: item.ID},
				{Key: "invoice_id", Value: item.InvoiceID},
				{Key: "quantity", Value: fmt.Sprintf("%.2f", item.Quantity)},
			}
			if item.Amount != 0 || item.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "amount", Value: fmt.Sprintf("%.2f %s", item.Amount, item.Currency)})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}
}

func newBillingSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Billing subscription management",
	}
	cmd.AddCommand(newBillingSubscriptionsListCmd())
	cmd.AddCommand(newBillingSubscriptionsGetCmd())
	cmd.AddCommand(newBillingSubscriptionsCreateCmd())
	cmd.AddCommand(newBillingSubscriptionsUpdateCmd())
	cmd.AddCommand(newBillingSubscriptionsCancelCmd())
	cmd.AddCommand(newBillingSubscriptionItemsCmd())
	return cmd
}

func newBillingSubscriptionsListCmd() *cobra.Command {
	var customerID string
	var status string
	var recurringPeriod int
	var recurringPeriodUnit string
	var from string
	var to string

	cmd := NewListCommand(ListConfig[api.BillingSubscription]{
		Use:          "list",
		Short:        "List billing subscriptions",
		Headers:      []string{"SUBSCRIPTION_ID", "CUSTOMER_ID", "STATUS", "CURRENT_PERIOD_END"},
		EmptyMessage: "No subscriptions found",
		RowFunc: func(s api.BillingSubscription) []string {
			return []string{billingSubscriptionID(s), s.CustomerID, s.Status, s.CurrentPeriodEndAt}
		},
		IDFunc: func(s api.BillingSubscription) string {
			return billingSubscriptionID(s)
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingSubscription], error) {
			if err := validateDate(from); err != nil {
				return ListResult[api.BillingSubscription]{}, fmt.Errorf("invalid --from date: %w", err)
			}
			if err := validateDate(to); err != nil {
				return ListResult[api.BillingSubscription]{}, fmt.Errorf("invalid --to date: %w", err)
			}
			if err := validateDateRange(from, to); err != nil {
				return ListResult[api.BillingSubscription]{}, err
			}

			fromRFC3339 := ""
			if from != "" {
				var err error
				fromRFC3339, err = convertDateToRFC3339(from)
				if err != nil {
					return ListResult[api.BillingSubscription]{}, fmt.Errorf("invalid --from date: %w", err)
				}
			}
			toRFC3339 := ""
			if to != "" {
				var err error
				toRFC3339, err = convertDateToRFC3339End(to)
				if err != nil {
					return ListResult[api.BillingSubscription]{}, fmt.Errorf("invalid --to date: %w", err)
				}
			}

			result, err := client.ListBillingSubscriptions(ctx, api.BillingSubscriptionListParams{
				CustomerID:          customerID,
				Status:              status,
				RecurringPeriod:     recurringPeriod,
				RecurringPeriodUnit: recurringPeriodUnit,
				FromCreatedAt:       fromRFC3339,
				ToCreatedAt:         toRFC3339,
				PageNum:             0,
				PageSize:            opts.Limit,
			})
			if err != nil {
				return ListResult[api.BillingSubscription]{}, err
			}
			return ListResult[api.BillingSubscription]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	cmd.Flags().StringVar(&customerID, "customer-id", "", "Filter by customer ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().IntVar(&recurringPeriod, "recurring-period", 0, "Filter by recurring period")
	cmd.Flags().StringVar(&recurringPeriodUnit, "recurring-period-unit", "", "Filter by recurring period unit")
	cmd.Flags().StringVar(&from, "from", "", "From created date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "To created date (YYYY-MM-DD)")
	return cmd
}

func newBillingSubscriptionsGetCmd() *cobra.Command {
	return NewGetCommand(GetConfig[*api.BillingSubscription]{
		Use:   "get <subscriptionId>",
		Short: "Get billing subscription",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.BillingSubscription, error) {
			return client.GetBillingSubscription(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, sub *api.BillingSubscription) error {
			rows := []outfmt.KV{
				{Key: "subscription_id", Value: billingSubscriptionID(*sub)},
				{Key: "customer_id", Value: sub.CustomerID},
				{Key: "status", Value: sub.Status},
				{Key: "current_period_start_at", Value: sub.CurrentPeriodStartAt},
				{Key: "current_period_end_at", Value: sub.CurrentPeriodEndAt},
				{Key: "next_billing_at", Value: sub.NextBillingAt},
				{Key: "created_at", Value: sub.CreatedAt},
				{Key: "updated_at", Value: sub.UpdatedAt},
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBillingSubscriptionsCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a billing subscription",
		Long: `Create a billing subscription using a JSON payload.

Examples:
  airwallex billing subscriptions create --data '{"customer_id":"cus_123","price_id":"price_123"}'
  airwallex billing subscriptions create --from-file subscription.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			sub, err := client.CreateBillingSubscription(cmd.Context(), payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, sub)
			}

			u.Success(fmt.Sprintf("Created billing subscription: %s", billingSubscriptionID(*sub)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingSubscriptionsUpdateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "update <subscriptionId>",
		Short: "Update a billing subscription",
		Long: `Update a billing subscription using a JSON payload.

Examples:
  airwallex billing subscriptions update sub_123 --data '{"cancel_at_period_end":true}'
  airwallex billing subscriptions update sub_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			sub, err := client.UpdateBillingSubscription(cmd.Context(), args[0], payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, sub)
			}

			u.Success(fmt.Sprintf("Updated billing subscription: %s", billingSubscriptionID(*sub)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingSubscriptionsCancelCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "cancel <subscriptionId>",
		Short: "Cancel a billing subscription",
		Long: `Cancel a billing subscription.

Examples:
  airwallex billing subscriptions cancel sub_123
  airwallex billing subscriptions cancel sub_123 --data '{"cancel_at_period_end":true}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			payload, err := readOptionalJSONPayload(data, fromFile)
			if err != nil {
				return err
			}

			sub, err := client.CancelBillingSubscription(cmd.Context(), args[0], payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, sub)
			}

			u.Success(fmt.Sprintf("Cancelled billing subscription: %s", billingSubscriptionID(*sub)))
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")
	return cmd
}

func newBillingSubscriptionItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "Subscription items",
	}
	cmd.AddCommand(newBillingSubscriptionItemsListCmd())
	cmd.AddCommand(newBillingSubscriptionItemsGetCmd())
	return cmd
}

func newBillingSubscriptionItemsListCmd() *cobra.Command {
	cmd := NewListCommand(ListConfig[api.BillingSubscriptionItem]{
		Use:          "list <subscriptionId>",
		Short:        "List subscription items",
		Headers:      []string{"ITEM_ID", "SUBSCRIPTION_ID", "PRICE_ID", "QTY"},
		EmptyMessage: "No subscription items found",
		RowFunc: func(i api.BillingSubscriptionItem) []string {
			priceID := ""
			if i.Price != nil {
				priceID = i.Price.ID
			}
			return []string{i.ID, i.SubscriptionID, priceID, fmt.Sprintf("%.2f", i.Quantity)}
		},
		IDFunc: func(i api.BillingSubscriptionItem) string {
			return i.ID
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingSubscriptionItem], error) {
			return ListResult[api.BillingSubscriptionItem]{}, fmt.Errorf("subscription ID is required")
		},
	}, getClient)

	cmd.Args = cobra.ExactArgs(1)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := getClient(cmd.Context())
		if err != nil {
			return err
		}

		limit, _ := cmd.Flags().GetInt("limit")
		if limit <= 0 {
			limit = 20
		}

		result, err := client.ListBillingSubscriptionItems(cmd.Context(), args[0], 0, limit)
		if err != nil {
			return err
		}

		f := outfmt.FromContext(cmd.Context())
		if len(result.Items) == 0 {
			if outfmt.IsJSON(cmd.Context()) {
				return f.Output(result)
			}
			f.Empty("No subscription items found")
			return nil
		}

		headers := []string{"ITEM_ID", "SUBSCRIPTION_ID", "PRICE_ID", "QTY"}
		rowFn := func(item any) []string {
			it := item.(api.BillingSubscriptionItem)
			priceID := ""
			if it.Price != nil {
				priceID = it.Price.ID
			}
			return []string{it.ID, it.SubscriptionID, priceID, fmt.Sprintf("%.2f", it.Quantity)}
		}

		if err := f.OutputList(result.Items, headers, rowFn); err != nil {
			return err
		}
		if !outfmt.IsJSON(cmd.Context()) && result.HasMore {
			_, _ = fmt.Fprintln(os.Stderr, "# More results available")
		}
		return nil
	}

	return cmd
}

func newBillingSubscriptionItemsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <subscriptionId> <itemId>",
		Short: "Get subscription item",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			item, err := client.GetBillingSubscriptionItem(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, item)
			}

			rows := []outfmt.KV{
				{Key: "item_id", Value: item.ID},
				{Key: "subscription_id", Value: item.SubscriptionID},
				{Key: "quantity", Value: fmt.Sprintf("%.2f", item.Quantity)},
			}
			if item.Price != nil {
				rows = append(rows, outfmt.KV{Key: "price_id", Value: item.Price.ID})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}
}
