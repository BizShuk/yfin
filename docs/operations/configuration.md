# 組態參考 (Configuration Reference)

`yfin` 的所有可調行為 — HTTP 連線、速率限制、重試、熔斷、scrape 引擎、FX 換算、觀測、密鑰 — 都透過 YAML 設定檔統一管理。單一 YAML 檔即承載全部設定。

## 1. 總覽 (Overview)

### 覆寫優先順序 (Precedence)

由高至低，依序覆寫：

```
CLI Flags > Environment Variables > Config File > Built-in Defaults
```

- CLI flag：root 級 persistent flags，例如 `--log-level` / `--qps` / `--timeout` / `--concurrency` / `--retry-max`，在任何子命令之前解析。
- 環境變數 (Environment Variables)：僅供 log / observability 設定使用（走 `gosdk` 的 `APP_*` 慣例）；YAML 內 `${VAR}` / `${VAR:-default}` 插值則於 `Loader.Load()` 階段展開。
- 設定檔 (Config File)：單一 YAML（`config/effective.yaml` 或 `--config` 指定路徑），由內建 loader 載入。
- 預設值 (Defaults)：YAML 缺欄位時由 `config/types/config.go` 內建。

### 設定檔位置 (File Locations)

| 路徑 | 用途 |
| --- | --- |
| `config/effective.yaml` | 預設設定檔（相對於工作目錄） |
| `./config/example.dev.yaml` | 開發環境範本 |
| `./config/example.staging.yaml` | 預上線環境範本 |
| `./config/example.prod.yaml` | 生產環境範本 |
| `--config <path>` | 自訂路徑（root flag） |

> `gosdk/config` 不接受自訂 app config dir；環境切換透過 `APP_*` 環境變數或切換工作目錄。

## 2. CLI 旗標 (CLI Flags)

以下為 root 級 persistent flags（`cmd/global.go`），每個子命令都會繼承：

| 旗標 | 型別 | 預設 | 用途 |
| --- | --- | --- | --- |
| `--config` | string | `""` | YAML 設定檔路徑 |
| `--log-level` | string | `"info"` | Log level：`info` / `debug` / `warn` / `error` |
| `--run-id` | string | `""` | Run ID（追蹤用；空白則自動產生） |
| `--concurrency` | int | `0` | Worker pool 大小（覆寫 YAML；`0` 表示沿用設定） |
| `--qps` | float | `0` | Per-host QPS（覆寫 YAML） |
| `--retry-max` | int | `0` | HTTP retry 嘗試次數 |
| `--sessions` | int | `0` | **Vestigial**：HTTP client 已無 session rotation；保留旗標為向後相容，無實際效果 |
| `--timeout` | duration | `0` | HTTP timeout（如 `6s`、`30s`） |
| `--observability-disable-tracing` | bool | `false` | 關閉 OpenTelemetry tracing |
| `--observability-disable-metrics` | bool | `false` | 關閉 Prometheus metrics |

## 3. 設定檔結構 (Config File)

`config/effective.yaml` 為樹狀 YAML；頂層直接對應 `Config` struct（`config/types/config.go`）：

| Top-level Key | 職責 |
| --- | --- |
| `app` | 環境標籤（`dev` / `staging` / `prod`）與 run 識別 |
| `yahoo` | Yahoo API base URL、HTTP timeout、connection pool |
| `concurrency` | Global / per-host worker pool 大小 |
| `rate_limit` | Per-host / per-session QPS 與 burst |
| `sessions` | Session pool 大小（仍保留欄位但已無作用，見 `--sessions` vestigial 說明） |
| `retry` | HTTP retry 通用設定（`utils/httpx` 使用） |
| `circuit_breaker` | 失敗熔斷閾值與 reset timeout |
| `markets` | Allowed intervals（`["1d"]` daily-only）、MIC allowlist、adjustment policy |
| `fx` | FX 換算（provider、target、cache、Yahoo web fallback） |
| `scrape` | 網頁爬蟲 fallback（Yahoo web 端點、robots policy） |
| `observability` | log level、Prometheus、OTLP tracing |
| `secrets` | URI 參照清單（`env:` / `file:` / `gcp-sm:` / `aws-sm:` / `secret:`） |

### 範例片段 (Excerpt)

```yaml
app:
  env: "dev"
  run_id: ""

yahoo:
  base_url: "https://query2.finance.yahoo.com"
  timeout_ms: 6000
  max_conns_per_host: 64
  user_agent: "yfin/1.x"
```

完整 default 值見 `config/effective.yaml`；四種環境差異範本見 `config/example.{dev,staging,prod}.yaml`。

## 4. Effective Config 檢視 (Effective Config Inspection)

`yfin config-effective` 透過 `config.Loader.GetEffectiveConfig()` 還原載入後、插值完成、敏感值 redact 後的最終設定。

```bash
# dot-notation key=value（排序、redact secrets）
yfin config-effective

# JSON 格式
yfin config-effective --json

# 載入自訂 YAML（驗證無誤即印出）
yfin --config ./config/example.dev.yaml config-effective
```

範例輸出：

```
EFFECTIVE CONFIG (redacted)
app.env=dev
app.run_id=
yahoo.base_url=https://query2.finance.yahoo.com
...
scrape.robots_policy=enforce
```

