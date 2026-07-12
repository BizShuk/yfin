// scrape.go — `scrape` cobra subcommand + DTO → stdout formatters grouped
// under one sub-package. The four scrape modes (--check connectivity /
// --preview-json extractor dry-run / --preview-news news parser dry-run /
// --preview-proto full proto emission dry-run) all share the same
// `scrape.Client` builder, URL builder, time/string helpers, and DTO
// formatters. Capacity: 1 `Register(rootCmd)` + 4 mode runners + DTO
// formatters in format.go.
package scrape

import "github.com/spf13/cobra"

// Register attaches the `scrape` subcommand onto rootCmd.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newScrapeCmd())
}
