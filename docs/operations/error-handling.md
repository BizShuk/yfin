# 錯誤處理與疑難排解指南 (Error Handling & Troubleshooting)

本指南涵蓋 `yfin` 的退出碼 (exit code) 語意、錯誤分類、CLI 端錯誤、爬蟲備援 (scrape fallback) 錯誤、HTTP 層錯誤、常見情境處置與除錯工具組。

## Exit Codes

`yfin` CLI 回傳穩定的退出碼 (定義於 `cmd/exitcodes.go`)，讓 shell script 與排程器可依錯誤類別分支處理。

| Code | 常數 | 語意 | 觸發情境 (典型) |
| ---- | ---- | ---- | --------------- |
| 0 | `ExitSuccess` | 命令正常完成 | 一般成功結束 |
| 1 | `ExitGeneral` | 一般 / 未分類錯誤 | unhandled error、網路失敗、parse 失敗、所有 ticker 全部失敗 |
| 2 | `ExitPaidFeature` | 此 endpoint 需付費訂閱 | `yfin fundamentals` 在沒有 Yahoo Finance 訂閱時收到 401 / `paid subscription` |
| 3 | `ExitConfigError` | 使用者 / 配置錯誤 | 旗標值非法、缺少必要旗標、YAML 載入失敗、scrape 停用、interval 非 `1d` |

**對應命令的發出點：**

- `ExitConfigError` — `cmd/market/pull.go` (validatePullFlags、parseDates、parseAdjusted、YAML load、interval validate)、`cmd/scrape/scrape_run.go` (validateScrapeFlags、YAML load、scrape disabled、observability init)、`cmd/fundamentals/*` (validate flags、YAML load、observability init)、`cmd/admin/admin.go`。
- `ExitGeneral` — `cmd/market/pull.go` (CreateClient 失敗、所有 symbol 全失敗)、`cmd/scrape/scrape_run.go` (CreateClient 失敗、未指定 mode)、`cmd/fundamentals/*` (CreateClient 失敗、未分類錯誤)。
- `ExitPaidFeature` — `cmd/fundamentals/fundamentals_run.go` (回傳錯誤字串含 `paid subscription` / `401` / `Unauthorized`)。

## Error Classification

錯誤依來源與重試行為分成六大類。

### Network Errors

**症狀：** 連線逾時、DNS 解析失敗、connection refused。

**錯誤訊息樣板：**

```
context deadline exceeded
connection refused
no such host
network is unreachable
```

**處置：** 檢查網路連線、確認 DNS、設定 proxy / firewall、調高 `--timeout`。

### Rate Limiting Errors

**症狀：** HTTP 429、響應緩慢、Yahoo Finance 節流 (throttling)。

**錯誤訊息樣板：**

```
429 Too Many Requests
rate limit exceeded
too many requests
```

**處置：** 降低 `--qps`、`scrape.qps` 與 `scrape.burst`、放大 `scrape.retry.max_delay_ms`、等待冷卻期。

### Authentication / Paid Subscription Errors

**症狀：** HTTP 401、`paid subscription required`。

**錯誤訊息樣板：**

```
401 Unauthorized
authentication required
paid subscription required
```

**處置：** 確認 Yahoo Finance 訂閱狀態；無訂閱時改走 scrape fallback（`yfin scrape --preview-proto --ticker X --endpoints financials,key-statistics,...`）。

### Parse Errors

**症狀：** JSON / HTML 解析失敗、欄位缺失、型別不符。

**錯誤訊息樣板：**

```
invalid character
unexpected end of JSON input
cannot unmarshal
parse error
schema validation failed
```

**處置：** 更新至最新版本、確認 Yahoo Finance HTML / JSON 結構未變動、實作 graceful degradation、回報 issue（附 debug log）。

### Validation / Data Quality Errors

**症狀：** 欄位為零、空 batch、負值 volume、currency 缺漏。

**錯誤訊息樣板：**

```
no bars in batch
missing or non-positive market price
missing currency code
non-positive prices detected
```

**處置：** 檢查 symbol 與日期範圍是否有效、將空 batch 視為正常結果 (`len(bars.Bars) == 0`)、檢查 nil 指標。

### Data Not Found

**症狀：** 空結果、`no data found`。

**錯誤訊息樣板：**

```
no quotes found
no bars found
no news articles found
symbol not found
```

**處置：** 驗證 symbol、檢查日期區間、辨識市場是否休市；以 graceful degradation 處理空結果。

## CLI-side Errors (Exit Code 3 / 4)

### Exit Code 3 — Configuration Errors

`cmd/market/pull.go`、`cmd/scrape/scrape_run.go`、`cmd/fundamentals/*`、`cmd/admin/admin.go` 都會在配置階段失敗時輸出 `ERROR:` 並退出 `ExitConfigError`。

