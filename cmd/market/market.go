// market.go — `pull` (每日 bars) + `quote` (即時報價) cobra subcommands
// 共用同一個 local-export sink（writeJSONFile），因此收進同一 sub-package。
// 兩個 subcommand 都需要 bus publishing + 本地 JSON 匯出，是 yfinance-go
// SDK 對應的「market data」API 集合。 Capacity: 1 `Register(rootCmd)` 註冊
// pull + quote；本檔只負責 cobra command 構建與各 command helper 之間的
// 依賴指向 (pull.go / quote.go / client_json.go)。
package market

import "github.com/spf13/cobra"

// Register attaches the `pull` and `quote` subcommands onto rootCmd.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newPullCmd())
	rootCmd.AddCommand(newQuoteCmd())
}