> Redact 規則：`secrets[].ref` 與 key 名稱匹配 `password` / `token` / `api_key` / `secret` / `key` 之值會自動替換為 `[REDACTED]`。

## 5. Scrape 設定 (Scrape-Specific Config)

對應 `ScrapeConfig`（`config/types/scrape.go`），控制 `svc/scrape` 引擎：

### 完整 scrape 區段

```yaml
scrape:
  enabled: true
  user_agent: "Mozilla/5.0 (yfin scraper)"
  timeout_ms: 15000
  qps: 3.0
  burst: 5
  retry:
    attempts: 4
    base_ms: 300
    max_delay_ms: 4000
  robots_policy: "enforce"        # enforce | warn | ignore
  cache_ttl_ms: 60000
  endpoints:
    key_statistics: true
    financials: true
    analysis: true
    profile: true
    news: true
```

### robots_policy 三種模式

| 模式 | 行為 |
| --- | --- |
| `"enforce"` | 嚴格遵守 robots.txt；違規時禁止抓取（生產推薦） |
| `"warn"` | 違規時 log warning 但放行（開發） |
| `"ignore"` | 完全跳過 robots.txt 檢查（僅測試） |

### Retry 結構（`ScrapeRetryConfig`）

| 欄位 | 型別 | 預設 | 說明 |
| --- | --- | --- | --- |
| `scrape.retry.attempts` | int | `4` | 最大重試次數（`>= 1`） |
| `scrape.retry.base_ms` | int | `300` | 基礎 backoff（毫秒） |
| `scrape.retry.max_delay_ms` | int | `4000` | 最大 backoff 上限（毫秒） |

> `ScrapeRetryConfig` 僅含上述三欄位；`backoff_multiplier` 與 `jitter` 並非此結構之欄位，倍數與抖動由 `utils/httpx` 內部計算。

### Endpoints 結構（`ScrapeEndpointConfig`）

| 欄位 | 對應 CLI `--endpoint` 名稱 |
| --- | --- |
| `scrape.endpoints.key_statistics` | `key-statistics` |
| `scrape.endpoints.financials` | `financials` |
| `scrape.endpoints.analysis` | `analysis` |
| `scrape.endpoints.profile` | `profile` |
| `scrape.endpoints.news` | `news` |

> CLI 額外可指定 `balance-sheet` / `cash-flow` / `analyst-insights`，共用 financials / key-statistics 的 extractor；這幾項未於 YAML 中獨立列出。

## 6. FX / Observability

### FX（`fx` 區段）

`FXConfig` 控制報價 / bars 的目標幣別換算：

- `fx.provider`：`none` / `yahoo-web`。
- `fx.target`：預設目標幣別（如 `EUR` / `JPY`）。
- `fx.cache_ttl_ms`：換算結果快取 TTL。
- `fx.yahoo_web.*`：Yahoo web FX 端點的 sub-rate-limit 與 retry。

### Observability（`observability` 區段）

`ObservabilityConfig` 控制 log / Prometheus / OTLP：

- `observability.logs.level`：覆寫 `gosdk` 初始 log level。
- `observability.metrics.prometheus.enabled` + `addr`：Prometheus exporter bind address。
- `observability.tracing.otlp.enabled` + `endpoint` + `sample_ratio`：OTLP tracing 端點與取樣率。

CLI 端可用 `--observability-disable-tracing` / `--observability-disable-metrics` 完全關閉對應收集器。

## 7. 驗證規則 (Validation Rules)

`config.Loader.validate()` 於 `Load()` 啟動時強制檢查：

| 規則 | 條件 | 失敗訊息 |
| --- | --- | --- |
| `concurrency.global_workers >= per_host_workers` | 必成立 | `must be >= per_host_workers` |
| `concurrency.per_host_workers >= sessions.n` | 必成立 | `must be >= sessions.n` |
| `markets.allowed_intervals == ["1d"]` | 嚴格等於 | `must be exactly ["1d"] for yfin (daily-only scope)` |
| `markets.default_adjustment_policy ∈ {raw, split_dividend}` | 列舉 | `must be 'raw' or 'split_dividend'` |
| `retry.attempts >= 1` | 範圍 | `must be >= 1` |
| `circuit_breaker.failure_threshold ∈ (0, 1]` | 範圍 | `must be between 0 and 1` |
| `observability.metrics.prometheus.addr`（啟用時） | 必填 | `is required when prometheus is enabled` |
| `observability.tracing.otlp.endpoint`（啟用時） | 必填 | `is required when OTLP tracing is enabled` |

任何違規都會在啟動時印錯並以 exit code `3`（`ExitConfigError`）退出。

## 8. 延伸閱讀 (See Also)

- `docs/install.md`：安裝與初次設定流程。
- `docs/usage.md`：各子命令 CLI 範例。
- `docs/scrape/cli.md`：`yfin scrape` 4 種 mode 與 endpoint 選擇。
- `docs/scrape/overview.md`：scrape 引擎架構與 proxy / payload 處理。
- `docs/operations/observability.md`：metric / log / trace 對接 `inf`（VictoriaMetrics / Loki / OTLP collector）。
