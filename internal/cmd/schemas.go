package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/flagmap"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

// schemaKeyToFlag finds the CLI flag that maps to a given schema field key or path
func schemaKeyToFlag(key, path string) string {
	// Try matching against the full path first, then the key
	for flag, mapping := range flagmap.AllMappings() {
		if mapping.SchemaPath == path || strings.HasSuffix(mapping.SchemaPath, "."+key) {
			return "--" + flag
		}
	}
	return ""
}

func newSchemasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schemas",
		Aliases: []string{"schema", "sc"},
		Short:   "API schema discovery",
		Long:    "Discover required fields for beneficiaries and transfers by country and payment method.",
	}
	cmd.AddCommand(newSchemasBeneficiaryCmd())
	cmd.AddCommand(newSchemasTransferCmd())
	return cmd
}

func newSchemasBeneficiaryCmd() *cobra.Command {
	var bankCountry, entityType, paymentMethod string

	cmd := &cobra.Command{
		Use:     "beneficiary",
		Aliases: []string{"ben", "b"},
		Short:   "Get required fields for creating a beneficiary",
		Long: `Discover which fields are required to create a beneficiary for a specific country and entity type.

Examples:
  airwallex schemas beneficiary --bank-country US --entity-type COMPANY
  airwallex schemas beneficiary --bank-country CA --entity-type PERSONAL --payment-method LOCAL`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			schema, err := client.GetBeneficiarySchema(cmd.Context(), bankCountry, entityType, paymentMethod)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return writeJSONOutput(cmd, schema)
			}

			formatter := outfmt.FromContext(cmd.Context())

			if len(schema.Fields) == 0 {
				formatter.Empty("No fields returned")
				return nil
			}

			formatter.StartTable([]string{"FIELD", "TYPE", "REQUIRED", "CLI FLAG", "DESCRIPTION"})
			for _, f := range schema.Fields {
				required := ""
				if f.Required {
					required = "*"
				}
				desc := f.Description
				if len(f.Enum()) > 0 && len(desc) == 0 {
					desc = fmt.Sprintf("enum: %v", f.Enum())
				}
				cliFlag := schemaKeyToFlag(f.Key, f.Path)
				formatter.Row(f.Name(), f.Type(), required, cliFlag, desc)
			}
			return formatter.EndTable()
		},
	}

	cmd.Flags().StringVar(&bankCountry, "bank-country", "", "Bank country code (required)")
	cmd.Flags().StringVar(&entityType, "entity-type", "", "Entity type: COMPANY or PERSONAL (required)")
	cmd.Flags().StringVar(&paymentMethod, "payment-method", "", "Payment method: LOCAL or SWIFT")
	mustMarkRequired(cmd, "bank-country")
	mustMarkRequired(cmd, "entity-type")
	flagAlias(cmd.Flags(), "bank-country", "bk")
	flagAlias(cmd.Flags(), "entity-type", "et")
	flagAlias(cmd.Flags(), "payment-method", "pm")
	return cmd
}

func newSchemasTransferCmd() *cobra.Command {
	var sourceCurrency, destCurrency, paymentMethod string

	cmd := &cobra.Command{
		Use:     "transfer",
		Aliases: []string{"xfer", "t"},
		Short:   "Get required fields for creating a transfer",
		Long: `Discover which fields are required to create a transfer for a specific currency pair.

Examples:
  airwallex schemas transfer --source-currency USD --dest-currency EUR
  airwallex schemas transfer --source-currency CAD --dest-currency CAD --payment-method LOCAL`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			schema, err := client.GetTransferSchema(cmd.Context(), sourceCurrency, destCurrency, paymentMethod)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return writeJSONOutput(cmd, schema)
			}

			formatter := outfmt.FromContext(cmd.Context())

			if len(schema.Fields) == 0 {
				formatter.Empty("No fields returned")
				return nil
			}

			formatter.StartTable([]string{"FIELD", "TYPE", "REQUIRED", "CLI FLAG", "DESCRIPTION"})
			for _, f := range schema.Fields {
				required := ""
				if f.Required {
					required = "*"
				}
				desc := f.Description
				if len(f.Enum()) > 0 && len(desc) == 0 {
					desc = fmt.Sprintf("enum: %v", f.Enum())
				}
				cliFlag := schemaKeyToFlag(f.Key, f.Path)
				formatter.Row(f.Name(), f.Type(), required, cliFlag, desc)
			}
			return formatter.EndTable()
		},
	}

	cmd.Flags().StringVar(&sourceCurrency, "source-currency", "", "Source currency (required)")
	cmd.Flags().StringVar(&destCurrency, "dest-currency", "", "Destination currency (required)")
	cmd.Flags().StringVar(&paymentMethod, "payment-method", "", "Payment method: LOCAL or SWIFT")
	mustMarkRequired(cmd, "source-currency")
	mustMarkRequired(cmd, "dest-currency")
	flagAlias(cmd.Flags(), "source-currency", "sc")
	flagAlias(cmd.Flags(), "dest-currency", "dc")
	flagAlias(cmd.Flags(), "payment-method", "pm")
	return cmd
}
