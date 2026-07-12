# Plan: cmd/ 一檔一職責重構

## Context

`cmd/root.go` 是 3014 行的 god file，把 8 個 cobra subcommand
（`pull` / `quote` / `fundamentals` / `scrape` / `comprehensive-stats` /
`comprehensive-profile` / `config` / `version`）加上 `init()` flag wiring、
`Execute()`、全部的 `run*` / `validate*` / `print*` / `handle*` /
`create*` / `format*` 邏輯全堆在一起。

同目錄下 `batch.go` (145) / `dispatch.go` (117) / `fetch.go` (129) /
`twse.go` (169) 已各自符合一檔一職責 — root.go 是唯一例外。

`cmd/` package 內無對外 caller（全在 `package cmd`），可自由重組 symbol。
`cmd/root_test.go` 同 package，直接搬遷測試函式即可。

目標：把 root.go 拆成多個聚焦檔案，每個檔案只負責一件事，
新結構遵循既有 `twse.go` 風格（一個 cobra command + 自有 `init()`），
並把 `root_test.go` 的測試搬到對應被測對象的檔案裡。

---

## 目標檔案結構

### 共用基礎 (3 檔)

| 檔案 | 職責 | 既有來源 |
|------|------|---------|
| `root.go` | 僅保留 cobra 根命令 + `Execute()` 入口 + 註冊所有 subcommand | `root.go:1-30, 141-158, 261-339, 341-346` |
| `exitcodes.go` | 5 個 `Exit*` 常數 | `root.go:43-49` |
| `global.go` | 全域 CLI flags：`GlobalConfig` struct + `globalConfig` var + persistent flag binding + `var _ = log.GetLogLevel` | `root.go:33, 52-61, 129-138, 150-158, 263-275` |

### 共用 client builders (1 檔)

| 檔案 | 職責 | 既有來源 |
|------|------|---------|
| `client.go` | `cliClient` 型別 + `createClient()` + `createBusConfig()` + `twseUserAgent` const + `buildTWSEClient()` + `writeJSONFile()`（pull & quote 共用 JSON 匯出器） | `root.go:999-1058, 1061-1140, 1442-1452, 1480-1522` |

### 各 subcommand 各自一檔 (7 檔)

每檔擁有：對應的 `*Config` struct + `*Config` 變數 + `*Cmd` 變數 +
自有 `init()`（bind flags + `rootCmd.AddCommand`）+ `run*` +
所有 `validate*` / `parse*` / `process*` / `print*` / `handle*` 私有 helper。

| 檔案 | 職責（對應 cobra subcommand） | 既有來源 |
|------|------|---------|
| `pull.go` | `pull` 每日 bars：validation + date/adjusted parse + symbol loading + per-symbol 處理 + bus publish + FX preview + 本地 JSON export | `root.go:64-79, 131, 161-172, 277-290, 349-461, 813-830, 941-991, 1143-1182, 1229-1247, 1287-1339, 1380-1414, 1455-1459` |
| `quote.go` | `quote` 即時報價：validation + ticker parse + 單 ticker 處理 + bus publish + 本地 JSON export | `root.go:82-90, 132, 175-184, 292-299, 463-523, 832-841, 1185-1210, 1249-1268, 1342-1377, 1417-1439, 1462-1466` |
| `fundamentals.go` | `fundamentals` 季報：validation + 處理 + preview + 401 paid-feature 偵測 | `root.go:93-96, 133, 187-196, 301-303, 525-561, 843-849, 1213-1226, 1270-1282, 1468-1474` |
| `scrape.go` | `scrape` 網頁爬蟲：4 種模式 (check / preview-json / preview-news / preview-proto) + `buildScrapeURL` + `createScrapeClient` + 共用時間/字串 helper | `root.go:99-109, 134, 199-212, 305-314, 563-645, 851-938, 1525-1578, 1581-1643, 1646-1678, 1681-1785, 2000-2023, 2734-2891` |
| `scrape_format.go` | scrape DTO → stdout formatter：`printAnalysisSummary` / `printAnalysisRow` / `printAnalysisCell` / `printAnalystInsightsSummary` / `printFundamentalsSnapshot` / `printProfileResult` / `printNewsArticles` / `getCurrencyFromLines` / `getTimeBounds` | `root.go:1788-1949, 1952-1997, 2894-3006` |
| `comprehensive_stats.go` | `comprehensive-stats` 5 年歷史財報：subcommand + fetch + DTO formatter | `root.go:111-115, 135, 215-225, 316-318, 648-709, 2026-2215` |
| `comprehensive_profile.go` | `comprehensive-profile` 公司基本資料：subcommand + fetch + DTO formatter | `root.go:117-121, 136, 228-238, 320-322, 2217-2403` |
| `version.go` | `version` 子命令 + `version` / `commit` / `date` vars + `runVersion` | `root.go:36-40, 254-259, 3008-3014` |
| `config.go` | `config` 子命令（`--print-effective`）+ effective config printer + flattenConfigMap | `root.go:123-127, 137, 241-251, 324-326, 712-810` |

