// version.go — `version` cobra subcommand plus the build-time version vars
// (`version` / `commit` / `date`) injected via `-ldflags`. Capacity: 3 build
// vars + 1 `versionCmd` + 1 `init()` (no flags) + 1 `runVersion`.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information set via ldflags during build
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version information including build details.`,
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// runVersion executes the version command
func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("yfin version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("build date: %s\n", date)
	return nil
}
