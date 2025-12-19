package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

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
				return outfmt.WriteJSON(cmd.OutOrStdout(), info)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "airwallex-cli %s\n", Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  commit:     %s\n", Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "  build date: %s\n", BuildDate)

			if updateResult != nil && updateResult.UpdateAvailable {
				fmt.Fprintf(cmd.OutOrStderr(), "\nUpdate available: %s → %s\n",
					updateResult.CurrentVersion, updateResult.LatestVersion)
				fmt.Fprintf(cmd.OutOrStderr(), "Run: brew upgrade airwallex-cli\n")
			}

			return nil
		},
	}

	return cmd
}
