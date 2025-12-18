package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the semantic version (set via ldflags)
	Version = "dev"
	// Commit is the git commit hash (set via ldflags)
	Commit = "unknown"
	// BuildDate is the build timestamp (set via ldflags)
	BuildDate = "unknown"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Display version, commit hash, and build date for this binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("airwallex-cli %s\n", Version)
			fmt.Printf("  commit:     %s\n", Commit)
			fmt.Printf("  build date: %s\n", BuildDate)
			return nil
		},
	}

	return cmd
}
