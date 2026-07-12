// build.go — build-time version variables injected via `-ldflags` at compile
// time. They are exported because scrape + comprehensive_stats/profile need
// to read them for observability metadata, and admin/* (`yfin version`) needs
// them for `--version` output. Capacity: 3 exported vars.
package cmd

// Version information set via ldflags during build
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
