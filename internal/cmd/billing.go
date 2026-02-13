package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func newBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "billing",
		Aliases: []string{"bill", "bi"},
		Short:   "Billing operations",
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
		Use:     "customers",
		Aliases: []string{"cust", "cu", "contacts", "contact"},
		Short:   "Billing customer management",
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
		Aliases:      []string{"ls", "l"},
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
			fromRFC3339, toRFC3339, err := parseDateRangeRFC3339(from, to, "--from", "--to", true)
			if err != nil {
				return ListResult[api.BillingCustomer]{}, err
			}

			result, err := client.ListBillingCustomers(ctx, api.BillingCustomerListParams{
				MerchantCustomerID: merchantCustomerID,
				FromCreatedAt:      fromRFC3339,
				ToCreatedAt:        toRFC3339,
				PageNum:            opts.Page - 1,
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
		Use:     "get <customerId>",
		Aliases: []string{"g"},
		Short:   "Get billing customer",
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
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingCustomer]{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a billing customer",
		Long: `Create a billing customer using a JSON payload.

Examples:
  airwallex billing customers create --data '{"business_name":"Acme Corp","email":"billing@example.com"}'
  airwallex billing customers create --from-file customer.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingCustomer, error) {
			return client.CreateBillingCustomer(ctx, payload)
		},
		SuccessMessage: func(customer *api.BillingCustomer) string {
			return fmt.Sprintf("Created billing customer: %s", billingCustomerID(*customer))
		},
	}, getClient)
}

func newBillingCustomersUpdateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingCustomer]{
		Use:     "update <customerId>",
		Aliases: []string{"up", "u"},
		Short:   "Update a billing customer",
		Long: `Update a billing customer using a JSON payload.

Examples:
  airwallex billing customers update cus_123 --data '{"business_name":"Updated"}'
  airwallex billing customers update cus_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingCustomer, error) {
			return client.UpdateBillingCustomer(ctx, NormalizeIDArg(args[0]), payload)
		},
		SuccessMessage: func(customer *api.BillingCustomer) string {
			return fmt.Sprintf("Updated billing customer: %s", billingCustomerID(*customer))
		},
	}, getClient)
}

func newBillingProductsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "products",
		Aliases: []string{"prod", "pr"},
		Short:   "Billing product management",
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
		Aliases:      []string{"ls", "l"},
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
				PageNum:  opts.Page - 1,
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
		Use:     "get <productId>",
		Aliases: []string{"g"},
		Short:   "Get billing product",
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
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingProduct]{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a billing product",
		Long: `Create a billing product using a JSON payload.

Examples:
  airwallex billing products create --data '{"name":"Starter"}'
  airwallex billing products create --from-file product.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingProduct, error) {
			return client.CreateBillingProduct(ctx, payload)
		},
		SuccessMessage: func(product *api.BillingProduct) string {
			return fmt.Sprintf("Created billing product: %s", billingProductID(*product))
		},
	}, getClient)
}

func newBillingProductsUpdateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingProduct]{
		Use:     "update <productId>",
		Aliases: []string{"up", "u"},
		Short:   "Update a billing product",
		Long: `Update a billing product using a JSON payload.

Examples:
  airwallex billing products update prod_123 --data '{"name":"Updated"}'
  airwallex billing products update prod_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingProduct, error) {
			return client.UpdateBillingProduct(ctx, NormalizeIDArg(args[0]), payload)
		},
		SuccessMessage: func(product *api.BillingProduct) string {
			return fmt.Sprintf("Updated billing product: %s", billingProductID(*product))
		},
	}, getClient)
}

func newBillingPricesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prices",
		Aliases: []string{"price", "pc"},
		Short:   "Billing price management",
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
		Aliases:      []string{"ls", "l"},
		Short:        "List billing prices",
		Headers:      []string{"PRICE_ID", "PRODUCT_ID", "AMOUNT", "CURRENCY", "ACTIVE"},
		EmptyMessage: "No prices found",
		RowFunc: func(p api.BillingPrice) []string {
			amount := p.UnitAmount
			if outfmt.MoneyFloat64(amount) == 0 {
				amount = p.FlatAmount
			}
			amountText := ""
			if outfmt.MoneyFloat64(amount) != 0 {
				amountText = outfmt.FormatMoney(amount)
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
				PageNum:             opts.Page - 1,
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
		Use:     "get <priceId>",
		Aliases: []string{"g"},
		Short:   "Get billing price",
		Fetch: func(ctx context.Context, client *api.Client, id string) (*api.BillingPrice, error) {
			return client.GetBillingPrice(ctx, id)
		},
		TextOutput: func(cmd *cobra.Command, price *api.BillingPrice) error {
			rows := []outfmt.KV{
				{Key: "price_id", Value: billingPriceID(*price)},
				{Key: "product_id", Value: price.ProductID},
				{Key: "active", Value: fmt.Sprintf("%t", price.Active)},
			}
			if outfmt.MoneyFloat64(price.UnitAmount) != 0 || outfmt.MoneyFloat64(price.FlatAmount) != 0 || price.Currency != "" {
				amount := price.UnitAmount
				if outfmt.MoneyFloat64(amount) == 0 {
					amount = price.FlatAmount
				}
				rows = append(rows, outfmt.KV{Key: "amount", Value: outfmt.FormatMoney(amount) + " " + price.Currency})
			}
			if price.Recurring != nil {
				rows = append(rows, outfmt.KV{Key: "recurring", Value: fmt.Sprintf("%d %s", price.Recurring.Period, price.Recurring.PeriodUnit)})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBillingPricesCreateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingPrice]{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a billing price",
		Long: `Create a billing price using a JSON payload.

Examples:
  airwallex billing prices create --data '{"product_id":"prod_123","currency":"USD","unit_amount":100}'
  airwallex billing prices create --from-file price.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingPrice, error) {
			return client.CreateBillingPrice(ctx, payload)
		},
		SuccessMessage: func(price *api.BillingPrice) string {
			return fmt.Sprintf("Created billing price: %s", billingPriceID(*price))
		},
	}, getClient)
}

func newBillingPricesUpdateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingPrice]{
		Use:     "update <priceId>",
		Aliases: []string{"up", "u"},
		Short:   "Update a billing price",
		Long: `Update a billing price using a JSON payload.

Examples:
  airwallex billing prices update price_123 --data '{"unit_amount":120}'
  airwallex billing prices update price_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingPrice, error) {
			return client.UpdateBillingPrice(ctx, NormalizeIDArg(args[0]), payload)
		},
		SuccessMessage: func(price *api.BillingPrice) string {
			return fmt.Sprintf("Updated billing price: %s", billingPriceID(*price))
		},
	}, getClient)
}

func newBillingInvoicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "invoices",
		Aliases: []string{"inv"},
		Short:   "Billing invoice management",
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
		Aliases:      []string{"ls", "l"},
		Short:        "List billing invoices",
		Headers:      []string{"INVOICE_ID", "STATUS", "TOTAL", "CURRENCY", "CUSTOMER_ID"},
		EmptyMessage: "No invoices found",
		RowFunc: func(i api.BillingInvoice) []string {
			amount := ""
			if outfmt.MoneyFloat64(i.TotalAmount) != 0 {
				amount = outfmt.FormatMoney(i.TotalAmount)
			}
			return []string{billingInvoiceID(i), i.Status, amount, i.Currency, i.CustomerID}
		},
		IDFunc: func(i api.BillingInvoice) string {
			return billingInvoiceID(i)
		},
		Fetch: func(ctx context.Context, client *api.Client, opts ListOptions) (ListResult[api.BillingInvoice], error) {
			fromRFC3339, toRFC3339, err := parseDateRangeRFC3339(from, to, "--from", "--to", true)
			if err != nil {
				return ListResult[api.BillingInvoice]{}, err
			}

			result, err := client.ListBillingInvoices(ctx, api.BillingInvoiceListParams{
				CustomerID:     customerID,
				SubscriptionID: subscriptionID,
				Status:         status,
				FromCreatedAt:  fromRFC3339,
				ToCreatedAt:    toRFC3339,
				PageNum:        opts.Page - 1,
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
		Use:     "get <invoiceId>",
		Aliases: []string{"g"},
		Short:   "Get billing invoice",
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
			if outfmt.MoneyFloat64(invoice.TotalAmount) != 0 || invoice.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "total", Value: outfmt.FormatMoney(invoice.TotalAmount) + " " + invoice.Currency})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}, getClient)
}

func newBillingInvoicesCreateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingInvoice]{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a billing invoice",
		Long: `Create a billing invoice using a JSON payload.

Examples:
  airwallex billing invoices create --data '{"customer_id":"cus_123","currency":"USD"}'
  airwallex billing invoices create --from-file invoice.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingInvoice, error) {
			return client.CreateBillingInvoice(ctx, payload)
		},
		SuccessMessage: func(invoice *api.BillingInvoice) string {
			return fmt.Sprintf("Created billing invoice: %s", billingInvoiceID(*invoice))
		},
	}, getClient)
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
				return writeJSONOutput(cmd, preview)
			}

			rows := []outfmt.KV{
				{Key: "customer_id", Value: preview.CustomerID},
				{Key: "subscription_id", Value: preview.SubscriptionID},
				{Key: "created_at", Value: preview.CreatedAt},
			}
			if outfmt.MoneyFloat64(preview.TotalAmount) != 0 || preview.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "total", Value: outfmt.FormatMoney(preview.TotalAmount) + " " + preview.Currency})
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
		Use:     "items",
		Aliases: []string{"item", "it"},
		Short:   "Invoice items",
	}
	cmd.AddCommand(newBillingInvoiceItemsListCmd())
	cmd.AddCommand(newBillingInvoiceItemsGetCmd())
	return cmd
}

func newBillingInvoiceItemsListCmd() *cobra.Command {
	cmd := NewListCommand(ListConfig[api.BillingInvoiceItem]{
		Use:          "list <invoiceId>",
		Aliases:      []string{"ls", "l"},
		Short:        "List invoice items",
		Headers:      []string{"ITEM_ID", "INVOICE_ID", "AMOUNT", "CURRENCY", "QTY"},
		EmptyMessage: "No invoice items found",
		Args:         cobra.ExactArgs(1),
		RowFunc: func(i api.BillingInvoiceItem) []string {
			amount := ""
			if outfmt.MoneyFloat64(i.Amount) != 0 {
				amount = outfmt.FormatMoney(i.Amount)
			}
			return []string{i.ID, i.InvoiceID, amount, i.Currency, outfmt.FormatMoney(i.Quantity)}
		},
		IDFunc: func(i api.BillingInvoiceItem) string {
			return i.ID
		},
		MoreHint: "# More results available",
		FetchWithArgs: func(ctx context.Context, client *api.Client, opts ListOptions, args []string) (ListResult[api.BillingInvoiceItem], error) {
			invoiceID := NormalizeIDArg(args[0])
			result, err := client.ListBillingInvoiceItems(ctx, invoiceID, opts.Page-1, opts.Limit)
			if err != nil {
				return ListResult[api.BillingInvoiceItem]{}, err
			}
			return ListResult[api.BillingInvoiceItem]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	return cmd
}

func newBillingInvoiceItemsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <invoiceId> <itemId>",
		Aliases: []string{"g"},
		Short:   "Get invoice item",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			invoiceID := NormalizeIDArg(args[0])
			itemID := NormalizeIDArg(args[1])
			item, err := client.GetBillingInvoiceItem(cmd.Context(), invoiceID, itemID)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return writeJSONOutput(cmd, item)
			}

			rows := []outfmt.KV{
				{Key: "item_id", Value: item.ID},
				{Key: "invoice_id", Value: item.InvoiceID},
				{Key: "quantity", Value: outfmt.FormatMoney(item.Quantity)},
			}
			if outfmt.MoneyFloat64(item.Amount) != 0 || item.Currency != "" {
				rows = append(rows, outfmt.KV{Key: "amount", Value: outfmt.FormatMoney(item.Amount) + " " + item.Currency})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}
}

func newBillingSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscriptions",
		Aliases: []string{"sub", "su"},
		Short:   "Billing subscription management",
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
		Aliases:      []string{"ls", "l"},
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
			fromRFC3339, toRFC3339, err := parseDateRangeRFC3339(from, to, "--from", "--to", true)
			if err != nil {
				return ListResult[api.BillingSubscription]{}, err
			}

			result, err := client.ListBillingSubscriptions(ctx, api.BillingSubscriptionListParams{
				CustomerID:          customerID,
				Status:              status,
				RecurringPeriod:     recurringPeriod,
				RecurringPeriodUnit: recurringPeriodUnit,
				FromCreatedAt:       fromRFC3339,
				ToCreatedAt:         toRFC3339,
				PageNum:             opts.Page - 1,
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
		Use:     "get <subscriptionId>",
		Aliases: []string{"g"},
		Short:   "Get billing subscription",
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
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingSubscription]{
		Use:     "create",
		Aliases: []string{"cr"},
		Short:   "Create a billing subscription",
		Long: `Create a billing subscription using a JSON payload.

Examples:
  airwallex billing subscriptions create --data '{"customer_id":"cus_123","price_id":"price_123"}'
  airwallex billing subscriptions create --from-file subscription.json`,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingSubscription, error) {
			return client.CreateBillingSubscription(ctx, payload)
		},
		SuccessMessage: func(sub *api.BillingSubscription) string {
			return fmt.Sprintf("Created billing subscription: %s", billingSubscriptionID(*sub))
		},
	}, getClient)
}

func newBillingSubscriptionsUpdateCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingSubscription]{
		Use:     "update <subscriptionId>",
		Aliases: []string{"up", "u"},
		Short:   "Update a billing subscription",
		Long: `Update a billing subscription using a JSON payload.

Examples:
  airwallex billing subscriptions update sub_123 --data '{"cancel_at_period_end":true}'
  airwallex billing subscriptions update sub_123 --from-file update.json`,
		Args: cobra.ExactArgs(1),
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingSubscription, error) {
			return client.UpdateBillingSubscription(ctx, NormalizeIDArg(args[0]), payload)
		},
		SuccessMessage: func(sub *api.BillingSubscription) string {
			return fmt.Sprintf("Updated billing subscription: %s", billingSubscriptionID(*sub))
		},
	}, getClient)
}

func newBillingSubscriptionsCancelCmd() *cobra.Command {
	return NewPayloadCommand(PayloadCommandConfig[*api.BillingSubscription]{
		Use:     "cancel <subscriptionId>",
		Aliases: []string{"x"},
		Short:   "Cancel a billing subscription",
		Long: `Cancel a billing subscription.

Examples:
  airwallex billing subscriptions cancel sub_123
  airwallex billing subscriptions cancel sub_123 --data '{"cancel_at_period_end":true}'`,
		Args:        cobra.ExactArgs(1),
		ReadPayload: readOptionalJSONPayload,
		Run: func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (*api.BillingSubscription, error) {
			return client.CancelBillingSubscription(ctx, NormalizeIDArg(args[0]), payload)
		},
		SuccessMessage: func(sub *api.BillingSubscription) string {
			return fmt.Sprintf("Cancelled billing subscription: %s", billingSubscriptionID(*sub))
		},
	}, getClient)
}

func newBillingSubscriptionItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "items",
		Aliases: []string{"item", "it"},
		Short:   "Subscription items",
	}
	cmd.AddCommand(newBillingSubscriptionItemsListCmd())
	cmd.AddCommand(newBillingSubscriptionItemsGetCmd())
	return cmd
}

func newBillingSubscriptionItemsListCmd() *cobra.Command {
	cmd := NewListCommand(ListConfig[api.BillingSubscriptionItem]{
		Use:          "list <subscriptionId>",
		Aliases:      []string{"ls", "l"},
		Short:        "List subscription items",
		Headers:      []string{"ITEM_ID", "SUBSCRIPTION_ID", "PRICE_ID", "QTY"},
		EmptyMessage: "No subscription items found",
		Args:         cobra.ExactArgs(1),
		RowFunc: func(i api.BillingSubscriptionItem) []string {
			priceID := ""
			if i.Price != nil {
				priceID = i.Price.ID
			}
			return []string{i.ID, i.SubscriptionID, priceID, outfmt.FormatMoney(i.Quantity)}
		},
		IDFunc: func(i api.BillingSubscriptionItem) string {
			return i.ID
		},
		MoreHint: "# More results available",
		FetchWithArgs: func(ctx context.Context, client *api.Client, opts ListOptions, args []string) (ListResult[api.BillingSubscriptionItem], error) {
			subscriptionID := NormalizeIDArg(args[0])
			result, err := client.ListBillingSubscriptionItems(ctx, subscriptionID, opts.Page-1, opts.Limit)
			if err != nil {
				return ListResult[api.BillingSubscriptionItem]{}, err
			}
			return ListResult[api.BillingSubscriptionItem]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)

	return cmd
}

func newBillingSubscriptionItemsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <subscriptionId> <itemId>",
		Aliases: []string{"g"},
		Short:   "Get subscription item",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			subscriptionID := NormalizeIDArg(args[0])
			itemID := NormalizeIDArg(args[1])
			item, err := client.GetBillingSubscriptionItem(cmd.Context(), subscriptionID, itemID)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return writeJSONOutput(cmd, item)
			}

			rows := []outfmt.KV{
				{Key: "item_id", Value: item.ID},
				{Key: "subscription_id", Value: item.SubscriptionID},
				{Key: "quantity", Value: outfmt.FormatMoney(item.Quantity)},
			}
			if item.Price != nil {
				rows = append(rows, outfmt.KV{Key: "price_id", Value: item.Price.ID})
			}
			return outfmt.WriteKV(cmd.OutOrStdout(), rows)
		},
	}
}
