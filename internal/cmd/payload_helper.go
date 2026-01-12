package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/ui"
)

// PayloadCommandConfig defines a JSON payload command with standard --data/--from-file flags.
type PayloadCommandConfig[T any] struct {
	Use     string
	Short   string
	Long    string
	Example string

	Args cobra.PositionalArgs

	ReadPayload    func(data, fromFile string) (map[string]interface{}, error)
	Run            func(ctx context.Context, client *api.Client, args []string, payload map[string]interface{}) (T, error)
	SuccessMessage func(T) string
}

// NewPayloadCommand builds a command that reads a JSON payload and executes a request.
func NewPayloadCommand[T any](cfg PayloadCommandConfig[T], getClient func(context.Context) (*api.Client, error)) *cobra.Command {
	var data string
	var fromFile string

	cmd := &cobra.Command{
		Use:     cfg.Use,
		Short:   cfg.Short,
		Long:    cfg.Long,
		Example: cfg.Example,
		Args:    cfg.Args,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Run == nil {
				return fmt.Errorf("payload command missing Run")
			}

			client, err := getClient(cmd.Context())
			if err != nil {
				return err
			}

			readPayload := cfg.ReadPayload
			if readPayload == nil {
				readPayload = readJSONPayload
			}

			payload, err := readPayload(data, fromFile)
			if err != nil {
				return err
			}

			result, err := cfg.Run(cmd.Context(), client, args, payload)
			if err != nil {
				return err
			}

			if outfmt.IsJSON(cmd.Context()) {
				return outfmt.WriteJSON(os.Stdout, result)
			}

			if cfg.SuccessMessage != nil {
				u := ui.FromContext(cmd.Context())
				u.Success(cfg.SuccessMessage(result))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Inline JSON payload")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to JSON payload file (- for stdin)")

	return cmd
}
