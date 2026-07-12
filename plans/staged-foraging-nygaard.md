# Plan: cmd/ 改用 sub-package 結構

## Context

前一輪已經把 `cmd/root.go`（3014 行 god file）拆成 22 個 `package cmd` 內的單一職責檔案。

這一輪再進化：把 10 個 cobra subcommands 用 **domain 分組** 收進 6 個
`cmd/` 下的 sub-package，每個 sub-package 只暴露一個 `Register(rootCmd)`
函式；`cmd/` 自身只剩根命令、shared infra（exitcodes / global flags /
client builders）。

目標結構：

```
cmd/
├── root.go            # rootCmd + Execute
├── exitcodes.go       # Exit* 常數
├── global.go          # GlobalConfig + 持久 flags
├── client.go          # cliClient + createClient + createBusConfig
├── market/            # pull + quote + writeJSONFile
├── fundamentals/      # fundamentals + comprehensive_stats + comprehensive_profile
├── scrape/            # scrape + scrape_format
├── dispatch/          # batch + dispatch (FetchContext + commandRegistry)
├── admin/             # config + version
└── twse/              # twse + buildTWSEClient + twseUserAgent + twseClientProvider
```

`cmd/soak/` 與 `cmd/tools/golden/` 維持原樣（已是獨立 binary entry）。

---

## Sub-package 對應

| Sub-package | 內含檔案 | 暴露 |
|-------------|---------|------|
| `cmd/market` | `pull.go`, `quote.go`, `client_json.go`（共用 `writeJSONFile`） | `Register(rootCmd)` 註冊 `pull` + `quote` |
| `cmd/fundamentals` | `fundamentals.go`, `stats.go`（comprehensive-stats）, `profile.go`（comprehensive-profile）| `Register(rootCmd)` 註冊 3 個 subcommands |
| `cmd/scrape` | `scrape.go`, `format.go`（scrape_format） | `Register(rootCmd)` 註冊 `scrape` |
| `cmd/dispatch` | `batch.go`, `registry.go`（dispatch：FetchContext + commandRegistry） | `Register(rootCmd)` 註冊 `batch` |
| `cmd/admin` | `config.go`, `version.go` | `Register(rootCmd)` 註冊 `config` + `version` |
| `cmd/twse` | `twse.go` + `client.go`（buildTWSEClient + twseUserAgent） | `Register(rootCmd)` 註冊 `twse` |

每個 sub-package 內的 `*Config` struct / 私有 helper 都是 package-local，
不再洩漏到外部。

---

## `cmd/` 根層只保留共用 infra

保留：

| 檔案 | 內容 |
|------|------|
| `root.go` | `rootCmd`, `Execute()`, `PersistentPreRunE` |
| `exitcodes.go` | `ExitSuccess/ExitGeneral/ExitPaidFeature/ExitConfigError/ExitPublishError` |
| `global.go` | `GlobalConfig` struct + exported `Global` var + 持久 flags binding + `var _ = log.GetLogLevel` |
| `client.go` | `cliClient` type, `CreateClient()`, `CreateBusConfig()` |

要 export 給 sub-packages 用：

| Symbol | 變更 |
|--------|------|
| `globalConfig` | rename → `Global`（exported var） |
| `createClient` | rename → `CreateClient`（exported func） |
| `createBusConfig` | rename → `CreateBusConfig`（exported func） |
| `version`, `commit`, `date` | rename → `Version`, `Commit`, `Date`（exported vars） |

`cliClient` 與 `Exit*` 已 exported，不變。

---

## Sub-package API 設計

每個 sub-package 統一介面：

```go
// e.g. cmd/market/market.go
package market

import "github.com/spf13/cobra"

// Register binds `pull` and `quote` subcommands onto rootCmd.
func Register(rootCmd *cobra.Command) {
    rootCmd.AddCommand(newPullCmd())
    rootCmd.AddCommand(newQuoteCmd())
}
```

`cmd/root.go` 一個地方呼叫 6 個 `Register`：

```go
import (
    "github.com/bizshuk/yfin/cmd/admin"
    "github.com/bizshuk/yfin/cmd/dispatch"
    "github.com/bizshuk/yfin/cmd/fundamentals"
    "github.com/bizshuk/yfin/cmd/market"
    "github.com/bizshuk/yfin/cmd/scrape"
    "github.com/bizshuk/yfin/cmd/twse"
)

func init() {
    admin.Register(rootCmd)
    dispatch.Register(rootCmd)
    fundamentals.Register(rootCmd)
    market.Register(rootCmd)
    scrape.Register(rootCmd)
    twse.Register(rootCmd)
}
```

