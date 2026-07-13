# Plan: 移除 `ampy-observability` 與 `ampy-config` 依賴

## TL;DR

兩個依賴的「接觸面」其實非常窄,而且彼此隔離:

- **`ampy-observability`** 的唯一直接接觸面是 `utils/obsv/obsv.go` 一個檔案,只有 5 處外部 caller(`utils/httpx/client.go` + 4 個 cmd entrypoint)。
- **`ampy-config`** 的唯一直接接觸面是 `config/types/loader.go` 兩個函式(`Load` / `GetEffectiveConfig`),且 ampy-config 本身只是 `os.ReadFile + yaml.Unmarshal → map[string]any`,**不做任何 secret 解析、檔案合併或 hot-reload**(從原始碼可證實)。
- `bizshuk/gosdk` v1.1.0 **不**依賴這兩個 module(已從 module cache grep 確認)。
- `ampy-bus` v1.2.0 / `ampy-proto/v2` v2.1.1 **不**依賴 ampy-obs(但 ampy-obs 自己 require ampy-bus 與 ampy-proto,所以兩者無傳遞衝突)。

`utils/obsv/metrics.go` 是本地 Prometheus metrics,**完全獨立於 ampy-observability**(只 import `prometheus/client_golang`);`ampyobs` 本身**不提供** Prometheus exporter(`grep prometheus` 在 ampyobs module 0 hits),所以刪除 ampy-obs 後本地 metrics 可原封不動保留。

---

## A. `ampy-observability` 移除面

### A.1 `utils/obsv/` 內檔案 + 匯出符號 + ampyobs.* 接觸點

#### `/Users/bytedance/projects/yfin/utils/obsv/obsv.go`
Import: `"github.com/AmpyFin/ampy-observability/go/ampyobs"`(L11)

**ampyobs.* 直接呼叫點**:
- `L70` `ampyobs.Init(ampyConfig)` — `Init` 路徑
- `L108` `ampyobs.Shutdown(ctx)` — `Shutdown` 路徑
- `L128` `ampyobs.L()` — `Logger` getter(回傳 `*slog.Logger`)
- `L146` `ampyobs.StartSpan(ctx, name, trace.SpanKindInternal)` — Span 建立

