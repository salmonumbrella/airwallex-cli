package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/airwallex-cli/internal/ui"
	"github.com/salmonumbrella/airwallex-cli/internal/update"
)

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade to the latest version",
		Long: `Check for and install the latest version of airwallex-cli.

On macOS with Homebrew:
  brew upgrade airwallex-cli

On other systems:
  go install github.com/salmonumbrella/airwallex-cli/cmd/awx@latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := ui.FromContext(cmd.Context())

			// Check current version
			result := update.CheckForUpdate(cmd.Context(), Version)
			if result == nil {
				u.Info("Unable to check for updates (dev build or network issue)")
				return nil
			}

			if !result.UpdateAvailable {
				u.Success(fmt.Sprintf("Already at latest version (%s)", result.CurrentVersion))
				return nil
			}

			u.Info(fmt.Sprintf("Update available: %s → %s", result.CurrentVersion, result.LatestVersion))

			// Try Homebrew on macOS
			if runtime.GOOS == "darwin" {
				if _, err := exec.LookPath("brew"); err == nil {
					u.Info("Upgrading via Homebrew...")
					upgradeCmd := exec.CommandContext(cmd.Context(), "brew", "upgrade", "airwallex-cli")
					upgradeCmd.Stdout = cmd.OutOrStdout()
					upgradeCmd.Stderr = cmd.OutOrStderr()
					if err := upgradeCmd.Run(); err != nil {
						return fmt.Errorf("homebrew upgrade failed: %w", err)
					}
					u.Success("Upgrade complete!")
					return nil
				}
			}

			// Fallback: show manual instructions
			u.Info("To upgrade manually, run:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  go install github.com/salmonumbrella/airwallex-cli/cmd/awx@latest")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Or download from:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  "+result.UpdateURL)

			return nil
		},
	}
}
