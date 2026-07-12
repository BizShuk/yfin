// root.go — cobra 根命令 + `Execute()` 入口點。每個 subcommand（pull / quote /
// fundamentals / scrape / comprehensive-stats / comprehensive-profile /
// config / version / batch / twse）各自檔案負責自己的 config struct、flag
// binding、init() 註冊、RunE 與所有 helper；root.go 只保留 ProcessTree 入口。
package cmd

import (
	sdkconfig "github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/gosdk/metric"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
// Subcommand registration happens in each subcommand file's init().
var rootCmd = &cobra.Command{
	Use:   "yfin",
	Short: "Yahoo Finance data fetcher and publisher",
	Long: `yfin is a command-line tool for fetching Yahoo Finance data including:
- Daily bars (adjusted/raw) - daily-only scope
- Snapshot quotes
- Fundamentals (requires paid subscription)

The tool supports FX conversion preview, bus publishing, and local export.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// gosdk bootstrap: load ~/.config/yfin + bind APP_* env (e.g. APP_LOG_LEVEL).
		// gosdk/log's package init installs slog.Default at import time (this
		// binary imports `log` via cmd/global.go's `var _ = log.GetLogLevel`),
		// so any APP_LOG_LEVEL set on the process environment is honored at startup.
		sdkconfig.Default(sdkconfig.WithAppName("yfin"))
		return nil
	},
}

// Execute runs the root cobra command. gosdk/metric's hook emits a
// command_line_trigger metric for every subcommand execution, then chains
// rootCmd's PersistentPreRunE.
func Execute() error {
	metric.CobraCMDHook(rootCmd)
	return rootCmd.Execute()
}
