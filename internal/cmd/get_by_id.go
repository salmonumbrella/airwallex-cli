package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

// newGetByIDCmd implements the "desire path" of "just give it the ID".
//
// It intentionally prefers JSON output (unless the user provided --template),
// because we cannot provide consistent human tables across many resource types.
func newGetByIDCmd(getClient func(context.Context) (*api.Client, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a resource by ID (auto-detect type)",
		Long: `Fetch a resource using only its ID.

This command auto-detects the resource type based on common Airwallex ID prefixes
(e.g. tfr_, ben_, pl_, inv_). It is designed for agents and scripting: if you
don't pass --template, it will emit JSON even when --output is text.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// If this is being used as an agent primitive, JSON is the most stable output.
			// Respect --template if present.
			if outfmt.GetTemplate(ctx) == "" && !outfmt.IsJSON(ctx) {
				ctx = outfmt.WithFormat(ctx, "json")
				cmd.SetContext(ctx)
			}

			id := NormalizeIDArg(args[0])

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			item, canonicalCmd, err := fetchByID(ctx, client, id)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(ctx)
			// Template output should apply to the resource itself, not the wrapper.
			if outfmt.GetTemplate(ctx) != "" {
				return f.Output(item)
			}

			links := map[string]string{"self": canonicalCmd}
			return f.OutputAnnotated(item, links)
		},
	}

	return cmd
}

func fetchByID(ctx context.Context, client *api.Client, id string) (any, string, error) {
	// Composite IDs: inv_xxx:item_yyy or sub_xxx:si_yyy
	if a, b, ok := strings.Cut(id, ":"); ok {
		a = NormalizeIDArg(a)
		b = NormalizeIDArg(b)
		switch {
		case strings.HasPrefix(a, "inv_"):
			item, err := client.GetBillingInvoiceItem(ctx, a, b)
			return item, fmt.Sprintf("airwallex billing invoices items get %s %s", a, b), err
		case strings.HasPrefix(a, "sub_"):
			item, err := client.GetBillingSubscriptionItem(ctx, a, b)
			return item, fmt.Sprintf("airwallex billing subscriptions items get %s %s", a, b), err
		default:
			return nil, "", fmt.Errorf("unknown composite id %q (expected inv_*:item_* or sub_*:si_*)", id)
		}
	}

	switch {
	// Transfers & beneficiaries.
	case strings.HasPrefix(id, "tfr_"):
		item, err := client.GetTransfer(ctx, id)
		return item, fmt.Sprintf("airwallex transfers get %s", id), err
	case strings.HasPrefix(id, "ben_"):
		item, err := client.GetBeneficiary(ctx, id)
		return item, fmt.Sprintf("airwallex beneficiaries get %s", id), err

	// Webhooks.
	case strings.HasPrefix(id, "wh_"):
		item, err := client.GetWebhook(ctx, id)
		return item, fmt.Sprintf("airwallex webhooks get %s", id), err

	// Linked accounts & deposits.
	case strings.HasPrefix(id, "la_"):
		item, err := client.GetLinkedAccount(ctx, id)
		return item, fmt.Sprintf("airwallex linked-accounts get %s", id), err
	case strings.HasPrefix(id, "dep_"):
		item, err := client.GetDeposit(ctx, id)
		return item, fmt.Sprintf("airwallex deposits get %s", id), err

	// Payment links.
	case strings.HasPrefix(id, "pl_"):
		item, err := client.GetPaymentLink(ctx, id)
		return item, fmt.Sprintf("airwallex payment-links get %s", id), err

	// FX.
	case strings.HasPrefix(id, "quote_"):
		item, err := client.GetQuote(ctx, id)
		return item, fmt.Sprintf("airwallex fx quotes get %s", id), err
	case strings.HasPrefix(id, "conv_"):
		item, err := client.GetConversion(ctx, id)
		return item, fmt.Sprintf("airwallex fx conversions get %s", id), err

	// Issuing.
	case strings.HasPrefix(id, "card_"):
		item, err := client.GetCard(ctx, id)
		return item, fmt.Sprintf("airwallex cards get %s", id), err
	case strings.HasPrefix(id, "card_holder_") || strings.HasPrefix(id, "cardholder_"):
		item, err := client.GetCardholder(ctx, id)
		return item, fmt.Sprintf("airwallex cardholders get %s", id), err
	case strings.HasPrefix(id, "txn_"):
		item, err := client.GetTransaction(ctx, id)
		return item, fmt.Sprintf("airwallex transactions get %s", id), err
	case strings.HasPrefix(id, "disp_"):
		item, err := client.GetTransactionDispute(ctx, id)
		return item, fmt.Sprintf("airwallex disputes get %s", id), err

	// Billing.
	case strings.HasPrefix(id, "prod_"):
		item, err := client.GetBillingProduct(ctx, id)
		return item, fmt.Sprintf("airwallex billing products get %s", id), err
	case strings.HasPrefix(id, "price_"):
		item, err := client.GetBillingPrice(ctx, id)
		return item, fmt.Sprintf("airwallex billing prices get %s", id), err
	case strings.HasPrefix(id, "inv_"):
		item, err := client.GetBillingInvoice(ctx, id)
		return item, fmt.Sprintf("airwallex billing invoices get %s", id), err
	case strings.HasPrefix(id, "sub_"):
		item, err := client.GetBillingSubscription(ctx, id)
		return item, fmt.Sprintf("airwallex billing subscriptions get %s", id), err
	case strings.HasPrefix(id, "cus_") || strings.HasPrefix(id, "cust_"):
		item, err := client.GetBillingCustomer(ctx, id)
		return item, fmt.Sprintf("airwallex billing customers get %s", id), err
	default:
		return nil, "", fmt.Errorf("unknown id %q (supported prefixes: tfr_, ben_, wh_, la_, dep_, pl_, quote_, conv_, card_, card_holder_, txn_, disp_, prod_, price_, inv_, sub_, cus_/cust_)", id)
	}
}
