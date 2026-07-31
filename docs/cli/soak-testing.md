# Soak Testing Guide

`cmd/soak` 是獨立的 live-test binary，不是 `yfin` subcommand。它在 shared `facade.Client` 上長時間執行 API/scrape requests，觀察 retry、rate limit、circuit breaker、robots policy、memory 與 goroutine 狀態。

Soak 會連線外部 providers；deterministic `go test ./...` 不會自動執行 live soak。

## 快速開始

```bash
go run ./cmd/soak \
  --config config/effective.yaml \
  --universe-file tests/testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --fallback auto \
  --duration 10m \
  --concurrency 4 \
  --qps 1 \
  --failure-rate 0
```

先以低 QPS、短 duration 和 `--failure-rate 0` 建立 upstream baseline，再逐步增加 load 或 failure injection。

## Runtime Flags

| Flag | Default | Contract |
| --- | --- | --- |
| `--config` | `""` | 空白時載入 `config/effective.yaml` |
| `--log-level` | `info` | `debug` / `info` / `warn` / `error` |
| `--run-id` | generated | 整個 run 共用同一 ID |
| `--universe-file` | required | newline-delimited ticker list |
| `--endpoints` | scrape surface list | comma-separated endpoints |
| `--fallback` | `auto` | `auto` / `api-only` / `scrape-only` |
| `--duration` | `2h` | 必須大於 0 |
| `--concurrency` | `12` | worker count，必須大於 0 |
| `--qps` | `5` | outer worker limiter 與 Yahoo host profiles 的 QPS |
| `--probe-interval` | `1h` | correctness probe interval |
| `--failure-rate` | `0.1` | shared HTTP transport 的 injected 503 比例，範圍 `0..1` |
| `--memory-check` | `true` | 啟用 memory/goroutine samples |
| `--retry-max` | loaded config | 非零時覆寫 shared HTTP max attempts |
| `--timeout` | loaded config | 非零時覆寫 shared HTTP timeout |

CLI overrides 會套在已載入的 typed config clone，不會修改 loader 回傳的 config snapshot。

## Endpoint and Strategy Matrix

API endpoints：

- `quote`
- `daily-bars`
- `fundamentals`

Scrape endpoints：

- `key-statistics`
- `financials`
- `balance-sheet`
- `cash-flow`
- `analysis`
- `analyst-insights`
- `profile`
- `news`

Strategy contract：

| Strategy | Behavior |
| --- | --- |
| `api-only` | 只接受 API endpoints；API error 直接保留 |
| `scrape-only` | 只接受 scrape endpoints |
| `auto` | API endpoints 先走 API；scrape-only endpoints 直接走 scrape |

`auto` 並不代表任意 endpoint 都可互換。唯一有明確 equivalent mapping 的 fallback 是：

```text
fundamentals API --retryable failure--> financials scrape
```

`quote` 與 `daily-bars` 沒有 scrape equivalent；遇到錯誤時不會改名 dispatch 到不存在的 scrape surface。

## Failure Injection

`--failure-rate` 裝飾 shared `http.RoundTripper`。被抽中的 request 會在真正的 request path 回傳 synthetic `503 Service Unavailable`，因此既有 retry/backoff/circuit-breaker logic 會實際處理它。

Failure injection：

- 不開固定 port。
- 不啟動另一個 disconnected HTTP server。
- 不回傳假的成功 payload。
- `0` 完全 delegate 到 upstream transport。
- `1` 對每次 transport attempt 注入 retryable 503。

## Correctness Probes

Probe round 對抽樣 ticker 讀取真實資料：

- Yahoo API `info` map。
- Scraped comprehensive key statistics。
- Scraped comprehensive financials（currency source）。
- Scraped comprehensive profile。

比較項目：

| Probe | Tolerance |
| --- | --- |
| Market cap | 5% |
| Trailing P/E；缺值時改比 forward P/E | 10% |
| Full-time employees | 15% |
| Sector | normalized exact match |
| Currency | case-insensitive exact match |

缺值會記錄為 probe failure；probe 不會填入固定 market cap、P/E、employee、sector 或 currency。

## Memory and Output

啟用 `--memory-check` 時，runner 取樣：

- allocated/system memory
- heap objects
- GC cycles
- goroutine count

Run 結束後 stdout 顯示 aggregate request counts、API/scrape attempts、fallback decisions、probe results、memory state 與 per-endpoint latency。Structured runtime logs 由 stdlib `slog` 寫到 stderr。

Soak runner 不建立獨立 Prometheus exporter；HTTP client 的通用 metrics ownership 請見 [Observability Guide](../operations/observability.md)。

## Acceptance Boundary

Soak result 會受 Yahoo/TWSE availability、network、robots policy 與 ticker data availability 影響。不要把一次 local pass 當成永久 SLA。

建議分開保存：

1. deterministic build/test evidence；
2. soak invocation 與 exact config；
3. run ID、開始/結束時間；
4. upstream failures 與 injected failures；
5. operator-defined acceptance thresholds。

## Troubleshooting

- `429` / repeated 5xx：降低 `--qps` 與 `--concurrency`，檢查 retry logs。
- robots denied：確認 `scrape.robots_policy`；production 不應用 `ignore` 繞過。
- invalid endpoint strategy：依 matrix 修正 `--fallback` 或 endpoint list。
- probe mismatch：先確認兩個 source 的 reporting date 與缺值，不要加入 fallback constants。
- memory growth：延長 observation window，區分 warmup、GC cycle 與持續 growth。

Package ownership 與 standalone binary placement 由 [CLAUDE.md](../../CLAUDE.md) 單一擁有。
