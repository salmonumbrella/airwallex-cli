package cmd

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
)

func writeJSONOutput(cmd *cobra.Command, value interface{}) error {
	return writeJSONOutputTo(cmd.Context(), commandOutputWriter(cmd), value)
}

func writeJSONOutputTo(ctx context.Context, w io.Writer, value interface{}) error {
	return outfmt.WriteJSONForContext(ctx, w, value)
}

func commandOutputWriter(cmd *cobra.Command) io.Writer {
	out := iocontext.GetIO(cmd.Context()).Out
	if out == os.Stdout {
		return cmd.OutOrStdout()
	}
	return out
}