**典型觸發點：**

| 觸發點 | 範例錯誤 |
| ------ | -------- |
| `--ticker` 與 `--universe-file` 兩者皆缺 / 同時給 | `either --ticker or --universe-file must be specified` |
| `--start` / `--end` 缺漏或日期格式錯誤 | `invalid start date: ...` |
| `--adjusted` 值非法 | `--adjusted must be 'raw' or 'split_dividend'` |
| `--out` 值非法 | `--out must be 'json' or 'parquet'` |
| YAML 載入失敗 | `Failed to load configuration: ...` |
| interval 違規（daily-only） | `markets.allowed_intervals must be exactly ["1d"]` |
| scrape 在 config 中停用 | `Scraping is disabled in configuration` |
| `scrape` mode 旗標缺漏 | `either --check, --preview-json, --preview-news, or --preview-proto flag is required` |
| `--endpoint` 不在白名單 | `--endpoint must be one of: [profile key-statistics financials balance-sheet cash-flow analysis analyst-insights news]` |
| observability init 失敗 | `Failed to initialize observability: ...` |

**修復：**

```bash
# 列印 effective config 確認目前生效值
yfin --config config/production.yaml config --print-effective
```

```yaml
# config/production.yaml 範例
scrape:
  qps: 1.0
  burst: 3
  retry:
    base_ms: 2000
    max_delay_ms: 30000
markets:
  allowed_intervals: ["1d"]
  default_adjustment_policy: split_dividend
observability:
  metrics:
    prometheus:
      enabled: true
      addr: ":9090"
```

**典型觸發情境：**

- Kafka / NATS broker 不可達
- 本地磁碟空間不足
- broken pipe（stdout 被 SIGPIPE 中斷，例如 `yfin ... | head`）
- parquet 寫入尚未實作（`--out parquet` 時回 `parquet export not implemented yet`）

## Scrape-specific Errors

由 `svc/scrape/errors.go` 集中管理 `ScrapeError` 型別與預設 sentinel。

### Parse Errors

| 錯誤常數 | 觸發原因 | 症狀 |
| -------- | -------- | ---- |
| `ErrNoQuoteSummary` | quote summary script payload 找不到 | 多個 ticker 同時失敗；HTTP 200 但解析失敗 |
| `ErrJSONUnescape` | embedded JSON 含錯誤 escape sequence | 部分 ticker 失敗、JSON parsing 階段報錯 |
| `ErrJSONDecode` | JSON 結構解碼失敗 | 結構變動、欄位型別不符 |
| `ErrMissingFieldBase` | 必要欄位缺失 | 對應 DTO 無法填值 |
| `ErrSchemaDriftBase` | 預期欄位型別變動 | `schema validation failed: ... expected number, got string` |

#### `ErrNoQuoteSummary` 修復流程

```bash
# 1. 確認問題是否廣泛
yfin scrape --ticker AAPL --endpoint key-statistics --check
yfin scrape --ticker MSFT --endpoint key-statistics --check

# 2. debug logging + 觀察 HTML 結構
yfin --log-level debug scrape --ticker AAPL --endpoint key-statistics --preview-json --endpoints key-statistics

# 3. 取得抽取結果比對缺失欄位
yfin scrape --ticker AAPL --endpoints key-statistics --preview-json
```

修復步驟：

1. **即時：** 切換到仍可用的 endpoint（`profile` / `financials`）。
2. **完整：** 至 `svc/scrape/` 對應 extractor（如 `key-statistics`）更新 selector / regex，並加上 fallback selector。
3. **驗證：** 跨 ticker 跑 `--preview-json` 確認。

#### `ErrJSONUnescape` 修復流程

```bash
yfin scrape --ticker "BRK-A" --endpoints financials --preview-json
yfin scrape --ticker "0700.HK" --endpoints financials --preview-json
yfin scrape --ticker "BRK.B" --endpoints financials --preview-json
```

修復：加強 `\\u[0-9a-fA-F]{4}` 反斜線處理與 unicode unescape 邏輯。

#### `ErrSchemaDriftBase` 修復流程

```bash
yfin scrape --ticker AAPL --endpoints key-statistics --preview-proto
yfin scrape --ticker MSFT --endpoints key-statistics --preview-proto
```

修復：於 `svc/emit/` 對應 mapper（如 `map_financials.go`）更新欄位處理，覆蓋 `string` / `float64` / `object` 等型別變化。

### Robots.txt 與合規錯誤

`ErrRobotsDenied` — Yahoo Finance 更新 robots.txt 後禁止存取特定路徑。

**症狀：** 原本能用的 endpoint 突然失敗；同一 endpoint 對所有 ticker 都失敗；錯誤在 HTTP 請求送出前觸發。

