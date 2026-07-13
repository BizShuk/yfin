# CLI 命令參考 (CLI Commands Reference)

`yfin` CLI 的扁平 (flat) 命令樹、旗標表與退出碼。本檔為 reference 索引；完整使用範例見 [usage.md](usage.md)。

## CLI 樹狀結構 (CLI Tree)

`yfin` 採 flat 命令設計，無 nested subcommand。每個 subcommand 直接掛在 root：

```tree
yfin
├── config-effective    (cmd/admin)
├── version             (cmd/admin)
├── batch               (cmd/dispatch)
├── fundamentals        (cmd/fundamentals)
├── comprehensive-stats (cmd/fundamentals)
├── comprehensive-profile (cmd/fundamentals)
├── pull                (cmd/market)
├── quote               (cmd/market)
├── scrape              (cmd/scrape)
└── twse                (cmd/twse)
```

註：表格中的 `cmd/dispatch` 是 Go sub-package 名稱，並非對外 CLI 命令；`batch` 直接掛在 root。

## 持續旗標 (Persistent Flags)

以下 10 個旗標定義於 `cmd/global.go` 的 `RootCmd.PersistentFlags()`，每個 subcommand 都可使用：

| 旗標 | 型別 | 預設值 | 用途 |
| --- | --- | --- | --- |
| `--config` | string | `""` | YAML 設定檔路徑 |
| `--log-level` | string | `"info"` | Log level (`info` / `debug` / `warn` / `error`) |
| `--run-id` | string | `""` | Run ID（空白則自動產生） |
| `--concurrency` | int | `0` | Worker pool 大小（覆寫 YAML 預設；`0` 採用 YAML 值） |
| `--qps` | float64 | `0` | Per-host QPS（覆寫 YAML 預設；`0` 採用 YAML 值） |
| `--retry-max` | int | `0` | HTTP retry 嘗試次數 |
| `--sessions` | int | `0` | **Vestigial**：session rotation 已從 HTTP client 移除，旗標保留為向後相容，無實際效果 |
| `--timeout` | duration | `0` | HTTP timeout（如 `6s`） |
| `--observability-disable-tracing` | bool | `false` | 關閉 OpenTelemetry tracing |
| `--observability-disable-metrics` | bool | `false` | 關閉 Prometheus metrics |

## 命令參考 (Command Reference)

### `config-effective`

列印 CLI 解析後的 effective configuration（扁平 `key=value` 或 JSON）。

| 旗標 | 型別 | 說明 |
| --- | --- | --- |
| `--print-effective` | bool | 印出 effective config（必填） |
| `--json` | bool | 以 JSON 格式輸出（預設為 `key=value`） |

```bash
yfin config-effective --print-effective
yfin config-effective --print-effective --json
```

### `version`

列印 CLI 版本與 build 細節（`Version` / `Commit` / `Date`，由 `-ldflags` 注入）。無旗標。

```bash
yfin version
```

### `batch`

對 universe 中所有 symbols 批次執行 `commandRegistry` 內每個 command，套用 tiered cache 與 worker pool。預設讀取 `yf/references/ticker_list.csv`。

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--ticker` | string | `""` | 單一 ticker（省略時讀 `ticker_list.csv`） |
| `--max-workers` | int | `10` | Worker pool 大小上限 |
| `--force` | bool | `false` | 強制重抓（忽略快取） |

```bash
yfin batch
yfin batch --ticker AAPL --max-workers 32 --force
```

### `fundamentals`

擷取單一 symbol 的 fundamentals 季報。**需 Yahoo Finance 付費訂閱**，未授權時 exit code 為 `2` (`ExitPaidFeature`)。

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--ticker` | string | `""` | 必填。股票代號（如 `AAPL`） |
| `--preview` | bool | `false` | 顯示 preview（前 5 行） |

```bash
yfin fundamentals --ticker AAPL --preview
```

### `comprehensive-stats`

擷取完整 key statistics（含當前值與 5 年歷史），使用 YAML regex 模式從 Yahoo Finance 萃取。

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--ticker` | string | `""` | 必填。股票代號 |
| `--preview` | bool | `false` | 顯示 preview |

```bash
yfin comprehensive-stats --ticker AAPL
```

### `comprehensive-profile`

擷取公司基本資料 + key executives + 業務摘要。

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--ticker` | string | `""` | 必填。股票代號 |
| `--preview` | bool | `false` | 顯示 preview |

```bash
yfin comprehensive-profile --ticker MSFT --preview
```

### `pull`

擷取單一 symbol 或 universe 的每日 bars（**僅支援 daily**），支援本地檔案匯出。

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--ticker` | string | `""` | 單一 ticker（與 `--universe-file` 互斥） |
| `--universe-file` | string | `""` | newline-delimited ticker 列表（與 `--ticker` 互斥） |
| `--start` | string | `""` | 必填。起始日（`YYYY-MM-DD`，UTC） |
| `--end` | string | `""` | 必填。結束日（`YYYY-MM-DD`，UTC） |
| `--adjusted` | string | `"split_dividend"` | 調整策略（`raw` / `split_dividend`）；可由 YAML `markets.default_adjustment_policy` 覆寫 |
| `--market` | string | `""` | Market MIC（MIC inference 提示，可選） |
| `--fx-target` | string | `""` | FX 轉換目標幣別（如 `EUR`、`JPY`） |
| `--preview` | bool | `false` | 顯示 preview，不實際寫檔 |
| `--out` | string | `""` | 輸出格式（`json` / `parquet`，parquet 尚未實作） |
| `--out-dir` | string | `""` | 輸出目錄 |

```bash
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./data --preview
```

### `quote`

擷取即時 quote 快照（單一或多個 symbols）。

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--tickers` | string | `""` | 必填。CSV ticker 列表（如 `AAPL,MSFT,TSLA`） |
| `--preview` | bool | `false` | 顯示 preview，不實際寫檔 |
| `--out` | string | `""` | 輸出格式（僅支援 `json`） |
| `--out-dir` | string | `""` | 輸出目錄 |

