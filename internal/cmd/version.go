package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/iocontext"
	"github.com/salmonumbrella/airwallex-cli/internal/outfmt"
	"github.com/salmonumbrella/airwallex-cli/internal/update"
)

var (
	// Version is the semantic version (set via ldflags)
	Version = "dev"
	// Commit is the git commit hash (set via ldflags)
	Commit = "unknown"
	// BuildDate is the build timestamp (set via ldflags)
	BuildDate = "unknown"
)

// versionInfo holds structured version information for JSON output
type versionInfo struct {
	Version         string `json:"version"`
	Commit          string `json:"commit"`
	BuildDate       string `json:"build_date"`
	LatestVersion   string `json:"latest_version,omitempty"`
	UpdateAvailable bool   `json:"update_available,omitempty"`
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Display version, commit hash, and build date for this binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for updates (non-blocking, ignores errors)
			updateResult := update.CheckForUpdate(cmd.Context(), Version)

			// Get IO streams - prefer context IO, fall back to cobra's writers
			io := iocontext.GetIO(cmd.Context())
			out := io.Out
			errOut := io.ErrOut

			// If IO from context is still the default (os.Stdout), check if cobra has custom writers
			if out == os.Stdout {
				out = cmd.OutOrStdout()
			}
			if errOut == os.Stderr {
				errOut = cmd.OutOrStderr()
			}

			if outfmt.IsJSON(cmd.Context()) {
				info := versionInfo{
					Version:   Version,
					Commit:    Commit,
					BuildDate: BuildDate,
				}
				if updateResult != nil {
					info.LatestVersion = updateResult.LatestVersion
					info.UpdateAvailable = updateResult.UpdateAvailable
				}
				return outfmt.WriteJSON(out, info)
			}

			_, _ = fmt.Fprintf(out, "airwallex-cli %s\n", Version)
			_, _ = fmt.Fprintf(out, "  commit:     %s\n", Commit)
			_, _ = fmt.Fprintf(out, "  build date: %s\n", BuildDate)

			if updateResult != nil && updateResult.UpdateAvailable {
				_, _ = fmt.Fprintf(errOut, "\nUpdate available: %s → %s\n",
					updateResult.CurrentVersion, updateResult.LatestVersion)
				_, _ = fmt.Fprintf(errOut, "Run: brew upgrade airwallex-cli\n")
			}

			return nil
		},
	}

	return cmd
}