```bash
# 1. 確認現行 robots.txt
curl -s https://finance.yahoo.com/robots.txt | grep -A 10 -B 10 "key-statistics"

# 2. 測試其他 endpoint
yfin scrape --ticker AAPL --endpoint profile --check
yfin scrape --ticker AAPL --endpoint news --check
```

修復：

```yaml
scrape:
  # 使用較通用的 user-agent（不要偽裝瀏覽器）
  user_agent: "yfin/1.x (contact@example.com)"
  # 開發期可暫時改為 warn
  robots_policy: warn
```

> **CRITICAL WARNING：** 絕對不要在生產環境使用 `--force` 或 `robots_policy: "ignore"`。這違反 Yahoo Finance 服務條款，可能導致 IP 被封鎖。

### Rate Limit 處理

`ErrRateLimited` — 持續高錯誤率（>5%）或暫時恢復後又出現錯誤。

```bash
# 1. 列印 effective config
yfin --config config/production.yaml config --print-effective

# 2. 用較低 QPS 測試
yfin --qps 0.5 scrape --ticker AAPL --endpoint key-statistics --check

# 3. 透過 --qps / --concurrency 旗標覆寫
yfin --qps 1.0 --concurrency 4 scrape --ticker AAPL --endpoint key-statistics --check
```

修復（YAML）：

```yaml
scrape:
  qps: 1.0      # 從 2.0 調降
  burst: 3      # 從 5 調降
  retry:
    base_ms: 2000
    max_delay_ms: 30000
circuit_breaker:
  window: 50
  failure_threshold: 0.30
  reset_timeout_ms: 30000
```

監控閾值：

- 429 / 503 rate > 5% 持續 5 分鐘 → warning
- 429 / 503 rate > 10% 持續 2 分鐘 → critical

### 資料品質錯誤

**News 去重問題：**

```bash
yfin scrape --ticker AAPL --preview-news
```

修復：於 `svc/scrape/extract_news.go` 使用複合 key（`url|title`）做為去重依據，並透過 `scrape.endpoints.news` 與 `cache_ttl_ms` 調整抓取行為。

**Time normalization caveats：**

Yahoo Finance 使用多種日期格式（`Q3 2023` 等），於 `svc/scrape/` 對應 time parser（`time.go`）加入 quarterly / yearly pattern。

## HTTP Layer Errors

由 `utils/httpx/errors.go` 集中管理 sentinel errors + 分類器 (classifier)。

### Sentinel Errors

| 錯誤常數 | 意義 | Retryable? |
| -------- | ---- | ---------- |
| `ErrTooManyRequests` | 429 速率限制 | 是 |
| `ErrServerUnavailable` | 5xx 伺服器錯誤 | 是 |
| `ErrDecode` | 解碼失敗 | 否 (fatal) |
| `ErrClientConfig` | 客戶端配置錯誤 | 否 (fatal) |
| `ErrRateLimited` | 通用 rate limit | 視分類器 |
| `ErrCircuitOpen` | 斷路器 (circuit breaker) 已開啟 | 否 (應等待 reset) |
| `ErrTimeout` | 請求逾時 | 是 |
| `ErrContextCanceled` | context 取消 | 視情境 |
| `ErrMiddleware` | middleware 拒絕請求 | 否 (fatal) |

### 分類器

- `httpx.IsRetryableError(err)` — HTTP 429 / 5xx、`ErrTooManyRequests`、`ErrServerUnavailable`、`ErrTimeout` 為 retryable。
- `httpx.IsFatalError(err)` — HTTP 400 / 401 / 403 / 404 / 422、`ErrClientConfig`、`ErrDecode` 為 fatal（不重試）。

### 斷路器 (Circuit Breaker)

當連續失敗達閾值，斷路器進入 open 狀態並 short-circuit 後續請求直到 `reset_timeout_ms` 屆滿。當下行為：

```text
circuit breaker is open
```

**處置：** 降低上游 QPS、放大 `circuit_breaker.reset_timeout_ms`、等待 reset 週期。

### 重試 (Retry) 與 Backoff

`utils/httpx` 內建 exponential backoff + jitter：

- base delay 由 `scrape.retry.base_ms` 起始
- max delay 上限 `scrape.retry.max_delay_ms`
- 最大嘗試次數 `scrape.retry.attempts`（CLI 覆寫為 `--retry-max`）

## Common Scenarios & Remediation

### 情境 1：「no news articles found」

**成因：** Yahoo Finance 對該 symbol 無新聞資料，或去重後為空。

**處置：** 視為正常結果（graceful degradation），不視為 fatal error。

```bash
yfin scrape --ticker AAPL --preview-news
```

輸出範例：

```
AAPL news: found=12 deduped=2 returned=10 as_of=2026-07-13T...
```

