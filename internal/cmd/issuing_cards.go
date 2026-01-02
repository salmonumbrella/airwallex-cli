package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

func newCardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cards",
		Short: "Card management",
	}
	cmd.AddCommand(newCardsListCmd())
	cmd.AddCommand(newCardsGetCmd())
	cmd.AddCommand(newCardsCreateCmd())
	cmd.AddCommand(newCardsUpdateCmd())
	cmd.AddCommand(newCardsActivateCmd())
	cmd.AddCommand(newCardsDetailsCmd())
	cmd.AddCommand(newCardsLimitsCmd())
	return cmd
}

func newCardsListCmd() *cobra.Command {
	var status string
	var cardholderID string
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cards",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			cards, err := client.ListCards(cmd.Context(), status, cardholderID, page, pageSize)
			if err != nil {
				return err
			}

			f := outfmt.FromContext(cmd.Context())

			if len(cards.Items) == 0 {
				if outfmt.IsJSON(cmd.Context()) {
					return f.Output(cards)
				}
				f.Empty("No cards found")
				return nil
			}

			headers := []string{"CARD_ID", "STATUS", "NICKNAME", "LAST4", "FORM_FACTOR", "CARDHOLDER"}
			rowFn := func(item any) []string {
				c := item.(api.Card)
				last4 := ""
				if len(c.CardNumber) >= 4 {
					last4 = c.CardNumber[len(c.CardNumber)-4:]
				}
				return []string{c.CardID, c.CardStatus, c.NickName, last4, c.FormFactor, c.CardholderID}
			}

			if err := f.OutputList(cards.Items, headers, rowFn); err != nil {
				return err
			}

			if !outfmt.IsJSON(cmd.Context()) && cards.HasMore {
				_, _ = fmt.Fprintln(os.Stderr, "# More results available")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (ACTIVE, INACTIVE, CLOSED)")
	cmd.Flags().StringVar(&cardholderID, "cardholder-id", "", "Filter by cardholder")
	cmd.Flags().IntVar(&page, "page", 0, "Page number (0 = first page)")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "API page size")
	return cmd
}

func newCardsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <cardId>",
		Short: "Get card details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			card, err := client.GetCard(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, card)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "card_id\t%s\n", card.CardID)
			_, _ = fmt.Fprintf(tw, "status\t%s\n", card.CardStatus)
			_, _ = fmt.Fprintf(tw, "nickname\t%s\n", card.NickName)
			_, _ = fmt.Fprintf(tw, "card_number\t%s\n", card.CardNumber)
			_, _ = fmt.Fprintf(tw, "brand\t%s\n", card.Brand)
			_, _ = fmt.Fprintf(tw, "form_factor\t%s\n", card.FormFactor)
			_, _ = fmt.Fprintf(tw, "cardholder_id\t%s\n", card.CardholderID)
			_, _ = fmt.Fprintf(tw, "created_at\t%s\n", card.CreatedAt)
			_ = tw.Flush()
			return nil
		},
	}
}

