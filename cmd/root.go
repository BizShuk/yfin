// root.go — cobra 根命令 + `Execute()` 入口點 + 共享 CLI 狀態。
// Subcommand 註冊由 `main.go`（composition root）依序呼叫每個 sub-package
// 的 `Register(RootCmd)` 完成，避免 cmd ↔ sub-package 之間的 import cycle。
//
// 各 sub-package 與職責對應：
// - cmd/admin       → config + version
// - cmd/dispatch    → batch + dispatch (FetchContext + commandRegistry)
// - cmd/fundamentals → fundamentals + comprehensive_stats + comprehensive_profile
// - cmd/market      → pull + quote + 本地 JSON 匯出
// - cmd/scrape      → scrape + scrape DTO formatters
// - cmd/twse        → twse (含 TWSE 專用 HTTP client builder)
package cmd

import (
	sdkconfig "github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/gosdk/metric"
	"github.com/spf13/cobra"
)

// RootCmd is the exported cobra root command. Sub-packages register onto it
// from main.go's composition path. Exporting it (rather than keeping a
// package-private rootCmd) lets `cmd/market` and friends register subcommands
// without `cmd/` itself needing to import those packages — avoiding an
// import cycle.
var RootCmd = &cobra.Command{
	Use:   "yfin",
	Short: "擷取並發布 Yahoo Finance 資料 (Yahoo Finance data fetcher and publisher)",
	Long: `yfin 是擷取 Yahoo Finance 資料的命令列工具 (yfin is a command-line tool for fetching Yahoo Finance data including)：
- 每日 bars（調整後 / 原始）— 僅限 daily 區間 (Daily bars, adjusted/raw — daily-only scope)
- 即時 quote 快照 (Snapshot quotes)
- fundamentals（需付費訂閱）(Fundamentals — requires paid subscription)

支援 FX 轉換預覽、bus 發布與本地匯出 (The tool supports FX conversion preview, bus publishing, and local export).`,
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
// RootCmd's PersistentPreRunE.
func Execute() error {
	metric.CobraCMDHook(RootCmd)
	return RootCmd.Execute()
}
