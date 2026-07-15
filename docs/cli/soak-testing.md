# Soak Testing Guide

`cmd/soak` 為獨立的 soak binary（非 `yfin` CLI 的一部分），用於在 production-like 條件下驗證穩定性、吞吐量與容錯能力。

## 重點摘要

- 與 `yfin` CLI 完全解耦：執行方式為 `go run ./cmd/soak ...` 或先 `go build -o soak ./cmd/soak` 後 `./soak ...`
- 覆蓋 API → fallback → 規範化 → 發佈 的端到端流程
- 內建失敗注入、robots.txt 守查與正確性探測
- Session 輪替功能已移除，HTTP client 改為共用 `http.Client` 搭配 rate-limit / retry / circuit breaker

## 快速開始

### 基本冒煙測試（10 分鐘）

```bash
go run ./cmd/soak \
  --config config/effective.yaml \
  --universe-file tests/testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --fallback auto \
  --duration 10m \
  --concurrency 8 \
  --qps 5 \
  --preview
```

### 完整 production soak（2 小時）

```bash
go run ./cmd/soak \
  --config config/effective.yaml \
  --universe-file tests/testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --fallback auto \
  --duration 2h \
  --concurrency 12 \
  --qps 5 \
  --preview
```

### 長時段測試

```bash
go run ./cmd/soak \
  --config config/effective.yaml \
  --universe-file tests/testdata/universe/soak.txt \
  --endpoints news \
  --fallback auto \
  --duration 30m \
  --concurrency 8 \
  --qps 3
```

## 設定旗標

旗標以 `cmd/soak/main.go` 的 `rootCmd.PersistentFlags()` 與 `rootCmd.Flags()` 為準。

### 全域旗標（Persistent）

| 旗標 | 預設值 | 說明 |
| --- | --- | --- |
| `--config` | `""` | YAML 設定檔路徑（選填） |
| `--log-level` | `info` | info \| debug \| warn \| error |
| `--run-id` | `""` | 識別本次 run 的 ID；留空自動產生 |
| `--concurrency` | `0` | worker pool 大小（留空由 config 決定） |
| `--qps` | `0` | per-host QPS（留空由 config 決定） |
| `--retry-max` | `0` | HTTP 重試次數（留空由 config 決定） |
| `--timeout` | `0` | HTTP timeout（例：`6s`） |

### Soak 旗標

| 旗標 | 預設值 | 說明 |
| --- | --- | --- |
| `--universe-file` | *必填* | 包含 ticker 清單的檔案路徑 |
| `--endpoints` | `key-statistics,financials,analysis,profile,news` | 逗號分隔的 endpoint 清單 |
| `--fallback` | `auto` | `auto` \| `api-only` \| `scrape-only` |
| `--duration` | `2h` | 測試持續時間 |
| `--concurrency` | `12` | 並行 worker 數 |
| `--qps` | `5.0` | 目標每秒查詢數 |
| `--probe-interval` | `1h` | 正確性探測週期間隔 |
| `--failure-rate` | `0.1` | 模擬失敗率（0.0–1.0） |
| `--memory-check` | `true` | 啟用記憶體與 goroutine 洩漏偵測 |

> 注意：CLI 的 `--qps` / `--retry-max` / `--timeout` 與 nested config 結構尚未在 orchestrator override hook 串接，未來會以 `cfg.HTTP()` / `cfg.RateLimit.PerHostQPS` / `cfg.Retry.Attempts` 等 setter 套用。

## Ticker Universe

`ticker` 名單位於 `tests/testdata/universe/soak.txt`，目前共 82 個 ticker，橫跨多種市場與產業：

- `US Large Cap`：AAPL, MSFT, GOOGL, AMZN, META, NVDA, TSLA …
- `US Mid/Small Cap`：ROKU, PLTR, RBLX, COIN, SNOW …
- `International`：ASML, SAP, 005930.KS, 0700.HK …
- `ADRs`：NIO, LI, XPEV, PDD, JD …
- `Specialized`：REITs、Crypto-related、Healthcare、Financial Services

## 支援的 Endpoint

### API + scrape fallback
- `quote` 即時報價
- `daily-bars` 日線歷史價
- `fundamentals` 季度基本面（需付費）

### Scrape only
- `key-statistics`、`financials`、`balance-sheet`、`cash-flow`
- `analysis`、`analyst-insights`、`profile`、`news`

## Fallback 策略

### `auto`（預設）
- 先走 API，遇 `401` / `429` / `5xx` 退回 scrape
- scrape-only endpoint 一律走 scrape

### `api-only`
- 僅 API；不可用即 fail，適合驗 API 可靠度

### `scrape-only`
- 僅 scrape；適合獨立驗證 scraping 與 robots.txt 守查