func newCardsCreateCmd() *cobra.Command {
	var cardholderID string
	var formFactor string
	var currency string
	var limitAmount float64
	var limitInterval string
	var limitCurrency string
	var createdBy string
	var programPurpose string
	var programType string
	var companyCard bool
	var additionalCardholders []string

	cmd := &cobra.Command{
		Use:   "create <nickname>",
		Short: "Create a new card",
		Long: `Create a new card with spending limits.

IMPORTANT: The --limit flag is required. Airwallex requires all cards to have
a spending limit configured.

Card types:
  - Employee card (default): Personalized for a single cardholder
  - Company card (--company): Shared card, supports up to 3 additional cardholders

Examples:
  # Create an employee card with a $100/month limit
  airwallex issuing cards create "DoorDash" --cardholder-id <id> --limit 100 --limit-interval MONTHLY

  # Create a company card shared by multiple employees (comma-separated IDs)
  airwallex issuing cards create "Office Supplies" --cardholder-id chld_123 --company \
    --additional-cardholders chld_456,chld_789 --limit 500

  # Create a card with a $500 all-time limit
  airwallex issuing cards create "Travel" --cardholder-id <id> --limit 500 --limit-interval ALL_TIME

Limit intervals: PER_TRANSACTION, DAILY, WEEKLY, MONTHLY, QUARTERLY, YEARLY, ALL_TIME
Program purposes: COMMERCIAL, CONSUMER
Program types: PREPAID, DEBIT, CREDIT, DEFERRED_DEBIT`,
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			func(cmd *cobra.Command, args []string) error {
				additionalCardholders, _ := cmd.Flags().GetStringSlice("additional-cardholders")
				companyCard, _ := cmd.Flags().GetBool("company")

				if len(additionalCardholders) > 0 && !companyCard {
					return fmt.Errorf("--additional-cardholders requires --company flag")
				}
				if len(additionalCardholders) > 3 {
					return fmt.Errorf("maximum 3 additional cardholders allowed")
				}
				return nil
			},
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			nickname := args[0]

			req := map[string]interface{}{
				"cardholder_id":   cardholderID,
				"form_factor":     formFactor,
				"nick_name":       nickname,
				"is_personalized": !companyCard,
				"created_by":      createdBy,
				"request_id":      fmt.Sprintf("cli-%d", os.Getpid()),
			}

			// Add additional cardholders for company cards
			if companyCard && len(additionalCardholders) > 0 {
				req["additional_cardholder_ids"] = additionalCardholders
			}

			// Only add program if explicitly set
			program := map[string]interface{}{
				"purpose": programPurpose,
			}
			if cmd.Flags().Changed("program-type") {
				program["type"] = programType
			}
			req["program"] = program

			if currency != "" {
				req["primary_currency"] = currency
			}

			// Build authorization controls with transaction limits
			authControls := map[string]interface{}{
				"allowed_transaction_count": "MULTIPLE",
			}

			if limitAmount > 0 {
				lc := limitCurrency
				if lc == "" {
					lc = "USD"
				}
				interval := limitInterval
				if interval == "" {
					interval = "MONTHLY"
				}

				authControls["transaction_limits"] = map[string]interface{}{
					"currency": lc,
					"limits": []map[string]interface{}{
						{
							"amount":   limitAmount,
							"interval": interval,
						},
					},
				}
			}

			req["authorization_controls"] = authControls

			card, err := client.CreateCard(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, card)
			}

			limitInfo := ""
			if limitAmount > 0 {
				lc := limitCurrency
				if lc == "" {
					lc = "USD"
				}
				interval := limitInterval
				if interval == "" {
					interval = "MONTHLY"
				}
				limitInfo = fmt.Sprintf(" with %s %.2f %s limit", interval, limitAmount, lc)
			}

			cardType := "employee"
			if companyCard {
				cardType = "company"
			}
			u.Success(fmt.Sprintf("Created %s card \"%s\"%s: %s", cardType, nickname, limitInfo, card.CardID))

			// For company cards, fetch and display card details (PAN, CVV, expiry)
			if companyCard {
				details, err := client.GetCardDetails(cmd.Context(), card.CardID)
				if err != nil {
					u.Error(fmt.Sprintf("Card created but could not fetch details: %v", err))
					u.Info("Use 'airwallex issuing cards details " + card.CardID + "' to retrieve them later")
				} else {
					defer details.Zeroize()
					fmt.Println()
					tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
					_, _ = fmt.Fprintln(tw, "CARD DETAILS (Company Card)")
					_, _ = fmt.Fprintf(tw, "card_number\t%s\n", details.CardNumber)
					_, _ = fmt.Fprintf(tw, "cvv\t%s\n", details.Cvv)
					_, _ = fmt.Fprintf(tw, "expiry\t%02d/%d\n", details.ExpiryMonth, details.ExpiryYear)
					_ = tw.Flush()
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&cardholderID, "cardholder-id", "", "Cardholder ID (required)")
	cmd.Flags().StringVar(&formFactor, "form-factor", "VIRTUAL", "VIRTUAL or PHYSICAL")
	cmd.Flags().StringVar(&currency, "currency", "", "Primary currency")
	cmd.Flags().Float64Var(&limitAmount, "limit", 0, "Spending limit amount (required)")
	cmd.Flags().StringVar(&limitInterval, "limit-interval", "MONTHLY", "Limit interval: PER_TRANSACTION, DAILY, WEEKLY, MONTHLY, QUARTERLY, YEARLY, ALL_TIME")
	cmd.Flags().StringVar(&limitCurrency, "limit-currency", "USD", "Limit currency (default: USD)")
	cmd.Flags().StringVar(&createdBy, "created-by", "Airwallex CLI", "Name of person creating the card")
	cmd.Flags().StringVar(&programPurpose, "program-purpose", "COMMERCIAL", "Program purpose: COMMERCIAL or CONSUMER")
	cmd.Flags().StringVar(&programType, "program-type", "PREPAID", "Program type: PREPAID, DEBIT, CREDIT, DEFERRED_DEBIT")
	cmd.Flags().BoolVar(&companyCard, "company", false, "Create a company card (shared, not personalized)")
	cmd.Flags().StringSliceVar(&additionalCardholders, "additional-cardholders", nil, "Additional cardholder IDs for company cards (max 3)")
	mustMarkRequired(cmd, "cardholder-id")
	mustMarkRequired(cmd, "limit")
	return cmd
}

func newCardsUpdateCmd() *cobra.Command {
	var nickname string
	var status string

	cmd := &cobra.Command{
		Use:   "update <cardId>",
		Short: "Update card (nickname, status)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			update := make(map[string]interface{})
			if cmd.Flags().Changed("nickname") {
				update["nick_name"] = nickname
			}
			if cmd.Flags().Changed("status") {
				update["card_status"] = status
			}

			if len(update) == 0 {
				return fmt.Errorf("no updates specified")
			}

			card, err := client.UpdateCard(cmd.Context(), args[0], update)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, card)
			}

			u.Success(fmt.Sprintf("Updated card: %s", card.CardID))
			return nil
		},
	}

	cmd.Flags().StringVar(&nickname, "nickname", "", "Card nickname")
	cmd.Flags().StringVar(&status, "status", "", "Card status (ACTIVE, INACTIVE, CLOSED)")
	return cmd
}

func newCardsActivateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "activate <cardId>",
		Short: "Activate a physical card",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			card, err := client.ActivateCard(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, card)
			}

			u.Success(fmt.Sprintf("Activated card: %s", card.CardID))
			return nil
		},
	}
}

func newCardsDetailsCmd() *cobra.Command {
	var showPAN bool

	cmd := &cobra.Command{
		Use:   "details <cardId>",
		Short: "Get sensitive card details (PAN, CVV, expiry)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			details, err := client.GetCardDetails(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			defer details.Zeroize()

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, details)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintf(tw, "card_id\t%s\n", details.CardID)
			if showPAN {
				_, _ = fmt.Fprintf(tw, "card_number\t%s\n", details.CardNumber)
			} else {
				_, _ = fmt.Fprintf(tw, "card_number\t%s\n", details.MaskedPAN())
			}
			_, _ = fmt.Fprintf(tw, "cvv\t%s\n", details.Cvv)
			_, _ = fmt.Fprintf(tw, "expiry\t%02d/%d\n", details.ExpiryMonth, details.ExpiryYear)
			_ = tw.Flush()
			return nil
		},
	}

	cmd.Flags().BoolVar(&showPAN, "show-pan", false, "Show full card number (PCI-sensitive)")
	return cmd
}

func newCardsLimitsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "limits <cardId>",
		Short: "Get card spending limits and remaining balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			limits, err := client.GetCardLimits(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, limits)
			}

			f := outfmt.FromContext(cmd.Context())
			fmt.Printf("currency\t%s\n\n", limits.Currency)
			f.StartTable([]string{"INTERVAL", "LIMIT", "REMAINING"})
			for _, l := range limits.Limits {
				f.Row(l.Interval, fmt.Sprintf("%.2f", l.Amount), fmt.Sprintf("%.2f", l.Remaining))
			}
			return f.EndTable()
		},
	}
}