### 測試檔案遷移 (3 檔調整)

| 來源 | 目標 |
|------|------|
| `root_test.go` `TestValidatePullFlags` / `TestParseDates` / `TestParseAdjusted` / `TestGetSymbols` / `TestWriteJSONFile` | `pull_test.go` (新增) |
| `root_test.go` `TestExitCodes` | `exitcodes_test.go` (新增) |
| `batch_test.go` / `dispatch_test.go` / `twse_test.go` | 維持原位不動 |

---

## 拆分規則

1. **同 package 拆分**：所有新檔案仍維持 `package cmd`，未匯出的 symbol 跨檔可直接呼叫。
2. **每檔自帶 `init()`**：bind 自己的 flags + `rootCmd.AddCommand(...)`。
   - `root.go` 的總 init() 拆掉 — 不再有集中註冊點。
3. **匯入精簡**：每檔只 import 它實際用到的 package；移除未用的 import。
4. **保留 doc comment 風格**：每檔頂部一行 `// <file>.go — <responsibility>. Capacity: <shape>.`
   與既有 `batch.go` / `dispatch.go` / `fetch.go` / `twse.go` 一致。
5. **`init()` 順序**：cobra 不在意子命令註冊順序；persistent flag 只要在
   任何 `RunE` 之前 bind 即可，而 init() 總在 `Execute()` 之前執行，
   所以分散到各檔的 init() 行為等價。

---

## 關鍵設計決策

| 決策 | 理由 |
|------|------|
| `scrape_format.go` 與 `scrape.go` 分開 | scrape.go 若含 formatters 會逼近 1300 行；formatters 全是純函式 (DTO → stdout text)，與 orchestration 是不同職責 |
| `comprehensive_stats.go` / `comprehensive_profile.go` 各自獨立 | 兩者用 scrape.Client 但屬獨立 subcommand；共用就違反一檔一 command |
| `writeJSONFile()` 放 `client.go` | pull + quote 都用，本來就不該綁在 pull.go 裡 |
| `buildTWSEClient()` + `twseUserAgent` 放 `client.go` | twse.go 透過 `twseClientProvider()` 引用，反向放 twse.go 會造成 twse.go 變胖且被併到 client 共用層 |
| `isPaidFeatureError()` 放 `fundamentals.go` | 唯一呼叫點是 `runFundamentals` |
| 測試檔案跟著被測對象搬 | 一檔一職責同時套用到測試；測試 ID 改變但不影響行為 |

---

## 修改清單（檔案視角）

- 修改：1（`root.go` 縮為 ~30 行 + 刪除大部分內容）
- 新增：12（exitcodes / global / version / config / pull / quote /
  fundamentals / scrape / scrape_format / comprehensive_stats /
  comprehensive_profile / client）
- 測試重組：3（拆 `root_test.go`、新增 `pull_test.go`、新增 `exitcodes_test.go`）

---

## Verification

1. `go build ./cmd/...` — 確認拆分後編譯過
2. `go vet ./cmd/...` — 檢查 init 順序、import 正確
3. `go test ./cmd/...` — 既有的 11 個測試函式（5 從 root_test.go 搬到
   pull_test.go、1 從 root_test.go 搬到 exitcodes_test.go、5 維持原位）
   全部維持綠燈
4. `./yfin --help` — 確認 8 個 subcommand（`pull` / `quote` /
   `fundamentals` / `scrape` / `comprehensive-stats` /
   `comprehensive-profile` / `config` / `version` + 既有 `batch` / `twse`）
   全部仍正確註冊
5. `wc -l cmd/*.go` — 確認沒有任何新檔案 > 700 行
6. `grep -L "^// .*\.go —" cmd/*.go` — 確認每個新檔案都有 doc-comment header

完成後 root.go 應縮為：

```go
// root.go — cobra 根命令 + Execute() 入口點。
package cmd

import (
    sdkconfig "github.com/bizshuk/gosdk/config"
    "github.com/bizshuk/gosdk/metric"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "yfin",
    Short: "Yahoo Finance data fetcher and publisher",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        sdkconfig.Default(sdkconfig.WithAppName("yfin"))
        return nil
    },
}

func Execute() error {
    metric.CobraCMDHook(rootCmd)
    return rootCmd.Execute()
}
```

（其餘子命令註冊改由各檔 init() 自行呼叫 `rootCmd.AddCommand`）