**匯出符號**:
- `Config` struct (L19-30):`ServiceName / ServiceVersion / Environment / CollectorEndpoint / TraceProtocol / SampleRatio / LogLevel / MetricsAddr / MetricsEnabled / TracingEnabled`
- `PrometheusConfig` struct (L33-36):`Enabled / Addr`
- `Observability` struct (L39-43):內部型別,**未真正對外暴露**給 caller
- `Init(ctx, *Config) error` (L52)
- `Shutdown(ctx) error` (L95)
- `Reset()` (L114,測試用)
- `Logger() *slog.Logger` (L121)— fallback 到 `slog.Default()`
- `Tracer() trace.Tracer` (L132)— fallback 到 `noop.NewTracerProvider().Tracer(...)`(L140 的註解坦承「ampy-observability doesn't expose tracer directly」,**目前其實就是 noop**)
- `StartSpan(ctx, name, opts...)` (L144)
- Span name constants (L152-160):`SpanNameRun / SpanNameIngestFetch / SpanNameIngestDecode / SpanNameIngestNormalize / SpanNameEmitProto / SpanNamePublishBus / SpanNameFXRates`
- Span factory functions:
  - `StartRunSpan(ctx, runID, env, args)` (L163)
  - `StartIngestFetchSpan(ctx, endpoint, symbol, mic, url, attempt)` (L174)
  - `UpdateIngestFetchSpan(span, status, bytes, elapsed)` (L187)
  - `StartIngestDecodeSpan(ctx, endpoint, symbol)` (L198)
  - `StartIngestNormalizeSpan(ctx, endpoint, symbol, mic)` (L208)
  - `StartEmitProtoSpan(ctx, messageType, symbol)` (L219)
  - `StartPublishBusSpan(ctx, topic, partitionKey, chunkIndex, bytes)` (L229)
  - `StartFXRatesSpan(ctx, fromCurrency, toCurrency)` (L241)
- `RecordSpanError(span, err)` (L251)
- `LogWithTrace(ctx, attrs...)` (L259)— 純 OTel `trace.SpanFromContext` 處理,**不依賴 ampyobs**
- `CommonLogAttrs(runID, symbol, mic, endpoint)` (L269)— 純屬性拼接,**不依賴 ampyobs**

#### `/Users/bytedance/projects/yfin/utils/obsv/metrics.go`
- **0 個 ampyobs 呼叫**。完全本地 Prometheus。
- 13 個 metrics + 12 個 recorders,只 import `github.com/prometheus/client_golang/prometheus{,/promhttp}`。
- 內部 `initMetrics(PrometheusConfig)` (L132) 用 `prometheus.MustRegister` + `promhttp.Handler()` 開 `/metrics` 端點。
- 內部 `shutdownMetrics(ctx)` (L177) 走 `metricsServer.Shutdown(ctx)`。
- **結論:這個檔案可原封不動保留**,完全不需替換。

#### `/Users/bytedance/projects/yfin/utils/obsv/obsv_test.go`
- 測試 `Init` / `Shutdown` / `Logger` / `Tracer` / 各 Span factory / `RecordSpanError` / `LogWithTrace` / `CommonLogAttrs` / 12 個 metric recorders。
- 全部走 `obsv` package 公開 API;若 `obsv` 換掉實作,測試可保持不動(只要公開介面不變)。

#### `/Users/bytedance/projects/yfin/utils/obsv/integration_test.go`
- 5 個整合測試,`Init` 後走完整 span hierarchy,測試 metric recorders。
- 同上,只要 `obsv` 公開介面不變,測試不需要改。

### A.2 全倉庫 `obsv.` 呼叫點(已 grep 驗證,只 5 個檔案 / 35 處)

| 檔案 | 行號 | 呼叫 |
|---|---|---|
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 126 | `obsv.StartIngestFetchSpan` |
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 131, 138, 159, 184, 204, 226, 237 | `obsv.RecordRequest` |
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 132, 139, 161, 206, 227, 239 | `obsv.RecordSpanError` |
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 155, 175 | `obsv.RecordRetry` |
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 160, 185, 205, 238 | `obsv.RecordRequestLatency` |
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 186 | `obsv.UpdateIngestFetchSpan` |
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 220 | `obsv.RecordBackoff` |
| `/Users/bytedance/projects/yfin/utils/httpx/client.go` | 221 | `obsv.RecordBackoffSleep` |
| `/Users/bytedance/projects/yfin/cmd/fundamentals/profile.go` | 69, 81, 85 | `obsv.Config`, `obsv.Init`, `obsv.Shutdown` |
| `/Users/bytedance/projects/yfin/cmd/fundamentals/stats.go` | 70, 82, 86 | `obsv.Config`, `obsv.Init`, `obsv.Shutdown` |
| `/Users/bytedance/projects/yfin/cmd/scrape/scrape_run.go` | 96, 109, 113 | `obsv.Config`, `obsv.Init`, `obsv.Shutdown` |
| `/Users/bytedance/projects/yfin/cmd/market/pull.go` | 124, 137, 141 | `obsv.Config`, `obsv.Init`, `obsv.Shutdown` |

> 已驗證: `cmd/soak/main.go`、`svc/`、`facade/`、其餘 `utils/` 子目錄**皆無** `obsv.` 呼叫。
> 已驗證: `cmd/admin/admin.go`、`cmd/global.go`、`cmd/client.go`、`cmd/soak/main.go`、`facade/scrape.go` 沒有 `obsv.*` 呼叫(只有 `obsv` 一個檔案對 `ampyobs.*` 有直接 import)。

### A.3 `utils/httpx/client.go` 對 obsv 的細節(L104-243,`Client.Do`)

呼叫順序(每一次 HTTP 請求):
1. **L126** `obsv.StartIngestFetchSpan(ctx, endpoint, "", "", req.URL.String(), 0)` — span 開始;`defer span.End()`(L127)
2. **L131-133** circuit breaker 開啟:`RecordRequest("error","circuit_open")` + `RecordSpanError`
3. **L137-141** rate limit 等待失敗:`RecordRequest("error","rate_limit")` + `RecordSpanError`
4. **L154-156** 重試計數:`RecordRetry("network_error")`
5. **L158-165** 網路錯誤中止:`RecordRequest("error","network_error")` + `RecordRequestLatency` + `RecordSpanError`
6. **L174-176** HTTP retry 計數:`RecordRetry("http_%d")`
7. **L183-186** 成功:`RecordRequest("success","%d")` + `RecordRequestLatency` + `UpdateIngestFetchSpan(status, contentLength, elapsed)`
8. **L203-206** 非 retry 失敗(4xx):`RecordRequest("error","%d")` + `RecordRequestLatency` + `RecordSpanError`
9. **L220-221** backoff:`RecordBackoff("retry")` + `RecordBackoffSleep(delay)`
10. **L225-227** context cancel:`RecordRequest("error","context_canceled")` + `RecordSpanError`
11. **L236-239** max attempts 超限:`RecordRequest("error","max_attempts")` + `RecordRequestLatency` + `RecordSpanError`

**重構方向**:
- `StartIngestFetchSpan` / `UpdateIngestFetchSpan` / `RecordSpanError` → 改用 `go.opentelemetry.io/otel/trace` 裸 API(本地 init 一個 `sdktrace.NewTracerProvider`,無 collector 時 fallback `noop`)。
- `RecordRequest*` / `RecordRetry` / `RecordBackoff*` → 保留 `obsv/metrics.go` 內已存在的 Prometheus recorder(它們已經是裸 `prometheus/client_golang`,**不需改**)。

### A.4 4 個 cmd entrypoint 的 obsv bootstrap 樣板(各約 L69-85 區段)

幾乎 100% 相同的 3 段:
1. **Config 物件** — 從 `ycfg.Observability.*` 欄位填入 `obsv.Config{...}`
2. **`obsv.Init(ctx, obsvConfig)`** — 失敗 `os.Exit(cmd.ExitConfigError)`
3. **`defer func() { _ = obsv.Shutdown(ctx) }()`** — 結束時關閉

| 檔案 | 行號區段 |
|---|---|
| `/Users/bytedance/projects/yfin/cmd/fundamentals/profile.go` | L66-85 |
| `/Users/bytedance/projects/yfin/cmd/fundamentals/stats.go` | L67-86 |
| `/Users/bytedance/projects/yfin/cmd/scrape/scrape_run.go` | L92-113 |
| `/Users/bytedance/projects/yfin/cmd/market/pull.go` | L118-141 |

**重複程式碼可抽 helper**(例如 `cmd.obsvInit(ctx, ycfg, disableTracing, disableMetrics)`),消重後 4 個檔案都精簡 15 行左右。

**重構方向**:把 `obsv.Config` 換成裸 OTel + slog,這段樣板改呼叫 `obsv.InitOtel(ctx, otelCfg)`(無外部 dep)。

### A.5 ampy-observability 是否提供 Prometheus exporter?

**否**。從 module cache grep:
- `ampyobs/metrics.go` 只有 `initMetrics() / BusProducedAdd / BusConsumedAdd / BusDeliveryLatencyMs / OMSOrderSubmitAdd / OMSOrderLatencyMs / OMSRejectAdd`(OTel `sdkmetric.MeterProvider`,不是 Prometheus)。
- `grep -i prometheus` 在 ampyobs 整個 module 0 hits。

**結論**:Prometheus exporter 完全是 `utils/obsv/metrics.go` 自行用 `prometheus/client_golang` 實作,與 ampyobs 無關。**刪除 ampy-obs 不會影響 Prometheus 端點**。

### A.6 其他套件衝突?

| 套件 | 與 ampy-obs 關係 |
|---|---|
| `github.com/bizshuk/gosdk v1.1.0` | **不依賴** ampy-obs/ampy-config(已從 module cache grep 驗證) |
| `github.com/AmpyFin/ampy-bus v1.2.0` | 不依賴 ampy-obs(其 go.mod 已查) |
| `github.com/AmpyFin/ampy-proto/v2 v2.1.1` | 不依賴 ampy-obs |
| `github.com/AmpyFin/ampy-observability v0.0.3` | **自己** require ampy-bus / ampy-config(反向) |
| `github.com/AmpyFin/ampy-config v1.1.4` | require nats.io/nats.go + gopkg.in/yaml.v3(已查 go.mod) |

**無衝突**;唯一連帶效果是刪 ampy-obs 後 `prometheus/client_golang` 仍是 yfin 直依,不會從 indirect 消失。

### A.7 go.mod 影響

需刪除:
```
github.com/AmpyFin/ampy-observability/go/ampyobs v0.0.3
```

ampy-obs 帶進來的 indirect 會被 `go mod tidy` 自動清掉(`go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc`、`otlpmetrichttp`、`otlptrace`、`otlptracegrpc`、`otlptracehttp`、`prometheus/client_golang` v1.19.1、`prometheus/client_model` v0.5.0、`prometheus/common` v0.48.0 等都已在 yfin 升版清單中,`go mod tidy` 會調整)。

`go.opentelemetry.io/otel` 與 `go.opentelemetry.io/otel/trace` 仍會保留(裸 OTel API 仍需使用)。

---

## B. `ampy-config` 移除面

### B.1 `config/types/loader.go` 的兩處 ampyconfig.NewLoader().Load() 呼叫

| 行號 | 程式碼 | 用途 |
|---|---|---|
| **L37-38** | `ampyLoader := ampyconfig.NewLoader(l.effectivePath); configMap, err := ampyLoader.Load()` | `Load()` — 主要 entrypoint,回傳 `map[string]interface{}` |
| **L216-217** | 同上,`GetEffectiveConfig()` 內重複呼叫 | `GetEffectiveConfig` — `--print-effective` 印出用 |

**ampyconfig 本身只是** `os.ReadFile + yaml.Unmarshal → map[string]any`(已讀 module source `/Users/bytedance/.local/go/pkg/mod/github.com/!ampy!fin/ampy-config/go/ampyconfig@v1.1.4/reload.go`):

```go
func (l *Loader) Load() (map[string]any, error) {
    b, err := os.ReadFile(l.Path)
    ...
    yaml.Unmarshal(b, &m)
    return m, nil
}
```

**沒有 secret 解析、沒有檔案合併、沒有 NATS hot-reload**(只有 `Client` struct 有 NATS,跟 Loader 無關)。

**替換方案**:
- 把 `config/types/loader.go` 的 `ampyLoader.Load()` 兩處直接換成 `os.ReadFile + yaml.Unmarshal`,刪掉 `ampyconfig` import。
- 或:把 `ampyconfig.NewLoader/Load` 在 `config/types/loader.go` 內封裝成一個 `loadYAMLFile(path string) (map[string]any, error)` private helper,`Load` 與 `GetEffectiveConfig` 都改呼叫它,程式碼更乾淨。

### B.2 `config/types/adapters.go` 的 getters 是否依賴 ampy-config?

**否**,完全是本地 `*Config` struct getter(已讀全檔):

| 函式 | 行號 | 實作 |
|---|---|---|
| `GetHTTPConfig()` | L22-39 | 從 `c.Yahoo / c.RateLimit / c.Retry / c.CircuitBreaker` 組裝出 `HTTPConfig`,**0 import** |
| `GetBusConfig()` | L42-44 | `return &c.Bus` |
| `GetFXConfig()` | L47-49 | `return &c.FX` |
| `GetScrapeConfig()` | L52-54 | `return &c.Scrape` |
| `ValidateInterval(interval)` | L57-64 | 本地比對 `c.Markets.AllowedIntervals` |
| `ValidateAdjustmentPolicy(policy)` | L67-72 | 本地 enum check |

**結論**:這些 getter 不需修改,刪 ampy-config 後行為完全不變。

### B.3 `config/effective.yaml` 結構

Top-level keys(L1-128 全文):
- `app` (L4-6): `env`, `run_id`
- `yahoo` (L8-13): `base_url`, `timeout_ms`, `idle_timeout_ms`, `max_conns_per_host`, `user_agent`
- `concurrency` (L15-17): `global_workers`, `per_host_workers`
- `rate_limit` (L19-23): `per_host_qps`, `per_host_burst`, `per_session_qps`, `per_session_burst`
- `sessions` (L25-28): `n`, `eject_after`, `recreate_cooldown_ms`
- `retry` (L30-33): `attempts`, `base_ms`, `max_delay_ms`
- `circuit_breaker` (L35-39): `window`, `failure_threshold`, `reset_timeout_ms`, `half_open_probes`
- `markets` (L41-46): `allowed_intervals`, `allowed_mics`, `default_adjustment_policy`
- `fx` (L48-61): `provider`, `target`, `cache_ttl_ms`, `rate_scale`, `rounding`, `yahoo_web` 子樹
- `bus` (L63-86): `enabled`, `env`, `topic_prefix`, `max_payload_bytes`, `publisher` (NATS + Kafka), `retry`, `circuit_breaker`
- `scrape` (L88-105): `enabled`, `user_agent`, `timeout_ms`, `qps`, `burst`, `retry`, `robots_policy`, `cache_ttl_ms`, `endpoints` (5 個 bool)
- `observability` (L107-118): `logs.level`, `metrics.prometheus.{enabled,addr}`, `tracing.otlp.{enabled,endpoint,sample_ratio}`
- `secrets` (L120-128): `[{name, ref, required}, ...]`

**env 變數插值**:`${NATS_URL:-nats://localhost:4222}` (L71) + `${OTEL_EXPORTER_OTLP_ENDPOINT:-http://localhost:4317}` (L117) — 兩處都靠 `config/types/loader.go:interpolateString` 處理,**不依賴** ampy-config。

**Secrets 區**:`ref: "env:NATS_PASSWORD"` 與 `ref: "file:/etc/ampy/secret/kafka.pass"`(L121-127)目前是**占位符**,`config/types/loader.go` 內**只做 `secrets[].ref` 的 redact 標記**,**不解析** `env:` / `file:` URI。`loader.go:131-145` 的 `redactSecrets` 只把 `ref` 改成 `"[REDACTED]"` 給 `--print-effective` 印出。

**結論**:
- 刪 ampy-config **不會**破壞 secrets 行為(本來就沒實作 secret resolution)。
- 如果未來要真的解析 `env:VAR` / `file:/path` → 自行在 `config/types/loader.go` 加 helper(約 20 行),不需引外部套件。

### B.4 其他耦合點(grep 結果)

| 檔案 | 行號 | 內容 |
|---|---|---|
| `/Users/bytedance/projects/yfin/config/types/loader.go` | L17 | `import "github.com/AmpyFin/ampy-config/go/ampyconfig"` — **唯一 import 點** |
| `/Users/bytedance/projects/yfin/cmd/admin/admin.go` | L67-75 | 用 `types.NewLoader(...).Load()` + `GetEffectiveConfig()` — **間接**呼叫,改 types 即可 |
| `/Users/bytedance/projects/yfin/cmd/client.go` | L32-34, L90-92 | 用 `types.NewLoader(...).Load()` — **間接**呼叫,改 types 即可 |
| `/Users/bytedance/projects/yfin/cmd/fundamentals/profile.go` | L54-55 | 同上 |
| `/Users/bytedance/projects/yfin/cmd/fundamentals/stats.go` | L55-56 | 同上 |
| `/Users/bytedance/projects/yfin/cmd/market/pull.go` | L105-106 | 同上 |
| `/Users/bytedance/projects/yfin/cmd/scrape/scrape_run.go` | L79-80 | 同上 |
| `/Users/bytedance/projects/yfin/cmd/soak/main.go` | L75-76 | 同上(獨立 binary,但呼叫同一個 `types.NewLoader`) |
| `/Users/bytedance/projects/yfin/config/ampy_config.go` | L11 | `import "github.com/bizshuk/yfin/config/types"` — 只是個 type alias,沒問題 |
| `/Users/bytedance/projects/yfin/cmd/global.go` | L39 | 註解提到 "ampy-config file" — doc-only,改文字 |
| `/Users/bytedance/projects/yfin/cmd/soak/main.go` | L45 | 註解同上 |
| `/Users/bytedance/projects/yfin/cmd/admin/admin.go` | L33, L39, L66 | 註解/help 文字同上 |
| `/Users/bytedance/projects/yfin/cmd/client.go` | L24, L32, L90 | 註解同上 |
| `/Users/bytedance/projects/yfin/facade/scrape.go` | L62 | 註解 "ampy-config's flat ScrapeConfig" — doc-only |
| `/Users/bytedance/projects/yfin/config/example.dev.yaml` | L1, L102 | YAML 註解 — doc-only |
| `/Users/bytedance/projects/yfin/config/example.prod.yaml` | L1 | 同上 |
| `/Users/bytedance/projects/yfin/config/example.staging.yaml` | L1 | 同上 |

**結論**:
- 唯一需要改的 Go 程式碼點:`config/types/loader.go`(L17 import + L37-38 + L216-217 兩處呼叫)。
- 上述 7 個 Go 檔案全部是**間接**呼叫,只要 `types.NewLoader` / `Loader.Load` / `GetEffectiveConfig` 公開 API 保留,這些檔案**不需改**。
- 純文字註解/help 文字:**不阻擋編譯**,可選改或保留。

### B.5 go.mod 影響

需刪除:
```
github.com/AmpyFin/ampy-config/go/ampyconfig v1.1.4
```

ampy-config 帶的 indirect `nats-io/nats.go` v1.35.0 / `klauspost/compress` v1.17.2 / `nats-io/nkeys` v0.4.7 / `nats-io/nuid` v1.0.1 / `golang.org/x/crypto` v0.18.0 — 這些都已被 yfin 升版(`nats.go` v1.47.0、`compress` v1.18.2),`go mod tidy` 會自然清掉舊版。

`gopkg.in/yaml.v3` 仍保留(yfin 直 import,在 `config/types/loader.go:18`)。

---

## 替換計畫摘要

### 階段 1 — `ampy-config` 移除(較簡單,工作量小)

1. 改寫 `config/types/loader.go`:
   - 刪 L17 `ampyconfig` import
   - 在檔案內加 private helper `loadYAMLFile(path string) (map[string]any, error) { os.ReadFile + yaml.Unmarshal }`
   - L37-38 / L216-217 改呼叫 `l.loadYAMLFile(l.effectivePath)`
2. 跑 `go mod tidy` 清 indirect
3. 跑既有 unit tests 確認 `config/types/loader_test.go` 通過
4. 7 個 `cmd/` 檔案 + 1 個 `cmd/soak/main.go` **不需改**(API 沒變)
5. (選)更新 6 處 doc 註解/help 文字,從 "ampy-config" 改為 "yaml config" 之類

### 階段 2 — `ampy-observability` 移除(較大,需替換 tracing/logger 部分)

1. 改寫 `utils/obsv/obsv.go`:
   - 刪 `ampyobs` import(L11)
   - 刪 `ampyConfig` 欄位(L41)
   - `Init`(L52) 改用裸 OTel:
     - `sdktrace.NewTracerProvider(WithBatcher(otlpgrpc.New(...)))` 或 HTTP exporter
     - `sdkmetric.NewMeterProvider(WithReader(otlpmetricgrpc.New()))`(用 OTel Metrics 取代,但**注意:yfin 內部 metric 還是走 Prometheus**)
     - `slog.SetDefault(...)` 取代 `ampyobs.L()`
   - `Logger()`(L121) 回傳 `slog.Default()`(或一個包裝過的 logger)
   - `Tracer()`(L132) 改用本地 `otel.GetTracerProvider().Tracer("yfinance-go")`(`globalObsv` 為 nil 時 fallback 到 `noop`)
   - `StartSpan`(L144) 改用 `Tracer().Start(ctx, name, opts...)`
   - Span factory / `RecordSpanError` / `LogWithTrace` / `CommonLogAttrs` **不需改**(本來就只用裸 OTel)
2. `utils/obsv/metrics.go` **不需改**(本地 Prometheus)
3. `utils/obsv/{obsv_test.go, integration_test.go}` **可能不需改**(公開介面不變)
4. `utils/httpx/client.go` 的 23 處 `obsv.*` 呼叫**不需改**(`obsv` 公開 API 保留)
5. 4 個 cmd entrypoint(`profile.go / stats.go / scrape_run.go / pull.go`)的 obsv bootstrap 區段**不需改**(仍呼叫 `obsv.Init` / `obsv.Shutdown`,但實作已換)
6. 跑 `go mod tidy` 清掉 ampy-obs 帶的舊版 OTel exporter indirect
7. 跑 `go test ./...` 確認 `utils/obsv` 整合測試通過

### 階段 3 — 驗證

- `go build ./...`
- `go vet ./...`
- `go test ./...`
- 手動驗證:`./yfin --config config/effective.yaml config --print-effective` 確認印出含 secrets redact
- 手動驗證:`./yfin pull --ticker AAPL --start ... --end ... --preview` 確認 Prometheus metrics 從 `:9090/metrics` 仍可拉
- (可選)啟 OTel collector 確認 trace 仍送出

---

## 風險與注意

1. **`utils/obsv/obsv.go:140` 目前的 `Tracer()` 本來就是 noop**(`// ampy-observability doesn't expose tracer directly, use context logger`),所以刪 ampy-obs 後行為**不變**,這個註解也可以順手清掉。
2. **OTel exporter 帶來的 indirect**:`ampy-obs` 用 OTLP gRPC + HTTP exporter,若改寫 `obsv.go` 沿用同樣 exporter,indirect 仍會存在(只是版本由 yfin 自身 require 控制,不走 ampy-obs)。
3. **不要刪 `utils/obsv/` 目錄本身**:它的 `metrics.go` 是專案自有的 Prometheus 基礎設施,跟 ampy-obs 無關。只需要把 `obsv.go` 改寫。
4. **`config/effective.yaml` 的 `secrets:` 區**目前只 redact 不解析;若團隊原本期望 ampy-config 會自動解析 `env:VAR`,這是**文件 vs 實作落差**——刪 ampy-config 後必須明確補上(或不補,在文件說明「secret 仍由 deployment 環境注入」)。
5. **`ampybus.Envelope` 的 trace context**:`utils/bus/publisher.go` 內自己處理 envelope 內的 trace header,不依賴 `ampyobs.NewTracedBus`,刪 ampy-obs 後 bus publish 的 trace 仍能運作(走裸 OTel context propagation)。
6. **go.mod 的 `// indirect` 清單**:`go mod tidy` 會自動處理,但刪 ampy-obs 後 `prometheus/client_golang` v1.19.1 仍可能以 indirect 殘留(因為 yfin 直 require v1.23.2,tidy 會升級)。

---

## 關鍵檔案路徑(absolute)

### 需要修改
- `/Users/bytedance/projects/yfin/config/types/loader.go` — 刪 ampyconfig import,加 `loadYAMLFile` helper
- `/Users/bytedance/projects/yfin/utils/obsv/obsv.go` — 刪 ampyobs import,改 `Init/Shutdown/Logger/Tracer/StartSpan` 走裸 OTel + slog
- `/Users/bytedance/projects/yfin/go.mod` — 刪 2 行 `require`
- `/Users/bytedance/projects/yfin/go.sum` — `go mod tidy` 自動清

### 不需修改(API 仍保留)
- `/Users/bytedance/projects/yfin/utils/obsv/metrics.go`
- `/Users/bytedance/projects/yfin/utils/obsv/obsv_test.go`
- `/Users/bytedance/projects/yfin/utils/obsv/integration_test.go`
- `/Users/bytedance/projects/yfin/utils/httpx/client.go`
- `/Users/bytedance/projects/yfin/cmd/fundamentals/profile.go`
- `/Users/bytedance/projects/yfin/cmd/fundamentals/stats.go`
- `/Users/bytedance/projects/yfin/cmd/scrape/scrape_run.go`
- `/Users/bytedance/projects/yfin/cmd/market/pull.go`
- `/Users/bytedance/projects/yfin/cmd/admin/admin.go`
- `/Users/bytedance/projects/yfin/cmd/client.go`
- `/Users/bytedance/projects/yfin/cmd/global.go`(只 doc 註解)
- `/Users/bytedance/projects/yfin/cmd/soak/main.go`(只 doc 註解)
- `/Users/bytedance/projects/yfin/facade/scrape.go`(只 doc 註解)
- `/Users/bytedance/projects/yfin/config/types/adapters.go`
- `/Users/bytedance/projects/yfin/config/ampy_config.go`
- `/Users/bytedance/projects/yfin/config/effective.yaml`
- `/Users/bytedance/projects/yfin/config/example.{dev,prod,staging}.yaml`(只 YAML header 註解)
