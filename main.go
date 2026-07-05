// main.go — yfin CLI composition root; forwards `cmd.Execute()` to the process and exits non-zero on error. Capacity: 1 entrypoint + 1 os.Exit fallback.
package main

import (
	"os"

	"github.com/bizshuk/yfin/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