顯式 import 清單，cobra 註冊順序一目了然；測試時也能個別 sub-package 隔離驗證。

---

## Critical files to modify

- 修改：
  - `cmd/root.go` — 加 import 區塊 + init() 註冊 6 個 sub-package
  - `cmd/global.go` — `globalConfig` → `Global`、`GlobalConfig` struct 仍 exported
  - `cmd/client.go` — `createClient` → `CreateClient`、`createBusConfig` → `CreateBusConfig`、移除 `buildTWSEClient` + `twseUserAgent` + `writeJSONFile`
  - `cmd/version.go` — `version`/`commit`/`date` → exported（會搬到 `cmd/admin/`）
  - `cmd/batch.go`, `cmd/dispatch.go` → 移入 `cmd/dispatch/`，測試 `cmd/batch_test.go` + `cmd/dispatch_test.go` 一起跟著搬
  - `cmd/twse.go`, `cmd/twse_test.go` → 移入 `cmd/twse/`，加上 `buildTWSEClient` + `twseUserAgent` 跟著搬
- 刪除（被 sub-package 取代）：
  - `cmd/pull.go`, `cmd/quote.go`（移入 `cmd/market/`）
  - `cmd/fundamentals.go`, `cmd/comprehensive_stats.go`, `cmd/comprehensive_profile.go`（移入 `cmd/fundamentals/`）
  - `cmd/scrape.go`, `cmd/scrape_format.go`（移入 `cmd/scrape/`）
  - `cmd/config.go`, `cmd/version.go`（移入 `cmd/admin/`）
  - `cmd/pull_test.go`（跟著搬到 `cmd/market/pull_test.go`）
- 新增：
  - `cmd/market/{pull.go, quote.go, client_json.go, market.go, pull_test.go}`
  - `cmd/fundamentals/{fundamentals.go, stats.go, profile.go, fundamentals.go}`
  - `cmd/scrape/{scrape.go, format.go, scrape.go}`
  - `cmd/dispatch/{batch.go, registry.go, batch_test.go, dispatch_test.go, dispatch.go}`
  - `cmd/admin/{config.go, version.go, admin.go}`
  - `cmd/twse/{twse.go, client.go, twse_test.go, twse.go}`
- 維持：
  - `cmd/root.go`, `cmd/exitcodes.go`, `cmd/global.go`, `cmd/client.go`
  - `cmd/exitcodes_test.go`, `cmd/fetch.go`, `cmd/soak/`, `cmd/tools/golden/`

---

## Verification

1. `go build ./cmd/...` — 確認 6 個 sub-package 編譯通過
2. `go vet ./cmd/...` — init 順序、import cycle 檢查
3. `go test ./cmd/...` — 既有的 11 個測試函式（5 從 pull_test.go 搬到 `cmd/market/`、1 exitcodes_test 維持原位、5 從 batch_test.go + dispatch_test.go + twse_test.go 跟著搬）全部綠燈
4. `./yfin --help` — 10 個 subcommand 全部仍正確註冊
5. `wc -l cmd/*.go cmd/*/*.go` — 確認沒有檔案 > 700 行；sub-package 內每檔也遵守上限
6. `grep -L "^// .*\.go —" cmd/*.go cmd/*/*.go` — 確認每個檔案都有 doc-comment header

---

## Out of scope

- `cmd/batch.go:88` 寫死 `~/.config/stock/data/raw` 的預先 bug
  （觀察 8719）— 不在本輪動，維持目前行為

## 風險

| 風險 | 緩解 |
|------|------|
| Sub-package 對 `cmd` 的 import cycle | 不存在：sub-package 只 import `cmd`，`cmd` 不 import sub-package（透過 `Register(rootCmd)` 反向注入） |
| 既有 `batch_test.go` / `dispatch_test.go` / `twse_test.go` / `pull_test.go` 改 package 名後遺失測試 | 改 package 名時一併改測試 import 與 helper（如 `setTwseClientForTest`），同 package 內測試仍可存取 unexported symbol |
| `createClient` / `globalConfig` rename 漏改某呼叫點 | 改完用 `go build` 把 compile error 全部抓出，一次修齊 |
| 重新命名後 `version` 的 doc-comment 用法破壞 cobra 顯示 | `Version` / `Commit` / `Date` 只在 `runVersion` 與 scrape/comprehensive-* 內部讀，export 不影響 `--help` 輸出 |