### 情境 2：Rate Limiting（429 / 503 上升）

**處置：**

```bash
# 立即降速
yfin --qps 0.5 --concurrency 2 scrape --ticker AAPL --endpoint key-statistics --check

# 提高 retry backoff
# 在 config/production.yaml：
# scrape.retry.base_ms: 2000
# scrape.retry.max_delay_ms: 30000
```

### 情境 3：FetchFundamentalsQuarterly 需付費

**錯誤訊息：** `paid subscription required` / `401 Unauthorized`。

**退出碼：** `ExitPaidFeature` (2)。

**處置：** 改走 scrape fallback：

```bash
yfin scrape --preview-proto --ticker AAPL --endpoints financials,key-statistics,analysis,profile
```

### 情境 4：空公司資訊

`FetchCompanyInfo` 只回傳基本 security 欄位；如需詳細 profile 改用 `yfin scrape --preview-proto --endpoints profile`。

### 情境 5：欄位命名 / ScaledDecimal 處理

價格與金額以 `ScaledDecimal` 內部儲存以保留精度；透過 `facade` 暴露為 `float64`。**必須**檢查 nil 指標：

```go
if quote.RegularMarketPrice != nil {
    price := float64(quote.RegularMarketPrice.Scaled) /
        float64(quote.RegularMarketPrice.Scale)
    _ = price
}
```

> 註：session rotation 已移除；HTTP client 由 `utils/httpx` 統一管理連線復用、QPS 與斷路器。`--sessions` 旗標雖仍可讀取但已無作用（vestigial）。

## Debugging Toolkit

### 全域旗標（持久）

| 旗標 | 用途 |
| ---- | ---- |
| `--log-level debug` | 開啟 debug 等級日誌 |
| `--qps N` | 覆寫 per-host QPS |
| `--concurrency N` | 覆寫 worker pool 大小 |
| `--retry-max N` | 覆寫 HTTP retry attempts |
| `--timeout 6s` | 覆寫 HTTP timeout |
| `--sessions N` | vestigial；不再生效 |
| `--config path` | 指定 YAML 配置 |
| `--observability-disable-tracing` | 停用 OpenTelemetry tracing |
| `--observability-disable-metrics` | 停用 Prometheus metrics |
| `--run-id` | 自訂 run ID |

### Scrape 模式（互斥）

| 模式 | 旗標組合 | 用途 |
| ---- | -------- | ---- |
| `--check` | `--ticker X --endpoint Y` | 連線測試（不下載、不解析） |
| `--preview-json` | `--ticker X --endpoints a,b,c` | extractor 乾跑（解析但不送 proto） |
| `--preview-news` | `--ticker X` | news parser 乾跑 |
| `--preview-proto` | `--ticker X --endpoints a,b,c` | proto 完整輸出乾跑 |

### 標準除錯流程

```bash
# Step 1：連線測試
yfin scrape --ticker AAPL --endpoint key-statistics --check

# Step 2：debug logging
yfin --log-level debug scrape --ticker AAPL --endpoint key-statistics --check

# Step 3：列印 effective config
yfin --config config/production.yaml config --print-effective

# Step 4：隔離問題
yfin scrape --ticker AAPL --endpoint key-statistics --check
yfin scrape --ticker AAPL --endpoints key-statistics --preview-json
yfin scrape --ticker AAPL --endpoints key-statistics --preview-proto

# Step 5：蒐集證據
yfin --log-level debug scrape --ticker AAPL --endpoints key-statistics --preview-json 2>&1 | tee debug.log

# Step 6：跨 ticker 驗證
for t in AAPL MSFT GOOGL; do
  yfin scrape --ticker "$t" --endpoint key-statistics --check
done
```

### 預防最佳實踐 (Prevention Best Practices)

- Parse 成功率 > 95%
- Rate limit 錯誤 < 1%
- 回應時間 P95 < 3s
- Schema 驗證成功率 > 99%

定期健康檢查範例：

```bash
#!/bin/bash
ENDPOINTS="profile key-statistics financials balance-sheet cash-flow analysis analyst-insights news"
TICKER="AAPL"

for endpoint in $ENDPOINTS; do
  yfin scrape --ticker "$TICKER" --endpoint "$endpoint" --check
done
```

## See also

- `docs/operations/configuration.md` — YAML 配置結構
- `docs/operations/observability.md` — tracing / metrics / logs
- `docs/cli/usage.md` — CLI 全域使用指南
- `docs/scrape/cli.md` — scrape 子命令詳解
- `docs/scrape/config.md` — scrape config 細節
- `docs/scrape/overview.md` — scrape 系統概觀
- `cmd/exitcodes.go` — exit code 常數定義
- `utils/httpx/errors.go` — HTTP sentinel errors
- `svc/scrape/errors.go` — scrape sentinel errors