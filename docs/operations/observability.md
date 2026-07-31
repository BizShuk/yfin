# Observability Guide

`yfin` 目前提供 structured logging、Prometheus pull-based metrics，以及保留 API 形狀的 no-op OpenTelemetry tracer。它不會自行寫入固定 log file、push metrics，或連接 distributed tracing backend。

## Runtime Boundary

| Signal | Current implementation | Ownership boundary |
| --- | --- | --- |
| Logs | `log/slog` text handler，寫入 process stderr | 檔案、retention 與 external sink 由 deployment owner 管理 |
| Metrics | 在 configured address 暴露 Prometheus `/metrics` | Prometheus scrape / remote storage 由 deployment owner 管理 |
| Traces | OpenTelemetry no-op tracer provider | 尚未建立 OTLP exporter；設定欄位只保留 schema compatibility |

`pull`、`comprehensive-stats`、`comprehensive-profile` 與 `scrape` 會初始化 `utils/obsv` lifecycle。其他 commands 不應被假設會啟動 metrics server。

Standalone `cmd/soak` 使用自己的 JSON `slog` 與 stdout aggregate report；它不啟動 Prometheus exporter。

## Configuration

Canonical YAML section：

```yaml
observability:
  logs:
    level: info
  metrics:
    prometheus:
      enabled: true
      addr: ":9090"
  tracing:
    otlp:
      enabled: true
      endpoint: "http://localhost:4317"
      sample_ratio: 0.05
```

- `logs.level` 控制 `utils/obsv` handler level；explicit `--log-level` 與 `APP_LOG_LEVEL` 依 precedence 覆寫。
- `metrics.prometheus.enabled` 決定是否啟動 exposition server。
- `metrics.prometheus.addr` 是 local bind address，不是 remote write endpoint。
- `tracing.otlp.*` 目前只會載入與驗證；`utils/obsv` 仍安裝 no-op tracer。
- `--observability-disable-metrics` 可在支援 lifecycle 的 command 上停用 metrics server。
- `--observability-disable-tracing` 是 compatibility flag；目前 tracer 本來就是 no-op。

完整 schema 與 precedence 見 [Configuration Reference](configuration.md)。

## Metrics Endpoint

以 `pull` 為例：

```bash
yfin pull \
  --ticker AAPL \
  --start 2026-07-01 \
  --end 2026-07-30
```

Command 執行期間可從 configured address 讀取：

```bash
curl -fsS http://127.0.0.1:9090/metrics |
  rg 'yfin_(requests|retries|backoff|request_latency|backoff_sleep)'
```

`utils/httpx` 目前會記錄：

| Metric | Labels / unit | Meaning |
| --- | --- | --- |
| `yfin_requests_total` | `endpoint`, `outcome`, `code` | 每個 logical request 的 final outcome |
| `yfin_retries_total` | `endpoint`, `reason` | retry attempts |
| `yfin_backoff_total` | `endpoint`, `reason` | backoff events |
| `yfin_decode_fail_total` | `reason` | decode failures |
| `yfin_cb_open_total` | `scope` | circuit open transitions |
| `yfin_cb_state` | `scope` | current circuit state |
| `yfin_inflight_requests` | `endpoint` | current inflight requests |
| `yfin_request_latency_ms` | `endpoint`, milliseconds | logical request latency |
| `yfin_backoff_sleep_ms` | `endpoint`, milliseconds | retry sleep duration |

News parsing lazily registers `yfin_scrape_news_total` 與 `yfin_scrape_news_parse_latency_ms`。Metric 未出現在 exposition output 不一定代表故障；labelled metric 通常要先被 observe 才會出現。

## Logging

`utils/obsv.Init` 建立 `slog` text handler，固定附加 `service` 與 `env` attributes。Explicit `--log-level` 也會更新 process default `slog`，讓 gosdk 或其他使用 default logger 的程式碼採用相同最低 level。

Log level 只控制已存在的 structured log emission，不會自行產生每個 HTTP request 的 debug event；request outcome、retry 與 backoff 的完整 machine-readable evidence 目前由 Prometheus metrics 擁有。

`./run.sh` 會建立 `~/.config/yfin/logs/`，但 yfin logger 不會自行寫入該目錄。若由 PM2、container runtime 或 shell redirect 執行，請由 deployment owner 將 stderr 指向該目錄並設定 rotation。

Debug example：

```bash
yfin --log-level debug scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics
```

分享 log 前應移除 response payload、cookies、tokens 與 operator-specific paths。

## Tracing Status

`StartRunSpan`、`StartIngestFetchSpan` 等 helpers 仍提供 stable call surface，但目前回傳 non-recording spans。以下能力尚未存在：

- OTLP network export
- trace sampling
- Jaeger / Tempo integration
- DNS、TCP 或 parser child-span auto-instrumentation

因此，OTLP endpoint 可通過 config validation 不代表 collector 已收到 traces。需要 real tracing 時，應先在 `utils/obsv` 建立 exporter lifecycle、shutdown semantics 與 integration tests，再更新本文件。

## Operational Checks

先驗證 config：

```bash
yfin config --print-effective |
  rg 'observability\\.(logs|metrics|tracing)'
```

若 command 回報 metrics bind failure：

1. 檢查 `observability.metrics.prometheus.addr`。
2. 確認 port 沒有被其他 process 使用。
3. 暫時加上 `--observability-disable-metrics`，判斷 data path 是否可獨立運作。
4. 不要把 `/metrics` 可連線誤當成 provider fetch 或 parser acceptance。

相關 incident 流程見 [Incident Response Playbook](../../monitoring/runbooks/incident-playbook.md)。
