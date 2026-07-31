// market.go — `pull` (每日 bars) + `quote` (即時報價) cobra subcommands
// 共用同一個 local-export sink（writeJSONFile），因此收進同一 sub-package。
// 兩個 subcommand 都提供本地 JSON 匯出，是 yfin
// SDK 對應的「market data」API 集合；本檔只負責 command registration。
package market

import "github.com/spf13/cobra"

// Register attaches the `pull` and `quote` subcommands onto rootCmd.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newPullCmd())
	rootCmd.AddCommand(newQuoteCmd())
}
