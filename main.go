// main.go — yfin CLI composition root; wires every sub-package's `Register`
// onto `cmd.RootCmd` then forwards `cmd.Execute()` to the process. Exit non-zero
// on error. Capacity: 1 entrypoint + 1 os.Exit fallback + 6 Register calls.
package main

import (
	"os"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/cmd/admin"
	"github.com/bizshuk/yfin/cmd/dispatch"
	"github.com/bizshuk/yfin/cmd/fundamentals"
	"github.com/bizshuk/yfin/cmd/market"
	"github.com/bizshuk/yfin/cmd/scrape"
	"github.com/bizshuk/yfin/cmd/twse"
)

func main() {
	// Composition root: explicit subcommand registration list.
	// Each sub-package owns its own cobra commands and exposes a single
	// `Register(rootCmd)` entrypoint.
	admin.Register(cmd.RootCmd)
	dispatch.Register(cmd.RootCmd)
	fundamentals.Register(cmd.RootCmd)
	market.Register(cmd.RootCmd)
	scrape.Register(cmd.RootCmd)
	twse.Register(cmd.RootCmd)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