```bash
yfin quote --tickers AAPL,MSFT,TSLA --preview
yfin quote --tickers AAPL --out json --out-dir ./quotes --preview
```

### `scrape`

Yahoo Finance 網頁爬蟲，提供 **4 種互斥 mode**（必填其一）。HTTP 連線由 `utils/httpx` 統一管理（限速、重試、circuit breaker）。

| Mode | 旗標 | 必要搭配 | 用途 |
| --- | --- | --- | --- |
| 連線測試 | `--check` | `--ticker` + `--endpoint` | 不解析、僅確認可達性 |
| extractor 乾跑 | `--preview-json` | `--ticker` + `--endpoints` | 解析 DTO 並輸出 JSON |
| news parser 乾跑 | `--preview-news` | `--ticker` | 解析 news 文章 |
| 完整輸出乾跑 | `--preview-proto` | `--ticker` + `--endpoints` | 完整 DTO 含 counts / periods / metadata |

**完整旗標表：**

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--check` | bool | `false` | 連線測試（與其他 mode 互斥） |
| `--preview-json` | bool | `false` | extractor 乾跑（互斥） |
| `--preview-news` | bool | `false` | news parser 乾跑（互斥） |
| `--preview-proto` | bool | `false` | 完整 DTO 輸出乾跑（互斥） |
| `--ticker` | string | `""` | 必填。股票代號 |
| `--endpoint` | string | `""` | 單一 endpoint，搭配 `--check`；可選值：`profile` / `key-statistics` / `financials` / `balance-sheet` / `cash-flow` / `analysis` / `analyst-insights` / `news` |
| `--endpoints` | string | `""` | CSV 多 endpoint，搭配 `--preview-json` / `--preview-proto` |
| `--preview` | bool | `false` | 顯示 preview |
| `--force` | bool | `false` | 忽略 API 可用性檢查，強制走爬蟲（僅測試用） |

```bash
yfin scrape --check --ticker AAPL --endpoint profile --preview
yfin scrape --preview-json --ticker AAPL --endpoints key-statistics,financials,analysis,profile
yfin scrape --preview-news --ticker TSLA
yfin scrape --preview-proto --ticker AAPL --endpoints financials,analysis,profile,news
```

### `twse`

查詢任一 TWSE（臺灣證券交易所）endpoint，輸出原始 JSON envelope。

| 旗標 | 型別 | 預設值 | 說明 |
| --- | --- | --- | --- |
| `--endpoint` | string | `""` | **必填**。TWSE endpoint 名稱 |
| `--date` | string | `""` | 日期（`YYYYMMDD`；`FMSRFK` / `STOCK_DAY_AVG` 用年份） |
| `--stock` | string | `""` | 個股代號（`STOCK_DAY` / `STOCK_DAY_AVG` / `BFIAUU_STOCK` / `FMSRFK` 必填） |
| `--month` | string | `""` | 月份（`YYYYMM`，`BFIMUU` / `FMTQIK` 必填） |
| `--timeout` | duration | `30s` | HTTP timeout |
| `--pretty` | bool | `false` | 將輸出 JSON 縮排美化 |

支援的 21 個 endpoint：`MI_INDEX` / `STOCK_DAY` / `BWIBBU_d` / `MI_INDEX_PLUS` / `MI_INDEX_ODD` / `MI_5MINS` / `TWTB4U` / `MI_MARGN` / `T86` / `MI_QFIIS` / `BFI82U` / `TWT38U` / `TWT43U` / `TWT44U` / `BFIAUU` / `BFIAUU_STOCK` / `BFIMUU` / `BFIAUU_YEAR` / `FMTQIK` / `STOCK_DAY_AVG` / `FMSRFK` / `BFIAMU` / `MI_WEEK`。

```bash
yfin twse --endpoint MI_INDEX --date 20221230
yfin twse --endpoint STOCK_DAY --date 20221230 --stock 2330
yfin twse --endpoint FMTQIK --month 202212
yfin twse --endpoint MI_WEEK --date 20221230 --pretty
```

## 退出碼 (Exit Codes)

定義於 `cmd/exitcodes.go`：

| Code | 常數 | 說明 |
| --- | --- | --- |
| `0` | `ExitSuccess` | 成功 |
| `1` | `ExitGeneral` | 一般錯誤（網路、執行失敗等） |
| `2` | `ExitPaidFeature` | 付費功能未授權（fundamentals） |
| `3` | `ExitConfigError` | 組態錯誤（CLI flag 驗證失敗、YAML 載入失敗） |

## 獨立 binary (Standalone Binary)

`cmd/soak/` 是獨立 binary，**不**是 `yfin` subcommand；以 `go run` 啟動：

```bash
go run ./cmd/soak
```

詳細使用見 [soak-testing.md](../soak-testing.md)。

## 參見 (See Also)

- [usage.md](usage.md) — 完整使用範例與整合腳本
- [operations/configuration.md](../operations/configuration.md) — YAML config 結構
- [operations/error-handling.md](../operations/error-handling.md) — 退出碼語義與錯誤處理
- [scrape/cli.md](../scrape/cli.md) — `scrape` 子命令深度指南
- [soak-testing.md](../soak-testing.md) — `cmd/soak` 獨立 binary