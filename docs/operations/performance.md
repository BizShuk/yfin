# Performance Tuning

本指南只描述目前可由 runtime 與 tests 驗證的 tuning surfaces。不要把固定 latency 或 throughput 數字視為契約；Yahoo Finance、TWSE、network path 與 response size 都會改變實測結果。

## Tuning Model

yfin 的 request pressure 由兩層控制：

1. `cmd/market` 的 bounded worker pool 控制同時處理的 symbols。
2. `utils/httpx` 的 per-host token bucket 控制送往每個 hostname 的 QPS 與 burst。

HTTP retries、backoff 與 circuit breaker 在同一個 shared `http.Client` / connection pool 上運作。Yahoo query hosts、Yahoo web host 與 cookie bootstrap host 使用各自的 immutable `HostProfile`；circuit breaker identity 則按 authority 加 optional logical group 隔離。

## CLI Overrides

Global flags 可在單次執行時覆寫 loaded YAML：

```bash
yfin \
  --concurrency 8 \
  --qps 1.5 \
  --retry-max 3 \
  --timeout 20s \
  pull \
  --universe-file symbols.txt \
  --start 2026-01-01 \
  --end 2026-01-31
```

- `--concurrency` 只由 multi-symbol `pull` / `quote` orchestration 消費。
- Effective worker count 是 CLI/YAML global limit、YAML per-host limit 與 item count 的最小值。
- `--qps`、`--retry-max`、`--timeout` 套用到 Yahoo API 與 configured scrape HTTP profile。
- `--sessions` 是無 runtime effect 的 compatibility flag。

單一 request 很慢時，增加 concurrency 不會縮短該 request；它只增加可同時處理的 items。遇到 `429`、`503` 或 circuit open，先降低 QPS / concurrency，再考慮增加 retry。

## YAML Ownership

API workload 的設定位於 `config/effective.yaml`：

- `yahoo.*`：base URL、timeout、idle timeout、per-host connection cap、user agent。
- `concurrency.*`：global / per-host worker limits。
- `rate_limit.*`：per-host QPS / burst。
- `retry.*`：attempts 與 backoff range。
- `circuit_breaker.*`：window、failure rate、minimum requests、reset timeout。

Scrape HTTP defaults 由 `utils/httpx.DefaultScrapeConfig()` 單一擁有，不在 YAML 複製一份。Go consumer 若需自訂，先修改 loaded `cfg.Scrape.HTTP`，再呼叫 `facade.Client.ConfigureScrape`。

## SDK Configuration

SDK consumer 可用 `facade.NewClientWithConfig(*httpx.Config)` 調整 API transport：

```go
cfg := httpx.DefaultConfig()
cfg.QPS = 2
cfg.Burst = 2
cfg.MaxAttempts = 3
cfg.Timeout = 20 * time.Second

client := facade.NewClientWithConfig(cfg)
```

同一個 `facade.Client` 應跨 requests 重用，才能共享 connection pool、cookie jar、rate limiter、circuit breakers 與 MIC cache。不要為每個 symbol 建立新 client。

`httpx.Config.WrapTransport` 是 process-local instrumentation / deterministic fault-injection hook；它不從 YAML 載入，也不應用來繞過 rate limit 或 provider policy。

## Caching

Cache optimization 必須走 facade boundary：

- `facade.NewCachedClient(rawRoot)`：daily bars request-identity read-through cache。
- `facade.NewCachedTwseClient(rawRoot)`：TWSE read-through cache。
- `facade.NewCachedFoundationClient(rawRoot)`：30-command foundation cache。
- `facade.NewRawCache(rawRoot)`：legacy artifact read boundary。

Consumer 提供 raw root；yfin 擁有 root confinement、request identity、freshness policy 與 atomic JSON persistence。不要在 downstream 重複實作 transport 或 raw persistence。

CLI `batch` 使用 daily / weekly / monthly / quarterly / annual refresh tiers，`--force` 只用於明確 bypass freshness。

## Measurement

啟用 Prometheus exposition 時，以 active metrics 判斷 bottleneck：

| Metric | 用途 |
| --- | --- |
| `yfin_requests_total` | Logical request outcome 與 status code |
| `yfin_request_latency_ms` | Endpoint latency distribution |
| `yfin_retries_total` | Retry pressure |
| `yfin_backoff_total` | Backoff frequency |
| `yfin_backoff_sleep_ms` | Backoff delay |
| `yfin_inflight_requests` | In-flight requests |
| `yfin_cb_open_total` | Circuit opens |
| `yfin_cb_state` | Circuit state |
| `yfin_decode_fail_total` | Decode failures |

完整 exporter boundary 見 [Observability](observability.md)。`cmd/soak` 自行輸出 run summary 與 memory analysis，不啟動另一套 Prometheus exporter。

## Safe Tuning Sequence

1. 以一個 symbol 跑通，記錄 endpoint、status、latency 與 retries。
2. 擴到小型 universe，先固定 QPS，再逐步增加 worker limit。
3. 若 retries / backoff / circuit opens 上升，降低 request pressure。
4. 若 in-flight 偏低且沒有 upstream errors，再提高 concurrency 或 QPS。
5. 對 long-running workload 使用 standalone [Soak Testing](../cli/soak-testing.md)，並把 deterministic failure injection 與 live upstream result 分開記錄。

## Verification

修改 performance-related code 後至少執行：

```bash
go test ./utils/httpx ./cmd/market ./facade
go test -tags=integration ./... -run '^$'
go vet ./...
```

需要 live evidence 時另行執行 integration / soak，且不要把 upstream failure 當成 deterministic test failure。

## See Also

- [Configuration](configuration.md)
- [Observability](observability.md)
- [Error Handling](error-handling.md)
- [Soak Testing](../cli/soak-testing.md)