## 觀測與診斷

### 即時指標
- 請求成功 / 失敗率
- 各 endpoint 延遲直方圖
- fallback 決策計數
- rate-limit 與 robots.txt 阻擋計數
- 記憶體用量與 goroutine 數

### 正確性探測
週期性比對 API vs scrape：
- `Market Cap`：容差 5%
- `P/E Ratio`：容差 10%
- `Employee Count`：容差 15%
- `Sector`、`Currency` 一致性

### 記憶體洩漏偵測
- heap / goroutine 持續監控
- 成長率分析
- GC 行為分析

### 失敗注入
內建 `FailureServer`（`cmd/soak/failure.go`）模擬：
- 429 rate-limiting
- 5xx server errors
- 401 auth failures
- connection timeouts
- bad gateway

## 輸出範例

```
=== SOAK TEST RESULTS ===
Duration: 2h0m15s
Total Requests: 7,234
Successful Requests: 7,198
Failed Requests: 36
Success Rate: 99.50%
Actual QPS: 1.00
API Requests: 2,156
Scrape Requests: 5,078
Fallback Decisions: 234
Rate Limit Hits: 12
Robots Blocked: 0
Correctness Probes Passed: 24
Correctness Probes Failed: 1

=== MEMORY ANALYSIS ===
Initial Memory: 2,313 KB
Peak Memory: 8,456 KB
Final Memory: 3,127 KB
Memory Growth: 814 KB
Initial Goroutines: 8
Peak Goroutines: 24
Final Goroutines: 12
Goroutine Growth: 4

=== ENDPOINT BREAKDOWN ===
key-statistics: 1,445 requests, 1,440 successes, 5 failures, avg latency: 1.2s
financials:     1,434 requests, 1,429 successes, 5 failures, avg latency: 1.8s
analysis:       1,445 requests, 1,441 successes, 4 failures, avg latency: 1.1s
profile:        1,456 requests, 1,444 successes, 12 failures, avg latency: 2.1s
news:           1,454 requests, 1,444 successes, 10 failures, avg latency: 1.5s
```

## 成功標準

- `Zero Memory Leaks`：warmup 後 heap / GC 穩定在 ±10%
- `Rate Limit Safety`：429/503 持續 < 1%
- `Robots Adherence`：enforce 模式下無 disallowed fetch
- `Correctness`：API vs scrape 容差內一致
- `No DLQ`：發佈乾淨且順序正確

## 疑難排解

### 失敗率偏高
- 檢查網路
- 確認 robots.txt 守查
- 降低 QPS
- 確認 Yahoo Finance 服務狀態

### 記憶體洩漏
- 檢視 goroutine lifecycle
- 確認 channel 正常關閉
- 確認 context cancel 傳遞
- 觀察物件留存模式

### 探測失敗
- 時差造成（多數情況）
- 匯率換算差異
- 報告週期不對齊
- 資料源更新延遲

### Rate limiting
- 調降 `--qps`
- 提高 `--concurrency` 分散流量
- 確認 backoff 行為出現在 log 中

## CI/CD 整合

### 每晚快速驗證（30 分鐘）

```bash
go run ./cmd/soak \
  --config config/ci.yaml \
  --universe-file tests/testdata/universe/soak.txt \
  --endpoints key-statistics,news \
  --duration 30m \
  --concurrency 4 \
  --qps 2 \
  --preview
```

### 上線前完整驗證（1 小時）

```bash
go run ./cmd/soak \
  --config config/staging.yaml \
  --universe-file tests/testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --duration 1h \
  --concurrency 8 \
  --qps 3 \
  --preview \
  --memory-check
```

## 最佳實踐

1. `Start Small`：短時、低 QPS 起步
2. `Monitor Resources`：留意 memory / goroutine 成長
3. `Validate Correctness`：定期審視探測結果
4. `Respect Rate Limits`：QPS 維持 < 10
5. `Use Preview Mode`：先無發佈測試
6. `Check Logs`：監看 robots.txt 違規
7. `Gradual Scaling`：循序漸進加壓

## 架構

`cmd/soak` 由下列元件組成（位於 `cmd/soak/`）：

- `Orchestrator` (`orchestrator.go`)：整體協調
- `Worker` (`worker.go`)：隨機 endpoint 派發
- `Metrics` (`metrics.go`)：Prometheus 相容指標
- `CorrectnessProbes` (`probes.go`)：API vs scrape 對齊
- `MemoryMonitor` (`memory.go`)：洩漏偵測
- `FailureServer` (`failure.go`)：失敗注入
- 共用 `rate.Limiter`：QPS 控制與 backoff

輸入 config 來自 `config/` 套件；執行期透過 `facade.Client` 與 `utils/bus` 對接匯流排。
