package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

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
	if c.ID != "" {
		return c.ID
	}
	return c.CustomerID
}

func billingProductID(p api.BillingProduct) string {
	if p.ID != "" {
		return p.ID
	}
	return p.ProductID
}

func billingPriceID(p api.BillingPrice) string {
	if p.ID != "" {
		return p.ID
	}
	return p.PriceID
}

func billingInvoiceID(i api.BillingInvoice) string {
	if i.ID != "" {
		return i.ID
	}
	return i.InvoiceID
}

func billingSubscriptionID(s api.BillingSubscription) string {
	if s.ID != "" {
		return s.ID
	}
	return s.SubscriptionID
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
	return NewListCommand(ListConfig[api.BillingCustomer]{
		Use:          "list",
		Short:        "List billing customers",
		Headers:      []string{"CUSTOMER_ID", "NAME", "EMAIL", "STATUS"},
		EmptyMessage: "No billing customers found",
		RowFunc: func(c api.BillingCustomer) []string {
			return []string{billingCustomerID(c), c.Name, c.Email, c.Status}
		},
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.BillingCustomer], error) {
			result, err := client.ListBillingCustomers(ctx, page, pageSize)
			if err != nil {
				return ListResult[api.BillingCustomer]{}, err
			}
			return ListResult[api.BillingCustomer]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newBillingCustomersGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <customerId>",
		Short: "Get billing customer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			customer, err := client.GetBillingCustomer(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, customer)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "customer_id\t%s\n", billingCustomerID(*customer))
			_, _ = fmt.Fprintf(tw, "name\t%s\n", customer.Name)
			_, _ = fmt.Fprintf(tw, "email\t%s\n", customer.Email)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", customer.Status)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", customer.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}

func newBillingCustomersCreateCmd() *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a billing customer",
		Long: `Create a billing customer using a JSON payload.

Examples:
  airwallex billing customers create --data '{"name":"Acme Corp","email":"billing@example.com"}'
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
  airwallex billing customers update cus_123 --data '{"name":"Updated"}'
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
	return NewListCommand(ListConfig[api.BillingProduct]{
		Use:          "list",
		Short:        "List billing products",
		Headers:      []string{"PRODUCT_ID", "NAME", "STATUS"},
		EmptyMessage: "No products found",
		RowFunc: func(p api.BillingProduct) []string {
			return []string{billingProductID(p), p.Name, p.Status}
		},
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.BillingProduct], error) {
			result, err := client.ListBillingProducts(ctx, page, pageSize)
			if err != nil {
				return ListResult[api.BillingProduct]{}, err
			}
			return ListResult[api.BillingProduct]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newBillingProductsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <productId>",
		Short: "Get billing product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			product, err := client.GetBillingProduct(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, product)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "product_id\t%s\n", billingProductID(*product))
			_, _ = fmt.Fprintf(tw, "name\t%s\n", product.Name)
			_, _ = fmt.Fprintf(tw, "description\t%s\n", product.Description)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", product.Status)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", product.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
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
	return NewListCommand(ListConfig[api.BillingPrice]{
		Use:          "list",
		Short:        "List billing prices",
		Headers:      []string{"PRICE_ID", "PRODUCT_ID", "AMOUNT", "CURRENCY", "STATUS"},
		EmptyMessage: "No prices found",
		RowFunc: func(p api.BillingPrice) []string {
			amount := ""
			if p.UnitAmount != 0 {
				amount = fmt.Sprintf("%.2f", p.UnitAmount)
			}
			return []string{billingPriceID(p), p.ProductID, amount, p.Currency, p.Status}
		},
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.BillingPrice], error) {
			result, err := client.ListBillingPrices(ctx, page, pageSize)
			if err != nil {
				return ListResult[api.BillingPrice]{}, err
			}
			return ListResult[api.BillingPrice]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newBillingPricesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <priceId>",
		Short: "Get billing price",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			price, err := client.GetBillingPrice(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, price)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "price_id\t%s\n", billingPriceID(*price))
			_, _ = fmt.Fprintf(tw, "product_id\t%s\n", price.ProductID)
			if price.UnitAmount != 0 || price.Currency != "" {
				_, _ = fmt.Fprintf(tw, "amount\t%.2f %s\n", price.UnitAmount, price.Currency)
			}
			_, _ = fmt.Fprintf(tw, "status\t%s\n", price.Status)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", price.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
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
	return cmd
}

func newBillingInvoicesListCmd() *cobra.Command {
	return NewListCommand(ListConfig[api.BillingInvoice]{
		Use:          "list",
		Short:        "List billing invoices",
		Headers:      []string{"INVOICE_ID", "NUMBER", "STATUS", "AMOUNT", "CURRENCY"},
		EmptyMessage: "No invoices found",
		RowFunc: func(i api.BillingInvoice) []string {
			amount := ""
			if i.Amount != 0 {
				amount = fmt.Sprintf("%.2f", i.Amount)
			}
			return []string{billingInvoiceID(i), i.InvoiceNumber, i.Status, amount, i.Currency}
		},
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.BillingInvoice], error) {
			result, err := client.ListBillingInvoices(ctx, page, pageSize)
			if err != nil {
				return ListResult[api.BillingInvoice]{}, err
			}
			return ListResult[api.BillingInvoice]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newBillingInvoicesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <invoiceId>",
		Short: "Get billing invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			invoice, err := client.GetBillingInvoice(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, invoice)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "invoice_id\t%s\n", billingInvoiceID(*invoice))
			_, _ = fmt.Fprintf(tw, "invoice_number\t%s\n", invoice.InvoiceNumber)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", invoice.Status)
			if invoice.Amount != 0 || invoice.Currency != "" {
				_, _ = fmt.Fprintf(tw, "amount\t%.2f %s\n", invoice.Amount, invoice.Currency)
			}
			_, _ = fmt.Fprintf(tw, "due_date\t%s\n", invoice.DueDate)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", invoice.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
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

func newBillingSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Billing subscription management",
	}
	cmd.AddCommand(newBillingSubscriptionsListCmd())
	cmd.AddCommand(newBillingSubscriptionsGetCmd())
	cmd.AddCommand(newBillingSubscriptionsCreateCmd())
	return cmd
}

func newBillingSubscriptionsListCmd() *cobra.Command {
	return NewListCommand(ListConfig[api.BillingSubscription]{
		Use:          "list",
		Short:        "List billing subscriptions",
		Headers:      []string{"SUBSCRIPTION_ID", "CUSTOMER_ID", "PRICE_ID", "STATUS"},
		EmptyMessage: "No subscriptions found",
		RowFunc: func(s api.BillingSubscription) []string {
			return []string{billingSubscriptionID(s), s.CustomerID, s.PriceID, s.Status}
		},
		Fetch: func(ctx context.Context, client *api.Client, page, pageSize int) (ListResult[api.BillingSubscription], error) {
			result, err := client.ListBillingSubscriptions(ctx, page, pageSize)
			if err != nil {
				return ListResult[api.BillingSubscription]{}, err
			}
			return ListResult[api.BillingSubscription]{
				Items:   result.Items,
				HasMore: result.HasMore,
			}, nil
		},
	}, getClient)
}

func newBillingSubscriptionsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <subscriptionId>",
		Short: "Get billing subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			sub, err := client.GetBillingSubscription(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, sub)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "subscription_id\t%s\n", billingSubscriptionID(*sub))
			_, _ = fmt.Fprintf(tw, "customer_id\t%s\n", sub.CustomerID)
			_, _ = fmt.Fprintf(tw, "price_id\t%s\n", sub.PriceID)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", sub.Status)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", sub.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
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
