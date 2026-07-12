// fundamentals.go — `fundamentals` + `comprehensive-stats` +
// `comprehensive-profile` cobra subcommands 共用 sub-package。三者邏輯上
// 都是 fundamentals 類資料；個別 subcommand 仍保留自己的 config struct
// 與 formatter，本檔只負責 Register 與共用規劃。
//
// - stats.go    → `comprehensive-stats` subcommand + stats formatter
// - profile.go  → `comprehensive-profile` subcommand + profile formatter
// - fundamentals.go（本檔）→ `fundamentals` subcommand + Register helper
package fundamentals

import "github.com/spf13/cobra"

// Register attaches the three fundamentals-family subcommands onto rootCmd.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newFundamentalsCmd())
	rootCmd.AddCommand(newComprehensiveStatsCmd())
	rootCmd.AddCommand(newComprehensiveProfileCmd())
}
